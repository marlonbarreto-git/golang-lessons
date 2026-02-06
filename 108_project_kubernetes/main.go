// Package main - Chapter 108: Kubernetes - Container Orchestration Platform
// Kubernetes (K8s) is the industry standard for container orchestration.
// Written in Go, it demonstrates distributed systems design, controller
// patterns, and declarative API-driven architecture.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 108: KUBERNETES - CONTAINER ORCHESTRATION ===")

	// ============================================
	// WHAT IS KUBERNETES
	// ============================================
	fmt.Println(`
WHAT IS KUBERNETES
------------------
Kubernetes (K8s) is an open-source container orchestration platform
that automates deployment, scaling, and management of containerized
applications.

Originally developed at Google (based on Borg/Omega), donated to the
Cloud Native Computing Foundation (CNCF) in 2014.

KEY FACTS:
- Written entirely in Go
- ~109,000+ GitHub stars (most starred Go project)
- Created the cloud-native ecosystem
- Runs on every major cloud provider
- Manages containers at massive scale (thousands of nodes)
- Declarative configuration via YAML/JSON`)

	// ============================================
	// WHY GO FOR KUBERNETES
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR KUBERNETES
---------------------------------
1. Concurrency model  - Goroutines for watching/reconciling resources
2. Static binaries    - Easy to deploy across diverse environments
3. Performance        - Low latency for API server operations
4. Simplicity         - Large contributor base needs readable code
5. Networking         - Excellent stdlib for HTTP, gRPC, TCP
6. Google heritage    - Go was already popular within Google

Kubernetes validated Go as THE language for cloud infrastructure.`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
KUBERNETES ARCHITECTURE
-----------------------

  +---------------------------------------------------+
  |                 CONTROL PLANE                      |
  |  +------------+  +----------+  +--------------+   |
  |  | API Server |  |Scheduler |  |Controller Mgr|   |
  |  +-----+------+  +----+-----+  +------+-------+   |
  |        |              |               |            |
  |  +-----v--------------v---------------v--------+  |
  |  |                   etcd                       |  |
  |  |           (distributed key-value store)      |  |
  |  +---------------------------------------------+  |
  +---------------------------------------------------+
           |
  +--------v------------------------------------------+
  |                    WORKER NODES                    |
  |  +----------+  +----------+  +----------+         |
  |  | kubelet  |  | kubelet  |  | kubelet  |         |
  |  +----+-----+  +----+-----+  +----+-----+         |
  |       |             |             |                |
  |  +----v-----+  +----v-----+  +----v-----+         |
  |  | Pod  Pod |  | Pod  Pod |  | Pod  Pod |         |
  |  +----------+  +----------+  +----------+         |
  |                                                    |
  |  +--------------------------------------------+   |
  |  |          kube-proxy (networking)            |   |
  |  +--------------------------------------------+   |
  +---------------------------------------------------+

COMPONENTS:
- API Server:       RESTful API gateway, validates and stores state
- Scheduler:        Assigns pods to nodes based on constraints
- Controller Mgr:   Runs reconciliation loops (desired vs actual)
- etcd:             Consistent key-value store for cluster state
- kubelet:          Node agent, manages pods on each node
- kube-proxy:       Network proxy, implements Service abstraction`)

	// ============================================
	// KEY GO PATTERNS IN KUBERNETES
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN KUBERNETES
-----------------------------

1. CONTROLLER / RECONCILIATION PATTERN
   The heart of Kubernetes. Every controller:
   - Watches for changes to resources
   - Compares desired state vs actual state
   - Takes action to reconcile the difference
   - Repeats continuously

   for {
       desired := getDesiredState()
       actual  := getActualState()
       diff    := computeDiff(desired, actual)
       apply(diff)
   }

2. INFORMER / WATCH PATTERN
   - Client watches API server for resource changes
   - Uses HTTP long-polling (watch=true)
   - Local cache updated via informers
   - Reduces API server load dramatically

3. WORKQUEUE PATTERN
   - Rate-limited queue for processing events
   - Deduplication of items
   - Retry with exponential backoff
   - Ensures at-least-once processing

4. INTERFACE-BASED EXTENSIBILITY
   - Runtime interface: Docker, containerd, CRI-O
   - Network interface: CNI plugins
   - Storage interface: CSI plugins
   - Admission webhooks for custom validation

5. CODE GENERATION
   - client-go generated from API definitions
   - deepcopy methods auto-generated
   - Typed clients for every resource kind`)

	// ============================================
	// RESOURCE MODEL
	// ============================================
	fmt.Println(`
