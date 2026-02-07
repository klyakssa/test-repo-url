package handler_test

import (
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/klyakssa/test-repo-url/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})

	httpc := resty.New().
		SetBaseURL("http://localhost:8080").
		SetRedirectPolicy(redirPolicy)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := httpc.R().
				SetHeader("Content-Type", "text/plain").
				SetBody([]byte(test.want.url)).
				Post("/")

			require.NoError(t, err)

			assert.Equal(t, test.want.codePost, res.StatusCode())
			assert.Equal(t, test.want.contentType, res.Header().Get("Content-Type"))
			assert.Equal(t, strconv.Itoa(len(res.Body())), res.Header().Get("Content-Length"))

			res2, err := httpc.R().
				Get(string(res.Body()))

			assert.Equal(t, test.want.codeGet, res2.StatusCode())
			assert.Equal(t, test.want.url, res2.Header().Get("Location"))

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

	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})

	httpc := resty.New().
		SetBaseURL("http://localhost:8080").
		SetRedirectPolicy(redirPolicy)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := httpc.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept-Encoding", "").
				SetBody(model.ShortenRequest{
					URL: tt.want.url,
				}).
				SetResult(model.ShortenResponse{}).
				Post("/api/shorten")

			require.NoError(t, err)

			assert.Equal(t, strconv.Itoa(len(res.Body())), res.Header().Get("Content-Length"))
			assert.Equal(t, tt.want.codePost, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			res2, err := httpc.R().
				Get(res.Result().(*model.ShortenResponse).Result)

			assert.Equal(t, tt.want.codeGet, res2.StatusCode())
			assert.Equal(t, tt.want.url, res2.Header().Get("Location"))
		})
	}
}

func TestGzipShortenHandler(t *testing.T) {
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

	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})

	httpc := resty.New().
		SetBaseURL("http://localhost:8080").
		SetRedirectPolicy(redirPolicy)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			res, err := httpc.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("Accept-Encoding", "gzip").
				SetBody(model.ShortenRequest{
					URL: tt.want.url,
				}).
				SetResult(model.ShortenResponse{}).
				Post("/api/shorten")

			require.NoError(t, err)

			assert.Equal(t, tt.want.codePost, res.StatusCode())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))

			res2, err := httpc.R().
				Get(res.Result().(*model.ShortenResponse).Result)

			assert.Equal(t, tt.want.codeGet, res2.StatusCode())
			assert.Equal(t, tt.want.url, res2.Header().Get("Location"))
		})
	}
}
