package internal

import (
	"flag"
	"fmt"
)

func MainUsage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: gokit <command/alias> [flags]\n\n")
	fmt.Fprintf(out, "Discover Go packages and go get them by short aliases.\n")
	fmt.Fprintf(out, "Running gokit <alias> runs go get on the package saved under that alias.\n\n")
	fmt.Fprintf(out, "Commands:\n")
	fmt.Fprintf(out, "  list      list saved aliases\n")
	fmt.Fprintf(out, "  save      save a package under an alias\n")
	fmt.Fprintf(out, "  remove    remove a saved alias\n")
	fmt.Fprintf(out, "  search    search pkg.go.dev for packages\n")
	fmt.Fprintf(out, "  help      show this help\n\n")
	fmt.Fprintf(out, "Run gokit <command> -h for a command's flags.\n")
}

func ListUsage(listCmd *flag.FlagSet) {
	out := listCmd.Output()
	fmt.Fprintf(out, "Usage: gokit list [flags]\n\n")
	fmt.Fprintf(out, "List saved aliases and the packages they point to.\n\n")
	fmt.Fprintf(out, "Flags:\n")
	listCmd.PrintDefaults()
}

func SaveUsage(saveCmd *flag.FlagSet) {
	out := saveCmd.Output()
	fmt.Fprintf(out, "Usage: gokit save [flags] <alias> <repo_url>\n\n")
	fmt.Fprintf(out, "Save a package under an alias, replacing any package already saved under it.\n\n")
	fmt.Fprintf(out, "Flags:\n")
	saveCmd.PrintDefaults()
}

func SearchUsage(searchCmd *flag.FlagSet) {
	out := searchCmd.Output()
	fmt.Fprintf(out, "Usage: gokit search [flags] <query>\n\n")
	fmt.Fprintf(out, "Search pkg.go.dev for packages matching the query, best match first.\n\n")
	fmt.Fprintf(out, "Flags:\n")
	searchCmd.PrintDefaults()
}

func RemoveUsage(removeCmd *flag.FlagSet) {
	out := removeCmd.Output()
	fmt.Fprintf(out, "Usage: gokit remove <alias>\n\n")
	fmt.Fprintf(out, "Remove a saved alias.\n\n")
}
