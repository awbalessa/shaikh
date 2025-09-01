-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.memories (
    id integer NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    source_message text NOT NULL,
    confidence REAL NOT NULL,
    unique_key TEXT NOT NULL,
    memory text NOT NULL
);

ALTER TABLE ONLY public.memories
    ADD CONSTRAINT memories_pkey PRIMARY KEY (id);

ALTER TABLE public.memories ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.memories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

ALTER TABLE ONLY public.memories
    ADD CONSTRAINT user_id_unique_key_unique UNIQUE (user_id, unique_key);

ALTER TABLE ONLY public.memories
    ADD CONSTRAINT memories_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

CREATE INDEX idx_memories_user_id ON public.memories USING btree (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS public.idx_memories_user_id;

ALTER TABLE IF EXISTS public.memories DROP CONSTRAINT IF EXISTS memories_user_id_fkey;

ALTER TABLE IF EXISTS public.memories ALTER COLUMN id DROP IDENTITY;

DROP TABLE IF EXISTS public.memories;

DROP SEQUENCE IF EXISTS public.memories_id_seq;
-- +goose StatementEnd
