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
	"syscall"
)

// Decode loads a surefile from an io.Reader.
func Decode(r io.Reader) (*Tree, error) {
	rd := bufio.NewReader(r)
	pd := NewPushDecoder()
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Trim the newline.
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		err = pd.add(line)
		if err != nil {
			return nil, err
		}
	}

	if pd.result == nil || len(pd.tree) != 0 {
		return nil, fmt.Errorf("Invalid ending state")
	}

	return pd.result, nil
}

// A PushDecoder is a stateful decoder that can be given a surefile
// one line at a time, and generate the full tree.
type PushDecoder struct {
	// The method
	add func(line string) error

	// The tree as we are building it.
	tree   []*Tree
	result *Tree
}

func NewPushDecoder() *PushDecoder {
	var pd PushDecoder
	pd.add = pd.needAsure

	return &pd
}

// Add adds a line to the push decoder.
func (pd *PushDecoder) Add(line string) error {
	return pd.add(line)
}

// Tree extracts the tree from the push decoder.  Returns an error if
// a full surefile wasn't pushed to it.
func (pd *PushDecoder) Tree() (*Tree, error) {
	if pd.result == nil || len(pd.tree) != 0 {
		return nil, fmt.Errorf("Invalid ending state")
	}

	return pd.result, nil
}

func (pd *PushDecoder) needAsure(line string) error {
	if line != "asure-2.0" {
		return errors.New("Invalid Magic")
	}
	pd.add = pd.needHyphens
	return nil
}

func (pd *PushDecoder) needHyphens(line string) error {
	if line != "-----" {
		return errors.New("Invalid Magic")
	}
	pd.add = pd.addDirs
	return nil
}

func (pd *PushDecoder) addDirs(line string) error {
	if len(line) == 0 {
		return errors.New("Invalid blank line")
	}
	if line[0] == 'd' {
		// New subdirectory.
		var newTree Tree
		err := parseNameAtts(line[1:], &newTree.Name, &newTree.Atts)
		if err != nil {
			return err
		}
		pd.tree = append(pd.tree, &newTree)

		// Still processing a directory
		return nil
	}

	if line[0] == '-' {
		// Done with this directory, add files to it.
		pd.add = pd.addFiles
		return nil
	}

	return fmt.Errorf("Unexpected line in directory state: %q", line)
}

func (pd *PushDecoder) addFiles(line string) error {
	if len(line) == 0 {
		return errors.New("Invalid blank line")
	}

	if line[0] == 'f' {
		// New File.
		tree := pd.tree[len(pd.tree)-1]

		var newFile File
		err := parseNameAtts(line[1:], &newFile.Name, &newFile.Atts)
		if err != nil {
			return err
		}
		tree.Files = append(tree.Files, &newFile)
		return nil
	}

	if line[0] == 'u' {
		tree := pd.tree[len(pd.tree)-1]
		pd.tree = pd.tree[:len(pd.tree)-1]

		if len(pd.tree) == 0 {
			// Last directory.
			pd.add = pd.nothing
			pd.result = tree
		} else {
			// Still inside a directory
			pd.add = pd.addDirs
			tf := pd.tree[len(pd.tree)-1]
			tf.Children = append(tf.Children, tree)
		}
		return nil
	}

	return fmt.Errorf("Unexpected line in file state: %q", line)
}

func (pd *PushDecoder) nothing(line string) error {
	return fmt.Errorf("Unexpected line in eof state: %q", line)
}

// parseNameAtts parses the name and attributes from the encoded line,
// and sets them.
func parseNameAtts(line string, name *string, atts *AttMap) error {
	i := 0
	tname, err := scanName(line, &i)
	if err != nil {
		return err
	}

	*name = tname

	if i >= len(line) || line[i] != '[' {
		return fmt.Errorf("Expecting '['")
	}
	i++

	allAtts := make(map[string]string)
	for {
		if i >= len(line) {
			return SyntaxError
		}
		if line[i] == ']' {
			at, err := decodeAtts(allAtts)
			if err == nil {
				*atts = at
			}
			return err
		}

		key, err := scanName(line, &i)
		if err != nil {
			return err
		}
		value, err := scanName(line, &i)
		if err != nil {
			return err
		}

		allAtts[key] = value
	}
}

func scanName(line string, pos *int) (string, error) {
	var out bytes.Buffer
	i := *pos

	for i < len(line) && line[i] != ' ' {
		if line[i] == '=' {
			if i+3 > len(line) {
				return "", fmt.Errorf("Encoded number truncated")
			}
			tmp, err := strconv.ParseUint(line[i+1:i+3], 16, 8)
			if err != nil {
				return "", fmt.Errorf("Invalid hex number in '='")
			}
			out.WriteByte(byte(tmp))
			i += 2
		} else {
			out.WriteByte(line[i])
		}
		i++
	}
	i++

	*pos = i
	return out.String(), nil
}

var SyntaxError = errors.New("Syntax error in surefile")

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

	delete(allAtts, "kind")

	for k := range allAtts {
		warnAtt(k, "decode")
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

		delete(atts, name)

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
var allKinds = map[string]uint32{
	"dir":  syscall.S_IFDIR,
	"file": syscall.S_IFREG,
	"lnk":  syscall.S_IFLNK,
	"fifo": syscall.S_IFIFO,
	"sock": syscall.S_IFSOCK,
	"chr":  syscall.S_IFCHR,
	"blk":  syscall.S_IFBLK,
}

// A mapping between the kind names and the structure type used to
// build it.
var kindToType = map[string]reflect.Type{
	"dir":  reflect.TypeOf((*DirAtts)(nil)).Elem(),
	"file": reflect.TypeOf((*RegAtts)(nil)).Elem(),
	"lnk":  reflect.TypeOf((*LinkAtts)(nil)).Elem(),
	"fifo": reflect.TypeOf((*FifoAtts)(nil)).Elem(),
	"sock": reflect.TypeOf((*FifoAtts)(nil)).Elem(),
	"chr":  reflect.TypeOf((*DevAtts)(nil)).Elem(),
	"blk":  reflect.TypeOf((*DevAtts)(nil)).Elem(),
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
