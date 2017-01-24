package sure

import (
	"fmt"
	"log"
	"path"
	"strconv"

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
		_, ok := f.Atts["sha1"]
		if !ok {
			e.Files += 1
			size, err := f.Atts.GetUint64("size")
			if err == NoKey {
				// Only nodes with a size are
				// considered for hashing.
				continue
			}
			if err != nil {
				panic(err)
			}
			e.Bytes += size
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
		_, ok := f.Atts["sha1"]
		size, sizeOk := f.Atts["size"]
		if !ok && sizeOk {
			hash, err := sha.HashFile(path.Join(name, f.Name))
			if err != nil {
				log.Printf("Unable to hash file: %v", err)
				continue
			}
			f.Atts["sha1"] = fmt.Sprintf("%x", hash)
			sz, _ := strconv.ParseUint(size, 10, 64)
			prog.Update(1, sz)
		}
	}

	// And the children.
	for _, c := range t.Children {
		c.hashWalk(prog, path.Join(name, c.Name))
	}
}
