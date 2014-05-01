package main

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strconv"
)

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

func nameToPath(object string) string {
	return ".git/objects/" + object[:2] + "/" + object[2:]
}

func openObjectFile(name string) (*os.File, error) {
	path := nameToPath(name)
	return os.Open(path)
}

type object struct {
	objectType string
	size       uint32
	file       *os.File
	reader     io.ReadCloser
}

func parseObject(f *os.File, r io.ReadCloser) (object, error) {
	br := bufio.NewReader(r)
	t, err := objectType(br)
	if err != nil {
		return object{}, err
	}
	s, err := objectSize(br)
	return object{objectType: t, size: s, file: f, reader: r}, nil
}

func parseObjectFile(path string) (object, error) {
	file, err := openObjectFile(path)
	if err != nil {
		return object{}, err
	}
	r, err := zlib.NewReader(file)
	if err != nil {
		file.Close()
		return object{}, err
	}
	return parseObject(file, r)
}

func dumpObjectType(name string) error {
	o, err := parseObjectFile(name)
	if err != nil {
		return err
	}
	fmt.Println(o.objectType)
	o.reader.Close()
	o.file.Close()
	return nil
}

func dumpObjectSize(name string) error {
	o, err := parseObjectFile(name)
	if err != nil {
		return err
	}
	fmt.Println(o.size)
	return nil
}

func dumpPrettyPrint(name string) error {
	o, err := parseObjectFile(name)
	if err != nil {
		return err
	}
	if o.objectType == "tree" {
		return fmt.Errorf("tree not supported, should do git ls-tree")
	}
	fmt.Print(o.reader)
	o.reader.Close()
	o.file.Close()
	return nil
}
