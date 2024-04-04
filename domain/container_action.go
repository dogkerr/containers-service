package domain

import (
	"time"

	"github.com/google/uuid"
)
type Action int

const (
	START Action = iota + 1
	STOP
)

func (s Status) ActionString() string {
	return [...]string{"START", "STOP"}[s-1]
}

type ContainerAction struct {
	ID uuid.UUID `json:"id"`
	ContainerId uuid.UUID `json:"container_id"`
	Timestamp time.Time `json:"timestamp"`
	Action Action `json:"action"`
}


