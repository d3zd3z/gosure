package sure

import (
	"log"
	"path"
	"syscall"

	"davidb.org/code/gosure/sha"
)

// A hash estimate
type Estimate struct {
	Files uint64
	Bytes uint64
}

// Estimate the amount of updates necessary to files hashes.
func (t *Tree) EstimateHashes() Estimate {
	est := Estimate{}
	est.update(t)
	return est
}

func (e *Estimate) update(t *Tree) {
	// Account for any files in this tree.
	for _, f := range t.Files {
		if f.Atts.Sha == nil {
			if f.Atts.Kind != syscall.S_IFREG {
				// Only regular files are considered
				// for hashing.
				continue
			}
			e.Files += 1
			e.Bytes += uint64(f.Atts.Size)
		}
	}

	// And visit all children.
	for _, c := range t.Children {
		e.update(c)
	}
}

// Update all of the file nodes that don't have hashes.
func (t *Tree) ComputeHashes(prog *Progress) {
	// TODO: Use proper tree location.
	t.hashWalk(prog, ".")
}

func (t *Tree) hashWalk(prog *Progress, name string) {
	for _, f := range t.Files {
		if f.Atts.Kind == syscall.S_IFREG && f.Atts.Sha == nil {
			hash, err := sha.HashFile(path.Join(name, f.Name))
			if err != nil {
				log.Printf("Unable to hash file: %v", err)
				continue
			}
			f.Atts.Sha = hash
			prog.Update(1, uint64(f.Atts.Size))
		}
	}

	// And the children.
	for _, c := range t.Children {
		c.hashWalk(prog, path.Join(name, c.Name))
	}
}
