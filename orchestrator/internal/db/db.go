// Package db owns all Postgres access for the orchestrator — the only
// component in the system with database credentials (see docs/database.md).
//
// Migrations are applied by a small hand-rolled runner instead of a
// migration library: the pipeline only ever needs to apply "up" migrations,
// in order, once, as a one-off administrative process (Twelve-Factor,
// Factor XII), so a full migration framework would be an extra dependency
// for no real benefit.
package db

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"neurolaw/orchestrator/migrations"
)

// Connect opens a connection pool to the given Postgres URL.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	return pool, nil
}

// Migrate applies any embedded *.up.sql migration not yet recorded in
// schema_migrations, in filename order, each inside its own transaction.
// It returns the filenames of the migrations it applied (empty if the
// schema was already up to date).
func Migrate(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return nil, fmt.Errorf("ensuring schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("reading embedded migrations: %w", err)
	}

	var filenames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)

	var applied []string
	for _, filename := range filenames {
		var alreadyApplied bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`,
			filename,
		).Scan(&alreadyApplied)
		if err != nil {
			return nil, fmt.Errorf("checking migration %s: %w", filename, err)
		}
		if alreadyApplied {
			continue
		}

		contents, err := migrations.FS.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("reading migration %s: %w", filename, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("beginning transaction for %s: %w", filename, err)
		}
		if _, err := tx.Exec(ctx, string(contents)); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("applying migration %s: %w", filename, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1)`, filename,
		); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("recording migration %s: %w", filename, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("committing migration %s: %w", filename, err)
		}

		applied = append(applied, filename)
	}

	return applied, nil
}
