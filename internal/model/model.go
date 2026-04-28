package model

// ShortenRequest is a request for shortening
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is a response for shortening
type ShortenResponse struct {
	Result string `json:"result"`
}

// FileStorageData is a data for file
type FileStorageData struct {
	UUID   string `json:"uuid"`
	SUrl   string `json:"short_url"`
	OUrl   string `json:"original_url"`
	UserID string `json:"user_id"`
}

// BatchShortenResponse is a response for batch shorten
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	SOrl          string `json:"short_url"`
}

// BatchShortenRequest is a request for batch shorten
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OUrl          string `json:"original_url"`
}

// AcceptedContentTypes is a list of accepted content types
var AcceptedContentTypes = []string{
	"application/json",
	"text/html",
}

// ShortURLModel is a model for short url
type ShortURLModel struct {
	OriginalURL string `db:"original_url"`
	UserID      string `db:"user_id"`
}

// GetShortURLInput is a request for getting short url
type GetShortURLInput struct {
	UUID   string
	UserID string
}

// CreateShortURLInput is a request for creating short url
type CreateShortURLInput struct {
	OriginalURL string
	UserID      string
}

// UrlsResponse is a response for getting urls
type UrlsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// UrlsStorage is a storage for urls
type UrlsStorage struct {
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
}

// Storage is a storage
type Storage struct {
	UUID        string `db:"uuid"`
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
	UserID      string `db:"user_id"`
	IsDeleted   bool   `db:"is_deleted"`
}

// DeleteTask is a task for deleting
type DeleteTask struct {
	UserID string
	UUIDs  []string
}

// AuditEntry is a entry for audit
type AuditEntry struct {
	Timestamp int64  `json:"ts"`
	Action    string `json:"action"`
	UserID    string `json:"user_id,omitempty"`
	URL       string `json:"url"`
}
