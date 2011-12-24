// Local filesystem traversal.

package main

import (
	"linuxdir"
	"os"
)

type DirWalker interface {
	Info() *Node
	Path() string
	NextDir() (dir DirWalker, err os.Error)
	NextNonDir() (node *Node, err os.Error)
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

	root := makeNode(path, stat)
	dir, err = buildLocalDir(path, root)
	return
}

// Accessors.
func (p *LocalDir) Info() *Node  { return p.info }
func (p *LocalDir) Path() string { return p.path }

func buildLocalDir(path string, dirStat *Node) (dir *LocalDir, err os.Error) {
	entries, err := linuxdir.Readdir(path)
	if err != nil {
		return
	}

	var dirs []*Node
	var nondirs []*Node

	for _, ent := range entries {
		tmp := makeNode(path, ent)
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
