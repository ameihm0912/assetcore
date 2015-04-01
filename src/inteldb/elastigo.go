// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package inteldb

import (
	"encoding/json"
	"errors"
	"fmt"
	elastigo "github.com/mattbaird/elastigo/lib"
)

type idbES struct {
	esConn  *elastigo.Conn
	esHost  string
	esIndex string
}

func (i *idbES) init(host string) {
	i.esConn = elastigo.NewConn()
	i.esConn.Domain = host
}

func (i *idbES) search(template string, doctype string) (ret []json.RawMessage, err error) {
	res, err := i.esConn.Search(i.esIndex, doctype, nil, template)
	if err != nil {
		fmt.Println("%v", err)
		return ret, err
	}
	havehits := res.Hits.Len()
	if havehits < res.Hits.Total {
		return ret, errors.New("partial result set returned on search")
	}
	// If we make it here we have all the data we needed from the query, package up
	// the raw JSON results and return
	for _, x := range res.Hits.Hits {
		newhit := *x.Source
		ret = append(ret, newhit)
	}
	return ret, nil
}

func (i *idbES) setIndex(index string, tryCreate bool) error {
	i.esIndex = index
	haveidx, err := i.esConn.IndicesExists(index)
	if err != nil {
		return err
	}
	if haveidx {
		return nil
	}
	if !tryCreate {
		return errors.New("setIndex(): index does not exist and will not create")
	}
	_, err = i.esConn.CreateIndex(index)
	if err != nil {
		return err
	}
	return nil
}
