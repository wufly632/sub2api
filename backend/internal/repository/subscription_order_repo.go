package repository

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionorder"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type subscriptionOrderRepository struct {
	client *ent.Client
}

func NewSubscriptionOrderRepository(client *ent.Client) service.SubscriptionOrderRepository {
	return &subscriptionOrderRepository{client: client}
}

func (r *subscriptionOrderRepository) Create(ctx context.Context, order *service.SubscriptionOrder) error {
	if order == nil {
		return service.ErrOrderNilInput
	}
	client := clientFromContext(ctx, r.client)
	created, err := client.SubscriptionOrder.Create().
		SetOrderNo(order.OrderNo).
		SetUserID(order.UserID).
		SetGroupID(order.GroupID).
		SetNillableSubscriptionID(order.SubscriptionID).
		SetPaymentProvider(order.PaymentProvider).
		SetNillablePaymentURL(nullableString(order.PaymentURL)).
		SetNillablePaymentQrcode(nullableString(order.PaymentQRCode)).
		SetNillablePaymentOpenOrderID(nullableString(order.PaymentOpenOrderID)).
		SetNillablePaymentTransactionID(nullableString(order.PaymentTransactionID)).
		SetNillablePaymentPlugin(nullableString(order.PaymentPlugin)).
		SetStatus(order.Status).
		SetAmount(order.Amount).
		SetCurrency(order.Currency).
		SetValidityDays(order.ValidityDays).
		SetNillablePaidAt(order.PaidAt).
		SetNillableCanceledAt(order.CanceledAt).
		SetNotes(order.Notes).
		Save(ctx)
	if err != nil {
		return translatePersistenceError(err, nil, nil)
	}
	applySubscriptionOrderEntityToService(order, created)
	return nil
}

func (r *subscriptionOrderRepository) GetByID(ctx context.Context, id int64) (*service.SubscriptionOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.SubscriptionOrder.Query().
		Where(subscriptionorder.IDEQ(id)).
		WithUser().
		WithGroup().
		WithSubscription().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrOrderNotFound, nil)
	}
	return subscriptionOrderEntityToService(m), nil
}

func (r *subscriptionOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.SubscriptionOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.SubscriptionOrder.Query().
		Where(subscriptionorder.OrderNoEQ(orderNo)).
		WithUser().
		WithGroup().
		WithSubscription().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrOrderNotFound, nil)
	}
	return subscriptionOrderEntityToService(m), nil
}

func (r *subscriptionOrderRepository) List(ctx context.Context, params pagination.PaginationParams, filters service.SubscriptionOrderFilters) ([]service.SubscriptionOrder, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	q := client.SubscriptionOrder.Query()
	if filters.OrderNo != "" {
		q = q.Where(subscriptionorder.OrderNoContainsFold(filters.OrderNo))
	}
	if filters.Status != "" {
		q = q.Where(subscriptionorder.StatusEQ(filters.Status))
	}
	if filters.UserID != nil {
		q = q.Where(subscriptionorder.UserIDEQ(*filters.UserID))
	}
	if filters.GroupID != nil {
		q = q.Where(subscriptionorder.GroupIDEQ(*filters.GroupID))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Order(ent.Desc(subscriptionorder.FieldCreatedAt)).
		Offset(params.Offset()).
		Limit(params.Limit()).
		WithUser().
		WithGroup().
		WithSubscription().
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	out := make([]service.SubscriptionOrder, 0, len(orders))
	for i := range orders {
		out = append(out, *subscriptionOrderEntityToService(orders[i]))
	}
	return out, paginationResultFromTotal(int64(total), params), nil
}

func (r *subscriptionOrderRepository) Update(ctx context.Context, order *service.SubscriptionOrder) error {
	if order == nil {
		return service.ErrOrderNilInput
	}
	client := clientFromContext(ctx, r.client)
	builder := client.SubscriptionOrder.UpdateOneID(order.ID).
		SetPaymentProvider(order.PaymentProvider).
		SetStatus(order.Status).
		SetAmount(order.Amount).
		SetCurrency(order.Currency).
		SetValidityDays(order.ValidityDays).
		SetNotes(order.Notes)

	if order.PaymentURL != "" {
		builder = builder.SetPaymentURL(order.PaymentURL)
	} else {
		builder = builder.ClearPaymentURL()
	}
	if order.PaymentQRCode != "" {
		builder = builder.SetPaymentQrcode(order.PaymentQRCode)
	} else {
		builder = builder.ClearPaymentQrcode()
	}
	if order.PaymentOpenOrderID != "" {
		builder = builder.SetPaymentOpenOrderID(order.PaymentOpenOrderID)
	} else {
		builder = builder.ClearPaymentOpenOrderID()
	}
	if order.PaymentTransactionID != "" {
		builder = builder.SetPaymentTransactionID(order.PaymentTransactionID)
	} else {
		builder = builder.ClearPaymentTransactionID()
	}
	if order.PaymentPlugin != "" {
		builder = builder.SetPaymentPlugin(order.PaymentPlugin)
	} else {
		builder = builder.ClearPaymentPlugin()
	}

	if order.SubscriptionID != nil {
		builder = builder.SetSubscriptionID(*order.SubscriptionID)
	} else {
		builder = builder.ClearSubscriptionID()
	}
	if order.PaidAt != nil {
		builder = builder.SetPaidAt(*order.PaidAt)
	} else {
		builder = builder.ClearPaidAt()
	}
	if order.CanceledAt != nil {
		builder = builder.SetCanceledAt(*order.CanceledAt)
	} else {
		builder = builder.ClearCanceledAt()
	}

	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrOrderNotFound, nil)
}

