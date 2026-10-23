package middleware

import (
	"agrawalManas/Valet/internal/config"
	"agrawalManas/Valet/internal/constants"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func AuthMiddleware(pool *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")

		if header == "" {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.NO_AUTH_HEADER))
			ctx.Abort()
			return
		}

		trimmedHeader := strings.TrimPrefix(header, "Bearer ")
		if trimmedHeader == "" || trimmedHeader == header {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.MALFORMED_TOKEN))
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(trimmedHeader, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf(constants.UNEXPECTED_SIGNING_METHOD+": %v", t.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(err.Error()))
			ctx.Abort()
			return
		}

		tokenClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.NO_TOKEN_CLAIMS_FOUND))
			ctx.Abort()
			return
		}

		userId, ok := tokenClaims["user_id"].(string)
		if !ok {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID_TOKEN_CLAIM+"[user_id]"))
			ctx.Abort()
			return
		}

		exp, ok := tokenClaims["exp"].(float64)
		if !ok {
			ctx.JSON(http.StatusBadRequest, CustomUtils.HandleErrorResponse(constants.INVALID_TOKEN_CLAIM+"[exp]"))
			ctx.Abort()
			return
		}

		expirationTime := time.Unix(int64(exp), 0)
		if time.Now().After(expirationTime) {
			ctx.JSON(http.StatusUnauthorized, CustomUtils.HandleErrorResponse(constants.UNAUTHORIZED))
			ctx.Abort()
			return
		}

		ctx.Set("user_id", userId)
		ctx.Next()
	}
}
