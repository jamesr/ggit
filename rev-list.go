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

func printCommitChain(hash string) error {
	for {
		c, err := readCommit(hash)
		if err != nil {
			return err
		}
		fmt.Println(hash)
		if len(c.parent) == 0 {
			return nil
		}
		hash = c.parent[0]
		if c.messageCloser != nil {
			c.messageCloser.Close()
		}
	}
}

func revList(narg int) {
	if len(os.Args) < 3 || !hashRe.MatchString(os.Args[len(os.Args)-1]) {
		fmt.Fprintf(os.Stderr, "Usage: ggit rev-list <hash>\n")
		os.Exit(1)
	}
	fs := flag.NewFlagSet("rev-list", flag.ExitOnError)
	memprofile := fs.String("memprofile", "", "write memory profile to file")
	bench := fs.Bool("bench", false, "loop for 5 seconds and report time taken")
	fs.Parse(os.Args[narg:])
	start := time.Now()
	count := 0
	if *memprofile != "" {
		runtime.MemProfileRate = 1
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < 10; i++ {
			err = printCommitChain(fs.Arg(fs.NArg() - 1))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}
	if *bench {
		for time.Since(start) < time.Duration(100)*time.Second {
			err := printCommitChain(fs.Arg(fs.NArg() - 1))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			count++
		}
		duration := time.Now().Sub(start).Seconds()
		fmt.Fprintf(os.Stderr, "%f iter/sec in %.1f seconds\n", float64(count)/duration, duration)
		return
	}
	err := printCommitChain(fs.Arg(fs.NArg() - 1))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
