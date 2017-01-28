package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type SccsFile struct {
	name string

	revs map[int]string
}

var deltaRe *regexp.Regexp = regexp.MustCompile("^\x01d D ([\\d\\.]+) .* (\\d+) \\d+$")

// Load and populate 'revs' from the header.
func (s *SccsFile) ScanRevs() error {
	fd, err := os.Open(s.name)
	if err != nil {
		return err
	}
	defer fd.Close()

	bfd := bufio.NewReader(fd)

	s.revs = make(map[int]string)

	for {
		line, _, err := bfd.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		res := deltaRe.FindStringSubmatch(string(line))
		if res != nil {
			num, err := strconv.Atoi(res[2])
			if err != nil {
				return err
			}
			s.revs[num] = res[1]
		}
	}

	return nil
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

// The state of a given request.
type stateMode int
type state struct {
	mode  stateMode
	delta int
}
type states struct{ state []state }

const (
	stKeep = iota
	stSkip
	stNext
)

// Remove the given numbered state.
func (st *states) pop(delta int) {
	found := -1
	for i := range st.state {
		if st.state[i].delta == delta {
			found = i
		}
	}
	if found == -1 {
		panic("State pop of not present")
	}

	ln := len(st.state)
	if found < ln-1 {
		copy(st.state[found:ln-1], st.state[found+1:ln])
	}
	st.state = st.state[:ln-1]
}

func (st *states) push(delta int, mode stateMode) {
	st.state = append(st.state, state{
		mode:  mode,
		delta: delta,
	})
}

// Are we currently keeping?
func (st *states) isKeep() bool {
	for i := len(st.state) - 1; i >= 0; i-- {
		switch st.state[i].mode {
		case stKeep:
			return true
		case stSkip:
			return false
		}
	}
	return false
}

// Retrieve the delta from an sccs file directly.
func (s *SccsFile) GetRev(delta int) ([]string, error) {
	fd, err := os.Open(s.name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	bfd := bufio.NewReader(fd)

	state := states{state: make([]state, 0, 10)}
	result := make([]string, 0)

	for {
		line, _, err := bfd.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(line) == 0 || line[0] != '\x01' {
			// Textual line.  Process if we should be.
			// fmt.Printf("line: %q, state: %v, keep: %b\n", line, state, state.isKeep())
			if state.isKeep() {
				result = append(result, string(line))
			}
			continue
		}

		// At this point, all should be control lines.  Skip
		// any that are too short.
		if len(line) < 4 {
			continue
		}

		// Ignore control lines other than the insert/delete
		// lines.
		if line[1] != 'I' && line[1] != 'D' && line[1] != 'E' {
			continue
		}

		dl, err := strconv.Atoi(string(line[3:]))
		if err != nil {
			// TODO: This could be more informative.
			return nil, err
		}

		// fmt.Printf("Control: '%c', %d\n", line[1], dl)

		if line[1] == 'E' {
			state.pop(dl)
			// fmt.Printf("state: %+v\n", state)
			continue
		}

		switch line[1] {
		case 'I':
			// Do the insert if this insert is older than
			// the requested delta.
			if delta >= dl {
				state.push(dl, stKeep)
			} else {
				state.push(dl, stSkip)
			}
		case 'D':
			// Do the delete if this delete is newer than
			// the current.  If not, don't account for it.
			if delta >= dl {
				state.push(dl, stSkip)
			} else {
				state.push(dl, stNext)
			}
		}
		// fmt.Printf("state: %+v\n", state)
	}

	return result, nil
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
