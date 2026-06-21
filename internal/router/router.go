package router

import (
	"booking-service/internal/handler"
	"booking-service/internal/middleware"
	"booking-service/internal/repository"
	"booking-service/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, allowedOrigins []string) *gin.Engine {
	txManager := repository.NewTransactionManager(db)
	eventRepo := repository.NewEventRepository(db)
	ticketCategoryRepo := repository.NewTicketCategoryRepository(db)
	ticketStockRepo := repository.NewTicketStockRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	bookingItemRepo := repository.NewBookingItemRepository(db)
	guestRepo := repository.NewGuestRepository(db)

	eventService := service.NewEventService(eventRepo)
	ticketCategoryService := service.NewTicketCategoryService(ticketCategoryRepo)
	ticketStockService := service.NewTicketStockService(ticketStockRepo)
	guestService := service.NewGuestService(guestRepo)
	bookingItemService := service.NewBookingItemService(bookingItemRepo)
	bookingService := service.NewBookingService(
		txManager,
		bookingRepo,
		eventService,
		ticketCategoryService,
		ticketStockService,
		guestService,
		bookingItemService,
	)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.CORS(allowedOrigins), middleware.Error())

	api := router.Group("/api/v1")
	handler.New(bookingService).RegisterRoutes(api)

	return router
}
