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
	chcoredone chan bool
	chhints    chan assetHint
	chcore     chan assetHint
	esHost     string
	maxHits    int
	window     time.Duration
}

func (cfg *acConfig) setDefaults() {
	cfg.esHost = "eshost"
	cfg.maxHits = 10
	cfg.window = time.Hour * 8
	cfg.chhints = make(chan assetHint)
	cfg.chcore = make(chan assetHint)
	cfg.chcoredone = make(chan bool)
}
