package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type userCtxKey string

const (
	// UserKey to store the user ID in context.
	UserKey = userCtxKey("user")
)

type config interface {
	DomainName() string
	JWTDuration() time.Duration
	JWTSecret() string
}

type Authorizer interface {
	Generator
	Validator
	UserIDFromContext(ctx context.Context) (int, error)
}

type Generator interface {
	GenerateJWT(int) (string, error)
}

type Validator interface {
	ParseJWT(string) (*jwt.Token, *Claims, error)
	ValidateClaims(*Claims) error
}

type authorizer struct {
	config config
}

func NewAuthoriser(cfg config) Authorizer { return &authorizer{cfg} }

// Claims to store in the JWT.
type Claims struct {
	UserID int `json:"userID"`
	jwt.RegisteredClaims
}

// GenerateJWT for the given user ID.
func (a *authorizer) GenerateJWT(id int) (string, error) {
	t := time.Now().UTC()
	claims := &Claims{
		UserID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(t),
			NotBefore: jwt.NewNumericDate(t),
			ExpiresAt: jwt.NewNumericDate(t.Add(a.config.JWTDuration())),
			Issuer:    a.config.DomainName(),
			Audience:  []string{a.config.DomainName()},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(a.config.JWTSecret()))
}

// UserIDFromContext attempts to retrieve the authenticated user ID from context.
func (a *authorizer) UserIDFromContext(ctx context.Context) (int, error) {
	uID, ok := ctx.Value(UserKey).(int)
	if !ok {
		return 0, errors.New("could not find user")
	}

	return uID, nil
}

// ParseJWT parses the given request JWT and extracts the Claims.
func (a *authorizer) ParseJWT(token string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return a.config.JWTSecret(), nil
	})
	if err != nil {
		return nil, nil, err
	}
	return tkn, claims, nil
}

func (a *authorizer) ValidateClaims(c *Claims) error {
	if c.UserID == 0 {
		return errors.New("no user associated with this token")
	}

	t := time.Now().UTC()
	if c.IssuedAt.Add(a.config.JWTDuration()).Before(t) {
		return jwt.ErrTokenExpired
	}

	if c.Issuer != a.config.DomainName() {
		return jwt.ErrTokenInvalidIssuer
	}

	for _, aud := range c.Audience {
		if aud == a.config.DomainName() {
			return nil
		}
	}

	return jwt.ErrTokenInvalidAudience
}
