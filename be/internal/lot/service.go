package lot

import (
	"context"
	"errors"
)

var ErrLotNotFound = errors.New("lot not found")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateLotRequest) (*Lot, error) {
	return s.repo.Create(ctx, req)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Lot, error) {
	l, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, ErrLotNotFound
	}
	return l, nil
}

func (s *Service) List(ctx context.Context) ([]Lot, error) {
	return s.repo.List(ctx)
}

func (s *Service) ListByProductFIFO(ctx context.Context, productID string) ([]Lot, error) {
	return s.repo.ListByProductFIFO(ctx, productID)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateLotRequest) (*Lot, error) {
	l, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if l == nil {
		return nil, ErrLotNotFound
	}
	return l, nil
}

func (s *Service) UpdateQuantity(ctx context.Context, id string, newQty float64) error {
	status := StatusAvailable
	if newQty <= 0 {
		status = StatusDepleted
		newQty = 0
	}
	return s.repo.UpdateQuantity(ctx, id, newQty, status)
}
