// Package main - Chapter 117: Consul - Service Mesh and Discovery
// Consul by HashiCorp provides service discovery, health checking,
// KV store, and service mesh. Written in Go, it demonstrates gossip
// protocols, consensus algorithms, and distributed service networking.
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 117: CONSUL - SERVICE MESH AND DISCOVERY ===")

	// ============================================
	// WHAT IS CONSUL
	// ============================================
	fmt.Println(`
WHAT IS CONSUL
--------------
Consul is a multi-cloud service networking platform that provides
service discovery, service mesh, and configuration management.
It enables services to discover and securely communicate with each other.

Created by HashiCorp in 2014, Consul combines several distributed
systems primitives into one tool.

KEY FACTS:
- Written entirely in Go
- ~28,000+ GitHub stars
- Four key features:
  1. Service Discovery  - Register and find services
  2. Health Checking    - Monitor service health
  3. KV Store           - Distributed configuration
  4. Service Mesh       - Secure service-to-service communication
- Uses Raft for consensus (servers)
- Uses Serf/Gossip for membership (agents)
- DNS and HTTP API interfaces
- Multi-datacenter support built-in`)

	// ============================================
	// WHY GO FOR CONSUL
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR CONSUL
------------------------------
1. Networking         - Excellent UDP/TCP support for gossip
2. Concurrency        - Goroutines for health checks and watches
3. Static binary      - Single binary for agent and server
4. Performance        - Low-latency service discovery
5. Cross-platform     - Runs everywhere
6. HashiCorp stack    - Consistent tooling with Terraform, Vault, Nomad`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
CONSUL ARCHITECTURE
-------------------

  DATACENTER 1                    DATACENTER 2
  +---------------------------+   +---------------------------+
  |  +-----+  +-----+        |   |        +-----+  +-----+  |
  |  |Srv 1|  |Srv 2|        |   |        |Srv 1|  |Srv 2|  |
  |  |LEADER| |FOLLWR|       |   |        |LEADER| |FOLLWR|  |
  |  +--+--+  +--+--+        |   |        +--+--+  +--+--+  |
  |     |        |   +-----+ | WAN|+-----+   |        |     |
  |     +--------+-->|Srv 3|<----->|Srv 3|<--+--------+     |
  |     Raft     |   |FOLLWR| |   |FOLLWR|   |  Raft        |
  |              |   +-----+ |   +-----+  |              |
  |  +-----------+---------+ |   +--------+----------+   |
  |  |  Gossip Pool (LAN)  | |   | Gossip Pool (LAN) |   |
  |  +-+----+----+----+---+  |   +-+----+----+----+--+   |
  |    |    |    |    |       |     |    |    |    |       |
  |  +-v-++-v-++-v-++-v-+    |   +-v-++-v-++-v-++-v-+    |
  |  |Agt||Agt||Agt||Agt|    |   |Agt||Agt||Agt||Agt|    |
  |  +---++---++---++---+    |   +---++---++---++---+    |
  +---------------------------+   +---------------------------+

LAYERS:
- Server Agents:  Run Raft consensus, store state
- Client Agents:  Forward requests, run health checks
- Gossip (LAN):   Membership and failure detection within DC
- Gossip (WAN):   Cross-datacenter server communication
- Raft:           Consensus for consistent state`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN CONSUL
-------------------------

1. GOSSIP PROTOCOL (Serf/SWIM)
   - Membership management via UDP gossip
   - Failure detection through health probes
   - Event dissemination: push-pull + piggyback
   - Scalable to thousands of nodes
   - Converges in O(log N) rounds

2. RAFT CONSENSUS (Server-side)
   - Strong consistency for KV store and catalog
   - Leader election and log replication
   - Snapshot and restore for state management
   - Built on hashicorp/raft library

3. BLOCKING QUERIES (Long Polling)
   - HTTP API supports ?wait=60s&index=N
   - Goroutine blocks until change or timeout
   - Efficient change notification without polling
   - Index-based consistency tracking

4. SERVICE MESH (Connect)
   - Sidecar proxies for service-to-service TLS
   - Intention-based access control
   - Automatic certificate rotation
   - Built-in Envoy integration

