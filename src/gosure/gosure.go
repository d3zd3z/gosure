// File integrity testing.

package main

import (
	"fmt"
	"linuxdir"
	"log"
)

var _ = linuxdir.Readdir
var _ = fmt.Printf
var _ = log.Printf

const magic = "asure-2.0\n-----\n"

func main() {
	if false {
		dir, err := WalkRoot(".")
		if err != nil {
			log.Fatalf("Unable to walk root directory: %s", err)
		}

		writeSure("2sure.0", dir)
	} else {
		in, err := ReadSure("2sure.0")
		if err != nil {
			log.Fatalf("Unable to read surefile: %s", err)
		}
		writeSure("tmp", in)
	}
}

type Node struct {
	name   string
	atts   map[string]string
	costly func() map[string]string // Get the atts that are costly to make.
}