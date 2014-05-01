package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"
)

func dumpIndex() {
	filename := ".git/index"
	version, entries, extensions, data, err := mapIndexFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse index file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("version %d entries %d extensions %d\n", version, len(entries), len(extensions))
	for i := range entries {
		fmt.Printf("entry %d: %v\n", i, entries[i])
		fmt.Printf("%s %x\n", string(entries[i].path), entries[i].hash)
	}
	for i := range extensions {
		fmt.Printf("extension %d: %v size %v\n", i, string(extensions[i].signature), extensions[i].size)
	}
	err = syscall.Munmap(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not unmap: %v\n", err)
		os.Exit(1)
	}
}

func catFile() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var typeOnly, sizeOnly, existsOnly, prettyPrint bool
	fs.BoolVar(&typeOnly, "t", false, "")
	fs.BoolVar(&sizeOnly, "s", false, "")
	fs.BoolVar(&existsOnly, "e", false, "")
	fs.BoolVar(&prettyPrint, "p", false, "")
	fs.Parse(os.Args[2:])
	name := fs.Arg(fs.NArg() - 1)
	if len(name) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: ggit cat-file [-t|-s|-e|-p] <object>")
		os.Exit(1)
	}
	err := error(nil)
	switch {
	case typeOnly:
		err = dumpObjectType(name)
	case sizeOnly:
		err = dumpObjectSize(name)
	case existsOnly:
		path := nameToPath(name)
		file, err := os.Open(path)
		if err != nil {
			os.Exit(1)
		}
		file.Close()
	case prettyPrint:
		err = dumpPrettyPrint(name)
	default:
		fmt.Fprintln(os.Stderr, "Usage: ggit cat-file [-t|-s|-e|-p] <object>")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func lsTree() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.Parse(os.Args[2:])
	treeish := fs.Arg(fs.NArg() - 1)
	if len(treeish) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: git ls-file <tree-ish>")
		os.Exit(1)
	}
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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: ggit <command>")
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch {
	case cmd == "dump-index":
		dumpIndex()
	case cmd == "cat-file":
		catFile()
	case cmd == "ls-tree":
		lsTree()
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", cmd)
	}
}
