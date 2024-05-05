// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: container_query.sql

package queries

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const deleteContainer = `-- name: DeleteContainer :exec
DELETE FROM containers
WHERE service_id=$1
`

func (q *Queries) DeleteContainer(ctx context.Context, serviceID string) error {
	_, err := q.db.ExecContext(ctx, deleteContainer, serviceID)
	return err
}

const getAllUserContainers = `-- name: GetAllUserContainers :many
SELECT c.id, c.user_id, c.image, c.status, c.name, c.container_port, c.public_port, c.created_time,c.service_id, c.terminated_time,
			cl.id as lifecycleId, cl.start_time as lifecycleStartTime, cl.stop_time as lifecycleStopTime, 
			cl.replica as lifecycleReplica, cl.status as lifecycleStatus FROM containers c  LEFT JOIN container_lifecycles cl ON cl.container_id=c.id
			WHERE c.user_id=$1
`

type GetAllUserContainersRow struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Image              string
	Status             ContainerStatus
	Name               string
	ContainerPort      int32
	PublicPort         sql.NullInt32
	CreatedTime        time.Time
	ServiceID          string
	TerminatedTime     sql.NullTime
	Lifecycleid        uuid.NullUUID
	Lifecyclestarttime sql.NullTime
	Lifecyclestoptime  sql.NullTime
	Lifecyclereplica   sql.NullInt32
	Lifecyclestatus    NullContainerStatus
}

