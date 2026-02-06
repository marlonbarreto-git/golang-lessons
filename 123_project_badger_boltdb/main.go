// Package main - Chapter 123: Badger and BoltDB - Embedded Key-Value Stores in Go
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 123: BADGER AND BOLTDB - EMBEDDED KEY-VALUE STORES IN GO ===")

	// ============================================
	// WHAT ARE BADGER AND BOLTDB?
	// ============================================
	fmt.Println(`
==============================
WHAT ARE BADGER AND BOLTDB?
==============================

BoltDB and Badger are embedded key-value databases written
in pure Go. They provide persistent storage without needing
a separate database server process.

BOLTDB (bbolt):
  - B+ tree based storage engine
  - ACID transactions with MVCC
  - Memory-mapped file for reads (mmap)
  - Single-writer, multiple-reader model
  - Originally by Ben Johnson, now maintained as bbolt
  - Used by: etcd, Consul, InfluxDB, CockroachDB (metadata)

BADGER (by Dgraph):
  - LSM tree + value log (WiscKey design)
  - Separates keys from values for efficiency
  - Concurrent reads and writes
  - Built-in garbage collection for value log
  - Supports transactions and iteration
  - Used by: Dgraph, Milvus, Dkron

WHEN TO USE WHICH:
  +-------------------+------------------+------------------+
  | Criteria          | BoltDB (bbolt)   | Badger           |
  +-------------------+------------------+------------------+
  | Write throughput  | Moderate         | High             |
  | Read latency      | Very low (mmap)  | Low              |
  | Write pattern     | Single writer    | Concurrent       |
  | Data structure    | B+ tree          | LSM tree + vlog  |
  | Memory usage      | Low              | Higher           |
  | Best for          | Read-heavy       | Write-heavy      |
  | Simplicity        | Very simple      | More complex     |
  +-------------------+------------------+------------------+`)

	// ============================================
	// ARCHITECTURE: BOLTDB
	// ============================================
	fmt.Println(`
==============================
ARCHITECTURE: BOLTDB (B+ TREE)
==============================

  +------------------------------------------+
  |              API Layer                    |
  |  (DB.Update, DB.View, Bucket, Cursor)    |
  +------------------------------------------+
  |           Transaction Layer               |
  |  (Read-only tx, Read-write tx, MVCC)     |
  +------------------------------------------+
  |            B+ Tree Engine                 |
  +------------------------------------------+
  |  Page Layout:                             |
  |  +--------+--------+--------+--------+   |
  |  | Meta   | Free   | Branch | Leaf   |   |
  |  | Pages  | List   | Pages  | Pages  |   |
  |  +--------+--------+--------+--------+   |
  +------------------------------------------+
  |          Memory-Mapped File (mmap)        |
  +------------------------------------------+
  |              OS File System               |
  +------------------------------------------+

B+ TREE STRUCTURE:
               [20 | 40]         <- Branch (internal)
              /    |    \
     [5|10|15] [25|30] [45|50]   <- Leaf nodes
     (data)    (data)  (data)`)

	// ============================================
	// ARCHITECTURE: BADGER
	// ============================================
	fmt.Println(`
==============================
ARCHITECTURE: BADGER (LSM + VLOG)
==============================

  +------------------------------------------+
  |              API Layer                    |
  |  (DB.Update, DB.View, Txn, Iterator)     |
  +------------------------------------------+
  |           Transaction Layer               |
  |  (Oracle for timestamps, conflict detect) |
  +------------------------------------------+
  |     +-------------------+  +-----------+ |
  |     |   LSM Tree        |  | Value Log | |
  |     |   (keys + ptrs)   |  | (values)  | |
  |     +-------------------+  +-----------+ |
  |     | MemTable (active) |  | vlog file | |
  |     | MemTable (immut.) |  | vlog file | |
  |     | Level 0 SSTs      |  | vlog file | |
  |     | Level 1 SSTs      |  +-----------+ |
  |     | Level 2 SSTs      |                |
  |     +-------------------+                |
  +------------------------------------------+

KEY INSIGHT (WiscKey paper):
  Only keys are stored in the LSM tree.
  Values are stored separately in a value log.
  This reduces write amplification significantly
  because compaction only moves small keys, not
  large values.`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
==============================
KEY GO PATTERNS
==============================

1. MEMORY-MAPPED I/O (BoltDB)
   Uses syscall.Mmap for zero-copy reads directly
   from the OS page cache. Go's unsafe.Pointer
   converts mmap'd regions to Go structs.

2. WRITE-AHEAD LOG (Both)
   All writes go to a log first, then to the main
   data structure. Ensures durability on crash.

3. TRANSACTION CLOSURES
   Both use a functional transaction API:
     db.Update(func(tx *Tx) error { ... })
     db.View(func(tx *Tx) error { ... })
   This ensures transactions are always committed
   or rolled back, using Go's defer pattern.

4. ITERATOR PATTERN (Badger)
   Prefix-based iteration using Go iterators.
   Memory-efficient scanning of large datasets.

5. SYNC.RWMUTEX FOR MVCC
   Read transactions can run concurrently.
   Write transactions are serialized.

6. GOROUTINE-BASED COMPACTION (Badger)
   Background goroutines handle LSM compaction
   and value log garbage collection.`)

	// ============================================
	// DEMO: SIMPLIFIED EMBEDDED KV STORES
	// ============================================
	fmt.Println(`
==============================
DEMO: SIMPLIFIED EMBEDDED KV STORES
==============================

This demo implements core concepts from both:
  - B+ tree (BoltDB-like) with buckets and cursors
  - LSM tree + value log (Badger-like) with compaction
  - ACID transactions for both
  - Iteration and prefix scanning
  - Write-ahead logging`)

	fmt.Println("\n--- B+ Tree (BoltDB-style) ---")
	demoBPlusTree()

	fmt.Println("\n--- LSM Tree (Badger-style) ---")
	demoLSMTree()

	fmt.Println("\n--- BoltDB-like API ---")
	demoBoltDBAPI()

	fmt.Println("\n--- Badger-like API ---")
	demoBadgerAPI()

	fmt.Println("\n--- Write-Ahead Log ---")
	demoWAL()

	fmt.Println("\n--- Full Mini Embedded DB ---")
	demoMiniEmbeddedDB()

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
==============================
PERFORMANCE TECHNIQUES
==============================

BOLTDB:
  1. Memory-mapped reads (zero syscall overhead)
  2. Copy-on-write B+ tree (no locks for readers)
  3. Page-aligned I/O for OS optimization
  4. Freelist for page reuse
  5. Batch writes for throughput

BADGER:
  1. Key-value separation (WiscKey design)
  2. Bloom filters on SSTs for negative lookups
  3. Block cache for frequently accessed data
  4. Concurrent memtable writes
  5. Parallel compaction goroutines
  6. Value log GC to reclaim space
  7. Stream framework for fast iteration`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
==============================
GO PHILOSOPHY
==============================

PURE GO, NO CGO:
  Both BoltDB and Badger are written in pure Go.
  No C dependencies, no CGo overhead. This means
  easy cross-compilation, predictable performance,
  and simple deployment.

SIMPLE API:
  The transaction closure pattern:
    db.Update(func(tx) error { ... })
  is idiomatic Go. It prevents resource leaks,
  ensures cleanup, and makes the API hard to misuse.

EMBED IN YOUR APPLICATION:
  No separate server process. Import as a Go package,
  open a file, and you have a database. This follows
  Go's philosophy of simple, composable tools.

INTERFACES FOR TESTING:
  Both expose interfaces that can be mocked for
  testing, following Go's interface-based design.`)

	fmt.Println("\n=== END OF CHAPTER 123 ===")
}

