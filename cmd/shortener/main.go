package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
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
	customMiddlewares "github.com/MowlCoder/go-url-shortener/internal/middlewares"
	"github.com/MowlCoder/go-url-shortener/internal/services"
	"github.com/MowlCoder/go-url-shortener/internal/storage"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	rand.Seed(time.Now().UnixNano())

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env provided")
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
	userService := services.NewUserService()
	deleteURLQueue := services.NewDeleteURLQueue(urlStorage, customLogger, 3)

	shortenerHandler := handlers.NewShortenerHandler(
		appConfig,
		urlStorage,
		stringGeneratorService,
		deleteURLQueue,
	)

	router := makeRouter(
		shortenerHandler,
		userService,
		customLogger,
		gzipWriter,
	)

	go deleteURLQueue.Start(context.Background())

	displayBuildInfo()
	fmt.Println("URL Shortener server is running on", appConfig.BaseHTTPAddr)
	fmt.Println("Config:", appConfig)

	if appConfig.EnableHTTPS {
		if err := http.ListenAndServeTLS(appConfig.BaseHTTPAddr, appConfig.SSLPemPath, appConfig.SSLKeyPath, router); err != nil {
			panic(err)
		}
	} else {
		if err := http.ListenAndServe(appConfig.BaseHTTPAddr, router); err != nil {
			panic(err)
		}
	}
}

// @title URL shortener
// @version 1.0
// @description URL shortener helps to work with long urls, allow to save your long url and give you a small url, that point to your long url
// @BasePath /
func makeRouter(
	shortenerHandler *handlers.ShortenerHandler,
	userService *services.UserService,
	customLogger *logger.Logger,
	gzipWriter *gzip.Writer,
) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(customMiddlewares.NewCompressMiddleware(gzipWriter).Handler)
	mux.Use(func(handler http.Handler) http.Handler {
		return customMiddlewares.WithLogging(handler, customLogger)
	})
	mux.Use(func(handler http.Handler) http.Handler {
		return customMiddlewares.AuthMiddleware(handler, userService)
	})

	mux.Post("/api/shorten/batch", shortenerHandler.ShortBatchURL)
	mux.Post("/api/shorten", shortenerHandler.ShortURLJSON)
	mux.Post("/", shortenerHandler.ShortURL)
	mux.Delete("/api/user/urls", shortenerHandler.DeleteURLs)
	mux.Get("/api/user/urls", shortenerHandler.GetMyURLs)
	mux.Get("/ping", shortenerHandler.Ping)
	mux.Get("/{id}", shortenerHandler.RedirectToURLByID)

	return mux
}

func displayBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	fmt.Println("========================Build Info========================")
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
	fmt.Println("==========================================================")
}
