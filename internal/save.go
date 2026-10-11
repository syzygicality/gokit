package internal

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gokit/internal/database"
	"gokit/internal/database/gen"
)

func HandleSave(saveCmd *flag.FlagSet) error {
	saveCmd.Usage = func() { SaveUsage(saveCmd) }

	var get bool
	saveCmd.BoolVar(&get, "get", false, "install the repo with go get after saving")
	saveCmd.BoolVar(&get, "g", false, "shorthand for --get")

	saveCmd.Parse(flag.Args()[1:])

	if saveCmd.NArg() != 2 {
		saveCmd.Usage()
		return errors.New("save requires exactly 2 arguments: <alias> <repo_url>")
	}

	alias := saveCmd.Arg(0)
	repoUrl := saveCmd.Arg(1)

	err := database.WithDB(func(db *gen.Queries) error {
		return db.UpsertAlias(context.Background(), gen.UpsertAliasParams{
			Alias:   alias,
			RepoUrl: repoUrl,
		})
	})

	if err != nil {
		return fmt.Errorf("Failed to save alias %s: %w", alias, err)
	}

	fmt.Printf("Saved %s -> %s\n", alias, repoUrl)

	if get {
		return goGet(repoUrl)
	}

	return nil
}