// ============================================
// B+ TREE (BOLTDB-STYLE)
// ============================================

type BPlusTree struct {
	mu   sync.RWMutex
	root *bpNode
	order int
}

type bpNode struct {
	keys     []string
	values   [][]byte
	children []*bpNode
	isLeaf   bool
	next     *bpNode
}

func NewBPlusTree(order int) *BPlusTree {
	return &BPlusTree{
		root:  &bpNode{isLeaf: true},
		order: order,
	}
}

func (t *BPlusTree) Get(key string) ([]byte, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.findLeaf(key)
	for i, k := range node.keys {
		if k == key {
			return node.values[i], true
		}
	}
	return nil, false
}

func (t *BPlusTree) Put(key string, value []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.findLeaf(key)

	for i, k := range node.keys {
		if k == key {
			node.values[i] = value
			return
		}
	}

	idx := sort.SearchStrings(node.keys, key)
	node.keys = append(node.keys, "")
	copy(node.keys[idx+1:], node.keys[idx:])
	node.keys[idx] = key

	node.values = append(node.values, nil)
	copy(node.values[idx+1:], node.values[idx:])
	node.values[idx] = value

	if len(node.keys) > t.order {
		t.splitLeaf(node)
	}
}

func (t *BPlusTree) Delete(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.findLeaf(key)
	for i, k := range node.keys {
		if k == key {
			node.keys = append(node.keys[:i], node.keys[i+1:]...)
			node.values = append(node.values[:i], node.values[i+1:]...)
			return true
		}
	}
	return false
}

