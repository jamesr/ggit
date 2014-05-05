package main

import (
	"bytes"
	"fmt"
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

func (p *pack) findHash(hash []byte) *object {
	// TODO: hashes are sorted, should binary search
	for i := 0; uint32(i) < p.idx.numEntries; i++ {
		if bytes.Compare(hash, p.idx.hash(i)) == 0 {
			if p.p == nil {
				err := p.parsePackFile()
				if err != nil {
					panic(err)
				}
			}
			fmt.Println("extracting from index offset ", p.idx.offset(i))
			o, err := p.p.extractObject(p.idx.offset(i))
			if err != nil {
				panic(err)
			}
			return &o
		}
	}
	return nil
}

var parsedPackFiles = []*pack(nil) // nil means not yet checked, empty means no pack files

func findHash(hash []byte) (*object, error) {
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
