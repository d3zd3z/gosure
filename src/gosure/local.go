// Local filesystem traversal.

package main

import (
	"fmt"
	"linuxdir"
	"log"
	"io"
	"os"
	"sha"
	"strconv"
)

type DirWalker interface {
	Info() *Node
	Path() string
	NextDir() (dir DirWalker, err os.Error)
	NextNonDir() (node *Node, err os.Error)
	io.Closer
}

// The iterator is destructive, and single pass.
type LocalDir struct {
	dirs    []*Node
	nondirs []*Node
	info    *Node
	path    string
}

func WalkRoot(path string) (dir DirWalker, err os.Error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return
	}

	if !stat.IsDirectory() {
		err = os.NewError("Expecting directory for walk")
		return
	}

	root := makeLocalNode(path, stat)
	dir, err = buildLocalDir(path, root)
	return
}

// Accessors.
func (p *LocalDir) Info() *Node     { return p.info }
func (p *LocalDir) Path() string    { return p.path }
func (p *LocalDir) Close() os.Error { return nil }

func buildLocalDir(path string, dirStat *Node) (dir *LocalDir, err os.Error) {
	entries, err := linuxdir.Readdir(path)
	if err != nil {
		return
	}

	var dirs []*Node
	var nondirs []*Node

	for _, ent := range entries {
		tmp := makeLocalNode(path, ent)
		if ent.IsDirectory() {
			dirs = append(dirs, tmp)
		} else {
			nondirs = append(nondirs, tmp)
		}
	}

	dir = &LocalDir{dirs, nondirs, dirStat, path}
	return
}

// Get the next directory.  Returns nil if there are none.
func (p *LocalDir) NextDir() (dir DirWalker, err os.Error) {
	if len(p.dirs) == 0 {
		return
	}

	n := p.dirs[0]
	p.dirs = p.dirs[1:]
	dir, err = buildLocalDir(p.path+"/"+n.name, n)
	return
}

// Get the next file.  Panics if there are still directories left.
func (p *LocalDir) NextNonDir() (node *Node, err os.Error) {
	if len(p.dirs) != 0 {
		panic("Iterator error, accessing files before directories")
	}

	if len(p.nondirs) == 0 {
		return
	}

	node = p.nondirs[0]
	p.nondirs = p.nondirs[1:]
	return
}

const hexDigits = "0123456789abcdef"

func makeLocalNode(path string, info *os.FileInfo) (n *Node) {
	atts := make(map[string]string)
	costly := noCostly

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
				log.Printf("Unable to hash file: %s", path+"/"+info.Name)
			}
			hex := make([]byte, 40)
			for i, ch := range hash {
				hex[2*i] = hexDigits[ch>>4]
				hex[2*i+1] = hexDigits[ch&0xf]
			}
			atts["sha1"] = string(hex)
			return
		}
	case info.IsSymlink():
		atts["kind"] = "lnk"
		target, err := os.Readlink(path + "/" + info.Name)
		if err != nil {
			log.Printf("Error reading symlink: %s", path+"/"+info.Name)
		} else {
			atts["targ"] = target
		}
	default:
		fmt.Printf("Node: %v\n", *info)
		panic("Unexpected file type")
	}

	n = &Node{name: info.Name, atts: atts, costly: costly}
	return

}