func (t *BPlusTree) findLeaf(key string) *bpNode {
	node := t.root
	for !node.isLeaf {
		idx := sort.SearchStrings(node.keys, key)
		if idx < len(node.children) {
			node = node.children[idx]
		} else {
			node = node.children[len(node.children)-1]
		}
	}
	return node
}

func (t *BPlusTree) splitLeaf(node *bpNode) {
	mid := len(node.keys) / 2

	newNode := &bpNode{
		isLeaf: true,
		keys:   make([]string, len(node.keys[mid:])),
		values: make([][]byte, len(node.values[mid:])),
		next:   node.next,
	}
	copy(newNode.keys, node.keys[mid:])
	copy(newNode.values, node.values[mid:])

	node.keys = node.keys[:mid]
	node.values = node.values[:mid]
	node.next = newNode

	if node == t.root {
		newRoot := &bpNode{
			isLeaf:   false,
			keys:     []string{newNode.keys[0]},
			children: []*bpNode{node, newNode},
		}
		t.root = newRoot
	}
}

func (t *BPlusTree) RangeScan(startKey, endKey string) []struct{ Key string; Value []byte } {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var results []struct{ Key string; Value []byte }

	node := t.findLeaf(startKey)
	for node != nil {
		for i, k := range node.keys {
			if k >= startKey && (endKey == "" || k <= endKey) {
				results = append(results, struct{ Key string; Value []byte }{k, node.values[i]})
			}
			if endKey != "" && k > endKey {
				return results
			}
		}
		node = node.next
	}
	return results
}

func demoBPlusTree() {
	tree := NewBPlusTree(4)

	entries := []struct{ key, val string }{
		{"user:001", "Alice"},
		{"user:002", "Bob"},
		{"user:003", "Charlie"},
		{"user:004", "Diana"},
		{"user:005", "Eve"},
		{"user:006", "Frank"},
	}

	for _, e := range entries {
		tree.Put(e.key, []byte(e.val))
	}
	fmt.Printf("  Inserted %d entries into B+ tree\n", len(entries))

	if val, ok := tree.Get("user:003"); ok {
		fmt.Printf("  Get 'user:003': %s\n", val)
	}

	tree.Put("user:003", []byte("Charles"))
	if val, ok := tree.Get("user:003"); ok {
		fmt.Printf("  Updated 'user:003': %s\n", val)
	}

	results := tree.RangeScan("user:002", "user:005")
	fmt.Printf("  Range scan [user:002, user:005]: %d results\n", len(results))
	for _, r := range results {
		fmt.Printf("    %s = %s\n", r.Key, r.Value)
	}

	tree.Delete("user:004")
	_, found := tree.Get("user:004")
	fmt.Printf("  After delete user:004: found=%v\n", found)
}

// ============================================
// LSM TREE (BADGER-STYLE)
// ============================================

type LSMTree struct {
	mu         sync.RWMutex
	memTable   *MemTable
	immutable  []*MemTable
	levels     [][]*SSTable
	maxMemSize int
	nextSST    int
}

