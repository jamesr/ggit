package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/jamesr/ggit"
)

func statusWalk(path string, info os.FileInfo, err error) error {
	if info.IsDir() && info.Name() == ".git" {
		return filepath.SkipDir
	}
	if info.IsDir() {
		return nil
	}
	tracked := false
	// TODO: map from path->entry would be faster for the common case of non-moved files
	for _, e := range indexEntries {
		if string(e.Path) == path {
			modified := info.ModTime() != e.Mtime
			if modified {
				fmt.Printf("%s modified\n", path)
			}
			tracked = true
			break
		}
	}
	if !tracked {
		fmt.Printf("%s untracked\n", path)
	}
	return nil
}

var indexEntries []ggit.Entry

func foo() error {
	_, entries, _, data, err := ggit.MapIndexFile(".git/index")
	if err != nil {
		return err
	}
	defer syscall.Munmap(data)
	indexEntries = entries

	err = filepath.Walk(".", statusWalk)
	if err != nil {
		return err
	}
	return err
}

func status(args []string) {
	err := foo()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
