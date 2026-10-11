package internal

import (
	"bytes"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// isolateHome points the alias database at a fresh temporary home directory.
func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

// fakeGo puts a go command on PATH that records its arguments and exits with
// exitCode, and returns a function reporting the recorded calls.
func fakeGo(t *testing.T, exitCode int) func() []string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake go toolchain is a shell script")
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "go.log")
	script := "#!/bin/sh\necho \"$@\" >> \"" + log + "\"\nexit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return func() []string {
		data, err := os.ReadFile(log)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			t.Fatal(err)
		}
		return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}
}

// captureStdout runs fn and returns what it wrote to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	defer func() {
		os.Stdout = orig
	}()
	fn()
	w.Close()
	return <-done
}

// runHandler runs a subcommand handler as if invoked with `gokit <name> args...`,
// returning its stdout, usage output and error.
func runHandler(t *testing.T, handler func(*flag.FlagSet) error, name string, args ...string) (stdout, usage string, err error) {
	t.Helper()
	orig := flag.CommandLine
	t.Cleanup(func() { flag.CommandLine = orig })

	flag.CommandLine = flag.NewFlagSet("gokit", flag.ContinueOnError)
	if err := flag.CommandLine.Parse(append([]string{name}, args...)); err != nil {
		t.Fatal(err)
	}

	var usageBuf bytes.Buffer
	cmd := flag.NewFlagSet(name, flag.ContinueOnError)
	cmd.SetOutput(&usageBuf)

	stdout = captureStdout(t, func() { err = handler(cmd) })
	return stdout, usageBuf.String(), err
}

// roundTripFunc adapts a function to http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// stubAPI serves every outgoing HTTP request with handler instead of the network,
// and returns a function reporting the requests made.
func stubAPI(t *testing.T, handler http.HandlerFunc) func() []*http.Request {
	t.Helper()
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })

	var reqs []*http.Request
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		reqs = append(reqs, r)
		rec := httptest.NewRecorder()
		handler(rec, r)
		return rec.Result(), nil
	})
	return func() []*http.Request { return reqs }
}
