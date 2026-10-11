package internal

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gokit/internal/consts"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type searchItem struct {
	PackagePath string `json:"packagePath"`
	ModulePath  string `json:"modulePath"`
	Version     string `json:"version"`
	Synopsis    string `json:"synopsis"`
}

type searchResponse struct {
	Items         []searchItem `json:"items"`
	Total         int          `json:"total"`
	NextPageToken string       `json:"nextPageToken"`
}

type apiError struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Fixes   []string `json:"fixes"`
}

func HandleSearch(searchCmd *flag.FlagSet) error {
	searchCmd.Usage = func() { SearchUsage(searchCmd) }

	var symbol, token, filter, alias string
	var limit int
	var top, get bool

	searchCmd.StringVar(&symbol, "symbol", "", "find symbols matching this string, within packages matching the query")
	searchCmd.IntVar(&limit, "limit", consts.SearchDefaultLimit, "max number of results to return")
	searchCmd.StringVar(&token, "token", "", "page token to resume listing from")
	searchCmd.StringVar(&filter, "filter", "", "only include results matching this filter expression")
	searchCmd.BoolVar(&top, "top", false, "only return the best match, same as -limit 1")
	searchCmd.BoolVar(&top, "t", false, "shorthand for -top")
	searchCmd.StringVar(&alias, "alias", "", "save the result under this alias, requires exactly one result (see -top)")
	searchCmd.StringVar(&alias, "a", "", "shorthand for -alias")
	searchCmd.BoolVar(&get, "get", false, "install the result with go get, requires exactly one result (see -top)")
	searchCmd.BoolVar(&get, "g", false, "shorthand for -get")

	searchCmd.Parse(flag.Args()[1:])

	if searchCmd.NArg() == 0 {
		searchCmd.Usage()
		return errors.New("search requires a query")
	}

	if alias != "" {
		if err := validateAlias(alias); err != nil {
			return err
		}
	}

	if limit < 0 {
		searchCmd.Usage()
		return fmt.Errorf("Invalid limit %d, expected 0 or more", limit)
	}

	if top {
		limit = 1
	}

	params := url.Values{}
	params.Set("q", strings.Join(searchCmd.Args(), " "))
	if symbol != "" {
		params.Set("symbol", symbol)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if token != "" {
		params.Set("token", token)
	}
	if filter != "" {
		params.Set("filter", filter)
	}

	result, err := search(params)
	if err != nil {
		return err
	}

	if len(result.Items) == 0 {
		fmt.Println("No results.")
		if alias != "" || get {
			return errors.New("-alias and -get need exactly one result, got 0")
		}
		return nil
	}

	for i, item := range result.Items {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s %s\n", item.PackagePath, item.Version)
		fmt.Printf("  module: %s\n", item.ModulePath)
		if item.Synopsis != "" {
			fmt.Printf("  %s\n", item.Synopsis)
		}
	}

	if result.NextPageToken != "" {
		fmt.Printf("\nMore results: -token %s\n", result.NextPageToken)
	}

	if (alias != "" || get) && len(result.Items) != 1 {
		return fmt.Errorf("-alias and -get need exactly one result, got %d; narrow the search or use -top", len(result.Items))
	}

	best := result.Items[0].ModulePath

	if alias != "" {
		if err := saveAlias(alias, best); err != nil {
			return err
		}
	}

	if get {
		return goGet(best)
	}

	return nil
}

func search(params url.Values) (*searchResponse, error) {
	client := http.Client{Timeout: 10 * time.Second}

	res, err := client.Get(consts.APIPrefix + "search?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("Failed to reach pkg.go.dev: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var apiErr apiError
		if err := json.NewDecoder(res.Body).Decode(&apiErr); err != nil || apiErr.Message == "" {
			return nil, fmt.Errorf("Search failed with status %s", res.Status)
		}

		msg := apiErr.Message
		for _, fix := range apiErr.Fixes {
			msg += "\n  " + fix
		}
		return nil, fmt.Errorf("Search failed: %s", msg)
	}

	var result searchResponse
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Failed to parse search results: %w", err)
	}

	return &result, nil
}
