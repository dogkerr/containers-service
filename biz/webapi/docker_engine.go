package webapi

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
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
					MemoryBytes: c.Limit.Memory * 1000000,
				},
				Reservations: &swarm.Resources{
					NanoCPUs:    c.Reservation.CPUs * 1000000,
					MemoryBytes: c.Reservation.Memory * 1000000,
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

/*


// ServiceStatus represents the number of running tasks in a service and the
// number of tasks desired to be running.
type ServiceStatus struct {
	// RunningTasks is the number of tasks for the service actually in the
	// Running state
	RunningTasks uint64




buat cari berapa yang masih running = replica-RunningTasks
*/

type dataServiceFromDB struct {
	TerminatedAt time.Time
	Lifecycles   []domain.ContainerLifecycle
	ID           string
}

// GetAllUserContainers
// @Description mendapatkan semua swarm service milik  user berdasarkan label user_id
func (d *DockerEngineAPI) GetAllUserContainers(ctx context.Context, userID string, cDB []domain.Container) (*[]domain.Container, error) {
	// var filterUserLabel map[string]string
	filterUserLabel := filters.Arg("label", "user_id="+userID)
	filter := filters.NewArgs(filterUserLabel)

	ctrDBData := make(map[string]dataServiceFromDB) // buat nyimpen data setiap service yg cuam disimpen di db
	for _, v := range cDB {
		ctrDBData[v.ServiceID] = dataServiceFromDB{
			ID:           v.ID,
			TerminatedAt: v.TerminatedTime,
			Lifecycles:   v.ContainerLifecycles,
		}
	}

	resp, err := d.Cli.ServiceList(ctx, types.ServiceListOptions{
		Filters: filter,
	})
	if err != nil {
		return nil, domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("user with id %s tidak memiliki container di dogker", userID))
	}
	var ctrs []domain.Container
	for _, v := range resp {
		filterServiceLabel := filters.Arg("service", v.Spec.Name)
		taskFilter := filters.NewArgs(filterServiceLabel)
		tasks, err := d.Cli.TaskList(ctx, types.TaskListOptions{Filters: taskFilter})

		if err != nil {
			zap.L().Error("d.Cli.ServiceInspectWithRaw", zap.Error(err), zap.String("serviceId", v.ID))
		}
		var runningTasks uint64 = 0
		for _, task := range tasks {
			if task.DesiredState == "running" {
				runningTasks += 1
			}
		}

		var status domain.ContainerStatus = domain.ContainerStatusRUN
		if runningTasks == 0 {
			status = domain.ContainerStatusSTOPPED
		}
		var ctrEndpoints []domain.Endpoint
		for _, portsConfig := range v.Endpoint.Ports {

			ctrEndpoints = append(ctrEndpoints, domain.Endpoint{
				TargetPort:    portsConfig.TargetPort,
				PublishedPort: uint64(portsConfig.PublishedPort),
				Protocol:      string(portsConfig.Protocol),
			})
		}
		ctrs = append(ctrs, domain.Container{
			UserID:              userID,
			Status:              status,
			Name:                v.Spec.Name,
			ContainerPort:       int(v.Spec.EndpointSpec.Ports[0].TargetPort),
			PublicPort:          int(v.Spec.EndpointSpec.Ports[0].PublishedPort),
			CreatedTime:         v.CreatedAt,
			TerminatedTime:      ctrDBData[v.ID].TerminatedAt,
			ContainerLifecycles: ctrDBData[v.ID].Lifecycles,
			ID:                  ctrDBData[v.ID].ID,
			ServiceID:           v.ID,
			Labels:              v.Spec.TaskTemplate.ContainerSpec.Labels,
			Replica:             *v.Spec.Mode.Replicated.Replicas,
			Limit: domain.Resource{
				CPUs:   v.Spec.TaskTemplate.Resources.Limits.NanoCPUs / 1000000,
				Memory: v.Spec.TaskTemplate.Resources.Limits.MemoryBytes / 1000000,
			},
			Reservation: domain.Resource{
				CPUs:   v.Spec.TaskTemplate.Resources.Reservations.NanoCPUs / 1000000,
				Memory: v.Spec.TaskTemplate.Resources.Reservations.MemoryBytes / 1000000,
			},
			Image:     v.Spec.TaskTemplate.ContainerSpec.Image,
			Env:       v.Spec.TaskTemplate.ContainerSpec.Env,
			Endpoint:  ctrEndpoints,
			Available: runningTasks,
		})
	}
	return &ctrs, nil
}

