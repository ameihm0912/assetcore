// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	"flag"
	idb "inteldb"
	"time"
)

var cfg acConfig

var idbconn idb.InteldbConn
var hintsconn idb.HintsConn

func main() {
	cfg.setDefaults()

	flag.BoolVar(&cfg.foreground, "f", false, "run and log in foreground")
	flag.IntVar(&cfg.previousMinutes, "p", 480, "begin hints fetch from now - minutes")
	flag.IntVar(&cfg.maxDocuments, "s", 10000, "maximum documents to fetch per request")
	flag.Parse()

	doneChan := loggerNotification()
	defer func() {
		<-doneChan
	}()
	cfg.logChan = make(chan string)
	go logger(cfg.logChan)
	defer func() {
		close(cfg.logChan)
	}()

	logMessage("assetcore initializing")

	logMessage("initializing inteldb index")
	err := idbconn.Init(cfg.inteldbHost, cfg.inteldbIndex)
	if err != nil {
		logMessage("initializing inteldb index: %v", err)
		return
	}

	logMessage("initializing hints index")
	err = hintsconn.Init(cfg.hintsHost, cfg.hintsIndex)
	if err != nil {
		logMessage("initializing hints index: %v", err)
		return
	}
	idb.SetMaxDocuments(cfg.maxDocuments)

	logMessage("spawning hints fetch goroutine")
	startAt := time.Now().UTC().Add(-1 * (time.Duration(cfg.previousMinutes) * time.Minute))
	logMessage("hints fetch will start at %v", startAt)
	go hintsconn.HintsFetch(cfg.hintsChan, cfg.hintsChanDone, startAt)

	doexit := false
	for {
		select {
		case nh := <-cfg.hintsChan:
			if nh.Err != nil {
				logMessage("hints channel: %v", nh.Err)
				doexit = true
				break
			} else if len(nh.Log) != 0 {
				logMessage(nh.Log)
				break
			}
			logMessage("%v", nh.Hint)
		}

		if doexit {
			break
		}
	}

	<-cfg.hintsChanDone
	logMessage("hints fetch has returned")
}
