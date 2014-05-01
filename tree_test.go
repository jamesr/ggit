package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestLsTree(t *testing.T) {
	treeBytes := []byte("tree 131\x00" +
		"100644 abc.txt\x00\x8b\xae\xf1\xb4\xab\xc4x\x17\x8b\x00Mb\x03\x1c\xf7\xfem\xb6\xf9\x03" +
		"40000 dir\x00@\xc5\xdbc\xe2\x83?!\t/\xfb\x06\xa2b\t\xdfSN\x91\xc9" +
		"100755 exe\x00\xe6\x9d\xe2\x9b\xb2\xd1\xd6CK\x8b)\xaewZ\xd8\xc2\xe4\x8cS\x91" +
		"120000 symlink\x00\xf2\xba\x8f\x84\xab\\\x1b\xce\x84\xa7\xb4A\xcb\x19Y\xcf\xc7\t;\x7f")

	prettyTree := "" +
		"100644 blob 8baef1b4abc478178b004d62031cf7fe6db6f903	abc.txt\n" +
		"040000 tree 40c5db63e2833f21092ffb06a26209df534e91c9	dir\n" +
		"100755 blob e69de29bb2d1d6434b8b29ae775ad8c2e48c5391	exe\n" +
		"120000 blob f2ba8f84ab5c1bce84a7b441cb1959cfc7093b7f	symlink"

	b := bytes.NewBuffer(treeBytes)

	origParseObjectFile := parseObjectFile
	parseObjectFile = func(name string) (object, error) {
		objectToType := map[string]string{
			"8baef1b4abc478178b004d62031cf7fe6db6f903": "blob",
			"40c5db63e2833f21092ffb06a26209df534e91c9": "tree",
			"e69de29bb2d1d6434b8b29ae775ad8c2e48c5391": "blob",
			"f2ba8f84ab5c1bce84a7b441cb1959cfc7093b7f": "blob",
		}
		t := objectToType[name]
		if t == "" {
			return object{}, fmt.Errorf("no such object %s", name)
		}
		return object{
			objectType: t,
			size:       1,
			file:       nil,
			reader:     nil,
		}, nil
	}
	defer func() { parseObjectFile = origParseObjectFile }()

	tree, err := parseObject(nopCloser{b})
	if err != nil {
		t.Errorf("error parsing object: %v\n", err)
	}

	actual, err := prettyPrintTree(tree)
	if err != nil {
		t.Error("error prettying tree: %v\n", err)
	}

	if actual != prettyTree {
		t.Errorf("expected \"%v\" got \"%v\"", prettyTree, actual)
	}
}
