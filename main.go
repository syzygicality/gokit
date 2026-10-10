package main

import (
	"errors"
	"flag"
	"os"

	"gokit/internal"
)

func main() {
	flag.Usage = internal.MainUsage

	helpCmd := flag.NewFlagSet("help", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	saveCmd := flag.NewFlagSet("save", flag.ExitOnError)
	removeCmd := flag.NewFlagSet("remove", flag.ExitOnError)
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)

	flag.Parse()

	var err error = nil

	switch flag.Arg(0) {
	case helpCmd.Name():
		flag.Usage()
	case listCmd.Name():
		err = internal.HandleList(listCmd)
	case saveCmd.Name():
		err = internal.HandleSave(saveCmd)
	case removeCmd.Name():
		err = internal.HandleRemove(removeCmd)
	case searchCmd.Name():
		err = internal.HandleSearch(searchCmd)
	default:
		err = internal.InstallPkg(flag.Arg(0))
		if err != nil {
			flag.Usage()
			err = errors.New("Invalid command.")
		}
	}

	if err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}
