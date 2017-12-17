package main

import (
	"log"

	"davidb.org/x/gosure/sure"

	"github.com/spf13/cobra"
)

func doScan(cmd *cobra.Command, args []string) {
	tree, err := sure.ScanFs(".")
	if err != nil {
		log.Fatal(err)
	}

	hashUpdate(tree)

	err = storeArg.Write(tree)
	if err != nil {
		log.Fatal(err)
	}
}

func hashUpdate(tree *sure.Tree) {
	est := tree.EstimateHashes()
	prog := sure.NewProgress(est.Files, est.Bytes)
	prog.Flush()
	tree.ComputeHashes(&prog)
	prog.Flush()
}
