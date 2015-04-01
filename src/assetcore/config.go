// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	idb "inteldb"
)

type acConfig struct {
	logChan       chan string
	hintsChan     chan idb.AssetHint
	hintsChanDone chan bool

	foreground bool

	inteldbIndex string
	inteldbHost  string
	hintsIndex   string
	hintsHost    string

	maxDocuments    int
	previousMinutes int
}

func (c *acConfig) setDefaults() {
	c.foreground = false

	c.inteldbIndex = "inteldb"
	c.inteldbHost = "eshost"
	c.hintsHost = "eshost"
	c.hintsIndex = "events"

	c.previousMinutes = 480
	c.maxDocuments = 10000

	c.hintsChan = make(chan idb.AssetHint)
	c.hintsChanDone = make(chan bool)
}
