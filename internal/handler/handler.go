package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/klyakssa/test-repo-url/internal/db/postgres"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/model"
	"github.com/klyakssa/test-repo-url/internal/repository"
	"github.com/klyakssa/test-repo-url/pkg/gzip"
	"github.com/klyakssa/test-repo-url/pkg/httperror"
)

type MyHandlerStruct struct {
	Logger  *logger.MyLogger
	service repository.UserService
}

func NewMyHandler(l *logger.MyLogger, uuid repository.UserService) *MyHandlerStruct {
	return &MyHandlerStruct{
		Logger:  l,
		service: uuid,
	}
}

func (h *MyHandlerStruct) GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gz := gzip.NewCompressWriter(c.Writer)
			c.Writer = gz
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

func (h *MyHandlerStruct) ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ew := httperror.NewErrorsWriter(c.Writer, c)
		c.Writer = ew
		c.Next()

		if rw, ok := c.Writer.(*httperror.ErrorsWriter); ok {
			if len(c.Errors) > 0 {
				err := c.Errors.Last().Err
				if errors.Is(err, postgres.ErrInsertUniqueViolation) {
					rw.WriteHeader(http.StatusConflict)
				} else {
					rw.WriteHeader(http.StatusInternalServerError)
				}
			}
			rw.MyFlush()
		}
	}
}

func (h *MyHandlerStruct) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("ShortenHandler")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shrt, err := h.service.Shorten(string(body), r.Context())
	if err != nil {
		if rw, ok := w.(*httperror.ErrorsWriter); ok {
			rw.AddError(err)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.Logger.Debug(string(body), " to ", shrt)

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len([]byte(shrt))))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shrt))
}

func (h *MyHandlerStruct) UnshortenHandler(w http.ResponseWriter, r *http.Request) {
	lng, err := h.service.Unshorten(r.URL.Path[1:], r.Context())
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

	if err = r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shrt, err := h.service.Shorten(req.URL, r.Context())
	if err != nil {
		if rw, ok := w.(*httperror.ErrorsWriter); ok {
			rw.AddError(err)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	h.Logger.Debug(req.URL, " to ", shrt)

	data, err := json.Marshal(model.ShortenResponse{
		Result: shrt,
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
	h.Logger.Debug("PingPostgresHandler")
	if err := h.service.PingContext(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *MyHandlerStruct) BatchHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("BatchHandler")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req []model.BatchShortenRequest
	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp []model.BatchShortenResponse
	for _, v := range req {
		shrt, err := h.service.Shorten(v.OUrl, r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp = append(resp, model.BatchShortenResponse{
			CorrelationID: v.CorrelationID,
			SOrl:          shrt,
		})
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
