// Package main - Chapter 030: Goroutines
// Las goroutines son la base de la concurrencia en Go.
// Aprenderás cómo crearlas, sincronizarlas, y evitar problemas comunes.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== GOROUTINES EN GO ===")

	// ============================================
	// GOROUTINE BÁSICA
	// ============================================
	fmt.Println("\n--- Goroutine Básica ---")

	// go lanza una función en una goroutine separada
	go saludar("Gopher")
	go saludar("Mundo")

	// Sin esperar, main termina antes que las goroutines
	time.Sleep(100 * time.Millisecond)

	// ============================================
	// GOROUTINE ANÓNIMA
	// ============================================
	fmt.Println("\n--- Goroutine Anónima ---")

	go func() {
		fmt.Println("Goroutine anónima ejecutándose")
	}()

	// Con parámetros (importante: pasar por valor)
	mensaje := "Hola desde goroutine"
	go func(msg string) {
		fmt.Println(msg)
	}(mensaje) // ← pasar el valor actual

	time.Sleep(50 * time.Millisecond)

	// ============================================
	// PROBLEMA: VARIABLE CAPTURADA
	// ============================================
	fmt.Println("\n--- Problema: Variable Capturada ---")

	// MAL: la variable i se comparte entre goroutines
	fmt.Println("MAL (variable capturada):")
	for i := 0; i < 3; i++ {
		go func() {
			fmt.Printf("  i = %d\n", i) // probablemente imprime 3,3,3
		}()
	}
	time.Sleep(50 * time.Millisecond)

	// BIEN: pasar la variable como parámetro
	fmt.Println("BIEN (pasada como parámetro):")
	for i := 0; i < 3; i++ {
		go func(n int) {
			fmt.Printf("  n = %d\n", n) // imprime 0,1,2 (en algún orden)
		}(i)
	}
	time.Sleep(50 * time.Millisecond)

	// Go 1.22+ fix: cada iteración tiene su propia variable
	fmt.Println("Go 1.22+ (loop variable fix):")
	for i := 0; i < 3; i++ {
		go func() {
			fmt.Printf("  i = %d\n", i) // ahora funciona correctamente
		}()
	}
	time.Sleep(50 * time.Millisecond)

	// ============================================
	// WAITGROUP
	// ============================================
	fmt.Println("\n--- WaitGroup ---")

	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1) // incrementar contador ANTES de lanzar goroutine

		go func(n int) {
			defer wg.Done() // decrementar cuando termine
			fmt.Printf("Worker %d ejecutándose\n", n)
			time.Sleep(time.Duration(n*10) * time.Millisecond)
			fmt.Printf("Worker %d terminó\n", n)
		}(i)
	}

	wg.Wait() // bloquea hasta que el contador llegue a 0
	fmt.Println("Todos los workers terminaron")

	// ============================================
	// WAITGROUP.GO (Go 1.25+)
	// ============================================
	fmt.Println("\n--- WaitGroup.Go (Go 1.25+) ---")

	var wg2 sync.WaitGroup

	// Nuevo método que combina Add(1) + go + defer Done()
	for i := 1; i <= 3; i++ {
		wg2.Go(func() {
			fmt.Printf("Worker (Go method) ejecutándose\n")
		})
	}

	wg2.Wait()
	fmt.Println("Todos terminaron con wg.Go()")

	// ============================================
	// NÚMERO DE GOROUTINES
	// ============================================
	fmt.Println("\n--- Información de Runtime ---")

	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
	fmt.Printf("NumGoroutine: %d\n", runtime.NumGoroutine())

	// Lanzar varias goroutines
	for i := 0; i < 10; i++ {
		go func() {
			time.Sleep(100 * time.Millisecond)
		}()
	}
	fmt.Printf("NumGoroutine (después de lanzar 10): %d\n", runtime.NumGoroutine())

	// ============================================
	// RACE CONDITIONS
	// ============================================
	fmt.Println("\n--- Race Conditions ---")

	// PROBLEMA: múltiples goroutines modificando sin sincronización
	contador := 0
	var wg3 sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg3.Add(1)
		go func() {
			defer wg3.Done()
			contador++ // DATA RACE: lectura-modificación-escritura no atómica
		}()
	}

	wg3.Wait()
	fmt.Printf("Contador (con race): %d (esperado 1000)\n", contador)

	// SOLUCIÓN 1: Mutex
	contador = 0
	var mu sync.Mutex
	var wg4 sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg4.Add(1)
		go func() {
			defer wg4.Done()
			mu.Lock()
			contador++
			mu.Unlock()
		}()
	}

	wg4.Wait()
	fmt.Printf("Contador (con Mutex): %d\n", contador)

	// SOLUCIÓN 2: Atomic
	var contadorAtomic int64
	var wg5 sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg5.Add(1)
		go func() {
			defer wg5.Done()
			atomic.AddInt64(&contadorAtomic, 1)
		}()
	}

	wg5.Wait()
	fmt.Printf("Contador (atomic): %d\n", contadorAtomic)

	// ============================================
	// GOROUTINE LEAKS
	// ============================================
	fmt.Println("\n--- Goroutine Leaks ---")
	fmt.Println(`
Una goroutine "leak" ocurre cuando:
- La goroutine espera en un channel que nunca recibe
- La goroutine espera un recurso que nunca se libera
- No hay forma de señalar que debe terminar

Prevención:
- Siempre tener una forma de señalar terminación (context, done channel)
- Cerrar channels cuando ya no se usarán
- Usar select con timeout o default`)
	// Ejemplo de goroutine que puede terminar
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				fmt.Println("Goroutine terminando limpiamente")
				return
			default:
				// trabajo...
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)
	close(done) // señalar que debe terminar
	time.Sleep(20 * time.Millisecond)

	// ============================================
	// PATRÓN: WORKER POOL SIMPLE
	// ============================================
	fmt.Println("\n--- Worker Pool Simple ---")

	jobs := make(chan int, 5)
	results := make(chan int, 5)
	var wgWorkers sync.WaitGroup

	// Lanzar workers
	for w := 1; w <= 3; w++ {
		wgWorkers.Add(1)
		go worker(w, jobs, results, &wgWorkers)
	}

	// Enviar jobs
	for j := 1; j <= 5; j++ {
		jobs <- j
	}
	close(jobs)

	// Esperar que terminen y cerrar results
	go func() {
		wgWorkers.Wait()
		close(results)
	}()

	// Recoger resultados
	for r := range results {
		fmt.Printf("Resultado: %d\n", r)
	}

	// ============================================
	// SCHEDULER DE GO
	// ============================================
	fmt.Println("\n--- Scheduler de Go ---")
	fmt.Println(`
El scheduler de Go usa un modelo M:N:
- M = OS threads
- N = Goroutines
- P = Processors (GOMAXPROCS)

Características:
- Goroutines son muy livianas (~2KB stack inicial)
- El stack crece/decrece dinámicamente
- El scheduler es preemptivo (Go 1.14+)
- Goroutines ceden el control en operaciones de I/O, llamadas al runtime, etc.

runtime.Gosched() - cede el control voluntariamente
runtime.GOMAXPROCS(n) - establece número de processors`)
	// ============================================
	// COMPARACIÓN CON THREADS
	// ============================================
	fmt.Println("\n--- Goroutines vs Threads ---")
	fmt.Println(`
Goroutines                    | OS Threads
------------------------------|---------------------------
~2KB stack inicial            | ~1MB stack fijo
Manejadas por Go runtime      | Manejadas por OS
Muy bajo costo de creación    | Alto costo de creación
Millones posibles             | Miles como máximo
Comunicación por channels     | Locks y memoria compartida
Scheduler cooperativo/preempt | Scheduler preemptivo OS`)}

