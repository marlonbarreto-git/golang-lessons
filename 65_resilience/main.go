// Package main - Capitulo 65: Resilience Patterns
// Patrones de resiliencia para sistemas en produccion: rate limiting,
// circuit breakers, retries, bulkhead, timeouts y graceful degradation.
package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== RESILIENCE PATTERNS EN GO ===")

	// ============================================
	// RATE LIMITING - TOKEN BUCKET
	// ============================================
	fmt.Println("\n--- Rate Limiting: Token Bucket ---")
	fmt.Print(`
CONCEPTO:
Un "token bucket" permite un rate de requests constante con bursts.
- Se agregan tokens al bucket a una tasa fija
- Cada request consume un token
- Si no hay tokens, se deniega o espera

LIBRERIA OFICIAL: golang.org/x/time/rate

go get golang.org/x/time/rate

import "golang.org/x/time/rate"

CREAR UN LIMITER:

// 10 requests por segundo, burst de 5
limiter := rate.NewLimiter(rate.Limit(10), 5)

// 1 request cada 100ms, burst de 3
limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 3)

METODOS PRINCIPALES:

// Allow() - Non-blocking, retorna true/false
if limiter.Allow() {
    // Procesar request
} else {
    // Rate limited
}

// Wait(ctx) - Blocking, espera hasta tener token
err := limiter.Wait(ctx)
if err != nil {
    // Context cancelado o deadline
}

// Reserve() - Reserva token, retorna info de cuando estara listo
reservation := limiter.Reserve()
if !reservation.OK() {
    // No se puede satisfacer (burst insuficiente)
}
delay := reservation.Delay()
time.Sleep(delay)

DIFERENCIAS:
- Allow()   -> Para rechazar inmediatamente (APIs publicas)
- Wait()    -> Para encolar y esperar (workers internos)
- Reserve() -> Para decisiones custom (prioridades)
`)

	fmt.Println("--- Demo: Token Bucket en accion ---")
	demoTokenBucket()

	// ============================================
	// PER-CLIENT RATE LIMITING
	// ============================================
	fmt.Println("\n--- Per-Client Rate Limiting ---")
	fmt.Print(`
RATE LIMIT POR CLIENTE (IP, API key, user ID):

import (
    "sync"
    "golang.org/x/time/rate"
)

type ClientRateLimiter struct {
    mu       sync.RWMutex
    clients  map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewClientRateLimiter(r rate.Limit, burst int) *ClientRateLimiter {
    return &ClientRateLimiter{
        clients: make(map[string]*rate.Limiter),
        rate:    r,
        burst:   burst,
    }
}

func (c *ClientRateLimiter) GetLimiter(clientID string) *rate.Limiter {
    c.mu.RLock()
    limiter, exists := c.clients[clientID]
    c.mu.RUnlock()

    if exists {
        return limiter
    }

    c.mu.Lock()
    defer c.mu.Unlock()

    // Double check
    if limiter, exists = c.clients[clientID]; exists {
        return limiter
    }

    limiter = rate.NewLimiter(c.rate, c.burst)
    c.clients[clientID] = limiter
    return limiter
}

// LIMPIEZA PERIODICA (evitar memory leak):

func (c *ClientRateLimiter) Cleanup(maxAge time.Duration) {
    // Ejecutar en goroutine con ticker
    ticker := time.NewTicker(maxAge)
    for range ticker.C {
        c.mu.Lock()
        // En produccion: trackear last access time
        c.clients = make(map[string]*rate.Limiter)
        c.mu.Unlock()
    }
}
`)

	// ============================================
	// HTTP RATE LIMITING MIDDLEWARE
	// ============================================
	fmt.Println("\n--- HTTP Rate Limiting Middleware ---")
	os.Stdout.WriteString(`
MIDDLEWARE GLOBAL:

func RateLimitMiddleware(rps float64, burst int) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(rps), burst)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                w.Header().Set("Retry-After", "1")
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

MIDDLEWARE PER-CLIENT:

func PerClientRateLimit(rps float64, burst int) func(http.Handler) http.Handler {
    clients := NewClientRateLimiter(rate.Limit(rps), burst)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr // o extraer de X-Forwarded-For

            limiter := clients.GetLimiter(ip)
            if !limiter.Allow() {
                w.Header().Set("Retry-After", "1")
                w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", burst))
                w.Header().Set("X-RateLimit-Remaining", "0")
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

USO:

mux := http.NewServeMux()
mux.HandleFunc("/api/", handleAPI)
handler := PerClientRateLimit(10, 20)(mux)
http.ListenAndServe(":8080", handler)

HEADERS ESTANDAR DE RATE LIMITING:
- X-RateLimit-Limit:     requests permitidos por ventana
- X-RateLimit-Remaining: requests restantes
- X-RateLimit-Reset:     timestamp cuando se resetea
- Retry-After:           segundos para reintentar
`)

	// ============================================
	// DISTRIBUTED RATE LIMITING
	// ============================================
	fmt.Println("\n--- Distributed Rate Limiting (Redis) ---")
	os.Stdout.WriteString(`
CONCEPTO:
Cuando tienes multiples instancias del servicio,
necesitas rate limiting centralizado (Redis).

SLIDING WINDOW CON REDIS:

import "github.com/redis/go-redis/v9"

func SlidingWindowRateLimit(
    ctx context.Context,
    rdb *redis.Client,
    key string,
    limit int,
    window time.Duration,
) (bool, error) {
    now := time.Now().UnixNano()
    windowStart := now - int64(window)

    pipe := rdb.Pipeline()

    // Remover requests fuera de la ventana
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

    // Contar requests en la ventana
    countCmd := pipe.ZCard(ctx, key)

    // Agregar request actual
    pipe.ZAdd(ctx, key, redis.Z{
        Score:  float64(now),
        Member: fmt.Sprintf("%d", now),
    })

    // Expirar key
    pipe.Expire(ctx, key, window)

    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }

    return countCmd.Val() < int64(limit), nil
}

USO:

allowed, err := SlidingWindowRateLimit(
    ctx, rdb,
    fmt.Sprintf("ratelimit:%s", clientIP),
    100,           // 100 requests
    time.Minute,   // por minuto
)

ALGORITMOS DE RATE LIMITING:

1. Token Bucket     - Burst-friendly, rate constante
2. Leaky Bucket     - Salida constante, suaviza trafico
3. Fixed Window     - Simple pero boundary spikes
4. Sliding Window   - Preciso, mas memoria
5. Sliding Window Log - Muy preciso, mas costoso
`)

	// ============================================
	// SLIDING WINDOW (LOCAL)
	// ============================================
	fmt.Println("\n--- Sliding Window Algorithm ---")
	fmt.Println("Demo: Sliding Window Rate Limiter")
	demoSlidingWindow()

	// ============================================
	// CIRCUIT BREAKER
	// ============================================
	fmt.Println("\n--- Circuit Breaker ---")
	os.Stdout.WriteString(`
CONCEPTO:
El Circuit Breaker protege tu sistema de llamadas repetidas
a servicios que estan fallando.

ESTADOS:

  CLOSED ──(fallos >= threshold)──> OPEN
     ^                                 |
     |                           (timeout expira)
     |                                 |
     └──(exito)── HALF-OPEN <──────────┘

CLOSED:   Requests pasan normalmente, contando fallos
OPEN:     Requests fallan inmediatamente (fail fast)
HALF-OPEN: Permite requests limitados para probar recovery

LIBRERIA: sony/gobreaker

go get github.com/sony/gobreaker/v2

import "github.com/sony/gobreaker/v2"

CONFIGURACION:

cb := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
    Name:        "external-api",
    MaxRequests: 3,              // requests en half-open
    Interval:    10 * time.Second, // ventana de conteo (closed)
    Timeout:     30 * time.Second, // tiempo en open antes de half-open

    ReadyToTrip: func(counts gobreaker.Counts) bool {
        // Abrir si >60% de requests fallan
        failRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 3 && failRatio >= 0.6
    },

    OnStateChange: func(name string, from, to gobreaker.State) {
        log.Printf("Circuit Breaker %s: %s -> %s", name, from, to)
    },

    IsSuccessful: func(err error) bool {
        // HTTP 429 no es un fallo del servicio
        if errors.Is(err, ErrRateLimited) {
            return true
        }
        return err == nil
    },
})

USO:

result, err := cb.Execute(func() ([]byte, error) {
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 500 {
        return nil, fmt.Errorf("server error: %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
})

if err != nil {
    if errors.Is(err, gobreaker.ErrOpenState) {
        // Circuit abierto: usar fallback
    }
}

CIRCUIT BREAKER CON HTTP CLIENT:

type ResilientClient struct {
    client *http.Client
    cb     *gobreaker.CircuitBreaker[*http.Response]
}

func NewResilientClient() *ResilientClient {
    return &ResilientClient{
        client: &http.Client{Timeout: 10 * time.Second},
        cb: gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
            Name:        "http-client",
            MaxRequests: 5,
            Timeout:     30 * time.Second,
            ReadyToTrip: func(counts gobreaker.Counts) bool {
                return counts.ConsecutiveFailures > 3
            },
        }),
    }
}

func (rc *ResilientClient) Do(req *http.Request) (*http.Response, error) {
    return rc.cb.Execute(func() (*http.Response, error) {
        resp, err := rc.client.Do(req)
        if err != nil {
            return nil, err
        }
        if resp.StatusCode >= 500 {
            return nil, fmt.Errorf("server error: %d", resp.StatusCode)
        }
        return resp, nil
    })
}
`)

	fmt.Println("--- Demo: Circuit Breaker Custom ---")
	demoCircuitBreaker()

	// ============================================
	// RETRY PATTERNS
	// ============================================
	fmt.Println("\n--- Retry Patterns ---")
	os.Stdout.WriteString(`
RETRY SIMPLE CON BACKOFF:

func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
    var err error
    for attempt := 0; attempt <= maxRetries; attempt++ {
        err = fn()
        if err == nil {
            return nil
        }

        if attempt == maxRetries {
            break
        }

        // Exponential backoff: 1s, 2s, 4s, 8s...
        backoff := time.Duration(1<<uint(attempt)) * time.Second

        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(backoff):
        }
    }
    return fmt.Errorf("after %d retries: %w", maxRetries, err)
}

EXPONENTIAL BACKOFF CON JITTER:

func BackoffWithJitter(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
    // Exponential: baseDelay * 2^attempt
    delay := baseDelay * time.Duration(1<<uint(attempt))

    // Cap al maximo
    if delay > maxDelay {
        delay = maxDelay
    }

    // Full jitter: [0, delay)
    jitter := time.Duration(rand.Int64N(int64(delay)))
    return jitter
}

POR QUE JITTER?
Sin jitter: todos los clientes reintentan al mismo tiempo ("thundering herd")
Con jitter: los reintentos se distribuyen uniformemente

TIPOS DE JITTER:
1. Full Jitter:      rand(0, delay)          <- recomendado
2. Equal Jitter:     delay/2 + rand(0, delay/2)
3. Decorrelated:     min(cap, rand(base, prev*3))

LIBRERIA: cenkalti/backoff

go get github.com/cenkalti/backoff/v5

import "github.com/cenkalti/backoff/v5"

// Exponential backoff
result, err := backoff.Retry(ctx, func() (string, error) {
    resp, err := callExternalAPI()
    if err != nil {
        // Retornar error retryable
        return "", err
    }
    return resp, nil
},
    backoff.WithBackOff(backoff.NewExponentialBackOff()),
    backoff.WithMaxElapsedTime(2*time.Minute),
)

// Con politica custom
bo := backoff.NewExponentialBackOff()
bo.InitialInterval = 500 * time.Millisecond
bo.MaxInterval = 30 * time.Second
bo.Multiplier = 2.0
bo.RandomizationFactor = 0.5

result, err := backoff.Retry(ctx, operation,
    backoff.WithBackOff(bo),
)

// Permanent error (no reintentar)
if isNotFound(err) {
    return "", backoff.Permanent(err)
}

RETRY WITH CONTEXT CANCELLATION:

func RetryWithContext(ctx context.Context, fn func(ctx context.Context) error) error {
    bo := backoff.NewExponentialBackOff()
    bo.MaxElapsedTime = 0 // Sin limite (controlado por ctx)

    _, err := backoff.Retry(ctx, func() (struct{}, error) {
        if err := fn(ctx); err != nil {
            return struct{}{}, err
        }
        return struct{}{}, nil
    }, backoff.WithBackOff(bo))
    return err
}

// El context controla cuando parar
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := RetryWithContext(ctx, func(ctx context.Context) error {
    return callService(ctx)
})
`)

	fmt.Println("--- Demo: Retry con Exponential Backoff + Jitter ---")
	demoRetryWithBackoff()

	// ============================================
	// IDEMPOTENCY
	// ============================================
	fmt.Println("\n--- Idempotency en Retries ---")
	fmt.Print(`
PROBLEMA:
Si reintentamos una operacion, puede ejecutarse multiples veces.
Debemos asegurar que sea seguro reintentar (idempotente).

OPERACIONES IDEMPOTENTES:
- GET, PUT, DELETE -> Naturalmente idempotentes
- POST, PATCH     -> Requieren idempotency key

PATRON IDEMPOTENCY KEY:

type IdempotencyStore struct {
    mu      sync.RWMutex
    results map[string]*IdempotencyResult
}

type IdempotencyResult struct {
    StatusCode int
    Body       []byte
    CreatedAt  time.Time
}

func (s *IdempotencyStore) GetOrExecute(
    key string,
    fn func() (*IdempotencyResult, error),
) (*IdempotencyResult, error) {
    // Verificar si ya existe resultado
    s.mu.RLock()
    if result, ok := s.results[key]; ok {
        s.mu.RUnlock()
        return result, nil
    }
    s.mu.RUnlock()

    // Ejecutar operacion
    result, err := fn()
    if err != nil {
        return nil, err
    }

    // Guardar resultado
    s.mu.Lock()
    s.results[key] = result
    s.mu.Unlock()

    return result, nil
}

HTTP HEADER:

// Cliente envia:
// Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000

func handler(w http.ResponseWriter, r *http.Request) {
    key := r.Header.Get("Idempotency-Key")
    if key == "" {
        http.Error(w, "Idempotency-Key required", 400)
        return
    }

    result, err := store.GetOrExecute(key, func() (*IdempotencyResult, error) {
        // Procesar payment, etc.
        return processPayment(r)
    })
    // ...
}

QUE ES SEGURO REINTENTAR?
- Lectura de datos              -> Siempre
- Crear con UUID del cliente    -> Si
- Incrementar contador          -> NO (usar idempotency key)
- Transferencia de dinero       -> NO (usar idempotency key)
- Enviar email                  -> NO (deduplicar)
`)

	// ============================================
	// BULKHEAD PATTERN
	// ============================================
	fmt.Println("\n--- Bulkhead Pattern ---")
	fmt.Print(`
CONCEPTO:
Aislar recursos para que el fallo de un componente
no consuma todos los recursos del sistema.
(Como compartimentos de un barco que evitan que se hunda todo)

SEMAPHORE-BASED:

type Bulkhead struct {
    sem chan struct{}
}

func NewBulkhead(maxConcurrent int) *Bulkhead {
    return &Bulkhead{
        sem: make(chan struct{}, maxConcurrent),
    }
}

func (b *Bulkhead) Execute(ctx context.Context, fn func() error) error {
    select {
    case b.sem <- struct{}{}:
        defer func() { <-b.sem }()
        return fn()
    case <-ctx.Done():
        return fmt.Errorf("bulkhead: %w", ctx.Err())
    }
}

func (b *Bulkhead) TryExecute(fn func() error) error {
    select {
    case b.sem <- struct{}{}:
        defer func() { <-b.sem }()
        return fn()
    default:
        return errors.New("bulkhead: capacity full")
    }
}

USO CON MULTIPLES SERVICIOS:

type ServiceBulkheads struct {
    payment  *Bulkhead  // Max 10 concurrent
    shipping *Bulkhead  // Max 20 concurrent
    email    *Bulkhead  // Max 5 concurrent
}

bulkheads := &ServiceBulkheads{
    payment:  NewBulkhead(10),
    shipping: NewBulkhead(20),
    email:    NewBulkhead(5),
}

// Payment puede saturarse sin afectar shipping
err := bulkheads.payment.Execute(ctx, func() error {
    return processPayment(order)
})
`)

	fmt.Println("--- Demo: Bulkhead con Semaphore ---")
	demoBulkhead()

	// ============================================
	// WORKER POOL AS BULKHEAD
	// ============================================
	fmt.Println("\n--- Worker Pool como Bulkhead ---")
	fmt.Print(`
WORKER POOL CON BACKPRESSURE:

type WorkerPool struct {
    tasks   chan func() error
    results chan error
    wg      sync.WaitGroup
}

func NewWorkerPool(workers, queueSize int) *WorkerPool {
    wp := &WorkerPool{
        tasks:   make(chan func() error, queueSize),
        results: make(chan error, queueSize),
    }

    for i := 0; i < workers; i++ {
        wp.wg.Add(1)
        go func() {
            defer wp.wg.Done()
            for task := range wp.tasks {
                wp.results <- task()
            }
        }()
    }

    return wp
}

func (wp *WorkerPool) Submit(ctx context.Context, task func() error) error {
    select {
    case wp.tasks <- task:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (wp *WorkerPool) Shutdown() {
    close(wp.tasks)
    wp.wg.Wait()
    close(wp.results)
}
`)

	// ============================================
	// TIMEOUT PATTERNS
	// ============================================
	fmt.Println("\n--- Timeout Patterns ---")
	fmt.Print(`
CONTEXT TIMEOUTS:

// Timeout para operacion individual
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := db.QueryContext(ctx, "SELECT ...")

HTTP SERVER TIMEOUTS:

server := &http.Server{
    Addr:              ":8080",
    ReadTimeout:       5 * time.Second,   // Leer request completo
    ReadHeaderTimeout: 2 * time.Second,   // Solo headers
    WriteTimeout:      10 * time.Second,  // Escribir response
    IdleTimeout:       120 * time.Second, // Keep-alive idle
    Handler:           mux,
}

// Handler timeout (per-handler)
handler := http.TimeoutHandler(mux, 30*time.Second, "Request timeout\n")

ANATOMIA DE TIMEOUTS HTTP:

  Client ──> [ReadHeaderTimeout] ──> [ReadTimeout] ──> Handler ──> [WriteTimeout] ──> Client
                                                           |
                                                    [HandlerTimeout]

ReadTimeout:       Incluye TLS handshake + headers + body
WriteTimeout:      Desde fin de read hasta fin de write
IdleTimeout:       Entre requests en keep-alive
HandlerTimeout:    Tiempo maximo que el handler puede ejecutar

DATABASE TIMEOUTS:

import "database/sql"

// Connection timeout
db, err := sql.Open("postgres", "host=... connect_timeout=5")

// Query timeout via context
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
rows, err := db.QueryContext(ctx, "SELECT ...")

// Connection pool settings
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
`)

	// ============================================
	// CASCADING TIMEOUT BUDGETS
	// ============================================
	fmt.Println("\n--- Cascading Timeout Budgets ---")
	fmt.Print(`
CONCEPTO:
Propagar un "budget" de tiempo a traves de la cadena de llamadas.
Si el request tiene 5s, y el primer servicio tarda 2s,
el segundo servicio solo tiene 3s.

PATRON:

func HandleOrder(ctx context.Context, orderID string) error {
    // Budget total: 5s (del context del caller)

    // Paso 1: Validar inventario (max 2s)
    invCtx, cancel1 := context.WithTimeout(ctx, 2*time.Second)
    defer cancel1()
    if err := checkInventory(invCtx, orderID); err != nil {
        return err
    }

    // Paso 2: Procesar pago (usa el tiempo restante del context padre)
    // Si paso 1 tardo 1s, quedan ~4s del budget original
    if err := processPayment(ctx, orderID); err != nil {
        return err
    }

    // Paso 3: Con budget explicito
    remaining := time.Until(deadline(ctx))
    shipCtx, cancel3 := context.WithTimeout(ctx, min(remaining, 3*time.Second))
    defer cancel3()
    return scheduleShipping(shipCtx, orderID)
}

func deadline(ctx context.Context) time.Time {
    d, ok := ctx.Deadline()
    if !ok {
        return time.Now().Add(30 * time.Second) // default
    }
    return d
}

PROPAGACION ENTRE MICROSERVICIOS:

// Servicio A -> Servicio B via HTTP
func callServiceB(ctx context.Context) error {
    deadline, ok := ctx.Deadline()
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

    if ok {
        // Propagar deadline via header
        remaining := time.Until(deadline)
        req.Header.Set("X-Request-Timeout", remaining.String())
    }

    resp, err := client.Do(req)
    // ...
}

// Servicio B: leer timeout del header
func handlerB(w http.ResponseWriter, r *http.Request) {
    timeoutStr := r.Header.Get("X-Request-Timeout")
    if timeoutStr != "" {
        timeout, _ := time.ParseDuration(timeoutStr)
        // Usar un poco menos para dejar margen al response
        ctx, cancel := context.WithTimeout(r.Context(), timeout*90/100)
        defer cancel()
        r = r.WithContext(ctx)
    }
    // ...
}
`)

	fmt.Println("--- Demo: Timeout Budget ---")
	demoTimeoutBudget()

	// ============================================
	// DEADLINE PROPAGATION
	// ============================================
	fmt.Println("\n--- Deadline Propagation ---")
	os.Stdout.WriteString(`
CONCEPTO:
Cada capa de tu sistema debe respetar el deadline del caller.
Si el deadline ya paso, fail fast sin hacer trabajo.

PATRON:

func ProcessRequest(ctx context.Context, data Request) (Response, error) {
    // Fail fast si ya expiro
    if ctx.Err() != nil {
        return Response{}, ctx.Err()
    }

    // Verificar que queda tiempo suficiente
    if deadline, ok := ctx.Deadline(); ok {
        remaining := time.Until(deadline)
        if remaining < 100*time.Millisecond {
            return Response{}, fmt.Errorf("insufficient time budget: %v", remaining)
        }
    }

    // Proceder con la operacion
    return doWork(ctx, data)
}

gRPC DEADLINE PROPAGATION:

// gRPC propaga deadlines automaticamente via metadata
// Cliente:
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: "123"})

// Servidor: el ctx ya tiene el deadline del cliente
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    // ctx.Deadline() retorna el deadline del cliente
    return s.repo.FindUser(ctx, req.Id)
}

HTTP DEADLINE PROPAGATION:

// Middleware que extrae deadline de headers
func DeadlinePropagation(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if deadlineStr := r.Header.Get("X-Request-Deadline"); deadlineStr != "" {
            deadline, err := time.Parse(time.RFC3339Nano, deadlineStr)
            if err == nil && deadline.After(time.Now()) {
                ctx, cancel := context.WithDeadline(r.Context(), deadline)
                defer cancel()
                r = r.WithContext(ctx)
            }
        }
        next.ServeHTTP(w, r)
    })
}
`)

	// ============================================
	// LOAD SHEDDING
	// ============================================
	fmt.Println("\n--- Load Shedding ---")
	fmt.Print(`
CONCEPTO:
Rechazar requests proactivamente cuando el sistema esta
sobrecargado, para proteger la capacidad restante.

"Es mejor servir 80% de requests bien que 100% mal"

PATRON POR CONCURRENCIA:

type LoadShedder struct {
    inFlight atomic.Int64
    maxLoad  int64
}

func NewLoadShedder(maxConcurrent int64) *LoadShedder {
    return &LoadShedder{maxLoad: maxConcurrent}
}

func (ls *LoadShedder) Allow() (func(), bool) {
    current := ls.inFlight.Add(1)
    if current > ls.maxLoad {
        ls.inFlight.Add(-1)
        return nil, false
    }
    return func() { ls.inFlight.Add(-1) }, true
}

// Middleware
func LoadSheddingMiddleware(maxConcurrent int64) func(http.Handler) http.Handler {
    shedder := NewLoadShedder(maxConcurrent)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            done, ok := shedder.Allow()
            if !ok {
                w.Header().Set("Retry-After", "5")
                http.Error(w, "Service Overloaded", http.StatusServiceUnavailable)
                return
            }
            defer done()
            next.ServeHTTP(w, r)
        })
    }
}

PATRON POR LATENCIA (adaptive):

type AdaptiveLoadShedder struct {
    mu           sync.Mutex
    avgLatency   time.Duration
    threshold    time.Duration
    shedding     bool
}

func (als *AdaptiveLoadShedder) RecordLatency(d time.Duration) {
    als.mu.Lock()
    defer als.mu.Unlock()

    // Exponential moving average
    als.avgLatency = (als.avgLatency*9 + d) / 10

    als.shedding = als.avgLatency > als.threshold
}

func (als *AdaptiveLoadShedder) ShouldShed() bool {
    als.mu.Lock()
    defer als.mu.Unlock()
    if !als.shedding {
        return false
    }
    // Shed probabilisticamente (no todo de golpe)
    return rand.Float64() < 0.5
}

LOAD SHEDDING POR PRIORIDAD:

type Priority int
const (
    PriorityCritical Priority = iota  // Nunca shed
    PriorityHigh                       // Shed ultimo
    PriorityNormal                     // Shed cuando sobrecargado
    PriorityLow                        // Shed primero
)

func (ls *LoadShedder) AllowWithPriority(p Priority) bool {
    load := float64(ls.inFlight.Load()) / float64(ls.maxLoad)

    switch p {
    case PriorityCritical:
        return true
    case PriorityHigh:
        return load < 0.95
    case PriorityNormal:
        return load < 0.80
    case PriorityLow:
        return load < 0.60
    }
    return false
}
`)

	fmt.Println("--- Demo: Load Shedder ---")
	demoLoadShedder()

	// ============================================
	// GRACEFUL DEGRADATION
	// ============================================
	fmt.Println("\n--- Graceful Degradation ---")
	os.Stdout.WriteString(`
CONCEPTO:
Cuando un componente falla, ofrecer funcionalidad reducida
en lugar de fallar completamente.

PATRON FALLBACK:

type RecommendationService struct {
    primary   RecommendationProvider
    cache     RecommendationCache
    defaults  []Product
}

func (s *RecommendationService) GetRecommendations(
    ctx context.Context, userID string,
) ([]Product, error) {
    // Nivel 1: Servicio ML en tiempo real
    recs, err := s.primary.Recommend(ctx, userID)
    if err == nil {
        s.cache.Set(userID, recs)
        return recs, nil
    }
    log.Printf("primary recommendations failed: %v", err)

    // Nivel 2: Cache de recomendaciones anteriores
    recs, err = s.cache.Get(userID)
    if err == nil {
        return recs, nil
    }
    log.Printf("cache recommendations failed: %v", err)

    // Nivel 3: Productos populares (siempre disponible)
    return s.defaults, nil
}

PATRON FEATURE FLAGS CON DEGRADATION:

type FeatureFlags struct {
    mu       sync.RWMutex
    features map[string]bool
}

func (f *FeatureFlags) IsEnabled(feature string) bool {
    f.mu.RLock()
    defer f.mu.RUnlock()
    enabled, ok := f.features[feature]
    return ok && enabled
}

func (f *FeatureFlags) Disable(feature string) {
    f.mu.Lock()
    defer f.mu.Unlock()
    f.features[feature] = false
}

type OrderService struct {
    flags *FeatureFlags
}

func (s *OrderService) PlaceOrder(ctx context.Context, order Order) error {
    // Funcionalidad core: siempre activa
    if err := s.validateOrder(order); err != nil {
        return err
    }
    if err := s.processPayment(ctx, order); err != nil {
        return err
    }

    // Funcionalidad degradable
    if s.flags.IsEnabled("fraud-detection") {
        if err := s.checkFraud(ctx, order); err != nil {
            log.Printf("fraud check failed (degraded): %v", err)
            // No bloquear el pedido
        }
    }

    if s.flags.IsEnabled("realtime-inventory") {
        if err := s.updateInventory(ctx, order); err != nil {
            log.Printf("inventory update failed, queuing: %v", err)
            s.queueInventoryUpdate(order)  // Async fallback
        }
    }

    if s.flags.IsEnabled("email-notifications") {
        go func() {
            if err := s.sendConfirmation(context.Background(), order); err != nil {
                log.Printf("email failed (non-critical): %v", err)
            }
        }()
    }

    return nil
}

HEALTH CHECK CON DEGRADATION LEVELS:

type HealthStatus string
const (
    HealthOK       HealthStatus = "ok"
    HealthDegraded HealthStatus = "degraded"
    HealthDown     HealthStatus = "down"
)

type HealthCheck struct {
    mu          sync.RWMutex
    components  map[string]HealthStatus
}

func (h *HealthCheck) Overall() HealthStatus {
    h.mu.RLock()
    defer h.mu.RUnlock()

    hasDegraded := false
    for _, status := range h.components {
        if status == HealthDown {
            return HealthDown
        }
        if status == HealthDegraded {
            hasDegraded = true
        }
    }
    if hasDegraded {
        return HealthDegraded
    }
    return HealthOK
}
`)

	// ============================================
	// COMBINANDO PATRONES
	// ============================================
	fmt.Println("\n--- Combinando Patrones de Resiliencia ---")
	fmt.Print(`
STACK DE RESILIENCIA COMPLETO:

Request
  |
  v
[Rate Limiter]        -- Controlar trafico de entrada
  |
  v
[Load Shedder]        -- Rechazar si sobrecargado
  |
  v
[Timeout]             -- Deadline para el request
  |
  v
[Bulkhead]            -- Aislar concurrencia
  |
  v
[Circuit Breaker]     -- Fail fast si downstream falla
  |
  v
[Retry + Backoff]     -- Reintentar transitorio
  |
  v
[Fallback]            -- Degradar si todo falla
  |
  v
Response

IMPLEMENTACION COMBINADA:

type ResilientService struct {
    rateLimiter  *rate.Limiter
    loadShedder  *LoadShedder
    bulkhead     *Bulkhead
    circuitBreaker *gobreaker.CircuitBreaker[Response]
    fallback     func(ctx context.Context) (Response, error)
}

func (s *ResilientService) Call(ctx context.Context, req Request) (Response, error) {
    // 1. Rate limit
    if !s.rateLimiter.Allow() {
        return Response{}, ErrRateLimited
    }

    // 2. Load shedding
    done, ok := s.loadShedder.Allow()
    if !ok {
        return Response{}, ErrOverloaded
    }
    defer done()

    // 3. Timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // 4. Bulkhead
    var result Response
    err := s.bulkhead.Execute(ctx, func() error {
        // 5. Circuit breaker + retry
        var callErr error
        result, callErr = s.circuitBreaker.Execute(func() (Response, error) {
            return retryWithBackoff(ctx, func() (Response, error) {
                return doActualCall(ctx, req)
            })
        })
        return callErr
    })

    // 6. Fallback
    if err != nil {
        return s.fallback(ctx)
    }

    return result, nil
}

ORDEN IMPORTA:
1. Rate Limit PRIMERO (mas barato, protege todo)
2. Load Shed antes de trabajo costoso
3. Timeout envuelve todo el pipeline
4. Bulkhead antes de I/O externo
5. Circuit Breaker por dependencia
6. Retry solo para errores transitorios
7. Fallback como ultima defensa
`)

	// ============================================
	// TESTING RESILIENCE
	// ============================================
	fmt.Println("\n--- Testing Resilience ---")
	fmt.Print(`
CHAOS TESTING:

func TestCircuitBreakerOpens(t *testing.T) {
    failCount := 0
    cb := NewCircuitBreaker(CircuitBreakerConfig{
        FailureThreshold: 3,
        ResetTimeout:     100 * time.Millisecond,
    })

    // Causar fallos
    for i := 0; i < 5; i++ {
        _, err := cb.Execute(func() (string, error) {
            return "", errors.New("service down")
        })
        if errors.Is(err, ErrCircuitOpen) {
            failCount++
        }
    }

    // Debe haber abierto despues de 3 fallos
    assert.Greater(t, failCount, 0)

    // Esperar reset
    time.Sleep(150 * time.Millisecond)

    // Debe estar en half-open
    result, err := cb.Execute(func() (string, error) {
        return "recovered", nil
    })
    assert.NoError(t, err)
    assert.Equal(t, "recovered", result)
}

SIMULAR LATENCIA:

type LatencyInjector struct {
    delay    time.Duration
    failRate float64
}

func (li *LatencyInjector) Wrap(fn func() error) func() error {
    return func() error {
        time.Sleep(li.delay)
        if rand.Float64() < li.failRate {
            return errors.New("injected failure")
        }
        return fn()
    }
}

PROPERTY-BASED TESTING:

func TestRateLimiter_NeverExceedsRate(t *testing.T) {
    limiter := NewRateLimiter(100, time.Second)

    var allowed int64
    start := time.Now()

    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if limiter.Allow() {
                atomic.AddInt64(&allowed, 1)
            }
        }()
    }
    wg.Wait()

    elapsed := time.Since(start)
    maxExpected := int64(float64(100) * elapsed.Seconds() * 1.1) // 10% margen
    assert.LessOrEqual(t, allowed, maxExpected)
}
`)

	fmt.Println("\n=== RESUMEN EJECUTADO ===")
	fmt.Println("Todos los demos ejecutados exitosamente.")
}

