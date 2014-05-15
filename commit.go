// Copyright 2014 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd

package ggit

import (
	"bufio"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func parseHashLine(line, leader string) (string, error) {
	if len(line) != 42+len(leader) {
		return "", fmt.Errorf("bad commit format: \"%s\"", line)
	}
	hash := line[len(leader)+1 : len(line)-1]
	for _, c := range hash {
		if c < 'a' && c > 'z' && c < '0' && c > '9' {
			return "", fmt.Errorf("bad commit format hash: \"%s\"", line)
		}
	}
	return hash, nil
}

// "whom" SP "name possibly with many spaces" SP "<" email ">" SP timestamp SP zone NL
func parsePersonLine(line, whom string) (name, email, zone string, t time.Time, err error) {
	// work from the end since arbitrary spaces only appear in second entry
	i := 0
	for i = len(line) - 1; i >= 0; i-- {
		if line[i] == ' ' {
			zone = line[i+1 : len(line)-1]
			break
		}
	}
	if i <= 9 {
		err = fmt.Errorf("bad person line %s", line)
		return
	}
	lastSpace := i
	for i--; i >= 0; i-- {
		if line[i] == ' ' {
			timeSec := int64(0)
			timeSec, err = strconv.ParseInt(line[i+1:lastSpace], 10, 32)
			if err != nil {
				return
			}
			t = time.Unix(timeSec, 0)
			break
		}
	}
	if i <= 7 {
		err = fmt.Errorf("bad person line %s", line)
		return
	}
	lastSpace = i
	for i--; i >= 0; i-- {
		if line[i] == ' ' {
			maybeEmail := line[i+1 : lastSpace]
			if maybeEmail[0] != '<' || maybeEmail[len(maybeEmail)-1] != '>' {
				err = fmt.Errorf("bad email %s", maybeEmail)
				return
			}
			email = maybeEmail[1 : len(maybeEmail)-1]
			break
		}
	}
	if i <= 3 {
		err = fmt.Errorf("bad person line %s", line)
		return
	}
	lastSpace = i
	for i = 0; i < len(whom); i++ {
		if line[i] != whom[i] {
			err = errors.New("bad person format")
			return
		}
	}
	name = line[i+1 : lastSpace]
	return
}

type commit struct {
	hash, tree                string
	Parent                    []string
	author, authorEmail       string
	committer, committerEmail string
	date                      time.Time
	zone                      string
	messageReader             *bufio.Reader
	zlibReader                zlib.ReadCloserReset
	messageStr                *string // lazily populated from reader
}

// time.ANSIC with s/_2/2/
const timeFormat = "Mon Jan 2 15:04:05 2006"

func (c commit) String() string {
	s := "commit " + c.hash + "\n"
	s += "Author: " + c.author + " <" + c.authorEmail + ">\n"
	s += "Date:   " + c.date.Format(timeFormat) + " " + c.zone + "\n\n"
	lines := strings.Split(c.Message(), "\n")
	for _, l := range lines {
		s += "    " + l + "\n"
	}
	return s
}

func parseKnownFields(c *commit, r io.Reader) error {
	br := getBufioReader(r)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return fmt.Errorf("ReadString err %v", err)
		}
		switch {
		case line == "\n":
			c.messageReader = br
			return nil
		case strings.HasPrefix(line, "tree "):
			c.tree, err = parseHashLine(line, "tree")
			if err != nil {
				return fmt.Errorf("hash %v", err)
			}
		case strings.HasPrefix(line, "parent "):
			parent, err := parseHashLine(line, "parent")
			if err != nil {
				return fmt.Errorf("parent %v", err)
			}
			c.Parent = append(c.Parent, parent)
		case strings.HasPrefix(line, "author "):
			c.author, c.authorEmail, c.zone, c.date, err = parsePersonLine(line, "author")
			if err != nil {
				return fmt.Errorf("author %v", err)
			}
		case strings.HasPrefix(line, "committer "):
			c.committer, c.committerEmail, _, _, err = parsePersonLine(line, "committer")
			if err != nil {
				return fmt.Errorf("committer %v", err)
			}
		default:
			// unknown line, ignore for now
			// fmt.Fprintf(os.Stderr, "unknown line \"%s\"\n", line)
		}
	}
}

func parseCommitObject(o Object) (commit, error) {
	c := commit{}

	err := parseKnownFields(&c, o.Reader)
	if err != nil {
		return commit{}, fmt.Errorf("parsing known fields %v", err)
	}

	c.zlibReader = o.zlibReader
	return c, nil
}

func (c *commit) Close() {
	if c.zlibReader != nil {
		returnZlibReader(c.zlibReader)
		c.zlibReader = nil
	}
	if c.messageReader != nil {
		returnBufioReader(c.messageReader)
		c.messageReader = nil
	}
}

func (c *commit) Message() string {
	if c.messageStr == nil {
		b, err := ioutil.ReadAll(c.messageReader)
		if err != nil {
			panic(err)
		}
		c.Close()
		s := string(b)
		c.messageStr = &s
	}
	return *c.messageStr
}

func ReadCommit(hash string) (commit, error) {
	object, err := LookupObject(hash)
	if err != nil {
		return commit{}, fmt.Errorf("error parsing object %v", err)
	}
	if object.ObjectType != "commit" {
		return commit{}, fmt.Errorf("object %s has bad type: %s", hash, object.ObjectType)
	}
	c, err := parseCommitObject(object)
	if err != nil {
		return commit{}, fmt.Errorf("error parsing commit %v", err)
	}
	c.hash = hash
	return c, nil
}

func showCommit(hash string) (string, error) {
	c, err := ReadCommit(hash)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}
