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
