-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS shorten_url (
   id serial PRIMARY KEY,
   short_url VARCHAR ( 20 ) UNIQUE NOT NULL,
   original_url TEXT NOT NULL,
   user_id VARCHAR( 100 ) NOT NULL,
   created_at TIMESTAMP NOT NULL DEFAULT NOW(),
   is_deleted BOOLEAN DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS short_url_idx ON shorten_url (short_url);
CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON shorten_url (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS short_url_idx;
DROP INDEX IF EXISTS original_url_idx;

DROP TABLE IF EXISTS shorten_url
-- +goose StatementEnd
