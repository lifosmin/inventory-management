package sale

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lifosmin/admin-backend/internal/lot"
)

var (
	ErrInsufficientStock   = errors.New("insufficient stock for this sale")
	ErrPaymentExceedsTotal = errors.New("payment exceeds total owed")
)

type Service struct {
	repo    *Repository
	lotRepo *lot.Repository
	pool    *pgxpool.Pool
}

func NewService(repo *Repository, lotRepo *lot.Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, lotRepo: lotRepo, pool: pool}
}

func (s *Service) Create(ctx context.Context, req CreateSaleRequest) (*Sale, error) {
	lots, err := s.lotRepo.ListAvailableByProductWarehouse(ctx, req.ProductID, req.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("fetching available lots: %w", err)
	}

	var totalAvailable float64
	for _, l := range lots {
		totalAvailable += l.Quantity
	}
	if totalAvailable < req.Qty {
		return nil, ErrInsufficientStock
	}

	var sale *Sale
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sale, err = s.repo.CreateTx(ctx, tx, req)
	if err != nil {
		return nil, err
	}

	remaining := req.Qty
	var allocations []Allocation
	for _, l := range lots {
		if remaining <= 0 {
			break
		}
		allocQty := l.Quantity
		if allocQty > remaining {
			allocQty = remaining
		}

		alloc, err := s.repo.CreateAllocationTx(ctx, tx, sale.ID, l.ID, allocQty, l.UnitCost)
		if err != nil {
			return nil, err
		}
		allocations = append(allocations, *alloc)

		newQty := l.Quantity - allocQty
		status := lot.StatusAvailable
		if newQty <= 0 {
			status = lot.StatusDepleted
			newQty = 0
		}
		_, err = tx.Exec(ctx,
			`UPDATE lots SET quantity = $2, status = $3, updated_at = now() WHERE id = $1`,
			l.ID, newQty, status,
		)
		if err != nil {
			return nil, fmt.Errorf("updating lot quantity: %w", err)
		}

		remaining -= allocQty
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing sale: %w", err)
	}

	sale.Allocations = allocations
	return sale, nil
}

func (s *Service) List(ctx context.Context) ([]Sale, error) {
	return s.repo.List(ctx)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) error {
	return s.repo.UpdateStatus(ctx, id, req)
}

func (s *Service) AddPayment(ctx context.Context, id string, amount float64) (*Sale, error) {
	return s.repo.AddPayment(ctx, id, amount)
}

func (s *Service) GetDashboard(ctx context.Context) (*Dashboard, error) {
	var d Dashboard

	err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(quantity * unit_cost), 0)
		 FROM lots WHERE status = 'available' AND quantity > 0`,
	).Scan(&d.NetWorth)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying net worth: %w", err)
	}

	err = s.pool.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(sa.qty * s.sell_price), 0),
			COALESCE(SUM(sa.qty * sa.unit_cost), 0)
		 FROM sale_allocations sa
		 JOIN sales s ON sa.sale_id = s.id`,
	).Scan(&d.TotalRevenue, &d.TotalCOGS)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("querying profit/loss: %w", err)
	}
	d.ProfitLoss = d.TotalRevenue - d.TotalCOGS

	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM lots`).Scan(&d.TotalRestocks)
	if err != nil {
		return nil, fmt.Errorf("counting restocks: %w", err)
	}

	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sales`).Scan(&d.TotalSales)
	if err != nil {
		return nil, fmt.Errorf("counting sales: %w", err)
	}

	return &d, nil
}

type Dashboard struct {
	NetWorth      float64 `json:"net_worth"`
	TotalRevenue  float64 `json:"total_revenue"`
	TotalCOGS     float64 `json:"total_cogs"`
	ProfitLoss    float64 `json:"profit_loss"`
	TotalRestocks int     `json:"total_restocks"`
	TotalSales    int     `json:"total_sales"`
}
