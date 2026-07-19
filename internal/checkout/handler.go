package checkout

import (
	domain "FairCheckout/internal/domain"
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CheckoutRequest struct {
	ProductID     string `json:"product_id" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	PaymentMethod string `json:"payment_mehtod" binding:"required"`

	ShippingAddress domain.ShippingAddress `json:"shipping_address" binding:"required"`
}

type Processor interface {
	ProcessCheckout(ctx context.Context, cr CheckoutRequest) Result
}

type Handler struct {
	service Processor
}

func NewHandler(s Processor) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) Checkout(gctx *gin.Context) {
	var checkoutRequest CheckoutRequest
	if err := gctx.ShouldBindJSON(&checkoutRequest); err != nil {
		slog.Error(
			"CheckoutRequest cannot be parsed",
			"error", err,
		)
		gctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx := gctx.Request.Context()
	checkoutResponse := h.service.ProcessCheckout(ctx, checkoutRequest)

	checkoutStatus := checkoutResponse.EventID.StatusCode()
	gctx.JSON(checkoutStatus, gin.H{
		"order_status": checkoutResponse.EventID.String(),
	})
}
