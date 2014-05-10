// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bufio"
	"compress/zlib"
	"io"
)

// zlib.ReadCloserReset is a proposed addition to the go std library that might
// appear (in almost certainly a cleaner form) in go 1.4, but doesn't exist today.
// To build this code against 1.3 change this to an io.ReadCloser and change the
// zr.Reset(r) line to allocate a new zlib reader.
var zlibReaders = make(chan zlib.ReadCloserReset, 128)

func getZlibReader(r io.Reader) (zlib.ReadCloserReset, error) {
	select {
	case zr := <-zlibReaders:
		err := zr.Reset(r)
		return zr, err
	default:
		return zlib.NewReader(r)
	}
}

func returnZlibReader(zr zlib.ReadCloserReset) {
	zr.Close()
	select {
	case zlibReaders <- zr:
	default:
		// adding to cache would block
	}
}

var bufioReaders = make(chan *bufio.Reader, 128)

func getBufioReader(r io.Reader) *bufio.Reader {
	select {
	case br := <-bufioReaders:
		br.Reset(r)
		return br
	default:
		return bufio.NewReaderSize(r, 512)
	}
}

func returnBufioReader(br *bufio.Reader) {
	select {
	case bufioReaders <- br:
	default:
	}
}
