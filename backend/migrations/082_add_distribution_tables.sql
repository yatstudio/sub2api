CREATE TABLE IF NOT EXISTS user_distributions (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    inviter_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    invite_code VARCHAR(32) NOT NULL UNIQUE,
    commission_rate DECIMAL(6,4) NOT NULL DEFAULT 0.1000,
    total_referrals BIGINT NOT NULL DEFAULT 0,
    total_commission_earned DECIMAL(20,8) NOT NULL DEFAULT 0,
    total_contribution DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (commission_rate >= 0 AND commission_rate <= 1)
);

CREATE INDEX IF NOT EXISTS idx_user_distributions_inviter_user_id
    ON user_distributions(inviter_user_id);

CREATE INDEX IF NOT EXISTS idx_user_distributions_invite_code
    ON user_distributions(invite_code);

CREATE TABLE IF NOT EXISTS distribution_commissions (
    id BIGSERIAL PRIMARY KEY,
    inviter_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topup_amount DECIMAL(20,8) NOT NULL,
    commission_rate DECIMAL(6,4) NOT NULL,
    commission_amount DECIMAL(20,8) NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (topup_amount >= 0),
    CHECK (commission_rate >= 0 AND commission_rate <= 1),
    CHECK (commission_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_distribution_commissions_inviter_created_at
    ON distribution_commissions(inviter_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_distribution_commissions_invitee_created_at
    ON distribution_commissions(invitee_user_id, created_at DESC);
