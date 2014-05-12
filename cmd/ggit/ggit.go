// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func runCommand(cmd string, args []string) {
	switch cmd {
	case "branch":
		branch(args)
	case "cat-file":
		catFile(args)
	case "dump-index":
		dumpIndex(args)
	case "ls-files":
		lsFiles(args)
	case "ls-tree":
		lsTree(args)
	case "rev-list":
		revList(args)
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", cmd)
	}
}
func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile := flag.String("memprofile", "", "write memory profile to file")
	bench := flag.Bool("bench", false, "loop for benchtime seconds and report time taken")
	benchtime := flag.Int("benchtime", 5, "time to loop for (seconds) when benchmarking")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: ggit <command>")
		os.Exit(1)
	}
	cmd := flag.Arg(0)
	args := flag.Args()[1:]
	start := time.Now()
	count := 0
	if *bench {
		for time.Since(start) < time.Duration(*benchtime)*time.Second {
			runCommand(cmd, args)
			count++
		}
		duration := time.Now().Sub(start).Seconds()
		fmt.Fprintf(os.Stderr, "%f iter/sec in %.1f seconds\n", float64(count)/duration, duration)
		return
	}
	if *memprofile != "" {
		runtime.MemProfileRate = 1
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			runCommand(cmd, args)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}
	runCommand(cmd, args)
}
