package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

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