func (q *Queries) GetAllUserContainers(ctx context.Context, userID uuid.UUID) ([]GetAllUserContainersRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllUserContainers, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllUserContainersRow
	for rows.Next() {
		var i GetAllUserContainersRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Image,
			&i.Status,
			&i.Name,
			&i.ContainerPort,
			&i.PublicPort,
			&i.CreatedTime,
			&i.ServiceID,
			&i.TerminatedTime,
			&i.Lifecycleid,
			&i.Lifecyclestarttime,
			&i.Lifecyclestoptime,
			&i.Lifecyclereplica,
			&i.Lifecyclestatus,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getContainer = `-- name: GetContainer :many
SELECT c.id, c.user_id, c.image, c.status, c.name, c.container_port, c.public_port,c.created_time,
	c.service_id,c.terminated_time, cl.id as lifeId, cl.start_time as lifecycleStartTime, cl.stop_time as lifecycleStopTime, cl.replica  as lifecycleReplica, cl.status as lifecycleStatus 
	FROM containers c LEFT JOIN container_lifecycles cl ON cl.container_id=c.id
	WHERE c.service_id=$1
`

type GetContainerRow struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Image              string
	Status             ContainerStatus
	Name               string
	ContainerPort      int32
	PublicPort         sql.NullInt32
	CreatedTime        time.Time
	ServiceID          string
	TerminatedTime     sql.NullTime
	Lifeid             uuid.NullUUID
	Lifecyclestarttime sql.NullTime
	Lifecyclestoptime  sql.NullTime
	Lifecyclereplica   sql.NullInt32
	Lifecyclestatus    NullContainerStatus
}

func (q *Queries) GetContainer(ctx context.Context, serviceID string) ([]GetContainerRow, error) {
	rows, err := q.db.QueryContext(ctx, getContainer, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetContainerRow
	for rows.Next() {
		var i GetContainerRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Image,
			&i.Status,
			&i.Name,
			&i.ContainerPort,
			&i.PublicPort,
			&i.CreatedTime,
			&i.ServiceID,
			&i.TerminatedTime,
			&i.Lifeid,
			&i.Lifecyclestarttime,
			&i.Lifecyclestoptime,
			&i.Lifecyclereplica,
			&i.Lifecyclestatus,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getContainerLifecycle = `-- name: GetContainerLifecycle :one
SELECT id, container_id, start_time, stop_time, status, replica
FROM container_lifecycles
WHERE id=$1
`

func (q *Queries) GetContainerLifecycle(ctx context.Context, id uuid.UUID) (ContainerLifecycle, error) {
	row := q.db.QueryRowContext(ctx, getContainerLifecycle, id)
	var i ContainerLifecycle
	err := row.Scan(
		&i.ID,
		&i.ContainerID,
		&i.StartTime,
		&i.StopTime,
		&i.Status,
		&i.Replica,
	)
	return i, err
}

const getContainerWithPagination = `-- name: GetContainerWithPagination :many
SELECT c.id, c.user_id, c.image, c.status, c.name, c.container_port, c.public_port,c.created_time,
	c.service_id,c.terminated_time, cl.id as lifeId, cl.start_time as lifecycleStartTime, cl.stop_time as lifecycleStopTime, cl.replica  as lifecycleReplica, cl.status as lifecycleStatus 
	FROM containers c LEFT JOIN container_lifecycles cl ON cl.container_id=c.id
	WHERE c.service_id=$1
	LIMIT $2 OFFSET $3
`

type GetContainerWithPaginationParams struct {
	ServiceID string
	Limit     int32
	Offset    int32
}

type GetContainerWithPaginationRow struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Image              string
	Status             ContainerStatus
	Name               string
	ContainerPort      int32
	PublicPort         sql.NullInt32
	CreatedTime        time.Time
	ServiceID          string
	TerminatedTime     sql.NullTime
	Lifeid             uuid.NullUUID
	Lifecyclestarttime sql.NullTime
	Lifecyclestoptime  sql.NullTime
	Lifecyclereplica   sql.NullInt32
	Lifecyclestatus    NullContainerStatus
}

func (q *Queries) GetContainerWithPagination(ctx context.Context, arg GetContainerWithPaginationParams) ([]GetContainerWithPaginationRow, error) {
	rows, err := q.db.QueryContext(ctx, getContainerWithPagination, arg.ServiceID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetContainerWithPaginationRow
	for rows.Next() {
		var i GetContainerWithPaginationRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Image,
			&i.Status,
			&i.Name,
			&i.ContainerPort,
			&i.PublicPort,
			&i.CreatedTime,
			&i.ServiceID,
			&i.TerminatedTime,
			&i.Lifeid,
			&i.Lifecyclestarttime,
			&i.Lifecyclestoptime,
			&i.Lifecyclereplica,
			&i.Lifecyclestatus,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertContainer = `-- name: InsertContainer :one
INSERT INTO containers (
	user_id, image, status, name, container_port, public_port, terminated_time, created_time, service_id
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING id, user_id, image, status, name, container_port, public_port, terminated_time, created_time, service_id
`

type InsertContainerParams struct {
	UserID         uuid.UUID
	Image          string
	Status         ContainerStatus
	Name           string
	ContainerPort  int32
	PublicPort     sql.NullInt32
	TerminatedTime sql.NullTime
	CreatedTime    time.Time
	ServiceID      string
}

func (q *Queries) InsertContainer(ctx context.Context, arg InsertContainerParams) (Container, error) {
	row := q.db.QueryRowContext(ctx, insertContainer,
		arg.UserID,
		arg.Image,
		arg.Status,
		arg.Name,
		arg.ContainerPort,
		arg.PublicPort,
		arg.TerminatedTime,
		arg.CreatedTime,
		arg.ServiceID,
	)
	var i Container
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Image,
		&i.Status,
		&i.Name,
		&i.ContainerPort,
		&i.PublicPort,
		&i.TerminatedTime,
		&i.CreatedTime,
		&i.ServiceID,
	)
	return i, err
}

const insertContainerLifecycle = `-- name: InsertContainerLifecycle :one
INSERT INTO container_lifecycles(
	container_id, start_time, stop_time, status, replica
) VALUES (
	$1, $2, $3, $4, $5
) RETURNING id, container_id, start_time, stop_time, status, replica
`

type InsertContainerLifecycleParams struct {
	ContainerID uuid.NullUUID
	StartTime   time.Time
	StopTime    sql.NullTime
	Status      ContainerStatus
	Replica     int32
}

func (q *Queries) InsertContainerLifecycle(ctx context.Context, arg InsertContainerLifecycleParams) (ContainerLifecycle, error) {
	row := q.db.QueryRowContext(ctx, insertContainerLifecycle,
		arg.ContainerID,
		arg.StartTime,
		arg.StopTime,
		arg.Status,
		arg.Replica,
	)
	var i ContainerLifecycle
	err := row.Scan(
		&i.ID,
		&i.ContainerID,
		&i.StartTime,
		&i.StopTime,
		&i.Status,
		&i.Replica,
	)
	return i, err
}

const updateContainer = `-- name: UpdateContainer :exec
UPDATE containers
SET 
	image=$2,
	status=$3,
	name=$4,
	container_port=$5,
	public_port=$6,
	terminated_time=$7,
	created_time=$8
WHERE service_id=$1
`

type UpdateContainerParams struct {
	ServiceID      string
	Image          string
	Status         ContainerStatus
	Name           string
	ContainerPort  int32
	PublicPort     sql.NullInt32
	TerminatedTime sql.NullTime
	CreatedTime    time.Time
}

func (q *Queries) UpdateContainer(ctx context.Context, arg UpdateContainerParams) error {
	_, err := q.db.ExecContext(ctx, updateContainer,
		arg.ServiceID,
		arg.Image,
		arg.Status,
		arg.Name,
		arg.ContainerPort,
		arg.PublicPort,
		arg.TerminatedTime,
		arg.CreatedTime,
	)
	return err
}

const updateContainerLifeCycleWithoutStopTime = `-- name: UpdateContainerLifeCycleWithoutStopTime :exec
UPDATE container_lifecycles
SET 
	replica=$2
WHERE id=$1
`

type UpdateContainerLifeCycleWithoutStopTimeParams struct {
	ID      uuid.UUID
	Replica int32
}

func (q *Queries) UpdateContainerLifeCycleWithoutStopTime(ctx context.Context, arg UpdateContainerLifeCycleWithoutStopTimeParams) error {
	_, err := q.db.ExecContext(ctx, updateContainerLifeCycleWithoutStopTime, arg.ID, arg.Replica)
	return err
}

const updateContainerLifecycle = `-- name: UpdateContainerLifecycle :exec
UPDATE container_lifecycles
SET 
	stop_time=$2,
	status=$3,
	replica=$4
WHERE id=$1
`

type UpdateContainerLifecycleParams struct {
	ID       uuid.UUID
	StopTime sql.NullTime
	Status   ContainerStatus
	Replica  int32
}

func (q *Queries) UpdateContainerLifecycle(ctx context.Context, arg UpdateContainerLifecycleParams) error {
	_, err := q.db.ExecContext(ctx, updateContainerLifecycle,
		arg.ID,
		arg.StopTime,
		arg.Status,
		arg.Replica,
	)
	return err
}
