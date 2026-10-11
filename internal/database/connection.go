package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/syzygicality/gokit/internal/consts"
	"github.com/syzygicality/gokit/internal/database/gen"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

//go:embed sql/schema.sql
var schema string

// resolveDatabasePath expands a leading "~/" in DatabasePath to the user's home directory.
func resolveDatabasePath() (string, error) {
	path := consts.DatabasePath
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("Failed to locate home directory: %w", err)
		}
		path = filepath.Join(home, rest)
	}
	return path, nil
}

func WithDB(callback func(db *gen.Queries) error) error {
	path, err := resolveDatabasePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), consts.DatabaseDirPerm); err != nil {
		return fmt.Errorf("Failed to create directory %s, please check your file permissions", filepath.Dir(path))
	}

	db, err := sql.Open("sqlite", path)

	if err != nil {
		return fmt.Errorf("Failed to open alias file, please check your file permissions on %s", path)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("Failed to open alias file %s, please check your file permissions: %w", path, err)
	}

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("Failed to initialize alias database %s: %w", path, err)
	}

	return callback(gen.New(db))
}
