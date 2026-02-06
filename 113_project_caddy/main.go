// Package main - Chapter 113: Caddy - Modern Web Server
// Caddy is a powerful, enterprise-ready web server with automatic HTTPS.
// Written in Go, it demonstrates modular architecture, automatic TLS
// certificate management, and configuration-driven design.
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
	fmt.Println("=== CHAPTER 113: CADDY - MODERN WEB SERVER ===")

	// ============================================
	// WHAT IS CADDY
	// ============================================
	fmt.Println(`
WHAT IS CADDY
-------------
Caddy is a modern, open-source web server that automatically manages
TLS certificates. It is the first major web server to provide HTTPS
by default using Let's Encrypt.

Created by Matt Holt in 2015, Caddy v2 was rewritten from scratch
with a modular, plugin-based architecture.

KEY FACTS:
- Written entirely in Go
- ~60,000+ GitHub stars
- Automatic HTTPS via ACME protocol (Let's Encrypt)
- JSON API for dynamic configuration
- Caddyfile for human-friendly config
- Modular plugin architecture
- HTTP/1.1, HTTP/2, and HTTP/3 support
- Used as reverse proxy, file server, load balancer`)

	// ============================================
	// WHY GO FOR CADDY
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR CADDY
-----------------------------
1. net/http stdlib     - Excellent HTTP server built-in
2. crypto/tls          - Native TLS support in stdlib
3. Static binary       - Single binary with all features included
4. Cross-compilation   - Build for any platform
5. Goroutines          - Handle thousands of concurrent connections
6. Plugin system       - Go interfaces enable clean module system`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
CADDY ARCHITECTURE
------------------

  +----------------------------------------------------+
  |                     CADDY                           |
  |  +---------------------------------------------+   |
  |  |              Admin API (:2019)               |   |
  |  |        JSON config / Caddyfile adapter       |   |
  |  +---------------------+-----------------------+   |
  |                         |                           |
  |  +---------------------v-----------------------+   |
  |  |              Module System                   |   |
  |  |  +----------+  +----------+  +----------+   |   |
  |  |  | HTTP App |  | TLS App  |  | PKI App  |   |   |
  |  |  +----+-----+  +----+-----+  +----------+   |   |
  |  |       |              |                       |   |
  |  |  +----v-----+  +----v-----------+            |   |
  |  |  | Routes   |  | Certificate    |            |   |
  |  |  | Matchers |  | Manager (ACME) |            |   |
  |  |  | Handlers |  +----------------+            |   |
  |  |  +----------+                                |   |
  |  +---------------------------------------------+   |
  |                                                      |
  |  Handler Chain (per request):                        |
  |  Request -> Matcher -> Handler -> Response           |
  |  [encode] -> [headers] -> [reverse_proxy] -> [resp] |
  +----------------------------------------------------+

KEY CONCEPTS:
- Apps:      Top-level modules (http, tls, pki)
- Matchers:  Decide which requests a route handles
- Handlers:  Process matched requests
- Middleware: Composable handler chain`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN CADDY
------------------------

1. MODULE SYSTEM
   - Every feature is a module (interface-based)
   - Modules registered in init() functions
   - Loaded dynamically from JSON config
   - Compile-time module inclusion via Go imports

   type Module interface {
       CaddyModule() ModuleInfo
   }
   type Provisioner interface {
       Provision(Context) error
   }
   type Validator interface {
       Validate() error
   }

2. HANDLER CHAIN (Middleware Pattern)
   - Each handler wraps the next
   - Request flows through chain
   - Any handler can short-circuit
   - Response flows back through chain

3. AUTOMATIC CERTIFICATE MANAGEMENT
   - ACME protocol implementation
   - Certificate storage abstraction
   - Automatic renewal goroutines
   - OCSP stapling for performance

4. CONFIG ADAPTERS
   - Caddyfile -> JSON adapter
   - Any format can adapt to JSON
   - JSON is the canonical config format
   - Hot-reloading without downtime

