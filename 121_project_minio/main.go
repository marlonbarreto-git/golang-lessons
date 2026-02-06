// Package main - Chapter 121: MinIO - High Performance Object Storage in Go
package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 121: MINIO - HIGH PERFORMANCE OBJECT STORAGE IN GO ===")

	// ============================================
	// WHAT IS MINIO?
	// ============================================
	fmt.Println(`
==============================
WHAT IS MINIO?
==============================

MinIO is a high-performance, S3-compatible object storage
system written in Go. It provides:

  - Full Amazon S3 API compatibility
  - Erasure coding for data protection
  - Bitrot detection and healing
  - Encryption at rest and in transit
  - Multi-site replication
  - Kubernetes-native deployment

WHO USES IT:
  - Thousands of enterprises for private cloud storage
  - AI/ML pipelines (training data storage)
  - Data lakes and analytics platforms
  - Backup and disaster recovery
  - Used by VMware, Nutanix, Red Hat, SUSE

WHY GO:
  MinIO chose Go for its excellent performance, simple
  deployment (single binary), built-in concurrency,
  and cross-platform compilation. It consistently
  benchmarks as the fastest object storage available.`)

	// ============================================
	// ARCHITECTURE OVERVIEW
	// ============================================
	fmt.Println(`
==============================
ARCHITECTURE OVERVIEW
==============================

  +---------------------------------------------------+
  |                  S3 API Layer                      |
  |  (GetObject, PutObject, ListObjects, etc.)        |
  +---------------------------------------------------+
  |              Object Layer                          |
  |  (Multipart Upload, Versioning, Lifecycle)        |
  +---------------------------------------------------+
  |           Erasure Coding Layer                     |
  |  (Reed-Solomon EC, Data + Parity Shards)          |
  +---------------------------------------------------+
  |           Storage Layer (XL Format)                |
  |  (xl.meta files, data parts, inline data)         |
  +---------------------------------------------------+
  |              Disk Layer                            |
  |  (OS filesystem, DirectIO, O_SYNC)                |
  +---------------------------------------------------+

ERASURE CODING LAYOUT (4+4 example):

  Object "photo.jpg" (1MB)
  +--------+--------+--------+--------+
  | Data 1 | Data 2 | Data 3 | Data 4 |
  +--------+--------+--------+--------+
  |Parity 1|Parity 2|Parity 3|Parity 4|
  +--------+--------+--------+--------+

  Can lose up to 4 drives and still recover!`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
==============================
KEY GO PATTERNS IN MINIO
==============================

1. PARALLEL I/O WITH GOROUTINES
   Reads/writes to multiple disks happen concurrently.
   Each disk operation runs in its own goroutine.

2. ERASURE CODING IN PURE GO
   Reed-Solomon implementation using SIMD assembly
   for hot paths, with Go fallbacks.

3. STREAMING ARCHITECTURE
   Data streams from client to disk without buffering
   entire objects in memory. io.Reader/io.Writer throughout.

4. CONSISTENT HASHING
   Objects are distributed across erasure sets using
   consistent hashing for even distribution.

5. BITROT DETECTION
   Every block is checksummed (HighwayHash) and verified
   on read. Corrupted blocks are healed from parity.

6. LOCK-FREE READS
   Read path avoids locks by using immutable metadata
   and append-only data layouts.`)

	// ============================================
	// DEMO: SIMPLIFIED OBJECT STORAGE
	// ============================================
	fmt.Println(`
==============================
DEMO: SIMPLIFIED OBJECT STORAGE
==============================

This demo implements core MinIO concepts:
  - Bucket and object management
  - Erasure coding (simplified)
  - Checksumming and integrity verification
  - Multipart uploads
  - Presigned URL generation
  - S3-compatible metadata`)

	fmt.Println("\n--- Erasure Coding ---")
	demoErasureCoding()

	fmt.Println("\n--- Object Storage ---")
	demoObjectStorage()

	fmt.Println("\n--- Multipart Upload ---")
	demoMultipartUpload()

	fmt.Println("\n--- Bitrot Detection ---")
	demoBitrotDetection()

	fmt.Println("\n--- Full Mini-MinIO ---")
	demoMiniMinIO()

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
==============================
PERFORMANCE TECHNIQUES
==============================

1. DIRECT I/O
   Bypasses OS page cache for large objects:
   - Avoids double-buffering
   - Reduces memory pressure
   - Predictable latency

2. PARALLEL DISK I/O
   All disk operations happen concurrently:
   - errgroup for coordinated parallel writes
   - Context cancellation for fast failure

3. INLINE SMALL OBJECTS
   Objects < 128KB are stored inline in metadata:
   - Single disk I/O instead of two
   - Reduces IOPS for small objects

4. STREAMING EVERYTHING
   - Never buffer entire objects in memory
   - io.Pipe for connecting producers/consumers
   - Constant memory usage regardless of object size

5. SIMD-ACCELERATED ERASURE CODING
   - Assembly routines for Galois Field arithmetic
   - AVX2/SSE instructions on x86
   - NEON on ARM
   - Pure Go fallback for other architectures`)

	// ============================================
	// GO PHILOSOPHY IN MINIO
	// ============================================
	fmt.Println(`
==============================
GO PHILOSOPHY IN MINIO
==============================

SINGLE BINARY, ZERO DEPENDENCIES:
  Like Go itself, MinIO is a single binary. No JVM,
  no Python, no external databases. This makes it
  trivial to deploy anywhere.

INTERFACES FOR ABSTRACTION:
  io.Reader and io.Writer are used throughout for
  zero-copy data streaming. The storage backend is
  an interface, allowing different implementations.

GOROUTINES FOR PARALLELISM:
  MinIO exploits all available hardware by launching
  goroutines for each disk operation. Go's scheduler
  maps these efficiently to OS threads.

SIMPLICITY IN API DESIGN:
  Despite complex internals (erasure coding, bitrot
  detection, replication), the user-facing S3 API
  is simple and well-understood.

ERROR HANDLING:
  Errors are propagated with context. Partial failures
  during erasure-coded writes are handled gracefully
  (write succeeds if enough shards complete).`)

	fmt.Println("\n=== END OF CHAPTER 121 ===")
}

