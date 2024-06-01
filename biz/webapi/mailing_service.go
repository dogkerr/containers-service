package webapi

import (
	"bytes"
	"context"
	"dogker/lintang/container-service/biz/domain"
	"dogker/lintang/container-service/config"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type MailingServiceWebAPI struct {
	MailingURL string
}
type CommonLabels struct {
	Alertname                       string `json:"alertname"`
	ContainerSwarmServiceID         string `json:"container_label_com_docker_swarm_service_id"`
	ContainerDockerSwarmServiceName string `json:"container_label_com_docker_swarm_service_name"`
	ContainerLabelUserID            string `json:"container_label_user_id"`
}

type PromeWebhookReq struct {
	Receiver     string       `json:"receiver"`
	CommonLabels CommonLabels `json:"commonLabels"`
}

func NewMailingServiceWEbAPI(cfg *config.Config) *MailingServiceWebAPI {
	return &MailingServiceWebAPI{
		MailingURL: cfg.Mailing.MailingURL,
	}
}

func (m *MailingServiceWebAPI) SendContainerDown(ctx context.Context, label CommonLabels) error {
	payload, err := json.Marshal(PromeWebhookReq{
		Receiver:     "webhook_receiver_emailing_service",
		CommonLabels: label,
	},
	)
	if err != nil {
		zap.L().Error("Marshal JSON (SendContainerDown) (MailingServiceWEBAPI)", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	req, err := http.NewRequest("POST", m.MailingURL+"/api/v1/email/down", bytes.NewBuffer(payload))
	if err != nil {
		zap.L().Error("NewRequest (SendContainerDown )  (MailingServiceWEBAPI)", zap.Error(err))
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


