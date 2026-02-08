// server.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Server represents our control plane.
type Server struct {
	pods     map[string]*Pod       // Acts as our 'etcd' for pods
	nodes    map[string]NodeStatus // A cache of the last known node status
	kubelets []*Kubelet            // Direct reference to Kubelets in our POC
	mu       sync.RWMutex
}

// NewServer creates and starts a new Server.
func NewServer() *Server {
	s := &Server{
		pods:     make(map[string]*Pod),
		nodes:    make(map[string]NodeStatus),
		kubelets: []*Kubelet{}, // Will be populated in main
	}

	go s.startPollingKubelets()
	return s
}

// startPollingKubelets periodically fetches status from all known Kubelets.
func (s *Server) startPollingKubelets() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		log.Println("Server: Polling Kubelets for status...")
		var wg sync.WaitGroup

		s.mu.RLock()
		kubeletPorts := []int{}
		for _, k := range s.kubelets {
			kubeletPorts = append(kubeletPorts, k.port)
		}
		s.mu.RUnlock()

		for _, port := range kubeletPorts {
			wg.Add(1)
			// Use goroutines to poll concurrently
			go func(port int) {
				defer wg.Done()
				url := fmt.Sprintf("http://localhost:%d/status", port)
				resp, err := http.Get(url)
				if err != nil {
					log.Printf("Server: Failed to get status from Kubelet on port %d: %v\n", port, err)
					return
				}
				defer resp.Body.Close()

				var status NodeStatus
				if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
					log.Printf("Server: Failed to decode status from Kubelet on port %d: %v\n", port, err)
					return
				}

				// Update the server's cache of node statuses
				s.mu.Lock()
				s.nodes[status.NodeName] = status
				s.mu.Unlock()
				log.Printf("Server: Received status from '%s': %+v\n", status.NodeName, status)
			}(port)
		}
		wg.Wait() // Wait for all polls in this cycle to complete
	}
}

// handleCreatePod is the API handler for creating and "scheduling" a pod.
func (s *Server) handleCreatePod(w http.ResponseWriter, r *http.Request) {
	var pod Pod
	if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	pod.ID = uuid.New().String()
	pod.CreatedAt = time.Now()

	// --- Scheduling Logic (Simplified) ---
	// Find a node to "schedule" this pod on. A real scheduler is much more complex.
	nodeToScheduleOn := s.findNodeForPod()
	if nodeToScheduleOn == nil {
		http.Error(w, "No available nodes to schedule pod", http.StatusInternalServerError)
		return
	}
	pod.NodeName = nodeToScheduleOn.nodeName

	// Store the pod in our "etcd" map.
	s.mu.Lock()
	s.pods[pod.ID] = &pod
	s.mu.Unlock()

	// "Tell" the Kubelet to run the pod.
	nodeToScheduleOn.AddPod(&pod)

	log.Printf("Server: Scheduled pod '%s' on node '%s'\n", pod.Name, pod.NodeName)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pod)
}

// findNodeForPod is a very simple scheduler that picks a node randomly.
func (s *Server) findNodeForPod() *Kubelet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.kubelets) == 0 {
		return nil
	}
	// Simple round-robin or random for this POC
	return s.kubelets[rand.Intn(len(s.kubelets))]
}

// handleListPods lists all pods known to the control plane.
func (s *Server) handleListPods(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	podList := make([]*Pod, 0, len(s.pods))
	for _, pod := range s.pods {
		podList = append(podList, pod)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(podList)
}

// handleListNodes lists the last known status of all nodes.
func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodeList := make([]NodeStatus, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodeList = append(nodeList, node)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodeList)
}

// Start runs the main HTTP server for the control plane.
func (s *Server) Start(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s.handleCreatePod(w, r)
		} else if r.Method == http.MethodGet {
			s.handleListPods(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/nodes", s.handleListNodes)

	log.Printf("Control Plane Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Control Plane Server failed to start: %v", err)
	}
}
