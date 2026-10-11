package internal

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func serveJSON(status int, body any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(body)
	}
}

var twoResults = searchResponse{
	Items: []searchItem{
		{PackagePath: "github.com/julienschmidt/httprouter", ModulePath: "github.com/julienschmidt/httprouter", Version: "v1.3.0", Synopsis: "Package httprouter is a trie based high performance HTTP request router."},
		{PackagePath: "github.com/go-chi/chi/v5/middleware", ModulePath: "github.com/go-chi/chi/v5", Version: "v5.3.2"},
	},
	Total:         2,
	NextPageToken: "cd0d8a5c",
}

var oneResult = searchResponse{
	Items: []searchItem{
		{PackagePath: "github.com/spf13/cobra/doc", ModulePath: "github.com/spf13/cobra", Version: "v1.9.1", Synopsis: "Cobra docs."},
	},
	Total: 1,
}

func TestSearchRequest(t *testing.T) {
	reqs := stubAPI(t, serveJSON(http.StatusOK, oneResult))
	params := url.Values{"q": {"http router"}, "limit": {"5"}}

	res, err := search(params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 1 || res.Items[0] != oneResult.Items[0] || res.Total != 1 {
		t.Errorf("result = %+v", res)
	}

	r := reqs()
	if len(r) != 1 {
		t.Fatalf("made %d requests", len(r))
	}
	if r[0].Method != http.MethodGet {
		t.Errorf("method = %s", r[0].Method)
	}
	if got := r[0].URL.Scheme + "://" + r[0].URL.Host + r[0].URL.Path; got != "https://pkg.go.dev/v1/search" {
		t.Errorf("endpoint = %s", got)
	}
	if r[0].URL.RawQuery != "limit=5&q=http+router" {
		t.Errorf("query = %s", r[0].URL.RawQuery)
	}
}

func TestSearchAPIError(t *testing.T) {
	stubAPI(t, serveJSON(http.StatusBadRequest, apiError{
		Code:    400,
		Message: "invalid filter",
		Fixes:   []string{"use a valid filter", "see the docs"},
	}))
	_, err := search(url.Values{"q": {"x"}})
	want := "Search failed: invalid filter\n  use a valid filter\n  see the docs"
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}

func TestSearchErrorStatusWithoutMessage(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"not json": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("<html>bad gateway</html>"))
		},
		"empty message": serveJSON(http.StatusBadGateway, apiError{Code: 502}),
	}
	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			stubAPI(t, handler)
			_, err := search(url.Values{"q": {"x"}})
			if err == nil || err.Error() != "Search failed with status 502 Bad Gateway" {
				t.Errorf("err = %v", err)
			}
		})
	}
}

func TestSearchBadJSON(t *testing.T) {
	stubAPI(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{not json"))
	})
	_, err := search(url.Values{"q": {"x"}})
	if err == nil || !strings.HasPrefix(err.Error(), "Failed to parse search results: ") {
		t.Errorf("err = %v", err)
	}
}

func TestHandleSearchParams(t *testing.T) {
	tests := []struct {
		args []string
		want url.Values
	}{
		{[]string{"cobra"}, url.Values{"q": {"cobra"}, "limit": {"26"}}},
		{[]string{"http", "router"}, url.Values{"q": {"http router"}, "limit": {"26"}}},
		{[]string{"-limit", "5", "x"}, url.Values{"q": {"x"}, "limit": {"5"}}},
		{[]string{"-limit", "0", "x"}, url.Values{"q": {"x"}}},
		{[]string{"-top", "x"}, url.Values{"q": {"x"}, "limit": {"1"}}},
		{[]string{"-t", "-limit", "9", "x"}, url.Values{"q": {"x"}, "limit": {"1"}}},
		{
			[]string{"-symbol", "Router", "-token", "abc", "-filter", "module:github.com", "x"},
			url.Values{"q": {"x"}, "limit": {"26"}, "symbol": {"Router"}, "token": {"abc"}, "filter": {"module:github.com"}},
		},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			reqs := stubAPI(t, serveJSON(http.StatusOK, searchResponse{}))
			if _, _, err := runHandler(t, HandleSearch, "search", tt.args...); err != nil {
				t.Fatal(err)
			}
			r := reqs()
			if len(r) != 1 {
				t.Fatalf("made %d requests", len(r))
			}
			if got := r[0].URL.Query(); got.Encode() != tt.want.Encode() {
				t.Errorf("query = %s, want %s", got.Encode(), tt.want.Encode())
			}
		})
	}
}

func TestHandleSearchOutput(t *testing.T) {
	stubAPI(t, serveJSON(http.StatusOK, twoResults))
	out, _, err := runHandler(t, HandleSearch, "search", "http", "router")
	if err != nil {
		t.Fatal(err)
	}
	want := `github.com/julienschmidt/httprouter v1.3.0
  module: github.com/julienschmidt/httprouter
  Package httprouter is a trie based high performance HTTP request router.

github.com/go-chi/chi/v5/middleware v5.3.2
  module: github.com/go-chi/chi/v5

More results: -token cd0d8a5c
`
	if out != want {
		t.Errorf("stdout =\n%s\nwant\n%s", out, want)
	}
}

