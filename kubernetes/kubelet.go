// kubelet.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

// Kubelet represents a node agent.
type Kubelet struct {
	nodeName        string
	port            int
	pods            map[string]*Pod // Pods "running" on this node
	mu              sync.RWMutex
	availableMemory int
}

// NewKubelet creates and starts a new Kubelet.
func NewKubelet(name string, port int) *Kubelet {
	k := &Kubelet{
		nodeName:        name,
		port:            port,
		pods:            make(map[string]*Pod),
		availableMemory: 2048, // Start with 2GB memory
	}

	// Start the Kubelet's own API server in a goroutine.
	go k.startApiServer()

	return k
}

// startApiServer runs the HTTP server for this Kubelet.
func (k *Kubelet) startApiServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", k.handleStatus)
	// In a real system, you'd have an endpoint to receive and run pods.
	// mux.HandleFunc("/pods", k.handleAddPod)

	addr := fmt.Sprintf(":%d", k.port)
	log.Printf("Kubelet '%s' listening on %s\n", k.nodeName, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Kubelet '%s' failed to start: %v", k.nodeName, err)
	}
}

// handleStatus reports the current status of the node.
func (k *Kubelet) handleStatus(w http.ResponseWriter, r *http.Request) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	status := NodeStatus{
		NodeName:         k.nodeName,
		AvailableMemory:  k.availableMemory - (len(k.pods) * 100), // Simulate memory usage
		RunningPodsCount: len(k.pods),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// AddPod simulates scheduling a pod to this Kubelet.
// This would be called by the server in a real implementation.
func (k *Kubelet) AddPod(pod *Pod) {
	k.mu.Lock()
	defer k.mu.Unlock()

	log.Printf("Kubelet '%s' is now running pod '%s'\n", k.nodeName, pod.Name)
	k.pods[pod.ID] = pod
	// Simulate memory consumption
	k.availableMemory -= rand.Intn(150) + 50
}
