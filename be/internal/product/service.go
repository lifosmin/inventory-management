package product

import (
	"context"
	"errors"
)

var ErrProductNotFound = errors.New("product not found")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateProductRequest) (*Product, error) {
	return s.repo.Create(ctx, req)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return p, nil
}

func (s *Service) List(ctx context.Context) ([]Product, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateProductRequest) (*Product, error) {
	p, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrProductNotFound
	}
	return p, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
