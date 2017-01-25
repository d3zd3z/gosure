package main

import (
	"log"

	"davidb.org/code/gosure/sure"
	"github.com/spf13/cobra"
)

func doSignoff(cmd *cobra.Command, args []string) {
	oldTree, err := loadTree("2sure.bak.gz")
	if err != nil {
		log.Fatal(err)
	}

	newTree, err := loadTree("2sure.dat.gz")
	if err != nil {
		log.Fatal(err)
	}

	sure.CompareTrees(oldTree, newTree)
}
