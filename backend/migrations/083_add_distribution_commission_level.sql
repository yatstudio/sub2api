ALTER TABLE IF EXISTS distribution_commissions
    ADD COLUMN IF NOT EXISTS commission_level SMALLINT NOT NULL DEFAULT 1;

ALTER TABLE IF EXISTS distribution_commissions
    DROP CONSTRAINT IF EXISTS distribution_commissions_commission_level_check;

ALTER TABLE IF EXISTS distribution_commissions
    ADD CONSTRAINT distribution_commissions_commission_level_check
        CHECK (commission_level IN (1, 2));

CREATE INDEX IF NOT EXISTS idx_distribution_commissions_invitee_level_created_at
    ON distribution_commissions(invitee_user_id, commission_level, created_at DESC);