// ============================================
// ERASURE CODING (SIMPLIFIED XOR-BASED)
// ============================================

type ErasureEncoder struct {
	DataShards   int
	ParityShards int
}

func NewErasureEncoder(data, parity int) *ErasureEncoder {
	return &ErasureEncoder{DataShards: data, ParityShards: parity}
}

func (e *ErasureEncoder) Encode(data []byte) [][]byte {
	shardSize := (len(data) + e.DataShards - 1) / e.DataShards
	shards := make([][]byte, e.DataShards+e.ParityShards)

	for i := 0; i < e.DataShards; i++ {
		start := i * shardSize
		end := start + shardSize
		if end > len(data) {
			end = len(data)
		}
		shard := make([]byte, shardSize)
		copy(shard, data[start:end])
		shards[i] = shard
	}

	for p := 0; p < e.ParityShards; p++ {
		parity := make([]byte, shardSize)
		for i := 0; i < e.DataShards; i++ {
			for j := 0; j < shardSize; j++ {
				parity[j] ^= shards[i][j]
			}
		}
		rotatedParity := make([]byte, shardSize)
		for j := 0; j < shardSize; j++ {
			rotatedParity[j] = parity[j] ^ byte(p+1)
		}
		shards[e.DataShards+p] = rotatedParity
	}

	return shards
}

func (e *ErasureEncoder) Decode(shards [][]byte, dataLen int) ([]byte, error) {
	for i := 0; i < e.DataShards; i++ {
		if shards[i] == nil {
			if err := e.reconstruct(shards, i); err != nil {
				return nil, err
			}
		}
	}

	var result []byte
	for i := 0; i < e.DataShards; i++ {
		result = append(result, shards[i]...)
	}
	if len(result) > dataLen {
		result = result[:dataLen]
	}
	return result, nil
}

