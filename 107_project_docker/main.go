// Package main - Chapter 107: Docker - Containerization Engine
// Docker revolutionized software deployment by popularizing Linux containers.
// Written in Go, it demonstrates systems programming, process isolation,
// and client-server architecture patterns.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 107: DOCKER - CONTAINERIZATION ENGINE ===")

	// ============================================
	// WHAT IS DOCKER
	// ============================================
	fmt.Println(`
WHAT IS DOCKER
--------------
Docker is a platform for building, shipping, and running applications
in containers. Containers are lightweight, standalone, executable packages
that include everything needed to run software: code, runtime, system
tools, libraries, and settings.

Created by Solomon Hykes at dotCloud (2013), Docker made Linux containers
accessible to everyday developers.

KEY FACTS:
- Written in Go since the beginning
- Open-sourced in March 2013
- Uses Linux kernel features: namespaces, cgroups, union filesystems
- Client-server architecture (CLI + daemon)
- ~68,000+ GitHub stars
- Used by millions of developers worldwide`)

	// ============================================
	// WHY GO FOR DOCKER
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR DOCKER
-----------------------------
1. Static binaries    - Single binary deployment, no runtime deps
2. Cross-compilation  - Build for any platform from any platform
3. Syscall access     - Direct Linux kernel interaction via syscall pkg
4. Concurrency        - Goroutines for managing multiple containers
5. Performance        - Near-C speed for systems operations
6. Fast compilation   - Quick iteration during development

Docker was one of the first major projects to validate Go for
systems-level infrastructure software.`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
DOCKER ARCHITECTURE
-------------------

  +------------------+
  |   Docker CLI     |  (docker build, run, push...)
  |   (client)       |
  +--------+---------+
           |  REST API (unix socket / TCP)
  +--------v---------+
  |  Docker Daemon   |  (dockerd)
  |  (server)        |
  +--+-----+------+--+
     |     |      |
  +--v-+ +-v--+ +-v--------+
  |img | |net | |containers|
  |mgr | |mgr | |  (runc)  |
  +----+ +----+ +----------+
                      |
           +----------v-----------+
           |   Linux Kernel       |
           |  namespaces, cgroups |
           |  union fs (overlay2) |
           +----------------------+

COMPONENTS:
- Docker CLI:     User-facing command-line interface
- Docker Daemon:  Long-running background service (dockerd)
- containerd:     Container runtime manager
- runc:           OCI-compliant container runtime
- BuildKit:       Next-generation build engine`)

	// ============================================
	// KEY GO PATTERNS IN DOCKER
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN DOCKER
-------------------------

1. CLIENT-SERVER via REST API
   - Docker CLI sends HTTP requests to daemon
   - Unix socket (/var/run/docker.sock) by default
   - Can also use TCP for remote access

2. INTERFACE-BASED DESIGN
   - Backend interface for different container runtimes
   - Driver interfaces for storage, networking, logging

3. PLUGIN ARCHITECTURE
   - Storage drivers: overlay2, aufs, btrfs, zfs
   - Network drivers: bridge, host, overlay, macvlan
   - Log drivers: json-file, syslog, journald, fluentd

4. BUILDER PATTERN
   - Dockerfile parsed into build instructions
   - Each instruction creates a layer
   - Layer caching for fast rebuilds

5. GOROUTINES FOR CONTAINER MANAGEMENT
   - Each container lifecycle managed concurrently
   - Event streaming uses goroutines + channels
   - Health checks run as background goroutines`)

	// ============================================
	// LINUX KERNEL FEATURES
	// ============================================
	fmt.Println(`
LINUX KERNEL FEATURES USED BY DOCKER
-------------------------------------

NAMESPACES (isolation):
- PID   - Process ID isolation (container sees its own PID 1)
- NET   - Network stack isolation (virtual interfaces)
- MNT   - Mount point isolation (container filesystem)
- UTS   - Hostname isolation
- IPC   - Inter-process communication isolation
- USER  - User/group ID mapping

CGROUPS (resource limits):
- CPU shares and quotas
- Memory limits and swap
- Block I/O bandwidth
- Device access control

