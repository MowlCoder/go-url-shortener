package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
		appConfig,
	)

	workersCtx, workersStopCtx := context.WithCancel(context.Background())
	go deleteURLQueue.Start(workersCtx)

	displayBuildInfo()
	log.Println("URL Shortener server is running on", appConfig.BaseHTTPAddr)
	log.Println("Config:", appConfig)

	server := http.Server{
		Addr:    appConfig.BaseHTTPAddr,
		Handler: router,
	}

	go func() {
		var err error

		if appConfig.EnableHTTPS {
			err = server.ListenAndServeTLS(appConfig.SSLPemPath, appConfig.SSLKeyPath)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGINT)
	<-sigs

	log.Println("start graceful shutdown...")

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer shutdownCtxCancel()

	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Fatal("graceful shutdown timed out... forcing exit")
		}
	}()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	workersStopCtx()

	log.Println("graceful shutdown server successfully")
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
	appConfig *config.AppConfig,
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

	mux.Group(func(privateRouter chi.Router) {
		privateRouter.Use(customMiddlewares.TrustedSubnetsMiddleware(appConfig.TrustedSubnet))
		privateRouter.Get("/api/internal/stats", shortenerHandler.GetStats)
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
