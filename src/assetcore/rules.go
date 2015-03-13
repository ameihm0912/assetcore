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

func relatedAssets(hint assetHint) []*asset {
	return aBlock.searchRelatedAssets(&hint)
}

func createNewAssetFromHint(hint assetHint) {
	var a asset
	a.AssetID = uuid.NewRandom().String()
	if len(hint.Details.Hostname) > 0 {
		a.Hostname = append(a.Hostname, hint.Details.Hostname)
	}
	if len(hint.Details.IPv4) > 0 {
		for _, x := range hint.Details.IPv4 {
			a.IPv4 = append(a.IPv4, x)
		}
	}
	a.updateHintTags(&hint)
	a.LastUpdated = time.Now().UTC()
	a.IsNew = true
	aBlock.addAsset(a)
}

func rulesPipeline(hint assetHint) {
	related := relatedAssets(hint)

	if len(related) == 0 {
		createNewAssetFromHint(hint)
	}
}