UNION FILESYSTEMS:
- overlay2 (default) - Efficient layered filesystem
- Each Docker image layer is read-only
- Container gets a thin read-write layer on top
- Copy-on-write for modified files`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN DOCKER
---------------------------------

1. LAYER CACHING
   - Each Dockerfile instruction creates a cached layer
   - Unchanged layers are reused across builds
   - Order instructions from least to most changing

2. COPY-ON-WRITE
   - Containers share base image layers (read-only)
   - Only modified files are copied to writable layer
   - Huge savings in disk space and startup time

3. CONTENT-ADDRESSABLE STORAGE
   - Images identified by SHA256 digest
   - Deduplication of identical layers
   - Efficient registry pulls (only missing layers)

4. CONNECTION MULTIPLEXING
   - Single socket handles multiple container streams
   - Stdout/stderr multiplexed over one connection
   - Efficient attach/exec operations

5. BUILDKIT OPTIMIZATIONS
   - Parallel build stages
   - Better caching with mount caches
   - Reduced context transfer`)

	// ============================================
	// GO PHILOSOPHY IN DOCKER
	// ============================================
	fmt.Println(`
HOW DOCKER EXPRESSES GO PHILOSOPHY
------------------------------------

"Do less, enable more"
  Docker provides simple primitives (build, run, push)
  that enable complex deployment workflows.

"A little copying is better than a little dependency"
  Docker vendored dependencies early on, ensuring
  reproducible builds before Go modules existed.

"Clear is better than clever"
  docker run, docker build, docker push - the CLI is
  intentionally simple and discoverable.

"Don't communicate by sharing memory; share memory
 by communicating"
  Container isolation embodies this at the OS level.
  Containers communicate via well-defined network APIs.`)

	// ============================================
	// SIMPLIFIED DOCKER-LIKE SYSTEM
	// ============================================
	fmt.Println("\n=== SIMPLIFIED DOCKER-LIKE CONTAINER MANAGER ===")

	manager := NewContainerManager()

	fmt.Println("\n--- Creating Images ---")
	manager.BuildImage("myapp:v1", []string{
		"FROM golang:1.22",
		"COPY . /app",
		"RUN go build -o /app/server",
		"CMD /app/server",
	})
	manager.BuildImage("redis:7", []string{
		"FROM alpine:3.19",
		"RUN apk add redis",
		"EXPOSE 6379",
		"CMD redis-server",
	})

	fmt.Println("\n--- Listing Images ---")
	manager.ListImages()

	fmt.Println("\n--- Running Containers ---")
	c1 := manager.RunContainer("myapp:v1", ContainerConfig{
		Name:   "web-1",
		Ports:  map[string]string{"8080": "80"},
		Env:    map[string]string{"ENV": "production"},
		Memory: "512m",
		CPU:    "0.5",
	})
	c2 := manager.RunContainer("redis:7", ContainerConfig{
		Name:   "cache-1",
		Ports:  map[string]string{"6379": "6379"},
		Memory: "256m",
		CPU:    "0.25",
	})

	fmt.Println("\n--- Container Status ---")
	manager.ListContainers()

	fmt.Println("\n--- Executing Command in Container ---")
	manager.ExecInContainer(c1, "ls /app")
	manager.ExecInContainer(c1, "cat /etc/hostname")

	fmt.Println("\n--- Container Logs ---")
	manager.GetLogs(c1, 5)
	manager.GetLogs(c2, 3)

	fmt.Println("\n--- Health Checks ---")
	time.Sleep(100 * time.Millisecond)
	manager.CheckHealth()

	fmt.Println("\n--- Stopping Containers ---")
	manager.StopContainer(c1)
	manager.StopContainer(c2)

	fmt.Println("\n--- Final Status ---")
	manager.ListContainers()

	// ============================================
	// DOCKER API SIMULATION
	// ============================================
	fmt.Println("\n=== DOCKER REST API SIMULATION ===")
	runDockerAPIDemo()

	fmt.Println(`
SUMMARY
-------
Docker transformed software deployment by making containers accessible.
Its Go codebase demonstrates:
- Systems programming with kernel APIs
- Clean client-server architecture
- Interface-driven extensibility
- Goroutine-based concurrent management
- Performance through caching and copy-on-write

Docker proved that Go is ideal for infrastructure tooling.`)
}

// ============================================
// CONTAINER TYPES
// ============================================

type ContainerState string

const (
	StateCreated  ContainerState = "created"
	StateRunning  ContainerState = "running"
	StateStopped  ContainerState = "stopped"
	StateExited   ContainerState = "exited"
)

