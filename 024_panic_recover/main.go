// Package main - Chapter 024: Panic and Recover
// panic es para errores irrecuperables, recover permite manejarlos.
// Aprenderás cuándo usar panic, cómo hacer recover, y patrones seguros.
package main

import (
	"fmt"
	"runtime/debug"
)

func main() {
	fmt.Println("=== PANIC Y RECOVER EN GO ===")

	// ============================================
	// PANIC BÁSICO
	// ============================================
	fmt.Println("\n--- Panic Básico ---")

	// panic detiene la ejecución normal
	// Descomenta para ver el comportamiento:
	// panic("algo terrible pasó")

	// panic con valores
	// panic(42)
	// panic(errors.New("error personalizado"))

	fmt.Println("Este código se ejecuta normalmente")

	// ============================================
	// RECOVER
	// ============================================
	fmt.Println("\n--- Recover ---")

	// recover solo funciona en defer
	resultado := funcionSegura()
	fmt.Printf("Resultado de función segura: %s\n", resultado)

	// ============================================
	// PATRÓN: RECOVER EN DEFER
	// ============================================
	fmt.Println("\n--- Patrón Recover en Defer ---")

	// La forma correcta de manejar panics
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recuperado de panic: %v\n", r)
			}
		}()

		fmt.Println("Antes del panic")
		panic("¡boom!")
		fmt.Println("Esto nunca se ejecuta") //nolint
	}()

	fmt.Println("Continuamos después del panic recuperado")

	// ============================================
	// STACK TRACE
	// ============================================
	fmt.Println("\n--- Stack Trace ---")

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic: %v\n", r)
				fmt.Println("Stack trace:")
				debug.PrintStack()
			}
		}()

		causarPanic()
	}()

	// ============================================
	// RECOVER NO CAPTURA TODO
	// ============================================
	fmt.Println("\n--- Limitaciones de Recover ---")
	fmt.Println(`
recover() NO puede capturar:
- runtime.Goexit() - terminación de goroutine
- os.Exit() - terminación del proceso
- Errores fatales del runtime (memoria agotada, stack overflow)

recover() SOLO funciona:
- Dentro de una función defer
- En la misma goroutine que el panic`)
	// ============================================
	// PANIC EN GOROUTINES
	// ============================================
	fmt.Println("\n--- Panic en Goroutines ---")

	// Un panic en una goroutine NO es capturado por el main
	// Cada goroutine debe manejar sus propios panics

	done := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Goroutine recuperó: %v\n", r)
			}
			done <- true
		}()

		panic("panic en goroutine")
	}()

	<-done
	fmt.Println("Main continúa después de goroutine panic")

	// ============================================
	// CUÁNDO USAR PANIC
	// ============================================
	fmt.Println("\n--- Cuándo Usar Panic ---")
	fmt.Println(`
USA PANIC para:
✓ Errores de programación (bugs) - no deberían ocurrir en producción
✓ Inicialización que no puede fallar (debe terminar el programa)
✓ Violación de invariantes (estado imposible)
✓ Índice fuera de rango (el runtime lo hace)
✓ nil pointer dereference (el runtime lo hace)

NO USES PANIC para:
✗ Errores esperados (archivo no existe, red falla)
✗ Validación de input del usuario
✗ Cualquier cosa que debería retornar error
✗ Control de flujo normal`)
	// ============================================
	// EJEMPLOS DE USO VÁLIDO DE PANIC
	// ============================================
	fmt.Println("\n--- Uso Válido de Panic ---")

	// 1. Must* functions (para inicialización)
	template := MustParseTemplate("Hello, {{.Name}}!")
	fmt.Printf("Template parseado: %s\n", template)

	// 2. Assertion de tipo garantizado
	var valor any = "string"
	str := valor.(string) // panic si no es string
	fmt.Printf("Valor como string: %s\n", str)

	// 3. Índice garantizado
	arr := []int{1, 2, 3}
	ultimo := arr[len(arr)-1] // panic si slice vacío
	fmt.Printf("Último elemento: %d\n", ultimo)

	// ============================================
	// PATRÓN: CONVERTIR PANIC A ERROR
	// ============================================
	fmt.Println("\n--- Convertir Panic a Error ---")

	resultado2, err := ejecutarSeguro(func() any {
		panic("algo salió mal")
	})

	if err != nil {
		fmt.Printf("Error capturado: %v\n", err)
	} else {
		fmt.Printf("Resultado: %v\n", resultado2)
	}

	// Con función que no hace panic
	resultado3, err := ejecutarSeguro(func() any {
		return "éxito"
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Resultado: %v\n", resultado3)
	}

	// ============================================
	// PATRÓN: SERVIDOR HTTP RESILIENTE
	// ============================================
	fmt.Println("\n--- Servidor Resiliente ---")

	// En servidores web, cada request debe manejar sus panics
	requests := []string{"request1", "bad_request", "request2"}

	for _, req := range requests {
		handleRequest(req)
	}

	// ============================================
	// PATRÓN: TRANSACTION ROLLBACK
	// ============================================
	fmt.Println("\n--- Transaction Rollback ---")

	err = realizarTransaccion(func() error {
		fmt.Println("Operación 1: OK")
		fmt.Println("Operación 2: falla")
		panic("error crítico en operación 2")
	})

	if err != nil {
		fmt.Printf("Transacción falló: %v\n", err)
	}

	// ============================================
	// PROPAGACIÓN DE PANIC
	// ============================================
	fmt.Println("\n--- Propagación de Panic ---")

	// A veces quieres recuperar, procesar, y re-panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Nivel 2 capturó: %v\n", r)
				// No re-panic - lo manejamos aquí
			}
		}()

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Nivel 1 capturó y re-panic: %v\n", r)
					panic(r) // re-panic
				}
			}()

			panic("panic original")
		}()
	}()

	fmt.Println("\nPrograma terminó normalmente")
}

