// Determine if we can successfully parse the storage file from SCCS.
// This is one half of understanding the format.  To do this, we'll
// need to generate a fairly complicated SCCS file.
//
// Start by performing random shuffles and such on the contents of the
// file.
package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
)

var sccsFile string
var playFile string
var genLines int
var genDeltas int

type Sccs struct {
	name string
}

func NewSccs(name string) Sccs {
	return Sccs{
		name: name,
	}
}

// Write data array to file.
func (s *Sccs) writeData(data []int) error {
	fd, err := os.Create(s.name)
	if err != nil {
		return err
	}
	defer fd.Close()

	bfd := bufio.NewWriter(fd)
	defer bfd.Flush()

	for _, n := range data {
		fmt.Fprintf(bfd, "%d\n", n)
	}

	return nil
}

// Initialize the Sccs system
func (s *Sccs) Initial(data []int) error {
	os.Mkdir("SCCS", 0755)

	err := s.writeData(data)
	if err != nil {
		return err
	}

	err = exec.Command("sccs", "admin", "-i"+s.name, "-n", sccsFile).Run()
	if err != nil {
		return err
	}

	if err = os.Remove(s.name); err != nil {
		return err
	}

	return nil
}

// Write out a delta
func (s *Sccs) WriteDelta(data []int) error {
	if err := exec.Command("sccs", "get", "-e", s.name).Run(); err != nil {
		return err
	}

	if err := s.writeData(data); err != nil {
		return err
	}

	if err := exec.Command("sccs", "delta", "-yMessage", s.name).Run(); err != nil {
		return err
	}

	return nil
}

// Shuffle the data in the array.
func shuffle(data []int) {
	for i := len(data) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		data[i], data[j] = data[j], data[i]
	}
}

func gen() {
	if _, err := os.Lstat(playFile); err == nil {
		log.Fatalf("plain file %q is present, remove and rerun", playFile)
	}

	// Clean up prior runs.
	os.Remove(sccsFile)

	// Start with the file sorted, and create the initial version.
	data := make([]int, genLines)

	for i := range data {
		data[i] = i
	}

	ss := NewSccs(playFile)
	if err := ss.Initial(data); err != nil {
		log.Fatal(err)
	}

	// Fill the delta with a thousand random shuffled variants.
	for i := 0; i < genDeltas; i++ {
		// Shuffle a subset of the data for the delta.
		a := rand.Intn(len(data))
		b := rand.Intn(len(data))

		if a > b {
			a, b = b, a
		}

		shuffle(data[a : b+1])
		if err := ss.WriteDelta(data); err != nil {
			log.Fatal(err)
		}
	}
}
