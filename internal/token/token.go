package token

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

type Token struct {
	jwt.StandardClaims
	NodeID string `json:"node_id"`
}

func GenerateJWT(signingKey []byte, claims Token) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("error signing JWT: %w", err)
	}
	return str, nil
}

func ParseJWT(tokenStr string, signingKey []byte) (*Token, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Token{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing method")
		}
		return signingKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing JWT: %w", err)
	}

	return token.Claims.(*Token), nil
}
