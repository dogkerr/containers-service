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
	BaseURL      string
	MyServiceURL string
}

func CreateDkronAPI(cfg *config.Config) *DkronAPI {
	return &DkronAPI{
		BaseURL:      cfg.Dkron.DkronURL,
		MyServiceURL: cfg.MyServiceURL,
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

func (d *DkronAPI) InstallCURL(ctx context.Context) error {
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
		zap.L().Error("Marshal JSON", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}

	req, err := http.NewRequest("POST", d.BaseURL, bytes.NewBuffer(payload))

	if err != nil {
		zap.L().Error("NewRequest ", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("client.Do(req) ", zap.Error(err))
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

	if action == domain.StartContainer {
		cronURL = fmt.Sprintf("http://%s:8888/api/v1/containers/scheduler/%s/start", d.MyServiceURL, ctrID)
	} else if action == domain.StopContainer {
		cronURL = fmt.Sprintf("http://%s:8888/api/v1/containers/scheduler/%s/stop", d.MyServiceURL, ctrID)
	} else if action == domain.TerminateContainer {
		cronURL = fmt.Sprintf("http://%s:8888/api/v1/containers/scheduler/%s/terminate", d.MyServiceURL, ctrID)
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

func (d *DkronAPI) AddCreateJob(ctx context.Context, schedule uint64, action domain.ContainerAction, userID string, ctr *domain.Container) error {
	randomString := uuid.New().String()
	var cronURL string

	cronURL = fmt.Sprintf("http://%s:8888/api/v1/containers/scheduler/create", d.MyServiceURL)

	jobName := userID + randomString

	var commandString map[string]string

	var reserv, labels, envs string = "", "", ""

	if ctr.Reservation.CPUs != 0 {
		reserv = `"reservation": {
			"cpus": ` + fmt.Sprint(ctr.Reservation.CPUs) + `,
			"memory": ` + fmt.Sprint(ctr.Reservation.Memory) + `
		},`
	}
	if ctr.Labels != nil {
		var labelsItems string = `"tes": "tes"`
		for k, v := range ctr.Labels {
			labelsItems += `, "` + k + `": "` + v + `"`
		}
		labels = `"labels": {
			` + labelsItems + `
		},
		`
	}
	if ctr.Env != nil {
		var envItems string = "TES=TES"
		for _, v := range ctr.Env {
			envItems += `
				,` + v + `
			`
		}

		envs = `"env": [
			"` + envItems + `"
		],`
	}

	var endpoints string
	var endpointItems string = ""

	for i, v := range ctr.Endpoint {
		if len(ctr.Endpoint) == 1 && endpointItems == "" {
			endpointItems += `{
				"target_port": ` + fmt.Sprint(v.TargetPort) + `,
				"published_port": ` + fmt.Sprint(v.PublishedPort) + `,
				"protocol": "` + v.Protocol + `"
			   }`
		} else {
			endpointItems += `{
				"target_port": ` + fmt.Sprint(v.TargetPort) + `,
				"published_port": ` + fmt.Sprint(v.PublishedPort) + `,
				"protocol": "` + v.Protocol + `"
			   },`
		}

		if i == len(ctr.Endpoint)-1 && len(ctr.Endpoint) != 1 {
			endpointItems += `{
				"target_port": ` + fmt.Sprint(v.TargetPort) + `,
				"published_port": ` + fmt.Sprint(v.PublishedPort) + `,
				"protocol": "` + v.Protocol + `"
			   }`
		}

	}
	endpoints = `
		[
		` + endpointItems + `
		]
	`

	commandString = map[string]string{
		"command": `curl --location  ` + cronURL + `  \
		--header 'Content-Type: application/json' \
		--data '{
			"name": "` + ctr.Name + `",
			"image": "` + ctr.Image + `",
			"limit": {
				"cpus": ` + fmt.Sprint(ctr.Limit.CPUs) + `,
				"memory": ` + fmt.Sprint(ctr.Limit.Memory) + `
			},
			` + reserv + `
			"replica": ` + fmt.Sprint(ctr.Replica) + `,
			` + labels + `
			` + envs + `
			"endpoint": ` + endpoints + `,
			"user_id": "`+ userID +`"
		}'`}

	at := time.Now().Add(time.Duration(schedule) * time.Second)
	payload, err := json.Marshal(JobReq{
		Name:           jobName,
		DisplayName:    jobName,
		Schedule:       fmt.Sprintf("@at " + at.Format(time.RFC3339)),
		Timezone:       "Asia/Jakarta",
		Owner:          "lintang birda saputra",
		OwnerEmail:     "lintangbirdasaputra23@gmail.com",
		Disabled:       false,
		Concurrency:    "allow",
		Executor:       "shell",
		ExecutorConfig: commandString,
	})
	if err != nil {
		zap.L().Error("Marshal JSON", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}

	req, err := http.NewRequest("POST", d.BaseURL, bytes.NewBuffer(payload))

	if err != nil {
		zap.L().Error("NewRequest ", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("client.Do(req) ", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	defer resp.Body.Close()

	return nil
}
