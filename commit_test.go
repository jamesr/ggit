package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestShow(t *testing.T) {
	commitHash := "919b32c0b3cdb2b80ed7daa741b1fe88176b4264"
	commitBytes := []byte("commit 247\x00tree 7e80d6c030ed0f3870dc2104f5b906b3fb2f9de2\nparent 6d4683dfec45407edb4e8124ce3c32c7ee570969\nauthor James Robinson <jamesr@chromium.org> 1398979283 -0700\ncommitter James Robinson <jamesr@chromium.org> 1398979283 -0700\n\npretty print index entries\n")

	show := "" +
		`commit 919b32c0b3cdb2b80ed7daa741b1fe88176b4264
Author: James Robinson <jamesr@chromium.org>
Date:   Thu May 1 14:21:23 2014 -0700

    pretty print index entries
    
`

	origParseObjectFile := parseObjectFile
	parseObjectFile = func(name string) (object, error) {
		if name != commitHash {
			return object{}, fmt.Errorf("unknown name %s", name)
		}
		b := bytes.NewBuffer(commitBytes)
		return parseObject(b)
	}
	defer func() { parseObjectFile = origParseObjectFile }()

	actual, err := showCommit(commitHash)
	if err != nil {
		t.Error("error prettying commit: %v\n", err)
	}

	if actual != show {
		t.Errorf("expected \"%v\" got \"%v\"", show, actual)
	}
}
