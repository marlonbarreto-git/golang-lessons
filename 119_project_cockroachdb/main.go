// Package main - Chapter 119: CockroachDB - Distributed SQL Database in Go
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 119: COCKROACHDB - DISTRIBUTED SQL DATABASE IN GO ===")

	// ============================================
	// WHAT IS COCKROACHDB?
	// ============================================
	fmt.Println(`
==============================
WHAT IS COCKROACHDB?
==============================

CockroachDB is a distributed SQL database built for global
cloud-native applications. Written entirely in Go, it provides:

  - PostgreSQL-compatible SQL interface
  - ACID transactions across distributed nodes
  - Automatic sharding and rebalancing
  - Survivability: tolerates disk, machine, rack, datacenter failures
  - Geo-partitioning for data locality compliance

WHO USES IT:
  - DoorDash (order management)
  - Netflix (control plane)
  - Bose (IoT data management)
  - Many fintech companies for transactional workloads

WHY GO:
  CockroachDB chose Go for its excellent concurrency model,
  garbage collection, strong standard library, and fast
  compilation. The entire database is a single binary.`)

	// ============================================
	// ARCHITECTURE OVERVIEW
	// ============================================
	fmt.Println(`
==============================
ARCHITECTURE OVERVIEW
==============================

CockroachDB has a layered architecture:

  +---------------------------------------------+
  |              SQL Layer                       |
  |  (Parser -> Optimizer -> Executor)           |
  +---------------------------------------------+
  |          Transaction Layer (KV)              |
  |  (MVCC, Timestamp Oracle, 2PC)              |
  +---------------------------------------------+
  |          Distribution Layer                  |
  |  (Ranges, Raft Consensus, Leaseholders)     |
  +---------------------------------------------+
  |          Replication Layer                   |
  |  (Raft Groups, Snapshots, Log)              |
  +---------------------------------------------+
  |          Storage Layer (Pebble)              |
  |  (LSM Tree, SSTs, Block Cache)              |
  +---------------------------------------------+

KEY CONCEPTS:
  - Data is split into ~512MB "Ranges"
  - Each Range is replicated via Raft (typically 3 replicas)
  - One replica is the "Leaseholder" that serves reads
  - Transactions use MVCC with hybrid-logical clocks
  - SQL is compiled to a distributed execution plan`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
==============================
KEY GO PATTERNS IN COCKROACHDB
==============================

1. LAYERED ARCHITECTURE WITH INTERFACES
   Each layer communicates through well-defined interfaces,
   allowing testing and swapping implementations.

2. RAFT CONSENSUS IN GO
   CockroachDB uses etcd's Raft library, wrapping it with
   Go channels and goroutines for message passing.

3. MVCC (Multi-Version Concurrency Control)
   Timestamps on every key-value pair enable snapshot
   isolation without read locks.

4. HYBRID-LOGICAL CLOCKS (HLC)
   Combines physical time with logical counters to provide
   causally consistent ordering across nodes.

5. RANGE-BASED SHARDING
   Automatic splitting/merging of data ranges with
   goroutine-per-range processing model.

6. MEMORY ACCOUNTING
   Careful tracking of memory usage per-query with
   Go's runtime.MemStats and custom allocators.`)

	// ============================================
	// DEMO: SIMPLIFIED DISTRIBUTED KV STORE
	// ============================================
	fmt.Println(`
==============================
DEMO: SIMPLIFIED DISTRIBUTED KV STORE
==============================

This demo implements core CockroachDB concepts:
  - MVCC key-value storage with timestamps
  - Range-based sharding across nodes
  - Raft-like consensus for replication
  - Hybrid-Logical Clocks (HLC)
  - Distributed transactions with 2PC`)

	fmt.Println("\n--- Hybrid-Logical Clock ---")
	demoHLC()

	fmt.Println("\n--- MVCC Storage Engine ---")
	demoMVCC()

	fmt.Println("\n--- Range-Based Sharding ---")
	demoRangeSharding()

	fmt.Println("\n--- Raft Consensus (Simplified) ---")
	demoRaftConsensus()

	fmt.Println("\n--- Distributed Transactions (2PC) ---")
	demoDistributedTxn()

	fmt.Println("\n--- Full Mini-CockroachDB ---")
	demoMiniCockroachDB()

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
==============================
PERFORMANCE TECHNIQUES
==============================

1. PEBBLE STORAGE ENGINE
   CockroachDB built Pebble (replacing RocksDB) in Go:
   - LSM-tree based key-value store
   - Optimized for CockroachDB's access patterns
   - Written in pure Go (no CGo overhead)
   - Bloom filters, compression, block cache

2. VECTORIZED EXECUTION
   SQL queries use columnar, vectorized processing:
   - Operates on batches of rows (1024 at a time)
   - Cache-friendly memory layout
   - Reduces per-row interpretation overhead
   - Uses Go's unsafe package for zero-copy access

3. MEMORY MANAGEMENT
   - Per-query memory budgets with monitors
   - Disk spilling for large operations
   - Custom arena allocators for hot paths
   - Careful avoidance of allocations in tight loops

4. CONNECTION POOLING
   - Goroutine-per-connection model
   - Efficient context cancellation
   - pgwire protocol implementation in Go

5. BATCH OPERATIONS
   - Coalesces multiple KV operations
   - Reduces Raft round-trips
   - Parallel dispatch to multiple ranges`)

	// ============================================
	// GO PHILOSOPHY IN COCKROACHDB
	// ============================================
	fmt.Println(`
==============================
GO PHILOSOPHY IN COCKROACHDB
==============================

SIMPLICITY OVER CLEVERNESS:
  CockroachDB is one of the most complex Go projects,
  yet each component follows Go's simplicity principle.
  Clear interfaces, explicit error handling, and
  straightforward concurrency patterns.

COMPOSITION OVER INHERITANCE:
  The storage engine, transaction layer, and SQL layer
  compose through interfaces. Each can be tested
  and reasoned about independently.

SINGLE BINARY DEPLOYMENT:
  Like Go itself, CockroachDB ships as a single binary.
  No JVM, no external dependencies. This dramatically
  simplifies operations.

GOROUTINES FOR CONCURRENCY:
  Each SQL connection, each Raft group, each background
  task runs in its own goroutine. The Go scheduler
  handles thousands of concurrent operations efficiently.

ERROR HANDLING:
  CockroachDB wraps errors with context at each layer,
  creating detailed error chains that aid debugging
  in distributed scenarios.`)

	fmt.Println("\n=== END OF CHAPTER 119 ===")
}

// ============================================
// HYBRID-LOGICAL CLOCK (HLC)
// ============================================

type HLC struct {
	mu          sync.Mutex
	physicalNow func() int64
	wallTime    int64
	logical     uint32
}

func NewHLC() *HLC {
	return &HLC{
		physicalNow: func() int64 {
			return time.Now().UnixNano()
		},
	}
}

type HLCTimestamp struct {
	WallTime int64
	Logical  uint32
}

func (ts HLCTimestamp) String() string {
	t := time.Unix(0, ts.WallTime)
	return fmt.Sprintf("%s.%d", t.Format("15:04:05.000"), ts.Logical)
}

func (ts HLCTimestamp) Less(other HLCTimestamp) bool {
	if ts.WallTime != other.WallTime {
		return ts.WallTime < other.WallTime
	}
	return ts.Logical < other.Logical
}

func (h *HLC) Now() HLCTimestamp {
	h.mu.Lock()
	defer h.mu.Unlock()

	physNow := h.physicalNow()
	if physNow > h.wallTime {
		h.wallTime = physNow
		h.logical = 0
	} else {
		h.logical++
	}
	return HLCTimestamp{WallTime: h.wallTime, Logical: h.logical}
}

func (h *HLC) Update(received HLCTimestamp) HLCTimestamp {
	h.mu.Lock()
	defer h.mu.Unlock()

	physNow := h.physicalNow()

	if physNow > h.wallTime && physNow > received.WallTime {
		h.wallTime = physNow
		h.logical = 0
	} else if received.WallTime > h.wallTime {
		h.wallTime = received.WallTime
		h.logical = received.Logical + 1
	} else if h.wallTime > received.WallTime {
		h.logical++
	} else {
		if received.Logical > h.logical {
			h.logical = received.Logical + 1
		} else {
			h.logical++
		}
	}

	return HLCTimestamp{WallTime: h.wallTime, Logical: h.logical}
}

func demoHLC() {
	clock1 := NewHLC()
	clock2 := NewHLC()

	ts1 := clock1.Now()
	fmt.Printf("  Node1 event:    %s\n", ts1)

	ts2 := clock1.Now()
	fmt.Printf("  Node1 event:    %s\n", ts2)

	ts3 := clock2.Update(ts2)
	fmt.Printf("  Node2 receives: %s\n", ts3)

	ts4 := clock2.Now()
	fmt.Printf("  Node2 event:    %s\n", ts4)

	fmt.Println("  Causality preserved:", ts1.Less(ts2), ts2.Less(ts3), ts3.Less(ts4))
}

// ============================================
// MVCC STORAGE ENGINE
// ============================================

type MVCCKey struct {
	Key       string
	Timestamp HLCTimestamp
}

type MVCCValue struct {
	Data    []byte
	Deleted bool
}

type MVCCStore struct {
	mu   sync.RWMutex
	data map[string][]mvccEntry
}

type mvccEntry struct {
	Timestamp HLCTimestamp
	Value     MVCCValue
}

func NewMVCCStore() *MVCCStore {
	return &MVCCStore{
		data: make(map[string][]mvccEntry),
	}
}

func (s *MVCCStore) Put(key string, value []byte, ts HLCTimestamp) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := mvccEntry{
		Timestamp: ts,
		Value:     MVCCValue{Data: value},
	}

	entries := s.data[key]
	entries = append(entries, entry)
	sort.Slice(entries, func(i, j int) bool {
		return entries[j].Timestamp.Less(entries[i].Timestamp)
	})
	s.data[key] = entries
}

func (s *MVCCStore) Get(key string, readTS HLCTimestamp) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.data[key]
	for _, e := range entries {
		if e.Timestamp.Less(readTS) || e.Timestamp == readTS {
			if e.Value.Deleted {
				return nil, false
			}
			return e.Value.Data, true
		}
	}
	return nil, false
}

func (s *MVCCStore) Delete(key string, ts HLCTimestamp) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := mvccEntry{
		Timestamp: ts,
		Value:     MVCCValue{Deleted: true},
	}

	entries := s.data[key]
	entries = append(entries, entry)
	sort.Slice(entries, func(i, j int) bool {
		return entries[j].Timestamp.Less(entries[i].Timestamp)
	})
	s.data[key] = entries
}

func (s *MVCCStore) Scan(prefix string, readTS HLCTimestamp) map[string][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]byte)
	for key, entries := range s.data {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		for _, e := range entries {
			if e.Timestamp.Less(readTS) || e.Timestamp == readTS {
				if !e.Value.Deleted {
					result[key] = e.Value.Data
				}
				break
			}
		}
	}
	return result
}

func demoMVCC() {
	store := NewMVCCStore()
	clock := NewHLC()

	ts1 := clock.Now()
	store.Put("user:1", []byte(`{"name":"Alice","age":30}`), ts1)
	fmt.Printf("  Write user:1 at %s\n", ts1)

	ts2 := clock.Now()
	store.Put("user:1", []byte(`{"name":"Alice","age":31}`), ts2)
	fmt.Printf("  Update user:1 at %s\n", ts2)

	if val, ok := store.Get("user:1", ts1); ok {
		fmt.Printf("  Read at ts1: %s (snapshot isolation)\n", val)
	}
	if val, ok := store.Get("user:1", ts2); ok {
		fmt.Printf("  Read at ts2: %s (latest version)\n", val)
	}

	ts3 := clock.Now()
	store.Put("user:2", []byte(`{"name":"Bob"}`), ts1)
	store.Put("user:3", []byte(`{"name":"Charlie"}`), ts1)
	results := store.Scan("user:", ts3)
	fmt.Printf("  Scan 'user:' found %d keys\n", len(results))
}

// ============================================
// RANGE-BASED SHARDING
// ============================================

type KeyRange struct {
	StartKey string
	EndKey   string
	RangeID  int
}

func (r KeyRange) Contains(key string) bool {
	if r.EndKey == "" {
		return key >= r.StartKey
	}
	return key >= r.StartKey && key < r.EndKey
}

type RangeDescriptor struct {
	Range    KeyRange
	Replicas []int
	Leader   int
}

type RangeManager struct {
	mu     sync.RWMutex
	ranges []RangeDescriptor
	nextID int
}

func NewRangeManager(nodeCount int) *RangeManager {
	rm := &RangeManager{nextID: 1}
	rm.ranges = append(rm.ranges, RangeDescriptor{
		Range:    KeyRange{StartKey: "", EndKey: "", RangeID: 0},
		Replicas: makeReplicas(nodeCount),
		Leader:   0,
	})
	return rm
}

func makeReplicas(n int) []int {
	replicas := make([]int, n)
	for i := range replicas {
		replicas[i] = i
	}
	return replicas
}

func (rm *RangeManager) FindRange(key string) *RangeDescriptor {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for i := range rm.ranges {
		if rm.ranges[i].Range.Contains(key) {
			return &rm.ranges[i]
		}
	}
	return nil
}

func (rm *RangeManager) Split(rangeID int, splitKey string) (int, int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for i, rd := range rm.ranges {
		if rd.Range.RangeID == rangeID {
			leftID := rm.nextID
			rightID := rm.nextID + 1
			rm.nextID += 2

			left := RangeDescriptor{
				Range:    KeyRange{StartKey: rd.Range.StartKey, EndKey: splitKey, RangeID: leftID},
				Replicas: append([]int{}, rd.Replicas...),
				Leader:   rd.Leader,
			}
			right := RangeDescriptor{
				Range:    KeyRange{StartKey: splitKey, EndKey: rd.Range.EndKey, RangeID: rightID},
				Replicas: append([]int{}, rd.Replicas...),
				Leader:   rd.Leader,
			}

			rm.ranges = append(rm.ranges[:i], append([]RangeDescriptor{left, right}, rm.ranges[i+1:]...)...)
			return leftID, rightID
		}
	}
	return -1, -1
}

func demoRangeSharding() {
	rm := NewRangeManager(3)

	fmt.Println("  Initial: 1 range covering all keys")
	rd := rm.FindRange("user:500")
	fmt.Printf("  'user:500' -> Range %d (leader: node %d)\n", rd.Range.RangeID, rd.Leader)

	leftID, rightID := rm.Split(0, "m")
	fmt.Printf("  Split at 'm' -> Range %d ['' - 'm'), Range %d ['m' - '')\n", leftID, rightID)

	rd1 := rm.FindRange("apple")
	rd2 := rm.FindRange("zebra")
	fmt.Printf("  'apple' -> Range %d, 'zebra' -> Range %d\n", rd1.Range.RangeID, rd2.Range.RangeID)

	_, _ = rm.Split(rightID, "t")
	fmt.Printf("  After second split: %d ranges total\n", len(rm.ranges))
}

// ============================================
// RAFT CONSENSUS (SIMPLIFIED)
// ============================================

type RaftState int

const (
	Follower RaftState = iota
	Candidate
	Leader
)

func (s RaftState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

type LogEntry struct {
	Term    int
	Index   int
	Command string
}

type RaftNode struct {
	mu          sync.Mutex
	id          int
	state       RaftState
	currentTerm int
	votedFor    int
	log         []LogEntry
	commitIndex int
	peers       []*RaftNode
}

func NewRaftNode(id int) *RaftNode {
	return &RaftNode{
		id:       id,
		state:    Follower,
		votedFor: -1,
	}
}

func (n *RaftNode) RequestVote(candidateID, term int) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if term > n.currentTerm {
		n.currentTerm = term
		n.votedFor = -1
		n.state = Follower
	}

	if term < n.currentTerm {
		return false
	}

	if n.votedFor == -1 || n.votedFor == candidateID {
		n.votedFor = candidateID
		return true
	}
	return false
}

func (n *RaftNode) AppendEntries(leaderID, term int, entries []LogEntry) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if term < n.currentTerm {
		return false
	}

	n.currentTerm = term
	n.state = Follower

	for _, entry := range entries {
		if entry.Index <= len(n.log) {
			n.log = n.log[:entry.Index-1]
		}
		n.log = append(n.log, entry)
	}
	return true
}

func (n *RaftNode) StartElection() bool {
	n.mu.Lock()
	n.currentTerm++
	n.state = Candidate
	n.votedFor = n.id
	term := n.currentTerm
	n.mu.Unlock()

	votes := 1
	for _, peer := range n.peers {
		if peer.RequestVote(n.id, term) {
			votes++
		}
	}

	majority := (len(n.peers)+1)/2 + 1
	if votes >= majority {
		n.mu.Lock()
		n.state = Leader
		n.mu.Unlock()
		return true
	}
	return false
}

func (n *RaftNode) Propose(command string) bool {
	n.mu.Lock()
	if n.state != Leader {
		n.mu.Unlock()
		return false
	}

	entry := LogEntry{
		Term:    n.currentTerm,
		Index:   len(n.log) + 1,
		Command: command,
	}
	n.log = append(n.log, entry)
	term := n.currentTerm
	n.mu.Unlock()

	successes := 1
	for _, peer := range n.peers {
		if peer.AppendEntries(n.id, term, []LogEntry{entry}) {
			successes++
		}
	}

	majority := (len(n.peers)+1)/2 + 1
	if successes >= majority {
		n.mu.Lock()
		n.commitIndex = entry.Index
		n.mu.Unlock()
		return true
	}
	return false
}

func demoRaftConsensus() {
	nodes := make([]*RaftNode, 3)
	for i := range nodes {
		nodes[i] = NewRaftNode(i)
	}
	for i, node := range nodes {
		for j, peer := range nodes {
			if i != j {
				node.peers = append(node.peers, peer)
			}
		}
	}

	won := nodes[0].StartElection()
	fmt.Printf("  Node 0 election: won=%v, state=%s\n", won, nodes[0].state)

	committed := nodes[0].Propose("SET user:1 Alice")
	fmt.Printf("  Propose 'SET user:1 Alice': committed=%v\n", committed)

	committed = nodes[0].Propose("SET user:2 Bob")
	fmt.Printf("  Propose 'SET user:2 Bob': committed=%v\n", committed)

	for i, node := range nodes {
		node.mu.Lock()
		fmt.Printf("  Node %d: state=%s, term=%d, log_len=%d, commit=%d\n",
			i, node.state, node.currentTerm, len(node.log), node.commitIndex)
		node.mu.Unlock()
	}
}

// ============================================
// DISTRIBUTED TRANSACTIONS (2PC)
// ============================================

type TxnStatus int

const (
	TxnPending TxnStatus = iota
	TxnCommitted
	TxnAborted
)

func (s TxnStatus) String() string {
	switch s {
	case TxnPending:
		return "PENDING"
	case TxnCommitted:
		return "COMMITTED"
	case TxnAborted:
		return "ABORTED"
	default:
		return "UNKNOWN"
	}
}

type TxnRecord struct {
	ID        string
	Status    TxnStatus
	Timestamp HLCTimestamp
	Writes    []WriteIntent
}

type WriteIntent struct {
	Key   string
	Value []byte
	TxnID string
}

type TxnCoordinator struct {
	mu      sync.Mutex
	records map[string]*TxnRecord
	store   *MVCCStore
	clock   *HLC
}

func NewTxnCoordinator(store *MVCCStore, clock *HLC) *TxnCoordinator {
	return &TxnCoordinator{
		records: make(map[string]*TxnRecord),
		store:   store,
		clock:   clock,
	}
}

func (tc *TxnCoordinator) Begin() *TxnRecord {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	ts := tc.clock.Now()
	txn := &TxnRecord{
		ID:        fmt.Sprintf("txn-%d-%d", ts.WallTime, ts.Logical),
		Status:    TxnPending,
		Timestamp: ts,
	}
	tc.records[txn.ID] = txn
	return txn
}

func (tc *TxnCoordinator) Write(txn *TxnRecord, key string, value []byte) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	txn.Writes = append(txn.Writes, WriteIntent{
		Key:   key,
		Value: value,
		TxnID: txn.ID,
	})
}

func (tc *TxnCoordinator) Commit(txn *TxnRecord) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	commitTS := tc.clock.Now()

	for _, w := range txn.Writes {
		tc.store.Put(w.Key, w.Value, commitTS)
	}

	txn.Status = TxnCommitted
	txn.Timestamp = commitTS
	return true
}

func (tc *TxnCoordinator) Abort(txn *TxnRecord) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	txn.Status = TxnAborted
	txn.Writes = nil
}

func demoDistributedTxn() {
	store := NewMVCCStore()
	clock := NewHLC()
	coord := NewTxnCoordinator(store, clock)

	txn1 := coord.Begin()
	fmt.Printf("  Begin txn: %s\n", txn1.ID)

	coord.Write(txn1, "account:alice", []byte(`{"balance":900}`))
	coord.Write(txn1, "account:bob", []byte(`{"balance":1100}`))
	fmt.Println("  Wrote: alice=900, bob=1100 (transfer $100)")

	committed := coord.Commit(txn1)
	fmt.Printf("  Commit: %v, status=%s\n", committed, txn1.Status)

	readTS := clock.Now()
	if val, ok := store.Get("account:alice", readTS); ok {
		fmt.Printf("  Read alice: %s\n", val)
	}
	if val, ok := store.Get("account:bob", readTS); ok {
		fmt.Printf("  Read bob: %s\n", val)
	}

	txn2 := coord.Begin()
	coord.Write(txn2, "account:alice", []byte(`{"balance":800}`))
	coord.Abort(txn2)
	fmt.Printf("  Aborted txn: status=%s, writes=%d\n", txn2.Status, len(txn2.Writes))
}

// ============================================
// FULL MINI-COCKROACHDB
// ============================================

type MiniCockroachDB struct {
	mu       sync.RWMutex
	nodes    map[int]*CRDBNode
	ranges   *RangeManager
	clock    *HLC
	nodeList []int
}

type CRDBNode struct {
	ID    int
	Store *MVCCStore
	Coord *TxnCoordinator
}

func NewMiniCockroachDB(nodeCount int) *MiniCockroachDB {
	db := &MiniCockroachDB{
		nodes:  make(map[int]*CRDBNode),
		ranges: NewRangeManager(nodeCount),
		clock:  NewHLC(),
	}

	for i := 0; i < nodeCount; i++ {
		store := NewMVCCStore()
		db.nodes[i] = &CRDBNode{
			ID:    i,
			Store: store,
			Coord: NewTxnCoordinator(store, db.clock),
		}
		db.nodeList = append(db.nodeList, i)
	}

	return db
}

func (db *MiniCockroachDB) routeKey(key string) *CRDBNode {
	h := fnv.New32a()
	h.Write([]byte(key))
	idx := int(h.Sum32()) % len(db.nodeList)
	if idx < 0 {
		idx = -idx
	}
	return db.nodes[db.nodeList[idx]]
}

func (db *MiniCockroachDB) Put(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	node := db.routeKey(key)
	txn := node.Coord.Begin()
	node.Coord.Write(txn, key, data)
	node.Coord.Commit(txn)
	return nil
}

func (db *MiniCockroachDB) Get(key string) (json.RawMessage, bool) {
	node := db.routeKey(key)
	ts := db.clock.Now()
	if val, ok := node.Store.Get(key, ts); ok {
		return json.RawMessage(val), true
	}
	return nil, false
}

func (db *MiniCockroachDB) Transfer(fromKey, toKey string, amount float64) error {
	fromNode := db.routeKey(fromKey)

	txn := fromNode.Coord.Begin()

	readTS := db.clock.Now()
	fromVal, ok := fromNode.Store.Get(fromKey, readTS)
	if !ok {
		fromNode.Coord.Abort(txn)
		return fmt.Errorf("source account not found")
	}

	toNode := db.routeKey(toKey)
	toVal, ok := toNode.Store.Get(toKey, readTS)
	if !ok {
		fromNode.Coord.Abort(txn)
		return fmt.Errorf("destination account not found")
	}

	var fromAcct, toAcct map[string]interface{}
	json.Unmarshal(fromVal, &fromAcct)
	json.Unmarshal(toVal, &toAcct)

	fromBal := fromAcct["balance"].(float64)
	toBal := toAcct["balance"].(float64)

	if fromBal < amount {
		fromNode.Coord.Abort(txn)
		return fmt.Errorf("insufficient funds")
	}

	fromAcct["balance"] = fromBal - amount
	toAcct["balance"] = toBal + amount

	fromData, _ := json.Marshal(fromAcct)
	toData, _ := json.Marshal(toAcct)

	fromNode.Coord.Write(txn, fromKey, fromData)
	fromNode.Coord.Write(txn, toKey, toData)
	fromNode.Coord.Commit(txn)

	return nil
}

func demoMiniCockroachDB() {
	db := NewMiniCockroachDB(3)

	db.Put("account:alice", map[string]interface{}{"name": "Alice", "balance": 1000.0})
	db.Put("account:bob", map[string]interface{}{"name": "Bob", "balance": 500.0})
	db.Put("account:charlie", map[string]interface{}{"name": "Charlie", "balance": 750.0})

	if val, ok := db.Get("account:alice"); ok {
		fmt.Printf("  Alice: %s\n", val)
	}
	if val, ok := db.Get("account:bob"); ok {
		fmt.Printf("  Bob: %s\n", val)
	}

	err := db.Transfer("account:alice", "account:bob", 200)
	fmt.Printf("  Transfer alice->bob $200: err=%v\n", err)

	if val, ok := db.Get("account:alice"); ok {
		fmt.Printf("  Alice after: %s\n", val)
	}
	if val, ok := db.Get("account:bob"); ok {
		fmt.Printf("  Bob after: %s\n", val)
	}

	err = db.Transfer("account:bob", "account:charlie", 10000)
	fmt.Printf("  Transfer bob->charlie $10000: err=%v\n", err)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("order:%d", i)
			db.Put(key, map[string]interface{}{
				"id":     i,
				"item":   "widget",
				"amount": float64(i) * 10,
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("order:%d", i)
		if val, ok := db.Get(key); ok {
			fmt.Printf("  %s: %s\n", key, val)
		}
	}
}

// Ensure we use all imported packages
var _ = binary.BigEndian
var _ = bytes.NewBuffer
var _ = rand.Intn
var _ = os.Stdout

/*
SUMMARY - COCKROACHDB: DISTRIBUTED SQL DATABASE IN GO:

WHAT IS COCKROACHDB:
- Distributed SQL database written entirely in Go
- PostgreSQL-compatible, ACID transactions across nodes
- Automatic sharding, rebalancing, and failure recovery
- Used by DoorDash, Netflix, Bose, fintech companies

ARCHITECTURE:
- Layered: SQL -> Transaction (KV) -> Distribution -> Replication -> Storage (Pebble)
- Data split into ~512MB Ranges, each replicated via Raft (3 replicas)
- Leaseholder serves reads, MVCC with hybrid-logical clocks

KEY GO PATTERNS DEMONSTRATED:
- Hybrid-Logical Clocks (HLC): combines physical time with logical counters
- MVCC Storage Engine: timestamped key-value pairs for snapshot isolation
- Range-Based Sharding: automatic splitting/merging of key ranges
- Raft Consensus: leader election, log replication, majority commits
- Distributed Transactions: two-phase commit with write intents

PERFORMANCE TECHNIQUES:
- Pebble storage engine (pure Go LSM-tree, replaced RocksDB)
- Vectorized SQL execution on column batches
- Per-query memory budgets with disk spilling
- Goroutine-per-connection with pgwire protocol
- Batch KV operations to reduce Raft round-trips

GO PHILOSOPHY:
- Single binary deployment, no external dependencies
- Composition through interfaces at each layer
- Explicit error handling with context wrapping
- Goroutines for concurrent range processing
*/
