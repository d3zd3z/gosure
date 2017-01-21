package main

import (
	"compress/gzip"
	"log"
	"os"

	"davidb.org/code/gosure/sure"
)

func main() {
	node, err := sure.ScanFs(".")
	if err != nil {
		log.Fatal(err)
	}

	est := node.EstimateHashes()
	prog := sure.NewProgress(est.Files, est.Bytes)
	prog.Flush()
	node.ComputeHashes(&prog)
	prog.Flush()

	// Dump it out for now in gob format.
	fd, err := os.Create("2sure.dat.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	zfd := gzip.NewWriter(fd)
	defer zfd.Close()

	err = node.Encode(zfd)
	if err != nil {
		log.Fatal(err)
	}
}
