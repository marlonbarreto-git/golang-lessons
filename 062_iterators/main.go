// Package main - Chapter 062: Iterators
// Go 1.23+ introduce iteradores como funciones que se pueden usar con range.
// Esta es una de las características más importantes del Go moderno.
package main

import (
	"fmt"
	"iter"
	"slices"
)

func main() {
	fmt.Println("=== ITERATORS EN GO (Go 1.23+) ===")

	// ============================================
	// CONCEPTO BÁSICO
	// ============================================
	fmt.Println("\n--- Concepto Básico ---")
	fmt.Println(`
Go 1.23 introduce "range over functions" - iteradores como funciones.

Antes (Go < 1.23):
for i, v := range slice { ... }     // Solo slices, maps, channels
for k, v := range myMap { ... }

Ahora (Go 1.23+):
for v := range myIterator { ... }   // ¡Funciones también!
for k, v := range myIterator2 { ... }

TIPOS DE ITERADORES:
- iter.Seq[V]      - Un valor por iteración
- iter.Seq2[K, V]  - Dos valores por iteración (key-value)`)
	// ============================================
	// ITER.SEQ - UN VALOR
	// ============================================
	fmt.Println("\n--- iter.Seq[V] - Un Valor ---")

	// Iterador simple que genera números
	fmt.Println("Range(5):")
	for n := range Range(5) {
		fmt.Printf("  %d\n", n)
	}

	// Iterador sobre slice filtrado
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println("\nNúmeros pares de 1-10:")
	for n := range Filter(slices.Values(nums), func(n int) bool { return n%2 == 0 }) {
		fmt.Printf("  %d\n", n)
	}

	// ============================================
	// ITER.SEQ2 - DOS VALORES
	// ============================================
	fmt.Println("\n--- iter.Seq2[K, V] - Dos Valores ---")

	// Iterador con índice
	words := []string{"go", "rust", "python"}
	fmt.Println("Enumerate:")
	for i, w := range Enumerate(slices.Values(words)) {
		fmt.Printf("  %d: %s\n", i, w)
	}

	// Iterador de pares clave-valor
	fmt.Println("\nZip dos slices:")
	keys := []string{"a", "b", "c"}
	values := []int{1, 2, 3}
	for k, v := range Zip(keys, values) {
		fmt.Printf("  %s -> %d\n", k, v)
	}

	// ============================================
	// SLICES PACKAGE CON ITERADORES
	// ============================================
	fmt.Println("\n--- Package slices con Iteradores ---")

	data := []int{3, 1, 4, 1, 5, 9, 2, 6}

	// slices.Values - iterador sobre valores
	fmt.Println("slices.Values:")
	for v := range slices.Values(data) {
		fmt.Printf(" %d", v)
	}
	fmt.Println()

	// slices.All - iterador con índices
	fmt.Println("slices.All (índice, valor):")
	for i, v := range slices.All(data) {
		fmt.Printf("  [%d]=%d\n", i, v)
	}

	// slices.Backward - iterador reverso
	fmt.Println("slices.Backward:")
	for i, v := range slices.Backward(data) {
		fmt.Printf("  [%d]=%d\n", i, v)
	}

	// slices.Collect - convertir iterador a slice
	doubled := slices.Collect(Map(slices.Values(data), func(n int) int { return n * 2 }))
	fmt.Printf("Doubled: %v\n", doubled)

	// ============================================
	// MAPS PACKAGE CON ITERADORES
	// ============================================
	fmt.Println("\n--- Package maps con Iteradores ---")
	fmt.Println(`
import "maps"

m := map[string]int{"a": 1, "b": 2, "c": 3}

// Iterar sobre keys
for k := range maps.Keys(m) {
    fmt.Println(k)
}

// Iterar sobre values
for v := range maps.Values(m) {
    fmt.Println(v)
}

// Iterar sobre pares
for k, v := range maps.All(m) {
    fmt.Println(k, v)
}

// Crear map desde iterador
newMap := maps.Collect(myIterator)`)
	// ============================================
	// CREAR ITERADORES PERSONALIZADOS
	// ============================================
	fmt.Println("\n--- Crear Iteradores Personalizados ---")
	fmt.Println(`
FIRMA DE ITERADORES:

// Un valor
type Seq[V any] func(yield func(V) bool)

// Dos valores
type Seq2[K, V any] func(yield func(K, V) bool)

EJEMPLO - Fibonacci infinito:

func Fibonacci() iter.Seq[int] {
    return func(yield func(int) bool) {
        a, b := 0, 1
        for {
            if !yield(a) {
                return  // El consumidor terminó
            }
            a, b = b, a+b
        }
    }
}

// Uso con Take para limitar
for n := range Take(Fibonacci(), 10) {
    fmt.Println(n)
}`)
	// Demostración de Fibonacci
	fmt.Println("Fibonacci (primeros 10):")
	for n := range Take(Fibonacci(), 10) {
		fmt.Printf("  %d\n", n)
	}

	// ============================================
	// PULL ITERATORS
	// ============================================
	fmt.Println("\n--- Pull Iterators ---")
	fmt.Println(`
iter.Pull convierte un "push iterator" en "pull iterator":

// Push: el iterador controla el flujo
for v := range mySeq { ... }

// Pull: tú controlas el flujo
next, stop := iter.Pull(mySeq)
defer stop()

v1, ok := next()  // Obtener siguiente valor
v2, ok := next()  // Obtener otro
stop()            // Terminar anticipadamente

CUÁNDO USAR PULL:
- Cuando necesitas pausar la iteración
- Cuando necesitas mezclar múltiples iteradores
- Cuando el control de flujo es complejo`)
	// Demostración de Pull
	next, stop := iter.Pull(Range(5))
	defer stop()

	fmt.Println("Pull iterator:")
	if v, ok := next(); ok {
		fmt.Printf("  Primer valor: %d\n", v)
	}
	if v, ok := next(); ok {
		fmt.Printf("  Segundo valor: %d\n", v)
	}
	// Saltar el resto

	// ============================================
	// COMPOSICIÓN DE ITERADORES
	// ============================================
	fmt.Println("\n--- Composición de Iteradores ---")

	// Pipeline: Range -> Filter -> Map -> Take
	result := slices.Collect(
		Take(
			Map(
				Filter(Range(100), func(n int) bool { return n%2 == 0 }),
				func(n int) int { return n * n },
			),
			5,
		),
	)
	fmt.Printf("Primeros 5 cuadrados de números pares: %v\n", result)

	// ============================================
	// ITERADORES Y EARLY RETURN
	// ============================================
	fmt.Println("\n--- Early Return en Iteradores ---")
	fmt.Println(`
El yield retorna false cuando el consumidor termina:

func MyIterator() iter.Seq[int] {
    return func(yield func(int) bool) {
        for i := 0; i < 1000; i++ {
            if !yield(i) {
                // Cleanup si es necesario
                return
            }
        }
    }
}

// El break hace que yield retorne false
for v := range MyIterator() {
    if v > 5 {
        break  // yield(6) retorna false
    }
    fmt.Println(v)
}`)
	// Demostración
	fmt.Println("Break en iterador:")
	for n := range Range(1000) {
		if n >= 3 {
			break
		}
		fmt.Printf("  %d\n", n)
	}
	fmt.Println("  (salió temprano)")

	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrones Comunes ---")
	fmt.Println(`
1. GENERAR SECUENCIAS:
func Range(n int) iter.Seq[int]
func Repeat[V any](v V) iter.Seq[V]
func Cycle[V any](s []V) iter.Seq[V]

2. TRANSFORMAR:
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U]
func Filter[V any](seq iter.Seq[V], pred func(V) bool) iter.Seq[V]
func FlatMap[T, U any](seq iter.Seq[T], f func(T) iter.Seq[U]) iter.Seq[U]

3. LIMITAR:
func Take[V any](seq iter.Seq[V], n int) iter.Seq[V]
func TakeWhile[V any](seq iter.Seq[V], pred func(V) bool) iter.Seq[V]
func Skip[V any](seq iter.Seq[V], n int) iter.Seq[V]

4. COMBINAR:
func Zip[K, V any](keys []K, vals []V) iter.Seq2[K, V]
func Chain[V any](seqs ...iter.Seq[V]) iter.Seq[V]
func Enumerate[V any](seq iter.Seq[V]) iter.Seq2[int, V]

5. REDUCIR:
func Reduce[T, U any](seq iter.Seq[T], init U, f func(U, T) U) U
func Count[V any](seq iter.Seq[V]) int
func Any[V any](seq iter.Seq[V], pred func(V) bool) bool
func All[V any](seq iter.Seq[V], pred func(V) bool) bool`)
	// ============================================
	// RENDIMIENTO
	// ============================================
	fmt.Println("\n--- Rendimiento ---")
	fmt.Println(`
Los iteradores tienen overhead mínimo:
- Inlining agresivo por el compilador
- Sin allocaciones para iteradores simples
- Comparable a loops tradicionales

BENCHMARK TÍPICO:
BenchmarkTraditionalLoop-8    1000000    1.2 ns/op    0 B/op
BenchmarkIterator-8           1000000    1.5 ns/op    0 B/op

TIPS:
1. Evitar closures que escapan al heap
2. Usar iteradores para código más limpio
3. El overhead es negligible para la mayoría de casos
4. Perfilar si el rendimiento es crítico`)}

