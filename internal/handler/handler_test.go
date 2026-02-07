package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/klyakssa/go-musthave-shortener-tpl/internal/config"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/handler"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/logger"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/model"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/service/fileservice"
	"github.com/klyakssa/go-musthave-shortener-tpl/internal/service/uuidservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log *logger.MyLogger
var cfg *config.Config
var hand *handler.MyHandlerStruct

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
	log = logger.NewLogger()
	cfg = config.InitFlagConfig()
	fs := fileservice.New(cfg)
	uuid := uuidservice.New()

	hand := handler.NewMyHandler(cfg, log, fs, uuid)
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(model.ShortenRequest{URL: tt.want.url})
			require.NoError(t, err)

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
			w := httptest.NewRecorder()
			hand.NewShortenHandler(w, request)

			res := w.Result()
			bufBody, err := io.ReadAll(res.Body)
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
