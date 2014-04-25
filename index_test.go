package ggit

import (
	"crypto/sha1"
	"encoding/binary"
	"testing"
)

type testCase struct {
	data             []byte
	version, entries int
	hasError         bool
}

func fillSignature(data []byte) {
	data[0] = 'D'
	data[1] = 'I'
	data[2] = 'R'
	data[3] = 'C'
}

func fillChecksum(data []byte) {
  idx := len(data) - sha1.Size
  checksum := sha1.Sum(data[0 : idx])
  for i := 0; i < sha1.Size; i++ {
    data[idx + i] = checksum[i]
  }
}

func TestValidIndexFile(t *testing.T) {
	var zeros [12 + sha1.Size]byte

	var badChecksum [12 + sha1.Size]byte
	fillSignature(badChecksum[:])

	var version2 [12 + sha1.Size]byte
	fillSignature(version2[:])
	binary.BigEndian.PutUint32(version2[4:8], 2)
  fillChecksum(version2[:])

	var entries4 [12 + sha1.Size]byte
	fillSignature(entries4[:])
	binary.BigEndian.PutUint32(entries4[8:12], 4)
  fillChecksum(entries4[:])

	cases := []testCase{{zeros[:], 0, 0, true},
		{badChecksum[:], 0, 0, true},
		{version2[:], 2, 0, false},
		{entries4[:], 0, 4, false}}

	for i := range cases {
		version, entries, err := parseIndexFile(cases[i].data)
    gotError := err != nil
		if cases[i].hasError != gotError {
			t.Errorf("Expected error %v did not match on case %v", cases[i].hasError, i)
		}
		if version != cases[i].version {
			t.Errorf("Expected version %v but got %v on case %v", cases[i].version,
				version, i)
		}
		if entries != cases[i].entries {
			t.Errorf("Expected entries %v but got %v on case %v", cases[i].entries,
				entries, i)
		}
	}
}
