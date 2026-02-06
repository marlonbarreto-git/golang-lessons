// Package main - Chapter 112: Grafana - Observability and Visualization
// Grafana is the leading open-source platform for monitoring visualization.
// Its Go backend demonstrates plugin architecture, multi-datasource
// querying, and dashboard-as-code patterns.
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
	fmt.Println("=== CHAPTER 112: GRAFANA - OBSERVABILITY AND VISUALIZATION ===")

	// ============================================
	// WHAT IS GRAFANA
	// ============================================
	fmt.Println(`
WHAT IS GRAFANA
---------------
Grafana is an open-source analytics and interactive visualization
platform. It provides charts, graphs, and alerts for monitoring
data from multiple sources including Prometheus, InfluxDB, Elasticsearch,
and many more.

Created by Torkel Odegaard in 2014, Grafana has become the de facto
standard for operational dashboards.

KEY FACTS:
- Backend written in Go, frontend in TypeScript/React
- ~66,000+ GitHub stars
- Supports 150+ data source plugins
- Dashboard-as-code with JSON models
- Unified alerting system
- Grafana Labs offers Grafana Cloud (SaaS)
- Powers dashboards at Netflix, EA, PayPal, etc.`)

	// ============================================
	// WHY GO FOR GRAFANA
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR GRAFANA BACKEND
--------------------------------------
1. HTTP performance   - Fast API server for dashboard queries
2. Concurrency        - Parallel data source queries via goroutines
3. Plugin system      - Go plugins for backend data sources
4. Static binary      - Simple deployment alongside frontend assets
5. Database access    - Good ORM support (XORM/GORM)
6. Cross-platform     - Runs on Linux, macOS, Windows, ARM`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
GRAFANA ARCHITECTURE
--------------------

  +--------------------------------------------------+
  |                 BROWSER                           |
  |  +--------------------------------------------+  |
  |  |    React Frontend (TypeScript)              |  |
  |  |  +----------+ +---------+ +-------------+  |  |
  |  |  |Dashboard | |Explore  | |Alert Viewer |  |  |
  |  |  |Renderer  | |Mode     | |             |  |  |
  |  |  +----------+ +---------+ +-------------+  |  |
  |  +---------------------+----------------------+  |
  +-------------------------+------------------------+
                            | HTTP/WebSocket
  +-------------------------v------------------------+
  |              GRAFANA BACKEND (Go)                |
  |  +----------+ +----------+ +----------------+    |
  |  | API      | | Auth     | | Alerting       |    |
  |  | Server   | | (RBAC)   | | Engine         |    |
  |  +----+-----+ +----------+ +-------+--------+    |
  |       |                             |             |
  |  +----v-----------------------------v---------+   |
  |  |           Plugin Manager                    |   |
  |  +--+-------+-------+-------+-------+--------+   |
  |     |       |       |       |       |             |
  +-----+-------+-------+-------+-------+-------------+
        |       |       |       |       |
  +-----v-+ +--v---+ +-v----+ +v-----+ +v---------+
  |Prom   | |PG/   | |Loki  | |Elastic| |CloudWatch|
  |etheus | |MySQL | |      | |search | |          |
  +-------+ +------+ +------+ +------+ +----------+
      Data Source Plugins (Go or gRPC)

LAYERS:
- Frontend:  React-based dashboard rendering
- API:       RESTful + WebSocket for real-time updates
- Backend:   Go services for auth, alerting, provisioning
- Plugins:   Extensible data source and panel plugins
- Storage:   SQLite/PostgreSQL/MySQL for configuration`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN GRAFANA
--------------------------

1. PLUGIN ARCHITECTURE (go-plugin / gRPC)
   - Data source plugins run as separate processes
   - Communicate via gRPC (HashiCorp go-plugin)
   - Hot-reloadable without restarting Grafana
   - Sandboxed execution for security

2. SERVICE REGISTRY PATTERN
   - Services registered at startup via dependency injection
   - Init order resolved automatically
   - Health checks for each service
   - Graceful shutdown in reverse order

3. BUS PATTERN (Event Bus)
   - Internal event bus for decoupled communication
   - Query/Command handlers for CQRS-like operations
   - Alert state changes propagated via events

4. MIDDLEWARE CHAIN
   - HTTP middleware for auth, logging, CORS, rate limiting
   - Composable middleware stack
   - Context propagation for tracing

