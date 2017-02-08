package main

import (
	"log"

	"davidb.org/code/gosure/sure"
	"github.com/spf13/cobra"
)

func doCheck(cmd *cobra.Command, args []string) {
	oldTree, err := storeArg.ReadDat()
	if err != nil {
		log.Fatal(err)
	}

	newTree, err := sure.ScanFs(".")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Factor this out between scan.
	est := newTree.EstimateHashes()
	prog := sure.NewProgress(est.Files, est.Bytes)
	prog.Flush()
	newTree.ComputeHashes(&prog)
	prog.Flush()

	sure.CompareTrees(oldTree, newTree)
}
