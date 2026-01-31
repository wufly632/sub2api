package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles admin subscription order management.
type OrderHandler struct {
	orderService *service.SubscriptionOrderService
}

// NewOrderHandler creates a new order handler.
func NewOrderHandler(orderService *service.SubscriptionOrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// List handles listing all orders with pagination
// GET /api/v1/admin/orders
func (h *OrderHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")
	orderNo := strings.TrimSpace(c.Query("order_no"))

	var userID *int64
	if v := c.Query("user_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			userID = &id
		}
	}
	var groupID *int64
	if v := c.Query("group_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			groupID = &id
		}
	}

	orders, pagination, err := h.orderService.ListOrders(c.Request.Context(), page, pageSize, service.SubscriptionOrderFilters{
		OrderNo: orderNo,
		Status:  status,
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.AdminSubscriptionOrder, 0, len(orders))
	for i := range orders {
		out = append(out, *dto.SubscriptionOrderFromServiceAdmin(&orders[i]))
	}
	response.Paginated(c, out, pagination.Total, page, pageSize)
}

// GetByID handles getting an order by ID
// GET /api/v1/admin/orders/:id
func (h *OrderHandler) GetByID(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.SubscriptionOrderFromServiceAdmin(order))
}

// MarkPaid handles marking an order as paid
// POST /api/v1/admin/orders/:id/mark-paid
func (h *OrderHandler) MarkPaid(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	order, err := h.orderService.MarkPaid(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.SubscriptionOrderFromServiceAdmin(order))
}

// Cancel handles canceling an order
// POST /api/v1/admin/orders/:id/cancel
func (h *OrderHandler) Cancel(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	order, err := h.orderService.Cancel(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.SubscriptionOrderFromServiceAdmin(order))
}
