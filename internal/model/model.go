package model

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type FileStorageData struct {
	Uuid  string `json:"uuid"`
	S_URL string `json:"short_url"`
	O_URL string `json:"original_url"`
}

var AcceptedContentTypes = []string{
	"application/json",
	"text/html",
}
