// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

type acConfig struct {
	logChan          chan string
	hintsChan        chan AssetHint
	failNotifyChan   chan bool
	exitPreviousChan chan bool
	exitHintsChan    chan bool
	exitCorChan      chan bool

	foreground      bool
	maxDocuments    int
	previousMinutes int

	hintsIndex string
	hintsHost  string
}

func (c *acConfig) setDefaults() {
	c.foreground = false
	c.maxDocuments = 10000
	c.previousMinutes = 480

	c.hintsIndex = "events"
	c.hintsHost = "eshost"

	c.logChan = make(chan string)
	c.hintsChan = make(chan AssetHint)
	// failNotifyChan:
	// Notification channel to collect goroutine failure indications; the
	// buffer size should be large enough we never block a sender
	c.failNotifyChan = make(chan bool, 10)
	c.exitPreviousChan = make(chan bool, 1)
	c.exitHintsChan = make(chan bool, 1)
	c.exitCorChan = make(chan bool, 1)
}
