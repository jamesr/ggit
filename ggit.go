package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: ggit <command>")
		os.Exit(1)
	}
	cmd := flag.Arg(0)
	switch {
	case cmd == "dump-index":
		dumpIndex()
	case cmd == "cat-file":
		catFile()
	case cmd == "ls-tree":
		lsTree()
	case cmd == "ls-files":
		lsFiles()
	case cmd == "rev-list":
		revList(flag.NFlag() + 2)
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", cmd)
	}
}
