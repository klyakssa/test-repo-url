package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/klyakssa/test-repo-url/internal/config"
	"github.com/klyakssa/test-repo-url/internal/db/postgres"
	"github.com/klyakssa/test-repo-url/internal/logger"
	"github.com/klyakssa/test-repo-url/internal/model"
	"github.com/klyakssa/test-repo-url/internal/repository"
	"github.com/klyakssa/test-repo-url/pkg/audit"
	"github.com/klyakssa/test-repo-url/pkg/gzip"
	"github.com/klyakssa/test-repo-url/pkg/httperror"
	"go.uber.org/zap"
)

type contextKey string

const UserIDKey contextKey = "user_id" // is a key for user id

// MyHandlerStruct is a struct for handler
type MyHandlerStruct struct {
	Logger     *logger.MyLogger       // is a logger
	service    repository.UserService // is a service
	cfg        *config.Config         // is a config
	subscriber *audit.Audit           // is variable for audit output
}

// NewMyHandler is a constructor
func NewMyHandler(l *logger.MyLogger, uuid repository.UserService, cfg *config.Config) *MyHandlerStruct {
	return &MyHandlerStruct{
		Logger:     l,
		service:    uuid,
		cfg:        cfg,
		subscriber: audit.NewAudit(cfg.Audit.AuditFile, cfg.Audit.AuditURL),
	}
}

// WithLogging is a middleware for logging
func (h *MyHandlerStruct) WithLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()
		h.Logger.Infow("request", "method", c.Request.Method, "path", c.Request.URL.Path, "duration", time.Since(start).Seconds())
		h.Logger.Infow("response", "status", c.Writer.Status(), "size", c.Writer.Size())
	}
}

// SecretMiddleware is a middleware for auth
func (h *MyHandlerStruct) SecretMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.Logger.Debug("SecretMiddleware")

		key := sha256.Sum256([]byte(h.cfg.WebConfig.Secret))

		aesblock, err := aes.NewCipher(key[:])
		if err != nil {
			h.Logger.Error(err)
			return
		}

		aesgcm, err := cipher.NewGCM(aesblock)
		if err != nil {
			h.Logger.Error(err)
			return
		}

		nonce := key[len(key)-aesgcm.NonceSize():]

		uuid := uuid.NewString()

		cookie, err := c.Cookie("shorten_user_id")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				c.SetCookieData(&http.Cookie{
					Name:     "shorten_user_id",
					Value:    hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(uuid), nil)),
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), UserIDKey, string(uuid)))
				c.Next()
				return
			}
		}

		cook, err := hex.DecodeString(cookie)
		if err != nil {
			h.Logger.Debug(err)
			c.SetCookieData(&http.Cookie{
				Name:     "shorten_user_id",
				Value:    hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(uuid), nil)),
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			c.Next()
			return
		}

		userID, err := aesgcm.Open(nil, nonce, cook, nil)
		if err != nil {
			h.Logger.Debug(err)
			c.SetCookieData(&http.Cookie{
				Name:     "shorten_user_id",
				Value:    hex.EncodeToString(aesgcm.Seal(nil, nonce, []byte(uuid), nil)),
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			c.Next()
			return
		}

		h.Logger.Debug("SecretMiddleware", zap.String("user_id", string(userID)))
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), UserIDKey, string(userID)))
		c.Next()
	}
}

// GzipMiddleware is a middleware for gzip
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

// ErrorMiddleware is a middleware for errors
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
					rw.WriteString(http.StatusText(http.StatusInternalServerError))
				}
			}
			rw.MyFlush()
		}
	}
}

// ShortenHandler is a handler for shorten link
// Body: original url text/plain
func (h *MyHandlerStruct) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("ShortenHandler",
		zap.Any("headers", r.Header))

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		userID = "unknown"
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err = r.Body.Close(); err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shrt, err := h.service.Shorten(r.Context(), model.CreateShortURLInput{OriginalURL: string(body), UserID: userID})
	if err != nil {
		if rw, ok := w.(*httperror.ErrorsWriter); ok {
			rw.AddError(err)
		} else {
			h.Logger.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	h.Logger.Debug("ShortenHandler",
		zap.String("from_body", string(body)),
		zap.String("to_short", shrt),
		zap.String("user_id", userID),
	)

	go h.subscriber.Subscribe(&model.AuditEntry{
		Action: "shorten",
		UserID: func() string {
			if ok {
				return userID
			}
			return ""
		}(),
		URL: string(body),
	})

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len([]byte(shrt))))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shrt))
}

