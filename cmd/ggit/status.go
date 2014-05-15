package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/jamesr/ggit"
)

func statusWalk(path string, info os.FileInfo, err error, indexEntries []ggit.Entry, modified, untracked *[]string) error {
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
			if info.ModTime() != e.Mtime {
				*modified = append(*modified, path)
			}
			tracked = true
			break
		}
	}
	if !tracked {
		*untracked = append(*untracked, path)
	}
	return nil
}

func findModifiedAndUntracked() (modified, untracked []string, err error) {
	_, entries, _, data, err := ggit.MapIndexFile(".git/index")
	if err != nil {
		return nil, nil, err
	}
	defer syscall.Munmap(data)

	modified = make([]string, 0)
	untracked = make([]string, 0)

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		return statusWalk(path, info, err, entries, &modified, &untracked)
	})
	if err != nil {
		return nil, nil, err
	}
	return modified, untracked, nil
}

func status(args []string) {
	branch, err := ggit.CurrentBranch()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println("On branch", branch)
	// TODO: staged
	modified, untracked, err := findModifiedAndUntracked()
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
