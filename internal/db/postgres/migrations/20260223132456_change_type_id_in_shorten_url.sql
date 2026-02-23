-- +goose Up
ALTER TABLE shorten_url 
    ALTER COLUMN id DROP DEFAULT;

DROP SEQUENCE IF EXISTS shorten_url_id_seq;

ALTER TABLE shorten_url 
    ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY 
    (START WITH (SELECT MAX(id) + 1 FROM shorten_url));

-- +goose Down
ALTER TABLE shorten_url 
    ALTER COLUMN id DROP IDENTITY IF EXISTS;

CREATE SEQUENCE shorten_url_id_seq
    START WITH (SELECT MAX(id) + 1 FROM shorten_url)
    OWNED BY shorten_url.id;

ALTER TABLE shorten_url 
    ALTER COLUMN id SET DEFAULT nextval('shorten_url_id_seq');