5. REVERSE PROXY WITH LOAD BALANCING
   - Multiple upstream selection policies
   - Health checking goroutines per upstream
   - Circuit breaker pattern
   - Passive health checks from response codes`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN CADDY
---------------------------------

1. HTTP/3 (QUIC) SUPPORT
   - Built on quic-go library
   - 0-RTT connection establishment
   - Multiplexed streams without head-of-line blocking

2. ON-DEMAND TLS
   - Certificates obtained on first request
   - Cached for subsequent requests
   - No pre-configuration needed
   - Rate limiting prevents abuse

3. CONNECTION POOLING
   - Reverse proxy reuses backend connections
   - Keep-alive management per upstream
   - Idle connection cleanup goroutines

4. RESPONSE BUFFERING
   - Optional response buffering
   - Streaming for large responses
   - Compression (gzip, zstd) middleware

5. GRACEFUL CONFIG RELOADS
   - New config loaded in background
   - Active connections finish on old config
   - Zero-downtime configuration changes`)

	// ============================================
	// CADDYFILE EXAMPLES
	// ============================================
	fmt.Println(`
CADDYFILE CONFIGURATION
-----------------------

# Simple file server with automatic HTTPS:
example.com {
    root * /var/www/html
    file_server
}

# Reverse proxy with load balancing:
api.example.com {
    reverse_proxy localhost:8001 localhost:8002 {
        lb_policy round_robin
        health_uri /health
        health_interval 10s
    }
}

# Full application setup:
app.example.com {
    encode gzip zstd
    header X-Frame-Options DENY
    log {
        output file /var/log/caddy/access.log
    }
    reverse_proxy /api/* localhost:9000
    file_server
}`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
HOW CADDY EXPRESSES GO PHILOSOPHY
-----------------------------------

"Make the zero value useful"
  Caddy with zero config still serves HTTPS.
  Default behavior is secure and sensible.
  No configuration needed for basic use cases.

"Accept interfaces, return structs"
  Module system built on interfaces.
  Any module implementing the interface works.
  Caddy core doesn't know about specific modules.

"Simplicity is prerequisite for reliability"
  Automatic HTTPS removes manual certificate management.
  One binary replaces nginx + certbot + config files.

"A little copying is better than a little dependency"
  Caddy includes its own ACME client, TLS utilities,
  and HTTP server rather than depending on external tools.`)

	// ============================================
	// SIMPLIFIED CADDY DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED CADDY-LIKE WEB SERVER ===")

	server := NewCaddyServer()

	fmt.Println("\n--- Configuring Routes ---")
	server.AddRoute(Route{
		Match:   HostMatcher("example.com"),
		Handler: FileServerHandler("/var/www/html"),
	})
	server.AddRoute(Route{
		Match:   PathMatcher("/api/*"),
		Handler: ReverseProxyHandler([]string{"localhost:8001", "localhost:8002", "localhost:8003"}),
	})
	server.AddRoute(Route{
		Match:   PathMatcher("/health"),
		Handler: HealthCheckHandler(),
	})
	server.AddRoute(Route{
		Match:   MatchAll(),
		Handler: FileServerHandler("/var/www/default"),
	})

	fmt.Println("\n--- Adding Middleware ---")
	server.AddMiddleware("gzip_compression")
	server.AddMiddleware("access_logging")
	server.AddMiddleware("security_headers")
	server.AddMiddleware("rate_limiter")

	fmt.Println("\n--- TLS Certificate Management ---")
	server.ManageCertificates([]string{"example.com", "api.example.com", "www.example.com"})

	fmt.Println("\n--- Processing Requests ---")
	requests := []Request{
		{Method: "GET", Host: "example.com", Path: "/index.html"},
		{Method: "GET", Host: "example.com", Path: "/api/users"},
		{Method: "POST", Host: "example.com", Path: "/api/orders"},
		{Method: "GET", Host: "example.com", Path: "/health"},
		{Method: "GET", Host: "example.com", Path: "/static/style.css"},
		{Method: "GET", Host: "unknown.com", Path: "/"},
	}

	for _, req := range requests {
		server.HandleRequest(req)
	}

	fmt.Println("\n--- Load Balancer Status ---")
	server.PrintUpstreamHealth()

	fmt.Println("\n--- Server Statistics ---")
	server.PrintStats()

	fmt.Println("\n--- Admin API ---")
	runCaddyAdminAPIDemo()

	fmt.Println(`
SUMMARY
-------
Caddy modernized web server deployment.
Its Go codebase demonstrates:
- Automatic TLS certificate management (ACME)
- Modular plugin architecture via Go interfaces
- Handler chain middleware pattern
- Zero-downtime configuration reloads
- HTTP/1.1, HTTP/2, and HTTP/3 support

Caddy proved that secure-by-default is achievable.`)
}

// ============================================
// TYPES
// ============================================

type Request struct {
	Method string
	Host   string
	Path   string
}

type Response struct {
	Status  int
	Body    string
	Headers map[string]string
}

type MatcherFunc func(Request) bool
type HandlerFunc func(Request) Response

type Route struct {
	Match   MatcherFunc
	Handler HandlerFunc
	Name    string
}

type Upstream struct {
	Address    string
	Healthy    bool
	Requests   int
	LastCheck  time.Time
}

// ============================================
// MATCHERS
// ============================================

func HostMatcher(host string) MatcherFunc {
	return func(r Request) bool {
		return r.Host == host
	}
}

func PathMatcher(pattern string) MatcherFunc {
	return func(r Request) bool {
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			return strings.HasPrefix(r.Path, prefix)
		}
		return r.Path == pattern
	}
}

func MatchAll() MatcherFunc {
	return func(r Request) bool { return true }
}

// ============================================
// HANDLERS
// ============================================

func FileServerHandler(root string) HandlerFunc {
	return func(r Request) Response {
		return Response{
			Status: 200,
			Body:   fmt.Sprintf("File: %s%s", root, r.Path),
			Headers: map[string]string{
				"Content-Type":  "text/html",
				"Cache-Control": "public, max-age=3600",
			},
		}
	}
}

func ReverseProxyHandler(upstreams []string) HandlerFunc {
	idx := 0
	return func(r Request) Response {
		upstream := upstreams[idx%len(upstreams)]
		idx++
		return Response{
			Status: 200,
			Body:   fmt.Sprintf("Proxied to %s", upstream),
			Headers: map[string]string{
				"X-Upstream":    upstream,
				"Content-Type":  "application/json",
			},
		}
	}
}

func HealthCheckHandler() HandlerFunc {
	return func(r Request) Response {
		return Response{
			Status: 200,
			Body:   `{"status":"healthy"}`,
			Headers: map[string]string{"Content-Type": "application/json"},
		}
	}
}

// ============================================
// CADDY SERVER
// ============================================

type CaddyServer struct {
	routes     []Route
	middleware []string
	upstreams  map[string]*Upstream
	certs      map[string]*TLSCert
	stats      ServerStats
	mu         sync.RWMutex
}

type TLSCert struct {
	Domain    string
	Issuer    string
	NotBefore time.Time
	NotAfter  time.Time
	AutoRenew bool
}

type ServerStats struct {
	TotalRequests int
	StatusCounts  map[int]int
	ByPath        map[string]int
}

func NewCaddyServer() *CaddyServer {
	return &CaddyServer{
		upstreams: make(map[string]*Upstream),
		certs:     make(map[string]*TLSCert),
		stats: ServerStats{
			StatusCounts: make(map[int]int),
			ByPath:       make(map[string]int),
		},
	}
}

func (s *CaddyServer) AddRoute(route Route) {
	s.routes = append(s.routes, route)
	fmt.Printf("  Route added: handler registered\n")
}

func (s *CaddyServer) AddMiddleware(name string) {
	s.middleware = append(s.middleware, name)
	fmt.Printf("  Middleware: %s\n", name)
}

func (s *CaddyServer) ManageCertificates(domains []string) {
	for _, domain := range domains {
		cert := &TLSCert{
			Domain:    domain,
			Issuer:    "Let's Encrypt",
			NotBefore: time.Now(),
			NotAfter:  time.Now().Add(90 * 24 * time.Hour),
			AutoRenew: true,
		}
		s.certs[domain] = cert
		fmt.Printf("  Certificate: %s (issuer: %s, expires: %s, auto-renew: true)\n",
			domain, cert.Issuer, cert.NotAfter.Format("2006-01-02"))
	}
}

func (s *CaddyServer) HandleRequest(req Request) {
	s.mu.Lock()
	s.stats.TotalRequests++
	s.mu.Unlock()

	fmt.Printf("\n  %s %s%s\n", req.Method, req.Host, req.Path)

	for _, mw := range s.middleware {
		fmt.Printf("    [middleware] %s\n", mw)
	}

	for _, route := range s.routes {
		if route.Match(req) {
			resp := route.Handler(req)
			s.mu.Lock()
			s.stats.StatusCounts[resp.Status]++
			s.stats.ByPath[req.Path]++
			s.mu.Unlock()

			fmt.Printf("    -> %d | %s\n", resp.Status, resp.Body)
			for k, v := range resp.Headers {
				fmt.Printf("       %s: %s\n", k, v)
			}
			return
		}
	}

	s.mu.Lock()
	s.stats.StatusCounts[404]++
	s.mu.Unlock()
	fmt.Printf("    -> 404 | Not Found\n")
}

func (s *CaddyServer) PrintUpstreamHealth() {
	upstreams := []string{"localhost:8001", "localhost:8002", "localhost:8003"}
	fmt.Println("  Upstream Health:")
	for _, addr := range upstreams {
		healthy := rand.Intn(10) > 0
		status := "healthy"
		if !healthy {
			status = "unhealthy"
		}
		fmt.Printf("    %s [%s] (requests: %d)\n", addr, status, rand.Intn(100))
	}
}

func (s *CaddyServer) PrintStats() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fmt.Printf("  Total requests: %d\n", s.stats.TotalRequests)
	fmt.Println("  Status codes:")
	for code, count := range s.stats.StatusCounts {
		fmt.Printf("    %d: %d requests\n", code, count)
	}
	fmt.Println("  TLS certificates:")
	for domain, cert := range s.certs {
		daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
		fmt.Printf("    %s: expires in %d days (auto-renew: %v)\n",
			domain, daysLeft, cert.AutoRenew)
	}
}

func runCaddyAdminAPIDemo() {
	mux := http.NewServeMux()
	mux.HandleFunc("/config/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"apps":{"http":{"servers":{}}}}`)
	})
	mux.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Caddy Admin API (port 2019):")
	fmt.Println("  GET    /config/             - Get full config")
	fmt.Println("  POST   /load                - Load new config")
	fmt.Println("  GET    /config/apps/http    - Get HTTP app config")
	fmt.Println("  PUT    /config/apps/http    - Update HTTP config")
	fmt.Println("  DELETE /config/apps/http    - Delete HTTP config")
	fmt.Println("  GET    /reverse_proxy/upstreams - List upstreams")

	_ = server
	_ = os.Stdout
}
