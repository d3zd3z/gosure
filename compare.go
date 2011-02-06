// Comparing two trees.

package main

import (
	"os"
)

// Lazily evaluated stream of nodes.
type LazyNodes struct {
	Node
	reader NodeReader
	next *LazyNodes
	hasNext bool
}

func makeLazyNodes(reader NodeReader) *LazyNodes {
	head, err := reader.ReadNode()
	if err == os.EOF { return nil }
	assert(err == nil)

	return &LazyNodes{head, reader, nil, false}
}

func (n *LazyNodes) Next() *LazyNodes {
	if n.hasNext { return n.next }

	n.next = makeLazyNodes(n.reader)
	n.hasNext = true
	return n.next
}

// Simple assertion.
func assert(condition bool) {
	if !condition {
		panic("Assertion failed")
	}
}

type Combiner struct {
	left, right *LazyNodes
}

// Both nodes are sitting at an enter, and should be visiting the same directory.
func (c *Combiner) enter() {
	assert(c.left.GetKind() == ENTER)
	assert(c.right.GetKind() == ENTER)
	assert(c.left.GetName() == c.right.GetName())
	c.left = c.left.Next()
	c.right = c.right.Next()

	c.descendDirs()

	assert(c.left.GetKind() == MARK)
	assert(c.right.GetKind() == MARK)
	c.left = c.left.Next()
	c.right = c.right.Next()

	c.walkFiles()

	assert(c.left.GetKind() == LEAVE)
	assert(c.right.GetKind() == LEAVE)
	c.left = c.left.Next()
	c.right = c.right.Next()
}

// Just after the ENTER nodes, descend the directories, appropriately.
func (c *Combiner) descendDirs() {
outer:
	for {
		leftKind := c.left.GetKind()
		rightKind := c.right.GetKind()

		switch {
		case leftKind == ENTER && rightKind == ENTER:
			// Both are subdirs.  Walk whichever one has the smallest name.
			leftName := c.left.GetName()
			rightName := c.right.GetName()

			switch {
			case leftName < rightName:
				c.left = c.skipDir(c.left)
			case leftName > rightName:
				c.right = c.skipDir(c.right)
			default:
				// Same child dir, recursively descend it.
				c.enter()
			}

		case leftKind == ENTER:
			// Right is out of directories.
			c.left = c.skipDir(c.left)
		case rightKind == ENTER:
			// Left is out of directories.
			c.right = c.skipDir(c.right)
		default:
			// No more subdirs.
			break outer
		}
	}
}

// Between the MARK and LEAVE contains the files.
func (c *Combiner) walkFiles() {
outer:
	for {
		leftKind := c.left.GetKind()
		rightKind := c.right.GetKind()

		switch {
		case leftKind == NODE && rightKind == NODE:
			// Still have files.  Use names to determine matches.
			leftName := c.left.GetName()
			rightName := c.right.GetName()

			switch {
			case leftName < rightName:
				// File in left, not in right.
				c.left = c.left.Next()
			case leftName > rightName:
				// File in right, not in left.
				c.right = c.right.Next()
			default:
				// Same name, compare and skip.
				c.left = c.left.Next()
				c.right = c.right.Next()
			}
		case leftKind == NODE:
			// Lone left node.
			c.left = c.left.Next()
		case rightKind == NODE:
			c.right = c.right.Next()
		default:
			// Done.
			break outer
		}
	}
}

// Walk a directory in one of the trees that is not present in the other.
func (c *Combiner) skipDir(ln *LazyNodes) *LazyNodes {
	ln = ln.Next()
	depth := 1
	for depth > 0 {
		switch ln.GetKind() {
		case ENTER:
			depth++
		case LEAVE:
			depth--
		}
		ln = ln.Next()
		assert(ln != nil)
	}

	return ln
}

func compareTrees(oldTree, newTree NodeReader) os.Error {
	left := makeLazyNodes(oldTree)
	right := makeLazyNodes(newTree)

	comb := &Combiner{left, right}
	comb.enter()
	
	return os.EOF
}