5. HEALTH CHECK GOROUTINES
   - Each registered health check runs in its own goroutine
   - HTTP, TCP, gRPC, script, and TTL check types
   - Configurable intervals and timeouts
   - Results stored locally, propagated via gossip`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN CONSUL
----------------------------------

1. GOSSIP PROTOCOL EFFICIENCY
   - O(log N) convergence for membership changes
   - Piggybacking state on health probes
   - Bounded gossip queue prevents flooding
   - UDP for minimal overhead

2. PREPARED QUERIES
   - Pre-compiled service queries for fast lookup
   - Supports failover across datacenters
   - Near-node routing for locality
   - Cached query results

3. ANTI-ENTROPY SYNC
   - Periodic full state sync between agent and servers
   - Catches any missed gossip updates
   - Runs in background goroutine
   - Interval-based with jitter

4. CONNECTION POOLING
   - RPC connection pool between agents and servers
   - Multiplexed yamux connections
   - Reduces TCP handshake overhead

5. BLOOM FILTER FOR KV WATCHES
   - Efficient change detection for KV watches
   - Reduces unnecessary notifications
   - Memory-efficient representation`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
HOW CONSUL EXPRESSES GO PHILOSOPHY
------------------------------------

"Don't communicate by sharing memory; share memory
 by communicating"
  Gossip protocol IS communication-based sharing.
  Nodes exchange state through messages, not shared memory.

"Errors are values"
  Health check results are first-class values.
  Service states (passing, warning, critical) drive
  routing decisions.

"Make the zero value useful"
  Default Consul agent joins cluster and begins
  health checking with minimal configuration.

"Composition over inheritance"
  Consul composes service discovery + health checking +
  KV store + service mesh. Each works independently.`)

	// ============================================
	// SIMPLIFIED CONSUL DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED CONSUL-LIKE SERVICE DISCOVERY ===")

	consul := NewConsul()

	fmt.Println("\n--- Starting Cluster ---")
	consul.AddServer("server-1", true)
	consul.AddServer("server-2", false)
	consul.AddServer("server-3", false)

	fmt.Println("\n--- Adding Client Agents ---")
	consul.AddAgent("agent-web-1")
	consul.AddAgent("agent-web-2")
	consul.AddAgent("agent-api-1")
	consul.AddAgent("agent-db-1")

	fmt.Println("\n--- Registering Services ---")
	consul.RegisterService(ConsulService{
		Name:    "web",
		ID:      "web-1",
		Address: "10.0.0.1",
		Port:    8080,
		Tags:    []string{"production", "v2.1"},
		Agent:   "agent-web-1",
		Check: HealthCheck{
			Type:     "http",
			Endpoint: "http://10.0.0.1:8080/health",
			Interval: 10 * time.Second,
		},
	})
	consul.RegisterService(ConsulService{
		Name:    "web",
		ID:      "web-2",
		Address: "10.0.0.2",
		Port:    8080,
		Tags:    []string{"production", "v2.1"},
		Agent:   "agent-web-2",
		Check: HealthCheck{
			Type:     "http",
			Endpoint: "http://10.0.0.2:8080/health",
			Interval: 10 * time.Second,
		},
	})
	consul.RegisterService(ConsulService{
		Name:    "api",
		ID:      "api-1",
		Address: "10.0.1.1",
		Port:    9090,
		Tags:    []string{"production", "v3.0"},
		Agent:   "agent-api-1",
		Check: HealthCheck{
			Type:     "http",
			Endpoint: "http://10.0.1.1:9090/health",
			Interval: 10 * time.Second,
		},
	})
	consul.RegisterService(ConsulService{
		Name:    "postgres",
		ID:      "db-1",
		Address: "10.0.2.1",
		Port:    5432,
		Tags:    []string{"production", "primary"},
		Agent:   "agent-db-1",
		Check: HealthCheck{
			Type:     "tcp",
			Endpoint: "10.0.2.1:5432",
			Interval: 15 * time.Second,
		},
	})

	fmt.Println("\n--- Service Discovery ---")
	consul.DiscoverService("web")
	consul.DiscoverService("api")
	consul.DiscoverService("postgres")
	consul.DiscoverService("nonexistent")

	fmt.Println("\n--- DNS Interface ---")
	consul.DNSLookup("web.service.consul")
	consul.DNSLookup("api.service.consul")
	consul.DNSLookup("db-1.node.consul")

	fmt.Println("\n--- Health Checks ---")
	consul.RunHealthChecks()

	fmt.Println("\n--- KV Store ---")
	consul.KVPut("config/web/max_connections", "100")
	consul.KVPut("config/web/timeout", "30s")
	consul.KVPut("config/api/rate_limit", "1000")
	consul.KVPut("config/db/pool_size", "20")
	consul.KVGet("config/web/max_connections")
	consul.KVGet("config/web/timeout")
	consul.KVList("config/web/")
	consul.KVList("config/")

	fmt.Println("\n--- Gossip Simulation ---")
	consul.SimulateGossip()

	fmt.Println("\n--- Service Mesh (Connect) ---")
	consul.SimulateConnect("web-1", "api-1")

	fmt.Println("\n--- Cluster Status ---")
	consul.PrintStatus()

	fmt.Println("\n--- HTTP API ---")
	runConsulAPIDemo()

	fmt.Println(`
