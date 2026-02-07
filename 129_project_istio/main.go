// Package main - Chapter 129: Istio Internals
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 129: ISTIO - SERVICE MESH CONTROL PLANE IN GO ===")

	// ============================================
	// WHAT IS ISTIO?
	// ============================================
	fmt.Println(`
==============================
WHAT IS ISTIO?
==============================

Istio is the most popular service mesh for Kubernetes, written in Go.
It provides traffic management, security, and observability for
microservices without changing application code.

WHAT IT DOES:
  - Traffic management (routing, load balancing, retries, timeouts)
  - Security (mTLS between services, authorization policies)
  - Observability (metrics, tracing, logging)
  - Policy enforcement (rate limiting, access control)

ARCHITECTURE:
  +------------------+     +------------------+
  |   Service A      |     |   Service B      |
  |  +------------+  |     |  +------------+  |
  |  | App Code   |  |     |  | App Code   |  |
  |  +-----+------+  |     |  +-----+------+  |
  |  | Envoy      |<----------| Envoy      |  |
  |  | Sidecar    |  | mTLS|  | Sidecar    |  |
  |  +-----+------+  |     |  +-----+------+  |
  +--------+---------+     +--------+---------+
           |         xDS API         |
           +----------+--------------+
                      |
              +-------v--------+
              |    istiod       |
              | (Control Plane) |
              +----------------+
              | - Pilot (config)|
              | - Citadel (certs)|
              | - Galley (valid)|
              +----------------+

KEY CONCEPTS:
  istiod:       Unified control plane (Pilot + Citadel + Galley)
  Envoy:        Data plane proxy (sidecar in each pod)
  xDS API:      Discovery Service protocol (gRPC streaming)
  VirtualService: Traffic routing rules
  DestinationRule: Load balancing and circuit breaking
  Gateway:       Ingress/egress traffic management

WHY GO:
  - gRPC native support for xDS streaming
  - Kubernetes client-go for cluster integration
  - Goroutines for watching 1000s of resources
  - Strong concurrency for reconciliation loops
  - Fast startup for sidecar injection webhook`)

	// ============================================
	// SERVICE REGISTRY
	// ============================================
	fmt.Println("\n--- Service Registry ---")
	demoServiceRegistry()

	// ============================================
	// xDS CONFIG SERVER
	// ============================================
	fmt.Println("\n--- xDS Configuration Server ---")
	demoXDSServer()

	// ============================================
	// mTLS CERTIFICATE MANAGEMENT
	// ============================================
	fmt.Println("\n--- mTLS Certificate Manager ---")
	demoCertificateManager()

	// ============================================
	// TRAFFIC ROUTING ENGINE
	// ============================================
	fmt.Println("\n--- Traffic Routing Engine ---")
	demoTrafficRouting()

	// ============================================
	// ENVOY SIDECAR PROXY
	// ============================================
	fmt.Println("\n--- Envoy Sidecar Simulation ---")
	demoEnvoySidecar()

	// ============================================
	// FULL MINI-ISTIO
	// ============================================
	fmt.Println("\n--- Mini-Istio Integration ---")
	demoMiniIstio()
}

// ============================================
// SERVICE REGISTRY
// ============================================

type ServiceInstance struct {
	ID        string
	Service   string
	Address   string
	Port      int
	Labels    map[string]string
	Healthy   bool
	Version   string
	Namespace string
}

type ServiceRegistry struct {
	mu        sync.RWMutex
	services  map[string][]*ServiceInstance
	watchers  map[string][]chan ServiceEvent
}

type ServiceEvent struct {
	Type     string
	Instance *ServiceInstance
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string][]*ServiceInstance),
		watchers: make(map[string][]chan ServiceEvent),
	}
}

func (r *ServiceRegistry) Register(inst *ServiceInstance) {
	r.mu.Lock()
	r.services[inst.Service] = append(r.services[inst.Service], inst)

	watchers := r.watchers[inst.Service]
	r.mu.Unlock()

	event := ServiceEvent{Type: "ADDED", Instance: inst}
	for _, ch := range watchers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (r *ServiceRegistry) Deregister(serviceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for svc, instances := range r.services {
		for i, inst := range instances {
			if inst.ID == serviceID {
				r.services[svc] = append(instances[:i], instances[i+1:]...)
				return
			}
		}
	}
}

func (r *ServiceRegistry) GetInstances(service string) []*ServiceInstance {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*ServiceInstance
	for _, inst := range r.services[service] {
		if inst.Healthy {
			result = append(result, inst)
		}
	}
	return result
}

func (r *ServiceRegistry) Watch(service string) <-chan ServiceEvent {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan ServiceEvent, 10)
	r.watchers[service] = append(r.watchers[service], ch)
	return ch
}

