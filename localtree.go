// Walk a local tree.

package main

import (
	"sort"
	"os"
	// "fmt"
	"container/list"
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

func (n *EnterNode) GetAtts() map[string]string { return n.atts }

func makeLeaveNode(name string) *LeaveNode {
	return &LeaveNode{BasicNode{LEAVE, name}}
}

func makeMarkNode(name string) *MarkNode {
	return &MarkNode{BasicNode{MARK, name}}
}

func makeEnterNode(path string, info *os.FileInfo) *EnterNode {
	return &EnterNode{path: path, atts: make(map[string]string),
		BasicNode: BasicNode{ENTER, info.Name}}
}

func makeFileNode(path string, info *os.FileInfo) *FileNode {
	return &FileNode{BasicNode{NODE, info.Name}, make(map[string]string)}
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
		r.nodes.PushFront(makeEnterNode(enter.path + "/" + enter.name, node))
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
	nodes.PushBack(makeEnterNode(base, info))
	return &walkReader{nodes}, nil

error:
	return nil, err
}
