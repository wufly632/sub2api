package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionorder"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrOrderNotFound           = infraerrors.NotFound("ORDER_NOT_FOUND", "order not found")
	ErrOrderInvalidStatus      = infraerrors.BadRequest("ORDER_INVALID_STATUS", "order status invalid")
	ErrOrderInvalidPlan        = infraerrors.BadRequest("ORDER_INVALID_PLAN", "invalid purchase plan")
	ErrOrderNilInput           = infraerrors.BadRequest("ORDER_NIL_INPUT", "order input cannot be nil")
	ErrPaymentNotConfigured    = infraerrors.BadRequest("PAYMENT_NOT_CONFIGURED", "payment provider not configured")
	ErrPaymentInvalidSignature = infraerrors.BadRequest("PAYMENT_INVALID_SIGNATURE", "payment signature invalid")
	ErrPaymentAmountMismatch   = infraerrors.BadRequest("PAYMENT_AMOUNT_MISMATCH", "payment amount mismatch")
)

const orderPaymentTimeout = 5 * time.Minute

// SubscriptionOrderService handles purchase orders for subscriptions.
type SubscriptionOrderService struct {
	entClient           *ent.Client
	groupRepo           GroupRepository
	orderRepo           SubscriptionOrderRepository
	subscriptionService *SubscriptionService
	settingService      *SettingService
	xunhuPayClient      *XunhuPayClient
}

// NewSubscriptionOrderService creates a new order service.
func NewSubscriptionOrderService(
	entClient *ent.Client,
	groupRepo GroupRepository,
	orderRepo SubscriptionOrderRepository,
	subscriptionService *SubscriptionService,
	settingService *SettingService,
	xunhuPayClient *XunhuPayClient,
) *SubscriptionOrderService {
	return &SubscriptionOrderService{
		entClient:           entClient,
		groupRepo:           groupRepo,
		orderRepo:           orderRepo,
		subscriptionService: subscriptionService,
		settingService:      settingService,
		xunhuPayClient:      xunhuPayClient,
	}
}

// ListPlans returns purchasable subscription plans.
func (s *SubscriptionOrderService) ListPlans(ctx context.Context) ([]Group, error) {
	return s.groupRepo.ListPurchasePlans(ctx)
}

