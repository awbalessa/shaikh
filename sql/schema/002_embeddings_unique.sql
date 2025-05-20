-- +goose Up
-- Remove old constraint
ALTER TABLE embeddings
DROP CONSTRAINT IF EXISTS embeddings_content_key;

-- Add new composite unique constraint
ALTER TABLE embeddings ADD CONSTRAINT unique_content_metadata UNIQUE (content, metadata);

-- Add a GIN index on metadata
CREATE INDEX idx_embeddings_metadata ON embeddings USING GIN (metadata);

-- +goose Down
-- Drop the GIN index
DROP INDEX IF EXISTS idx_embeddings_metadata;

-- Drop composite unique constraint
ALTER TABLE embeddings
DROP CONSTRAINT IF EXISTS unique_content_metadata;

-- Restore old unique constraint
ALTER TABLE embeddings ADD CONSTRAINT embeddings_content_key UNIQUE (content);
