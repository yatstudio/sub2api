-- +goose Up
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
