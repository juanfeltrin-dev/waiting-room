package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaims struct {
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

func NewSessionID() string {
	return uuid.New().String()
}
