package main

import (
	"bytes"
	"io"
	"testing"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestLsTree(t *testing.T) {
	treeBytes := []byte("tree 114\x00100644 README.md\x00\xab\xa2b\xd3\x81N\x19|\xde\xfa\xe7\xaf\xfcC\xabg\xa5\xf5!\x97" +
		"100644 index.go\x00\x82]*^9\xf5\x1d:\xdf\x14\xff|\xa4\x9b+E\x97\x96T\xcf" +
		"100644 index_test.go\x002\xc9\x18\xe9{9\x9e\x07\x1d9\xd3\x0e\x0c\xee\xfe\x08|\xadY\x0c")

	prettyTree := "100644 blob aba262d3814e197cdefae7affc43ab67a5f52197	README.md\n" +
		"100644 blob 825d2a5e39f51d3adf14ff7ca49b2b45979654cf	index.go\n" +
		"100644 blob 32c918e97b399e071d39d30e0ceefe087cad590c	index_test.go"

	b := bytes.NewBuffer(treeBytes)
	tree, err := parseObject(nopCloser{b})
	if err != nil {
		t.Error(err)
	}

	actual, err := prettyPrintTree(tree)
	if err != nil {
		t.Error(err)
	}

	if actual != prettyTree {
		t.Errorf("expected \"%v\" got \"%v\"", prettyTree, actual)
	}
}
