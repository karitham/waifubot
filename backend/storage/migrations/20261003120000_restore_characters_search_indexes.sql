-- migrate:up
-- The pg_trgm indexes from 20251015184814 are attached to characters_backup,
-- left behind when 20251123142500/01 rebuilt characters. Index names are
-- schema-scoped, so the leading drops free those names where that table still
-- exists. No query uses full-text operators, so characters_name_idx is unused.
DROP INDEX IF EXISTS characters_name_trgm_idx;
DROP INDEX IF EXISTS characters_id_varchar_prefix_idx;

CREATE INDEX characters_name_trgm_idx ON characters USING GIN (name gin_trgm_ops);
CREATE INDEX characters_id_varchar_prefix_idx ON characters ((id::VARCHAR) varchar_pattern_ops);

DROP INDEX IF EXISTS characters_name_idx;

-- migrate:down
-- Leaves characters_backup without its indexes; no query reads that table.
DROP INDEX IF EXISTS characters_id_varchar_prefix_idx;
DROP INDEX IF EXISTS characters_name_trgm_idx;

CREATE INDEX IF NOT EXISTS characters_name_idx ON characters USING GIN (to_tsvector('english', name));
