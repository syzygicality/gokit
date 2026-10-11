package internal

import (
	"context"
	"flag"
	"fmt"
	"gokit/internal/database"
	"gokit/internal/database/gen"
	"os"
	"slices"
	"text/tabwriter"
)

func HandleList(listCmd *flag.FlagSet) error {
	listCmd.Usage = func() { ListUsage(listCmd) }

	sortBy := listCmd.String("sort", "alias", "column to sort by: alias, updated, created")
	direction := listCmd.String("dir", "desc", "sort direction: asc, desc")

	listCmd.Parse(flag.Args()[1:])

	if *direction != "asc" && *direction != "desc" {
		listCmd.Usage()
		return fmt.Errorf("Invalid sort direction %q, expected asc or desc", *direction)
	}

	var aliases []gen.Alias

	err := database.WithDB(func(db *gen.Queries) error {
		var err error
		ctx := context.Background()

		switch *sortBy {
		case "alias":
			aliases, err = db.ListAliasesByAlias(ctx)
		case "updated":
			aliases, err = db.ListAliasesByRecency(ctx)
		case "created":
			aliases, err = db.ListAliasesByCreation(ctx)
		default:
			listCmd.Usage()
			return fmt.Errorf("Invalid sort column %q, expected alias, updated or created", *sortBy)
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

	if *direction == "asc" {
		slices.Reverse(aliases)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ALIAS\tREPO URL\tUPDATED\tCREATED")
	for _, a := range aliases {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Alias, a.RepoUrl, a.UpdatedAt, a.CreatedAt)
	}

	return w.Flush()
}
