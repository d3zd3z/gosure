// Comparing two trees.

package main

import (
	"fmt"
	"os"
	"sort"
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

type Comparer interface {
	DeleteNode(path string, left Node)
	AddNode(path string, right Node)
	SameNode(path string, left, right Node)
}

type Combiner struct {
	left, right *LazyNodes
	path string
	comp Comparer
}

// Both nodes are sitting at an enter, and should be visiting the same directory.
func (c *Combiner) enter() {
	oldPath := c.path
	assert(c.left.GetKind() == ENTER)
	assert(c.right.GetKind() == ENTER)
	assert(c.left.GetName() == c.right.GetName())
	c.path = c.path + "/" + c.left.GetName()
	c.comp.SameNode(c.path, c.left, c.right)
	c.left = c.left.Next()
	c.right = c.right.Next()

	c.descendDirs()

	assert(c.left.GetKind() == MARK)
	assert(c.right.GetKind() == MARK)
	c.comp.SameNode(c.path, c.left, c.right)
	c.left = c.left.Next()
	c.right = c.right.Next()

	c.walkFiles()

	assert(c.left.GetKind() == LEAVE)
	assert(c.right.GetKind() == LEAVE)
	c.comp.SameNode(c.path, c.left, c.right)
	c.left = c.left.Next()
	c.right = c.right.Next()
	c.path = oldPath
}

// Just after the ENTER nodes, descend the directories, appropriately.
func (c *Combiner) descendDirs() {
	delVisit := func(what Node) { c.comp.DeleteNode(c.path, what) }
	addVisit := func(what Node) { c.comp.AddNode(c.path, what) }
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
				c.left = c.skipDir(c.left, delVisit)
			case leftName > rightName:
				c.right = c.skipDir(c.right, addVisit)
			default:
				// Same child dir, recursively descend it.
				c.enter()
			}

		case leftKind == ENTER:
			// Right is out of directories.
			c.left = c.skipDir(c.left, delVisit)
		case rightKind == ENTER:
			// Left is out of directories.
			c.right = c.skipDir(c.right, addVisit)
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
				c.comp.DeleteNode(c.path, c.left)
				c.left = c.left.Next()
			case leftName > rightName:
				// File in right, not in left.
				c.comp.AddNode(c.path, c.right)
				c.right = c.right.Next()
			default:
				// Same name, compare and skip.
				c.comp.SameNode(c.path, c.left, c.right)
				c.left = c.left.Next()
				c.right = c.right.Next()
			}
		case leftKind == NODE:
			// Lone left node.
			c.comp.DeleteNode(c.path, c.left)
			c.left = c.left.Next()
		case rightKind == NODE:
			c.comp.AddNode(c.path, c.right)
			c.right = c.right.Next()
		default:
			// Done.
			break outer
		}
	}
}

// Walk a directory in one of the trees that is not present in the other.
func (c *Combiner) skipDir(ln *LazyNodes, visit func(what Node)) *LazyNodes {
	visit(ln)
	ln = ln.Next()
	depth := 1
	for depth > 0 {
		visit(ln)
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

type ListCompare struct {
	hide int		// Hide output if > 0.
}

func niceName(name string, node Node) string {
	name = name + "/" + node.GetName()
	if len(name) > 10 && name[:10] == "/__root__/" {
		return name[10:]
	}
	return name
}

func (c *ListCompare) DeleteNode(path string, left Node) {
	kind := left.GetKind()
	if c.hide == 0 {
		fmt.Printf("- %-20s   %s\n", left.GetAtts()["kind"], niceName(path, left))
	}
	if kind == ENTER { c.hide++ }
	if kind == LEAVE { c.hide-- }
}

func (c *ListCompare) AddNode(path string, right Node) {
	kind := right.GetKind()
	if c.hide == 0 {
		fmt.Printf("+ %-20s   %s\n", right.GetAtts()["kind"], niceName(path, right))
	}
	if kind == ENTER { c.hide++ }
	if kind == LEAVE { c.hide-- }
}

type StringSort []string

func (ary StringSort) Len() int { return len(ary) }
func (ary StringSort) Less(i, j int) bool {
	return ary[i] < ary[j]
}
func (ary StringSort) Swap(i, j int) {
	ary[i], ary[j] = ary[j], ary[i]
}

func (c *ListCompare) SameNode(path string, left, right Node) {
	kind := left.GetKind()
	if kind != ENTER && kind != NODE { return }

	// Compute the attributes that differ.
	latts := GetAllAtts(left)
	ratts := GetAllAtts(right)

	changed := make([]string, 0, 10)

	for key, lvalue := range latts {
		// ctime and ino are not for comparison, but update.
		if key == "ctime" || key == "ino" { continue }

		rvalue, present := ratts[key]
		if !present {
			fmt.Printf("Missing attribute: %s\n", key)
		} else if lvalue != rvalue {
			changed = append(changed, key)
		}
	}

	// Check for Extra attributes.
	for key, _ := range ratts {
		_, present := latts[key]
		if !present {
			fmt.Printf("Extra attribute: %s\n", key)
		}
	}

	if len(changed) == 0 { return }

	sort.Sort(StringSort(changed))

	msg := ""
	for _, key := range changed {
		msg += "," + key
	}

	fmt.Printf("  [%-20s] %s\n", msg[1:], niceName(path, left))
}

func compareTrees(oldTree, newTree NodeReader) os.Error {
	left := makeLazyNodes(oldTree)
	right := makeLazyNodes(newTree)

	comb := &Combiner{left, right, "", &ListCompare{}}
	comb.enter()
	
	return os.EOF
}
