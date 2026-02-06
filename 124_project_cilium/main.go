// Package main - Chapter 124: Cilium - eBPF-based Networking for Kubernetes in Go
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 124: CILIUM - EBPF-BASED NETWORKING FOR KUBERNETES IN GO ===")

	// ============================================
	// WHAT IS CILIUM?
	// ============================================
	fmt.Println(`
==============================
WHAT IS CILIUM?
==============================

Cilium is an open-source networking, security, and
observability platform for Kubernetes, powered by eBPF.
Written primarily in Go, it provides:

  - CNI plugin for Kubernetes pod networking
  - Network policy enforcement (L3/L4/L7)
  - Load balancing (replacing kube-proxy)
  - Service mesh (without sidecars)
  - Network observability via Hubble
  - Transparent encryption (WireGuard/IPsec)
  - Multi-cluster connectivity (ClusterMesh)

WHO USES IT:
  - Google (GKE Dataplane V2)
  - AWS (EKS networking)
  - Adobe, Capital One, Datadog
  - Bell Canada, Sky, Meltwater
  - Default CNI in many Kubernetes distributions

WHY GO:
  Cilium chose Go for the control plane because of its
  excellent Kubernetes ecosystem integration, strong
  concurrency primitives for managing eBPF programs
  and network state, and simple deployment as a
  static binary. The data plane uses eBPF/C for
  kernel-level packet processing.`)

	// ============================================
	// ARCHITECTURE OVERVIEW
	// ============================================
	fmt.Println(`
==============================
ARCHITECTURE OVERVIEW
==============================

  +---------------------------------------------------+
  |              Hubble (Observability)                |
  |  (Flow logs, Service map, Metrics)                |
  +---------------------------------------------------+
  |           Cilium Agent (Go)                       |
  |  +-----------+  +-----------+  +-----------+     |
  |  | Policy    |  | Endpoint  |  | Service   |     |
  |  | Engine    |  | Manager   |  | Manager   |     |
  |  +-----------+  +-----------+  +-----------+     |
  |  +-----------+  +-----------+  +-----------+     |
  |  | IPAM      |  | Identity  |  | ClusterMesh|    |
  |  | Manager   |  | Manager   |  | Manager   |     |
  |  +-----------+  +-----------+  +-----------+     |
  +---------------------------------------------------+
  |           eBPF Dataplane (C)                      |
  |  +--------+  +--------+  +--------+  +--------+ |
  |  | tc/XDP |  | Socket |  | Cgroup |  | kprobe | |
  |  | hooks  |  | hooks  |  | hooks  |  | hooks  | |
  |  +--------+  +--------+  +--------+  +--------+ |
  +---------------------------------------------------+
  |              Linux Kernel                         |
  +---------------------------------------------------+

KEY CONCEPTS:
  - Endpoints: network-attached pods
  - Identity: label-based security identity
  - Policy: allow/deny rules based on identities
  - eBPF: kernel programs for fast packet processing
  - Hubble: observability layer for network flows`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
==============================
KEY GO PATTERNS IN CILIUM
==============================

1. CONTROLLER PATTERN
   Cilium uses a controller manager that runs reconciliation
   loops. Each controller watches Kubernetes resources and
   reconciles desired state with actual state.

2. IDENTITY-BASED SECURITY
   Instead of IP-based rules, Cilium assigns cryptographic
   identities to pods based on labels. Policies reference
   identities, not IPs.

3. EBPF MAP MANAGEMENT
   Go code manages eBPF maps (hash maps, LRU, arrays)
   in the kernel. The Go control plane writes policy
   rules; the eBPF data plane enforces them.

4. EVENT-DRIVEN ARCHITECTURE
   Changes in Kubernetes (pod creation, policy updates)
   trigger events that flow through channels to
   appropriate handlers.

5. INFORMER/WATCHER PATTERN
   Uses Kubernetes informers (list+watch) to maintain
   a local cache of cluster state, avoiding expensive
   API server queries.

6. ENDPOINT REGENERATION
   When policies change, affected endpoints are
   "regenerated" - their eBPF programs are recompiled
   and atomically replaced in the kernel.`)

	// ============================================
	// DEMO: SIMPLIFIED CILIUM
	// ============================================
	fmt.Println(`
==============================
DEMO: SIMPLIFIED CILIUM
==============================

This demo implements core Cilium concepts:
  - Identity-based security model
  - Network policy engine
  - Endpoint management
  - Service load balancing
  - eBPF map simulation
  - Flow observability (Hubble-like)`)

	fmt.Println("\n--- Identity Manager ---")
	demoIdentityManager()

	fmt.Println("\n--- Network Policy Engine ---")
	demoPolicyEngine()

	fmt.Println("\n--- Endpoint Manager ---")
	demoEndpointManager()

	fmt.Println("\n--- Service Load Balancer ---")
	demoServiceLB()

	fmt.Println("\n--- eBPF Map Simulation ---")
	demoEBPFMaps()

	fmt.Println("\n--- Full Mini-Cilium ---")
	demoMiniCilium()

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
==============================
PERFORMANCE TECHNIQUES
==============================

1. EBPF FOR DATA PLANE
   Packet processing happens in the Linux kernel via
   eBPF programs - no context switches to userspace.
   This achieves near-line-rate performance.

2. XDP (EXPRESS DATA PATH)
   For load balancing, Cilium can attach eBPF programs
   at the earliest point in the network stack (XDP),
   processing packets before they even reach the kernel
   networking stack.

3. CONNECTION TRACKING IN eBPF
   Conntrack is implemented in eBPF maps, avoiding
   the overhead of the kernel's netfilter conntrack.

4. BATCH MAP OPERATIONS
   Go control plane batches eBPF map updates to
   minimize syscall overhead.

5. LOCK-FREE eBPF MAPS
   Per-CPU maps and RCU-protected data structures
   in the eBPF dataplane avoid lock contention.

6. IDENTITY CACHING
   Numeric identity lookups are cached in eBPF maps,
   avoiding userspace roundtrips for policy decisions.`)

	// ============================================
	// GO PHILOSOPHY IN CILIUM
	// ============================================
	fmt.Println(`
==============================
GO PHILOSOPHY IN CILIUM
==============================

CLEAR SEPARATION OF CONCERNS:
  Go handles the control plane (configuration, policy
  computation, Kubernetes integration). C/eBPF handles
  the data plane (packet processing). Each language
  is used where it excels.

INTERFACES FOR EXTENSIBILITY:
  The datapath interface allows different eBPF program
  implementations. Testing uses mock datapaths that
  don't require a real kernel.

GOROUTINES FOR CONTROLLERS:
  Each reconciliation loop runs in its own goroutine.
  The Go runtime efficiently multiplexes hundreds of
  controllers across a few OS threads.

ERROR HANDLING WITH CONTEXT:
  Cilium wraps errors with context at each layer,
  creating detailed error chains essential for
  debugging distributed networking issues.

KUBERNETES-NATIVE:
  Cilium deeply integrates with Kubernetes using Go's
  client-go library, informers, and custom resource
  definitions (CRDs) - all idiomatic Go patterns.`)

	fmt.Println("\n=== END OF CHAPTER 124 ===")
}

