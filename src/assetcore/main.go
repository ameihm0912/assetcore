// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	"flag"
	"sync"
	"time"
)

var cfg acConfig
var wgLogger sync.WaitGroup
var wgMain sync.WaitGroup

func failNotify() {
	cfg.failNotifyChan <- true
}

func fetchPreviousHints() {
	now := time.Now().UTC()
	upuntil := now.Add(-1 * (time.Minute * 5))
	startFrom := upuntil.Add(-1 * (time.Minute * time.Duration(cfg.previousMinutes)))
	windowSize := time.Minute * 30

	curStart := startFrom
	curEnd := curStart.Add(windowSize)
	for {
		select {
		case <-cfg.exitPreviousChan:
			logMessage("fetchPreviousHints: exit notification")
			return
		default:
		}

		if curEnd.After(upuntil) {
			curEnd = upuntil
		}

		err := fetchHints(curStart, curEnd, cfg.exitPreviousChan)
		if err != nil {
			logMessage("fetchPreviousHints: %v", err)
			failNotify()
			return
		}

		if curEnd == upuntil {
			break
		}

		curStart = curEnd
		curEnd = curEnd.Add(windowSize)
	}

	logMessage("fetchPreviousHints: done")
}

func main() {
	defer func() {
		wgMain.Wait()
		close(cfg.logChan)
		wgLogger.Wait()
	}()

	cfg.setDefaults()

	flag.BoolVar(&cfg.foreground, "f", false, "run in foreground")
	flag.IntVar(&cfg.previousMinutes, "p", 480, "begin hints fetch from now - mins")
	flag.IntVar(&cfg.maxDocuments, "s", 10000, "max documents to fetch per request")
	flag.Parse()

	wgLogger.Add(1)
	go func() {
		logger()
		wgLogger.Done()
	}()

	wgMain.Add(1)
	go func() {
		fetchPreviousHints()
		wgMain.Done()
	}()

	wgMain.Add(1)
	go func() {
		hintsLoop()
		wgMain.Done()
	}()

	wgMain.Add(1)
	go func() {
		corLoop()
		wgMain.Done()
	}()

	<-cfg.failNotifyChan
	cfg.exitPreviousChan <- true
	cfg.exitHintsChan <- true
	cfg.exitCorChan <- true
}
