package main

import (
	"log"
	"time"

	"davidb.org/x/gosure/status"
	"davidb.org/x/gosure/sure"

	"github.com/spf13/cobra"
)

func doUpdate(cmd *cobra.Command, args []string) {
	st := status.NewManager()
	defer st.Close()

	oldTree, err := storeArg.ReadDat()
	if err != nil {
		log.Fatal(err)
	}

	meter := st.Meter(250 * time.Millisecond)
	newTree, err := sure.ScanFs(scanDir, meter)
	meter.Close()
	if err != nil {
		log.Fatal(err)
	}

	sure.MigrateHashes(oldTree, newTree)
	hashUpdate(newTree, scanDir, st)
	err = storeArg.Write(newTree)
	if err != nil {
		log.Fatal(err)
	}
}
