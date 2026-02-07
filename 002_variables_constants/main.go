// Package main - Chapter 002: Variables and Constants
// Aprenderás todos los tipos de datos en Go, cómo declarar variables
// y constantes, y el concepto de zero values.
package main

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

// Variables a nivel de paquete (globales dentro del paquete)
// Deben usar var, no pueden usar :=
var (
	packageVar1 int
	packageVar2 string = "global"
	packageVar3        = 42 // tipo inferido
)

// Constantes a nivel de paquete
const (
	Pi      = 3.14159265359
	AppName = "GoLessons"
	Debug   = false
)

func main() {
	fmt.Println("=== TIPOS BÁSICOS EN GO ===")
	// ============================================
	// BOOLEAN
	// ============================================
	fmt.Println("--- Boolean ---")
	var b1 bool       // zero value: false
	var b2 bool = true
	b3 := false

	fmt.Printf("b1 (zero value): %t\n", b1)
	fmt.Printf("b2: %t, b3: %t\n", b2, b3)

	// ============================================
	// ENTEROS CON SIGNO
	// ============================================
	fmt.Println("\n--- Enteros con signo ---")

	var i int   // tamaño depende de la plataforma (32 o 64 bits)
	var i8 int8 // -128 a 127
	var i16 int16
	var i32 int32
	var i64 int64

	fmt.Printf("int:   %d bits, rango: %d a %d\n",
		unsafe.Sizeof(i)*8, math.MinInt, math.MaxInt)
	fmt.Printf("int8:  %d bits, rango: %d a %d\n",
		unsafe.Sizeof(i8)*8, math.MinInt8, math.MaxInt8)
	fmt.Printf("int16: %d bits, rango: %d a %d\n",
		unsafe.Sizeof(i16)*8, math.MinInt16, math.MaxInt16)
	fmt.Printf("int32: %d bits, rango: %d a %d\n",
		unsafe.Sizeof(i32)*8, math.MinInt32, math.MaxInt32)
	fmt.Printf("int64: %d bits, rango: %d a %d\n",
		unsafe.Sizeof(i64)*8, math.MinInt64, math.MaxInt64)

	// ============================================
	// ENTEROS SIN SIGNO
	// ============================================
	fmt.Println("\n--- Enteros sin signo ---")

	var u uint   // tamaño depende de la plataforma
	var u8 uint8 // 0 a 255 (alias: byte)
	var u16 uint16
	var u32 uint32
	var u64 uint64
	var uptr uintptr // para almacenar punteros

	fmt.Printf("uint:    %d bits, max: %d\n", unsafe.Sizeof(u)*8, uint(math.MaxUint))
	fmt.Printf("uint8:   %d bits, max: %d\n", unsafe.Sizeof(u8)*8, math.MaxUint8)
	fmt.Printf("uint16:  %d bits, max: %d\n", unsafe.Sizeof(u16)*8, math.MaxUint16)
	fmt.Printf("uint32:  %d bits, max: %d\n", unsafe.Sizeof(u32)*8, math.MaxUint32)
	fmt.Printf("uint64:  %d bits, max: %d\n", unsafe.Sizeof(u64)*8, uint64(math.MaxUint64))
	fmt.Printf("uintptr: %d bits\n", unsafe.Sizeof(uptr)*8)

	// ============================================
	// FLOTANTES
	// ============================================
	fmt.Println("\n--- Flotantes ---")

	var f32 float32
	var f64 float64

	fmt.Printf("float32: %d bits, max: %e\n", unsafe.Sizeof(f32)*8, math.MaxFloat32)
	fmt.Printf("float64: %d bits, max: %e\n", unsafe.Sizeof(f64)*8, math.MaxFloat64)

	// Precisión de flotantes
	fmt.Println("\nPrecisión:")
	var precio32 float32 = 19.99
	var precio64 float64 = 19.99
	fmt.Printf("float32: %.20f\n", precio32)
	fmt.Printf("float64: %.20f\n", precio64)

	// ============================================
	// COMPLEJOS
	// ============================================
	fmt.Println("\n--- Números Complejos ---")

	var c64 complex64   // float32 real + float32 imaginario
	var c128 complex128 // float64 real + float64 imaginario

	c64 = 3 + 4i
	c128 = complex(5, 12) // otra forma de crear

	fmt.Printf("c64:  %v (tipo: %T)\n", c64, c64)
	fmt.Printf("c128: %v (tipo: %T)\n", c128, c128)
	fmt.Printf("Real: %.2f, Imaginario: %.2f\n", real(c128), imag(c128))

	// ============================================
	// STRINGS
	// ============================================
	fmt.Println("\n--- Strings ---")

	var s1 string     // zero value: "" (string vacío)
	s2 := "Hola"
	s3 := `String literal
	multilínea
	sin escape: \n no es newline aquí`

	fmt.Printf("s1 (zero value): %q (len: %d)\n", s1, len(s1))
	fmt.Printf("s2: %s\n", s2)
	fmt.Printf("s3: %s\n", s3)

	// Strings son inmutables
	// s2[0] = 'h' // ERROR: cannot assign

	// ============================================
	// BYTE Y RUNE
	// ============================================
	fmt.Println("\n--- Byte y Rune ---")

	var byteVar byte = 'A'   // alias de uint8
	var runeVar rune = '世'   // alias de int32, para Unicode

	fmt.Printf("byte 'A': %d (char: %c)\n", byteVar, byteVar)
	fmt.Printf("rune '世': %d (char: %c)\n", runeVar, runeVar)

	// ============================================
	// ZERO VALUES
	// ============================================
	fmt.Println("\n=== ZERO VALUES ===")
	fmt.Println("Go inicializa todas las variables a su 'zero value':")

	var zBool bool
	var zInt int
	var zFloat float64
	var zString string
	var zPointer *int
	var zSlice []int
	var zMap map[string]int
	var zChan chan int
	var zFunc func()
	var zInterface interface{}

	fmt.Printf("bool:      %v\n", zBool)
	fmt.Printf("int:       %v\n", zInt)
	fmt.Printf("float64:   %v\n", zFloat)
	fmt.Printf("string:    %q\n", zString)
	fmt.Printf("pointer:   %v\n", zPointer)
	fmt.Printf("slice:     %v (nil: %t)\n", zSlice, zSlice == nil)
	fmt.Printf("map:       %v (nil: %t)\n", zMap, zMap == nil)
	fmt.Printf("channel:   %v (nil: %t)\n", zChan, zChan == nil)
	fmt.Printf("func:      %v (nil: %t)\n", zFunc, zFunc == nil)
	fmt.Printf("interface: %v (nil: %t)\n", zInterface, zInterface == nil)

	// ============================================
	// DECLARACIÓN DE VARIABLES
	// ============================================
	fmt.Println("\n=== FORMAS DE DECLARAR VARIABLES ===")

	// 1. var con tipo explícito
	var edad int = 30

	// 2. var con tipo inferido
	var nombre = "Gopher"

	// 3. var sin valor (usa zero value)
	var contador int

	// 4. Declaración corta := (solo dentro de funciones)
	salario := 50000.50

	// 5. Múltiples variables
	var a, b, c int = 1, 2, 3
	x, y := 10, "texto"

	// 6. Bloque var
	var (
		ciudad  = "Madrid"
		pais    = "España"
		codigo  int
	)

	fmt.Printf("edad: %d, nombre: %s, contador: %d\n", edad, nombre, contador)
	fmt.Printf("salario: %.2f\n", salario)
	fmt.Printf("a=%d, b=%d, c=%d\n", a, b, c)
	fmt.Printf("x=%d, y=%s\n", x, y)
	fmt.Printf("ciudad=%s, pais=%s, codigo=%d\n", ciudad, pais, codigo)

	// ============================================
	// CONSTANTES
	// ============================================
	fmt.Println("\n=== CONSTANTES ===")

	const pi = 3.14159
	const (
		StatusOK       = 200
		StatusNotFound = 404
		StatusError    = 500
	)

	// Constantes no tipadas (untyped constants)
	const untypedInt = 42      // puede usarse como int, int64, float64, etc.
	const untypedFloat = 3.14

	var i32Val int32 = untypedInt    // OK: constante se adapta
	var f64Val float64 = untypedInt  // OK: constante se adapta

	fmt.Printf("untypedInt como int32: %d\n", i32Val)
	fmt.Printf("untypedInt como float64: %.2f\n", f64Val)

	// ============================================
	// CONVERSIÓN DE TIPOS
	// ============================================
	fmt.Println("\n=== CONVERSIÓN DE TIPOS ===")
	fmt.Println("Go NO hace conversiones implícitas. Debes ser explícito.")

	var entero int = 42
	var flotante float64 = float64(entero) // conversión explícita
	var numFlotante float64 = 3.99
	var enteroDeFlotante int = int(numFlotante) // trunca, no redondea

	fmt.Printf("int a float64: %d -> %.2f\n", entero, flotante)
	fmt.Printf("float64 a int: 3.99 -> %d (trunca)\n", enteroDeFlotante)

	// String <-> []byte
	str := "Hello"
	bytes := []byte(str)
	strBack := string(bytes)
	fmt.Printf("string -> []byte -> string: %s -> %v -> %s\n", str, bytes, strBack)

	// int <-> string (NO es conversión directa)
	num := 65
	charFromNum := string(rune(num)) // convierte a caracter Unicode
	fmt.Printf("int 65 a string (como rune): %s\n", charFromNum)

	// Para convertir número a string como texto, usar strconv o fmt
	// numStr := strconv.Itoa(42) // "42"

	// ============================================
	// TYPE INFERENCE
	// ============================================
	fmt.Println("\n=== INFERENCIA DE TIPOS ===")

	inferInt := 42                    // int
	inferFloat := 3.14                // float64
	inferComplex := 1 + 2i            // complex128
	inferString := "hello"            // string
	inferBool := true                 // bool

	fmt.Printf("%v es %s\n", inferInt, reflect.TypeOf(inferInt))
	fmt.Printf("%v es %s\n", inferFloat, reflect.TypeOf(inferFloat))
	fmt.Printf("%v es %s\n", inferComplex, reflect.TypeOf(inferComplex))
	fmt.Printf("%v es %s\n", inferString, reflect.TypeOf(inferString))
	fmt.Printf("%v es %s\n", inferBool, reflect.TypeOf(inferBool))

	// ============================================
	// ANY (INTERFACE{})
	// ============================================
	fmt.Println("\n=== ANY (interface{}) ===")

	var cualquiera any = 42
	fmt.Printf("any con int: %v (tipo: %T)\n", cualquiera, cualquiera)

	cualquiera = "ahora soy string"
	fmt.Printf("any con string: %v (tipo: %T)\n", cualquiera, cualquiera)

	cualquiera = []int{1, 2, 3}
	fmt.Printf("any con slice: %v (tipo: %T)\n", cualquiera, cualquiera)

	// Type assertion
	if slice, ok := cualquiera.([]int); ok {
		fmt.Printf("Type assertion exitosa: %v\n", slice)
	}
}

