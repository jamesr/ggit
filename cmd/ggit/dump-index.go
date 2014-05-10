// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/jamesr/ggit"
)

func dumpIndex(args []string) {
	filename := ".git/index"
	if len(args) > 0 {
		filename = args[0]
	}
	version, entries, extensions, data, err := ggit.MapIndexFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse index file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("version %d entries %d extensions %d\n", version, len(entries), len(extensions))
	for i := range entries {
		fmt.Printf("entry %d: %v\n", i, entries[i])
		fmt.Printf("%s %x\n", string(entries[i].Path), entries[i].Hash)
	}
	for i := range extensions {
		fmt.Printf("extension %d: %v size %v\n", i, string(extensions[i].Signature), extensions[i].Size)
	}
	err = syscall.Munmap(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not unmap: %v\n", err)
		os.Exit(1)
	}
}
