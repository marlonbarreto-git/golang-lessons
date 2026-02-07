// Package main - Chapter 035: Concurrency Patterns
// Go tiene patrones idiomáticos para resolver problemas comunes
// de concurrencia de manera elegante y segura.
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================
// TIPOS AUXILIARES
// ============================================

type Job struct {
	ID   int
	Data string
}

type Result struct {
	JobID  int
	Output string
}

type errGroup struct {
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

func (g *errGroup) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
			})
		}
	}()
}

func (g *errGroup) Wait() error {
	g.wg.Wait()
	return g.err
}

type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan string
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]chan string),
	}
}

func (ps *PubSub) Subscribe(topic string) <-chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan string, 10)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

func (ps *PubSub) Publish(topic, msg string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	for _, ch := range ps.subscribers[topic] {
		select {
		case ch <- msg:
		default:
		}
	}
}

type CircuitBreaker struct {
	mu          sync.Mutex
	failures    int
	threshold   int
	lastFailure time.Time
	timeout     time.Duration
	state       string
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
		state:     "closed",
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "open" {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = "half-open"
		} else {
			return errors.New("circuit breaker is open")
		}
	}

	err := fn()
	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.threshold {
			cb.state = "open"
		}
		return err
	}

	cb.failures = 0
	cb.state = "closed"
	return nil
}

