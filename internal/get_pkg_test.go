package internal

import (
	"testing"
)

func TestGoGetStripsURL(t *testing.T) {
	tests := []struct {
		url, pkg string
	}{
		{"github.com/a/b", "github.com/a/b"},
		{"https://github.com/a/b", "github.com/a/b"},
		{"http://github.com/a/b", "github.com/a/b"},
		{"github.com/a/b/", "github.com/a/b"},
		{"github.com/a/b.git", "github.com/a/b"},
		{"https://github.com/a/b.git", "github.com/a/b"},
		{"https://github.com/a/b/v2", "github.com/a/b/v2"},
		// Only one of each prefix and suffix is stripped, trailing slash before .git.
		{"https://github.com/a/b.git/", "github.com/a/b"},
		{"https://github.com/a/b//", "github.com/a/b/"},
		{"https://http://x", "x"},
		{"http://https://x", "https://x"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			calls := fakeGo(t, 0)
			var err error
			out := captureStdout(t, func() { err = goGet(tt.url) })
			if err != nil {
				t.Fatal(err)
			}
			if want := "getting " + tt.pkg + "...\n"; out != want {
				t.Errorf("stdout = %q, want %q", out, want)
			}
			if got := calls(); len(got) != 1 || got[0] != "get "+tt.pkg {
				t.Errorf("go calls = %q, want [get %s]", got, tt.pkg)
			}
		})
	}
}

func TestGoGetFailure(t *testing.T) {
	fakeGo(t, 3)
	var err error
	captureStdout(t, func() { err = goGet("github.com/a/b") })
	if err == nil || err.Error() != "go get github.com/a/b failed: exit status 3" {
		t.Errorf("err = %v", err)
	}
}

func TestInstallPkg(t *testing.T) {
	isolateHome(t)
	calls := fakeGo(t, 0)
	captureStdout(t, func() {
		if err := saveAlias("b", "https://github.com/a/b.git"); err != nil {
			t.Fatal(err)
		}
	})

	var err error
	out := captureStdout(t, func() { err = InstallPkg("b") })
	if err != nil {
		t.Fatal(err)
	}
	if out != "getting github.com/a/b...\n" {
		t.Errorf("stdout = %q", out)
	}
	if got := calls(); len(got) != 1 || got[0] != "get github.com/a/b" {
		t.Errorf("go calls = %q", got)
	}
}

func TestInstallPkgErrors(t *testing.T) {
	isolateHome(t)
	calls := fakeGo(t, 0)

	if err := InstallPkg(""); err == nil || err.Error() != "No alias given" {
		t.Errorf("empty alias err = %v", err)
	}
	if err := InstallPkg("nope"); err == nil || err.Error() != "Unknown alias nope" {
		t.Errorf("unknown alias err = %v", err)
	}
	if got := calls(); got != nil {
		t.Errorf("go ran: %q", got)
	}
}
