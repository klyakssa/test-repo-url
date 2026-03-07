package mapper

import "github.com/klyakssa/test-repo-url/internal/model"

func ToUrlsResponse(urls []model.UrlsStorage) []model.UrlsResponse {
	responses := make([]model.UrlsResponse, len(urls))
	for i, url := range urls {
		responses[i] = model.UrlsResponse{
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
		}
	}
	return responses
}
