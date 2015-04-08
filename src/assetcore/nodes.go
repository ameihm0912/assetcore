// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	"code.google.com/p/go-uuid/uuid"
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

func (n *nodestore) fetchNodeHostname(node Node) *Node {
	for x := range n.nodes {
		if n.nodes[x].Hostname.Hostname == node.Hostname.Hostname {
			return &n.nodes[x]
		}
	}
	return nil
}

func (n *nodestore) fetchNodeIPv4(node Node) *Node {
	for x := range n.nodes {
		if n.nodes[x].IPv4.IPv4 == node.IPv4.IPv4 {
			return &n.nodes[x]
		}
	}
	return nil
}

func (n *nodestore) fetchNode(node Node) *Node {
	switch node.NodeType.NodeType {
	case NODETYPE_IPV4:
		return n.fetchNodeIPv4(node)
	case NODETYPE_HOSTNAME:
		return n.fetchNodeHostname(node)
	default:
		panic("fetchNode: unknown node type")
	}
	return nil
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

	Related []NodeRelation `json:"related,omitempty"`
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

func nodesFromHintHostname(hint *AssetHint) (nodes []Node) {
	if len(hint.Details.Hostname) == 0 {
		return nodes
	}
	var nn Node
	nn.NodeType.NodeType = NODETYPE_HOSTNAME
	nn.NodeType.NodeTypeStr = "hostname"
	nn.Hostname.Hostname = hint.Details.Hostname
	nn.LastUpdated = hint.Timestamp
	nn.Confidence = 100
	nodes = append(nodes, nn)
	return nodes
}

func nodesFromHint(hint *AssetHint) (nodes []Node) {
	nodes = append(nodes, nodesFromHintIPv4(hint)...)
	nodes = append(nodes, nodesFromHintHostname(hint)...)
	for x := range nodes {
		nodes[x].NodeID = uuid.NewRandom().String()
	}
	return nodes
}

func relateNodeGroup(innodes []Node) []Node {
	for i := range innodes {
		for j := range innodes {
			if j == i {
				continue
			}
			var nr NodeRelation
			nr.NodeType = innodes[j].NodeType
			nr.LastUpdated = innodes[j].LastUpdated
			nr.Confidence = 100
			nr.NodeLinkID = innodes[j].NodeID
			innodes[i].Related = append(innodes[i].Related, nr)
		}
	}
	return innodes
}