func (r *ServiceRegistry) ListServices() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []string
	for svc := range r.services {
		services = append(services, svc)
	}
	sort.Strings(services)
	return services
}

func demoServiceRegistry() {
	registry := NewServiceRegistry()

	watcher := registry.Watch("payment-service")

	instances := []*ServiceInstance{
		{ID: "pay-1", Service: "payment-service", Address: "10.0.0.1", Port: 8080, Healthy: true, Version: "v1", Namespace: "default", Labels: map[string]string{"version": "v1", "env": "prod"}},
		{ID: "pay-2", Service: "payment-service", Address: "10.0.0.2", Port: 8080, Healthy: true, Version: "v2", Namespace: "default", Labels: map[string]string{"version": "v2", "env": "prod"}},
		{ID: "pay-3", Service: "payment-service", Address: "10.0.0.3", Port: 8080, Healthy: false, Version: "v1", Namespace: "default", Labels: map[string]string{"version": "v1", "env": "prod"}},
		{ID: "order-1", Service: "order-service", Address: "10.0.1.1", Port: 8081, Healthy: true, Version: "v1", Namespace: "default", Labels: map[string]string{"version": "v1"}},
		{ID: "order-2", Service: "order-service", Address: "10.0.1.2", Port: 8081, Healthy: true, Version: "v1", Namespace: "default", Labels: map[string]string{"version": "v1"}},
	}

	for _, inst := range instances {
		registry.Register(inst)
	}

	healthy := registry.GetInstances("payment-service")
	os.Stdout.WriteString(fmt.Sprintf("  Healthy payment instances: %d\n", len(healthy)))
	for _, inst := range healthy {
		os.Stdout.WriteString(fmt.Sprintf("    %s -> %s:%d (version=%s)\n", inst.ID, inst.Address, inst.Port, inst.Version))
	}

	services := registry.ListServices()
	os.Stdout.WriteString(fmt.Sprintf("  Registered services: %v\n", services))

	eventCount := 0
	for {
		select {
		case <-watcher:
			eventCount++
		default:
			goto done
		}
	}
done:
	os.Stdout.WriteString(fmt.Sprintf("  Watcher received %d events\n", eventCount))
}

// ============================================
// xDS CONFIGURATION SERVER
// ============================================

type XDSResourceType string

const (
	ClusterResource  XDSResourceType = "type.googleapis.com/envoy.config.cluster.v3.Cluster"
	ListenerResource XDSResourceType = "type.googleapis.com/envoy.config.listener.v3.Listener"
	RouteResource    XDSResourceType = "type.googleapis.com/envoy.config.route.v3.RouteConfiguration"
	EndpointResource XDSResourceType = "type.googleapis.com/envoy.api.v2.ClusterLoadAssignment"
)

type XDSResource struct {
	Name    string
	Type    XDSResourceType
	Version string
	Config  map[string]interface{}
}

type DiscoveryRequest struct {
	VersionInfo   string
	Node          string
	ResourceNames []string
	TypeURL       XDSResourceType
}

type DiscoveryResponse struct {
	VersionInfo string
	Resources   []XDSResource
	TypeURL     XDSResourceType
	Nonce       string
}

type XDSServer struct {
	mu        sync.RWMutex
	resources map[XDSResourceType]map[string]*XDSResource
	version   int
	pushCh    chan XDSResourceType
}

func NewXDSServer() *XDSServer {
	return &XDSServer{
		resources: map[XDSResourceType]map[string]*XDSResource{
			ClusterResource:  {},
			ListenerResource: {},
			RouteResource:    {},
			EndpointResource: {},
		},
		pushCh: make(chan XDSResourceType, 100),
	}
}

