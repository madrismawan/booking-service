package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"booking-service/internal/dto"
	"booking-service/internal/mapper"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getTicketStock(ctx *gin.Context) {
	ticketCategoryID, err := strconv.ParseInt(ctx.Param("ticket_category_id"), 10, 64)
	if err != nil || ticketCategoryID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "invalid ticket category id",
			Errors:  nil,
		})
		return
	}

	stock, err := h.ticketStockService.FindByTicketCategoryID(ticketCategoryID)
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	etag := ticketStockETag(stock.TicketCategoryID, stock.Version)
	ctx.Header("ETag", etag)
	ctx.Header("Cache-Control", "no-cache")

	if etagMatches(ctx.GetHeader("If-None-Match"), etag) {
		ctx.Status(http.StatusNotModified)
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "ticket stock retrieved",
		Data:    mapper.ToTicketStockResponse(stock),
	})
}

func ticketStockETag(ticketCategoryID, version int64) string {
	return fmt.Sprintf(`"ticket-stock-%d-v%d"`, ticketCategoryID, version)
}

func etagMatches(ifNoneMatch, currentETag string) bool {
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}
		candidate = strings.TrimPrefix(candidate, "W/")
		if candidate == currentETag {
			return true
		}
	}
	return false
}
