package model

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type FileStorageData struct {
	UUID string `json:"uuid"`
	SUrl string `json:"short_url"`
	OUrl string `json:"original_url"`
}

var AcceptedContentTypes = []string{
	"application/json",
	"text/html",
}
