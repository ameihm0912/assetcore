// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package inteldb

// The maximum number of documents the library will fetch from ES for any
// given request, can be set using SetMaxDocuments().
var maxDocuments int = 10000

func SetMaxDocuments(v int) {
	maxDocuments = v
}

type InteldbConn struct {
	idbConn idbES
}

func (i *InteldbConn) Init(host string, index string) (err error) {
	i.idbConn.init(host)
	err = i.idbConn.setIndex(index, true)
	if err != nil {
		return err
	}

	return nil
}
