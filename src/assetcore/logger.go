// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	"fmt"
	"time"
)

var logNotifyChan chan bool

func loggerNotification() chan bool {
	logNotifyChan = make(chan bool)
	return logNotifyChan
}

func logMessage(s string, args ...interface{}) {
	buf := fmt.Sprintf(s, args...)
	buf = "[" + time.Now().UTC().Format(time.RFC3339) + "] [assetcore] " + buf + "\n"
	cfg.logChan <- buf
}

func logger(ch chan string) {
	defer func() {
		logNotifyChan <- true
	}()

	for {
		s, status := <-ch
		if !status {
			return
		}
		if cfg.foreground {
			fmt.Printf(s)
		}
	}
}
