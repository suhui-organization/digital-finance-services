package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes all .sql files from dir in filename order.
// SQL files should be idempotent (use IF NOT EXISTS / ON CONFLICT DO NOTHING).
func RunMigrations(db *sql.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}

	// Collect .sql files and sort by name (000001_init.sql < 000002_...)
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	if len(files) == 0 {
		slog.Warn("no SQL files found, skipping", "dir", dir)
		return nil
	}

	for _, name := range files {
		path := filepath.Join(dir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		sqlStr := string(sqlBytes)
		if strings.TrimSpace(sqlStr) == "" {
			continue
		}

		slog.Info("executing migration", "file", name, "bytes", len(sqlBytes))
		if _, err := db.Exec(sqlStr); err != nil {
			return fmt.Errorf("execute %s: %w", name, err)
		}
		slog.Info("migration applied", "file", name)
	}

	slog.Info("all migrations applied", "count", len(files))
	return nil
}
