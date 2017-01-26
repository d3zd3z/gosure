package main

import (
	"compress/gzip"
	"log"
	"os"

	"davidb.org/code/gosure/sure"

	"github.com/spf13/cobra"
)

func doScan(cmd *cobra.Command, args []string) {
	tree, err := sure.ScanFs(".")
	if err != nil {
		log.Fatal(err)
	}

	hashSave(tree)
}

func hashSave(tree *sure.Tree) {
	est := tree.EstimateHashes()
	prog := sure.NewProgress(est.Files, est.Bytes)
	prog.Flush()
	tree.ComputeHashes(&prog)
	prog.Flush()

	writeSure(tree)
	os.Rename("2sure.dat.gz", "2sure.bak.gz")
	err := os.Rename("2sure.0.gz", "2sure.dat.gz")
	if err != nil {
		log.Printf("Unable to rename 2sure.0.gz: %v", err)
	}
}

func writeSure(tree *sure.Tree) {
	// Dump it out for now in gob format.
	// TODO: Choose a tmp name that is unique by incrementing the
	// digit.
	fd, err := os.Create("2sure.0.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	zfd := gzip.NewWriter(fd)
	defer zfd.Close()

	err = tree.Encode(zfd)
	if err != nil {
		log.Fatal(err)
	}
}
