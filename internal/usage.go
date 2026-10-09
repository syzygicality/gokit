package internal

import (
	"flag"
	"fmt"
)

// Usage prints the command-line usage message.
func Usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit <command> [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}
