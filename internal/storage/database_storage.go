package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/MowlCoder/go-url-shortener/internal/domain"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// DatabaseStorage is storage that store all information in database. DB is PostgreSQL.
type DatabaseStorage struct {
	pool *pgxpool.Pool
}

// NewDatabaseStorage create database storage and run migrations.
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

	if err := dbStorage.runMigrations(databaseDNS); err != nil {
		return nil, err
	}

	return &dbStorage, nil
}

// GetByShortURL return model where short url equal given short url.
func (storage *DatabaseStorage) GetByShortURL(ctx context.Context, shortURL string) (*domain.ShortenedURL, error) {
	query := `
		SELECT id, short_url, original_url, is_deleted
		FROM shorten_url
		WHERE short_url = $1
	`
	row := storage.pool.QueryRow(ctx, query, shortURL)

	if row == nil {
		return nil, domain.ErrURLNotFound
	}

	shortenedURL := domain.ShortenedURL{}

	if err := row.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.OriginalURL, &shortenedURL.IsDeleted); err != nil {
		return nil, err
	}

	return &shortenedURL, nil
}

// GetURLsByUserID return list of models where user id equal given user id.
func (storage *DatabaseStorage) GetURLsByUserID(ctx context.Context, userID string) ([]domain.ShortenedURL, error) {
	urls := make([]domain.ShortenedURL, 0)
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
		shortenedURL := domain.ShortenedURL{}

		if err := rows.Scan(&shortenedURL.ID, &shortenedURL.ShortURL, &shortenedURL.UserID, &shortenedURL.OriginalURL); err != nil {
			return nil, err
		}

		urls = append(urls, shortenedURL)
	}

	return urls, nil
}

// SaveURL save short url to the database.
func (storage *DatabaseStorage) SaveURL(ctx context.Context, dto domain.SaveShortURLDto) (*domain.ShortenedURL, error) {
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

	shortenedURL := domain.ShortenedURL{}

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

// SaveSeveralURL save several short url to the database.
func (storage *DatabaseStorage) SaveSeveralURL(ctx context.Context, dtos []domain.SaveShortURLDto) ([]domain.ShortenedURL, error) {
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

	if batchCloseErr := batchResult.Close(); batchCloseErr != nil {
		return nil, batchCloseErr
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

	shortenedURLs := make([]domain.ShortenedURL, 0, len(dtos))

	for rows.Next() {
		shortenedURL := domain.ShortenedURL{}

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

// DeleteByShortURLs delete short urls from the database.
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

// DoDeleteURLTasks execute delete tasks and save result to the database.
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

// GetInternalStats get internal stats for metrics.
func (storage *DatabaseStorage) GetInternalStats(ctx context.Context) (*domain.InternalStats, error) {
	query := `
		SELECT COUNT(DISTINCT user_id) AS users_count, COUNT(id) AS urls_count
		FROM shorten_url
	`

	stats := domain.InternalStats{}
	err := storage.pool.QueryRow(ctx, query).Scan(&stats.Users, &stats.URLs)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// Ping check if storage is available.
func (storage *DatabaseStorage) Ping(ctx context.Context) error {
	return storage.pool.Ping(ctx)
}

func (storage *DatabaseStorage) runMigrations(databaseDNS string) error {
	db, err := sql.Open("pgx", databaseDNS)
	if err != nil {
		return err
	}

	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
