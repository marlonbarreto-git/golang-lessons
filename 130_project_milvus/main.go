// Package main - Chapter 130: Milvus Internals
package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 130: MILVUS - VECTOR DATABASE FOR AI IN GO ===")

	// ============================================
	// WHAT IS MILVUS?
	// ============================================
	fmt.Println(`
==============================
WHAT IS MILVUS?
==============================

Milvus is the most popular open-source vector database, built in Go.
It stores, indexes, and searches high-dimensional vectors at scale,
powering AI applications like RAG, image search, and recommendations.

WHAT IT DOES:
  - Stores billions of vectors (embeddings from ML models)
  - Similarity search (find nearest neighbors)
  - Multiple index types (HNSW, IVF, FLAT, etc.)
  - Hybrid search (vector + scalar filtering)
  - Real-time insert and search

ARCHITECTURE:
  +------------------+     +------------------+
  |   SDK Client     |---->|   Proxy Node     |
  |   (Python/Go/..) |     |   (gRPC API)     |
  +------------------+     +--------+---------+
                                    |
                    +---------------+---------------+
                    |               |               |
           +--------v------+ +-----v-------+ +-----v-------+
           |  Query Node   | | Data Node   | | Index Node  |
           |  (search)     | | (insert)    | | (build idx) |
           +---------------+ +-------------+ +-------------+
                    |               |               |
              +-----v---------------v---------------v-----+
              |            Object Storage (S3/MinIO)       |
              |            + Message Queue (Pulsar/Kafka)   |
              +---------------------------------------------+
                    |
              +-----v-----------+
              |   Root Coord    |
              |   (metadata,    |
              |    timestamps)  |
              +--+-----------+--+
                 |           |
           +-----v---+ +----v------+
           | Data    | | Query     |
           | Coord   | | Coord     |
           +---------+ +-----------+

KEY CONCEPTS:
  Collection:   Like a table, holds vectors + scalar fields
  Segment:      Storage unit within a collection
  Index:        Search acceleration structure (HNSW, IVF, etc.)
  Embedding:    Dense vector from ML model (e.g., 768-dim from BERT)
  Similarity:   Cosine, L2 (Euclidean), Inner Product

USE CASES:
  - RAG (Retrieval Augmented Generation) for LLMs
  - Semantic search (find similar documents/images)
  - Recommendation systems
  - Anomaly detection
  - Drug discovery (molecular similarity)

WHY GO:
  Go handles the distributed coordination layer: gRPC services,
  metadata management, segment scheduling, and cluster coordination.
  Actual vector computation uses CGo bindings to C++ (knowhere engine)
  for SIMD-optimized similarity search.`)

	// ============================================
	// VECTOR STORAGE AND SIMILARITY
	// ============================================
	fmt.Println("\n--- Vector Storage and Similarity ---")
	demoVectorSimilarity()

	// ============================================
	// COLLECTION MANAGEMENT
	// ============================================
	fmt.Println("\n--- Collection Management ---")
	demoCollectionManager()

	// ============================================
	// HNSW INDEX
	// ============================================
	fmt.Println("\n--- HNSW-like Index ---")
	demoHNSWIndex()

	// ============================================
	// SEGMENT MANAGEMENT
	// ============================================
	fmt.Println("\n--- Segment Management ---")
	demoSegmentManager()

	// ============================================
	// QUERY ENGINE
	// ============================================
	fmt.Println("\n--- Query Engine ---")
	demoQueryEngine()

	// ============================================
	// FULL MINI-MILVUS
	// ============================================
	fmt.Println("\n--- Mini-Milvus Integration ---")
	demoMiniMilvus()
}

// ============================================
// VECTOR STORAGE AND SIMILARITY
// ============================================

type Vector []float64

func NewRandomVector(dim int, rng *rand.Rand) Vector {
	v := make(Vector, dim)
	for i := range v {
		v[i] = rng.Float64()*2 - 1
	}
	return v
}

func CosineSimilarity(a, b Vector) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func EuclideanDistance(a, b Vector) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}
	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func InnerProduct(a, b Vector) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}

func NormalizeVector(v Vector) Vector {
	var norm float64
	for _, val := range v {
		norm += val * val
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return v
	}
	result := make(Vector, len(v))
	for i := range v {
		result[i] = v[i] / norm
	}
	return result
}

