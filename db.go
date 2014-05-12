// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bytes"
	"os"
	"sort"
	"strings"
	"syscall"
)

// Object database. Strategy for finding an object based on name:
// 1.) normalize name to hash (TODO: implement)
// 2.) look in .git/packed-refs (TODO: figure out what this is)
// 3.) if pack files not yet parsed:
//   3.a) open .git/objects/pack/
//   3.b) go through pack files there, parse indices
// 4.) search pack files for hash
// 5.) if not found, open .git/objects/ha/sh.....

type pack struct {
	p              *packFile
	idx            packIndexFile
	baseFileName   string
	pFile, idxFile *os.File
}

func (p pack) Close() {
	if p.p != nil {
		_ = syscall.Munmap(p.p.data)
		_ = p.pFile.Close()
	}
	_ = syscall.Munmap(p.idx.data)
	_ = p.idxFile.Close()
}

func (p *pack) parsePackFile() error {
	data, err := mmapFile(".git/objects/pack/" + p.baseFileName + ".pack")
	if err != nil {
		return err
	}
	pp, err := parsePackFile(data)
	if err != nil {
		return err
	}
	p.p = &pp
	return nil
}

func (p *pack) findHash(hash []byte) *Object {
	lo := 0
	if hash[0] > 0 {
		lo = p.idx.fanOut[int(hash[0])-1]
	}
	hi := p.idx.fanOut[hash[0]]
	idx := sort.Search(hi-lo, func(i int) bool {
		return bytes.Compare(hash, p.idx.hash(i+lo)) <= 0
	}) + lo
	if idx == hi {
		return nil
	}
	cmp := bytes.Compare(hash, p.idx.hash(idx))
	if cmp != 0 {
		return nil
	}
	if p.p == nil {
		err := p.parsePackFile()
		if err != nil {
			panic(err)
		}
	}
	o, err := p.p.extractObject(p.idx.offset(idx))
	if err != nil {
		panic(err)
	}
	return &o
}

var parsedPackFiles = []*pack(nil) // nil means not yet checked, empty means no pack files

func findHash(hash []byte) (*Object, error) {
	if parsedPackFiles == nil {
		f, err := os.Open(".git/objects/pack")
		if err != nil {
			return nil, err
		}
		names, err := f.Readdirnames(0)
		if err != nil {
			return nil, err
		}
		for _, n := range names {
			if !strings.HasPrefix(n, "pack-") {
				continue
			}
			if strings.HasSuffix(n, ".idx") {
				data, err := mmapFile(".git/objects/pack/" + n)
				idx, err := parsePackIndexFile(data)
				if err != nil {
					return nil, err
				}
				parsedPackFiles = append(parsedPackFiles,
					&pack{p: nil,
						pFile:        nil,
						baseFileName: n[:len(n)-len(".idx")],
						idx:          idx,
						idxFile:      f})
			}
		}
	}

	for _, p := range parsedPackFiles {
		o := p.findHash(hash)
		if o != nil {
			return o, nil
		}
	}

	return nil, nil
}
