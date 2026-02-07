// Package main - Chapter 065: Go 1.24 Features
// Go 1.24 (Feb 2025) trajo mejoras significativas en rendimiento
// y nuevas características del lenguaje y biblioteca estándar.
package main

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"time"
)

func main() {
	fmt.Println("=== GO 1.24 FEATURES ===")
	fmt.Printf("Running on Go %s\n", runtime.Version())

	// ============================================
	// GENERIC TYPE ALIASES
	// ============================================
	fmt.Println("\n--- Generic Type Aliases ---")

	// Go 1.24 permite type aliases con parámetros de tipo
	// Antes solo podías hacer: type MySlice = []int
	// Ahora puedes hacer: type MySlice[T any] = []T

	type Pair[T any] = struct {
		First, Second T
	}

	intPair := Pair[int]{First: 1, Second: 2}
	strPair := Pair[string]{First: "hello", Second: "world"}

	fmt.Printf("Int pair: %+v\n", intPair)
	fmt.Printf("String pair: %+v\n", strPair)

	// Type alias genérico útil para simplificar tipos complejos
	type Result[T any] = struct {
		Value T
		Err   error
	}

	r := Result[string]{Value: "success", Err: nil}
	fmt.Printf("Result: %+v\n", r)

	// ============================================
	// SWISS TABLES MAP IMPLEMENTATION
	// ============================================
	fmt.Println("\n--- Swiss Tables (Map Performance) ---")
	fmt.Println(`
Go 1.24 reemplaza la implementación interna de maps con Swiss Tables:
- Hasta 60% más rápido en microbenchmarks
- ~1.5% mejora de CPU en benchmarks de aplicaciones reales
- Mejor uso de cache

No requiere cambios de código - es transparente.
Desactivar con: GOEXPERIMENT=noswissmap`)
	// Benchmark simple
	m := make(map[int]int)
	start := time.Now()
	for i := 0; i < 100000; i++ {
		m[i] = i * 2
	}
	for i := 0; i < 100000; i++ {
		_ = m[i]
	}
	fmt.Printf("Map operations (100k writes + 100k reads): %v\n", time.Since(start))

	// ============================================
	// TOOL DIRECTIVES IN GO.MOD
	// ============================================
	fmt.Println("\n--- Tool Directives in go.mod ---")
	fmt.Println(`
Antes (workaround con tools.go):
  //go:build tools
  package tools
  import _ "golang.org/x/tools/cmd/stringer"

Ahora (Go 1.24+):
  // go.mod
  module myproject
  go 1.24
  tool golang.org/x/tools/cmd/stringer

Comandos:
  go get -tool golang.org/x/tools/cmd/stringer
  go tool stringer -type=MyType
  go install tool  // instalar todos los tools`)
	// ============================================
	// CRYPTO IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Crypto Improvements ---")

	// crypto/rand.Text - generar texto aleatorio seguro
	text := rand.Text()
	fmt.Printf("Random text (22 chars): %s\n", text)

	// Nuevo: crypto/mlkem para post-quantum key exchange
	fmt.Println(`
Nuevos paquetes crypto:
- crypto/mlkem: ML-KEM-768 y ML-KEM-1024 (post-quantum, FIPS 203)
- crypto/hkdf: HMAC-based Key Derivation (RFC 5869)
- crypto/pbkdf2: Password-based key derivation (RFC 8018)
- crypto/sha3: SHA-3, SHAKE, cSHAKE (FIPS 202)

FIPS 140-3 compliance:
  GOFIPS140=v1.0.0 go build  // habilitar módulo FIPS`)
	// ============================================
	// TESTING IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Testing Improvements ---")
	fmt.Println(`
Nuevo método b.Loop() para benchmarks:

func BenchmarkFoo(b *testing.B) {
    // Antes:
    for i := 0; i < b.N; i++ {
        Foo()
    }

    // Ahora (Go 1.24+):
    for b.Loop() {
        Foo()
    }
}

Ventajas:
- Setup/cleanup se ejecutan solo una vez
- Previene optimización del compilador
- Código más limpio

Nuevos métodos:
- t.Context() / b.Context(): context cancelado después del test
- t.Chdir() / b.Chdir(): cambiar directorio temporalmente`)
	// ============================================
	// WEAK POINTERS
	// ============================================
	fmt.Println("\n--- Weak Pointers (package weak) ---")
	fmt.Println(`
Nuevo paquete 'weak' para punteros débiles:

import "weak"

ptr := weak.Make(&myObject)
// ptr no previene garbage collection

if obj := ptr.Value(); obj != nil {
    // objeto aún existe
    use(obj)
}

Casos de uso:
- Caches eficientes en memoria
- Canonicalización de valores
- Evitar memory leaks en estructuras cíclicas`)
	// ============================================
	// SYNCTEST (EXPERIMENTAL)
	// ============================================
	fmt.Println("\n--- synctest Package (Experimental) ---")
	fmt.Println(`
Nuevo paquete para testing de código concurrente:

import "testing/synctest"

func TestConcurrent(t *testing.T) {
    synctest.Run(func() {
        // time.Sleep usa clock virtual
        // time.After usa clock virtual
        // Todas las goroutines usan tiempo simulado

        go worker()
        time.Sleep(time.Hour)  // instantáneo
        synctest.Wait()        // esperar que goroutines bloqueen
    })
}

Muy útil para tests deterministas de código concurrente.`)
	// ============================================
	// OS.ROOT
	// ============================================
	fmt.Println("\n--- os.Root (Directory-Confined FS) ---")
	fmt.Println(`
Nuevo tipo os.Root para acceso a filesystem confinado:

root, err := os.OpenRoot("/var/data")
if err != nil {
    log.Fatal(err)
}
defer root.Close()

// Todas las operaciones están confinadas al directorio
file, err := root.Open("config.json")  // OK: /var/data/config.json
file, err := root.Open("../secret")    // ERROR: fuera del root

Previene path traversal attacks.`)
	// ============================================
	// JSON OMITZERO
	// ============================================
	fmt.Println("\n--- JSON omitzero Tag ---")

	type Config struct {
		Name     string    `json:"name"`
		Count    int       `json:"count,omitzero"`
		Enabled  bool      `json:"enabled,omitzero"`
		Created  time.Time `json:"created,omitzero"`
	}

	// omitzero vs omitempty:
	// omitempty: omite si es zero value O empty (para slices/maps/strings)
	// omitzero: omite solo si es zero value exacto

	config := Config{Name: "test"} // Count=0, Enabled=false, Created=zero
	// Con omitzero, estos campos se omiten del JSON

	fmt.Printf("Config: %+v\n", config)
	fmt.Println(`
Diferencia:
- omitempty: "", 0, false, nil, []
- omitzero: solo el zero value del tipo

Para time.Time:
- omitempty: no funciona bien (siempre incluye)
- omitzero: omite time.Time{} correctamente`)
	// ============================================
	// HTTP IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- HTTP/2 Improvements ---")
	fmt.Println(`
Soporte mejorado para HTTP/2:

// Nuevo: HTTP/2 no encriptado (h2c)
client := &http.Client{
    Transport: &http.Transport{
        // Habilitar h2c para desarrollo local
    },
}

// Nuevos campos de configuración:
server := &http.Server{
    HTTP2: &http.HTTP2Config{
        MaxConcurrentStreams: 100,
    },
}

// Encrypted Client Hello (ECH)
tlsConfig := &tls.Config{
    EncryptedClientHelloKeys: keys,
}`)
	// ============================================
	// RUNTIME IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Runtime Improvements ---")
	fmt.Println(`
Mejoras de rendimiento (2-3% CPU):

1. Swiss Tables para maps (ya mencionado)
2. Small object allocation más eficiente
3. Nuevo mutex interno del runtime
4. runtime.AddCleanup() - alternativa a SetFinalizer

// AddCleanup es más eficiente que SetFinalizer
runtime.AddCleanup(ptr, func(value int) {
    fmt.Println("Cleanup:", value)
}, 42)`)
	// ============================================
	// ITERATOR IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Iterator Functions ---")
	fmt.Println(`
Nuevas funciones de iteración en bytes/strings:

import "strings"

for line := range strings.Lines(text) {
    process(line)
}

for part := range strings.SplitSeq(text, ",") {
    process(part)
}

for field := range strings.FieldsSeq(text) {
    process(field)
}

También disponible en package bytes.`)
	// ============================================
	// PLATFORM CHANGES
	// ============================================
	fmt.Println("\n--- Platform Changes ---")
	fmt.Println(`
Linux:
- Requiere kernel 3.2 o posterior (antes 2.6.32)

macOS:
- Go 1.24 es la ÚLTIMA versión que soporta macOS 11 Big Sur
- Go 1.25+ requiere macOS 12 Monterey

WebAssembly:
- Nueva directiva go:wasmexport
- Soporte para reactor/library mode
- Reducción significativa de memoria inicial

Windows:
- 32-bit windows/arm marcado como broken`)}

