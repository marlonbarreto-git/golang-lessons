// Package main - Chapter 025: Error Handling Patterns
// Go tiene un sistema de errores explícito. Aquí aprenderás patrones
// avanzados para manejar errores de forma idiomática y robusta.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// ============================================
// CUSTOM ERROR TYPES (deben estar fuera de main)
// ============================================

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
}

type ContextError struct {
	Op      string
	Err     error
	Context map[string]any
}

func (e *ContextError) Error() string {
	return fmt.Sprintf("%s: %v (context: %v)", e.Op, e.Err, e.Context)
}

func (e *ContextError) Unwrap() error {
	return e.Err
}

// Sentinel errors
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
)

func main() {
	fmt.Println("=== PATRONES DE MANEJO DE ERRORES ===")

	// ============================================
	// EARLY RETURN (PATRÓN FUNDAMENTAL)
	// ============================================
	fmt.Println("\n--- Early Return ---")

	// BIEN: early return mantiene el happy path limpio
	processFile := func(path string) error {
		if path == "" {
			return errors.New("path cannot be empty")
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("getting file info: %w", err)
		}

		if info.Size() == 0 {
			return errors.New("file is empty")
		}

		fmt.Printf("File size: %d bytes\n", info.Size())
		return nil
	}

	if err := processFile("/etc/hosts"); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// ============================================
	// SENTINEL ERRORS
	// ============================================
	fmt.Println("\n--- Sentinel Errors ---")

	findUser := func(id int) error {
		if id <= 0 {
			return ErrInvalidInput
		}
		if id == 999 {
			return ErrUnauthorized
		}
		if id > 100 {
			return ErrNotFound
		}
		return nil
	}

	err := findUser(150)
	switch {
	case errors.Is(err, ErrNotFound):
		fmt.Println("Usuario no encontrado")
	case errors.Is(err, ErrUnauthorized):
		fmt.Println("No autorizado")
	case errors.Is(err, ErrInvalidInput):
		fmt.Println("Input inválido")
	case err != nil:
		fmt.Printf("Error desconocido: %v\n", err)
	default:
		fmt.Println("Usuario encontrado")
	}

	// ============================================
	// ERROR WRAPPING (%w)
	// ============================================
	fmt.Println("\n--- Error Wrapping ---")

	readConfig := func(path string) ([]byte, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading config %s: %w", path, err)
		}
		return data, nil
	}

	_, err = readConfig("/nonexistent/config.json")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("El archivo no existe")
		}
	}

	// ============================================
	// CUSTOM ERROR TYPES
	// ============================================
	fmt.Println("\n--- Custom Error Types ---")

	validate := func(email string) error {
		if email == "" {
			return &ValidationError{Field: "email", Message: "cannot be empty"}
		}
		if len(email) < 5 {
			return &ValidationError{Field: "email", Message: "too short"}
		}
		return nil
	}

	err = validate("")
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		fmt.Printf("Campo: %s, Mensaje: %s\n", valErr.Field, valErr.Message)
	}

	// ============================================
	// MULTI-ERROR
	// ============================================
	fmt.Println("\n--- Multi-Error (errors.Join) ---")

	validateUser := func(name, email string, age int) error {
		var errs []error
		if name == "" {
			errs = append(errs, errors.New("name is required"))
		}
		if email == "" {
			errs = append(errs, errors.New("email is required"))
		}
		if age < 0 {
			errs = append(errs, errors.New("age cannot be negative"))
		}
		return errors.Join(errs...)
	}

	err = validateUser("", "", -5)
	if err != nil {
		fmt.Printf("Validation errors:\n%v\n", err)
	}

	// ============================================
	// ERROR HANDLING EN DEFER
	// ============================================
	fmt.Println("\n--- Error Handling en Defer ---")

	copyFile := func(src, dst string) (err error) {
		srcFile, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("opening source: %w", err)
		}
		defer func() {
			closeErr := srcFile.Close()
			if err == nil {
				err = closeErr
			}
		}()

		dstFile, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("creating destination: %w", err)
		}
		defer func() {
			closeErr := dstFile.Close()
			if err == nil {
				err = closeErr
			}
		}()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return fmt.Errorf("copying: %w", err)
		}
		return nil
	}
	_ = copyFile

	// ============================================
	// RETRY PATTERN
	// ============================================
	fmt.Println("\n--- Retry Pattern ---")

	retry := func(attempts int, delay time.Duration, fn func() error) error {
		var lastErr error
		for i := 0; i < attempts; i++ {
			if err := fn(); err == nil {
				return nil
			} else {
				lastErr = err
				fmt.Printf("Attempt %d failed: %v\n", i+1, err)
				if i < attempts-1 {
					time.Sleep(delay)
				}
			}
		}
		return fmt.Errorf("all %d attempts failed: %w", attempts, lastErr)
	}

	counter := 0
	err = retry(3, 10*time.Millisecond, func() error {
		counter++
		if counter < 3 {
			return errors.New("temporary error")
		}
		return nil
	})
	fmt.Printf("Retry result: %v\n", err)

	// ============================================
	// MUST PATTERN
	// ============================================
	fmt.Println("\n--- Must Pattern ---")

	must := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	_ = must
	fmt.Println("Must pattern se usa para errores irrecuperables en init")

	// ============================================
	// ERROR CHECKING IDIOMÁTICO
	// ============================================
	fmt.Println("\n--- Error Checking Idiomático ---")

	if f, err := os.Open("/etc/hosts"); err != nil {
		fmt.Printf("Error abriendo: %v\n", err)
	} else {
		fmt.Printf("Archivo abierto: %s\n", f.Name())
		f.Close()
	}

	var errNoRows = sql.ErrNoRows
	getUser := func(id int) (string, error) {
		if id == 0 {
			return "", errNoRows
		}
		return "user", nil
	}

	name, err := getUser(0)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("Usuario no encontrado en DB")
	} else if err != nil {
		fmt.Printf("Error de DB: %v\n", err)
	} else {
		fmt.Printf("Usuario: %s\n", name)
	}

	// ============================================
	// ERROR CONTEXT PATTERN
	// ============================================
	fmt.Println("\n--- Error Context Pattern ---")

	wrapWithContext := func(op string, err error, ctx map[string]any) error {
		if err == nil {
			return nil
		}
		return &ContextError{Op: op, Err: err, Context: ctx}
	}

	err = wrapWithContext("fetchUser", ErrNotFound, map[string]any{
		"userID":    123,
		"requestID": "abc-123",
	})
	fmt.Printf("Error con contexto: %v\n", err)
}

