package main

import (
	"flag"
	"os"

	"gokit/internal"
)

func main() {
	flag.Usage = internal.Usage

	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)

	flag.Parse()

	switch flag.Arg(0) {
	case helpCmd.Name():
		flag.Usage()
		os.Exit(0)
	default:
		flag.Usage()
		os.Exit(2)
	}
}
