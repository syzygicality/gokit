package internal

import "flag"


func HandleSave(saveCmd *flag.FlagSet) error {
	saveCmd.Usage = SaveUsage

	return nil
}