func (s *XDSServer) SetResource(res XDSResource) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.version++
	res.Version = fmt.Sprintf("v%d", s.version)

	if _, ok := s.resources[res.Type]; !ok {
		s.resources[res.Type] = make(map[string]*XDSResource)
	}
	s.resources[res.Type][res.Name] = &res

	select {
	case s.pushCh <- res.Type:
	default:
	}
}

func (s *XDSServer) HandleDiscovery(req DiscoveryRequest) DiscoveryResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var resources []XDSResource
	resourceMap := s.resources[req.TypeURL]

	if len(req.ResourceNames) == 0 {
		for _, res := range resourceMap {
			resources = append(resources, *res)
		}
	} else {
		for _, name := range req.ResourceNames {
			if res, ok := resourceMap[name]; ok {
				resources = append(resources, *res)
			}
		}
	}

	nonce := fmt.Sprintf("%d", time.Now().UnixNano())

	return DiscoveryResponse{
		VersionInfo: fmt.Sprintf("v%d", s.version),
		Resources:   resources,
		TypeURL:     req.TypeURL,
		Nonce:       nonce,
	}
}

func demoXDSServer() {
	xds := NewXDSServer()

	xds.SetResource(XDSResource{
		Name: "payment-service",
		Type: ClusterResource,
		Config: map[string]interface{}{
			"name":              "payment-service",
			"type":              "EDS",
			"lb_policy":         "ROUND_ROBIN",
			"connect_timeout":   "5s",
			"circuit_breakers":  map[string]int{"max_connections": 1000},
		},
	})

	xds.SetResource(XDSResource{
		Name: "order-service",
		Type: ClusterResource,
		Config: map[string]interface{}{
			"name":            "order-service",
			"type":            "EDS",
			"lb_policy":       "LEAST_REQUEST",
			"connect_timeout": "3s",
		},
	})

	xds.SetResource(XDSResource{
		Name: "0.0.0.0:8080",
		Type: ListenerResource,
		Config: map[string]interface{}{
			"name":    "inbound:8080",
			"address": "0.0.0.0:8080",
			"filter_chains": []string{"http_connection_manager"},
		},
	})

	xds.SetResource(XDSResource{
		Name: "default-route",
		Type: RouteResource,
		Config: map[string]interface{}{
			"name": "default-route",
			"virtual_hosts": []map[string]interface{}{
				{
					"name":    "payment",
					"domains": []string{"payment-service"},
					"routes": []map[string]string{
						{"prefix": "/", "cluster": "payment-service"},
					},
				},
			},
		},
	})

	req := DiscoveryRequest{
		Node:    "sidecar~10.0.0.1~payment-pod-abc",
		TypeURL: ClusterResource,
	}

	resp := xds.HandleDiscovery(req)
	os.Stdout.WriteString(fmt.Sprintf("  CDS Response: version=%s, resources=%d\n", resp.VersionInfo, len(resp.Resources)))
	for _, res := range resp.Resources {
		configJSON, _ := json.Marshal(res.Config)
		os.Stdout.WriteString(fmt.Sprintf("    Cluster: %s -> %s\n", res.Name, configJSON))
	}

	routeReq := DiscoveryRequest{
		Node:    "sidecar~10.0.0.1~payment-pod-abc",
		TypeURL: RouteResource,
	}
	routeResp := xds.HandleDiscovery(routeReq)
	os.Stdout.WriteString(fmt.Sprintf("  RDS Response: version=%s, resources=%d\n", routeResp.VersionInfo, len(routeResp.Resources)))

	listenerReq := DiscoveryRequest{
		Node:    "sidecar~10.0.0.1~payment-pod-abc",
		TypeURL: ListenerResource,
	}
	listenerResp := xds.HandleDiscovery(listenerReq)
	os.Stdout.WriteString(fmt.Sprintf("  LDS Response: version=%s, resources=%d\n", listenerResp.VersionInfo, len(listenerResp.Resources)))

	fmt.Println(`
  xDS PROTOCOL (real Istio):
    CDS (Cluster Discovery):    Service cluster definitions
    EDS (Endpoint Discovery):   IP:port for each cluster
    LDS (Listener Discovery):   Ports to listen on
    RDS (Route Discovery):      HTTP routing rules
    SDS (Secret Discovery):     TLS certificates

    Flow: LDS -> RDS -> CDS -> EDS (dependency order)
    Push model: istiod pushes config changes via gRPC stream`)
}

