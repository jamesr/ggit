package main

import (
	"flag"
	"fmt"
	"os"
)

func lsTree() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	d := fs.Bool("d", false, "Show only the named tree entry itself, not its children")
	r := fs.Bool("r", false, "Recurse into sub-trees")
	fs.Parse(os.Args[2:])
	treeish := fs.Arg(fs.NArg() - 1)
	if len(treeish) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: git ls-file <tree-ish>")
		os.Exit(1)
	}
	dumpTree(treeish, *r, *d)
}

func dumpTree(treeish string, r, d bool) {
	o, err := parseObjectFile(treeish)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading tree-ish %s: %v\n", treeish, err)
		os.Exit(1)
	}
	defer o.Close()
	s, err := prettyPrintTree(o, r, d, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error processing tree-ish %s: %v\n", treeish, err)
		os.Exit(1)
	}
	fmt.Println(s)
}
