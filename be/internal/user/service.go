package user

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateUserRequest) (*User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	return s.repo.Create(ctx, req, string(hashed))
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	u, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
