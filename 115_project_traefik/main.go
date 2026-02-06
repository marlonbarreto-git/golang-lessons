// Package main - Chapter 115: Traefik - Cloud-Native Edge Router
// Traefik is a modern HTTP reverse proxy and load balancer designed
// for microservices. Written in Go, it demonstrates dynamic configuration,
// service discovery integration, and middleware chain patterns.
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 115: TRAEFIK - CLOUD-NATIVE EDGE ROUTER ===")

	// ============================================
	// WHAT IS TRAEFIK
	// ============================================
	fmt.Println(`
WHAT IS TRAEFIK
---------------
Traefik (pronounced "traffic") is a modern reverse proxy and load
balancer that makes deploying microservices easy. It automatically
discovers services and configures itself dynamically.

Created by Emile Vauge (Traefik Labs) in 2015, Traefik integrates
natively with container orchestrators like Docker and Kubernetes.

KEY FACTS:
- Written entirely in Go
- ~52,000+ GitHub stars
- Auto-discovers services from Docker, K8s, Consul, etc.
- Automatic HTTPS via Let's Encrypt
- Built-in metrics, tracing, and access logs
- Middleware for auth, rate limiting, circuit breaking
- Dashboard for real-time monitoring
- Supports HTTP, TCP, UDP, gRPC`)

	// ============================================
	// WHY GO FOR TRAEFIK
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR TRAEFIK
-------------------------------
1. Performance       - High-throughput reverse proxy
2. Concurrency       - Handle thousands of connections
3. Docker integration - Go Docker client SDK
4. Static binary      - Easy container deployment
5. HTTP/2 native      - Built-in in net/http
6. Plugin system      - Go plugins via Yaegi interpreter`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
TRAEFIK ARCHITECTURE
--------------------

  +-----------------------------------------------------+
  |                    ENTRYPOINTS                       |
  |  :80 (HTTP)    :443 (HTTPS)    :8080 (Dashboard)    |
  +----------+----------+-----------+-------------------+
             |          |           |
  +----------v----------v-----------v-------------------+
  |                    ROUTERS                           |
  |  Rules: Host, Path, Headers, Query                  |
  |  TLS: automatic cert management                     |
  +------------------------+----------------------------+
                           |
  +------------------------v----------------------------+
  |                   MIDDLEWARES                        |
  |  [Auth] -> [RateLimit] -> [Headers] -> [Compress]   |
  +------------------------+----------------------------+
                           |
  +------------------------v----------------------------+
  |                    SERVICES                          |
  |  Load Balancing: Round Robin, Weighted, Sticky      |
  |  Health Checks: Active + Passive                    |
  +--+---------+---------+---------+--------------------+
     |         |         |         |
  +--v--+  +--v--+  +--v--+  +--v--+
  |svc-1|  |svc-2|  |svc-3|  |svc-4|
  +-----+  +-----+  +-----+  +-----+

PROVIDERS (Dynamic Configuration):
  Docker:      Labels on containers
  Kubernetes:  Ingress/IngressRoute CRDs
  Consul:      Service catalog
  File:        YAML/TOML configuration
  etcd:        Key-value store`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN TRAEFIK
--------------------------

1. PROVIDER PATTERN (Dynamic Configuration)
   - Each provider watches for changes
   - Goroutine per provider with watch loop
   - Providers emit configuration via channels
   - Aggregator merges configs from all providers

   type Provider interface {
       Provide(chan<- dynamic.Message) error
       Init() error
   }

2. MIDDLEWARE CHAIN
   - HTTP middleware as handler wrappers
   - Configurable per-router middleware stack
   - Built-in: auth, compress, headers, rate limit
   - Plugin middleware via Yaegi Go interpreter

3. ENTRYPOINT PATTERN
   - Each listening port is an entrypoint
   - Entrypoints accept connections as goroutines
   - TLS termination at entrypoint level
   - Protocol detection (HTTP, TCP, UDP)

4. ROUND-TRIPPER PATTERN
   - Custom http.RoundTripper for backend requests
   - Connection pooling per backend
   - Retry logic with configurable attempts
   - Circuit breaker integration