type MemTable struct {
	entries map[string]memEntry
	size    int
}

type memEntry struct {
	Value     []byte
	Deleted   bool
	Timestamp int64
}

type SSTable struct {
	ID      int
	Level   int
	Entries []sstEntry
	Bloom   map[string]bool
}

type sstEntry struct {
	Key       string
	Value     []byte
	Deleted   bool
	Timestamp int64
}

func NewLSMTree() *LSMTree {
	return &LSMTree{
		memTable:   &MemTable{entries: make(map[string]memEntry)},
		levels:     make([][]*SSTable, 4),
		maxMemSize: 10,
	}
}

func (l *LSMTree) Put(key string, value []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.memTable.entries[key] = memEntry{
		Value:     value,
		Timestamp: time.Now().UnixNano(),
	}
	l.memTable.size++

	if l.memTable.size >= l.maxMemSize {
		l.flushMemTable()
	}
}

func (l *LSMTree) Delete(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.memTable.entries[key] = memEntry{
		Deleted:   true,
		Timestamp: time.Now().UnixNano(),
	}
	l.memTable.size++
}

func (l *LSMTree) Get(key string) ([]byte, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if entry, ok := l.memTable.entries[key]; ok {
		if entry.Deleted {
			return nil, false
		}
		return entry.Value, true
	}

	for i := len(l.immutable) - 1; i >= 0; i-- {
		if entry, ok := l.immutable[i].entries[key]; ok {
			if entry.Deleted {
				return nil, false
			}
			return entry.Value, true
		}
	}

	for level := 0; level < len(l.levels); level++ {
		for i := len(l.levels[level]) - 1; i >= 0; i-- {
			sst := l.levels[level][i]
			if sst.Bloom != nil && !sst.Bloom[key] {
				continue
			}
			for _, entry := range sst.Entries {
				if entry.Key == key {
					if entry.Deleted {
						return nil, false
					}
					return entry.Value, true
				}
			}
		}
	}

	return nil, false
}

func (l *LSMTree) flushMemTable() {
	old := l.memTable
	l.memTable = &MemTable{entries: make(map[string]memEntry)}

	var entries []sstEntry
	bloom := make(map[string]bool)
	for k, v := range old.entries {
		entries = append(entries, sstEntry{
			Key:       k,
			Value:     v.Value,
			Deleted:   v.Deleted,
			Timestamp: v.Timestamp,
		})
		bloom[k] = true
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	sst := &SSTable{
		ID:      l.nextSST,
		Level:   0,
		Entries: entries,
		Bloom:   bloom,
	}
	l.nextSST++
	l.levels[0] = append(l.levels[0], sst)

	if len(l.levels[0]) >= 4 {
		l.compact(0)
	}
}

func (l *LSMTree) compact(level int) {
	if level >= len(l.levels)-1 {
		return
	}

	var merged []sstEntry
	seen := make(map[string]bool)

	for i := len(l.levels[level]) - 1; i >= 0; i-- {
		for _, entry := range l.levels[level][i].Entries {
			if !seen[entry.Key] {
				merged = append(merged, entry)
				seen[entry.Key] = true
			}
		}
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Key < merged[j].Key
	})

	bloom := make(map[string]bool)
	for _, e := range merged {
		bloom[e.Key] = true
	}

	newSST := &SSTable{
		ID:      l.nextSST,
		Level:   level + 1,
		Entries: merged,
		Bloom:   bloom,
	}
	l.nextSST++

	l.levels[level] = nil
	l.levels[level+1] = append(l.levels[level+1], newSST)
}

func (l *LSMTree) Stats() map[string]int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := map[string]int{
		"memtable_size": l.memTable.size,
		"immutable":     len(l.immutable),
	}
	for i, level := range l.levels {
		stats[fmt.Sprintf("level_%d_ssts", i)] = len(level)
	}
	return stats
}

