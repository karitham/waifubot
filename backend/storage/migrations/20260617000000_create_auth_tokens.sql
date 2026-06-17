-- migrate:up
CREATE TABLE IF NOT EXISTS public.auth_tokens (
    token TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES public.users(user_id) ON DELETE CASCADE,
    scopes TEXT NOT NULL DEFAULT 'all',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS auth_tokens_user_id_idx ON public.auth_tokens(user_id);

-- migrate:down
DROP INDEX IF EXISTS auth_tokens_user_id_idx;
DROP TABLE IF EXISTS public.auth_tokens;
