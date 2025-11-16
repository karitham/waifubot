CREATE TABLE IF NOT EXISTS users (
  "id" serial NOT NULL,
  "user_id" BIGINT NOT NULL PRIMARY KEY,
  "quote" TEXT NOT NULL DEFAULT '',
  "date" TIMESTAMP NOT NULL DEFAULT '1970-01-01 00:00:00-00',
  "favorite" BIGINT,
  "tokens" INT NOT NULL DEFAULT 0,
  "anilist_url" VARCHAR(255) NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS characters (
  "user_id" BIGINT NOT NULL,
  "id" BIGINT NOT NULL,
  "image" CHARACTER VARYING(256) NOT NULL DEFAULT '',
  "name" CHARACTER VARYING(128) NOT NULL DEFAULT '',
  "date" TIMESTAMP NOT NULL DEFAULT NOW(),
  "type" VARCHAR NOT NULL DEFAULT '',
  PRIMARY KEY ("id", "user_id"),
  CONSTRAINT "users_characters_fk" FOREIGN key (user_id) REFERENCES users (user_id)
);

-- constraint
ALTER TABLE users
ADD CONSTRAINT "characters_users_fk" FOREIGN key (favorite, user_id) REFERENCES characters (id, user_id);

-- index
CREATE INDEX characters_id_user_id_idx ON characters (id, user_id);

CREATE INDEX characters_user_id_idx ON characters (user_id);

CREATE INDEX users_user_id_idx ON users (user_id);

CREATE INDEX characters_id_user_id_date_idx ON characters (id, user_id, date);

-- CREATE INDEX ON characters (user_id) STORING (image, name, date, type);
CREATE TYPE indexing_status AS ENUM ('pending', 'in_progress', 'completed');

CREATE TABLE guild_members (
    guild_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    indexed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (guild_id, user_id)
);

CREATE TABLE guild_indexing_jobs (
    guild_id BIGINT PRIMARY KEY,
    status indexing_status NOT NULL DEFAULT 'pending',
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
