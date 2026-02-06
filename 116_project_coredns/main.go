// Package main - Chapter 116: CoreDNS - DNS Server and Service Discovery
// CoreDNS is a flexible, extensible DNS server written in Go.
// It is the default DNS server in Kubernetes, demonstrating plugin chain
// architecture, middleware composition, and network programming.
package main

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 116: COREDNS - DNS SERVER AND SERVICE DISCOVERY ===")

	// ============================================
	// WHAT IS COREDNS
	// ============================================
	fmt.Println(`
WHAT IS COREDNS
---------------
CoreDNS is a DNS server that chains plugins. It is the successor to
SkyDNS and the default DNS service in Kubernetes clusters, handling
service discovery and name resolution.

Created by Miek Gieben in 2016, CoreDNS is a CNCF graduated project
that replaced kube-dns as the Kubernetes default.

KEY FACTS:
- Written entirely in Go
- ~12,000+ GitHub stars
- Default DNS in Kubernetes since 1.13
- Plugin-based architecture (30+ built-in plugins)
- Supports DNS, DNS over TLS (DoT), DNS over HTTPS (DoH)
- Serves from files, databases, cloud APIs
- CNCF graduated project`)

	// ============================================
	// WHY GO FOR COREDNS
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR COREDNS
-------------------------------
1. Net package        - Excellent UDP/TCP networking in stdlib
2. Goroutines         - Handle thousands of DNS queries concurrently
3. dns library        - Miek's Go DNS library (github.com/miekg/dns)
4. Static binary      - Easy deployment in containers
5. Plugin system      - Go interfaces for composable plugins
6. Performance        - Low-latency DNS resolution`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
COREDNS ARCHITECTURE
--------------------

  +-----------------------------------------------+
  |                   CoreDNS                      |
  |  +------------------------------------------+ |
  |  |             Corefile Config               | |
  |  |  .:53 {                                   | |
  |  |    kubernetes cluster.local               | |
  |  |    forward . 8.8.8.8                      | |
  |  |    cache 30                               | |
  |  |    log                                    | |
  |  |  }                                        | |
  |  +------------------------------------------+ |
  |                                                |
  |  DNS Query Flow (Plugin Chain):                |
  |                                                |
  |  Query -> [log] -> [cache] -> [kubernetes]     |
  |                                   |            |
  |           [forward] <-- miss <----+            |
  |               |                                |
  |           8.8.8.8                              |
  |               |                                |
  |           Response -> [cache store] -> Client  |
  +-----------------------------------------------+

  Kubernetes Integration:

  +----------+     +----------+     +----------+
  | Pod A    |     | CoreDNS  |     | K8s API  |
  | nslookup | --> | (cache + | --> | Server   |
  | svc-b    |     |  watch)  |     | (etcd)   |
  +----------+     +----------+     +----------+

  DNS Query: svc-b.default.svc.cluster.local
  Answer:    10.96.45.123 (ClusterIP of svc-b)`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN COREDNS
--------------------------

1. PLUGIN CHAIN (Middleware Pattern)
   - Each plugin processes the DNS query
   - Can respond, modify, or pass to next plugin
   - Order defined in Corefile configuration
   - Inspired by Caddy's middleware architecture

   type Plugin interface {
       ServeDNS(ctx context.Context, w ResponseWriter,
               r *dns.Msg) (int, error)
       Name() string
   }

2. SERVER BLOCKS
   - Each zone/port combination is a server block
   - Independent plugin chains per block
   - Wildcard zones for catch-all handling
   - Zone-based request routing

3. WATCHER PATTERN (Kubernetes Plugin)
   - Watches K8s API for Service/Endpoint changes
   - Maintains in-memory DNS records
   - Updates are immediate (no polling delay)
   - Graceful handling of API disconnections

4. CACHE WITH TTL
   - Response caching with configurable TTL
   - Negative caching for NXDOMAIN responses
   - Cache prefetching before TTL expiry
   - Per-zone cache configuration

