package router

import (
	"agrawalManas/Valet/internal/config"
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/handler"
	"agrawalManas/Valet/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Initialize(cfg *config.Config, pool *pgxpool.Pool) {
	println("Initializing router")

	router := gin.Default()
	router.SetTrustedProxies(nil)
	setupApiRoutes(router, pool, cfg)

	router.Run("localhost:" + cfg.PORT)
}

func setupApiRoutes(router *gin.Engine, pool *pgxpool.Pool, cfg *config.Config) {
	authApi := router.Group("/api/auth")
	{
		authApi.GET("/get-otp", handler.GenerateOTPHandler(pool))
		authApi.POST("/customer/verify-otp", handler.VerifyOTPHandler(pool, cfg, constants.ROLE_CUSTOMER))
		authApi.POST("/supervisor/verify-otp", handler.VerifyOTPHandler(pool, cfg, constants.ROLE_SUPERVISOR))
		authApi.POST("/driver/verify-otp", handler.VerifyOTPHandler(pool, cfg, constants.ROLE_DRIVER))
		authApi.POST("/admin/verify-otp", handler.VerifyOTPHandler(pool, cfg, constants.ROLE_ADMIN))
	}

	userApi := router.Group("api/user")
	userApi.Use(middleware.AuthMiddleware(pool, cfg))
	{
		userApi.PATCH("/update-profile", handler.UpdateProfileHandler(pool))
	}

	adminApi := router.Group("api/admin")
	adminApi.Use(middleware.AuthMiddleware(pool, cfg), middleware.RequireRoles(constants.ROLE_ADMIN))
	{
		adminApi.PATCH("/user-access", handler.AdminUpdateUserHandler(pool))
		adminApi.POST("/add-site", handler.AddValetSiteHandler(pool))
		adminApi.GET("/get-sites", handler.GetValetSitesHandler(pool))
		adminApi.PATCH("/update-site/:id", handler.UpdateSiteHandler(pool))
	}
}
