package main

// Reading data from surefiles

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"os"
	"strings"
)

type sureState struct {
	// Needed at the end to close everything up.
	file  *os.File
	zfile *gzip.Decompressor

	buf *bufio.Reader

	// Each directory visited has a unique number.  Track the one
	// whose iterator is currently valid.
	curDir  int
	lastDir int
}

type dirStateType int

const (
	readingDirs dirStateType = iota
	readingFiles
	readingDone
)

type SureDir struct {
	state *sureState
	info  *Node

	// Number of this directory, must match 'curDir' in the state
	// for this directory to be valid.
	number    int
	parentNum int

	dirState dirStateType
}

func ReadSure(path string) (dir DirWalker, err os.Error) {
	var state sureState

	state.file, err = os.Open(path)
	if err != nil {
		return
	}

	var zread io.Reader
	if strings.HasSuffix(path, ".gz") {
		state.zfile, err = gzip.NewReader(state.file)
		if err != nil {
			state.file.Close()
			return
		}
		zread = state.zfile
	} else {
		zread = state.file
	}

	state.buf = bufio.NewReader(zread)
	dir, err = scanSure(&state)
	return
}

func scanSure(state *sureState) (dir DirWalker, err os.Error) {
	line, err := state.buf.ReadString('\n')
	if err != nil {
		return
	}
	if line != "asure-2.0\n" {
		err = os.NewError("Invalid magic")
		return
	}

	line, err = state.buf.ReadString('\n')
	if err != nil {
		return
	}
	if line != "-----\n" {
		err = os.NewError("Expecting '-----' delimiter line")
		return
	}

	dir, err = sureDir(state)
	return
}

func sureDir(state *sureState) (dir DirWalker, err os.Error) {
	ch, err := state.buf.ReadByte()
	if err != nil {
		// Shouldn't ever get EOF.  If so, we're being
		// misused.
		return
	}
	if ch != 'd' {
		log.Fatal("Expecting 'd' in surefile")
	}

	name, err := readName(state.buf)
	if err != nil {
		return
	}

	atts, err := readAtts(state.buf)
	if err != nil {
		return
	}

	info := &Node{name: name, atts: atts, costly: noCostly}

	// TODO: How do we verify this?
	parent := state.curDir
	state.lastDir++
	state.curDir = state.lastDir
	sdir := &SureDir{state: state, info: info, number: state.lastDir, parentNum: parent}
	dir = sdir
	return
}

func (p *SureDir) Info() *Node  { return p.info }
func (p *SureDir) Path() string { return "TODO" }

func (p *SureDir) Close() (err os.Error) {
	if p.state.zfile != nil {
		err = p.state.zfile.Close()
	}
	err = p.state.file.Close()
	p.state.zfile = nil
	p.state.file = nil
	p.state.buf = nil
	return
}

func (p *SureDir) NextDir() (dir DirWalker, err os.Error) {
	if p.number != p.state.curDir {
		log.Fatalf("incorrect sureread iterator use")
	}
	if p.dirState != readingDirs {
		log.Fatalf("incorrect sure state, not expecing to NextDir()")
	}
	peek, err := p.state.buf.Peek(1)
	if err != nil {
		return
	}
	switch peek[0] {
	case '-':
		p.dirState = readingFiles
		err = mustLine(p.state.buf, "-\n")
		if err != nil {
			return
		}
	case 'd':
		dir, err = sureDir(p.state)
	default:
		log.Fatalf("Unexpected line in surefile")
	}
	return
}

func (p *SureDir) NextNonDir() (node *Node, err os.Error) {
	if p.number != p.state.curDir {
		log.Fatal("incorrect sureread iterator use")
	}
	if p.dirState != readingFiles {
		log.Fatalf("incorrect sure state, not expecint NextNonDir()")
	}
	peek, err := p.state.buf.Peek(1)
	if err != nil {
		return
	}
	switch peek[0] {
	case 'u':
		p.dirState = readingDone
		err = mustLine(p.state.buf, "u\n")
		p.state.curDir = p.parentNum
		return
	case 'f':
		var name string
		var atts map[string]string

		err = mustRead(p.state.buf, 'f')
		if err != nil {
			return
		}

		name, err = readName(p.state.buf)
		if err != nil {
			return
		}

		atts, err = readAtts(p.state.buf)
		if err != nil {
			return
		}

		node = &Node{name: name, atts: atts, costly: noCostly}
	default:
		log.Fatalf("Unexpected line in surefile")
	}
	return
}

func noCostly() map[string]string { return nil }

func mustRead(buf *bufio.Reader, expect byte) (err os.Error) {
	ch, err := buf.ReadByte()
	if err != nil {
		return
	}
	if ch != expect {
		log.Fatalf("Expecting '%c' in surefile, got '%c'", expect, ch)
	}
	return
}

func mustLine(buf *bufio.Reader, expect string) (err os.Error) {
	line, err := buf.ReadString('\n')
	if err != nil {
		return
	}
	if line != expect {
		log.Fatalf("Expecting %+v in surefile, got %+v", expect, line)
	}
	return
}

// Read a space delimited string from the input, reversing the quoted
// printable encoding.
func readName(buf *bufio.Reader) (name string, err os.Error) {
	raw, err := buf.ReadString(' ')
	if err != nil {
		return
	}

	if len(raw) == 0 || raw[len(raw)-1] != ' ' {
		log.Fatal("Error reading name from surefile")
	}
	raw = raw[:len(raw)-1]

	var out bytes.Buffer
	i := 0
	for i < len(raw) {
		if raw[i] == '=' {
			log.Fatal("TODO: handle = decoding")
		} else {
			out.WriteByte(raw[i])
		}

		i++
	}
	name = out.String()
	return
}

// Reads the attributes, including the '[' and ']' characters and the
// terminating line.
func readAtts(buf *bufio.Reader) (atts map[string]string, err os.Error) {
	err = mustRead(buf, '[')
	if err != nil {
		return
	}

	atts = make(map[string]string)
	for {
		p, err := buf.Peek(1)
		if err != nil {
			return
		}
		if p[0] == ']' {
			err = mustRead(buf, ']')
			if err == nil {
				err = mustRead(buf, '\n')
			}
			return
		}

		var key, value string
		key, err = readName(buf)
		if err != nil {
			return
		}
		value, err = readName(buf)
		if err != nil {
			return
		}

		atts[key] = value
	}
	panic("Not reached")
}
