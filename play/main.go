// Determine if we can successfully parse the storage file from SCCS.
// This is one half of understanding the format.  To do this, we'll
// need to generate a fairly complicated SCCS file.
//
// Start by performing random shuffles and such on the contents of the
// file.
package main

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/spf13/cobra"
)

var cpuProfile string
var memProfile string

func main() {
	root := &cobra.Command{
		Use:   "play command args ...",
		Short: "Playing with sccs",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
			log.Fatal("Invalid usage")
		},
	}

	fl := root.PersistentFlags()
	fl.StringVarP(&sccsFile, "sccs-file", "f", "SCCS/s.foo", "SCCS file to use")
	fl.StringVar(&cpuProfile, "cpu-profile", "", "Output cpu profile to file")
	fl.StringVar(&memProfile, "mem-profile", "", "Output memory profile to file")

	gen := &cobra.Command{
		Use:   "gen",
		Short: "Generate a SCCS data file",
		Run:   doGen,
	}

	fl = gen.Flags()
	fl.IntVarP(&genLines, "lines", "l", 100, "# Lines in data file")
	fl.IntVarP(&genDeltas, "deltas", "d", 3, "# Deltas to generate")

	root.AddCommand(gen)

	scan := &cobra.Command{
		Use:   "scan",
		Short: "Scan SCCS file",
		Run:   doScan,
	}

	root.AddCommand(scan)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func doGen(cmd *cobra.Command, args []string) {
	checkSccsFile(cmd)
	gen()
}

func doScan(cmd *cobra.Command, args []string) {
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	checkSccsFile(cmd)
	if err := scan(); err != nil {
		log.Fatal(err)
	}

	if memProfile != "" {
		f, err := os.Create(memProfile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}

func checkSccsFile(cmd *cobra.Command) {
	if !strings.HasPrefix(sccsFile, "SCCS/s.") {
		log.Fatal("sccs-file must start with SCCS/s.")
	}

	playFile = sccsFile[7:]

	if len(playFile) == 0 {
		log.Fatal("Filename after SCCS/s. must not be empty")
	}

	if strings.Contains(playFile, "/") {
		log.Fatalf("plain part of name %q must not contain a slash", playFile)
	}
}
