package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"strconv"
	"syscall"
)

func objectToPath(object string) string {
	return ".git/objects/" + object[:2] + "/" + object[2:]
}

func objectType(buf *bufio.Reader) (string, error) {
	types := []string{"", // OBJ_NONE
		"commit", // OBJ_COMMIT
		"tree",   // OBJ_TREE
		"blob",   // OBJ_BLOB
		"tag",    // OBJ_TAG
	}
	t, err := buf.ReadString(' ')
	if err != nil {
		return "", err
	}
	t = t[:len(t)-1]
	for _, u := range types {
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

func mapObjectFile(object string) ([]byte, error) {
	path := objectToPath(object)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileinfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	length := int(fileinfo.Size())
	if length < 4 {
		return nil, fmt.Errorf("File too short to be valid object: %d", length)
	}
	flags := 0
	data, err := syscall.Mmap(int(file.Fd()), 0, length, syscall.PROT_READ, flags)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseHeader(br *bufio.Reader) (t string, s uint32, err error) {
	t, err = objectType(br)
	if err != nil {
		return
	}
	s, err = objectSize(br)
	return
}

func parseTypeAndSize(object string) (string, uint32, error) {
	data, err := mapObjectFile(object)
	if err != nil {
		return "", 0, err
	}
	defer syscall.Munmap(data)
	buf := bytes.NewBuffer(data)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return "", 0, err
	}
	defer r.Close()
	br := bufio.NewReader(r)
	return parseHeader(br)
}

func dumpObjectType(object string) error {
	t, _, err := parseTypeAndSize(object)
	if err != nil {
		return err
	}
	fmt.Println(t)
	return nil
}

func dumpObjectSize(object string) error {
	_, s, err := parseTypeAndSize(object)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}