func demoLSMTree() {
	lsm := NewLSMTree()

	for i := 0; i < 15; i++ {
		key := fmt.Sprintf("key:%03d", i)
		val := fmt.Sprintf("value-%d", i)
		lsm.Put(key, []byte(val))
	}

	stats := lsm.Stats()
	fmt.Printf("  After 15 inserts: memtable=%d, L0 SSTs=%d\n",
		stats["memtable_size"], stats["level_0_ssts"])

	if val, ok := lsm.Get("key:005"); ok {
		fmt.Printf("  Get 'key:005': %s\n", val)
	}

	lsm.Put("key:005", []byte("updated-value-5"))
	if val, ok := lsm.Get("key:005"); ok {
		fmt.Printf("  Updated 'key:005': %s\n", val)
	}

	lsm.Delete("key:003")
	_, found := lsm.Get("key:003")
	fmt.Printf("  After delete key:003: found=%v (tombstone)\n", found)

	if val, ok := lsm.Get("key:010"); ok {
		fmt.Printf("  Get 'key:010': %s (from SSTable)\n", val)
	}
}

// ============================================
// BOLTDB-LIKE API
// ============================================

type BoltDB struct {
	mu      sync.RWMutex
	buckets map[string]*BoltBucket
}

type BoltBucket struct {
	tree *BPlusTree
}

type BoltTx struct {
	db       *BoltDB
	writable bool
}

func NewBoltDB() *BoltDB {
	return &BoltDB{
		buckets: make(map[string]*BoltBucket),
	}
}

func (db *BoltDB) Update(fn func(tx *BoltTx) error) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx := &BoltTx{db: db, writable: true}
	if err := fn(tx); err != nil {
		return err
	}
	return nil
}

func (db *BoltDB) View(fn func(tx *BoltTx) error) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	tx := &BoltTx{db: db, writable: false}
	return fn(tx)
}

func (tx *BoltTx) CreateBucketIfNotExists(name string) *BoltBucket {
	if !tx.writable {
		return nil
	}
	if b, ok := tx.db.buckets[name]; ok {
		return b
	}
	b := &BoltBucket{tree: NewBPlusTree(4)}
	tx.db.buckets[name] = b
	return b
}

func (tx *BoltTx) Bucket(name string) *BoltBucket {
	return tx.db.buckets[name]
}

func (b *BoltBucket) Put(key, value []byte) error {
	b.tree.Put(string(key), value)
	return nil
}

func (b *BoltBucket) Get(key []byte) []byte {
	if val, ok := b.tree.Get(string(key)); ok {
		return val
	}
	return nil
}

func (b *BoltBucket) Delete(key []byte) error {
	b.tree.Delete(string(key))
	return nil
}

func demoBoltDBAPI() {
	db := NewBoltDB()

	err := db.Update(func(tx *BoltTx) error {
		b := tx.CreateBucketIfNotExists("users")

		users := []struct{ id, name string }{
			{"1", "Alice"},
			{"2", "Bob"},
			{"3", "Charlie"},
		}
		for _, u := range users {
			data, _ := json.Marshal(map[string]string{"name": u.name})
			b.Put([]byte("user:"+u.id), data)
		}
		return nil
	})
	fmt.Printf("  BoltDB Update (write 3 users): err=%v\n", err)

	err = db.View(func(tx *BoltTx) error {
		b := tx.Bucket("users")
		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		val := b.Get([]byte("user:2"))
		fmt.Printf("  BoltDB View (read user:2): %s\n", val)
		return nil
	})
	fmt.Printf("  BoltDB View: err=%v\n", err)

	err = db.Update(func(tx *BoltTx) error {
		b := tx.Bucket("users")
		return b.Delete([]byte("user:3"))
	})

	db.View(func(tx *BoltTx) error {
		b := tx.Bucket("users")
		val := b.Get([]byte("user:3"))
		fmt.Printf("  After delete user:3: value=%v\n", val)
		return nil
	})
}

// ============================================
// BADGER-LIKE API
// ============================================

type BadgerDB struct {
	lsm *LSMTree
}

type BadgerTxn struct {
	db       *BadgerDB
	writable bool
	pending  map[string][]byte
	deletes  map[string]bool
}

