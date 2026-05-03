package user

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

func (r *Repository) Create(ctx context.Context, req CreateUserRequest, hashedPassword string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, $3)
		 RETURNING id, email, role, created_at, updated_at`,
		req.Email, hashedPassword, req.Role,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return &u, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return &u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return &u, nil
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, role, created_at, updated_at FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) Update(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`UPDATE users SET
			email = COALESCE($2, email),
			role = COALESCE($3, role),
			updated_at = now()
		 WHERE id = $1
		 RETURNING id, email, role, created_at, updated_at`,
		id, req.Email, req.Role,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}
	return &u, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}
