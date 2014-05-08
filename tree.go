package main

import (
	"bufio"
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
		r := bufio.NewReaderSize(tree.reader, 64)
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

func prettyPrintTree(tree object, recurse, dirsOnly bool, dir string) (string, error) {
	entries, err := parseTreeEntries(tree)
	if err != nil {
		return "", err
	}
	s := ""
	add := func(line string) {
		if len(s) != 0 {
			s += "\n"
		}
		s += line
	}
	for _, e := range entries {
		o, err := parseObjectFile(fmt.Sprintf("%x", e.hash))
		if err != nil {
			return "", fmt.Errorf("error on entry %v %x: %v", e, e.hash, err)
		}
		isTree := o.objectType == "tree"
		if (!recurse && !dirsOnly) || (dirsOnly && isTree) || (recurse && !isTree && !dirsOnly) {
			add(fmt.Sprintf("%s %s %x\t%s%s", e.mode, o.objectType, e.hash, dir, e.name))
		}
		if recurse && isTree {
			subTree, err := prettyPrintTree(o, recurse, dirsOnly, dir+e.name+"/")
			if err != nil {
				return "", err
			}
			if len(subTree) > 0 {
				add(subTree)
			}
		}
	}
	return s, nil
}
