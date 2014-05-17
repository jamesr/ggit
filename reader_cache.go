// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"os"
)

type readCloserReset interface {
	io.ReadCloser
	Reset(io.Reader) error
}

// zlib.ReadCloserReset is a proposed addition to the go std library that might
// appear (in almost certainly a cleaner form) in go 1.4, but doesn't exist today.
// To build this code against 1.3 change this to an io.ReadCloser and change the
// zr.Reset(r) line to allocate a new zlib reader.
var zlibReaders = make(chan readCloserReset, 128)

func getZlibReader(r io.Reader) (io.ReadCloser, error) {
	select {
	case zr := <-zlibReaders:
		err := zr.Reset(r)
		return zr, err
	default:
		return zlib.NewReader(r)
	}
}

func returnZlibReader(r io.ReadCloser) {
	err := r.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing reader %v\n", err)
		return
	}
	if rcr, ok := r.(readCloserReset); ok {
		select {
		case zlibReaders <- rcr:
		default:
			// adding to cache would block
		}
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
