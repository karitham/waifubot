-- migrate:up
ALTER TABLE characters ADD COLUMN IF NOT EXISTS favorites INTEGER NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE characters DROP COLUMN IF EXISTS favorites;