// ============================================
// IDENTITY MANAGER
// ============================================

type SecurityIdentity struct {
	ID     uint32
	Labels map[string]string
}

func (si SecurityIdentity) String() string {
	var parts []string
	for k, v := range si.Labels {
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	return fmt.Sprintf("ID=%d [%s]", si.ID, strings.Join(parts, ", "))
}

func (si SecurityIdentity) MatchesSelector(selector map[string]string) bool {
	for k, v := range selector {
		if si.Labels[k] != v {
			return false
		}
	}
	return true
}

type IdentityManager struct {
	mu         sync.RWMutex
	identities map[uint32]*SecurityIdentity
	labelIndex map[string]uint32
	nextID     uint32
}

func NewIdentityManager() *IdentityManager {
	return &IdentityManager{
		identities: make(map[uint32]*SecurityIdentity),
		labelIndex: make(map[string]uint32),
		nextID:     1000,
	}
}

func (im *IdentityManager) labelKey(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (im *IdentityManager) AllocateIdentity(labels map[string]string) *SecurityIdentity {
	im.mu.Lock()
	defer im.mu.Unlock()

	key := im.labelKey(labels)
	if id, ok := im.labelIndex[key]; ok {
		return im.identities[id]
	}

	id := im.nextID
	im.nextID++

	identity := &SecurityIdentity{
		ID:     id,
		Labels: labels,
	}

	im.identities[id] = identity
	im.labelIndex[key] = id
	return identity
}

func (im *IdentityManager) GetIdentity(id uint32) *SecurityIdentity {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.identities[id]
}

func (im *IdentityManager) FindByLabels(labels map[string]string) *SecurityIdentity {
	im.mu.RLock()
	defer im.mu.RUnlock()

	key := im.labelKey(labels)
	if id, ok := im.labelIndex[key]; ok {
		return im.identities[id]
	}
	return nil
}

func demoIdentityManager() {
	im := NewIdentityManager()

	frontend := im.AllocateIdentity(map[string]string{
		"app": "frontend", "env": "production",
	})
	backend := im.AllocateIdentity(map[string]string{
		"app": "backend", "env": "production",
	})
	db := im.AllocateIdentity(map[string]string{
		"app": "database", "env": "production",
	})

	fmt.Printf("  Frontend: %s\n", frontend)
	fmt.Printf("  Backend:  %s\n", backend)
	fmt.Printf("  Database: %s\n", db)

	sameLabels := im.AllocateIdentity(map[string]string{
		"app": "frontend", "env": "production",
	})
	fmt.Printf("  Same labels reuse ID: %v (frontend.ID=%d, realloc.ID=%d)\n",
		frontend.ID == sameLabels.ID, frontend.ID, sameLabels.ID)

	matches := frontend.MatchesSelector(map[string]string{"env": "production"})
	fmt.Printf("  Frontend matches {env=production}: %v\n", matches)
}

// ============================================
// NETWORK POLICY ENGINE
// ============================================

type NetworkPolicy struct {
	Name         string
	Namespace    string
	EndpointSelector map[string]string
	IngressRules []PolicyRule
	EgressRules  []PolicyRule
}

type PolicyRule struct {
	FromEndpoints []map[string]string
	ToEndpoints   []map[string]string
	ToPorts       []PortRule
}

type PortRule struct {
	Port     int
	Protocol string
}

type PolicyEngine struct {
	mu       sync.RWMutex
	policies map[string]*NetworkPolicy
}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: make(map[string]*NetworkPolicy),
	}
}

