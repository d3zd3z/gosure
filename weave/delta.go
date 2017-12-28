package weave

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// A DeltaWriter is used to write a new version to a weave file.
type DeltaWriter struct {
	file *os.File
	wr   *bufio.Writer
	nc   NamingConvention
	base int
	name string
	tags map[string]string
}

// NewDeltaWriter create a new DeltaWriter.  The contents should be
// written to the DeltaWriter, and the 'Close' method called.  Note
// that most of the work is done as part of the Close method.  'base'
// should be set to the delta number that this change is based on.
// This will typically be the last delta written.  Note that the tags
// should not be changed until after Close is called.
func NewDeltaWriter(nc NamingConvention, base int, name string, tags map[string]string) (*DeltaWriter, error) {
	file, err := TempFile(nc, false)
	if err != nil {
		return nil, err
	}

	wr := bufio.NewWriter(file)

	return &DeltaWriter{
		file: file,
		wr:   wr,
		nc:   nc,
		base: base,
		name: name,
		tags: tags,
	}, nil
}

func (w *DeltaWriter) Write(p []byte) (n int, err error) {
	return w.wr.Write(p)
}

// Close closes the delta writer.  This causes the delta to actually
// be generated, so it is important to check the error status from
// this method.
func (w *DeltaWriter) Close() error {
	err := w.wr.Flush()
	if err != nil {
		return err
	}
	err = w.file.Close()
	if err != nil {
		return err
	}

	priorName, err := w.getPrior()
	if err != nil {
		return err
	}

	diffs, err := w.getDiff(priorName, w.file.Name())
	if err != nil {
		return err
	}

	newName, _, err := w.applyDiff(diffs)
	if err != nil {
		return err
	}
	// fmt.Printf("new delta: %d\n", newDelta)

	os.Rename(w.nc.MainFile(), w.nc.BackupFile())
	err = os.Rename(newName, w.nc.MainFile())
	if err != nil {
		return err
	}

	os.Remove(priorName)
	os.Remove(w.file.Name())

	return nil
}

// getPrior reads the base delta into a new temporary file.  Returns
// the name of the new temporary file
func (w *DeltaWriter) getPrior() (string, error) {
	file, err := TempFile(w.nc, false)
	if err != nil {
		return "", err
	}
	defer file.Close()
	wr := bufio.NewWriter(file)

	inpfile, err := os.Open(w.nc.MainFile())
	if err != nil {
		return "", err
	}
	defer inpfile.Close()

	gz, err := gzip.NewReader(inpfile)
	if err != nil {
		return "", err
	}

	err = NewParser(gz, NewWriteDelta(wr), w.base).ParseTo(0)
	if err != io.EOF {
		if err == nil {
			panic("Unexpected flow")
		}
		return "", err
	}
	// fmt.Printf("Flushing writer for %q\n", file.Name())
	err = wr.Flush()
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

// getDiff Runs diff on the two filenames, returning it as lines.
// TODO: This should stream the diff result instead of reading it all
// into memory.
func (w *DeltaWriter) getDiff(a, b string) ([]string, error) {
	diffBuf, err := exec.Command("diff", a, b).Output()
	if err != nil {
		state, ok := err.(*exec.ExitError)
		if !ok {
			panic("Guessed wrong type of exit code")
		}
		status, ok := state.Sys().(syscall.WaitStatus)
		if !ok {
			panic("return code is not helpful")
		}
		switch status {
		case 0, 256:
		default:
			return nil, fmt.Errorf("Unable to run diff: %s", err)
		}
	}
	lines := strings.Split(string(diffBuf), "\n")
	lines = lines[:len(lines)-1]

	return lines, nil
}

// Regexp for parsing what we care about in diff lines.  There are one
// or two numbers (comma separated), one of the letters a, c or d, and
// then stuff describing the destination file, which really doesn't
// matter to us.  The regexp result [1] is the range start [3] is the
// optional range end, and [4] is the command.
var commandRe = regexp.MustCompile(`^(\d+)(,(\d+))?([acd]).*$`)

// applyDelta uses the output of diff to generate a new weave file
// with the new delta as an additional revision.  Returns the new
// delta number.
func (w *DeltaWriter) applyDiff(diff []string) (string, int, error) {
	file, rd, err := weaveOpen(w.nc)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	bufrd := bufio.NewReader(rd)

	hdr, err := LoadHeader(bufrd)
	if err != nil {
		return "", 0, err
	}

	newDelta := hdr.AddDelta(w.name, w.tags)

	wfile, wr, err := weaveCreate(w.nc, hdr)
	if err != nil {
		return "", 0, err
	}
	defer wfile.Close()
	defer wr.Close() // Calls flush.

	weaveWr := Writer{wr}

	parser := NewParser(bufrd, weaveWr, w.base)

	isDone := false
	isAdding := false

	// Go through the diff, and apply
	for _, diffLine := range diff {
		m := commandRe.FindStringSubmatch(diffLine)
		// fmt.Printf("diff: %q, %#v\n", diffLine, m)
		if m != nil {
			if isAdding {
				err = weaveWr.End(newDelta)
				if err != nil {
					return "", 0, err
				}
				isAdding = false
			}

			left, err := strconv.Atoi(m[1])
			if err != nil {
				panic(err)
			}

			right := left
			if m[3] != "" {
				right, err = strconv.Atoi(m[3])
				if err != nil {
					panic(err)
				}
			}

			cmd := m[4][0]
			if cmd == 'd' || cmd == 'c' {
				// These include deletions.
				err = parser.ParseTo(left)
				// fmt.Printf("p1: %s\n", err)
				if err != nil {
					return "", 0, err
				}
				err = weaveWr.Delete(newDelta)
				if err != nil {
					return "", 0, err
				}
				err = parser.ParseTo(right + 1)
				// fmt.Printf("p2: %s\n", err)
				if err == io.EOF {
					isDone = true
				} else if err != nil {
					return "", 0, err
				}
				err = weaveWr.End(newDelta)
				if err != nil {
					return "", 0, err
				}
			} else {
				err = parser.ParseTo(right + 1)
				// fmt.Printf("p3: %s\n", err)
				if err == io.EOF {
					isDone = true
				} else if err != nil {
					return "", 0, err
				}
			}

			if cmd == 'c' || cmd == 'a' {
				err = weaveWr.Insert(newDelta)
				if err != nil {
					return "", 0, err
				}
				isAdding = true
			}

			continue
		}

		if diffLine[0] == '<' || diffLine[0] == '-' {
			continue
		}

		if diffLine[0] == '>' {
			// Add lines should just be written as-is.
			err = weaveWr.Plain(diffLine[2:], true)
			if err != nil {
				return "", 0, err
			}

			continue
		}

		panic(fmt.Sprintf("Unexpected line from diff command: %q", diffLine))
	}

	if isAdding {
		err = weaveWr.End(newDelta)
		if err != nil {
			return "", 0, err
		}
	}

	if !isDone {
		err = parser.ParseTo(0)
		// fmt.Printf("p4: %s\n", err)
		if err == nil {
			panic("Unexpected non-eof")
		}
		if err != io.EOF {
			return "", 0, err
		}
	}

	return wfile.Name(), newDelta, nil
}
