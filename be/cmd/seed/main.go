package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lifosmin/admin-backend/internal/config"
	"github.com/lifosmin/admin-backend/internal/db"
	"github.com/lifosmin/admin-backend/internal/user"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	repo := user.NewRepository(pool)
	svc := user.NewService(repo)

	admin, err := svc.Create(ctx, user.CreateUserRequest{
		Email:    "admin@admin.com",
		Password: "admin123",
		Role:     user.RoleAdmin,
	})
	if err != nil {
		log.Fatalf("creating admin user: %v", err)
	}

	fmt.Printf("Admin user created:\n  ID:    %s\n  Email: %s\n  Role:  %s\n", admin.ID, admin.Email, admin.Role)
}
