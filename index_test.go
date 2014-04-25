package ggit

import (
  "crypto/sha1"
  "testing"
)

type testCase struct {
  data []byte
  version, entries int
  hasError bool
}

func TestValidIndexFile(t *testing.T) {
  cases := []testCase{}
  zeros := make([]byte, 12 + sha1.Size)
  for i := range zeros {
    zeros[i] = 0
  }
  cases = append(cases, testCase{zeros, 0, 0, true})
  for i := range cases {
    version, entries, err := parseIndexFile(cases[i].data)
    if cases[i].hasError && err == nil {
      t.Errorf("Expected but did not get error on case %v", i)
    }
    if version != cases[i].version {
      t.Errorf("Expected version %v but got %v", cases[i].version, version)
    }
    if entries != cases[i].entries {
      t.Errorf("Expected entries %v but got %v", cases[i].entries, entries)
    }
  }
}
