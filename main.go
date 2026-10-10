package main

import (
	"flag"
	"os"

	"gokit/internal"
)

func main() {
	flag.Usage = internal.Usage

	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	upsertCmd := flag.NewFlagSet("upsert", flag.ExitOnError)
	removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)

	flag.Parse()

	switch flag.Arg(0) {
	case helpCmd.Name():
		flag.Usage()
		os.Exit(0)
	case listCmd.Name():

	case upsertCmd.Name():

	case removeCmd.Name():

	case searchCmd.Name():
		
	default:
		flag.Usage()
		os.Exit(2)
	}
}
