package db

import (
	"context"
	"database/sql"
	"dogker/lintang/container-service/biz/dal/db/queries"
	"dogker/lintang/container-service/biz/domain"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/gofrs/uuid"
	googleuuid "github.com/google/uuid"
)

type ContainerRepository struct {
	db *Postgres
}

func NewContainerRepo(db *Postgres) *ContainerRepository {
	return &ContainerRepository{db}
}

func (r *ContainerRepository) GetAllUserContainers(ctx context.Context, userID string) (*[]domain.Container, error) {
	q := queries.New(r.db.Pool)
	uid, err := uuid.FromString(userID)
	if err != nil {
		hlog.Error("uuid.FromString(userID)", err)
		return nil, err
	}
	ctrs, err := q.GetAllUserContainers(ctx, googleuuid.UUID(uid))
	if err != nil {
		if err == sql.ErrNoRows {
			hlog.Debug("container milik userId: "+userID+"tidak ada", err.Error())
			return nil, domain.ErrNotFound
		}
		hlog.Error("q.GetAllUserContainers(ctx, googleuuid.UUID(uid))", err.Error())
		return nil, err
	}
	var res []domain.Container
	for _, ctr := range ctrs {
		cLife := domain.ContainerLifecycle{
			ID:        ctr.Lifecycleid.UUID.String(),
			StartTime: ctr.Lifecyclestarttime.Time,
			StopTime:  ctr.Lifecyclestoptime.Time,
			Replica:   uint64(ctr.Lifecyclereplica.Int32),
			Status:    domain.ContainerStatus(ctr.Lifecyclestatus.ContainerStatus),
		}

		if (len(res) > 0 && res[len(res)-1].ID != ctr.ID.String()) || len(res) == 0 {
			var newCl []domain.ContainerLifecycle
			var terminatedtime time.Time
			var publicPort int
			if ctr.TerminatedTime.Valid {
				terminatedtime = ctr.TerminatedTime.Time
			}
			if ctr.PublicPort.Valid {
				publicPort = int(ctr.PublicPort.Int32)
			}
			res = append(res, domain.Container{
				ID:                  ctr.ID.String(),
				UserID:              ctr.UserID.String(),
				Image:               ctr.Image,
				Status:              domain.ContainerStatus(ctr.Status),
				Name:                ctr.Name,
				ContainerPort:       int(ctr.ContainerPort),
				PublicPort:          int(publicPort),
				CreatedTime:         ctr.CreatedTime,
				ServiceID:           ctr.ServiceID,
				TerminatedTime:      terminatedtime,
				ContainerLifecycles: append(newCl, cLife),
			})
		} else {
			res[len(res)-1].ContainerLifecycles = append(res[len(res)-1].ContainerLifecycles,
				cLife,
			)
		}
	}
	return &res, nil
}

