package middleware

import (
	"errors"
	"net/http"

	"booking-service/internal/dto"
	"booking-service/internal/repository"

	"github.com/gin-gonic/gin"
)

func Error() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) == 0 || ctx.Writer.Written() {
			return
		}

		err := ctx.Errors.Last().Err
		status, message := resolveHTTPError(err)
		ctx.JSON(status, dto.Response{
			Success: false,
			Message: message,
			Errors:  nil,
		})
	}
}

func resolveHTTPError(err error) (int, string) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, repository.ErrInsufficientStock):
		return http.StatusConflict, err.Error()
	case errors.Is(err, repository.ErrQueuePublish):
		return http.StatusServiceUnavailable, "queue service unavailable"
	case errors.Is(err, repository.ErrInvalidCheckout):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, repository.ErrExpiredCheckout):
		return http.StatusGone, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
