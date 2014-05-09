// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"crypto/sha1"
	"encoding/binary"
	"reflect"
	"testing"
	"time"
)

type entryTestCase struct {
	data     []byte
	e        entry
	length   uint32
	hasError bool
}

func TestParseEntry(t *testing.T) {
	tooShort := make([]byte, 69)

	good := entryTestCase{
		data:     make([]byte, 80),
		e:        entry{},
		length:   80,
		hasError: false}
	good.e.ctime = time.Unix(1, 2)
	copy(good.data[0:8], []byte{0, 0, 0, 1, 0, 0, 0, 2})
	good.e.mtime = time.Unix(3, 4)
	copy(good.data[8:16], []byte{0, 0, 0, 3, 0, 0, 0, 4})
	good.e.dev = 5
	copy(good.data[16:20], []byte{0, 0, 0, 5})
	good.e.ino = 6
	copy(good.data[20:24], []byte{0, 0, 0, 6})
	// high 4 bits 1000 (regular file), low 9 bits 0644
	good.e.mode = 0x8000<<16 | 0x01a4
	binary.BigEndian.PutUint32(good.data[24:28], good.e.mode)
	good.e.uid = 7
	copy(good.data[28:32], []byte{0, 0, 0, 7})
	good.e.gid = 8
	copy(good.data[32:36], []byte{0, 0, 0, 8})
	good.e.size = 9
	copy(good.data[36:40], []byte{0, 0, 0, 9})
	emptyHash := sha1.Sum(make([]byte, 0))
	copy(good.e.hash[:], emptyHash[:])
	copy(good.data[40:60], emptyHash[:])
	// high bit 1 for assume-valid, rest 0
	good.e.flags = 0x80
	copy(good.data[60:62], []byte{0, 0x80})
	good.e.path = []byte("foo/bar/file.txt")
	copy(good.data[62:78], []byte(good.e.path))
	// Rest of the slice is NUL, so no need to set terminator or padding.

	// Same as good, but the length divides 8 evenly without any padding.
	goodLengthDivides8 := entryTestCase{
		data:     make([]byte, 80),
		e:        good.e,
		length:   good.length,
		hasError: false}
	copy(goodLengthDivides8.data, good.data)
	goodLengthDivides8.e.path = []byte(string(good.e.path) + "2")
	goodLengthDivides8.data[78] = '2'
	goodLengthDivides8.length = 80

	cases := []entryTestCase{
		{tooShort, entry{}, 0, true},
		good, goodLengthDivides8}

	for i := range cases {
		e, length, err := parseEntry(cases[i].data)
		gotError := err != nil
		if cases[i].hasError != gotError {
			t.Errorf("Expected error %v but instead got %v on case %v",
				cases[i].hasError, err, i)
		}
		if cases[i].length != length {
			t.Errorf("Expected length %v but got %v on case %v", cases[i].length,
				length, i)
		}
		if !reflect.DeepEqual(cases[i].e, e) {
			t.Errorf("Expected entry %v but got %v on case %v", cases[i].e, e,
				i)
		}
	}
}

type indexHeaderTestCase struct {
	data             []byte
	version, entries uint32
	hasError         bool
}

func fillSignature(data []byte) {
	data[0] = 'D'
	data[1] = 'I'
	data[2] = 'R'
	data[3] = 'C'
}

func fillChecksum(data []byte) {
	idx := len(data) - sha1.Size
	checksum := sha1.Sum(data[0:idx])
	for i := 0; i < sha1.Size; i++ {
		data[idx+i] = checksum[i]
	}
}

func TestValidIndexHeader(t *testing.T) {
	var zeros [12 + sha1.Size]byte

	var badChecksum [12 + sha1.Size]byte
	fillSignature(badChecksum[:])

	var version2 [12 + sha1.Size]byte
	fillSignature(version2[:])
	binary.BigEndian.PutUint32(version2[4:8], 2)
	fillChecksum(version2[:])

	var entries4 [12 + sha1.Size]byte
	fillSignature(entries4[:])
	binary.BigEndian.PutUint32(entries4[8:12], 4)
	fillChecksum(entries4[:])

	cases := []indexHeaderTestCase{{zeros[:], 0, 0, true},
		{badChecksum[:], 0, 0, true},
		{version2[:], 2, 0, false},
		{entries4[:], 0, 4, false}}

	for i := range cases {
		version, entries, err := parseIndexFileHeader(cases[i].data)
		gotError := err != nil
		if cases[i].hasError != gotError {
			t.Errorf("Expected error %v did not match on case %v", cases[i].hasError, i)
		}
		if version != cases[i].version {
			t.Errorf("Expected version %v but got %v on case %v", cases[i].version,
				version, i)
		}
		if entries != cases[i].entries {
			t.Errorf("Expected entries %v but got %v on case %v", cases[i].entries,
				entries, i)
		}
	}
}