// ============================================
// mTLS CERTIFICATE MANAGER
// ============================================

type CertificateInfo struct {
	ServiceName string
	Namespace   string
	SPIFFEID    string
	SerialNum   string
	NotBefore   time.Time
	NotAfter    time.Time
	PublicKey   string
	IsCA        bool
}

type CertificateManager struct {
	mu          sync.RWMutex
	rootCA      *CertificateInfo
	certs       map[string]*CertificateInfo
	rotationDur time.Duration
	serialSeq   int
}

func NewCertificateManager(rootDomain string) *CertificateManager {
	cm := &CertificateManager{
		certs:       make(map[string]*CertificateInfo),
		rotationDur: 24 * time.Hour,
	}

	cm.rootCA = &CertificateInfo{
		ServiceName: "istio-ca",
		Namespace:   "istio-system",
		SPIFFEID:    fmt.Sprintf("spiffe://%s", rootDomain),
		SerialNum:   cm.nextSerial(),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		PublicKey:    generateKeyFingerprint(),
		IsCA:        true,
	}

	return cm
}

func (cm *CertificateManager) nextSerial() string {
	cm.serialSeq++
	return fmt.Sprintf("SN-%06d", cm.serialSeq)
}

func generateKeyFingerprint() string {
	b := make([]byte, 16)
	rand.Read(b)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:8])
}

func (cm *CertificateManager) IssueCert(service, namespace string) *CertificateInfo {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	spiffeID := fmt.Sprintf("spiffe://cluster.local/ns/%s/sa/%s", namespace, service)

	cert := &CertificateInfo{
		ServiceName: service,
		Namespace:   namespace,
		SPIFFEID:    spiffeID,
		SerialNum:   cm.nextSerial(),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(cm.rotationDur),
		PublicKey:    generateKeyFingerprint(),
		IsCA:        false,
	}

	cm.certs[spiffeID] = cert
	return cert
}

func (cm *CertificateManager) VerifyCert(spiffeID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	cert, ok := cm.certs[spiffeID]
	if !ok {
		return false
	}

	now := time.Now()
	return now.After(cert.NotBefore) && now.Before(cert.NotAfter)
}

func (cm *CertificateManager) RotateCert(spiffeID string) *CertificateInfo {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	old, ok := cm.certs[spiffeID]
	if !ok {
		return nil
	}

	newCert := &CertificateInfo{
		ServiceName: old.ServiceName,
		Namespace:   old.Namespace,
		SPIFFEID:    old.SPIFFEID,
		SerialNum:   cm.nextSerial(),
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(cm.rotationDur),
		PublicKey:    generateKeyFingerprint(),
		IsCA:        false,
	}

	cm.certs[spiffeID] = newCert
	return newCert
}

func (cm *CertificateManager) ListCerts() []*CertificateInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var result []*CertificateInfo
	for _, cert := range cm.certs {
		result = append(result, cert)
	}
	return result
}

func demoCertificateManager() {
	cm := NewCertificateManager("cluster.local")

	os.Stdout.WriteString(fmt.Sprintf("  Root CA: %s (serial=%s, key=%s)\n",
		cm.rootCA.SPIFFEID, cm.rootCA.SerialNum, cm.rootCA.PublicKey))

	services := []struct{ name, ns string }{
		{"payment-service", "default"},
		{"order-service", "default"},
		{"user-service", "auth"},
		{"gateway", "istio-system"},
	}

	for _, svc := range services {
		cert := cm.IssueCert(svc.name, svc.ns)
		os.Stdout.WriteString(fmt.Sprintf("  Issued: %s (serial=%s, expires=%s)\n",
			cert.SPIFFEID, cert.SerialNum,
			cert.NotAfter.Format("2006-01-02 15:04")))
	}

	paymentSPIFFE := "spiffe://cluster.local/ns/default/sa/payment-service"
	valid := cm.VerifyCert(paymentSPIFFE)
	os.Stdout.WriteString(fmt.Sprintf("  Verify %s: %v\n", paymentSPIFFE, valid))

	newCert := cm.RotateCert(paymentSPIFFE)
	os.Stdout.WriteString(fmt.Sprintf("  Rotated: %s (new serial=%s, new key=%s)\n",
		newCert.SPIFFEID, newCert.SerialNum, newCert.PublicKey))

	bogusValid := cm.VerifyCert("spiffe://cluster.local/ns/default/sa/evil-service")
	os.Stdout.WriteString(fmt.Sprintf("  Verify unknown service: %v\n", bogusValid))

	fmt.Println(`
  mTLS IN ISTIO (real implementation):
    - Citadel (now part of istiod) acts as CA
    - SPIFFE identity: spiffe://cluster.local/ns/NS/sa/SA
    - SDS (Secret Discovery Service) delivers certs to Envoy
    - Automatic rotation before expiry (default 24h)
    - Root CA can be plugged in (Vault, cert-manager, etc.)
    - PeerAuthentication CRD controls mTLS mode`)
}

