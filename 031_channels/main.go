// Package main - Chapter 031: Channels
// Los channels son el mecanismo de comunicación entre goroutines.
// Aprenderás unbuffered, buffered, direccionalidad, y patrones comunes.
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHANNELS EN GO ===")

	// ============================================
	// CHANNEL BÁSICO (UNBUFFERED)
	// ============================================
	fmt.Println("\n--- Channel Unbuffered ---")

	// Crear channel
	ch := make(chan string)

	// Enviar en goroutine (unbuffered bloquea hasta que alguien reciba)
	go func() {
		ch <- "Hola desde goroutine" // enviar
	}()

	mensaje := <-ch // recibir (bloquea hasta que haya mensaje)
	fmt.Println(mensaje)

	// ============================================
	// CHANNEL BUFFERED
	// ============================================
	fmt.Println("\n--- Channel Buffered ---")

	// Channel con buffer de tamaño 3
	buffered := make(chan int, 3)

	// Enviar no bloquea mientras haya espacio en el buffer
	buffered <- 1
	buffered <- 2
	buffered <- 3
	// buffered <- 4 // esto bloquearía (buffer lleno)

	fmt.Printf("Buffer len: %d, cap: %d\n", len(buffered), cap(buffered))

	// Recibir
	fmt.Println(<-buffered) // 1
	fmt.Println(<-buffered) // 2
	fmt.Println(<-buffered) // 3

	// ============================================
	// CERRAR CHANNELS
	// ============================================
	fmt.Println("\n--- Cerrar Channels ---")

	ch2 := make(chan int, 5)

	// Enviar datos
	for i := 1; i <= 5; i++ {
		ch2 <- i
	}
	close(ch2) // cerrar cuando ya no enviarás más

	// Range sobre channel - termina cuando se cierra
	fmt.Print("Valores: ")
	for v := range ch2 {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	// Verificar si channel está cerrado
	ch3 := make(chan string, 1)
	ch3 <- "dato"
	close(ch3)

	v1, ok1 := <-ch3
	fmt.Printf("v1=%q, ok=%t (dato recibido)\n", v1, ok1)

	v2, ok2 := <-ch3
	fmt.Printf("v2=%q, ok=%t (channel cerrado)\n", v2, ok2)

	// ============================================
	// DIRECCIONALIDAD DE CHANNELS
	// ============================================
	fmt.Println("\n--- Direccionalidad ---")

	ch4 := make(chan int)

	// Channel solo-envío: chan<- T
	// Channel solo-recepción: <-chan T

	go productor(ch4)   // solo puede enviar
	consumidor(ch4)     // solo puede recibir

	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrón: Done Channel ---")

	done := make(chan struct{}) // struct{} no ocupa memoria

	go func() {
		fmt.Println("Trabajando...")
		time.Sleep(50 * time.Millisecond)
		close(done) // señalar que terminó
	}()

	<-done // esperar señal
	fmt.Println("Trabajo completado")

	// ============================================
	// PATRÓN: GENERATOR
	// ============================================
	fmt.Println("\n--- Patrón: Generator ---")

	numeros := generador(1, 2, 3, 4, 5)
	for n := range numeros {
		fmt.Printf("%d ", n)
	}
	fmt.Println()

	// ============================================
	// PATRÓN: PIPELINE
	// ============================================
	fmt.Println("\n--- Patrón: Pipeline ---")

	// Etapa 1: generar números
	nums := generador(1, 2, 3, 4, 5)

	// Etapa 2: cuadrado
	cuadrados := cuadrado(nums)

	// Etapa 3: sumar 10
	sumados := sumar10(cuadrados)

	// Consumir
	fmt.Print("Pipeline (x² + 10): ")
	for v := range sumados {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	// ============================================
	// PATRÓN: FAN-OUT
	// ============================================
	fmt.Println("\n--- Patrón: Fan-Out ---")

	input := generador(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// Múltiples workers procesan del mismo channel
	workers := 3
	results := make([]<-chan int, workers)

	for i := 0; i < workers; i++ {
		results[i] = workerFanOut(i, input)
	}

	// Fan-in: combinar resultados
	merged := fanIn(results...)

	fmt.Print("Fan-out/Fan-in results: ")
	for v := range merged {
		fmt.Printf("%d ", v)
	}
	fmt.Println()

	// ============================================
	// PATRÓN: SEMÁFORO
	// ============================================
	fmt.Println("\n--- Patrón: Semáforo ---")

	// Limitar concurrencia con buffered channel
	maxConcurrent := 3
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			sem <- struct{}{}        // adquirir (bloquea si lleno)
			defer func() { <-sem }() // liberar

			fmt.Printf("Worker %d ejecutando (máx %d concurrentes)\n", id, maxConcurrent)
			time.Sleep(20 * time.Millisecond)
		}(i)
	}

	wg.Wait()

	// ============================================
	// PATRÓN: TIMEOUT CON CHANNEL
	// ============================================
	fmt.Println("\n--- Patrón: Timeout ---")

	resultado := make(chan string, 1)

	go func() {
		time.Sleep(200 * time.Millisecond) // simular trabajo lento
		resultado <- "completado"
	}()

	select {
	case r := <-resultado:
		fmt.Println("Resultado:", r)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Timeout!")
	}

	// ============================================
	// NIL CHANNELS
	// ============================================
	fmt.Println("\n--- Nil Channels ---")
	fmt.Println(`
Comportamiento de nil channels:
- Enviar a nil channel: bloquea forever
- Recibir de nil channel: bloquea forever
- Cerrar nil channel: PANIC

Uso útil:
- En select, un nil channel nunca está listo
- Permite "desactivar" un case dinámicamente`)
	// ============================================
	// ERRORES COMUNES
	// ============================================
	fmt.Println("\n--- Errores Comunes ---")
	fmt.Println(`
1. Enviar a channel cerrado → PANIC
   close(ch)
   ch <- valor // panic!

2. Cerrar channel dos veces → PANIC
   close(ch)
   close(ch) // panic!

3. Deadlock - todos bloqueados esperando
   ch := make(chan int)
   ch <- 1 // deadlock: nadie recibe

4. Goroutine leak - goroutine esperando forever
   ch := make(chan int)
   go func() {
       <-ch // nunca recibe si nadie envía
   }()`)
	// ============================================
	// EJEMPLO: BROADCAST
	// ============================================
	fmt.Println("\n--- Patrón: Broadcast ---")

	broadcaster := NewBroadcaster[string]()

	// Crear suscriptores
	sub1 := broadcaster.Subscribe()
	sub2 := broadcaster.Subscribe()
	sub3 := broadcaster.Subscribe()

	// Enviar mensaje broadcast
	var wg2 sync.WaitGroup
	wg2.Add(3)

	go func() { defer wg2.Done(); fmt.Printf("Sub1 recibió: %s\n", <-sub1) }()
	go func() { defer wg2.Done(); fmt.Printf("Sub2 recibió: %s\n", <-sub2) }()
	go func() { defer wg2.Done(); fmt.Printf("Sub3 recibió: %s\n", <-sub3) }()

	broadcaster.Broadcast("¡Mensaje para todos!")
	wg2.Wait()

	broadcaster.Close()
}

