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
	query := `
		SELECT id, short_url, original_url, is_deleted
		FROM shorten_url
		WHERE short_url = $1
	`
	row := storage.pool.QueryRow(ctx, query, shortURL)

	if row == nil {
		return nil, domain.ErrURLNotFound
	}

	shortenedURL := models.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL, &shortenedURL.IsDeleted); err != nil {
		return nil, err
	}

	return &shortenedURL, nil
}

func (storage *DatabaseStorage) GetURLsByUserID(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
	urls := make([]models.ShortenedURL, 0)
	query := `
		SELECT id, short_url, user_id, original_url
		FROM shorten_url
		WHERE user_id = $1
	`

	rows, err := storage.pool.Query(ctx, query, userID)

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
	query := `
		INSERT INTO shorten_url (short_url, original_url, user_id) VALUES ($1, $2, $3)
		ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING id, short_url, user_id, original_url;
	`
	row := storage.pool.QueryRow(
		ctx,
		query,
		dto.ShortURL, dto.OriginalURL, dto.UserID,
	)

	shortenedURL := models.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.UserID, &shortenedURL.OriginalURL); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == PgUniqueIndexErrorCode {
			return nil, domain.ErrShortURLConflict
		}

		return nil, err
	}

	if dto.ShortURL != shortenedURL.ShortURL {
		return &shortenedURL, domain.ErrURLConflict
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
	query := `
		INSERT INTO shorten_url (short_url, original_url, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (original_url) DO NOTHING
	`

	for _, dto := range dtos {
		batch.Queue(
			query,
			dto.ShortURL, dto.OriginalURL, dto.UserID,
		)
		originalURLs = append(originalURLs, dto.OriginalURL)
	}

	batchResult := tx.SendBatch(ctx, batch)

	if err := batchResult.Close(); err != nil {
		return nil, err
	}

	query = `
		SELECT id, short_url, user_id, original_url
		FROM shorten_url
		WHERE original_url = ANY($1)
	`
	rows, err := tx.Query(
		ctx,
		query,
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
	query := `
		UPDATE shorten_url
		SET is_deleted = TRUE
	   	WHERE user_id = $1 AND short_url = ANY($2)
	`
	_, err := storage.pool.Exec(
		ctx,
		query,
		userID, shortURLs,
	)
	return err
}

func (storage *DatabaseStorage) DoDeleteURLTasks(ctx context.Context, tasks []domain.DeleteURLsTask) error {
	batch := &pgx.Batch{}

	query := `
		UPDATE shorten_url
		SET is_deleted = TRUE
		WHERE user_id = $1 AND short_url = ANY($2)
	`

	for _, task := range tasks {
		batch.Queue(
			query,
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
