// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
)

type packFile struct {
	numObjects uint32
	data       []byte // includes the header
}

type packIndexFile struct {
	fanOut                           []int
	numEntries                       uint32
	hashes, crc32s, smallByteOffsets []byte
	data                             []byte
}

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

func (p packFile) parseHeader(offset uint32) (byte, int, uint32, error) {
	used := uint32(0)
	c := p.data[offset]
	used++
	t := (c >> 4) & 7

	size := int(c & 0x0f)
	shift := uint(4)

	for (c & 0x80) != 0 {
		if used+offset >= uint32(len(p.data)) {
			return 0, 0, 0, errors.New("bad object header")
		}
		c = p.data[used+offset]
		used++
		size += int(c&0x7f) << shift
		shift += 7
	}
	return t, size, used, nil
}

var verifyChecksums = false

func parsePackFile(data []byte) (packFile, error) {
	if bytes.Compare(data[:4], []byte("PACK")) != 0 {
		return packFile{}, fmt.Errorf("invalid signature %s", string(data[:4]))
	}
	version := binary.BigEndian.Uint32(data[4:8])
	if version != 2 {
		return packFile{}, fmt.Errorf("unsupported version %d", 2)
	}
	if verifyChecksums {
		checksum := sha1.Sum(data[:len(data)-sha1.Size])
		if bytes.Compare(data[len(data)-sha1.Size:], checksum[:]) != 0 {
			return packFile{}, errors.New("bad checksum")
		}
	}
	numObjects := binary.BigEndian.Uint32(data[8:12])

	return packFile{numObjects: numObjects, data: data}, nil
}

func (p packFile) extractObject(offset uint32) (Object, error) {
	t, size, used, err := p.parseHeader(offset)
	if err != nil {
		return Object{}, err
	}

	deltasCompressed := [][]byte{}
	for t == OBJ_OFS_DELTA {
		c := p.data[offset+used]
		used++
		deltaOffset := uint32(c & 0x7f)
		for (c & 0x80) != 0 {
			deltaOffset++
			c = p.data[offset+used]
			used++
			deltaOffset = (deltaOffset << 7) + uint32(c&0x7f)
		}
		if deltaOffset > offset {
			return Object{}, fmt.Errorf("bad object header delta offset %d %d", deltaOffset, offset)
		}
		// at this point, next size bytes are a delta against base. store it for use in constructing the
		// object's reader later on
		deltasCompressed = append(deltasCompressed, p.data[offset+used:offset+used+uint32(size)+8]) // TODO: +8 is a hack, figure out
		t, size, used, err = p.parseHeader(offset - deltaOffset)
		offset -= deltaOffset
	}

	if t < OBJ_COMMIT || t > OBJ_BLOB {
		return Object{}, fmt.Errorf("unsupported type %d", t)
	}
	o := Object{ObjectType: objectTypeStrings[t], Size: uint32(size), file: nil}
	if len(deltasCompressed) != 0 {
		o.Reader = &compressedDeltaReader{
			baseCompressed:   p.data[offset+used : offset+used+uint32(size)],
			deltasCompressed: deltasCompressed}

	} else {
		br := bytes.NewReader(p.data[offset+used : offset+used+uint32(size)])
		zr, err := getZlibReader(br)
		if err != nil {
			return Object{}, err
		}
		o.Reader = zr
		o.zlibReader = zr
	}
	return o, nil
}

var verifyPackChecksum = false

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
	if verifyPackChecksum {
		checksum := sha1.Sum(data[:len(data)-sha1.Size])
		if bytes.Compare(checksum[:], data[len(data)-sha1.Size:]) != 0 {
			return idx, fmt.Errorf("bad checksum: %x", checksum)
		}
	}

	fanOut := make([]int, 256)
	for i := 0; i < 256; i++ {
		fanOut[i] = int(binary.BigEndian.Uint32(data[8+i*4 : 12+i*4]))
	}
	numEntries := fanOut[255]
	const fanOutTableSize = 256 * 4
	hashesTableOffset := 8 + fanOutTableSize
	crc32TableOffset := hashesTableOffset + numEntries*sha1.Size
	smallByteOffsetTableOffset := crc32TableOffset + numEntries*4
	largeByteOffsetTableOffset := smallByteOffsetTableOffset + numEntries*4

	p := packIndexFile{
		fanOut:           fanOut,
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
	smallByteOffset := binary.BigEndian.Uint32(idx.smallByteOffsets[i*4 : (i+1)*4])
	if (smallByteOffset & (1 << 31)) != 0 {
		panic("do not support large byte offsets")
	}
	return smallByteOffset
}
