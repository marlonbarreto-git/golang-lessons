// Package main - Chapter 033: Context
// context.Context maneja cancelación, deadlines y valores a través del call stack.
// Es esencial para APIs bien diseñadas y operaciones cancelables.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== CONTEXT EN GO ===")

	// ============================================
	// CONTEXTOS BASE
	// ============================================
	fmt.Println("\n--- Contextos Base ---")

	// context.Background() - contexto raíz, nunca se cancela
	ctx := context.Background()
	fmt.Printf("Background: %v\n", ctx)

	// context.TODO() - marcador para código que aún no sabe qué context usar
	ctxTodo := context.TODO()
	fmt.Printf("TODO: %v\n", ctxTodo)

	// ============================================
	// CONTEXT WITH CANCEL
	// ============================================
	fmt.Println("\n--- Context WithCancel ---")

	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Goroutine: recibí señal de cancelación")
				return
			default:
				fmt.Println("Goroutine: trabajando...")
				time.Sleep(50 * time.Millisecond)
			}
		}
	}(ctx)

	time.Sleep(150 * time.Millisecond)
	cancel() // cancelar
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("Context error: %v\n", ctx.Err())

	// ============================================
	// CONTEXT WITH TIMEOUT
	// ============================================
	fmt.Println("\n--- Context WithTimeout ---")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2() // siempre llamar cancel, incluso si timeout ocurre

	select {
	case <-time.After(200 * time.Millisecond):
		fmt.Println("Operación completada")
	case <-ctx2.Done():
		fmt.Printf("Timeout: %v\n", ctx2.Err())
	}

	// ============================================
	// CONTEXT WITH DEADLINE
	// ============================================
	fmt.Println("\n--- Context WithDeadline ---")

	deadline := time.Now().Add(100 * time.Millisecond)
	ctx3, cancel3 := context.WithDeadline(context.Background(), deadline)
	defer cancel3()

	if d, ok := ctx3.Deadline(); ok {
		fmt.Printf("Deadline: %v (en %.0fms)\n", d.Format("15:04:05.000"), time.Until(d).Seconds()*1000)
	}

	<-ctx3.Done()
	fmt.Printf("Deadline alcanzado: %v\n", ctx3.Err())

	// ============================================
	// CONTEXT WITH VALUE
	// ============================================
	fmt.Println("\n--- Context WithValue ---")

	// Usar tipos propios como keys (no strings)
	type ctxKey string
	const userIDKey ctxKey = "userID"
	const requestIDKey ctxKey = "requestID"

	ctx4 := context.WithValue(context.Background(), userIDKey, 12345)
	ctx4 = context.WithValue(ctx4, requestIDKey, "req-abc-123")

	// Extraer valores
	if userID, ok := ctx4.Value(userIDKey).(int); ok {
		fmt.Printf("User ID: %d\n", userID)
	}

	if reqID, ok := ctx4.Value(requestIDKey).(string); ok {
		fmt.Printf("Request ID: %s\n", reqID)
	}

	// Valor no existente
	if val := ctx4.Value("otra-key"); val == nil {
		fmt.Println("Key 'otra-key' no existe")
	}

	// ============================================
	// CONTEXT WITH CAUSE (Go 1.20+)
	// ============================================
	fmt.Println("\n--- Context WithCancelCause (Go 1.20+) ---")

	ctx5, cancelWithCause := context.WithCancelCause(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancelWithCause(errors.New("razón personalizada"))
	}()

	<-ctx5.Done()
	fmt.Printf("Error: %v\n", ctx5.Err())
	fmt.Printf("Causa: %v\n", context.Cause(ctx5))

	// ============================================
	// CONTEXT AFTER FUNC (Go 1.21+)
	// ============================================
	fmt.Println("\n--- Context AfterFunc (Go 1.21+) ---")

	ctx6, cancel6 := context.WithCancel(context.Background())

	context.AfterFunc(ctx6, func() {
		fmt.Println("AfterFunc: context fue cancelado")
	})

	cancel6()
	time.Sleep(10 * time.Millisecond)

	// ============================================
	// PROPAGACIÓN DE CONTEXT
	// ============================================
	fmt.Println("\n--- Propagación de Context ---")

	ctx7, cancel7 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel7()

	resultado, err := operacionCompleja(ctx7)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Resultado: %s\n", resultado)
	}

	// ============================================
	// PATRÓN: REQUEST SCOPED VALUES
	// ============================================
	fmt.Println("\n--- Request Scoped Values ---")

	// Simular un request HTTP
	handler := func(ctx context.Context) {
		// Extraer información del request
		reqCtx := RequestContext{
			UserID:    ctx.Value(userIDKey).(int),
			RequestID: ctx.Value(requestIDKey).(string),
		}

		fmt.Printf("Procesando request: UserID=%d, RequestID=%s\n",
			reqCtx.UserID, reqCtx.RequestID)

		// Propagar a servicios internos
		buscarUsuario(ctx, reqCtx.UserID)
	}

	// Crear context con datos del request
	reqCtx := context.WithValue(context.Background(), userIDKey, 42)
	reqCtx = context.WithValue(reqCtx, requestIDKey, "req-xyz-789")

	handler(reqCtx)

	// ============================================
	// CONTEXT EN HTTP
	// ============================================
	fmt.Println("\n--- Context en HTTP ---")

	// El request tiene su propio context
	handler2 := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Verificar cancelación antes de operaciones costosas
		select {
		case <-ctx.Done():
			http.Error(w, "Request cancelado", http.StatusServiceUnavailable)
			return
		default:
		}

		// Operación con el context del request
		resultado, err := operacionConContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Resultado: %s", resultado)
	}
	_ = handler2

	fmt.Println("Los handlers HTTP reciben context via r.Context()")

	// ============================================
	// ANTI-PATTERNS
	// ============================================
	fmt.Println("\n--- Anti-Patterns ---")
	fmt.Println(`
❌ Context en structs
   type Service struct {
       ctx context.Context  // ¡NO!
   }

✓ Context como primer parámetro
   func (s *Service) DoSomething(ctx context.Context) error

❌ Pasar nil context
   DoSomething(nil, data)  // ¡NO!

✓ Usar context.Background() o context.TODO()
   DoSomething(context.Background(), data)

❌ Usar context para pasar datos opcionales
   ctx = context.WithValue(ctx, "timeout", 30)  // ¡NO!

✓ Usar parámetros explícitos
   DoSomething(ctx, timeout)

❌ Cancelar context en la función que lo recibe
   func Process(ctx context.Context) {
       ctx, cancel := context.WithCancel(ctx)  // confuso
   }

✓ El llamador controla la cancelación
   ctx, cancel := context.WithTimeout(ctx, timeout)
   defer cancel()
   Process(ctx)`)
	// ============================================
	// BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Best Practices ---")
	fmt.Println(`
1. Context es el PRIMER parámetro: func Foo(ctx context.Context, ...)
2. No almacenar context en structs
3. Siempre llamar cancel() aunque el contexto expire
4. Usar tipos propios como keys de WithValue
5. Solo usar WithValue para request-scoped data
6. Propagar context a través de todo el call stack
7. Verificar ctx.Done() en operaciones largas
8. Usar context.Background() como raíz, context.TODO() temporalmente`)}

