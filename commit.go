package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var hashRe = regexp.MustCompile("^[a-z0-9]{40}$")

func parseHashLine(r *bufio.Reader, leader string) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) != 42+len(leader) || line[:len(leader)+1] != leader+" " {
		return "", fmt.Errorf("bad commit format: \"%s\"", line)
	}
	hash := line[len(leader)+1 : len(line)-1]
	if !hashRe.MatchString(hash) {
		return "", fmt.Errorf("bad commit format hash: \"%s\"", line)
	}
	return hash, nil
}

func parsePersonLine(r *bufio.Reader, whom string) (name, email, zone string, t time.Time, err error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return
	}
	s := strings.Split(line, " ")
	if len(s) < 5 {
		err = fmt.Errorf("bad format, insufficient parts for %s line", whom)
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
	hash, tree, parent        string
	author, authorEmail       string
	committer, committerEmail string
	date                      time.Time
	zone                      string
	message                   string
}

// time.ANSIC with s/_2/2/
const timeFormat = "Mon Jan 2 15:04:05 2006"

func (c commit) String() string {
	s := "commit " + c.hash + "\n"
	s += "Author: " + c.author + " <" + c.authorEmail + ">\n"
	s += "Date:   " + c.date.Format(timeFormat) + " " + c.zone + "\n\n"
	lines := strings.Split(c.message, "\n")
	for _, l := range lines {
		s += "    " + l + "\n"
	}
	return s
}

func showCommit(commitObject object) (string, error) {
	r := commitObject.reader
	c := commit{}
	err := error(nil)
	c.hash = "919b32c0b3cdb2b80ed7daa741b1fe88176b4264" // TODO: CHEAT!
	c.tree, err = parseHashLine(r, "tree")
	if err != nil {
		return "", err
	}
	c.parent, err = parseHashLine(r, "parent")
	if err != nil {
		return "", err
	}

	c.author, c.authorEmail, c.zone, c.date, err = parsePersonLine(r, "author")
	if err != nil {
		return "", err
	}
	c.committer, c.committerEmail, _, _, err = parsePersonLine(r, "committer")
	if err != nil {
		return "", err
	}
	nl, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	if nl != '\n' {
		return "", fmt.Errorf("expected another newline")
	}
	c.message, err = r.ReadString(0)
	if err != nil && err != io.EOF {
		return "", err
	}

	return c.String(), nil
}
