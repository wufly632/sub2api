package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// PurchaseHandler handles user purchase operations.
type PurchaseHandler struct {
	orderService *service.SubscriptionOrderService
}

// NewPurchaseHandler creates a new purchase handler.
func NewPurchaseHandler(orderService *service.SubscriptionOrderService) *PurchaseHandler {
	return &PurchaseHandler{orderService: orderService}
}

// ListPlans handles listing purchasable plans
// GET /api/v1/purchase/plans
func (h *PurchaseHandler) ListPlans(c *gin.Context) {
	plans, err := h.orderService.ListPlans(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.Group, 0, len(plans))
	for i := range plans {
		out = append(out, *dto.GroupFromServiceShallow(&plans[i]))
	}
	response.Success(c, out)
}

// CreateOrder handles creating a new order
// POST /api/v1/purchase/orders
func (h *PurchaseHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	var req struct {
		GroupID int64  `json:"group_id" binding:"required"`
		Notes   string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), subject.UserID, req.GroupID, req.Notes)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.SubscriptionOrderFromService(order))
}

// ListOrders handles listing current user's orders
// GET /api/v1/purchase/orders
func (h *PurchaseHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")

	orders, pagination, err := h.orderService.ListUserOrders(c.Request.Context(), subject.UserID, page, pageSize, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]dto.SubscriptionOrder, 0, len(orders))
	for i := range orders {
		out = append(out, *dto.SubscriptionOrderFromService(&orders[i]))
	}
	response.Paginated(c, out, pagination.Total, page, pageSize)
}

// GetOrder handles getting current user's order by ID
// GET /api/v1/purchase/orders/:id
func (h *PurchaseHandler) GetOrder(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
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
	if order.UserID != subject.UserID {
		response.Forbidden(c, "Forbidden")
		return
	}
	response.Success(c, dto.SubscriptionOrderFromService(order))
}
