-- migrate:up
CREATE EXTENSION if NOT EXISTS pg_trgm;

CREATE INDEX characters_name_trgm_idx ON characters USING gin (name gin_trgm_ops);

CREATE INDEX characters_id_varchar_prefix_idx ON characters ((id::VARCHAR) varchar_pattern_ops);

-- migrate:down
DROP INDEX if EXISTS characters_id_varchar_prefix_idx;

DROP INDEX if EXISTS characters_name_trgm_idx;

DROP EXTENSION if EXISTS pg_trgm;
