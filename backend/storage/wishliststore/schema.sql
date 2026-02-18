CREATE TABLE public.characters (
  id BIGINT CONSTRAINT characters_new_id_not_null NOT NULL,
  name CHARACTER VARYING(128) CONSTRAINT characters_new_name_not_null NOT NULL,
  image CHARACTER VARYING(256) CONSTRAINT characters_new_image_not_null NOT NULL
);

CREATE TABLE public.character_wishlist (
  user_id BIGINT NOT NULL,
  character_id BIGINT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE public.collection (
  user_id BIGINT NOT NULL,
  character_id BIGINT NOT NULL,
  source CHARACTER VARYING(50) DEFAULT 'ROLL'::CHARACTER VARYING NOT NULL,
  acquired_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE TABLE public.guild_members (
  guild_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  indexed_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL
);
