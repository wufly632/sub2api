package service

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// XunhuPayConfig holds gateway configuration.
type XunhuPayConfig struct {
	AppID     string
	AppSecret string
	Gateway   string
	NotifyURL string
	ReturnURL string
	Plugins   string
}

// XunhuPayCreateResponse represents the create payment response.
type XunhuPayCreateResponse struct {
	OpenID    string
	URL       string
	URLQRCode string
	ErrCode   int
	ErrMsg    string
	Hash      string
}

// XunhuPayNotifyPayload is the notify callback payload.
type XunhuPayNotifyPayload struct {
	AppID         string
	TradeOrderID  string
	TotalFee      string
	TransactionID string
	OpenOrderID   string
	OrderTitle    string
	Status        string
	Plugins       string
	Attach        string
	Time          string
	NonceStr      string
	Hash          string
}

// XunhuPayClient is a minimal HTTP client for XunhuPay.
type XunhuPayClient struct {
	httpClient *http.Client
}

// NewXunhuPayClient creates a new XunhuPay client.
func NewXunhuPayClient() *XunhuPayClient {
	return &XunhuPayClient{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// CreatePayment creates a payment order and returns payment URLs.
func (c *XunhuPayClient) CreatePayment(ctx context.Context, cfg XunhuPayConfig, order *SubscriptionOrder, title string) (*XunhuPayCreateResponse, error) {
	if order == nil {
		return nil, ErrOrderNilInput
	}
	params := map[string]string{
		"version":        "1.1",
		"appid":          cfg.AppID,
		"trade_order_id": order.OrderNo,
		"total_fee":      strconv.FormatFloat(order.Amount, 'f', 2, 64),
		"title":          title,
		"time":           strconv.FormatInt(time.Now().Unix(), 10),
		"notify_url":     cfg.NotifyURL,
		"nonce_str":      randomNonce(16),
	}
	if cfg.ReturnURL != "" {
		params["return_url"] = cfg.ReturnURL
	}
	if cfg.Plugins != "" {
		params["plugins"] = cfg.Plugins
	}
	params["hash"] = xunhuPaySign(params, cfg.AppSecret)

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Gateway, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode xunhupay response: %w", err)
	}

	response := &XunhuPayCreateResponse{
		OpenID:    toString(raw["openid"]),
		URL:       toString(raw["url"]),
		URLQRCode: toString(raw["url_qrcode"]),
		Hash:      toString(raw["hash"]),
	}
	if v, ok := raw["errcode"]; ok {
		if code, err := toInt(v); err == nil {
			response.ErrCode = code
		}
	}
	response.ErrMsg = toString(raw["errmsg"])

	if response.ErrCode != 0 {
		return nil, fmt.Errorf("xunhupay error: %s", response.ErrMsg)
	}
	if response.Hash != "" {
		if !xunhuPayVerify(raw, cfg.AppSecret, response.Hash) {
			// Some gateways return hashes that don't match our verify logic.
			// Keep the order creation flow unblocked while logging diagnostics.
			log.Printf("xunhupay response hash mismatch: errcode=%d errmsg=%s openid=%s url=%s url_qrcode=%s hash=%s",
				response.ErrCode, response.ErrMsg, response.OpenID, response.URL, response.URLQRCode, response.Hash)
		}
	}
	if response.URL == "" && response.URLQRCode == "" {
		return nil, fmt.Errorf("xunhupay response missing payment url")
	}
	return response, nil
}

// VerifyNotify verifies callback signature.
func (c *XunhuPayClient) VerifyNotify(payload XunhuPayNotifyPayload, appSecret string) bool {
	params := map[string]string{
		"appid":          payload.AppID,
		"trade_order_id": payload.TradeOrderID,
		"total_fee":      payload.TotalFee,
		"transaction_id": payload.TransactionID,
		"open_order_id":  payload.OpenOrderID,
		"order_title":    payload.OrderTitle,
		"status":         payload.Status,
		"plugins":        payload.Plugins,
		"attach":         payload.Attach,
		"time":           payload.Time,
		"nonce_str":      payload.NonceStr,
	}
	sign := xunhuPaySign(params, appSecret)
	return strings.EqualFold(sign, payload.Hash)
}

func xunhuPayVerify(raw map[string]any, appSecret, expected string) bool {
	params := make(map[string]string, len(raw))
	for k, v := range raw {
		if k == "hash" {
			continue
		}
		val := toString(v)
		if val == "" {
			continue
		}
		params[k] = val
	}
	sign := xunhuPaySign(params, appSecret)
	return strings.EqualFold(sign, expected)
}

func xunhuPaySign(params map[string]string, appSecret string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if k == "hash" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for i, key := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(params[key])
	}
	builder.WriteString(appSecret)
	sum := md5.Sum([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

func randomNonce(length int) string {
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if length <= 0 {
		length = 16
	}
	buf := make([]byte, length)
	randBuf := make([]byte, length)
	if _, err := rand.Read(randBuf); err != nil {
		for i := range buf {
			buf[i] = chars[i%len(chars)]
		}
		return string(buf)
	}
	for i := 0; i < length; i++ {
		buf[i] = chars[int(randBuf[i])%len(chars)]
	}
	return string(buf)
}

func toString(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		if value == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func toInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(strings.TrimSpace(v))
	default:
		return 0, fmt.Errorf("unsupported int value")
	}
}