KUBERNETES RESOURCE MODEL
-------------------------

Every Kubernetes resource follows the same structure:

  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: my-app
    namespace: default
    labels:
      app: my-app
  spec:          # Desired state (user declares this)
    replicas: 3
    selector:
      matchLabels:
        app: my-app
    template:
      spec:
        containers:
        - name: app
          image: myapp:v1
  status:        # Actual state (system updates this)
    replicas: 3
    readyReplicas: 3

KEY RESOURCES:
- Pod:         Smallest deployable unit (1+ containers)
- Deployment:  Manages ReplicaSets for rolling updates
- Service:     Stable networking endpoint for pods
- ConfigMap:   Configuration data
- Secret:      Sensitive configuration data
- Ingress:     HTTP routing from external to services
- Namespace:   Resource isolation boundary`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN KUBERNETES
-------------------------------------

1. SHARED INFORMER CACHE
   - Each controller shares a local cache of resources
   - Watch events update cache incrementally
   - List operations hit cache, not API server
   - Dramatically reduces etcd load

2. PROTOBUF SERIALIZATION
   - Internal communication uses protobuf (not JSON)
   - 3-10x faster than JSON marshaling
   - Smaller wire format reduces network I/O

3. LEADER ELECTION
   - Only one controller instance is active at a time
   - Others are standby for fast failover
   - Uses etcd leases for distributed locking

4. RATE-LIMITED WORKQUEUES
   - Prevents thundering herd on mass updates
   - Exponential backoff on retries
   - Item deduplication reduces redundant work

5. API PAGINATION
   - Large resource lists are paginated
   - continue token for stateless pagination
   - Prevents OOM on clusters with millions of objects`)

	// ============================================
	// GO PHILOSOPHY IN K8S
	// ============================================
	fmt.Println(`
HOW KUBERNETES EXPRESSES GO PHILOSOPHY
---------------------------------------

"Make the zero value useful"
  Default Pod restartPolicy is Always. Default Service
  type is ClusterIP. Sensible defaults reduce config.

"Errors are values"
  Kubernetes controllers handle errors as data:
  requeue with backoff, update status conditions,
  emit events. Never panic on transient failures.

"Don't communicate by sharing memory; share memory
 by communicating"
  Controllers communicate via the API server (etcd).
  No direct inter-controller communication.
  The API server IS the shared communication channel.

"A little copying is better than a little dependency"
  client-go, apimachinery, and api are separate modules.
  Code generation preferred over runtime reflection.`)

	// ============================================
	// SIMPLIFIED KUBERNETES DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED KUBERNETES-LIKE SCHEDULER ===")

	cluster := NewCluster()

	fmt.Println("\n--- Adding Nodes ---")
	cluster.AddNode("node-1", NodeResources{CPU: 4000, Memory: 8192})
	cluster.AddNode("node-2", NodeResources{CPU: 4000, Memory: 8192})
	cluster.AddNode("node-3", NodeResources{CPU: 2000, Memory: 4096})

	fmt.Println("\n--- Deploying Application ---")
	cluster.ApplyDeployment(Deployment{
		Name:     "web-frontend",
		Replicas: 3,
		Image:    "nginx:1.25",
		CPU:      500,
		Memory:   256,
		Labels:   map[string]string{"app": "web", "tier": "frontend"},
	})

	cluster.ApplyDeployment(Deployment{
		Name:     "api-server",
		Replicas: 2,
		Image:    "myapi:v2",
		CPU:      1000,
		Memory:   512,
		Labels:   map[string]string{"app": "api", "tier": "backend"},
	})

	fmt.Println("\n--- Creating Services ---")
	cluster.CreateService(Service{
		Name:     "web-svc",
		Selector: map[string]string{"app": "web"},
		Port:     80,
		Type:     "ClusterIP",
	})
	cluster.CreateService(Service{
		Name:     "api-svc",
		Selector: map[string]string{"app": "api"},
		Port:     8080,
		Type:     "LoadBalancer",
	})

	fmt.Println("\n--- Cluster Status ---")
	cluster.PrintStatus()

	fmt.Println("\n--- Running Reconciliation Loop ---")
	cluster.Reconcile()

	fmt.Println("\n--- Scaling Deployment ---")
	cluster.ScaleDeployment("web-frontend", 5)

	fmt.Println("\n--- After Scaling ---")
	cluster.PrintStatus()

	fmt.Println("\n--- Simulating Node Failure ---")
	cluster.FailNode("node-2")

	fmt.Println("\n--- After Recovery ---")
	cluster.Reconcile()
	cluster.PrintStatus()

	fmt.Println(`
SUMMARY
-------
Kubernetes is the defining project of the cloud-native era.
Its Go codebase demonstrates:
- Controller/reconciliation pattern (watch + diff + act)
- Declarative API design (desired state vs actual state)
- Distributed systems primitives (leader election, watches)
- Interface-driven extensibility (CRI, CNI, CSI)
- Code generation for type safety at scale

Kubernetes made Go the default language for cloud infrastructure.`)
}