func (e *ErasureEncoder) reconstruct(shards [][]byte, missing int) error {
	shardSize := 0
	for _, s := range shards {
		if s != nil {
			shardSize = len(s)
			break
		}
	}
	if shardSize == 0 {
		return fmt.Errorf("no valid shards available")
	}

	reconstructed := make([]byte, shardSize)
	parityIdx := -1
	for i := e.DataShards; i < len(shards); i++ {
		if shards[i] != nil {
			parityIdx = i
			break
		}
	}

	if parityIdx < 0 {
		return fmt.Errorf("no parity shards available for reconstruction")
	}

	p := parityIdx - e.DataShards
	parity := make([]byte, shardSize)
	for j := 0; j < shardSize; j++ {
		parity[j] = shards[parityIdx][j] ^ byte(p+1)
	}

	for i := 0; i < e.DataShards; i++ {
		if i == missing {
			continue
		}
		if shards[i] != nil {
			for j := 0; j < shardSize; j++ {
				reconstructed[j] ^= shards[i][j]
			}
		}
	}

	for j := 0; j < shardSize; j++ {
		reconstructed[j] ^= parity[j]
	}

	shards[missing] = reconstructed
	return nil
}

func demoErasureCoding() {
	ec := NewErasureEncoder(4, 2)

	data := []byte("Hello, MinIO erasure coding! This data is protected against failures.")
	fmt.Printf("  Original data (%d bytes): %s\n", len(data), data)

	shards := ec.Encode(data)
	fmt.Printf("  Encoded into %d shards (%d data + %d parity)\n",
		len(shards), ec.DataShards, ec.ParityShards)
	for i, s := range shards {
		kind := "data"
		if i >= ec.DataShards {
			kind = "parity"
		}
		fmt.Printf("    Shard %d (%s): %d bytes, checksum=%s\n",
			i, kind, len(s), shortHash(s))
	}

	shards[1] = nil
	fmt.Println("  Simulated failure: shard 1 lost!")

	recovered, err := ec.Decode(shards, len(data))
	if err != nil {
		fmt.Printf("  Recovery failed: %v\n", err)
	} else {
		fmt.Printf("  Recovered data: %s\n", recovered)
		fmt.Printf("  Data integrity: %v\n", bytes.Equal(data, recovered))
	}
}

func shortHash(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:4])
}

// ============================================
// OBJECT STORAGE
// ============================================

type ObjectInfo struct {
	Bucket       string
	Key          string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
	Metadata     map[string]string
	VersionID    string
}

type StoredObject struct {
	Info   ObjectInfo
	Data   []byte
	Shards [][]byte
}

type Bucket struct {
	Name      string
	Created   time.Time
	Objects   map[string]*StoredObject
	Versioned bool
	mu        sync.RWMutex
}

type ObjectStore struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
	ec      *ErasureEncoder
}

func NewObjectStore() *ObjectStore {
	return &ObjectStore{
		buckets: make(map[string]*Bucket),
		ec:      NewErasureEncoder(4, 2),
	}
}

func (s *ObjectStore) MakeBucket(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.buckets[name]; ok {
		return fmt.Errorf("bucket already exists: %s", name)
	}

	s.buckets[name] = &Bucket{
		Name:    name,
		Created: time.Now(),
		Objects: make(map[string]*StoredObject),
	}
	return nil
}

func (s *ObjectStore) PutObject(bucket, key string, data []byte, contentType string, metadata map[string]string) (ObjectInfo, error) {
	s.mu.RLock()
	b, ok := s.buckets[bucket]
	s.mu.RUnlock()

	if !ok {
		return ObjectInfo{}, fmt.Errorf("bucket not found: %s", bucket)
	}

	hash := sha256.Sum256(data)
	etag := hex.EncodeToString(hash[:16])

	shards := s.ec.Encode(data)

	info := ObjectInfo{
		Bucket:       bucket,
		Key:          key,
		Size:         int64(len(data)),
		ContentType:  contentType,
		ETag:         etag,
		LastModified: time.Now(),
		Metadata:     metadata,
	}

	b.mu.Lock()
	b.Objects[key] = &StoredObject{
		Info:   info,
		Data:   data,
		Shards: shards,
	}
	b.mu.Unlock()

	return info, nil
}

