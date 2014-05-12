package ggit

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Branch struct {
	Name, Hash string
}

func ListBranches() ([]Branch, int, error) {
	current, err := CurrentBranch()
	if err != nil {
		return nil, -1, err
	}
	// TODO: understand .git/packed-refs
	branches := []Branch(nil)
	dir := ".git/refs/heads/"
	dirs, err := ioutil.ReadDir(".git/refs/heads/")
	if err != nil {
		return nil, -1, err
	}
	currentIdx := -1

	for i, n := range dirs {
		if n.IsDir() {
			continue
		}
		f, err := os.Open(dir + n.Name())
		if err != nil {
			return nil, -1, err
		}
		asciiHash := make([]byte, 40)
		_, err = io.ReadFull(f, asciiHash)
		if err != nil {
			return nil, -1, err
		}
		if n.Name() == current {
			currentIdx = i
		}
		branches = append(branches, Branch{Name: n.Name(), Hash: string(asciiHash)})
	}
	return branches, currentIdx, nil
}

func CurrentBranch() (string, error) {
	b, err := ioutil.ReadFile(".git/HEAD")
	if err != nil {
		return "", err
	}
	s := string(b)
	const refPrefix = "ref: "
	if strings.HasPrefix(s, refPrefix) {
		ref := s[len(refPrefix):]
		if strings.HasPrefix(ref, "refs/heads/") {
			return ref[len("refs/heads/") : len(ref)-1], nil
		}
	}

	return s, nil
}
