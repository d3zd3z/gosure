// File integrity testing.

package main

import (
	"fmt"
	"linuxdir"
	"log"
	"os"
)

var _ = linuxdir.Readdir
var _ = fmt.Printf
var _ = log.Printf

const magic = "asure-2.0\n-----\n"

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	switch os.Args[1] {
	case "scan":
		dir, err := WalkRoot(".")
		if err != nil {
			log.Fatalf("Unable to walk root directory: %s", err)
		}
		defer dir.Close()

		err = writeSure("2sure.0.gz", dir)
		if err != nil {
			log.Fatalf("Error writing surefile: %s", err)
		}
		// TODO: Handle these?
		_ = os.Rename("2sure.dat.gz", "2sure.bak.gz")
		_ = os.Rename("2sure.0.gz", "2sure.dat.gz")
	case "check":
		dir1, err := ReadSure("2sure.dat.gz")
		if err != nil {
			log.Fatalf("Unable to read surefile: %s", err)
		}
		defer dir1.Close()

		dir2, err := WalkRoot(".")
		if err != nil {
			log.Fatalf("Error walking root dir: %s", err)
		}
		defer dir2.Close()

		Compare(dir1, dir2)

	case "signoff":
		left, err := ReadSure("2sure.bak.gz")
		if err != nil {
			log.Fatalf("Unable to read backup surefile: %s", err)
		}
		right, err := ReadSure("2sure.dat.gz")
		if err != nil {
			log.Fatalf("Unable to read surefile: %s", err)
		}
		Compare(left, right)
	case "tmp":
		in, err := ReadSure("2sure.0.gz")
		if err != nil {
			log.Fatalf("Unable to read surefile: %s", err)
		}
		defer in.Close()

		writeSure("tmp", in)
	default:
		usage()
	}
}

func usage() {
	log.Fatalf("Usage: gosure {scan|check|tmp}\n")
}

type Node struct {
	name   string
	atts   map[string]string
	costly func() map[string]string // Get the atts that are costly to make.
}

// TODO: costly should return an error.
func getAllAtts(node *Node) (atts map[string]string) {
	atts = make(map[string]string)

	for k, v := range node.atts {
		atts[k] = v
	}
	for k, v := range node.costly() {
		atts[k] = v
	}
	return
}