func (pe *PolicyEngine) AddPolicy(policy *NetworkPolicy) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.policies[policy.Namespace+"/"+policy.Name] = policy
}

func (pe *PolicyEngine) DeletePolicy(namespace, name string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	delete(pe.policies, namespace+"/"+name)
}

type PolicyVerdict struct {
	Allowed bool
	Policy  string
	Reason  string
}

func (pe *PolicyEngine) EvaluateIngress(destIdentity, srcIdentity *SecurityIdentity, port int) PolicyVerdict {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	matchingPolicies := 0

	for _, policy := range pe.policies {
		if !destIdentity.MatchesSelector(policy.EndpointSelector) {
			continue
		}
		matchingPolicies++

		for _, rule := range policy.IngressRules {
			srcAllowed := len(rule.FromEndpoints) == 0
			for _, selector := range rule.FromEndpoints {
				if srcIdentity.MatchesSelector(selector) {
					srcAllowed = true
					break
				}
			}

			portAllowed := len(rule.ToPorts) == 0
			for _, pr := range rule.ToPorts {
				if pr.Port == port {
					portAllowed = true
					break
				}
			}

			if srcAllowed && portAllowed {
				return PolicyVerdict{
					Allowed: true,
					Policy:  policy.Name,
					Reason:  "explicitly allowed by policy",
				}
			}
		}
	}

	if matchingPolicies == 0 {
		return PolicyVerdict{
			Allowed: true,
			Reason:  "no matching policy (default allow)",
		}
	}

	return PolicyVerdict{
		Allowed: false,
		Reason:  "denied by default (policy exists but no matching rule)",
	}
}

