// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"flag"
	"fmt"
	"os"
)

func lsFiles() {
	fs := flag.NewFlagSet("ls-files", flag.ExitOnError)
	stage := fs.Bool("s", false, "Show staged contents' object name, mode bits and stage number in the output")
	fs.Parse(os.Args[2:])
	_, entries, _, _, err := mapIndexFile(".git/index")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, e := range entries {
		if *stage {
			stageNumber := 0 // TODO: what is this?
			fmt.Printf("%o %x %d\t%s\n", e.mode, e.hash, stageNumber, string(e.path))
		} else {
			fmt.Println(string(e.path))
		}
	}
	_ = stage
}
