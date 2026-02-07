// Package main - Chapter 010: Collections
// Arrays, slices y maps son las estructuras de datos fundamentales en Go.
// Aprenderás sus diferencias, cuándo usar cada una, y operaciones comunes.
package main

import (
	"fmt"
	"slices"
	"sort"
)

func main() {
	fmt.Println("=== COLECCIONES EN GO ===")

	// ============================================
	// ARRAYS
	// ============================================
	fmt.Println("\n--- Arrays ---")

	// Arrays tienen tamaño FIJO, parte del tipo
	var arr1 [5]int                          // [0 0 0 0 0] - zero values
	arr2 := [5]int{1, 2, 3, 4, 5}             // inicializado
	arr3 := [...]int{1, 2, 3}                 // tamaño inferido (3)
	arr4 := [5]int{0: 10, 4: 40}              // índices específicos

	fmt.Printf("arr1: %v (len: %d)\n", arr1, len(arr1))
	fmt.Printf("arr2: %v\n", arr2)
	fmt.Printf("arr3: %v (len: %d)\n", arr3, len(arr3))
	fmt.Printf("arr4: %v\n", arr4)

	// Acceso y modificación
	arr2[0] = 100
	fmt.Printf("arr2[0] = 100: %v\n", arr2)

	// Arrays son VALORES, no referencias
	// Se copian al asignar o pasar a funciones
	copia := arr2
	copia[0] = 999
	fmt.Printf("Original: %v\n", arr2) // no cambió
	fmt.Printf("Copia: %v\n", copia)

	// [5]int y [3]int son TIPOS DIFERENTES
	// var x [5]int = arr3 // NO COMPILA

	// ============================================
	// SLICES
	// ============================================
	fmt.Println("\n--- Slices ---")

	// Slices son vistas flexibles sobre arrays
	// Tienen tamaño DINÁMICO

	// Formas de crear slices
	var s1 []int                    // nil slice (len=0, cap=0)
	s2 := []int{}                   // slice vacío (len=0, cap=0)
	s3 := []int{1, 2, 3, 4, 5}      // literal
	s4 := make([]int, 5)            // len=5, cap=5, zero values
	s5 := make([]int, 3, 10)        // len=3, cap=10

	fmt.Printf("s1 (nil): %v, len=%d, cap=%d, nil=%t\n", s1, len(s1), cap(s1), s1 == nil)
	fmt.Printf("s2 (vacío): %v, len=%d, cap=%d, nil=%t\n", s2, len(s2), cap(s2), s2 == nil)
	fmt.Printf("s3: %v, len=%d, cap=%d\n", s3, len(s3), cap(s3))
	fmt.Printf("s4: %v, len=%d, cap=%d\n", s4, len(s4), cap(s4))
	fmt.Printf("s5: %v, len=%d, cap=%d\n", s5, len(s5), cap(s5))

	// ============================================
	// SLICE OPERATIONS
	// ============================================
	fmt.Println("\n--- Operaciones de Slice ---")

	nums := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	// Slicing: [inicio:fin] (fin exclusivo)
	fmt.Printf("nums:      %v\n", nums)
	fmt.Printf("nums[2:5]: %v\n", nums[2:5])   // [2 3 4]
	fmt.Printf("nums[:5]:  %v\n", nums[:5])    // [0 1 2 3 4]
	fmt.Printf("nums[5:]:  %v\n", nums[5:])    // [5 6 7 8 9]
	fmt.Printf("nums[:]:   %v\n", nums[:])     // copia referencia

	// Slicing con capacidad: [inicio:fin:capacidad]
	limited := nums[2:5:7]  // cap = 7-2 = 5
	fmt.Printf("nums[2:5:7]: %v, len=%d, cap=%d\n", limited, len(limited), cap(limited))

	// ============================================
	// APPEND
	// ============================================
	fmt.Println("\n--- Append ---")

	slice := []int{1, 2, 3}
	fmt.Printf("Original: %v (cap=%d)\n", slice, cap(slice))

	// append puede realocar si no hay capacidad
	slice = append(slice, 4)
	fmt.Printf("Después de append(4): %v (cap=%d)\n", slice, cap(slice))

	slice = append(slice, 5, 6, 7)
	fmt.Printf("Después de append(5,6,7): %v (cap=%d)\n", slice, cap(slice))

	// Append otro slice (spread operator ...)
	otro := []int{8, 9, 10}
	slice = append(slice, otro...)
	fmt.Printf("Después de append(otro...): %v (cap=%d)\n", slice, cap(slice))

	// IMPORTANTE: siempre reasignar el resultado de append
	// slice = append(slice, x) // CORRECTO
	// append(slice, x)         // INCORRECTO - resultado perdido

	// ============================================
	// COPY
	// ============================================
	fmt.Println("\n--- Copy ---")

	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, 3)

	n := copy(dst, src)
	fmt.Printf("copy(dst, src): copió %d elementos\n", n)
	fmt.Printf("src: %v\n", src)
	fmt.Printf("dst: %v\n", dst) // [1 2 3]

	// Copia completa
	fullCopy := make([]int, len(src))
	copy(fullCopy, src)
	fmt.Printf("fullCopy: %v\n", fullCopy)

	// ============================================
	// SLICES COMPARTEN MEMORIA
	// ============================================
	fmt.Println("\n--- Slices Comparten Memoria ---")

	original := []int{1, 2, 3, 4, 5}
	vista := original[1:4]

	fmt.Printf("Original: %v\n", original)
	fmt.Printf("Vista [1:4]: %v\n", vista)

	// Modificar vista afecta original
	vista[0] = 999
	fmt.Printf("Después de vista[0]=999:\n")
	fmt.Printf("  Original: %v\n", original)
	fmt.Printf("  Vista: %v\n", vista)

	// Para una copia independiente:
	independiente := make([]int, len(vista))
	copy(independiente, vista)
	independiente[0] = 111
	fmt.Printf("Copia independiente: %v\n", independiente)
	fmt.Printf("Vista sin cambios: %v\n", vista)

	// ============================================
	// ELIMINAR ELEMENTOS
	// ============================================
	fmt.Println("\n--- Eliminar Elementos ---")

	// Eliminar elemento en índice i
	data := []int{0, 1, 2, 3, 4, 5}
	i := 2
	data = append(data[:i], data[i+1:]...)
	fmt.Printf("Después de eliminar índice 2: %v\n", data)

	// Eliminar sin preservar orden (más eficiente)
	data2 := []int{0, 1, 2, 3, 4, 5}
	i = 2
	data2[i] = data2[len(data2)-1]  // mover último al hueco
	data2 = data2[:len(data2)-1]     // truncar
	fmt.Printf("Eliminación rápida índice 2: %v\n", data2)

	// ============================================
	// SLICES PACKAGE (Go 1.21+)
	// ============================================
	fmt.Println("\n--- Package slices (Go 1.21+) ---")

	nums1 := []int{3, 1, 4, 1, 5, 9, 2, 6}

	// Sort
	slices.Sort(nums1)
	fmt.Printf("Sorted: %v\n", nums1)

	// Contains
	fmt.Printf("Contains(5): %t\n", slices.Contains(nums1, 5))
	fmt.Printf("Contains(7): %t\n", slices.Contains(nums1, 7))

	// Index
	fmt.Printf("Index(5): %d\n", slices.Index(nums1, 5))

	// Min, Max
	fmt.Printf("Min: %d, Max: %d\n", slices.Min(nums1), slices.Max(nums1))

	// Reverse
	slices.Reverse(nums1)
	fmt.Printf("Reversed: %v\n", nums1)

	// Clone (copia profunda para slices simples)
	clonado := slices.Clone(nums1)
	clonado[0] = 999
	fmt.Printf("Original: %v, Clonado: %v\n", nums1, clonado)

	// Equal
	a1 := []int{1, 2, 3}
	a2 := []int{1, 2, 3}
	fmt.Printf("Equal: %t\n", slices.Equal(a1, a2))

	// ============================================
	// MAPS
	// ============================================
	fmt.Println("\n--- Maps ---")

	// Formas de crear maps
	var m1 map[string]int                      // nil map (solo lectura)
	m2 := map[string]int{}                     // map vacío
	m3 := map[string]int{"uno": 1, "dos": 2}   // literal
	m4 := make(map[string]int)                 // make
	m5 := make(map[string]int, 100)            // con capacidad inicial

	fmt.Printf("m1 (nil): %v, nil=%t\n", m1, m1 == nil)
	fmt.Printf("m2 (vacío): %v\n", m2)
	fmt.Printf("m3: %v\n", m3)
	_ = m4
	_ = m5

	// IMPORTANTE: no puedes escribir en un nil map
	// m1["key"] = 1 // PANIC: assignment to entry in nil map

	// ============================================
	// OPERACIONES CON MAPS
	// ============================================
	fmt.Println("\n--- Operaciones con Maps ---")

	edades := map[string]int{
		"Alice": 30,
		"Bob":   25,
		"Carol": 35,
	}

	// Leer
	fmt.Printf("Alice tiene %d años\n", edades["Alice"])

	// Leer clave inexistente devuelve zero value
	fmt.Printf("David tiene %d años (zero value)\n", edades["David"])

	// Verificar existencia (comma ok idiom)
	if edad, existe := edades["David"]; existe {
		fmt.Printf("David tiene %d años\n", edad)
	} else {
		fmt.Println("David no existe en el mapa")
	}

	// Escribir
	edades["David"] = 28
	fmt.Printf("Después de agregar David: %v\n", edades)

	// Eliminar
	delete(edades, "Bob")
	fmt.Printf("Después de eliminar Bob: %v\n", edades)

	// Iterar (orden NO garantizado)
	fmt.Println("Iterando:")
	for nombre, edad := range edades {
		fmt.Printf("  %s: %d\n", nombre, edad)
	}

	// Tamaño
	fmt.Printf("Tamaño: %d\n", len(edades))

	// ============================================
	// MAPS CON STRUCTS
	// ============================================
	fmt.Println("\n--- Maps con Structs ---")

	type Usuario struct {
		Nombre string
		Email  string
	}

	usuarios := map[int]Usuario{
		1: {Nombre: "Alice", Email: "alice@example.com"},
		2: {Nombre: "Bob", Email: "bob@example.com"},
	}

	// No puedes modificar campos de struct directamente en map
	// usuarios[1].Nombre = "Alicia" // NO COMPILA

	// Solución: reasignar todo el struct
	u := usuarios[1]
	u.Nombre = "Alicia"
	usuarios[1] = u
	fmt.Printf("Usuario 1: %+v\n", usuarios[1])

	// O usar punteros
	usuariosPtr := map[int]*Usuario{
		1: {Nombre: "Alice", Email: "alice@example.com"},
	}
	usuariosPtr[1].Nombre = "Alicia Modificada"
	fmt.Printf("Usuario 1 (ptr): %+v\n", usuariosPtr[1])

	// ============================================
	// MAPS COMO SETS
	// ============================================
	fmt.Println("\n--- Maps como Sets ---")

	// Go no tiene tipo Set, usa map[T]bool o map[T]struct{}
	set := map[string]struct{}{
		"apple":  {},
		"banana": {},
		"cherry": {},
	}

	// Verificar existencia
	if _, existe := set["apple"]; existe {
		fmt.Println("apple está en el set")
	}

	// Agregar
	set["date"] = struct{}{}

	// Eliminar
	delete(set, "banana")

	fmt.Printf("Set: %v\n", set)

	// ============================================
	// ORDENAR MAPS
	// ============================================
	fmt.Println("\n--- Ordenar Maps ---")

	puntuaciones := map[string]int{
		"Charlie": 95,
		"Alice":   88,
		"Bob":     92,
	}

	// Extraer claves y ordenarlas
	claves := make([]string, 0, len(puntuaciones))
	for k := range puntuaciones {
		claves = append(claves, k)
	}
	sort.Strings(claves)

	fmt.Println("Ordenado por clave:")
	for _, k := range claves {
		fmt.Printf("  %s: %d\n", k, puntuaciones[k])
	}

	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrones Comunes ---")

	// 1. Contar ocurrencias
	palabras := []string{"go", "go", "java", "python", "go", "rust", "go"}
	conteo := make(map[string]int)
	for _, p := range palabras {
		conteo[p]++
	}
	fmt.Printf("Conteo: %v\n", conteo)

	// 2. Agrupar por clave
	personas := []struct {
		Nombre string
		Ciudad string
	}{
		{"Alice", "NYC"},
		{"Bob", "LA"},
		{"Carol", "NYC"},
		{"David", "LA"},
	}

	porCiudad := make(map[string][]string)
	for _, p := range personas {
		porCiudad[p.Ciudad] = append(porCiudad[p.Ciudad], p.Nombre)
	}
	fmt.Printf("Por ciudad: %v\n", porCiudad)

	// 3. Filtrar slice
	numeros := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	pares := []int{}
	for _, n := range numeros {
		if n%2 == 0 {
			pares = append(pares, n)
		}
	}
	fmt.Printf("Pares: %v\n", pares)
}

