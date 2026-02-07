// Package main - Chapter 100: Performance
// Go es rápido por defecto, pero hay técnicas para
// optimizar cuando es necesario.
package main

import (
	"os"
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("=== PERFORMANCE Y OPTIMIZACIÓN ===")

	// ============================================
	// PROFILING
	// ============================================
	fmt.Println("\n--- Profiling ---")
	fmt.Println(`
TIPOS DE PROFILES:
- CPU: dónde pasa tiempo el programa
- Memory (heap): allocations
- Block: bloqueos en sync primitives
- Mutex: contención de locks
- Goroutine: stack traces de goroutines

BENCHMARK PROFILING:
go test -bench=. -cpuprofile=cpu.prof
go test -bench=. -memprofile=mem.prof
go test -bench=. -blockprofile=block.prof
go test -bench=. -mutexprofile=mutex.prof

RUNTIME PROFILING:
import "runtime/pprof"

f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()
// ... código a perfilar ...

// Heap profile
f, _ := os.Create("mem.prof")
pprof.WriteHeapProfile(f)
f.Close()

HTTP PROFILING (net/http/pprof):
import _ "net/http/pprof"
go http.ListenAndServe("localhost:6060", nil)

// Acceder:
// http://localhost:6060/debug/pprof/
// http://localhost:6060/debug/pprof/profile?seconds=30
// http://localhost:6060/debug/pprof/heap

ANALIZAR:
go tool pprof cpu.prof
> top10
> list FunctionName
> web  # Abre en navegador

go tool pprof -http=:8080 cpu.prof  # Web UI`)
	// ============================================
	// TRACE
	// ============================================
	fmt.Println("\n--- Execution Trace ---")
	fmt.Println(`
import "runtime/trace"

f, _ := os.Create("trace.out")
trace.Start(f)
defer trace.Stop()
// ... código ...

// O con tests:
go test -trace=trace.out

// Analizar:
go tool trace trace.out

TRACE MUESTRA:
- Goroutine scheduling
- System calls
- GC events
- Network I/O
- Blocking operations

FLIGHT RECORDER (Go 1.25+):
recorder := trace.NewFlightRecorder()
defer recorder.Stop()

// Cuando algo interesante pasa:
recorder.WriteTo(file)`)
	// ============================================
	// MEMORY
	// ============================================
	fmt.Println("\n--- Optimización de Memoria ---")
	fmt.Println(`
1. PRE-ALLOCATE SLICES:
// MAL
var s []int
for i := 0; i < 1000; i++ {
    s = append(s, i)  // Múltiples reallocations
}

// BIEN
s := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
    s = append(s, i)  // Sin reallocations
}

2. REUTILIZAR BUFFERS (sync.Pool):
var bufPool = sync.Pool{
    New: func() any {
        return make([]byte, 4096)
    },
}

func process(data []byte) {
    buf := bufPool.Get().([]byte)
    defer bufPool.Put(buf)
    // usar buf
}

3. EVITAR ESCAPES AL HEAP:
// MAL - escapa al heap
func newUser() *User {
    return &User{Name: "Alice"}
}

// Verificar escapes:
go build -gcflags="-m"

4. STRINGS VS []BYTE:
// Strings son inmutables - concatenar crea copias
// Para muchas operaciones, usar []byte y strings.Builder

var sb strings.Builder
sb.Grow(100)  // Pre-allocate
for _, s := range parts {
    sb.WriteString(s)
}
result := sb.String()

5. STRUCT PADDING:
// MAL - 24 bytes con padding
type Bad struct {
    a bool   // 1 + 7 padding
    b int64  // 8
    c bool   // 1 + 7 padding
}

// BIEN - 16 bytes
type Good struct {
    b int64  // 8
    a bool   // 1
    c bool   // 1 + 6 padding
}`)
	// ============================================
	// CONCURRENCIA
	// ============================================
	fmt.Println("\n--- Optimización de Concurrencia ---")
	fmt.Println(`
1. GOMAXPROCS:
runtime.GOMAXPROCS(runtime.NumCPU())  // Por defecto
// Container-aware en Go 1.25+

2. EVITAR CONTENCIÓN:
// MAL - un lock para todo
type Cache struct {
    mu   sync.Mutex
    data map[string]string
}

// BIEN - sharding
type ShardedCache struct {
    shards [256]struct {
        mu   sync.Mutex
        data map[string]string
    }
}

func (c *ShardedCache) getShard(key string) *shard {
    hash := fnv.New32a()
    hash.Write([]byte(key))
    return &c.shards[hash.Sum32()%256]
}

3. ATOMIC VS MUTEX:
// Para contadores simples, usar atomic
var counter atomic.Int64
counter.Add(1)

// Para operaciones complejas, usar mutex

4. CHANNEL BUFFER SIZE:
// Unbuffered: sincronización
// Buffered: throughput
ch := make(chan Item, 100)

5. WORKER POOLS:
// Reutilizar goroutines
jobs := make(chan Job, 100)
for i := 0; i < numWorkers; i++ {
    go worker(jobs)
}`)
	// ============================================
	// GC TUNING
	// ============================================
	fmt.Println("\n--- GC Tuning ---")
	os.Stdout.WriteString(`
GOGC:
GOGC=100  # Default - GC cuando heap crece 100%
GOGC=200  # Menos GC, más memoria
GOGC=50   # Más GC, menos memoria
GOGC=off  # Deshabilitar GC (cuidado!)

GOMEMLIMIT (Go 1.19+):
GOMEMLIMIT=1GiB  # Límite suave de memoria
// GC se ejecuta más agresivamente cerca del límite

RUNTIME:
// Forzar GC
runtime.GC()

// Estadísticas
var stats runtime.MemStats
runtime.ReadMemStats(&stats)
fmt.Printf("Alloc: %d MB\n", stats.Alloc/1024/1024)
fmt.Printf("NumGC: %d\n", stats.NumGC)

// Debug GC
GODEBUG=gctrace=1 ./myapp

GREEN TEA GC (Go 1.25+ experimental):
GOEXPERIMENT=greenteagc go build
// 10-40% menos overhead de GC
`)

	// ============================================
	// INLINING Y COMPILER
	// ============================================
	fmt.Println("\n--- Compiler Optimizations ---")
	fmt.Println(`
INLINING:
// Funciones pequeñas se inlinean automáticamente
// Ver qué se inlinea:
go build -gcflags="-m"

// Forzar no-inline (para debugging)
//go:noinline
func myFunc() {}

BOUNDS CHECK ELIMINATION:
// El compilador elimina bounds checks cuando puede probar seguridad
for i := 0; i < len(s); i++ {
    _ = s[i]  // Sin bounds check (compilador sabe que i < len)
}

// Forzar eliminación:
_ = s[len(s)-1]  // Hint al compilador

ESCAPE ANALYSIS:
go build -gcflags="-m"

ASSEMBLY:
go build -gcflags="-S"  # Ver assembly generado
go tool compile -S main.go`)
	// ============================================
	// BENCHMARKING TIPS
	// ============================================
	fmt.Println("\n--- Benchmarking Tips ---")
	fmt.Println(`
1. WARMUP:
// El compilador optimiza durante la ejecución
// Primeras iteraciones pueden ser más lentas

2. EVITAR OPTIMIZACIÓN DEL COMPILADOR:
var result int  // Global

func BenchmarkFoo(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = Foo()
    }
    result = r  // Evita que el compilador elimine el código
}

3. USAR b.ResetTimer():
func BenchmarkWithSetup(b *testing.B) {
    // Setup costoso
    data := loadData()
    b.ResetTimer()  // No contar setup
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

4. COMPARAR CON BENCHSTAT:
go test -bench=. -count=10 > old.txt
# hacer cambios
go test -bench=. -count=10 > new.txt
benchstat old.txt new.txt

5. PROFILE BENCHMARKS:
go test -bench=BenchmarkFoo -cpuprofile=cpu.prof`)
	// ============================================
	// HERRAMIENTAS
	// ============================================
	fmt.Println("\n--- Herramientas ---")
	fmt.Println(`
PROFILING:
- go tool pprof
- go tool trace
- Pyroscope (continuous profiling)
- Datadog profiler

BENCHMARKING:
- go test -bench
- benchstat
- hey (HTTP load testing)
- wrk

MEMORY:
- go tool pprof mem.prof
- GODEBUG=allocfreetrace=1

CONCURRENCY:
- go build -race
- go vet

LINTING:
- staticcheck
- golangci-lint

VISUALIZATION:
- go tool pprof -http=:8080 profile
- go tool trace
- flamegraph`)
	fmt.Printf("\nGOMAXPROCS actual: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
	fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())
}

