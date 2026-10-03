-- migrate:up
-- idx_collection_user_character from 20260503000000 duplicates the
-- collection_pkey column order from 20251123142500. collection_character_idx
-- from that migration is a prefix of collection_character_user_idx, which
-- 251123223946 creates last: dbmate sorts that filename after every 2026
-- timestamp, so the composite does not exist yet when this runs.
DROP INDEX IF EXISTS idx_collection_user_character;
DROP INDEX IF EXISTS collection_character_idx;

-- migrate:down
CREATE INDEX IF NOT EXISTS collection_character_idx ON collection (character_id);
CREATE INDEX IF NOT EXISTS idx_collection_user_character ON collection (user_id, character_id);
