package internal

import (
	"strings"
	"testing"
)

func TestTrimMillis(t *testing.T) {
	tests := map[string]string{
		"2026-10-11 21:06:02.123": "2026-10-11 21:06:02",
		"2026-10-11 21:06:02":     "2026-10-11 21:06:02",
		"":                        "",
		"a.b.c":                   "a",
	}
	for in, want := range tests {
		if got := trimMillis(in); got != want {
			t.Errorf("trimMillis(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHandleListEmpty(t *testing.T) {
	isolateHome(t)
	out, _, err := runHandler(t, HandleList, "list")
	if err != nil {
		t.Fatal(err)
	}
	if out != "No aliases saved.\n" {
		t.Errorf("stdout = %q", out)
	}
}

func TestHandleListOrder(t *testing.T) {
	isolateHome(t)
	captureStdout(t, func() {
		for _, a := range []string{"b", "a", "c"} {
			if err := saveAlias(a, "github.com/x/"+a); err != nil {
				t.Fatal(err)
			}
		}
	})

	tests := []struct {
		args []string
		want string
	}{
		{nil, "a b c"},
		{[]string{"-dir", "desc"}, "c b a"},
		{[]string{"-sort", "created"}, "c a b"},
		{[]string{"-s", "updated", "-d", "asc"}, "b a c"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out, _, err := runHandler(t, HandleList, "list", tt.args...)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
			var got []string
			for _, l := range lines[1:] {
				got = append(got, strings.Fields(l)[0])
			}
			if strings.Join(got, " ") != tt.want {
				t.Errorf("order = %v, want %s\n%s", got, tt.want, out)
			}
		})
	}
}

func TestHandleListInvalid(t *testing.T) {
	isolateHome(t)
	tests := []struct {
		args []string
		err  string
	}{
		{[]string{"x"}, "list takes no arguments"},
		{[]string{"-s", "repo"}, `Invalid sort column "repo", expected alias, updated or created`},
		{[]string{"-d", "ASC"}, `Invalid sort direction "ASC", expected asc or desc`},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, usage, err := runHandler(t, HandleList, "list", tt.args...)
			if err == nil || err.Error() != tt.err {
				t.Errorf("err = %v, want %q", err, tt.err)
			}
			if !strings.HasPrefix(usage, "Usage: gokit list [flags]\n") {
				t.Errorf("usage = %q", usage)
			}
		})
	}
}