// ============================================
// TIPOS Y FUNCIONES
// ============================================

type RequestContext struct {
	UserID    int
	RequestID string
}

func operacionCompleja(ctx context.Context) (string, error) {
	// Verificar si ya está cancelado
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context ya cancelado: %w", err)
	}

	// Simular operación larga dividida en pasos
	for i := 1; i <= 3; i++ {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("cancelado en paso %d: %w", i, ctx.Err())
		case <-time.After(50 * time.Millisecond):
			fmt.Printf("Paso %d completado\n", i)
		}
	}

	return "operación exitosa", nil
}

func buscarUsuario(ctx context.Context, userID int) {
	// En producción: verificar ctx.Done() y propagar a queries
	fmt.Printf("Buscando usuario %d...\n", userID)
}

func operacionConContext(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(10 * time.Millisecond):
		return "OK", nil
	}
}

/*
RESUMEN:

CREAR CONTEXTS:
context.Background()                    // contexto raíz
context.TODO()                          // placeholder
context.WithCancel(parent)              // cancelable
context.WithTimeout(parent, duration)   // con timeout
context.WithDeadline(parent, time)      // con deadline
context.WithValue(parent, key, value)   // con valor
context.WithCancelCause(parent)         // con causa (1.20+)

MÉTODOS DE CONTEXT:
ctx.Done() <-chan struct{}              // channel cerrado cuando cancela
ctx.Err() error                         // error de cancelación
ctx.Deadline() (time.Time, bool)        // deadline si existe
ctx.Value(key) any                      // valor asociado

ERRORES:
context.Canceled                        // cancelación manual
context.DeadlineExceeded                // timeout/deadline

BUENAS PRÁCTICAS:

1. Primer parámetro de funciones
2. No guardar en structs
3. Siempre llamar cancel()
4. Tipos propios como keys
5. Solo request-scoped data en WithValue
6. Propagar siempre el context
7. Verificar ctx.Done() en loops
*/

/*
SUMMARY - CHAPTER 033: CONTEXT

CONTEXT PURPOSE:
- Carry cancellation signals across API boundaries
- Set deadlines and timeouts
- Pass request-scoped values
- Essential for well-behaved concurrent programs

CONTEXT TYPES:
- Background(): root context, never canceled
- TODO(): placeholder when unsure what to use
- WithCancel(): manual cancellation
- WithTimeout(): automatic timeout
- WithDeadline(): deadline at specific time
- WithValue(): carry request-scoped data

METHODS:
- Done() <-chan struct{}: closed when canceled
- Err() error: why canceled (Canceled or DeadlineExceeded)
- Deadline() (time.Time, bool): deadline if set
- Value(key) any: retrieve stored value

CANCELLATION:
- Call cancel() to propagate cancellation
- Always defer cancel() even if timeout expires
- Check ctx.Done() in long operations
- Use select with ctx.Done() for responsiveness

WITHVALUE USAGE:
- Use custom types as keys, not strings
- Only for request-scoped data (user ID, trace ID)
- Not for optional parameters (use explicit params)

CONTEXT IN HTTP:
- r.Context() provides request context
- Canceled when client disconnects
- Propagate to all downstream operations

BEST PRACTICES:
- First parameter: func(ctx context.Context, ...)
- Never store context in struct
- Always call cancel(), even on timeout
- Use custom types for WithValue keys
- Check ctx.Done() in loops and long operations
*/