func demoPolicyEngine() {
	im := NewIdentityManager()
	pe := NewPolicyEngine()

	frontend := im.AllocateIdentity(map[string]string{"app": "frontend"})
	backend := im.AllocateIdentity(map[string]string{"app": "backend"})
	database := im.AllocateIdentity(map[string]string{"app": "database"})
	external := im.AllocateIdentity(map[string]string{"app": "external"})

	pe.AddPolicy(&NetworkPolicy{
		Name:             "allow-backend-ingress",
		Namespace:        "default",
		EndpointSelector: map[string]string{"app": "backend"},
		IngressRules: []PolicyRule{
			{
				FromEndpoints: []map[string]string{
					{"app": "frontend"},
				},
				ToPorts: []PortRule{
					{Port: 8080, Protocol: "TCP"},
				},
			},
		},
	})

	pe.AddPolicy(&NetworkPolicy{
		Name:             "allow-db-ingress",
		Namespace:        "default",
		EndpointSelector: map[string]string{"app": "database"},
		IngressRules: []PolicyRule{
			{
				FromEndpoints: []map[string]string{
					{"app": "backend"},
				},
				ToPorts: []PortRule{
					{Port: 5432, Protocol: "TCP"},
				},
			},
		},
	})

	tests := []struct {
		desc string
		dst  *SecurityIdentity
		src  *SecurityIdentity
		port int
	}{
		{"frontend -> backend:8080", backend, frontend, 8080},
		{"frontend -> backend:9090", backend, frontend, 9090},
		{"external -> backend:8080", backend, external, 8080},
		{"backend -> database:5432", database, backend, 5432},
		{"frontend -> database:5432", database, frontend, 5432},
	}

	for _, t := range tests {
		verdict := pe.EvaluateIngress(t.dst, t.src, t.port)
		allowed := "ALLOW"
		if !verdict.Allowed {
			allowed = "DENY"
		}
		fmt.Printf("  %s: %s (%s)\n", t.desc, allowed, verdict.Reason)
	}
}

// ============================================
// ENDPOINT MANAGER
// ============================================

type EndpointState string

const (
	EndpointCreating    EndpointState = "creating"
	EndpointReady       EndpointState = "ready"
	EndpointRegenerating EndpointState = "regenerating"
	EndpointDisconnected EndpointState = "disconnected"
)

type Endpoint struct {
	mu         sync.RWMutex
	ID         uint16
	PodName    string
	Namespace  string
	IPv4       net.IP
	Labels     map[string]string
	Identity   *SecurityIdentity
	State      EndpointState
	PolicyMap  map[string]bool
	Created    time.Time
}

type EndpointManager struct {
	mu        sync.RWMutex
	endpoints map[uint16]*Endpoint
	nextID    uint16
}

func NewEndpointManager() *EndpointManager {
	return &EndpointManager{
		endpoints: make(map[uint16]*Endpoint),
		nextID:    1,
	}
}

func (em *EndpointManager) CreateEndpoint(podName, namespace string, ipv4 net.IP, labels map[string]string, identity *SecurityIdentity) *Endpoint {
	em.mu.Lock()
	defer em.mu.Unlock()

	ep := &Endpoint{
		ID:        em.nextID,
		PodName:   podName,
		Namespace: namespace,
		IPv4:      ipv4,
		Labels:    labels,
		Identity:  identity,
		State:     EndpointCreating,
		PolicyMap: make(map[string]bool),
		Created:   time.Now(),
	}
	em.nextID++
	em.endpoints[ep.ID] = ep
	return ep
}