func NewBadgerDB() *BadgerDB {
	return &BadgerDB{lsm: NewLSMTree()}
}

func (db *BadgerDB) Update(fn func(txn *BadgerTxn) error) error {
	txn := &BadgerTxn{
		db:       db,
		writable: true,
		pending:  make(map[string][]byte),
		deletes:  make(map[string]bool),
	}

	if err := fn(txn); err != nil {
		return err
	}

	for key, val := range txn.pending {
		db.lsm.Put(key, val)
	}
	for key := range txn.deletes {
		db.lsm.Delete(key)
	}

	return nil
}

func (db *BadgerDB) View(fn func(txn *BadgerTxn) error) error {
	txn := &BadgerTxn{
		db:       db,
		writable: false,
	}
	return fn(txn)
}

func (txn *BadgerTxn) Set(key, value []byte) error {
	if !txn.writable {
		return fmt.Errorf("read-only transaction")
	}
	txn.pending[string(key)] = value
	return nil
}

func (txn *BadgerTxn) Get(key []byte) ([]byte, error) {
	if txn.writable {
		if val, ok := txn.pending[string(key)]; ok {
			return val, nil
		}
	}
	if val, ok := txn.db.lsm.Get(string(key)); ok {
		return val, nil
	}
	return nil, fmt.Errorf("key not found")
}

func (txn *BadgerTxn) DeleteKey(key []byte) error {
	if !txn.writable {
		return fmt.Errorf("read-only transaction")
	}
	txn.deletes[string(key)] = true
	return nil
}

func demoBadgerAPI() {
	db := NewBadgerDB()

	err := db.Update(func(txn *BadgerTxn) error {
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("item:%03d", i)
			val := fmt.Sprintf(`{"name":"item-%d","price":%d}`, i, (i+1)*10)
			if err := txn.Set([]byte(key), []byte(val)); err != nil {
				return err
			}
		}
		return nil
	})
	fmt.Printf("  Badger Update (write 5 items): err=%v\n", err)

	err = db.View(func(txn *BadgerTxn) error {
		val, err := txn.Get([]byte("item:002"))
		if err != nil {
			return err
		}
		fmt.Printf("  Badger View (read item:002): %s\n", val)
		return nil
	})
	fmt.Printf("  Badger View: err=%v\n", err)

	db.Update(func(txn *BadgerTxn) error {
		return txn.DeleteKey([]byte("item:003"))
	})

	db.View(func(txn *BadgerTxn) error {
		_, err := txn.Get([]byte("item:003"))
		fmt.Printf("  After delete item:003: err=%v\n", err)
		return nil
	})
}

// ============================================
// WRITE-AHEAD LOG
// ============================================

type WALEntry struct {
	Sequence  uint64
	Operation byte
	Key       []byte
	Value     []byte
	Checksum  uint32
}

type WAL struct {
	mu      sync.Mutex
	entries []WALEntry
	nextSeq uint64
}

func NewWAL() *WAL {
	return &WAL{nextSeq: 1}
}

func (w *WAL) Append(op byte, key, value []byte) uint64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	var buf bytes.Buffer
	buf.WriteByte(op)
	buf.Write(key)
	buf.Write(value)
	checksum := crc32.ChecksumIEEE(buf.Bytes())

	entry := WALEntry{
		Sequence:  w.nextSeq,
		Operation: op,
		Key:       key,
		Value:     value,
		Checksum:  checksum,
	}
	w.entries = append(w.entries, entry)
	w.nextSeq++
	return entry.Sequence
}

func (w *WAL) Verify() (int, int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	valid, corrupted := 0, 0
	for _, entry := range w.entries {
		var buf bytes.Buffer
		buf.WriteByte(entry.Operation)
		buf.Write(entry.Key)
		buf.Write(entry.Value)
		expected := crc32.ChecksumIEEE(buf.Bytes())

		if expected == entry.Checksum {
			valid++
		} else {
			corrupted++
		}
	}
	return valid, corrupted
}

func (w *WAL) Replay() []WALEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	result := make([]WALEntry, len(w.entries))
	copy(result, w.entries)
	return result
}

