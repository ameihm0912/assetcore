// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	elastigo "github.com/mattbaird/elastigo/lib"
	"os"
	"time"
)

var es *elastigo.Conn

var aBlock assetBlock
var cfg acConfig

func esSetup() {
	es = elastigo.NewConn()
	es.Domain = cfg.esHost
}

func pullHintsWorker(start time.Time, end time.Time) {
	qs := start.Format(time.RFC3339)
	qe := end.Format(time.RFC3339)
	logmsg("new hints worker %v -> %v", qs, qe)

	template := `{
		"size": %v,
		"query": {
			"term": {
				category: "asset_hint"
			}
		},
		"filter": {
			"range": {
				"utctimestamp": {
					"from": "%v",
					"to": "%v"
				}
			}
		}
	}`
	sj := fmt.Sprintf(template, cfg.maxHits, qs, qe)

	res, err := es.Search("events", "event", nil, sj)
	if err != nil {
		logmsg("error fetching hints: %v", err)
		return
	}
	if res.Hits.Total == 0 {
		return
	}
	havehints := res.Hits.Len()
	if havehints < res.Hits.Total {
		logmsg("WARNING: some hints not returned, increase maxHits (got %v of %v)", havehints, res.Hits.Total)
	}
	logmsg("hints worker sending %v hits", res.Hits.Len())
	for _, x := range res.Hits.Hits {
		var h assetHint
		err = json.Unmarshal(*x.Source, &h)
		if err != nil {
			continue
		}
		cfg.chhints <- h
	}
}

func pullAssets() error {
	logmsg("initializing asset block from es")

	template := `{
		"size": %v
	}`
	sj := fmt.Sprintf(template, cfg.maxAssetHits)

	res, err := es.Search("assets", "asset", nil, sj)
	if err != nil {
		logmsg("error fetching assets: %v", err)
		return errors.New("error fetching assets")
	}
	if res.Hits.Total == 0 {
		return nil
	}
	haveassets := res.Hits.Len()
	if haveassets < res.Hits.Total {
		logmsg("error: incomplete asset list, will not continue (got %v of %v)", haveassets, res.Hits.Total)
		return errors.New("fetched incomplete asset list")
	}
	logmsg("fetched %v assets", haveassets)
	aBlock.Lock()
	for _, x := range res.Hits.Hits {
		var a asset
		err = json.Unmarshal(*x.Source, &a)
		if err != nil {
			logmsg("error unmarshalling asset: %v", err)
			return errors.New("error unmarshalling asset")
		}
	}
	aBlock.Unlock()
	logmsg("assets inserted into asset block")
	return nil
}

func pushAssets() {
	aBlock.Lock()
	for _, x := range aBlock.assets {
		buf, err := json.Marshal(x)
		if err != nil {
			logmsg("error marshalling asset: %v", err)
			continue
		}
		_, err = es.Index(cfg.assetIndex, "asset", x.AssetID, nil, buf)
		if err != nil {
			logmsg("error indexing asset: %v", err)
		}
	}
	aBlock.Unlock()
}

func pullHints() {
	end := time.Now().UTC()
	start := end.Add(-1 * cfg.window)

	logmsg("hints fetch started")
	index_s := start
	index_e := start.Add(time.Hour)
	for index_s.Before(end) {
		pullHintsWorker(index_s, index_e)
		index_s = index_s.Add(time.Hour)
		index_e = index_e.Add(time.Hour)
	}

	logmsg("hints fetch complete")
	close(cfg.chhints)
}

func processAssetHint(hint assetHint) {
	rulesPipeline(hint)
}

func assetCorWorker(hintbuf []assetHint) {
	logmsg("new correlation worker processing %v hints", len(hintbuf))
	for _, x := range hintbuf {
		processAssetHint(x)
	}
	cfg.chcoreworker <- true
}

func assetCorrelator() {
	var hintBuffer []assetHint = nil
	logmsg("asset correlator started")
	wrkcnt := 0
	for {
		if hintBuffer == nil {
			hintBuffer = make([]assetHint, 0, 250)
		}
		newhint, status := <-cfg.chhints
		if !status {
			break
		}
		newhint.sanitize()
		hintBuffer = append(hintBuffer, newhint)
		if len(hintBuffer) == cap(hintBuffer) {
			wrkcnt += 1
			go assetCorWorker(hintBuffer)
			hintBuffer = nil
		}
	}
	if len(hintBuffer) > 0 {
		wrkcnt += 1
		go assetCorWorker(hintBuffer)
	}

	for wrkcnt > 0 {
		<-cfg.chcoreworker
		wrkcnt -= 1
	}

	logmsg("correlation complete, %v assets in block", aBlock.count)

	logmsg("pushing updated asset data")
	//pushAssets()

	logmsg("asset correlator exiting")
	cfg.chcore <- true
}

func doexit(rc int) {
	close(cfg.chlogger)
	<-cfg.chloggerexit
	os.Exit(rc)
}

func main() {
	cfg.setDefaults()

	go logger()
	logmsg("assetcore initializing")
	esSetup()

	err := pullAssets()
	if err != nil {
		doexit(1)
	}

	go pullHints()
	go assetCorrelator()

	<-cfg.chcore
	logmsg("assetcore exiting")
	doexit(0)
}
