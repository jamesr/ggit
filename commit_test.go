// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"
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

	origParseObjectFile := ParseObjectFile
	ParseObjectFile = func(name string) (Object, error) {
		idx := -1
		for i := range commitHash {
			if name == commitHash[i] {
				idx = i
			}
		}
		if idx == -1 {
			return Object{}, fmt.Errorf("unknown name %s", name)
		}
		b := bytes.NewBuffer(commitBytes[idx])
		o, err := parseObject(nopCloser{b}, nil)
		if err != nil {
			return Object{}, err
		}
		return *o, err
	}
	defer func() { ParseObjectFile = origParseObjectFile }()

	for i := range commitHash {
		actual, err := showCommit(commitHash[i])
		if err != nil {
			t.Errorf("error prettying commit: %v case %d\n", err, i)
		}

		if actual != expected[i] {
			t.Errorf("expected \"%v\" got \"%v\" case %d", expected[i], actual, i)
		}
	}
}

func TestParseKnownFields(t *testing.T) {
	str := `tree 1c5641428ab2aad75d9874abedb821fd9ad01205
parent 8fe3ee67adcd2ee9372c7044fa311ce55eb285b4
parent fe191fcaa58cb785c804465a0da9bcba9fd9e822
author Junio C Hamano <gitster@pobox.com> 1398102789 -0700
committer Junio C Hamano <gitster@pobox.com> 1398102789 -0700

Merge git://bogomips.org/git-svn

* git://bogomips.org/git-svn:
  Git 2.0: git svn: Set default --prefix='origin/' if --prefix is not given`

	r := bytes.NewBuffer([]byte(str))
	c := commit{}
	err := parseKnownFields(&c, r)
	if err != nil {
		t.Fatal(err)
	}
	s := `Merge git://bogomips.org/git-svn

* git://bogomips.org/git-svn:
  Git 2.0: git svn: Set default --prefix='origin/' if --prefix is not given`

	expected := commit{
		tree: "1c5641428ab2aad75d9874abedb821fd9ad01205",
		Parent: []string{"8fe3ee67adcd2ee9372c7044fa311ce55eb285b4",
			"fe191fcaa58cb785c804465a0da9bcba9fd9e822"},
		author:         "Junio C Hamano",
		authorEmail:    "gitster@pobox.com",
		committer:      "Junio C Hamano",
		committerEmail: "gitster@pobox.com",
		date:           time.Unix(1398102789, 0),
		zone:           "-0700",
		messageStr:     &s}

	c.message()

	if !reflect.DeepEqual(c, expected) {
		t.Errorf("does not match %v vs %v", c, expected)
	}
}

func BenchmarkParsePersonLine(b *testing.B) {
	lines := []string{"author Junio C Hamano <gitster@pobox.com> 1398106469 -0700",
		"author Elia Pinto <gitter.spiros@gmail.com> 1397669398 -0700",
		"committer Linus Torvalds <torvalds@ppc970.osdl.org> 1112976998 -0700"}
	whom := []string{"author", "author", "committer"}
	for i := 0; i < b.N; i++ {
		for j := range lines {
			parsePersonLine(lines[j], whom[j])
		}
	}
}
