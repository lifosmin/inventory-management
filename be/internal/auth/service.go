package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lifosmin/admin-backend/internal/config"
	"github.com/lifosmin/admin-backend/internal/user"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

type Service struct {
	userRepo *user.Repository
	jwtCfg   config.JWTConfig
}

func NewService(userRepo *user.Repository, jwtCfg config.JWTConfig) *Service {
	return &Service{userRepo: userRepo, jwtCfg: jwtCfg}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*TokenPair, []*http.Cookie, error) {
	u, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, err
	}
	if u == nil {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	accessToken, err := GenerateAccessToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.AccessExpiration)
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := GenerateRefreshToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.RefreshExpiration)
	if err != nil {
		return nil, nil, err
	}

	cookies := []*http.Cookie{
		{
			Name:     "access_token",
			Value:    accessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			MaxAge:   int(s.jwtCfg.AccessExpiration.Seconds()),
		},
		{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			MaxAge:   int(s.jwtCfg.RefreshExpiration.Seconds()),
		},
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, cookies, nil
}

func (s *Service) Refresh(ctx context.Context, refreshTokenStr string) (*TokenPair, []*http.Cookie, error) {
	claims, err := ValidateRefreshToken(s.jwtCfg.Secret, refreshTokenStr)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	userID, _ := claims.GetSubject()
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if u == nil {
		return nil, nil, ErrUserNotFound
	}

	newAccess, err := GenerateAccessToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.AccessExpiration)
	if err != nil {
		return nil, nil, err
	}
	newRefresh, err := GenerateRefreshToken(s.jwtCfg.Secret, u.ID, string(u.Role), s.jwtCfg.RefreshExpiration)
	if err != nil {
		return nil, nil, err
	}

	cookies := []*http.Cookie{
		{
			Name:     "access_token",
			Value:    newAccess,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			MaxAge:   int(s.jwtCfg.AccessExpiration.Seconds()),
		},
		{
			Name:     "refresh_token",
			Value:    newRefresh,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			MaxAge:   int(s.jwtCfg.RefreshExpiration.Seconds()),
		},
	}

	return &TokenPair{AccessToken: newAccess, RefreshToken: newRefresh}, cookies, nil
}

func (s *Service) Logout() []*http.Cookie {
	return []*http.Cookie{
		{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
		},
		{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
		},
	}
}