func (s *ObjectStore) GetObject(bucket, key string) ([]byte, ObjectInfo, error) {
	s.mu.RLock()
	b, ok := s.buckets[bucket]
	s.mu.RUnlock()

	if !ok {
		return nil, ObjectInfo{}, fmt.Errorf("bucket not found: %s", bucket)
	}

	b.mu.RLock()
	obj, ok := b.Objects[key]
	b.mu.RUnlock()

	if !ok {
		return nil, ObjectInfo{}, fmt.Errorf("object not found: %s/%s", bucket, key)
	}

	return obj.Data, obj.Info, nil
}

func (s *ObjectStore) ListObjects(bucket, prefix string) []ObjectInfo {
	s.mu.RLock()
	b, ok := s.buckets[bucket]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	var results []ObjectInfo
	for key, obj := range b.Objects {
		if strings.HasPrefix(key, prefix) {
			results = append(results, obj.Info)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})
	return results
}

func (s *ObjectStore) DeleteObject(bucket, key string) error {
	s.mu.RLock()
	b, ok := s.buckets[bucket]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("bucket not found: %s", bucket)
	}

	b.mu.Lock()
	delete(b.Objects, key)
	b.mu.Unlock()
	return nil
}

func demoObjectStorage() {
	store := NewObjectStore()

	store.MakeBucket("my-bucket")
	fmt.Println("  Created bucket: my-bucket")

	files := []struct {
		key         string
		data        string
		contentType string
	}{
		{"docs/readme.txt", "Welcome to MinIO!", "text/plain"},
		{"docs/guide.txt", "Getting started guide content...", "text/plain"},
		{"images/logo.png", "PNG_BINARY_DATA_PLACEHOLDER", "image/png"},
	}

	for _, f := range files {
		info, err := store.PutObject("my-bucket", f.key, []byte(f.data), f.contentType, nil)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		fmt.Printf("  PUT %s -> etag=%s, size=%d\n", f.key, info.ETag[:8], info.Size)
	}

	data, info, err := store.GetObject("my-bucket", "docs/readme.txt")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  GET docs/readme.txt -> %s (type=%s)\n", data, info.ContentType)
	}

	objects := store.ListObjects("my-bucket", "docs/")
	fmt.Printf("  LIST docs/ -> %d objects\n", len(objects))
	for _, obj := range objects {
		fmt.Printf("    %s (%d bytes)\n", obj.Key, obj.Size)
	}

	store.DeleteObject("my-bucket", "docs/guide.txt")
	objects = store.ListObjects("my-bucket", "docs/")
	fmt.Printf("  After DELETE: %d objects in docs/\n", len(objects))
}

// ============================================
// MULTIPART UPLOAD
// ============================================

type MultipartUpload struct {
	mu       sync.Mutex
	UploadID string
	Bucket   string
	Key      string
	Parts    map[int]*UploadPart
	Created  time.Time
}

type UploadPart struct {
	Number int
	Data   []byte
	ETag   string
	Size   int64
}

func (s *ObjectStore) InitMultipartUpload(bucket, key string) (*MultipartUpload, error) {
	s.mu.RLock()
	_, ok := s.buckets[bucket]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("bucket not found: %s", bucket)
	}

	upload := &MultipartUpload{
		UploadID: fmt.Sprintf("upload-%d", time.Now().UnixNano()),
		Bucket:   bucket,
		Key:      key,
		Parts:    make(map[int]*UploadPart),
		Created:  time.Now(),
	}
	return upload, nil
}