// ============================================
// FUNCIONES Y TIPOS
// ============================================

// Direccionalidad
func productor(ch chan<- int) {
	for i := 1; i <= 3; i++ {
		ch <- i
	}
	close(ch)
}

func consumidor(ch <-chan int) {
	for v := range ch {
		fmt.Printf("Recibido: %d\n", v)
	}
}

// Generator pattern
func generador(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// Pipeline stages
func cuadrado(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func sumar10(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n + 10
		}
	}()
	return out
}

// Fan-out worker
func workerFanOut(id int, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * 2
		}
	}()
	return out
}

// Fan-in: merge multiple channels
func fanIn(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Broadcaster pattern
type Broadcaster[T any] struct {
	mu      sync.RWMutex
	subs    []chan T
	closed  bool
}

func NewBroadcaster[T any]() *Broadcaster[T] {
	return &Broadcaster[T]{
		subs: make([]chan T, 0),
	}
}

func (b *Broadcaster[T]) Subscribe() <-chan T {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan T, 1)
	b.subs = append(b.subs, ch)
	return ch
}

func (b *Broadcaster[T]) Broadcast(msg T) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subs {
		select {
		case ch <- msg:
		default:
			// skip si el buffer está lleno
		}
	}
}

func (b *Broadcaster[T]) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.closed = true
		for _, ch := range b.subs {
			close(ch)
		}
	}
}

/*
RESUMEN:

CREAR CHANNELS:
ch := make(chan T)       // unbuffered
ch := make(chan T, n)    // buffered con capacidad n

OPERACIONES:
ch <- valor              // enviar
valor := <-ch            // recibir
valor, ok := <-ch        // recibir con verificación
close(ch)                // cerrar
for v := range ch { }    // iterar hasta cerrar

DIRECCIONALIDAD:
chan T                   // bidireccional
chan<- T                 // solo envío
<-chan T                 // solo recepción

PATRONES:
- Done channel: señalizar terminación
- Generator: producir secuencia de valores
- Pipeline: procesar en etapas
- Fan-out: distribuir trabajo
- Fan-in: combinar resultados
- Semáforo: limitar concurrencia
- Timeout: operación con tiempo límite

BUENAS PRÁCTICAS:

1. El productor cierra el channel, nunca el consumidor
2. Usa channels direccionales en signatures de funciones
3. Usa buffered channels para desacoplar productor/consumidor
4. struct{} para channels de señalización (0 bytes)
5. Verifica ok para saber si channel está cerrado
6. Evita cerrar channels si no es necesario (GC lo maneja)
*/
