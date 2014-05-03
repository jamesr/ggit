package main

import (
	"crypto/sha1"
	"fmt"
	"io"
)

type treeEntry struct {
	mode, name string
	hash       [20]byte
}

func parseTreeEntries(tree object) ([]treeEntry, error) {
	entries := make([]treeEntry, 0)
	for {
		entry := treeEntry{}
		r := tree.reader
		mode, err := r.ReadString(' ')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entry.mode = mode[:len(mode)-1]
		if len(entry.mode) < 6 {
			entry.mode = "00000"[:6-len(entry.mode)] + entry.mode
		}

		name, err := r.ReadString(0)
		if err != nil {
			return nil, err
		}
		entry.name = name[:len(name)-1]

		n, err := r.Read(entry.hash[:])
		if err != nil {
			return nil, err
		}
		if n != sha1.Size {
			return nil, fmt.Errorf("invalid hash for tree entry, only %v bytes", n)
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

func prettyPrintTree(tree object, recurse bool, dir string) (string, error) {
	entries, err := parseTreeEntries(tree)
	if err != nil {
		return "", err
	}
	s := ""
	for _, e := range entries {
		if len(s) != 0 {
			s += "\n"
		}
		o, err := parseObjectFile(fmt.Sprintf("%x", e.hash))
		if err != nil {
			return "", fmt.Errorf("error on entry %v %x: %v", e, e.hash, err)
		}
		if recurse && o.objectType == "tree" {
			fmt.Printf("should recurse into tree %x, dir %s\n", 5, e.name)
			subTree, err := prettyPrintTree(o, recurse, e.name+"/")
			if err != nil {
				return "", err
			}
			s += subTree
		} else {
			s += fmt.Sprintf("%s %s %x\t%s%s", e.mode, o.objectType, e.hash, dir, e.name)
		}
	}
	return s, nil
}
