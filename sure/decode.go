package sure

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"syscall"
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

	var tmpu uint64
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

		switch key {
		case "kind":
			val, ok := allKinds[value]
			if !ok {
				warnAtt(value, "Unknown file kind")
			} else {
				atts.Kind = val
			}
		case "uid":
			tmpu, err = strconv.ParseUint(value, 10, 32)
			if err != nil {
				return
			}
			atts.Uid = uint32(tmpu)
		case "gid":
			tmpu, err = strconv.ParseUint(value, 10, 32)
			if err != nil {
				return
			}
			atts.Gid = uint32(tmpu)
		case "perm":
			tmpu, err = strconv.ParseUint(value, 10, 32)
			if err != nil {
				return
			}
			atts.Perm = uint32(tmpu)
		case "devmaj":
			tmpu, err = strconv.ParseUint(value, 10, 32)
			if err != nil {
				return
			}
			atts.Devmaj = uint32(tmpu)
		case "devmin":
			tmpu, err = strconv.ParseUint(value, 10, 32)
			if err != nil {
				return
			}
			atts.Devmin = uint32(tmpu)
		case "mtime":
			atts.Mtime, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				return
			}
		case "ctime":
			atts.Ctime, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				return
			}
		case "ino":
			atts.Ino, err = strconv.ParseUint(value, 10, 64)
			if err != nil {
				return
			}
		case "size":
			atts.Size, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				return
			}
		case "sha1":
			_, err = fmt.Sscanf(value, "%x", &atts.Sha)
			if err != nil {
				return
			}
		case "targ":
			atts.Targ = value
		default:
			warnAtt(key, "Unknown attribute")
		}
	}
}

var allKinds = make(map[string]uint32)

func init() {
	allKinds["dir"] = syscall.S_IFDIR
	allKinds["file"] = syscall.S_IFREG
	allKinds["lnk"] = syscall.S_IFLNK
	allKinds["fifo"] = syscall.S_IFIFO
	allKinds["sock"] = syscall.S_IFSOCK
	allKinds["chr"] = syscall.S_IFCHR
	allKinds["blk"] = syscall.S_IFBLK
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