/*
RESUMEN DE PERFORMANCE:

PROFILING:
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof
net/http/pprof para runtime

TRACE:
go tool trace trace.out

MEMORIA:
- Pre-allocate slices
- sync.Pool para buffers
- strings.Builder
- Struct padding
- Escape analysis: -gcflags="-m"

CONCURRENCIA:
- Sharding para reducir contención
- atomic para contadores
- Sized channels
- Worker pools

GC:
GOGC=100 (default)
GOMEMLIMIT=1GiB
runtime.ReadMemStats()

COMPILER:
-gcflags="-m" escape analysis
-gcflags="-S" assembly
//go:noinline

BENCHMARK:
- b.ResetTimer()
- Evitar optimización
- benchstat
- -count=10

HERRAMIENTAS:
pprof, trace, benchstat
race detector, vet
golangci-lint
*/

/*
SUMMARY - CHAPTER 100: Performance

TOPIC: Performance Optimization and Profiling
- CPU profiling: go test -cpuprofile, runtime/pprof.StartCPUProfile(), analyze with pprof tool (top10, list, web)
- Memory profiling: go test -memprofile, pprof.WriteHeapProfile(), find allocations and memory leaks
- Execution tracing: runtime/trace for goroutine scheduling, GC events, network I/O, blocking operations
- HTTP profiling: net/http/pprof automatic endpoints, access /debug/pprof/ for live profiling
- Memory optimization: Pre-allocate slices with make([]T, 0, capacity), sync.Pool for buffer reuse
- Escape analysis: go build -gcflags="-m" shows heap allocations, avoid returning pointers for stack allocation
- String optimization: Use strings.Builder with Grow() for concatenation, []byte for mutable operations
- Struct padding: Arrange fields by size (largest first) to minimize memory waste from alignment
- Concurrency optimization: Sharded maps to reduce lock contention, atomic for simple counters, buffered channels
- GC tuning: GOGC (default 100), GOMEMLIMIT for soft memory limit, runtime.ReadMemStats() for diagnostics
- Compiler optimizations: Inlining for small functions, bounds check elimination, escape analysis for stack allocation
- Benchmarking: b.ResetTimer() to exclude setup, assign result to global to prevent dead code elimination
- benchstat: Statistical comparison of benchmark results, detects significant performance changes
- Best practices: Profile before optimizing, use benchstat for comparisons, run with -count=10 for stability
*/
