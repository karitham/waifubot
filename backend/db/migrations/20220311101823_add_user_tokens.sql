-- migrate:up
ALTER TABLE users
ADD tokens INT NOT NULL DEFAULT 0;
-- migrate:down
ALTER TABLE users DROP COLUMN tokens;