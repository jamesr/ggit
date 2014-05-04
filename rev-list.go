package main

import (
	"fmt"
	"os"
)

func printCommitChain(hash string) error {
	c, err := readCommit(hash)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(hash)
	if len(c.parent) > 0 {
		printCommitChain(c.parent[0])
	}
	return nil
}

func revList() {
	if len(os.Args) != 3 || !hashRe.MatchString(os.Args[2]) {
		fmt.Fprintf(os.Stderr, "Usage: ggit rev-list <hash>\n")
		os.Exit(1)
	}
	err := printCommitChain(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
