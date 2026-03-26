-- migrate:up
ALTER TABLE characters ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW();

UPDATE characters SET updated_at = NOW() WHERE updated_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_characters_updated_at_id ON characters (updated_at, id);

-- migrate:down
DROP INDEX IF EXISTS idx_characters_updated_at_id;
ALTER TABLE characters DROP COLUMN IF EXISTS updated_at;
