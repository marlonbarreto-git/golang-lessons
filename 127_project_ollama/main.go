// Package main - Chapter 127: Ollama Internals
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 127: OLLAMA - LOCAL LLM SERVING IN GO ===")

	// ============================================
	// WHAT IS OLLAMA?
	// ============================================
	fmt.Println(`
==============================
WHAT IS OLLAMA?
==============================

Ollama is the #2 most popular Go project on GitHub with 161k+ stars.
It makes running large language models locally as easy as Docker
makes running containers.

WHAT IT DOES:
  - Downloads and manages LLM models (llama, mistral, phi, etc.)
  - Serves models via a REST API (OpenAI-compatible)
  - Handles GPU memory allocation and model loading
  - Supports custom models via Modelfile (like Dockerfile)
  - Streaming token generation with Server-Sent Events

ARCHITECTURE:
  +------------------+     +------------------+
  |   CLI Client     |---->|   REST API       |
  +------------------+     |   (net/http)     |
                           +--------+---------+
                                    |
                           +--------v---------+
                           | Model Manager    |
                           | (download, cache)|
                           +--------+---------+
                                    |
                           +--------v---------+
                           | GGML Runtime     |
                           | (llama.cpp CGo)  |
                           +--------+---------+
                                    |
                           +--------v---------+
                           | GPU Scheduler    |
                           | (CUDA/Metal/ROCm)|
                           +------------------+

WHY GO:
  - Single static binary for easy distribution
  - net/http for the API server with streaming
  - Goroutines for concurrent model loading/inference
  - CGo bindings to llama.cpp for actual inference
  - Cross-platform: macOS (Metal), Linux (CUDA), Windows

COMPARISON WITH ALTERNATIVES:
  vLLM (Python):    Production serving, PagedAttention, higher throughput
  TGI (Rust):       HuggingFace's server, optimized for cloud
  Ollama (Go):      Local-first, simplicity, Docker-like UX
  llama.cpp (C++):  Raw engine Ollama wraps, maximum performance`)

	// ============================================
	// MODELFILE PARSER
	// ============================================
	fmt.Println("\n--- Modelfile Parser ---")
	demoModelfileParser()

	// ============================================
	// MODEL REGISTRY
	// ============================================
	fmt.Println("\n--- Model Registry ---")
	demoModelRegistry()

	// ============================================
	// STREAMING INFERENCE SERVER
	// ============================================
	fmt.Println("\n--- Streaming Inference Server ---")
	demoStreamingServer()

	// ============================================
	// GPU MEMORY MANAGER
	// ============================================
	fmt.Println("\n--- GPU Memory Manager ---")
	demoGPUMemoryManager()

	// ============================================
	// CONCURRENT MODEL SCHEDULER
	// ============================================
	fmt.Println("\n--- Concurrent Model Scheduler ---")
	demoConcurrentScheduler()

	// ============================================
	// FULL MINI-OLLAMA
	// ============================================
	fmt.Println("\n--- Mini-Ollama Integration ---")
	demoMiniOllama()
}

// ============================================
// MODELFILE PARSER
// ============================================

type Modelfile struct {
	From       string
	Parameters map[string]string
	Template   string
	System     string
	Adapter    string
	License    string
	Messages   []Message
}

type Message struct {
	Role    string
	Content string
}

type ModelfileParser struct {
	lines   []string
	current int
}

func NewModelfileParser(content string) *ModelfileParser {
	return &ModelfileParser{
		lines: strings.Split(content, "\n"),
	}
}

func (p *ModelfileParser) Parse() (*Modelfile, error) {
	mf := &Modelfile{
		Parameters: make(map[string]string),
	}

	for p.current < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.current])
		p.current++

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		instruction := strings.ToUpper(parts[0])
		value := strings.TrimSpace(parts[1])

		switch instruction {
		case "FROM":
			mf.From = value
		case "PARAMETER":
			paramParts := strings.SplitN(value, " ", 2)
			if len(paramParts) == 2 {
				mf.Parameters[paramParts[0]] = paramParts[1]
			}
		case "TEMPLATE":
			mf.Template = p.parseMultiline(value)
		case "SYSTEM":
			mf.System = p.parseMultiline(value)
		case "ADAPTER":
			mf.Adapter = value
		case "LICENSE":
			mf.License = p.parseMultiline(value)
		case "MESSAGE":
			msgParts := strings.SplitN(value, " ", 2)
			if len(msgParts) == 2 {
				mf.Messages = append(mf.Messages, Message{
					Role:    msgParts[0],
					Content: msgParts[1],
				})
			}
		}
	}

	if mf.From == "" {
		return nil, fmt.Errorf("Modelfile must contain a FROM instruction")
	}

	return mf, nil
}

