package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
)

// TODO add encryption / decryption into here

type userCtxKey string

const (
	// UserKey to store the user ID in context.
	UserKey = userCtxKey("user")
)

type Config interface {
	DomainName() string
	JWTDuration() time.Duration
	JWTSecret() string
}

type Authorizer struct {
	config Config
}

func NewAuthorizer(cfg Config) *Authorizer { return &Authorizer{cfg} }

// Claims to store in the JWT.
type Claims struct {
	UserID int `json:"userID"`
	jwt.RegisteredClaims
}

// Authenticate the given user into a JWT.
func (a *Authorizer) Authenticate(user *muzz.User, password string) (string, *apperror.Error) {
	// TODO compare with encryption lib
	if user.Password != password {
		return "", apperror.IncorrectCredentials()
	}

	t := time.Now().UTC()
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(t),
			NotBefore: jwt.NewNumericDate(t),
			ExpiresAt: jwt.NewNumericDate(t.Add(a.config.JWTDuration())),
			Issuer:    a.config.DomainName(),
			Audience:  []string{a.config.DomainName()},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tString, err := token.SignedString([]byte(a.config.JWTSecret()))
	if err != nil {
		return "", apperror.Internal(err)
	}

	return tString, nil
}

// Authorize parses and authorizes the given request JWT, extracts the Claims and authorizes them,
// returning the userID once successfully authorized.
func (a *Authorizer) Authorize(token string) (int, *apperror.Error) {
	c := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, c, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.JWTSecret()), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return 0, apperror.Unauthorised(err)
		}

		return 0, apperror.Internal(err)
	}

	_, ok := tkn.Claims.(*Claims)
	if !ok {
		return 0, apperror.Unauthorised(errors.New("unknown claims type"))
	}

	if !tkn.Valid {
		return 0, apperror.Unauthorised(errors.New("invalid jwt"))
	}

	err = a.validateClaims(c)
	if err != nil {
		return 0, apperror.Unauthorised(err)
	}

	return c.UserID, nil
}

// validateClaims using the most recent Config values.
func (a *Authorizer) validateClaims(c *Claims) error {
	if c.UserID == 0 {
		return errors.New("token has no user associated with it")
	}

	dur := a.config.JWTDuration()

	t := time.Now().UTC()
	if c.IssuedAt.Add(dur).Before(t) {
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

// UserFromContext attempts to retrieve the authenticated user from context.
func (*Authorizer) UserFromContext(ctx context.Context) (int, error) {
	uID, ok := ctx.Value(UserKey).(int)
	if !ok {
		return 0, errors.New("authenticated user not on context")
	}

	return uID, nil
}

// UserOnContext sets the authenticated user on context.
func (*Authorizer) UserOnContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserKey, userID)
}
