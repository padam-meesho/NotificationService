package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/padam-meesho/NotificationService/internal/handlers"
	"github.com/padam-meesho/NotificationService/internal/middlewares"
)

func SetUpRoutes() {
	// create a router.
	// define a base route and try to group routes, and within that grouping apply the middleware.
	// now try to define the different endpoints
	router := gin.Default()
	router.GET("/health", healthHandler)
	api := router.Group("/v1", middlewares.AuthCheck(), middlewares.TraceMiddleware()) // this is to add the base route and apply middleware on it.

	// sms apis
	smsApi := api.Group("/sms") // these need to be put in the route handlers
	smsApi.POST("/send", handlers.SendSmsController)
	smsApi.GET("/:request_id", handlers.GetSmsController) // this shall act as a path variable

	// blacklist apis
	blacklistApi := api.Group("/blacklist")
	blacklistApi.GET("", handlers.GetBlacklistController)
	blacklistApi.POST("", handlers.AddToBlacklistController)
	blacklistApi.DELETE("/:number", handlers.RemoveFromBlacklistController) // this shall act as the route which shall be hit to remove a number from a blacklist.

	router.Run(":3333")
}

func healthHandler(c *gin.Context) {
	// note that all gin methods are capitalized.
	c.JSON(200, gin.H{
		"system": "up",
	})
}
