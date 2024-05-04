package service

import (
	"context"
	"dogker/lintang/container-service/biz/domain"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type ContainerRepository interface {
	Get(ctx context.Context, serviceID string) (*domain.Container, error)
	GetAllUserContainer(ctx context.Context, userID string) (*[]domain.Container, error)
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
