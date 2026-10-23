package router

import (
	"agrawalManas/Valet/internal/config"
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
	println("Router Initialized")

	router.Run("localhost:" + cfg.PORT)
}

func setupApiRoutes(router *gin.Engine, pool *pgxpool.Pool, cfg *config.Config) {
	authApi := router.Group("/api/auth")
	{
		authApi.GET("/get-otp", handler.GenerateOTPHandler(pool))
		authApi.POST("/verify-otp", handler.VerifyOTPHandler(pool, cfg))
	}

	userApi := router.Group("api/user")
	userApi.Use(middleware.AuthMiddleware(pool, cfg))
	{
		userApi.PATCH("/update", handler.UpdateUserHandler(pool))
	}
}