func (p *ModelfileParser) parseMultiline(firstLine string) string {
	if !strings.HasPrefix(firstLine, `"""`) {
		return firstLine
	}

	var lines []string
	content := strings.TrimPrefix(firstLine, `"""`)
	if strings.HasSuffix(content, `"""`) {
		return strings.TrimSuffix(content, `"""`)
	}
	if content != "" {
		lines = append(lines, content)
	}

	for p.current < len(p.lines) {
		line := p.lines[p.current]
		p.current++
		if strings.Contains(line, `"""`) {
			before := strings.TrimSuffix(strings.TrimSpace(line), `"""`)
			if before != "" {
				lines = append(lines, before)
			}
			break
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func demoModelfileParser() {
	modelfileContent := `# Custom coding assistant model
FROM llama3.2

# Inference parameters
PARAMETER temperature 0.7
PARAMETER top_p 0.9
PARAMETER top_k 40
PARAMETER num_ctx 4096
PARAMETER repeat_penalty 1.1
PARAMETER stop <|eot_id|>

TEMPLATE """{{ if .System }}<|start_header_id|>system<|end_header_id|>
{{ .System }}<|eot_id|>{{ end }}{{ if .Prompt }}<|start_header_id|>user<|end_header_id|>
{{ .Prompt }}<|eot_id|>{{ end }}<|start_header_id|>assistant<|end_header_id|>
{{ .Response }}<|eot_id|>"""

SYSTEM """You are a helpful coding assistant specializing in Go.
Always provide working code examples."""

MESSAGE user How do I create a goroutine?
MESSAGE assistant Use the go keyword: go myFunction()

LICENSE """MIT License - Free to use and modify."""`

	parser := NewModelfileParser(modelfileContent)
	mf, err := parser.Parse()
	if err != nil {
		fmt.Printf("  Parse error: %v\n", err)
		return
	}

	fmt.Printf("  FROM: %s\n", mf.From)
	fmt.Printf("  Parameters:\n")
	for k, v := range mf.Parameters {
		fmt.Printf("    %s = %s\n", k, v)
	}
	fmt.Printf("  System: %s\n", mf.System[:50]+"...")
	fmt.Printf("  Template length: %d chars\n", len(mf.Template))
	fmt.Printf("  Messages: %d\n", len(mf.Messages))
	for _, msg := range mf.Messages {
		fmt.Printf("    [%s]: %s\n", msg.Role, msg.Content)
	}
}

// ============================================
// MODEL REGISTRY
// ============================================

type ModelInfo struct {
	Name         string
	Tag          string
	Size         int64
	Digest       string
	Family       string
	ParameterSize string
	QuantLevel   string
	Format       string
	ModifiedAt   time.Time
}

type LayerInfo struct {
	Digest    string
	Size      int64
	MediaType string
	Status    string
}

type ModelRegistry struct {
	mu     sync.RWMutex
	models map[string]*ModelInfo
	layers map[string]*LayerInfo
	blobs  map[string][]byte
}

func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]*ModelInfo),
		layers: make(map[string]*LayerInfo),
		blobs:  make(map[string][]byte),
	}
}

