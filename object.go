package main

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strconv"
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

func openObjectFile(object string) (*os.File, error) {
	path := objectToPath(object)
	return os.Open(path)
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
	file, err := openObjectFile(object)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()
	r, err := zlib.NewReader(file)
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

func dumpPrettyPrint(object string) error {
	file, err := openObjectFile(object)
	if err != nil {
		return err
	}
	defer file.Close()
	r, err := zlib.NewReader(file)
	if err != nil {
		return err
	}
	defer r.Close()
	br := bufio.NewReader(r)
	t, _, err := parseHeader(br)
	if err != nil {
		return err
	}
	if t == "tree" {
		return fmt.Errorf("tree not supported, should do git ls-tree")
	}
	for {
		b := make([]byte, 4096)
		n, err := br.Read(b)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Print(string(b[:n]))
	}
	return nil
}