func (em *EndpointManager) Regenerate(ep *Endpoint, pe *PolicyEngine) {
	ep.mu.Lock()
	ep.State = EndpointRegenerating
	ep.PolicyMap = make(map[string]bool)

	pe.mu.RLock()
	for _, policy := range pe.policies {
		if ep.Identity.MatchesSelector(policy.EndpointSelector) {
			ep.PolicyMap[policy.Name] = true
		}
	}
	pe.mu.RUnlock()

	ep.State = EndpointReady
	ep.mu.Unlock()
}

func (em *EndpointManager) GetEndpoint(id uint16) *Endpoint {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return em.endpoints[id]
}

func (em *EndpointManager) ListEndpoints() []*Endpoint {
	em.mu.RLock()
	defer em.mu.RUnlock()

	var eps []*Endpoint
	for _, ep := range em.endpoints {
		eps = append(eps, ep)
	}
	sort.Slice(eps, func(i, j int) bool {
		return eps[i].ID < eps[j].ID
	})
	return eps
}

func demoEndpointManager() {
	im := NewIdentityManager()
	pe := NewPolicyEngine()
	em := NewEndpointManager()

	pe.AddPolicy(&NetworkPolicy{
		Name:             "backend-policy",
		Namespace:        "default",
		EndpointSelector: map[string]string{"app": "backend"},
		IngressRules:     []PolicyRule{{FromEndpoints: []map[string]string{{"app": "frontend"}}}},
	})

	pods := []struct {
		name, ns string
		ip       string
		labels   map[string]string
	}{
		{"frontend-abc", "default", "10.0.0.1", map[string]string{"app": "frontend"}},
		{"backend-xyz", "default", "10.0.0.2", map[string]string{"app": "backend"}},
		{"backend-qrs", "default", "10.0.0.3", map[string]string{"app": "backend"}},
	}

	for _, p := range pods {
		identity := im.AllocateIdentity(p.labels)
		ep := em.CreateEndpoint(p.name, p.ns, net.ParseIP(p.ip), p.labels, identity)
		em.Regenerate(ep, pe)

		ep.mu.RLock()
		fmt.Printf("  Endpoint %d: %s (%s) identity=%d state=%s policies=%v\n",
			ep.ID, ep.PodName, ep.IPv4, ep.Identity.ID, ep.State, ep.PolicyMap)
		ep.mu.RUnlock()
	}
}

// ============================================
// SERVICE LOAD BALANCER
// ============================================

type Service struct {
	mu        sync.RWMutex
	Name      string
	Namespace string
	ClusterIP net.IP
	Port      int
	Backends  []Backend
	counter   uint64
}

type Backend struct {
	IP     net.IP
	Port   int
	Weight int
	Active bool
}

type ServiceManager struct {
	mu       sync.RWMutex
	services map[string]*Service
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		services: make(map[string]*Service),
	}
}

func (sm *ServiceManager) UpsertService(name, namespace string, clusterIP net.IP, port int, backends []Backend) *Service {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := namespace + "/" + name
	svc := &Service{
		Name:      name,
		Namespace: namespace,
		ClusterIP: clusterIP,
		Port:      port,
		Backends:  backends,
	}
	sm.services[key] = svc
	return svc
}

func (svc *Service) SelectBackend() *Backend {
	svc.mu.RLock()
	defer svc.mu.RUnlock()

	var active []Backend
	for _, b := range svc.Backends {
		if b.Active {
			active = append(active, b)
		}
	}
	if len(active) == 0 {
		return nil
	}

	idx := atomic.AddUint64(&svc.counter, 1) % uint64(len(active))
	return &active[idx]
}

func (sm *ServiceManager) GetService(namespace, name string) *Service {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.services[namespace+"/"+name]
}

