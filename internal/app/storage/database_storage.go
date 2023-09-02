package storage

import (
	"context"
	"database/sql"
	"math/rand"
	"time"

	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"
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

func (storage *DatabaseStorage) GetOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	row := storage.db.QueryRowContext(ctx, "SELECT original_url FROM shorten_url WHERE short_url = $1", shortURL)

	if row == nil || row.Err() != nil {
		return "", errorURLNotFound
	}

	var originalURL string
	row.Scan(&originalURL)

	return originalURL, nil
}

func (storage *DatabaseStorage) SaveURL(ctx context.Context, url string) (models.ShortenedURL, error) {
	shortURL := storage.generateUniqueShortSlug(ctx)

	row := storage.db.QueryRowContext(
		ctx,
		`
			INSERT INTO shorten_url (short_url, original_url, created_at) VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
			RETURNING id, short_url, original_url;
		`,
		shortURL, url, time.Now(),
	)

	if row.Err() != nil {
		return models.ShortenedURL{}, row.Err()
	}

	shortenedURL := models.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL); err != nil {
		return models.ShortenedURL{}, err
	}

	if shortURL != shortenedURL.ShortURL {
		return shortenedURL, ErrRowConflict
	}

	return shortenedURL, nil
}

func (storage *DatabaseStorage) SaveSeveralURL(ctx context.Context, urls []string) ([]models.ShortenedURL, error) {
	tx, err := storage.db.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	shortenedURLs := make([]models.ShortenedURL, 0, len(urls))
	isResultWithConflict := false

	for _, url := range urls {
		shortURL := storage.generateUniqueShortSlug(ctx)
		row := tx.QueryRowContext(
			ctx,
			`
				INSERT INTO shorten_url (short_url, original_url, created_at) VALUES ($1, $2, $3)
				ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url                                                              
			  	RETURNING id, short_url, original_url;
		`,
			shortURL, url, time.Now(),
		)

		if row.Err() != nil {
			return nil, row.Err()
		}

		shortenedURL := models.ShortenedURL{}

		if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL); err != nil {
			return nil, err
		}

		if shortURL != shortenedURL.ShortURL {
			isResultWithConflict = true
		}

		shortenedURLs = append(shortenedURLs, shortenedURL)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if isResultWithConflict {
		return shortenedURLs, ErrRowConflict
	}

	return shortenedURLs, nil
}

func (storage *DatabaseStorage) Ping(ctx context.Context) error {
	return storage.db.Ping()
}

func (storage *DatabaseStorage) generateUniqueShortSlug(ctx context.Context) string {
	shortURL := ""

	for original := "original"; original != ""; original, _ = storage.GetOriginalURLByShortURL(ctx, shortURL) {
		shortURL = util.Base62Encode(rand.Uint64())
	}

	return shortURL
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
	tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON shorten_url (original_url)`)

	return tx.Commit()
}
