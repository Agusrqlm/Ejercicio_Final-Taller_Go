package api

import (
	"net/http"

	"parte3/internal/sales"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// salesHandler holds the sales service and implements HTTP handlers for sales operations.
type salesHandler struct {
	salesService *sales.Service
	logger       *zap.Logger
}

// NewSalesHandler creates a new sales handler.
func NewSalesHandler(salesService *sales.Service, logger *zap.Logger) *salesHandler {
	return &salesHandler{
		salesService: salesService,
		logger:       logger,
	}
}

// handleCreateSale handles the POST /sales endpoint.
func (h *salesHandler) handleCreateSale(ctx *gin.Context) {
	var req struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("failed to bind JSON request", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	sale, err := h.salesService.CreateSale(req.UserID, req.Amount)
	if err != nil {
		h.logger.Error("failed to create sale", zap.Error(err), zap.String("user_id", req.UserID), zap.Float64("amount", req.Amount))
		if err.Error() == "amount must be greater than zero" || err.Error() == "user not found" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Consider more specific error handling based en el tipo de error
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sale"})
		return
	}

	ctx.JSON(http.StatusCreated, sale)
}
