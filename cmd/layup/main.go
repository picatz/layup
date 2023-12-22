package main

import (
	"flag"
	"fmt"
	"os"

	layupv1 "github.com/picatz/layup/pkg/layup/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var usage = `Layup enables anyone to model relationships between data in a graph using 
"layers" containing "nodes" and "links" to represent relationships. It is 
designed to be a simple, flexible, and extensible way to model anything. 

Because everything is a graph.

Usage: layup <path/to/layup.hcl>`

func getArgOrExit() string {
	if len(os.Args) != 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	var help bool
	flagSet := flag.NewFlagSet("layup", flag.ExitOnError)
	flagSet.BoolVar(&help, "help", false, "Show this help message")
	flagSet.Parse(os.Args[1:])

	if help {
		fmt.Println(usage)
		os.Exit(0)
	}

	return flagSet.Arg(0)
}

func main() {
	// TODO: use cobra to parse flags and args.
	fh, err := os.Open(getArgOrExit())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer fh.Close()

	m, err := layupv1.ParseHCL(fh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	b, err := protojson.Marshal(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(b))
}
