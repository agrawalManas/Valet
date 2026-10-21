package router

import "github.com/gin-gonic/gin"

func Initialize(port string) {
	println("Initializing router")

	router := gin.Default()
	router.SetTrustedProxies(nil)
	setupApiRoutes(router)
	println("Router Initialized")

	router.Run("localhost:" + port)
}

func setupApiRoutes(router *gin.Engine) {
	authApi := router.Group("/api/auth")
	{
		authApi.POST("/get-otp")
		authApi.POST("login")
	}
}
