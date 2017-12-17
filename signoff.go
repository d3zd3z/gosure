package main

import (
	"log"

	"davidb.org/x/gosure/sure"
	"github.com/spf13/cobra"
)

func doSignoff(cmd *cobra.Command, args []string) {
	oldTree, err := storeArg.ReadBak()
	if err != nil {
		log.Fatal(err)
	}

	newTree, err := storeArg.ReadDat()
	if err != nil {
		log.Fatal(err)
	}

	sure.CompareTrees(oldTree, newTree)
}