5. HEALTH AND READY CHECKS
   - Each plugin reports health status
   - Ready endpoint for Kubernetes probes
   - Aggregated health from all plugins`)

	// ============================================
	// DNS FUNDAMENTALS
	// ============================================
	fmt.Println(`
DNS FUNDAMENTALS (for context)
------------------------------

RECORD TYPES:
  A      - IPv4 address          example.com -> 93.184.216.34
  AAAA   - IPv6 address          example.com -> 2606:2800:220:1:...
  CNAME  - Canonical name alias  www.example.com -> example.com
  MX     - Mail exchange         example.com -> mail.example.com
  TXT    - Text record           example.com -> "v=spf1 ..."
  SRV    - Service locator       _http._tcp.example.com -> ...
  NS     - Name server           example.com -> ns1.example.com
  SOA    - Start of authority    Zone metadata

KUBERNETES DNS:
  <service>.<namespace>.svc.cluster.local -> ClusterIP
  <pod-ip>.<namespace>.pod.cluster.local  -> Pod IP
  _<port>._<proto>.<svc>.<ns>.svc.cluster.local -> SRV record`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN COREDNS
----------------------------------

1. PLUGIN CHAIN SHORT-CIRCUIT
   - Cache plugin responds immediately for cached queries
   - No need to traverse remaining plugins
   - Reduces latency for hot queries

2. UDP CONNECTION REUSE
   - DNS primarily uses UDP (fast, no handshake)
   - Connection pooling for TCP fallback
   - Goroutine per query for parallelism

3. PREFETCH CACHE
   - Proactively refreshes popular entries before TTL
   - Prevents cache stampede on expiry
   - Background goroutine for prefetch

4. AUTOPATH
   - Reduces DNS query count for K8s pods
   - Short-circuits search domain traversal
   - Single query instead of 4-5 for typical lookup

5. MINIMAL RESPONSE
   - Only returns necessary records
   - Reduces packet size for faster transmission
   - Important for DNS over UDP (512 byte limit)`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
HOW COREDNS EXPRESSES GO PHILOSOPHY
-------------------------------------

"Do less, enable more"
  CoreDNS core is tiny. Plugins do all the work.
  Adding functionality = adding a plugin.

"Accept interfaces, return structs"
  Plugin interface is just ServeDNS + Name.
  Any plugin implementing this interface works.
  Simple contract, powerful composition.

"Make the zero value useful"
  Default CoreDNS serves as a simple forwarder.
  Works with minimal Corefile configuration.