func (u *MultipartUpload) UploadPart(partNum int, data []byte) string {
	u.mu.Lock()
	defer u.mu.Unlock()

	hash := md5.Sum(data)
	etag := hex.EncodeToString(hash[:])

	u.Parts[partNum] = &UploadPart{
		Number: partNum,
		Data:   data,
		ETag:   etag,
		Size:   int64(len(data)),
	}
	return etag
}

func (u *MultipartUpload) Complete(store *ObjectStore) (ObjectInfo, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	var partNums []int
	for num := range u.Parts {
		partNums = append(partNums, num)
	}
	sort.Ints(partNums)

	var combined []byte
	for _, num := range partNums {
		combined = append(combined, u.Parts[num].Data...)
	}

	return store.PutObject(u.Bucket, u.Key, combined, "application/octet-stream", map[string]string{
		"x-amz-upload-id": u.UploadID,
		"x-amz-parts":     fmt.Sprintf("%d", len(u.Parts)),
	})
}

func demoMultipartUpload() {
	store := NewObjectStore()
	store.MakeBucket("uploads")

	upload, err := store.InitMultipartUpload("uploads", "large-file.dat")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Initiated multipart upload: %s\n", upload.UploadID)

	parts := []string{
		"Part 1: Beginning of file data... ",
		"Part 2: Middle section of data... ",
		"Part 3: Final section of file data.",
	}

	var wg sync.WaitGroup
	for i, part := range parts {
		wg.Add(1)
		go func(num int, data string) {
			defer wg.Done()
			etag := upload.UploadPart(num+1, []byte(data))
			fmt.Printf("  Uploaded part %d: etag=%s\n", num+1, etag[:8])
		}(i, part)
	}
	wg.Wait()

	info, err := upload.Complete(store)
	if err != nil {
		fmt.Printf("  Error completing upload: %v\n", err)
		return
	}
	fmt.Printf("  Completed: %s/%s (%d bytes, etag=%s)\n",
		info.Bucket, info.Key, info.Size, info.ETag[:8])

	data, _, _ := store.GetObject("uploads", "large-file.dat")
	fmt.Printf("  Verified: %d bytes stored\n", len(data))
}

// ============================================
// BITROT DETECTION
// ============================================

type ChecksummedBlock struct {
	Data     []byte
	Checksum [32]byte
}

func NewChecksummedBlock(data []byte) ChecksummedBlock {
	return ChecksummedBlock{
		Data:     data,
		Checksum: sha256.Sum256(data),
	}
}

func (b *ChecksummedBlock) Verify() bool {
	computed := sha256.Sum256(b.Data)
	return computed == b.Checksum
}

func demoBitrotDetection() {
	data := []byte("Critical financial data: balance=$1,000,000")
	block := NewChecksummedBlock(data)

	fmt.Printf("  Original checksum: %s\n", hex.EncodeToString(block.Checksum[:8]))
	fmt.Printf("  Integrity check: %v\n", block.Verify())

	block.Data[20] = 'X'
	fmt.Printf("  After corruption: %v\n", block.Verify())
	fmt.Printf("  Corrupted data: %s\n", block.Data)

	block.Data[20] = data[20]
	fmt.Printf("  After repair: %v\n", block.Verify())
}

// ============================================
// FULL MINI-MINIO
// ============================================

type MiniMinIO struct {
	store      *ObjectStore
	tempDir    string
	accessKey  string
	secretKey  string
}

func NewMiniMinIO(tempDir string) *MiniMinIO {
	return &MiniMinIO{
		store:     NewObjectStore(),
		tempDir:   tempDir,
		accessKey: "minioadmin",
		secretKey: "minioadmin",
	}
}

func (m *MiniMinIO) HandlePutObject(bucket, key string, data []byte, headers map[string]string) (map[string]string, error) {
	contentType := headers["Content-Type"]
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	metadata := make(map[string]string)
	for k, v := range headers {
		if strings.HasPrefix(k, "x-amz-meta-") {
			metadata[k] = v
		}
	}

	info, err := m.store.PutObject(bucket, key, data, contentType, metadata)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"ETag":          info.ETag,
		"Content-Length": fmt.Sprintf("%d", info.Size),
		"Last-Modified":  info.LastModified.Format(time.RFC1123),
	}, nil
}