// ============================================
// TRAFFIC ROUTING ENGINE
// ============================================

type VirtualService struct {
	Name     string
	Hosts    []string
	HTTP     []HTTPRoute
}

type HTTPRoute struct {
	Match []HTTPMatchRequest
	Route []HTTPRouteDestination
	Fault *FaultInjection
	Retry *RetryPolicy
}

type HTTPMatchRequest struct {
	URI     *StringMatch
	Headers map[string]*StringMatch
}

type StringMatch struct {
	Exact  string
	Prefix string
	Regex  string
}

type HTTPRouteDestination struct {
	Host   string
	Subset string
	Port   int
	Weight int
}

type FaultInjection struct {
	Delay *FaultDelay
	Abort *FaultAbort
}

type FaultDelay struct {
	FixedDelay time.Duration
	Percentage float64
}

type FaultAbort struct {
	HTTPStatus int
	Percentage float64
}

type RetryPolicy struct {
	Attempts      int
	PerTryTimeout time.Duration
	RetryOn       string
}

type DestinationRule struct {
	Host    string
	Subsets []Subset
}

type Subset struct {
	Name   string
	Labels map[string]string
}

type TrafficRouter struct {
	mu               sync.RWMutex
	virtualServices  map[string]*VirtualService
	destinationRules map[string]*DestinationRule
	registry         *ServiceRegistry
}

func NewTrafficRouter(registry *ServiceRegistry) *TrafficRouter {
	return &TrafficRouter{
		virtualServices:  make(map[string]*VirtualService),
		destinationRules: make(map[string]*DestinationRule),
		registry:         registry,
	}
}

func (tr *TrafficRouter) ApplyVirtualService(vs *VirtualService) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.virtualServices[vs.Name] = vs
}

func (tr *TrafficRouter) ApplyDestinationRule(dr *DestinationRule) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.destinationRules[dr.Host] = dr
}

func (tr *TrafficRouter) Route(host, path string, headers map[string]string) *HTTPRouteDestination {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	for _, vs := range tr.virtualServices {
		hostMatch := false
		for _, h := range vs.Hosts {
			if h == host || h == "*" {
				hostMatch = true
				break
			}
		}
		if !hostMatch {
			continue
		}

		for _, route := range vs.HTTP {
			if matchesRoute(route, path, headers) {
				return selectDestination(route.Route)
			}
		}
	}

	return &HTTPRouteDestination{Host: host, Port: 80, Weight: 100}
}

func matchesRoute(route HTTPRoute, path string, headers map[string]string) bool {
	if len(route.Match) == 0 {
		return true
	}

	for _, match := range route.Match {
		if match.URI != nil {
			if match.URI.Exact != "" && path != match.URI.Exact {
				continue
			}
			if match.URI.Prefix != "" && !strings.HasPrefix(path, match.URI.Prefix) {
				continue
			}
		}

		headerMatch := true
		for key, sm := range match.Headers {
			val, ok := headers[key]
			if !ok {
				headerMatch = false
				break
			}
			if sm.Exact != "" && val != sm.Exact {
				headerMatch = false
				break
			}
		}
		if headerMatch {
			return true
		}
	}
	return false
}

func selectDestination(routes []HTTPRouteDestination) *HTTPRouteDestination {
	if len(routes) == 0 {
		return nil
	}
	if len(routes) == 1 {
		return &routes[0]
	}

	totalWeight := 0
	for _, r := range routes {
		totalWeight += r.Weight
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(int64(totalWeight)))
	point := int(n.Int64())

	cumulative := 0
	for i := range routes {
		cumulative += routes[i].Weight
		if point < cumulative {
			return &routes[i]
		}
	}
	return &routes[0]
}

