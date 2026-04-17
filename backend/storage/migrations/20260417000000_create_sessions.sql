CREATE TABLE public.sessions (
    token CHARACTER VARYING(64) PRIMARY KEY NOT NULL,
    user_id BIGINT NOT NULL REFERENCES public.users(user_id),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL,
    expires_at TIMESTAMP WITHOUT TIME ZONE DEFAULT (NOW() + INTERVAL '7 days') NOT NULL
);

CREATE INDEX idx_sessions_user_id ON public.sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON public.sessions(expires_at);
