package main

import (
	"log"

	"davidb.org/x/gosure/sure"

	"github.com/spf13/cobra"
)

func doUpdate(cmd *cobra.Command, args []string) {
	oldTree, err := storeArg.ReadDat()
	if err != nil {
		log.Fatal(err)
	}

	newTree, err := sure.ScanFs(scanDir)
	if err != nil {
		log.Fatal(err)
	}

	sure.MigrateHashes(oldTree, newTree)
	hashUpdate(newTree, scanDir)
	err = storeArg.Write(newTree)
	if err != nil {
		log.Fatal(err)
	}
}