// ============================================
// DEMO: TOKEN BUCKET
// ============================================

func demoTokenBucket() {
	type SimpleTokenBucket struct {
		mu         sync.Mutex
		tokens     float64
		maxTokens  float64
		refillRate float64
		lastRefill time.Time
	}

	newBucket := func(maxTokens, refillPerSecond float64) *SimpleTokenBucket {
		return &SimpleTokenBucket{
			tokens:     maxTokens,
			maxTokens:  maxTokens,
			refillRate: refillPerSecond,
			lastRefill: time.Now(),
		}
	}

	allow := func(tb *SimpleTokenBucket) bool {
		tb.mu.Lock()
		defer tb.mu.Unlock()

		now := time.Now()
		elapsed := now.Sub(tb.lastRefill).Seconds()
		tb.tokens = math.Min(tb.maxTokens, tb.tokens+elapsed*tb.refillRate)
		tb.lastRefill = now

		if tb.tokens >= 1 {
			tb.tokens--
			return true
		}
		return false
	}

	bucket := newBucket(3, 2)

	fmt.Println("Bucket: capacity=3, refill=2/sec")
	for i := 1; i <= 6; i++ {
		ok := allow(bucket)
		fmt.Printf("  Request %d: allowed=%v\n", i, ok)
	}

	fmt.Println("  (esperando 1 segundo para refill...)")
	time.Sleep(1 * time.Second)

	for i := 7; i <= 9; i++ {
		ok := allow(bucket)
		fmt.Printf("  Request %d: allowed=%v\n", i, ok)
	}
}

