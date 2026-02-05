// Package main - Capítulo 49: Observabilidad
// Logging, métricas y tracing son esenciales para
// sistemas en producción. Go tiene excelente soporte.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	fmt.Println("=== OBSERVABILIDAD EN GO ===")

	// ============================================
	// LOGGING CON SLOG (Go 1.21+)
	// ============================================
	fmt.Println("\n--- slog (Structured Logging) ---")

	// Logger por defecto (text)
	slog.Info("Application started",
		"version", "1.0.0",
		"env", "production",
	)

	// Logger JSON
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	jsonLogger := slog.New(jsonHandler)

	jsonLogger.Info("Request processed",
		"method", "GET",
		"path", "/api/users",
		"duration_ms", 42,
		"status", 200,
	)

	// Con grupos
	jsonLogger.Info("User action",
		slog.Group("user",
			slog.Int("id", 123),
			slog.String("name", "Alice"),
		),
		slog.Group("request",
			slog.String("ip", "192.168.1.1"),
		),
	)

	// Logger con atributos predefinidos
	requestLogger := jsonLogger.With(
		"request_id", "abc-123",
		"user_id", 456,
	)
	requestLogger.Info("Processing order")

	// Niveles
	jsonLogger.Debug("Debug message")
	jsonLogger.Info("Info message")
	jsonLogger.Warn("Warning message")
	jsonLogger.Error("Error message", "error", fmt.Errorf("something failed"))

	// ============================================
	// LOGGING PATTERNS
	// ============================================
	fmt.Println("\n--- Logging Patterns ---")
	fmt.Println(`
1. LOGGER POR REQUEST:
func handleRequest(w http.ResponseWriter, r *http.Request) {
    logger := slog.Default().With(
        "request_id", r.Header.Get("X-Request-ID"),
        "method", r.Method,
        "path", r.URL.Path,
    )

    ctx := context.WithValue(r.Context(), loggerKey, logger)
    // ... usar logger del context ...
}

2. MIDDLEWARE DE LOGGING:
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w}

        next.ServeHTTP(wrapped, r)

        slog.Info("Request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "status", wrapped.status,
            "duration", time.Since(start),
        )
    })
}

3. ERROR LOGGING:
if err != nil {
    slog.Error("Operation failed",
        "error", err,
        "operation", "createUser",
        "user_id", userID,
    )
    return fmt.Errorf("creating user: %w", err)
}

4. CUSTOM HANDLER:
type CustomHandler struct {
    slog.Handler
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
    // Agregar campos automáticos
    r.AddAttrs(slog.String("service", "myapp"))
    return h.Handler.Handle(ctx, r)
}`)
	// ============================================
	// MÉTRICAS
	// ============================================
	fmt.Println("\n--- Métricas ---")
	fmt.Println(`
PROMETHEUS METRICS:
import "github.com/prometheus/client_golang/prometheus"
import "github.com/prometheus/client_golang/prometheus/promauto"
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Counter
var requestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

// Incrementar
requestsTotal.WithLabelValues("GET", "/api/users", "200").Inc()

// Gauge
var activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
    Name: "active_connections",
    Help: "Number of active connections",
})

activeConnections.Inc()
activeConnections.Dec()
activeConnections.Set(42)

// Histogram
var requestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration",
        Buckets: prometheus.DefBuckets,
    },
    []string{"method", "path"},
)

timer := prometheus.NewTimer(requestDuration.WithLabelValues("GET", "/api"))
defer timer.ObserveDuration()

// Summary (percentiles)
var responseSize = promauto.NewSummary(prometheus.SummaryOpts{
    Name:       "response_size_bytes",
    Help:       "Response size in bytes",
    Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
})

// Exponer métricas
http.Handle("/metrics", promhttp.Handler())

OPENTELEMETRY METRICS:
import "go.opentelemetry.io/otel/metric"

meter := otel.Meter("myapp")

counter, _ := meter.Int64Counter("requests_total")
counter.Add(ctx, 1, metric.WithAttributes(
    attribute.String("method", "GET"),
))

histogram, _ := meter.Float64Histogram("request_duration")
histogram.Record(ctx, duration.Seconds())`)
	// ============================================
	// TRACING
	// ============================================
	fmt.Println("\n--- Distributed Tracing ---")
	fmt.Println(`
OPENTELEMETRY:
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/trace"
)

// Configurar exporter
exp, _ := jaeger.New(jaeger.WithCollectorEndpoint(
    jaeger.WithEndpoint("http://jaeger:14268/api/traces"),
))

tp := trace.NewTracerProvider(
    trace.WithBatcher(exp),
    trace.WithResource(resource.NewWithAttributes(
        semconv.ServiceName("myservice"),
    )),
)
otel.SetTracerProvider(tp)
defer tp.Shutdown(ctx)

// Crear spans
tracer := otel.Tracer("myapp")

ctx, span := tracer.Start(ctx, "handleRequest")
defer span.End()

// Agregar atributos
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.Int("items.count", len(items)),
)

// Eventos
span.AddEvent("processing started")

// Errores
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}

// Span anidado
ctx, childSpan := tracer.Start(ctx, "database.query")
defer childSpan.End()

HTTP PROPAGATION:
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// Cliente
client := &http.Client{
    Transport: otelhttp.NewTransport(http.DefaultTransport),
}

// Servidor
handler := otelhttp.NewHandler(myHandler, "myserver")`)
	// ============================================
	// HEALTH CHECKS
	// ============================================
	fmt.Println("\n--- Health Checks ---")
	fmt.Println(`
type HealthChecker interface {
    Check(ctx context.Context) error
}

type HealthStatus struct {
    Status    string            ` + "`" + `json:"status"` + "`" + `
    Checks    map[string]string ` + "`" + `json:"checks"` + "`" + `
    Timestamp time.Time         ` + "`" + `json:"timestamp"` + "`" + `
}

func healthHandler(checkers map[string]HealthChecker) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        status := HealthStatus{
            Status:    "healthy",
            Checks:    make(map[string]string),
            Timestamp: time.Now(),
        }

        for name, checker := range checkers {
            if err := checker.Check(ctx); err != nil {
                status.Status = "unhealthy"
                status.Checks[name] = err.Error()
            } else {
                status.Checks[name] = "ok"
            }
        }

        if status.Status != "healthy" {
            w.WriteHeader(http.StatusServiceUnavailable)
        }

        json.NewEncoder(w).Encode(status)
    }
}

// Liveness vs Readiness
// /healthz - liveness: ¿el proceso está vivo?
// /readyz - readiness: ¿puede recibir tráfico?`)
	// ============================================
	// EJEMPLO COMPLETO
	// ============================================
	fmt.Println("\n--- Ejemplo: Middleware Observable ---")
	fmt.Println(`
func observabilityMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Tracing
        ctx, span := tracer.Start(ctx, "http.request",
            trace.WithAttributes(
                semconv.HTTPMethod(r.Method),
                semconv.HTTPURL(r.URL.String()),
            ),
        )
        defer span.End()

        // Request ID
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.NewString()
        }

        // Logger con context
        logger := slog.Default().With(
            "request_id", requestID,
            "trace_id", span.SpanContext().TraceID().String(),
        )

        // Response wrapper
        wrapped := &responseWriter{ResponseWriter: w, status: 200}

        // Timer
        start := time.Now()

        // Execute
        next.ServeHTTP(wrapped, r.WithContext(ctx))

        // Metrics
        duration := time.Since(start)
        requestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
        requestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprint(wrapped.status)).Inc()

        // Log
        logger.Info("Request completed",
            "method", r.Method,
            "path", r.URL.Path,
            "status", wrapped.status,
            "duration_ms", duration.Milliseconds(),
        )

        // Update span
        span.SetAttributes(semconv.HTTPStatusCode(wrapped.status))
        if wrapped.status >= 400 {
            span.SetStatus(codes.Error, "HTTP error")
        }
    })
}`)}

