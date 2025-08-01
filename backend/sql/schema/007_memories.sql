-- +goose Up
CREATE TABLE IF NOT EXISTS memories (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    memory TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_memories_user_id ON memories(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_memories_user_id;
DROP TABLE IF EXISTS memories;