// ============================================
// DEMO: SLIDING WINDOW
// ============================================

func demoSlidingWindow() {
	type SlidingWindow struct {
		mu        sync.Mutex
		requests  []time.Time
		limit     int
		window    time.Duration
	}

	newWindow := func(limit int, window time.Duration) *SlidingWindow {
		return &SlidingWindow{
			requests: make([]time.Time, 0, limit),
			limit:    limit,
			window:   window,
		}
	}

	allow := func(sw *SlidingWindow) bool {
		sw.mu.Lock()
		defer sw.mu.Unlock()

		now := time.Now()
		windowStart := now.Add(-sw.window)

		valid := sw.requests[:0]
		for _, t := range sw.requests {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		sw.requests = valid

		if len(sw.requests) >= sw.limit {
			return false
		}

		sw.requests = append(sw.requests, now)
		return true
	}

	sw := newWindow(3, 500*time.Millisecond)

	fmt.Println("Sliding Window: limit=3 per 500ms")
	for i := 1; i <= 5; i++ {
		ok := allow(sw)
		fmt.Printf("  Request %d: allowed=%v\n", i, ok)
	}

	fmt.Println("  (esperando 600ms para que expire la ventana...)")
	time.Sleep(600 * time.Millisecond)

	for i := 6; i <= 8; i++ {
		ok := allow(sw)
		fmt.Printf("  Request %d: allowed=%v\n", i, ok)
	}
}

// ============================================
// DEMO: CIRCUIT BREAKER
// ============================================

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu               sync.Mutex
	state            CircuitState
	failures         int
	successes        int
	failureThreshold int
	successThreshold int
	resetTimeout     time.Duration
	lastFailure      time.Time
}

