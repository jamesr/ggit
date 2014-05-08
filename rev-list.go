package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"
)

var chain []string

func printCommitChain(hash string) error {
	for {
		c, err := readCommit(hash)
		if err != nil {
			return err
		}
		chain = append(chain, hash)
		if len(c.parent) == 0 {
			break
		}
		hash = c.parent[0]
		c.discardZlibReader()
	}
	// The syscall overhead of fmt.Print* to stdout is pretty high, so join into one string first
	fmt.Println(strings.Join(chain, "\n"))
	return nil
}

var hashRe = regexp.MustCompile("^[a-z0-9]{40}$")

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
		for time.Since(start) < time.Duration(30)*time.Second {
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