// UnshortenHandler is a handler for unshorten link
// Param: uuid
func (h *MyHandlerStruct) UnshortenHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("UnshortenHandler",
		zap.Any("headers", r.Header))

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		userID = "unknown"
	}

	lng, err := h.service.Unshorten(r.Context(), model.GetShortURLInput{
		UserID: userID,
		UUID:   r.URL.Path[1:],
	})
	if err != nil {
		if errors.Is(err, postgres.ErrURLDeleted) {
			http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
			return
		}
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	h.Logger.Debug("UnshortenHandler",
		zap.String("from_url", r.URL.Path[1:]),
		zap.String("to_original", lng),
		zap.String("user_id", userID),
	)

	go h.subscriber.Subscribe(&model.AuditEntry{
		Action: "follow",
		UserID: func() string {
			if ok {
				return userID
			}
			return ""
		}(),
		URL: lng,
	})

	w.Header().Add("Location", lng)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// NewShortenHandler is a handler for shorten link
// Body: original url application/json
func (h *MyHandlerStruct) NewShortenHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("NewShortenHandler",
		zap.Any("headers", r.Header))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var req model.ShortenRequest
	if err = json.Unmarshal(body, &req); err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusBadRequest)
		return
	}

	if err = r.Body.Close(); err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		userID = "unknown"
	}

	shrt, err := h.service.Shorten(r.Context(), model.CreateShortURLInput{OriginalURL: req.URL, UserID: userID})
	if err != nil {
		if rw, ok := w.(*httperror.ErrorsWriter); ok {
			rw.AddError(err)
		} else {
			h.Logger.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	h.Logger.Debug("NewShortenHandler",
		zap.String("from_url", req.URL),
		zap.String("to_short", shrt),
		zap.String("user_id", userID),
	)

	data, err := json.Marshal(model.ShortenResponse{
		Result: shrt,
	})
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	go h.subscriber.Subscribe(&model.AuditEntry{
		Action: "shorten",
		UserID: func() string {
			if ok {
				return userID
			}
			return ""
		}(),
		URL: req.URL,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// PingPostgresHandler is a handler for ping postgres
func (h *MyHandlerStruct) PingPostgresHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("PingPostgresHandler",
		zap.Any("headers", r.Header))

	if err := h.service.PingContext(r.Context()); err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// BatchHandler is a handler for batch shorten
// Body: original urls application/json
func (h *MyHandlerStruct) BatchHandler(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("BatchHandler",
		zap.Any("headers", r.Header))

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var req []model.BatchShortenRequest
	if err = json.Unmarshal(body, &req); err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusBadRequest)
		return
	}

	if err = r.Body.Close(); err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var resp []model.BatchShortenResponse
	for _, v := range req {
		shrt, err := h.service.Shorten(r.Context(), model.CreateShortURLInput{OriginalURL: v.OUrl, UserID: userID})
		if err != nil {
			h.Logger.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		resp = append(resp, model.BatchShortenResponse{
			CorrelationID: v.CorrelationID,
			SOrl:          shrt,
		})
	}

	data, err := json.Marshal(resp)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

// GetUrlsHandler is a handler for get urls
// Response: application/json original urls by user
func (h *MyHandlerStruct) GetUrlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.Logger.Debug("GetUrlsHandler",
			zap.Any("headers", c.Request.Header))

		userID, ok := c.Request.Context().Value(UserIDKey).(string)
		if !ok {
			http.Error(c.Writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		urls, err := h.service.GetUrlsByUserID(c.Request.Context(), userID)
		if err != nil {
			h.Logger.Error(err)
			http.Error(c.Writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if len(urls) == 0 {
			http.Error(c.Writer, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		}

		c.JSON(http.StatusOK, urls)
	}
}

// DeleteUrlsHandler is a handler for delete urls by user
// Body: uuids text/plain
func (h *MyHandlerStruct) DeleteUrlsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		h.Logger.Debug("DeleteUrlsHandler",
			zap.Any("headers", c.Request.Header))
		var uuids []string
		if err := c.BindJSON(&uuids); err != nil {
			h.Logger.Error(err)
			http.Error(c.Writer, http.StatusText(http.StatusInternalServerError), http.StatusBadRequest)
			return
		}

		userID, ok := c.Request.Context().Value(UserIDKey).(string)
		if !ok {
			http.Error(c.Writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		h.service.DeleteUrlsByUserID(c.Request.Context(), userID, uuids)
		c.Writer.WriteHeader(http.StatusAccepted)
	}
}
