package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"davidb.org/x/gosure/store"
	"github.com/spf13/cobra"
)

var scanDir string
var storeArg store.Store
var tags = store.NewTags(&storeArg)

var version = "compiled manually"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	root := &cobra.Command{
		Use:   "gosure command args ...",
		Short: "File integrity management",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
			log.Fatal("Invalid usage, TODO")
		},
	}

	pf := root.PersistentFlags()
	pf.VarP(&storeArg, "file", "f", "Surefile to write to")
	pf.VarP(&tags, "tag", "t", "Tags for new delta")

	scan := &cobra.Command{
		Use:   "scan",
		Short: "Scan tree",
		Long:  "Scan the tree and record integrity",
		Run:   doScan,
	}

	root.AddCommand(scan)

	pf = scan.PersistentFlags()
	pf.StringVarP(&scanDir, "dir", "d", ".", "Directory to scan")

	update := &cobra.Command{
		Use:   "update",
		Short: "Update tree",
		Long:  "Scan the tree, updating files that have changed",
		Run:   doUpdate,
	}

	pf = update.PersistentFlags()
	pf.StringVarP(&scanDir, "dir", "d", ".", "Directory to scan")

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

	pf = check.PersistentFlags()
	pf.StringVarP(&scanDir, "dir", "d", ".", "Directory to scan")
	pf.IntVarP(&checkRev, "rev", "r", -1, "Revision to check")

	root.AddCommand(check)

	list := &cobra.Command{
		Use:   "list",
		Short: "List revisions in surefile",
		Run:   doList,
	}

	root.AddCommand(list)

	version := &cobra.Command{
		Use:   "version",
		Short: "Show program version",
		Run:   doVersion,
	}

	root.AddCommand(version)

	if err := root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func doVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("gosure %s\nCompiled with %v on %v/%v\n",
		version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
