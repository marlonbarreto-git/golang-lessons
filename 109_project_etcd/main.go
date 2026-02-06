// Package main - Chapter 109: etcd - Distributed Key-Value Store
// etcd is a strongly consistent, distributed key-value store used as
// the backbone for Kubernetes and many distributed systems.
// Written in Go, it demonstrates consensus algorithms, WAL logging,
// and distributed systems fundamentals.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 109: ETCD - DISTRIBUTED KEY-VALUE STORE ===")

	// ============================================
	// WHAT IS ETCD
	// ============================================
	fmt.Println(`
WHAT IS ETCD
------------
etcd is a distributed, reliable key-value store for the most critical
data of a distributed system. It uses the Raft consensus algorithm to
ensure data consistency across a cluster of machines.

Created at CoreOS in 2013 (now part of CNCF), etcd is the data store
behind Kubernetes, storing all cluster state.

KEY FACTS:
- Written entirely in Go
- ~48,000+ GitHub stars
- Uses Raft consensus algorithm
- Provides strong consistency (linearizability)
- Supports watch API for change notifications
- mvcc (multi-version concurrency control) storage
- gRPC API with HTTP/JSON gateway`)

	// ============================================
	// WHY GO FOR ETCD
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR ETCD
---------------------------
1. Concurrency      - Goroutines for handling client connections
2. Performance      - Low-latency reads/writes critical for K8s
3. Static binaries  - Single binary deployment
4. gRPC native      - Go has first-class gRPC support
5. Cross-platform   - Runs on Linux, macOS, Windows
6. Simplicity       - Raft implementation needs clear, correct code`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
ETCD ARCHITECTURE
-----------------

  +--------------------------------------------------+
  |                    CLIENT                         |
  |          (gRPC / HTTP+JSON gateway)               |
  +------------------+-------------------------------+
                     |
  +------------------v-------------------------------+
  |                etcd SERVER                        |
  |  +------------+  +-----------+  +-------------+  |
  |  |  gRPC API  |  |  Auth &   |  |   Watch     |  |
  |  |  Server    |  |  RBAC     |  |   Hub       |  |
  |  +-----+------+  +-----------+  +------+------+  |
  |        |                               |          |
  |  +-----v-------------------------------v------+   |
  |  |              MVCC Store                     |   |
  |  |  (multi-version concurrency control)        |   |
  |  +-----+--------------------------------------+   |
  |        |                                          |
  |  +-----v-----------+  +-----------------------+   |
  |  |    BoltDB       |  |   WAL (Write-Ahead    |   |
  |  |  (B+ tree)      |  |    Log) + Snapshots   |   |
  |  +-----------------+  +-----------------------+   |
  |                                                    |
  |  +------------------------------------------------+
  |  |              Raft Consensus Layer               |
  |  |  (leader election, log replication)             |
  |  +----+----------------+----------------+---------+
  +-------+----------------+----------------+---------+
          |                |                |
     +----v----+     +----v----+     +----v----+
     | Node 1  |     | Node 2  |     | Node 3  |
     | (Leader)|     |(Follower)|    |(Follower)|
     +---------+     +---------+     +---------+

RAFT CONSENSUS:
- Leader elected by majority vote
- All writes go through leader
- Leader replicates to followers
- Committed when majority acknowledges
- Tolerates (N-1)/2 failures in N-node cluster`)

	// ============================================
	// KEY GO PATTERNS IN ETCD
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN ETCD
-----------------------

1. RAFT CONSENSUS IMPLEMENTATION
   - State machine: Follower -> Candidate -> Leader
   - Heartbeat via goroutine timers
   - Log replication through channels
   - Term-based leader election

2. MVCC (Multi-Version Concurrency Control)
   - Every key has a revision number
   - Historical versions retained until compaction
   - Enables consistent snapshots and watches
   - Key: /registry/pods/default/my-pod
   - Value: serialized pod object + revision

3. WATCH MECHANISM
   - Long-lived gRPC streams for change notifications
   - Watch hub multiplexes watchers efficiently
   - Goroutine per watcher with channel delivery
   - Supports watch from specific revision

