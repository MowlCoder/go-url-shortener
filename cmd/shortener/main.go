package main

import (
	"compress/gzip"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	"github.com/MowlCoder/go-url-shortener/internal/handlers"
	"github.com/MowlCoder/go-url-shortener/internal/logger"
	middlewares2 "github.com/MowlCoder/go-url-shortener/internal/middlewares"
	"github.com/MowlCoder/go-url-shortener/internal/services"
	"github.com/MowlCoder/go-url-shortener/internal/storage"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

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

	urlStorage, err := storage.New(appConfig)
	if err != nil {
		panic(err)
	}

	stringGeneratorService := services.NewStringGenerator()

	shortenerHandler := handlers.NewShortenerHandler(appConfig, urlStorage, stringGeneratorService)

	mux := chi.NewRouter()

	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(middlewares2.NewCompressMiddleware(gzipWriter).Handler)
	mux.Use(func(handler http.Handler) http.Handler {
		return middlewares2.WithLogging(handler, customLogger)
	})

	mux.Post("/api/shorten/batch", shortenerHandler.ShortBatchURL)
	mux.Post("/api/shorten", shortenerHandler.ShortURLJSON)
	mux.Post("/", shortenerHandler.ShortURL)
	mux.Get("/ping", shortenerHandler.Ping)
	mux.Get("/{id}", shortenerHandler.RedirectToURLByID)

	fmt.Println("URL Shortener server is running on", appConfig.BaseHTTPAddr)

	if err := http.ListenAndServe(appConfig.BaseHTTPAddr, mux); err != nil {
		panic(err)
	}
}
