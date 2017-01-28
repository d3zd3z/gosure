// Determine if we can successfully parse the storage file from SCCS.
// This is one half of understanding the format.  To do this, we'll
// need to generate a fairly complicated SCCS file.
//
// Start by performing random shuffles and such on the contents of the
// file.
package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Expecting a single command")
	}

	switch os.Args[1] {
	case "gen":
		// Generate new data.
		gen()
	case "scan":
		// Verify that we can read the file.
		if err := scan(); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Unknown command: %q", os.Args[1])
	}
}
