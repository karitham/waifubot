CREATE TABLE public.characters (
  id BIGINT CONSTRAINT characters_new_id_not_null NOT NULL,
  name CHARACTER VARYING(128) CONSTRAINT characters_new_name_not_null NOT NULL,
  image CHARACTER VARYING(256) CONSTRAINT characters_new_image_not_null NOT NULL,
  media_title TEXT NOT NULL DEFAULT ''
);

CREATE TABLE public.channel_drops (
  channel_id BIGINT PRIMARY KEY,
  character_id BIGINT NOT NULL REFERENCES public.characters (id) ON DELETE CASCADE
);
