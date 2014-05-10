// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"strconv"
)

var objectTypeStrings = []string{"", // OBJ_NONE
	"commit", // OBJ_COMMIT
	"tree",   // OBJ_TREE
	"blob",   // OBJ_BLOB
	"tag",    // OBJ_TAG
}

func objectType(buf *bufio.Reader) (string, error) {
	t, err := buf.ReadString(' ')
	if err != nil {
		return "", err
	}
	t = t[:len(t)-1]
	for _, u := range objectTypeStrings {
		if t == u {
			return t, nil
		}
	}
	return "", fmt.Errorf("Invalid object type %s", t)
}

func objectSize(buf *bufio.Reader) (uint32, error) {
	s, err := buf.ReadString(0)
	if err != nil {
		return 0, nil
	}
	size, err := strconv.ParseUint(s[:len(s)-1], 10, 32)
	return uint32(size), err
}

type Object struct {
	ObjectType string
	Size       uint32
	file       *os.File
	zlibReader zlib.ReadCloserReset
	Reader     io.Reader
}

func (o Object) Close() {
	if o.file != nil {
		o.file.Close()
	}
}

func parseObject(r io.ReadCloser, zr zlib.ReadCloserReset) (*Object, error) {
	br := bufio.NewReader(r)
	t, err := objectType(br)
	if err != nil {
		returnZlibReader(zr)
		return nil, err
	}
	s, err := objectSize(br)
	return &Object{ObjectType: t, Size: s, zlibReader: zr, Reader: br}, nil
}

func NameToPath(object string) string {
	return ".git/objects/" + object[:2] + "/" + object[2:]
}

func openObjectFile(name string) (*os.File, error) {
	path := NameToPath(name)
	return os.Open(path)
}

func nameToHash(name string) []byte {
	h := make([]byte, sha1.Size)
	for i := 0; i < sha1.Size; i++ {
		n, _ := strconv.ParseUint(name[2*i:2*(i+1)], 16, 8)
		h[i] = byte(n)
	}
	return h
}

var ParseObjectFile = func(name string) (Object, error) {
	o, err := findHash(nameToHash(name))
	if err != nil {
		return Object{}, err
	}
	if o != nil {
		return *o, nil
	}
	file, err := openObjectFile(name)
	if err != nil {
		return Object{}, err
	}
	r, err := getZlibReader(file)
	if err != nil {
		file.Close()
		return Object{}, err
	}
	o, err = parseObject(r, r)
	if err != nil {
		file.Close()
		return Object{}, err
	}
	o.file = file
	return *o, nil
}
