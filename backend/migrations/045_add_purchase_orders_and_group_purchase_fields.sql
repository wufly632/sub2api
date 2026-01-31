-- Add purchase fields to groups
ALTER TABLE groups ADD COLUMN IF NOT EXISTS purchase_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS purchase_price DECIMAL(20, 8) DEFAULT NULL;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS purchase_display_order INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_groups_purchase_enabled ON groups(purchase_enabled);
CREATE INDEX IF NOT EXISTS idx_groups_purchase_display_order ON groups(purchase_display_order);

-- Create subscription_orders table
CREATE TABLE IF NOT EXISTS subscription_orders (
    id               BIGSERIAL PRIMARY KEY,
    order_no         VARCHAR(32) NOT NULL UNIQUE,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id         BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    subscription_id  BIGINT REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending',
    amount           DECIMAL(20, 8) NOT NULL DEFAULT 0,
    currency         VARCHAR(10) NOT NULL DEFAULT 'CNY',
    validity_days    INT NOT NULL DEFAULT 30,
    paid_at          TIMESTAMPTZ,
    canceled_at      TIMESTAMPTZ,
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscription_orders_user_id ON subscription_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_subscription_orders_group_id ON subscription_orders(group_id);
CREATE INDEX IF NOT EXISTS idx_subscription_orders_status ON subscription_orders(status);
CREATE INDEX IF NOT EXISTS idx_subscription_orders_created_at ON subscription_orders(created_at);
