package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var hashRe = regexp.MustCompile("^[a-z0-9]{40}$")

func parseHashLine(line, leader string) (string, error) {
	if len(line) != 42+len(leader) {
		return "", fmt.Errorf("bad commit format: \"%s\"", line)
	}
	hash := line[len(leader)+1 : len(line)-1]
	if !hashRe.MatchString(hash) {
		return "", fmt.Errorf("bad commit format hash: \"%s\"", line)
	}
	return hash, nil
}

func parsePersonLine(line, whom string) (name, email, zone string, t time.Time, err error) {
	s := strings.Split(line, " ")
	if len(s) < 5 {
		err = fmt.Errorf("bad format, insufficient parts for %s line in \"%s\"", whom, line)
		return
	}
	if s[0] != whom {
		err = fmt.Errorf("bad format, got %s but expected %s", s[0], whom)
	}
	timeSec, err := strconv.ParseInt(s[len(s)-2], 10, 64)
	if err != nil {
		return
	}
	maybeEmail := s[len(s)-3]
	if maybeEmail[0] != '<' || maybeEmail[len(maybeEmail)-1] != '>' {
		err = fmt.Errorf("bad email %s", maybeEmail)
		return
	}
	name = strings.Join(s[1:len(s)-3], " ")
	email = maybeEmail[1 : len(maybeEmail)-1]
	t = time.Unix(timeSec, 0)
	zone = s[len(s)-1]
	zone = zone[:len(zone)-1] // zone has trailing \n
	return
}

type commit struct {
	hash, tree                string
	parent                    []string
	author, authorEmail       string
	committer, committerEmail string
	date                      time.Time
	zone                      string
	messageReader             io.Reader
	zlibReader                zlib.ReadCloserReset
	messageStr                *string // lazily populated from reader
}

// time.ANSIC with s/_2/2/
const timeFormat = "Mon Jan 2 15:04:05 2006"

func (c commit) String() string {
	s := "commit " + c.hash + "\n"
	s += "Author: " + c.author + " <" + c.authorEmail + ">\n"
	s += "Date:   " + c.date.Format(timeFormat) + " " + c.zone + "\n\n"
	lines := strings.Split(c.message(), "\n")
	for _, l := range lines {
		s += "    " + l + "\n"
	}
	return s
}

func parseKnownFields(c *commit, r io.Reader, size int) error {
	br := bufio.NewReaderSize(r, int(size))
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return err
		}
		switch {
		case line == "\n":
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
			c.parent = append(c.parent, parent)
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

func parseCommitObject(o object) (commit, error) {
	c := commit{}

	err := parseKnownFields(&c, o.reader, int(o.size))
	if err != nil {
		return commit{}, fmt.Errorf("parsing known fields %v", err)
	}

	c.messageReader = o.reader
	c.zlibReader = o.zlibReader
	return c, nil
}

func (c *commit) discardZlibReader() {
	if c.zlibReader != nil {
		returnZlibReader(c.zlibReader)
		c.zlibReader = nil
	}
}

func (c *commit) message() string {
	if c.messageStr == nil {
		b := bytes.NewBuffer(nil) // TODO: size this buffer?
		_, err := io.Copy(b, c.messageReader)
		c.discardZlibReader()
		if err != nil {
			panic(err)
		}
		s := b.String()
		c.messageStr = &s
	}
	return *c.messageStr
}

func readCommit(hash string) (commit, error) {
	object, err := parseObjectFile(hash)
	if err != nil {
		return commit{}, fmt.Errorf("error parsing object %v", err)
	}
	if object.objectType != "commit" {
		return commit{}, fmt.Errorf("object %s has bad type: %s", hash, object.objectType)
	}
	c, err := parseCommitObject(object)
	if err != nil {
		return commit{}, fmt.Errorf("error parsing commit %v", err)
	}
	c.hash = hash
	return c, nil
}

func showCommit(hash string) (string, error) {
	c, err := readCommit(hash)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}
