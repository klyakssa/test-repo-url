-- +goose Up
ALTER TABLE shorten_url 
    ALTER COLUMN uuid TYPE VARCHAR(200),
    ALTER COLUMN short_url TYPE VARCHAR(200),
    ALTER COLUMN original_url TYPE VARCHAR(2048);

-- +goose Down
ALTER TABLE shorten_url 
    ALTER COLUMN uuid TYPE TEXT,
    ALTER COLUMN short_url TYPE TEXT,
    ALTER COLUMN original_url TYPE TEXT;
