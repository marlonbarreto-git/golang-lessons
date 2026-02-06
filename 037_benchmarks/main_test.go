package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

// ============================================
// BENCHMARK BÁSICO
// ============================================

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(2, 3)
	}
}

// Go 1.24+ style con b.Loop()
func BenchmarkAddLoop(b *testing.B) {
	for b.Loop() {
		Add(2, 3)
	}
}

// ============================================
// COMPARAR IMPLEMENTACIONES
// ============================================

func BenchmarkFibonacci(b *testing.B) {
	b.Run("recursive-10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Fibonacci(10)
		}
	})

	b.Run("iterative-10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FibonacciIterative(10)
		}
	})

	b.Run("recursive-20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Fibonacci(20)
		}
	})

	b.Run("iterative-20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FibonacciIterative(20)
		}
	})
}

// ============================================
// BENCHMARK CON DIFERENTES TAMAÑOS
// ============================================

func BenchmarkIsPrime(b *testing.B) {
	cases := []struct {
		name string
		n    int
	}{
		{"small", 17},
		{"medium", 7919},
		{"large", 104729},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsPrime(tc.n)
			}
		})
	}
}

// ============================================
// BENCHMARK CON SETUP
// ============================================

func BenchmarkSort(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			// Setup: crear slice
			original := make([]int, size)
			for i := range original {
				original[i] = size - i
			}

			b.ResetTimer() // No contar setup

			for i := 0; i < b.N; i++ {
				// Copiar para no ordenar slice ya ordenado
				data := make([]int, len(original))
				copy(data, original)
				sort.Ints(data)
			}
		})
	}
}

// ============================================
// BENCHMARK DE MEMORIA
// ============================================

func BenchmarkAllocation(b *testing.B) {
	b.Run("slice-grow", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var s []int
			for j := 0; j < 100; j++ {
				s = append(s, j)
			}
		}
	})

	b.Run("slice-preallocated", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := make([]int, 0, 100)
			for j := 0; j < 100; j++ {
				s = append(s, j)
			}
		}
	})
}

// ============================================
// BENCHMARK STRING CONCATENATION
// ============================================

func BenchmarkStringConcat(b *testing.B) {
	words := []string{"hello", "world", "from", "go"}

	b.Run("plus", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := ""
			for _, w := range words {
				s += w + " "
			}
			_ = s
		}
	})

	b.Run("sprintf", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := fmt.Sprintf("%s %s %s %s", words[0], words[1], words[2], words[3])
			_ = s
		}
	})

	b.Run("join", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			s := strings.Join(words, " ")
			_ = s
		}
	})

	b.Run("builder", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			for _, w := range words {
				sb.WriteString(w)
				sb.WriteString(" ")
			}
			_ = sb.String()
		}
	})
}

// ============================================
// BENCHMARK PARALELO
// ============================================

func BenchmarkParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			IsPrime(104729)
		}
	})
}

// ============================================
// BENCHMARK CON CUSTOM METRICS
// ============================================

func BenchmarkWithMetrics(b *testing.B) {
	var totalOps int64

	for i := 0; i < b.N; i++ {
		// Simular operación que procesa items
		items := 100
		totalOps += int64(items)
	}

	b.ReportMetric(float64(totalOps)/float64(b.N), "items/op")
}

// ============================================
// BENCHMARK MAP VS SLICE LOOKUP
// ============================================

func BenchmarkLookup(b *testing.B) {
	size := 1000

	// Preparar datos
	slice := make([]int, size)
	mapData := make(map[int]bool, size)
	for i := 0; i < size; i++ {
		slice[i] = i
		mapData[i] = true
	}

	target := size / 2 // Buscar elemento en el medio

	b.Run("slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, v := range slice {
				if v == target {
					break
				}
			}
		}
	})

	b.Run("map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mapData[target]
		}
	})
}

// ============================================
// EVITAR OPTIMIZACIÓN DEL COMPILADOR
// ============================================

var result int // Variable global para evitar optimización

func BenchmarkNoOptimization(b *testing.B) {
	var r int
	for i := 0; i < b.N; i++ {
		r = Fibonacci(10)
	}
	result = r // Asignar a global
}

/*
EJECUTAR ESTOS BENCHMARKS:

# Todos los benchmarks
go test -bench=. -benchmem

# Solo Fibonacci
go test -bench=BenchmarkFibonacci -benchmem

# Con múltiples runs para estadísticas
go test -bench=. -count=5

# Generar profile de CPU
go test -bench=BenchmarkFibonacci -cpuprofile=cpu.prof

# Analizar profile
go tool pprof cpu.prof

# Comparar resultados con benchstat
go test -bench=. -count=10 > before.txt
# ... hacer cambios ...
go test -bench=. -count=10 > after.txt
benchstat before.txt after.txt
*/
