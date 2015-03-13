// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package main

import (
	"strings"
	"time"
)

type assetHint struct {
	Category  string    `json:"category"`
	Hostname  string    `json:"hostname"`
	Severity  string    `json:"severity"`
	Pid       string    `json:"processid"`
	ProcName  string    `json:"processname"`
	Summary   string    `json:"summary"`
	Tags      []string  `json:"tags"`
	Timestamp time.Time `json:"utctimestamp"`
	Details   assetHintDetails
}

type assetHintDetails struct {
	Hostname   string   `json:"hostname"`
	IPv4       []string `json:"ipv4"`
	IPv6       []string `json:"ipv6"`
	NexAssetId string   `json:"nexassetid"`
	MAC        []string `json:"macaddress"`
	Software   []string `json:"software"`

	// Some compatability for the way MIG sends hint data, we should
	// standardize on fields later.
	HostnameMig string `json:"name"`
}

func (h *assetHint) sanitize() {
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
