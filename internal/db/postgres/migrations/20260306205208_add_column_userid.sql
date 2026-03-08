-- +goose Up
-- +goose StatementBegin
ALTER TABLE shorten_url ADD COLUMN user_id VARCHAR(36) NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shorten_url DROP COLUMN IF EXISTS user_id;
-- +goose StatementEnd