func NewCircuitBreaker(failureThreshold, successThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		resetTimeout:     resetTimeout,
	}
}

func (cb *CircuitBreaker) Execute(fn func() (string, error)) (string, error) {
	cb.mu.Lock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			fmt.Printf("    [CB] State: OPEN -> HALF-OPEN\n")
		} else {
			cb.mu.Unlock()
			return "", ErrCircuitOpen
		}
	}

	cb.mu.Unlock()

	result, err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.state == StateHalfOpen || cb.failures >= cb.failureThreshold {
			prev := cb.state
			cb.state = StateOpen
			fmt.Printf("    [CB] State: %s -> OPEN (failures=%d)\n", prev, cb.failures)
		}
		return "", err
	}

	if cb.state == StateHalfOpen {
		cb.successes++
		if cb.successes >= cb.successThreshold {
			cb.state = StateClosed
			cb.failures = 0
			fmt.Printf("    [CB] State: HALF-OPEN -> CLOSED (recovered!)\n")
		}
	} else {
		cb.failures = 0
	}

	return result, nil
}

func demoCircuitBreaker() {
	cb := NewCircuitBreaker(3, 2, 500*time.Millisecond)
	callCount := 0
	shouldFail := true

	externalCall := func() (string, error) {
		callCount++
		if shouldFail {
			return "", errors.New("service unavailable")
		}
		return "OK", nil
	}

	fmt.Println("Circuit Breaker: threshold=3 failures, reset=500ms")
	fmt.Println("  Fase 1: Servicio fallando")
	for i := 1; i <= 5; i++ {
		result, err := cb.Execute(externalCall)
		if err != nil {
			fmt.Printf("  Call %d: error=%v\n", i, err)
		} else {
			fmt.Printf("  Call %d: result=%s\n", i, result)
		}
	}

	fmt.Printf("  Total calls al servicio: %d (ahorradas: %d)\n", callCount, 5-callCount)

	fmt.Println("  Fase 2: Esperando reset timeout...")
	time.Sleep(600 * time.Millisecond)
	shouldFail = false

	fmt.Println("  Fase 3: Servicio recuperado")
	for i := 6; i <= 9; i++ {
		result, err := cb.Execute(externalCall)
		if err != nil {
			fmt.Printf("  Call %d: error=%v\n", i, err)
		} else {
			fmt.Printf("  Call %d: result=%s\n", i, result)
		}
	}
}

