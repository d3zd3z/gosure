// File integrity testing.

package main

import (
	"bytes"
	"fmt"
	"io"
	"linuxdir"
	"log"
	"os"
	"sort"
	"strconv"
)

var _ = linuxdir.Readdir
var _ = fmt.Printf
var _ = log.Printf

const magic = "asure-2.0\n-----\n"

func main() {
	/*
		entries, err := linuxdir.Readdir(".")
		if err != nil {
			log.Fatal("Unable to read from the directory")
		}
		for _, e := range entries {
			fmt.Printf("%#v\n", *e)
		}
	*/

	dir, err := walk(".")
	if err != nil {
		log.Fatalf("Unable to walk root directory: %s", err)
	}

	// fmt.Printf("%v\n", dir)
	fmt.Fprintf(os.Stdout, "%s", magic)
	dumpDir(os.Stdout, "__root__", dir)
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

func walk(path string) (dir *dirInfo, err os.Error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return
	}

	if !stat.IsDirectory() {
		err = os.NewError("Expecting directory for walk")
	}

	root := makeNode(path, stat)

	dir, err = buildDir(path, root)

	return
}

func buildDir(path string, dirStat *node) (dir *dirInfo, err os.Error) {
	entries, err := linuxdir.Readdir(path)
	if err != nil {
		return
	}

	dirs := make([]*node, 0)
	nondirs := make([]*node, 0)

	for _, ent := range entries {
		tmp := makeNode(path, ent)
		if ent.IsDirectory() {
			dirs = append(dirs, tmp)
		} else {
			nondirs = append(nondirs, tmp)
		}
	}

	dir = &dirInfo{dirs, nondirs, dirStat, path}
	return
}

type dirInfo struct {
	dirs    []*node
	nondirs []*node
	info    *node
	path string
}

type node struct {
	name   string
	atts   map[string]string
	costly func() map[string]string // Get the atts that are costly to make.
}

func makeNode(path string, info *os.FileInfo) (n *node) {
	atts := make(map[string]string)

	switch {
	case info.IsDirectory():
		atts["kind"] = "dir"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
	case info.IsRegular():
		atts["kind"] = "file"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
		atts["mtime"] = strconv.Itoa64(info.Mtime_ns / 1000000000)
		atts["ctime"] = strconv.Itoa64(info.Ctime_ns / 1000000000)
		atts["ino"] = strconv.Uitoa64(info.Ino)
		// TODO: Setup expensive fetch for hash.
	case info.IsSymlink():
		atts["kind"] = "lnk"
		target, err := os.Readlink(path + "/" + info.Name)
		// TODO: Warn about the errors.
		if err == nil {
			atts["targ"] = target
		}
	default:
		fmt.Printf("Node: %v\n", *info)
		panic("Unexpected file type")
	}

	n = &node{name: info.Name, atts: atts}
	return

}
