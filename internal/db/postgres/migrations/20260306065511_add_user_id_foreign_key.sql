-- +goose Up
-- +goose StatementBegin
ALTER TABLE shorten_url
ALTER COLUMN user_id SET NOT NULL,
ADD CONSTRAINT fk_user_id
FOREIGN KEY (user_id) REFERENCES users(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shorten_url
DROP CONSTRAINT IF EXISTS fk_user_id,
ALTER COLUMN user_id DROP NOT NULL;
-- +goose StatementEnd
