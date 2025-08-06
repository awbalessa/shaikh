-- +goose Up

CREATE TYPE messages_role AS enum ('user', 'model', 'function');
CREATE TYPE messages_model AS enum ('gemini-2.5-flash', 'gemini-2.5-flash-lite');


CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    session_id UUID REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    role messages_role NOT NULL,
    content TEXT NOT NULL,
    model messages_model NOT NULL,
    turn INTEGER NOT NULL,
    token_count INTEGER,
    function_name TEXT,
    CONSTRAINT unique_session_id_turn_role_key UNIQUE(session_id, role, turn)
);

CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);

-- +goose Down
DROP INDEX IF EXISTS idx_messages_user_id;
DROP INDEX IF EXISTS idx_messages_session_id;
DROP TABLE IF EXISTS messages;
DROP TYPE IF EXISTS messages_role;
DROP TYPE IF EXISTS messages_model;