4. WRITE-AHEAD LOG (WAL)
   - All mutations written to WAL before applying
   - Enables crash recovery
   - Periodic snapshots reduce replay time
   - Segment-based file management

5. LEASE MECHANISM
   - TTL-based key expiration
   - Used for leader election, service discovery
   - Lease keepalive via streaming RPC
   - Efficient lease revocation cascades`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN ETCD
-------------------------------

1. BBOLT (formerly BoltDB) STORAGE
   - B+ tree for ordered key storage
   - Memory-mapped file I/O
   - Copy-on-write for transactions
   - Read transactions never block writes

2. BATCHED RAFT PROPOSALS
   - Multiple client requests batched into single Raft proposal
   - Reduces consensus round trips
   - Pipeline mode for high-throughput writes

3. STREAMING WATCHES
   - gRPC bidirectional streaming
   - Coalesced watch events reduce network I/O
   - Efficient fan-out from watch hub

4. COMPACTION
   - Old revisions compacted to save space
   - Defragmentation reclaims disk space
   - Configurable auto-compaction (periodic/revision)

5. LEARNER NODES
   - Non-voting members that replicate data
   - Reduces cluster disruption during scaling
   - Can be promoted to voting member`)

	// ============================================
	// GO PHILOSOPHY IN ETCD
	// ============================================
	fmt.Println(`
HOW ETCD EXPRESSES GO PHILOSOPHY
---------------------------------

"Don't communicate by sharing memory; share memory
 by communicating"
  Raft nodes communicate via message passing.
  The Raft log IS the shared state, replicated
  through explicit message channels.

"Errors are values"
  etcd uses rich error types with gRPC status codes.
  Clients can programmatically handle different
  failure modes (leader change, timeout, etc).

"Clear is better than clever"
  The Raft implementation prioritizes correctness
  and readability over micro-optimizations.
  Distributed consensus must be correct first.

"Make the zero value useful"
  etcd client with default config connects to
  localhost:2379. Zero-value timeouts use sensible
  defaults.`)

	// ============================================
	// SIMPLIFIED ETCD DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED ETCD-LIKE KEY-VALUE STORE ===")

	cluster := NewEtcdCluster(3)
	cluster.Start()

	fmt.Println("\n--- Writing Keys ---")
	cluster.Put("/services/web", "10.0.0.1:8080")
	cluster.Put("/services/api", "10.0.0.2:9090")
	cluster.Put("/services/db", "10.0.0.3:5432")
	cluster.Put("/config/app/name", "my-application")
	cluster.Put("/config/app/version", "2.1.0")
	cluster.Put("/config/app/debug", "false")

	fmt.Println("\n--- Reading Keys ---")
	cluster.Get("/services/web")
	cluster.Get("/services/api")
	cluster.Get("/config/app/name")
	cluster.Get("/nonexistent/key")

	fmt.Println("\n--- Range Query ---")
	cluster.GetRange("/services/", "/services0")
	cluster.GetRange("/config/app/", "/config/app0")

	fmt.Println("\n--- Watch Demo ---")
	watcher := cluster.Watch("/services/")
	go func() {
		time.Sleep(50 * time.Millisecond)
		cluster.Put("/services/cache", "10.0.0.4:6379")
		cluster.Put("/services/web", "10.0.0.5:8080")
		cluster.Delete("/services/db")
		time.Sleep(50 * time.Millisecond)
		close(watcher.stop)
	}()
	watcher.Drain()

	fmt.Println("\n--- Lease Demo ---")
	leaseID := cluster.GrantLease(5)
	cluster.PutWithLease("/ephemeral/leader", "node-1", leaseID)
	cluster.Get("/ephemeral/leader")
	fmt.Println("  (Lease would expire after TTL in real system)")

	fmt.Println("\n--- Transaction (Compare-and-Swap) ---")
	cluster.Transaction("/services/web", "10.0.0.5:8080", "10.0.0.6:8080")
	cluster.Transaction("/services/web", "10.0.0.5:8080", "10.0.0.7:8080")

	fmt.Println("\n--- Cluster Status ---")
	cluster.PrintStatus()

	fmt.Println("\n--- Compaction ---")
	cluster.Compact(3)

	fmt.Println(`
SUMMARY
-------
etcd is the critical data store behind Kubernetes.
Its Go codebase demonstrates:
- Raft consensus for distributed consistency
- MVCC for concurrent access and history
- Watch API for reactive change notifications
- WAL for crash recovery and durability
- gRPC for high-performance client communication

etcd proves Go can implement correct distributed consensus.`)
}

