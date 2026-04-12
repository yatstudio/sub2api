ALTER TABLE IF EXISTS user_distributions
    ADD COLUMN IF NOT EXISTS total_commission_withdrawn DECIMAL(20,8) NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS distribution_withdrawal_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20,8) NOT NULL,
    account_type VARCHAR(32),
    account_ref TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    notes TEXT,
    review_note TEXT,
    reviewed_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (amount > 0),
    CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_user_status_created_at
    ON distribution_withdrawal_requests(user_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_distribution_withdrawals_status_created_at
    ON distribution_withdrawal_requests(status, created_at DESC);
