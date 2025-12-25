package models

import "time"

// K8sContext represents a Kubernetes context
type K8sContext struct {
	Name      string `json:"name" yaml:"name"`
	Cluster   string `json:"cluster" yaml:"cluster"`
	User      string `json:"user" yaml:"user"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Current   bool   `json:"current" yaml:"current"`
}

// K8sResource represents a Kubernetes resource
type K8sResource struct {
	Kind       string            `json:"kind"`
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Age        string            `json:"age"`
	Labels     map[string]string `json:"labels,omitempty"`
	Replicas   string            `json:"replicas,omitempty"`
	Ready      string            `json:"ready,omitempty"`
	Restarts   int               `json:"restarts,omitempty"`
	Node       string            `json:"node,omitempty"`
	IP         string            `json:"ip,omitempty"`
}

// K8sPodStatus represents pod status details
type K8sPodStatus struct {
	Name          string            `json:"name"`
	Namespace     string            `json:"namespace"`
	Status        string            `json:"status"`
	Ready         string            `json:"ready"`
	Restarts      int               `json:"restarts"`
	Age           string            `json:"age"`
	Node          string            `json:"node"`
	IP            string            `json:"ip"`
	Containers    []K8sContainer    `json:"containers"`
	Conditions    []K8sCondition    `json:"conditions"`
	Events        []K8sEvent        `json:"events,omitempty"`
}

// K8sContainer represents container status
type K8sContainer struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Ready        bool   `json:"ready"`
	RestartCount int    `json:"restart_count"`
	State        string `json:"state"`
	Reason       string `json:"reason,omitempty"`
}

// K8sCondition represents a condition
type K8sCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// K8sEvent represents a Kubernetes event
type K8sEvent struct {
	Type      string    `json:"type"` // Normal, Warning
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// K8sClusterInfo represents cluster information
type K8sClusterInfo struct {
	Version       string            `json:"version"`
	Platform      string            `json:"platform"`
	NodeCount     int               `json:"node_count"`
	Nodes         []K8sNodeInfo     `json:"nodes"`
	NamespaceCount int              `json:"namespace_count"`
	PodCount      int               `json:"pod_count"`
	ServiceCount  int               `json:"service_count"`
}

// K8sNodeInfo represents node information
type K8sNodeInfo struct {
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Roles          []string `json:"roles"`
	Version        string   `json:"version"`
	InternalIP     string   `json:"internal_ip"`
	ExternalIP     string   `json:"external_ip,omitempty"`
	OS             string   `json:"os"`
	Architecture   string   `json:"architecture"`
	CPUCapacity    string   `json:"cpu_capacity"`
	MemoryCapacity string   `json:"memory_capacity"`
	PodCount       int      `json:"pod_count"`
}

// K8sResourceUsage represents resource usage
type K8sResourceUsage struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
	CPUUsage   string `json:"cpu_usage"`
	CPULimit   string `json:"cpu_limit,omitempty"`
	MemUsage   string `json:"memory_usage"`
	MemLimit   string `json:"memory_limit,omitempty"`
}
