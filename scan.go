package main

import (
	"log"
	"time"

	"davidb.org/x/gosure/status"
	"davidb.org/x/gosure/sure"

	"github.com/spf13/cobra"
)

func doScan(cmd *cobra.Command, args []string) {
	st := status.NewManager()
	defer st.Close()

	tree, err := sure.ScanFs(scanDir)
	if err != nil {
		log.Fatal(err)
	}

	hashUpdate(tree, scanDir, st)

	err = storeArg.Write(tree)
	if err != nil {
		log.Fatal(err)
	}
}

func hashUpdate(tree *sure.Tree, dir string, st *status.Manager) {
	est := tree.EstimateHashes()
	meter := st.Meter(250 * time.Millisecond)
	prog := sure.NewProgress(est.Files, est.Bytes, meter)
	prog.Flush()
	tree.ComputeHashes(&prog, dir)
	meter.Close()
}
