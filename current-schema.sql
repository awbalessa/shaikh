--
-- PostgreSQL database dump
--

\restrict ibiowusqdtDMM1v9X08mwnq3O8hMreclbFA6eMEfOgCbKYw5rBIoPPqpodGyQae

-- Dumped from database version 17.5 (Homebrew)
-- Dumped by pg_dump version 17.6 (Homebrew)

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
-- Name: paradedb; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA paradedb;


--
-- Name: rag; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA rag;


--
-- Name: pg_search; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_search WITH SCHEMA paradedb;


--
-- Name: vector; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;


--
-- Name: vectorscale; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS vectorscale WITH SCHEMA public;


--
-- Name: large_language_model; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.large_language_model AS ENUM (
    'gemini-2.5-flash',
    'gemini-2.5-flash-lite'
);


--
-- Name: messages_role; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.messages_role AS ENUM (
    'user',
    'model',
    'function'
);


--
-- Name: ayah; Type: TYPE; Schema: rag; Owner: -
--

CREATE TYPE rag.ayah AS ENUM (
    '1',
    '2',
    '3',
    '4',
    '5',
    '6',
    '7',
    '8',
    '9',
    '10',
    '11',
    '12',
    '13',
    '14',
    '15',
    '16',
    '17',
    '18',
    '19',
    '20',
    '21',
    '22',
    '23',
    '24',
    '25',
    '26',
    '27',
    '28',
    '29',
    '30',
    '31',
    '32',
    '33',
    '34',
    '35',
    '36',
    '37',
    '38',
    '39',
    '40',
    '41',
    '42',
    '43',
    '44',
    '45',
    '46',
    '47',
    '48',
    '49',
    '50',
    '51',
    '52',
    '53',
    '54',
    '55',
    '56',
    '57',
    '58',
    '59',
    '60',
    '61',
    '62',
    '63',
    '64',
    '65',
    '66',
    '67',
    '68',
    '69',
    '70',
    '71',
    '72',
    '73',
    '74',
    '75',
    '76',
    '77',
    '78',
    '79',
    '80',
    '81',
    '82',
    '83',
    '84',
    '85',
    '86',
    '87',
    '88',
    '89',
    '90',
    '91',
    '92',
    '93',
    '94',
    '95',
    '96',
    '97',
    '98',
    '99',
    '100',
    '101',
    '102',
    '103',
    '104',
    '105',
    '106',
    '107',
    '108',
    '109',
    '110',
    '111',
    '112',
    '113',
    '114',
    '115',
    '116',
    '117',
    '118',
    '119',
    '120',
    '121',
    '122',
    '123',
    '124',
    '125',
    '126',
    '127',
    '128',
    '129',
    '130',
    '131',
    '132',
    '133',
    '134',
    '135',
    '136',
    '137',
    '138',
    '139',
    '140',
    '141',
    '142',
    '143',
    '144',
    '145',
    '146',
    '147',
    '148',
    '149',
    '150',
    '151',
    '152',
    '153',
    '154',
    '155',
    '156',
    '157',
    '158',
    '159',
    '160',
    '161',
    '162',
    '163',
    '164',
    '165',
    '166',
    '167',
    '168',
    '169',
    '170',
    '171',
    '172',
    '173',
    '174',
    '175',
    '176',
    '177',
    '178',
    '179',
    '180',
    '181',
    '182',
    '183',
    '184',
    '185',
    '186',
    '187',
    '188',
    '189',
    '190',
    '191',
    '192',
    '193',
    '194',
    '195',
    '196',
    '197',
    '198',
    '199',
    '200',
    '201',
    '202',
    '203',
    '204',
    '205',
    '206',
    '207',
    '208',
    '209',
    '210',
    '211',
    '212',
    '213',
    '214',
    '215',
    '216',
    '217',
    '218',
    '219',
    '220',
    '221',
    '222',
    '223',
    '224',
    '225',
    '226',
    '227',
    '228',
    '229',
    '230',
    '231',
    '232',
    '233',
    '234',
    '235',
    '236',
    '237',
    '238',
    '239',
    '240',
    '241',
    '242',
    '243',
    '244',
    '245',
    '246',
    '247',
    '248',
    '249',
    '250',
    '251',
    '252',
    '253',
    '254',
    '255',
    '256',
    '257',
    '258',
    '259',
    '260',
    '261',
    '262',
    '263',
    '264',
    '265',
    '266',
    '267',
    '268',
    '269',
    '270',
    '271',
    '272',
    '273',
    '274',
    '275',
    '276',
    '277',
    '278',
    '279',
    '280',
    '281',
    '282',
    '283',
    '284',
    '285',
    '286'
);


--
-- Name: content_type; Type: TYPE; Schema: rag; Owner: -
--

