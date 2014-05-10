// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bytes"
	"os"
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
		syscall.Munmap(p.p.data)
		p.pFile.Close()
	}
	syscall.Munmap(p.idx.data)
	p.idxFile.Close()
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
	// TODO: binary search is fast, but given that these hashes are likely to be very
	// evenly distributed we could do some newtonion something and perhaps do better.
	lo, hi := 0, int(p.idx.numEntries)
	for hi > lo {
		i := lo + (hi-lo)/2
		cmp := bytes.Compare(hash, p.idx.hash(i))
		if cmp == 0 {
			if p.p == nil {
				err := p.parsePackFile()
				if err != nil {
					panic(err)
				}
			}
			o, err := p.p.extractObject(p.idx.offset(i))
			if err != nil {
				panic(err)
			}
			return &o
		} else if cmp > 0 {
			lo = i + 1
		} else {
			hi = i
		}
	}
	return nil
}

var parsedPackFiles = []*pack(nil) // nil means not yet checked, empty means no pack files

func findHash(hash []byte) (*Object, error) {
	if parsedPackFiles == nil {
		f, err := os.Open(".git/objects/pack")
		if err != nil {
			return nil, err
		}
		names, err := f.Readdirnames(0)
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
