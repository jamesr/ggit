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

func printCommitChain(hash string) error {
	// syscall overhead in Println is surprisingly high, so do the printing itself
	// in a goroutine
	const bufSize = 256
	hashes := make(chan string, bufSize)
	done := make(chan interface{})
	go func() {
		// it's faster to send big strings to Println(), even though it means extra
		// allocations to form the joined string.
		buf := make([]string, bufSize)
		i := 0
		for h := range hashes {
			buf[i] = h
			i++
			if i == bufSize {
				fmt.Println(strings.Join(buf, "\n"))
				i = 0
			}
		}
		fmt.Println(strings.Join(buf[:i], "\n"))
		done <- nil
	}()
	for {
		c, err := readCommit(hash)
		if err != nil {
			return err
		}
		hashes <- hash
		if len(c.parent) == 0 {
			break
		}
		hash = c.parent[0]
		c.Close()
	}
	close(hashes)
	<-done
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
	bench := fs.Bool("bench", false, "loop for benchtime seconds and report time taken")
	benchtime := fs.Int("benchtime", 5, "time to loop for (seconds) when benchmarking")
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
		for time.Since(start) < time.Duration(*benchtime)*time.Second {
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
