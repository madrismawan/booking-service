package router

import (
	"booking-service/internal/app"
	"booking-service/internal/handler"
	"booking-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(container *app.Container, allowedOrigins []string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORS(allowedOrigins), middleware.Error())

	api := router.Group("/api/v1")
	handler.New(container.BookingService, container.WaitingRoomService).RegisterRoutes(api)

	return router
}
