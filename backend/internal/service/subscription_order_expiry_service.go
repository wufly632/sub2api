package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/subscriptionorder"
)

// SubscriptionOrderExpiryService periodically cancels expired pending orders.
type SubscriptionOrderExpiryService struct {
	entClient *ent.Client
	interval  time.Duration
	stopCh    chan struct{}
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

func NewSubscriptionOrderExpiryService(entClient *ent.Client, interval time.Duration) *SubscriptionOrderExpiryService {
	return &SubscriptionOrderExpiryService{
		entClient: entClient,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

func (s *SubscriptionOrderExpiryService) Start() {
	if s == nil || s.entClient == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *SubscriptionOrderExpiryService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *SubscriptionOrderExpiryService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cutoff := time.Now().Add(-orderPaymentTimeout)
	canceledAt := time.Now()
	updated, err := s.entClient.SubscriptionOrder.Update().
		Where(
			subscriptionorder.StatusEQ(OrderStatusPending),
			subscriptionorder.CreatedAtLT(cutoff),
		).
		SetStatus(OrderStatusCanceled).
		SetCanceledAt(canceledAt).
		Save(ctx)
	if err != nil {
		log.Printf("[OrderExpiry] cancel expired pending failed: %v", err)
		return
	}
	if updated > 0 {
		log.Printf("[OrderExpiry] canceled %d expired pending orders", updated)
	}
}