/*
RESUMEN DE TIPOS EN GO:

Básicos:
- bool
- string
- int, int8, int16, int32, int64
- uint, uint8 (byte), uint16, uint32, uint64, uintptr
- float32, float64
- complex64, complex128
- rune (alias de int32, para Unicode)

Compuestos:
- array [n]T
- slice []T
- map map[K]V
- struct
- pointer *T
- function func(...)
- interface
- channel chan T

BUENAS PRÁCTICAS:

1. Usa int para enteros a menos que necesites un tamaño específico
2. Usa float64 para flotantes (más precisión, es el default)
3. Prefiere := para variables locales (más conciso)
4. Usa var para zero values o cuando el tipo no es obvio
5. Agrupa declaraciones relacionadas con var ()
6. Las constantes no tipadas son muy flexibles, úsalas
7. Siempre convierte tipos explícitamente
*/

/*
SUMMARY - VARIABLES AND CONSTANTS:

TIPOS BÁSICOS:
- bool, string, int/int8/int16/int32/int64
- uint/uint8(byte)/uint16/uint32/uint64/uintptr
- float32/float64, complex64/complex128, rune (int32)

ZERO VALUES:
- Todas las variables se inicializan automáticamente
- bool=false, int=0, float=0.0, string="", pointer/slice/map/chan/func=nil

DECLARACIÓN DE VARIABLES:
- var x int = 10 (explícito), var x = 10 (inferido)
- x := 10 (declaración corta, solo en funciones)
- var () para bloques de declaraciones relacionadas

CONSTANTES:
- const Pi = 3.14159 (no tipada, flexible)
- const blocks con tipo explícito para enums
- Las constantes no tipadas se adaptan al contexto

CONVERSIÓN DE TIPOS:
- Go NO hace conversiones implícitas
- Siempre explícitas: float64(entero), int(flotante)
- int a float trunca (no redondea)

TYPE INFERENCE:
- := infiere el tipo del valor asignado
- any (interface{}) puede contener cualquier valor
- Type assertion para extraer el tipo concreto
*/
