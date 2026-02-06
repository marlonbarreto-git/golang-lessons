// Package main - Chapter 066: Go 1.25 Features
// Go 1.25 (Aug 2025) introdujo WaitGroup.Go(), JSON v2 experimental,
// y mejoras significativas en el runtime y la biblioteca estándar.
package main

import (
	"os"
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== GO 1.25 FEATURES ===")
	fmt.Printf("Running on Go %s\n", runtime.Version())

	// ============================================
	// WAITGROUP.GO()
	// ============================================
	fmt.Println("\n--- WaitGroup.Go() ---")

	// Antes (Go 1.24 y anteriores):
	var wgOld sync.WaitGroup
	for i := 0; i < 3; i++ {
		wgOld.Add(1)
		go func(n int) {
			defer wgOld.Done()
			fmt.Printf("Old worker %d\n", n)
		}(i)
	}
	wgOld.Wait()

	// Ahora (Go 1.25+):
	var wgNew sync.WaitGroup
	for i := 0; i < 3; i++ {
		n := i // capturar variable
		wgNew.Go(func() {
			fmt.Printf("New worker %d\n", n)
		})
	}
	wgNew.Wait()

	fmt.Println(`
WaitGroup.Go() combina:
- wg.Add(1)
- go func()
- defer wg.Done()

Más conciso y menos propenso a errores.`)
	// ============================================
	// JSON V2 (EXPERIMENTAL)
	// ============================================
	fmt.Println("\n--- JSON v2 (Experimental) ---")
	fmt.Println(`
Habilitar con: GOEXPERIMENT=jsonv2 go build

Nuevos paquetes:
- encoding/json/v2: revisión major de encoding/json
- encoding/json/jsontext: procesamiento de bajo nivel

Mejoras:
- Mejor rendimiento de decodificación
- Nuevas opciones de configuración
- Mejor manejo de errores

Ejemplo:
  import jsonv2 "encoding/json/v2"

  opts := jsonv2.Options{
      Indent: "  ",
      OmitEmptyStructs: true,
  }
  data, err := jsonv2.Marshal(value, opts)`)
	// ============================================
	// SYNCTEST PACKAGE (ESTABILIZADO)
	// ============================================
	fmt.Println("\n--- synctest Package ---")
	fmt.Println(`
El paquete testing/synctest ahora es estable:

func TestConcurrent(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        // Tiempo virtual - clock avanza cuando goroutines bloquean
        done := make(chan bool)

        go func() {
            time.Sleep(time.Hour) // instantáneo en virtual time
            done <- true
        }()

        <-done
        synctest.Wait() // esperar que todas las goroutines bloqueen
    })
}

Características:
- Tiempo virtual determinista
- synctest.Wait() sincroniza goroutines
- Ideal para tests de código concurrente`)
	// ============================================
	// CONTAINER-AWARE GOMAXPROCS
	// ============================================
	fmt.Println("\n--- Container-Aware GOMAXPROCS ---")
	fmt.Println(`
Go 1.25 detecta automáticamente límites de CPU en containers:

En Docker/Kubernetes:
- Considera cgroup CPU bandwidth limits
- Ajusta GOMAXPROCS automáticamente
- Actualiza periódicamente si los límites cambian

Desactivar con:
  GOMAXPROCS=N (variable de entorno)
  runtime.GOMAXPROCS(n)
  GODEBUG=containercpu=0`)
	fmt.Printf("GOMAXPROCS actual: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())

	// ============================================
	// EXPERIMENTAL GREEN TEA GC
	// ============================================
	fmt.Println("\n--- Green Tea GC (Experimental) ---")
	fmt.Println(`
Nuevo garbage collector experimental:

Habilitar: GOEXPERIMENT=greenteagc go build

Mejoras esperadas:
- 10-40% reducción en overhead de GC
- Mejor localidad de marcado/escaneo
- Mayor escalabilidad en CPUs múltiples

Para workloads con mucha memoria/objetos.`)
	// ============================================
	// FLIGHT RECORDER
	// ============================================
	fmt.Println("\n--- Trace Flight Recorder ---")
	fmt.Println(`
Nueva API para captura de traces on-demand:

import "runtime/trace"

recorder := trace.NewFlightRecorder()
defer recorder.Stop()

// ... aplicación ejecutando ...

// Cuando ocurre algo interesante:
recorder.WriteTo(file)

Características:
- Ring buffer en memoria
- Bajo overhead cuando no se captura
- Snapshot en cualquier momento
- Útil para debugging de producción`)
	// ============================================
	// CRYPTO IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Crypto Improvements ---")
	fmt.Println(`
Nuevas características:

1. crypto.SignMessage():
   signer := privateKey.(crypto.MessageSigner)
   signature, err := crypto.SignMessage(signer, nil, message)

2. RSA 3x más rápido en generación de claves

3. SHA-1 2x más rápido en amd64 con SHA-NI

4. SHA-3 2x más rápido en Apple M series

5. Nuevos métodos:
   - hash.Clone() en todas las implementaciones
   - crypto/sha3.Clone()
   - crypto/ecdsa.ParseRawPrivateKey()`)
	// ============================================
	// DWARF5 DEBUG INFO
	// ============================================
	fmt.Println("\n--- DWARF5 Debug Info ---")
	fmt.Println(`
Go 1.25 genera debug info en DWARF version 5:

Ventajas:
- Menor tamaño de binarios
- Linking más rápido (especialmente binarios grandes)

Desactivar: GOEXPERIMENT=nodwarf5

Afecta principalmente herramientas de debugging.`)
	// ============================================
	// NET/HTTP CSRF PROTECTION
	// ============================================
	fmt.Println("\n--- HTTP CSRF Protection ---")
	fmt.Println(`
Nuevo: net/http.CrossOriginProtection

handler := http.CrossOriginProtection(myHandler)

Protege contra:
- Cross-Origin Request Forgery
- Clickjacking
- Algunos tipos de XSS

Configuración:
  http.CrossOriginProtectionConfig{
      AllowedOrigins: []string{"https://myapp.com"},
      SameSite: http.SameSiteStrictMode,
  }`)
	// ============================================
	// REFLECT.TYPEASSERT
	// ============================================
	fmt.Println("\n--- reflect.TypeAssert ---")
	fmt.Println(`
Nueva función para type assertion sin allocations:

import "reflect"

// Antes (causa allocation):
if v, ok := iface.(MyType); ok { ... }

// Ahora (zero allocation):
var target MyType
if reflect.TypeAssert(iface, &target) {
    // target contiene el valor
}

Útil en código sensible a performance.`)
	// ============================================
	// IO/FS READLINKFS
	// ============================================
	fmt.Println("\n--- io/fs.ReadLinkFS ---")
	fmt.Println(`
Nueva interface para leer symlinks:

type ReadLinkFS interface {
    FS
    ReadLink(name string) (string, error)
}

os.DirFS y os.Root ahora implementan ReadLinkFS.`)
	// ============================================
	// LOG/SLOG IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- log/slog Improvements ---")
	os.Stdout.WriteString(`
Nuevas funciones:

// GroupAttrs crea grupo de atributos
attrs := slog.GroupAttrs("request",
    slog.String("method", "GET"),
    slog.String("path", "/api"),
)

// Record.Source() obtiene información de ubicación
rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
source := rec.Source()
fmt.Printf("%s:%d", source.File, source.Line)
`)

	// ============================================
	// TESTING IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Testing Improvements ---")
	fmt.Println(`
Nuevos métodos:

// Attr para metadatos de test
t.Attr("key", "value")

// Output writer para custom output
w := t.Output()
fmt.Fprintf(w, "Custom output")

// AllocsPerRun ahora panic si se usa en parallel tests
// (era un bug silencioso antes)`)
	// ============================================
	// EJEMPLO: CÓDIGO MODERNIZADO
	// ============================================
	fmt.Println("\n--- Ejemplo: Código Modernizado ---")

	// Worker pool con WaitGroup.Go()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	results := make(chan int, 10)
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		n := i
		wg.Go(func() {
			select {
			case results <- n * n:
			case <-ctx.Done():
			}
		})
	}

	// Cerrar results cuando todos terminen
	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Print("Resultados: ")
	for r := range results {
		fmt.Printf("%d ", r)
	}
	fmt.Println()

	// ============================================
	// PLATFORM CHANGES
	// ============================================
	fmt.Println("\n--- Platform Changes ---")
	fmt.Println(`
macOS:
- Requiere macOS 12 Monterey o posterior
- macOS 11 Big Sur ya no soportado

Windows:
- Go 1.25 es la última versión para 32-bit windows/arm

AMD64:
- GOAMD64=v3+: usa FMA para math más rápido

Loong64:
- Soporte para race detector
- Plugin build mode en linux/riscv64`)}

/*
RESUMEN GO 1.25:

HIGHLIGHTS:
- sync.WaitGroup.Go() - método más conveniente
- encoding/json/v2 experimental
- testing/synctest estabilizado
- Container-aware GOMAXPROCS
- Green Tea GC experimental
- Trace Flight Recorder

RENDIMIENTO:
- RSA keygen 3x más rápido
- SHA-1 2x más rápido (amd64 SHA-NI)
- SHA-3 2x más rápido (Apple M)
- DWARF5 reduce tamaño de binarios

STDLIB:
- net/http.CrossOriginProtection
- reflect.TypeAssert (zero alloc)
- io/fs.ReadLinkFS
- slog.GroupAttrs, Record.Source()

TESTING:
- synctest.Test() estable
- t.Attr(), t.Output()

PLATFORM:
- macOS 12+ requerido
- Última versión para windows/arm 32-bit
*/
