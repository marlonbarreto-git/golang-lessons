// Package main - Chapter 070: Unsafe
// El paquete unsafe permite operaciones de bajo nivel
// que evitan el sistema de tipos de Go. Usar con extremo cuidado.
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("=== UNSAFE EN GO ===")

	fmt.Println(`
⚠️  ADVERTENCIA: unsafe puede causar:
- Undefined behavior
- Memory corruption
- Security vulnerabilities
- Non-portable code

Solo usar cuando absolutamente necesario y entendiendo
completamente las implicaciones.`)
	// ============================================
	// SIZEOF, ALIGNOF, OFFSETOF
	// ============================================
	fmt.Println("\n--- Sizeof, Alignof, Offsetof ---")

	type Example struct {
		a bool
		b int64
		c bool
		d int32
	}

	fmt.Printf("Sizeof:\n")
	fmt.Printf("  bool:    %d bytes\n", unsafe.Sizeof(true))
	fmt.Printf("  int32:   %d bytes\n", unsafe.Sizeof(int32(0)))
	fmt.Printf("  int64:   %d bytes\n", unsafe.Sizeof(int64(0)))
	fmt.Printf("  string:  %d bytes\n", unsafe.Sizeof(""))
	fmt.Printf("  []int:   %d bytes\n", unsafe.Sizeof([]int{}))
	fmt.Printf("  Example: %d bytes\n", unsafe.Sizeof(Example{}))

	fmt.Printf("\nAlignof:\n")
	fmt.Printf("  bool:    %d\n", unsafe.Alignof(true))
	fmt.Printf("  int32:   %d\n", unsafe.Alignof(int32(0)))
	fmt.Printf("  int64:   %d\n", unsafe.Alignof(int64(0)))
	fmt.Printf("  string:  %d\n", unsafe.Alignof(""))

	var ex Example
	fmt.Printf("\nOffsetof in Example:\n")
	fmt.Printf("  a: %d\n", unsafe.Offsetof(ex.a))
	fmt.Printf("  b: %d\n", unsafe.Offsetof(ex.b))
	fmt.Printf("  c: %d\n", unsafe.Offsetof(ex.c))
	fmt.Printf("  d: %d\n", unsafe.Offsetof(ex.d))

	// ============================================
	// PADDING Y ALINEAMIENTO
	// ============================================
	fmt.Println("\n--- Padding y Alineamiento ---")

	// Struct con padding ineficiente
	type BadLayout struct {
		a bool   // 1 byte + 7 padding
		b int64  // 8 bytes
		c bool   // 1 byte + 3 padding
		d int32  // 4 bytes
	} // Total: 24 bytes

	// Struct optimizado
	type GoodLayout struct {
		b int64  // 8 bytes
		d int32  // 4 bytes
		a bool   // 1 byte
		c bool   // 1 byte + 2 padding
	} // Total: 16 bytes

	fmt.Printf("BadLayout size:  %d bytes\n", unsafe.Sizeof(BadLayout{}))
	fmt.Printf("GoodLayout size: %d bytes\n", unsafe.Sizeof(GoodLayout{}))
	fmt.Println("Ordenar campos de mayor a menor reduce padding")

	// ============================================
	// UNSAFE.POINTER
	// ============================================
	fmt.Println("\n--- unsafe.Pointer ---")

	// Conversión entre tipos de puntero
	var x int64 = 42
	ptr := unsafe.Pointer(&x)
	xPtr := (*int64)(ptr)
	fmt.Printf("Original: %d, Via unsafe.Pointer: %d\n", x, *xPtr)

	// Conversión a uintptr para aritmética
	addr := uintptr(ptr)
	fmt.Printf("Address: %#x\n", addr)

	// ============================================
	// ACCESO A CAMPOS PRIVADOS (NO RECOMENDADO)
	// ============================================
	fmt.Println("\n--- Acceso a Campos Privados ---")

	type private struct {
		public  int
		private int // Normalmente inaccesible desde fuera
	}

	p := private{public: 1, private: 2}
	fmt.Printf("Original: %+v\n", p)

	// Acceder al campo privado vía unsafe
	privatePtr := (*int)(unsafe.Pointer(
		uintptr(unsafe.Pointer(&p)) + unsafe.Offsetof(p.private),
	))
	*privatePtr = 99
	fmt.Printf("Modified: %+v\n", p)
	fmt.Println("⚠️  Esto rompe encapsulación - evitar en producción")

	// ============================================
	// STRING <-> []BYTE SIN COPIA
	// ============================================
	fmt.Println("\n--- String <-> []byte Sin Copia ---")

	// ADVERTENCIA: El resultado es read-only si el original es un string literal
	s := "Hello, World!"
	b := stringToBytes(s)
	fmt.Printf("String: %s\n", s)
	fmt.Printf("Bytes: %v\n", b)

	// []byte -> string
	bytes := []byte{72, 101, 108, 108, 111}
	str := bytesToString(bytes)
	fmt.Printf("Bytes to String: %s\n", str)

	fmt.Println("⚠️  Cuidado: modificar bytes modifica el string original")

	// ============================================
	// SLICE HEADER MANIPULATION
	// ============================================
	fmt.Println("\n--- Slice Header ---")

	// Un slice es un struct interno:
	// type SliceHeader struct {
	//     Data uintptr
	//     Len  int
	//     Cap  int
	// }

	slice := []int{1, 2, 3, 4, 5}
	slicePtr := unsafe.Pointer(&slice)

	// Leer el header (solo lectura, no modificar así)
	data := *(*uintptr)(slicePtr)
	length := *(*int)(unsafe.Pointer(uintptr(slicePtr) + unsafe.Sizeof(uintptr(0))))
	capacity := *(*int)(unsafe.Pointer(uintptr(slicePtr) + 2*unsafe.Sizeof(uintptr(0))))

	fmt.Printf("Slice: %v\n", slice)
	fmt.Printf("Data ptr: %#x\n", data)
	fmt.Printf("Length: %d\n", length)
	fmt.Printf("Capacity: %d\n", capacity)

	// ============================================
	// UNSAFE.ADD Y UNSAFE.SLICE (Go 1.17+)
	// ============================================
	fmt.Println("\n--- unsafe.Add y unsafe.Slice (Go 1.17+) ---")

	arr := [5]int{10, 20, 30, 40, 50}
	ptr2 := unsafe.Pointer(&arr[0])

	// unsafe.Add: aritmética de punteros más segura
	ptr3 := unsafe.Add(ptr2, 2*unsafe.Sizeof(arr[0]))
	value := *(*int)(ptr3)
	fmt.Printf("arr[2] via unsafe.Add: %d\n", value)

	// unsafe.Slice: crear slice desde puntero
	sliceFromArr := unsafe.Slice(&arr[0], 3)
	fmt.Printf("Slice from array: %v\n", sliceFromArr)

	// ============================================
	// UNSAFE.STRING Y UNSAFE.STRINGDATA (Go 1.20+)
	// ============================================
	fmt.Println("\n--- unsafe.String/StringData (Go 1.20+) ---")

	// Crear string desde puntero y longitud
	data2 := []byte{'H', 'e', 'l', 'l', 'o'}
	str2 := unsafe.String(&data2[0], len(data2))
	fmt.Printf("String from bytes: %s\n", str2)

	// Obtener puntero a datos del string
	str3 := "World"
	dataPtr := unsafe.StringData(str3)
	fmt.Printf("First byte of 'World': %c\n", *dataPtr)

	// ============================================
	// REGLAS DE USO SEGURO
	// ============================================
	fmt.Println("\n--- Reglas de Uso Seguro ---")
	fmt.Println(`
PATRONES VÁLIDOS:

1. Conversión *T1 -> unsafe.Pointer -> *T2
   ptr := unsafe.Pointer(&x)
   y := (*T2)(ptr)

2. Conversión a uintptr para aritmética
   addr := uintptr(unsafe.Pointer(&x))
   // Operación atómica, no guardar en variable

3. unsafe.Add para aritmética de punteros
   ptr2 := unsafe.Add(ptr, offset)

4. unsafe.Slice para crear slices
   slice := unsafe.Slice(ptr, length)

PATRONES PELIGROSOS:

✗ Guardar uintptr en variable (puede ser invalidado por GC)
   addr := uintptr(ptr)  // Peligroso si GC mueve el objeto
   // ... tiempo pasa ...
   ptr2 := (*T)(unsafe.Pointer(addr))  // Puede fallar

✗ Modificar string literals
   s := "hello"
   b := *(*[]byte)(unsafe.Pointer(&s))
   b[0] = 'H'  // UNDEFINED BEHAVIOR

✗ Crear punteros a memoria no inicializada
   var x int
   ptr := (*[1000]int)(unsafe.Pointer(&x))  // Peligroso`)
	// ============================================
	// CUÁNDO USAR UNSAFE
	// ============================================
	fmt.Println("\n--- Cuándo Usar Unsafe ---")
	fmt.Println(`
CASOS LEGÍTIMOS:

1. Interoperabilidad con C (cgo)
2. Optimizaciones críticas de rendimiento
3. Serialización de bajo nivel
4. Implementación de primitivas del runtime

ALTERNATIVAS PREFERIDAS:

1. encoding/binary para serialización
2. reflect para inspección de tipos
3. Generics para código genérico
4. Copiar datos en vez de reinterpretar

CHECKLIST ANTES DE USAR:

☐ ¿Es realmente necesario?
☐ ¿Hay alternativa segura?
☐ ¿Está bien documentado?
☐ ¿Hay tests exhaustivos?
☐ ¿Entiendo las implicaciones de GC?
☐ ¿Es portable entre plataformas?`)}