func demoVectorSimilarity() {
	docVectors := map[string]Vector{
		"Go concurrency":  {0.8, 0.1, 0.3, 0.9, 0.2, 0.1, 0.7, 0.4},
		"Python async":    {0.7, 0.2, 0.4, 0.8, 0.3, 0.2, 0.5, 0.3},
		"Rust ownership":  {0.3, 0.9, 0.7, 0.2, 0.8, 0.6, 0.1, 0.4},
		"Java threads":    {0.6, 0.3, 0.3, 0.7, 0.2, 0.1, 0.6, 0.5},
		"Machine learning": {0.1, 0.2, 0.8, 0.1, 0.9, 0.7, 0.2, 0.3},
	}

	query := Vector{0.7, 0.15, 0.35, 0.85, 0.25, 0.15, 0.6, 0.35}
	fmt.Println("  Query: 'Go async patterns' (similar to Go concurrency)")

	type result struct {
		name       string
		cosine     float64
		euclidean  float64
		innerProd  float64
	}

	var results []result
	for name, vec := range docVectors {
		results = append(results, result{
			name:      name,
			cosine:    CosineSimilarity(query, vec),
			euclidean: EuclideanDistance(query, vec),
			innerProd: InnerProduct(query, vec),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].cosine > results[j].cosine
	})

	fmt.Println("  Results (sorted by cosine similarity):")
	for i, r := range results {
		os.Stdout.WriteString(fmt.Sprintf("    %d. %-20s cosine=%.4f  L2=%.4f  IP=%.4f\n",
			i+1, r.name, r.cosine, r.euclidean, r.innerProd))
	}

	v := Vector{3.0, 4.0}
	normalized := NormalizeVector(v)
	os.Stdout.WriteString(fmt.Sprintf("  Normalize [3,4] -> [%.4f, %.4f] (magnitude=%.4f)\n",
		normalized[0], normalized[1],
		math.Sqrt(normalized[0]*normalized[0]+normalized[1]*normalized[1])))
}

// ============================================
// COLLECTION MANAGEMENT
// ============================================

type FieldType int

const (
	FieldInt64 FieldType = iota
	FieldFloat64
	FieldVarChar
	FieldFloatVector
)

func (f FieldType) String() string {
	switch f {
	case FieldInt64:
		return "Int64"
	case FieldFloat64:
		return "Float64"
	case FieldVarChar:
		return "VarChar"
	case FieldFloatVector:
		return "FloatVector"
	default:
		return "Unknown"
	}
}

type FieldSchema struct {
	Name       string
	Type       FieldType
	IsPrimary  bool
	AutoID     bool
	Dimension  int
	MaxLength  int
}

type CollectionSchema struct {
	Name        string
	Description string
	Fields      []FieldSchema
}

type Collection struct {
	mu         sync.RWMutex
	Schema     CollectionSchema
	Segments   []*Segment
	RowCount   int64
	CreatedAt  time.Time
	Loaded     bool
	IndexBuilt bool
}

type CollectionManager struct {
	mu          sync.RWMutex
	collections map[string]*Collection
}

func NewCollectionManager() *CollectionManager {
	return &CollectionManager{
		collections: make(map[string]*Collection),
	}
}

func (cm *CollectionManager) CreateCollection(schema CollectionSchema) (*Collection, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.collections[schema.Name]; exists {
		return nil, fmt.Errorf("collection %s already exists", schema.Name)
	}

	hasPrimary := false
	hasVector := false
	for _, f := range schema.Fields {
		if f.IsPrimary {
			hasPrimary = true
		}
		if f.Type == FieldFloatVector {
			hasVector = true
		}
	}
	if !hasPrimary {
		return nil, fmt.Errorf("collection must have a primary key field")
	}
	if !hasVector {
		return nil, fmt.Errorf("collection must have at least one vector field")
	}

	col := &Collection{
		Schema:    schema,
		CreatedAt: time.Now(),
	}
	cm.collections[schema.Name] = col
	return col, nil
}

func (cm *CollectionManager) DropCollection(name string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.collections[name]; !exists {
		return fmt.Errorf("collection %s does not exist", name)
	}
	delete(cm.collections, name)
	return nil
}

