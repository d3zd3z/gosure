package sure

import (
	"log"
	"path"

	"davidb.org/x/gosure/sha"
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
		atts, ok := f.Atts.(*RegAtts)
		if ok && atts.Sha1 == nil {
			e.Files += 1
			e.Bytes += uint64(atts.Size)
		}
	}

	// And visit all children.
	for _, c := range t.Children {
		e.update(c)
	}
}

// Update all of the file nodes that don't have hashes.
func (t *Tree) ComputeHashes(prog *Progress, dir string) {
	t.hashWalk(prog, dir)
}

func (t *Tree) hashWalk(prog *Progress, name string) {
	for _, f := range t.Files {
		atts, ok := f.Atts.(*RegAtts)
		if ok && atts.Sha1 == nil {
			hash, err := sha.HashFile(path.Join(name, f.Name))
			if err != nil {
				log.Printf("Unable to hash file: %v", err)
				continue
			}
			atts.Sha1 = hash
			prog.Update(1, uint64(atts.Size))
		}
	}

	// And the children.
	for _, c := range t.Children {
		c.hashWalk(prog, path.Join(name, c.Name))
	}
}
