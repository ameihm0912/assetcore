// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package main

import (
	"sync"
	"time"
)

type assetBlock struct {
	assets       []asset
	count        int
	newcount     int
	existedcount int

	// Search lock.
	sync.Mutex
}

func (a *assetBlock) addAsset(newasset asset) {
	a.Lock()
	a.assets = append(a.assets, newasset)
	a.count += 1
	if newasset.IsNew {
		a.newcount += 1
	}
	a.Unlock()
}

func (a *assetBlock) searchRelatedAssets(hint *assetHint) (ret []*asset) {
	a.Lock()
	for _, x := range a.assets {
		x.Lock()
		ret = x.testIPv4Related(hint, ret)
		ret = x.testHostnameRelated(hint, ret)
		x.Unlock()
	}
	a.Unlock()
	return ret
}

type asset struct {
	// Asset UUID
	AssetID string `json:"assetid"`

	// Fold source information
	Fold foldSource `json:"fold,omitempty"`

	// Hostnames known to be assigned to this device.
	Hostname []string `json:"hostname,omitempty"`

	// Addresses known to be assigned to this device.
	IPv4 []string `json:"ipv4,omitempty"`
	IPv6 []string `json:"ipv6,omitempty"`

	// MAC addresses known to be associated with this device.
	MAC []string `json:"macaddress,omitempty"`

	// The last time this object was updated.
	LastUpdated time.Time `json:"lastupdated"`

	IsNew bool `json:"-"`
	sync.Mutex
}

type foldSource struct {
	Tags []foldTag `json:"tags,omitempty"`
}

type foldTag struct {
	Name     string    `json:"name"`
	Provided time.Time `json:"provided"`
}

func (a *asset) updateHintTags(hint *assetHint) {
	var foundidx int
	for _, x := range hint.Tags {
		foundidx = -1
		for i, y := range a.Fold.Tags {
			if x == y.Name {
				foundidx = i
				break
			}
		}
		if foundidx != -1 {
			a.Fold.Tags[foundidx].Provided = hint.Timestamp
		} else {
			newtag := foldTag{x, hint.Timestamp}
			a.Fold.Tags = append(a.Fold.Tags, newtag)
		}
	}
}

func (a *asset) testIPv4Related(hint *assetHint, l []*asset) []*asset {
	for _, x := range a.IPv4 {
		for _, y := range hint.Details.IPv4 {
			if x == y {
				l = append(l, a)
				return l
			}
		}
	}
	return l
}

func (a *asset) testHostnameRelated(hint *assetHint, l []*asset) []*asset {
	for _, x := range a.Hostname {
		if hint.Details.Hostname == x {
			l = append(l, a)
			return l
		}
	}
	return l
}
