// Package main - Chapter 064: Build Tags
// Go permite compilar código condicionalmente usando build tags.
// Esencial para cross-platform, features opcionales, y configuraciones.
package main

import (
	"fmt"
	"runtime"
)

func main() {
	fmt.Println("=== BUILD TAGS Y CONDITIONAL COMPILATION ===")

	// ============================================
	// SINTAXIS BÁSICA
	// ============================================
	fmt.Println("\n--- Sintaxis Básica ---")
	fmt.Println(`
SINTAXIS MODERNA (Go 1.17+):
// go:build expression

SINTAXIS LEGACY (aún soportada):
// +build expression

EJEMPLOS:
// go:build linux
// go:build windows && amd64
// go:build !cgo
// go:build debug || test

UBICACIÓN:
- Primera línea del archivo (antes de package)
- Línea en blanco después del build tag`)
	// ============================================
	// OPERADORES
	// ============================================
	fmt.Println("\n--- Operadores ---")
	fmt.Println(`
OPERADORES LÓGICOS:

&&    AND    // go:build linux && amd64
||    OR     // go:build linux || darwin
!     NOT    // go:build !windows
()    GROUP  // go:build (linux && amd64) || (darwin && arm64)

EJEMPLOS COMUNES:

// Solo Linux o macOS
// go:build linux || darwin

// Solo 64-bit
// go:build amd64 || arm64

// No Windows
// go:build !windows

// Desarrollo solamente
// go:build dev

// Producción (no dev ni test)
// go:build !dev && !test`)
	// ============================================
	// TAGS PREDEFINIDOS
	// ============================================
	fmt.Println("\n--- Tags Predefinidos ---")
	fmt.Println(`
SISTEMA OPERATIVO (GOOS):
linux, darwin, windows, freebsd, netbsd, openbsd
android, ios, js, wasip1, plan9, aix, dragonfly
illumos, solaris, zos

ARQUITECTURA (GOARCH):
amd64, arm64, arm, 386, ppc64, ppc64le
mips, mipsle, mips64, mips64le
riscv64, s390x, wasm, loong64

VERSIÓN DE GO:
go1.21, go1.22, go1.23, go1.24, go1.25, go1.26

OTROS:
cgo        - CGO habilitado
race       - Race detector habilitado
msan       - Memory sanitizer habilitado
asan       - Address sanitizer habilitado
purego     - Sin assembly (pure Go)
unix       - Sistemas Unix-like (linux, darwin, freebsd, etc.)`)
	// ============================================
	// ARCHIVOS POR PLATAFORMA
	// ============================================
	fmt.Println("\n--- Archivos por Plataforma ---")
	fmt.Println(`
CONVENCIÓN DE NOMBRES:

file_GOOS.go           // file_linux.go, file_windows.go
file_GOARCH.go         // file_amd64.go, file_arm64.go
file_GOOS_GOARCH.go    // file_linux_amd64.go

EJEMPLOS:

network_linux.go       // Solo compila en Linux
network_windows.go     // Solo compila en Windows
crypto_amd64.go        // Solo compila en amd64
simd_linux_amd64.go    // Solo Linux + amd64

NO NECESITAN // go:build:
- El nombre del archivo implica el tag
- file_linux.go es equivalente a // go:build linux`)
	// ============================================
	// CUSTOM TAGS
	// ============================================
	fmt.Println("\n--- Custom Tags ---")
	fmt.Println(`
CREAR TUS PROPIOS TAGS:

// debug.go
// go:build debug

package main

func debugLog(msg string) {
    fmt.Println("[DEBUG]", msg)
}

// release.go
// go:build !debug

package main

func debugLog(msg string) {
    // No-op en release
}

COMPILAR CON TAGS:
go build -tags debug          # Con debug
go build -tags "debug trace"  # Múltiples tags
go build                      # Sin tags (release)

MÚLTIPLES TAGS:
go build -tags "integration postgres"
go test -tags "e2e slow"`)
	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrones Comunes ---")
	fmt.Println(`
1. FEATURES OPCIONALES:

// feature_enabled.go
// go:build feature_x

func FeatureX() { /* implementación */ }

// feature_disabled.go
// go:build !feature_x

func FeatureX() { /* no-op o error */ }

2. CONFIGURACIÓN POR AMBIENTE:

// config_dev.go
// go:build dev

var Config = DevConfig{}

// config_prod.go
// go:build !dev

var Config = ProdConfig{}

3. DRIVERS DE BASE DE DATOS:

// db_postgres.go
// go:build postgres

import _ "github.com/lib/pq"

// db_mysql.go
// go:build mysql

import _ "github.com/go-sql-driver/mysql"

4. IMPLEMENTACIONES OPTIMIZADAS:

// hash_generic.go
// go:build !amd64

func Hash(data []byte) uint64 {
    // Implementación genérica
}

// hash_amd64.go
// go:build amd64

func Hash(data []byte) uint64 {
    // Implementación con SIMD
}`)
	// ============================================
	// TESTING CON TAGS
	// ============================================
	fmt.Println("\n--- Testing con Tags ---")
	fmt.Println(`
TESTS CONDICIONALES:

// integration_test.go
// go:build integration

func TestDatabaseConnection(t *testing.T) {
    // Test que necesita DB real
}

// slow_test.go
// go:build slow

func TestHeavyComputation(t *testing.T) {
    // Test que toma mucho tiempo
}

EJECUTAR:
go test                           # Solo unit tests
go test -tags integration         # + integration tests
go test -tags "integration slow"  # + slow tests
go test -tags all                 # Todos

EXCLUIR:
// unit_test.go
// go:build !integration

func TestQuickUnit(t *testing.T) {
    // Solo en unit tests
}`)
	// ============================================
	// CGO CONDITIONAL
	// ============================================
	fmt.Println("\n--- CGO Conditional ---")
	fmt.Println(`
CÓDIGO CGO OPCIONAL:

// sqlite_cgo.go
// go:build cgo

import "database/sql"
import _ "github.com/mattn/go-sqlite3"

func OpenDB(path string) (*sql.DB, error) {
    return sql.Open("sqlite3", path)
}

// sqlite_nocgo.go
// go:build !cgo

import "modernc.org/sqlite"

func OpenDB(path string) (*sql.DB, error) {
    return sql.Open("sqlite", path)
}

BUILD:
CGO_ENABLED=1 go build    # Usa versión CGO
CGO_ENABLED=0 go build    # Usa versión pure Go`)
	// ============================================
	// GO VERSION TAGS
	// ============================================
	fmt.Println("\n--- Go Version Tags ---")
	fmt.Println(`
CÓDIGO PARA VERSIONES ESPECÍFICAS:

// iter_new.go
// go:build go1.23

import "iter"

func MyIterator() iter.Seq[int] { ... }

// iter_old.go
// go:build !go1.23

func MyIterator() func(func(int) bool) { ... }

FEATURE DETECTION:

// maps_new.go
// go:build go1.21

import "maps"

func CloneMap[K comparable, V any](m map[K]V) map[K]V {
    return maps.Clone(m)
}

// maps_old.go
// go:build !go1.21

func CloneMap[K comparable, V any](m map[K]V) map[K]V {
    result := make(map[K]V, len(m))
    for k, v := range m {
        result[k] = v
    }
    return result
}`)
	// ============================================
	// IGNORE FILES
	// ============================================
	fmt.Println("\n--- Ignorar Archivos ---")
	fmt.Println(`
EXCLUIR COMPLETAMENTE:

// Este archivo nunca se compila
// go:build ignore

package main

// Código de ejemplo, documentación, etc.

TAMBIÉN FUNCIONAN:
- Archivos que empiezan con _ : _ignored.go
- Archivos que empiezan con . : .hidden.go
- Directorio testdata/

ÚTIL PARA:
- Ejemplos en documentación
- Código deprecated
- Generadores de código
- Scripts de build`)
	// ============================================
	// MIGRACIÓN LEGACY → MODERNO
	// ============================================
	fmt.Println("\n--- Migración de Sintaxis ---")
	fmt.Println(`
LEGACY (// +build):
// +build linux,amd64 darwin,amd64
// +build !windows

MODERNO (// go:build):
// go:build (linux && amd64) || (darwin && amd64)
// go:build !windows

CONVERSIÓN:
// +build linux darwin     → // go:build linux || darwin
// +build amd64            → // go:build amd64
// +build linux,amd64      → // go:build linux && amd64
// +build !cgo             → // go:build !cgo

HERRAMIENTA:
gofmt -w file.go    # Actualiza automáticamente a // go:build

AMBOS EN ARCHIVO (compatibilidad):
// go:build linux
// +build linux

package main`)
	// ============================================
	// VERIFICAR BUILD TAGS
	// ============================================
	fmt.Println("\n--- Verificar Build Tags ---")
	fmt.Println(`
LISTAR ARCHIVOS QUE SE COMPILAN:

go list -f '{{.GoFiles}}' ./...
go list -tags debug -f '{{.GoFiles}}' ./...

VER QUÉ TAGS AFECTAN:
go list -f '{{.IgnoredGoFiles}}' ./...

ERRORES COMUNES:
go vet ./...    # Detecta build tags mal formados

DEBUG:
GODEBUG=gocacheverify=1 go build -tags debug`)
	// Información del sistema actual
	fmt.Printf("\n--- Sistema Actual ---\n")
	fmt.Printf("GOOS: %s\n", runtime.GOOS)
	fmt.Printf("GOARCH: %s\n", runtime.GOARCH)
	fmt.Printf("Compiler: %s\n", runtime.Compiler)
	fmt.Printf("Go Version: %s\n", runtime.Version())
}

/*
RESUMEN DE BUILD TAGS:

SINTAXIS:
  //go:build expression    (moderna, Go 1.17+)
  // +build expression     (legacy, aún soportada)

OPERADORES:
&&   AND
||   OR
!    NOT
()   Agrupación

TAGS PREDEFINIDOS:
- GOOS: linux, darwin, windows, etc.
- GOARCH: amd64, arm64, arm, etc.
- go1.XX: versión de Go
- cgo: CGO habilitado
- unix: sistemas Unix-like

CONVENCIÓN DE ARCHIVOS:
file_GOOS.go           # file_linux.go
file_GOARCH.go         # file_amd64.go
file_GOOS_GOARCH.go    # file_linux_amd64.go

CUSTOM TAGS:
  // go:build mytag
go build -tags mytag

PATRONES:
1. Features opcionales
2. Config por ambiente (dev/prod)
3. Drivers de DB
4. Optimizaciones por arch
5. Tests condicionales

COMANDOS:
go build -tags "tag1 tag2"
go test -tags integration
go list -f '{{.GoFiles}}'

TESTING:
  // go:build integration
  // go:build slow
  // go:build !unit

IGNORAR:
  // go:build ignore
_archivo.go
.archivo.go
testdata/
*/
