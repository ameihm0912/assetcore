// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
//"time"
)

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
		logMessage("%+v", nh)
	}
}