func (r *ModelRegistry) Pull(name string) (*ModelInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	parts := strings.SplitN(name, ":", 2)
	modelName := parts[0]
	tag := "latest"
	if len(parts) == 2 {
		tag = parts[1]
	}

	key := modelName + ":" + tag

	catalog := map[string]ModelInfo{
		"llama3.2:latest": {
			Name: "llama3.2", Tag: "latest", Size: 2_000_000_000,
			Family: "llama", ParameterSize: "3B", QuantLevel: "Q4_0",
			Format: "gguf",
		},
		"mistral:latest": {
			Name: "mistral", Tag: "latest", Size: 4_100_000_000,
			Family: "mistral", ParameterSize: "7B", QuantLevel: "Q4_0",
			Format: "gguf",
		},
		"phi3:latest": {
			Name: "phi3", Tag: "latest", Size: 2_300_000_000,
			Family: "phi3", ParameterSize: "3.8B", QuantLevel: "Q4_K_M",
			Format: "gguf",
		},
		"codellama:latest": {
			Name: "codellama", Tag: "latest", Size: 3_800_000_000,
			Family: "llama", ParameterSize: "7B", QuantLevel: "Q4_0",
			Format: "gguf",
		},
	}

	info, exists := catalog[key]
	if !exists {
		return nil, fmt.Errorf("model %s not found in registry", key)
	}

	info.ModifiedAt = time.Now()
	info.Digest = fmt.Sprintf("sha256:%x", rand.Int63())
	r.models[key] = &info

	layers := []LayerInfo{
		{Digest: fmt.Sprintf("sha256:%x", rand.Int63()), Size: info.Size, MediaType: "application/vnd.ollama.image.model"},
		{Digest: fmt.Sprintf("sha256:%x", rand.Int63()), Size: 512, MediaType: "application/vnd.ollama.image.template"},
		{Digest: fmt.Sprintf("sha256:%x", rand.Int63()), Size: 256, MediaType: "application/vnd.ollama.image.params"},
	}

	for i := range layers {
		layers[i].Status = "success"
		r.layers[layers[i].Digest] = &layers[i]
	}

	return &info, nil
}

func (r *ModelRegistry) List() []ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []ModelInfo
	for _, m := range r.models {
		result = append(result, *m)
	}
	return result
}

func (r *ModelRegistry) Get(name string) (*ModelInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !strings.Contains(name, ":") {
		name += ":latest"
	}

	m, ok := r.models[name]
	if !ok {
		return nil, false
	}
	return m, true
}

func (r *ModelRegistry) Delete(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !strings.Contains(name, ":") {
		name += ":latest"
	}

	if _, ok := r.models[name]; ok {
		delete(r.models, name)
		return true
	}
	return false
}

func demoModelRegistry() {
	registry := NewModelRegistry()

	models := []string{"llama3.2", "mistral", "phi3", "codellama"}
	for _, name := range models {
		info, err := registry.Pull(name)
		if err != nil {
			fmt.Printf("  Error pulling %s: %v\n", name, err)
			continue
		}
		fmt.Printf("  Pulled %s:%s (%s, %s, %.1f GB)\n",
			info.Name, info.Tag, info.Family, info.ParameterSize,
			float64(info.Size)/1e9)
	}

	fmt.Printf("  Registry has %d models\n", len(registry.List()))

	if info, ok := registry.Get("llama3.2"); ok {
		fmt.Printf("  Found: %s (format=%s, quant=%s)\n", info.Name, info.Format, info.QuantLevel)
	}

	registry.Delete("phi3")
	fmt.Printf("  After delete: %d models\n", len(registry.List()))

	_, err := registry.Pull("nonexistent")
	fmt.Printf("  Pull nonexistent: %v\n", err)
}

// ============================================
// STREAMING INFERENCE SERVER
// ============================================

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type GenerateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	TotalDuration    int64 `json:"total_duration,omitempty"`
	LoadDuration     int64 `json:"load_duration,omitempty"`
	PromptEvalCount  int   `json:"prompt_eval_count,omitempty"`
	EvalCount        int   `json:"eval_count,omitempty"`
	EvalDuration     int64 `json:"eval_duration,omitempty"`
}

