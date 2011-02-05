// File integrity checking

package main

import "flag"
import "fmt"
import "os"

var surefileArg = flag.String("file", "2sure", "name of surefile, will have .dat appended")
var helpArg = flag.Bool("help", false, "Ask for help")

func usage(message string) {
	fmt.Printf("error: %s\n", message)
	fmt.Printf("usage: gosure [{-f|--file} name] {scan|update|check|signoff|show|walk}\n\n")
	os.Exit(1)
}

type NodeReader interface {
	ReadNode() (Node, os.Error)
	Close()
}

// Show the entire tree.
func show(nodes NodeReader) {
	for {
		node, err := nodes.ReadNode()
		if err == os.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("Node %d, %s, %v\n", node.GetKind(), node.GetName(), node.GetAtts())
	}
}

func parseArgs() {
	flag.Parse()
	if *helpArg {
		usage("Help")
	}
	if flag.NArg() != 1 {
		usage("Expecting a single command")
	}

	switch flag.Arg(0) {
	case "show":
		sf, err := surefile(*surefileArg)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
		defer sf.Close()

		show(sf)

	case "walk":
		nodes, err := walkTree(".")
		if err != nil {
			fmt.Printf("error: %s\n", err)
			os.Exit(1)
		}
		defer nodes.Close()

		show(nodes)
	default:
		usage(fmt.Sprintf("Unknown command: '%s'", flag.Arg(0)))
	}
}

func main() {
	parseArgs()
}
