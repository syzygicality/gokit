package internal

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gokit/internal/database"
	"gokit/internal/database/gen"
)

func HandleRemove(removeCmd *flag.FlagSet) error {
	removeCmd.Usage = func() { RemoveUsage(removeCmd) }

	removeCmd.Parse(flag.Args()[1:])

	if removeCmd.NArg() != 1 {
		removeCmd.Usage()
		return errors.New("remove requires exactly 1 argument: <alias>")
	}

	alias := removeCmd.Arg(0)

	var removed int64

	err := database.WithDB(func(db *gen.Queries) error {
		var err error
		removed, err = db.DeleteAlias(context.Background(), alias)
		return err
	})

	if err != nil {
		return fmt.Errorf("Failed to remove alias %s: %w", alias, err)
	}
	if removed == 0 {
		return fmt.Errorf("Unknown alias %s", alias)
	}

	fmt.Printf("Removed %s\n", alias)

	return nil
}
