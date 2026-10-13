--
-- PostgreSQL database dump
--


-- Dumped from database version 17.9
-- Dumped by pg_dump version 17.9

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: indexing_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.indexing_status AS ENUM (
    'pending',
    'in_progress',
    'completed'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: channel_drops; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.channel_drops (
    channel_id bigint NOT NULL,
    character_id bigint NOT NULL
);


--
-- Name: channel_interactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.channel_interactions (
    channel_id bigint NOT NULL,
    interaction_count bigint DEFAULT 0 NOT NULL
);


--
-- Name: character_wishlist; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.character_wishlist (
    user_id bigint NOT NULL,
    character_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);


--
-- Name: characters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.characters (
    id bigint NOT NULL,
    name character varying(128) NOT NULL,
    image character varying(256) NOT NULL,
    media_title text DEFAULT ''::text NOT NULL,
    favorites integer DEFAULT 0 NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    is_active boolean DEFAULT true NOT NULL
);


--
-- Name: characters_backup; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.characters_backup (
    user_id bigint NOT NULL,
    id bigint NOT NULL,
    image character varying(256) DEFAULT ''::character varying NOT NULL,
    name character varying(128) DEFAULT ''::character varying NOT NULL,
    date timestamp without time zone DEFAULT now() NOT NULL,
    type character varying DEFAULT ''::character varying NOT NULL
);


--
-- Name: collection; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collection (
    user_id bigint NOT NULL,
    character_id bigint NOT NULL,
    source character varying(50) DEFAULT 'ROLL'::character varying NOT NULL,
    acquired_at timestamp without time zone DEFAULT now()
);


--
-- Name: command_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.command_migrations (
    hash text NOT NULL,
    deployed_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: guild_indexing_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_indexing_jobs (
    guild_id bigint NOT NULL,
    status public.indexing_status DEFAULT 'pending'::public.indexing_status NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: guild_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.guild_members (
    guild_id bigint NOT NULL,
    user_id bigint NOT NULL,
    indexed_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

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


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: channel_drops channel_drops_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_drops
    ADD CONSTRAINT channel_drops_pkey PRIMARY KEY (channel_id);


--
-- Name: channel_interactions channel_interactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_interactions
    ADD CONSTRAINT channel_interactions_pkey PRIMARY KEY (channel_id);


--
-- Name: character_wishlist character_wishlist_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_wishlist
    ADD CONSTRAINT character_wishlist_pkey PRIMARY KEY (user_id, character_id);


--
-- Name: characters characters_new_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters
    ADD CONSTRAINT characters_new_pkey PRIMARY KEY (id);


--
-- Name: characters_backup characters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.characters_backup
    ADD CONSTRAINT characters_pkey PRIMARY KEY (id, user_id);


--
-- Name: collection collection_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection
    ADD CONSTRAINT collection_pkey PRIMARY KEY (user_id, character_id);


--
-- Name: guild_indexing_jobs guild_indexing_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_indexing_jobs
    ADD CONSTRAINT guild_indexing_jobs_pkey PRIMARY KEY (guild_id);


--
-- Name: guild_members guild_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.guild_members
    ADD CONSTRAINT guild_members_pkey PRIMARY KEY (guild_id, user_id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);


--
-- Name: character_wishlist_character_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX character_wishlist_character_idx ON public.character_wishlist USING btree (character_id);


--
-- Name: character_wishlist_character_user_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX character_wishlist_character_user_idx ON public.character_wishlist USING btree (character_id, user_id);


--
-- Name: character_wishlist_user_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX character_wishlist_user_idx ON public.character_wishlist USING btree (user_id);


--
-- Name: characters_id_user_id_date_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX characters_id_user_id_date_idx ON public.characters_backup USING btree (id, user_id, date);


--
-- Name: characters_id_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX characters_id_user_id_idx ON public.characters_backup USING btree (id, user_id);


--
-- Name: characters_id_varchar_prefix_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX characters_id_varchar_prefix_idx ON public.characters USING btree (((id)::character varying) varchar_pattern_ops);


--
-- Name: characters_name_trgm_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX characters_name_trgm_idx ON public.characters USING gin (name public.gin_trgm_ops);


--
-- Name: characters_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX characters_user_id_idx ON public.characters_backup USING btree (user_id);


--
-- Name: collection_character_user_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX collection_character_user_idx ON public.collection USING btree (character_id, user_id);


--
-- Name: collection_user_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX collection_user_idx ON public.collection USING btree (user_id, acquired_at DESC);


--
-- Name: idx_characters_active_favorites; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_characters_active_favorites ON public.characters USING btree (is_active, favorites DESC) WHERE (is_active = true);


--
-- Name: idx_characters_updated_at_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_characters_updated_at_id ON public.characters USING btree (updated_at, id);


--
-- Name: users_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_user_id_idx ON public.users USING btree (user_id);


--
-- Name: channel_drops channel_drops_character_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.channel_drops
    ADD CONSTRAINT channel_drops_character_id_fkey FOREIGN KEY (character_id) REFERENCES public.characters(id) ON DELETE CASCADE;


--
-- Name: character_wishlist character_wishlist_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.character_wishlist
    ADD CONSTRAINT character_wishlist_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- Name: collection collection_character_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection
    ADD CONSTRAINT collection_character_id_fkey FOREIGN KEY (character_id) REFERENCES public.characters(id);


--
-- Name: collection collection_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection
    ADD CONSTRAINT collection_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id);


--
-- PostgreSQL database dump complete
--


