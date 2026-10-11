package internal

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/syzygicality/gokit/internal/consts"
	"github.com/syzygicality/gokit/internal/database"
	"github.com/syzygicality/gokit/internal/database/gen"
	"slices"
	"strings"
	"unicode"
)

func HandleSave(saveCmd *flag.FlagSet) error {
	saveCmd.Usage = func() { SaveUsage(saveCmd) }

	var get bool
	saveCmd.BoolVar(&get, "get", false, "install the repo with go get after saving")
	saveCmd.BoolVar(&get, "g", false, "shorthand for -get")

	saveCmd.Parse(flag.Args()[1:])

	if saveCmd.NArg() != 2 {
		saveCmd.Usage()
		return errors.New("save requires exactly 2 arguments: <alias> <repo_url>")
	}

	alias := saveCmd.Arg(0)
	repoUrl := saveCmd.Arg(1)

	if err := saveAlias(alias, repoUrl); err != nil {
		return err
	}

	if get {
		return goGet(repoUrl)
	}

	return nil
}

func validateAlias(alias string) error {
	switch {
	case alias == "":
		return errors.New("Alias must not be empty")
	case strings.HasPrefix(alias, "-"):
		return fmt.Errorf("Invalid alias %q, aliases cannot start with -", alias)
	case strings.ContainsFunc(alias, unicode.IsSpace):
		return fmt.Errorf("Invalid alias %q, aliases cannot contain whitespace", alias)
	case len(alias) > consts.AliasMaxLength:
		return fmt.Errorf("Invalid alias %q, aliases can be at most %d characters", alias, consts.AliasMaxLength)
	case slices.Contains(consts.Commands, alias):
		return fmt.Errorf("Invalid alias %q, it is a gokit command", alias)
	}
	return nil
}

func saveAlias(alias string, repoUrl string) error {
	if err := validateAlias(alias); err != nil {
		return err
	}
	if strings.TrimSpace(repoUrl) == "" {
		return errors.New("Repo URL must not be empty")
	}

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

	return nil
}
