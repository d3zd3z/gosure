package main

// Writing data to surefiles.

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"io"
	"os"
	"sort"
	"strings"
)

func writeSure(path string, info DirWalker) (err os.Error) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	var zfile io.Writer
	if strings.HasSuffix(path, ".gz") {
		var tmp *gzip.Compressor
		tmp, err = gzip.NewWriter(file)
		if err != nil {
			return
		}
		defer tmp.Close()
		zfile = tmp
	} else {
		zfile = file
	}

	io.WriteString(zfile, magic)
	dumpDir(zfile, "__root__", info)
	return
}

func dumpDir(w io.Writer, name string, info DirWalker) {
	fmt.Fprintf(w, "d%s [%s]\n", escapeString(name), encodeAtts(info.Info()))
	for {
		child, err := info.NextDir()
		if err != nil {
			log.Fatalf("can't walk dir: %s (%s)", err, info.Path())
		}
		if child == nil {
			break
		}
		dumpDir(w, child.Info().name, child)
	}
	fmt.Fprintf(w, "-\n")

	for {
		n, err := info.NextNonDir()
		if err != nil {
			log.Fatalf("Error walking files: %s (%s)", err, info.Path())
		}
		if n == nil {
			break
		}
		fmt.Fprintf(w, "f%s [%s]\n", escapeString(n.name), encodeAtts(n))
	}

	fmt.Fprintf(w, "u\n")
}

func encodeAtts(node *Node) string {
	var buf bytes.Buffer

	// TODO: Compute expensive atts.
	var costly map[string]string
	if node.costly != nil {
		costly = node.costly()
	} else {
		costly = make(map[string]string)
	}
	pairs := make([]stringPair, 0, len(node.atts)+len(costly))

	for k, v := range node.atts {
		pairs = append(pairs, stringPair{k, v})
	}
	for k, v := range costly {
		pairs = append(pairs, stringPair{k, v})
	}
	sort.Sort(stringPairSlice(pairs))

	for _, p := range pairs {
		// TODO: Escape this as well.
		buf.WriteString(p.key)
		buf.WriteRune(' ')
		buf.WriteString(escapeString(p.value))
		buf.WriteRune(' ')
	}

	return buf.String()
}

type stringPair struct {
	key   string
	value string
}
type stringPairSlice []stringPair

func (p stringPairSlice) Len() int           { return len(p) }
func (p stringPairSlice) Less(i, j int) bool { return p[i].key < p[j].key }
func (p stringPairSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Escape strings in quoted-printable style (roughly).  UTF-8 bytes will
// always be escaped.
func escapeString(text string) string {
	var buf bytes.Buffer
	for _, ch := range ([]byte)(text) {
		if '!' <= ch && ch <= '~' && ch != '=' {
			buf.WriteByte(ch)
		} else {
			fmt.Fprintf(&buf, "=%02x", ch)
		}
	}
	return buf.String()
}