func (cm *CollectionManager) GetCollection(name string) (*Collection, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	col, ok := cm.collections[name]
	return col, ok
}

func (cm *CollectionManager) ListCollections() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var names []string
	for name := range cm.collections {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func demoCollectionManager() {
	cm := NewCollectionManager()

	col, err := cm.CreateCollection(CollectionSchema{
		Name:        "documents",
		Description: "Document embeddings for RAG",
		Fields: []FieldSchema{
			{Name: "id", Type: FieldInt64, IsPrimary: true, AutoID: true},
			{Name: "title", Type: FieldVarChar, MaxLength: 256},
			{Name: "embedding", Type: FieldFloatVector, Dimension: 768},
		},
	})
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Created collection: %s\n", col.Schema.Name)
	for _, f := range col.Schema.Fields {
		extra := ""
		if f.IsPrimary {
			extra = " [PRIMARY]"
		}
		if f.Dimension > 0 {
			extra = fmt.Sprintf(" [dim=%d]", f.Dimension)
		}
		os.Stdout.WriteString(fmt.Sprintf("    Field: %s (%s%s)\n", f.Name, f.Type, extra))
	}

	cm.CreateCollection(CollectionSchema{
		Name: "images",
		Fields: []FieldSchema{
			{Name: "id", Type: FieldInt64, IsPrimary: true},
			{Name: "url", Type: FieldVarChar, MaxLength: 512},
			{Name: "feature", Type: FieldFloatVector, Dimension: 512},
		},
	})

	fmt.Printf("  Collections: %v\n", cm.ListCollections())

	_, err = cm.CreateCollection(CollectionSchema{
		Name:   "bad_collection",
		Fields: []FieldSchema{{Name: "id", Type: FieldInt64, IsPrimary: true}},
	})
	fmt.Printf("  No vector field error: %v\n", err)

	cm.DropCollection("images")
	fmt.Printf("  After drop: %v\n", cm.ListCollections())
}

// ============================================
// HNSW-LIKE INDEX
// ============================================

type HNSWNode struct {
	ID        int64
	Vector    Vector
	Neighbors [][]int64
	Metadata  map[string]string
}

type HNSWIndex struct {
	mu         sync.RWMutex
	nodes      map[int64]*HNSWNode
	maxLevel   int
	efConst    int
	M          int
	entryPoint int64
	dimension  int
	rng        *rand.Rand
}

func NewHNSWIndex(dim, M, efConstruction int) *HNSWIndex {
	return &HNSWIndex{
		nodes:     make(map[int64]*HNSWNode),
		M:         M,
		efConst:   efConstruction,
		dimension: dim,
		maxLevel:  0,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (h *HNSWIndex) randomLevel() int {
	level := 0
	for h.rng.Float64() < 0.5 && level < 4 {
		level++
	}
	return level
}

func (h *HNSWIndex) Insert(id int64, vector Vector, metadata map[string]string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	level := h.randomLevel()

	node := &HNSWNode{
		ID:        id,
		Vector:    vector,
		Neighbors: make([][]int64, level+1),
		Metadata:  metadata,
	}

	for i := range node.Neighbors {
		node.Neighbors[i] = make([]int64, 0)
	}

	h.nodes[id] = node

	if len(h.nodes) == 1 {
		h.entryPoint = id
		h.maxLevel = level
		return
	}

	for l := min64(level, h.maxLevel); l >= 0; l-- {
		neighbors := h.findNeighborsAtLevel(vector, l, h.M)
		for _, nID := range neighbors {
			if l < len(node.Neighbors) {
				node.Neighbors[l] = append(node.Neighbors[l], nID)
			}
			neighbor := h.nodes[nID]
			if neighbor != nil && l < len(neighbor.Neighbors) {
				neighbor.Neighbors[l] = append(neighbor.Neighbors[l], id)
				if len(neighbor.Neighbors[l]) > h.M*2 {
					neighbor.Neighbors[l] = neighbor.Neighbors[l][:h.M*2]
				}
			}
		}
	}

	if level > h.maxLevel {
		h.maxLevel = level
		h.entryPoint = id
	}
}

func (h *HNSWIndex) findNeighborsAtLevel(query Vector, level, count int) []int64 {
	type candidate struct {
		id   int64
		dist float64
	}

	var candidates []candidate
	for id, node := range h.nodes {
		if level < len(node.Neighbors) || level == 0 {
			dist := EuclideanDistance(query, node.Vector)
			candidates = append(candidates, candidate{id, dist})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist < candidates[j].dist
	})

	result := make([]int64, 0, count)
	for i := 0; i < len(candidates) && i < count; i++ {
		result = append(result, candidates[i].id)
	}
	return result
}

type SearchResult struct {
	ID       int64
	Distance float64
	Score    float64
	Metadata map[string]string
}

func (h *HNSWIndex) Search(query Vector, topK int) []SearchResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	type candidate struct {
		id   int64
		dist float64
	}

	var candidates []candidate
	for id, node := range h.nodes {
		dist := EuclideanDistance(query, node.Vector)
		candidates = append(candidates, candidate{id, dist})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist < candidates[j].dist
	})

	var results []SearchResult
	for i := 0; i < len(candidates) && i < topK; i++ {
		c := candidates[i]
		node := h.nodes[c.id]
		score := 1.0 / (1.0 + c.dist)
		results = append(results, SearchResult{
			ID:       c.id,
			Distance: c.dist,
			Score:    score,
			Metadata: node.Metadata,
		})
	}

	return results
}

func (h *HNSWIndex) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.nodes)
}

