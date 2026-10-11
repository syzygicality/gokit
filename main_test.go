package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// binPath is the gokit binary built once for all end-to-end tests.
var binPath string

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		// The fake go toolchain below is a shell script.
		fmt.Println("skipping end-to-end tests on windows")
		os.Exit(0)
	}

	dir, err := os.MkdirTemp("", "gokit-e2e-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	binPath = filepath.Join(dir, "gokit")
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to build gokit:", err)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// env is an isolated environment: its own home directory (and so its own alias
// database) and a fake go toolchain that records its arguments instead of running.
type env struct {
	t     *testing.T
	home  string
	goLog string
	// goExit is the exit code of the fake go command.
	goExit int
}

type result struct {
	stdout, stderr string
	code           int
}

const fakeGo = `#!/bin/sh
echo "$@" >> "$GOKIT_TEST_GO_LOG"
exit "${GOKIT_TEST_GO_EXIT:-0}"
`

func newEnv(t *testing.T) *env {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	bin := filepath.Join(root, "bin")
	for _, d := range []string{home, bin} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(fakeGo), 0o755); err != nil {
		t.Fatal(err)
	}
	return &env{t: t, home: home, goLog: filepath.Join(root, "go.log")}
}

func (e *env) run(args ...string) result {
	e.t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(),
		"HOME="+e.home,
		"PATH="+filepath.Join(filepath.Dir(e.home), "bin")+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GOKIT_TEST_GO_LOG="+e.goLog,
		fmt.Sprintf("GOKIT_TEST_GO_EXIT=%d", e.goExit),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	code := 0
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		code = exitErr.ExitCode()
	} else if err != nil {
		e.t.Fatalf("running gokit %v: %v", args, err)
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

// mustRun runs gokit and fails the test unless it exits 0.
func (e *env) mustRun(args ...string) result {
	e.t.Helper()
	r := e.run(args...)
	if r.code != 0 {
		e.t.Fatalf("gokit %v exited %d\nstdout:\n%s\nstderr:\n%s", args, r.code, r.stdout, r.stderr)
	}
	return r
}

// goCalls returns the argument lines the fake go command was invoked with.
func (e *env) goCalls() []string {
	e.t.Helper()
	data, err := os.ReadFile(e.goLog)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		e.t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
}

func expectError(t *testing.T, r result, msg string) {
	t.Helper()
	if r.code != 2 {
		t.Errorf("exit code = %d, want 2\nstderr:\n%s", r.code, r.stderr)
	}
	if want := "Error: " + msg + "\n"; !strings.HasSuffix(r.stderr, want) {
		t.Errorf("stderr does not end with %q\nstderr:\n%s", want, r.stderr)
	}
}

// listedAliases returns the alias column of `gokit list` output, in order.
func listedAliases(t *testing.T, out string) []string {
	t.Helper()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "ALIAS") {
		t.Fatalf("list output has no header:\n%s", out)
	}
	var aliases []string
	for _, line := range lines[1:] {
		aliases = append(aliases, strings.Fields(line)[0])
	}
	return aliases
}

func TestHelp(t *testing.T) {
	e := newEnv(t)
	r := e.mustRun("help")
	if r.stdout != "" {
		t.Errorf("stdout = %q, want empty", r.stdout)
	}
	for _, want := range []string{
		"Usage: gokit <command/alias> [flags]",
		"  list      list saved aliases",
		"  save      save a package under an alias",
		"  remove    remove a saved alias",
		"  search    search pkg.go.dev for packages",
		"  help      show this help",
		"Run gokit <command> -h for a command's flags.",
	} {
		if !strings.Contains(r.stderr, want) {
			t.Errorf("help output missing %q\n%s", want, r.stderr)
		}
	}
}

func TestMainHelpFlag(t *testing.T) {
	e := newEnv(t)
	r := e.mustRun("-h")
	if !strings.Contains(r.stderr, "Usage: gokit <command/alias> [flags]") {
		t.Errorf("-h output missing usage:\n%s", r.stderr)
	}
}

