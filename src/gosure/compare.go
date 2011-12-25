// Comparison of two trees.

package main

import (
	"fmt"
	"log"
	"os"
	"sort"
)

// Call into the comparison, capturing any error that is thrown.
func Compare(left, right DirWalker) (err os.Error) {
	defer func() {
		e := recover()
		var correct bool
		err, correct = e.(os.Error)
		if !correct {
			err = os.NewError("Unknown error")
		}
	}()

	compare(left, right)
	return
}

func compare(left, right DirWalker) {
	compareSubdirs(left, right)
	compareNodes(left, right)
}

func compareSubdirs(left, right DirWalker) {
	leftEnt := nextDir(left)
	rightEnt := nextDir(right)

	// Compare as long as there are entries.
	for leftEnt != nil && rightEnt != nil {
		ln := leftEnt.Info()
		rn := rightEnt.Info()

		switch {
		case ln.name == rn.name:
			compareAtts(right, ln, rn)
			compare(leftEnt, rightEnt)
			leftEnt = nextDir(left)
			rightEnt = nextDir(right)

		case ln.name < rn.name:
			fmt.Printf("- dir                    %s\n", leftEnt.Path())
			skipDir(leftEnt)
			leftEnt = nextDir(left)

		default:
			fmt.Printf("+ dir                    %s\n", rightEnt.Path())
			skipDir(rightEnt)
			rightEnt = nextDir(right)
		}
	}

	for leftEnt != nil {
		fmt.Printf("- dir                    %s\n", leftEnt.Path())
		skipDir(leftEnt)
		leftEnt = nextDir(left)
	}
	for rightEnt != nil {
		fmt.Printf("+ dir                    %s\n", rightEnt.Path())
		skipDir(rightEnt)
		rightEnt = nextDir(right)
	}
	return
}

func nextDir(dir DirWalker) (child DirWalker) {
	child, err := dir.NextDir()
	if err != nil {
		panic(err)
	}
	return
}

func nextNonDir(dir DirWalker) (node *Node) {
	node, err := dir.NextNonDir()
	if err != nil {
		panic(err)
	}
	return
}

func compareNodes(left, right DirWalker) {
	leftEnt := nextNonDir(left)
	rightEnt := nextNonDir(right)

	for leftEnt != nil && rightEnt != nil {
		switch {
		case leftEnt.name == rightEnt.name:
			compareAtts(right, leftEnt, rightEnt)
			leftEnt = nextNonDir(left)
			rightEnt = nextNonDir(right)

		case leftEnt.name < rightEnt.name:
			fmt.Printf("- file                   %s/%s\n", left.Path(), leftEnt.name)
			leftEnt = nextNonDir(left)
		default:
			fmt.Printf("+ file                   %s/%s\n", right.Path(), rightEnt.name)
			rightEnt = nextNonDir(right)
		}
	}

	for leftEnt != nil {
		fmt.Printf("- file                   %s/%s\n", left.Path(), leftEnt.name)
		leftEnt = nextNonDir(left)
	}
	for rightEnt != nil {
		fmt.Printf("+ file                   %s/%s\n", right.Path(), rightEnt.name)
		rightEnt = nextNonDir(right)
	}
	return
}

func compareAtts(rightWalk DirWalker, left, right *Node) {
	latts := getAllAtts(left)
	ratts := getAllAtts(right)

	changed := make([]string, 0, 10)
	for key, lvalue := range latts {
		// Ignore 'ctime' and 'ino', which are expected to
		// change.
		if key == "ctime" || key == "ino" {
			continue
		}

		rvalue, present := ratts[key]
		if !present {
			log.Printf("Missing attribute: %s\n", key)
		} else if lvalue != rvalue {
			changed = append(changed, key)
		}
	}

	for key, _ := range ratts {
		_, present := latts[key]
		if !present {
			log.Printf("Extra attribute: %s\n", key)
		}
	}

	if len(changed) == 0 {
		return
	}

	sort.Sort(sort.StringSlice(changed))

	msg := ""
	for _, key := range changed {
		msg += "," + key
	}

	fmt.Printf("  [%-20s] %s/%s\n", msg[1:], rightWalk.Path(), left.name)
}

// Skip a directory tree.  TODO: Make this a smarter method to avoid
// having to traverse an entire FS.
func skipDir(dir DirWalker) {
	err := dir.Skip()
	if err != nil {
		panic(err)
	}
}
