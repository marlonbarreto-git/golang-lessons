// Package main - Chapter 073: Plugin System & Expvar Monitoring
// Sistema de plugins con el paquete plugin (carga dinamica de .so),
// y monitoreo de aplicaciones con expvar (variables publicas, mapas,
// variables custom, integracion HTTP).
package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== PLUGIN SYSTEM & EXPVAR MONITORING ===")

	// ============================================
	// 1. PLUGIN PACKAGE - CONCEPTOS
	// ============================================
	fmt.Println("\n--- 1. plugin - Sistema de plugins ---")
	os.Stdout.WriteString(`
PLUGIN PACKAGE - CARGA DINAMICA DE CODIGO:

  plugin.Open(path string) (*Plugin, error)    // Abrir archivo .so
  p.Lookup(symName string) (Symbol, error)     // Buscar simbolo exportado
  type Symbol = interface{}                     // Cualquier valor exportado

CREAR UN PLUGIN:
  1. Escribir codigo con package main
  2. Compilar: go build -buildmode=plugin -o myplugin.so plugin.go
  3. Cargar y usar desde la aplicacion host

LIMITACIONES:
  - Solo funciona en Linux, FreeBSD y macOS
  - El plugin debe compilarse con la MISMA version de Go
  - Las dependencias deben ser compatibles
  - No se puede descargar un plugin una vez cargado
  - No funciona con CGO_ENABLED=0
  - No funciona en Windows

EJEMPLO DE PLUGIN (greeter.go):

  package main

  import "fmt"

  var Name = "GreeterPlugin"
  var Version = "1.0.0"

  type Greeter struct{}

  func (g Greeter) Greet(name string) string {
      return fmt.Sprintf("Hello, %s!", name)
  }

  // Variable exportada que la app host buscara
  var Impl Greeter

EJEMPLO DE HOST (main.go):

  p, err := plugin.Open("greeter.so")
  if err != nil { log.Fatal(err) }

  nameSym, _ := p.Lookup("Name")
  name := *nameSym.(*string)
  fmt.Println("Plugin:", name)

  implSym, _ := p.Lookup("Impl")
  // Type assert a la interfaz esperada
`)

	fmt.Println("  Plugin package disponible en: Linux, FreeBSD, macOS")
	fmt.Printf("  Sistema actual: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// ============================================
	// 2. PLUGIN - PATRON DE INTERFAZ
	// ============================================
	fmt.Println("\n--- 2. Plugin - Patron de interfaz ---")
	os.Stdout.WriteString(`
PATRON RECOMENDADO PARA PLUGINS:

  1. Definir interfaz en paquete compartido:
     type Processor interface {
         Name() string
         Process(data []byte) ([]byte, error)
     }

  2. Plugin implementa la interfaz:
     var Plugin Processor = &myProcessor{}

  3. Host busca el simbolo y hace type assertion:
     sym, _ := p.Lookup("Plugin")
     proc := sym.(Processor)
     result, _ := proc.Process(data)

ALTERNATIVA SIN PLUGIN PACKAGE:
  - Subprocesos con stdin/stdout (JSON-RPC, gRPC)
  - Hashicorp go-plugin (usa gRPC sobre proceso hijo)
  - Yaegi (interprete Go embebido)
  - WASM plugins (portable, sandboxed)
`)

	fmt.Println("  Simulando sistema de plugins con registro:")
	registry := NewPluginRegistry()

	registry.Register("uppercase", &UppercasePlugin{})
	registry.Register("reverse", &ReversePlugin{})
	registry.Register("caesar", &CaesarPlugin{shift: 3})

	input := "Hello World"
	fmt.Printf("  Input: %q\n", input)

	for _, name := range registry.List() {
		p, _ := registry.Get(name)
		result, _ := p.Process([]byte(input))
		fmt.Printf("  Plugin %-12s -> %q\n", name, string(result))
	}

	// ============================================
	// 3. EXPVAR - TIPOS BASICOS
	// ============================================
	fmt.Println("\n--- 3. expvar - Tipos basicos ---")
	os.Stdout.WriteString(`
EXPVAR - VARIABLES ATOMICAS PARA MONITOREO:

  expvar.Int:
    NewInt(name) *Int, Add(delta int64), Set(value int64), Value() int64

  expvar.Float:
    NewFloat(name) *Float, Add(delta float64), Set(value float64), Value() float64

  expvar.String:
    NewString(name) *String, Set(value string), Value() string

  Todas implementan expvar.Var (interfaz con String() string)
  y son seguras para acceso concurrente (atomicas).
`)

	httpReqs := expvar.NewInt("http_requests")
	httpErrs := expvar.NewInt("http_errors")
	avgResp := expvar.NewFloat("avg_response_ms")
	svcName := expvar.NewString("service_name")

	svcName.Set("my-api-service")

	for i := 0; i < 150; i++ {
		httpReqs.Add(1)
		if i%15 == 0 {
			httpErrs.Add(1)
		}
	}
	avgResp.Set(23.7)

	fmt.Printf("  service_name:    %s\n", svcName.Value())
	fmt.Printf("  http_requests:   %d\n", httpReqs.Value())
	fmt.Printf("  http_errors:     %d\n", httpErrs.Value())
	fmt.Printf("  avg_response_ms: %.1f\n", avgResp.Value())
	fmt.Printf("  error_rate:      %.2f%%\n", float64(httpErrs.Value())/float64(httpReqs.Value())*100)

	// ============================================
	// 4. EXPVAR - MAP
	// ============================================
	fmt.Println("\n--- 4. expvar - Map ---")
	os.Stdout.WriteString(`
EXPVAR.MAP - MAPA DE VARIABLES:

  expvar.NewMap(name) *Map
  m.Add(key string, delta int64)           // Incrementar contador
  m.Set(key string, var Var)               // Establecer variable
  m.Get(key string) Var                    // Obtener variable
  m.Delete(key string)                     // Eliminar clave
  m.Do(func(KeyValue))                     // Iterar
  m.Init() *Map                            // Reiniciar

  Ideal para metricas por endpoint, por metodo HTTP, etc.
`)

	statusCodes := expvar.NewMap("status_codes")
	statusCodes.Add("200", 1500)
	statusCodes.Add("201", 200)
	statusCodes.Add("400", 45)
	statusCodes.Add("404", 120)
	statusCodes.Add("500", 8)

	fmt.Println("  Status codes:")
	statusCodes.Do(func(kv expvar.KeyValue) {
		fmt.Printf("    HTTP %s: %s requests\n", kv.Key, kv.Value.String())
	})

	methodStats := expvar.NewMap("method_stats")
	methodStats.Add("GET", 800)
	methodStats.Add("POST", 300)
	methodStats.Add("PUT", 100)
	methodStats.Add("DELETE", 50)
	methodStats.Add("PATCH", 25)

	fmt.Println("\n  Method stats:")
	methodStats.Do(func(kv expvar.KeyValue) {
		fmt.Printf("    %-8s: %s\n", kv.Key, kv.Value.String())
	})

	// Map anidado
	endpointMap := expvar.NewMap("endpoints_detail")
	getUsersMap := new(expvar.Map).Init()
	getUsersMap.Add("count", 500)
	getUsersMap.Add("errors", 3)
	endpointMap.Set("GET /users", getUsersMap)

	createOrderMap := new(expvar.Map).Init()
	createOrderMap.Add("count", 150)
	createOrderMap.Add("errors", 12)
	endpointMap.Set("POST /orders", createOrderMap)

	fmt.Println("\n  Endpoints detalle (maps anidados):")
	endpointMap.Do(func(kv expvar.KeyValue) {
		fmt.Printf("    %s: %s\n", kv.Key, kv.Value.String())
	})

	// ============================================
	// 5. EXPVAR - PUBLISH & VARIABLE CUSTOM
	// ============================================
	fmt.Println("\n--- 5. expvar - Variables custom ---")
	os.Stdout.WriteString(`
EXPVAR.PUBLISH - PUBLICAR VARIABLE CUSTOM:

  expvar.Publish(name string, v Var)

  type Var interface {
      String() string   // Debe retornar JSON valido
  }

  expvar.Func(func() any)                 // Funcion como variable
    - Se ejecuta cada vez que se lee
    - Ideal para metricas calculadas en tiempo real
`)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("memory_mb", expvar.Func(func() any {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return map[string]float64{
			"alloc":       float64(m.Alloc) / 1024 / 1024,
			"sys":         float64(m.Sys) / 1024 / 1024,
			"heap_inuse":  float64(m.HeapInuse) / 1024 / 1024,
			"stack_inuse": float64(m.StackInuse) / 1024 / 1024,
		}
	}))

	expvar.Publish("build_info", expvar.Func(func() any {
		return map[string]string{
			"go_version": runtime.Version(),
			"os":         runtime.GOOS,
			"arch":       runtime.GOARCH,
			"cpus":       fmt.Sprintf("%d", runtime.NumCPU()),
		}
	}))

	fmt.Println("  Variables Func (calculadas al leer):")
	readExpvar := func(name string) {
		v := expvar.Get(name)
		if v != nil {
			val := v.String()
			if len(val) > 80 {
				val = val[:80] + "..."
			}
			fmt.Printf("    %-15s = %s\n", name, val)
		}
	}
	readExpvar("goroutines")
	readExpvar("memory_mb")
	readExpvar("build_info")

	// Custom Var con historico
	latencyTracker := &LatencyTracker{}
	expvar.Publish("latency_stats", latencyTracker)

	latencies := []float64{12.5, 15.3, 8.7, 22.1, 11.0, 45.6, 9.8, 14.2, 18.9, 7.3}
	for _, l := range latencies {
		latencyTracker.Record(l)
	}
	fmt.Printf("\n  latency_stats: %s\n", latencyTracker.String())

	// ============================================
	// 6. EXPVAR - HANDLER HTTP
	// ============================================
	fmt.Println("\n--- 6. expvar - HTTP handler ---")
	os.Stdout.WriteString(`
EXPVAR HTTP HANDLER:

  Importar "expvar" automaticamente registra:
    /debug/vars en http.DefaultServeMux

  expvar.Handler() http.Handler            // Handler independiente

  Respuesta JSON:
  {
    "cmdline": [...],
    "memstats": {...},
    "http_requests": 150,
    "status_codes": {"200": 1500, "404": 120},
    "goroutines": 5,
    ...
  }

  Integracion con Prometheus/Grafana:
  - Usar expvar_exporter para convertir a metricas Prometheus
  - O usar un middleware custom que actualice expvars

  Integracion con mux custom:
    mux := http.NewServeMux()
    mux.Handle("/debug/vars", expvar.Handler())
    mux.Handle("/metrics", customMetricsHandler())
`)

	fmt.Println("  Handler HTTP registrado en /debug/vars")
	fmt.Println("  Para usar: http.ListenAndServe(':6060', nil)")
	_ = expvar.Handler()

	// ============================================
	// 7. EXPVAR - DO (ITERAR TODAS)
	// ============================================
	fmt.Println("\n--- 7. expvar - Iterar todas las variables ---")

	fmt.Println("  Todas las variables expvar registradas:")
	count := 0
	expvar.Do(func(kv expvar.KeyValue) {
		count++
		val := kv.Value.String()
		if kv.Key == "memstats" || kv.Key == "cmdline" {
			val = fmt.Sprintf("(%d chars)", len(val))
		} else if len(val) > 50 {
			val = val[:50] + "..."
		}
		fmt.Printf("    [%02d] %-20s = %s\n", count, kv.Key, val)
	})

	// ============================================
	// 8. EJEMPLO: MIDDLEWARE DE METRICAS
	// ============================================
	fmt.Println("\n--- 8. Ejemplo: Middleware de metricas ---")

	metricsMiddleware := NewMetricsMiddleware()

	simulatedRequests := []struct {
		method string
		path   string
		status int
		dur    time.Duration
	}{
		{"GET", "/api/users", 200, 15 * time.Millisecond},
		{"GET", "/api/users/1", 200, 12 * time.Millisecond},
		{"POST", "/api/users", 201, 45 * time.Millisecond},
		{"GET", "/api/products", 200, 8 * time.Millisecond},
		{"GET", "/api/missing", 404, 2 * time.Millisecond},
		{"POST", "/api/orders", 500, 120 * time.Millisecond},
		{"GET", "/api/users", 200, 18 * time.Millisecond},
		{"PUT", "/api/users/1", 200, 25 * time.Millisecond},
		{"DELETE", "/api/users/2", 204, 10 * time.Millisecond},
		{"GET", "/api/products", 200, 9 * time.Millisecond},
	}

	for _, req := range simulatedRequests {
		metricsMiddleware.RecordRequest(req.method, req.path, req.status, req.dur)
	}

	fmt.Println("  Reporte de metricas:")
	metricsMiddleware.PrintReport()

	// ============================================
	// 9. EJEMPLO: SISTEMA DE PLUGINS SIMULADO
	// ============================================
	fmt.Println("\n--- 9. Plugin system completo (simulado) ---")

	engine := NewPluginEngine()

	engine.LoadPlugin("json-formatter", &JSONFormatterPlugin{indent: true})
	engine.LoadPlugin("word-counter", &WordCounterPlugin{})
	engine.LoadPlugin("hash-generator", &HashPlugin{})

	fmt.Println("  Plugins cargados:")
	for _, info := range engine.ListPlugins() {
		fmt.Printf("    - %s (version: %s)\n", info.Name, info.Version)
	}

	testInput := "Go es un lenguaje de programacion open source"
	fmt.Printf("\n  Input: %q\n", testInput)

	for _, name := range []string{"json-formatter", "word-counter", "hash-generator"} {
		result, err := engine.Execute(name, []byte(testInput))
		if err != nil {
			fmt.Printf("  [%s] Error: %v\n", name, err)
		} else {
			fmt.Printf("  [%s] -> %s\n", name, string(result))
		}
	}

	fmt.Println("\n=== FIN CAPITULO 073 ===")
}

