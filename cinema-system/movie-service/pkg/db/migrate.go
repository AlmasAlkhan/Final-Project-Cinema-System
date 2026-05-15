package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)
	for _, name := range files {
		path := filepath.Join(dir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}
