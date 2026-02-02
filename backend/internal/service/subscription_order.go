package service

import (
	"crypto/rand"
	"fmt"
	"time"
)

// SubscriptionOrder represents a purchase order for a subscription plan.
type SubscriptionOrder struct {
	ID                   int64
	OrderNo              string
	UserID               int64
	GroupID              int64
	SubscriptionID       *int64
	PaymentProvider      string
	PaymentURL           string
	PaymentQRCode        string
	PaymentOpenOrderID   string
	PaymentTransactionID string
	PaymentPlugin        string
	Status               string
	Amount               float64
	Currency             string
	ValidityDays         int
	PaidAt               *time.Time
	CanceledAt           *time.Time
	Notes                string
	CreatedAt            time.Time
	UpdatedAt            time.Time

	User         *User
	Group        *Group
	Subscription *UserSubscription
}

// GenerateOrderNo creates a unique order number.
func GenerateOrderNo() (string, error) {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("S%s%02x%02x%02x", time.Now().UTC().Format("20060102150405"), buf[0], buf[1], buf[2]), nil
}
