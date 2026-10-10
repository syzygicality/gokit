package internal

import (
	"flag"
	"fmt"
)

func MainUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit <command/alias> [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}

func ListUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit list [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}

func SaveUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit save <alias> <repo_url> [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}

func SearchUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit search [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}

func RemoveUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit remove <alias> [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}