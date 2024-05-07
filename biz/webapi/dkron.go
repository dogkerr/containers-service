package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"dogker/lintang/container-service/biz/domain"
	"dogker/lintang/container-service/config"
)

type DkronAPI struct {
	BaseURL string
}

func CreateDkronAPI(cfg *config.Config) *DkronAPI {
	return &DkronAPI{
		BaseURL: cfg.Dkron.DkronURL,
	}
}

type JobReq struct {
	Name           string            `json:"name"`
	DisplayName    string            `json:"displayname"`
	Schedule       string            `json:"schedule"`
	Timezone       string            `json:"timezone"`
	Owner          string            `json:"owner"`
	OwnerEmail     string            `json:"owner_email"`
	Disabled       bool              `json:"disabled"`
	Concurrency    string            `json:"concurrency"`
	Executor       string            `json:"executor"`
	ExecutorConfig map[string]string `json:"executor_config"`
}


func (d *DkronAPI) InstallCURL(ctx context.Context) error  {
	at := time.Now().Add(time.Duration(2) * time.Second)

	payload, err := json.Marshal(JobReq{
		Name:        "insatll curl",
		DisplayName: "insatll curl",
		Schedule:    fmt.Sprintf("@at " + at.Format(time.RFC3339)),
		Timezone:    "Asia/Jakarta",
		Owner:       "lintang birda saputra",
		OwnerEmail:  "lintangbirdasaputra23@gmail.com",
		Disabled:    false,
		Concurrency: "allow",
		Executor:    "shell",
		ExecutorConfig: map[string]string{
			"command": `sh /curl/curl.sh'`,
		},
	})
	if err != nil {
		zap.L().Error("Marshal JSON", zap.Error(err), )
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}

	req, err := http.NewRequest("POST", d.BaseURL, bytes.NewBuffer(payload))

	if err != nil {
		zap.L().Error("NewRequest ", zap.Error(err), )
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("client.Do(req) ", zap.Error(err), )
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	defer resp.Body.Close()
	return nil
}

// AddJob -.
// @Description menambah cron job baru di dkron
// schedule dalam satuan second (1- tak hingga)
func (d *DkronAPI) AddJob(ctx context.Context, schedule uint64, ctrID string, action domain.ContainerAction, userID string) error {
	randomString := uuid.New().String()
	var cronURL string

	if action == domain.CreateContainer {
		cronURL = fmt.Sprintf("http://container-service:8888/api/v1/containers/scheduler/%s/create", ctrID)
	} else if action == domain.StartContainer {
		cronURL = fmt.Sprintf("http://container-service:8888/api/v1/containers/scheduler/%s/start", ctrID)
	} else if action == domain.StopContainer {
		cronURL = fmt.Sprintf("http://container-service:8888/api/v1/containers/scheduler/%s/stop", ctrID)
	} else if action == domain.TerminateContainer {
		cronURL = fmt.Sprintf("http://container-service:8888/api/v1/containers/scheduler/%s/terminate", ctrID)
	}

	jobName := ctrID + randomString

	at := time.Now().Add(time.Duration(schedule) * time.Second)

	payload, err := json.Marshal(JobReq{
		Name:        jobName,
		DisplayName: jobName,
		Schedule:    fmt.Sprintf("@at " + at.Format(time.RFC3339)),
		Timezone:    "Asia/Jakarta",
		Owner:       "lintang birda saputra",
		OwnerEmail:  "lintangbirdasaputra23@gmail.com",
		Disabled:    false,
		Concurrency: "allow",
		Executor:    "shell",
		ExecutorConfig: map[string]string{
			// "shell": "true",
			"command": `curl --location ` + cronURL + ` \
			--header 'Content-Type: application/json' \
			--data '{
				"user_id": "` + userID + `"
			}'`,
		},
	})
	if err != nil {
		zap.L().Error("Marshal JSON", zap.Error(err), zap.String("ctrID", ctrID))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}

	req, err := http.NewRequest("POST", d.BaseURL, bytes.NewBuffer(payload))

	if err != nil {
		zap.L().Error("NewRequest ", zap.Error(err), zap.String("ctrID", ctrID))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("client.Do(req) ", zap.Error(err), zap.String("ctrID", ctrID))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	defer resp.Body.Close()

	return nil
}
