package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"booking-service/internal/dto"
	"booking-service/internal/repository"

	"github.com/gin-gonic/gin"
)

const maxPaymentWebhookBody = 1 << 20

func (h *Handler) paymentWebhook(ctx *gin.Context) {
	body, err := io.ReadAll(http.MaxBytesReader(
		ctx.Writer,
		ctx.Request.Body,
		maxPaymentWebhookBody,
	))
	if err != nil {
		_ = ctx.Error(repository.ErrInvalidPayment)
		return
	}

	if !h.paymentService.VerifySignature(body, ctx.GetHeader("X-Payment-Signature")) {
		_ = ctx.Error(repository.ErrInvalidSignature)
		return
	}

	var req dto.PaymentWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil || !req.Valid() {
		_ = ctx.Error(repository.ErrInvalidPayment)
		return
	}

	result, err := h.paymentService.ProcessWebhook(req, body)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "payment webhook processed",
		Data: dto.PaymentWebhookResponse{
			PaymentTransactionID: result.Payment.ID,
			BookingID:            result.Booking.ID,
			BookingStatus:        result.Booking.Status,
			PaymentStatus:        result.Payment.Status,
			Duplicate:            result.Duplicate,
		},
	})
}
