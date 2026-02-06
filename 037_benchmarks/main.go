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

	fmt.Printf("\nGOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
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