func (r *ContainerRepository) Get(ctx context.Context, serviceID string) (*domain.Container, error) {
	q := queries.New(r.db.Pool)

	ctrs, err := q.GetContainer(ctx, serviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			hlog.Debug("container dengan id: "+serviceID+" tidak ada di database", err.Error())
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var res domain.Container
	for _, ctr := range ctrs {
		cLife := domain.ContainerLifecycle{
			ID:        ctr.Lifeid.UUID.String(),
			StartTime: ctr.Lifecyclestarttime.Time,
			StopTime:  ctr.Lifecyclestoptime.Time,
			Replica:   uint64(ctr.Lifecyclereplica.Int32),
			Status:    domain.ContainerStatus(ctr.Lifecyclestatus.ContainerStatus),
		}

		if res.Name == "" {
			var newCl []domain.ContainerLifecycle
			var publicPort int
			var terminatedtime time.Time
			if ctr.PublicPort.Valid {
				publicPort = int(ctr.PublicPort.Int32)
			}
			if ctr.TerminatedTime.Valid {
				terminatedtime = ctr.TerminatedTime.Time
			}
			res = domain.Container{
				ID:                  ctr.ID.String(),
				UserID:              ctr.UserID.String(),
				Image:               ctr.Image,
				Status:              domain.ContainerStatus(ctr.Status),
				Name:                ctr.Name,
				ContainerPort:       int(ctr.ContainerPort),
				PublicPort:          publicPort,
				CreatedTime:         ctr.CreatedTime,
				ServiceID:           serviceID,
				TerminatedTime:      terminatedtime,
				ContainerLifecycles: append(newCl, cLife),
			}
		} else {
			res.ContainerLifecycles = append(res.ContainerLifecycles,
				cLife,
			)
		}
	}
	return &res, nil
}

func (r *ContainerRepository) Insert(ctx context.Context, c *domain.Container) (*domain.Container, error) {
	q := queries.New(r.db.Pool)
	uid, err := uuid.FromString(c.UserID)
	if err != nil {
		hlog.Error(" uuid.FromString(c.UserID)", err.Error())
		return nil, err
	}
	ctr, err := q.InsertContainer(ctx, queries.InsertContainerParams{
		UserID: googleuuid.UUID(uid),
		Image:  c.Image,
	})
	c.ID = ctr.ID.String()
	return c, nil
}

func (r *ContainerRepository) Update(ctx context.Context, c *domain.Container) error {
	q := queries.New(r.db.Pool)
	err := q.UpdateContainer(ctx, queries.UpdateContainerParams{
		ServiceID:      c.ServiceID,
		Image:          c.Image,
		Status:         queries.ContainerStatus(c.Status),
		Name:           c.Name,
		ContainerPort:  int32(c.ContainerPort),
		PublicPort:     sql.NullInt32{Valid: true, Int32: int32(c.PublicPort)},
		TerminatedTime: sql.NullTime{Valid: true, Time: c.TerminatedTime},
		CreatedTime:    c.CreatedTime,
	})
	if err != nil {
		hlog.Error("q.UpdateContainer", err.Error())
		return err
	}
	return nil
}

func (r *ContainerRepository) Delete(ctx context.Context, serviceID string) error {
	q := queries.New(r.db.Pool)
	err := q.DeleteContainer(ctx, serviceID)
	if err != nil {
		hlog.Error("q.DeletContainer", err.Error())
		return err
	}
	return nil
}

func (r *ContainerRepository) InsertLifecycle(ctx context.Context, c *domain.ContainerLifecycle) (*domain.ContainerLifecycle, error) {
	q := queries.New(r.db.Pool)
	cID, err := uuid.FromString(c.ContainerID)
	if err != nil {
		hlog.Error("uuid.FromString(c.ContainerID)", err.Error())
		return nil, err
	}
	ctr, err := q.InsertContainerLifecycle(ctx, queries.InsertContainerLifecycleParams{
		ContainerID: googleuuid.UUID(cID),
		StartTime:   c.StartTime,
		Status:      queries.ContainerStatus(c.Status),
		Replica:     int32(c.Replica),
	})

	if err != nil {
		hlog.Error("q.InsertContainerLifecycle", err.Error())
		return nil, err
	}
	res := &domain.ContainerLifecycle{
		ID:          ctr.ID.String(),
		ContainerID: c.ContainerID,
		StartTime:   c.StartTime,
		Replica:     c.Replica,
		Status:      c.Status,
		StopTime:    c.StopTime,
	}
	return res, nil
}

func (r *ContainerRepository) GetLifecycle(ctx context.Context, lifeId string) (*domain.ContainerLifecycle, error) {
	q := queries.New(r.db.Pool)
	iuid, err := uuid.FromString(lifeId)
	if err != nil {
		return nil, err
	}
	life, err := q.GetContainerLifecycle(ctx, googleuuid.UUID(iuid))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	var stopTime time.Time
	if life.StopTime.Valid {
		stopTime = life.StopTime.Time
	}
	res := &domain.ContainerLifecycle{
		ID:          life.ID.String(),
		ContainerID: life.ContainerID.String(),
		StartTime:   life.StartTime,
		StopTime:    stopTime,
		Status:      domain.ContainerStatus(life.Status),
		Replica:     uint64(life.Replica),
	}
	return res, nil
}

func (r *ContainerRepository) UpdateLifecycle(ctx context.Context, lifeId string, stopTime time.Time, status domain.ContainerStatus, replica uint32) error {
	q := queries.New(r.db.Pool)
	lifeIduuid, err := uuid.FromString(lifeId)
	if err != nil {
		return err
	}
	err = q.UpdateContainerLifecycle(ctx, queries.UpdateContainerLifecycleParams{
		ID:       googleuuid.UUID(lifeIduuid),
		StopTime: sql.NullTime{Valid: true, Time: stopTime},
		Status:   queries.ContainerStatus(status),
		Replica:  int32(replica),
	})
	if err != nil {
		return err
	}
	return nil
}
