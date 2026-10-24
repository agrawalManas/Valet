package CustomUtils

import (
	"agrawalManas/Valet/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewJWT(role string, userId string, config *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   userId,
		"user_role": role,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))

	if err != nil {
		return "", err
	}
	return tokenString, nil
}
