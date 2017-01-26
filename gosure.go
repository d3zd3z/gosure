package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var scanDir string

func main() {
	root := &cobra.Command{
		Use:   "gosure command args ...",
		Short: "File integrity management",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
			log.Fatal("Invalid usage, TODO")
		},
	}

	scan := &cobra.Command{
		Use:   "scan",
		Short: "Scan tree",
		Long:  "Scan the tree and record integrity",
		Run:   doScan,
	}

	root.AddCommand(scan)

	pf := scan.PersistentFlags()
	pf.StringVarP(&scanDir, "dir", "d", ".", "Directory to scan")

	update := &cobra.Command{
		Use:   "update",
		Short: "Update tree",
		Long:  "Scan the tree, updating files that have changed",
		Run:   doUpdate,
	}

	root.AddCommand(update)

	signoff := &cobra.Command{
		Use:   "signoff",
		Short: "Compare prior scan with current",
		Run:   doSignoff,
	}

	root.AddCommand(signoff)

	check := &cobra.Command{
		Use:   "check",
		Short: "Compare current scan with tree",
		Run:   doCheck,
	}

	root.AddCommand(check)

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