5. PROVISIONING SYSTEM
   - Dashboard/datasource as YAML/JSON files
   - File watchers detect changes
   - Git-ops friendly configuration
   - Idempotent provisioning on startup`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN GRAFANA
----------------------------------

1. PARALLEL DATA SOURCE QUERIES
   - Dashboard panels query data sources concurrently
   - Goroutine per panel with WaitGroup synchronization
   - Context cancellation on dashboard navigation

2. QUERY CACHING
   - Results cached by query hash + time range
   - Configurable TTL per data source
   - Reduces load on backend data sources

3. LIVE STREAMING (WebSocket)
   - Real-time data push via WebSocket
   - Avoids polling overhead for live dashboards
   - Goroutine per connection with fan-out

4. ALERT EVALUATION OPTIMIZATION
   - Stateful alert evaluator avoids redundant queries
   - Multi-dimensional alerts evaluated in batch
   - Configurable evaluation interval per rule group

5. FRONTEND QUERY SPLITTING
   - Large time ranges split into smaller chunks
   - Parallel chunk execution on backend
   - Results merged before returning to frontend`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
HOW GRAFANA EXPRESSES GO PHILOSOPHY
-------------------------------------

"Accept interfaces, return structs"
  Data source plugins implement a query interface.
  Any data source (Prometheus, SQL, Elastic) returns
  the same DataFrame struct. Uniform data model.

"Don't communicate by sharing memory; share memory
 by communicating"
  Plugin processes communicate via gRPC channels.
  Event bus decouples internal services.

"Errors are values"
  Query errors displayed inline on panels.
  Partial failures don't crash the dashboard.
  Each panel independently handles its errors.

"Clear is better than clever"
  Dashboard JSON model is human-readable.
  API follows REST conventions.
  Configuration via simple YAML files.`)

	// ============================================
	// SIMPLIFIED GRAFANA DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED GRAFANA-LIKE DASHBOARD SYSTEM ===")

	grafana := NewGrafana()

	fmt.Println("\n--- Registering Data Sources ---")
	grafana.AddDataSource(DataSource{
		Name: "Prometheus",
		Type: "prometheus",
		URL:  "http://prometheus:9090",
	})
	grafana.AddDataSource(DataSource{
		Name: "PostgreSQL",
		Type: "postgres",
		URL:  "postgres://grafana:5432/metrics",
	})
	grafana.AddDataSource(DataSource{
		Name: "Loki",
		Type: "loki",
		URL:  "http://loki:3100",
	})

	fmt.Println("\n--- Creating Dashboard ---")
	dashboard := grafana.CreateDashboard(Dashboard{
		Title: "Production Overview",
		Tags:  []string{"production", "overview"},
		Panels: []Panel{
			{
				ID:         1,
				Title:      "Request Rate",
				Type:       "graph",
				DataSource: "Prometheus",
				Query:      "rate(http_requests_total[5m])",
				GridPos:    GridPos{X: 0, Y: 0, W: 12, H: 8},
			},
			{
				ID:         2,
				Title:      "Error Rate",
				Type:       "stat",
				DataSource: "Prometheus",
				Query:      "sum(rate(http_requests_total{status=~\"5..\"}[5m]))",
				GridPos:    GridPos{X: 12, Y: 0, W: 6, H: 4},
			},
			{
				ID:         3,
				Title:      "P99 Latency",
				Type:       "gauge",
				DataSource: "Prometheus",
				Query:      "histogram_quantile(0.99, rate(http_duration_seconds_bucket[5m]))",
				GridPos:    GridPos{X: 18, Y: 0, W: 6, H: 4},
			},
			{
				ID:         4,
				Title:      "Active Users",
				Type:       "stat",
				DataSource: "PostgreSQL",
				Query:      "SELECT count(*) FROM sessions WHERE active = true",
				GridPos:    GridPos{X: 12, Y: 4, W: 12, H: 4},
			},
			{
				ID:         5,
				Title:      "Application Logs",
				Type:       "logs",
				DataSource: "Loki",
				Query:      "{app=\"production\"} |= \"error\"",
				GridPos:    GridPos{X: 0, Y: 8, W: 24, H: 8},
			},
		},
	})

	fmt.Println("\n--- Dashboard Layout ---")
	grafana.RenderDashboard(dashboard)

	fmt.Println("\n--- Executing Panel Queries (Parallel) ---")
	grafana.ExecuteQueries(dashboard)

	fmt.Println("\n--- Alert Rules ---")
	grafana.AddAlertRule(GrafanaAlertRule{
		Name:       "High Error Rate",
		Dashboard:  dashboard,
		PanelID:    2,
		Condition:  "> 0.05",
		Duration:   5 * time.Minute,
		Severity:   "critical",
		NotifyVia:  []string{"slack", "pagerduty"},
	})
	grafana.AddAlertRule(GrafanaAlertRule{
		Name:       "High P99 Latency",
		Dashboard:  dashboard,
		PanelID:    3,
		Condition:  "> 1.0",
		Duration:   10 * time.Minute,
		Severity:   "warning",
		NotifyVia:  []string{"slack"},
	})
	grafana.EvaluateAlerts()

	fmt.Println("\n--- User and Organization ---")
	grafana.CreateOrg("Acme Corp")
	grafana.CreateUser("admin@acme.com", "Admin", "admin")
	grafana.CreateUser("viewer@acme.com", "Alice", "viewer")
	grafana.ListPermissions(dashboard)

	fmt.Println("\n--- Dashboard Export (JSON) ---")
	grafana.ExportDashboard(dashboard)

	fmt.Println("\n--- API Endpoints ---")
	runGrafanaAPIDemo()

	fmt.Println(`
