package main

import (
	"log"
	"time"

	"davidb.org/x/gosure/status"
	"davidb.org/x/gosure/sure"
	"github.com/spf13/cobra"
)

var checkRev int

func doCheck(cmd *cobra.Command, args []string) {
	st := status.NewManager()
	defer st.Close()

	oldTree, err := storeArg.ReadDelta(checkRev)
	if err != nil {
		log.Fatal(err)
	}

	meter := st.Meter(250 * time.Millisecond)
	newTree, err := sure.ScanFs(scanDir, meter)
	meter.Close()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Factor this out between scan.
	est := newTree.EstimateHashes()
	meter = st.Meter(250 * time.Millisecond)
	prog := sure.NewProgress(est.Files, est.Bytes, meter)
	prog.Flush()
	newTree.ComputeHashes(&prog, scanDir)
	meter.Close()

	sure.CompareTrees(oldTree, newTree)
}