func demoTrafficRouting() {
	registry := NewServiceRegistry()
	router := NewTrafficRouter(registry)

	router.ApplyVirtualService(&VirtualService{
		Name:  "payment-routing",
		Hosts: []string{"payment-service"},
		HTTP: []HTTPRoute{
			{
				Match: []HTTPMatchRequest{
					{Headers: map[string]*StringMatch{"x-canary": {Exact: "true"}}},
				},
				Route: []HTTPRouteDestination{
					{Host: "payment-service", Subset: "v2", Port: 8080, Weight: 100},
				},
			},
			{
				Route: []HTTPRouteDestination{
					{Host: "payment-service", Subset: "v1", Port: 8080, Weight: 90},
					{Host: "payment-service", Subset: "v2", Port: 8080, Weight: 10},
				},
			},
		},
	})

	router.ApplyDestinationRule(&DestinationRule{
		Host: "payment-service",
		Subsets: []Subset{
			{Name: "v1", Labels: map[string]string{"version": "v1"}},
			{Name: "v2", Labels: map[string]string{"version": "v2"}},
		},
	})

	fmt.Println("  Canary header routing:")
	dest := router.Route("payment-service", "/api/pay", map[string]string{"x-canary": "true"})
	os.Stdout.WriteString(fmt.Sprintf("    -> %s (subset=%s, port=%d)\n", dest.Host, dest.Subset, dest.Port))

	fmt.Println("  Weighted routing (100 requests):")
	v1Count, v2Count := 0, 0
	for i := 0; i < 100; i++ {
		d := router.Route("payment-service", "/api/pay", nil)
		if d.Subset == "v1" {
			v1Count++
		} else {
			v2Count++
		}
	}
	os.Stdout.WriteString(fmt.Sprintf("    v1: %d requests, v2: %d requests\n", v1Count, v2Count))

	fmt.Println("  Default routing (unknown service):")
	def := router.Route("unknown-service", "/api/test", nil)
	os.Stdout.WriteString(fmt.Sprintf("    -> %s (port=%d)\n", def.Host, def.Port))
}

// ============================================
// ENVOY SIDECAR SIMULATION
// ============================================

type EnvoySidecar struct {
	mu          sync.Mutex
	nodeID      string
	clusters    map[string]*EnvoyCluster
	listeners   []*EnvoyListener
	routes      map[string]*EnvoyRoute
	stats       ProxyStats
	certManager *CertificateManager
	spiffeID    string
}

type EnvoyCluster struct {
	Name      string
	Endpoints []Endpoint
	LBPolicy  string
	nextIdx   int
}

type Endpoint struct {
	Address string
	Port    int
	Weight  int
	Healthy bool
}

type EnvoyListener struct {
	Name    string
	Address string
	Port    int
}

type EnvoyRoute struct {
	Name        string
	VirtualHost string
	Prefix      string
	Cluster     string
}

type ProxyStats struct {
	RequestsTotal   int
	RequestsSuccess int
	RequestsFailed  int
	BytesSent       int64
	BytesReceived   int64
	ActiveConns     int
}

func NewEnvoySidecar(nodeID string, cm *CertificateManager) *EnvoySidecar {
	return &EnvoySidecar{
		nodeID:      nodeID,
		clusters:    make(map[string]*EnvoyCluster),
		routes:      make(map[string]*EnvoyRoute),
		certManager: cm,
	}
}

func (e *EnvoySidecar) ApplyCluster(name string, endpoints []Endpoint, lbPolicy string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clusters[name] = &EnvoyCluster{
		Name:      name,
		Endpoints: endpoints,
		LBPolicy:  lbPolicy,
	}
}

func (e *EnvoySidecar) ApplyListener(name, address string, port int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners = append(e.listeners, &EnvoyListener{
		Name: name, Address: address, Port: port,
	})
}

func (e *EnvoySidecar) ApplyRoute(name, prefix, cluster string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.routes[prefix] = &EnvoyRoute{
		Name: name, Prefix: prefix, Cluster: cluster,
	}
}