func min64(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func demoHNSWIndex() {
	index := NewHNSWIndex(8, 16, 200)

	rng := rand.New(rand.NewSource(42))
	documents := []struct {
		id    int64
		title string
	}{
		{1, "Introduction to Go"},
		{2, "Go Concurrency Patterns"},
		{3, "Rust Memory Safety"},
		{4, "Python Machine Learning"},
		{5, "Go HTTP Servers"},
		{6, "Docker Containers"},
		{7, "Kubernetes Orchestration"},
		{8, "Go Channels and Goroutines"},
		{9, "React Frontend Development"},
		{10, "Go Testing Best Practices"},
	}

	for _, doc := range documents {
		vec := NewRandomVector(8, rng)
		index.Insert(doc.id, vec, map[string]string{"title": doc.title})
	}

	os.Stdout.WriteString(fmt.Sprintf("  Index size: %d vectors\n", index.Size()))
	os.Stdout.WriteString(fmt.Sprintf("  Max level: %d\n", index.maxLevel))

	queryVec := NewRandomVector(8, rng)
	results := index.Search(queryVec, 5)

	fmt.Println("  Top-5 search results:")
	for i, r := range results {
		os.Stdout.WriteString(fmt.Sprintf("    %d. [ID=%d] %s (distance=%.4f, score=%.4f)\n",
			i+1, r.ID, r.Metadata["title"], r.Distance, r.Score))
	}

	fmt.Println(`
  HNSW (Hierarchical Navigable Small World):
    - Multi-layer graph structure
    - Top layers: long-range connections (express highways)
    - Bottom layers: short-range connections (local streets)
    - Search: start at top, greedily descend
    - Insert: random level, connect to neighbors
    - O(log N) search complexity
    - Most popular ANN index for vector databases`)
}

// ============================================
// SEGMENT MANAGEMENT
// ============================================

type SegmentState int

const (
	SegmentGrowing SegmentState = iota
	SegmentSealed
	SegmentFlushed
)

func (s SegmentState) String() string {
	switch s {
	case SegmentGrowing:
		return "Growing"
	case SegmentSealed:
		return "Sealed"
	case SegmentFlushed:
		return "Flushed"
	default:
		return "Unknown"
	}
}

type Segment struct {
	mu       sync.RWMutex
	ID       int64
	State    SegmentState
	Vectors  map[int64]Vector
	Metadata map[int64]map[string]string
	MaxRows  int
	Index    *HNSWIndex
}

func NewSegment(id int64, dim, maxRows int) *Segment {
	return &Segment{
		ID:       id,
		State:    SegmentGrowing,
		Vectors:  make(map[int64]Vector),
		Metadata: make(map[int64]map[string]string),
		MaxRows:  maxRows,
		Index:    NewHNSWIndex(dim, 16, 200),
	}
}

func (s *Segment) Insert(id int64, vec Vector, meta map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != SegmentGrowing {
		return fmt.Errorf("segment %d is not in growing state", s.ID)
	}

	s.Vectors[id] = vec
	s.Metadata[id] = meta
	s.Index.Insert(id, vec, meta)

	if len(s.Vectors) >= s.MaxRows {
		s.State = SegmentSealed
	}

	return nil
}

func (s *Segment) Search(query Vector, topK int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Index.Search(query, topK)
}

func (s *Segment) RowCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Vectors)
}

