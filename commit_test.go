package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestShowCommit(t *testing.T) {
	commitHash := []string{
		"919b32c0b3cdb2b80ed7daa741b1fe88176b4264",
		"9072f9473cd87dcc76b213853cce7acd380b689f"}
	commitBytes := [][]byte{
		[]byte("commit 247\x00tree 7e80d6c030ed0f3870dc2104f5b906b3fb2f9de2\n" +
			"parent 6d4683dfec45407edb4e8124ce3c32c7ee570969\n" +
			"author James Robinson <jamesr@chromium.org> 1398979283 -0700\n" +
			"committer James Robinson <jamesr@chromium.org> 1398979283 -0700\n\n" +
			"pretty print index entries\n"),
		[]byte("commit 183\x00tree fbe461fb502beff7c0075f7179fe168599502491\n" +
			"author James Robinson <jamesr@chromium.org> 1398372819 -0700\n" +
			"committer James Robinson <jamesr@chromium.org> 1398372819 -0700\n\n" +
			"Add readme\n")}

	expected := []string{`commit 919b32c0b3cdb2b80ed7daa741b1fe88176b4264
Author: James Robinson <jamesr@chromium.org>
Date:   Thu May 1 14:21:23 2014 -0700

    pretty print index entries
    
`, `commit 9072f9473cd87dcc76b213853cce7acd380b689f
Author: James Robinson <jamesr@chromium.org>
Date:   Thu Apr 24 13:53:39 2014 -0700

    Add readme
    
`}

	origParseObjectFile := parseObjectFile
	parseObjectFile = func(name string) (object, error) {
		idx := -1
		for i := range commitHash {
			if name == commitHash[i] {
				idx = i
			}
		}
		if idx == -1 {
			return object{}, fmt.Errorf("unknown name %s", name)
		}
		b := bytes.NewBuffer(commitBytes[idx])
		o, err := parseObject(b)
		if err != nil {
			return object{}, err
		}
		return *o, err
	}
	defer func() { parseObjectFile = origParseObjectFile }()

	for i := range commitHash {
		actual, err := showCommit(commitHash[i])
		if err != nil {
			t.Error("error prettying commit: %v case %d\n", err, i)
		}

		if actual != expected[i] {
			t.Errorf("expected \"%v\" got \"%v\" case %d", expected[i], actual, i)
		}
	}
}
