package storage

import (
	"database/sql"
	"math/rand"
	"time"

	"github.com/MowlCoder/go-url-shortener/internal/app/util"
)

type DatabaseStorage struct {
	db *sql.DB
}

func NewDatabaseStorage(databaseDNS string) (*DatabaseStorage, error) {
	db, err := sql.Open("pgx", databaseDNS)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	dbStorage := DatabaseStorage{
		db: db,
	}

	if err := dbStorage.bootstrap(); err != nil {
		return nil, err
	}

	return &dbStorage, nil
}

func (storage *DatabaseStorage) GetOriginalURLByShortURL(shortURL string) (string, error) {
	row := storage.db.QueryRow("SELECT original_url FROM shorten_url WHERE short_url = $1", shortURL)

	if row == nil || row.Err() != nil {
		return "", errorURLNotFound
	}

	var originalURL string
	row.Scan(&originalURL)

	return originalURL, nil
}

func (storage *DatabaseStorage) SaveURL(url string) (string, error) {
	shortURL := util.Base62Encode(rand.Uint64())

	_, err := storage.db.Exec(
		"INSERT INTO shorten_url (short_url, original_url, created_at) VALUES ($1, $2, $3) RETURNING id, short_url, original_url;",
		shortURL, url, time.Now(),
	)

	return shortURL, err
}

func (storage *DatabaseStorage) Ping() error {
	return storage.db.Ping()
}

func (storage *DatabaseStorage) bootstrap() error {
	tx, err := storage.db.Begin()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	tx.Exec(`
		CREATE TABLE IF NOT EXISTS shorten_url (
	  		id serial PRIMARY KEY,
	  		short_url VARCHAR ( 20 ) UNIQUE NOT NULL,
	  		original_url TEXT NOT NULL,
	  		created_at TIMESTAMP NOT NULL
		)
	`)
	tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS short_url_idx ON shorten_url (short_url)`)

	return tx.Commit()
}
