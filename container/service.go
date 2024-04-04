package container

import (
	"containerServiceLintang/domain"
	"containerServiceLintang/internal/repository"
	"context"

	"github.com/google/uuid"
)

// ContainerRepository represent the container's repository contract
type ContainerRepository interface {
	Fetch(ctx context.Context, created_time string, limit int, page int) (res []domain.Container, pagination repository.Pagination, err error)
	GetById(ctx context.Context, id uuid.UUID) (domain.Container, error)
	SearchContainer(ctx context.Context, query string) (domain.Container, error)
	Update(ctx context.Context, c *domain.Container) error
	Store(ctx context.Context, c *domain.Container) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ContainerActionRepository interface {
	Fetch(ctx context.Context, created_time string, limit int, page int) (res []domain.ContainerAction, pagination repository.Pagination, err error)
	Store(ctx context.Context, ca *domain.ContainerAction) error
}


