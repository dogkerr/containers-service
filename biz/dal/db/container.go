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
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
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
		zap.L().Error("uuid.FromString(userID)", zap.Error(err))
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}
	ctrs, err := q.GetAllUserContainers(ctx, googleuuid.UUID(uid))
	if err != nil {
		if err == sql.ErrNoRows {
			hlog.Debug("container milik userId: "+userID+"tidak ada", err)
			return nil, domain.WrapErrorf(err, domain.ErrNotFound, "container milik userId: "+userID+"tidak ada")
		}
		zap.L().Error("q.GetAllUserContainers(ctx, googleuuid.UUID(uid))", zap.Error(err))
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
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
				CreatedTime:         ctr.CreatedTime.Time,
				ServiceID:           ctr.ServiceID,
				TerminatedTime:      terminatedtime,
				ContainerLifecycles: append(newCl, cLife),
				Replica:             uint64(ctr.Lifecyclereplica.Int32),
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
			zap.L().Debug("GetContainer (containerRepository)", zap.Error(err), zap.String("serviceID", serviceID))

			return nil, domain.WrapErrorf(err, domain.ErrNotFound, "container dengan id: "+serviceID+" tidak ada di database")
		}
		zap.L().Error("GetContainer (containerRepository)", zap.Error(err), zap.String("serviceID", serviceID))
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}
	var res domain.Container
	for _, ctr := range ctrs {
		cLife := domain.ContainerLifecycle{
			ID:          ctr.Lifeid.UUID.String(),
			ContainerID: ctr.ID.String(),
			StartTime:   ctr.Lifecyclestarttime.Time,
			StopTime:    ctr.Lifecyclestoptime.Time,
			Replica:     uint64(ctr.Lifecyclereplica.Int32),
			Status:      domain.ContainerStatus(ctr.Lifecyclestatus.ContainerStatus),
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
				CreatedTime:         ctr.CreatedTime.Time,
				ServiceID:           serviceID,
				TerminatedTime:      terminatedtime,
				ContainerLifecycles: append(newCl, cLife),
				Replica:             uint64(ctr.Lifecyclereplica.Int32),
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
		zap.L().Error("uuid.FromString(c.UserID)", zap.Error(err))
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}
	ctr, err := q.InsertContainer(ctx, queries.InsertContainerParams{
		UserID:         googleuuid.UUID(uid),
		Image:          c.Image,
		Status:         queries.ContainerStatusRUN,
		Name:           c.Name,
		ContainerPort:  int32(c.Endpoint[0].TargetPort),
		PublicPort:     pgtype.Int4{Valid: true, Int32: int32(c.Endpoint[0].PublishedPort)},
		TerminatedTime: pgtype.Timestamptz{Valid: false},
		CreatedTime:    pgtype.Timestamptz{Valid: true, Time: c.CreatedTime},
		ServiceID:      c.ServiceID,
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
		PublicPort:     pgtype.Int4{Valid: true, Int32: int32(c.PublicPort)},
		TerminatedTime: pgtype.Timestamptz{Valid: true, Time: c.TerminatedTime},
		CreatedTime:    pgtype.Timestamptz{Time: c.CreatedTime, Valid: true},
	})
	if err != nil {
		zap.L().Error(" q.UpdateContainer", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	return nil
}

func (r *ContainerRepository) Delete(ctx context.Context, serviceID string) error {
	q := queries.New(r.db.Pool)
	err := q.DeleteContainer(ctx, serviceID)
	if err != nil {
		zap.L().Error("q.DeleteContainer", zap.Error(err))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	return nil
}

func (r *ContainerRepository) InsertLifecycle(ctx context.Context, c *domain.ContainerLifecycle) (*domain.ContainerLifecycle, error) {
	q := queries.New(r.db.Pool)
	cID, err := uuid.FromString(c.ContainerID)
	if err != nil {
		zap.L().Error("uuid.FromString(c.ContainerID)", zap.Error(err), zap.String("cid", c.ContainerID))
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}
	ctr, err := q.InsertContainerLifecycle(ctx, queries.InsertContainerLifecycleParams{
		ContainerID: googleuuid.NullUUID{Valid: true, UUID: googleuuid.UUID(cID)},
		StartTime:   pgtype.Timestamptz{Time: c.StartTime, Valid: true},
		Status:      queries.ContainerStatus(c.Status),
		Replica:     int32(c.Replica),
	})

	if err != nil {
		zap.L().Error("q.InsertContainerLifecycle", zap.Error(err))
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
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
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	life, err := q.GetContainerLifecycle(ctx, googleuuid.UUID(iuid))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.WrapErrorf(err, domain.ErrNotFound, "container lifecycles dengan id: "+lifeId+" tidak ada dalam database")
		}
		return nil, domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	var stopTime time.Time
	if life.StopTime.Valid {
		stopTime = life.StopTime.Time
	}
	res := &domain.ContainerLifecycle{
		ID:          life.ID.String(),
		ContainerID: life.ContainerID.UUID.String(),
		StartTime:   life.StartTime.Time,
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
		zap.L().Error("uuid fromString", zap.Error(err), zap.String("lifeID", lifeId))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	err = q.UpdateContainerLifecycle(ctx, queries.UpdateContainerLifecycleParams{
		ID:       googleuuid.UUID(lifeIduuid),
		StopTime: pgtype.Timestamptz{Valid: true, Time: stopTime},
		Status:   queries.ContainerStatus(status),
		Replica:  int32(replica),
	})
	if err != nil {
		zap.L().Error("UpdateContainerLifecycle", zap.Error(err), zap.String("lifeID", lifeId))

		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	return nil
}

func (r *ContainerRepository) UpdateCtrLifeCycleWithoutStopTime(ctx context.Context, replica uint64, lifeID string) error {
	q := queries.New(r.db.Pool)
	lifeIduuid, err := uuid.FromString(lifeID)
	if err != nil {
		zap.L().Error("uuid fromString", zap.Error(err), zap.String("lifeID", lifeID))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, domain.MessageInternalServerError)
	}
	err = q.UpdateContainerLifeCycleWithoutStopTime(ctx, queries.UpdateContainerLifeCycleWithoutStopTimeParams{
		ID:      googleuuid.UUID(lifeIduuid),
		Replica: int32(replica),
	})
	return nil
}

func (r *ContainerRepository) InsertContainerMetrics(ctx context.Context, metrics domain.Metric) error {
	q := queries.New(r.db.Pool)

	ctrID, err := uuid.FromString(metrics.ContainerID)
	err = q.InsertIntoContainerMetrics(ctx, queries.InsertIntoContainerMetricsParams{
		ContainerID:    googleuuid.UUID(ctrID),
		Cpus:           float64(metrics.CpuUsage),
		Memory:         float64(metrics.MemoryUsage),
		NetworkIngress: float64(metrics.NetworkIngressUsage),
		NetworkEgress:  float64(metrics.NetworkEgressUsage),
	})
	if err != nil {
		zap.L().Error("InsertIntoContainerMetrics sqlc", zap.Error(err), zap.String("ctrID", metrics.ContainerID), zap.Float32("cpus", metrics.CpuUsage))
		return domain.WrapErrorf(err, domain.ErrInternalServerError, "internal server error")
	}

	return nil
}
