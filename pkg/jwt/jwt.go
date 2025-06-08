// jwt pacakge provides functionality to generate, sign and velidate JWT tokens
package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

// TimeNow is (or at least should be) used in ensure stable tests that do not depend on real time. In other words TimeNow will be mocked in unit tests.
var TimeNow = time.Now
var ErrTokenNotValid = errors.New("token is not valid")
var ErrTokenExpired = errors.New("token expired")

// TokenInfo is a struct that is used to transfer data for a jwt.Token to be generated
type TokenInfo struct {
	UserName   string
	Subject    string
	Expiration time.Duration
}

// JWTService is a wrapper around the github.com/golang-jwt/jwt functionality
type JWTService struct {
	secret []byte
}

// NewJWTService is a constructor function for the JWTService type
func NewJWTService(jwtSecret []byte) *JWTService {
	return &JWTService{
		secret: jwtSecret,
	}
}

// Generate receives token info and creates a jwt.Token
func (gen *JWTService) Generate(data TokenInfo) *jwt.Token {
	claims := jwt.MapClaims{
		"sub":      data.Subject,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(data.Expiration * time.Hour).Unix(),
		"username": data.UserName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token
}

// Sign receives a token and creates a signature using the '*JWTService.secret' key
func (gen *JWTService) Sign(token *jwt.Token) (string, error) {
	tokenStr, err := token.SignedString(gen.secret)
	if err != nil {
		return "", fmt.Errorf("get signing string: %w", err)
	}
	return tokenStr, nil
}

// Validate receives a string jwt token, parses the claims and validates the token signature
func (gen *JWTService) Validate(token string) (jwt.MapClaims, error) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return gen.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt parse: %w: %w", err, ErrTokenNotValid)
	}

	if !jwtToken.Valid {
		return nil, ErrTokenNotValid
	}

	var claims jwt.MapClaims
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("jwt claims type assertion failed")
	}

	if expVal, ok := claims["exp"].(float64); ok {
		if int64(expVal) < TimeNow().Unix() {
			return nil, fmt.Errorf("token expired at %v: %w", time.Unix(int64(expVal), 0), ErrTokenExpired)
		}
	}

	return claims, nil
}
