package app

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// LoginRequest is the accepted request format to log in.
type LoginRequest struct {
	Email    string
	Password string
}

// Validate the LoginRequest fields.
func (req *LoginRequest) Validate() error {
	if req.Email == "" {
		return errors.New("email is a required field")
	}
	if req.Password == "" {
		return errors.New("password is a required field")
	}

	return nil
}

// Claims to store in the JWT.
type Claims struct {
	ID int `json:"id"`
	jwt.RegisteredClaims
}

func generateJWT(id int, cfg configuration) (string, error) {
	t := time.Now().UTC()
	claims := &Claims{
		ID: id,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(t),
			NotBefore: jwt.NewNumericDate(t),
			ExpiresAt: jwt.NewNumericDate(t.Add(cfg.JWTDuration())),
			Issuer:    cfg.DomainName(),
			Audience:  []string{cfg.DomainName()},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(cfg.JWTSecret()))
}
