-- +goose Up
CREATE TABLE IF NOT EXISTS distribution_invite_attributions (
  id BIGSERIAL PRIMARY KEY,
  invitee_user_id BIGINT NOT NULL,
  inviter_user_id BIGINT NOT NULL,
  invite_code TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT 'direct',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT distribution_invite_attributions_invitee_unique UNIQUE (invitee_user_id)
);

CREATE INDEX IF NOT EXISTS idx_distribution_invite_attr_inviter
  ON distribution_invite_attributions (inviter_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_distribution_invite_attr_source
  ON distribution_invite_attributions (source, created_at DESC);

ALTER TABLE distribution_invite_attributions
  ADD COLUMN IF NOT EXISTS material TEXT NOT NULL DEFAULT '';

ALTER TABLE distribution_invite_attributions
  ADD COLUMN IF NOT EXISTS version TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_distribution_invite_attr_source_material_version
  ON distribution_invite_attributions (source, material, version, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_distribution_invite_attr_source_material_version;

ALTER TABLE distribution_invite_attributions
  DROP COLUMN IF EXISTS version;

ALTER TABLE distribution_invite_attributions
  DROP COLUMN IF EXISTS material;
