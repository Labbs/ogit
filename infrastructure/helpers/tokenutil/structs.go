package tokenutil

import "github.com/golang-jwt/jwt/v4"

type JwtCustomClaims struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	jwt.RegisteredClaims
}

type JwtCustomRefreshClaims struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	jwt.RegisteredClaims
}