"Composition over inheritance"
  Plugin chain is pure composition.
  Complex DNS behavior emerges from simple plugins
  chained together.`)

	// ============================================
	// SIMPLIFIED COREDNS DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED COREDNS-LIKE DNS SERVER ===")

	dns := NewCoreDNS()

	fmt.Println("\n--- Loading Zone Data ---")
	dns.AddZone("cluster.local", map[string][]DNSRecord{
		"web-svc.default.svc.cluster.local.": {
			{Type: "A", Value: "10.96.1.100", TTL: 30},
		},
		"api-svc.default.svc.cluster.local.": {
			{Type: "A", Value: "10.96.1.101", TTL: 30},
		},
		"db-svc.data.svc.cluster.local.": {
			{Type: "A", Value: "10.96.2.50", TTL: 60},
		},
		"cache-svc.data.svc.cluster.local.": {
			{Type: "A", Value: "10.96.2.51", TTL: 30},
			{Type: "A", Value: "10.96.2.52", TTL: 30},
		},
	})

	dns.AddZone("example.com", map[string][]DNSRecord{
		"example.com.": {
			{Type: "A", Value: "93.184.216.34", TTL: 300},
			{Type: "AAAA", Value: "2606:2800:220:1:248:1893:25c8:1946", TTL: 300},
		},
		"www.example.com.": {
			{Type: "CNAME", Value: "example.com.", TTL: 300},
		},
		"mail.example.com.": {
			{Type: "A", Value: "93.184.216.35", TTL: 300},
		},
		"example.com.MX": {
			{Type: "MX", Value: "10 mail.example.com.", TTL: 3600},
		},
	})

	fmt.Println("\n--- Configuring Plugin Chain ---")
	dns.AddPlugin("log", LogPlugin{})
	dns.AddPlugin("cache", NewCachePlugin(30 * time.Second))
	dns.AddPlugin("kubernetes", &KubernetesPlugin{zones: dns.zones})
	dns.AddPlugin("forward", &ForwardPlugin{upstream: "8.8.8.8:53"})

	fmt.Println("\n--- Processing DNS Queries ---")
	queries := []DNSQuery{
		{Name: "web-svc.default.svc.cluster.local.", Type: "A"},
		{Name: "api-svc.default.svc.cluster.local.", Type: "A"},
		{Name: "web-svc.default.svc.cluster.local.", Type: "A"},
		{Name: "cache-svc.data.svc.cluster.local.", Type: "A"},
		{Name: "example.com.", Type: "A"},
		{Name: "www.example.com.", Type: "CNAME"},
		{Name: "nonexistent.cluster.local.", Type: "A"},
		{Name: "google.com.", Type: "A"},
	}

	for _, q := range queries {
		dns.HandleQuery(q)
	}

	fmt.Println("\n--- Cache Statistics ---")
	dns.PrintCacheStats()

	fmt.Println("\n--- Server Statistics ---")
	dns.PrintStats()

	fmt.Println("\n--- Kubernetes Service Watch Simulation ---")
	dns.SimulateK8sEvent("ADDED", "new-svc.default.svc.cluster.local.", "10.96.1.200")
	dns.HandleQuery(DNSQuery{Name: "new-svc.default.svc.cluster.local.", Type: "A"})

	dns.SimulateK8sEvent("DELETED", "api-svc.default.svc.cluster.local.", "")
	dns.HandleQuery(DNSQuery{Name: "api-svc.default.svc.cluster.local.", Type: "A"})

	fmt.Println(`
SUMMARY
-------
CoreDNS is the DNS backbone of Kubernetes.
Its Go codebase demonstrates:
- Plugin chain architecture for composable functionality
- Efficient UDP/TCP network programming
- Kubernetes API watching for service discovery
- Cache with TTL and prefetch optimization
- Zone-based DNS server configuration

