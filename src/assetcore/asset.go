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
	updatecount  int

	// Search lock.
	sync.Mutex
}

// Some additional counters used during correlation operations that are
// not directly related to the asset block.
type correlationCounters struct {
	hints_ignored_tagts int
	sync.Mutex
}

var corCntrs correlationCounters

func (a *assetBlock) addAsset(newasset asset) {
	a.Lock()
	a.assets = append(a.assets, newasset)
	a.count += 1
	if newasset.IsNew {
		a.newcount += 1
	} else {
		a.existedcount += 1
	}
	a.Unlock()
}

func (a *assetBlock) searchRelatedAssets(hint *assetHint) []*asset {
	ret := make([]*asset, 0)

	a.Lock()
	// Iterate through assets and test relation to the hint. We don't use
	// range here as we want to return pointers to the asset values
	// for later modification.
	for i := 0; i < len(a.assets); i++ {
		x := &a.assets[i]
		x.Lock()
		added := false
		ret, added = x.testIPv4Related(hint, ret)
		if added {
			x.Unlock()
			continue
		}
		ret, added = x.testHostnameRelated(hint, ret)
		if added {
			x.Unlock()
			continue
		}
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

func (a *asset) tagExpired(hint *assetHint) bool {
	ts := hint.Timestamp
	for _, x := range a.Fold.Tags {
		// Don't compare the asset tag here, this is present in all
		// hint events, and should always be set to the timestamp
		// provided be the newest hint for the asset.
		if x.Name == "asset" {
			continue
		}
		for _, y := range hint.Tags {
			if x.Name == y {
				if x.Provided.After(ts) {
					return true
				} else {
					// The tag existed in both the hint and
					// the asset, but the hint is newer so
					// we will want to integrate it.
					return false
				}
			}
		}
	}
	return false
}

func (a *asset) updateFromHint(hint *assetHint) {
	// See if the asset already has data integrated from a hint from
	// the same provider that is newer; if so this older hint is just
	// ignored.
	aBlock.Lock()
	a.Lock()
	corCntrs.Lock()
	defer func() {
		corCntrs.Unlock()
		a.Unlock()
		aBlock.Unlock()
	}()

	if a.tagExpired(hint) {
		corCntrs.hints_ignored_tagts++
		return
	}

	aBlock.updatecount += 1
}

func (a *asset) testIPv4Related(hint *assetHint, l []*asset) ([]*asset, bool) {
	for _, x := range a.IPv4 {
		for _, y := range hint.Details.IPv4 {
			if x == y {
				l = append(l, a)
				return l, true
			}
		}
	}
	return l, false
}

func (a *asset) testHostnameRelated(hint *assetHint, l []*asset) ([]*asset, bool) {
	for _, x := range a.Hostname {
		if hint.Details.Hostname == x {
			l = append(l, a)
			return l, true
		}
	}
	return l, false
}