func demoServiceLB() {
	sm := NewServiceManager()

	svc := sm.UpsertService("api-gateway", "default",
		net.ParseIP("10.96.0.100"), 80,
		[]Backend{
			{IP: net.ParseIP("10.0.1.1"), Port: 8080, Weight: 1, Active: true},
			{IP: net.ParseIP("10.0.1.2"), Port: 8080, Weight: 1, Active: true},
			{IP: net.ParseIP("10.0.1.3"), Port: 8080, Weight: 1, Active: false},
		})
	fmt.Printf("  Service: %s/%s -> %s:%d (%d backends)\n",
		svc.Namespace, svc.Name, svc.ClusterIP, svc.Port, len(svc.Backends))

	distribution := make(map[string]int)
	for i := 0; i < 10; i++ {
		backend := svc.SelectBackend()
		if backend != nil {
			key := backend.IP.String()
			distribution[key]++
		}
	}
	fmt.Println("  Load distribution (10 requests):")
	for ip, count := range distribution {
		fmt.Printf("    %s: %d requests\n", ip, count)
	}
}

// ============================================
// EBPF MAP SIMULATION
// ============================================

type BPFMapType int

const (
	BPFMapHash BPFMapType = iota
	BPFMapArray
	BPFMapLRUHash
	BPFMapPerCPU
)

func (t BPFMapType) String() string {
	switch t {
	case BPFMapHash:
		return "HASH"
	case BPFMapArray:
		return "ARRAY"
	case BPFMapLRUHash:
		return "LRU_HASH"
	case BPFMapPerCPU:
		return "PERCPU_HASH"
	default:
		return "UNKNOWN"
	}
}

type BPFMap struct {
	mu         sync.RWMutex
	Name       string
	Type       BPFMapType
	MaxEntries int
	entries    map[string][]byte
	accessOrder []string
}

func NewBPFMap(name string, mapType BPFMapType, maxEntries int) *BPFMap {
	return &BPFMap{
		Name:       name,
		Type:       mapType,
		MaxEntries: maxEntries,
		entries:    make(map[string][]byte),
	}
}

func (m *BPFMap) Update(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.entries[key]; !exists && len(m.entries) >= m.MaxEntries {
		if m.Type == BPFMapLRUHash && len(m.accessOrder) > 0 {
			evictKey := m.accessOrder[0]
			m.accessOrder = m.accessOrder[1:]
			delete(m.entries, evictKey)
		} else {
			return fmt.Errorf("map full: %s (max=%d)", m.Name, m.MaxEntries)
		}
	}

	m.entries[key] = value
	m.accessOrder = append(m.accessOrder, key)
	return nil
}

func (m *BPFMap) Lookup(key string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.entries[key]
	return val, ok
}

func (m *BPFMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.entries, key)
}

func (m *BPFMap) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

type BPFMapManager struct {
	mu   sync.RWMutex
	maps map[string]*BPFMap
}

func NewBPFMapManager() *BPFMapManager {
	return &BPFMapManager{
		maps: make(map[string]*BPFMap),
	}
}

func (mm *BPFMapManager) CreateMap(name string, mapType BPFMapType, maxEntries int) *BPFMap {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	m := NewBPFMap(name, mapType, maxEntries)
	mm.maps[name] = m
	return m
}

func (mm *BPFMapManager) GetMap(name string) *BPFMap {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.maps[name]
}

func demoEBPFMaps() {
	mm := NewBPFMapManager()

	ctMap := mm.CreateMap("cilium_ct4_global", BPFMapHash, 1000)
	policyMap := mm.CreateMap("cilium_policy", BPFMapHash, 100)
	lruMap := mm.CreateMap("cilium_lru_cache", BPFMapLRUHash, 3)

	connEntries := []struct {
		key string
		val map[string]interface{}
	}{
		{"10.0.0.1:8080->10.0.0.2:80", map[string]interface{}{"state": "established", "packets": 42}},
		{"10.0.0.3:9090->10.0.0.4:443", map[string]interface{}{"state": "syn_sent", "packets": 1}},
	}
	for _, e := range connEntries {
		data, _ := json.Marshal(e.val)
		ctMap.Update(e.key, data)
	}
	fmt.Printf("  CT map entries: %d\n", ctMap.Len())

	policyEntries := []struct {
		key string
		val string
	}{
		{"1000->1001:8080/TCP", "ALLOW"},
		{"1000->1002:5432/TCP", "DENY"},
	}
	for _, e := range policyEntries {
		policyMap.Update(e.key, []byte(e.val))
	}
	if val, ok := policyMap.Lookup("1000->1001:8080/TCP"); ok {
		fmt.Printf("  Policy lookup 1000->1001:8080/TCP: %s\n", val)
	}

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("cache-key-%d", i)
		lruMap.Update(key, []byte(fmt.Sprintf("value-%d", i)))
	}
	fmt.Printf("  LRU map (max=3) after 5 inserts: %d entries\n", lruMap.Len())
}

