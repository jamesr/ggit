package main

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

type object struct {
	objectType string
	size       uint32
	file       *os.File
	reader     *bufio.Reader
}

func (o object) Close() {
	if o.file != nil {
		o.file.Close()
	}
}

func parseObject(r io.Reader) (*object, error) {
	br := bufio.NewReader(r)
	t, err := objectType(br)
	if err != nil {
		return nil, err
	}
	s, err := objectSize(br)
	return &object{objectType: t, size: s, file: nil, reader: br}, nil
}

func nameToPath(object string) string {
	return ".git/objects/" + object[:2] + "/" + object[2:]
}

func openObjectFile(name string) (*os.File, error) {
	path := nameToPath(name)
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

var parseObjectFile = func(name string) (object, error) {
	o, err := findHash(nameToHash(name))
	if err != nil {
		return object{}, err
	}
	if o != nil {
		return *o, nil
	}
	file, err := openObjectFile(name)
	if err != nil {
		return object{}, err
	}
	r, err := zlib.NewReader(file)
	if err != nil {
		file.Close()
		return object{}, err
	}
	o, err = parseObject(r)
	if err != nil {
		file.Close()
		return object{}, err
	}
	o.file = file
	return *o, nil
}
