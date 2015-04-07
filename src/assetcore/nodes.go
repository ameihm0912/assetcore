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

var nodestore []Node

type NodeDescriptor struct {
	NodeType    int    `json:"type"`
	NodeTypeStr string `json:"typestr"`
}

type NodeRelation struct {
	NodeType    NodeDescriptor `json:"type"`
	NodeLinkID  string         `json:"linkid"`
	NodeLinkURL string         `json:"linkurl"`
	lastUpdated time.Time      `json:"lastupdated"`
	confidence  int            `json:"confidence"`
}

type NodeHost struct {
	Hostname string `json:"hostname,omitempty"`
}

type NodeIPv4 struct {
	IPv4 string `json:"ipv4,omitempty"`
}

type Node struct {
	NodeID string `json:"nodeid"`

	NodeType NodeDescriptor `json:"type"`

	lastUpdated time.Time `json:"lastupdated"`
	confidence  int       `json:"confidence"`

	Hostname NodeHost `json:"hostdetails,omitempty"`
	IPv4     NodeIPv4 `json:"ipv4details,omitempty"`
}
