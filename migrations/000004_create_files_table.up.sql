CREATE TABLE IF NOT EXISTS files (
    id            SERIAL PRIMARY KEY,
    storage_key   TEXT NOT NULL UNIQUE,
    original_name VARCHAR(255) NOT NULL,
    content_type  VARCHAR(100) NOT NULL,
    size_bytes    BIGINT NOT NULL CHECK (size_bytes > 0),
    kind          VARCHAR(10) NOT NULL CHECK (kind IN ('IMAGE', 'FILE')),
    visibility    VARCHAR(10) NOT NULL CHECK (visibility IN ('PUBLIC', 'PRIVATE')),
    url VARCHAR(255) NOT NULL,
    uploaded_by   UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);