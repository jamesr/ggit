package ggit

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"syscall"
	"time"
)

type entry struct {
	ctime, mtime             time.Time
	dev, ino, mode, uid, gid uint32
	size                     uint32
	hash                     [sha1.Size]byte
	flags                    uint16
	extendedFlags            uint16
	path                     string
}

func parseTime(data []byte) time.Time {
	ctimeSeconds := binary.BigEndian.Uint32(data[:4])
	ctimeNanos := binary.BigEndian.Uint32(data[4:8])
	return time.Unix(int64(ctimeSeconds), int64(ctimeNanos))
}

// parseEntry parses an entry from the file into an entry struct and returns the
// length of the entry in bytes. data should be a slice pointing at the start of
// the entry.
// If the entry cannot be parsed, an error is returned and the other return
// values are not meaningful.
func parseEntry(data []byte) (e entry, length uint32, err error) {
	const minEntryLen = 70
	if len(data) < minEntryLen {
		err = fmt.Errorf("Entry too short: %v, must be at least %v bytes",
			len(data), minEntryLen)
		return
	}
	consume := func(numBytes uint32) []byte {
		oldData := data
		data = data[numBytes:]
		length += numBytes
		return oldData
	}
	e.ctime = parseTime(consume(8))
	e.mtime = parseTime(consume(8))
	e.dev = binary.BigEndian.Uint32(consume(4))
	e.ino = binary.BigEndian.Uint32(consume(4))
	e.mode = binary.BigEndian.Uint32(consume(4))
	e.uid = binary.BigEndian.Uint32(consume(4))
	e.gid = binary.BigEndian.Uint32(consume(4))
	e.size = binary.BigEndian.Uint32(consume(4))
	copy(e.hash[:], consume(sha1.Size))
	e.flags = binary.BigEndian.Uint16(consume(2))
	// TODO(jamesr): If version >= 3, 16 bit extended flags
	// data now points to the first byte of the path. In versions <= 3, this is a
	// NUL-terminated string followed by 0-7 bytes of additional padding to round
	// the length out to a multiple of 8 bytes.
	// TODO(jamesr): If version >= 4, this is prefix-compressed relative to the
	// previous path.
	pathLength := uint32(0)
	for data[pathLength] != 0 {
		pathLength++
	}
	e.path = string(data[:pathLength])
	length += pathLength
	// There are between 1-8 NUL bytes at the end of each entry to pad it to an
	// 8-byte boundary.
	length = ((length / 8) + 1) * 8
	return
}

func parseIndexFileHeader(data []byte) (version, entries uint32, err error) {
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
	version, entries, err = parseIndexFileHeader(data)
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
