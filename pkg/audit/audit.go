package audit

import (
	"encoding/json"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/klyakssa/test-repo-url/internal/model"
)

type Audit struct {
	file   *os.File
	client *resty.Client
}

func NewAudit(filePath, serverURL string) *Audit {
	audit := &Audit{}
	if filePath != "" {
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		audit.file = file
	}
	if serverURL != "" {
		audit.client = resty.New().SetTimeout(50 * time.Second).SetRetryWaitTime(1 * time.Second).SetRetryCount(3).SetBaseURL(serverURL)
	}
	return audit
}

func (a *Audit) Subscribe(audit *model.AuditEntry) {
	audit.Timestamp = time.Now().Unix()
	if a.file != nil {
		jsonData, err := json.Marshal(audit)
		if err != nil {
			return
		}
		a.file.WriteString(string(jsonData) + "\n")
	}
	if a.client != nil {
		a.client.R().SetBody(audit).Post("/")
	}
}
