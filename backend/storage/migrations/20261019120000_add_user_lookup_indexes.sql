-- migrate:up
CREATE INDEX if NOT EXISTS users_discord_username_idx ON users (discord_username)
WHERE
  discord_username <> '';

CREATE INDEX if NOT EXISTS users_anilist_url_lower_idx ON users (LOWER(anilist_url));

-- migrate:down
DROP INDEX if EXISTS users_anilist_url_lower_idx;

DROP INDEX if EXISTS users_discord_username_idx;