// Get
// @Description mendapatkan container by id
func (d *DockerEngineAPI) Get(ctx context.Context, ctrID string, cDB *domain.Container) (*domain.Container, error) {
	resp, _, err := d.Cli.ServiceInspectWithRaw(ctx, ctrID, types.ServiceInspectOptions{})
	if err != nil {
		// munngkin emang service dg id ctrID gak ada di docker
		zap.L().Debug("ServiceInspectWithRaw docker cli ", zap.Error(err), zap.String("ctrID", ctrID))
		return nil, domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf("container dengan id %s tidak terdaftar dalam sistem dogker", ctrID))
	}
	// alg buat tau service masih running gak
	filterServiceLabel := filters.Arg("service", resp.Spec.Name)
	taskFilter := filters.NewArgs(filterServiceLabel)
	tasks, err := d.Cli.TaskList(ctx, types.TaskListOptions{Filters: taskFilter})

	if err != nil {
		zap.L().Error("d.Cli.ServiceInspectWithRaw", zap.Error(err), zap.String("serviceId", resp.ID))
	}
	var runningTasks uint64 = 0
	for _, task := range tasks {
		if task.DesiredState == "running" {
			runningTasks += 1
		}
	}

	var status domain.ContainerStatus = domain.ContainerStatusRUN
	if runningTasks == 0 {
		status = domain.ContainerStatusSTOPPED
	}

	var ctrEndpoints []domain.Endpoint
	for _, portsConfig := range resp.Endpoint.Ports {

		ctrEndpoints = append(ctrEndpoints, domain.Endpoint{
			TargetPort:    portsConfig.TargetPort,
			PublishedPort: uint64(portsConfig.PublishedPort),
			Protocol:      string(portsConfig.Protocol),
		})
	}

	ctr := &domain.Container{
		ID:                  cDB.ID,
		UserID:              cDB.UserID,
		Status:              status,
		Name:                resp.Spec.Name,
		ContainerPort:       int(resp.Spec.EndpointSpec.Ports[0].TargetPort),
		PublicPort:          int(resp.Spec.EndpointSpec.Ports[0].PublishedPort),
		CreatedTime:         resp.CreatedAt,
		TerminatedTime:      cDB.TerminatedTime,
		ContainerLifecycles: cDB.ContainerLifecycles,
		ServiceID:           resp.ID,
		Labels:              resp.Spec.TaskTemplate.ContainerSpec.Labels,
		Replica:             *resp.Spec.Mode.Replicated.Replicas,
		Limit: domain.Resource{
			CPUs:   resp.Spec.TaskTemplate.Resources.Limits.NanoCPUs / 1000000,
			Memory: resp.Spec.TaskTemplate.Resources.Limits.MemoryBytes / 1000000,
		},
		Reservation: domain.Resource{
			CPUs:   resp.Spec.TaskTemplate.Resources.Reservations.NanoCPUs / 1000000,
			Memory: resp.Spec.TaskTemplate.Resources.Reservations.MemoryBytes / 1000000,
		},
		Image:     resp.Spec.TaskTemplate.ContainerSpec.Image,
		Env:       resp.Spec.TaskTemplate.ContainerSpec.Env,
		Endpoint:  ctrEndpoints,
		Available: runningTasks,
	}

	return ctr, nil

}
