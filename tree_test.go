// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestLsTree(t *testing.T) {
	treeBytes := []byte("tree 131\x00" +
		"100644 abc.txt\x00\x8b\xae\xf1\xb4\xab\xc4x\x17\x8b\x00Mb\x03\x1c\xf7\xfem\xb6\xf9\x03" +
		"40000 dir\x00@\xc5\xdbc\xe2\x83?!\t/\xfb\x06\xa2b\t\xdfSN\x91\xc9" +
		"100755 exe\x00\xe6\x9d\xe2\x9b\xb2\xd1\xd6CK\x8b)\xaewZ\xd8\xc2\xe4\x8cS\x91" +
		"120000 symlink\x00\xf2\xba\x8f\x84\xab\\\x1b\xce\x84\xa7\xb4A\xcb\x19Y\xcf\xc7\t;\x7f")

	prettyTree := "" +
		"100644 blob 8baef1b4abc478178b004d62031cf7fe6db6f903	abc.txt\n" +
		"040000 tree 40c5db63e2833f21092ffb06a26209df534e91c9	dir\n" +
		"100755 blob e69de29bb2d1d6434b8b29ae775ad8c2e48c5391	exe\n" +
		"120000 blob f2ba8f84ab5c1bce84a7b441cb1959cfc7093b7f	symlink"

	b := bytes.NewBuffer(treeBytes)

	origLookupObject := LookupObject
	LookupObject = func(name string) (Object, error) {
		objectToType := map[string]string{
			"8baef1b4abc478178b004d62031cf7fe6db6f903": "blob",
			"40c5db63e2833f21092ffb06a26209df534e91c9": "tree",
			"e69de29bb2d1d6434b8b29ae775ad8c2e48c5391": "blob",
			"f2ba8f84ab5c1bce84a7b441cb1959cfc7093b7f": "blob",
		}
		t := objectToType[name]
		if t == "" {
			return Object{}, fmt.Errorf("no such object %s", name)
		}
		return Object{
			ObjectType: t,
			Size:       1,
			file:       nil,
			Reader:     nil,
		}, nil
	}
	defer func() { LookupObject = origLookupObject }()

	tree, err := parseObject(ioutil.NopCloser(b), nil)
	if err != nil {
		t.Errorf("error parsing object: %v\n", err)
	}

	actual, err := PrettyPrintTree(*tree, false, false, "")
	if err != nil {
		t.Errorf("error prettying tree: %s\n", err)
	}

	if actual != prettyTree {
		t.Errorf("expected \"%v\" got \"%v\"", prettyTree, actual)
	}
}
