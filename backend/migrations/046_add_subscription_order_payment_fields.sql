ALTER TABLE subscription_orders ADD COLUMN IF NOT EXISTS payment_provider VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE subscription_orders ADD COLUMN IF NOT EXISTS payment_url TEXT;
ALTER TABLE subscription_orders ADD COLUMN IF NOT EXISTS payment_qrcode TEXT;
ALTER TABLE subscription_orders ADD COLUMN IF NOT EXISTS payment_open_order_id VARCHAR(64);
ALTER TABLE subscription_orders ADD COLUMN IF NOT EXISTS payment_transaction_id VARCHAR(64);
ALTER TABLE subscription_orders ADD COLUMN IF NOT EXISTS payment_plugin VARCHAR(32);

CREATE INDEX IF NOT EXISTS idx_subscription_orders_payment_transaction_id ON subscription_orders(payment_transaction_id);
