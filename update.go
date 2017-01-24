package main

import (
	"compress/gzip"
	"log"
	"os"

	"davidb.org/code/gosure/sure"

	"github.com/spf13/cobra"
)

func doUpdate(cmd *cobra.Command, args []string) {
	oldTree, err := loadTree("2sure.dat.gz")
	if err != nil {
		log.Fatal(err)
	}

	newTree, err := sure.ScanFs(".")
	if err != nil {
		log.Fatal(err)
	}

	sure.MigrateHashes(oldTree, newTree)
	hashSave(newTree)
}

func loadTree(name string) (tree *sure.Tree, err error) {
	fd, err := os.Open(name)
	if err != nil {
		return
	}
	defer fd.Close()

	zfd, err := gzip.NewReader(fd)
	if err != nil {
		return
	}
	defer zfd.Close()

	tree, err = sure.Decode(zfd)
	return
}