// ============================================
// SISTEMA DE PLUGINS (SIMULADO EN PROCESO)
// ============================================

type TextPlugin interface {
	Name() string
	Version() string
	Process(data []byte) ([]byte, error)
}

type PluginRegistry struct {
	mu      sync.RWMutex
	plugins map[string]TextPlugin
}

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{plugins: make(map[string]TextPlugin)}
}

func (r *PluginRegistry) Register(name string, p TextPlugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins[name] = p
}

func (r *PluginRegistry) Get(name string) (TextPlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.plugins[name]
	return p, ok
}

func (r *PluginRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type UppercasePlugin struct{}

func (p *UppercasePlugin) Name() string                        { return "uppercase" }
func (p *UppercasePlugin) Version() string                     { return "1.0.0" }
func (p *UppercasePlugin) Process(data []byte) ([]byte, error) { return []byte(strings.ToUpper(string(data))), nil }

type ReversePlugin struct{}

func (p *ReversePlugin) Name() string    { return "reverse" }
func (p *ReversePlugin) Version() string { return "1.0.0" }
func (p *ReversePlugin) Process(data []byte) ([]byte, error) {
	runes := []rune(string(data))
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return []byte(string(runes)), nil
}

type CaesarPlugin struct{ shift int }

func (p *CaesarPlugin) Name() string    { return "caesar" }
func (p *CaesarPlugin) Version() string { return "1.0.0" }
func (p *CaesarPlugin) Process(data []byte) ([]byte, error) {
	result := make([]byte, len(data))
	for i, b := range data {
		if b >= 'a' && b <= 'z' {
			result[i] = 'a' + byte((int(b-'a')+p.shift)%26)
		} else if b >= 'A' && b <= 'Z' {
			result[i] = 'A' + byte((int(b-'A')+p.shift)%26)
		} else {
			result[i] = b
		}
	}
	return result, nil
}

// ============================================
// LATENCY TRACKER (CUSTOM EXPVAR.VAR)
// ============================================

type LatencyTracker struct {
	mu      sync.Mutex
	values  []float64
	count   int64
	sum     float64
	min     float64
	max     float64
}

func (lt *LatencyTracker) Record(ms float64) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.values = append(lt.values, ms)
	lt.count++
	lt.sum += ms
	if lt.count == 1 || ms < lt.min {
		lt.min = ms
	}
	if ms > lt.max {
		lt.max = ms
	}
}

