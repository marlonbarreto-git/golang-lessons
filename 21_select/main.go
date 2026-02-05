// Package main - Capítulo 21: Select
// select permite esperar en múltiples operaciones de channel.
// Aprenderás multiplexación, timeouts, default, y patrones avanzados.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("=== SELECT EN GO ===")

	// ============================================
	// SELECT BÁSICO
	// ============================================
	fmt.Println("\n--- Select Básico ---")

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		time.Sleep(50 * time.Millisecond)
		ch1 <- "mensaje de ch1"
	}()

	go func() {
		time.Sleep(30 * time.Millisecond)
		ch2 <- "mensaje de ch2"
	}()

	// select espera en múltiples channels
	// ejecuta el primer case que esté listo
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Println("Recibido de ch1:", msg)
		case msg := <-ch2:
			fmt.Println("Recibido de ch2:", msg)
		}
	}

	// ============================================
	// SELECT CON DEFAULT
	// ============================================
	fmt.Println("\n--- Select con Default ---")

	ch := make(chan int, 1)

	// Non-blocking receive
	select {
	case v := <-ch:
		fmt.Printf("Recibido: %d\n", v)
	default:
		fmt.Println("No hay nada que recibir (non-blocking)")
	}

	// Non-blocking send
	ch <- 42
	select {
	case ch <- 100:
		fmt.Println("Enviado 100")
	default:
		fmt.Println("Buffer lleno, no se pudo enviar (non-blocking)")
	}

	// ============================================
	// TIMEOUT
	// ============================================
	fmt.Println("\n--- Timeout ---")

	resultado := make(chan string, 1)

	go func() {
		time.Sleep(200 * time.Millisecond) // operación lenta
		resultado <- "completado"
	}()

	select {
	case r := <-resultado:
		fmt.Println("Resultado:", r)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("Timeout: operación tardó demasiado")
	}

	// ============================================
	// TICKER Y TIME.AFTER
	// ============================================
	fmt.Println("\n--- Ticker ---")

	ticker := time.NewTicker(50 * time.Millisecond)
	done := make(chan bool)

	go func() {
		time.Sleep(180 * time.Millisecond)
		done <- true
	}()

	count := 0
loop:
	for {
		select {
		case t := <-ticker.C:
			count++
			fmt.Printf("Tick %d at %v\n", count, t.Format("15:04:05.000"))
		case <-done:
			ticker.Stop()
			fmt.Println("Ticker detenido")
			break loop
		}
	}

	// ============================================
	// SELECT EN LOOP (Event Loop)
	// ============================================
	fmt.Println("\n--- Event Loop ---")

	events := make(chan string, 10)
	quit := make(chan struct{})

	// Productor de eventos
	go func() {
		eventos := []string{"click", "scroll", "keypress", "resize"}
		for i := 0; i < 5; i++ {
			time.Sleep(30 * time.Millisecond)
			events <- eventos[rand.Intn(len(eventos))]
		}
		close(quit)
	}()

	// Event loop
eventLoop:
	for {
		select {
		case event := <-events:
			fmt.Printf("Procesando evento: %s\n", event)
		case <-quit:
			fmt.Println("Event loop terminado")
			break eventLoop
		}
	}

	// ============================================
	// SELECT CON CONTEXT
	// ============================================
	fmt.Println("\n--- Select con Context ---")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	results := make(chan int, 1)

	go func() {
		// Simular trabajo
		time.Sleep(50 * time.Millisecond)
		results <- 42
	}()

	select {
	case r := <-results:
		fmt.Printf("Resultado: %d\n", r)
	case <-ctx.Done():
		fmt.Printf("Context cancelado: %v\n", ctx.Err())
	}

	// ============================================
	// SELECT RANDOM
	// ============================================
	fmt.Println("\n--- Select Random ---")

	// Si múltiples cases están listos, select elige uno al azar
	c1 := make(chan int, 1)
	c2 := make(chan int, 1)
	c1 <- 1
	c2 <- 2

	conteo := map[string]int{"c1": 0, "c2": 0}

	for i := 0; i < 100; i++ {
		c1 <- 1
		c2 <- 2

		select {
		case <-c1:
			conteo["c1"]++
		case <-c2:
			conteo["c2"]++
		}
	}

	fmt.Printf("Distribución: c1=%d, c2=%d (aprox. 50/50)\n", conteo["c1"], conteo["c2"])

	// ============================================
	// PATRÓN: FAN-IN CON SELECT
	// ============================================
	fmt.Println("\n--- Fan-In con Select ---")

	input1 := generadorNumeros("A", 3)
	input2 := generadorNumeros("B", 3)

	merged := fanInSelect(input1, input2)

	for v := range merged {
		fmt.Println(v)
	}

	// ============================================
	// PATRÓN: FIRST RESPONSE (RACING)
	// ============================================
	fmt.Println("\n--- First Response (Racing) ---")

	servers := []string{"server1", "server2", "server3"}
	response := firstResponse(servers)
	fmt.Printf("Primera respuesta: %s\n", response)

	// ============================================
	// NIL CHANNEL EN SELECT
	// ============================================
	fmt.Println("\n--- Nil Channel en Select ---")

	// Un nil channel nunca está listo
	// Útil para "desactivar" un case dinámicamente

	var chA, chB chan int
	chA = make(chan int, 1)
	chB = nil // desactivado

	chA <- 42

	select {
	case v := <-chA:
		fmt.Printf("De chA: %d\n", v)
	case v := <-chB:
		fmt.Printf("De chB: %d\n", v) // nunca se ejecuta
	default:
		fmt.Println("Ninguno listo")
	}

	// ============================================
	// PATRÓN: HEARTBEAT
	// ============================================
	fmt.Println("\n--- Patrón: Heartbeat ---")

	heartbeat, results2 := workerConHeartbeat(100 * time.Millisecond)

	// Monitorear worker
	timeout := time.After(500 * time.Millisecond)
	heartbeatCount := 0