5. WATCHER PATTERN (for providers)
   - Long-polling Docker/K8s APIs
   - Debouncing rapid configuration changes
   - Atomic config swaps (no partial updates)
   - Graceful transition between configs`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN TRAEFIK
----------------------------------

1. CONNECTION POOLING
   - Persistent connections to backends
   - Configurable max connections per host
   - Idle connection cleanup

2. BUFFER POOLING
   - sync.Pool for request/response buffers
   - Reduces garbage collection pressure
   - Significant improvement under high load

3. CONFIGURATION HOT-RELOAD
   - Atomic configuration swap
   - No connection drops during reload
   - New connections use new config immediately

4. PASSIVE HEALTH CHECKS
   - Monitor response codes from real traffic
   - No extra health check requests needed
   - Fast failover on backend errors

5. HTTP/2 MULTIPLEXING
   - Multiple requests over single connection
   - Header compression (HPACK)
   - Server push support`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
HOW TRAEFIK EXPRESSES GO PHILOSOPHY
-------------------------------------

"Accept interfaces, return structs"
  Provider interface allows any service discovery backend.
  Middleware interface enables composable request processing.
  All return concrete configuration structs.

"Don't communicate by sharing memory; share memory
 by communicating"
  Providers communicate config changes via channels.
  No shared mutable state between providers.

"Make the zero value useful"
  Default load balancing is round-robin.
  Default health check is passive.
  Works with minimal configuration.

"Errors are values"
  Backend errors trigger circuit breaker.
  Retry middleware treats errors as retryable conditions.
  Health status derived from error patterns.`)

	// ============================================
	// SIMPLIFIED TRAEFIK DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED TRAEFIK-LIKE EDGE ROUTER ===")

	router := NewTraefikRouter()

	fmt.Println("\n--- Configuring Entrypoints ---")
	router.AddEntrypoint("web", 80, false)
	router.AddEntrypoint("websecure", 443, true)

	fmt.Println("\n--- Adding Services ---")
	router.AddService("web-app", []Backend{
		{Address: "10.0.0.1:8080", Weight: 3},
		{Address: "10.0.0.2:8080", Weight: 3},
		{Address: "10.0.0.3:8080", Weight: 1},
	})
	router.AddService("api", []Backend{
		{Address: "10.0.1.1:9090", Weight: 1},
		{Address: "10.0.1.2:9090", Weight: 1},
	})
	router.AddService("static-files", []Backend{
		{Address: "10.0.2.1:80", Weight: 1},
	})

	fmt.Println("\n--- Configuring Routes ---")
	router.AddRouterRule(RouterRule{
		Name:       "web-router",
		Entrypoint: "websecure",
		Rule:       "Host(`example.com`) && PathPrefix(`/`)",
		Service:    "web-app",
		Middlewares: []string{"rate-limit", "compress", "security-headers"},
		Priority:   1,
	})
	router.AddRouterRule(RouterRule{
		Name:       "api-router",
		Entrypoint: "websecure",
		Rule:       "Host(`api.example.com`) && PathPrefix(`/v1`)",
		Service:    "api",
		Middlewares: []string{"rate-limit", "auth", "cors"},
		Priority:   10,
	})
	router.AddRouterRule(RouterRule{
		Name:       "static-router",
		Entrypoint: "web",
		Rule:       "Host(`static.example.com`)",
		Service:    "static-files",
		Middlewares: []string{"compress", "cache-headers"},
		Priority:   5,
	})

	fmt.Println("\n--- Registering Middleware ---")
	router.RegisterMiddleware("rate-limit", "Rate limit: 100 req/s")
	router.RegisterMiddleware("compress", "Gzip compression")
	router.RegisterMiddleware("security-headers", "X-Frame-Options, CSP, HSTS")
	router.RegisterMiddleware("auth", "JWT Bearer token validation")
	router.RegisterMiddleware("cors", "CORS headers for API")
	router.RegisterMiddleware("cache-headers", "Cache-Control: public, max-age=86400")

	fmt.Println("\n--- Processing Requests ---")
	requests := []TraefikRequest{
		{Host: "example.com", Path: "/index.html", Method: "GET"},
		{Host: "api.example.com", Path: "/v1/users", Method: "GET"},
		{Host: "api.example.com", Path: "/v1/orders", Method: "POST"},
		{Host: "static.example.com", Path: "/css/style.css", Method: "GET"},
		{Host: "unknown.example.com", Path: "/", Method: "GET"},
	}

	for _, req := range requests {
		router.HandleRequest(req)
	}

	fmt.Println("\n--- Health Check Status ---")
	router.RunHealthChecks()

	fmt.Println("\n--- Load Balancer Statistics ---")
	router.PrintStats()

	fmt.Println("\n--- Dynamic Config Reload ---")
	router.SimulateDockerEvent("container_start", "new-api-instance", "10.0.1.3:9090")
	router.SimulateDockerEvent("container_stop", "old-api-instance", "10.0.1.1:9090")

	fmt.Println("\n--- Dashboard API ---")
	runTraefikDashboard()

	fmt.Println(`
