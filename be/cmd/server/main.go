package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lifosmin/admin-backend/internal/auth"
	"github.com/lifosmin/admin-backend/internal/config"
	"github.com/lifosmin/admin-backend/internal/db"
	"github.com/lifosmin/admin-backend/internal/inventory"
	"github.com/lifosmin/admin-backend/internal/lot"
	"github.com/lifosmin/admin-backend/internal/middleware"
	"github.com/lifosmin/admin-backend/internal/product"
	"github.com/lifosmin/admin-backend/internal/report"
	"github.com/lifosmin/admin-backend/internal/sale"
	"github.com/lifosmin/admin-backend/internal/user"
	"github.com/lifosmin/admin-backend/internal/warehouse"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	cfg, err := config.Load(ctx)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := user.NewRepository(pool)
	productRepo := product.NewRepository(pool)
	warehouseRepo := warehouse.NewRepository(pool)
	lotRepo := lot.NewRepository(pool)
	inventoryRepo := inventory.NewRepository(pool)

	userSvc := user.NewService(userRepo)
	productSvc := product.NewService(productRepo)
	warehouseSvc := warehouse.NewService(warehouseRepo)
	lotSvc := lot.NewService(lotRepo)
	inventorySvc := inventory.NewService(inventoryRepo, lotRepo, pool)
	authSvc := auth.NewService(userRepo, cfg.JWT)
	reportSvc := report.NewService(pool)
	saleRepo := sale.NewRepository(pool)
	saleSvc := sale.NewService(saleRepo, lotRepo, pool)

	authHandler := auth.NewHandler(authSvc)
	userHandler := user.NewHandler(userSvc)
	productHandler := product.NewHandler(productSvc)
	warehouseHandler := warehouse.NewHandler(warehouseSvc)
	lotHandler := lot.NewHandler(lotSvc, lotRepo)
	inventoryHandler := inventory.NewHandler(inventorySvc)
	reportHandler := report.NewHandler(reportSvc)
	saleHandler := sale.NewHandler(saleSvc)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(logger))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigin))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(20, time.Minute))
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWT.Secret))

			r.Get("/auth/me", authHandler.Me)

			r.Route("/users", func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Get("/", userHandler.List)
				r.Post("/", userHandler.Create)
				r.Get("/{id}", userHandler.GetByID)
				r.Put("/{id}", userHandler.Update)
				r.Delete("/{id}", userHandler.Delete)
			})

			r.Route("/products", func(r chi.Router) {
				r.Get("/", productHandler.List)
				r.Get("/{id}", productHandler.GetByID)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager"))
					r.Post("/", productHandler.Create)
					r.Put("/{id}", productHandler.Update)
					r.Delete("/{id}", productHandler.Delete)
				})
			})

			r.Route("/warehouses", func(r chi.Router) {
				r.Get("/", warehouseHandler.List)
				r.Get("/{id}", warehouseHandler.GetByID)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager"))
					r.Post("/", warehouseHandler.Create)
					r.Put("/{id}", warehouseHandler.Update)
				})
			})

			r.Route("/lots", func(r chi.Router) {
				r.Get("/", lotHandler.List)
				r.Get("/{id}", lotHandler.GetByID)
				r.Get("/{id}/movements", inventoryHandler.ListByLot)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager"))
					r.Post("/", lotHandler.Create)
					r.Put("/{id}", lotHandler.Update)
					r.Patch("/{id}/shipment-status", lotHandler.UpdateShipmentStatus)
					r.Patch("/{id}/payment", lotHandler.AddPayment)
				})
			})

			r.Route("/movements", func(r chi.Router) {
				r.Get("/", inventoryHandler.List)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager"))
					r.Post("/", inventoryHandler.RecordMovement)
				})
			})

			r.Route("/sales", func(r chi.Router) {
				r.Get("/", saleHandler.List)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireRole("admin", "manager"))
					r.Post("/", saleHandler.Create)
					r.Patch("/{id}/status", saleHandler.UpdateStatus)
					r.Patch("/{id}/payment", saleHandler.AddPayment)
				})
			})

			r.Get("/dashboard", saleHandler.Dashboard)

			r.Route("/reports", func(r chi.Router) {
				r.Get("/valuation", reportHandler.Valuation)
				r.Get("/aging", reportHandler.Aging)
				r.Get("/export", reportHandler.Export)
			})
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}
	logger.Info("server stopped")
}