// ============================================
// TYPES
// ============================================

type EventType string

const (
	EventPut    EventType = "PUT"
	EventDelete EventType = "DELETE"
)

type KeyValue struct {
	Key            string
	Value          string
	CreateRevision int64
	ModRevision    int64
	Version        int64
}

type WatchEvent struct {
	Type EventType
	KV   KeyValue
}

type Watcher struct {
	prefix string
	events chan WatchEvent
	stop   chan struct{}
}

type RaftNode struct {
	ID       int
	IsLeader bool
	Term     int64
	Log      []RaftEntry
}

type RaftEntry struct {
	Term    int64
	Index   int64
	Command string
}

type Lease struct {
	ID  int64
	TTL int64
	Keys []string
}

// ============================================
// ETCD CLUSTER
// ============================================

type EtcdCluster struct {
	nodes    []*RaftNode
	store    map[string]*KeyValue
	revision int64
	watchers []*Watcher
	leases   map[int64]*Lease
	nextLease int64
	mu       sync.RWMutex
}

func NewEtcdCluster(size int) *EtcdCluster {
	nodes := make([]*RaftNode, size)
	for i := 0; i < size; i++ {
		nodes[i] = &RaftNode{
			ID:   i + 1,
			Term: 1,
		}
	}
	return &EtcdCluster{
		nodes:     nodes,
		store:     make(map[string]*KeyValue),
		leases:    make(map[int64]*Lease),
		nextLease: 1000,
	}
}

func (c *EtcdCluster) Start() {
	c.nodes[0].IsLeader = true
	fmt.Printf("etcd cluster started with %d nodes\n", len(c.nodes))
	fmt.Printf("Leader elected: node-%d (term: %d)\n", c.nodes[0].ID, c.nodes[0].Term)
}

func (c *EtcdCluster) Put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.revision++

	existing, exists := c.store[key]
	var version int64 = 1
	var createRev int64 = c.revision
	if exists {
		version = existing.Version + 1
		createRev = existing.CreateRevision
	}

	kv := &KeyValue{
		Key:            key,
		Value:          value,
		CreateRevision: createRev,
		ModRevision:    c.revision,
		Version:        version,
	}
	c.store[key] = kv

	leader := c.getLeader()
	leader.Log = append(leader.Log, RaftEntry{
		Term:    leader.Term,
		Index:   c.revision,
		Command: fmt.Sprintf("PUT %s=%s", key, value),
	})

	action := "created"
	if exists {
		action = "updated"
	}
	fmt.Printf("  PUT %s = %s (rev: %d, %s)\n", key, value, c.revision, action)

	c.notifyWatchers(EventPut, *kv)
}

func (c *EtcdCluster) Get(key string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	kv, exists := c.store[key]
	if !exists {
		fmt.Printf("  GET %s = (not found)\n", key)
		return
	}
	fmt.Printf("  GET %s = %s (rev: %d, ver: %d)\n",
		key, kv.Value, kv.ModRevision, kv.Version)
}

