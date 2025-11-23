CREATE EXTENSION if NOT EXISTS pg_trgm
WITH  schema public;

CREATE TYPE public.indexing_status AS ENUM('pending', 'in_progress', 'completed');

CREATE TABLE public.characters (
  id BIGINT CONSTRAINT characters_new_id_not_null NOT NULL,
  name CHARACTER VARYING(128) CONSTRAINT characters_new_name_not_null NOT NULL,
  image CHARACTER VARYING(256) CONSTRAINT characters_new_image_not_null NOT NULL
);

CREATE TABLE public.characters_backup (
  user_id BIGINT CONSTRAINT characters_user_id_not_null NOT NULL,
  id BIGINT CONSTRAINT characters_id_not_null NOT NULL,
  image CHARACTER VARYING(256) DEFAULT ''::CHARACTER VARYING CONSTRAINT characters_image_not_null NOT NULL,
  name CHARACTER VARYING(128) DEFAULT ''::CHARACTER VARYING CONSTRAINT characters_name_not_null NOT NULL,
  date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() CONSTRAINT characters_date_not_null NOT NULL,
  type CHARACTER VARYING DEFAULT ''::CHARACTER VARYING CONSTRAINT characters_type_not_null NOT NULL
);

CREATE TABLE public.collection (
  user_id BIGINT NOT NULL,
  character_id BIGINT NOT NULL,
  source CHARACTER VARYING(50) DEFAULT 'ROLL'::CHARACTER VARYING NOT NULL,
  acquired_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

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

CREATE TABLE public.schema_migrations (version CHARACTER VARYING NOT NULL);

CREATE TABLE public.users (
  id INTEGER NOT NULL,
  user_id BIGINT NOT NULL,
  quote TEXT DEFAULT ''::TEXT NOT NULL,
  date TIMESTAMP WITHOUT TIME ZONE DEFAULT '1970-01-01 00:00:00'::TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  favorite BIGINT,
  tokens INTEGER DEFAULT 0 NOT NULL,
  anilist_url CHARACTER VARYING(255) DEFAULT ''::CHARACTER VARYING NOT NULL
);

CREATE TABLE public.character_wishlist (
  user_id BIGINT NOT NULL,
  character_id BIGINT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);

