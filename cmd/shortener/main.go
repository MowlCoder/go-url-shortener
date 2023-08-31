package main

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MowlCoder/go-url-shortener/internal/app/handlers"
	"github.com/MowlCoder/go-url-shortener/internal/app/middlewares"
	"github.com/MowlCoder/go-url-shortener/internal/app/storage"

	"github.com/MowlCoder/go-url-shortener/internal/app/logger"

	"github.com/MowlCoder/go-url-shortener/internal/app/config"
)

func main() {
	appConfig := &config.AppConfig{}
	appConfig.ParseFlags()

	customLogger, err := logger.NewLogger(logger.Options{
		Level:        logger.LogInfo,
		IsProduction: appConfig.AppEnvironment == config.AppProductionEnv,
	})

	if err != nil {
		panic(err)
	}

	gzipWriter, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)

	if err != nil {
		panic(err)
	}

	urlStorage := storage.NewFileStorage(appConfig.FileStoragePath)
	db, err := sql.Open("pgx", appConfig.DatabaseDSN)

	if err != nil {
		panic(err)
	}

	shortenerHandler := handlers.NewShortenerHandler(appConfig, urlStorage)

	mux := chi.NewRouter()

	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(middlewares.NewCompressMiddleware(gzipWriter).Handler)
	mux.Use(func(handler http.Handler) http.Handler {
		return middlewares.WithLogging(handler, customLogger)
	})

	mux.Post("/api/shorten", shortenerHandler.ShortURLJSON)
	mux.Post("/", shortenerHandler.ShortURL)
	mux.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()

		if err != nil {
			handlers.SendStatusCode(w, http.StatusInternalServerError)
			return
		}

		handlers.SendStatusCode(w, http.StatusOK)
	})
	mux.Get("/{id}", shortenerHandler.RedirectToURLByID)

	fmt.Println("URL Shortener server is running on", appConfig.BaseHTTPAddr)

	if err := http.ListenAndServe(appConfig.BaseHTTPAddr, mux); err != nil {
		panic(err)
	}
}