/*
RESUMEN GO 1.24:

LENGUAJE:
- Generic type aliases completos

RENDIMIENTO:
- Swiss Tables para maps (~60% más rápido)
- Mejor allocation de objetos pequeños
- Nuevo mutex interno

HERRAMIENTAS:
- Tool directives en go.mod
- go run/tool caching mejorado
- Nuevo vet analyzer para tests

STDLIB:
- crypto/mlkem, crypto/hkdf, crypto/pbkdf2
- package weak para weak pointers
- testing/synctest para tests concurrentes
- os.Root para filesystem confinado
- JSON omitzero tag
- HTTP/2 improvements

TESTING:
- b.Loop() para benchmarks
- t.Context() / b.Context()
- t.Chdir() / b.Chdir()

BREAKING CHANGES:
- Linux kernel 3.2+ requerido
- Última versión para macOS 11
*/

/* SUMMARY - CHAPTER 065: Go 1.24 Features
Topics covered in this chapter:

• Generic type aliases: complete support without restrictions or edge cases
• Swiss Tables: new hash map implementation ~60% faster with better memory locality
• Tool directives: go.mod tool directives for specifying required tools
• Crypto improvements: crypto/mlkem (post-quantum), crypto/hkdf, crypto/pbkdf2
• Package weak: weak pointers for preventing memory leaks in caches
• testing/synctest: testing concurrent code with deterministic behavior
• os.Root: confined filesystem access for security
• JSON omitzero: new struct tag for omitting zero values
• Testing improvements: b.Loop() for benchmarks, t.Context()/b.Context(), t.Chdir()
• Performance: better small object allocation, new internal mutex
• Tool improvements: go run/tool caching, new vet analyzer for tests
• Breaking changes: Linux kernel 3.2+ required, last version for macOS 11
*/