CREATE TYPE rag.content_type AS ENUM (
    'tafsir'
);


--
-- Name: granularity; Type: TYPE; Schema: rag; Owner: -
--

CREATE TYPE rag.granularity AS ENUM (
    'phrase',
    'ayah',
    'surah',
    'quran'
);


--
-- Name: source; Type: TYPE; Schema: rag; Owner: -
--

CREATE TYPE rag.source AS ENUM (
    'Tafsir Ibn Kathir'
);


--
-- Name: surah; Type: TYPE; Schema: rag; Owner: -
--

CREATE TYPE rag.surah AS ENUM (
    '1',
    '2',
    '3',
    '4',
    '5',
    '6',
    '7',
    '8',
    '9',
    '10',
    '11',
    '12',
    '13',
    '14',
    '15',
    '16',
    '17',
    '18',
    '19',
    '20',
    '21',
    '22',
    '23',
    '24',
    '25',
    '26',
    '27',
    '28',
    '29',
    '30',
    '31',
    '32',
    '33',
    '34',
    '35',
    '36',
    '37',
    '38',
    '39',
    '40',
    '41',
    '42',
    '43',
    '44',
    '45',
    '46',
    '47',
    '48',
    '49',
    '50',
    '51',
    '52',
    '53',
    '54',
    '55',
    '56',
    '57',
    '58',
    '59',
    '60',
    '61',
    '62',
    '63',
    '64',
    '65',
    '66',
    '67',
    '68',
    '69',
    '70',
    '71',
    '72',
    '73',
    '74',
    '75',
    '76',
    '77',
    '78',
    '79',
    '80',
    '81',
    '82',
    '83',
    '84',
    '85',
    '86',
    '87',
    '88',
    '89',
    '90',
    '91',
    '92',
    '93',
    '94',
    '95',
    '96',
    '97',
    '98',
    '99',
    '100',
    '101',
    '102',
    '103',
    '104',
    '105',
    '106',
    '107',
    '108',
    '109',
    '110',
    '111',
    '112',
    '113',
    '114'
);


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.goose_db_version ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: memories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.memories (
    id integer NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    memory text NOT NULL,
    source_message text NOT NULL,
    confidence real NOT NULL,
    unique_key text NOT NULL
);


--
-- Name: memories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.memories ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.memories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    id integer NOT NULL,
    session_id uuid NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    role public.messages_role NOT NULL,
    model public.large_language_model,
    turn integer NOT NULL,
    total_input_tokens integer,
    total_output_tokens integer,
    content text,
    function_name text,
    function_call jsonb,
    function_response jsonb
);


--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.messages ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: refresh_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.refresh_tokens (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id uuid NOT NULL,
    user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    archived_at timestamp with time zone,
    summary text,
    max_turn integer DEFAULT 0 NOT NULL,
    max_turn_summarized integer DEFAULT 0 NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    total_messages integer DEFAULT 0 NOT NULL,
    total_messages_memorized integer DEFAULT 0 NOT NULL,
    password_hash text NOT NULL,
    is_admin boolean DEFAULT false NOT NULL
);


--
-- Name: ayat; Type: TABLE; Schema: rag; Owner: -
--

CREATE TABLE rag.ayat (
    surah rag.surah NOT NULL,
    ayah rag.ayah NOT NULL,
    ar text NOT NULL,
    ar_uthmani text NOT NULL,
    en text NOT NULL
);


--
-- Name: chunks; Type: TABLE; Schema: rag; Owner: -
--

CREATE TABLE rag.chunks (
    id bigint NOT NULL,
    sequence_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    granularity rag.granularity NOT NULL,
    content_type rag.content_type NOT NULL,
    source rag.source NOT NULL,
    raw_chunk text NOT NULL,
    tokenized_chunk text NOT NULL,
    chunk_title text NOT NULL,
    tokenized_chunk_title text NOT NULL,
    context_header text NOT NULL,
    embedded_chunk text NOT NULL,
    labels smallint[] NOT NULL,
    embedding public.vector(1024) NOT NULL,
    parent_id integer,
    surah rag.surah,
    ayah rag.ayah
);


--
-- Name: chunks_id_seq; Type: SEQUENCE; Schema: rag; Owner: -
--

ALTER TABLE rag.chunks ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME rag.chunks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: documents; Type: TABLE; Schema: rag; Owner: -
--

CREATE TABLE rag.documents (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    granularity rag.granularity NOT NULL,
    content_type rag.content_type NOT NULL,
    source rag.source NOT NULL,
    context_header text NOT NULL,
    document text NOT NULL,
    surah rag.surah,
    ayah rag.ayah
);


