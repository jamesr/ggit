// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"time"
)

func parseTime(data []byte) time.Time {
	ctimeSeconds := binary.BigEndian.Uint32(data[:4])
	ctimeNanos := binary.BigEndian.Uint32(data[4:8])
	return time.Unix(int64(ctimeSeconds), int64(ctimeNanos))
}

type Entry struct {
	Ctime, Mtime             time.Time
	Dev, Ino, Mode, Uid, Gid uint32
	Size                     uint32
	Hash                     [sha1.Size]byte
	Flags                    uint16
	ExtendedFlags            uint16
	Path                     []byte
}

func (e Entry) String() string {
	const layout = "Jan 2 15:04"
	s := fmt.Sprintf("ctime %s mtime %s ", e.Ctime.Format(layout), e.Mtime.Format(layout))
	s += fmt.Sprintf("dev %d ino %d mode %o uid %d gid %d size %d ", e.Dev, e.Ino, e.Mode, e.Uid, e.Gid, e.Size)
	s += fmt.Sprintf("sha1 %x ", e.Hash)
	s += fmt.Sprintf("flags v: %d extended: %d stage: %d name length %d ", e.Flags>>15, e.Flags&0x4000>>14, e.Flags&0x3000>>12, e.Flags&0x0fff)
	s += fmt.Sprintf("extended flags: %d ", e.ExtendedFlags)
	s += fmt.Sprintf("path: %s", string(e.Path))
	return s
}

// parseEntry parses an entry from the file into an entry struct and returns the
// length of the entry in bytes. data should be a slice pointing at the start of
// the entry.
// If the entry cannot be parsed, an error is returned and the other return
// values are not meaningful.
func parseEntry(data []byte) (e Entry, length uint32, err error) {
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
	e.Ctime = parseTime(consume(8))
	e.Mtime = parseTime(consume(8))
	e.Dev = binary.BigEndian.Uint32(consume(4))
	e.Ino = binary.BigEndian.Uint32(consume(4))
	e.Mode = binary.BigEndian.Uint32(consume(4))
	e.Uid = binary.BigEndian.Uint32(consume(4))
	e.Gid = binary.BigEndian.Uint32(consume(4))
	e.Size = binary.BigEndian.Uint32(consume(4))
	copy(e.Hash[:], consume(sha1.Size))
	e.Flags = binary.BigEndian.Uint16(consume(2))
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
	e.Path = data[:pathLength]
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

func parseEntries(data []byte, numEntries uint32) ([]Entry, uint32, error) {
	entries := make([]Entry, numEntries)
	entriesLen := uint32(0)
	for i := 0; i < int(numEntries); i++ {
		e, entryLen, err := parseEntry(data)
		if err != nil {
			return nil, 0, err
		}
		entries[i] = e
		data = data[entryLen:]
		entriesLen += entryLen
	}
	return entries, entriesLen, nil
}

type extension struct {
	Signature []byte
	Size      uint32
}

func parseExtensions(data []byte) ([]extension, error) {
	extensions := make([]extension, 0)
	for len(data) > 0 {
		if len(data) < 8 {
			return nil, fmt.Errorf("Not enough bytes for signature and size: %v", len(data))
		}
		e := extension{
			Signature: data[:4],
			Size:      binary.BigEndian.Uint32(data[4:8])}
		if len(data) < 8+int(e.Size) {
			return nil, fmt.Errorf("Not enough bytes for extension data, expecting %v but only have %v",
				e.Size, len(data)-8)
		}
		data = data[8+e.Size:]
		extensions = append(extensions, e)
	}
	return extensions, nil
}

var mmapFile = func(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Error opening %s: %v", path, err)
	}
	fileinfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("Error statting %s: %v", path, err)
	}
	length := int(fileinfo.Size())
	if length < 12+sha1.Size { // 12 byte header at start, SHA-1 checksum at end.
		return nil, fmt.Errorf("Index file too small, %d", length)
	}
	flags := 0
	return syscall.Mmap(int(file.Fd()), 0, length, syscall.PROT_READ, flags)
}

func MapIndexFile(filename string) (version uint32, entries []Entry, extensions []extension, data []byte, err error) {
	data, err = mmapFile(filename)
	if err != nil {
		return
	}
	numEntries := uint32(0)
	version, numEntries, err = parseIndexFileHeader(data)
	if err != nil {
		return
	}
	entries, entriesLen, entriesErr := parseEntries(data[12:], numEntries)
	if entriesErr != nil {
		err = fmt.Errorf("Error parsing entries: %v", entriesErr)
		return
	}
	extensions, extensionsErr := parseExtensions(data[12+entriesLen : len(data)-sha1.Size])
	if extensionsErr != nil {
		err = fmt.Errorf("Error parsing extensions: %v", extensionsErr)
		return
	}
	return
}