func TestHandleSearchNoResults(t *testing.T) {
	stubAPI(t, serveJSON(http.StatusOK, searchResponse{}))
	out, _, err := runHandler(t, HandleSearch, "search", "zzz")
	if err != nil {
		t.Fatal(err)
	}
	if out != "No results.\n" {
		t.Errorf("stdout = %q", out)
	}
}

func TestHandleSearchNoResultsWithAction(t *testing.T) {
	isolateHome(t)
	calls := fakeGo(t, 0)
	for _, args := range [][]string{{"-a", "x", "zzz"}, {"-g", "zzz"}} {
		stubAPI(t, serveJSON(http.StatusOK, searchResponse{}))
		out, _, err := runHandler(t, HandleSearch, "search", args...)
		if err == nil || err.Error() != "-alias and -get need exactly one result, got 0" {
			t.Errorf("%v: err = %v", args, err)
		}
		if out != "No results.\n" {
			t.Errorf("%v: stdout = %q", args, out)
		}
	}
	if got := calls(); got != nil {
		t.Errorf("go ran: %q", got)
	}
}

func TestHandleSearchActionNeedsOneResult(t *testing.T) {
	isolateHome(t)
	calls := fakeGo(t, 0)
	for _, args := range [][]string{{"-alias", "x", "http"}, {"-get", "http"}} {
		stubAPI(t, serveJSON(http.StatusOK, twoResults))
		out, _, err := runHandler(t, HandleSearch, "search", args...)
		want := "-alias and -get need exactly one result, got 2; narrow the search or use -top"
		if err == nil || err.Error() != want {
			t.Errorf("%v: err = %v", args, err)
		}
		// Results are still printed before the error.
		if !strings.Contains(out, "github.com/julienschmidt/httprouter v1.3.0") {
			t.Errorf("%v: stdout = %q", args, out)
		}
	}
	if _, err := getAlias(t, "x"); err == nil {
		t.Error("alias saved despite multiple results")
	}
	if got := calls(); got != nil {
		t.Errorf("go ran: %q", got)
	}
}

func TestHandleSearchAliasAndGet(t *testing.T) {
	isolateHome(t)
	calls := fakeGo(t, 0)
	stubAPI(t, serveJSON(http.StatusOK, oneResult))

	out, _, err := runHandler(t, HandleSearch, "search", "-t", "-a", "cobra", "-g", "cobra")
	if err != nil {
		t.Fatal(err)
	}
	want := `github.com/spf13/cobra/doc v1.9.1
  module: github.com/spf13/cobra
  Cobra docs.
Saved cobra -> github.com/spf13/cobra
getting github.com/spf13/cobra...
`
	if out != want {
		t.Errorf("stdout =\n%s\nwant\n%s", out, want)
	}

	// The alias points at the module path, not the package path.
	row, err := getAlias(t, "cobra")
	if err != nil {
		t.Fatal(err)
	}
	if row.RepoUrl != "github.com/spf13/cobra" {
		t.Errorf("RepoUrl = %q", row.RepoUrl)
	}
	if got := calls(); len(got) != 1 || got[0] != "get github.com/spf13/cobra" {
		t.Errorf("go calls = %q", got)
	}
}

func TestHandleSearchInvalid(t *testing.T) {
	reqs := stubAPI(t, serveJSON(http.StatusOK, oneResult))
	tests := []struct {
		args      []string
		err       string
		showUsage bool
	}{
		{nil, "search requires a query", true},
		{[]string{"-alias", "a b", "x"}, `Invalid alias "a b", aliases cannot contain whitespace`, false},
		{[]string{"-limit", "-2", "x"}, "Invalid limit -2, expected 0 or more", true},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, usage, err := runHandler(t, HandleSearch, "search", tt.args...)
			if err == nil || err.Error() != tt.err {
				t.Errorf("err = %v, want %q", err, tt.err)
			}
			if got := strings.HasPrefix(usage, "Usage: gokit search [flags] <query>\n"); got != tt.showUsage {
				t.Errorf("usage shown = %v, want %v: %q", got, tt.showUsage, usage)
			}
		})
	}
	if n := len(reqs()); n != 0 {
		t.Errorf("invalid searches made %d requests", n)
	}
}

func TestHandleSearchAPIErrorPropagates(t *testing.T) {
	stubAPI(t, serveJSON(http.StatusInternalServerError, apiError{Message: "boom"}))
	out, _, err := runHandler(t, HandleSearch, "search", "x")
	if err == nil || err.Error() != "Search failed: boom" {
		t.Errorf("err = %v", err)
	}
	if out != "" {
		t.Errorf("stdout = %q", out)
	}
}
