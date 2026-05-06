package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lifosmin/admin-backend/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run cmd/migrate/main.go [up|down]")
	}

	ctx := context.Background()
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	m, err := migrate.New("file:///migrations", cfg.DB.DSN())
	if err != nil {
		log.Fatalf("creating migrator: %v", err)
	}
	defer m.Close()

	cmd := os.Args[1]
	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration up: %v", err)
		}
		fmt.Println("migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration down: %v", err)
		}
		fmt.Println("migrations rolled back successfully")
	default:
		log.Fatalf("unknown command: %s (use 'up' or 'down')", cmd)
	}
}
