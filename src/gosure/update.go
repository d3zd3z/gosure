package main

import (
	"os"
)

// Update is an iterator that returns the results of the 'right'
// iterator.  However, it also uses information available in the
// 'left' iterator to get the 'costly' attributes if it is clear that
// the file hasn't changed.

func NewUpdater(left, right DirWalker) (dir DirWalker, err os.Error) {
	var tmp UpdateDir

	tmp.left = left
	tmp.right = right

	tmp.lchild, err = left.NextDir()
	if err != nil {
		return
	}

	tmp.rchild, err = right.NextDir()
	if err != nil {
		return
	}

	dir = &tmp
	return
}

type UpdateDir struct {
	left, right DirWalker

	// Cursors pointing to the next value.  lchild.Info().name
	// should always be <= rchild.Info().name unless one is nil.
	lchild, rchild DirWalker

	// Cursors pointing to the next nodes for non-dirs
	lnode, rnode *Node
}

// Close this combining iterator.  Does _not_ close the child
// iterators.
func (p *UpdateDir) Close() (err os.Error) { return }

func (p *UpdateDir) Info() *Node  { return p.right.Info() }
func (p *UpdateDir) Path() string { return p.right.Path() }

func (p *UpdateDir) NextDir() (dir DirWalker, err os.Error) {
	var tmp DirWalker
	if p.rchild != nil {
		tmp, err = p.right.NextDir()
		if err != nil {
			return
		}
	}
	dir = p.rchild
	p.rchild = tmp

	left := p.lchild
	// Skip the left child if it comes before this one.
	for left != nil && (dir == nil || left.Info().name < dir.Info().name) {
		err = left.Skip()
		if err != nil {
			return
		}
		left, err = p.left.NextDir()
		p.lchild = left
		if err != nil {
			return
		}
	}

	// If the left and right match, create a new UpdateDir node to
	// track it.
	if left != nil && dir != nil && left.Info().name == dir.Info().name {
		dir, err = NewUpdater(p.lchild, dir)
	}

	// If we've concluded directories, setup the file iterators.
	if dir == nil && err == nil {
		p.lnode, err = p.left.NextNonDir()
		if err != nil {
			return
		}

		p.rnode, err = p.right.NextNonDir()
		if err != nil {
			return
		}
	}

	return
}

func (p *UpdateDir) NextNonDir() (node *Node, err os.Error) {
	var tmp *Node
	if p.rnode != nil {
		tmp, err = p.right.NextNonDir()
		if err != nil {
			return
		}
	}
	node = p.rnode
	p.rnode = tmp

	// Skip the left node, if it comes before this one.
	left := p.lnode
	for left != nil && (node == nil || left.name < node.name) {
		left, err = p.left.NextNonDir()
		p.lnode = left
		if err != nil {
			return
		}
	}

	// If the nodes match sufficiencly, then construct a new node.
	sameNode(left, &node)

	return
}

func sameNode(left *Node, right **Node) {
	if left == nil || *right == nil {
		return
	}
	if left.name != (*right).name {
		return
	}

	latts := left.atts
	ratts := (*right).atts

	if !keySame(latts, ratts, "ino") {
		return
	}
	if !keySame(latts, ratts, "ctime") {
		return
	}
	if latts["kind"] != "file" || ratts["kind"] != "file" {
		return
	}
	_, present := ratts["sha1"]
	if present {
		return
	}

	latts = getAllAtts(left)

	sha, present := latts["sha1"]
	if !present {
		return
	}

	ratts = make(map[string]string)
	for k, v := range (*right).atts {
		ratts[k] = v
	}
	ratts["sha1"] = sha

	var tmp Node
	tmp.name = (*right).name
	tmp.atts = ratts
	tmp.costly = noCostly

	*right = &tmp
}

// Compares to maps by a given key.  The value must be present in
// both, and have the same value.
func keySame(a, b map[string]string, key string) bool {
	avalue, present := a[key]
	if !present {
		return false
	}
	bvalue, present := b[key]
	return present && avalue == bvalue
}

func (p *UpdateDir) Skip() (err os.Error) {
	e1 := p.left.Skip()
	err = p.right.Skip()
	if err == nil {
		err = e1
	}
	return
}
