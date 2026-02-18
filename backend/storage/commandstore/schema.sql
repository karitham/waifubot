CREATE TABLE public.command_migrations (
  hash TEXT NOT NULL,
  deployed_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);
