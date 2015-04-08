// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

func corLoop() {
	for {
		select {
		case <-cfg.exitCorChan:
			logMessage("corLoop: exit notification")
			return
		default:
		}

		var nh AssetHint
		select {
		case nh = <-cfg.hintsChan:
		case <-cfg.exitCorChan:
			logMessage("corLoop: exit notification")
			return
		}

		nh.sanitize()

		// Get a list of all nodes we could create from the asset hint.
		// If the nodes do not already exist in the nodestore they
		// will be added with the required relations. If they already
		// exist updates will occur.
		newnodes := nodesFromHint(&nh)
		newnodes = relateNodeGroup(newnodes)

		for _, x := range newnodes {
			logMessage("%+v", x)
			_ = acns.fetchNode(x)
		}

		for x := range newnodes {
			acns.updateNode(newnodes[x])
		}
	}
}
