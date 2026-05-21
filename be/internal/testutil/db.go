// Package testutil provides shared helpers for integration tests.
//
// Tests use a real Postgres so we can exercise row-level locking and FIFO
// allocation under concurrency — mocks would defeat the purpose. Each call to
// SetupDB creates a fresh schema, applies the full migration history into it,
// and drops the schema on cleanup.
package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupDB returns a pool scoped to an isolated, per-test schema with all
// migrations applied. Connection details come from TEST_DATABASE_URL if set,
// otherwise the standard DB_* env vars used by the app. If neither is
// configured the test is skipped.
func SetupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := buildDSN(t)
	ctx := context.Background()

	schema := "test_" + randHex(8)
	createSchema(t, ctx, dsn, schema)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	cfg.MaxConns = 20
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, fmt.Sprintf(`SET search_path TO %q, public`, schema))
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		dropSchema(ctx, dsn, schema)
		t.Fatalf("create test pool: %v", err)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		pool.Close()
		dropSchema(ctx, dsn, schema)
		t.Fatalf("apply migrations: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		dropSchema(context.Background(), dsn, schema)
	})

	return pool
}

func buildDSN(t *testing.T) string {
	t.Helper()
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		t.Skip("TEST_DATABASE_URL not set and DB_USER not configured — skipping integration test")
	}
	host := envOr("DB_HOST", "localhost")
	port := envOr("DB_PORT", "5432")
	pw := os.Getenv("DB_PASSWORD")
	name := envOr("DB_NAME", "postgres")
	ssl := envOr("DB_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, pw, host, port, name, ssl)
}

func createSchema(t *testing.T, ctx context.Context, dsn, schema string) {
	t.Helper()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("bootstrap connect: %v", err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA %q`, schema)); err != nil {
		t.Fatalf("create schema %s: %v", schema, err)
	}
}

func dropSchema(ctx context.Context, dsn, schema string) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return
	}
	defer conn.Close(ctx)
	_, _ = conn.Exec(ctx, fmt.Sprintf(`DROP SCHEMA %q CASCADE`, schema))
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var ups []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			ups = append(ups, filepath.Join(dir, name))
		}
	}
	sort.Strings(ups)

	for _, path := range ups {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("exec %s: %w", filepath.Base(path), err)
		}
	}
	return nil
}

// migrationsDir resolves be/migrations relative to this source file so it
// works regardless of which package's tests are running.
func migrationsDir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot resolve testutil source location")
	}
	// thisFile = .../be/internal/testutil/db.go → up three to reach be/, then migrations/
	beDir := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	return filepath.Join(beDir, "migrations"), nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
