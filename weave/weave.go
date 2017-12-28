// Package weave reads and writes multiple revisions of line-oriented
// data to a weave file.
//
// A weave is a type of delta encoding that stores all of the
// revisions of the file linearily.
//
// The weave sequence consists of 3 control messages interspersed with
// lines of plain data.  The control messages are: Insert to begin
// data associated with a given delta, Delete to indicate a section no
// longer applies as of a given delta, and End to indicate a previous
// Insert or Delete no longer applies.
//
// Each control message has an associated delta number, which begin
// with '1' and are increasing integers.
package weave

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strconv"
)

// A Sink accepts a weave data stream, and processes it in some
// manner.  All return a possible error, which will propagate upward.
type Sink interface {
	// Begin an insert sequence for the given delta.
	Insert(delta int) error
	// Begin a delete sequence for the given delta.
	Delete(delta int) error
	// End a previous Insert or Delete
	End(delta int) error
	// A line of plain text.  `keep` indicates whether or not the
	// text should be included in the Delta specified in the
	// Reader.
	Plain(text string, keep bool) error
}

// Parser scans as input a weave file.  The behavior is controlled by
// some of the fields.
type Parser struct {
	// The source of data for the reader.
	Source *bufio.Reader

	// A possible sync for the data read.  The reader will call
	// into the Sync for each record of the weave file.
	Sink Sink

	// Delta is the desired delta number to retrieve.  This
	// affects the ParseTo call, as well as the 'keep' argument
	// passed to Sink's Plain call.
	Delta int

	dstate    deltaState // States of all open deltas.
	keeping   bool       // Are we currently keeping records?
	lineNo    int
	isPending bool   // Is there a pending line?
	pending   string // The pending line.
}

// NewParser constructs a new reader that reads from the given Reader,
// targeting the given delta.  Will call into sink for each record.
func NewParser(rd io.Reader, sink Sink, delta int) *Parser {
	return &Parser{
		Source: bufio.NewReader(rd),
		Sink:   sink,
		Delta:  delta,

		dstate: make(deltaState, 0),
	}
}

// ParseTo runs the parser until we reach line 'lineno'.  Lines are
// numbered from 1, so calling with a lineno of zero will run the
// parser until the end of the input.  Returns an error if there is
// one, can also return io.EOF to indicate the end of parsing.
func (p *Parser) ParseTo(lineno int) error {
	if p.isPending {
		if err := p.Sink.Plain(p.pending, p.keeping); err != nil {
			return err
		}
		p.isPending = false
		p.pending = "" // Not really needed, but keeps memory a little cleaner.
	}

	for {
		line, err := p.Source.ReadBytes('\n')
		if err != nil {
			// return io.EOF up.
			return err
		}
		line = line[:len(line)-1]

		if len(line) == 0 || line[0] != '\x01' {
			// Textual line.  Count line numbers for the
			// lines we're keeing
			strLine := string(line)
			if p.keeping {
				p.lineNo++
				if p.lineNo == lineno {
					// This is the desired stopping point,
					// return to the caller.
					p.pending = strLine
					p.isPending = true
					return nil
				}
			}
			err = p.Sink.Plain(strLine, p.keeping)
			if err != nil {
				return err
			}
			continue
		}

		// At this point, all should be control lines.  Skip
		// any that are too short (header lines).
		if len(line) < 4 {
			continue
		}

		// Ignore control lines other than the
		// insert/delete/end lines.
		if line[1] != 'I' && line[1] != 'D' && line[1] != 'E' {
			continue
		}

		thisDelta, err := strconv.Atoi(string(line[3:]))
		if err != nil {
			// TODO: This could be more informative
			return err
		}

		switch line[1] {
		case 'E':
			err = p.Sink.End(thisDelta)
			if err != nil {
				return err
			}
			p.pop(thisDelta)
		case 'I':
			err = p.Sink.Insert(thisDelta)
			if err != nil {
				return err
			}

			// Do this insert if this insert is at least
			// as old as the request delta.
			if p.Delta >= thisDelta {
				p.push(thisDelta, stKeep)
			} else {
				p.push(thisDelta, stSkip)
			}
		case 'D':
			err = p.Sink.Delete(thisDelta)
			if err != nil {
				return err
			}

			// Do this delete if this delete is newer than
			// current.  If not, don't account for it.
			if p.Delta >= thisDelta {
				p.push(thisDelta, stSkip)
			} else {
				p.push(thisDelta, stNext)
			}
		}
		p.updateKeep()
	}
}

// As deltas open and close "insert" and "delete" records, we keep
// track of a state.
type oneDelta struct {
	delta int
	mode  stateMode
}

// deltaState is kept sorted with newest (largest) delta at element 0.
type deltaState []oneDelta

// When pushed, this indicates if the delta itself applies.
type stateMode int

const (
	// stKeep means records should be kept.
	stKeep stateMode = iota

	// stSkip means records should be skipped.
	stSkip

	// stNext means this delta marker doesn't apply, and we should
	// look at the next oldest delta.
	stNext
)

// Remove the given numbered state.
func (p *Parser) pop(delta int) {
	found := -1
	for i := range p.dstate {
		if p.dstate[i].delta == delta {
			found = i
			break
		}
	}
	if found == -1 {
		panic("State of pop not present")
	}

	ln := len(p.dstate)
	if found < ln-1 {
		copy(p.dstate[found:ln-1], p.dstate[found+1:ln])
	}
	p.dstate = p.dstate[:ln-1]
}

// Add a new state.  It will be inserted in the proper place in the
// array, based on the delta number.
func (p *Parser) push(delta int, mode stateMode) {
	p.dstate = append(p.dstate, oneDelta{
		mode:  mode,
		delta: delta,
	})

	// Move the node to its proper sort position.
	pos := len(p.dstate) - 1
	for pos > 0 && p.dstate[pos].delta > p.dstate[pos-1].delta {
		p.dstate[pos-1], p.dstate[pos] = p.dstate[pos], p.dstate[pos-1]
		pos--
	}
}

// Update the keep field.
func (p *Parser) updateKeep() {
	for i := range p.dstate {
		switch p.dstate[i].mode {
		case stKeep:
			p.keeping = true
			return
		case stSkip:
			p.keeping = false
			return
		}
	}

	// This shouldn't be reached if there are any more content
	// lines, but we may get here when we reach the end of the
	// input.
	p.keeping = false
}

// Represents a woven file.
type file struct {
	Name string
}

var deltaRe = regexp.MustCompile("^\x01d D ([\\d\\.]+) .* (\\d+) \\d+$")

// ScanRevs retrieves retrieves the mapping from a delta number to an
// SCCS revision.  For files written by SCCS, this can be used to
// retrieve specific deltas using SCCS.
func (s *file) ScanRevs() (map[int]string, error) {
	fd, err := os.Open(s.Name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	bfd := bufio.NewReader(fd)

	revs := make(map[int]string)

	for {
		line, _, err := bfd.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Detect the end of the header early to avoid
		// scanning the entire file.
		if len(line) >= 2 && string(line[:2]) == "\x01I" {
			break
		}

		res := deltaRe.FindStringSubmatch(string(line))
		if res != nil {
			num, err := strconv.Atoi(res[2])
			if err != nil {
				return nil, err
			}
			revs[num] = res[1]
		}
	}

	return revs, nil
}

// There are two basic ways we process a weave file.  In either case,
// we will walk through the data.  The simple case retrieves a
// particular version