// ============================================
// FUNCIONES
// ============================================

func funcionSegura() string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recuperado internamente: %v\n", r)
		}
	}()

	// Simular código que puede hacer panic
	provocarPanic()

	return "éxito"
}

func provocarPanic() {
	panic("panic interno")
}

func causarPanic() {
	nivel1()
}

func nivel1() {
	nivel2()
}

func nivel2() {
	panic("panic en nivel 2")
}

// Must* pattern
func MustParseTemplate(text string) string {
	if text == "" {
		panic("template vacío")
	}
	// Simular parsing
	return text
}

// Ejecutar función con recover
func ejecutarSeguro(fn func() any) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recuperado: %v", r)
		}
	}()

	result = fn()
	return result, nil
}

// Handler de request simulado
func handleRequest(req string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  [ERROR] Request %s causó panic: %v\n", req, r)
		}
	}()

	fmt.Printf("Procesando: %s\n", req)

	if req == "bad_request" {
		panic("request inválido")
	}

	fmt.Printf("  [OK] %s completado\n", req)
}

// Transacción con rollback automático
func realizarTransaccion(fn func() error) (err error) {
	// Simular inicio de transacción
	fmt.Println("BEGIN TRANSACTION")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ROLLBACK (panic)")
			err = fmt.Errorf("panic en transacción: %v", r)
		} else if err != nil {
			fmt.Println("ROLLBACK (error)")
		} else {
			fmt.Println("COMMIT")
		}
	}()

	err = fn()
	return err
}

/*
RESUMEN:

PANIC:
- Detiene ejecución normal de la función
- Ejecuta defers en orden inverso
- Propaga hacia arriba en el call stack
- Termina el programa si no se recupera

RECOVER:
- Solo funciona dentro de defer
- Retorna el valor pasado a panic
- Retorna nil si no hubo panic
- Permite que el programa continúe

BUENAS PRÁCTICAS:

1. Usa error return, no panic, para errores esperados
2. Usa panic solo para errores de programación/invariantes
3. Must* functions son válidas para inicialización
4. Cada goroutine debe manejar sus propios panics
5. En servidores, recupera panics por request
6. Incluye stack trace en logs de panic
7. Considera re-panic después de logging

ANTI-PATRONES:
- Usar panic para control de flujo
- No recuperar panics en goroutines de servidor
- Ignorar panic sin logging
*/
