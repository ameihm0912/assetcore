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

const (
	_ = iota
	NODETYPE_IPV4
	NODETYPE_HOSTNAME
)

type nodestore struct {
	nodes []Node
}

var acns nodestore

func (n *nodestore) updateNode(newnode Node) {
}

type NodeDescriptor struct {
	NodeType    int    `json:"type"`
	NodeTypeStr string `json:"typestr"`
}

type NodeRelation struct {
	NodeType    NodeDescriptor `json:"type"`
	NodeLinkID  string         `json:"linkid"`
	NodeLinkURL string         `json:"linkurl"`
	LastUpdated time.Time      `json:"lastupdated"`
	Confidence  int            `json:"confidence"`
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

	LastUpdated time.Time `json:"lastupdated"`
	Confidence  int       `json:"confidence"`

	Hostname NodeHost `json:"hostdetails,omitempty"`
	IPv4     NodeIPv4 `json:"ipv4details,omitempty"`
}

func nodesFromHintIPv4(hint *AssetHint) (nodes []Node) {
	for _, x := range hint.Details.IPv4 {
		var nn Node
		nn.NodeType.NodeType = NODETYPE_IPV4
		nn.NodeType.NodeTypeStr = "ipv4"
		nn.IPv4.IPv4 = x
		nn.LastUpdated = hint.Timestamp
		nn.Confidence = 100
		nodes = append(nodes, nn)
	}
	return nodes
}

func nodesFromHint(hint *AssetHint) (nodes []Node) {
	nodes = append(nodes, nodesFromHintIPv4(hint)...)
	logMessage("%v", nodes)
	return nodes
}
