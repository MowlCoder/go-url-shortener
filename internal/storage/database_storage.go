package storage

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
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

func (storage *DatabaseStorage) GetURLsByUserID(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
	urls := make([]models.ShortenedURL, 0)

	rows, err := storage.db.QueryContext(ctx, "SELECT id, short_url, user_id, original_url FROM shorten_url WHERE user_id = $1", userID)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	for rows.Next() {
		shortenedURL := models.ShortenedURL{}

		if err := rows.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.UserID, &shortenedURL.OriginalURL); err != nil {
			return nil, err
		}

		urls = append(urls, shortenedURL)
	}

	return urls, nil
}

func (storage *DatabaseStorage) SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*models.ShortenedURL, error) {
	row := storage.db.QueryRowContext(
		ctx,
		`
			INSERT INTO shorten_url (short_url, original_url, user_id) VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
			RETURNING id, short_url, user_id, original_url;
		`,
		dto.ShortURL, dto.OriginalURL, dto.UserID,
	)

	if row.Err() != nil {
		var pgErr *pgconn.PgError

		if errors.As(row.Err(), &pgErr) && pgErr.Code == PgUniqueIndexErrorCode {
			return nil, ErrShortURLConflict
		}

		return nil, row.Err()
	}

	shortenedURL := models.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.UserID, &shortenedURL.OriginalURL); err != nil {
		return nil, err
	}

	if dto.ShortURL != shortenedURL.ShortURL {
		return &shortenedURL, ErrRowConflict
	}

	return &shortenedURL, nil
}

func (storage *DatabaseStorage) SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]models.ShortenedURL, error) {
	tx, err := storage.db.Begin()

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	originalUrls := make([]string, 0, len(dtos))

	sqlStmtBuffer := bytes.Buffer{}
	sqlStmtBuffer.WriteString("INSERT INTO shorten_url (short_url, original_url, user_id) VALUES ")

	vals := []interface{}{}

	for idx, dto := range dtos {
		sqlStmtBuffer.WriteString(fmt.Sprintf("($%d, $%d, $%d)", idx*3+1, idx*3+2, idx*3+3))

		if idx+1 != len(dtos) {
			sqlStmtBuffer.WriteString(",")
		}

		vals = append(vals, dto.ShortURL, dto.OriginalURL, dto.UserID)
		originalUrls = append(originalUrls, dto.OriginalURL)
	}

	sqlStmtBuffer.WriteString(" ON CONFLICT (original_url) DO NOTHING")

	_, err = tx.ExecContext(ctx, sqlStmtBuffer.String(), vals...)

	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(
		ctx,
		"SELECT id, short_url, user_id, original_url FROM shorten_url WHERE original_url = ANY($1::text[])",
		"{"+strings.Join(originalUrls, ",")+"}",
	)

	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	shortenedURLs := make([]models.ShortenedURL, 0, len(dtos))

	for rows.Next() {
		shortenedURL := models.ShortenedURL{}

		if err := rows.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.UserID, &shortenedURL.OriginalURL); err != nil {
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
	  		user_id VARCHAR( 100 ) NOT NULL,
	  		created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS short_url_idx ON shorten_url (short_url)`)
	tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON shorten_url (original_url)`)

	return tx.Commit()
}
