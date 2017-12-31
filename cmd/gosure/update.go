package main

import (
	"log"

	"davidb.org/x/gosure/status"
	"davidb.org/x/gosure/suredrive"

	"github.com/spf13/cobra"
)

func doUpdate(cmd *cobra.Command, args []string) {
	mgr := status.NewManager()
	defer mgr.Close()

	err := suredrive.Scan(&storeArg, scanDir, mgr)
	if err != nil {
		log.Fatal(err)
	}
}
