// Package main - Chapter 052: Structured Logging with log/slog
// Go 1.21+ introduced log/slog for structured, leveled logging.
// This chapter covers the slog API, handlers, and production patterns.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	fmt.Println("=== STRUCTURED LOGGING WITH LOG/SLOG ===")

	// ============================================
	// BASIC USAGE
	// ============================================
	fmt.Println("\n--- Basic Usage ---")

	slog.Info("application started",
		"version", "1.0.0",
		"environment", "development",
	)

	slog.Debug("debug message (won't show at default level)")
	slog.Warn("disk space low", "available_gb", 5)
	slog.Error("connection failed",
		"host", "db.example.com",
		"error", "timeout",
	)

	fmt.Println(`
LOG LEVELS (ascending severity):
slog.LevelDebug  = -4
slog.LevelInfo   = 0   (default)
slog.LevelWarn   = 4
slog.LevelError  = 8

BASIC FUNCTIONS:
slog.Debug(msg, key, val, ...)
slog.Info(msg, key, val, ...)
slog.Warn(msg, key, val, ...)
slog.Error(msg, key, val, ...)`)

	// ============================================
	// TEXT HANDLER
	// ============================================
	fmt.Println("\n--- TextHandler ---")

	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	textLogger := slog.New(textHandler)
	textLogger.Info("text format", "user", "alice", "action", "login")
	textLogger.Debug("debug visible now", "detail", "verbose")

	fmt.Println(`
TEXT HANDLER OUTPUT:
time=2024-01-01T00:00:00Z level=INFO msg="text format" user=alice action=login

OPTIONS:
slog.HandlerOptions{
    Level:     slog.LevelDebug,  // Minimum level
    AddSource: true,             // Add file:line
    ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
        // Modify attributes before output
        return a
    },
}`)

	// ============================================
	// JSON HANDLER
	// ============================================
	fmt.Println("\n--- JSONHandler ---")

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	jsonLogger := slog.New(jsonHandler)
	jsonLogger.Info("json format", "user_id", 12345, "latency_ms", 42)

	fmt.Println(`
JSON HANDLER OUTPUT:
{"time":"2024-01-01T00:00:00Z","level":"INFO","msg":"json format","user_id":12345,"latency_ms":42}

USE CASES:
TextHandler - Development, human reading
JSONHandler - Production, log aggregation (ELK, Datadog, Loki)`)

	// ============================================
	// TYPED ATTRIBUTES
	// ============================================
	fmt.Println("\n--- Typed Attributes ---")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	logger.Info("typed attrs",
		slog.String("name", "alice"),
		slog.Int("age", 30),
		slog.Float64("score", 95.5),
		slog.Bool("active", true),
		slog.Time("created", time.Now()),
		slog.Duration("elapsed", 250*time.Millisecond),
		slog.Any("data", map[string]int{"a": 1}),
	)

	fmt.Println(`
TYPED ATTRIBUTE CONSTRUCTORS (avoid allocations):
slog.String(key, val)
slog.Int(key, val)
slog.Int64(key, val)
slog.Uint64(key, val)
slog.Float64(key, val)
slog.Bool(key, val)
slog.Time(key, val)
slog.Duration(key, val)
slog.Any(key, val)       // For complex types
slog.Group(key, attrs)   // Nested group`)

	// ============================================
	// GROUPS
	// ============================================
	fmt.Println("\n--- Groups ---")

	logger.Info("request completed",
		slog.Group("request",
			slog.String("method", "GET"),
			slog.String("path", "/api/users"),
			slog.Int("status", 200),
		),
		slog.Group("response",
			slog.Duration("latency", 45*time.Millisecond),
			slog.Int("bytes", 1024),
		),
	)

	fmt.Println(`
TEXT OUTPUT:
request.method=GET request.path=/api/users request.status=200 response.latency=45ms response.bytes=1024

JSON OUTPUT:
{"request":{"method":"GET","path":"/api/users","status":200},"response":{"latency":"45ms","bytes":1024}}`)

	// ============================================
	// WITH (CHILD LOGGERS)
	// ============================================
	fmt.Println("\n--- With (Child Loggers) ---")

	baseLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	requestLogger := baseLogger.With(
		"request_id", "abc-123",
		"user_id", 42,
	)

	requestLogger.Info("processing request")
	requestLogger.Info("fetching data", "table", "users")
	requestLogger.Warn("slow query", "duration_ms", 500)

	fmt.Println(`
WITH adds attributes to ALL subsequent log entries:

logger := slog.Default().With("service", "api", "version", "2.0")
logger.Info("started")   // includes service=api version=2.0
logger.Error("failed")   // includes service=api version=2.0

WITH GROUP adds a group prefix:
logger := slog.Default().WithGroup("http")
logger.Info("request", "method", "GET")
// Output: http.method=GET`)

	// ============================================
	// CONTEXT INTEGRATION
	// ============================================
	fmt.Println("\n--- Context Integration ---")

	ctx := context.Background()

	slog.InfoContext(ctx, "with context",
		"operation", "database_query",
	)

	fmt.Println(`
CONTEXT FUNCTIONS:
slog.DebugContext(ctx, msg, attrs...)
slog.InfoContext(ctx, msg, attrs...)
slog.WarnContext(ctx, msg, attrs...)
slog.ErrorContext(ctx, msg, attrs...)
slog.Log(ctx, level, msg, attrs...)
slog.LogAttrs(ctx, level, msg, ...slog.Attr)

EXTRACT VALUES FROM CONTEXT:
A custom Handler can extract request_id, trace_id, etc.
from context and add them automatically to every log entry`)

	// ============================================
	// LOGVALUER INTERFACE
	// ============================================
	fmt.Println("\n--- LogValuer Interface ---")

	type User struct {
		ID       int
		Name     string
		Email    string
		Password string
	}

	// Implement LogValuer to control what gets logged
	userLogValue := func(u User) slog.Value {
		return slog.GroupValue(
			slog.Int("id", u.ID),
			slog.String("name", u.Name),
			// Password intentionally omitted!
		)
	}

	u := User{ID: 1, Name: "Alice", Email: "alice@example.com", Password: "secret"}
	logger.Info("user action", "user", userLogValue(u))

	fmt.Println(`
type LogValuer interface {
    LogValue() slog.Value
}

IMPLEMENT ON YOUR TYPES:
func (u User) LogValue() slog.Value {
    return slog.GroupValue(
        slog.Int("id", u.ID),
        slog.String("name", u.Name),
        // Omit sensitive fields!
    )
}

// Then use directly:
logger.Info("login", "user", user)  // Calls LogValue() automatically

USE CASES:
- Redact sensitive data (passwords, tokens, PII)
- Custom formatting for complex types
- Lazy evaluation (expensive to compute values)`)

	// ============================================
	// REPLACE ATTR
	// ============================================
	fmt.Println("\n--- ReplaceAttr ---")

	customHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "ts"
				a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339))
			}
			if a.Key == slog.LevelKey {
				a.Key = "severity"
			}
			return a
		},
	})
	customLogger := slog.New(customHandler)
	customLogger.Info("custom attributes")

	fmt.Println(`
COMMON ReplaceAttr USES:

1. RENAME KEYS:
   slog.TimeKey -> "ts" or "timestamp"
   slog.LevelKey -> "severity"
   slog.MessageKey -> "message"

2. CHANGE TIME FORMAT:
   a.Value = slog.StringValue(t.Format(time.RFC3339))

3. REMOVE ATTRIBUTES:
   return slog.Attr{}   // Empty attr is removed

4. REDACT SENSITIVE:
   if a.Key == "password" {
       a.Value = slog.StringValue("***REDACTED***")
   }`)

	// ============================================
	// SET DEFAULT LOGGER
	// ============================================
	fmt.Println("\n--- Set Default Logger ---")

	fmt.Println(`
SET GLOBAL DEFAULT:

handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level:     slog.LevelInfo,
    AddSource: true,
})
slog.SetDefault(slog.New(handler))

// Now all slog.Info(), slog.Error() etc use this handler
slog.Info("uses default handler")

// Also redirects standard log package:
log.Println("this goes through slog too")

BRIDGE TO log.Logger:
stdLogger := slog.NewLogLogger(handler, slog.LevelError)
// Use stdLogger where *log.Logger is needed
server.ErrorLog = stdLogger`)

	// ============================================
	// DYNAMIC LEVEL
	// ============================================
	fmt.Println("\n--- Dynamic Level ---")

	var level slog.LevelVar
	level.Set(slog.LevelInfo)

	dynHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: &level,
	})
	dynLogger := slog.New(dynHandler)

	dynLogger.Debug("not visible at INFO level")
	dynLogger.Info("visible", "level", level.Level())

	level.Set(slog.LevelDebug)
	dynLogger.Debug("now visible after level change", "level", level.Level())

	os.Stdout.WriteString(`
slog.LevelVar:
- Thread-safe level that can change at runtime
- Use for admin endpoints to adjust verbosity
- No need to restart application

HTTP ENDPOINT:
http.HandleFunc("/log/level", func(w http.ResponseWriter, r *http.Request) {
    if r.Method == "PUT" {
        var l slog.Level
        l.UnmarshalText([]byte(r.URL.Query().Get("level")))
        levelVar.Set(l)
    }
    fmt.Fprintf(w, "level: %s", levelVar.Level())
})
`)

	// ============================================
	// PERFORMANCE TIPS
	// ============================================
	fmt.Println("\n--- Performance Tips ---")
	os.Stdout.WriteString(`
1. USE TYPED ATTRS (avoid reflection):
   // Slow: slog.Info("x", "count", 42)
   // Fast: slog.Info("x", slog.Int("count", 42))

2. CHECK LEVEL BEFORE EXPENSIVE OPERATIONS:
   if logger.Enabled(ctx, slog.LevelDebug) {
       logger.Debug("details", "dump", expensiveSerialize(data))
   }

3. USE LogAttrs FOR BEST PERFORMANCE:
   logger.LogAttrs(ctx, slog.LevelInfo, "request",
       slog.String("method", r.Method),
       slog.String("path", r.URL.Path),
       slog.Int("status", status),
   )

4. PRE-COMPUTE With LOGGERS:
   // Do once at init:
   reqLogger := logger.With("service", "api")
   // Use many times:
   reqLogger.Info("request")

5. AVOID STRING FORMATTING:
   // Bad: slog.Info(fmt.Sprintf("user %d logged in", id))
   // Good: slog.Info("user logged in", "user_id", id)
`)

	// ============================================
	// COMPARISON WITH ZEROLOG AND ZAP
	// ============================================
	fmt.Println("\n--- Comparison: slog vs zerolog vs zap ---")
	fmt.Println(`
FEATURE         | slog        | zerolog     | zap
-------------------------------------------------
Stdlib          | Yes         | No          | No
Zero alloc      | Partial     | Yes         | Yes
JSON native     | Yes         | Yes         | Yes
Structured      | Yes         | Yes         | Yes
Levels          | 4           | 7           | 6
Context         | Yes         | Yes         | Yes
Sampling        | No          | Yes         | Yes
Hook support    | Via Handler | Yes         | Yes

RECOMMENDATION:
- New projects: Start with slog (stdlib, good enough)
- High-perf logging: zerolog or zap
- Most projects: slog with JSONHandler is sufficient`)

	// ============================================
	// PRODUCTION SETUP
	// ============================================
	fmt.Println("\n--- Production Setup Pattern ---")
	os.Stdout.WriteString(`
func setupLogger(env string) *slog.Logger {
    var handler slog.Handler

    switch env {
    case "production":
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level:     slog.LevelInfo,
            AddSource: true,
            ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
                if a.Key == slog.TimeKey {
                    a.Value = slog.StringValue(
                        a.Value.Time().Format(time.RFC3339Nano),
                    )
                }
                return a
            },
        })
    default:
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }

    logger := slog.New(handler)
    slog.SetDefault(logger)
    return logger
}
`)
}

/*
SUMMARY - LOG/SLOG:

HANDLERS:
slog.NewTextHandler(w, opts) - Human-readable
slog.NewJSONHandler(w, opts) - Machine-readable

FUNCTIONS:
slog.Debug/Info/Warn/Error(msg, key, val, ...)
slog.DebugContext/InfoContext/WarnContext/ErrorContext(ctx, ...)
slog.LogAttrs(ctx, level, msg, ...slog.Attr)

TYPED ATTRS (avoid allocations):
slog.String, slog.Int, slog.Float64, slog.Bool
slog.Time, slog.Duration, slog.Any, slog.Group

CHILD LOGGERS:
logger.With(key, val)   - Add persistent attrs
logger.WithGroup(name)  - Add group prefix

INTERFACES:
slog.Handler    - Custom output handlers
slog.LogValuer  - Custom type logging

OPTIONS:
Level     - Min log level (or *slog.LevelVar for dynamic)
AddSource - Include file:line
ReplaceAttr - Modify attrs before output

PERFORMANCE:
1. Use typed attrs (slog.Int, slog.String)
2. Check Enabled() before expensive ops
3. Use LogAttrs for hot paths
4. Pre-compute With loggers
*/
