package internal

import "flag"


func HandleSearch(searchCmd *flag.FlagSet) error {
	searchCmd.Usage = SearchUsage
	
	return nil
}