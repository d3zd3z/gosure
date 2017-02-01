package sure

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
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

	allAtts := make(map[string]string)
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
			if err != nil {
				return
			}

			return decodeAtts(allAtts)
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

		allAtts[key] = value
	}
}

// Given string value attributes as a map, decode them into the
// desired structure.
func decodeAtts(allAtts map[string]string) (AttMap, error) {
	kind, ok := allAtts["kind"]
	if !ok {
		return nil, errors.New("Record has no 'kind'")
	}

	typ, ok := kindToType[kind]
	if !ok {
		return nil, errors.New("Unrecognized file kind")
	}

	itemPtr := reflect.New(typ)
	item := itemPtr.Elem()

	err := decWalk(item, allAtts)
	if err != nil {
		return nil, err
	}

	return itemPtr.Interface().(AttMap), nil
}

// Walk a structure, pulling field values out of the attribute map,
// warning if there are fields not taken.
func decWalk(v reflect.Value, atts map[string]string) error {
	t := v.Type()
	nField := t.NumField()

	for i := 0; i < nField; i++ {
		fld := v.Field(i)
		ftyp := t.Field(i)

		// Flatten structs
		if ftyp.Type.Kind() == reflect.Struct {
			err := decWalk(fld, atts)
			if err != nil {
				return err
			}
			continue
		}

		name := strings.ToLower(ftyp.Name)

		// Special case for the kind field.
		if name == "kind" {
			fld.SetUint(uint64(allKinds[atts["kind"]]))
			continue
		}

		// Otherwise, just set the field based on type.
		value, ok := atts[name]
		if !ok {
			warnAtt(name, "decWalk")
			continue
		}
		_ = value

		switch fld.Interface().(type) {
		case uint32:
			v, err := strconv.ParseUint(value, 10, 32)
			if err != nil {
				return err
			}
			fld.SetUint(v)
		case uint64:
			v, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			fld.SetUint(v)
		case int64:
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			fld.SetInt(v)
		case []byte:
			var buf []byte
			_, err := fmt.Sscanf(value, "%x", &buf)
			if err != nil {
				return err
			}
			fld.SetBytes(buf)
		case string:
			// TODO: Need to unescape this
			fld.SetString(value)
		default:
			log.Fatalf("Unknown field type: %v", ftyp)
		}
	}

	return nil
}

// A mapping between kind names and the integer codes for them.
var allKinds = make(map[string]uint32)

func init() {
	allKinds["dir"] = S_IFDIR
	allKinds["file"] = S_IFREG
	allKinds["lnk"] = S_IFLNK
	allKinds["fifo"] = S_IFIFO
	allKinds["sock"] = S_IFSOCK
	allKinds["chr"] = S_IFCHR
	allKinds["blk"] = S_IFBLK
}

// A mapping between the kind names and the structure type used to
// build it.
var kindToType = make(map[string]reflect.Type)

func init() {
	kindToType["dir"] = reflect.TypeOf(DirAtts{})
	kindToType["file"] = reflect.TypeOf(RegAtts{})
	kindToType["lnk"] = reflect.TypeOf(LinkAtts{})
	kindToType["fifo"] = reflect.TypeOf(FifoAtts{})
	kindToType["sock"] = reflect.TypeOf(FifoAtts{})
	kindToType["chr"] = reflect.TypeOf(DevAtts{})
	kindToType["blk"] = reflect.TypeOf(DevAtts{})
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