func (e *EnvoySidecar) HandleRequest(path string) (string, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stats.RequestsTotal++

	var matchedRoute *EnvoyRoute
	for prefix, route := range e.routes {
		if strings.HasPrefix(path, prefix) {
			matchedRoute = route
			break
		}
	}

	if matchedRoute == nil {
		e.stats.RequestsFailed++
		return "", false
	}

	cluster, ok := e.clusters[matchedRoute.Cluster]
	if !ok || len(cluster.Endpoints) == 0 {
		e.stats.RequestsFailed++
		return "", false
	}

	var endpoint *Endpoint
	switch cluster.LBPolicy {
	case "ROUND_ROBIN":
		for attempts := 0; attempts < len(cluster.Endpoints); attempts++ {
			ep := &cluster.Endpoints[cluster.nextIdx%len(cluster.Endpoints)]
			cluster.nextIdx++
			if ep.Healthy {
				endpoint = ep
				break
			}
		}
	default:
		for i := range cluster.Endpoints {
			if cluster.Endpoints[i].Healthy {
				endpoint = &cluster.Endpoints[i]
				break
			}
		}
	}

	if endpoint == nil {
		e.stats.RequestsFailed++
		return "", false
	}

	e.stats.RequestsSuccess++
	target := fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port)
	return target, true
}

func (e *EnvoySidecar) GetStats() ProxyStats {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.stats
}

func demoEnvoySidecar() {
	cm := NewCertificateManager("cluster.local")
	sidecar := NewEnvoySidecar("sidecar~10.0.0.1~payment-pod", cm)

	sidecar.ApplyCluster("order-service", []Endpoint{
		{Address: "10.0.1.1", Port: 8081, Weight: 1, Healthy: true},
		{Address: "10.0.1.2", Port: 8081, Weight: 1, Healthy: true},
		{Address: "10.0.1.3", Port: 8081, Weight: 1, Healthy: false},
	}, "ROUND_ROBIN")

	sidecar.ApplyListener("inbound", "0.0.0.0", 8080)
	sidecar.ApplyRoute("order-route", "/api/orders", "order-service")

	for i := 0; i < 6; i++ {
		target, ok := sidecar.HandleRequest("/api/orders/123")
		if ok {
			os.Stdout.WriteString(fmt.Sprintf("  Request %d -> %s\n", i+1, target))
		} else {
			os.Stdout.WriteString(fmt.Sprintf("  Request %d -> FAILED (no healthy upstream)\n", i+1))
		}
	}

	_, ok := sidecar.HandleRequest("/api/unknown")
	os.Stdout.WriteString(fmt.Sprintf("  Unknown route: success=%v\n", ok))

	stats := sidecar.GetStats()
	os.Stdout.WriteString(fmt.Sprintf("  Stats: total=%d, success=%d, failed=%d\n",
		stats.RequestsTotal, stats.RequestsSuccess, stats.RequestsFailed))
}

// ============================================
// FULL MINI-ISTIO INTEGRATION
// ============================================

type MiniIstio struct {
	registry    *ServiceRegistry
	xds         *XDSServer
	certManager *CertificateManager
	router      *TrafficRouter
	sidecars    map[string]*EnvoySidecar
	mu          sync.RWMutex
}

func NewMiniIstio() *MiniIstio {
	registry := NewServiceRegistry()
	return &MiniIstio{
		registry:    registry,
		xds:         NewXDSServer(),
		certManager: NewCertificateManager("cluster.local"),
		router:      NewTrafficRouter(registry),
		sidecars:    make(map[string]*EnvoySidecar),
	}
}

func (m *MiniIstio) RegisterService(inst *ServiceInstance) {
	m.registry.Register(inst)

	cert := m.certManager.IssueCert(inst.Service, inst.Namespace)

	m.xds.SetResource(XDSResource{
		Name: inst.Service,
		Type: ClusterResource,
		Config: map[string]interface{}{
			"name":      inst.Service,
			"type":      "EDS",
			"lb_policy": "ROUND_ROBIN",
		},
	})

	m.mu.Lock()
	sidecar := NewEnvoySidecar(
		fmt.Sprintf("sidecar~%s~%s", inst.Address, inst.ID),
		m.certManager,
	)
	sidecar.spiffeID = cert.SPIFFEID
	m.sidecars[inst.ID] = sidecar
	m.mu.Unlock()
}

func (m *MiniIstio) ApplyRoutingRule(vs *VirtualService) {
	m.router.ApplyVirtualService(vs)
}

