package main

import (
	"log"

	"davidb.org/code/gosure/sure"

	"github.com/spf13/cobra"
)

func doUpdate(cmd *cobra.Command, args []string) {
	oldTree, err := storeArg.ReadDat()
	if err != nil {
		log.Fatal(err)
	}

	newTree, err := sure.ScanFs(".")
	if err != nil {
		log.Fatal(err)
	}

	sure.MigrateHashes(oldTree, newTree)
	hashUpdate(newTree)
	err = storeArg.Write(newTree)
	if err != nil {
		log.Fatal(err)
	}
}
