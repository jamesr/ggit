package main

import (
	"fmt"
	"io"
)

func prettyPrintTree(tree object) (string, error) {
	fmt.Print(tree)
	b := make([]byte, 4096)
	s := ""
	for {
		_, err := tree.reader.Read(b)
		if err == io.EOF {
			return s, nil
		}
		if err != nil {
			return "", err
		}
		fmt.Printf("appending \"%s\"\n", string(b))
		s += string(b)
	}
}