func (lt *LatencyTracker) String() string {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	if lt.count == 0 {
		return `{"count":0}`
	}
	avg := lt.sum / float64(lt.count)
	p95 := lt.percentile(95)
	data := map[string]any{
		"count": lt.count,
		"avg":   math.Round(avg*100) / 100,
		"min":   lt.min,
		"max":   lt.max,
		"p95":   p95,
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func (lt *LatencyTracker) percentile(p float64) float64 {
	if len(lt.values) == 0 {
		return 0
	}
	sorted := make([]float64, len(lt.values))
	copy(sorted, lt.values)
	sort.Float64s(sorted)
	idx := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

// ============================================
// METRICS MIDDLEWARE
// ============================================

type MetricsMiddleware struct {
	totalRequests *expvar.Int
	totalErrors   *expvar.Int
	byMethod      *expvar.Map
	byStatus      *expvar.Map
	latencies     []float64
	mu            sync.Mutex
}

func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		totalRequests: new(expvar.Int),
		totalErrors:   new(expvar.Int),
		byMethod:      new(expvar.Map).Init(),
		byStatus:      new(expvar.Map).Init(),
	}
}

func (m *MetricsMiddleware) RecordRequest(method, _ string, status int, dur time.Duration) {
	m.totalRequests.Add(1)
	if status >= 400 {
		m.totalErrors.Add(1)
	}
	m.byMethod.Add(method, 1)
	m.byStatus.Add(fmt.Sprintf("%d", status), 1)
	m.mu.Lock()
	m.latencies = append(m.latencies, float64(dur.Milliseconds()))
	m.mu.Unlock()
}

func (m *MetricsMiddleware) PrintReport() {
	fmt.Printf("    Total requests: %d\n", m.totalRequests.Value())
	fmt.Printf("    Total errors:   %d\n", m.totalErrors.Value())

	fmt.Println("    By method:")
	m.byMethod.Do(func(kv expvar.KeyValue) {
		fmt.Printf("      %-8s: %s\n", kv.Key, kv.Value.String())
	})

	fmt.Println("    By status:")
	m.byStatus.Do(func(kv expvar.KeyValue) {
		fmt.Printf("      %s: %s\n", kv.Key, kv.Value.String())
	})

	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.latencies) > 0 {
		var sum float64
		for _, l := range m.latencies {
			sum += l
		}
		avg := sum / float64(len(m.latencies))
		fmt.Printf("    Avg latency: %.1f ms\n", avg)
	}
}

