package sure

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
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

	pairs := make([]stringPair, 0, len(atts))

	for k, v := range atts {
		pairs = append(pairs, stringPair{k, v})
	}
	sort.Sort(stringPairSlice(pairs))

	for _, p := range pairs {
		buf.WriteString(p.key)
		buf.WriteRune(' ')
		buf.WriteString(escapeString(p.value))
		buf.WriteRune(' ')
	}

	return buf.String()
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
