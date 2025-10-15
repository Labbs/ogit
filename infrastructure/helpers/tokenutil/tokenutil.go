package tokenutil

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labbs/ogit/infrastructure/config"
)

func CreateAccessToken(user_id, sessionId string, config config.Config) (accessToken string, err error) {
	exp := time.Now().Add(time.Minute * time.Duration(config.Session.ExpirationMinutes)).Unix()
	claims := &JwtCustomClaims{
		SessionID: sessionId,
		UserID:    user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Session.Issuer,
			ExpiresAt: &jwt.NumericDate{Time: time.Unix(exp, 0)},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(config.Session.SecretKey))
	if err != nil {
		return "", err
	}
	return t, nil
}