SUMMARY
-------
Consul unifies service discovery and networking.
Its Go codebase demonstrates:
- Gossip protocol for scalable membership (Serf/SWIM)
- Raft consensus for consistent state
- Multi-datacenter architecture
- Health checking with multiple check types
- Service mesh with automatic mTLS

Consul showed how Go can implement complex distributed protocols.`)
}

// ============================================
// TYPES
// ============================================

type NodeRole string

const (
	RoleServer NodeRole = "server"
	RoleClient NodeRole = "client"
)

type ServiceStatus string

const (
	StatusPassing  ServiceStatus = "passing"
	StatusWarning  ServiceStatus = "warning"
	StatusCritical ServiceStatus = "critical"
)

type ConsulNode struct {
	Name     string
	Role     NodeRole
	IsLeader bool
	Address  string
}

type ConsulService struct {
	Name    string
	ID      string
	Address string
	Port    int
	Tags    []string
	Agent   string
	Status  ServiceStatus
	Check   HealthCheck
}

type HealthCheck struct {
	Type     string
	Endpoint string
	Interval time.Duration
}

// ============================================
// CONSUL CLUSTER
// ============================================

type Consul struct {
	servers  []*ConsulNode
	agents   []*ConsulNode
	services map[string][]*ConsulService
	kv       map[string]string
	mu       sync.RWMutex
}

func NewConsul() *Consul {
	return &Consul{
		services: make(map[string][]*ConsulService),
		kv:       make(map[string]string),
	}
}

func (c *Consul) AddServer(name string, leader bool) {
	node := &ConsulNode{
		Name:     name,
		Role:     RoleServer,
		IsLeader: leader,
		Address:  fmt.Sprintf("10.10.0.%d", len(c.servers)+1),
	}
	c.servers = append(c.servers, node)
	role := "follower"
	if leader {
		role = "leader"
	}
	fmt.Printf("  Server %s started (%s, %s)\n", name, node.Address, role)
}

func (c *Consul) AddAgent(name string) {
	node := &ConsulNode{
		Name:    name,
		Role:    RoleClient,
		Address: fmt.Sprintf("10.10.1.%d", len(c.agents)+1),
	}
	c.agents = append(c.agents, node)
	fmt.Printf("  Agent %s joined cluster (%s)\n", name, node.Address)
}

func (c *Consul) RegisterService(svc ConsulService) {
	c.mu.Lock()
	defer c.mu.Unlock()

	svc.Status = StatusPassing
	c.services[svc.Name] = append(c.services[svc.Name], &svc)
	fmt.Printf("  Service: %s (id: %s, addr: %s:%d, tags: %v)\n",
		svc.Name, svc.ID, svc.Address, svc.Port, svc.Tags)
	fmt.Printf("    Health check: %s %s (every %s)\n",
		svc.Check.Type, svc.Check.Endpoint, svc.Check.Interval)
}

func (c *Consul) DiscoverService(name string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	instances := c.services[name]
	if len(instances) == 0 {
		fmt.Printf("\n  Discover %q: no instances found\n", name)
		return
	}

	fmt.Printf("\n  Discover %q: %d instance(s)\n", name, len(instances))
	for _, svc := range instances {
		fmt.Printf("    %s -> %s:%d [%s] tags:%v\n",
			svc.ID, svc.Address, svc.Port, string(svc.Status), svc.Tags)
	}
}

func (c *Consul) DNSLookup(query string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Printf("\n  DNS: %s\n", query)

	parts := strings.Split(query, ".")
	if len(parts) >= 3 && parts[1] == "service" {
		svcName := parts[0]
		instances := c.services[svcName]
		for _, svc := range instances {
			if svc.Status == StatusPassing {
				fmt.Printf("    A %s:%d\n", svc.Address, svc.Port)
			}
		}
		if len(instances) == 0 {
			fmt.Printf("    NXDOMAIN\n")
		}
	} else if len(parts) >= 3 && parts[1] == "node" {
		nodeName := parts[0]
		for _, a := range c.agents {
			if a.Name == nodeName || strings.HasPrefix(a.Name, "agent-"+nodeName) {
				fmt.Printf("    A %s\n", a.Address)
				return
			}
		}
		fmt.Printf("    NXDOMAIN\n")
	}
}

func (c *Consul) RunHealthChecks() {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("Running health checks:")
	for _, instances := range c.services {
		for _, svc := range instances {
			status := StatusPassing
			if rand.Intn(10) == 0 {
				status = StatusCritical
			} else if rand.Intn(5) == 0 {
				status = StatusWarning
			}
			svc.Status = status

			checkResult := "OK"
			switch status {
			case StatusWarning:
				checkResult = "WARN (slow response)"
			case StatusCritical:
				checkResult = "FAIL (connection refused)"
			}
			fmt.Printf("  %s (%s check %s): %s\n",
				svc.ID, svc.Check.Type, svc.Check.Endpoint, checkResult)
		}
	}
}

func (c *Consul) KVPut(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.kv[key] = value
	fmt.Printf("  KV PUT %s = %s\n", key, value)
}

func (c *Consul) KVGet(key string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.kv[key]
	if !exists {
		fmt.Printf("  KV GET %s = (not found)\n", key)
		return
	}
	fmt.Printf("  KV GET %s = %s\n", key, value)
}

func (c *Consul) KVList(prefix string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Printf("  KV LIST %s\n", prefix)
	var keys []string
	for k := range c.kv {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("    %s = %s\n", k, c.kv[k])
	}
}

func (c *Consul) SimulateGossip() {
	allNodes := len(c.servers) + len(c.agents)
	rounds := 0
	informed := 1
	for informed < allNodes {
		informed *= 2
		if informed > allNodes {
			informed = allNodes
		}
		rounds++
		fmt.Printf("  Gossip round %d: %d/%d nodes informed\n", rounds, informed, allNodes)
	}
	fmt.Printf("  Convergence: %d rounds for %d nodes (O(log N))\n", rounds, allNodes)
}

func (c *Consul) SimulateConnect(from, to string) {
	fmt.Printf("  Connect: %s -> %s\n", from, to)
	fmt.Println("    1. Sidecar proxy initiated on both services")
	fmt.Println("    2. mTLS certificate obtained from Consul CA")
	fmt.Println("    3. Intention checked: ALLOW (default allow)")
	fmt.Println("    4. Encrypted connection established")
	fmt.Println("    5. Traffic flows through sidecar proxies")
}

func (c *Consul) PrintStatus() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Println("\nCluster Status:")
	fmt.Printf("  Servers: %d\n", len(c.servers))
	for _, s := range c.servers {
		role := "follower"
		if s.IsLeader {
			role = "leader"
		}
		fmt.Printf("    %s (%s) [%s]\n", s.Name, s.Address, role)
	}
	fmt.Printf("  Agents: %d\n", len(c.agents))
	for _, a := range c.agents {
		fmt.Printf("    %s (%s)\n", a.Name, a.Address)
	}

	totalServices := 0
	for name, instances := range c.services {
		passing := 0
		for _, svc := range instances {
			if svc.Status == StatusPassing {
				passing++
			}
		}
		fmt.Printf("  Service %s: %d/%d healthy\n", name, passing, len(instances))
		totalServices += len(instances)
	}
	fmt.Printf("  Total service instances: %d\n", totalServices)
	fmt.Printf("  KV entries: %d\n", len(c.kv))
}

func runConsulAPIDemo() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/catalog/services", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"web":["production"],"api":["production"]}`)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"Service":{"ID":"web-1","Address":"10.0.0.1"}}]`)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Consul HTTP API endpoints (simulated):")
	fmt.Println("  GET  /v1/catalog/services         - List services")
	fmt.Println("  GET  /v1/health/service/:name      - Healthy instances")
	fmt.Println("  PUT  /v1/agent/service/register    - Register service")
	fmt.Println("  PUT  /v1/agent/service/deregister  - Deregister service")
	fmt.Println("  GET  /v1/kv/:key                   - Get KV")
	fmt.Println("  PUT  /v1/kv/:key                   - Set KV")
	fmt.Println("  GET  /v1/agent/members             - List members")

	_ = server
	_ = os.Stdout
}
