package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/model"
	"github.com/klyakssa/test-repo-url/internal/service/fileservice"
	"github.com/klyakssa/test-repo-url/internal/service/pgxservice"
	"github.com/klyakssa/test-repo-url/internal/service/uuidservice"
	"github.com/klyakssa/test-repo-url/pkg/gzip"
)

type MyHandlerStruct struct {
	cfg    *config.Config
	Logger *logger.MyLogger
	FS     *fileservice.FileService
	UUID   *uuidservice.UUIDService
	DB     *pgxservice.NewPgxService
}

func NewMyHandler(cfg *config.Config, l *logger.MyLogger, fs *fileservice.FileService, uuid *uuidservice.UUIDService, db *pgxservice.NewPgxService) *MyHandlerStruct {
	return &MyHandlerStruct{
		cfg:    cfg,
		Logger: l,
		FS:     fs,
		UUID:   uuid,
		DB:     db,
	}
}

func (h *MyHandlerStruct) GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gz := gzip.NewCompressWriter(c.Writer)
			c.Writer = gz
			h.Logger.Debug(c.Writer.Header())
			defer gz.Close()
		}

		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := gzip.NewCompressReader(c.Request.Body)
			if err != nil {
				h.Logger.Error(err)
				c.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			c.Request.Body = cr
			defer cr.Close()
		}

		c.Next()
	}
}

func (h *MyHandlerStruct) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shrt, err := h.UUID.Shorten(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Logger.Debug(string(body), " to ", shrt)

	if err = r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(h.cfg.WebConfig.BaseUrl+"/"+shrt)))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.cfg.WebConfig.BaseUrl + "/" + shrt))
}

func (h *MyHandlerStruct) UnshortenHandler(w http.ResponseWriter, r *http.Request) {
	lng, err := h.UUID.Unshorten(r.URL.Path[1:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Logger.Debug(r.URL.Path[1:], " to ", lng)
	w.Header().Add("Location", lng)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *MyHandlerStruct) NewShortenHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug(r.Header)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req model.ShortenRequest
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shrt, err := h.UUID.Shorten(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Logger.Debug(req.URL, " to ", shrt)

	if err = r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(model.ShortenResponse{
		Result: h.cfg.WebConfig.BaseUrl + "/" + shrt,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (h *MyHandlerStruct) PingPostgresHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
