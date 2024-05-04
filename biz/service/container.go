package service

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
)

type ContainerRepository interface {
	Get(ctx context.Context, serviceID string) (*domain.Container, error)
	GetAllUserContainer(ctx context.Context, userID string) (*[]domain.Container, error)
}

type ContainerService struct {
	containerRepo ContainerRepository
}

func NewContainerService(c ContainerRepository) *ContainerService {
	return &ContainerService{
		containerRepo: c,
	}
}

func (s *ContainerService) Hello(ctx context.Context) (string, error) {
	return "hello", nil
}
