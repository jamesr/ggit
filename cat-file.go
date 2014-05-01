package main

import (
	"flag"
	"fmt"
	"os"
)

func dumpObjectType(name string) error {
	o, err := parseObjectFile(name)
	if err != nil {
		return err
	}
	fmt.Println(o.objectType)
	o.Close()
	return nil
}

func dumpObjectSize(name string) error {
	o, err := parseObjectFile(name)
	if err != nil {
		return err
	}
	fmt.Println(o.size)
	return nil
}

func dumpPrettyPrint(name string) error {
	o, err := parseObjectFile(name)
	if err != nil {
		return err
	}
	if o.objectType == "tree" {
		dumpTree(name)
		return nil
	}
	o.Close()
	return nil
}
func catFile() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var typeOnly, sizeOnly, existsOnly, prettyPrint bool
	fs.BoolVar(&typeOnly, "t", false, "")
	fs.BoolVar(&sizeOnly, "s", false, "")
	fs.BoolVar(&existsOnly, "e", false, "")
	fs.BoolVar(&prettyPrint, "p", false, "")
	fs.Parse(os.Args[2:])
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
		path := nameToPath(name)
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
