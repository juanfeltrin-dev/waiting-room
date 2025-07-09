package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
	"waitingroom/internal/infra/container"
	"waitingroom/internal/schemas/auth"
)

func CreateToken(sessionID string) (string, error) {
	customClaims := auth.CustomClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "sala-espera-api",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	tokenString, err := token.SignedString(container.GetSecretKey())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*auth.CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &auth.CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return container.GetSecretKey(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*auth.CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func ParseUnverifiedToken(tokenString string) (*auth.CustomClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &auth.CustomClaims{})
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(*auth.CustomClaims)

	return claims, err
}
