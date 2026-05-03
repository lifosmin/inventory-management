package inventory

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

func (r *Repository) CreateMovement(ctx context.Context, req CreateMovementRequest, performedBy string) (*Movement, error) {
	var m Movement
	err := r.pool.QueryRow(ctx,
		`INSERT INTO lot_movements (lot_id, movement_type, quantity, reference_doc, notes, performed_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, lot_id, movement_type, quantity, reference_doc, notes, performed_by, created_at`,
		req.LotID, req.MovementType, req.Quantity, req.ReferenceDoc, req.Notes, performedBy,
	).Scan(&m.ID, &m.LotID, &m.MovementType, &m.Quantity, &m.ReferenceDoc, &m.Notes, &m.PerformedBy, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating movement: %w", err)
	}
	return &m, nil
}

func (r *Repository) CreateMovementTx(ctx context.Context, tx pgx.Tx, req CreateMovementRequest, performedBy string) (*Movement, error) {
	var m Movement
	err := tx.QueryRow(ctx,
		`INSERT INTO lot_movements (lot_id, movement_type, quantity, reference_doc, notes, performed_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, lot_id, movement_type, quantity, reference_doc, notes, performed_by, created_at`,
		req.LotID, req.MovementType, req.Quantity, req.ReferenceDoc, req.Notes, performedBy,
	).Scan(&m.ID, &m.LotID, &m.MovementType, &m.Quantity, &m.ReferenceDoc, &m.Notes, &m.PerformedBy, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating movement in tx: %w", err)
	}
	return &m, nil
}

func (r *Repository) ListByLot(ctx context.Context, lotID string) ([]Movement, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, lot_id, movement_type, quantity, reference_doc, notes, performed_by, created_at
		 FROM lot_movements WHERE lot_id = $1 ORDER BY created_at DESC`, lotID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing movements by lot: %w", err)
	}
	defer rows.Close()

	var movements []Movement
	for rows.Next() {
		var m Movement
		if err := rows.Scan(&m.ID, &m.LotID, &m.MovementType, &m.Quantity, &m.ReferenceDoc, &m.Notes, &m.PerformedBy, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning movement: %w", err)
		}
		movements = append(movements, m)
	}
	return movements, nil
}

func (r *Repository) List(ctx context.Context) ([]Movement, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, lot_id, movement_type, quantity, reference_doc, notes, performed_by, created_at
		 FROM lot_movements ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing movements: %w", err)
	}
	defer rows.Close()

	var movements []Movement
	for rows.Next() {
		var m Movement
		if err := rows.Scan(&m.ID, &m.LotID, &m.MovementType, &m.Quantity, &m.ReferenceDoc, &m.Notes, &m.PerformedBy, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning movement: %w", err)
		}
		movements = append(movements, m)
	}
	return movements, nil
}