type SegmentManager struct {
	mu         sync.RWMutex
	segments   map[int64]*Segment
	nextSegID  int64
	dimension  int
	maxPerSeg  int
}

func NewSegmentManager(dim, maxPerSeg int) *SegmentManager {
	return &SegmentManager{
		segments:  make(map[int64]*Segment),
		dimension: dim,
		maxPerSeg: maxPerSeg,
	}
}

func (sm *SegmentManager) GetGrowingSegment() *Segment {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, seg := range sm.segments {
		if seg.State == SegmentGrowing {
			return seg
		}
	}

	sm.nextSegID++
	seg := NewSegment(sm.nextSegID, sm.dimension, sm.maxPerSeg)
	sm.segments[sm.nextSegID] = seg
	return seg
}

func (sm *SegmentManager) GetAllSegments() []*Segment {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var result []*Segment
	for _, seg := range sm.segments {
		result = append(result, seg)
	}
	return result
}

func (sm *SegmentManager) SearchAll(query Vector, topK int) []SearchResult {
	segments := sm.GetAllSegments()

	var allResults []SearchResult
	for _, seg := range segments {
		results := seg.Search(query, topK)
		allResults = append(allResults, results...)
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Distance < allResults[j].Distance
	})

	if len(allResults) > topK {
		allResults = allResults[:topK]
	}
	return allResults
}

func demoSegmentManager() {
	sm := NewSegmentManager(8, 5)

	rng := rand.New(rand.NewSource(99))
	for i := int64(1); i <= 12; i++ {
		seg := sm.GetGrowingSegment()
		vec := NewRandomVector(8, rng)
		meta := map[string]string{
			"doc": fmt.Sprintf("document-%d", i),
		}
		err := seg.Insert(i, vec, meta)
		if err != nil {
			fmt.Printf("  Insert error: %v\n", err)
		}
	}

	segments := sm.GetAllSegments()
	fmt.Printf("  Total segments: %d\n", len(segments))
	for _, seg := range segments {
		os.Stdout.WriteString(fmt.Sprintf("    Segment %d: state=%s, rows=%d\n",
			seg.ID, seg.State, seg.RowCount()))
	}

	query := NewRandomVector(8, rng)
	results := sm.SearchAll(query, 3)
	fmt.Println("  Cross-segment search (top 3):")
	for i, r := range results {
		os.Stdout.WriteString(fmt.Sprintf("    %d. [ID=%d] %s (dist=%.4f)\n",
			i+1, r.ID, r.Metadata["doc"], r.Distance))
	}
}

// ============================================
// QUERY ENGINE
// ============================================

type MetricType string

const (
	MetricCosine MetricType = "COSINE"
	MetricL2     MetricType = "L2"
	MetricIP     MetricType = "IP"
)

type SearchParams struct {
	CollectionName string
	VectorField    string
	QueryVectors   []Vector
	TopK           int
	MetricType     MetricType
	Filter         *ScalarFilter
	OutputFields   []string
}

type ScalarFilter struct {
	Field    string
	Operator string
	Value    string
}

type QueryResult struct {
	QueryIdx int
	Results  []SearchResult
}

type QueryEngine struct {
	mu          sync.RWMutex
	collections map[string]*QueryCollection
}

type QueryCollection struct {
	Schema   CollectionSchema
	Segments *SegmentManager
}

func NewQueryEngine() *QueryEngine {
	return &QueryEngine{
		collections: make(map[string]*QueryCollection),
	}
}

func (qe *QueryEngine) RegisterCollection(schema CollectionSchema, segMgr *SegmentManager) {
	qe.mu.Lock()
	defer qe.mu.Unlock()

	qe.collections[schema.Name] = &QueryCollection{
		Schema:   schema,
		Segments: segMgr,
	}
}

