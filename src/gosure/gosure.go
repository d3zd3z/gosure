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

		writeSure("2sure.0", dir)
	case "tmp":
		in, err := ReadSure("2sure.0")
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
	log.Fatalf("Usage: gosure {scan|tmp}\n")
}

type Node struct {
	name   string
	atts   map[string]string
	costly func() map[string]string // Get the atts that are costly to make.
}
