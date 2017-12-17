package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"

	"davidb.org/x/gosure/weave"
)

type SccsFile struct {
	name string

	revs map[int]string
}

// Load and populate 'revs' from the header.
func (s *SccsFile) ScanRevs() error {
	wv := weave.File{
		Name: s.name,
	}

	var err error
	s.revs, err = wv.ScanRevs()

	return err
}

// Get the text of a revision using SCCS.
func (s *SccsFile) GetRevSccs(delta int) ([]string, error) {
	rev, ok := s.revs[delta]
	if !ok {
		return nil, errors.New("Invalid delta")
	}
	out, err := exec.Command("sccs", "get", "-ksp", "-r"+rev, s.name).Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")

	// The last line will be blank.
	lines = lines[:len(lines)-1]

	return lines, nil
}

// Retrieve the delta from an sccs file directly.
func (s *SccsFile) GetRev(delta int) ([]string, error) {
	// fmt.Printf("GetRev: %d\n", delta)
	fd, err := os.Open(s.name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var sink captureSink
	wv := weave.NewParser(fd, &sink, delta)
	err = wv.ParseTo(0)
	if err == nil {
		return nil, err
	}
	return sink, nil
}

// The captureSink is just an array of the lines of the delta.
type captureSink []string

func (s *captureSink) Insert(delta int) error { return nil }
func (s *captureSink) Delete(delta int) error { return nil }
func (s *captureSink) End(delta int) error    { return nil }

func (s *captureSink) Plain(text string, keep bool) error {
	if keep {
		*s = append(*s, text)
	}
	return nil
}

func scan() error {
	sf := SccsFile{
		name: sccsFile,
	}

	if err := sf.ScanRevs(); err != nil {
		return err
	}

	// Get all of the deltas in order:
	deltas := make([]int, 0, len(sf.revs))
	for delta := range sf.revs {
		deltas = append(deltas, delta)
	}
	sort.Sort(sort.IntSlice(deltas))

	for _, delta := range deltas {
		expected, err := sf.GetRevSccs(delta)
		if err != nil {
			return err
		}

		gotten, err := sf.GetRev(delta)
		if err != nil {
			return err
		}

		if !reflect.DeepEqual(gotten, expected) {
			fmt.Printf("Mismatch on %d:\ngotten: %v\nexpect: %v\n", delta, gotten, expected)
			break
		}
	}

	return nil
}
