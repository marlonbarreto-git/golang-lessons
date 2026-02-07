// Package main - Chapter 037: Benchmarks
// Go incluye herramientas de benchmarking integradas en el paquete testing.
// Aprenderás a medir y optimizar el rendimiento de tu código.
package main

import (
	"os"
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("=== BENCHMARKS EN GO ===")

	// ============================================
	// BENCHMARK BASICS
	// ============================================
	os.Stdout.WriteString(`
Este archivo explica benchmarks. Los benchmarks reales están en main_test.go.

Ejecutar benchmarks:
  go test -bench=.              # Todos los benchmarks
  go test -bench=BenchmarkAdd   # Benchmark específico
  go test -bench=. -benchmem    # Con estadísticas de memoria
  go test -bench=. -count=5     # Repetir 5 veces
  go test -bench=. -benchtime=5s # Ejecutar por 5 segundos
  go test -bench=. -cpuprofile=cpu.prof  # Generar profile

ANATOMÍA DE UN BENCHMARK:

func BenchmarkXxx(b *testing.B) {
    // Setup (no se mide)
    data := prepareData()

    b.ResetTimer() // Reiniciar timer después de setup

    for i := 0; i < b.N; i++ {
        // Código a medir
        processData(data)
    }
}

b.N es determinado automáticamente para obtener mediciones estables.

---

BENCHMARK CON b.Loop() (Go 1.24+):

func BenchmarkModern(b *testing.B) {
    for b.Loop() {
        // Código a medir
    }
}

Ventajas:
- Setup/teardown fuera del loop no se mide
- Previene optimización del compilador
- Más ergonómico

---

MÉTRICAS COMUNES:

BenchmarkExample-8   1000000   1234 ns/op   256 B/op   4 allocs/op
                     ↑         ↑            ↑          ↑
                     |         |            |          allocations por operación
                     |         |            bytes por operación
                     |         nanosegundos por operación
                     iteraciones

---

BENCHSTAT (análisis estadístico):

go install golang.org/x/perf/cmd/benchstat@latest

go test -bench=. -count=10 > old.txt
# hacer cambios
go test -bench=. -count=10 > new.txt
benchstat old.txt new.txt

---

SUB-BENCHMARKS:

func BenchmarkSort(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}
    for _, size := range sizes {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            data := generateData(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                sort.Ints(data)
            }
        })
    }
}

---

REPORTING CUSTOM METRICS:

func BenchmarkCustom(b *testing.B) {
    var totalBytes int64
    for i := 0; i < b.N; i++ {
        totalBytes += process()
    }
    b.ReportMetric(float64(totalBytes)/float64(b.N), "bytes/op")
}

---

PARALLEL BENCHMARKS:

func BenchmarkParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            // Operación thread-safe
        }
    })
}

---

EVITAR OPTIMIZACIÓN DEL COMPILADOR:

var result int  // Variable global

func BenchmarkFoo(b *testing.B) {
    var r int
    for i := 0; i < b.N; i++ {
        r = Foo()  // Asignar a variable local
    }
    result = r  // Asignar a global para evitar optimización
}

---

MEMORY PROFILING:

func BenchmarkAlloc(b *testing.B) {
    b.ReportAllocs()  // Reportar allocations
    for i := 0; i < b.N; i++ {
        _ = make([]byte, 1024)
    }
}

---

BENCHMARK DE FUNCIONES CON SETUP COSTOSO:

func BenchmarkExpensive(b *testing.B) {
    // Setup costoso
    db := connectToDatabase()
    defer db.Close()

    b.ResetTimer()  // No contar el setup
    for i := 0; i < b.N; i++ {
        db.Query("SELECT 1")
    }
}

---

COMPARAR IMPLEMENTACIONES:

func BenchmarkStringConcat(b *testing.B) {
    b.Run("plus", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            s := "hello" + " " + "world"
            _ = s
        }
    })
    b.Run("sprintf", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            s := fmt.Sprintf("%s %s", "hello", "world")
            _ = s
        }
    })
    b.Run("builder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var sb strings.Builder
            sb.WriteString("hello")
            sb.WriteString(" ")
            sb.WriteString("world")
            _ = sb.String()
        }
    })
}

`)

	// ============================================
	// RUNTIME INFO
	// ============================================
	fmt.Printf("\nGOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())

	// ============================================
	// b.ReportAllocs() DEEP DIVE
	// ============================================
	fmt.Println("\n--- b.ReportAllocs() ---")
	os.Stdout.WriteString(`
b.ReportAllocs() enables memory allocation statistics for the benchmark.
Equivalent to running with -benchmem flag but per-benchmark.

func BenchmarkSliceAppend(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        s := make([]int, 0)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}

func BenchmarkSlicePrealloc(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        s := make([]int, 0, 100)
        for j := 0; j < 100; j++ {
            s = append(s, j)
        }
    }
}

Output shows:
  BenchmarkSliceAppend-8     500000   3200 ns/op   2040 B/op   8 allocs/op
  BenchmarkSlicePrealloc-8  1000000   1100 ns/op    896 B/op   1 allocs/op

The allocs/op and B/op columns reveal allocation behavior.
Use this to identify and reduce allocations in hot paths.
`)

	// ============================================
	// b.ResetTimer() DEEP DIVE
	// ============================================
	fmt.Println("--- b.ResetTimer() ---")
	os.Stdout.WriteString(`
b.ResetTimer() zeroes the elapsed benchmark time and memory counters.
Use it after expensive setup that should NOT count toward the measurement.

func BenchmarkDatabaseQuery(b *testing.B) {
    // Expensive setup
    db := connectToTestDB()
    populateTestData(db, 10000)

    b.ResetTimer() // Reset: setup time doesn't count

    for i := 0; i < b.N; i++ {
        db.Query("SELECT * FROM users WHERE id = ?", i)
    }
}

b.StopTimer() / b.StartTimer() for per-iteration setup:

func BenchmarkSort(b *testing.B) {
    for i := 0; i < b.N; i++ {
        b.StopTimer()
        data := generateRandomSlice(1000) // don't measure this
        b.StartTimer()

        sort.Ints(data) // only measure this
    }
}

WARNING: StopTimer/StartTimer add overhead. Prefer ResetTimer
for one-time setup. Use StopTimer/StartTimer only when needed.
`)

	// ============================================
	// SUB-BENCHMARKS
	// ============================================
	fmt.Println("--- Sub-Benchmarks ---")
	os.Stdout.WriteString(`
Sub-benchmarks use b.Run() to organize and parameterize benchmarks.

func BenchmarkMapAccess(b *testing.B) {
    for _, size := range []int{10, 100, 1000, 10000} {
        b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
            m := make(map[int]int, size)
            for i := 0; i < size; i++ {
                m[i] = i
            }
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                _ = m[i%size]
            }
        })
    }
}

Output:
  BenchmarkMapAccess/size-10-8      200000000   8 ns/op
  BenchmarkMapAccess/size-100-8     100000000   12 ns/op
  BenchmarkMapAccess/size-1000-8     50000000   18 ns/op
  BenchmarkMapAccess/size-10000-8    30000000   35 ns/op

Run specific sub-benchmark:
  go test -bench=BenchmarkMapAccess/size-1000
  go test -bench=/size-10$   (regex match)
`)

	// ============================================
	// BENCHMARK COMPARISON TOOLS
	// ============================================
	fmt.Println("--- Benchmark Comparison Tools ---")
	os.Stdout.WriteString(`
BENCHSTAT: Statistical comparison of benchmark results.

Install:
  go install golang.org/x/perf/cmd/benchstat@latest

Usage:
  # Run benchmarks multiple times for statistical significance
  go test -bench=. -count=10 > old.txt

  # Make your optimization changes

  go test -bench=. -count=10 > new.txt

  # Compare
  benchstat old.txt new.txt

Output:
  name       old time/op  new time/op  delta
  Parse-8    450ns ± 3%   320ns ± 2%   -28.89%  (p=0.000 n=10+10)
  Format-8   1.2µs ± 1%   1.1µs ± 2%    -8.33%  (p=0.001 n=10+10)

The p-value and confidence interval tell you if the change is real.

---

BENCHCMP (deprecated, use benchstat):
  benchcmp old.txt new.txt

---

TIPS FOR RELIABLE COMPARISONS:
  1. Use -count=10 or higher for statistical significance
  2. Close other applications during benchmarking
  3. Run on the same hardware with consistent load
  4. Use -benchtime=3s for more stable results
  5. Check the ± variance - high variance means unreliable
`)

	// ============================================
	// PROFILING FROM BENCHMARKS
	// ============================================
	fmt.Println("--- Profiling from Benchmarks ---")
	os.Stdout.WriteString(`
Benchmarks can generate CPU and memory profiles for pprof analysis.

CPU PROFILE:
  go test -bench=BenchmarkFoo -cpuprofile=cpu.prof
  go tool pprof cpu.prof

MEMORY PROFILE:
  go test -bench=BenchmarkFoo -memprofile=mem.prof
  go tool pprof mem.prof

BLOCK PROFILE (goroutine blocking):
  go test -bench=BenchmarkFoo -blockprofile=block.prof
  go tool pprof block.prof

MUTEX PROFILE (mutex contention):
  go test -bench=BenchmarkFoo -mutexprofile=mutex.prof
  go tool pprof mutex.prof

PPROF COMMANDS:
  (pprof) top 10          # top 10 functions by time
  (pprof) list FuncName   # show annotated source
  (pprof) web             # open flame graph in browser
  (pprof) svg > out.svg   # export as SVG

TRACE:
  go test -bench=BenchmarkFoo -trace=trace.out
  go tool trace trace.out

EXAMPLE WORKFLOW:
  # 1. Identify slow function with benchmark
  go test -bench=BenchmarkParse -cpuprofile=cpu.prof

  # 2. Analyze with pprof
  go tool pprof -http=:8080 cpu.prof

  # 3. Look at flame graph, find hot paths
  # 4. Optimize the hot path
  # 5. Re-run benchmark to verify improvement
  go test -bench=BenchmarkParse -count=10 > new.txt
  benchstat old.txt new.txt
`)
}

// Funciones para benchmarking (ver main_test.go)

func Add(a, b int) int {
	return a + b
}

func Fibonacci(n int) int {
	if n < 2 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

func FibonacciIterative(n int) int {
	if n < 2 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

/*
RESUMEN DE BENCHMARKS:

ESTRUCTURA BÁSICA:
func BenchmarkXxx(b *testing.B) {
    for i := 0; i < b.N; i++ { ... }
}

GO 1.24+ STYLE:
func BenchmarkXxx(b *testing.B) {
    for b.Loop() { ... }
}

COMANDOS:
go test -bench=.              # Ejecutar todos
go test -bench=. -benchmem    # Con memoria
go test -bench=. -count=10    # Múltiples runs
go test -bench=. -benchtime=5s # Duración

MÉTODOS DE testing.B:
- b.N: número de iteraciones
- b.Loop(): loop moderno (Go 1.24+)
- b.ResetTimer(): reiniciar timer
- b.StopTimer() / b.StartTimer(): pausar timer
- b.ReportAllocs(): reportar allocations
- b.SetBytes(n): bytes por operación
- b.ReportMetric(n, unit): métrica custom
- b.Run(name, fn): sub-benchmarks
- b.RunParallel(fn): benchmark paralelo

BUENAS PRÁCTICAS:
1. Usar b.ResetTimer() después de setup
2. Usar -benchmem para memory stats
3. Ejecutar múltiples veces con -count=N
4. Usar benchstat para comparar
5. Evitar optimización del compilador
6. Benchmarkear casos reales
7. Usar sub-benchmarks para comparar
*/

/*
SUMMARY - CHAPTER 037: BENCHMARKS

BENCHMARK BASICS:
- Integrated in testing package
- Measure performance and memory usage
- Coverage-guided optimization
- Statistical comparison with benchstat

STRUCTURE:
- func BenchmarkXxx(b *testing.B)
- Loop b.N times (determined automatically)
- b.Loop() modern style (Go 1.24+)

RUNNING BENCHMARKS:
- go test -bench=.: run all benchmarks
- go test -bench=BenchmarkName: specific benchmark
- -benchmem: include memory statistics
- -count=N: run N times for statistical significance
- -benchtime=Ns: run for N seconds

TIMING CONTROL:
- b.ResetTimer(): zero timer after setup
- b.StopTimer()/b.StartTimer(): pause/resume
- Only measured code contributes to results

MEMORY PROFILING:
- b.ReportAllocs(): enable allocation tracking
- Shows B/op (bytes per op) and allocs/op
- Identify allocation hotspots

SUB-BENCHMARKS:
- b.Run(name, func): organize related benchmarks
- Compare implementations side-by-side
- Parameterize with different inputs

AVOIDING COMPILER OPTIMIZATION:
- Assign to global variable
- Use result in meaningful way
- Prevent dead code elimination

BENCHSTAT:
- Statistical comparison tool
- go test -count=10 > old.txt
- benchstat old.txt new.txt
- Shows confidence intervals and p-values

PROFILING FROM BENCHMARKS:
- -cpuprofile=cpu.prof
- -memprofile=mem.prof
- go tool pprof for analysis
*/
