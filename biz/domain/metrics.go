package domain 

type Metric struct {
	ContainerID string `json:"container_id"`
	CpuUsage            float32                `json:"cpu_usage"`
	MemoryUsage         float32                `json:"memory_usage"`
	NetworkIngressUsage float32                `json:"network_ingress_usage"`
	NetworkEgressUsage  float32                `json:"network_egress_usage"`
}




