package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrInvalidAlgorithm = errors.New("unexpected signing algorithm")
)

type Claims struct {
	UserID uuid.UUID `json:"sub"`
	Email  string    `json:"email"`
	OrgID  uuid.UUID `json:"org_id,omitempty"`
	Role   string    `json:"role,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secretKey []byte
	issuer    string
}

type Config struct {
	SecretKey              string `env:"SECRET_KEY,required"`
	AccessTokenExpireMins  int    `env:"ACCESS_TOKEN_EXPIRE_MINUTES" envDefault:"15"`
	RefreshTokenExpireDays int    `env:"REFRESH_TOKEN_EXPIRE_DAYS" envDefault:"30"`
	Issuer                 string `env:"JWT_ISSUER" envDefault:"crm-ai"`
}

func (c Config) AccessDuration() time.Duration {
	return time.Duration(c.AccessTokenExpireMins) * time.Minute
}

func New(cfg Config) *Manager {
	return &Manager{
		secretKey: []byte(cfg.SecretKey),
		issuer:    cfg.Issuer,
	}
}

func (m *Manager) Generate(userID, orgID uuid.UUID, email, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		OrgID:  orgID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *Manager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidAlgorithm, t.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
