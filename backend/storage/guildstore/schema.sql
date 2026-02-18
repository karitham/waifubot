CREATE TYPE public.indexing_status AS ENUM('pending', 'in_progress', 'completed');

CREATE TABLE public.guild_indexing_jobs (
  guild_id BIGINT NOT NULL,
  status public.indexing_status DEFAULT 'pending'::public.indexing_status NOT NULL,
  updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE public.guild_members (
  guild_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  indexed_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE public.collection (
  user_id BIGINT NOT NULL,
  character_id BIGINT NOT NULL,
  source CHARACTER VARYING(50) DEFAULT 'ROLL'::CHARACTER VARYING NOT NULL,
  acquired_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);
