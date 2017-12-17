package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"davidb.org/x/gosure/weave"
)

var deltaLines int
var deltaDeltas int
var weaveFile string = "w.delta"

var writeDebug = false

func delta() {
	if _, err := os.Lstat(playFile); err == nil {
		log.Fatalf("plain file %q is present, remove and rerun", playFile)
	}

	// Clean up prior runs.
	os.Remove(sccsFile)

	data := make([]int, deltaLines)

	for i := range data {
		data[i] = i + 1
	}

	// Generate initial SCCS file.
	ss := NewSccs(playFile)
	if err := ss.Initial(data); err != nil {
		log.Fatal(err)
	}

	// Generate initial weave file.
	if err := writeWeave(data); err != nil {
		log.Fatal(err)
	}

	if err := writeDebugDelta(data, 1); err != nil {
		log.Fatal(err)
	}

	// Make sure they are the same.
	if err := compareWeaves(sccsFile, weaveFile); err != nil {
		log.Fatal(err)
	}

	for i := 1; i < deltaDeltas; i++ {

		log.Printf("Applying delta %d", i+1)

		// Shuffle a subset of the data for the delta.
		a := rand.Intn(len(data))
		b := rand.Intn(len(data))
		if a > b {
			a, b = b, a
		}
		shuffle(data[a : b+1])

		// Update the SCCS file.
		sfile := Sccs{
			name: playFile,
		}
		if err := sfile.WriteDelta(data); err != nil {
			panic(err)
			log.Fatal(err)
		}

		diff, err := getDiff(i, data)
		if err != nil {
			log.Fatal(err)
		}

		err = applyDelta(diff, weaveFile, weaveFile+".tmp", i, i+1)
		if err != nil {
			panic(err)
		}
		if err := os.Rename(weaveFile+".tmp", weaveFile); err != nil {
			panic(err)
		}

		if err := writeDebugDelta(data, i+1); err != nil {
			panic(err)
		}

		// Make sure they are the same.
		if err := compareWeaves(sccsFile, weaveFile); err != nil {
			log.Fatal(err)
		}
	}

}

// Emit a given delta from the sccs file, and generate a diff based on
// newly provided data.
func getDiff(delta int, data []int) ([]string, error) {
	// Write the weave delta information out.
	oldName, err := weaveTo(weaveFile, delta)
	if err != nil {
		return nil, err
	}
	defer os.Remove(oldName)

	// Write out the new contents.
	newName, err := writePlainTo(data)
	if err != nil {
		return nil, err
	}
	defer os.Remove(newName)

	// Run 'diff' on these two files, and capture its output.
	// Diff has annoying output, where 0 means no diffs, 1 means
	// there were diffs, and 2 means "trouble".
	diffBuf, err := exec.Command("diff", oldName, newName).Output()
	if err != nil {
		state, ok := err.(*exec.ExitError)
		if !ok {
			log.Fatal("Guessed type wrong")
		}
		status, ok := state.Sys().(syscall.WaitStatus)
		if !ok {
			log.Fatal("return code is not helpful")
		}
		switch status {
		case 0, 256:
		default:
			log.Fatal("Unable to run diff: %v", err)
		}
	}
	lines := strings.Split(string(diffBuf), "\n")
	lines = lines[:len(lines)-1]
	// log.Printf("diff: %#v", lines)

	return lines, nil
}

// Regexp for parsing what we care about in diff lines.  There are one
// or two numbers (comma separated), one of the letters a, c or d, and
// then stuff describing the destination file, which doesn't really
// matter to us.  The regexp result [1] is range start [3] is the
// optional range end, and [4] is the command.
var commandRe = regexp.MustCompile(`^(\d+)(,(\d+))?([acd]).*$`)

