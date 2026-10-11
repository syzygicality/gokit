package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/syzygicality/gokit/internal/database/gen"
)

func TestResolveDatabasePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := resolveDatabasePath()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".gokit", "aliases"); path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
}

func TestWithDBCreatesDatabase(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	err := WithDB(func(db *gen.Queries) error {
		aliases, err := db.ListAliasesByAlias(context.Background())
		if len(aliases) != 0 {
			t.Errorf("new database has aliases: %v", aliases)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".gokit", "aliases")); err != nil {
		t.Errorf("database file not created: %v", err)
	}
}

func TestWithDBPersists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()

	err := WithDB(func(db *gen.Queries) error {
		return db.UpsertAlias(ctx, gen.UpsertAliasParams{Alias: "a", RepoUrl: "github.com/a/a"})
	})
	if err != nil {
		t.Fatal(err)
	}

	err = WithDB(func(db *gen.Queries) error {
		row, err := db.GetAlias(ctx, "a")
		if err != nil {
			return err
		}
		if row.RepoUrl != "github.com/a/a" {
			t.Errorf("RepoUrl = %q", row.RepoUrl)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithDBReturnsCallbackError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	sentinel := errors.New("sentinel")
	if err := WithDB(func(*gen.Queries) error { return sentinel }); err != sentinel {
		t.Errorf("err = %v, want sentinel", err)
	}
}

func TestWithDBUnwritableHome(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	home := t.TempDir()
	if err := os.Chmod(home, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(home, 0o700) })
	t.Setenv("HOME", home)

	called := false
	err := WithDB(func(*gen.Queries) error { called = true; return nil })
	want := "Failed to create directory " + filepath.Join(home, ".gokit") + ", please check your file permissions"
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
	if called {
		t.Error("callback ran without a database")
	}
}

func TestQueries(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()

	err := WithDB(func(db *gen.Queries) error {
		for _, a := range []string{"b", "a", "c"} {
			if err := db.UpsertAlias(ctx, gen.UpsertAliasParams{Alias: a, RepoUrl: "u/" + a}); err != nil {
				return err
			}
		}

		order := func(rows []gen.Alias) string {
			s := ""
			for _, r := range rows {
				s += r.Alias
			}
			return s
		}

		// All list queries return newest/largest first; the CLI reverses for asc.
		byAlias, err := db.ListAliasesByAlias(ctx)
		if err != nil {
			return err
		}
		if got := order(byAlias); got != "cba" {
			t.Errorf("ListAliasesByAlias = %s, want cba", got)
		}
		byCreation, err := db.ListAliasesByCreation(ctx)
		if err != nil {
			return err
		}
		if got := order(byCreation); got != "cab" {
			t.Errorf("ListAliasesByCreation = %s, want cab", got)
		}
		byRecency, err := db.ListAliasesByRecency(ctx)
		if err != nil {
			return err
		}
		if got := order(byRecency); got != "cab" {
			t.Errorf("ListAliasesByRecency = %s, want cab", got)
		}

		n, err := db.DeleteAlias(ctx, "a")
		if err != nil {
			return err
		}
		if n != 1 {
			t.Errorf("DeleteAlias rows = %d, want 1", n)
		}
		n, err = db.DeleteAlias(ctx, "a")
		if err != nil {
			return err
		}
		if n != 0 {
			t.Errorf("second DeleteAlias rows = %d, want 0", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