func TestNoCommand(t *testing.T) {
	e := newEnv(t)
	r := e.run()
	expectError(t, r, "No command or alias given.")
	if !strings.Contains(r.stderr, "Usage: gokit <command/alias> [flags]") {
		t.Errorf("missing usage:\n%s", r.stderr)
	}
}

func TestUnknownGlobalFlag(t *testing.T) {
	e := newEnv(t)
	r := e.run("-nope")
	if r.code != 2 {
		t.Errorf("exit code = %d, want 2", r.code)
	}
	if !strings.Contains(r.stderr, "flag provided but not defined: -nope") {
		t.Errorf("stderr:\n%s", r.stderr)
	}
}

func TestSubcommandHelp(t *testing.T) {
	tests := map[string][]string{
		"list":   {"Usage: gokit list [flags]", "-sort", "-dir"},
		"save":   {"Usage: gokit save [flags] <alias> <repo_url>", "-get"},
		"search": {"Usage: gokit search [flags] <query>", "-limit", "-top", "-alias", "-get", "-symbol", "-filter", "-token"},
		"remove": {"Usage: gokit remove <alias>"},
	}
	for cmd, wants := range tests {
		t.Run(cmd, func(t *testing.T) {
			e := newEnv(t)
			r := e.mustRun(cmd, "-h")
			for _, want := range wants {
				if !strings.Contains(r.stderr, want) {
					t.Errorf("missing %q in:\n%s", want, r.stderr)
				}
			}
		})
	}
}

func TestUnknownSubcommandFlag(t *testing.T) {
	e := newEnv(t)
	r := e.run("list", "-nope")
	if r.code != 2 {
		t.Errorf("exit code = %d, want 2", r.code)
	}
	if !strings.Contains(r.stderr, "flag provided but not defined: -nope") {
		t.Errorf("stderr:\n%s", r.stderr)
	}
}

func TestListEmpty(t *testing.T) {
	e := newEnv(t)
	r := e.mustRun("list")
	if r.stdout != "No aliases saved.\n" {
		t.Errorf("stdout = %q", r.stdout)
	}
}

func TestDatabaseLocation(t *testing.T) {
	e := newEnv(t)
	e.mustRun("list")
	if _, err := os.Stat(filepath.Join(e.home, ".gokit", "aliases")); err != nil {
		t.Errorf("database not created in home directory: %v", err)
	}
}

