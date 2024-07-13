package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nickbadlose/muzz"
	"github.com/nickbadlose/muzz/internal/apperror"
)

type userCtxKey string

const (
	// userKey to store the user ID in context.
	userKey userCtxKey = "user"
)

// Config is the interface to retrieve configuration secrets from the environment.
type Config interface {
	// DomainName retrieves the domain name of the server.
	DomainName() string
	// JWTDuration retrieves the expiry duration of the generated JWTs.
	JWTDuration() time.Duration
	// JWTSecret retrieves the JWT signature to authorize with.
	JWTSecret() string
}

// Repository is the interface to authenticate a user against the valid users records.
type Repository interface {
	// Authenticate the provided user credentials.
	Authenticate(ctx context.Context, email, password string) (*muzz.User, error)
}

// Authoriser is the service which handles all authentication and authorization based logic.
type Authoriser struct {
	config     Config
	repository Repository
}

// NewAuthoriser builds a new *Authoriser.
func NewAuthoriser(cfg Config, repo Repository) (*Authoriser, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if repo == nil {
		return nil, errors.New("repository cannot be nil")
	}
	return &Authoriser{config: cfg, repository: repo}, nil
}

// Claims to store in the JWT.
type Claims struct {
	// UserID is the id of the authenticated user, this can be retrieved in subsequent requests.
	UserID int `json:"userID"`
	jwt.RegisteredClaims
}

// Authenticate the given credentials.
func (a *Authoriser) Authenticate(ctx context.Context, email, password string) (string, *muzz.User, *apperror.Error) {
	user, err := a.repository.Authenticate(ctx, email, password)
	if errors.Is(err, apperror.ErrNoResults) {
		return "", nil, apperror.IncorrectCredentials()
	}
	if err != nil {
		return "", nil, apperror.Internal(err)
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
		return "", nil, apperror.Internal(err)
	}

	return tString, user, nil
}

// Authorise parses and authorises the given request JWT, extracts the Claims and authorises them,
// returning the userID once successfully authorised.
func (a *Authoriser) Authorise(token string) (int, *apperror.Error) {
	c := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, c, func(_ *jwt.Token) (any, error) {
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

// validateClaims using the current Config values.
func (a *Authoriser) validateClaims(c *Claims) error {
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
func (*Authoriser) UserFromContext(ctx context.Context) (int, error) {
	uID, ok := ctx.Value(userKey).(int)
	if !ok {
		return 0, errors.New("authenticated user not on context")
	}

	return uID, nil
}

// UserOnContext sets the authenticated user on context.
func (*Authoriser) UserOnContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userKey, userID)
}
