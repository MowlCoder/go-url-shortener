package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/MowlCoder/go-url-shortener/internal/app/domain"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage/models"
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

func (storage *DatabaseStorage) SaveURL(ctx context.Context, dto domain.SaveShortUrlDto) (*models.ShortenedURL, error) {
	row := storage.db.QueryRowContext(
		ctx,
		`
			INSERT INTO shorten_url (short_url, original_url, created_at) VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
			RETURNING id, short_url, original_url;
		`,
		dto.ShortURL, dto.OriginalURL, time.Now(),
	)

	if row.Err() != nil {
		return nil, row.Err()
	}

	shortenedURL := models.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL); err != nil {
		return nil, err
	}

	if dto.ShortURL != shortenedURL.ShortURL {
		return &shortenedURL, ErrRowConflict
	}

	return &shortenedURL, nil
}

func (storage *DatabaseStorage) SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortUrlDto) ([]models.ShortenedURL, error) {
	tx, err := storage.db.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	shortenedURLs := make([]models.ShortenedURL, 0, len(dtos))
	originalUrls := make([]string, 0, len(dtos))

	sqlStr := "INSERT INTO shorten_url (short_url, original_url, created_at) VALUES "
	vals := []interface{}{}

	for idx, dto := range dtos {
		sqlStr += fmt.Sprintf("($%d, $%d, $%d),", idx*3+1, idx*3+2, idx*3+3)
		vals = append(vals, dto.ShortURL, dto.OriginalURL, time.Now())

		originalUrls = append(originalUrls, dto.OriginalURL)
	}

	sqlStr = sqlStr[0 : len(sqlStr)-1]
	sqlStr += " ON CONFLICT (original_url) DO NOTHING"

	_, err = tx.ExecContext(ctx, sqlStr, vals...)

	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(
		ctx,
		"SELECT id, short_url, original_url FROM shorten_url WHERE original_url = ANY($1::text[])",
		"{"+strings.Join(originalUrls, ",")+"}",
	)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		shortenedURL := models.ShortenedURL{}

		if err := rows.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL); err != nil {
			return nil, err
		}

		shortenedURLs = append(shortenedURLs, shortenedURL)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return shortenedURLs, nil
}

func (storage *DatabaseStorage) Ping(ctx context.Context) error {
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
	tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON shorten_url (original_url)`)

	return tx.Commit()
}