heartbeatLoop:
	for {
		select {
		case <-heartbeat:
			heartbeatCount++
			fmt.Printf("💓 Heartbeat #%d\n", heartbeatCount)
		case r := <-results2:
			fmt.Printf("Resultado: %d\n", r)
			break heartbeatLoop
		case <-timeout:
			fmt.Println("Worker no responde (timeout)")
			break heartbeatLoop
		}
	}

	// ============================================
	// SELECT VACÍO
	// ============================================
	fmt.Println("\n--- Select Vacío ---")
	fmt.Println(`
select {}
Bloquea forever. Útil cuando:
- El programa principal solo sirve para mantener goroutines vivas
- En servers que deben correr indefinidamente`)
	// ============================================
	// BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Best Practices ---")
	fmt.Println(`
1. Siempre incluir timeout o context.Done() para operaciones que pueden bloquear
2. Usar default para operaciones non-blocking
3. Cuidado con goroutine leaks: asegurar que los channels se cierren
4. nil channels son útiles para desactivar cases dinámicamente
5. time.After crea un timer cada vez - para loops usar time.NewTicker
6. Recordar que select elige al azar entre cases listos`)}

// ============================================
// FUNCIONES AUXILIARES
// ============================================

func generadorNumeros(prefix string, count int) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for i := 1; i <= count; i++ {
			time.Sleep(10 * time.Millisecond)
			out <- fmt.Sprintf("%s-%d", prefix, i)
		}
	}()
	return out
}

func fanInSelect(ch1, ch2 <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for ch1 != nil || ch2 != nil {
			select {
			case v, ok := <-ch1:
				if !ok {
					ch1 = nil // desactivar este case
					continue
				}
				out <- v
			case v, ok := <-ch2:
				if !ok {
					ch2 = nil
					continue
				}
				out <- v
			}
		}
	}()
	return out
}

func firstResponse(servers []string) string {
	results := make(chan string, len(servers))

	for _, server := range servers {
		go func(s string) {
			// Simular latencia variable
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			results <- fmt.Sprintf("Respuesta de %s", s)
		}(server)
	}

	return <-results // retorna la primera
}

func workerConHeartbeat(interval time.Duration) (<-chan struct{}, <-chan int) {
	heartbeat := make(chan struct{})
	results := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(results)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Simular trabajo con heartbeats
		for i := 0; i < 3; i++ {
			select {
			case <-ticker.C:
				select {
				case heartbeat <- struct{}{}:
				default:
				}
			}
		}

		// Enviar resultado
		results <- 42
	}()

	return heartbeat, results
}

/*
RESUMEN:

SELECT SYNTAX:
select {
case <-ch1:
    // ch1 listo para recibir
case ch2 <- valor:
    // ch2 listo para enviar
case <-time.After(timeout):
    // timeout
case <-ctx.Done():
    // context cancelado
default:
    // ningún case listo (non-blocking)
}

COMPORTAMIENTO:
- Bloquea hasta que un case esté listo
- Si múltiples están listos, elige al azar
- default se ejecuta si ningún case está listo
- nil channel nunca está listo

PATRONES:
- Timeout: time.After(duration)
- Context cancellation: ctx.Done()
- Non-blocking: default
- Event loop: for { select { } }
- Fan-in: merge channels
- Racing: first response wins
- Heartbeat: monitorear workers

BUENAS PRÁCTICAS:

1. Siempre tener salida (timeout, context, quit channel)
2. Usar Ticker para loops, After para one-shot
3. Nil channels para desactivar cases
4. Cerrar channels para terminar range/select
5. Buffered channels para evitar bloqueos
*/