func (qe *QueryEngine) Search(params SearchParams) ([]QueryResult, error) {
	qe.mu.RLock()
	col, ok := qe.collections[params.CollectionName]
	qe.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("collection %s not found", params.CollectionName)
	}

	var queryResults []QueryResult

	for idx, queryVec := range params.QueryVectors {
		segments := col.Segments.GetAllSegments()

		var allResults []SearchResult
		for _, seg := range segments {
			segResults := seg.Search(queryVec, params.TopK*2)

			for _, r := range segResults {
				if params.Filter != nil {
					val, ok := r.Metadata[params.Filter.Field]
					if !ok {
						continue
					}
					switch params.Filter.Operator {
					case "==":
						if val != params.Filter.Value {
							continue
						}
					case "!=":
						if val == params.Filter.Value {
							continue
						}
					}
				}

				if params.MetricType == MetricCosine {
					vec := seg.Vectors[r.ID]
					r.Score = CosineSimilarity(queryVec, vec)
					r.Distance = 1.0 - r.Score
				}

				allResults = append(allResults, r)
			}
		}

		sort.Slice(allResults, func(i, j int) bool {
			return allResults[i].Distance < allResults[j].Distance
		})

		if len(allResults) > params.TopK {
			allResults = allResults[:params.TopK]
		}

		queryResults = append(queryResults, QueryResult{
			QueryIdx: idx,
			Results:  allResults,
		})
	}

	return queryResults, nil
}

func demoQueryEngine() {
	segMgr := NewSegmentManager(4, 100)
	engine := NewQueryEngine()

	schema := CollectionSchema{
		Name: "articles",
		Fields: []FieldSchema{
			{Name: "id", Type: FieldInt64, IsPrimary: true},
			{Name: "category", Type: FieldVarChar, MaxLength: 64},
			{Name: "embedding", Type: FieldFloatVector, Dimension: 4},
		},
	}

	engine.RegisterCollection(schema, segMgr)

	articles := []struct {
		id       int64
		category string
		vec      Vector
	}{
		{1, "tech", Vector{0.9, 0.1, 0.2, 0.8}},
		{2, "tech", Vector{0.8, 0.2, 0.1, 0.9}},
		{3, "science", Vector{0.3, 0.8, 0.7, 0.2}},
		{4, "science", Vector{0.2, 0.9, 0.8, 0.1}},
		{5, "tech", Vector{0.7, 0.3, 0.2, 0.7}},
		{6, "sports", Vector{0.1, 0.2, 0.9, 0.1}},
		{7, "sports", Vector{0.2, 0.1, 0.8, 0.2}},
		{8, "tech", Vector{0.85, 0.15, 0.15, 0.85}},
	}

	seg := segMgr.GetGrowingSegment()
	for _, a := range articles {
		seg.Insert(a.id, a.vec, map[string]string{"category": a.category})
	}

	fmt.Println("  Search without filter:")
	results, _ := engine.Search(SearchParams{
		CollectionName: "articles",
		QueryVectors:   []Vector{{0.85, 0.15, 0.2, 0.8}},
		TopK:           3,
		MetricType:     MetricL2,
	})
	for _, qr := range results {
		for i, r := range qr.Results {
			os.Stdout.WriteString(fmt.Sprintf("    %d. [ID=%d] category=%s dist=%.4f\n",
				i+1, r.ID, r.Metadata["category"], r.Distance))
		}
	}

	fmt.Println("  Search with filter (category=tech):")
	filtered, _ := engine.Search(SearchParams{
		CollectionName: "articles",
		QueryVectors:   []Vector{{0.85, 0.15, 0.2, 0.8}},
		TopK:           3,
		MetricType:     MetricL2,
		Filter:         &ScalarFilter{Field: "category", Operator: "==", Value: "tech"},
	})
	for _, qr := range filtered {
		for i, r := range qr.Results {
			os.Stdout.WriteString(fmt.Sprintf("    %d. [ID=%d] category=%s dist=%.4f\n",
				i+1, r.ID, r.Metadata["category"], r.Distance))
		}
	}

	fmt.Println("  Batch search (2 queries):")
	batch, _ := engine.Search(SearchParams{
		CollectionName: "articles",
		QueryVectors: []Vector{
			{0.9, 0.1, 0.1, 0.9},
			{0.1, 0.8, 0.8, 0.1},
		},
		TopK:       2,
		MetricType: MetricL2,
	})
	for _, qr := range batch {
		os.Stdout.WriteString(fmt.Sprintf("    Query %d:\n", qr.QueryIdx+1))
		for i, r := range qr.Results {
			os.Stdout.WriteString(fmt.Sprintf("      %d. [ID=%d] category=%s dist=%.4f\n",
				i+1, r.ID, r.Metadata["category"], r.Distance))
		}
	}
}

