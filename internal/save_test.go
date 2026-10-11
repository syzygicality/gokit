package internal

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/syzygicality/gokit/internal/consts"
	"github.com/syzygicality/gokit/internal/database"
	"github.com/syzygicality/gokit/internal/database/gen"
)

func TestValidateAlias(t *testing.T) {
	tests := []struct {
		alias string
		err   string
	}{
		{"cobra", ""},
		{"go-chi", ""},
		{"a.b_c/d", ""},
		{"ünïcode", ""},
		{strings.Repeat("a", consts.AliasMaxLength), ""},
		{"", "Alias must not be empty"},
		{"-x", `Invalid alias "-x", aliases cannot start with -`},
		{"a b", `Invalid alias "a b", aliases cannot contain whitespace`},
		{"a\tb", `Invalid alias "a\tb", aliases cannot contain whitespace`},
		{"a\n", `Invalid alias "a\n", aliases cannot contain whitespace`},
		{strings.Repeat("a", consts.AliasMaxLength+1), `aliases can be at most 32 characters`},
		{"help", `Invalid alias "help", it is a gokit command`},
		{"list", `Invalid alias "list", it is a gokit command`},
		{"save", `Invalid alias "save", it is a gokit command`},
		{"remove", `Invalid alias "remove", it is a gokit command`},
		{"search", `Invalid alias "search", it is a gokit command`},
		// Commands are matched exactly.
		{"List", ""},
		{"lists", ""},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			err := validateAlias(tt.alias)
			if tt.err == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Errorf("error = %v, want %q", err, tt.err)
			}
		})
	}
}

func TestValidateAliasLengthIsBytes(t *testing.T) {
	// 11 three-byte runes is 33 bytes, over the limit even though it is 11 characters.
	if err := validateAlias(strings.Repeat("あ", 11)); err == nil {
		t.Error("expected multi-byte alias over 32 bytes to be rejected")
	}
}

func TestCommandsListsEveryCommand(t *testing.T) {
	want := []string{consts.HelpCommand, consts.ListCommand, consts.SaveCommand, consts.RemoveCommand, consts.SearchCommand}
	if strings.Join(consts.Commands, ",") != strings.Join(want, ",") {
		t.Errorf("Commands = %v, want %v", consts.Commands, want)
	}
}

func getAlias(t *testing.T, alias string) (gen.Alias, error) {
	t.Helper()
	var row gen.Alias
	err := database.WithDB(func(db *gen.Queries) error {
		var err error
		row, err = db.GetAlias(context.Background(), alias)
		return err
	})
	return row, err
}

func TestSaveAlias(t *testing.T) {
	isolateHome(t)

	var err error
	out := captureStdout(t, func() { err = saveAlias("chi", "github.com/go-chi/chi/v5") })
	if err != nil {
		t.Fatal(err)
	}
	if out != "Saved chi -> github.com/go-chi/chi/v5\n" {
		t.Errorf("stdout = %q", out)
	}

	row, err := getAlias(t, "chi")
	if err != nil {
		t.Fatal(err)
	}
	// The URL is stored as given; stripping happens at go get time.
	if row.RepoUrl != "github.com/go-chi/chi/v5" {
		t.Errorf("RepoUrl = %q", row.RepoUrl)
	}
	if row.CreatedAt != row.UpdatedAt {
		t.Errorf("new alias has CreatedAt %q != UpdatedAt %q", row.CreatedAt, row.UpdatedAt)
	}
	// Timestamps keep millisecond precision.
	if len(row.CreatedAt) != len("2006-01-02 15:04:05.000") {
		t.Errorf("CreatedAt = %q, want millisecond precision", row.CreatedAt)
	}
}

func TestSaveAliasUpsert(t *testing.T) {
	isolateHome(t)
	captureStdout(t, func() {
		if err := saveAlias("x", "github.com/a/old"); err != nil {
			t.Fatal(err)
		}
	})
	before, err := getAlias(t, "x")
	if err != nil {
		t.Fatal(err)
	}

	// Make sure the update lands in a later millisecond.
	time.Sleep(5 * time.Millisecond)
	captureStdout(t, func() {
		if err := saveAlias("x", "github.com/a/new"); err != nil {
			t.Fatal(err)
		}
	})
	after, err := getAlias(t, "x")
	if err != nil {
		t.Fatal(err)
	}

	if after.RepoUrl != "github.com/a/new" {
		t.Errorf("RepoUrl = %q", after.RepoUrl)
	}
	if after.CreatedAt != before.CreatedAt {
		t.Errorf("CreatedAt changed: %q -> %q", before.CreatedAt, after.CreatedAt)
	}
	if after.UpdatedAt <= before.UpdatedAt {
		t.Errorf("UpdatedAt did not advance: %q -> %q", before.UpdatedAt, after.UpdatedAt)
	}
}

func TestSaveAliasRejects(t *testing.T) {
	isolateHome(t)
	tests := []struct {
		alias, url, err string
	}{
		{"help", "github.com/a/b", "it is a gokit command"},
		{"ok", "", "Repo URL must not be empty"},
		{"ok", " \t", "Repo URL must not be empty"},
	}
	for _, tt := range tests {
		out := captureStdout(t, func() {
			err := saveAlias(tt.alias, tt.url)
			if err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Errorf("saveAlias(%q, %q) error = %v, want %q", tt.alias, tt.url, err, tt.err)
			}
		})
		if out != "" {
			t.Errorf("rejected save printed %q", out)
		}
	}
	if _, err := getAlias(t, "ok"); err == nil {
		t.Error("rejected save was stored")
	}
}

func TestHandleSaveUsageOnWrongArgs(t *testing.T) {
	isolateHome(t)
	_, usage, err := runHandler(t, HandleSave, "save", "only-one")
	if err == nil || err.Error() != "save requires exactly 2 arguments: <alias> <repo_url>" {
		t.Errorf("err = %v", err)
	}
	if !strings.HasPrefix(usage, "Usage: gokit save [flags] <alias> <repo_url>\n") {
		t.Errorf("usage = %q", usage)
	}
}

func TestHandleSaveGet(t *testing.T) {
	isolateHome(t)
	calls := fakeGo(t, 0)
	out, _, err := runHandler(t, HandleSave, "save", "-g", "x", "https://github.com/a/b/")
	if err != nil {
		t.Fatal(err)
	}
	if want := "Saved x -> https://github.com/a/b/\ngetting github.com/a/b...\n"; out != want {
		t.Errorf("stdout = %q, want %q", out, want)
	}
	if got := calls(); len(got) != 1 || got[0] != "get github.com/a/b" {
		t.Errorf("go calls = %q", got)
	}
}
