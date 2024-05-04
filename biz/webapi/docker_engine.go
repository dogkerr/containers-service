package webapi

import (
	"context"
	"dogker/lintang/container-service/biz/domain"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
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
		hlog.Error(" d.Cli.ServiceCreate", err)
		return "", err
	}
	return resp.ID, nil
}