func (m *MiniIstio) GetSidecar(instanceID string) *EnvoySidecar {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sidecars[instanceID]
}

func demoMiniIstio() {
	istio := NewMiniIstio()

	instances := []*ServiceInstance{
		{ID: "web-1", Service: "web-frontend", Address: "10.0.0.1", Port: 80, Healthy: true, Version: "v1", Namespace: "default"},
		{ID: "api-1", Service: "api-gateway", Address: "10.0.1.1", Port: 8080, Healthy: true, Version: "v1", Namespace: "default"},
		{ID: "pay-1", Service: "payment-svc", Address: "10.0.2.1", Port: 8081, Healthy: true, Version: "v1", Namespace: "default"},
		{ID: "pay-2", Service: "payment-svc", Address: "10.0.2.2", Port: 8081, Healthy: true, Version: "v2", Namespace: "default"},
	}

	for _, inst := range instances {
		istio.RegisterService(inst)
		os.Stdout.WriteString(fmt.Sprintf("  Registered: %s (%s:%d)\n", inst.ID, inst.Address, inst.Port))
	}

	istio.ApplyRoutingRule(&VirtualService{
		Name:  "canary-payment",
		Hosts: []string{"payment-svc"},
		HTTP: []HTTPRoute{
			{
				Route: []HTTPRouteDestination{
					{Host: "payment-svc", Subset: "v1", Weight: 80},
					{Host: "payment-svc", Subset: "v2", Weight: 20},
				},
			},
		},
	})

	services := istio.registry.ListServices()
	os.Stdout.WriteString(fmt.Sprintf("  Services: %v\n", services))

	certs := istio.certManager.ListCerts()
	os.Stdout.WriteString(fmt.Sprintf("  Active certificates: %d\n", len(certs)))
	for _, cert := range certs {
		os.Stdout.WriteString(fmt.Sprintf("    %s (serial=%s)\n", cert.SPIFFEID, cert.SerialNum))
	}

	cdsResp := istio.xds.HandleDiscovery(DiscoveryRequest{TypeURL: ClusterResource})
	os.Stdout.WriteString(fmt.Sprintf("  xDS clusters: %d\n", len(cdsResp.Resources)))

	fmt.Println(`
  ISTIO REAL FEATURES:
    - Automatic sidecar injection via MutatingWebhook
    - Envoy filter chains for protocol detection
    - Wasm plugin support for custom filters
    - Multi-cluster mesh federation
    - External authorization (ext-authz)
    - Rate limiting with Envoy filters
    - Distributed tracing (Jaeger, Zipkin)
    - Prometheus metrics + Grafana dashboards
    - Kiali service mesh visualization`)
}

// Ensure imports are used
var _ = sort.Strings
var _ = big.NewInt
var _ = hex.EncodeToString

/* SUMMARY - CHAPTER 129: ISTIO INTERNALS

Key Concepts:

1. SERVICE REGISTRY
   - Service instance registration and discovery
   - Health-based filtering
   - Watch pattern for real-time updates

2. xDS CONFIGURATION SERVER
   - CDS (Cluster), EDS (Endpoint), LDS (Listener), RDS (Route)
   - Discovery request/response protocol
   - Version-based config push

3. mTLS CERTIFICATE MANAGEMENT
   - SPIFFE identity (spiffe://cluster.local/ns/NS/sa/SA)
   - Certificate issuance and rotation
   - Root CA trust chain

4. TRAFFIC ROUTING ENGINE
   - VirtualService for routing rules
   - DestinationRule for subsets
   - Weighted routing (canary deployments)
   - Header-based routing

5. ENVOY SIDECAR PROXY
   - Cluster/listener/route configuration
   - Round-robin load balancing
   - Health-aware endpoint selection
   - Request statistics tracking

Go Patterns Demonstrated:
- Observer pattern (service watchers)
- gRPC streaming simulation (xDS push)
- Middleware/proxy pattern (sidecar)
- Configuration reconciliation
- Weighted random selection
- Mutex-protected concurrent state

Real Istio Implementation:
- Full xDS v3 protocol over gRPC
- Envoy as data plane proxy (C++)
- Kubernetes informers for resource watching
- Admission webhook for sidecar injection
- Wasm plugin support
- Multi-cluster service mesh
- External authorization integration
*/
