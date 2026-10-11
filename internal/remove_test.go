package internal

import (
	"strings"
	"testing"
)

func TestHandleRemove(t *testing.T) {
	isolateHome(t)
	captureStdout(t, func() {
		if err := saveAlias("x", "github.com/a/b"); err != nil {
			t.Fatal(err)
		}
	})

	out, _, err := runHandler(t, HandleRemove, "remove", "x")
	if err != nil {
		t.Fatal(err)
	}
	if out != "Removed x\n" {
		t.Errorf("stdout = %q", out)
	}
	if _, err := getAlias(t, "x"); err == nil {
		t.Error("alias still stored after remove")
	}

	out, _, err = runHandler(t, HandleRemove, "remove", "x")
	if err == nil || err.Error() != "Unknown alias x" {
		t.Errorf("second remove err = %v", err)
	}
	if out != "" {
		t.Errorf("failed remove printed %q", out)
	}
}

func TestHandleRemoveArgs(t *testing.T) {
	isolateHome(t)
	for _, args := range [][]string{nil, {"a", "b"}} {
		_, usage, err := runHandler(t, HandleRemove, "remove", args...)
		if err == nil || err.Error() != "remove requires exactly 1 argument: <alias>" {
			t.Errorf("args %v: err = %v", args, err)
		}
		if !strings.HasPrefix(usage, "Usage: gokit remove <alias>\n") {
			t.Errorf("usage = %q", usage)
		}
	}
}