/*
RESUMEN DE PATRONES:

EARLY RETURN:
- Manejar errores primero y retornar
- El happy path queda sin indentación

SENTINEL ERRORS:
var ErrNotFound = errors.New("not found")
- Para comparación con errors.Is()

ERROR WRAPPING:
return fmt.Errorf("context: %w", err)
- Agrega contexto manteniendo el error original

CUSTOM ERRORS:
type MyError struct { ... }
func (e *MyError) Error() string { ... }
- Para errores con datos estructurados

MULTI-ERROR (Go 1.20+):
errors.Join(err1, err2, err3)
- Combina múltiples errores

MUST PATTERN:
- Para inicialización donde error = fatal

RETRY PATTERN:
- Reintentar operaciones que pueden fallar
*/

/*
SUMMARY - ERROR HANDLING PATTERNS:

EARLY RETURN:
- Manejar errores primero con return inmediato
- Happy path sin indentación
- Cada paso del pipeline verifica y wrappea

SENTINEL ERRORS:
- var ErrNotFound = errors.New("not found")
- Switch con errors.Is() para manejar distintos sentinels
- Stdlib: os.ErrNotExist, sql.ErrNoRows

ERROR WRAPPING:
- fmt.Errorf("context: %w", err) preserva el original
- errors.Is() busca a través de toda la cadena wrap
- errors.As() extrae tipo específico de la cadena

CUSTOM ERROR TYPES:
- ValidationError con Field y Message
- ContextError con Op, Err y Context map
- Implementar Unwrap() para participar en la cadena

MULTI-ERROR (Go 1.20+):
- errors.Join(errs...) combina múltiples errores
- errors.Is() funciona con cualquiera de los combinados

DEFER CON ERROR:
- Named return (err error) permite modificar en defer
- Capturar errores de Close() sin perder el error original

RETRY PATTERN:
- Reintentar N veces con delay configurable
- Retornar último error si todos los intentos fallan

MUST PATTERN:
- log.Fatal(err) para errores irrecuperables en init
- Solo para inicialización del programa
*/
