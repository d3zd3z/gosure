// Walk a local tree.

package main

import (
	"sort"
	"strconv"
	"os"
	"container/list"
	"crypto/sha1"
	"strings"
	"syscall"
)

type walkReader struct {
	nodes *list.List
}

type BasicNode struct {
	kind int
	name string
}

func (n *BasicNode) GetKind() int { return n.kind }
func (n *BasicNode) GetAtts() map[string]string { return make(map[string]string) }
func (n *BasicNode) GetExpensiveAtts() map[string]string { return make(map[string]string) }
func (n *BasicNode) GetName() string { return n.name }

type EnterNode struct {
	path string
	atts map[string]string
	BasicNode
}

type LeaveNode struct {
	BasicNode
}

type MarkNode struct {
	BasicNode
}

type FileNode struct {
	BasicNode
	atts map[string]string
}

type DataFileNode struct {
	FileNode
	fullPath string
}

func (n *FileNode) GetAtts() map[string]string { return n.atts }

const hexDigits = "0123456789abcdef"

func (n *DataFileNode) GetExpensiveAtts() (atts map[string]string) {
	atts = make(map[string]string)
	hash := sha1.New()
	file, err := os.Open(n.fullPath, os.O_RDONLY | syscall.O_NOATIME, 0)
	if err != nil {
		file, err = os.Open(n.fullPath, os.O_RDONLY, 0)
	}
	if err != nil {
		// Don't set hash if we can't read the file.
		return
	}
	defer file.Close()

	buffer := make([]byte, 65536)
	for {
		n, err := file.Read(buffer)
		if err == os.EOF {
			break
		}
		if err != nil {
			// TODO: Warn about read problem.
			return
		}

		nn, err := hash.Write(buffer[0:n])
		if err != nil || nn != n {
			// TODO: Warn about hash problem.
			return
		}
	}
	sum := hash.Sum()
	result := make([]byte, 40, 40)
	for i, ch := range sum {
		result[2*i] = hexDigits[ch >> 4]
		result[2*i+1] = hexDigits[ch & 0x0f]
	}

	atts["sha1"] = string(result)

	return
}

func (n *EnterNode) GetAtts() map[string]string { return n.atts }

func makeLeaveNode(name string) *LeaveNode {
	return &LeaveNode{BasicNode{LEAVE, name}}
}

func makeMarkNode(name string) *MarkNode {
	return &MarkNode{BasicNode{MARK, name}}
}

func makeEnterNode(path, name string, info *os.FileInfo) *EnterNode {
	atts := make(map[string]string)
	atts["kind"] = "dir"
	atts["uid"] = strconv.Itoa(info.Uid)
	atts["gid"] = strconv.Itoa(info.Gid)
	atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
	return &EnterNode{path: path, atts: atts,
		BasicNode: BasicNode{ENTER, name}}
}

func makeFileNode(path string, info *os.FileInfo) Node {
	atts := make(map[string]string)

	switch {
	case info.IsRegular():
		atts["kind"] = "file"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
		atts["mtime"] = strconv.Itoa64(info.Mtime_ns / 1000000000)
		atts["ctime"] = strconv.Itoa64(info.Ctime_ns / 1000000000)
		atts["ino"] = strconv.Uitoa64(info.Ino)
	case info.IsSymlink():
		atts["kind"] = "lnk"
		target, err := os.Readlink(path + "/" + info.Name)
		if err == nil {
			atts["targ"] = target
		}
	case info.IsSocket():
		atts["kind"] = "sock"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
	case info.IsFifo():
		atts["kind"] = "fifo"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
	case info.IsBlock():
		atts["kind"] = "blk"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
		// This is non-portable.
		atts["devmaj"] = strconv.Uitoa64(info.Rdev >> 8)
		atts["devmin"] = strconv.Uitoa64(info.Rdev & 0xFF)
	case info.IsChar():
		atts["kind"] = "chr"
		atts["uid"] = strconv.Itoa(info.Uid)
		atts["gid"] = strconv.Itoa(info.Gid)
		atts["perm"] = strconv.Uitoa64(uint64(info.Permission()))
		// This is non-portable.
		atts["devmaj"] = strconv.Uitoa64(info.Rdev >> 8)
		atts["devmin"] = strconv.Uitoa64(info.Rdev & 0xFF)
	default:
		panic("Unsupported file type")
	}

	if info.IsRegular() {
		return &DataFileNode{
			FileNode{BasicNode{NODE, info.Name}, atts},
			path + "/" + info.Name}
	}
	return &FileNode{BasicNode{NODE, info.Name}, atts}
}

func (r *walkReader) Close() {}

// Sort the names in the FileInfo pointers in reverse order.
type NameSorter []*os.FileInfo

func (n NameSorter) Len() int { return len(n) }
func (n NameSorter) Less(i, j int) bool {
	return n[i].Name > n[j].Name
}
func (n NameSorter) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (r *walkReader) insertDir(enter *EnterNode) os.Error {
	fd, err := os.Open(enter.path, os.O_RDONLY, 0)
	if err != nil {
		goto error
	}
	defer fd.Close()

	// Note that Readdir doesn't process names in inode
	// order, so will probably be inefficient on large
	// directories.
	children, err := fd.Readdir(-1)
	if err != nil {
		goto error
	}
	dirs := make([]*os.FileInfo, 0, len(children))
	files := make([]*os.FileInfo, 0, len(children))
	for i, _ := range children {
		child := &children[i]
		if strings.HasPrefix(child.Name, "2sure.") { continue }
		if child.IsDirectory() {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}

	sort.Sort(NameSorter(files))
	sort.Sort(NameSorter(dirs))

	r.nodes.PushFront(makeLeaveNode(enter.name))
	for _, node := range files {
		r.nodes.PushFront(makeFileNode(enter.path, node))
	}
	r.nodes.PushFront(makeMarkNode(enter.name))
	for _, node := range dirs {
		r.nodes.PushFront(makeEnterNode(enter.path + "/" + node.Name, node.Name, node))
	}

	return nil

error:
	// Note that we don't return an error, but just push an empty
	// directory.  Probably should warn here.
	r.nodes.PushFront(makeLeaveNode(enter.name))
	r.nodes.PushFront(makeMarkNode(enter.name))
	return nil
}

func (r *walkReader) ReadNode() (Node, os.Error) {
	nodeElement := r.nodes.Front()
	if nodeElement == nil {
		return nil, os.EOF
	}

	node := r.nodes.Remove(nodeElement)

	enter, ok := node.(*EnterNode)
	if ok {
		r.insertDir(enter)
	}

	return node.(Node), nil
}

func dirWalker(base string)

func walkTree(base string) (NodeReader, os.Error) {
	info, err := os.Lstat(base)
	if err != nil {
		goto error
	}

	nodes := list.New()
	nodes.PushBack(makeEnterNode(base, "__root__", info))
	return &walkReader{nodes}, nil

error:
	return nil, err
}
