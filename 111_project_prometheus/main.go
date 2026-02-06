// Package main - Chapter 111: Prometheus - Monitoring and Alerting
// Prometheus is the standard for cloud-native monitoring and alerting.
// Written in Go, it demonstrates time-series databases, pull-based
// metrics collection, and efficient data compression.
package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 111: PROMETHEUS - MONITORING AND ALERTING ===")

	// ============================================
	// WHAT IS PROMETHEUS
	// ============================================
	fmt.Println(`
WHAT IS PROMETHEUS
------------------
Prometheus is an open-source monitoring and alerting toolkit designed
for reliability and scalability. It collects metrics from targets by
scraping HTTP endpoints and stores them in a time-series database.

Originally built at SoundCloud in 2012, Prometheus became the second
CNCF graduated project (after Kubernetes) in 2018.

KEY FACTS:
- Written entirely in Go
- ~56,000+ GitHub stars
- Pull-based metrics collection (scraping)
- Multi-dimensional data model (labels)
- PromQL query language
- Built-in alerting via Alertmanager
- No external dependencies for storage`)

	// ============================================
	// WHY GO FOR PROMETHEUS
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR PROMETHEUS
---------------------------------
1. Performance       - Efficient time-series ingestion and queries
2. Concurrency       - Goroutines for parallel scraping of targets
3. Static binaries   - Easy deployment, no JVM or runtime needed
4. Low memory        - Efficient memory management for large datasets
5. HTTP native       - Excellent stdlib for scraping and serving APIs
6. Ecosystem         - Client libraries easy to write in Go`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
PROMETHEUS ARCHITECTURE
-----------------------

                    +-------------------+
                    |   Alertmanager    |
                    | (dedupe, group,   |
                    |  route, notify)   |
                    +--------^----------+
                             |
  +----------------------------------------------+
  |              Prometheus Server                |
  |  +----------+  +----------+  +----------+    |
  |  | Scraper  |  |  TSDB    |  |  PromQL  |    |
  |  | (pull)   |  | (storage)|  |  Engine  |    |
  |  +----+-----+  +----+-----+  +----+-----+    |
  |       |              |              |         |
  |  +----v--------------v--------------v-----+   |
  |  |          HTTP API (/api/v1/...)         |   |
  |  +-----------------------------------------+   |
  +--+---+---+---+---+---+---+---+---+--------+
     |   |   |   |   |   |   |   |   |
  +--v-+ | +-v-+ | +-v-+ | +-v-+ | +-v--+
  |app1| | |app2| | |app3| | |app4| | |node|
  |/met| | |/met| | |/met| | |/met| | |exp |
  +----+ | +----+ | +----+ | +----+ | +----+
         |        |        |        |
     Targets expose /metrics endpoint

  +-------------------+
  |    Grafana         |  (visualization)
  |  (queries PromQL)  |
  +-------------------+

PULL MODEL:
- Prometheus PULLS metrics from targets
- Each target exposes /metrics HTTP endpoint
- Scrape interval configurable (default: 15s)
- Advantages: targets don't need to know about Prometheus
  service discovery finds new targets automatically`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN PROMETHEUS
-----------------------------

1. TIME-SERIES DATA MODEL
   metric_name{label1="val1", label2="val2"} value timestamp

   Example:
   http_requests_total{method="GET", path="/api", status="200"} 1234
   process_cpu_seconds_total 47.89
   go_goroutines 142

2. METRIC TYPES
   - Counter:   Only goes up (requests, errors, bytes)
   - Gauge:     Goes up and down (temperature, queue size)
   - Histogram: Observation buckets (request latency)
   - Summary:   Quantiles over sliding window

3. CONCURRENT SCRAPING
   - Goroutine per target for parallel scraping
   - Scrape pool manages target lifecycle
   - Stagger scrapes to avoid thundering herd
   - Context cancellation for timeouts

4. TSDB (Time Series Database)
   - Custom storage engine optimized for time series
   - Head block (in-memory) for recent data
   - Compacted blocks on disk for historical data
   - Write-ahead log for crash recovery

