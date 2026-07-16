package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwtlib.RegisteredClaims
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, ttl time.Duration) (*Manager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 characters")
	}
	if ttl <= 0 {
		return nil, errors.New("JWT TTL must be positive")
	}
	return &Manager{secret: []byte(secret), ttl: ttl}, nil
}

func (m *Manager) GenerateToken(userID uint, username, email string) (string, error) {
	now := time.Now()
	claims := &Claims{UserID: userID, Username: username, Email: email, RegisteredClaims: jwtlib.RegisteredClaims{ExpiresAt: jwtlib.NewNumericDate(now.Add(m.ttl)), IssuedAt: jwtlib.NewNumericDate(now)}}
	return jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwtlib.ParseWithClaims(tokenString, claims, func(token *jwtlib.Token) (interface{}, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
