package sure

import (
	"log"
	"path"
	"runtime"
	"sync"

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
	cpus := runtime.NumCPU()

	req := make(chan hashUpdate, 2*cpus)
	var wg sync.WaitGroup

	wg.Add(cpus)

	for i := 0; i < cpus; i++ {
		go updateWorker(req, &wg, prog)
	}

	t.hashWalk(prog, dir, req)
	close(req)

	// Wait for everyone to finish.
	wg.Wait()
}

// This message indicates a single path to compute a hash for.  The
// result will be added to the given attributes.
type hashUpdate struct {
	path string
	atts *RegAtts
}

func (t *Tree) hashWalk(prog *Progress, name string, req chan<- hashUpdate) {
	for _, f := range t.Files {
		atts, ok := f.Atts.(*RegAtts)
		if ok && atts.Sha1 == nil {
			req <- hashUpdate{
				path: path.Join(name, f.Name),
				atts: atts,
			}
		}
	}

	// And the children.
	for _, c := range t.Children {
		c.hashWalk(prog, path.Join(name, c.Name), req)
	}
}

// updateWorker pulls messages from 'req', hashes the file, and then
// tells the wg when it is done.
func updateWorker(req <-chan hashUpdate, wg *sync.WaitGroup, prog *Progress) {
	for hu := range req {
		hash, err := sha.HashFile(hu.path)
		if err != nil {
			log.Printf("Unable to hash file: %s", err)
			continue
		}
		hu.atts.Sha1 = hash
		prog.Update(1, uint64(hu.atts.Size))
	}

	wg.Done()
}
