// Package main - Chapter 020: Generics
// Generics (Go 1.18+) permiten escribir código que funciona con múltiples tipos.
// Aprenderás type parameters, constraints, y cuándo usar generics.
package main

import (
	"cmp"
	"fmt"
	"slices"
)

func main() {
	fmt.Println("=== GENERICS EN GO ===")

	// ============================================
	// PROBLEMA SIN GENERICS
	// ============================================
	fmt.Println("\n--- Sin Generics (el problema) ---")

	// Antes de generics, necesitabas funciones separadas o interface{}
	fmt.Printf("MinInt(3, 5) = %d\n", MinInt(3, 5))
	fmt.Printf("MinFloat(3.14, 2.71) = %.2f\n", MinFloat(3.14, 2.71))
	fmt.Printf("MinString(\"b\", \"a\") = %s\n", MinString("b", "a"))

	// O usar interface{} (perdiendo type safety)
	fmt.Printf("MinAny(3, 5) = %v\n", MinAny(3, 5))

	// ============================================
	// FUNCIONES GENÉRICAS
	// ============================================
	fmt.Println("\n--- Funciones Genéricas ---")

	// Una función que funciona con cualquier tipo ordenado
	fmt.Printf("Min(3, 5) = %d\n", Min(3, 5))
	fmt.Printf("Min(3.14, 2.71) = %.2f\n", Min(3.14, 2.71))
	fmt.Printf("Min(\"banana\", \"apple\") = %s\n", Min("banana", "apple"))

	// El tipo se infiere automáticamente
	result := Min(10, 20)
	fmt.Printf("Min inferido: %d (tipo: %T)\n", result, result)

	// También puedes especificar el tipo explícitamente
	result2 := Min[int64](10, 20)
	fmt.Printf("Min[int64]: %d (tipo: %T)\n", result2, result2)

	// ============================================
	// CONSTRAINTS
	// ============================================
	fmt.Println("\n--- Constraints ---")

	// any = interface{} - acepta cualquier tipo
	PrintAny(42)
	PrintAny("hello")
	PrintAny([]int{1, 2, 3})

	// comparable - tipos que soportan == y !=
	fmt.Printf("Equal(5, 5) = %t\n", Equal(5, 5))
	fmt.Printf("Equal(\"a\", \"b\") = %t\n", Equal("a", "b"))

	// cmp.Ordered - tipos que soportan < <= > >=
	fmt.Printf("Max(10, 20) = %d\n", Max(10, 20))
	fmt.Printf("Max(3.14, 2.71) = %.2f\n", Max(3.14, 2.71))

	// ============================================
	// CONSTRAINTS PERSONALIZADOS
	// ============================================
	fmt.Println("\n--- Constraints Personalizados ---")

	// Constraint que acepta solo números
	fmt.Printf("Sum(1, 2, 3, 4, 5) = %d\n", Sum(1, 2, 3, 4, 5))
	fmt.Printf("Sum(1.1, 2.2, 3.3) = %.1f\n", Sum(1.1, 2.2, 3.3))

	// Constraint con tipo subyacente (~)
	type MyInt int
	var a, b MyInt = 10, 20
	fmt.Printf("Sum con MyInt: %d\n", Sum(a, b))

	// ============================================
	// TIPOS GENÉRICOS
	// ============================================
	fmt.Println("\n--- Tipos Genéricos ---")

	// Stack genérico
	intStack := NewStack[int]()
	intStack.Push(1)
	intStack.Push(2)
	intStack.Push(3)
	fmt.Printf("Stack: %v\n", intStack.items)
	fmt.Printf("Pop: %d\n", intStack.Pop())
	fmt.Printf("Peek: %d\n", intStack.Peek())

	strStack := NewStack[string]()
	strStack.Push("a")
	strStack.Push("b")
	fmt.Printf("String Stack: %v\n", strStack.items)

	// Pair genérico
	pair := Pair[string, int]{First: "edad", Second: 30}
	fmt.Printf("Pair: %+v\n", pair)

	// ============================================
	// MÉTODOS GENÉRICOS
	// ============================================
	fmt.Println("\n--- Métodos Genéricos ---")

	// Los métodos pueden usar los type parameters del tipo
	list := List[int]{items: []int{3, 1, 4, 1, 5, 9}}
	fmt.Printf("Lista: %v\n", list.items)
	fmt.Printf("Contiene 4: %t\n", list.Contains(4))
	fmt.Printf("Contiene 7: %t\n", list.Contains(7))
	fmt.Printf("Índice de 5: %d\n", list.IndexOf(5))

	// Map y Filter
	doubled := list.Map(func(x int) int { return x * 2 })
	fmt.Printf("Duplicados: %v\n", doubled)

	evens := list.Filter(func(x int) bool { return x%2 == 0 })
	fmt.Printf("Pares: %v\n", evens)

	// ============================================
	// MÚLTIPLES TYPE PARAMETERS
	// ============================================
	fmt.Println("\n--- Múltiples Type Parameters ---")

	// Función con dos tipos
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := Keys(m)
	values := Values(m)
	fmt.Printf("Keys: %v\n", keys)
	fmt.Printf("Values: %v\n", values)

	// Invertir map
	inverted := Invert(m)
	fmt.Printf("Invertido: %v\n", inverted)

	// ============================================
	// TYPE INFERENCE
	// ============================================
	fmt.Println("\n--- Type Inference ---")

	// Go infiere los tipos cuando es posible
	nums := []int{5, 2, 8, 1, 9}
	fmt.Printf("Min de slice: %d\n", MinSlice(nums))

	strs := []string{"banana", "apple", "cherry"}
	fmt.Printf("Min de strings: %s\n", MinSlice(strs))

	// A veces necesitas especificar
	emptySlice := []int{}
	if val, ok := FirstOrDefault(emptySlice, 42); !ok {
		fmt.Printf("Slice vacío, default: %d\n", val)
	}

	// ============================================
	// SLICES PACKAGE (stdlib con generics)
	// ============================================
	fmt.Println("\n--- Package slices (Go 1.21+) ---")

	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}

	// Operaciones genéricas de slices
	sorted := slices.Clone(numbers)
	slices.Sort(sorted)
	fmt.Printf("Ordenado: %v\n", sorted)

	fmt.Printf("Contiene 5: %t\n", slices.Contains(numbers, 5))
	fmt.Printf("Índice de 9: %d\n", slices.Index(numbers, 9))
	fmt.Printf("Min: %d, Max: %d\n", slices.Min(numbers), slices.Max(numbers))

	// Equal
	a1 := []int{1, 2, 3}
	a2 := []int{1, 2, 3}
	fmt.Printf("Son iguales: %t\n", slices.Equal(a1, a2))

	// ============================================
	// MAPS PACKAGE (stdlib)
	// ============================================
	fmt.Println("\n--- Package maps (Go 1.21+) ---")

	// Nota: maps es un paquete experimental, mostramos patrones similares
	map1 := map[string]int{"a": 1, "b": 2}
	map2 := map[string]int{"a": 1, "b": 2}
	fmt.Printf("Maps iguales: %t\n", MapsEqual(map1, map2))

	// ============================================
	// CUÁNDO USAR GENERICS
	// ============================================
	fmt.Println("\n--- Cuándo Usar Generics ---")
	fmt.Println(`
USA GENERICS cuando:
✓ Escribes funciones que operan sobre slices/maps de cualquier tipo
✓ Implementas estructuras de datos (Stack, Queue, Tree, etc.)
✓ Necesitas type safety pero el algoritmo es el mismo para varios tipos
✓ Quieres evitar duplicación de código tipado

NO USES GENERICS cuando:
✗ Una interface específica es más clara
✗ Solo tienes uno o dos tipos específicos
✗ El código se vuelve más difícil de entender
✗ Puedes usar interface{}/any con type assertion segura

REGLA: "Don't design with generics, refactor to them"
- Escribe código concreto primero
- Refactoriza a generics cuando veas patrones repetidos`)
	// ============================================
	// EJEMPLO PRÁCTICO: CACHE GENÉRICO
	// ============================================
	fmt.Println("\n--- Ejemplo: Cache Genérico ---")

	userCache := NewCache[int, string]()
	userCache.Set(1, "Alice")
	userCache.Set(2, "Bob")

	if name, ok := userCache.Get(1); ok {
		fmt.Printf("Usuario 1: %s\n", name)
	}

	if _, ok := userCache.Get(3); !ok {
		fmt.Println("Usuario 3 no encontrado")
	}

	// ============================================
	// EJEMPLO: RESULT TYPE (como Rust)
	// ============================================
	fmt.Println("\n--- Ejemplo: Result Type ---")

	result3 := divide(10, 2)
	if result3.IsOk() {
		fmt.Printf("10 / 2 = %.2f\n", result3.Unwrap())
	}

	result4 := divide(10, 0)
	if result4.IsErr() {
		fmt.Printf("Error: %v\n", result4.UnwrapErr())
	}
}

