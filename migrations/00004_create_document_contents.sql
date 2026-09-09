-- +goose Up
CREATE TABLE document_contents (
    document_id UUID PRIMARY KEY REFERENCES documents(id) ON DELETE CASCADE,
    json JSONB,
    file_path TEXT,

    CONSTRAINT document_contents_not_empty_check
        CHECK (json IS NOT NULL OR file_path IS NOT NULL),
    CONSTRAINT document_contents_file_path_not_empty_check
        CHECK (file_path IS NULL OR btrim(file_path) <> '')
);

-- +goose Down
DROP TABLE document_contents;
