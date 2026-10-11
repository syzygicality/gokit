package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gokit/internal/database"
	"gokit/internal/database/gen"
	"os"
	"os/exec"
	"strings"
)

func InstallPkg(alias string) error {
	if alias == "" {
		return errors.New("No alias given")
	}

	var repoUrl string

	err := database.WithDB(func(db *gen.Queries) error {
		row, err := db.GetAlias(context.Background(), alias)
		if err != nil {
			return err
		}
		repoUrl = row.RepoUrl
		return nil
	})

	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("Unknown alias %s", alias)
	}
	if err != nil {
		return fmt.Errorf("Failed to look up alias %s: %w", alias, err)
	}

	return goGet(repoUrl)
}

func goGet(repoUrl string) error {
	// go get expects a module path, not a URL
	pkg := strings.TrimPrefix(strings.TrimPrefix(repoUrl, "https://"), "http://")
	pkg = strings.TrimSuffix(strings.TrimSuffix(pkg, "/"), ".git")

	fmt.Printf("getting %s...\n", pkg)

	cmd := exec.Command("go", "get", pkg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go get %s failed: %w", pkg, err)
	}

	return nil
}
