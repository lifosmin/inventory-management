package product

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, req CreateProductRequest) (*Product, error) {
	var p Product
	err := r.pool.QueryRow(ctx,
		`INSERT INTO products (name, category) VALUES ($1, $2)
		 RETURNING id, name, category, created_at`,
		req.Name, req.Category,
	).Scan(&p.ID, &p.Name, &p.Category, &p.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("a product with this name already exists")
		}
		return nil, fmt.Errorf("creating product: %w", err)
	}
	return &p, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Product, error) {
	var p Product
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, category, created_at FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Category, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting product: %w", err)
	}
	return &p, nil
}

func (r *Repository) List(ctx context.Context) ([]Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, category, created_at FROM products ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *Repository) Update(ctx context.Context, id string, req UpdateProductRequest) (*Product, error) {
	var p Product
	err := r.pool.QueryRow(ctx,
		`UPDATE products SET
			name = COALESCE($2, name),
			category = COALESCE($3, category)
		 WHERE id = $1
		 RETURNING id, name, category, created_at`,
		id, req.Name, req.Category,
	).Scan(&p.ID, &p.Name, &p.Category, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("updating product: %w", err)
	}
	return &p, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}
	return nil
}
