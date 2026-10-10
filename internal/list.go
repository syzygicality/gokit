package internal

import "flag"


func HandleList(listCmd *flag.FlagSet) error {
	listCmd.Usage = ListUsage
	
	return nil
}