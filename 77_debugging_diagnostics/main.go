// Package main - Chapter 77: Debugging and Diagnostics
// Go provides powerful tools for debugging: Delve debugger,
// runtime/debug package, race detector, and memory/goroutine
// leak detection. Mastering these is essential for production systems.
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== DEBUGGING AND DIAGNOSTICS ===")

	// ============================================
	// DELVE DEBUGGER
	// ============================================
	fmt.Println("\n--- Delve Debugger ---")
	fmt.Println(`
INSTALACION:
go install github.com/go-delve/delve/cmd/dlv@latest

MODOS DE USO:

1. Debug un programa:
   dlv debug main.go
   dlv debug ./cmd/server

2. Attach a proceso existente:
   dlv attach <PID>

3. Debug tests:
   dlv test ./pkg/mypackage
   dlv test ./... -- -run TestSpecific

4. Core dumps:
   GOTRACEBACK=crash ./myapp
   dlv core ./myapp core.1234

COMANDOS BASICOS DE DELVE:
   break main.main          # Breakpoint en funcion
   break main.go:42         # Breakpoint en linea
   break myFunc if x > 10   # Breakpoint condicional
   continue (c)             # Continuar hasta breakpoint
   next (n)                 # Siguiente linea (step over)
   step (s)                 # Entrar en funcion (step into)
   stepout (so)             # Salir de funcion
   print (p) varName        # Imprimir variable
   locals                   # Ver variables locales
   args                     # Ver argumentos de funcion
   goroutines               # Listar goroutines
   goroutine <id>           # Cambiar a goroutine
   stack (bt)               # Ver stack trace
   on <bp> print x          # Ejecutar comando en breakpoint

DEBUGGING REMOTO:
   dlv debug --headless --listen=:2345 --api-version=2
   # Conectar desde VS Code o GoLand

VS CODE launch.json:
{
    "version": "0.2.0",
    "configurations": [{
        "name": "Launch",
        "type": "go",
        "request": "launch",
        "mode": "debug",
        "program": "${workspaceFolder}/cmd/server"
    }, {
        "name": "Attach",
        "type": "go",
        "request": "attach",
        "mode": "remote",
        "remotePath": "",
        "port": 2345,
        "host": "127.0.0.1"
    }]
}`)

	// ============================================
	// RUNTIME/DEBUG PACKAGE
	// ============================================
	fmt.Println("\n--- runtime/debug Package ---")

	// Build Info
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("Go version: %s\n", info.GoVersion)
		fmt.Printf("Module path: %s\n", info.Path)
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" || setting.Key == "vcs.time" {
				fmt.Printf("  %s = %s\n", setting.Key, setting.Value)
			}
		}
	}

	// Stack traces
	fmt.Println("\nStack trace actual:")
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false)
	fmt.Printf("(primeros 200 bytes): %s...\n", string(stackBuf[:min(n, 200)]))

	// SetGCPercent
	fmt.Println(`
debug.SetGCPercent(percent):
  debug.SetGCPercent(100) // Default: GC cuando heap dobla
  debug.SetGCPercent(200) // Menos GC, mas memoria
  debug.SetGCPercent(-1)  // Deshabilitar GC

debug.SetMemoryLimit(limit):
  debug.SetMemoryLimit(1 << 30) // 1 GiB soft limit

debug.FreeOSMemory():
  // Forzar devolver memoria al OS
  debug.FreeOSMemory()`)

	// SetMaxStack
	fmt.Println(`
debug.SetMaxStack(bytes):
  // Limitar tamano maximo de stack (default 1GB)
  debug.SetMaxStack(32 * 1024 * 1024) // 32 MB

debug.SetMaxThreads(threads):
  // Limitar threads del OS (default 10000)
  debug.SetMaxThreads(500)`)

	// SetTraceback
	fmt.Println(`
GOTRACEBACK controla detalle de panic:
  GOTRACEBACK=none    // Solo el panic message
  GOTRACEBACK=single  // Stack del goroutine que hizo panic (default)
  GOTRACEBACK=all     // Stack de todos los goroutines
  GOTRACEBACK=system  // Incluye goroutines del runtime
  GOTRACEBACK=crash   // Genera core dump`)

	// ============================================
	// RACE DETECTOR DEEP DIVE
	// ============================================
	fmt.Println("\n--- Race Detector Deep Dive ---")
	fmt.Println(`
ACTIVAR:
  go run -race main.go
  go test -race ./...
  go build -race -o myapp

COMO FUNCIONA:
- Instrumenta cada acceso a memoria (read/write)
- Usa algoritmo ThreadSanitizer (TSan)
- Detecta accesos concurrentes sin sincronizacion
- Ralentiza 5-10x y usa 5-10x mas memoria
- SOLO detecta races que OCURREN en la ejecucion

EJEMPLO DE DATA RACE:`)

	// Demostrar deteccion de race condition
	demonstrateRaceFreeCode()

	fmt.Println(`
PATRONES COMUNES DE RACES:

1. MAP CONCURRENTE:
   // RACE! Maps no son goroutine-safe
   go func() { m["key"] = "val" }()
   go func() { _ = m["key"] }()
   // FIX: usar sync.Map o mutex

2. SLICE APPEND:
   // RACE! Append puede modificar underlying array
   go func() { s = append(s, 1) }()
   go func() { s = append(s, 2) }()
   // FIX: usar mutex o channel

3. INTERFACE CHECK:
   // RACE! Interface es (type, pointer) - no atomico
   go func() { err = errors.New("x") }()
   go func() { if err != nil {} }()
   // FIX: usar mutex o atomic.Value

4. LOOP VARIABLE CAPTURE (pre Go 1.22):
   for _, v := range items {
       go func() { use(v) }()  // RACE!
   }
   // FIX (pre 1.22): go func(v T) { use(v) }(v)
   // Go 1.22+: cada iteracion tiene su propia variable

CI INTEGRATION:
  # En CI, ejecutar tests con race detector
  go test -race -count=1 ./...

  # Suprimir falsos positivos (raro pero posible):
  // Archivo: race_test.go
  //go:build !race`)

	// ============================================
	// GOROUTINE LEAK DETECTION
	// ============================================
	fmt.Println("\n--- Goroutine Leak Detection ---")

	// Mostrar conteo de goroutines
	before := runtime.NumGoroutine()
	fmt.Printf("Goroutines antes: %d\n", before)

	// Simular leak y fix
	demonstrateGoroutineLeak()
	time.Sleep(50 * time.Millisecond)

	after := runtime.NumGoroutine()
	fmt.Printf("Goroutines despues: %d (leak si crece sin parar)\n", after)

	fmt.Println(`
DETECTAR LEAKS CON goleak (Uber):
  go get go.uber.org/goleak

  // En tests:
  func TestMain(m *testing.M) {
      goleak.VerifyTestMain(m)
  }

  // O por test individual:
  func TestNoLeak(t *testing.T) {
      defer goleak.VerifyNone(t)
      // ... test code ...
  }

PATRONES QUE CAUSAN LEAKS:

1. CHANNEL SIN RECEPTOR:
   ch := make(chan int)
   go func() {
       result := compute()
       ch <- result  // Bloqueado para siempre si nadie lee
   }()
   // FIX: usar buffered channel o context

2. CONTEXT SIN CANCEL:
   ctx, cancel := context.WithCancel(ctx)
   // Si olvidas cancel(), el goroutine asociado leaks
   defer cancel()  // SIEMPRE defer cancel

3. HTTP BODY SIN CERRAR:
   resp, _ := http.Get(url)
   // resp.Body.Close() olvidado = leak de goroutine + fd
   defer resp.Body.Close()

4. TICKER SIN STOP:
   ticker := time.NewTicker(time.Second)
   // ticker.Stop() olvidado = leak
   defer ticker.Stop()

DIAGNOSTICO:
  // Dump de todos los goroutines
  pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)

  // Via HTTP
  curl http://localhost:6060/debug/pprof/goroutine?debug=2`)

	// ============================================
	// MEMORY LEAK DETECTION
	// ============================================
	fmt.Println("\n--- Memory Leak Detection ---")

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("HeapAlloc: %d KB\n", stats.HeapAlloc/1024)
	fmt.Printf("HeapObjects: %d\n", stats.HeapObjects)
	fmt.Printf("NumGC: %d\n", stats.NumGC)
	fmt.Printf("GoroutineCount: %d\n", runtime.NumGoroutine())

	os.Stdout.WriteString(`
CAMPOS IMPORTANTES DE runtime.MemStats:
  Alloc        // Bytes allocados actualmente (heap)
  TotalAlloc   // Total bytes allocados (acumulado)
  Sys          // Bytes obtenidos del OS
  HeapAlloc    // Bytes en heap activos
  HeapIdle     // Bytes en heap no usados
  HeapInuse    // Bytes en heap en uso
  HeapObjects  // Numero de objetos en heap
  NumGC        // Numero de ciclos de GC
  PauseTotalNs // Tiempo total de pausa de GC
  LastGC       // Timestamp del ultimo GC

MONITOREO CONTINUO:
  go func() {
      var stats runtime.MemStats
      ticker := time.NewTicker(10 * time.Second)
      defer ticker.Stop()
      for range ticker.C {
          runtime.ReadMemStats(&stats)
          log.Printf("HeapAlloc=%dMB HeapObjects=%d NumGC=%d",
              stats.HeapAlloc/1024/1024,
              stats.HeapObjects,
              stats.NumGC,
          )
      }
  }()

HEAP PROFILE ANALYSIS:
  go tool pprof http://localhost:6060/debug/pprof/heap
  > top10
  > list FunctionName

  # Comparar snapshots para encontrar leaks:
  curl -o before.prof localhost:6060/debug/pprof/heap
  # ... esperar ...
  curl -o after.prof localhost:6060/debug/pprof/heap
  go tool pprof -diff_base=before.prof after.prof
  > top10  # Muestra que crecio
`)

	// ============================================
	// GODEBUG VARIABLES
	// ============================================
	fmt.Println("\n--- GODEBUG Environment Variables ---")
	fmt.Println(`
GODEBUG permite activar diagnosticos del runtime:

GC:
  GODEBUG=gctrace=1           // Traza de cada GC
  // gc 1 @0.012s 2%: 0.024+1.2+0.003 ms clock, 0.096+0/1.2/0+0.012 ms cpu

SCHEDULER:
  GODEBUG=schedtrace=1000     // Traza del scheduler cada 1000ms
  GODEBUG=scheddetail=1       // Detalle por-P del scheduler

HTTP:
  GODEBUG=http2debug=1        // Debug de HTTP/2
  GODEBUG=http2debug=2        // Verbose HTTP/2

TLS:
  GODEBUG=tls13=0             // Deshabilitar TLS 1.3
  GODEBUG=tlsmaxrsasize=8192  // Max RSA key size

ALLOCATOR:
  GODEBUG=madvdontneed=1      // Devolver memoria al OS mas rapido

COMBINACION:
  GODEBUG=gctrace=1,schedtrace=1000 ./myapp`)

	// ============================================
	// STACK TRACE ANALYSIS
	// ============================================
	fmt.Println("\n--- Stack Trace Analysis ---")

	// runtime.Caller
	demonstrateRuntimeCaller()

	os.Stdout.WriteString(`
LEER STACK TRACES DE PANIC:

goroutine 1 [running]:
main.foo(0x1, 0x2)          // Funcion y argumentos
    /path/main.go:42 +0x1a  // Archivo:linea + offset
main.bar(...)                // Funcion llamadora
    /path/main.go:38 +0x2f
main.main()
    /path/main.go:10 +0x3b

INTERPRETAR:
- goroutine N [estado]:  estado puede ser:
  - running: ejecutando
  - runnable: listo para ejecutar
  - sleep: en time.Sleep
  - chan receive: bloqueado en <-ch
  - chan send: bloqueado en ch<-
  - select: bloqueado en select
  - IO wait: esperando I/O
  - semacquire: esperando mutex/semaphore

- Los argumentos se muestran como valores raw (hex)
- +0x1a es el offset desde el inicio de la funcion

FULL GOROUTINE DUMP (en cualquier momento):
  kill -SIGQUIT <PID>   // Unix: imprime todas las goroutines
  kill -SIGABRT <PID>   // Genera core dump

  // O programaticamente:
  buf := make([]byte, 1<<20)
  n := runtime.Stack(buf, true) // true = todos los goroutines
  fmt.Printf("%%s\n", buf[:n])
`)

	// ============================================
	// COMMON DEBUGGING TECHNIQUES
	// ============================================
	fmt.Println("\n--- Debugging Techniques ---")
	os.Stdout.WriteString(`
1. PRINTF DEBUGGING (rapido pero temporal):
   fmt.Fprintf(os.Stderr, "DEBUG: x=%v at %s\n", x, time.Now())

2. CONDITIONAL BREAKPOINTS con build tags:
   //go:build debug

   func debugLog(msg string, args ...any) {
       fmt.Fprintf(os.Stderr, msg+"\n", args...)
   }

3. DEADLOCK DETECTION:
   // Go detecta deadlocks simples automaticamente:
   // "fatal error: all goroutines are asleep - deadlock!"

   // Para deadlocks parciales, usar timeout:
   select {
   case result := <-ch:
       process(result)
   case <-time.After(5 * time.Second):
       log.Fatal("probable deadlock: timeout waiting for result")
   }

4. COMPILE-TIME ASSERTIONS:
   // Verificar que un tipo implementa interface en compile time
   var _ io.Reader = (*MyType)(nil)

   // Verificar tamano de struct
   const _ = uint(unsafe.Sizeof(MyStruct{}) - 64) // Panic si != 64

5. ESCAPE ANALYSIS:
   go build -gcflags="-m -m" 2>&1 | grep "escapes"
   // Muestra que variables escapan al heap

6. BINARY INSPECTION:
   go tool objdump -s "main.myFunc" ./myapp  // Disassembly
   go tool nm ./myapp | grep myFunc          // Symbols
   go version -m ./myapp                     // Build info
`)

	// ============================================
	// PRODUCTION DEBUGGING
	// ============================================
	fmt.Println("\n--- Production Debugging ---")
	fmt.Println(`
PRINCIPIOS:
- Nunca debuggear con Delve en produccion directamente
- Usar observabilidad: logs, metrics, traces
- Habilitar pprof endpoint (protegido!)
- Configurar GOTRACEBACK=crash para core dumps

DIAGNOSTICS ENDPOINT SEGURO:
  mux := http.NewServeMux()

  // Solo accesible internamente
  debugMux := http.NewServeMux()
  debugMux.HandleFunc("/debug/pprof/", pprof.Index)
  debugMux.HandleFunc("/debug/pprof/profile", pprof.Profile)

  // Escuchar en puerto interno solamente
  go http.ListenAndServe("localhost:6060", debugMux)

  // Puerto publico no expone debug
  http.ListenAndServe(":8080", mux)

CONTINUOUS PROFILING:
  // Usar servicios como:
  // - Pyroscope (open source)
  // - Datadog Continuous Profiler
  // - Google Cloud Profiler

  // Ejemplo con pyroscope:
  pyroscope.Start(pyroscope.Config{
      ApplicationName: "my-service",
      ServerAddress:   "http://pyroscope:4040",
      ProfileTypes: []pyroscope.ProfileType{
          pyroscope.ProfileCPU,
          pyroscope.ProfileAllocObjects,
          pyroscope.ProfileGoroutines,
      },
  })`)

	fmt.Println("\n=== FIN CAPITULO 77: DEBUGGING AND DIAGNOSTICS ===")
}