SUMMARY
-------
Traefik simplified cloud-native edge routing.
Its Go codebase demonstrates:
- Dynamic configuration from service discovery
- Middleware chain for composable request processing
- Provider pattern for multi-source configuration
- Connection pooling and buffer reuse
- Graceful hot-reload without dropping connections

Traefik made reverse proxy configuration automatic.`)
}

// ============================================
// TYPES
// ============================================

type Backend struct {
	Address string
	Weight  int
	Healthy bool
	Requests int
}

type Entrypoint struct {
	Name string
	Port int
	TLS  bool
}

type RouterRule struct {
	Name        string
	Entrypoint  string
	Rule        string
	Service     string
	Middlewares  []string
	Priority    int
}

type TraefikRequest struct {
	Host   string
	Path   string
	Method string
}

type ServiceGroup struct {
	Name     string
	Backends []Backend
	NextIdx  int
}

type MiddlewareDef struct {
	Name        string
	Description string
}

// ============================================
// TRAEFIK ROUTER
// ============================================

type TraefikRouter struct {
	entrypoints map[string]*Entrypoint
	services    map[string]*ServiceGroup
	rules       []RouterRule
	middlewares  map[string]*MiddlewareDef
	stats       TraefikStats
	mu          sync.RWMutex
}

type TraefikStats struct {
	TotalRequests int
	ByService     map[string]int
	ByStatus      map[int]int
}

func NewTraefikRouter() *TraefikRouter {
	return &TraefikRouter{
		entrypoints: make(map[string]*Entrypoint),
		services:    make(map[string]*ServiceGroup),
		middlewares:  make(map[string]*MiddlewareDef),
		stats: TraefikStats{
			ByService: make(map[string]int),
			ByStatus:  make(map[int]int),
		},
	}
}

func (t *TraefikRouter) AddEntrypoint(name string, port int, tls bool) {
	t.entrypoints[name] = &Entrypoint{Name: name, Port: port, TLS: tls}
	proto := "HTTP"
	if tls {
		proto = "HTTPS"
	}
	fmt.Printf("  Entrypoint %q: :%d (%s)\n", name, port, proto)
}

func (t *TraefikRouter) AddService(name string, backends []Backend) {
	for i := range backends {
		backends[i].Healthy = true
	}
	t.services[name] = &ServiceGroup{
		Name:     name,
		Backends: backends,
	}
	var addrs []string
	for _, b := range backends {
		addrs = append(addrs, fmt.Sprintf("%s(w:%d)", b.Address, b.Weight))
	}
	fmt.Printf("  Service %q: [%s]\n", name, strings.Join(addrs, ", "))
}

func (t *TraefikRouter) AddRouterRule(rule RouterRule) {
	t.rules = append(t.rules, rule)
	fmt.Printf("  Router %q: %s -> %s (middlewares: %s)\n",
		rule.Name, rule.Rule, rule.Service, strings.Join(rule.Middlewares, " -> "))
}

func (t *TraefikRouter) RegisterMiddleware(name, desc string) {
	t.middlewares[name] = &MiddlewareDef{Name: name, Description: desc}
	fmt.Printf("  Middleware %q: %s\n", name, desc)
}

func (t *TraefikRouter) HandleRequest(req TraefikRequest) {
	t.mu.Lock()
	t.stats.TotalRequests++
	t.mu.Unlock()

	fmt.Printf("\n  %s %s%s\n", req.Method, req.Host, req.Path)

	var matched *RouterRule
	for i := range t.rules {
		if t.matchRule(&t.rules[i], req) {
			if matched == nil || t.rules[i].Priority > matched.Priority {
				matched = &t.rules[i]
			}
		}
	}

	if matched == nil {
		t.mu.Lock()
		t.stats.ByStatus[404]++
		t.mu.Unlock()
		fmt.Printf("    -> 404 No matching route\n")
		return
	}

	fmt.Printf("    Matched: %s\n", matched.Name)
	for _, mw := range matched.Middlewares {
		if def, ok := t.middlewares[mw]; ok {
			fmt.Printf("    [%s] %s\n", mw, def.Description)
		}
	}

	svc, exists := t.services[matched.Service]
	if !exists {
		fmt.Printf("    -> 503 Service unavailable\n")
		return
	}

	backend := t.selectBackend(svc)
	if backend == nil {
		fmt.Printf("    -> 503 No healthy backends\n")
		return
	}

	t.mu.Lock()
	backend.Requests++
	t.stats.ByService[matched.Service]++
	t.stats.ByStatus[200]++
	t.mu.Unlock()

	fmt.Printf("    -> 200 Proxied to %s\n", backend.Address)
}

func (t *TraefikRouter) matchRule(rule *RouterRule, req TraefikRequest) bool {
	ruleStr := rule.Rule
	if strings.Contains(ruleStr, "Host(") {
		hostStart := strings.Index(ruleStr, "Host(`") + 6
		hostEnd := strings.Index(ruleStr[hostStart:], "`")
		expectedHost := ruleStr[hostStart : hostStart+hostEnd]
		if req.Host != expectedHost {
			return false
		}
	}
	if strings.Contains(ruleStr, "PathPrefix(") {
		pathStart := strings.Index(ruleStr, "PathPrefix(`") + 12
		pathEnd := strings.Index(ruleStr[pathStart:], "`")
		prefix := ruleStr[pathStart : pathStart+pathEnd]
		if !strings.HasPrefix(req.Path, prefix) {
			return false
		}
	}
	return true
}

