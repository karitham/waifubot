-- migrate:up
ALTER TABLE public.characters 
ADD COLUMN IF NOT EXISTS media_title TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS public.channel_interactions (
    channel_id BIGINT PRIMARY KEY,
    interaction_count BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS public.channel_drops (
    channel_id BIGINT PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES public.characters(id) ON DELETE CASCADE
);

-- migrate:down
DROP TABLE IF EXISTS public.channel_drops;
DROP TABLE IF EXISTS public.channel_interactions;
ALTER TABLE public.characters 
DROP COLUMN IF EXISTS media_title;
