CREATE TABLE IF NOT EXISTS shorten_url (
  id serial PRIMARY KEY,
  short_url VARCHAR ( 20 ) UNIQUE NOT NULL,
  original_url TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS short_url_idx ON shorten_url (short_url)
CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON shorten_url (original_url)