// CreateOrder creates a new order. Paid plans remain pending until payment is confirmed.
func (s *SubscriptionOrderService) CreateOrder(ctx context.Context, userID, groupID int64, notes string) (*SubscriptionOrder, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if !group.IsActive() || !group.IsSubscriptionType() || !group.PurchaseEnabled || group.PurchasePrice == nil {
		return nil, ErrOrderInvalidPlan
	}

	validityDays := normalizeValidityDays(group.DefaultValidityDays)
	amount := *group.PurchasePrice

	orderNo, err := GenerateOrderNo()
	if err != nil {
		return nil, err
	}

	paymentProvider, cfg, err := s.resolvePaymentConfig(ctx, amount)
	if err != nil {
		return nil, err
	}
	order := &SubscriptionOrder{
		OrderNo:         orderNo,
		UserID:          userID,
		GroupID:         groupID,
		PaymentProvider: paymentProvider,
		Status:          OrderStatusPending,
		Amount:          amount,
		Currency:        CurrencyCNY,
		ValidityDays:    validityDays,
		Notes:           notes,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	if amount <= 0 {
		info := &PaymentResult{Provider: PaymentProviderManual}
		return s.activateOrder(ctx, order, info)
	}

	if paymentProvider == PaymentProviderManual {
		return s.orderRepo.GetByID(ctx, order.ID)
	}

	if paymentProvider == PaymentProviderXunhuPay {
		title := fmt.Sprintf("%s %d天", group.Name, validityDays)
		payResp, err := s.xunhuPayClient.CreatePayment(ctx, *cfg, order, title)
		if err != nil {
			canceledAt := time.Now()
			_ = s.orderRepo.UpdateStatus(ctx, order.ID, OrderStatusCanceled, nil, &canceledAt)
			return nil, err
		}
		order.PaymentURL = payResp.URL
		order.PaymentQRCode = payResp.URLQRCode
		if cfg.Plugins != "" {
			order.PaymentPlugin = cfg.Plugins
		}
		if err := s.orderRepo.Update(ctx, order); err != nil {
			return nil, err
		}
		return s.orderRepo.GetByID(ctx, order.ID)
	}

	return nil, ErrPaymentNotConfigured
}

// ListUserOrders lists orders for a user.
func (s *SubscriptionOrderService) ListUserOrders(ctx context.Context, userID int64, page, pageSize int, status string) ([]SubscriptionOrder, *pagination.PaginationResult, error) {
	if err := s.cancelExpiredPending(ctx); err != nil {
		log.Printf("[Order] cancel expired pending failed: %v", err)
	}
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	filters := SubscriptionOrderFilters{
		Status: status,
		UserID: &userID,
	}
	return s.orderRepo.List(ctx, params, filters)
}

// ListOrders lists orders with filters (admin).
func (s *SubscriptionOrderService) ListOrders(ctx context.Context, page, pageSize int, filters SubscriptionOrderFilters) ([]SubscriptionOrder, *pagination.PaginationResult, error) {
	if err := s.cancelExpiredPending(ctx); err != nil {
		log.Printf("[Order] cancel expired pending failed: %v", err)
	}
	params := pagination.PaginationParams{Page: page, PageSize: pageSize}
	return s.orderRepo.List(ctx, params, filters)
}

// GetOrderByID returns an order by ID.
func (s *SubscriptionOrderService) GetOrderByID(ctx context.Context, id int64) (*SubscriptionOrder, error) {
	if err := s.cancelExpiredPending(ctx); err != nil {
		log.Printf("[Order] cancel expired pending failed: %v", err)
	}
	return s.orderRepo.GetByID(ctx, id)
}

// MarkPaid marks an order as paid.
func (s *SubscriptionOrderService) MarkPaid(ctx context.Context, id int64) (*SubscriptionOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	info := &PaymentResult{Provider: order.PaymentProvider}
	if info.Provider == "" {
		info.Provider = PaymentProviderManual
	}
	return s.activateOrder(ctx, order, info)
}

// Cancel cancels an order.
func (s *SubscriptionOrderService) Cancel(ctx context.Context, id int64) (*SubscriptionOrder, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.Status != OrderStatusPending {
		return nil, ErrOrderInvalidStatus
	}
	canceledAt := time.Now()
	if err := s.orderRepo.UpdateStatus(ctx, id, OrderStatusCanceled, nil, &canceledAt); err != nil {
		return nil, err
	}
	return s.orderRepo.GetByID(ctx, id)
}

// HandleXunhuPayNotify handles payment notify callback.
func (s *SubscriptionOrderService) HandleXunhuPayNotify(ctx context.Context, payload XunhuPayNotifyPayload) error {
	if err := s.cancelExpiredPending(ctx); err != nil {
		log.Printf("[Order] cancel expired pending failed: %v", err)
	}
	if s.settingService == nil || s.xunhuPayClient == nil {
		return ErrPaymentNotConfigured
	}
	settings, err := s.settingService.GetAllSettings(ctx)
	if err != nil {
		return err
	}
	if settings.PaymentProvider != PaymentProviderXunhuPay {
		return ErrPaymentNotConfigured
	}
	if settings.XunhuPayAppSecret == "" {
		return ErrPaymentNotConfigured
	}
	if payload.AppID != settings.XunhuPayAppID {
		return ErrPaymentInvalidSignature
	}
	if !s.xunhuPayClient.VerifyNotify(payload, settings.XunhuPayAppSecret) {
		return ErrPaymentInvalidSignature
	}
	if payload.Status != "OD" {
		return nil
	}

	order, err := s.orderRepo.GetByOrderNo(ctx, payload.TradeOrderID)
	if err != nil {
		return err
	}
	if order.Status == OrderStatusPaid {
		return nil
	}
	if order.Status != OrderStatusPending {
		return ErrOrderInvalidStatus
	}

	if payload.TotalFee != "" {
		if amount, err := strconv.ParseFloat(payload.TotalFee, 64); err == nil {
			if !amountMatches(order.Amount, amount) {
				return ErrPaymentAmountMismatch
			}
		}
	}

	info := &PaymentResult{
		Provider:      PaymentProviderXunhuPay,
		TransactionID: payload.TransactionID,
		OpenOrderID:   payload.OpenOrderID,
		Plugin:        payload.Plugins,
	}
	returnValue, err := s.activateOrder(ctx, order, info)
	if err != nil {
		return err
	}
	_ = returnValue
	return nil
}

type PaymentResult struct {
	Provider      string
	TransactionID string
	OpenOrderID   string
	Plugin        string
}

func (s *SubscriptionOrderService) activateOrder(ctx context.Context, order *SubscriptionOrder, info *PaymentResult) (*SubscriptionOrder, error) {
	if order.Status != OrderStatusPending {
		return nil, ErrOrderInvalidStatus
	}
	paidAt := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := ent.NewTxContext(ctx, tx)

	sub, _, err := s.subscriptionService.AssignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
		UserID:       order.UserID,
		GroupID:      order.GroupID,
		ValidityDays: order.ValidityDays,
		AssignedBy:   0,
		Notes:        fmt.Sprintf("订单 %s", order.OrderNo),
	})
	if err != nil {
		return nil, err
	}

	order.Status = OrderStatusPaid
	order.PaidAt = &paidAt
	order.SubscriptionID = &sub.ID
	if info != nil {
		if info.Provider != "" {
			order.PaymentProvider = info.Provider
		}
		if info.TransactionID != "" {
			order.PaymentTransactionID = info.TransactionID
		}
		if info.OpenOrderID != "" {
			order.PaymentOpenOrderID = info.OpenOrderID
		}
		if info.Plugin != "" {
			order.PaymentPlugin = info.Plugin
		}
	}

	if err := s.orderRepo.Update(txCtx, order); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return s.orderRepo.GetByID(ctx, order.ID)
}

