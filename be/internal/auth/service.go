package auth

import (
	"context"
	"errors"

	"github.com/lifosmin/admin-backend/internal/config"
	"github.com/lifosmin/admin-backend/internal/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	userRepo  *user.Repository
	jwtCfg    config.JWTConfig
	dummyHash []byte
}

func NewService(userRepo *user.Repository, jwtCfg config.JWTConfig) *Service {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), 12)
	return &Service{userRepo: userRepo, jwtCfg: jwtCfg, dummyHash: dummy}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenPair, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		// Run bcrypt against a dummy hash so timing matches the "user exists, wrong password" path.
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(req.Password))
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := GenerateAccessToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.AccessExpiration)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateRefreshToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.RefreshExpiration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshTokenStr string) (*TokenPair, error) {
	claims, err := ValidateRefreshToken(s.jwtCfg.Secret, refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	userID, _ := claims.GetSubject()
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	newAccess, err := GenerateAccessToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.AccessExpiration)
	if err != nil {
		return nil, err
	}
	newRefresh, err := GenerateRefreshToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.RefreshExpiration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: newAccess, RefreshToken: newRefresh}, nil
}
