// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jamesr/ggit"
)

func dumpObjectType(name string) error {
	o, err := ggit.LookupObject(name)
	if err != nil {
		return err
	}
	fmt.Println(o.ObjectType)
	o.Close()
	return nil
}

func dumpObjectSize(name string) error {
	o, err := ggit.LookupObject(name)
	if err != nil {
		return err
	}
	fmt.Println(o.Size)
	return nil
}

func dumpPrettyPrint(name string) error {
	o, err := ggit.LookupObject(name)
	if err != nil {
		return err
	}
	if o.ObjectType == "tree" {
		recurse, dirsOnly := false, false
		dumpTree(name, recurse, dirsOnly)
		return nil
	} else {
		return dumpPrettyPrintObject(o)
	}
}

func dumpPrettyPrintObject(o ggit.Object) error {
	b := bytes.NewBuffer(nil)
	_, err := io.Copy(b, o.Reader)
	if err != nil {
		return err
	}
	fmt.Print(b.String())
	o.Close()
	return nil
}

func catFile(args []string) {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var typeOnly, sizeOnly, existsOnly, prettyPrint bool
	fs.BoolVar(&typeOnly, "t", false, "")
	fs.BoolVar(&sizeOnly, "s", false, "")
	fs.BoolVar(&existsOnly, "e", false, "")
	fs.BoolVar(&prettyPrint, "p", false, "")
	fs.Parse(args)
	name := fs.Arg(fs.NArg() - 1)
	if len(name) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: ggit cat-file [-t|-s|-e|-p] <object>")
		os.Exit(1)
	}
	err := error(nil)
	switch {
	case typeOnly:
		err = dumpObjectType(name)
	case sizeOnly:
		err = dumpObjectSize(name)
	case existsOnly:
		path, err := ggit.NameToPath(name)
		if err != nil {
			os.Exit(1)
		}
		file, err := os.Open(path)
		if err != nil {
			os.Exit(1)
		}
		file.Close()
	case prettyPrint:
		err = dumpPrettyPrint(name)
	default:
		fmt.Fprintln(os.Stderr, "Usage: ggit cat-file [-t|-s|-e|-p] <object>")
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