// ============================================
// DEMO: RETRY WITH BACKOFF
// ============================================

func demoRetryWithBackoff() {
	attempt := 0

	unreliableService := func() error {
		attempt++
		if attempt < 4 {
			return errors.New("temporary error")
		}
		return nil
	}

	retryWithJitter := func(ctx context.Context, maxRetries int, baseDelay time.Duration, fn func() error) error {
		var err error
		for i := 0; i <= maxRetries; i++ {
			err = fn()
			if err == nil {
				fmt.Printf("  Exito en intento %d\n", i+1)
				return nil
			}

			if i == maxRetries {
				break
			}

			delay := baseDelay * time.Duration(1<<uint(i))
			maxDelay := 5 * time.Second
			if delay > maxDelay {
				delay = maxDelay
			}
			jitter := time.Duration(rand.Int64N(int64(delay)))

			fmt.Printf("  Intento %d fallo: %v (retry en %v)\n", i+1, err, jitter)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(jitter):
			}
		}
		return fmt.Errorf("after %d retries: %w", maxRetries, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Retry: max=5, baseDelay=100ms, exponential+jitter")
	err := retryWithJitter(ctx, 5, 100*time.Millisecond, unreliableService)
	if err != nil {
		fmt.Printf("  Final error: %v\n", err)
	}
}

// ============================================
// DEMO: BULKHEAD
// ============================================

func demoBulkhead() {
	type Bulkhead struct {
		sem chan struct{}
	}

	newBulkhead := func(max int) *Bulkhead {
		return &Bulkhead{sem: make(chan struct{}, max)}
	}

	execute := func(b *Bulkhead, ctx context.Context, name string, fn func() error) error {
		select {
		case b.sem <- struct{}{}:
			defer func() { <-b.sem }()
			return fn()
		case <-ctx.Done():
			return fmt.Errorf("bulkhead full for %s: %w", name, ctx.Err())
		}
	}

	bh := newBulkhead(3)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fmt.Println("Bulkhead: max_concurrent=3, launching 6 tasks")

	var wg sync.WaitGroup
	for i := 1; i <= 6; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := execute(bh, ctx, fmt.Sprintf("task-%d", id), func() error {
				fmt.Printf("  Task %d: ejecutando\n", id)
				time.Sleep(200 * time.Millisecond)
				fmt.Printf("  Task %d: completada\n", id)
				return nil
			})
			if err != nil {
				fmt.Printf("  Task %d: %v\n", id, err)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("  Todas las tareas procesadas (max 3 concurrentes)")
}

// ============================================
// DEMO: TIMEOUT BUDGET
// ============================================

func demoTimeoutBudget() {
	step := func(ctx context.Context, name string, duration time.Duration) error {
		if ctx.Err() != nil {
			return fmt.Errorf("%s: context already done: %w", name, ctx.Err())
		}

		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			fmt.Printf("  %s: budget restante=%v\n", name, remaining.Round(time.Millisecond))
			if remaining < duration {
				return fmt.Errorf("%s: insufficient budget (%v < %v)", name, remaining.Round(time.Millisecond), duration)
			}
		}

		select {
		case <-time.After(duration):
			fmt.Printf("  %s: completado en %v\n", name, duration)
			return nil
		case <-ctx.Done():
			return fmt.Errorf("%s: timeout: %w", name, ctx.Err())
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	fmt.Println("Budget total: 500ms")

	if err := step(ctx, "Validate", 100*time.Millisecond); err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}

	if err := step(ctx, "Payment", 200*time.Millisecond); err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}

	if err := step(ctx, "Shipping", 300*time.Millisecond); err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}

	fmt.Println("  Todos los pasos completados dentro del budget!")
}