5. PROMQL ENGINE
   - Custom query language for time series
   - Lazy evaluation of range vectors
   - Parallel execution of independent sub-queries
   - Memory-bounded evaluation`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN PROMETHEUS
-------------------------------------

1. GORILLA COMPRESSION (XOR encoding)
   - Time deltas stored as delta-of-deltas
   - Values XOR'd with previous value
   - 12x compression ratio for typical metrics
   - ~1.37 bytes per sample

2. INVERTED INDEX
   - Labels indexed for fast lookups
   - Posting lists intersected for multi-label queries
   - Memory-mapped for efficient disk access

3. HEAD BLOCK (in-memory)
   - Most recent 2 hours in memory
   - Lock-free append for high throughput
   - Periodic compaction to disk blocks

4. BLOCK COMPACTION
   - Small blocks merged into larger ones
   - Reduces file count and improves query speed
   - Tombstones for efficient deletion

5. EXEMPLAR STORAGE
   - Links between metrics and traces
   - Stored alongside series data
   - Enables metric-to-trace correlation`)

	// ============================================
	// PROMQL
	// ============================================
	fmt.Println(`
PROMQL - PROMETHEUS QUERY LANGUAGE
----------------------------------

INSTANT VECTOR (current value):
  http_requests_total{status="200"}
  rate(http_requests_total[5m])

RANGE VECTOR (time range):
  http_requests_total[5m]

AGGREGATION:
  sum(rate(http_requests_total[5m])) by (method)
  avg(node_cpu_seconds_total) by (instance)
  histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

FUNCTIONS:
  rate()       - Per-second rate over range
  increase()   - Total increase over range
  avg_over_time() - Average over range
  predict_linear() - Linear prediction`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
HOW PROMETHEUS EXPRESSES GO PHILOSOPHY
---------------------------------------

"Simplicity is prerequisite for reliability"
  Pull-based model is simpler than push-based.
  No complex message queues or buffering needed.
  Targets just serve an HTTP endpoint.

"Make the zero value useful"
  Default scrape interval, default retention,
  default alert evaluation. Works out of the box.

"A little copying is better than a little dependency"
  Prometheus has its own TSDB rather than using
  an external database. Full control over storage.

"Don't communicate by sharing memory; share memory
 by communicating"
  Scraping IS communication. Each target independently
  exposes its state. No shared state between targets.`)

	// ============================================
	// SIMPLIFIED PROMETHEUS DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED PROMETHEUS-LIKE MONITORING SYSTEM ===")

	prom := NewPrometheus()

	fmt.Println("\n--- Registering Metrics ---")
	prom.RegisterCounter("http_requests_total", "Total HTTP requests", []string{"method", "status"})
	prom.RegisterGauge("goroutines_count", "Number of goroutines", []string{"service"})
	prom.RegisterHistogram("http_request_duration_seconds", "Request latency", []string{"method"},
		[]float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0})

	fmt.Println("\n--- Recording Metrics ---")
	methods := []string{"GET", "POST", "PUT"}
	statuses := []string{"200", "201", "404", "500"}

	for i := 0; i < 100; i++ {
		method := methods[rand.Intn(len(methods))]
		status := statuses[rand.Intn(len(statuses))]
		prom.IncCounter("http_requests_total", map[string]string{"method": method, "status": status})

		latency := rand.ExpFloat64() * 0.1
		prom.ObserveHistogram("http_request_duration_seconds", map[string]string{"method": method}, latency)
	}

	prom.SetGauge("goroutines_count", map[string]string{"service": "web"}, 142)
	prom.SetGauge("goroutines_count", map[string]string{"service": "worker"}, 89)

	fmt.Printf("  Recorded 100 HTTP requests across %d methods and %d statuses\n",
		len(methods), len(statuses))

	fmt.Println("\n--- Metrics Endpoint (/metrics) ---")
	prom.ExposeMetrics()

	fmt.Println("\n--- PromQL Queries ---")
	prom.Query("http_requests_total")
	prom.QueryRate("http_requests_total", map[string]string{"method": "GET"})
	prom.QuerySum("http_requests_total", "method")
	prom.QueryHistogramQuantile("http_request_duration_seconds", 0.99)

	fmt.Println("\n--- Alerting Rules ---")
	prom.AddAlertRule(AlertRule{
		Name:     "HighErrorRate",
		Expr:     "rate(http_requests_total{status='500'}[5m]) > 0.1",
		Duration: 5 * time.Minute,
		Labels:   map[string]string{"severity": "critical"},
		Annotations: map[string]string{
			"summary": "High error rate detected",
		},
	})
	prom.AddAlertRule(AlertRule{
		Name:     "HighLatency",
		Expr:     "histogram_quantile(0.99, http_request_duration_seconds) > 1.0",
		Duration: 10 * time.Minute,
		Labels:   map[string]string{"severity": "warning"},
		Annotations: map[string]string{
			"summary": "P99 latency above 1 second",
		},
	})
	prom.EvaluateAlerts()

	fmt.Println("\n--- Scrape Targets ---")
	prom.AddTarget("web-1", "http://10.0.0.1:8080/metrics")
	prom.AddTarget("web-2", "http://10.0.0.2:8080/metrics")
	prom.AddTarget("api-1", "http://10.0.0.3:9090/metrics")
	prom.ScrapeTargets()

	fmt.Println("\n--- HTTP API Demo ---")
	runPrometheusAPIDemo()

	fmt.Println(`
SUMMARY
-------
Prometheus defined cloud-native monitoring.
Its Go codebase demonstrates:
- Custom time-series database with Gorilla compression
- Pull-based scraping with goroutine concurrency
- PromQL query engine for data analysis
- Multi-dimensional data model with labels
- Alert evaluation and notification routing

Prometheus made metrics-driven operations standard practice.`)
}

// ============================================
// TYPES
// ============================================

type MetricType string

const (
	MetricCounter   MetricType = "counter"
	MetricGauge     MetricType = "gauge"
	MetricHistogram MetricType = "histogram"
)

type MetricFamily struct {
	Name       string
	Help       string
	Type       MetricType
	LabelNames []string
	Series     map[string]*Series
	Buckets    []float64
}

type Series struct {
	Labels map[string]string
	Value  float64
	Counts map[float64]int
	Sum    float64
	Count  int
}

type AlertRule struct {
	Name        string
	Expr        string
	Duration    time.Duration
	Labels      map[string]string
	Annotations map[string]string
}

type ScrapeTarget struct {
	Name     string
	URL      string
	Up       bool
	LastScrape time.Time
}

// ============================================
// PROMETHEUS ENGINE
// ============================================

type Prometheus struct {
	metrics map[string]*MetricFamily
	alerts  []AlertRule
	targets map[string]*ScrapeTarget
	mu      sync.RWMutex
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		metrics: make(map[string]*MetricFamily),
		targets: make(map[string]*ScrapeTarget),
	}
}

func (p *Prometheus) RegisterCounter(name, help string, labelNames []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics[name] = &MetricFamily{
		Name:       name,
		Help:       help,
		Type:       MetricCounter,
		LabelNames: labelNames,
		Series:     make(map[string]*Series),
	}
	fmt.Printf("  Registered counter: %s (%s)\n", name, help)
}

func (p *Prometheus) RegisterGauge(name, help string, labelNames []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics[name] = &MetricFamily{
		Name:       name,
		Help:       help,
		Type:       MetricGauge,
		LabelNames: labelNames,
		Series:     make(map[string]*Series),
	}
	fmt.Printf("  Registered gauge: %s (%s)\n", name, help)
}

func (p *Prometheus) RegisterHistogram(name, help string, labelNames []string, buckets []float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metrics[name] = &MetricFamily{
		Name:       name,
		Help:       help,
		Type:       MetricHistogram,
		LabelNames: labelNames,
		Series:     make(map[string]*Series),
		Buckets:    buckets,
	}
	fmt.Printf("  Registered histogram: %s (buckets: %v)\n", name, buckets)
}

func (p *Prometheus) IncCounter(name string, labels map[string]string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	mf, exists := p.metrics[name]
	if !exists {
		return
	}

	key := labelsKey(labels)
	series, exists := mf.Series[key]
	if !exists {
		series = &Series{Labels: labels}
		mf.Series[key] = series
	}
	series.Value++
}

func (p *Prometheus) SetGauge(name string, labels map[string]string, value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	mf, exists := p.metrics[name]
	if !exists {
		return
	}

	key := labelsKey(labels)
	mf.Series[key] = &Series{Labels: labels, Value: value}
}

func (p *Prometheus) ObserveHistogram(name string, labels map[string]string, value float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	mf, exists := p.metrics[name]
	if !exists {
		return
	}

	key := labelsKey(labels)
	series, exists := mf.Series[key]
	if !exists {
		series = &Series{
			Labels: labels,
			Counts: make(map[float64]int),
		}
		mf.Series[key] = series
	}
	series.Sum += value
	series.Count++
	for _, bucket := range mf.Buckets {
		if value <= bucket {
			series.Counts[bucket]++
		}
	}
}

func (p *Prometheus) ExposeMetrics() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var names []string
	for name := range p.metrics {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		mf := p.metrics[name]
		os.Stdout.WriteString(fmt.Sprintf("# HELP %s %s\n", mf.Name, mf.Help))
		os.Stdout.WriteString(fmt.Sprintf("# TYPE %s %s\n", mf.Name, string(mf.Type)))

		var keys []string
		for k := range mf.Series {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			series := mf.Series[key]
			labelStr := formatLabels(series.Labels)

			switch mf.Type {
			case MetricCounter, MetricGauge:
				fmt.Printf("%s{%s} %.0f\n", mf.Name, labelStr, series.Value)
			case MetricHistogram:
				for _, bucket := range mf.Buckets {
					count := series.Counts[bucket]
					fmt.Printf("%s_bucket{%s,le=\"%.2f\"} %d\n",
						mf.Name, labelStr, bucket, count)
				}
				fmt.Printf("%s_bucket{%s,le=\"+Inf\"} %d\n", mf.Name, labelStr, series.Count)
				fmt.Printf("%s_sum{%s} %.4f\n", mf.Name, labelStr, series.Sum)
				fmt.Printf("%s_count{%s} %d\n", mf.Name, labelStr, series.Count)
			}
		}
		fmt.Println()
	}
}

func (p *Prometheus) Query(name string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	mf, exists := p.metrics[name]
	if !exists {
		fmt.Printf("  Query: %s => (no data)\n", name)
		return
	}

	fmt.Printf("  Query: %s\n", name)
	for _, series := range mf.Series {
		labelStr := formatLabels(series.Labels)
		fmt.Printf("    %s{%s} => %.0f\n", name, labelStr, series.Value)
	}
}

func (p *Prometheus) QueryRate(name string, filter map[string]string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	mf, exists := p.metrics[name]
	if !exists {
		return
	}

	fmt.Printf("  Query: rate(%s{%s}[5m])\n", name, formatLabels(filter))
	for _, series := range mf.Series {
		match := true
		for k, v := range filter {
			if series.Labels[k] != v {
				match = false
				break
			}
		}
		if match {
			rate := series.Value / 300.0
			fmt.Printf("    {%s} => %.4f req/s\n", formatLabels(series.Labels), rate)
		}
	}
}

func (p *Prometheus) QuerySum(name, by string) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	mf, exists := p.metrics[name]
	if !exists {
		return
	}

	sums := make(map[string]float64)
	for _, series := range mf.Series {
		key := series.Labels[by]
		sums[key] += series.Value
	}

	fmt.Printf("  Query: sum(%s) by (%s)\n", name, by)
	for k, v := range sums {
		fmt.Printf("    {%s=%q} => %.0f\n", by, k, v)
	}
}

func (p *Prometheus) QueryHistogramQuantile(name string, quantile float64) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	mf, exists := p.metrics[name]
	if !exists {
		return
	}

	fmt.Printf("  Query: histogram_quantile(%.2f, %s)\n", quantile, name)
	for _, series := range mf.Series {
		if series.Count == 0 {
			continue
		}
		target := int(math.Ceil(float64(series.Count) * quantile))
		cumulative := 0
		for _, bucket := range mf.Buckets {
			cumulative += series.Counts[bucket]
			if cumulative >= target {
				fmt.Printf("    {%s} => %.4fs (p%.0f)\n",
					formatLabels(series.Labels), bucket, quantile*100)
				break
			}
		}
	}
}

func (p *Prometheus) AddAlertRule(rule AlertRule) {
	p.alerts = append(p.alerts, rule)
	fmt.Printf("  Alert rule: %s (expr: %s, for: %s)\n",
		rule.Name, rule.Expr, rule.Duration)
}

func (p *Prometheus) EvaluateAlerts() {
	fmt.Println("  Evaluating alert rules...")
	for _, rule := range p.alerts {
		firing := rand.Intn(3) == 0
		state := "inactive"
		if firing {
			state = "FIRING"
		}
		fmt.Printf("    %s [%s]: %s (severity: %s)\n",
			rule.Name, state, rule.Annotations["summary"], rule.Labels["severity"])
	}
}

func (p *Prometheus) AddTarget(name, url string) {
	p.targets[name] = &ScrapeTarget{
		Name: name,
		URL:  url,
		Up:   true,
	}
}

func (p *Prometheus) ScrapeTargets() {
	fmt.Println("Scraping targets:")
	for _, target := range p.targets {
		target.LastScrape = time.Now()
		target.Up = rand.Intn(10) > 0
		status := "UP"
		if !target.Up {
			status = "DOWN"
		}
		fmt.Printf("  %s (%s) [%s] scraped at %s\n",
			target.Name, target.URL, status,
			target.LastScrape.Format("15:04:05"))
	}
}

func runPrometheusAPIDemo() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"success","data":{"resultType":"vector"}}`)
	})
	mux.HandleFunc("/api/v1/targets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"success","data":{"activeTargets":[]}}`)
	})

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Prometheus HTTP API endpoints (simulated):")
	fmt.Println("  GET  /api/v1/query           - Instant query")
	fmt.Println("  GET  /api/v1/query_range     - Range query")
	fmt.Println("  GET  /api/v1/targets         - Target discovery")
	fmt.Println("  GET  /api/v1/alerts          - Active alerts")
	fmt.Println("  GET  /metrics                - Self-monitoring")
	fmt.Println("  POST /api/v1/admin/tsdb/snapshot - Create snapshot")

	_ = server
}

// ============================================
// HELPERS
// ============================================

func labelsKey(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func formatLabels(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%q", k, v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}