func (c *EtcdCluster) GetRange(start, end string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Printf("  RANGE [%s, %s):\n", start, end)

	var keys []string
	for k := range c.store {
		if k >= start && k < end {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	for _, k := range keys {
		kv := c.store[k]
		fmt.Printf("    %s = %s (rev: %d)\n", k, kv.Value, kv.ModRevision)
	}
	fmt.Printf("  Total: %d keys\n", len(keys))
}

func (c *EtcdCluster) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	kv, exists := c.store[key]
	if !exists {
		fmt.Printf("  DELETE %s = (not found)\n", key)
		return
	}

	c.revision++
	deletedKV := *kv
	deletedKV.ModRevision = c.revision
	delete(c.store, key)

	leader := c.getLeader()
	leader.Log = append(leader.Log, RaftEntry{
		Term:    leader.Term,
		Index:   c.revision,
		Command: fmt.Sprintf("DELETE %s", key),
	})

	fmt.Printf("  DELETE %s (rev: %d)\n", key, c.revision)
	c.notifyWatchers(EventDelete, deletedKV)
}

func (c *EtcdCluster) Watch(prefix string) *Watcher {
	c.mu.Lock()
	defer c.mu.Unlock()

	w := &Watcher{
		prefix: prefix,
		events: make(chan WatchEvent, 100),
		stop:   make(chan struct{}),
	}
	c.watchers = append(c.watchers, w)
	fmt.Printf("  WATCH prefix=%s (watcher created)\n", prefix)
	return w
}

func (w *Watcher) Drain() {
	fmt.Printf("  Watching for changes on prefix %s...\n", w.prefix)
	for {
		select {
		case event := <-w.events:
			fmt.Printf("  WATCH EVENT: %s %s = %s (rev: %d)\n",
				event.Type, event.KV.Key, event.KV.Value, event.KV.ModRevision)
		case <-w.stop:
			fmt.Printf("  Watch on %s closed\n", w.prefix)
			return
		}
	}
}

func (c *EtcdCluster) notifyWatchers(eventType EventType, kv KeyValue) {
	for _, w := range c.watchers {
		if strings.HasPrefix(kv.Key, w.prefix) {
			select {
			case w.events <- WatchEvent{Type: eventType, KV: kv}:
			default:
			}
		}
	}
}

func (c *EtcdCluster) GrantLease(ttl int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nextLease++
	lease := &Lease{
		ID:  c.nextLease,
		TTL: ttl,
	}
	c.leases[lease.ID] = lease
	fmt.Printf("  Lease granted: ID=%d, TTL=%ds\n", lease.ID, ttl)
	return lease.ID
}

func (c *EtcdCluster) PutWithLease(key, value string, leaseID int64) {
	c.mu.Lock()

	lease, exists := c.leases[leaseID]
	if !exists {
		fmt.Printf("  Error: lease %d not found\n", leaseID)
		c.mu.Unlock()
		return
	}
	lease.Keys = append(lease.Keys, key)
	c.mu.Unlock()

	c.Put(key, value)
	fmt.Printf("    (attached to lease %d, TTL: %ds)\n", leaseID, lease.TTL)
}

func (c *EtcdCluster) Transaction(key, expectedValue, newValue string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	kv, exists := c.store[key]
	if !exists || kv.Value != expectedValue {
		actual := "(not found)"
		if exists {
			actual = kv.Value
		}
		fmt.Printf("  TXN FAILED: %s expected=%s actual=%s\n", key, expectedValue, actual)
		return
	}

	c.revision++
	kv.Value = newValue
	kv.ModRevision = c.revision
	kv.Version++

	leader := c.getLeader()
	leader.Log = append(leader.Log, RaftEntry{
		Term:    leader.Term,
		Index:   c.revision,
		Command: fmt.Sprintf("CAS %s %s->%s", key, expectedValue, newValue),
	})

	fmt.Printf("  TXN OK: %s = %s -> %s (rev: %d)\n", key, expectedValue, newValue, c.revision)
}

func (c *EtcdCluster) Compact(revision int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Printf("  Compacting revisions <= %d\n", revision)
	fmt.Printf("  Store size: %d keys, Raft log: %d entries\n",
		len(c.store), len(c.getLeader().Log))
}

func (c *EtcdCluster) PrintStatus() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Println("Cluster Status:")
	for _, node := range c.nodes {
		role := "Follower"
		if node.IsLeader {
			role = "Leader"
		}
		fmt.Printf("  Node %d: %s (term: %d, log entries: %d)\n",
			node.ID, role, node.Term, len(node.Log))
	}
	fmt.Printf("  Total keys: %d\n", len(c.store))
	fmt.Printf("  Current revision: %d\n", c.revision)
	fmt.Printf("  Active watchers: %d\n", len(c.watchers))
	fmt.Printf("  Active leases: %d\n", len(c.leases))
}

func (c *EtcdCluster) getLeader() *RaftNode {
	for _, n := range c.nodes {
		if n.IsLeader {
			return n
		}
	}
	return c.nodes[0]
}

func init() {
	_ = os.Stdout
}
