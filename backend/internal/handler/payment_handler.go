package handler

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// PaymentHandler handles payment callbacks.
type PaymentHandler struct {
	orderService *service.SubscriptionOrderService
}

// NewPaymentHandler creates a payment handler.
func NewPaymentHandler(orderService *service.SubscriptionOrderService) *PaymentHandler {
	return &PaymentHandler{orderService: orderService}
}

// XunhuPayNotify handles XunhuPay notify callbacks.
// POST /api/v1/payment/xunhupay/notify
func (h *PaymentHandler) XunhuPayNotify(c *gin.Context) {
	payload := service.XunhuPayNotifyPayload{
		AppID:         c.PostForm("appid"),
		TradeOrderID:  c.PostForm("trade_order_id"),
		TotalFee:      c.PostForm("total_fee"),
		TransactionID: c.PostForm("transaction_id"),
		OpenOrderID:   c.PostForm("open_order_id"),
		OrderTitle:    c.PostForm("order_title"),
		Status:        c.PostForm("status"),
		Plugins:       c.PostForm("plugins"),
		Attach:        c.PostForm("attach"),
		Time:          c.PostForm("time"),
		NonceStr:      c.PostForm("nonce_str"),
		Hash:          c.PostForm("hash"),
	}

	if err := h.orderService.HandleXunhuPayNotify(c.Request.Context(), payload); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}