/*
RESUMEN:

ARRAYS:
- Tamaño fijo (parte del tipo)
- Se copian al asignar
- Uso: cuando conoces el tamaño exacto y necesitas tipo por valor

SLICES:
- Tamaño dinámico
- Referencia a array subyacente
- Comparten memoria cuando se hacen sub-slices
- Uso: la mayoría de los casos (99%)

MAPS:
- Key-value dinámico
- Orden no garantizado
- No thread-safe (usar sync.Map o mutex)
- Uso: búsquedas O(1), conteos, agrupaciones

BUENAS PRÁCTICAS:

1. Usa slices, no arrays (a menos que necesites el tamaño en el tipo)
2. Pre-aloca con make() si conoces el tamaño aproximado
3. Cuidado con slices que comparten memoria
4. Siempre verifica existencia en maps con comma-ok
5. Nunca escribas en nil maps (causa panic)
6. Usa struct{} como valor en sets (0 bytes)
*/

/*
SUMMARY - COLLECTIONS:

ARRAYS:
- Tamaño fijo, parte del tipo: [5]int != [3]int
- Se copian por valor al asignar o pasar a funciones
- Uso raro, preferir slices

SLICES:
- Vista dinámica sobre un array subyacente
- Creación: literal []int{1,2,3}, make([]int, len, cap)
- nil slice vs slice vacío (ambos len=0)
- Slicing: s[inicio:fin], s[inicio:fin:capacidad]
- append: puede realocar, siempre reasignar resultado
- copy: copia elementos entre slices
- Slices comparten memoria con sub-slices

PAQUETE slices (Go 1.21+):
- Sort, Contains, Index, Min, Max, Reverse
- Clone para copias independientes, Equal para comparación

MAPS:
- Key-value dinámico con búsqueda O(1)
- Creación: make(map[K]V) o literal
- Comma-ok idiom: if v, ok := m[key]; ok { }
- delete(m, key) para eliminar
- Orden de iteración NO garantizado
- Nunca escribir en nil map (panic)

PATRONES:
- Maps como sets con map[T]struct{} (0 bytes por valor)
- Contar ocurrencias con map[string]int
- Agrupar por clave con map[string][]T
- Ordenar maps: extraer claves, sort, iterar
*/