func main() {
	fmt.Println("=== PATRONES DE CONCURRENCIA ===")

	// ============================================
	// FAN-OUT / FAN-IN
	// ============================================
	fmt.Println("\n--- Fan-Out / Fan-In ---")

	fanOutFanIn := func() {
		jobs := make(chan int, 10)
		results := make(chan int, 10)

		var wg sync.WaitGroup
		for w := 1; w <= 3; w++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for job := range jobs {
					time.Sleep(10 * time.Millisecond)
					results <- job * 2
				}
			}(w)
		}

		for j := 1; j <= 9; j++ {
			jobs <- j
		}
		close(jobs)

		go func() {
			wg.Wait()
			close(results)
		}()

		for result := range results {
			fmt.Printf("Result: %d\n", result)
		}
	}
	fanOutFanIn()

	// ============================================
	// PIPELINE
	// ============================================
	fmt.Println("\n--- Pipeline ---")

	generate := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			for _, n := range nums {
				out <- n
			}
			close(out)
		}()
		return out
	}

	square := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n * n
			}
			close(out)
		}()
		return out
	}

	filterEven := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				if n%2 == 0 {
					out <- n
				}
			}
			close(out)
		}()
		return out
	}

	nums := generate(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	squared := square(nums)
	evens := filterEven(squared)

	fmt.Print("Pipeline result: ")
	for n := range evens {
		fmt.Printf("%d ", n)
	}
	fmt.Println()

	// ============================================
	// WORKER POOL
	// ============================================
	fmt.Println("\n--- Worker Pool ---")

	workerPool := func(numWorkers int, jobs <-chan Job) <-chan Result {
		results := make(chan Result)

		var wg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for job := range jobs {
					time.Sleep(20 * time.Millisecond)
					results <- Result{
						JobID:  job.ID,
						Output: fmt.Sprintf("Worker %d processed: %s", workerID, job.Data),
					}
				}
			}(i)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		return results
	}

	jobsChan := make(chan Job, 5)
	for i := 1; i <= 5; i++ {
		jobsChan <- Job{ID: i, Data: fmt.Sprintf("task-%d", i)}
	}
	close(jobsChan)

	for result := range workerPool(3, jobsChan) {
		fmt.Printf("  %s\n", result.Output)
	}

	// ============================================
	// SEMAPHORE
	// ============================================
	fmt.Println("\n--- Semaphore ---")

	semaphore := make(chan struct{}, 3)

	var semWg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		semWg.Add(1)
		go func(id int) {
			defer semWg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("Task %d running (concurrent limit: 3)\n", id)
			time.Sleep(30 * time.Millisecond)
		}(i)
	}
	semWg.Wait()

	// ============================================
	// OR-DONE CHANNEL
	// ============================================
	fmt.Println("\n--- Or-Done Channel ---")

	orDone := func(done <-chan struct{}, c <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if !ok {
						return
					}
					select {
					case out <- v:
					case <-done:
						return
					}
				}
			}
		}()
		return out
	}

	done := make(chan struct{})
	values := make(chan int)

	go func() {
		for i := 0; i < 10; i++ {
			values <- i
			time.Sleep(10 * time.Millisecond)
		}
		close(values)
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(done)
	}()

	for v := range orDone(done, values) {
		fmt.Printf("Received: %d\n", v)
	}

	// ============================================
	// GENERATOR
	// ============================================
	fmt.Println("\n--- Generator ---")

	fibonacci := func(n int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			a, b := 0, 1
			for i := 0; i < n; i++ {
				out <- a
				a, b = b, a+b
			}
		}()
		return out
	}

	fmt.Print("Fibonacci: ")
	for n := range fibonacci(10) {
		fmt.Printf("%d ", n)
	}
	fmt.Println()

	// ============================================
	// RATE LIMITER
	// ============================================
	fmt.Println("\n--- Rate Limiter ---")

	rateLimiter := time.NewTicker(50 * time.Millisecond)
	defer rateLimiter.Stop()

	requests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		requests <- i
	}
	close(requests)

	for req := range requests {
		<-rateLimiter.C
		fmt.Printf("Request %d at %v\n", req, time.Now().Format("15:04:05.000"))
	}

	// ============================================
	// ERRGROUP PATTERN
	// ============================================
	fmt.Println("\n--- ErrGroup Pattern ---")

	eg := &errGroup{}
	eg.Go(func() error {
		return nil
	})
	eg.Go(func() error {
		return errors.New("task 2 failed")
	})
	eg.Go(func() error {
		return nil
	})

	if err := eg.Wait(); err != nil {
		fmt.Printf("ErrGroup error: %v\n", err)
	}

	// ============================================
	// PUBSUB
	// ============================================
	fmt.Println("\n--- PubSub ---")

	ps := NewPubSub()
	sub1 := ps.Subscribe("news")
	sub2 := ps.Subscribe("news")

	ps.Publish("news", "Breaking: Go 1.26 released!")

	fmt.Printf("Sub1 received: %s\n", <-sub1)
	fmt.Printf("Sub2 received: %s\n", <-sub2)

	// ============================================
	// TIMEOUT PATTERN
	// ============================================
	fmt.Println("\n--- Timeout Pattern ---")

	doWork := func(ctx context.Context) (string, error) {
		select {
		case <-time.After(100 * time.Millisecond):
			return "work completed", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := doWork(ctx)
	if err != nil {
		fmt.Printf("Timeout: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}

	// ============================================
	// CIRCUIT BREAKER
	// ============================================
	fmt.Println("\n--- Circuit Breaker ---")

	cb := NewCircuitBreaker(3, time.Second)

	for i := 0; i < 5; i++ {
		err := cb.Execute(func() error {
			if rand.Float32() < 0.7 {
				return errors.New("service unavailable")
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Call %d: %v (state: %s)\n", i+1, err, cb.state)
		} else {
			fmt.Printf("Call %d: success\n", i+1)
		}
	}
}

/*
RESUMEN DE PATRONES:

FAN-OUT / FAN-IN:
- Distribuir trabajo a múltiples workers
- Combinar resultados de múltiples sources

PIPELINE:
- Cadena de stages conectados por channels

WORKER POOL:
- N workers procesando jobs de un channel

SEMAPHORE:
sem := make(chan struct{}, N)
- Limitar concurrencia máxima

OR-DONE:
- Propagar cancelación a través de channels

GENERATOR:
- Producir valores bajo demanda

RATE LIMITER:
- Controlar velocidad de operaciones

ERRGROUP:
- Ejecutar tareas concurrentes con manejo de errores

PUBSUB:
- Publicar mensajes a múltiples subscribers

TIMEOUT:
- Limitar tiempo de operaciones con context

CIRCUIT BREAKER:
- Evitar llamadas a servicios fallidos
*/

/*
SUMMARY - CHAPTER 035: CONCURRENCY PATTERNS

FAN-OUT / FAN-IN:
- Fan-out: distribute work to multiple workers
- Fan-in: merge results from multiple channels
- Scale processing with multiple goroutines

PIPELINE:
- Chain processing stages via channels
- Each stage reads from input, writes to output
- Composable and testable

WORKER POOL:
- Fixed number of workers consuming from job channel
- Control concurrency and resource usage
- Close jobs channel to shut down workers

SEMAPHORE:
- Buffered channel limits concurrent operations
- Send to acquire, receive to release
- Prevents resource exhaustion

OR-DONE CHANNEL:
- Wrap channel operations with done signal
- Propagate cancellation through pipeline
- Prevents goroutine leaks

GENERATOR:
- Function returns channel producing values
- Lazy evaluation, on-demand production
- Close channel when done

RATE LIMITER:
- time.Ticker to control operation rate
- Prevents overwhelming external services
- Burstable with buffered channel

ERRGROUP:
- Run tasks concurrently, return first error
- WaitGroup + error handling combined
- Stop on first error or collect all

PUBSUB:
- Broadcast messages to multiple subscribers
- Each subscriber gets own channel
- Non-blocking sends with select + default

TIMEOUT PATTERN:
- context.WithTimeout for operation limits
- Check ctx.Done() in loops
- Clean resource cleanup on timeout

CIRCUIT BREAKER:
- Prevent calls to failing services
- States: closed, open, half-open
- Fail fast during outages
*/
