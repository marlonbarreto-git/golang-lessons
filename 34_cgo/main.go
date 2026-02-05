// Package main - Capítulo 34: CGO
// CGO permite llamar código C desde Go y viceversa.
// Útil para usar bibliotecas C existentes.
package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Función C simple
int add(int a, int b) {
    return a + b;
}

// Función que modifica un buffer
void fill_buffer(char* buf, int size) {
    for (int i = 0; i < size - 1; i++) {
        buf[i] = 'A' + (i % 26);
    }
    buf[size - 1] = '\0';
}

// Struct C
typedef struct {
    int id;
    char name[50];
    double value;
} CItem;

// Función que recibe struct
void print_item(CItem* item) {
    printf("C: Item id=%d, name=%s, value=%.2f\n",
           item->id, item->name, item->value);
}
*/
import "C"

import (
	"os"
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("=== CGO EN GO ===")

	// ============================================
	// LLAMAR FUNCIONES C
	// ============================================
	fmt.Println("\n--- Llamar Funciones C ---")

	// Llamar función C simple
	result := C.add(10, 20)
	fmt.Printf("C.add(10, 20) = %d\n", result)

	// ============================================
	// TIPOS C
	// ============================================
	fmt.Println("\n--- Tipos C ---")
	fmt.Println(`
Tipos C en Go:
  C.char      -> int8
  C.schar     -> int8
  C.uchar     -> uint8
  C.short     -> int16
  C.ushort    -> uint16
  C.int       -> int32
  C.uint      -> uint32
  C.long      -> int32 o int64
  C.ulong     -> uint32 o uint64
  C.longlong  -> int64
  C.ulonglong -> uint64
  C.float     -> float32
  C.double    -> float64
  C.size_t    -> uintptr`)
	// ============================================
	// STRINGS
	// ============================================
	fmt.Println("\n--- Strings ---")

	// Go string -> C string
	goStr := "Hello from Go!"
	cStr := C.CString(goStr)
	defer C.free(unsafe.Pointer(cStr)) // IMPORTANTE: liberar memoria

	fmt.Printf("Go string: %s\n", goStr)

	// C string -> Go string
	goStrBack := C.GoString(cStr)
	fmt.Printf("Back to Go: %s\n", goStrBack)

	// ============================================
	// ARRAYS Y SLICES
	// ============================================
	fmt.Println("\n--- Arrays y Slices ---")

	// Crear buffer C y llenarlo
	bufSize := 27
	buf := (*C.char)(C.malloc(C.size_t(bufSize)))
	defer C.free(unsafe.Pointer(buf))

	C.fill_buffer(buf, C.int(bufSize))

	// Convertir a Go string
	goBuffer := C.GoString(buf)
	fmt.Printf("Buffer filled by C: %s\n", goBuffer)

	// Go slice -> C array
	goSlice := []C.int{1, 2, 3, 4, 5}
	cArray := &goSlice[0] // Puntero al primer elemento
	_ = cArray

	// ============================================
	// STRUCTS
	// ============================================
	fmt.Println("\n--- Structs ---")

	// Crear y usar struct C
	item := C.CItem{
		id:    42,
		value: 99.99,
	}

	// Copiar string a char array
	name := "TestItem"
	for i, c := range name {
		item.name[i] = C.char(c)
	}
	item.name[len(name)] = 0 // Null terminator

	C.print_item(&item)

	// ============================================
	// DOCUMENTACIÓN CGO
	// ============================================
	fmt.Println("\n--- Documentación CGO ---")
	fmt.Println(`
ESTRUCTURA DE CÓDIGO CGO:

// El comentario especial DEBE estar JUSTO ANTES de "import C"
/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lmylib
#cgo pkg-config: libpng

#include <stdlib.h>
#include "myheader.h"

// Código C inline
int myfunction(int x) {
    return x * 2;
}
*/
import "C"

DIRECTIVAS #cgo:

#cgo CFLAGS: -I<path>       Flags del compilador C
#cgo LDFLAGS: -L<path> -l<lib>  Flags del linker
#cgo pkg-config: <pkg>      Usar pkg-config
#cgo CPPFLAGS: <flags>      Flags del preprocesador
#cgo CXXFLAGS: <flags>      Flags del compilador C++

CONDICIONES DE PLATAFORMA:

#cgo linux LDFLAGS: -lm
#cgo darwin LDFLAGS: -framework CoreFoundation
#cgo windows LDFLAGS: -lws2_32
#cgo !windows LDFLAGS: -lpthread`)
	// ============================================
	// FUNCIONES HELPER
	// ============================================
	fmt.Println("\n--- Funciones Helper ---")
	fmt.Println(`
CONVERSIÓN DE TIPOS:

// Go string -> C string (REQUIERE free!)
cStr := C.CString(goStr)
defer C.free(unsafe.Pointer(cStr))

// C string -> Go string
goStr := C.GoString(cStr)

// C string con longitud -> Go string
goStr := C.GoStringN(cStr, length)

// Go []byte -> C array
cBytes := C.CBytes(goBytes)
defer C.free(cBytes)

// C array -> Go []byte
goBytes := C.GoBytes(unsafe.Pointer(cPtr), length)`)
	// ============================================
	// CALLBACKS
	// ============================================
	fmt.Println("\n--- Callbacks ---")
	os.Stdout.WriteString(`
EXPORTAR FUNCIÓN GO PARA C:

//export MyGoCallback
func MyGoCallback(x C.int) C.int {
    return x * 2
}

USO EN C:
extern int MyGoCallback(int);

void useCallback() {
    int result = MyGoCallback(21);
    printf("Result: %d\n", result);
}
`)

	// ============================================
	// BUENAS PRÁCTICAS
	// ============================================
	fmt.Println("\n--- Buenas Prácticas ---")
	fmt.Println(`
1. SIEMPRE liberar memoria C:
   cStr := C.CString(s)
   defer C.free(unsafe.Pointer(cStr))

2. Manejar errores C:
   // errno después de llamadas C
   _, err := C.someFunction()
   if err != nil {
       // error de errno
   }

3. Evitar pasar punteros Go a C si:
   - Apuntan a memoria Go con punteros
   - La función C guarda el puntero
   // go build -gcflags="-d=checkptr" para verificar

4. Minimizar llamadas CGO:
   - Cada llamada tiene overhead (~100ns)
   - Batching si es posible

5. Testing:
   CGO_ENABLED=1 go test

6. Cross-compilation es compleja:
   - Necesita compilador C para target
   - Considerar bibliotecas Go puras`)
	// ============================================
	// ALTERNATIVAS A CGO
	// ============================================
	fmt.Println("\n--- Alternativas a CGO ---")
	fmt.Println(`
CUÁNDO EVITAR CGO:

1. Cross-compilation requerida
2. Deployment en containers mínimos
3. Alto rendimiento (overhead de llamadas)
4. Complejidad de build no justificada

ALTERNATIVAS:

1. Bibliotecas Go puras
   - image en lugar de libpng
   - encoding/json en lugar de jansson

2. FFI sin CGO
   - purego (github.com/ebitengine/purego)
   - Carga dinámica de .so/.dll

3. IPC/RPC
   - Proceso separado en C
   - Comunicación via socket/pipe

4. WASM
   - Compilar C a WASM
   - Ejecutar en Go via wazero

5. Reescribir en Go
   - Para código C pequeño/mediano`)
	// ============================================
	// BUILD Y DEPLOY
	// ============================================
	fmt.Println("\n--- Build y Deploy ---")
	fmt.Println(`
BUILD:

# Habilitar CGO (default si hay código C)
CGO_ENABLED=1 go build

# Deshabilitar CGO
CGO_ENABLED=0 go build

# Especificar compilador
CC=gcc go build
CC=clang go build

# Cross-compile (necesita toolchain C para target)
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
  CC=x86_64-linux-gnu-gcc go build

STATIC LINKING:

# Forzar linking estático (Linux)
CGO_LDFLAGS="-static" go build

# Con musl-libc
CC=musl-gcc go build

DOCKER:

# Imagen con toolchain C
FROM golang:1.26 AS builder
RUN apt-get update && apt-get install -y gcc
COPY . .
RUN CGO_ENABLED=1 go build -o app

# Imagen final necesita libc
FROM debian:bookworm-slim
COPY --from=builder /app/app /app
CMD ["/app"]`)}

/*
RESUMEN DE CGO:

SINTAXIS:
// Comentario especial seguido de import "C"
// #include <stdlib.h>
// import "C"

DIRECTIVAS:
// #cgo CFLAGS: -I<path>
// #cgo LDFLAGS: -L<path> -l<lib>
// #cgo pkg-config: <pkg>

TIPOS:
C.int, C.char, C.double, etc.
C.struct_Name, C.enum_Name, C.union_Name

CONVERSIONES:
C.CString(s)      - Go string -> C char* (requiere free)
C.GoString(cStr)  - C char* -> Go string
C.CBytes(b)       - Go []byte -> C void* (requiere free)
C.GoBytes(p, n)   - C void* -> Go []byte

MEMORIA:
C.malloc(size)
C.free(ptr)
SIEMPRE liberar memoria C con defer

EXPORTAR:
//export FunctionName
func FunctionName(x C.int) C.int { ... }

BUILD:
CGO_ENABLED=1 go build
CC=gcc go build

BUENAS PRÁCTICAS:
1. Minimizar llamadas CGO
2. Siempre liberar memoria C
3. Documentar dependencias C
4. Considerar alternativas puras Go
*/
