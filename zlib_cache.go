package main

import (
	"compress/zlib"
	"io"
)

// TODO: locking, or use a channel?
var readers []zlib.ReadCloserReset

func getZlibReader(r io.Reader) (zlib.ReadCloserReset, error) {
	if len(readers) > 0 {
		zr := readers[len(readers)-1]
		err := zr.Reset(r)
		readers = readers[:len(readers)-1]
		return zr, err
	}
	return zlib.NewReader(r)
}

func returnZlibReader(zr zlib.ReadCloserReset) {
	zr.Close()
	readers = append(readers, zr)
	if zr == nil {
		panic("oh shit")
	}
}
