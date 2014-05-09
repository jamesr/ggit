// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type compressedDeltaReader struct {
	baseCompressed   []byte
	deltasCompressed [][]byte
	r                io.Reader // Lazily set on first access
}

func readAllBytes(compressed []byte) ([]byte, error) {
	r, err := getZlibReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer returnZlibReader(r)
	b := bytes.NewBuffer(nil)
	_, err = io.Copy(b, r)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (d *compressedDeltaReader) Read(b []byte) (int, error) {
	if d.r == nil {
		base, err := readAllBytes(d.baseCompressed)
		if err != nil {
			return 0, fmt.Errorf("error decompressing base: %v", err)
		}
		patched := base
		for i := range d.deltasCompressed {
			delta, err := readAllBytes(d.deltasCompressed[len(d.deltasCompressed)-i-1])
			if err != nil {
				return 0, fmt.Errorf("error decompressing delta: %v", err)
			}
			patched, err = patchDelta(patched, delta)
			if err != nil {
				return 0, fmt.Errorf("error applying delta: %v", err)
			}
		}
		d.r = bytes.NewReader(patched)
	}
	return d.r.Read(b)
}

func deltaHeaderSize(delta []byte) (uint32, int) {
	used := 0
	size := uint32(0)
	shift := uint32(0)
	for {
		c := delta[used]
		used++
		size |= uint32(c&0x7f) << shift
		shift += 7
		if (c&0x80) == 0 || used == len(delta) {
			return size, used
		}
	}
}

func patchDelta(base, delta []byte) ([]byte, error) {
	if len(delta) < 4 {
		return nil, fmt.Errorf("delta too small: %d", delta)
	}

	expectedSourceSize, used := deltaHeaderSize(delta)
	if expectedSourceSize != uint32(len(base)) {
		return nil, fmt.Errorf("source size %d but delta header says %d",
			len(base), expectedSourceSize)
	}
	resultSize, moreUsed := deltaHeaderSize(delta[used:])
	used += moreUsed
	result := make([]byte, resultSize)
	resultOff := 0

	for i := used; i < len(delta); {
		cmd := delta[i]
		i++
		if (cmd & 0x80) != 0 {
			copyOffset, copySize := 0, 0
			if (cmd & 0x01) != 0 {
				copyOffset = int(delta[i])
				i++
			}
			if (cmd & 0x02) != 0 {
				copyOffset |= int(delta[i]) << 8
				i++
			}
			if (cmd & 0x04) != 0 {
				copyOffset |= int(delta[i]) << 16
				i++
			}
			if (cmd & 0x08) != 0 {
				copyOffset |= int(delta[i]) << 24
				i++
			}
			if (cmd & 0x10) != 0 {
				copySize = int(delta[i])
				i++
			}
			if (cmd & 0x20) != 0 {
				copySize |= int(delta[i]) << 8
				i++
			}
			if (cmd & 0x40) != 0 {
				copySize |= int(delta[i]) << 16
				i++
			}
			if copySize == 0 {
				copySize = 0x10000
			}
			copy(result[resultOff:resultOff+copySize], base[copyOffset:copyOffset+copySize])
			resultOff += copySize
		} else if cmd != 0 {
			copySize := int(cmd)
			copy(result[resultOff:resultOff+copySize], delta[i:i+copySize])
			resultOff += copySize
			i += copySize
		} else {
			return nil, errors.New("unexpected delta opcode 0")
		}
	}
	return result, nil
}
