package main

import (
	"bytes"
	"testing"
)

func TestShow(t *testing.T) {
	commitBytes := []byte("commit 247\x00tree 7e80d6c030ed0f3870dc2104f5b906b3fb2f9de2\nparent 6d4683dfec45407edb4e8124ce3c32c7ee570969\nauthor James Robinson <jamesr@chromium.org> 1398979283 -0700\ncommitter James Robinson <jamesr@chromium.org> 1398979283 -0700\n\npretty print index entries\n")

	show := "" +
		`commit 919b32c0b3cdb2b80ed7daa741b1fe88176b4264
Author: James Robinson <jamesr@chromium.org>
Date:   Thu May 1 14:21:23 2014 -0700

    pretty print index entries
    
`

	b := bytes.NewBuffer(commitBytes)

	/*
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
	*/

	commit, err := parseObject(nopCloser{b})
	if err != nil {
		t.Errorf("error parsing object: %v\n", err)
	}

	actual, err := showCommit(commit)
	if err != nil {
		t.Error("error prettying commit: %v\n", err)
	}

	if actual != show {
		t.Errorf("expected \"%v\" got \"%v\"", show, actual)
	}
}
