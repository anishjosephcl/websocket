// types.go
package main

import "time"

// Pod represents the desired state of a workload.
type Pod struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	NodeName  string    `json:"nodeName,omitempty"` // The node this pod is assigned to
	CreatedAt time.Time `json:"createdAt"`
}

// NodeStatus represents the current state of a Kubelet/node.
type NodeStatus struct {
	NodeName         string `json:"nodeName"`
	AvailableMemory  int    `json:"availableMemory"` // in MB
	RunningPodsCount int    `json:"runningPodsCount"`
}
