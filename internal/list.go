package internal

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"gokit/internal/database"
	"gokit/internal/database/gen"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
)

func HandleList(listCmd *flag.FlagSet) error {
	listCmd.Usage = func() { ListUsage(listCmd) }

	var sortBy, direction string
	listCmd.StringVar(&sortBy, "s", "alias", "shorthand for -sort")
	listCmd.StringVar(&sortBy, "sort", "alias", "column to sort by: alias, updated, created")
	listCmd.StringVar(&direction, "d", "", "shorthand for -dir")
	listCmd.StringVar(&direction, "dir", "", "sort direction: asc, desc (default asc for alias, desc otherwise)")

	listCmd.Parse(flag.Args()[1:])

	if listCmd.NArg() != 0 {
		listCmd.Usage()
		return errors.New("list takes no arguments")
	}

	if sortBy != "alias" && sortBy != "updated" && sortBy != "created" {
		listCmd.Usage()
		return fmt.Errorf("Invalid sort column %q, expected alias, updated or created", sortBy)
	}

	if direction == "" {
		direction = "desc"
		if sortBy == "alias" {
			direction = "asc"
		}
	}

	if direction != "asc" && direction != "desc" {
		listCmd.Usage()
		return fmt.Errorf("Invalid sort direction %q, expected asc or desc", direction)
	}

	var aliases []gen.Alias

	err := database.WithDB(func(db *gen.Queries) error {
		var err error
		ctx := context.Background()

		switch sortBy {
		case "alias":
			aliases, err = db.ListAliasesByAlias(ctx)
		case "updated":
			aliases, err = db.ListAliasesByRecency(ctx)
		case "created":
			aliases, err = db.ListAliasesByCreation(ctx)
		}

		return err
	})

	if err != nil {
		return fmt.Errorf("Failed to list aliases: %w", err)
	}

	if len(aliases) == 0 {
		fmt.Println("No aliases saved.")
		return nil
	}

	if direction == "asc" {
		slices.Reverse(aliases)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ALIAS\tREPO URL\tUPDATED\tCREATED")
	for _, a := range aliases {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Alias, a.RepoUrl, trimMillis(a.UpdatedAt), trimMillis(a.CreatedAt))
	}

	return w.Flush()
}

// trimMillis drops the fractional seconds from a stored timestamp for display.
func trimMillis(timestamp string) string {
	before, _, _ := strings.Cut(timestamp, ".")
	return before
}
