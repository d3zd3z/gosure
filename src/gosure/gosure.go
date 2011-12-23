// File integrity testing.

package main

import (
	"fmt"
	"linuxdir"
	"log"
	"os"
	"sha"
	"strconv"
)

var _ = linuxdir.Readdir
var _ = fmt.Printf
var _ = log.Printf

const magic = "asure-2.0\n-----\n"

func main() {
	dir, err := walk(".")
	if err != nil {
		log.Fatalf("Unable to walk root directory: %s", err)
	}

	// fmt.Printf("%v\n", dir)
	// fmt.Fprintf(os.Stdout, "%s", magic)
	// dumpDir(os.Stdout, "__root__", dir)
	writeSure("2sure.0.gz", dir)
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
	path    string
}

type node struct {
	name   string
	atts   map[string]string
	costly func() map[string]string // Get the atts that are costly to make.
}

const hexDigits = "0123456789abcdef"

func makeNode(path string, info *os.FileInfo) (n *node) {
	atts := make(map[string]string)
	var costly func() map[string]string

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

		costly = func() (atts map[string]string) {
			atts = make(map[string]string)
			hash, err := sha.HashFile(path + "/" + info.Name)
			if err != nil {
				log.Printf("Unable to hash file: %s", path + "/" + info.Name)
			}
			hex := make([]byte, 40)
			for i, ch := range hash {
				hex[2*i] = hexDigits[ch >> 4]
				hex[2*i+1] = hexDigits[ch & 0xf]
			}
			atts["sha1"] = string(hex)
			return
		}
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

	n = &node{name: info.Name, atts: atts, costly: costly}
	return

}