// ============================================
// FULL MINI-MILVUS INTEGRATION
// ============================================

type MiniMilvus struct {
	mu          sync.RWMutex
	colManager  *CollectionManager
	segManagers map[string]*SegmentManager
	engine      *QueryEngine
	nextID      int64
}

func NewMiniMilvus() *MiniMilvus {
	return &MiniMilvus{
		colManager:  NewCollectionManager(),
		segManagers: make(map[string]*SegmentManager),
		engine:      NewQueryEngine(),
	}
}

func (m *MiniMilvus) CreateCollection(schema CollectionSchema) error {
	_, err := m.colManager.CreateCollection(schema)
	if err != nil {
		return err
	}

	dim := 0
	for _, f := range schema.Fields {
		if f.Type == FieldFloatVector {
			dim = f.Dimension
			break
		}
	}

	sm := NewSegmentManager(dim, 1000)
	m.mu.Lock()
	m.segManagers[schema.Name] = sm
	m.mu.Unlock()

	m.engine.RegisterCollection(schema, sm)
	return nil
}

func (m *MiniMilvus) Insert(collection string, vectors []Vector, metadata []map[string]string) ([]int64, error) {
	m.mu.Lock()
	sm, ok := m.segManagers[collection]
	if !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("collection %s not found", collection)
	}
	m.mu.Unlock()

	var ids []int64
	for i, vec := range vectors {
		m.mu.Lock()
		m.nextID++
		id := m.nextID
		m.mu.Unlock()

		seg := sm.GetGrowingSegment()
		meta := map[string]string{}
		if i < len(metadata) {
			meta = metadata[i]
		}
		err := seg.Insert(id, vec, meta)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *MiniMilvus) Search(collection string, queryVecs []Vector, topK int, metric MetricType, filter *ScalarFilter) ([]QueryResult, error) {
	return m.engine.Search(SearchParams{
		CollectionName: collection,
		QueryVectors:   queryVecs,
		TopK:           topK,
		MetricType:     metric,
		Filter:         filter,
	})
}

func (m *MiniMilvus) ListCollections() []string {
	return m.colManager.ListCollections()
}

func (m *MiniMilvus) CollectionInfo(name string) string {
	col, ok := m.colManager.GetCollection(name)
	if !ok {
		return "not found"
	}

	m.mu.RLock()
	sm := m.segManagers[name]
	m.mu.RUnlock()

	totalRows := 0
	segs := sm.GetAllSegments()
	for _, seg := range segs {
		totalRows += seg.RowCount()
	}

	var fields []string
	for _, f := range col.Schema.Fields {
		fields = append(fields, fmt.Sprintf("%s(%s)", f.Name, f.Type))
	}

	return fmt.Sprintf("rows=%d, segments=%d, fields=[%s]",
		totalRows, len(segs), strings.Join(fields, ", "))
}

