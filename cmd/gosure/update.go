package main

import (
	"log"

	"davidb.org/x/gosure"
	"davidb.org/x/gosure/status"

	"github.com/spf13/cobra"
)

func doUpdate(cmd *cobra.Command, args []string) {
	mgr := status.NewManager()
	defer mgr.Close()

	err := gosure.Scan(&storeArg, scanDir, mgr)
	if err != nil {
		log.Fatal(err)
	}
}
