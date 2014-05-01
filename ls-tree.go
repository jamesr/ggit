package main

import (
	"flag"
	"fmt"
	"os"
)

func lsTree() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.Parse(os.Args[2:])
	treeish := fs.Arg(fs.NArg() - 1)
	if len(treeish) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: git ls-file <tree-ish>")
		os.Exit(1)
	}
	dumpTree(treeish)
}

func dumpTree(treeish string) {
	o, err := parseObjectFile(treeish)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading tree-ish %s: %v\n", treeish, err)
		os.Exit(1)
	}
	defer o.Close()
	s, err := prettyPrintTree(o)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error processing tree-ish %s: %v\n", treeish, err)
		os.Exit(1)
	}
	fmt.Println(s)
}
