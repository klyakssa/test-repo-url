-- +goose Up
-- +goose StatementBegin
ALTER TABLE shorten_url ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shorten_url DROP COLUMN is_deleted;
-- +goose StatementEnd
