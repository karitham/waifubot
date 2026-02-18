-- migrate:up
CREATE TABLE public.command_migrations (
  hash TEXT NOT NULL,
  deployed_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);

-- migrate:down
DROP TABLE IF EXISTS public.command_migrations;