type Container struct {
	ID        string
	Name      string
	Image     string
	State     ContainerState
	Config    ContainerConfig
	CreatedAt time.Time
	StartedAt time.Time
	Logs      []LogEntry
	PID       int
	mu        sync.Mutex
}

type ContainerConfig struct {
	Name   string
	Ports  map[string]string
	Env    map[string]string
	Memory string
	CPU    string
}

type Image struct {
	Name      string
	Tag       string
	Layers    []string
	Size      int64
	CreatedAt time.Time
	Digest    string
}

type LogEntry struct {
	Timestamp time.Time
	Stream    string
	Message   string
}

// ============================================
// CONTAINER MANAGER
// ============================================

type ContainerManager struct {
	containers map[string]*Container
	images     map[string]*Image
	nextID     int
	mu         sync.RWMutex
}

func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		containers: make(map[string]*Container),
		images:     make(map[string]*Image),
		nextID:     1000,
	}
}

func (cm *ContainerManager) BuildImage(nameTag string, dockerfile []string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	parts := strings.SplitN(nameTag, ":", 2)
	name := parts[0]
	tag := "latest"
	if len(parts) > 1 {
		tag = parts[1]
	}

	var totalSize int64
	fmt.Printf("Building image %s...\n", nameTag)
	for i, instruction := range dockerfile {
		layer := fmt.Sprintf("sha256:%x", time.Now().UnixNano()+int64(i))
		fmt.Printf("  Step %d/%d: %s\n", i+1, len(dockerfile), instruction)
		fmt.Printf("   ---> layer %s\n", layer[:20])
		totalSize += int64(1024 * (i + 1))
	}

	digest := fmt.Sprintf("sha256:%x", time.Now().UnixNano())
	img := &Image{
		Name:      name,
		Tag:       tag,
		Layers:    dockerfile,
		Size:      totalSize,
		CreatedAt: time.Now(),
		Digest:    digest[:30],
	}
	cm.images[nameTag] = img
	fmt.Printf("Successfully built %s (digest: %s)\n", nameTag, digest[:20])
}

func (cm *ContainerManager) ListImages() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	fmt.Printf("%-20s %-10s %-15s %s\n", "REPOSITORY", "TAG", "SIZE", "CREATED")
	for _, img := range cm.images {
		fmt.Printf("%-20s %-10s %-15s %s\n",
			img.Name, img.Tag,
			formatBytes(img.Size),
			img.CreatedAt.Format("15:04:05"))
	}
}

func (cm *ContainerManager) RunContainer(image string, config ContainerConfig) string {
	cm.mu.Lock()

	if _, exists := cm.images[image]; !exists {
		fmt.Printf("Error: image %s not found\n", image)
		cm.mu.Unlock()
		return ""
	}

	cm.nextID++
	id := fmt.Sprintf("%012x", cm.nextID)

	name := config.Name
	if name == "" {
		name = fmt.Sprintf("container_%s", id[:8])
	}

	container := &Container{
		ID:        id,
		Name:      name,
		Image:     image,
		State:     StateRunning,
		Config:    config,
		CreatedAt: time.Now(),
		StartedAt: time.Now(),
		PID:       cm.nextID + 30000,
	}

	cm.containers[id] = container
	cm.mu.Unlock()

	go cm.generateLogs(container)
	go cm.runHealthCheck(container)

	fmt.Printf("Container %s (%s) started from %s\n", name, id[:12], image)
	for host, cont := range config.Ports {
		fmt.Printf("  Port mapping: %s -> %s\n", host, cont)
	}
	for k, v := range config.Env {
		fmt.Printf("  Env: %s=%s\n", k, v)
	}
	if config.Memory != "" {
		fmt.Printf("  Memory limit: %s\n", config.Memory)
	}
	if config.CPU != "" {
		fmt.Printf("  CPU limit: %s\n", config.CPU)
	}

	return id
}