func demoMiniMilvus() {
	milvus := NewMiniMilvus()

	err := milvus.CreateCollection(CollectionSchema{
		Name:        "knowledge_base",
		Description: "RAG knowledge base",
		Fields: []FieldSchema{
			{Name: "id", Type: FieldInt64, IsPrimary: true, AutoID: true},
			{Name: "source", Type: FieldVarChar, MaxLength: 128},
			{Name: "embedding", Type: FieldFloatVector, Dimension: 8},
		},
	})
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	rng := rand.New(rand.NewSource(2024))
	sources := []string{"go-docs", "go-blog", "go-spec", "go-wiki", "go-talks"}

	var vectors []Vector
	var metas []map[string]string
	for i := 0; i < 20; i++ {
		vectors = append(vectors, NewRandomVector(8, rng))
		metas = append(metas, map[string]string{
			"source": sources[i%len(sources)],
			"chunk":  fmt.Sprintf("chunk-%d", i),
		})
	}

	ids, err := milvus.Insert("knowledge_base", vectors, metas)
	if err != nil {
		fmt.Printf("  Insert error: %v\n", err)
		return
	}
	os.Stdout.WriteString(fmt.Sprintf("  Inserted %d vectors (IDs: %d to %d)\n", len(ids), ids[0], ids[len(ids)-1]))

	fmt.Printf("  Collections: %v\n", milvus.ListCollections())
	fmt.Printf("  Info: %s\n", milvus.CollectionInfo("knowledge_base"))

	fmt.Println("  RAG similarity search:")
	queryVec := NewRandomVector(8, rng)
	results, err := milvus.Search("knowledge_base", []Vector{queryVec}, 5, MetricL2, nil)
	if err != nil {
		fmt.Printf("  Search error: %v\n", err)
		return
	}

	for _, qr := range results {
		for i, r := range qr.Results {
			os.Stdout.WriteString(fmt.Sprintf("    %d. [ID=%d] source=%s, chunk=%s (dist=%.4f)\n",
				i+1, r.ID, r.Metadata["source"], r.Metadata["chunk"], r.Distance))
		}
	}

	fmt.Println("  Filtered search (source=go-docs):")
	filtered, _ := milvus.Search("knowledge_base", []Vector{queryVec}, 3, MetricL2,
		&ScalarFilter{Field: "source", Operator: "==", Value: "go-docs"})

	for _, qr := range filtered {
		for i, r := range qr.Results {
			os.Stdout.WriteString(fmt.Sprintf("    %d. [ID=%d] source=%s, chunk=%s (dist=%.4f)\n",
				i+1, r.ID, r.Metadata["source"], r.Metadata["chunk"], r.Distance))
		}
	}

	fmt.Println(`
  MILVUS REAL FEATURES:
    - Billions of vectors at millisecond latency
    - GPU-accelerated index building (NVIDIA RAFT)
    - Multiple index types: HNSW, IVF_FLAT, IVF_PQ, DISKANN
    - Hybrid search: vector + scalar filtering
    - Multi-vector search and reranking
    - Time travel (query historical data)
    - CDC (Change Data Capture)
    - Role-based access control (RBAC)
    - Cloud-native: S3 + Pulsar/Kafka + etcd

  COMMON RAG PIPELINE:
    1. Document -> Chunks (text splitter)
    2. Chunks -> Embeddings (OpenAI/BERT/etc.)
    3. Embeddings -> Milvus (insert)
    4. User Query -> Query Embedding
    5. Query Embedding -> Milvus Search (top-K)
    6. Top-K Chunks -> LLM Context (prompt)
    7. LLM -> Answer (grounded in retrieved docs)`)
}

// Ensure imports are used
var _ = math.Sqrt
var _ = sort.Slice
var _ = time.Now

/* SUMMARY - CHAPTER 130: MILVUS INTERNALS

Key Concepts:

1. VECTOR SIMILARITY
   - Cosine similarity (angle between vectors)
   - Euclidean distance (L2 norm)
   - Inner product (dot product)
   - Vector normalization

2. COLLECTION MANAGEMENT
   - Schema validation (primary key + vector field required)
   - Field types: Int64, Float64, VarChar, FloatVector
   - CRUD operations on collections

3. HNSW INDEX
   - Hierarchical Navigable Small World graph
   - Multi-level neighbor connections
   - O(log N) approximate nearest neighbor search
   - Configurable M (max connections) and efConstruction

4. SEGMENT MANAGEMENT
   - Growing -> Sealed -> Flushed lifecycle
   - Auto-seal when max rows reached
   - Cross-segment search with result merging

5. QUERY ENGINE
   - Multi-query batch search
   - Scalar filtering (hybrid search)
   - Multiple metric types (L2, Cosine, IP)
   - Top-K result aggregation

Go Patterns Demonstrated:
- Interface-based metric functions
- Segment lifecycle state machine
- Cross-segment search aggregation
- Mutex-protected concurrent operations
- Builder pattern for search params
- Sort-based top-K selection

Real Milvus Implementation:
- knowhere engine (C++) for SIMD-optimized search
- GPU index building with NVIDIA RAFT
- Distributed query across nodes
- Log-structured merge for segments
- etcd for metadata coordination
- Pulsar/Kafka for write-ahead log
- S3/MinIO for persistent storage
*/
