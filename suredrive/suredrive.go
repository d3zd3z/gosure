// Package suredrive implements higher-level wrappers around the various
// gosure libraries, making it easier both for a command-line tool, as
// well as other tools to perform the integrity scans.
package suredrive

import (
	"log"
	"time"

	"davidb.org/x/gosure/status"
	"davidb.org/x/gosure/store"
	"davidb.org/x/gosure/sure"
)

// Scan performs a scan or an update.
func Scan(st *store.Store, dir string, mgr *status.Manager) error {
	oldTree, err := st.ReadDat()
	if err != nil {
		log.Printf("no prior scan, doing initial scan\n")
	}

	meter := mgr.Meter(250 * time.Millisecond)
	newTree, err := sure.ScanFs(dir, meter)
	meter.Close()
	if err != nil {
		return err
	}

	if oldTree != nil {
		sure.MigrateHashes(oldTree, newTree)
	}
	HashUpdate(newTree, dir, mgr)

	err = st.Write(newTree)
	if err != nil {
		return err
	}

	return nil
}

// HashUpdate updates the hashes of any files that are needed.
func HashUpdate(tree *sure.Tree, dir string, mgr *status.Manager) {
	est := tree.EstimateHashes()
	meter := mgr.Meter(250 * time.Millisecond)
	prog := sure.NewProgress(est.Files, est.Bytes, meter)
	prog.Flush()
	tree.ComputeHashes(&prog, dir)
	meter.Close()
}
