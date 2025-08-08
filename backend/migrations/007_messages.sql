-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.messages (
    id integer NOT NULL,
    session_id uuid,
    user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    role public.messages_role NOT NULL,
    content text NOT NULL,
    turn integer NOT NULL,
    function_name text
);

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);

ALTER TABLE public.messages ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT unique_session_id_turn_role_key UNIQUE (session_id, role, turn);

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.sessions(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;

CREATE INDEX idx_messages_session_id ON public.messages USING btree (session_id);
CREATE INDEX idx_messages_user_id ON public.messages USING btree (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS public.idx_messages_user_id;
DROP INDEX IF EXISTS public.idx_messages_session_id;

ALTER TABLE IF EXISTS public.messages DROP CONSTRAINT IF EXISTS messages_user_id_fkey;
ALTER TABLE IF EXISTS public.messages DROP CONSTRAINT IF EXISTS messages_session_id_fkey;

ALTER TABLE IF EXISTS public.messages DROP CONSTRAINT IF EXISTS unique_session_id_turn_role_key;

ALTER TABLE IF EXISTS public.messages ALTER COLUMN id DROP IDENTITY;

DROP TABLE IF EXISTS public.messages;

DROP SEQUENCE IF EXISTS public.messages_id_seq;
-- +goose StatementEnd