func (t *TraefikRouter) selectBackend(svc *ServiceGroup) *Backend {
	var healthy []int
	for i := range svc.Backends {
		if svc.Backends[i].Healthy {
			healthy = append(healthy, i)
		}
	}
	if len(healthy) == 0 {
		return nil
	}

	totalWeight := 0
	for _, idx := range healthy {
		totalWeight += svc.Backends[idx].Weight
	}

	svc.NextIdx++
	pick := svc.NextIdx % totalWeight
	cumulative := 0
	for _, idx := range healthy {
		cumulative += svc.Backends[idx].Weight
		if pick < cumulative {
			return &svc.Backends[idx]
		}
	}
	return &svc.Backends[healthy[0]]
}

func (t *TraefikRouter) RunHealthChecks() {
	fmt.Println("  Active Health Checks:")
	for name, svc := range t.services {
		for i := range svc.Backends {
			healthy := rand.Intn(10) > 0
			svc.Backends[i].Healthy = healthy
			status := "OK"
			if !healthy {
				status = "FAIL"
			}
			fmt.Printf("    %s/%s [%s] (requests: %d)\n",
				name, svc.Backends[i].Address, status, svc.Backends[i].Requests)
		}
	}
}

func (t *TraefikRouter) PrintStats() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fmt.Printf("  Total requests: %d\n", t.stats.TotalRequests)
	fmt.Println("  By service:")
	for name, count := range t.stats.ByService {
		fmt.Printf("    %s: %d requests\n", name, count)
	}
	fmt.Println("  By status:")
	for status, count := range t.stats.ByStatus {
		fmt.Printf("    %d: %d\n", status, count)
	}
}

func (t *TraefikRouter) SimulateDockerEvent(event, container, address string) {
	fmt.Printf("  Docker event: %s %s (%s)\n", event, container, address)
	switch event {
	case "container_start":
		fmt.Printf("    Adding backend %s to service pool\n", address)
	case "container_stop":
		fmt.Printf("    Removing backend %s from service pool\n", address)
	}
	fmt.Println("    Configuration reloaded (zero downtime)")
}

func runTraefikDashboard() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/overview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"http":{"routers":3,"services":3,"middlewares":6}}`)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Traefik Dashboard API (port 8080):")
	fmt.Println("  GET /api/overview        - System overview")
	fmt.Println("  GET /api/http/routers    - List HTTP routers")
	fmt.Println("  GET /api/http/services   - List HTTP services")
	fmt.Println("  GET /api/http/middlewares - List middlewares")
	fmt.Println("  GET /api/entrypoints     - List entrypoints")
	fmt.Println("  GET /api/rawdata         - Raw dynamic config")

	_ = server
	_ = os.Stdout
}
