-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.users (
    id uuid NOT NULL,
    email text NOT NULL,
    password_hash TEXT NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    total_messages integer NOT NULL DEFAULT 0,
    total_messages_memorized integer NOT NULL DEFAULT 0
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.users;
-- +goose StatementEnd
