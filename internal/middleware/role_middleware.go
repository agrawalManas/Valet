package middleware

import (
	"agrawalManas/Valet/internal/constants"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireRoles(allowed ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		set[r] = struct{}{}
	}
	return func(ctx *gin.Context) {
		u, ok := CurrentUser(ctx)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, CustomUtils.HandleErrorResponse(constants.UNAUTHORIZED))
			return
		}
		if _, allowed := set[u.Role]; !allowed {
			ctx.AbortWithStatusJSON(http.StatusForbidden, CustomUtils.HandleErrorResponse(constants.FORBIDDEN))
			return
		}
		ctx.Next()
	}
}