func demoWAL() {
	wal := NewWAL()

	seq1 := wal.Append(1, []byte("user:1"), []byte("Alice"))
	seq2 := wal.Append(1, []byte("user:2"), []byte("Bob"))
	seq3 := wal.Append(2, []byte("user:1"), nil)
	fmt.Printf("  WAL entries: seq=%d, %d, %d\n", seq1, seq2, seq3)

	valid, corrupted := wal.Verify()
	fmt.Printf("  WAL integrity: valid=%d, corrupted=%d\n", valid, corrupted)

	entries := wal.Replay()
	fmt.Printf("  WAL replay: %d entries\n", len(entries))
	for _, e := range entries {
		op := "PUT"
		if e.Operation == 2 {
			op = "DEL"
		}
		fmt.Printf("    [seq=%d] %s key=%s val=%s checksum=%08x\n",
			e.Sequence, op, e.Key, e.Value, e.Checksum)
	}
}

// ============================================
// FULL MINI EMBEDDED DB
// ============================================

type MiniEmbeddedDB struct {
	mu      sync.RWMutex
	bolt    *BoltDB
	badger  *BadgerDB
	wal     *WAL
	engine  string
}

func NewMiniEmbeddedDB(engine string) *MiniEmbeddedDB {
	db := &MiniEmbeddedDB{
		wal:    NewWAL(),
		engine: engine,
	}
	switch engine {
	case "bolt":
		db.bolt = NewBoltDB()
	case "badger":
		db.badger = NewBadgerDB()
	}
	return db
}

func (db *MiniEmbeddedDB) Put(bucket, key string, value []byte) error {
	db.wal.Append(1, []byte(bucket+"/"+key), value)

	switch db.engine {
	case "bolt":
		return db.bolt.Update(func(tx *BoltTx) error {
			b := tx.CreateBucketIfNotExists(bucket)
			return b.Put([]byte(key), value)
		})
	case "badger":
		return db.badger.Update(func(txn *BadgerTxn) error {
			return txn.Set([]byte(bucket+":"+key), value)
		})
	}
	return fmt.Errorf("unknown engine: %s", db.engine)
}

func (db *MiniEmbeddedDB) Get(bucket, key string) ([]byte, error) {
	switch db.engine {
	case "bolt":
		var result []byte
		err := db.bolt.View(func(tx *BoltTx) error {
			b := tx.Bucket(bucket)
			if b == nil {
				return fmt.Errorf("bucket not found: %s", bucket)
			}
			result = b.Get([]byte(key))
			if result == nil {
				return fmt.Errorf("key not found: %s", key)
			}
			return nil
		})
		return result, err
	case "badger":
		var result []byte
		err := db.badger.View(func(txn *BadgerTxn) error {
			val, err := txn.Get([]byte(bucket + ":" + key))
			if err != nil {
				return err
			}
			result = val
			return nil
		})
		return result, err
	}
	return nil, fmt.Errorf("unknown engine: %s", db.engine)
}

func demoMiniEmbeddedDB() {
	for _, engine := range []string{"bolt", "badger"} {
		fmt.Printf("\n  === Engine: %s ===\n", strings.ToUpper(engine))
		db := NewMiniEmbeddedDB(engine)

		users := []struct{ id, data string }{
			{"alice", `{"name":"Alice","age":30}`},
			{"bob", `{"name":"Bob","age":25}`},
			{"charlie", `{"name":"Charlie","age":35}`},
		}

		for _, u := range users {
			err := db.Put("users", u.id, []byte(u.data))
			if err != nil {
				fmt.Printf("    Error: %v\n", err)
			}
		}
		fmt.Printf("    Wrote %d users\n", len(users))

		val, err := db.Get("users", "bob")
		if err != nil {
			fmt.Printf("    Get bob: err=%v\n", err)
		} else {
			fmt.Printf("    Get bob: %s\n", val)
		}

		valid, corrupted := db.wal.Verify()
		fmt.Printf("    WAL: %d valid entries, %d corrupted\n", valid, corrupted)
	}
}

// Ensure all imports are used
var _ = binary.BigEndian
var _ = os.Stdout
