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
)

func writeSure(path string, info *dirInfo) (err os.Error) {
	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	zfile, err := gzip.NewWriter(file)
	if err != nil {
		return
	}
	defer zfile.Close()

	io.WriteString(zfile, magic)
	dumpDir(zfile, "__root__", info)
	return
}

func dumpDir(w io.Writer, name string, info *dirInfo) {
	fmt.Fprintf(w, "d%s [%s]\n", escapeString(name), encodeAtts(info.info))
	for _, n := range info.dirs {
		path := info.path + "/" + n.name
		child, err := buildDir(path, n)
		if err != nil {
			log.Fatalf("can't walk dir: %s (%s)", err, path)
		}
		dumpDir(w, n.name, child)
	}
	fmt.Fprintf(w, "-\n")

	for _, n := range info.nondirs {
		fmt.Fprintf(w, "f%s [%s]\n", escapeString(n.name), encodeAtts(n))
	}

	fmt.Fprintf(w, "u\n")
}

func encodeAtts(node *node) string {
	var buf bytes.Buffer

	// TODO: Compute expensive atts.
	var costly map[string]string
	if node.costly != nil {
		costly = node.costly()
	} else {
		costly = make(map[string]string)
	}
	pairs := make([]stringPair, 0, len(node.atts) + len(costly))

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
	key string
	value string
}
type stringPairSlice []stringPair

func (p stringPairSlice) Len() int { return len(p) }
func (p stringPairSlice) Less(i, j int) bool { return p[i].key < p[j].key }
func (p stringPairSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

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
