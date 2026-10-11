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

func ListUsage(listCmd *flag.FlagSet) {
	out := listCmd.Output()
	fmt.Fprintf(out, "Usage: gokit list [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	listCmd.PrintDefaults()
}

func SaveUsage(saveCmd *flag.FlagSet) {
	out := saveCmd.Output()
	fmt.Fprintf(out, "Usage: gokit save [flags] <alias> <repo_url>\n\n")
	fmt.Fprintf(out, "Flags:\n")
	saveCmd.PrintDefaults()
}

func SearchUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit search [flags]\n\n")
	fmt.Fprintf(out, "Flags:\n")
	flag.PrintDefaults()
}

func RemoveUsage(removeCmd *flag.FlagSet) {
	out := removeCmd.Output()
	fmt.Fprintf(out, "Usage: gokit remove [flags] <alias>\n\n")
	fmt.Fprintf(out, "Flags:\n")
	removeCmd.PrintDefaults()
}