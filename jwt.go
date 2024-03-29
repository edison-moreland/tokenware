package tokenware

// echo.labstack.com/cookbook/jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Generate creates and signs a new JWT
func Generate(identity interface{}) (string, error) {
	config := pkgConfig()

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Add identity and expiration to claims
	claims := token.Claims.(jwt.MapClaims)
	claims[config.IdentityClaim] = identity
	claims["exp"] = time.Now().Add(config.TimeToLive).Unix()

	// Sign token
	tokenSigned, err := token.SignedString([]byte(config.SigningKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %#v", err.Error())
	}

	return tokenSigned, nil
}

// Validate de-signs and parses a JWTString and returns the identity associated
func Validate(tokenString string) (interface{}, error) {
	config := pkgConfig()

	if IsRevoked(tokenString) {
		return nil, errors.New("token has been revoked")
	}

	// Decrypt and parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check JWT signing method is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Everything looks good, return signing key
		return []byte(config.SigningKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not validate token (%v). Reason: %v", tokenString, err.Error())
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Grab identity from JWT claims
		return claims[config.IdentityClaim], nil
		// TODO add option for function that gets identity model from database?
	}
	return nil, fmt.Errorf("could not validate token (%v)", tokenString)
}

// GetRawToken grabs the token from request headers
func GetRawToken(r *http.Request) (string, error) {
	config := pkgConfig()

	rawToken := r.Header.Get(config.Header)
	if !strings.HasPrefix(rawToken, config.HeaderPrefix) {
		return "", errors.New("token not in headers")
	}

	token := strings.TrimPrefix(rawToken, config.HeaderPrefix)

	return token, nil
}

// ValidateFromRequest gets token from request headers and validates it
func ValidateFromRequest(r *http.Request) (interface{}, error) {
	rawToken, err := GetRawToken(r)
	if err != nil {
		return nil, err
	}

	return Validate(rawToken)
}
