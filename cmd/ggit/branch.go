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
	"text/tabwriter"

	"github.com/jamesr/ggit"
)

func branch(args []string) {
	fs := flag.NewFlagSet("branch", flag.ExitOnError)
	var verbose bool
	fs.BoolVar(&verbose, "v", false, "")
	fs.Parse(args)

	// assume --list for now
	branches, err := ggit.ReadBranches()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to fetch branches", err)
		os.Exit(1)
	}
	current, err := ggit.CurrentBranch()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to fetch current branch", err)
		os.Exit(1)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)
	defer tw.Flush()
	for _, b := range branches {
		prefix := " "
		if b.Name == current {
			prefix = "*" // TODO: color code?
		}
		if verbose {
			c, err := ggit.ReadCommit(b.Hash)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error reading commit", b.Hash, err)
				os.Exit(1)
			}
			defer c.Close()
			m := c.Message()
			n := strings.IndexByte(m, '\n')
			if n != -1 {
				m = m[:n]
			}
			fmt.Fprintln(tw, strings.Join([]string{prefix, b.Name, b.Hash[:7], m}, "\t"))
		} else {
			fmt.Fprintf(tw, "%s\t%s\n", prefix, b.Name)
		}
	}
}