// ============================================
// TYPES
// ============================================

type PodPhase string

const (
	PodPending   PodPhase = "Pending"
	PodRunning   PodPhase = "Running"
	PodSucceeded PodPhase = "Succeeded"
	PodFailed    PodPhase = "Failed"
)

type NodeResources struct {
	CPU    int
	Memory int
}

type Node struct {
	Name      string
	Resources NodeResources
	Used      NodeResources
	Pods      []*Pod
	Ready     bool
}

type Pod struct {
	Name      string
	Namespace string
	Image     string
	Node      string
	Phase     PodPhase
	Labels    map[string]string
	CPU       int
	Memory    int
}

type Deployment struct {
	Name     string
	Replicas int
	Image    string
	CPU      int
	Memory   int
	Labels   map[string]string
}

type Service struct {
	Name      string
	Selector  map[string]string
	Port      int
	Type      string
	ClusterIP string
}

// ============================================
// CLUSTER
// ============================================

type Cluster struct {
	nodes       map[string]*Node
	pods        []*Pod
	deployments map[string]*Deployment
	services    map[string]*Service
	mu          sync.RWMutex
}

func NewCluster() *Cluster {
	return &Cluster{
		nodes:       make(map[string]*Node),
		deployments: make(map[string]*Deployment),
		services:    make(map[string]*Service),
	}
}

func (c *Cluster) AddNode(name string, resources NodeResources) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nodes[name] = &Node{
		Name:      name,
		Resources: resources,
		Ready:     true,
	}
	fmt.Printf("Node %s added (CPU: %dm, Memory: %dMi)\n",
		name, resources.CPU, resources.Memory)
}

func (c *Cluster) ApplyDeployment(d Deployment) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deployments[d.Name] = &d
	fmt.Printf("Deployment %s created (replicas: %d, image: %s)\n",
		d.Name, d.Replicas, d.Image)

	for i := 0; i < d.Replicas; i++ {
		pod := &Pod{
			Name:      fmt.Sprintf("%s-%s", d.Name, randomSuffix()),
			Namespace: "default",
			Image:     d.Image,
			Phase:     PodPending,
			Labels:    d.Labels,
			CPU:       d.CPU,
			Memory:    d.Memory,
		}
		c.pods = append(c.pods, pod)
		c.schedulePod(pod)
	}
}

func (c *Cluster) schedulePod(pod *Pod) {
	var bestNode *Node
	var bestScore int

	for _, node := range c.nodes {
		if !node.Ready {
			continue
		}
		availCPU := node.Resources.CPU - node.Used.CPU
		availMem := node.Resources.Memory - node.Used.Memory

		if availCPU >= pod.CPU && availMem >= pod.Memory {
			score := availCPU + availMem
			if bestNode == nil || score > bestScore {
				bestNode = node
				bestScore = score
			}
		}
	}

	if bestNode != nil {
		pod.Node = bestNode.Name
		pod.Phase = PodRunning
		bestNode.Pods = append(bestNode.Pods, pod)
		bestNode.Used.CPU += pod.CPU
		bestNode.Used.Memory += pod.Memory
		fmt.Printf("  Pod %s scheduled on %s\n", pod.Name, bestNode.Name)
	} else {
		fmt.Printf("  Pod %s PENDING - no suitable node found\n", pod.Name)
	}
}

func (c *Cluster) CreateService(svc Service) {
	c.mu.Lock()
	defer c.mu.Unlock()

	svc.ClusterIP = fmt.Sprintf("10.96.%d.%d", rand.Intn(255), rand.Intn(255)+1)
	c.services[svc.Name] = &svc

	endpoints := c.findEndpoints(svc.Selector)
	fmt.Printf("Service %s created (type: %s, clusterIP: %s, endpoints: %d)\n",
		svc.Name, svc.Type, svc.ClusterIP, len(endpoints))
}

