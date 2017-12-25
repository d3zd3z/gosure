package main

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

func doList(cmd *cobra.Command, args []string) {
	hdr, err := storeArg.ReadHeader()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("vers | Time captured       | name\n")
	fmt.Printf("-----+---------------------+----------------\n")
	for i := len(hdr.Deltas) - 1; i >= 0; i-- {
		d := hdr.Deltas[i]

		var timeText string
		t, err := time.Parse(time.RFC3339Nano, d.Time)
		if err != nil {
			timeText = "???"
		} else {
			timeText = t.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%4d | %-19s | %s\n", d.Number, timeText, d.Name)
	}
}
