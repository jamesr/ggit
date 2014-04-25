package ggit

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"syscall"
)

func parseIndexFile(data []byte) (version, entries uint32, err error) {
	if data[0] != 'D' || data[1] != 'I' || data[2] != 'R' || data[3] != 'C' {
		err = fmt.Errorf("Invalid signature")
		return
	}
	fileChecksum := data[len(data)-sha1.Size:]
	computedChecksum := sha1.Sum(data[0 : len(data)-sha1.Size])
	if !bytes.Equal(fileChecksum, computedChecksum[:]) {
		err = fmt.Errorf("Invalid checksum")
		return
	}
	version = binary.BigEndian.Uint32(data[4:8])
	entries = binary.BigEndian.Uint32(data[8:12])
	return
}

func mapIndexFile(filename string) (version, entries uint32, data []byte, err error) {
	file, err := os.Open(filename)
	if err != nil {
		err = fmt.Errorf("Error opening %s: %v", filename, err)
		return
	}
	fileinfo, err := file.Stat()
	if err != nil {
		err = fmt.Errorf("Error statting %s: %v", filename, err)
		return
	}
	length := int(fileinfo.Size())
	if length < 12+sha1.Size { // 12 byte header at start, SHA-1 checksum at end.
		err = fmt.Errorf("Index file too small, %d", length)
		return
	}
	flags := 0
	data, err = syscall.Mmap(int(file.Fd()), 0, length, syscall.PROT_READ, flags)
	version, entries, err = parseIndexFile(data)
	return
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: index <path to index file>\n")
		os.Exit(1)
	}
	filename := flag.Args()[0]
	version, entries, data, err := mapIndexFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse index file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("version %d entries %d\n", version, entries)
	err = syscall.Munmap(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not unmap: %v\n", err)
		os.Exit(1)
	}
}
