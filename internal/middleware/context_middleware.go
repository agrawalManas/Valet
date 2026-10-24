package middleware

import (
	"agrawalManas/Valet/internal/constants"

	"github.com/gin-gonic/gin"
)

type AuthUser struct {
	ID       string
	Role     string
	IsActive bool
}

func CurrentUser(ctx *gin.Context) (AuthUser, bool) {
	v, ok := ctx.Get(constants.CTX_AUTH_USER)
	if !ok {
		return AuthUser{}, false
	}
	u, ok := v.(AuthUser)
	return u, ok
}

// For handlers behind Auth. A missing user means a route was registered outside
// the protected group, which is a programmer error, so panic (500 + log) is correct.
func MustCurrentUser(ctx *gin.Context) AuthUser {
	u, ok := CurrentUser(ctx)
	if !ok {
		panic("MustCurrentUser called on a route without the Auth middleware")
	}
	return u
}
