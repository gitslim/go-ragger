-- +goose Up
-- +goose StatementBegin
CREATE TYPE document_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed'
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    file_data BYTEA NOT NULL,
    file_size BIGINT NOT NULL,
    file_hash BYTEA NOT NULL,
    status document_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_documents_file_name ON documents(file_name);
CREATE INDEX idx_documents_mime_type ON documents(mime_type);
CREATE INDEX idx_documents_file_hash ON documents(file_hash);
CREATE INDEX idx_documents_status ON documents(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS documents;
DROP TYPE IF EXISTS document_status;
DROP INDEX IF EXISTS idx_documents_user_id;
DROP INDEX IF EXISTS idx_documents_file_name;
DROP INDEX IF EXISTS idx_documents_mime_type;
DROP INDEX IF EXISTS idx_documents_file_hash;
DROP INDEX IF EXISTS idx_documents_status;
-- +goose StatementEnd
