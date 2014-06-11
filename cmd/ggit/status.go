// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/jamesr/ggit"
)

func statusWalk(path string, info os.FileInfo, err error, entries []ggit.Entry) error {
	if info.IsDir() && info.Name() == ".git" {
		return filepath.SkipDir
	}
	if info.IsDir() {
		return nil
	}
	tracked := false
	// TODO: map from path->entry would be faster for the common case of non-moved files
	for _, e := range entries {
		if string(e.Path) == path {
			if info.ModTime() != e.Mtime {
				modified = append(modified, path)
			}
			tracked = true
			break
		}
	}
	if !tracked {
		untracked = append(untracked, path)
	}
	return nil
}

var modified, untracked, ignorePatterns []string

func ignored(filepath string) bool {
	if filepath == ".git" {
		return true
	}
	for _, p := range ignorePatterns {
		matched, _ := path.Match(p, filepath)
		if matched {
			return true
		}
	}
	return false
}

func parseIgnored(filepath string) {
	f, err := os.Open(filepath)
	if err == nil {
		r := bufio.NewReader(f)
		for l, err := r.ReadString('\n'); err == nil; l, err = r.ReadString('\n') {
			if l[0] == '#' {
				continue
			}
			ignorePatterns = append(ignorePatterns, l[:len(l)-1])
		}
	}

}

func walkDir(base string, entries []ggit.Entry) error {
	if ignored(base) {
		return nil
	}
	infos, err := ioutil.ReadDir(base)
	if err != nil {
		return err
	}
	for _, info := range infos {
		filepath := path.Clean(base + string(os.PathSeparator) + info.Name())
		if info.IsDir() {
			walkDir(filepath, entries)
		} else {
			if info.Name() == ".gitignore" {
				parseIgnored(filepath)
				continue
			}
			if ignored(filepath) {
				continue
			}
			tracked := false
			// TODO: map from path->entry
			for _, e := range entries {
				if string(e.Path) == filepath {
					tracked = true
					if info.ModTime() != e.Mtime {
						modified = append(modified, filepath)
					}
					// TODO: ctime, etc
				}
			}
			if !tracked {
				untracked = append(untracked, filepath)
			}

		}
	}
	return nil
}

func findModifiedAndUntracked() (err error) {
	_, entries, _, data, err := ggit.MapIndexFile(".git/index")
	if err != nil {
		return err
	}
	defer syscall.Munmap(data)

	parseIgnored(".git/info/exclude")

	modified = make([]string, 0)
	untracked = make([]string, 0)
	ignorePatterns = make([]string, 0)

	err = walkDir(".", entries)

	/*
		err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			return statusWalk(path, info, err, entries)
		})
	*/
	if err != nil {
		return err
	}
	return nil
}

func status(args []string) {
	branch, err := ggit.CurrentBranch()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println("On branch", branch)
	// TODO: staged
	err = findModifiedAndUntracked()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	if len(modified) > 0 {
		fmt.Println(`Changes not staged for commit:
  (use "git add <file>..." to update what will be committed)
  (use "git checkout -- <file>..." to discard changes in working directory)`)
		fmt.Println()
		for _, p := range modified {
			fmt.Println("\tmodified:  ", p)
		}
		fmt.Println()
	}
	if len(untracked) > 0 {
		fmt.Println(`Untracked files:
  (use "git add <file>..." to include in what will be committed)`)
		fmt.Println()
		for _, p := range untracked {
			fmt.Printf("\t%s\n", p)
		}
		fmt.Println()
	}
	// TODO: one more status line here
}
