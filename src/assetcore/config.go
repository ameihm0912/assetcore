// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package main

import (
	"time"
)

type acConfig struct {
	chhints      chan assetHint
	chcore       chan bool
	chcoreworker chan bool
	chlogger     chan string
	chloggerexit chan bool
	assetIndex   string
	esHost       string
	maxHits      int
	window       time.Duration
}

func (cfg *acConfig) setDefaults() {
	cfg.assetIndex = "assets"
	cfg.esHost = "eshost"
	cfg.maxHits = 50
	cfg.window = time.Hour * 8
	cfg.chhints = make(chan assetHint)
	cfg.chcore = make(chan bool)
	cfg.chcoreworker = make(chan bool)
	cfg.chlogger = make(chan string)
	cfg.chloggerexit = make(chan bool)
}