// ============================================
// FLOW OBSERVABILITY (HUBBLE-LIKE)
// ============================================

type Flow struct {
	Timestamp time.Time
	SrcIP     string
	DstIP     string
	SrcPort   int
	DstPort   int
	Protocol  string
	Verdict   string
	Type      string
	Summary   string
}

type FlowObserver struct {
	mu    sync.Mutex
	flows []Flow
	max   int
}

func NewFlowObserver(maxFlows int) *FlowObserver {
	return &FlowObserver{max: maxFlows}
}

func (fo *FlowObserver) RecordFlow(flow Flow) {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	flow.Timestamp = time.Now()
	fo.flows = append(fo.flows, flow)
	if len(fo.flows) > fo.max {
		fo.flows = fo.flows[1:]
	}
}

func (fo *FlowObserver) GetFlows(filter func(Flow) bool) []Flow {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	var result []Flow
	for _, f := range fo.flows {
		if filter == nil || filter(f) {
			result = append(result, f)
		}
	}
	return result
}

// ============================================
// FULL MINI-CILIUM
// ============================================

type MiniCilium struct {
	identityMgr  *IdentityManager
	policyEngine *PolicyEngine
	endpointMgr  *EndpointManager
	serviceMgr   *ServiceManager
	bpfMaps      *BPFMapManager
	observer     *FlowObserver
}

func NewMiniCilium() *MiniCilium {
	return &MiniCilium{
		identityMgr:  NewIdentityManager(),
		policyEngine: NewPolicyEngine(),
		endpointMgr:  NewEndpointManager(),
		serviceMgr:   NewServiceManager(),
		bpfMaps:      NewBPFMapManager(),
		observer:     NewFlowObserver(1000),
	}
}

func (c *MiniCilium) OnPodCreate(podName, namespace string, ip net.IP, labels map[string]string) *Endpoint {
	identity := c.identityMgr.AllocateIdentity(labels)
	ep := c.endpointMgr.CreateEndpoint(podName, namespace, ip, labels, identity)
	c.endpointMgr.Regenerate(ep, c.policyEngine)
	return ep
}

func (c *MiniCilium) OnPolicyCreate(policy *NetworkPolicy) {
	c.policyEngine.AddPolicy(policy)

	for _, ep := range c.endpointMgr.ListEndpoints() {
		c.endpointMgr.Regenerate(ep, c.policyEngine)
	}
}

func (c *MiniCilium) SimulateTraffic(srcIP, dstIP string, srcPort, dstPort int, protocol string) string {
	var srcIdentity, dstIdentity *SecurityIdentity

	for _, ep := range c.endpointMgr.ListEndpoints() {
		ep.mu.RLock()
		if ep.IPv4.String() == srcIP {
			srcIdentity = ep.Identity
		}
		if ep.IPv4.String() == dstIP {
			dstIdentity = ep.Identity
		}
		ep.mu.RUnlock()
	}

	verdict := "ALLOW"
	reason := "no policy"
	if srcIdentity != nil && dstIdentity != nil {
		pv := c.policyEngine.EvaluateIngress(dstIdentity, srcIdentity, dstPort)
		if pv.Allowed {
			verdict = "ALLOW"
		} else {
			verdict = "DROP"
		}
		reason = pv.Reason
	}

	c.observer.RecordFlow(Flow{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		SrcPort:  srcPort,
		DstPort:  dstPort,
		Protocol: protocol,
		Verdict:  verdict,
		Type:     "L4",
		Summary:  reason,
	})

	return verdict
}