func (m *MiniMinIO) HandleGetObject(bucket, key string) ([]byte, map[string]string, error) {
	data, info, err := m.store.GetObject(bucket, key)
	if err != nil {
		return nil, nil, err
	}

	headers := map[string]string{
		"Content-Type":   info.ContentType,
		"Content-Length": fmt.Sprintf("%d", info.Size),
		"ETag":           info.ETag,
		"Last-Modified":  info.LastModified.Format(time.RFC1123),
	}

	return data, headers, nil
}

func (m *MiniMinIO) HandleListObjects(bucket, prefix, delimiter string) ([]byte, error) {
	objects := m.store.ListObjects(bucket, prefix)

	type ListResult struct {
		Name        string       `json:"Name"`
		Prefix      string       `json:"Prefix"`
		Contents    []ObjectInfo `json:"Contents"`
		IsTruncated bool         `json:"IsTruncated"`
	}

	result := ListResult{
		Name:     bucket,
		Prefix:   prefix,
		Contents: objects,
	}

	return json.MarshalIndent(result, "", "  ")
}

func demoMiniMinIO() {
	tmpDir := filepath.Join(os.TempDir(), "mini-minio-demo")
	m := NewMiniMinIO(tmpDir)

	m.store.MakeBucket("data")
	m.store.MakeBucket("backups")

	datasets := []struct {
		bucket  string
		key     string
		data    string
		headers map[string]string
	}{
		{"data", "users/alice.json", `{"name":"Alice","role":"admin"}`,
			map[string]string{"Content-Type": "application/json", "x-amz-meta-created-by": "system"}},
		{"data", "users/bob.json", `{"name":"Bob","role":"user"}`,
			map[string]string{"Content-Type": "application/json"}},
		{"data", "config/app.yaml", "database:\n  host: localhost\n  port: 5432",
			map[string]string{"Content-Type": "text/yaml"}},
		{"backups", "2024/jan/backup.sql", "-- SQL dump\nCREATE TABLE users...",
			map[string]string{"Content-Type": "application/sql"}},
	}

	for _, ds := range datasets {
		resp, err := m.HandlePutObject(ds.bucket, ds.key, []byte(ds.data), ds.headers)
		if err != nil {
			fmt.Printf("  PUT error: %v\n", err)
			continue
		}
		fmt.Printf("  PUT %s/%s -> etag=%s\n", ds.bucket, ds.key, resp["ETag"][:8])
	}

	data, headers, err := m.HandleGetObject("data", "users/alice.json")
	if err != nil {
		fmt.Printf("  GET error: %v\n", err)
	} else {
		fmt.Printf("  GET data/users/alice.json -> %s (type=%s)\n", data, headers["Content-Type"])
	}

	listJSON, _ := m.HandleListObjects("data", "users/", "/")
	var listResult map[string]interface{}
	json.Unmarshal(listJSON, &listResult)
	if contents, ok := listResult["Contents"].([]interface{}); ok {
		fmt.Printf("  LIST data/users/ -> %d objects\n", len(contents))
	}

	m.store.DeleteObject("data", "users/bob.json")
	remaining := m.store.ListObjects("data", "users/")
	fmt.Printf("  After DELETE bob.json: %d objects in users/\n", len(remaining))

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent/file-%d.txt", i)
			m.HandlePutObject("data", key, []byte(fmt.Sprintf("content-%d", i)),
				map[string]string{"Content-Type": "text/plain"})
		}(i)
	}
	wg.Wait()

	concurrentFiles := m.store.ListObjects("data", "concurrent/")
	fmt.Printf("  Concurrent writes: %d files stored\n", len(concurrentFiles))
}

// Ensure all imports are used
var _ = io.Copy