// ============================================
// DEMO: LOAD SHEDDER
// ============================================

func demoLoadShedder() {
	var inFlight atomic.Int64
	var maxLoad int64 = 3

	allow := func() (func(), bool) {
		current := inFlight.Add(1)
		if current > maxLoad {
			inFlight.Add(-1)
			return nil, false
		}
		return func() { inFlight.Add(-1) }, true
	}

	fmt.Println("Load Shedder: max_concurrent=3, sending 8 requests")

	var wg sync.WaitGroup
	var accepted, rejected atomic.Int64

	for i := 1; i <= 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			done, ok := allow()
			if !ok {
				rejected.Add(1)
				fmt.Printf("  Request %d: SHED (rechazado)\n", id)
				return
			}
			defer done()
			accepted.Add(1)
			fmt.Printf("  Request %d: ACCEPTED (in-flight=%d)\n", id, inFlight.Load())
			time.Sleep(150 * time.Millisecond)
		}(i)
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	fmt.Printf("  Resultado: accepted=%d, rejected=%d\n", accepted.Load(), rejected.Load())
}

// ============================================
// TIPOS AUXILIARES (para compilacion)
// ============================================

type Request struct{}
type Response struct{}
type Order struct{}
type Product struct{}

type RecommendationProvider interface {
	Recommend(ctx context.Context, userID string) ([]Product, error)
}

