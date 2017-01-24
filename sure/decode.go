package sure

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"strconv"
)

// Load a surefile from an external file.
func Decode(r io.Reader) (*Tree, error) {
	buf := bufio.NewReader(r)

	line, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if line != "asure-2.0\n" {
		return nil, errors.New("Invalid magic")
	}

	line, err = buf.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if line != "-----\n" {
		return nil, errors.New("Expecting '-----' delimiter line")
	}

	return readTree(buf)
}

func readTree(buf *bufio.Reader) (tree *Tree, err error) {
	ch, err := buf.ReadByte()
	if err != nil {
		// Shouldn't ever get an EOF.
		return
	}
	if ch != 'd' {
		log.Fatalf("Expecting 'd' in surefile (got '%c')", ch)
	}

	name, err := readName(buf)
	if err != nil {
		return
	}

	atts, err := readAtts(buf)
	if err != nil {
		return
	}

	tree = &Tree{
		Name: name,
		Atts: atts,
	}

	// Read children until we get the '-' delimiter.
	for {
		var chs []byte
		chs, err = buf.Peek(1)
		if err != nil {
			tree = nil
			return
		}
		if chs[0] == '-' {
			break
		}

		var child *Tree
		child, err = readTree(buf)
		if err != nil {
			tree = nil
			return
		}

		tree.Children = append(tree.Children, child)
	}

	line, err := buf.ReadString('\n')
	if err != nil {
		tree = nil
		return
	}
	if line != "-\n" {
		tree = nil
		err = errors.New("Expecting '-\\n' line")
		return
	}

	// Read the files.
	for {
		var chs []byte
		chs, err = buf.Peek(1)
		if err != nil {
			tree = nil
			return
		}
		if chs[0] == 'u' {
			break
		}

		var file *File
		file, err = readFile(buf)
		if err != nil {
			tree = nil
			return
		}

		tree.Files = append(tree.Files, file)
	}

	line, err = buf.ReadString('\n')
	if err != nil {
		tree = nil
		return
	}
	if line != "u\n" {
		tree = nil
		err = errors.New("Expecting 'u\\n' line")
		return
	}

	return
}

func readFile(buf *bufio.Reader) (file *File, err error) {
	err = mustRead(buf, 'f')
	if err != nil {
		return
	}

	name, err := readName(buf)
	if err != nil {
		return
	}

	atts, err := readAtts(buf)
	if err != nil {
		return
	}

	file = &File{
		Name: name,
		Atts: atts,
	}
	return
}

// Read a space delimited string from the input, reversing the quoted
// printable encoding.
func readName(buf *bufio.Reader) (name string, err error) {
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
			if i+3 > len(raw) {
				log.Fatal("Encoded number beyond range")
			}
			var tmp uint64
			tmp, err = strconv.ParseUint(raw[i+1:i+3], 16, 64)
			if err != nil {
				log.Fatal("Unable to decode hext number in '='")
			}
			out.WriteByte(byte(tmp))
			i += 2
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
func readAtts(buf *bufio.Reader) (atts AttMap, err error) {
	err = mustRead(buf, '[')
	if err != nil {
		return
	}

	atts = make(AttMap)
	for {
		var p []byte
		p, err = buf.Peek(1)
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
}

func mustRead(buf *bufio.Reader, expect byte) (err error) {
	ch, err := buf.ReadByte()
	if err != nil {
		return
	}
	if ch != expect {
		log.Fatalf("Expecting '%c' in surefile, got '%c'", expect, ch)
	}
	return
}
