package service

import (
	"context"
	"dogker/lintang/container-service/biz/domain"
	"fmt"
	"time"

	"go.uber.org/zap"
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
	GetAllUserContainers(ctx context.Context, userID string, cDB []domain.Container) (*[]domain.Container, error)
	Get(ctx context.Context, ctrID string, cDB *domain.Container) (*domain.Container, error)
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

func (s *ContainerService) CreateNewService(ctx context.Context, d *domain.Container) (string, time.Time, *domain.ContainerLifecycle, error) {
	serviceId, err := s.dockerAPI.CreateService(ctx, d)
	if err != nil {
		zap.L().Error("s.dockerAPI.CreateService", zap.Error(err))
		return "", time.Now(), nil, err
	}

	d.ServiceID = serviceId
	d.CreatedTime = time.Now()

	ctrRowId, err := s.containerRepo.Insert(ctx, d)
	if err != nil {
		return "", time.Now(), nil, err
	}

	ctrLife, err := s.containerRepo.InsertLifecycle(ctx, &domain.ContainerLifecycle{
		ID:        ctrRowId.ID,
		StartTime: d.CreatedTime,
		Status:    domain.ContainerStatusRUN,
		Replica:   d.Replica,
	})
	if err != nil {
		return "", time.Now(), nil, err
	}
	return serviceId, d.CreatedTime, ctrLife, nil
}

func (s *ContainerService) GetUserContainers(ctx context.Context, userID string, offset uint64, limit uint64) (*[]domain.Container, error) {
	userCtrsDb, err := s.containerRepo.GetAllUserContainers(ctx, userID)
	if err != nil {
		return nil, err
	}
	ctrsDocker, err := s.dockerAPI.GetAllUserContainers(ctx, userID, *userCtrsDb)
	if err != nil {
		return nil, err
	}

	
	return ctrsDocker, nil
}

func (s *ContainerService) GetContainer(ctx context.Context, ctrID string, userID string) (*domain.Container, error) {
	ctrDB, err := s.containerRepo.Get(ctx, ctrID)
	if err != nil {
		return nil, err 
	}
	if ctrDB.UserID != userID {
		return nil, domain.WrapErrorf(err, domain.ErrBadParamInput, fmt.Sprintf( "container %s bukan milik anda", ctrID))
	}
	ctrDocker, err := s.dockerAPI.Get(ctx, ctrID, ctrDB)
	if err  != nil {
		return nil, err
	}
	return ctrDocker, nil
}



