package sure

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// Encode the sure tree to the given writer.
func (t *Tree) Encode(w io.Writer) error {
	out := bufio.NewWriter(w)
	defer out.Flush()

	_, err := fmt.Fprintf(out, "asure-2.0\n-----\n")
	if err != nil {
		return err
	}
	return t.outWalk(out)
}

func (t *Tree) outWalk(out *bufio.Writer) error {
	_, err := fmt.Fprintf(out, "d%s [%s]\n",
		escapeString(t.Name),
		encodeAtts(t.Atts))
	if err != nil {
		return err
	}

	for _, c := range t.Children {
		err = c.outWalk(out)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(out, "-\n")
	if err != nil {
		return err
	}

	for _, f := range t.Files {
		err = f.outFile(out)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(out, "u\n")
	if err != nil {
		return err
	}

	return nil
}

func (f *File) outFile(out *bufio.Writer) error {
	_, err := fmt.Fprintf(out, "f%s [%s]\n",
		escapeString(f.Name),
		encodeAtts(f.Atts))
	if err != nil {
		return err
	}
	return nil
}

func encodeAtts(atts AttMap) string {
	var buf bytes.Buffer

	// The attributes present must match those set by the local
	// value, since zero is valid for many of these.

	v := reflect.ValueOf(atts)
	v = v.Elem()

	alist := []stringPair{
		{
			key:   "kind",
			value: atts.GetKind(),
		},
	}
	alist = encWalk(v, alist)

	sort.Sort(stringPairSlice(alist))

	for _, p := range alist {
		fmt.Fprintf(&buf, "%s %s ", p.key, p.value)
	}

	return buf.String()
}

func encWalk(v reflect.Value, atts []stringPair) []stringPair {
	t := v.Type()
	nField := t.NumField()

	for i := 0; i < nField; i++ {
		fld := v.Field(i)
		ftyp := t.Field(i)

		// Flatten structs.
		if ftyp.Type.Kind() == reflect.Struct {
			atts = encWalk(fld, atts)
			continue
		}

		name := strings.ToLower(ftyp.Name)

		if name == "kind" {
			continue
		}

		switch v := fld.Interface().(type) {
		case uint32:
			atts = append(atts, stringPair{
				key:   name,
				value: strconv.FormatUint(uint64(v), 10),
			})
		case int64:
			atts = append(atts, stringPair{
				key:   name,
				value: strconv.FormatInt(v, 10),
			})
		case uint64:
			atts = append(atts, stringPair{
				key:   name,
				value: strconv.FormatUint(v, 10),
			})
		case []byte:
			// Parser fails if there are no bytes output.
			// Just skip the attribute.
			if len(v) == 0 {
				continue
			}

			atts = append(atts, stringPair{
				key:   name,
				value: fmt.Sprintf("%x", v),
			})
		case string:
			atts = append(atts, stringPair{
				key:   name,
				value: escapeString(v),
			})
		default:
			log.Printf("%q, type: %v", ftyp.Name, ftyp.Type)
			panic("Unknown type")
		}

	}

	return atts
}

var kindNames = make(map[uint32]string)

func init() {
	kindNames[syscall.S_IFDIR] = "dir"
	kindNames[syscall.S_IFREG] = "file"
	kindNames[syscall.S_IFLNK] = "lnk"
	kindNames[syscall.S_IFIFO] = "fifo"
	kindNames[syscall.S_IFSOCK] = "sock"
	kindNames[syscall.S_IFCHR] = "chr"
	kindNames[syscall.S_IFBLK] = "blk"
}

// A pair of key/value, that can be sorted by key.
type stringPair struct {
	key   string
	value string
}
type stringPairSlice []stringPair

func (p stringPairSlice) Len() int           { return len(p) }
func (p stringPairSlice) Less(i, j int) bool { return p[i].key < p[j].key }
func (p stringPairSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Escape strings in a quoted-printable style (roughly).  UTF-8 bytes
// will always be escaped.
func escapeString(text string) string {
	var buf bytes.Buffer
	for _, ch := range ([]byte)(text) {
		if '!' <= ch && ch <= '~' && ch != '=' && ch != '[' && ch != ']' {
			buf.WriteByte(ch)
		} else {
			fmt.Fprintf(&buf, "=%02x", ch)
		}
	}
	return buf.String()
}
