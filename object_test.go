// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"testing"
)

type testCase struct {
	data        []byte
	objectType  string
	size        uint32
	prettyPrint string
}

func TestObjectParsing(t *testing.T) {
	dataABlob := []byte{0x78, 0x01, 0x4b, 0xca, 0xc9, 0x4f, 0x52, 0x30, 0x62, 0x48, 0xe4, 0x02, 0x00, 0x0e, 0x64, 0x02, 0x5d}

	dataTestContentBlob := []byte{0x78, 0x01, 0x4b, 0xca, 0xc9, 0x4f, 0x52, 0x30, 0x34, 0x66, 0x28, 0x49, 0x2d, 0x2e, 0x51, 0x48, 0xce, 0xcf, 0x2b, 0x49, 0xcd, 0x2b, 0xe1, 0x02, 0x00, 0x4b, 0xdf, 0x07, 0x09}

	cases := []testCase{
		{data: dataABlob,
			objectType:  "blob",
			size:        2,
			prettyPrint: "a"},
		{data: dataTestContentBlob,
			objectType:  "blob",
			size:        13,
			prettyPrint: "test content"}}

	for i := range cases {
		c := cases[i]
		buf := bytes.NewBuffer(cases[i].data)
		r, err := zlib.NewReader(buf)
		br := bufio.NewReader(r)
		objectType, err := objectType(br)
		if err != nil {
			t.Errorf("error parsing type %v", err)
		}
		if objectType != c.objectType {
			t.Errorf("Expected type %v but got %v on case %v\n",
				c.objectType, objectType, i)
		}
		size, err := objectSize(br)
		if err != nil {
			t.Errorf("error parsing size %v", err)
		}
		if size != c.size {
			t.Errorf("Expected size %v but got %v on case %v\n",
				c.size, size, i)
		}
	}
}
