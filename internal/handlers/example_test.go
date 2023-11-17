package handlers_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/MowlCoder/go-url-shortener/internal/config"
	"github.com/MowlCoder/go-url-shortener/internal/handlers"
	"github.com/MowlCoder/go-url-shortener/internal/logger"
	"github.com/MowlCoder/go-url-shortener/internal/services"
	"github.com/MowlCoder/go-url-shortener/internal/storage"
)

func Example() {
	// Initialize all dependencies for handler
	appConfig := &config.AppConfig{
		BaseHTTPAddr:     ":8080",
		BaseShortURLAddr: "http://localhost:8080",
		AppEnvironment:   config.AppDevEnv,
	}
	urlStorage, _ := storage.New(appConfig)
	strGeneratorService := services.NewStringGenerator()
	customLogger, _ := logger.NewLogger(logger.Options{
		Level:        logger.LogInfo,
		IsProduction: appConfig.AppEnvironment == config.AppProductionEnv,
	})
	queue := services.NewDeleteURLQueue(urlStorage, customLogger, 3)

	// Create handler
	handler := handlers.NewShortenerHandler(
		appConfig,
		urlStorage,
		strGeneratorService,
		queue,
	)

	// Short url
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://test.url"))
	w := httptest.NewRecorder()
	handler.ShortURL(w, r)
	response := w.Result()
	defer response.Body.Close()

	shortURL, _ := io.ReadAll(response.Body)
	fmt.Println("Short URL:", shortURL)

	// Send GET request to shortened url and check if location header is set to original url
	redirectRequest := httptest.NewRequest(http.MethodGet, string(shortURL), nil)
	w = httptest.NewRecorder()
	handler.RedirectToURLByID(w, redirectRequest)
	response = w.Result()
	defer response.Body.Close()

	fmt.Printf("Status code: %d\nLocation: %s\n", response.StatusCode, response.Header.Get("Location"))
}
