package CustomUtils

import (
	"agrawalManas/Valet/internal/config"
	"agrawalManas/Valet/internal/constants"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewJWT(role string, userId string, isActive bool, config *config.Config) (string, error) {
	claims := jwt.MapClaims{
		constants.CLAIM_USER_ID:   userId,
		constants.CLAIM_USER_ROLE: role,
		constants.CLAIM_IS_ACTIVE: isActive,
		constants.CLAIM_EXP:       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))

	if err != nil {
		return "", err
	}
	return tokenString, nil
}
