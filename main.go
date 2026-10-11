package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gokit/internal"
	"gokit/internal/consts"
)

func main() {
	flag.Usage = internal.MainUsage

	helpCmd := flag.NewFlagSet(consts.HelpCommand, flag.ExitOnError)
	listCmd := flag.NewFlagSet(consts.ListCommand, flag.ExitOnError)
	saveCmd := flag.NewFlagSet(consts.SaveCommand, flag.ExitOnError)
	removeCmd := flag.NewFlagSet(consts.RemoveCommand, flag.ExitOnError)
	searchCmd := flag.NewFlagSet(consts.SearchCommand, flag.ExitOnError)

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
	case "":
		flag.Usage()
		err = errors.New("No command or alias given.")
	default:
		err = internal.InstallPkg(flag.Arg(0))
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(2)
	}
	os.Exit(0)
}
