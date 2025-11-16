-- migrate:up
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX characters_name_trgm_idx ON characters USING GIN (name gin_trgm_ops);
CREATE INDEX characters_id_varchar_prefix_idx ON characters ((id::VARCHAR) varchar_pattern_ops);

-- migrate:down
DROP INDEX IF EXISTS characters_id_varchar_prefix_idx;
DROP INDEX IF EXISTS characters_name_trgm_idx;
DROP EXTENSION IF EXISTS pg_trgm;
