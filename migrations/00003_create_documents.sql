-- +goose Up
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    mime TEXT NOT NULL,
    file BOOLEAN NOT NULL,
    public BOOLEAN NOT NULL,
    created TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT documents_name_not_empty_check
        CHECK (btrim(name) <> ''),
    CONSTRAINT documents_mime_not_empty_check
        CHECK (btrim(mime) <> '')
);

CREATE INDEX documents_owner_id_name_created_idx
    ON documents (owner_id, name, created);

-- +goose Down
DROP TABLE documents;
