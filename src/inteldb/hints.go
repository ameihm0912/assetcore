// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package inteldb

import (
	"encoding/json"
	"fmt"
	"time"
)

type AssetHint struct {
	Category  string    `json:"category"`
	Hostname  string    `json:"hostname"`
	Severity  string    `json:"severity"`
	Pid       string    `json:"processid"`
	ProcName  string    `json:"processname"`
	Summary   string    `json:"summary"`
	Tags      []string  `json:"tags"`
	Timestamp time.Time `json:"utctimestamp"`
	Details   AssetHintDetails
}

type AssetHintDetails struct {
	Hostname   string   `json:"hostname"`
	IPv4       []string `json:"ipv4"`
	IPv6       []string `json:"ipv6"`
	NexAssetId string   `json:"nexassetid"`
	MAC        []string `json:"macaddress"`
	Software   []string `json:"software"`

	// MIG hint compatibility
	HostnameMig string `json:"name"`
}

type HintsMessage struct {
	Hint AssetHint
	Err  error
	Log  string
}

type HintsConn struct {
	idbConn idbES
}

func (h *HintsConn) Init(host string, index string) (err error) {
	h.idbConn.init(host)
	err = h.idbConn.setIndex(index, false)
	if err != nil {
		return err
	}

	return nil
}

func (h *HintsConn) Search(template string) (ret []AssetHint, err error) {
	res, err := h.idbConn.search(template, "event")
	if err != nil {
		return ret, err
	}
	for _, x := range res {
		var nh AssetHint
		err = json.Unmarshal(x, &nh)
		ret = append(ret, nh)
	}
	return ret, nil
}

func (h *HintsConn) HintsFetch(hintsChan chan HintsMessage, doneChan chan bool, startAt time.Time) {
	defer func() {
		doneChan <- true
	}()

	// Start by fetching blocks of hints from the past in one hour intervals, until we reach
	// the current time - 5 minutes.
	now := time.Now().UTC()
	endpoint := now.Add(-1 * (time.Minute * 5))
	qs := startAt
	var qe time.Time
	window := time.Hour
	last := false
	for {
		qe = qs.Add(window)
		if qe.After(endpoint) {
			qe = endpoint
			last = true
		}

		template := createHintsTemplate(qs, qe, maxDocuments)
		res, err := h.Search(template)
		if err != nil {
			hintsChan <- HintsMessage{Err: err}
			return
		}
		for _, x := range res {
			hintsChan <- HintsMessage{Hint: x}
		}

		if last {
			break
		}

		qs = qe
	}
	hintsChan <- HintsMessage{Log: "completed hints prefetch"}

	// endpoint will be our start position, and we will always fetch one minute worth
	// of data from this period, remaining approximately 4 - 5 minutes behind ES to
	// ensure new data has been indexed before we try to read it.
	startpoint := endpoint
	for {
		now = time.Now().UTC()
		we := now.Add(-1 * (time.Minute * 4))
		time.Sleep(time.Minute)
		qdiff := we.Sub(startpoint)
		delaydelta := now.Sub(we)
		hintsChan <- HintsMessage{Log: fmt.Sprintf("hints fetch %v -> %v", startpoint, we)}
		hintsChan <- HintsMessage{Log: fmt.Sprintf("qdiff %v, delaydelta %v", qdiff, delaydelta)}

		template := createHintsTemplate(startpoint, we, maxDocuments)
		res, err := h.Search(template)
		if err != nil {
			hintsChan <- HintsMessage{Err: err}
			return
		}
		for _, x := range res {
			hintsChan <- HintsMessage{Hint: x}
		}

		startpoint = we
	}
}

func createHintsTemplate(start time.Time, end time.Time, size int) (ret string) {
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
	ret = fmt.Sprintf(template, size, start.Format(time.RFC3339), end.Format(time.RFC3339))

	return ret
}
