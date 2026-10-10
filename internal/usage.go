package internal

import (
	"flag"
	"fmt"
)

func MainUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit <command> [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}