func saludar(nombre string) {
	fmt.Printf("Hola, %s!\n", nombre)
}

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		fmt.Printf("Worker %d procesando job %d\n", id, j)
		time.Sleep(20 * time.Millisecond) // simular trabajo
		results <- j * 2
	}
}

/*
RESUMEN:

CREAR GOROUTINE:
go funcion()
go func() { }()
go func(param tipo) { }(valor)

SINCRONIZACIÓN:
- sync.WaitGroup: esperar múltiples goroutines
- sync.Mutex: exclusión mutua
- sync/atomic: operaciones atómicas
- channels: comunicación (siguiente capítulo)

WaitGroup API:
- wg.Add(n): incrementar contador
- wg.Done(): decrementar contador
- wg.Wait(): bloquear hasta contador = 0
- wg.Go(fn): Add(1) + go fn() + defer Done() (Go 1.25+)

BUENAS PRÁCTICAS:

1. Pasar variables por parámetro a goroutines anónimas
2. Siempre poder señalar a una goroutine que debe terminar
3. Usar -race flag para detectar data races: go run -race main.go
4. Preferir channels sobre memoria compartida
5. No lanzar goroutines sin forma de esperar o terminar
6. Usar WaitGroup.Go() cuando esté disponible (Go 1.25+)

DETECTAR RACE CONDITIONS:
go run -race main.go
go test -race ./...
go build -race
*/

/*
SUMMARY - CHAPTER 030: GOROUTINES

GOROUTINE BASICS:
- Lightweight concurrent execution units (~2KB initial stack)
- Launch with go keyword
- Managed by Go runtime scheduler
- Use goroutines for concurrent operations, not parallelism by default

CREATING GOROUTINES:
- go function(): launch named function
- go func() { }(): anonymous function
- Pass variables as parameters to avoid capture issues

SYNCHRONIZATION:
- sync.WaitGroup: wait for multiple goroutines to complete
- wg.Add(n), wg.Done(), wg.Wait()
- wg.Go(fn): modern helper (Go 1.25+)

COMMON PITFALLS:
- Variable capture in loops: pass as parameter
- Race conditions: multiple goroutines accessing shared data
- Goroutine leaks: goroutines blocked forever

RACE DETECTION:
- go run -race: detect data races at runtime
- go test -race: run tests with race detector
- Use sync.Mutex or atomic for shared data

SCHEDULER:
- M:N scheduling (M OS threads, N goroutines)
- Preemptive scheduler (Go 1.14+)
- GOMAXPROCS controls parallelism level
- Goroutines yield on I/O, channel ops, runtime calls

GOROUTINES VS THREADS:
- Much lighter than OS threads
- Millions possible vs thousands
- Communicate via channels, not shared memory
*/
