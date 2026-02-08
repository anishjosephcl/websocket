// main.go
package main

import (
	"log"
)

func main() {
	log.Println("--- Starting Miniature Kubernetes POC ---")

	// 1. Create the Control Plane Server.
	server := NewServer()

	// 2. Create our Kubelet "nodes".
	kubelet1 := NewKubelet("node-01", 8081)
	kubelet2 := NewKubelet("node-02", 8082)
	kubelet3 := NewKubelet("node-03", 8083)

	// 3. Register the Kubelets with the server so it knows who to poll.
	server.kubelets = append(server.kubelets, kubelet1, kubelet2, kubelet3)

	// 4. Start the main Control Plane API server. This will block.
	server.Start(":8090")
}