func (cm *ContainerManager) generateLogs(c *Container) {
	messages := []string{
		"Starting application...",
		"Loading configuration...",
		"Listening on port 80",
		"Ready to accept connections",
		"Health check passed",
	}

	for _, msg := range messages {
		c.mu.Lock()
		if c.State != StateRunning {
			c.mu.Unlock()
			return
		}
		c.Logs = append(c.Logs, LogEntry{
			Timestamp: time.Now(),
			Stream:    "stdout",
			Message:   msg,
		})
		c.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

func (cm *ContainerManager) runHealthCheck(c *Container) {
	time.Sleep(50 * time.Millisecond)
	c.mu.Lock()
	if c.State == StateRunning {
		c.Logs = append(c.Logs, LogEntry{
			Timestamp: time.Now(),
			Stream:    "stdout",
			Message:   "Health: OK",
		})
	}
	c.mu.Unlock()
}

func (cm *ContainerManager) ListContainers() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	fmt.Printf("%-14s %-12s %-15s %-10s %s\n",
		"CONTAINER ID", "NAME", "IMAGE", "STATUS", "PORTS")
	for _, c := range cm.containers {
		ports := ""
		for h, p := range c.Config.Ports {
			if ports != "" {
				ports += ", "
			}
			ports += fmt.Sprintf("%s->%s", h, p)
		}
		fmt.Printf("%-14s %-12s %-15s %-10s %s\n",
			c.ID[:12], c.Name, c.Image, string(c.State), ports)
	}
}

func (cm *ContainerManager) ExecInContainer(id, command string) {
	cm.mu.RLock()
	c, exists := cm.containers[id]
	cm.mu.RUnlock()

	if !exists {
		fmt.Printf("Error: container %s not found\n", id)
		return
	}

	c.mu.Lock()
	if c.State != StateRunning {
		fmt.Printf("Error: container %s is not running\n", c.Name)
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	fmt.Printf("Exec in %s: %s\n", c.Name, command)
	switch {
	case strings.Contains(command, "ls"):
		fmt.Println("  server  config.yaml  go.mod  go.sum")
	case strings.Contains(command, "hostname"):
		fmt.Printf("  %s\n", c.ID[:12])
	default:
		fmt.Printf("  (executed: %s)\n", command)
	}
}

func (cm *ContainerManager) GetLogs(id string, lines int) {
	cm.mu.RLock()
	c, exists := cm.containers[id]
	cm.mu.RUnlock()

	if !exists {
		fmt.Printf("Error: container %s not found\n", id)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Printf("--- Logs for %s ---\n", c.Name)
	start := 0
	if len(c.Logs) > lines {
		start = len(c.Logs) - lines
	}
	for _, entry := range c.Logs[start:] {
		fmt.Printf("  %s [%s] %s\n",
			entry.Timestamp.Format("15:04:05.000"),
			entry.Stream, entry.Message)
	}
}

func (cm *ContainerManager) CheckHealth() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	fmt.Println("Health Check Results:")
	for _, c := range cm.containers {
		c.mu.Lock()
		status := "healthy"
		if c.State != StateRunning {
			status = "unhealthy"
		}
		fmt.Printf("  %s (%s): %s\n", c.Name, c.Image, status)
		c.mu.Unlock()
	}
}

func (cm *ContainerManager) StopContainer(id string) {
	cm.mu.RLock()
	c, exists := cm.containers[id]
	cm.mu.RUnlock()

	if !exists {
		fmt.Printf("Error: container %s not found\n", id)
		return
	}

	c.mu.Lock()
	c.State = StateStopped
	c.Logs = append(c.Logs, LogEntry{
		Timestamp: time.Now(),
		Stream:    "stdout",
		Message:   "Received SIGTERM, shutting down gracefully...",
	})
	c.mu.Unlock()

	fmt.Printf("Container %s stopped\n", c.Name)
}

func formatBytes(b int64) string {
	if b < 1024 {
		return fmt.Sprintf("%d B", b)
	}
	return fmt.Sprintf("%.1f KB", float64(b)/1024)
}

// ============================================
// DOCKER REST API SIMULATION
// ============================================

func runDockerAPIDemo() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/containers/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"Id":"abc123","Image":"myapp:v1","State":"running"}]`)
	})

	mux.HandleFunc("/v1/images/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"Repository":"myapp","Tag":"v1","Size":52428800}]`)
	})

	mux.HandleFunc("/v1/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"Containers":2,"Images":5,"Driver":"overlay2"}`)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Docker API endpoints (simulated):")
	fmt.Println("  GET /v1/containers/json - List containers")
	fmt.Println("  GET /v1/images/json     - List images")
	fmt.Println("  GET /v1/info            - System information")
	fmt.Println("  POST /v1/containers/create - Create container")
	fmt.Println("  POST /v1/containers/{id}/start - Start container")
	fmt.Println("  POST /v1/containers/{id}/stop  - Stop container")
	fmt.Println("  DELETE /v1/containers/{id}      - Remove container")

	_ = server
	_ = os.Stdout
}
