-- +goose Up
ALTER TABLE shorten_url 
    ALTER COLUMN id DROP DEFAULT;

DROP SEQUENCE IF EXISTS shorten_url_id_seq;

ALTER TABLE shorten_url 
    ALTER COLUMN id TYPE BIGINT,
    ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY;

SELECT setval(pg_get_serial_sequence('shorten_url', 'id'), COALESCE(MAX(id), 1)) FROM shorten_url;

-- +goose Down
ALTER TABLE shorten_url 
    ALTER COLUMN id DROP IDENTITY IF EXISTS;

CREATE SEQUENCE shorten_url_id_seq
    OWNED BY shorten_url.id;

SELECT setval('shorten_url_id_seq', COALESCE(MAX(id), 1)) FROM shorten_url;

ALTER TABLE shorten_url 
    ALTER COLUMN id SET DEFAULT nextval('shorten_url_id_seq'::regclass);
