package warehouse

import (
	"context"
	"errors"
)

var ErrWarehouseNotFound = errors.New("warehouse not found")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateWarehouseRequest) (*Warehouse, error) {
	return s.repo.Create(ctx, req)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Warehouse, error) {
	w, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrWarehouseNotFound
	}
	return w, nil
}

func (s *Service) List(ctx context.Context) ([]Warehouse, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateWarehouseRequest) (*Warehouse, error) {
	w, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrWarehouseNotFound
	}
	return w, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
