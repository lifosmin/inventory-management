package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidTokenType = errors.New("invalid token type")

func GenerateToken(secret string, userID string, role string, expiration time.Duration, tokenType string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"typ":  tokenType,
		"exp":  time.Now().Add(expiration).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateAccessToken(secret string, userID string, role string, expiration time.Duration) (string, error) {
	return GenerateToken(secret, userID, role, expiration, "access")
}

func GenerateRefreshToken(secret string, userID string, role string, expiration time.Duration) (string, error) {
	return GenerateToken(secret, userID, role, expiration, "refresh")
}

func ValidateToken(secret string, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func ValidateRefreshToken(secret string, tokenString string) (jwt.MapClaims, error) {
	claims, err := ValidateToken(secret, tokenString)
	if err != nil {
		return nil, err
	}
	if tokenType, _ := claims["typ"].(string); tokenType != "refresh" {
		return nil, ErrInvalidTokenType
	}
	return claims, nil
}
