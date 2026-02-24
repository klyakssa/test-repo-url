-- +goose Up
-- +goose StatementBegin
CREATE TABLE shorten_url (
    id            SERIAL PRIMARY KEY,
    uuid          TEXT NOT NULL UNIQUE,
    short_url     TEXT NOT NULL UNIQUE,
    original_url  TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shorten_url CASCADE;
-- +goose StatementEnd
