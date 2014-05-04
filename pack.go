package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	OBJ_BAD    = -1
	OBJ_NONE   = 0
	OBJ_COMMIT = 1
	OBJ_TREE   = 2
	OBJ_BLOB   = 3
	OBJ_TAG    = 4
	/* 5 for future expansion */
	OBJ_OFS_DELTA = 6
	OBJ_REF_DELTA = 7
	OBJ_ANY
	OBJ_MAX
)

func parsePackObjectHeader(data []byte) (byte, int, uint32, error) {
	used := uint32(0)
	c := data[used]
	used++
	t := (c >> 4) & 0x07

	size := int(c & 0x15)
	shift := uint(4)

	for (c & 0x80) != 0 {
		if used >= uint32(len(data)) {
			return 0, 0, 0, errors.New("bad object header")
		}
		c = data[used]
		used++
		size += int(c&0x7f) << shift
		shift += 7
	}

	return t, size, used, nil
}

type packFile struct {
	numObjects uint32
	data       []byte // includes the header
}

type packIndexFile struct {
	numEntries                       uint32
	hashes, crc32s, smallByteOffsets []byte
	data                             []byte
}

func parsePackFile(data []byte) (packFile, error) {
	if bytes.Compare(data[:4], []byte("PACK")) != 0 {
		return packFile{}, fmt.Errorf("invalid signature %s", string(data[:4]))
	}
	version := binary.BigEndian.Uint32(data[4:8])
	if version != 2 {
		return packFile{}, fmt.Errorf("unsupported version %d", 2)
	}
	checksum := sha1.Sum(data[:len(data)-sha1.Size])
	if bytes.Compare(data[len(data)-sha1.Size:], checksum[:]) != 0 {
		return packFile{}, errors.New("bad checksum")
	}
	numObjects := binary.BigEndian.Uint32(data[8:12])

	return packFile{numObjects: numObjects, data: data}, nil
}

func (p packFile) extractObject(offset uint32) (object, error) {
	t, size, used, err := parsePackObjectHeader(p.data[offset:])
	if err != nil {
		return object{}, err
	}
	if t < OBJ_COMMIT && t > OBJ_BLOB {
		return object{}, fmt.Errorf("unsupported type %d", t)
	}
	br := bytes.NewReader(p.data[offset+uint32(used):])
	zr, err := zlib.NewReader(br)
	if err != nil {
		return object{}, err
	}
	return object{objectType: objectTypeStrings[t], size: uint32(size), file: nil, reader: bufio.NewReader(zr)}, nil
}

func parsePackIndexFile(data []byte) (packIndexFile, error) {
	const magic = "\377tOc"
	idx := packIndexFile{}
	if bytes.Compare(data[:4], []byte(magic)) != 0 {
		return idx, fmt.Errorf("bad magic: %s", string(data[:4]))
	}
	version := binary.BigEndian.Uint32(data[4:8])
	if version != 2 {
		return idx, fmt.Errorf("unsupported index format %d", version)
	}
	checksum := sha1.Sum(data[:len(data)-sha1.Size])
	if bytes.Compare(checksum[:], data[len(data)-sha1.Size:]) != 0 {
		return idx, fmt.Errorf("bad checksum: %x", checksum)
	}

	entriesPerByte := make([]uint32, 256)
	numEntries := uint32(0)
	for i := 0; i < 256; i++ {
		fanOut := binary.BigEndian.Uint32(data[8+i*4 : 12+i*4])
		if fanOut != 0 {
			entriesPerByte[i] = fanOut - numEntries
			numEntries = fanOut
		}
	}
	const fanOutTableSize = 256 * 4
	hashesTableOffset := 8 + uint32(fanOutTableSize)
	crc32TableOffset := hashesTableOffset + numEntries*sha1.Size
	smallByteOffsetTableOffset := crc32TableOffset + numEntries*4
	largeByteOffsetTableOffset := smallByteOffsetTableOffset + numEntries*4

	p := packIndexFile{
		numEntries:       numEntries,
		hashes:           data[hashesTableOffset:crc32TableOffset],
		crc32s:           data[crc32TableOffset:smallByteOffsetTableOffset],
		smallByteOffsets: data[smallByteOffsetTableOffset:largeByteOffsetTableOffset],
		data:             data}
	return p, nil
}

func (idx *packIndexFile) hash(i int) []byte {
	return idx.hashes[i*sha1.Size : (i+1)*sha1.Size]
}

func (idx *packIndexFile) offset(i int) uint32 {
	return binary.BigEndian.Uint32(idx.smallByteOffsets[i*4 : (i+1)*4])
}