func TestSaveAndList(t *testing.T) {
	e := newEnv(t)
	r := e.mustRun("save", "chi", "github.com/go-chi/chi/v5")
	if r.stdout != "Saved chi -> github.com/go-chi/chi/v5\n" {
		t.Errorf("save stdout = %q", r.stdout)
	}
	if calls := e.goCalls(); calls != nil {
		t.Errorf("save without -get ran go: %v", calls)
	}

	r = e.mustRun("list")
	lines := strings.Split(strings.TrimRight(r.stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("list output:\n%s", r.stdout)
	}
	header := regexp.MustCompile(`^ALIAS +REPO URL +UPDATED +CREATED$`)
	if !header.MatchString(lines[0]) {
		t.Errorf("header = %q", lines[0])
	}
	// Timestamps are shown to the second, without the stored milliseconds.
	ts := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`
	row := regexp.MustCompile(`^chi +github\.com/go-chi/chi/v5 +` + ts + ` +` + ts + `$`)
	if !row.MatchString(lines[1]) {
		t.Errorf("row = %q", lines[1])
	}
	// Columns are aligned with tabwriter.
	if strings.Index(lines[0], "REPO URL") != strings.Index(lines[1], "github.com") {
		t.Errorf("columns not aligned:\n%s", r.stdout)
	}
}

func TestSaveOverwrites(t *testing.T) {
	e := newEnv(t)
	e.mustRun("save", "x", "github.com/a/old")
	created := strings.Fields(strings.Split(e.mustRun("list").stdout, "\n")[1])[4:6]

	r := e.mustRun("save", "x", "github.com/a/new")
	if r.stdout != "Saved x -> github.com/a/new\n" {
		t.Errorf("stdout = %q", r.stdout)
	}

	out := e.mustRun("list").stdout
	fields := strings.Fields(strings.Split(out, "\n")[1])
	if fields[1] != "github.com/a/new" {
		t.Errorf("repo url = %q, want github.com/a/new", fields[1])
	}
	if got := fields[4:6]; got[0] != created[0] || got[1] != created[1] {
		t.Errorf("created changed on overwrite: %v -> %v", created, got)
	}
	if got := listedAliases(t, out); len(got) != 1 {
		t.Errorf("aliases = %v, want one", got)
	}
}

func TestSaveWithGet(t *testing.T) {
	for _, flag := range []string{"-get", "-g"} {
		t.Run(flag, func(t *testing.T) {
			e := newEnv(t)
			r := e.mustRun("save", flag, "gin", "https://github.com/gin-gonic/gin.git")
			want := "Saved gin -> https://github.com/gin-gonic/gin.git\ngetting github.com/gin-gonic/gin...\n"
			if r.stdout != want {
				t.Errorf("stdout = %q, want %q", r.stdout, want)
			}
			if calls := e.goCalls(); len(calls) != 1 || calls[0] != "get github.com/gin-gonic/gin" {
				t.Errorf("go calls = %q", calls)
			}
		})
	}
}

func TestSaveFlagsAfterArgsAreArgs(t *testing.T) {
	// Flags must come before positional arguments.
	e := newEnv(t)
	r := e.run("save", "gin", "github.com/gin-gonic/gin", "-g")
	expectError(t, r, "save requires exactly 2 arguments: <alias> <repo_url>")
}

func TestSaveInvalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
		msg  string
	}{
		{"no args", []string{"save"}, "save requires exactly 2 arguments: <alias> <repo_url>"},
		{"one arg", []string{"save", "a"}, "save requires exactly 2 arguments: <alias> <repo_url>"},
		{"three args", []string{"save", "a", "b", "c"}, "save requires exactly 2 arguments: <alias> <repo_url>"},
		{"empty alias", []string{"save", "", "github.com/a/b"}, "Alias must not be empty"},
		{"dash alias", []string{"save", "--", "-x", "github.com/a/b"}, `Invalid alias "-x", aliases cannot start with -`},
		{"space alias", []string{"save", "a b", "github.com/a/b"}, `Invalid alias "a b", aliases cannot contain whitespace`},
		{"long alias", []string{"save", strings.Repeat("a", 33), "github.com/a/b"}, fmt.Sprintf("Invalid alias %q, aliases can be at most 32 characters", strings.Repeat("a", 33))},
		{"command alias", []string{"save", "list", "github.com/a/b"}, `Invalid alias "list", it is a gokit command`},
		{"empty url", []string{"save", "a", "  "}, "Repo URL must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			expectError(t, e.run(tt.args...), tt.msg)
			if out := e.mustRun("list").stdout; out != "No aliases saved.\n" {
				t.Errorf("invalid save stored an alias:\n%s", out)
			}
		})
	}
}

func TestSaveMaxLengthAlias(t *testing.T) {
	e := newEnv(t)
	alias := strings.Repeat("a", 32)
	e.mustRun("save", alias, "github.com/a/b")
	if got := listedAliases(t, e.mustRun("list").stdout); len(got) != 1 || got[0] != alias {
		t.Errorf("aliases = %v", got)
	}
}

func TestListSorting(t *testing.T) {
	e := newEnv(t)
	// Created order: b, a, c. Then b is updated, making it the most recently updated.
	e.mustRun("save", "b", "github.com/x/b")
	e.mustRun("save", "a", "github.com/x/a")
	e.mustRun("save", "c", "github.com/x/c")
	e.mustRun("save", "b", "github.com/x/b2")

	tests := []struct {
		args []string
		want string
	}{
		{nil, "a b c"},
		{[]string{"-s", "alias"}, "a b c"},
		{[]string{"-sort", "alias", "-dir", "asc"}, "a b c"},
		{[]string{"-d", "desc"}, "c b a"},
		{[]string{"-s", "created"}, "c a b"},
		{[]string{"-s", "created", "-d", "desc"}, "c a b"},
		{[]string{"-sort", "created", "-dir", "asc"}, "b a c"},
		{[]string{"-s", "updated"}, "b c a"},
		{[]string{"-s", "updated", "-d", "asc"}, "a c b"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out := e.mustRun(append([]string{"list"}, tt.args...)...).stdout
			if got := strings.Join(listedAliases(t, out), " "); got != tt.want {
				t.Errorf("order = %q, want %q\n%s", got, tt.want, out)
			}
		})
	}
}

func TestListInvalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
		msg  string
	}{
		{"positional arg", []string{"list", "x"}, "list takes no arguments"},
		{"bad sort", []string{"list", "-s", "name"}, `Invalid sort column "name", expected alias, updated or created`},
		{"bad dir", []string{"list", "-d", "up"}, `Invalid sort direction "up", expected asc or desc`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			r := e.run(tt.args...)
			expectError(t, r, tt.msg)
			if !strings.Contains(r.stderr, "Usage: gokit list [flags]") {
				t.Errorf("missing usage:\n%s", r.stderr)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	e := newEnv(t)
	e.mustRun("save", "a", "github.com/x/a")
	e.mustRun("save", "b", "github.com/x/b")

	r := e.mustRun("remove", "a")
	if r.stdout != "Removed a\n" {
		t.Errorf("stdout = %q", r.stdout)
	}
	if got := strings.Join(listedAliases(t, e.mustRun("list").stdout), " "); got != "b" {
		t.Errorf("aliases after remove = %q, want b", got)
	}

	expectError(t, e.run("remove", "a"), "Unknown alias a")
}

func TestRemoveInvalid(t *testing.T) {
	for _, args := range [][]string{{"remove"}, {"remove", "a", "b"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			e := newEnv(t)
			r := e.run(args...)
			expectError(t, r, "remove requires exactly 1 argument: <alias>")
			if !strings.Contains(r.stderr, "Usage: gokit remove <alias>") {
				t.Errorf("missing usage:\n%s", r.stderr)
			}
		})
	}
}

func TestInstallAlias(t *testing.T) {
	tests := []struct {
		url, pkg string
	}{
		{"github.com/spf13/cobra", "github.com/spf13/cobra"},
		{"https://github.com/spf13/cobra", "github.com/spf13/cobra"},
		{"http://github.com/spf13/cobra", "github.com/spf13/cobra"},
		{"https://github.com/spf13/cobra/", "github.com/spf13/cobra"},
		{"https://github.com/spf13/cobra.git", "github.com/spf13/cobra"},
		{"github.com/go-chi/chi/v5", "github.com/go-chi/chi/v5"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			e := newEnv(t)
			e.mustRun("save", "pkg", tt.url)
			r := e.mustRun("pkg")
			if want := "getting " + tt.pkg + "...\n"; r.stdout != want {
				t.Errorf("stdout = %q, want %q", r.stdout, want)
			}
			if calls := e.goCalls(); len(calls) != 1 || calls[0] != "get "+tt.pkg {
				t.Errorf("go calls = %q, want [get %s]", calls, tt.pkg)
			}
		})
	}
}

func TestInstallUnknownAlias(t *testing.T) {
	e := newEnv(t)
	expectError(t, e.run("nope"), "Unknown alias nope")
	if calls := e.goCalls(); calls != nil {
		t.Errorf("go ran for unknown alias: %v", calls)
	}
}

func TestInstallGoGetFails(t *testing.T) {
	e := newEnv(t)
	e.mustRun("save", "pkg", "github.com/a/b")
	e.goExit = 1
	expectError(t, e.run("pkg"), "go get github.com/a/b failed: exit status 1")
}

func TestSearchInvalidBeforeNetwork(t *testing.T) {
	tests := []struct {
		name string
		args []string
		msg  string
	}{
		{"no query", []string{"search"}, "search requires a query"},
		{"flags only", []string{"search", "-top"}, "search requires a query"},
		{"bad alias", []string{"search", "-a", "help", "cobra"}, `Invalid alias "help", it is a gokit command`},
		{"negative limit", []string{"search", "-limit", "-1", "cobra"}, "Invalid limit -1, expected 0 or more"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			expectError(t, e.run(tt.args...), tt.msg)
		})
	}
}
