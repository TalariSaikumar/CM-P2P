// Run SQL migrations without psql: go run ./cmd/migrate [up|down]
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"carmanage/backend/internal/config"

	"github.com/jackc/pgx/v5"
)

func main() {
	direction := "up"
	if len(os.Args) > 1 {
		direction = strings.ToLower(strings.TrimSpace(os.Args[1]))
	}
	if direction != "up" && direction != "down" {
		log.Fatalf("usage: go run ./cmd/migrate [up|down]")
	}

	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		root = filepath.Join(root, "backend")
	}

	cfg, err := config.Load(root)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	migrationsDir := filepath.Join(root, "migrations")
	var pattern string
	if direction == "up" {
		pattern = "*.up.sql"
	} else {
		pattern = "*.down.sql"
	}

	entries, err := filepath.Glob(filepath.Join(migrationsDir, pattern))
	if err != nil {
		log.Fatalf("glob: %v", err)
	}
	sort.Strings(entries)
	if direction == "down" {
		sort.Sort(sort.Reverse(sort.StringSlice(entries)))
	}
	if len(entries) == 0 {
		log.Printf("no %s migration files", direction)
		return
	}

	for _, path := range entries {
		name := filepath.Base(path)
		log.Printf("applying %s: %s", direction, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read %s: %v", name, err)
		}
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			log.Fatalf("migration failed %s: %v", name, err)
		}
	}

	fmt.Printf("All '%s' migrations applied successfully (%d files).\n", direction, len(entries))
}