--
-- Name: documents_id_seq; Type: SEQUENCE; Schema: rag; Owner: -
--

ALTER TABLE rag.documents ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME rag.documents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: memories memories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memories
    ADD CONSTRAINT memories_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);


--
-- Name: refresh_tokens refresh_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: messages unique_session_id_turn_role_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT unique_session_id_turn_role_key UNIQUE (session_id, role, turn);


--
-- Name: memories user_id_unique_key_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memories
    ADD CONSTRAINT user_id_unique_key_unique UNIQUE (user_id, unique_key);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: ayat ayat_pkey; Type: CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.ayat
    ADD CONSTRAINT ayat_pkey PRIMARY KEY (surah, ayah);


--
-- Name: chunks chunks_pkey; Type: CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT chunks_pkey PRIMARY KEY (id);


--
-- Name: documents documents_context_header_key; Type: CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.documents
    ADD CONSTRAINT documents_context_header_key UNIQUE (context_header);


--
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- Name: chunks unique_context_header_sequence_id_key; Type: CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT unique_context_header_sequence_id_key UNIQUE (context_header, sequence_id);


--
-- Name: idx_memories_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_memories_user_id ON public.memories USING btree (user_id);


--
-- Name: idx_messages_function_call_gin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_function_call_gin ON public.messages USING gin (function_call jsonb_path_ops);


--
-- Name: idx_messages_function_response_gin; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_function_response_gin ON public.messages USING gin (function_response jsonb_path_ops);


--
-- Name: idx_messages_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_session_id ON public.messages USING btree (session_id);


--
-- Name: idx_messages_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_messages_user_id ON public.messages USING btree (user_id);


--
-- Name: idx_refresh_tokens_hash; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_refresh_tokens_hash ON public.refresh_tokens USING btree (token_hash);


--
-- Name: idx_refresh_tokens_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_refresh_tokens_user ON public.refresh_tokens USING btree (user_id);


--
-- Name: idx_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_user_id ON public.sessions USING btree (user_id);


--
-- Name: bm25_chunks_tokenized_chunk; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX bm25_chunks_tokenized_chunk ON rag.chunks USING bm25 (id, tokenized_chunk, tokenized_chunk_title, content_type, source, surah, ayah) WITH (key_field=id, text_fields='{
        "tokenized_chunk": {
            "tokenizer": {"type": "whitespace", "stemmer": "Arabic"}
        },
        "tokenized_chunk_title": {
            "tokenizer": {"type": "whitespace", "stemmer": "Arabic"}
        }
    }', numeric_fields='{
        "surah": {"fast": true},
        "ayah": {"fast": true},
        "content_type": {"fast": true},
        "source": {"fast": true}
    }');


--
-- Name: btree_chunks_content_type; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX btree_chunks_content_type ON rag.chunks USING btree (content_type);


--
-- Name: btree_chunks_source; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX btree_chunks_source ON rag.chunks USING btree (source);


--
-- Name: btree_chunks_surah_ayah; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX btree_chunks_surah_ayah ON rag.chunks USING btree (surah, ayah);


--
-- Name: diskann_chunks_embedding_labels; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX diskann_chunks_embedding_labels ON rag.chunks USING diskann (embedding, labels);


--
-- Name: idx_documents_content_type; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX idx_documents_content_type ON rag.documents USING btree (content_type);


--
-- Name: idx_documents_source; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX idx_documents_source ON rag.documents USING btree (source);


--
-- Name: idx_documents_surah_ayah; Type: INDEX; Schema: rag; Owner: -
--

CREATE INDEX idx_documents_surah_ayah ON rag.documents USING btree (surah, ayah);


--
-- Name: memories memories_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.memories
    ADD CONSTRAINT memories_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: messages messages_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.sessions(id) ON DELETE CASCADE;


--
-- Name: messages messages_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: refresh_tokens refresh_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: chunks chunks_parent_id_fkey; Type: FK CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT chunks_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES rag.documents(id);


--
-- Name: chunks chunks_surah_ayah_fkey; Type: FK CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.chunks
    ADD CONSTRAINT chunks_surah_ayah_fkey FOREIGN KEY (surah, ayah) REFERENCES rag.ayat(surah, ayah);


--
-- Name: documents documents_surah_ayah_fkey; Type: FK CONSTRAINT; Schema: rag; Owner: -
--

ALTER TABLE ONLY rag.documents
    ADD CONSTRAINT documents_surah_ayah_fkey FOREIGN KEY (surah, ayah) REFERENCES rag.ayat(surah, ayah);


--
-- PostgreSQL database dump complete
--

\unrestrict ibiowusqdtDMM1v9X08mwnq3O8hMreclbFA6eMEfOgCbKYw5rBIoPPqpodGyQae

