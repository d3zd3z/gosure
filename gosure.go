package main

import (
	"compress/gzip"
	"fmt"
	"log"
	"os"

	"davidb.org/code/gosure/sure"

	"github.com/spf13/cobra"
)

var scanDir string

func main() {
	root := &cobra.Command{
		Use:   "gosure command args ...",
		Short: "File integrity management",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
			log.Fatal("Invalid usage, TODO")
		},
	}

	scan := &cobra.Command{
		Use:   "scan",
		Short: "Scan tree",
		Long:  "Scan the tree and record integrity",
		Run:   doScan,
	}

	root.AddCommand(scan)

	pf := scan.PersistentFlags()
	pf.StringVarP(&scanDir, "dir", "d", ".", "Directory to scan")

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func doScan(cmd *cobra.Command, args []string) {
	tree, err := sure.ScanFs(".")
	if err != nil {
		log.Fatal(err)
	}

	est := tree.EstimateHashes()
	prog := sure.NewProgress(est.Files, est.Bytes)
	prog.Flush()
	tree.ComputeHashes(&prog)
	prog.Flush()

	writeSure(tree)
	os.Rename("2sure.dat.gz", "2sure.bak.gz")
	err = os.Rename("2sure.0.gz", "2sure.dat.gz")
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
