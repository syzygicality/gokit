package internal

import "flag"


func HandleRemove(removeCmd *flag.FlagSet) error {
	removeCmd.Usage = RemoveUsage
	
	return nil
}