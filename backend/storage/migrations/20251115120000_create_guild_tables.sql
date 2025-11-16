-- migrate:up
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

-- migrate:down
DROP TABLE IF EXISTS guild_indexing_jobs;
DROP TABLE IF EXISTS guild_members;
DROP TYPE IF EXISTS indexing_status;
