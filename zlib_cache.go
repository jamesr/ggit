package main

import (
	"compress/zlib"
	"io"
)

var readers = make(chan zlib.ReadCloserReset, 128)

func getZlibReader(r io.Reader) (zlib.ReadCloserReset, error) {
	select {
	case zr := <-readers:
		err := zr.Reset(r)
		return zr, err
	default:
		return zlib.NewReader(r)
	}
}

func returnZlibReader(zr zlib.ReadCloserReset) {
	zr.Close()
	select {
	case readers <- zr:
	default:
		// adding to cache would block
	}
}