// ============================================
// IMPLEMENTACIÓN DE ITERADORES
// ============================================

// Range genera números de 0 a n-1
func Range(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 0; i < n; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

// Filter filtra elementos que cumplen el predicado
func Filter[V any](seq iter.Seq[V], pred func(V) bool) iter.Seq[V] {
	return func(yield func(V) bool) {
		for v := range seq {
			if pred(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Map transforma cada elemento
func Map[T, U any](seq iter.Seq[T], f func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Take toma los primeros n elementos
func Take[V any](seq iter.Seq[V], n int) iter.Seq[V] {
	return func(yield func(V) bool) {
		count := 0
		for v := range seq {
			if count >= n {
				return
			}
			if !yield(v) {
				return
			}
			count++
		}
	}
}

// Enumerate agrega índices a un iterador
func Enumerate[V any](seq iter.Seq[V]) iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		i := 0
		for v := range seq {
			if !yield(i, v) {
				return
			}
			i++
		}
	}
}

// Zip combina dos slices en pares
func Zip[K, V any](keys []K, vals []V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		n := len(keys)
		if len(vals) < n {
			n = len(vals)
		}
		for i := 0; i < n; i++ {
			if !yield(keys[i], vals[i]) {
				return
			}
		}
	}
}

// Fibonacci genera la secuencia de Fibonacci infinita
func Fibonacci() iter.Seq[int] {
	return func(yield func(int) bool) {
		a, b := 0, 1
		for {
			if !yield(a) {
				return
			}
			a, b = b, a+b
		}
	}
}

/*
RESUMEN DE ITERATORS:

TIPOS:
iter.Seq[V]       - func(yield func(V) bool)
iter.Seq2[K, V]   - func(yield func(K, V) bool)

PACKAGES CON SOPORTE:
slices.Values(s)     - valores de slice
slices.All(s)        - índice y valor
slices.Backward(s)   - reverso
slices.Collect(seq)  - iterador a slice
maps.Keys(m)         - claves de map
maps.Values(m)       - valores de map
maps.All(m)          - clave y valor

CREAR ITERADORES:
func MySeq() iter.Seq[T] {
    return func(yield func(T) bool) {
        for ... {
            if !yield(value) {
                return  // Consumidor terminó
            }
        }
    }
}

PULL ITERATORS:
next, stop := iter.Pull(seq)
defer stop()
v, ok := next()

PATRONES:
- Filter, Map, Take, Skip
- Enumerate, Zip, Chain
- Composición: seq | filter | map | take

VENTAJAS:
1. Código más expresivo y componible
2. Lazy evaluation (eficiente)
3. Integración con range nativo
4. Sin allocaciones innecesarias

CUÁNDO USAR:
- Procesamiento de colecciones
- Generación lazy de datos
- Pipelines de transformación
- APIs más expresivas
*/
