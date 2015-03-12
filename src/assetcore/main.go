// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package main

import (
	"encoding/json"
	"github.com/mattbaird/elastigo/api"
	"github.com/mattbaird/elastigo/search"
	"strconv"
	"time"
)

var assetBlock []asset

var cfg acConfig

func esSetup() {
	api.Domain = cfg.esHost
}

func pullHintsWorker(start time.Time, end time.Time) {
	qs := start.Format(time.RFC3339)
	qe := end.Format(time.RFC3339)
	res, err := search.Search("events").Size(strconv.Itoa(cfg.maxHits)).Filter(
		search.Filter().Terms("category", "asset_hint"),
		search.Range().Field("utctimestamp").From(qs).To(qe),
	).Result()
	if err != nil {
		return
	}
	if res.Hits.Total == 0 {
		return
	}
	for _, x := range res.Hits.Hits {
		var h assetHint
		err = json.Unmarshal(*x.Source, &h)
		if err != nil {
			continue
		}
		cfg.chhints <- h
	}
}

func pullHints() {
	end := time.Now().UTC()
	start := end.Add(-1 * cfg.window)

	index_s := start
	index_e := start.Add(time.Hour)
	for index_s.Before(end) {
		pullHintsWorker(index_s, index_e)
		index_s = index_s.Add(time.Hour)
		index_e = index_e.Add(time.Hour)
	}

	close(cfg.chhints)
}

func processAssetHint(hint assetHint) {
	rulesPipeline(hint)
}

func assetCorWorker(hintbuf []assetHint) {
	for _, x := range hintbuf {
		processAssetHint(x)
	}
	cfg.chcoreworker <- true
}

func assetCorrelator() {
	var hintBuffer []assetHint = nil
	wrkcnt := 0
	for {
		if hintBuffer == nil {
			hintBuffer = make([]assetHint, 0, 250)
		}
		newhint, status := <-cfg.chhints
		if !status {
			break
		}
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

	cfg.chcore <- true
}

func main() {
	cfg.setDefaults()

	esSetup()

	go pullHints()
	go assetCorrelator()
	<-cfg.chcore
}