func demoMiniCilium() {
	cilium := NewMiniCilium()

	cilium.OnPolicyCreate(&NetworkPolicy{
		Name:             "frontend-to-backend",
		Namespace:        "production",
		EndpointSelector: map[string]string{"app": "api"},
		IngressRules: []PolicyRule{
			{
				FromEndpoints: []map[string]string{
					{"app": "web"},
				},
				ToPorts: []PortRule{
					{Port: 8080, Protocol: "TCP"},
				},
			},
		},
	})

	cilium.OnPolicyCreate(&NetworkPolicy{
		Name:             "api-to-db",
		Namespace:        "production",
		EndpointSelector: map[string]string{"app": "postgres"},
		IngressRules: []PolicyRule{
			{
				FromEndpoints: []map[string]string{
					{"app": "api"},
				},
				ToPorts: []PortRule{
					{Port: 5432, Protocol: "TCP"},
				},
			},
		},
	})

	pods := []struct {
		name   string
		ip     string
		labels map[string]string
	}{
		{"web-frontend-1", "10.0.1.1", map[string]string{"app": "web"}},
		{"web-frontend-2", "10.0.1.2", map[string]string{"app": "web"}},
		{"api-server-1", "10.0.2.1", map[string]string{"app": "api"}},
		{"api-server-2", "10.0.2.2", map[string]string{"app": "api"}},
		{"postgres-primary", "10.0.3.1", map[string]string{"app": "postgres"}},
	}

	for _, p := range pods {
		ep := cilium.OnPodCreate(p.name, "production", net.ParseIP(p.ip), p.labels)
		ep.mu.RLock()
		fmt.Printf("  Pod %s (%s) -> identity=%d, state=%s\n",
			p.name, p.ip, ep.Identity.ID, ep.State)
		ep.mu.RUnlock()
	}

	svc := cilium.serviceMgr.UpsertService("api-gateway", "production",
		net.ParseIP("10.96.0.1"), 80,
		[]Backend{
			{IP: net.ParseIP("10.0.2.1"), Port: 8080, Weight: 1, Active: true},
			{IP: net.ParseIP("10.0.2.2"), Port: 8080, Weight: 1, Active: true},
		})
	fmt.Printf("  Service: %s -> %s:%d\n", svc.Name, svc.ClusterIP, svc.Port)

	trafficTests := []struct {
		desc                    string
		srcIP, dstIP            string
		srcPort, dstPort        int
	}{
		{"web -> api:8080", "10.0.1.1", "10.0.2.1", 45000, 8080},
		{"web -> api:9090", "10.0.1.1", "10.0.2.1", 45001, 9090},
		{"api -> postgres:5432", "10.0.2.1", "10.0.3.1", 45002, 5432},
		{"web -> postgres:5432", "10.0.1.1", "10.0.3.1", 45003, 5432},
	}

	fmt.Println("\n  Traffic simulation:")
	for _, t := range trafficTests {
		verdict := cilium.SimulateTraffic(t.srcIP, t.dstIP, t.srcPort, t.dstPort, "TCP")
		fmt.Printf("    %s: %s\n", t.desc, verdict)
	}

	flows := cilium.observer.GetFlows(func(f Flow) bool {
		return f.Verdict == "DROP"
	})
	fmt.Printf("\n  Dropped flows: %d\n", len(flows))
	for _, f := range flows {
		fmt.Printf("    %s:%d -> %s:%d [%s] %s\n",
			f.SrcIP, f.SrcPort, f.DstIP, f.DstPort, f.Verdict, f.Summary)
	}

	allFlows := cilium.observer.GetFlows(nil)
	fmt.Printf("  Total observed flows: %d\n", len(allFlows))
}
