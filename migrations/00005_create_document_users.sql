-- +goose Up
CREATE TABLE document_users (
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    PRIMARY KEY (document_id, user_id)
);

CREATE INDEX document_users_user_id_document_id_idx
    ON document_users (user_id, document_id);

-- +goose Down
DROP TABLE document_users;
