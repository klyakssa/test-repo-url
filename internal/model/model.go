package model

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type FileStorageData struct {
	UUID   string `json:"uuid"`
	SUrl   string `json:"short_url"`
	OUrl   string `json:"original_url"`
	UserID string `json:"user_id"`
}

type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	SOrl          string `json:"short_url"`
}

type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OUrl          string `json:"original_url"`
}

var AcceptedContentTypes = []string{
	"application/json",
	"text/html",
}

type ShortURLModel struct {
	OriginalURL string `db:"original_url"`
	UserID      string `db:"user_id"`
}

type GetShortURLInput struct {
	UUID   string
	UserID string
}

type CreateShortURLInput struct {
	OriginalURL string
	UserID      string
}

type UrlsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type UrlsStorage struct {
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
}

type Storage struct {
	UUID        string `db:"uuid"`
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
	UserID      string `db:"user_id"`
}
