package main

import (
	"bufio"
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
