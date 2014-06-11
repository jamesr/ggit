// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type Branch struct {
	Name, Hash string
}

func readBranches() ([]Branch, error) {
	// TODO: understand .git/packed-refs
	branches := []Branch(nil)
	dir := ".git/refs/heads/"
	dirs, err := ioutil.ReadDir(".git/refs/heads/")
	if err != nil {
		return nil, err
	}

	for _, n := range dirs {
		if n.IsDir() {
			continue
		}
		f, err := os.Open(dir + n.Name())
		if err != nil {
			return nil, err
		}
		asciiHash := make([]byte, 40)
		_, err = io.ReadFull(f, asciiHash)
		if err != nil {
			return nil, err
		}
		branches = append(branches, Branch{Name: n.Name(), Hash: string(asciiHash)})
	}
	return branches, nil
}

var branches []Branch
var branchesErr error
var branchesOnce sync.Once

func ReadBranches() ([]Branch, error) {
	branchesOnce.Do(func() {
		branches, branchesErr = readBranches()
	})
	return branches, branchesErr
}

const refPrefix = "ref: "

func CurrentBranch() (string, error) {
	b, err := ioutil.ReadFile(".git/HEAD")
	if err != nil {
		return "", err
	}
	s := string(b)
	if strings.HasPrefix(s, refPrefix) {
		ref := s[len(refPrefix):]
		if strings.HasPrefix(ref, "refs/heads/") {
			return ref[len("refs/heads/") : len(ref)-1], nil
		}
	}

	return s, nil
}
