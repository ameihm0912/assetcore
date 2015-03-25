// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package main

import (
	"encoding/json"
	"io"
	"os"
)

type cacheState struct {
	assetsPost *os.File
	assetsPre  *os.File
	hints      *os.File
}

var cs cacheState

func cache_asset(a asset, ispre bool) error {
	var f *os.File
	if ispre {
		f = cs.assetsPre
	} else {
		f = cs.assetsPost
	}

	enc := json.NewEncoder(f)
	err := enc.Encode(a)
	if err != nil {
		return err
	}

	return nil
}

func cache_hint(h assetHint) error {
	enc := json.NewEncoder(cs.hints)
	err := enc.Encode(h)
	if err != nil {
		return err
	}
	return nil
}

func cache_get_assets_pre() (ret []asset, err error) {
	dec := json.NewDecoder(cs.assetsPre)
	for {
		var a asset
		err := dec.Decode(&a)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return ret, err
			}
		}
		ret = append(ret, a)
	}
	return ret, nil
}

func cache_get_hints() (ret []assetHint, err error) {
	dec := json.NewDecoder(cs.hints)
	for {
		var a assetHint
		err := dec.Decode(&a)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return ret, err
			}
		}
		ret = append(ret, a)
	}
	return ret, nil
}

func cache_init() error {
	var err error
	fmode := os.O_RDWR | os.O_CREATE
	cs.assetsPre, err = os.OpenFile(cfg.cacheAssetsPrePath, fmode, 0644)
	if err != nil {
		return err
	}
	cs.assetsPost, err = os.OpenFile(cfg.cacheAssetsPostPath, fmode|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	cs.hints, err = os.OpenFile(cfg.cacheHintsPath, fmode, 0644)
	if err != nil {
		return err
	}
	return nil
}

func cache_close() error {
	if cs.assetsPost != nil {
		err := cs.assetsPost.Close()
		if err != nil {
			return err
		}
	}
	if cs.assetsPre != nil {
		err := cs.assetsPre.Close()
		if err != nil {
			return err
		}
	}
	if cs.hints != nil {
		err := cs.hints.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
