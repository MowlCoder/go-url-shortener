package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	grpcHandlers "github.com/MowlCoder/go-url-shortener/internal/handlers/grpc"
	httpHandlers "github.com/MowlCoder/go-url-shortener/internal/handlers/http"
	"github.com/MowlCoder/go-url-shortener/internal/interceptors"
	"github.com/MowlCoder/go-url-shortener/internal/logger"
	customMiddlewares "github.com/MowlCoder/go-url-shortener/internal/middlewares"
	"github.com/MowlCoder/go-url-shortener/internal/services"
	"github.com/MowlCoder/go-url-shortener/internal/storage"
	"github.com/MowlCoder/go-url-shortener/proto"
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
	shortenerService := services.NewShortenerService(
		urlStorage,
		stringGeneratorService,
		deleteURLQueue,
	)

	httpShortenerHandler := httpHandlers.NewShortenerHandler(
		appConfig,
		shortenerService,
	)
	grpcShortenerHandler := grpcHandlers.NewShortenerHandler(
		appConfig,
		shortenerService,
	)

	httpRouter := makeRouter(
		httpShortenerHandler,
		userService,
		customLogger,
		gzipWriter,
		appConfig,
	)
	grpcServer := makeGRPCServer(
		grpcShortenerHandler,
		userService,
	)

	workersCtx, workersStopCtx := context.WithCancel(context.Background())
	go deleteURLQueue.Start(workersCtx)

	displayBuildInfo()
	log.Println("URL Shortener server is running on", appConfig.BaseHTTPAddr)
	log.Println("Config:", appConfig)

	httpServer := http.Server{
		Addr:    appConfig.BaseHTTPAddr,
		Handler: httpRouter,
	}

	go func() {
		var err error

		if appConfig.EnableHTTPS {
			err = httpServer.ListenAndServeTLS(appConfig.SSLPemPath, appConfig.SSLKeyPath)
		} else {
			err = httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		listen, err := net.Listen("tcp", appConfig.BaseGRPCAddr)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("gRPC server started on", appConfig.BaseGRPCAddr)
		if err := grpcServer.Serve(listen); err != nil {
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

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	grpcServer.Stop()
	workersStopCtx()

	log.Println("graceful shutdown server successfully")
}

// @title URL shortener
// @version 1.0
// @description URL shortener helps to work with long urls, allow to save your long url and give you a small url, that point to your long url
// @BasePath /
func makeRouter(
	shortenerHandler *httpHandlers.ShortenerHandler,
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

func makeGRPCServer(
	shortenerHandler *grpcHandlers.ShortenerHandler,
	userService *services.UserService,
) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.CreateAuthInterceptor(userService)),
	)
	proto.RegisterShortenerServer(grpcServer, shortenerHandler)

	return grpcServer
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
