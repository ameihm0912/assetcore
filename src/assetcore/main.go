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
	"flag"
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

func pullHintsWorker(start time.Time, end time.Time, cchan chan assetHint) {
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

		// If the cache channel is not nil, that means we are in active
		// cacheing mode. Also send the hint to the hint cache
		// goroutine so it can be written to disk.
		if cchan != nil {
			cchan <- h
		}
	}
}

func pullAssets() error {
	template := `{
		"size": %v
	}`
	sj := fmt.Sprintf(template, cfg.maxAssetHits)

	if cfg.dataCache {
		logmsg("initializing asset block from cache")
		adata, err := cache_get_assets_pre()
		if err != nil {
			logmsg("error reading cached assets: %v", err)
			return err
		}
		if len(adata) > 0 {
			// If cached data was returned from the cache, use that
			// instead of an ES query.
			for _, x := range adata {
				aBlock.addAsset(x)
				aBlock.existedcount += 1
			}
			return nil
		}
		logmsg("no cached assets, will populate this time")
	}

	logmsg("initializing asset block from es")

	haveidx, err := es.IndicesExists("assets")
	if err != nil {
		logmsg("error obtaining index status: %v", err)
		return err
	}
	if !haveidx {
		return nil
	}

	res, err := es.Search("assets", "asset", nil, sj)
	if err != nil {
		logmsg("error fetching assets: %v", err)
		return err
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
	for _, x := range res.Hits.Hits {
		var a asset
		err = json.Unmarshal(*x.Source, &a)
		if err != nil {
			logmsg("error unmarshalling asset: %v", err)
			return err
		}
		aBlock.addAsset(a)
		aBlock.existedcount += 1
	}
	logmsg("assets inserted into asset block")

	if cfg.dataCache {
		logmsg("writing asset pre cache")
		for _, x := range aBlock.assets {
			err = cache_asset(x, true)
			if err != nil {
				logmsg("error cacheing asset: %v", err)
				return err
			}
		}
	}

	return nil
}

func pushAssets() {
	aBlock.Lock()
	defer func() {
		aBlock.Unlock()
	}()

	if cfg.dataCache {
		logmsg("writing post asset cache")
		for _, x := range aBlock.assets {
			err := cache_asset(x, false)
			if err != nil {
				logmsg("error caching post assets: %v", err)
				return
			}
		}
		return
	}

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
}

func pullHints() {
	end := time.Now().UTC()
	start := end.Add(-1 * cfg.window)

	var cchan chan assetHint
	var cchandone chan bool

	if cfg.dataCache {
		logmsg("initializing hints from cache")
		hdata, err := cache_get_hints()
		if err != nil {
			logmsg("error reading cached hints: %v", err)
			doexit(1)
		}
		if len(hdata) > 0 {
			for _, x := range hdata {
				cfg.chhints <- x
			}
			close(cfg.chhints)
			return
		}
		logmsg("no cached hints, will populate this time")

		cchan = make(chan assetHint)
		cchandone = make(chan bool)
		go func() {
			for {
				h, ok := <-cchan
				if !ok {
					break
				}
				err = cache_hint(h)
				if err != nil {
					logmsg("error cacheing hint: %v", err)
					doexit(1)
				}
			}
			cchandone <- true
		}()
	}

	logmsg("hints fetch started")
	index_s := start
	index_e := start.Add(time.Hour)
	for index_s.Before(end) {
		pullHintsWorker(index_s, index_e, cchan)
		index_s = index_s.Add(time.Hour)
		index_e = index_e.Add(time.Hour)
	}

	logmsg("hints fetch complete")
	if cchan != nil {
		close(cchan)
		<-cchandone
	}
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
	logmsg("%v assets existed before run", aBlock.existedcount)
	logmsg("%v assets are new", aBlock.newcount)

	logmsg("pushing updated asset data")
	pushAssets()

	logmsg("asset correlator exiting")
	cfg.chcore <- true
}

func doexit(rc int) {
	cache_close()
	close(cfg.chlogger)
	<-cfg.chloggerexit
	os.Exit(rc)
}

func main() {
	cfg.setDefaults()
	flag.BoolVar(&cfg.dataCache, "c", false, "Enable offline data cache")
	flag.Parse()

	go logger()
	logmsg("assetcore initializing")
	esSetup()

	if cfg.dataCache {
		logmsg("initializing data cache")
		err := cache_init()
		if err != nil {
			logmsg("error initializing cache: %v", err)
			doexit(1)
		}
	}

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
