package service

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type ContainerRepository interface {
	Get(ctx context.Context, serviceID string) (*domain.Container, error)
	GetAllUserContainers(ctx context.Context, userID string) (*[]domain.Container, error)
	Insert(ctx context.Context, c *domain.Container) (*domain.Container, error)
	Update(ctx context.Context, c *domain.Container) error
	Delete(ctx context.Context, serviceID string) error
	InsertLifecycle(ctx context.Context, c *domain.ContainerLifecycle) (*domain.ContainerLifecycle, error)
	GetLifecycle(ctx context.Context, lifeId string) (*domain.ContainerLifecycle, error)
	UpdateLifecycle(ctx context.Context, lifeId string, stopTime time.Time, status domain.ContainerStatus, replica uint32) error
}

type DockerEngineAPI interface {
	CreateService(ctx context.Context, c *domain.Container) (string, error)
}

type ContainerService struct {
	containerRepo ContainerRepository
	dockerAPI     DockerEngineAPI
}

func NewContainerService(c ContainerRepository, d DockerEngineAPI) *ContainerService {
	return &ContainerService{
		containerRepo: c,
		dockerAPI:     d,
	}
}

func (s *ContainerService) Hello(ctx context.Context) (string, error) {
	return "hello", nil
}

func (s *ContainerService) CreateNewService(ctx context.Context, d *domain.Container) (string, error) {
	serviceId, err := s.dockerAPI.CreateService(ctx, d)
	if err != nil {
		hlog.Error("s.dockerAPI.CreateService", err)
		return "", err
	}

	return serviceId, nil
}