func (r *subscriptionOrderRepository) UpdateStatus(ctx context.Context, id int64, status string, paidAt, canceledAt *time.Time) error {
	client := clientFromContext(ctx, r.client)
	builder := client.SubscriptionOrder.UpdateOneID(id).SetStatus(status)
	if paidAt != nil {
		builder = builder.SetPaidAt(*paidAt)
	} else {
		builder = builder.ClearPaidAt()
	}
	if canceledAt != nil {
		builder = builder.SetCanceledAt(*canceledAt)
	} else {
		builder = builder.ClearCanceledAt()
	}
	_, err := builder.Save(ctx)
	return translatePersistenceError(err, service.ErrOrderNotFound, nil)
}

func (r *subscriptionOrderRepository) SetSubscriptionID(ctx context.Context, id int64, subscriptionID int64) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.SubscriptionOrder.UpdateOneID(id).
		SetSubscriptionID(subscriptionID).
		Save(ctx)
	return translatePersistenceError(err, service.ErrOrderNotFound, nil)
}

func subscriptionOrderEntityToService(m *ent.SubscriptionOrder) *service.SubscriptionOrder {
	if m == nil {
		return nil
	}
	out := &service.SubscriptionOrder{}
	applySubscriptionOrderEntityToService(out, m)
	if m.Edges.User != nil {
		out.User = userEntityToService(m.Edges.User)
	}
	if m.Edges.Group != nil {
		out.Group = groupEntityToService(m.Edges.Group)
	}
	if m.Edges.Subscription != nil {
		out.Subscription = userSubscriptionEntityToService(m.Edges.Subscription)
	}
	return out
}

func applySubscriptionOrderEntityToService(out *service.SubscriptionOrder, m *ent.SubscriptionOrder) {
	if out == nil || m == nil {
		return
	}
	out.ID = m.ID
	out.OrderNo = m.OrderNo
	out.UserID = m.UserID
	out.GroupID = m.GroupID
	if m.SubscriptionID != nil {
		value := *m.SubscriptionID
		out.SubscriptionID = &value
	} else {
		out.SubscriptionID = nil
	}
	out.PaymentProvider = m.PaymentProvider
	out.PaymentURL = derefString(m.PaymentURL)
	out.PaymentQRCode = derefString(m.PaymentQrcode)
	out.PaymentOpenOrderID = derefString(m.PaymentOpenOrderID)
	out.PaymentTransactionID = derefString(m.PaymentTransactionID)
	out.PaymentPlugin = derefString(m.PaymentPlugin)
	out.Status = m.Status
	out.Amount = m.Amount
	out.Currency = m.Currency
	out.ValidityDays = m.ValidityDays
	out.PaidAt = m.PaidAt
	out.CanceledAt = m.CanceledAt
	out.Notes = derefString(m.Notes)
	out.CreatedAt = m.CreatedAt
	out.UpdatedAt = m.UpdatedAt
}

func nullableString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