// ============================================
// PLUGIN ENGINE (SIMULADO)
// ============================================

type PluginInfo struct {
	Name    string
	Version string
}

type PluginEngine struct {
	plugins map[string]TextPlugin
}

func NewPluginEngine() *PluginEngine {
	return &PluginEngine{plugins: make(map[string]TextPlugin)}
}

func (e *PluginEngine) LoadPlugin(name string, p TextPlugin) {
	e.plugins[name] = p
}

func (e *PluginEngine) Execute(name string, data []byte) ([]byte, error) {
	p, ok := e.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %q not found", name)
	}
	return p.Process(data)
}

func (e *PluginEngine) ListPlugins() []PluginInfo {
	infos := make([]PluginInfo, 0, len(e.plugins))
	for _, p := range e.plugins {
		infos = append(infos, PluginInfo{Name: p.Name(), Version: p.Version()})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

type JSONFormatterPlugin struct{ indent bool }

func (p *JSONFormatterPlugin) Name() string    { return "json-formatter" }
func (p *JSONFormatterPlugin) Version() string { return "1.1.0" }
func (p *JSONFormatterPlugin) Process(data []byte) ([]byte, error) {
	words := strings.Fields(string(data))
	result := map[string]any{
		"original":   string(data),
		"words":      words,
		"word_count": len(words),
		"char_count": len(data),
	}
	if p.indent {
		return json.MarshalIndent(result, "", "  ")
	}
	return json.Marshal(result)
}

type WordCounterPlugin struct{}

func (p *WordCounterPlugin) Name() string    { return "word-counter" }
func (p *WordCounterPlugin) Version() string { return "1.0.0" }
func (p *WordCounterPlugin) Process(data []byte) ([]byte, error) {
	words := strings.Fields(string(data))
	freq := map[string]int{}
	for _, w := range words {
		freq[strings.ToLower(w)]++
	}
	b, _ := json.Marshal(freq)
	return b, nil
}

type HashPlugin struct{}

func (p *HashPlugin) Name() string    { return "hash-generator" }
func (p *HashPlugin) Version() string { return "1.0.0" }
func (p *HashPlugin) Process(data []byte) ([]byte, error) {
	var hash uint64
	for _, b := range data {
		hash = hash*31 + uint64(b)
	}
	return []byte(fmt.Sprintf("hash:%016x (len:%d)", hash, len(data))), nil
}
