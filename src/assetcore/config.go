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

	assetIndex     string
	esHost         string
	esExHintsHosts []string
	maxAssetHits   int
	maxHits        int
	window         time.Duration

	dataCache           bool
	cacheAssetsPrePath  string
	cacheAssetsPostPath string
	cacheHintsPath      string
}

func (cfg *acConfig) setDefaults() {
	cfg.assetIndex = "assets"
	cfg.esHost = "eshost"
	cfg.esExHintsHosts = []string{"eshost2"}
	cfg.maxAssetHits = 10000
	cfg.maxHits = 10000
	cfg.window = time.Hour * 8

	cfg.dataCache = false
	cfg.cacheAssetsPrePath = "./assets-pre.acc"
	cfg.cacheAssetsPostPath = "./assets-post.acc"
	cfg.cacheHintsPath = "./hints.acc"

	cfg.chhints = make(chan assetHint)
	cfg.chcore = make(chan bool)
	cfg.chcoreworker = make(chan bool)
	cfg.chlogger = make(chan string)
	cfg.chloggerexit = make(chan bool)
}
