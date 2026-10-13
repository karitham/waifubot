CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;
CREATE TYPE public.indexing_status AS ENUM (
    'pending',
    'in_progress',
    'completed'
);
CREATE TABLE public.channel_drops (
    channel_id bigint NOT NULL,
    character_id bigint NOT NULL
);
CREATE TABLE public.channel_interactions (
    channel_id bigint NOT NULL,
    interaction_count bigint DEFAULT 0 NOT NULL
);
CREATE TABLE public.character_wishlist (
    user_id bigint NOT NULL,
    character_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);
CREATE TABLE public.characters (
    id bigint CONSTRAINT characters_new_id_not_null NOT NULL,
    name character varying(128) CONSTRAINT characters_new_name_not_null NOT NULL,
    image character varying(256) CONSTRAINT characters_new_image_not_null NOT NULL,
    media_title text DEFAULT ''::text NOT NULL,
    favorites integer DEFAULT 0 NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);
CREATE TABLE public.characters_backup (
    user_id bigint CONSTRAINT characters_user_id_not_null NOT NULL,
    id bigint CONSTRAINT characters_id_not_null NOT NULL,
    image character varying(256) DEFAULT ''::character varying CONSTRAINT characters_image_not_null NOT NULL,
    name character varying(128) DEFAULT ''::character varying CONSTRAINT characters_name_not_null NOT NULL,
    date timestamp without time zone DEFAULT now() CONSTRAINT characters_date_not_null NOT NULL,
    type character varying DEFAULT ''::character varying CONSTRAINT characters_type_not_null NOT NULL
);
CREATE TABLE public.collection (
    user_id bigint NOT NULL,
    character_id bigint NOT NULL,
    source character varying(50) DEFAULT 'ROLL'::character varying NOT NULL,
    acquired_at timestamp without time zone DEFAULT now()
);
CREATE TABLE public.command_migrations (
    hash text NOT NULL,
    deployed_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE TABLE public.guild_indexing_jobs (
    guild_id bigint NOT NULL,
    status public.indexing_status DEFAULT 'pending'::public.indexing_status NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE TABLE public.guild_members (
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    indexed_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE TABLE public.users (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    quote text DEFAULT ''::text NOT NULL,
    date timestamp without time zone DEFAULT '1970-01-01 00:00:00'::timestamp without time zone NOT NULL,
    favorite bigint,
    tokens integer DEFAULT 0 NOT NULL,
    anilist_url character varying(255) DEFAULT ''::character varying NOT NULL,
    discord_username character varying(32) DEFAULT ''::character varying NOT NULL,
    discord_avatar character varying(34) DEFAULT ''::character varying NOT NULL,
    last_updated timestamp without time zone DEFAULT '1970-01-01 00:00:00'::timestamp without time zone NOT NULL
);
CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;
ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
ALTER TABLE ONLY public.channel_drops
    ADD CONSTRAINT channel_drops_pkey PRIMARY KEY (channel_id);
ALTER TABLE ONLY public.channel_interactions
    ADD CONSTRAINT channel_interactions_pkey PRIMARY KEY (channel_id);
ALTER TABLE ONLY public.character_wishlist
    ADD CONSTRAINT character_wishlist_pkey PRIMARY KEY (user_id, character_id);
ALTER TABLE ONLY public.characters
    ADD CONSTRAINT characters_new_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.characters_backup
    ADD CONSTRAINT characters_pkey PRIMARY KEY (id, user_id);
ALTER TABLE ONLY public.collection
    ADD CONSTRAINT collection_pkey PRIMARY KEY (user_id, character_id);
ALTER TABLE ONLY public.guild_indexing_jobs
    ADD CONSTRAINT guild_indexing_jobs_pkey PRIMARY KEY (guild_id);
ALTER TABLE ONLY public.guild_members
    ADD CONSTRAINT guild_members_pkey PRIMARY KEY (guild_id, user_id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);
CREATE INDEX character_wishlist_character_idx ON public.character_wishlist USING btree (character_id);
CREATE INDEX character_wishlist_character_user_idx ON public.character_wishlist USING btree (character_id, user_id);
CREATE INDEX character_wishlist_user_idx ON public.character_wishlist USING btree (user_id);
CREATE INDEX characters_id_user_id_date_idx ON public.characters_backup USING btree (id, user_id, date);
CREATE INDEX characters_id_user_id_idx ON public.characters_backup USING btree (id, user_id);
CREATE INDEX characters_id_varchar_prefix_idx ON public.characters USING btree (((id)::character varying) varchar_pattern_ops);
CREATE INDEX characters_name_trgm_idx ON public.characters USING gin (name public.gin_trgm_ops);
CREATE INDEX characters_user_id_idx ON public.characters_backup USING btree (user_id);
CREATE INDEX collection_character_user_idx ON public.collection USING btree (character_id, user_id);
CREATE INDEX collection_user_idx ON public.collection USING btree (user_id, acquired_at DESC);
CREATE INDEX idx_characters_active_favorites ON public.characters USING btree (is_active, favorites DESC) WHERE (is_active = true);
CREATE INDEX idx_characters_updated_at_id ON public.characters USING btree (updated_at, id);
CREATE INDEX users_user_id_idx ON public.users USING btree (user_id);
ALTER TABLE ONLY public.channel_drops
    ADD CONSTRAINT channel_drops_character_id_fkey FOREIGN KEY (character_id) REFERENCES public.characters(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.character_wishlist
    ADD CONSTRAINT character_wishlist_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);
ALTER TABLE ONLY public.collection
    ADD CONSTRAINT collection_character_id_fkey FOREIGN KEY (character_id) REFERENCES public.characters(id);
ALTER TABLE ONLY public.collection
    ADD CONSTRAINT collection_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);
