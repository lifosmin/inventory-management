package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lifosmin/admin-backend/internal/db"
	"github.com/lifosmin/admin-backend/internal/lot"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock in lot")
	ErrInvalidMovement   = errors.New("invalid movement type")
)

type Service struct {
	repo    *Repository
	lotRepo *lot.Repository
	pool    *pgxpool.Pool
}

func NewService(repo *Repository, lotRepo *lot.Repository, pool *pgxpool.Pool) *Service {
	return &Service{repo: repo, lotRepo: lotRepo, pool: pool}
}

func (s *Service) RecordMovement(ctx context.Context, req CreateMovementRequest, performedBy string) (*Movement, error) {
	switch req.MovementType {
	case MovementReceipt, MovementReturn:
		return s.handleInbound(ctx, req, performedBy)
	case MovementIssue:
		return s.handleOutbound(ctx, req, performedBy)
	case MovementAdjustment:
		return s.handleAdjustment(ctx, req, performedBy)
	default:
		return nil, ErrInvalidMovement
	}
}

func (s *Service) handleInbound(ctx context.Context, req CreateMovementRequest, performedBy string) (*Movement, error) {
	var movement *Movement
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		l, err := s.lotRepo.GetByID(ctx, req.LotID)
		if err != nil {
			return err
		}
		if l == nil {
			return lot.ErrLotNotFound
		}

		newQty := l.Quantity + req.Quantity
		status := lot.StatusAvailable
		if err := s.lotRepo.UpdateQuantity(ctx, l.ID, newQty, status); err != nil {
			return err
		}

		movement, err = s.repo.CreateMovementTx(ctx, tx, req, performedBy)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("recording inbound movement: %w", err)
	}
	return movement, nil
}

func (s *Service) handleOutbound(ctx context.Context, req CreateMovementRequest, performedBy string) (*Movement, error) {
	var movement *Movement
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		l, err := s.lotRepo.GetByID(ctx, req.LotID)
		if err != nil {
			return err
		}
		if l == nil {
			return lot.ErrLotNotFound
		}
		if l.Quantity < req.Quantity {
			return ErrInsufficientStock
		}

		newQty := l.Quantity - req.Quantity
		status := lot.StatusAvailable
		if newQty <= 0 {
			status = lot.StatusDepleted
			newQty = 0
		}
		if err := s.lotRepo.UpdateQuantity(ctx, l.ID, newQty, status); err != nil {
			return err
		}

		movement, err = s.repo.CreateMovementTx(ctx, tx, req, performedBy)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("recording outbound movement: %w", err)
	}
	return movement, nil
}

func (s *Service) handleAdjustment(ctx context.Context, req CreateMovementRequest, performedBy string) (*Movement, error) {
	var movement *Movement
	err := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		l, err := s.lotRepo.GetByID(ctx, req.LotID)
		if err != nil {
			return err
		}
		if l == nil {
			return lot.ErrLotNotFound
		}

		newQty := l.Quantity + req.Quantity
		if newQty < 0 {
			newQty = 0
		}
		status := lot.StatusAvailable
		if newQty <= 0 {
			status = lot.StatusDepleted
		}
		if err := s.lotRepo.UpdateQuantity(ctx, l.ID, newQty, status); err != nil {
			return err
		}

		movement, err = s.repo.CreateMovementTx(ctx, tx, req, performedBy)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("recording adjustment: %w", err)
	}
	return movement, nil
}

func (s *Service) ListByLot(ctx context.Context, lotID string) ([]Movement, error) {
	return s.repo.ListByLot(ctx, lotID)
}

func (s *Service) List(ctx context.Context) ([]Movement, error) {
	return s.repo.List(ctx)
}