type RecommendationCache interface {
	Get(userID string) ([]Product, error)
	Set(userID string, products []Product)
}

// Evitar warnings de imports no usados en los demos
var _ = http.StatusOK
var _ = rand.Int64N

/*
RESUMEN CAPITULO 65: RESILIENCE PATTERNS

RATE LIMITING:
- Token Bucket:      golang.org/x/time/rate (Allow, Wait, Reserve)
- Per-Client:        Map de limiters por IP/API key
- Sliding Window:    Ventana deslizante, mas preciso
- Distribuido:       Redis sorted sets para multiples instancias
- Headers:           X-RateLimit-Limit, Remaining, Reset, Retry-After

CIRCUIT BREAKER:
- Estados:           Closed -> Open -> Half-Open -> Closed
- sony/gobreaker:    Libreria madura y configurable
- Configuracion:     MaxRequests, Interval, Timeout, ReadyToTrip
- Fail Fast:         Evitar cascading failures
- IsSuccessful:      Definir que errores cuentan como fallo

RETRY PATTERNS:
- Exponential Backoff: delay * 2^attempt
- Jitter:              Evitar thundering herd (Full Jitter recomendado)
- cenkalti/backoff:    Libreria completa con context support
- Permanent errors:    No reintentar errores definitivos
- Idempotency:         Asegurar que reintentos son seguros

BULKHEAD:
- Semaphore:           Channel como semaforo de concurrencia
- Worker Pool:         Pool fijo con queue limitada
- Per-Service:         Aislar recursos por dependencia

TIMEOUTS:
- Context:             WithTimeout/WithDeadline para todo I/O
- HTTP Server:         Read, Write, Idle, Handler timeouts
- Database:            connect_timeout + query context
- Cascading Budget:    Propagar tiempo restante entre capas

DEADLINE PROPAGATION:
- Context chain:       Toda operacion recibe y respeta ctx
- HTTP headers:        X-Request-Timeout, X-Request-Deadline
- gRPC:                Propagacion automatica de deadlines
- Fail fast:           Si deadline expiro, no empezar trabajo

LOAD SHEDDING:
- Concurrencia:        Rechazar si in-flight > threshold
- Latencia:            Shed adaptivo basado en avg latency
- Prioridad:           Critical nunca shed, Low shed primero
- HTTP 503:            Service Unavailable + Retry-After

GRACEFUL DEGRADATION:
- Fallback chain:      Primary -> Cache -> Defaults
- Feature flags:       Desactivar features no-criticas
- Async fallback:      Queue para procesar despues
- Health levels:       OK, Degraded, Down

STACK COMPLETO (orden):
1. Rate Limiter     (barato, protege todo)
2. Load Shedder     (protege capacidad)
3. Timeout          (deadline global)
4. Bulkhead         (aislar recursos)
5. Circuit Breaker  (fail fast por dependencia)
6. Retry + Backoff  (errores transitorios)
7. Fallback         (ultima defensa)

TESTING:
- Chaos testing:       Inyectar fallos y latencia
- Property testing:    Verificar invariantes (nunca excede rate)
- Integration:         Simular escenarios de fallo reales
*/
