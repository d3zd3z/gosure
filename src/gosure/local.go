// Local filesystem traversal.

package main

import (
	"errors"
	"fmt"
	"io"
	"linuxdir"
	"log"
	"os"
	"sha"
	"strconv"
	"syscall"
)

type DirWalker interface {
	Info() *Node
	Path() string
	NextDir() (dir DirWalker, err error)
	NextNonDir() (node *Node, err error)
	io.Closer

	Skip() (err error)
}

// The iterator is destructive, and single pass.
type LocalDir struct {
	dirs    []*Node
	nondirs []*Node
	info    *Node
	path    string
}

func WalkRoot(path string) (dir DirWalker, err error) {
	stat, err := os.Lstat(path)
	if err != nil {
		return
	}

	if !stat.IsDir() {
		err = errors.New("Expecting directory for walk")
		return
	}

	root := makeLocalNode(path, stat)
	dir, err = buildLocalDir(path, root)
	return
}

// Accessors.
func (p *LocalDir) Info() *Node  { return p.info }
func (p *LocalDir) Path() string { return p.path }
func (p *LocalDir) Close() error { return nil }

// No work needed to skip local dirs.
func (p *LocalDir) Skip() error { return nil }

func buildLocalDir(path string, dirStat *Node) (dir *LocalDir, err error) {
	entries, err := linuxdir.Readdir(path)
	if err != nil {
		return
	}

	var dirs []*Node
	var nondirs []*Node

	for _, ent := range entries {
		tmp := makeLocalNode(path, ent)
		if ent.IsDir() {
			dirs = append(dirs, tmp)
		} else {
			if ent.Name() == "2sure.dat.gz" || ent.Name() == "2sure.bak.gz" || ent.Name() == "2sure.0.gz" {
				continue
			}
			nondirs = append(nondirs, tmp)
		}
	}

	dir = &LocalDir{dirs, nondirs, dirStat, path}
	return
}

// Get the next directory.  Returns nil if there are none.
func (p *LocalDir) NextDir() (dir DirWalker, err error) {
	if len(p.dirs) == 0 {
		return
	}

	n := p.dirs[0]
	p.dirs = p.dirs[1:]
	dir, err = buildLocalDir(p.path+"/"+n.name, n)
	return
}

// Get the next file.  Panics if there are still directories left.
func (p *LocalDir) NextNonDir() (node *Node, err error) {
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

func i64toa(i int64) string {
	return strconv.FormatInt(i, 10)
}

func u64toa(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func makeLocalNode(path string, info os.FileInfo) (n *Node) {
	atts := make(map[string]string)
	costly := noCostly
	sys := info.Sys().(*syscall.Stat_t)

	switch sys.Mode & syscall.S_IFMT {
	case syscall.S_IFDIR:
		basePerms(atts, "dir", sys)
	case syscall.S_IFREG:
		basePerms(atts, "file", sys)
		timeInfo(atts, sys)
		atts["ino"] = u64toa(sys.Ino)

		costly = func() (atts map[string]string) {
			atts = make(map[string]string)
			hash, err := sha.HashFile(path + "/" + info.Name())
			if err != nil {
				log.Printf("Unable to hash file: %s (%s)", path+"/"+info.Name(), err)
			}
			hex := make([]byte, 40)
			for i, ch := range hash {
				hex[2*i] = hexDigits[ch>>4]
				hex[2*i+1] = hexDigits[ch&0xf]
			}
			atts["sha1"] = string(hex)
			return
		}
	case syscall.S_IFLNK:
		atts["kind"] = "lnk"
		target, err := os.Readlink(path + "/" + info.Name())
		if err != nil {
			log.Printf("Error reading symlink: %s", path+"/"+info.Name())
		} else {
			atts["targ"] = target
		}
	case syscall.S_IFIFO:
		basePerms(atts, "fifo", sys)
	case syscall.S_IFSOCK:
		basePerms(atts, "sock", sys)
	case syscall.S_IFCHR:
		basePerms(atts, "chr", sys)
		devInfo(atts, sys)
		// TODO: These should have time info on them?
	case syscall.S_IFBLK:
		basePerms(atts, "blk", sys)
		devInfo(atts, sys)
	default:
		fmt.Printf("Node: %+v\n", info)
		panic("Unexpected file type")
	}

	n = &Node{name: info.Name(), atts: atts, costly: costly}
	return

}

// Set base permissions
func basePerms(atts map[string]string, kind string, sys *syscall.Stat_t) {
	atts["kind"] = kind
	atts["uid"] = u64toa(uint64(sys.Uid))
	atts["gid"] = u64toa(uint64(sys.Gid))
	atts["perm"] = permission(sys)
}

func devInfo(atts map[string]string, sys *syscall.Stat_t) {
	atts["devmaj"] = u64toa(linuxdir.Major(sys.Rdev))
	atts["devmin"] = u64toa(linuxdir.Minor(sys.Rdev))
}

func timeInfo(atts map[string]string, sys *syscall.Stat_t) {
	atts["mtime"] = i64toa(sys.Mtim.Sec)
	atts["ctime"] = i64toa(sys.Ctim.Sec)
}

// The Permission() call in 'os' is incorrect, and masks off too many
// bits.
func permission(sys *syscall.Stat_t) string {
	return u64toa(uint64(sys.Mode &^ syscall.S_IFMT))
}
