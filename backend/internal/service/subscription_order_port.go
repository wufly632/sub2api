package service

import (
	"context"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// SubscriptionOrderFilters represents filters for listing orders.
type SubscriptionOrderFilters struct {
	OrderNo string
	Status  string
	UserID  *int64
	GroupID *int64
}

// SubscriptionOrderRepository provides persistence for subscription orders.
type SubscriptionOrderRepository interface {
	Create(ctx context.Context, order *SubscriptionOrder) error
	GetByID(ctx context.Context, id int64) (*SubscriptionOrder, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*SubscriptionOrder, error)
	List(ctx context.Context, params pagination.PaginationParams, filters SubscriptionOrderFilters) ([]SubscriptionOrder, *pagination.PaginationResult, error)
	Update(ctx context.Context, order *SubscriptionOrder) error
	UpdateStatus(ctx context.Context, id int64, status string, paidAt, canceledAt *time.Time) error
	SetSubscriptionID(ctx context.Context, id int64, subscriptionID int64) error
}
