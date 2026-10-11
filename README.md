# gokit

Golang CLI tool to improve package discovery and import aliasing.

Search pkg.go.dev from your terminal, save the packages you use under short aliases, and `go get` them by alias instead of typing out the full module path.

```sh
gokit search -top -alias cobra cobra   # find a package and save it as "cobra"
gokit cobra                            # go get github.com/spf13/cobra
```

## Installation

Requires Go 1.26 or newer.

```sh
go install github.com/syzygicality/gokit@latest
```

This puts the `gokit` binary in `$(go env GOPATH)/bin` (or `$GOBIN` if set). Make sure that directory is on your `PATH`.

## Usage

```
gokit <command/alias> [flags]
```

Run `gokit help` for an overview, or `gokit <command> -h` for a command's flags.

Flags must come before positional arguments, for example `gokit save -g gin github.com/gin-gonic/gin`.

### Installing a package by alias

```sh
gokit <alias>
```

Runs `go get` on the package saved under `<alias>`. Like `go get`, this must be run inside a Go module.

A leading `http://` or `https://`, a trailing `/` and a trailing `.git` are stripped from the saved URL, so a URL copied from a browser or a clone link works.

### search

Search pkg.go.dev for packages, best match first.

```sh
gokit search [flags] <query>
```

| Flag | Description |
| --- | --- |
| `-limit <n>` | Maximum number of results (default 26) |
| `-top`, `-t` | Only return the best match, same as `-limit 1` |
| `-alias <name>`, `-a <name>` | Save the result under this alias |
| `-get`, `-g` | Install the result with `go get` |
| `-symbol <name>` | Find symbols matching this string, within packages matching the query |
| `-filter <expr>` | Only include results matching this filter expression (see the [pkg.go.dev API docs](https://pkg.go.dev/v1/api#filters)) |
| `-token <token>` | Page token to resume from |

`-alias` and `-get` need exactly one result, so pair them with `-top` or a query narrow enough to return a single package. The saved alias points at the result's module path.

When there are more results, the output ends with a `-token` value. Repeat the same search with that token to get the next page:

```sh
gokit search -limit 5 http router
gokit search -limit 5 -token <token> http router
```

### save

Save a package under an alias, replacing any package already saved under that alias.

```sh
gokit save [flags] <alias> <repo_url>
```

| Flag | Description |
| --- | --- |
| `-get`, `-g` | Install the package with `go get` after saving |

The repo URL can be a module path (`github.com/spf13/cobra`) or a URL (`https://github.com/spf13/cobra`).

Aliases must:

- be at most 32 characters
- not contain whitespace
- not start with `-`
- not be a gokit command name (`help`, `list`, `save`, `remove`, `search`)

### list

List saved aliases and the packages they point to.

```sh
gokit list [flags]
```

| Flag | Description |
| --- | --- |
| `-sort <column>`, `-s <column>` | Column to sort by: `alias` (default), `updated` or `created` |
| `-dir <direction>`, `-d <direction>` | Sort direction: `asc` or `desc`. Defaults to `asc` for `alias` and `desc` for the timestamp columns |

Timestamps are shown in UTC.

### remove

Remove a saved alias.

```sh
gokit remove <alias>
```

## Example

```sh
$ gokit search -limit 2 http router
github.com/julienschmidt/httprouter v1.3.0
  module: github.com/julienschmidt/httprouter
  Package httprouter is a trie based high performance HTTP request router.

github.com/go-chi/chi/v5 v5.3.2
  module: github.com/go-chi/chi/v5
  Package chi is a small, idiomatic and composable router for building HTTP services.

More results: -token cd0d8a5c13fa5aae260e155416c61da6

$ gokit save chi github.com/go-chi/chi/v5
Saved chi -> github.com/go-chi/chi/v5

$ gokit list
ALIAS  REPO URL                  UPDATED              CREATED
chi    github.com/go-chi/chi/v5  2026-10-11 21:06:02  2026-10-11 21:06:02

$ gokit chi
getting github.com/go-chi/chi/v5...
go: added github.com/go-chi/chi/v5 v5.3.2
```

## Data

Aliases are stored in a SQLite database at `~/.gokit/aliases`, created on first use. Delete that file to clear all aliases.

## License

[MIT](LICENSE)
