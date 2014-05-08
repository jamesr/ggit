package main

import (
	"compress/zlib"
	"io"
)

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
