package sure

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

	isReg := atts.Kind == syscall.S_IFREG
	hasTime := atts.Kind == syscall.S_IFREG
	hasDev := atts.Kind == syscall.S_IFCHR || atts.Kind == syscall.S_IFBLK
	isLink := atts.Kind == syscall.S_IFLNK

	if hasTime {
		fmt.Fprintf(&buf, "ctime %d ", atts.Ctime)
	}
	if hasDev {
		fmt.Fprintf(&buf, "devmaj %d ", atts.Devmaj)
		fmt.Fprintf(&buf, "devmin %d ", atts.Devmin)
	}
	if !isLink {
		fmt.Fprintf(&buf, "gid %d ", atts.Gid)
		fmt.Fprintf(&buf, "ino %d ", atts.Ino)
	}

	ktext, ok := kindNames[atts.Kind]
	if !ok {
		panic("Invalid kind")
	}
	fmt.Fprintf(&buf, "kind %s ", ktext)

	if hasTime {
		fmt.Fprintf(&buf, "mtime %d ", atts.Mtime)
	}
	if !isLink {
		fmt.Fprintf(&buf, "perm %d ", atts.Perm)
	}
	if isReg && atts.Sha != nil {
		fmt.Fprintf(&buf, "sha1 %x ", atts.Sha)
	}
	if isReg {
		fmt.Fprintf(&buf, "size %d ", atts.Size)
	}
	if isLink {
		fmt.Fprintf(&buf, "targ %s ", escapeString(atts.Targ))
	}
	if !isLink {
		fmt.Fprintf(&buf, "uid %d ", atts.Uid)
	}

	return buf.String()
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