SUMMARY
-------
Grafana is the visualization layer of the observability stack.
Its Go backend demonstrates:
- Plugin architecture for extensible data sources
- Parallel query execution across data sources
- Service registry with dependency injection
- Dashboard-as-code with JSON models
- Unified alerting across multiple data sources

Grafana proved that Go can power complex web application backends.`)
}

// ============================================
// TYPES
// ============================================

type DataSource struct {
	Name string
	Type string
	URL  string
}

type GridPos struct {
	X, Y, W, H int
}

type Panel struct {
	ID         int
	Title      string
	Type       string
	DataSource string
	Query      string
	GridPos    GridPos
}

type Dashboard struct {
	UID    string
	Title  string
	Tags   []string
	Panels []Panel
}

type GrafanaAlertRule struct {
	Name      string
	Dashboard string
	PanelID   int
	Condition string
	Duration  time.Duration
	Severity  string
	NotifyVia []string
}

type QueryResult struct {
	PanelID int
	Data    []DataPoint
	Error   error
}

type DataPoint struct {
	Timestamp time.Time
	Value     float64
	Labels    map[string]string
}

// ============================================
// GRAFANA ENGINE
// ============================================

type Grafana struct {
	dataSources map[string]*DataSource
	dashboards  map[string]*Dashboard
	alerts      []GrafanaAlertRule
	orgs        []string
	users       map[string]string
	mu          sync.RWMutex
}

func NewGrafana() *Grafana {
	return &Grafana{
		dataSources: make(map[string]*DataSource),
		dashboards:  make(map[string]*Dashboard),
		users:       make(map[string]string),
	}
}

func (g *Grafana) AddDataSource(ds DataSource) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.dataSources[ds.Name] = &ds
	fmt.Printf("  Data source: %s (type: %s, url: %s)\n", ds.Name, ds.Type, ds.URL)
}

func (g *Grafana) CreateDashboard(d Dashboard) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	d.UID = fmt.Sprintf("d-%d", time.Now().UnixNano()%100000)
	g.dashboards[d.UID] = &d
	fmt.Printf("  Dashboard: %q (uid: %s, panels: %d, tags: %v)\n",
		d.Title, d.UID, len(d.Panels), d.Tags)
	return d.UID
}

func (g *Grafana) RenderDashboard(uid string) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	dash, exists := g.dashboards[uid]
	if !exists {
		fmt.Printf("  Dashboard %s not found\n", uid)
		return
	}

	fmt.Printf("  Dashboard: %s\n", dash.Title)
	fmt.Println("  +--------------------------------------------------+")

	maxY := 0
	for _, p := range dash.Panels {
		endY := p.GridPos.Y + p.GridPos.H
		if endY > maxY {
			maxY = endY
		}
	}

	grid := make([][]string, maxY)
	for i := range grid {
		grid[i] = make([]string, 24)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	for _, p := range dash.Panels {
		for y := p.GridPos.Y; y < p.GridPos.Y+p.GridPos.H && y < maxY; y++ {
			for x := p.GridPos.X; x < p.GridPos.X+p.GridPos.W && x < 24; x++ {
				if y == p.GridPos.Y && x == p.GridPos.X {
					title := p.Title
					if len(title) > p.GridPos.W-2 {
						title = title[:p.GridPos.W-2]
					}
					grid[y][x] = fmt.Sprintf("[%s]", title)
				}
			}
		}
	}

	for _, p := range dash.Panels {
		fmt.Printf("  | Panel %d: %-20s | %-8s | pos(%d,%d) %dx%d |\n",
			p.ID, p.Title, p.Type, p.GridPos.X, p.GridPos.Y, p.GridPos.W, p.GridPos.H)
	}
	fmt.Println("  +--------------------------------------------------+")
}

func (g *Grafana) ExecuteQueries(uid string) {
	g.mu.RLock()
	dash, exists := g.dashboards[uid]
	if !exists {
		g.mu.RUnlock()
		return
	}
	panels := dash.Panels
	g.mu.RUnlock()

	results := make(chan QueryResult, len(panels))
	var wg sync.WaitGroup

	for _, panel := range panels {
		wg.Add(1)
		go func(p Panel) {
			defer wg.Done()

			start := time.Now()
			time.Sleep(time.Duration(rand.Intn(50)+10) * time.Millisecond)

			var data []DataPoint
			for i := 0; i < 5; i++ {
				data = append(data, DataPoint{
					Timestamp: time.Now().Add(time.Duration(-i) * time.Minute),
					Value:     rand.Float64() * 100,
				})
			}

			elapsed := time.Since(start)
			results <- QueryResult{PanelID: p.ID, Data: data}
			fmt.Printf("  Panel %d (%s): queried %s in %v (%d points)\n",
				p.ID, p.Title, p.DataSource, elapsed.Round(time.Millisecond), len(data))
		}(panel)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []QueryResult
	for r := range results {
		allResults = append(allResults, r)
	}
	fmt.Printf("  All %d panels loaded successfully\n", len(allResults))
}

func (g *Grafana) AddAlertRule(rule GrafanaAlertRule) {
	g.alerts = append(g.alerts, rule)
	fmt.Printf("  Alert: %s (panel %d, condition: value %s, via: %v)\n",
		rule.Name, rule.PanelID, rule.Condition, rule.NotifyVia)
}

func (g *Grafana) EvaluateAlerts() {
	fmt.Println("  Evaluating alert rules...")
	for _, rule := range g.alerts {
		value := rand.Float64() * 2
		state := "OK"
		if rand.Intn(3) == 0 {
			state = "ALERTING"
		}
		fmt.Printf("    %s: value=%.4f %s -> [%s]\n",
			rule.Name, value, rule.Condition, state)
		if state == "ALERTING" {
			fmt.Printf("      Notifications sent to: %s\n",
				strings.Join(rule.NotifyVia, ", "))
		}
	}
}

func (g *Grafana) CreateOrg(name string) {
	g.orgs = append(g.orgs, name)
	fmt.Printf("  Organization: %s created\n", name)
}

func (g *Grafana) CreateUser(email, name, role string) {
	g.users[email] = role
	fmt.Printf("  User: %s (%s) - role: %s\n", name, email, role)
}

func (g *Grafana) ListPermissions(uid string) {
	g.mu.RLock()
	dash, exists := g.dashboards[uid]
	g.mu.RUnlock()

	if !exists {
		return
	}

	fmt.Printf("  Permissions for dashboard %q:\n", dash.Title)
	for email, role := range g.users {
		perm := "View"
		if role == "admin" {
			perm = "Edit/Admin"
		}
		fmt.Printf("    %s (%s): %s\n", email, role, perm)
	}
}

func (g *Grafana) ExportDashboard(uid string) {
	g.mu.RLock()
	dash, exists := g.dashboards[uid]
	g.mu.RUnlock()

	if !exists {
		return
	}

	fmt.Println("  {")
	fmt.Printf("    \"uid\": %q,\n", dash.UID)
	fmt.Printf("    \"title\": %q,\n", dash.Title)
	fmt.Printf("    \"tags\": [%s],\n", formatStringSlice(dash.Tags))
	fmt.Println("    \"panels\": [")
	for i, p := range dash.Panels {
		comma := ","
		if i == len(dash.Panels)-1 {
			comma = ""
		}
		fmt.Printf("      {\"id\": %d, \"title\": %q, \"type\": %q, \"datasource\": %q}%s\n",
			p.ID, p.Title, p.Type, p.DataSource, comma)
	}
	fmt.Println("    ]")
	fmt.Println("  }")
}

func formatStringSlice(ss []string) string {
	var quoted []string
	for _, s := range ss {
		quoted = append(quoted, fmt.Sprintf("%q", s))
	}
	return strings.Join(quoted, ", ")
}

func runGrafanaAPIDemo() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/dashboards/db", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"success","uid":"abc123"}`)
	})
	mux.HandleFunc("/api/datasources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"name":"Prometheus","type":"prometheus"}]`)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Grafana HTTP API endpoints (simulated):")
	fmt.Println("  GET    /api/dashboards/uid/:uid   - Get dashboard")
	fmt.Println("  POST   /api/dashboards/db         - Create/update dashboard")
	fmt.Println("  GET    /api/datasources           - List data sources")
	fmt.Println("  GET    /api/alerts                 - List alerts")
	fmt.Println("  POST   /api/alert-notifications    - Create notification channel")
	fmt.Println("  GET    /api/org/users              - List organization users")
	fmt.Println("  POST   /api/annotations            - Create annotation")

	_ = server
	_ = os.Stdout
}
