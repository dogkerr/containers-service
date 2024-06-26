-- name: GetAllUserContainers :many
SELECT c.id, c.user_id, c.image, c.status, c.name, c.container_port, c.public_port, c.created_time,c.service_id, c.terminated_time,
			cl.id as lifecycleId, cl.start_time as lifecycleStartTime, cl.stop_time as lifecycleStopTime, 
			cl.replica as lifecycleReplica, cl.status as lifecycleStatus FROM containers c  LEFT JOIN container_lifecycles cl ON cl.container_id=c.id
			WHERE c.user_id=$1;



-- name: GetContainer :many
SELECT c.id, c.user_id, c.image, c.status, c.name, c.container_port, c.public_port,c.created_time,
	c.service_id,c.terminated_time, cl.id as lifeId, cl.start_time as lifecycleStartTime, cl.stop_time as lifecycleStopTime, cl.replica  as lifecycleReplica, cl.status as lifecycleStatus 
	FROM containers c LEFT JOIN container_lifecycles cl ON cl.container_id=c.id
	WHERE c.service_id=$1;

-- name: GetContainerWithPagination :many
SELECT c.id, c.user_id, c.image, c.status, c.name, c.container_port, c.public_port,c.created_time,
	c.service_id,c.terminated_time, cl.id as lifeId, cl.start_time as lifecycleStartTime, cl.stop_time as lifecycleStopTime, cl.replica  as lifecycleReplica, cl.status as lifecycleStatus 
	FROM containers c LEFT JOIN container_lifecycles cl ON cl.container_id=c.id
	WHERE c.service_id=$1
	LIMIT $2 OFFSET $3;


-- name: GetContainersByIDs :many
SELECT c.id, c.user_id, c.image, c.status, c.name, c.container_port, c.public_port,c.created_time,
	c.service_id,c.terminated_time
	FROM containers c 
	WHERE c.service_id = ANY($1::varchar[]);


-- name: GetStoppedContainer :many
SELECT c.id, c.service_id, c.name
	FROM containers c
	WHERE c.status = $1;

-- name: GetContainersByStatus :many
SELECT c.id, c.service_id, c.name, cl.id as lifeId, cl.start_time as lifecycleStartTime, cl.stop_time as lifecycleStopTime,  cl.status as lifecycleStatus ,
	cl.container_id as cLifeCtrID
	FROM containers c
	LEFT JOIN container_lifecycles cl ON cl.container_id=c.id
	WHERE c.status = $1;

-- name: InsertContainer :one
INSERT INTO containers (
	user_id, image, status, name, container_port, public_port, terminated_time, created_time, service_id
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: BatchInsertContainerMetrics :copyfrom
INSERT INTO container_metrics(
	container_id, cpus, memory, network_ingress, network_egress
) VALUES (
	$1, $2, $3, $4, $5
);

-- name: UpdateContainer :exec
UPDATE containers
SET 
	image=$2,
	status=$3,
	name=$4,
	container_port=$5,
	public_port=$6,
	terminated_time=$7,
	created_time=$8
WHERE service_id=$1;


-- name: BatchUpdateStatusContainer :exec
UPDATE containers
SET 
	status=$2
WHERE service_id = ANY($1::varchar[]);

-- name: BatchUpdateStatusContainerLifecycle :exec
UPDATE container_lifecycles
SET 
	status=$2
WHERE container_id = ANY($1::UUID[]);



-- name: DeleteContainer :exec
DELETE FROM containers
WHERE service_id=$1;

-- name: InsertContainerLifecycle :one
INSERT INTO container_lifecycles(
	container_id, start_time, stop_time, status, replica
) VALUES (
	$1, $2, $3, $4, $5
) RETURNING *;

-- name: GetContainerLifecycle :one
SELECT id, container_id, start_time, stop_time, status, replica
FROM container_lifecycles
WHERE id=$1;


-- name: UpdateContainerLifecycle :exec
UPDATE container_lifecycles
SET 
	stop_time=$2,
	status=$3,
	replica=$4
WHERE id=$1;

-- name: UpdateContainerLifecycleStatus :exec
UPDATE container_lifecycles
SET 
	status=$2
WHERE id=$1;




-- name: UpdateContainerLifeCycleWithoutStopTime :exec
UPDATE container_lifecycles
SET 
	replica=$2
WHERE id=$1;




-- name: InsertIntoContainerMetrics :exec
INSERT INTO container_metrics(
	container_id, cpus, memory, network_ingress, network_egress
) VALUES (
	$1, $2, $3, $4, $5
) RETURNING *;