func (s *SubscriptionOrderService) resolvePaymentConfig(ctx context.Context, amount float64) (string, *XunhuPayConfig, error) {
	if amount <= 0 {
		return PaymentProviderManual, nil, nil
	}
	if s.settingService == nil || s.xunhuPayClient == nil {
		return "", nil, ErrPaymentNotConfigured
	}
	settings, err := s.settingService.GetAllSettings(ctx)
	if err != nil {
		return "", nil, err
	}
	provider := strings.TrimSpace(settings.PaymentProvider)
	if provider == "" {
		provider = PaymentProviderManual
	}
	if provider == PaymentProviderManual {
		return provider, nil, nil
	}
	if provider != PaymentProviderXunhuPay {
		return provider, nil, ErrPaymentNotConfigured
	}
	cfg := &XunhuPayConfig{
		AppID:     settings.XunhuPayAppID,
		AppSecret: settings.XunhuPayAppSecret,
		Gateway:   settings.XunhuPayGateway,
		NotifyURL: settings.XunhuPayNotifyURL,
		ReturnURL: settings.XunhuPayReturnURL,
		Plugins:   settings.XunhuPayPlugins,
	}
	if cfg.AppID == "" || cfg.AppSecret == "" || cfg.Gateway == "" || cfg.NotifyURL == "" {
		return provider, nil, ErrPaymentNotConfigured
	}
	return provider, cfg, nil
}

func (s *SubscriptionOrderService) cancelExpiredPending(ctx context.Context) error {
	if s.entClient == nil {
		return nil
	}
	cutoff := time.Now().Add(-orderPaymentTimeout)
	canceledAt := time.Now()
	_, err := s.entClient.SubscriptionOrder.Update().
		Where(
			subscriptionorder.StatusEQ(OrderStatusPending),
			subscriptionorder.CreatedAtLT(cutoff),
		).
		SetStatus(OrderStatusCanceled).
		SetCanceledAt(canceledAt).
		Save(ctx)
	return err
}

func amountMatches(expected, actual float64) bool {
	diff := expected - actual
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.01
}

func normalizeValidityDays(days int) int {
	if days <= 0 {
		days = 30
	}
	if days > MaxValidityDays {
		days = MaxValidityDays
	}
	return days
}