// ============================================
// FUNCIONES SIN GENERICS (para comparación)
// ============================================

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MinFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func MinString(a, b string) string {
	if a < b {
		return a
	}
	return b
}

func MinAny(a, b any) any {
	// Perdemos type safety
	switch va := a.(type) {
	case int:
		if vb, ok := b.(int); ok && va < vb {
			return va
		}
		return b
	default:
		return a
	}
}

// ============================================
// FUNCIONES GENÉRICAS
// ============================================

// Función genérica con constraint Ordered
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Con any constraint
func PrintAny[T any](v T) {
	fmt.Printf("Valor: %v (tipo: %T)\n", v, v)
}

// Con comparable constraint
func Equal[T comparable](a, b T) bool {
	return a == b
}

// ============================================
// CONSTRAINTS PERSONALIZADOS
// ============================================

// Constraint para números
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func Sum[T Number](nums ...T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// Constraint que requiere métodos
type Stringer interface {
	String() string
}

func PrintStringer[T Stringer](v T) {
	fmt.Println(v.String())
}

// ============================================
// TIPOS GENÉRICOS
// ============================================

// Stack genérico
type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0)}
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		var zero T
		return zero
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item
}

func (s *Stack[T]) Peek() T {
	if len(s.items) == 0 {
		var zero T
		return zero
	}
	return s.items[len(s.items)-1]
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Pair genérico
type Pair[T, U any] struct {
	First  T
	Second U
}

// List con métodos genéricos
type List[T comparable] struct {
	items []T
}

func (l *List[T]) Contains(item T) bool {
	for _, v := range l.items {
		if v == item {
			return true
		}
	}
	return false
}

func (l *List[T]) IndexOf(item T) int {
	for i, v := range l.items {
		if v == item {
			return i
		}
	}
	return -1
}

func (l *List[T]) Map(f func(T) T) []T {
	result := make([]T, len(l.items))
	for i, v := range l.items {
		result[i] = f(v)
	}
	return result
}

func (l *List[T]) Filter(f func(T) bool) []T {
	var result []T
	for _, v := range l.items {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

// ============================================
// FUNCIONES CON MÚLTIPLES TYPE PARAMETERS
// ============================================

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func Invert[K, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// ============================================
// UTILIDADES
// ============================================

func MinSlice[T cmp.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}
	min := s[0]
	for _, v := range s[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func FirstOrDefault[T any](s []T, defaultVal T) (T, bool) {
	if len(s) == 0 {
		return defaultVal, false
	}
	return s[0], true
}

func MapsEqual[K, V comparable](m1, m2 map[K]V) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if v2, ok := m2[k]; !ok || v1 != v2 {
			return false
		}
	}
	return true
}

// ============================================
// CACHE GENÉRICO
// ============================================

type Cache[K comparable, V any] struct {
	data map[K]V
}

func NewCache[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{data: make(map[K]V)}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.data[key] = value
}

func (c *Cache[K, V]) Delete(key K) {
	delete(c.data, key)
}

// ============================================
// RESULT TYPE
// ============================================

type Result[T any] struct {
	value T
	err   error
}

func Ok[T any](value T) Result[T] {
	return Result[T]{value: value}
}

func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

func (r Result[T]) IsOk() bool {
	return r.err == nil
}

func (r Result[T]) IsErr() bool {
	return r.err != nil
}

func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(r.err)
	}
	return r.value
}

func (r Result[T]) UnwrapErr() error {
	return r.err
}

func divide(a, b float64) Result[float64] {
	if b == 0 {
		return Err[float64](fmt.Errorf("división por cero"))
	}
	return Ok(a / b)
}

/*
RESUMEN:

SINTAXIS:
func Nombre[T constraint](param T) T { }
type Tipo[T constraint] struct { campo T }

CONSTRAINTS COMUNES:
- any: cualquier tipo
- comparable: soporta == y !=
- cmp.Ordered: soporta < <= > >= (Go 1.21+)
- constraints.Integer/Float/etc. (x/exp/constraints)

CONSTRAINTS PERSONALIZADOS:
type MiConstraint interface {
    ~int | ~string | MiInterface
}

~ = incluye tipos con ese underlying type

BUENAS PRÁCTICAS:

1. Empieza con código concreto, refactoriza a generics cuando haya patrón
2. Usa constraints lo más específicos posible
3. Prefiere interfaces para comportamiento, generics para contenedores
4. No sobreuses generics - a veces el código específico es más claro
5. Aprovecha los paquetes slices y maps de la stdlib
*/

/*
SUMMARY - GENERICS:

FUNCIONES GENÉRICAS:
- func Min[T cmp.Ordered](a, b T) T
- El tipo se infiere automáticamente en la mayoría de casos
- Especificar explícitamente: Min[int64](10, 20)

CONSTRAINTS:
- any: cualquier tipo
- comparable: soporta == y !=
- cmp.Ordered: soporta <, <=, >, >= (Go 1.21+)
- Personalizados: type Number interface { ~int | ~float64 | ... }
- ~ incluye tipos con ese underlying type

TIPOS GENÉRICOS:
- type Stack[T any] struct { items []T }
- type Pair[T, U any] struct { First T; Second U }
- type Cache[K comparable, V any] struct { data map[K]V }

MÉTODOS GENÉRICOS:
- Los métodos usan los type parameters del tipo
- List[T].Contains(), Map(), Filter()

PAQUETES STDLIB:
- slices: Sort, Contains, Index, Min, Max, Clone, Equal
- maps: Keys, Values, Equal (experimental)

CUÁNDO USAR:
- Estructuras de datos genéricas (Stack, Queue, Tree)
- Funciones sobre slices/maps de cualquier tipo
- Evitar duplicación de código tipado
- NO usar cuando una interface específica es más clara

RESULT TYPE (inspirado en Rust):
- type Result[T any] struct { value T; err error }
- Métodos IsOk(), IsErr(), Unwrap(), UnwrapErr()
*/
