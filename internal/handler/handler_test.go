package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/handler"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/model"
	"github.com/klyakssa/test-repo-url/internal/service/uuidservice"
	"github.com/klyakssa/test-repo-url/internal/uuidstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testConfig *config.Config

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	testConfig = config.InitFlagConfig()
}

func TestMainHandler(t *testing.T) {
	type want struct {
		codePost    int
		codeGet     int
		url         string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test #1",
			want: want{
				codePost:    201,
				codeGet:     307,
				url:         "https://yandex.ru",
				contentType: "text/plain",
			},
		},
		{
			name: "test #2",
			want: want{
				codePost:    201,
				codeGet:     307,
				url:         "https://practicum.yandex.ru",
				contentType: "text/plain",
			},
		},
	}
	log := logger.NewLogger()

	ctx := context.Background()
	userService := uuidservice.New(ctx, uuidstorage.New(testConfig), log)

	hand := handler.NewMyHandler(log, userService, testConfig)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.want.url))
			w := httptest.NewRecorder()
			hand.ShortenHandler(w, request)

			res := w.Result()
			assert.Equal(t, test.want.codePost, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			assert.Equal(t, strconv.Itoa(len(resBody)), res.Header.Get("Content-Length"))

			require.NoError(t, err)

			request2 := httptest.NewRequest(http.MethodGet, string(resBody), nil)
			w2 := httptest.NewRecorder()
			hand.UnshortenHandler(w2, request2)

			res2 := w2.Result()

			assert.Equal(t, test.want.codeGet, res2.StatusCode)
			assert.Equal(t, test.want.url, res2.Header.Get("Location"))

			defer res2.Body.Close()
		})
	}
}

func TestNewShortenHandler(t *testing.T) {
	type want struct {
		codePost    int
		codeGet     int
		url         string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "test #1",
			want: want{
				codePost:    201,
				codeGet:     307,
				url:         "https://pkg.go.dev/",
				contentType: "application/json",
			},
		},
		{
			name: "test #2",
			want: want{
				codePost:    201,
				codeGet:     307,
				url:         "https://github.com/gin-gonic",
				contentType: "application/json",
			},
		},
	}
	log := logger.NewLogger()

	ctx := context.Background()
	userService := uuidservice.New(ctx, uuidstorage.New(testConfig), log)

	hand := handler.NewMyHandler(log, userService, testConfig)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(model.ShortenRequest{URL: tt.want.url})
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(data))
			w := httptest.NewRecorder()

			hand.NewShortenHandler(w, request)

			res := w.Result()
			bufBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, strconv.Itoa(len(bufBody)), res.Header.Get("Content-Length"))

			var resBody model.ShortenResponse
			err = json.NewDecoder(bytes.NewReader(bufBody)).Decode(&resBody)
			require.NoError(t, err)

			assert.Equal(t, tt.want.codePost, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			defer res.Body.Close()

			request2 := httptest.NewRequest(http.MethodGet, resBody.Result, nil)
			w2 := httptest.NewRecorder()
			hand.UnshortenHandler(w2, request2)

			res2 := w2.Result()

			assert.Equal(t, tt.want.codeGet, res2.StatusCode)
			assert.Equal(t, tt.want.url, res2.Header.Get("Location"))

			defer res2.Body.Close()
		})
	}
}

func BenchmarkMainHandler(b *testing.B) {
	log := &logger.MyLogger{
		SugaredLogger: zap.NewNop().Sugar(), // NoOp логгер
	}

	ctx := context.Background()
	userService := uuidservice.New(ctx, uuidstorage.New(testConfig), log)

	hand := handler.NewMyHandler(log, userService, testConfig)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("1"))
		request.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		hand.ShortenHandler(w, request)

		if w.Code != http.StatusCreated {
			b.Fatalf("unexpected status: %d", w.Code)
		}
	}
}

type mockService struct {
	urls map[string]string
}

func newMockService() *mockService {
	return &mockService{
		urls: make(map[string]string),
	}
}

func (m *mockService) Shorten(ctx context.Context, input model.CreateShortURLInput) (string, error) {
	uuid := "short-" + uuid.New().String()[:8]
	m.urls[uuid] = input.OriginalURL
	return "http://localhost:8080/" + uuid, nil
}

func (m *mockService) Unshorten(ctx context.Context, input model.GetShortURLInput) (string, error) {
	uuid := strings.TrimPrefix(input.UUID, "http://localhost:8080/")
	if url, ok := m.urls[uuid]; ok {
		return url, nil
	}
	return "", fmt.Errorf("url not found")
}

func (m *mockService) GetUrlsByUserID(ctx context.Context, userID string) ([]model.UrlsResponse, error) {
	var result []model.UrlsResponse
	for short, original := range m.urls {
		result = append(result, model.UrlsResponse{
			ShortURL:    short,
			OriginalURL: original,
		})
	}
	return result, nil
}

func (m *mockService) DeleteUrlsByUserID(ctx context.Context, userID string, uuids []string) {
	for _, uuid := range uuids {
		delete(m.urls, uuid)
	}
}

func (m *mockService) PingContext(ctx context.Context) error {
	return nil
}

func (m *mockService) Close() error {
	return nil
}