// Ejemplo de logger configurado
func configureLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		AddSource: true,
	})
	return slog.New(handler)
}

// Simulación de request processing con logging
func processRequest(ctx context.Context, logger *slog.Logger) {
	start := time.Now()
	defer func() {
		logger.Info("Request processed",
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	// Simular trabajo
	time.Sleep(10 * time.Millisecond)
}

/*
RESUMEN DE OBSERVABILIDAD:

LOGGING (slog):
slog.Info("message", "key", value)
slog.With("key", value)
slog.NewJSONHandler()

MÉTRICAS:
Prometheus:
- Counter: eventos incrementales
- Gauge: valores que suben/bajan
- Histogram: distribución de valores
- Summary: percentiles

promhttp.Handler() para exponer

TRACING:
OpenTelemetry:
tracer.Start(ctx, "span-name")
span.SetAttributes()
span.AddEvent()
span.RecordError()

HEALTH CHECKS:
/healthz - liveness
/readyz - readiness
Timeouts en checks

PATTERNS:
1. Logger por request
2. Request ID propagation
3. Metrics middleware
4. Tracing propagation
5. Structured errors

HERRAMIENTAS:
- Prometheus + Grafana
- Jaeger / Zipkin / Tempo
- ELK / Loki
- Datadog / New Relic
*/
