package warehouse

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

func (r *Repository) Create(ctx context.Context, req CreateWarehouseRequest) (*Warehouse, error) {
	var w Warehouse
	err := r.pool.QueryRow(ctx,
		`INSERT INTO warehouses (name, address) VALUES ($1, $2)
		 RETURNING id, name, address, created_at`,
		req.Name, req.Address,
	).Scan(&w.ID, &w.Name, &w.Address, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating warehouse: %w", err)
	}
	return &w, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Warehouse, error) {
	var w Warehouse
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, address, created_at FROM warehouses WHERE id = $1`, id,
	).Scan(&w.ID, &w.Name, &w.Address, &w.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting warehouse: %w", err)
	}
	return &w, nil
}

func (r *Repository) List(ctx context.Context) ([]Warehouse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, address, created_at FROM warehouses ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing warehouses: %w", err)
	}
	defer rows.Close()

	var warehouses []Warehouse
	for rows.Next() {
		var w Warehouse
		if err := rows.Scan(&w.ID, &w.Name, &w.Address, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning warehouse: %w", err)
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

func (r *Repository) Update(ctx context.Context, id string, req UpdateWarehouseRequest) (*Warehouse, error) {
	var w Warehouse
	err := r.pool.QueryRow(ctx,
		`UPDATE warehouses SET
			name = COALESCE($2, name),
			address = COALESCE($3, address)
		 WHERE id = $1
		 RETURNING id, name, address, created_at`,
		id, req.Name, req.Address,
	).Scan(&w.ID, &w.Name, &w.Address, &w.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("updating warehouse: %w", err)
	}
	return &w, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM warehouses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting warehouse: %w", err)
	}
	return nil
}