// setupTestHandler создает handler для тестирования
func setupTestHandler() (*handler.MyHandlerStruct, *mockService) {
	log := &logger.MyLogger{
		SugaredLogger: zap.NewNop().Sugar(),
	}

	cfg := &config.Config{
		WebConfig: config.WebConfig{
			HostPort: "localhost:8080",
			BaseURL:  "http://localhost:8080",
			Secret:   "test-secret-key",
		},
	}

	mockSvc := newMockService()
	h := handler.NewMyHandler(log, mockSvc, cfg)

	return h, mockSvc
}

// ExampleMyHandlerStruct_ShortenHandler демонстрирует работу с эндпоинтом сокращения ссылки (text/plain)
func ExampleMyHandlerStruct_ShortenHandler() {
	h, _ := setupTestHandler()

	// Создание запроса с телом в формате text/plain
	body := strings.NewReader("https://example.com/very/long/url/that/needs/shortening")
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", "text/plain")

	w := httptest.NewRecorder()

	h.ShortenHandler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	fmt.Println("Short URL: (generated UUID)")

	// Output:
	// Status: 201
	// Content-Type: text/plain
	// Short URL: (generated UUID)
}

// ExampleMyHandlerStruct_UnshortenHandler демонстрирует работу с эндпоинтом восстановления оригинальной ссылки
func ExampleMyHandlerStruct_UnshortenHandler() {
	h, mockSvc := setupTestHandler()

	shortURL, _ := mockSvc.Shorten(context.Background(), model.CreateShortURLInput{
		OriginalURL: "https://example.com/original/url",
		UserID:      "test-user",
	})

	uuid := strings.TrimPrefix(shortURL, "http://localhost:8080/")

	req := httptest.NewRequest(http.MethodGet, "/"+uuid, nil)
	w := httptest.NewRecorder()

	h.UnshortenHandler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Location: %s\n", w.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://example.com/original/url
}

// ExampleMyHandlerStruct_NewShortenHandler демонстрирует работу с эндпоинтом сокращения ссылки (application/json)
func ExampleMyHandlerStruct_NewShortenHandler() {
	h, _ := setupTestHandler()

	reqBody := model.ShortenRequest{
		URL: "https://api.example.com/long/endpoint/with/many/segments",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h.NewShortenHandler(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	// Output:
	// Status: 201
	// Content-Type: application/json
}

// ExampleMyHandlerStruct_BatchHandler демонстрирует работу с эндпоинтом пакетного сокращения ссылок
func ExampleMyHandlerStruct_BatchHandler() {
	h, _ := setupTestHandler()

	reqBody := []model.BatchShortenRequest{
		{CorrelationID: "1", OUrl: "https://example.com/first"},
		{CorrelationID: "2", OUrl: "https://example.com/second"},
		{CorrelationID: "3", OUrl: "https://example.com/third"},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/batch", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), handler.UserIDKey, "test-user")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	h.BatchHandler(w, req)

	var resp []model.BatchShortenResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	fmt.Printf("Number of results: %d\n", len(resp))

	// Output:
	// Status: 201
	// Content-Type: application/json
	// Number of results: 3
}

// ExampleMyHandlerStruct_PingPostgresHandler демонстрирует работу с эндпоинтом проверки соединения с БД
func ExampleMyHandlerStruct_PingPostgresHandler() {
	h, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h.PingPostgresHandler(w, req)

	fmt.Printf("Status: %d\n", w.Code)

	// Output:
	// Status: 200
}

// ExampleMyHandlerStruct_GetUrlsHandler демонстрирует работу с эндпоинтом получения списка URL пользователя
func ExampleMyHandlerStruct_GetUrlsHandler() {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h, mockSvc := setupTestHandler()

	mockSvc.Shorten(context.Background(), model.CreateShortURLInput{
		OriginalURL: "https://example.com/url1",
		UserID:      "test-user",
	})
	mockSvc.Shorten(context.Background(), model.CreateShortURLInput{
		OriginalURL: "https://example.com/url2",
		UserID:      "test-user",
	})

	r.GET("/api/user/urls", h.SecretMiddleware(), h.GetUrlsHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)

	// Output:
	// Status: 200
}

// ExampleMyHandlerStruct_DeleteUrlsHandler демонстрирует работу с эндпоинтом удаления ссылок
func ExampleMyHandlerStruct_DeleteUrlsHandler() {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	h, mockSvc := setupTestHandler()

	short1, _ := mockSvc.Shorten(context.Background(), model.CreateShortURLInput{
		OriginalURL: "https://example.com/to-delete-1",
		UserID:      "test-user",
	})
	short2, _ := mockSvc.Shorten(context.Background(), model.CreateShortURLInput{
		OriginalURL: "https://example.com/to-delete-2",
		UserID:      "test-user",
	})

	uuid1 := strings.TrimPrefix(short1, "http://localhost:8080/")
	uuid2 := strings.TrimPrefix(short2, "http://localhost:8080/")

	r.DELETE("/api/user/urls", h.SecretMiddleware(), h.DeleteUrlsHandler())

	deleteBody := []string{uuid1, uuid2}
	jsonBody, _ := json.Marshal(deleteBody)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)

	// Output:
	// Status: 202
}
