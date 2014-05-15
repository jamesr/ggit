// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jamesr/ggit"
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
		c, err := ggit.ReadCommit(hash)
		if err != nil {
			return err
		}
		hashes <- c.Hash
		if len(c.Parent) == 0 {
			break
		}
		hash = c.Parent[0]
		c.Close()
	}
	close(hashes)
	<-done
	return nil
}

func revList(args []string) {
	fs := flag.NewFlagSet("rev-list", flag.ExitOnError)
	_ = fs.Bool("first-parent", false, "prints only the first parent")
	fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: ggit rev-list <hash>\n")
		os.Exit(1)
	}
	hash := fs.Arg(0)
	err := printCommitChain(hash)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