// Conversión string -> []byte sin copia
// ADVERTENCIA: No modificar el resultado si el string es literal
func stringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Conversión []byte -> string sin copia
// ADVERTENCIA: El string se invalida si los bytes se modifican
func bytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

/*
RESUMEN DE UNSAFE:

FUNCIONES:
unsafe.Sizeof(x)    - Tamaño en bytes
unsafe.Alignof(x)   - Alineamiento requerido
unsafe.Offsetof(f)  - Offset de campo en struct

TIPOS:
unsafe.Pointer      - Puntero genérico
uintptr             - Entero del tamaño de un puntero

Go 1.17+:
unsafe.Add(ptr, len) - Aritmética de punteros
unsafe.Slice(ptr, len) - Crear slice desde puntero

Go 1.20+:
unsafe.String(ptr, len) - Crear string desde bytes
unsafe.StringData(s)    - Obtener puntero a datos de string
unsafe.SliceData(s)     - Obtener puntero a datos de slice

OPTIMIZACIÓN DE STRUCTS:
- Ordenar campos de mayor a menor tamaño
- Reduce padding y tamaño total

REGLAS:
1. Nunca guardar uintptr en variable
2. Operaciones atómicas pointer -> uintptr -> pointer
3. No modificar strings via []byte
4. Documentar todo uso de unsafe
5. Tests exhaustivos

⚠️  USAR SOLO CUANDO ABSOLUTAMENTE NECESARIO
*/
