-- +goose Up
-- +goose StatementBegin
ALTER TABLE shorten_url ADD UNIQUE (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shorten_url DROP UNIQUE (original_url);
-- +goose StatementEnd
