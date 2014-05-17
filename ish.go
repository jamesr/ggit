package ggit

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
)

var readFile = ioutil.ReadFile

func CommitishToHash(committish string) string {
	if len(committish) <= sha1.Size*2 {
		allHex := true
		for _, c := range committish {
			if c < '0' && c > '9' && c < 'a' && c > 'f' {
				allHex = false
				break
			}
		}
		if allHex {
			return committish
		}
	}
	if committish == "HEAD" {
		h, err := readFile(".git/HEAD")
		if err != nil {
			panic(err)
		}
		if bytes.HasPrefix(h, []byte(refPrefix)) {
			ref := h[len(refPrefix):]
			if bytes.HasPrefix(ref, []byte("refs/heads/")) {
				ref = ref[len("refs/heads/"):]
				branches, err := ReadBranches()
				if err != nil {
					panic(err)
				}
				for _, b := range branches {
					if bytes.HasPrefix(ref, []byte(b.Name)) {
						return b.Hash
					}
				}
				panic(fmt.Sprintf("unknown ref %s", string(ref)))
			} else {
				panic(fmt.Sprintf("no idea how to handle ref %s", string(ref)))
			}
		}
		if len(h) < sha1.Size*2 {
			panic(fmt.Sprintf(".git/HEAD too short: %d", len(h)))
		}
		return string(h[:sha1.Size*2])
	}

	return ""
}
