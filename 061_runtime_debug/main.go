// Package main - Chapter 061: Runtime, Debug & Diagnostics
// Informacion de build con runtime/debug, metricas del runtime,
// tracing con runtime/trace, inspeccion de binarios con debug/*,
// y variables de monitoreo con expvar.
package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== RUNTIME, DEBUG & DIAGNOSTICS ===")

	// ============================================
	// 1. RUNTIME/DEBUG - BUILD INFO
	// ============================================
	fmt.Println("\n--- 1. runtime/debug - BuildInfo ---")
	os.Stdout.WriteString(`
RUNTIME/DEBUG - INFORMACION DE BUILD:

  debug.ReadBuildInfo() (*BuildInfo, bool)  // Info del binario actual

  BuildInfo contiene:
    GoVersion   string          // Version de Go
    Path        string          // Modulo principal
    Main        Module          // Info del modulo principal
    Deps        []*Module       // Dependencias
    Settings    []BuildSetting  // Configuracion de build

  BuildSetting: Key/Value pairs como:
    -compiler, GOARCH, GOOS, GOAMD64, vcs, vcs.revision, vcs.time, vcs.modified
`)

	info, ok := debug.ReadBuildInfo()
	if ok {
		fmt.Printf("  Go Version: %s\n", info.GoVersion)
		fmt.Printf("  Path:       %s\n", info.Path)
		if info.Main.Path != "" {
			fmt.Printf("  Main Module: %s@%s\n", info.Main.Path, info.Main.Version)
		}

		fmt.Println("  Build Settings:")
		for _, s := range info.Settings {
			if s.Value != "" {
				fmt.Printf("    %s = %s\n", s.Key, s.Value)
			}
		}

		if len(info.Deps) > 0 {
			fmt.Printf("  Dependencies (%d):\n", len(info.Deps))
			for _, dep := range info.Deps {
				fmt.Printf("    %s@%s\n", dep.Path, dep.Version)
			}
		} else {
			fmt.Println("  Dependencies: (ninguna)")
		}
	}

	// ============================================
	// 2. RUNTIME/DEBUG - GC CONTROL
	// ============================================
	fmt.Println("\n--- 2. runtime/debug - Control del GC ---")
	os.Stdout.WriteString(`
RUNTIME/DEBUG - CONTROL DEL GARBAGE COLLECTOR:

  debug.SetGCPercent(percent int) int       // Ajustar umbral del GC
    - 100 (default): GC cuando heap crece 100%
    - 50: GC mas agresivo (mas frecuente)
    - 200: GC menos frecuente
    - -1: Desactivar GC automatico

  debug.FreeOSMemory()                      // Forzar devolucion de memoria al OS
  debug.SetMaxStack(bytes int) int          // Tamano maximo del stack
  debug.SetMaxThreads(threads int) int      // Maximo de threads del OS

  debug.SetMemoryLimit(limit int64) int64   // Limite de memoria (Go 1.19+)
    - Soft limit para el runtime de Go
    - El GC se vuelve mas agresivo al acercarse al limite
`)

	prevGC := debug.SetGCPercent(100)
	fmt.Printf("  GC Percent anterior: %d\n", prevGC)
	debug.SetGCPercent(prevGC)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("  HeapAlloc:  %d KB\n", memStats.HeapAlloc/1024)
	fmt.Printf("  HeapSys:    %d KB\n", memStats.HeapSys/1024)
	fmt.Printf("  NumGC:      %d\n", memStats.NumGC)

	allocateAndRelease()

	runtime.ReadMemStats(&memStats)
	fmt.Printf("  Despues de alloc/release:\n")
	fmt.Printf("    HeapAlloc: %d KB\n", memStats.HeapAlloc/1024)
	fmt.Printf("    NumGC:     %d\n", memStats.NumGC)

	debug.FreeOSMemory()
	runtime.ReadMemStats(&memStats)
	fmt.Printf("  Despues de FreeOSMemory:\n")
	fmt.Printf("    HeapAlloc: %d KB\n", memStats.HeapAlloc/1024)

	// ============================================
	// 3. RUNTIME/DEBUG - STACK TRACE
	// ============================================
	fmt.Println("\n--- 3. runtime/debug - Stack trace ---")
	os.Stdout.WriteString(`
RUNTIME/DEBUG - STACK TRACES:

  debug.Stack() []byte                      // Stack trace actual
  debug.PrintStack()                        // Imprimir stack a stderr

  Util para:
  - Debugging en produccion
  - Logging de errores con contexto
  - Panic recovery con informacion detallada
`)

	stack := debug.Stack()
	lines := strings.Split(string(stack), "\n")
	fmt.Println("  Stack trace actual (primeras lineas):")
	limit := 10
	if len(lines) < limit {
		limit = len(lines)
	}
	for _, line := range lines[:limit] {
		fmt.Printf("    %s\n", line)
	}
	if len(lines) > limit {
		fmt.Printf("    ... (%d lineas mas)\n", len(lines)-limit)
	}

	// ============================================
	// 4. RUNTIME/METRICS - METRICAS DEL RUNTIME
	// ============================================
	fmt.Println("\n--- 4. runtime/metrics ---")
	os.Stdout.WriteString(`
RUNTIME/METRICS (Go 1.16+):

  metrics.All() []Description              // Todas las metricas disponibles
  metrics.Read(samples []Sample)           // Leer valores de metricas

  Description:
    Name        string      // Nombre (ej: /gc/cycles/total:gc-cycles)
    Description string      // Descripcion
    Kind        ValueKind   // Tipo (Uint64, Float64, Float64Histogram)
    Cumulative  bool        // Es acumulativa?

  Categorias de metricas:
    /gc/*         - Garbage collector
    /memory/*     - Memoria
    /sched/*      - Scheduler
    /cpu/*        - CPU
    /godebug/*    - Debug settings
`)

	allMetrics := metrics.All()
	fmt.Printf("  Total metricas disponibles: %d\n", len(allMetrics))

	fmt.Println("\n  Metricas de memoria y GC:")
	interestingMetrics := []string{
		"/memory/classes/heap/objects:bytes",
		"/memory/classes/total:bytes",
		"/gc/cycles/total:gc-cycles",
		"/gc/heap/allocs:bytes",
		"/sched/goroutines:goroutines",
	}

	samples := make([]metrics.Sample, len(interestingMetrics))
	for i, name := range interestingMetrics {
		samples[i].Name = name
	}
	metrics.Read(samples)

	for _, s := range samples {
		switch s.Value.Kind() {
		case metrics.KindUint64:
			fmt.Printf("    %-50s = %d\n", s.Name, s.Value.Uint64())
		case metrics.KindFloat64:
			fmt.Printf("    %-50s = %.2f\n", s.Name, s.Value.Float64())
		case metrics.KindFloat64Histogram:
			h := s.Value.Float64Histogram()
			fmt.Printf("    %-50s = histogram (%d buckets)\n", s.Name, len(h.Buckets))
		default:
			fmt.Printf("    %-50s = (tipo: %v)\n", s.Name, s.Value.Kind())
		}
	}

	fmt.Println("\n  Listado de metricas por categoria:")
	categories := map[string]int{}
	for _, m := range allMetrics {
		parts := strings.SplitN(m.Name, "/", 3)
		if len(parts) >= 2 {
			cat := "/" + parts[1]
			categories[cat]++
		}
	}
	for cat, count := range categories {
		fmt.Printf("    %-15s: %d metricas\n", cat, count)
	}

	// ============================================
	// 5. RUNTIME/TRACE - TRACING
	// ============================================
	fmt.Println("\n--- 5. runtime/trace (conceptos) ---")
	os.Stdout.WriteString(`
RUNTIME/TRACE - TRACING DE EJECUCION:

  trace.Start(w io.Writer)                 // Iniciar captura de trace
  trace.Stop()                             // Detener captura

  trace.Log(ctx, category, message)        // Agregar log al trace
  trace.Logf(ctx, category, format, args)  // Log con formato

  trace.NewTask(ctx, type) (ctx, *Task)    // Crear tarea
  task.End()                               // Finalizar tarea

  trace.WithRegion(ctx, type, func())      // Region con nombre
  trace.StartRegion(ctx, type) *Region     // Region manual
  region.End()                             // Fin de region

USO:
  f, _ := os.Create("trace.out")
  trace.Start(f)
  defer trace.Stop()
  // ... ejecutar programa ...
  // Analizar: go tool trace trace.out

EJEMPLO (no ejecutamos para no crear archivos):
  ctx, task := trace.NewTask(context.Background(), "processOrder")
  defer task.End()

  trace.WithRegion(ctx, "validateInput", func() {
      // validar...
  })

  trace.WithRegion(ctx, "saveToDatabase", func() {
      // guardar...
  })
`)

	fmt.Println("  (trace.Start/Stop no se ejecuta para evitar crear archivos)")
	fmt.Println("  Para capturar trace en tu programa:")
	fmt.Println("    1. go test -trace=trace.out ./...")
	fmt.Println("    2. go tool trace trace.out")
	fmt.Println("    3. Abrir en navegador para visualizacion interactiva")

	// ============================================
	// 6. DEBUG/BUILDINFO - LEER DESDE ARCHIVO
	// ============================================
	fmt.Println("\n--- 6. debug/buildinfo ---")
	os.Stdout.WriteString(`
DEBUG/BUILDINFO:

  buildinfo.ReadFile(name string)          // Leer build info de un binario Go

  Permite inspeccionar cualquier binario de Go:
  - Version de Go con la que fue compilado
  - Modulo principal y dependencias
  - Settings de build

EJEMPLO:
  info, err := buildinfo.ReadFile("/usr/local/bin/gopls")
  fmt.Println(info.GoVersion)    // go1.22.0
  fmt.Println(info.Main.Path)   // golang.org/x/tools/gopls
`)

	fmt.Println("  (No leemos binarios externos por portabilidad)")

	// ============================================
	// 7. EXPVAR - VARIABLES PUBLICADAS
	// ============================================
	fmt.Println("\n--- 7. expvar - Variables de monitoreo ---")
	os.Stdout.WriteString(`
EXPVAR - VARIABLES PUBLICAS PARA MONITOREO:

  expvar.NewInt(name) *expvar.Int          // Entero atomico
  expvar.NewFloat(name) *expvar.Float      // Float atomico
  expvar.NewString(name) *expvar.String    // String atomico
  expvar.NewMap(name) *expvar.Map          // Mapa de variables

  Metodos de Int: Add(delta), Set(value), Value()
  Metodos de Float: Add(delta), Set(value), Value()
  Metodos de String: Set(value), Value()
  Metodos de Map: Add(key, delta), Set(key, var), Get(key), Do(func)

  expvar.Publish(name, var)                // Publicar variable custom
  expvar.Do(func(KeyValue))                // Iterar todas las variables
  expvar.Handler()                         // HTTP handler para /debug/vars

  Al importar expvar, se registra automaticamente:
  - /debug/vars endpoint en DefaultServeMux
  - Variables "cmdline" y "memstats" por defecto
`)

	requestCount := expvar.NewInt("requests_total")
	errorCount := expvar.NewInt("errors_total")
	latency := expvar.NewFloat("avg_latency_ms")
	appVersion := expvar.NewString("app_version")
	endpoints := expvar.NewMap("endpoints")

	appVersion.Set("2.1.0")
	for i := 0; i < 100; i++ {
		requestCount.Add(1)
		if i%10 == 0 {
			errorCount.Add(1)
		}
	}
	latency.Set(12.5)

	endpoints.Add("GET /api/users", 45)
	endpoints.Add("POST /api/orders", 23)
	endpoints.Add("GET /api/products", 67)
	endpoints.Add("DELETE /api/cache", 5)

	fmt.Printf("  requests_total: %d\n", requestCount.Value())
	fmt.Printf("  errors_total:   %d\n", errorCount.Value())
	fmt.Printf("  avg_latency_ms: %.1f\n", latency.Value())
	fmt.Printf("  app_version:    %s\n", appVersion.Value())

	fmt.Println("  endpoints:")
	endpoints.Do(func(kv expvar.KeyValue) {
		fmt.Printf("    %-25s = %s\n", kv.Key, kv.Value.String())
	})

	// ============================================
	// 8. EXPVAR - VARIABLE CUSTOM
	// ============================================
	fmt.Println("\n--- 8. expvar - Variable custom ---")
	os.Stdout.WriteString(`
EXPVAR - IMPLEMENTAR expvar.Var:

  type Var interface {
      String() string   // Retorna representacion JSON
  }

  Cualquier tipo que implemente String() puede publicarse.
`)

	sysInfo := &SystemInfo{
		Hostname: "app-server-01",
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUs:     runtime.NumCPU(),
		Started:  time.Now(),
	}
	expvar.Publish("system_info", sysInfo)
	fmt.Printf("  system_info: %s\n", sysInfo.String())

	uptime := &UptimeVar{start: time.Now().Add(-2 * time.Hour)}
	expvar.Publish("uptime_seconds", uptime)
	fmt.Printf("  uptime_seconds: %s\n", uptime.String())

	// ============================================
	// 9. EXPVAR - ENDPOINT HTTP
	// ============================================
	fmt.Println("\n--- 9. expvar - HTTP endpoint ---")
	os.Stdout.WriteString(`
EXPVAR HTTP ENDPOINT:

  Al importar "expvar", se registra automaticamente:
    http.HandleFunc("/debug/vars", expvar.Handler())

  Respuesta JSON con todas las variables publicadas:
  {
    "cmdline": ["./myapp", "-port", "8080"],
    "memstats": {...},
    "requests_total": 100,
    "errors_total": 10,
    "app_version": "2.1.0"
  }

  Uso tipico:
    go func() {
        log.Fatal(http.ListenAndServe(":6060", nil))
    }()

  Consultar: curl http://localhost:6060/debug/vars | jq .
`)

	fmt.Println("  Todas las variables expvar publicadas:")
	expvar.Do(func(kv expvar.KeyValue) {
		val := kv.Value.String()
		if len(val) > 60 {
			val = val[:60] + "..."
		}
		fmt.Printf("    %-20s = %s\n", kv.Key, val)
	})

	fmt.Println("\n  Handler HTTP disponible en: expvar.Handler()")
	_ = http.DefaultServeMux

	// ============================================
	// 10. EJEMPLO: DASHBOARD DE RUNTIME
	// ============================================
	fmt.Println("\n--- 10. Ejemplo: Dashboard de runtime ---")

	runtime.ReadMemStats(&memStats)
	fmt.Println("  === Runtime Dashboard ===")
	fmt.Printf("  Go Version:    %s\n", runtime.Version())
	fmt.Printf("  GOOS/GOARCH:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  NumCPU:        %d\n", runtime.NumCPU())
	fmt.Printf("  NumGoroutine:  %d\n", runtime.NumGoroutine())
	fmt.Printf("  GOMAXPROCS:    %d\n", runtime.GOMAXPROCS(0))
	fmt.Println("  --- Memoria ---")
	fmt.Printf("  Alloc:         %d KB\n", memStats.Alloc/1024)
	fmt.Printf("  TotalAlloc:    %d KB\n", memStats.TotalAlloc/1024)
	fmt.Printf("  Sys:           %d KB\n", memStats.Sys/1024)
	fmt.Printf("  HeapObjects:   %d\n", memStats.HeapObjects)
	fmt.Printf("  HeapInuse:     %d KB\n", memStats.HeapInuse/1024)
	fmt.Println("  --- GC ---")
	fmt.Printf("  NumGC:         %d\n", memStats.NumGC)
	fmt.Printf("  GCSys:         %d KB\n", memStats.GCSys/1024)
	if memStats.NumGC > 0 {
		lastGC := time.Unix(0, int64(memStats.LastGC))
		fmt.Printf("  LastGC:        %s\n", lastGC.Format(time.RFC3339))
	}

	fmt.Println("\n=== FIN CAPITULO 061 ===")
}

func allocateAndRelease() {
	data := make([][]byte, 1000)
	for i := range data {
		data[i] = make([]byte, 1024)
	}
	runtime.GC()
}

type SystemInfo struct {
	Hostname string    `json:"hostname"`
	OS       string    `json:"os"`
	Arch     string    `json:"arch"`
	CPUs     int       `json:"cpus"`
	Started  time.Time `json:"started"`
}

func (s *SystemInfo) String() string {
	b, _ := json.Marshal(s)
	return string(b)
}

type UptimeVar struct {
	start time.Time
}

func (u *UptimeVar) String() string {
	return fmt.Sprintf("%.0f", time.Since(u.start).Seconds())
}
