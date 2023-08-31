CREATE TABLE IF NOT EXISTS shorten_url (
  id serial PRIMARY KEY,
  short_url VARCHAR ( 20 ) UNIQUE NOT NULL,
  original_url TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL
);