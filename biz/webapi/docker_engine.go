package webapi

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type DockerEngineAPI struct {
	Cli *client.Client
}

func CreateNewDockerEngineAPI() *DockerEngineAPI {
	apiclient, err := client.NewClientWithOpts(client.WithHost("unix:///var/run/docker.sock"), client.WithAPIVersionNegotiation())

	if err != nil {
		hlog.Fatal("client.NewClientWithOpts ", err)
	}

	return &DockerEngineAPI{Cli: apiclient}

}

/*
misal c.Limit.CPUs = 1000 / 1 cpus / 1000 milicpus ,
kalo ke nannocpu berarti 1000000000 / 10^9
bearrti ke nano cpu -> cpus * 10^6

*/

func (d *DockerEngineAPI) CreateService(ctx context.Context, c *domain.Container) (string, error) {

	var portsConfig []swarm.PortConfig
	for _, v := range c.Endpoint {
		portsConfig = append(portsConfig, swarm.PortConfig{
			TargetPort:    uint32(v.TargetPort),
			PublishedPort: uint32(v.PublishedPort),
			Protocol:      swarm.PortConfigProtocol(v.Protocol),
		})
	}
	if len(c.Labels) == 0 {
		var ownLabel map[string]string = map[string]string{"user_id": c.UserID}
		c.Labels = ownLabel

	} else {
		c.Labels["user_id"] = c.UserID
	}

	resp, err := d.Cli.ServiceCreate(ctx, swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{

			ContainerSpec: &swarm.ContainerSpec{
				Image:  c.Image,
				Labels: c.Labels,
				Env:    c.Env,
			},
			Resources: &swarm.ResourceRequirements{
				Limits: &swarm.Limit{
					NanoCPUs:    c.Limit.CPUs * 1000000,
					MemoryBytes: c.Limit.Memory / 1000000,
				},
				Reservations: &swarm.Resources{
					NanoCPUs:    c.Reservation.CPUs * 1000000,
					MemoryBytes: c.Reservation.Memory / 1000000,
				},
			},
			LogDriver: &swarm.Driver{
				Name: "loki",
				Options: map[string]string{
					"loki-url":             "http://localhost:3100/loki/api/v1/push",
					"loki-retries":         "5",
					"loki-batch-size":      "400",
					"loki-external-labels": "job=docker,container_name=go_container_log1,userId=" + c.UserID,
				},
			},
		},
		Annotations: swarm.Annotations{
			Name:   c.Name,
			Labels: c.Labels,
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &c.Replica,
			},
		},
		EndpointSpec: &swarm.EndpointSpec{
			Ports: portsConfig,
		},
	}, types.ServiceCreateOptions{})
	if err != nil {
		fmt.Println(c.Endpoint[0].PublishedPort)
		if strings.Contains(err.Error(), "already in use") {
			return "", domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("port %d already in use", c.Endpoint[0].PublishedPort))
		}
		// hlog.Error(" d.Cli.ServiceCreate", err)
		zap.L().Error(" d.Cli.ServiceCreate", zap.Error(err))
		return "", domain.WrapErrorf(err, domain.ErrInternalServerError, " internal server error")
	}

	return resp.ID, nil
}
