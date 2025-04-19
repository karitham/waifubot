-- migrate:up
ALTER TABLE users
ADD COLUMN anilist_url VARCHAR(255) NOT NULL DEFAULT '';
-- migrate:down
ALTER TABLE users DROP COLUMN anilist_url;