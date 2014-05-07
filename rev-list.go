package main

import (
	"flag"
	"fmt"
	"os"
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
	}
}

func revList() {
	if len(os.Args) < 3 || !hashRe.MatchString(os.Args[len(os.Args)-1]) {
		fmt.Fprintf(os.Stderr, "Usage: ggit rev-list <hash>\n")
		os.Exit(1)
	}
	fs := flag.NewFlagSet("rev-list", flag.ExitOnError)
	fs.Parse(os.Args[2:])
	start := time.Now()
	count := 0
	for time.Since(start) < time.Duration(5)*time.Second {
		err := printCommitChain(fs.Arg(fs.NArg() - 1))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		count++
	}
	duration := time.Now().Sub(start).Seconds()
	fmt.Fprintf(os.Stderr, "%f iter/sec in %.1f seconds\n", float64(count)/duration, duration)
}