CoreDNS proved that Go is ideal for network infrastructure.`)
}

// ============================================
// TYPES
// ============================================

type DNSRecord struct {
	Type  string
	Value string
	TTL   int
}

type DNSQuery struct {
	Name string
	Type string
}

type DNSResponse struct {
	Name    string
	Records []DNSRecord
	Rcode   int
	Cached  bool
}

type CacheEntry struct {
	Response  DNSResponse
	ExpiresAt time.Time
}

// ============================================
// PLUGIN INTERFACES
// ============================================

type DNSPlugin interface {
	Name() string
	ServeDNS(query DNSQuery, next func(DNSQuery) DNSResponse) DNSResponse
}

// ============================================
// LOG PLUGIN
// ============================================

type LogPlugin struct{}

func (l LogPlugin) Name() string { return "log" }

func (l LogPlugin) ServeDNS(query DNSQuery, next func(DNSQuery) DNSResponse) DNSResponse {
	start := time.Now()
	resp := next(query)
	elapsed := time.Since(start)
	rcode := "NOERROR"
	if resp.Rcode == 3 {
		rcode = "NXDOMAIN"
	}
	fmt.Printf("    [log] %s %s -> %s (%v)\n", query.Type, query.Name, rcode, elapsed.Round(time.Microsecond))
	return resp
}

// ============================================
// CACHE PLUGIN
// ============================================

type CachePlugin struct {
	cache map[string]*CacheEntry
	ttl   time.Duration
	hits  int
	misses int
	mu    sync.RWMutex
}

func NewCachePlugin(ttl time.Duration) *CachePlugin {
	return &CachePlugin{
		cache: make(map[string]*CacheEntry),
		ttl:   ttl,
	}
}

func (c *CachePlugin) Name() string { return "cache" }

func (c *CachePlugin) ServeDNS(query DNSQuery, next func(DNSQuery) DNSResponse) DNSResponse {
	key := query.Name + ":" + query.Type

	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if exists && time.Now().Before(entry.ExpiresAt) {
		c.mu.Lock()
		c.hits++
		c.mu.Unlock()
		resp := entry.Response
		resp.Cached = true
		fmt.Printf("    [cache] HIT %s\n", key)
		return resp
	}

	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
	fmt.Printf("    [cache] MISS %s\n", key)

	resp := next(query)

	if resp.Rcode == 0 {
		c.mu.Lock()
		c.cache[key] = &CacheEntry{
			Response:  resp,
			ExpiresAt: time.Now().Add(c.ttl),
		}
		c.mu.Unlock()
	}

	return resp
}

func (c *CachePlugin) Stats() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// ============================================
// KUBERNETES PLUGIN
// ============================================

type KubernetesPlugin struct {
	zones map[string]map[string][]DNSRecord
	mu    sync.RWMutex
}

func (k *KubernetesPlugin) Name() string { return "kubernetes" }

func (k *KubernetesPlugin) ServeDNS(query DNSQuery, next func(DNSQuery) DNSResponse) DNSResponse {
	k.mu.RLock()
	defer k.mu.RUnlock()

	for _, zone := range k.zones {
		if records, exists := zone[query.Name]; exists {
			var matched []DNSRecord
			for _, r := range records {
				if r.Type == query.Type {
					matched = append(matched, r)
				}
			}
			if len(matched) > 0 {
				fmt.Printf("    [kubernetes] FOUND %s (%d records)\n", query.Name, len(matched))
				return DNSResponse{
					Name:    query.Name,
					Records: matched,
					Rcode:   0,
				}
			}
		}
	}

	fmt.Printf("    [kubernetes] NOT FOUND %s, forwarding...\n", query.Name)
	return next(query)
}

func (k *KubernetesPlugin) AddRecord(name string, record DNSRecord) {
	k.mu.Lock()
	defer k.mu.Unlock()

	for _, zone := range k.zones {
		zone[name] = append(zone[name], record)
		return
	}
}

func (k *KubernetesPlugin) DeleteRecord(name string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	for _, zone := range k.zones {
		delete(zone, name)
	}
}

// ============================================
// FORWARD PLUGIN
// ============================================

type ForwardPlugin struct {
	upstream string
}

func (f *ForwardPlugin) Name() string { return "forward" }

func (f *ForwardPlugin) ServeDNS(query DNSQuery, _ func(DNSQuery) DNSResponse) DNSResponse {
	if strings.HasSuffix(query.Name, "cluster.local.") {
		fmt.Printf("    [forward] NXDOMAIN %s (no upstream record)\n", query.Name)
		return DNSResponse{Name: query.Name, Rcode: 3}
	}

	fmt.Printf("    [forward] -> %s for %s\n", f.upstream, query.Name)
	ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(200)+1, rand.Intn(255), rand.Intn(255), rand.Intn(255))
	return DNSResponse{
		Name: query.Name,
		Records: []DNSRecord{
			{Type: "A", Value: ip, TTL: 300},
		},
		Rcode: 0,
	}
}

// ============================================
// COREDNS SERVER
// ============================================

type CoreDNS struct {
	zones   map[string]map[string][]DNSRecord
	plugins []DNSPlugin
	stats   DNSStats
	mu      sync.RWMutex
}

type DNSStats struct {
	TotalQueries int
	ByType       map[string]int
	ByRcode      map[int]int
}

func NewCoreDNS() *CoreDNS {
	return &CoreDNS{
		zones: make(map[string]map[string][]DNSRecord),
		stats: DNSStats{
			ByType:  make(map[string]int),
			ByRcode: make(map[int]int),
		},
	}
}

func (d *CoreDNS) AddZone(name string, records map[string][]DNSRecord) {
	d.zones[name] = records
	total := 0
	for _, recs := range records {
		total += len(recs)
	}
	fmt.Printf("  Zone %q loaded: %d names, %d records\n", name, len(records), total)
}

func (d *CoreDNS) AddPlugin(name string, plugin DNSPlugin) {
	d.plugins = append(d.plugins, plugin)
	fmt.Printf("  Plugin: %s\n", name)
}

func (d *CoreDNS) HandleQuery(query DNSQuery) {
	d.mu.Lock()
	d.stats.TotalQueries++
	d.stats.ByType[query.Type]++
	d.mu.Unlock()

	fmt.Printf("\n  Query: %s %s\n", query.Type, query.Name)

	chain := d.buildChain(0)
	resp := chain(query)

	d.mu.Lock()
	d.stats.ByRcode[resp.Rcode]++
	d.mu.Unlock()

	if len(resp.Records) > 0 {
		for _, rec := range resp.Records {
			cached := ""
			if resp.Cached {
				cached = " (cached)"
			}
			fmt.Printf("  Answer: %s %s TTL=%d%s\n", rec.Type, rec.Value, rec.TTL, cached)
		}
	} else {
		rcode := "NOERROR"
		if resp.Rcode == 3 {
			rcode = "NXDOMAIN"
		}
		fmt.Printf("  Answer: %s (no records)\n", rcode)
	}
}

func (d *CoreDNS) buildChain(idx int) func(DNSQuery) DNSResponse {
	if idx >= len(d.plugins) {
		return func(q DNSQuery) DNSResponse {
			return DNSResponse{Name: q.Name, Rcode: 3}
		}
	}

	plugin := d.plugins[idx]
	next := d.buildChain(idx + 1)

	return func(q DNSQuery) DNSResponse {
		return plugin.ServeDNS(q, next)
	}
}

func (d *CoreDNS) PrintCacheStats() {
	for _, p := range d.plugins {
		if cp, ok := p.(*CachePlugin); ok {
			hits, misses := cp.Stats()
			total := hits + misses
			hitRate := 0.0
			if total > 0 {
				hitRate = float64(hits) / float64(total) * 100
			}
			fmt.Printf("  Cache: %d hits, %d misses (%.1f%% hit rate)\n",
				hits, misses, hitRate)
			fmt.Printf("  Cached entries: %d\n", len(cp.cache))
		}
	}
}

func (d *CoreDNS) PrintStats() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	fmt.Printf("  Total queries: %d\n", d.stats.TotalQueries)
	fmt.Println("  By type:")
	for t, count := range d.stats.ByType {
		fmt.Printf("    %s: %d\n", t, count)
	}
	fmt.Println("  By response code:")
	for code, count := range d.stats.ByRcode {
		name := "NOERROR"
		if code == 3 {
			name = "NXDOMAIN"
		}
		fmt.Printf("    %s(%d): %d\n", name, code, count)
	}
}

func (d *CoreDNS) SimulateK8sEvent(eventType, name, ip string) {
	fmt.Printf("\n  K8s Watch Event: %s %s", eventType, name)
	if ip != "" {
		fmt.Printf(" -> %s", ip)
	}
	fmt.Println()

	for _, p := range d.plugins {
		if kp, ok := p.(*KubernetesPlugin); ok {
			switch eventType {
			case "ADDED":
				kp.AddRecord(name, DNSRecord{Type: "A", Value: ip, TTL: 30})
				fmt.Printf("    Record added to zone\n")
			case "DELETED":
				kp.DeleteRecord(name)
				fmt.Printf("    Record removed from zone\n")
			}
		}
	}
}

func init() {
	_ = os.Stdout
	_ = net.IPv4
}
