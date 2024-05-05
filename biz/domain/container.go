package domain

import (
	"time"
)

// type Status int

// const (
// 	RUN Status = iota + 1
// 	STOPPED
// )

// func (s Status) String() string {
// 	return [...]string{"RUN", "STOPPED"}[s-1]
// }

type ContainerStatus string

const (
	ContainerStatusRUN     ContainerStatus = "RUN"
	ContainerStatusSTOPPED ContainerStatus = "STOPPED"
)

type ContainerLifecycle struct {
	ID          string          `json:"id"`
	ContainerID string          `json:"containerId"`
	StartTime   time.Time       `json:"start_time"`
	StopTime    time.Time       `json:"stop_time"`
	CPUCore     float64         `json:"cpu_core"`
	MemCapacity float64         `json:"mem_capacity"`
	Replica     uint64          `json:"replica"`
	Status      ContainerStatus `json:"status"`
}

type Container struct {
	// ini cuma id row di table container
	ID                  string               `json:"id"`
	UserID              string               `json:"user_id"`
	Status              ContainerStatus      `json:"status"`
	Name                string               `json:"name"`
	ContainerPort       int                  `json:"container_port"`
	PublicPort          int                  `json:"public_port"`
	CreatedTime         time.Time            `json:"created_at"`
	TerminatedTime      time.Time            `json:"terminated_time"`
	ContainerLifecycles []ContainerLifecycle `json:"all_container_lifecycles"`
	// id dari containernya/servicenya
	ServiceID string `json:"serviceId"`

	/// field dibawah ini cuma dari docker engine && bukan dari db
	Labels      map[string]string `json:"labels"`
	Replica     uint64            `json:"replica"`
	Limit       Resource          `json:"limit"`
	Reservation Resource          `json:"reservation,omitempty"`
	Image       string            `json:"image"`
	Env         []string          `json:"env"`
	Endpoint    []Endpoint        `json:"endpoint"`
}
type Endpoint struct {
	TargetPort    uint32 `json:"target_port,required" vd:"$<65555 && $>0"`
	PublishedPort uint64 `json:"published_port,required" vd:"$<65555 && $>0"`
	Protocol      string `json:"protocol" default:"tcp" vd:"in($, 'tcp','udp','sctp')" `
}

// Resource
// @Description ini resource cpus & memory buat setiap container nya
type Resource struct {
	// cpu dalam milicpu (1000 cpus = 1 vcpu)
	CPUs int64 `json:"cpus" vd:"len($)<20000 && $>0"`
	// memory dalam satuan mb (1000mb = 1gb)
	Memory int64 `json:"memory" vd:"len($)<50000  && $>0"`
}

// type container struct {
// 	CreatedAt   time.Time         `json:"created_at"`
// 	ID          string            `json:"id"`
// 	Name        string            `json:"name"`

// }
