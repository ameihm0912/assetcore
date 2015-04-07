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
	"fmt"
	elastigo "github.com/mattbaird/elastigo/lib"
	"strings"
	"time"
)

type AssetHint struct {
	Category  string           `json:"category"`
	Hostname  string           `json:"hostname"`
	Severity  string           `json:"severity"`
	Pid       string           `json:"processid"`
	ProcName  string           `json:"processname"`
	Summary   string           `json:"summary"`
	Tags      []string         `json:"tags"`
	Timestamp time.Time        `json:"utctimestamp"`
	Details   AssetHintDetails `json:"details"`
}

type AssetHintDetails struct {
	Hostname   string   `json:"hostname"`
	IPv4       []string `json:"ipv4"`
	IPv6       []string `json:"ipv6"`
	NexAssetId string   `json:"nexassetid"`
	MAC        []string `json:"macaddress"`
	Software   []string `json:"software"`

	// MIG compatibility
	HostnameMig string `json:"name"`
}

func (h *AssetHint) sanitize() {
	// If the IPv4 address has been stored in the hint with a subnet
	// mask, strip that out. We could probably incorporate knowledge
	// of the subnet mask into the asset.
	for i, x := range h.Details.IPv4 {
		r := strings.Index(x, "/")
		if r != -1 {
			h.Details.IPv4[i] = x[:r]
		}
	}

	if len(h.Details.HostnameMig) != 0 {
		h.Details.Hostname = h.Details.HostnameMig
	}
}

func hintsLoop() {
	defer func() {
		close(cfg.hintsChan)
	}()

	now := time.Now().UTC()

	qs := now.Add(-1 * (time.Minute * 5))
	qe := qs.Add(time.Minute)
	for {
		tchan := make(chan bool)
		go func() {
			time.Sleep(time.Second * 30)
			tchan <- true
		}()
		select {
		case <-tchan:
		case <-cfg.exitHintsChan:
			logMessage("hintsLoop: exit notification")
			return
		}

		select {
		case <-cfg.exitHintsChan:
			logMessage("hintsLoop: exit notification")
			return
		default:
		}

		err := fetchHints(qs, qe, cfg.exitHintsChan)
		if err != nil {
			logMessage("hintsLoop: %v", err)
			failNotify()
			return
		}
		qs = qe
		qe = time.Now().UTC().Add(-1 * (time.Minute * 4))
	}
}

func fetchHints(start time.Time, end time.Time, exitchan chan bool) error {
	logMessage("fetchHints: %v -> %v", start, end)

	conn := elastigo.NewConn()
	conn.Domain = cfg.hintsHost

	template := `{
		"size": %v,
		"query": {
			"term": {
				"category": "asset_hint"
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
	tempbuf := fmt.Sprintf(template, cfg.maxDocuments, start.Format(time.RFC3339), end.Format(time.RFC3339))

	res, err := conn.Search(cfg.hintsIndex, "event", nil, tempbuf)
	if err != nil {
		return err
	}
	haveHints := res.Hits.Len()
	logMessage("fetchHints: %v -> %v: %v", start, end, haveHints)
	for _, x := range res.Hits.Hits {
		var nh AssetHint
		err = json.Unmarshal(*x.Source, &nh)
		if err != nil {
			return err
		}
		select {
		case cfg.hintsChan <- nh:
		case <-exitchan:
			return errors.New("fetchHints: exit channel notification")
		}
	}

	return nil
}
