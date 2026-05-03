package sale

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateTx(ctx context.Context, tx pgx.Tx, req CreateSaleRequest) (*Sale, error) {
	var s Sale
	err := tx.QueryRow(ctx,
		`INSERT INTO sales (product_id, warehouse_id, buyer_name, qty, sell_price)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, product_id, warehouse_id, buyer_name, qty, sell_price, payment_status, shipment_status, created_at`,
		req.ProductID, req.WarehouseID, req.BuyerName, req.Qty, req.SellPrice,
	).Scan(&s.ID, &s.ProductID, &s.WarehouseID, &s.BuyerName, &s.Qty, &s.SellPrice,
		&s.PaymentStatus, &s.ShipmentStatus, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating sale: %w", err)
	}
	return &s, nil
}

func (r *Repository) CreateAllocationTx(ctx context.Context, tx pgx.Tx, saleID, lotID string, qty, unitCost float64) (*Allocation, error) {
	var a Allocation
	err := tx.QueryRow(ctx,
		`INSERT INTO sale_allocations (sale_id, lot_id, qty, unit_cost)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, sale_id, lot_id, qty, unit_cost, created_at`,
		saleID, lotID, qty, unitCost,
	).Scan(&a.ID, &a.SaleID, &a.LotID, &a.Qty, &a.UnitCost, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating allocation: %w", err)
	}
	return &a, nil
}

func (r *Repository) List(ctx context.Context) ([]Sale, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.product_id, p.name, s.warehouse_id, w.name,
		        s.buyer_name, s.qty, s.sell_price, s.payment_status, s.shipment_status, s.created_at
		 FROM sales s
		 JOIN products p ON s.product_id = p.id
		 JOIN warehouses w ON s.warehouse_id = w.id
		 ORDER BY s.created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing sales: %w", err)
	}
	defer rows.Close()

	var sales []Sale
	for rows.Next() {
		var s Sale
		if err := rows.Scan(&s.ID, &s.ProductID, &s.ProductName, &s.WarehouseID, &s.WarehouseName,
			&s.BuyerName, &s.Qty, &s.SellPrice, &s.PaymentStatus, &s.ShipmentStatus, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning sale: %w", err)
		}
		sales = append(sales, s)
	}
	return sales, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sales SET
			payment_status = COALESCE($2, payment_status),
			shipment_status = COALESCE($3, shipment_status)
		 WHERE id = $1`,
		id, req.PaymentStatus, req.ShipmentStatus,
	)
	if err != nil {
		return fmt.Errorf("updating sale status: %w", err)
	}
	return nil
}