func (c *Cluster) findEndpoints(selector map[string]string) []*Pod {
	var endpoints []*Pod
	for _, pod := range c.pods {
		if pod.Phase != PodRunning {
			continue
		}
		match := true
		for k, v := range selector {
			if pod.Labels[k] != v {
				match = false
				break
			}
		}
		if match {
			endpoints = append(endpoints, pod)
		}
	}
	return endpoints
}

func (c *Cluster) PrintStatus() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Println("\nNODES:")
	fmt.Printf("  %-10s %-8s %-15s %-15s %s\n",
		"NAME", "STATUS", "CPU (used/cap)", "MEM (used/cap)", "PODS")
	for _, node := range c.nodes {
		status := "Ready"
		if !node.Ready {
			status = "NotReady"
		}
		fmt.Printf("  %-10s %-8s %5d/%-9d %5d/%-9d %d\n",
			node.Name, status,
			node.Used.CPU, node.Resources.CPU,
			node.Used.Memory, node.Resources.Memory,
			len(node.Pods))
	}

	fmt.Println("\nPODS:")
	fmt.Printf("  %-30s %-10s %-10s %s\n", "NAME", "STATUS", "NODE", "IMAGE")
	for _, pod := range c.pods {
		fmt.Printf("  %-30s %-10s %-10s %s\n",
			pod.Name, string(pod.Phase), pod.Node, pod.Image)
	}

	fmt.Println("\nSERVICES:")
	fmt.Printf("  %-12s %-14s %-16s %s\n", "NAME", "TYPE", "CLUSTER-IP", "ENDPOINTS")
	for _, svc := range c.services {
		endpoints := c.findEndpoints(svc.Selector)
		var epNames []string
		for _, ep := range endpoints {
			epNames = append(epNames, ep.Name)
		}
		epStr := strings.Join(epNames, ", ")
		if len(epStr) > 40 {
			epStr = epStr[:37] + "..."
		}
		fmt.Printf("  %-12s %-14s %-16s %s\n",
			svc.Name, svc.Type, svc.ClusterIP, epStr)
	}
}

func (c *Cluster) Reconcile() {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("Running reconciliation loop...")

	for name, deploy := range c.deployments {
		var running int
		for _, pod := range c.pods {
			if pod.Phase == PodRunning && matchesDeployment(pod, deploy) {
				running++
			}
		}

		diff := deploy.Replicas - running
		if diff > 0 {
			fmt.Printf("  Deployment %s: need %d more replicas (have %d, want %d)\n",
				name, diff, running, deploy.Replicas)
			for i := 0; i < diff; i++ {
				pod := &Pod{
					Name:      fmt.Sprintf("%s-%s", name, randomSuffix()),
					Namespace: "default",
					Image:     deploy.Image,
					Phase:     PodPending,
					Labels:    deploy.Labels,
					CPU:       deploy.CPU,
					Memory:    deploy.Memory,
				}
				c.pods = append(c.pods, pod)
				c.schedulePod(pod)
			}
		} else if diff == 0 {
			fmt.Printf("  Deployment %s: OK (%d/%d replicas running)\n",
				name, running, deploy.Replicas)
		}
	}
}

func (c *Cluster) ScaleDeployment(name string, replicas int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	deploy, exists := c.deployments[name]
	if !exists {
		fmt.Printf("Deployment %s not found\n", name)
		return
	}

	old := deploy.Replicas
	deploy.Replicas = replicas
	fmt.Printf("Deployment %s scaled from %d to %d replicas\n", name, old, replicas)

	if replicas > old {
		for i := 0; i < replicas-old; i++ {
			pod := &Pod{
				Name:      fmt.Sprintf("%s-%s", name, randomSuffix()),
				Namespace: "default",
				Image:     deploy.Image,
				Phase:     PodPending,
				Labels:    deploy.Labels,
				CPU:       deploy.CPU,
				Memory:    deploy.Memory,
			}
			c.pods = append(c.pods, pod)
			c.schedulePod(pod)
		}
	}
}

func (c *Cluster) FailNode(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.nodes[name]
	if !exists {
		return
	}

	node.Ready = false
	fmt.Printf("Node %s marked NotReady - evicting %d pods\n", name, len(node.Pods))

	for _, pod := range node.Pods {
		pod.Phase = PodFailed
		pod.Node = ""
	}
	node.Pods = nil
	node.Used = NodeResources{}
}

func matchesDeployment(pod *Pod, deploy *Deployment) bool {
	return strings.HasPrefix(pod.Name, deploy.Name+"-")
}

func randomSuffix() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func init() {
	_ = os.Stdout
	_ = time.Now
}