type ChatRequest struct {
	Model    string          `json:"model"`
	Messages []ChatMessage   `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type InferenceEngine struct {
	mu         sync.Mutex
	loadedModel string
	vocabulary []string
}

func NewInferenceEngine() *InferenceEngine {
	return &InferenceEngine{
		vocabulary: []string{
			"the", "a", "is", "in", "Go", "function", "returns",
			"error", "nil", "struct", "interface", "channel",
			"goroutine", "package", "import", "func", "var",
			"type", "map", "slice", "string", "int", "bool",
			"for", "range", "if", "else", "switch", "case",
			"defer", "select", "make", "new", "append", "len",
			"concurrent", "safe", "pattern", "using", "with",
			"can", "be", "used", "to", "create", "handle",
			"multiple", "requests", "simultaneously", ".",
		},
	}
}

func (e *InferenceEngine) LoadModel(name string) time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()

	start := time.Now()
	time.Sleep(5 * time.Millisecond)
	e.loadedModel = name
	return time.Since(start)
}

func (e *InferenceEngine) GenerateTokens(prompt string, count int) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < count; i++ {
			token := e.vocabulary[r.Intn(len(e.vocabulary))]
			if i == 0 {
				token = strings.Title(token)
			}
			ch <- token
			time.Sleep(time.Millisecond)
		}
	}()

	return ch
}

type OllamaServer struct {
	engine   *InferenceEngine
	registry *ModelRegistry
}

func NewOllamaServer() *OllamaServer {
	return &OllamaServer{
		engine:   NewInferenceEngine(),
		registry: NewModelRegistry(),
	}
}

func (s *OllamaServer) HandleGenerate(req GenerateRequest) []GenerateResponse {
	start := time.Now()

	_, ok := s.registry.Get(req.Model)
	if !ok {
		s.registry.Pull(req.Model)
	}

	loadDur := s.engine.LoadModel(req.Model)

	var responses []GenerateResponse
	tokens := s.engine.GenerateTokens(req.Prompt, 15)

	var fullResponse strings.Builder
	tokenCount := 0

	for token := range tokens {
		tokenCount++
		if fullResponse.Len() > 0 {
			fullResponse.WriteString(" ")
		}
		fullResponse.WriteString(token)

		if req.Stream {
			responses = append(responses, GenerateResponse{
				Model:     req.Model,
				Response:  token + " ",
				Done:      false,
				CreatedAt: time.Now().Format(time.RFC3339),
			})
		}
	}

	finalResp := GenerateResponse{
		Model:           req.Model,
		Response:        "",
		Done:            true,
		CreatedAt:       time.Now().Format(time.RFC3339),
		TotalDuration:   time.Since(start).Nanoseconds(),
		LoadDuration:    loadDur.Nanoseconds(),
		PromptEvalCount: len(strings.Fields(req.Prompt)),
		EvalCount:       tokenCount,
		EvalDuration:    time.Since(start).Nanoseconds() - loadDur.Nanoseconds(),
	}

	if !req.Stream {
		finalResp.Response = fullResponse.String()
	}

	responses = append(responses, finalResp)
	return responses
}

func demoStreamingServer() {
	server := NewOllamaServer()

	server.registry.Pull("llama3.2")

	fmt.Println("  Streaming generate:")
	req := GenerateRequest{
		Model:  "llama3.2",
		Prompt: "Explain goroutines in Go",
		Stream: true,
	}

	responses := server.HandleGenerate(req)
	fmt.Print("    ")
	for _, resp := range responses {
		if !resp.Done {
			fmt.Print(resp.Response)
		} else {
			os.Stdout.WriteString(fmt.Sprintf(
				"\n    [done] tokens=%d, total=%dms, eval=%dms\n",
				resp.EvalCount,
				resp.TotalDuration/1e6,
				resp.EvalDuration/1e6,
			))
		}
	}

	fmt.Println("  Non-streaming generate:")
	req2 := GenerateRequest{
		Model:  "llama3.2",
		Prompt: "What is a channel?",
		Stream: false,
	}
	responses2 := server.HandleGenerate(req2)
	for _, resp := range responses2 {
		if resp.Done {
			os.Stdout.WriteString(fmt.Sprintf(
				"    Response: %s\n    [tokens=%d]\n",
				resp.Response, resp.EvalCount,
			))
		}
	}

	fmt.Println("  Chat endpoint simulation:")
	chatReq := ChatRequest{
		Model: "llama3.2",
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a Go expert."},
			{Role: "user", Content: "What is a goroutine?"},
		},
		Stream: false,
	}
	chatJSON, _ := json.MarshalIndent(chatReq, "    ", "  ")
	fmt.Printf("    Request: %s\n", chatJSON)

	genReq := GenerateRequest{
		Model:  chatReq.Model,
		Prompt: chatReq.Messages[len(chatReq.Messages)-1].Content,
		Stream: false,
	}
	chatResponses := server.HandleGenerate(genReq)
	for _, resp := range chatResponses {
		if resp.Done {
			fmt.Printf("    Assistant: %s\n", resp.Response)
		}
	}
}

// ============================================
// GPU MEMORY MANAGER
// ============================================

type GPUDevice struct {
	ID           int
	Name         string
	TotalMemory  int64
	UsedMemory   int64
	Backend      string
}

type ModelAllocation struct {
	ModelName  string
	DeviceID   int
	MemoryUsed int64
	Layers     int
	LoadedAt   time.Time
}

type GPUMemoryManager struct {
	mu          sync.RWMutex
	devices     []*GPUDevice
	allocations map[string]*ModelAllocation
	maxModels   int
}

func NewGPUMemoryManager(devices []*GPUDevice) *GPUMemoryManager {
	return &GPUMemoryManager{
		devices:     devices,
		allocations: make(map[string]*ModelAllocation),
		maxModels:   len(devices) * 2,
	}
}

func (m *GPUMemoryManager) AvailableMemory(deviceID int) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, d := range m.devices {
		if d.ID == deviceID {
			return d.TotalMemory - d.UsedMemory
		}
	}
	return 0
}

func (m *GPUMemoryManager) AllocateModel(modelName string, requiredMemory int64, layers int) (*ModelAllocation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.allocations[modelName]; exists {
		return m.allocations[modelName], nil
	}

	var bestDevice *GPUDevice
	var bestAvailable int64

	for _, d := range m.devices {
		available := d.TotalMemory - d.UsedMemory
		if available >= requiredMemory && (bestDevice == nil || available > bestAvailable) {
			bestDevice = d
			bestAvailable = available
		}
	}

	if bestDevice == nil {
		if len(m.allocations) > 0 {
			var oldest string
			var oldestTime time.Time
			for name, alloc := range m.allocations {
				if oldest == "" || alloc.LoadedAt.Before(oldestTime) {
					oldest = name
					oldestTime = alloc.LoadedAt
				}
			}
			if oldest != "" {
				m.evictModelLocked(oldest)
				return m.allocateOnDevice(modelName, requiredMemory, layers)
			}
		}
		return nil, fmt.Errorf("insufficient GPU memory for model %s (need %d bytes)", modelName, requiredMemory)
	}

	alloc := &ModelAllocation{
		ModelName:  modelName,
		DeviceID:   bestDevice.ID,
		MemoryUsed: requiredMemory,
		Layers:     layers,
		LoadedAt:   time.Now(),
	}

	bestDevice.UsedMemory += requiredMemory
	m.allocations[modelName] = alloc
	return alloc, nil
}

func (m *GPUMemoryManager) allocateOnDevice(modelName string, requiredMemory int64, layers int) (*ModelAllocation, error) {
	for _, d := range m.devices {
		available := d.TotalMemory - d.UsedMemory
		if available >= requiredMemory {
			alloc := &ModelAllocation{
				ModelName:  modelName,
				DeviceID:   d.ID,
				MemoryUsed: requiredMemory,
				Layers:     layers,
				LoadedAt:   time.Now(),
			}
			d.UsedMemory += requiredMemory
			m.allocations[modelName] = alloc
			return alloc, nil
		}
	}
	return nil, fmt.Errorf("no device with enough memory after eviction")
}

func (m *GPUMemoryManager) evictModelLocked(modelName string) {
	alloc, exists := m.allocations[modelName]
	if !exists {
		return
	}

	for _, d := range m.devices {
		if d.ID == alloc.DeviceID {
			d.UsedMemory -= alloc.MemoryUsed
			break
		}
	}
	delete(m.allocations, modelName)
}

func (m *GPUMemoryManager) EvictModel(modelName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evictModelLocked(modelName)
}

func (m *GPUMemoryManager) Status() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	for _, d := range m.devices {
		usedPct := float64(d.UsedMemory) / float64(d.TotalMemory) * 100
		sb.WriteString(fmt.Sprintf("    GPU %d (%s/%s): %.1f GB / %.1f GB (%.0f%% used)\n",
			d.ID, d.Name, d.Backend,
			float64(d.UsedMemory)/1e9,
			float64(d.TotalMemory)/1e9,
			usedPct))
	}

	sb.WriteString(fmt.Sprintf("    Loaded models: %d\n", len(m.allocations)))
	for name, alloc := range m.allocations {
		sb.WriteString(fmt.Sprintf("      %s -> GPU %d (%d layers, %.1f GB)\n",
			name, alloc.DeviceID, alloc.Layers, float64(alloc.MemoryUsed)/1e9))
	}

	return sb.String()
}

func demoGPUMemoryManager() {
	devices := []*GPUDevice{
		{ID: 0, Name: "Apple M2 Ultra", TotalMemory: 24_000_000_000, Backend: "Metal"},
		{ID: 1, Name: "NVIDIA RTX 4090", TotalMemory: 24_000_000_000, Backend: "CUDA"},
	}

	mgr := NewGPUMemoryManager(devices)

	allocations := []struct {
		name   string
		memory int64
		layers int
	}{
		{"llama3.2:3B", 2_000_000_000, 26},
		{"mistral:7B", 4_100_000_000, 32},
		{"codellama:7B", 3_800_000_000, 32},
		{"phi3:3.8B", 2_300_000_000, 32},
	}

	for _, a := range allocations {
		alloc, err := mgr.AllocateModel(a.name, a.memory, a.layers)
		if err != nil {
			fmt.Printf("  Failed to allocate %s: %v\n", a.name, err)
			continue
		}
		fmt.Printf("  Allocated %s on GPU %d (%d layers)\n", a.name, alloc.DeviceID, alloc.Layers)
	}

	fmt.Println("  GPU Status:")
	fmt.Print(mgr.Status())

	fmt.Println("  Evicting mistral:7B...")
	mgr.EvictModel("mistral:7B")

	largeAlloc, err := mgr.AllocateModel("llama3.1:70B", 40_000_000_000, 80)
	if err != nil {
		fmt.Printf("  Expected: %v\n", err)
	} else {
		fmt.Printf("  Allocated large model on GPU %d\n", largeAlloc.DeviceID)
	}

	fmt.Println("  Final status:")
	fmt.Print(mgr.Status())
}

// ============================================
// CONCURRENT MODEL SCHEDULER
// ============================================

type InferenceRequest struct {
	ID     string
	Model  string
	Prompt string
	Result chan<- InferenceResult
}

type InferenceResult struct {
	RequestID string
	Response  string
	Duration  time.Duration
	Error     error
}

type ModelScheduler struct {
	mu       sync.Mutex
	engine   *InferenceEngine
	registry *ModelRegistry
	queue    chan InferenceRequest
	active   int
	maxConc  int
	wg       sync.WaitGroup
}

func NewModelScheduler(maxConcurrent int) *ModelScheduler {
	s := &ModelScheduler{
		engine:   NewInferenceEngine(),
		registry: NewModelRegistry(),
		queue:    make(chan InferenceRequest, 100),
		maxConc:  maxConcurrent,
	}
	return s
}

func (s *ModelScheduler) Start() {
	for i := 0; i < s.maxConc; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

func (s *ModelScheduler) Stop() {
	close(s.queue)
	s.wg.Wait()
}

func (s *ModelScheduler) worker(id int) {
	defer s.wg.Done()

	for req := range s.queue {
		start := time.Now()

		s.engine.LoadModel(req.Model)

		tokens := s.engine.GenerateTokens(req.Prompt, 10)
		var response strings.Builder
		for token := range tokens {
			if response.Len() > 0 {
				response.WriteString(" ")
			}
			response.WriteString(token)
		}

		req.Result <- InferenceResult{
			RequestID: req.ID,
			Response:  response.String(),
			Duration:  time.Since(start),
		}
	}
}

func (s *ModelScheduler) Submit(req InferenceRequest) {
	s.queue <- req
}

func demoConcurrentScheduler() {
	scheduler := NewModelScheduler(3)
	scheduler.Start()

	requests := []struct {
		id     string
		model  string
		prompt string
	}{
		{"req-1", "llama3.2", "What is Go?"},
		{"req-2", "mistral", "Explain channels"},
		{"req-3", "llama3.2", "What are goroutines?"},
		{"req-4", "codellama", "Write a HTTP server"},
		{"req-5", "phi3", "Explain interfaces"},
	}

	results := make(chan InferenceResult, len(requests))

	for _, r := range requests {
		scheduler.Submit(InferenceRequest{
			ID:     r.id,
			Model:  r.model,
			Prompt: r.prompt,
			Result: results,
		})
	}

	for i := 0; i < len(requests); i++ {
		result := <-results
		fmt.Printf("  [%s] %s (took %v)\n", result.RequestID, result.Response[:min(40, len(result.Response))], result.Duration.Round(time.Millisecond))
	}

	scheduler.Stop()
	fmt.Println("  Scheduler stopped gracefully")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================
// FULL MINI-OLLAMA INTEGRATION
// ============================================

type MiniOllama struct {
	registry  *ModelRegistry
	gpuMgr    *GPUMemoryManager
	scheduler *ModelScheduler
	server    *OllamaServer
}

func NewMiniOllama() *MiniOllama {
	devices := []*GPUDevice{
		{ID: 0, Name: "Simulated GPU", TotalMemory: 16_000_000_000, Backend: "Metal"},
	}

	return &MiniOllama{
		registry:  NewModelRegistry(),
		gpuMgr:    NewGPUMemoryManager(devices),
		scheduler: NewModelScheduler(2),
		server:    NewOllamaServer(),
	}
}

func (o *MiniOllama) Start() {
	o.scheduler.Start()
}

func (o *MiniOllama) Stop() {
	o.scheduler.Stop()
}

func (o *MiniOllama) Pull(name string) error {
	info, err := o.registry.Pull(name)
	if err != nil {
		return err
	}
	_, err = o.gpuMgr.AllocateModel(info.Name+":"+info.Tag, info.Size, 32)
	return err
}

func (o *MiniOllama) Generate(model, prompt string) (string, error) {
	result := make(chan InferenceResult, 1)
	o.scheduler.Submit(InferenceRequest{
		ID:     fmt.Sprintf("gen-%d", time.Now().UnixNano()),
		Model:  model,
		Prompt: prompt,
		Result: result,
	})

	res := <-result
	return res.Response, res.Error
}

func (o *MiniOllama) ListModels() []ModelInfo {
	return o.registry.List()
}

func demoMiniOllama() {
	ollama := NewMiniOllama()
	ollama.Start()
	defer ollama.Stop()

	models := []string{"llama3.2", "mistral", "phi3"}
	for _, m := range models {
		err := ollama.Pull(m)
		if err != nil {
			fmt.Printf("  Error pulling %s: %v\n", m, err)
			continue
		}
		fmt.Printf("  Pulled %s\n", m)
	}

	fmt.Printf("  Available models: %d\n", len(ollama.ListModels()))

	fmt.Println("  GPU status:")
	fmt.Print(ollama.gpuMgr.Status())

	prompts := []struct {
		model  string
		prompt string
	}{
		{"llama3.2", "Explain error handling in Go"},
		{"mistral", "What is a context in Go?"},
	}

	for _, p := range prompts {
		resp, err := ollama.Generate(p.model, p.prompt)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		fmt.Printf("  [%s] Q: %s\n", p.model, p.prompt)
		fmt.Printf("          A: %s\n", resp)
	}

	fmt.Println(`
  OLLAMA API ENDPOINTS (real implementation):
    POST /api/generate     - Generate completion (streaming)
    POST /api/chat         - Chat completion (streaming)
    POST /api/embeddings   - Generate embeddings
    POST /api/pull         - Pull a model
    POST /api/push         - Push a model
    POST /api/create       - Create model from Modelfile
    GET  /api/tags         - List local models
    POST /api/show         - Show model info
    POST /api/copy         - Copy a model
    DELETE /api/delete     - Delete a model
    GET  /api/ps           - List running models`)
}

/* SUMMARY - CHAPTER 127: OLLAMA INTERNALS

Key Concepts:

1. MODELFILE PARSER
   - Docker-like syntax (FROM, PARAMETER, TEMPLATE, SYSTEM)
   - Multiline string support with triple quotes
   - Instruction-based model configuration

2. MODEL REGISTRY
   - Pull/list/delete model management
   - Layer-based storage (like Docker images)
   - GGUF format for quantized models

3. STREAMING INFERENCE
   - Server-Sent Events pattern for token streaming
   - Generate and Chat API endpoints
   - Response includes timing metrics

4. GPU MEMORY MANAGEMENT
   - Device discovery and allocation
   - LRU eviction when memory is full
   - Multi-GPU support with best-fit allocation

5. CONCURRENT MODEL SCHEDULING
   - Worker pool for parallel inference
   - Request queuing with bounded concurrency
   - Graceful shutdown with WaitGroup

Go Patterns Demonstrated:
- Channel-based streaming (token generation)
- Worker pool pattern (concurrent inference)
- LRU cache eviction (GPU memory manager)
- Builder/parser pattern (Modelfile)
- Interface composition (server components)
- Mutex-protected shared state (registry, GPU manager)

Real Ollama Implementation:
- CGo bindings to llama.cpp for actual inference
- Metal/CUDA/ROCm GPU acceleration
- KV-cache management for context
- Partial model offloading (CPU+GPU split)
- OpenAI-compatible API endpoints
- Model layer deduplication and caching
*/
