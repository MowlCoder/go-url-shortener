package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
	"github.com/MowlCoder/go-url-shortener/internal/storage/models"
)

type DatabaseStorage struct {
	pool *pgxpool.Pool
}

func NewDatabaseStorage(databaseDNS string) (*DatabaseStorage, error) {
	dbpool, err := pgxpool.New(context.Background(), databaseDNS)

	if err != nil {
		return nil, err
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		return nil, err
	}

	dbStorage := DatabaseStorage{
		pool: dbpool,
	}

	if err := dbStorage.bootstrap(); err != nil {
		return nil, err
	}

	return &dbStorage, nil
}

func (storage *DatabaseStorage) GetByShortURL(ctx context.Context, shortURL string) (*models.ShortenedURL, error) {
	row := storage.pool.QueryRow(ctx, "SELECT id, short_url, original_url, is_deleted FROM shorten_url WHERE short_url = $1", shortURL)

	if row == nil {
		return nil, errorURLNotFound
	}

	shortenedURL := models.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL, &shortenedURL.IsDeleted); err != nil {
		return nil, err
	}

	return &shortenedURL, nil
}

func (storage *DatabaseStorage) GetURLsByUserID(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
	urls := make([]models.ShortenedURL, 0)

	rows, err := storage.pool.Query(ctx, "SELECT id, short_url, user_id, original_url FROM shorten_url WHERE user_id = $1", userID)

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
	row := storage.pool.QueryRow(
		ctx,
		`
			INSERT INTO shorten_url (short_url, original_url, user_id) VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
			RETURNING id, short_url, user_id, original_url;
		`,
		dto.ShortURL, dto.OriginalURL, dto.UserID,
	)

	shortenedURL := models.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.UserID, &shortenedURL.OriginalURL); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == PgUniqueIndexErrorCode {
			return nil, ErrShortURLConflict
		}

		return nil, err
	}

	if dto.ShortURL != shortenedURL.ShortURL {
		return &shortenedURL, ErrRowConflict
	}

	return &shortenedURL, nil
}

func (storage *DatabaseStorage) SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]models.ShortenedURL, error) {
	tx, err := storage.pool.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	originalURLs := make([]string, 0, len(dtos))

	for _, dto := range dtos {
		batch.Queue(
			"INSERT INTO shorten_url (short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING",
			dto.ShortURL, dto.OriginalURL, dto.UserID,
		)
		originalURLs = append(originalURLs, dto.OriginalURL)
	}

	batchResult := tx.SendBatch(ctx, batch)

	if err := batchResult.Close(); err != nil {
		return nil, err
	}

	rows, err := tx.Query(
		ctx,
		"SELECT id, short_url, user_id, original_url FROM shorten_url WHERE original_url = ANY($1)",
		originalURLs,
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

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return shortenedURLs, nil
}

func (storage *DatabaseStorage) DeleteByShortURLs(ctx context.Context, shortURLs []string, userID string) error {
	_, err := storage.pool.Exec(
		ctx,
		"UPDATE shorten_url SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2)",
		userID, shortURLs,
	)
	return err
}

func (storage *DatabaseStorage) DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error {
	batch := &pgx.Batch{}

	for _, task := range tasks {
		batch.Queue(
			"UPDATE shorten_url SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2)",
			task.UserID, task.ShortURLs,
		)
	}

	batchResult := storage.pool.SendBatch(ctx, batch)
	return batchResult.Close()
}

func (storage *DatabaseStorage) Ping(ctx context.Context) error {
	return storage.pool.Ping(ctx)
}

func (storage *DatabaseStorage) bootstrap() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	defer cancel()

	tx, err := storage.pool.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS shorten_url (
	  		id serial PRIMARY KEY,
	  		short_url VARCHAR ( 20 ) UNIQUE NOT NULL,
	  		original_url TEXT NOT NULL,
	  		user_id VARCHAR( 100 ) NOT NULL,
	  		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		   	is_deleted BOOLEAN DEFAULT FALSE
		)
	`)
	tx.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS short_url_idx ON shorten_url (short_url)`)
	tx.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS original_url_idx ON shorten_url (original_url)`)

	return tx.Commit(ctx)
}
