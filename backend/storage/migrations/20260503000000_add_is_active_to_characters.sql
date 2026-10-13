-- migrate:up
ALTER TABLE characters
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX if NOT EXISTS idx_characters_active_favorites ON characters (is_active, favorites DESC)
WHERE
  is_active = TRUE;

CREATE INDEX if NOT EXISTS idx_collection_user_character ON collection (user_id, character_id);

-- migrate:down
DROP INDEX if EXISTS idx_collection_user_character;

DROP INDEX if EXISTS idx_characters_active_favorites;

ALTER TABLE characters
DROP COLUMN IF EXISTS is_active;