// Copy a weave file, adding a new delta to it based on the contents
// of a diff.
func applyDelta(diff []string, srcWeave, destWeave string, base, newDelta int) error {
	rfd, err := os.Open(srcWeave)
	if err != nil {
		return err
	}
	defer rfd.Close()

	// TODO: This is a security problem, this shouldn't allow
	// writes to an existing name.
	wfd, err := os.Create(destWeave)
	if err != nil {
		return err
	}
	wbuf := bufio.NewWriter(wfd)
	defer wbuf.Flush()

	wr := weave.NewWriter(wbuf)
	parser := weave.NewParser(rfd, wr, base)
	_ = parser

	isDone := false
	isAdding := false

	// Go through the diff, and apply
	for _, diffLine := range diff {
		m := commandRe.FindStringSubmatch(diffLine)
		// log.Printf("diff: %q, %#v", diffLine, m)
		if m != nil {
			if isAdding {
				err = wr.End(newDelta)
				if err != nil {
					return err
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
				if err != nil {
					return err
				}
				err = wr.Delete(newDelta)
				if err != nil {
					return err
				}
				err = parser.ParseTo(right + 1)
				if err == io.EOF {
					isDone = true
				} else if err != nil {
					return err
				}
				err = wr.End(newDelta)
				if err != nil {
					return err
				}
			} else {
				err = parser.ParseTo(right + 1)
				if err == io.EOF {
					isDone = true
				} else if err != nil {
					return err
				}
			}

			if cmd == 'c' || cmd == 'a' {
				err = wr.Insert(newDelta)
				if err != nil {
					return err
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
			err = wr.Plain(diffLine[2:], true)
			if err != nil {
				return err
			}

			continue
		}

		panic(fmt.Sprintf("Unexpected line from diff command: %q", diffLine))
	}

	if isAdding {
		err = wr.End(newDelta)
		if err != nil {
			return err
		}
	}

	if !isDone {
		err = parser.ParseTo(0)
		if err == nil {
			return errors.New("Unexpected non-eof")
		}
		if err != io.EOF {
			return err
		}
	}

	return nil
}

// Write an initial weave file, writing the data as the first delta.
func writeWeave(data []int) error {
	f, err := os.Create(weaveFile)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewWriter(f)
	defer buf.Flush()

	wr := weave.NewWriter(buf)

	if err = wr.Insert(1); err != nil {
		return err
	}

	for _, x := range data {
		if err = wr.Plain(fmt.Sprintf("%d", x), true); err != nil {
			return nil
		}
	}

	if err = wr.End(1); err != nil {
		return err
	}

	return nil
}

// Output a given delta from the weave file to a temporary file, and
// return the name of that file.
func weaveTo(weaveName string, delta int) (string, error) {
	rfd, err := os.Open(weaveName)
	if err != nil {
		return "", err
	}
	defer rfd.Close()

	wfd, err := ioutil.TempFile("", "weave")
	if err != nil {
		return "", err
	}
	defer wfd.Close()

	name := wfd.Name()
	buf := bufio.NewWriter(wfd)
	defer buf.Flush()

	wr := weave.NewDeltaWriter(buf)
	err = weave.NewParser(rfd, wr, delta).ParseTo(0)
	if err != io.EOF {
		wfd.Close()
		os.Remove(name)
		return "", err
	}

	return name, nil
}

// Generate a plaintext file (in tmp) holding the new data.
func writePlainTo(data []int) (string, error) {
	wfd, err := ioutil.TempFile("", "weave")
	if err != nil {
		return "", nil
	}
	defer wfd.Close()

	buf := bufio.NewWriter(wfd)
	defer buf.Flush()
	for _, x := range data {
		fmt.Fprintf(buf, "%d\n", x)
	}

	return wfd.Name(), nil
}

// For debugging, write the plaintext a file corresponding to its
// delta.
func writeDebugDelta(data []int, delta int) error {
	if !writeDebug {
		return nil
	}

	name := fmt.Sprintf("debug-%03d", delta)
	wfd, err := os.Create(name)
	if err != nil {
		return err
	}
	defer wfd.Close()

	buf := bufio.NewWriter(wfd)
	defer buf.Flush()
	for _, x := range data {
		_, err = fmt.Fprintf(buf, "%d\n", x)
		if err != nil {
			return err
		}
	}

	return nil
}

func compareWeaves(a, b string) error {
	af, err := os.Open(a)
	if err != nil {
		return err
	}
	defer af.Close()

	bf, err := os.Open(b)
	if err != nil {
		return err
	}
	defer bf.Close()

	achan := make(chan controlItem, 0)
	bchan := make(chan controlItem, 0)

	aparse := weave.NewParser(af, chanWriter(achan), 0)
	bparse := weave.NewParser(bf, chanWriter(bchan), 0)

	go func() {
		err := aparse.ParseTo(0)
		if err != io.EOF {
			log.Fatal(err, " achan")
		}
		close(achan)
	}()

	go func() {
		err := bparse.ParseTo(0)
		if err != io.EOF {
			log.Fatal(err, " bchan")
		}
		close(bchan)
	}()

	for {
		aitem, aok := <-achan
		bitem, bok := <-bchan

		// Both are finished.
		if !aok && !bok {
			break
		}

		if aok != bok {
			return errors.New("Early EOF")
		}

		if aitem.kind != bitem.kind || aitem.delta != bitem.delta || aitem.text != bitem.text {
			return errors.New("Item mismatch")
		}
	}

	return nil
}

// A ChanWriter writes the encoded data to a channel.
type controlType int

const (
	controlInsert controlType = iota
	controlDelete
	controlEnd
	controlPlain
)

// Either delta or text should be set to the zero value.
type controlItem struct {
	kind  controlType
	delta int
	text  string
}

type chanWriter chan<- controlItem

func (w chanWriter) Insert(delta int) error {
	w <- controlItem{
		kind:  controlInsert,
		delta: delta,
	}
	return nil
}

func (w chanWriter) Delete(delta int) error {
	w <- controlItem{
		kind:  controlDelete,
		delta: delta,
	}
	return nil
}

func (w chanWriter) End(delta int) error {
	w <- controlItem{
		kind:  controlEnd,
		delta: delta,
	}
	return nil
}

func (w chanWriter) Plain(text string, keep bool) error {
	w <- controlItem{
		kind: controlPlain,
		text: text,
	}
	return nil
}