// demonstrateRaceFreeCode muestra codigo seguro con sincronizacion
func demonstrateRaceFreeCode() {
	var mu sync.Mutex
	counter := 0

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}
	wg.Wait()
	fmt.Printf("Counter (race-free con mutex): %d\n", counter)
}

// demonstrateGoroutineLeak muestra como prevenir leaks
func demonstrateGoroutineLeak() {
	// MAL: goroutine que nunca termina si nadie lee
	// ch := make(chan int)
	// go func() { ch <- expensiveComputation() }()
	// return // ch nunca se lee -> goroutine leak

	// BIEN: usar context para cancelar
	done := make(chan struct{})
	ch := make(chan int, 1) // Buffered: no bloquea aunque nadie lea

	go func() {
		select {
		case ch <- 42:
		case <-done:
			return
		}
	}()

	result := <-ch
	close(done)
	fmt.Printf("Resultado sin leak: %d\n", result)
}

// demonstrateRuntimeCaller muestra como obtener info de caller
func demonstrateRuntimeCaller() {
	pc, file, line, ok := runtime.Caller(0)
	if ok {
		fn := runtime.FuncForPC(pc)
		fmt.Printf("Caller info:\n")
		fmt.Printf("  Function: %s\n", fn.Name())
		fmt.Printf("  File: %s\n", file)
		fmt.Printf("  Line: %d\n", line)
	}

	// runtime.Callers para full stack
	var pcs [10]uintptr
	n := runtime.Callers(1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	fmt.Println("Call stack (primeros 3):")
	for i := 0; i < 3; i++ {
		frame, more := frames.Next()
		fmt.Printf("  %s (%s:%d)\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
