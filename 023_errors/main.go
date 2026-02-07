// Package main - Chapter 023: Errors
// El manejo de errores en Go es explícito y basado en valores.
// Aprenderás la interface error, errors package, y patrones comunes.
package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
)

func main() {
	fmt.Println("=== MANEJO DE ERRORES EN GO ===")

	// ============================================
	// ERROR INTERFACE
	// ============================================
	fmt.Println("\n--- Error Interface ---")

	// error es una interface con un solo método:
	// type error interface {
	//     Error() string
	// }

	err := errors.New("algo salió mal")
	fmt.Printf("Error: %v\n", err)
	fmt.Printf("Tipo: %T\n", err)

	// ============================================
	// CREANDO ERRORES
	// ============================================
	fmt.Println("\n--- Creando Errores ---")

	// 1. errors.New - para errores simples
	err1 := errors.New("error simple")
	fmt.Printf("errors.New: %v\n", err1)

	// 2. fmt.Errorf - para errores con formato
	nombre := "archivo.txt"
	err2 := fmt.Errorf("no se pudo abrir %s", nombre)
	fmt.Printf("fmt.Errorf: %v\n", err2)

	// 3. Error personalizado (implementando interface)
	err3 := &MiError{
		Codigo:  404,
		Mensaje: "recurso no encontrado",
	}
	fmt.Printf("Error personalizado: %v\n", err3)

	// ============================================
	// PATRÓN BÁSICO DE MANEJO
	// ============================================
	fmt.Println("\n--- Patrón Básico ---")

	// El patrón más común: verificar error inmediatamente
	resultado, err := dividir(10, 2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Resultado: %d\n", resultado)
	}

	// Con división por cero
	_, err = dividir(10, 0)
	if err != nil {
		fmt.Printf("Error esperado: %v\n", err)
	}

	// ============================================
	// EARLY RETURN CON ERRORES
	// ============================================
	fmt.Println("\n--- Early Return ---")

	if err := procesarArchivo("config.txt"); err != nil {
		fmt.Printf("Error procesando: %v\n", err)
	}

	// ============================================
	// SENTINEL ERRORS
	// ============================================
	fmt.Println("\n--- Sentinel Errors ---")

	// Errores predefinidos que se pueden comparar
	_, err = encontrarUsuario(0)
	if err == ErrUsuarioNoEncontrado {
		fmt.Println("Usuario no existe (sentinel error)")
	}

	_, err = encontrarUsuario(-1)
	if err == ErrIDInvalido {
		fmt.Println("ID inválido (sentinel error)")
	}

	// Sentinels de la stdlib
	_, err = os.Open("archivo_que_no_existe.txt")
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println("Archivo no existe (fs.ErrNotExist)")
	}

	// ============================================
	// ERROR WRAPPING (Go 1.13+)
	// ============================================
	fmt.Println("\n--- Error Wrapping ---")

	// Envolver errores para agregar contexto
	err = abrirConfiguracion("config.json")
	if err != nil {
		fmt.Printf("Error con contexto: %v\n", err)
	}

	// %w en fmt.Errorf crea un wrapped error
	originalErr := errors.New("conexión rechazada")
	wrappedErr := fmt.Errorf("fallo al conectar a DB: %w", originalErr)
	fmt.Printf("Wrapped: %v\n", wrappedErr)

	// ============================================
	// ERRORS.IS
	// ============================================
	fmt.Println("\n--- errors.Is ---")

	// errors.Is verifica si un error es o wrappea otro error
	err = fmt.Errorf("capa 2: %w", fmt.Errorf("capa 1: %w", ErrUsuarioNoEncontrado))

	if errors.Is(err, ErrUsuarioNoEncontrado) {
		fmt.Println("Es (o contiene) ErrUsuarioNoEncontrado")
	}

	// Funciona con múltiples niveles de wrapping
	fmt.Printf("Cadena completa: %v\n", err)

	// ============================================
	// ERRORS.AS
	// ============================================
	fmt.Println("\n--- errors.As ---")

	// errors.As extrae un error de tipo específico
	err = realizarOperacion()

	var miErr *MiError
	if errors.As(err, &miErr) {
		fmt.Printf("Código: %d, Mensaje: %s\n", miErr.Codigo, miErr.Mensaje)
	}

	// Con errores del sistema
	_, err = os.Open("/root/secreto.txt")
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		fmt.Printf("PathError - Op: %s, Path: %s\n", pathErr.Op, pathErr.Path)
	}

	// ============================================
	// ERRORS.JOIN (Go 1.20+)
	// ============================================
	fmt.Println("\n--- errors.Join (Go 1.20+) ---")

	// Combinar múltiples errores
	err1 = errors.New("error 1")
	err2 = errors.New("error 2")
	err3General := errors.New("error 3")

	combinado := errors.Join(err1, err2, err3General)
	fmt.Printf("Errores combinados: %v\n", combinado)

	// Is funciona con cualquiera de los errores
	fmt.Printf("Contiene error 1: %t\n", errors.Is(combinado, err1))
	fmt.Printf("Contiene error 2: %t\n", errors.Is(combinado, err2))

	// ============================================
	// CUSTOM ERRORS CON UNWRAP
	// ============================================
	fmt.Println("\n--- Custom Errors con Unwrap ---")

	err = &ErrorConContexto{
		Op:    "leer",
		Path:  "/etc/config",
		Causa: fs.ErrPermission,
	}

	fmt.Printf("Error: %v\n", err)

	// Unwrap permite extraer la causa
	if errors.Is(err, fs.ErrPermission) {
		fmt.Println("Es un error de permisos")
	}

	// ============================================
	// VALIDACIÓN CON MÚLTIPLES ERRORES
	// ============================================
	fmt.Println("\n--- Validación con Múltiples Errores ---")

	errs := validarUsuario("", "ab", -5)
	if errs != nil {
		fmt.Println("Errores de validación:")
		for _, e := range errs {
			fmt.Printf("  - %v\n", e)
		}
	}

	// ============================================
	// PATRÓN: DEFER CON ERROR
	// ============================================
	fmt.Println("\n--- Defer con Error ---")

	err = operacionConCleanup()
	fmt.Printf("Resultado: %v\n", err)

	// ============================================
	// PATRÓN: ERROR AS CONTROL FLOW
	// ============================================
	fmt.Println("\n--- Error como Control de Flujo ---")

	_, err = strconv.Atoi("no es número")
	if err != nil {
		var numErr *strconv.NumError
		if errors.As(err, &numErr) {
			fmt.Printf("No se pudo convertir %q: %v\n", numErr.Num, numErr.Err)
		}
	}

	// ============================================
	// ANTI-PATRONES
	// ============================================
	fmt.Println("\n--- Anti-Patrones (evitar) ---")
	fmt.Println(`
❌ Ignorar errores:
   result, _ := funcion()  // ¡NO!

❌ Solo loguear sin manejar:
   if err != nil { log.Println(err) }  // el programa continúa mal

❌ Crear strings para comparar:
   if err.Error() == "not found" { }  // frágil

❌ Panic para errores esperados:
   if err != nil { panic(err) }  // solo para errores irrecuperables

✅ Patrones correctos:
   if err != nil { return fmt.Errorf("context: %w", err) }
   if errors.Is(err, ErrNotFound) { ... }
   var myErr *MyError; if errors.As(err, &myErr) { ... }`)}

// ============================================
// TIPOS Y FUNCIONES
// ============================================

// Sentinel errors (variables de error predefinidas)
var (
	ErrUsuarioNoEncontrado = errors.New("usuario no encontrado")
	ErrIDInvalido          = errors.New("ID inválido")
	ErrDivisionPorCero     = errors.New("división por cero")
)

// Error personalizado
type MiError struct {
	Codigo  int
	Mensaje string
}

func (e *MiError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Codigo, e.Mensaje)
}

// Error con Unwrap (para wrapping chain)
type ErrorConContexto struct {
	Op    string
	Path  string
	Causa error
}

func (e *ErrorConContexto) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Causa)
}

func (e *ErrorConContexto) Unwrap() error {
	return e.Causa
}

// Funciones que retornan errores
func dividir(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivisionPorCero
	}
	return a / b, nil
}

func encontrarUsuario(id int) (string, error) {
	if id < 0 {
		return "", ErrIDInvalido
	}
	if id == 0 {
		return "", ErrUsuarioNoEncontrado
	}
	return "Usuario", nil
}

func procesarArchivo(nombre string) error {
	// Simulación de múltiples operaciones
	if nombre == "" {
		return errors.New("nombre de archivo vacío")
	}

	// Early return en cada paso
	if err := validarNombre(nombre); err != nil {
		return fmt.Errorf("validación fallida: %w", err)
	}

	if err := cargarArchivo(nombre); err != nil {
		return fmt.Errorf("carga fallida: %w", err)
	}

	return nil
}

func validarNombre(nombre string) error {
	if len(nombre) < 3 {
		return errors.New("nombre muy corto")
	}
	return nil
}

func cargarArchivo(nombre string) error {
	// Simular que el archivo no existe
	return fmt.Errorf("archivo %s no encontrado", nombre)
}

func abrirConfiguracion(path string) error {
	_, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("abrirConfiguracion %s: %w", path, err)
	}
	return nil
}

func realizarOperacion() error {
	// Simular error en operación interna
	err := &MiError{Codigo: 500, Mensaje: "error interno"}
	return fmt.Errorf("operación fallida: %w", err)
}

func validarUsuario(nombre, password string, edad int) []error {
	var errs []error

	if nombre == "" {
		errs = append(errs, errors.New("nombre requerido"))
	}

	if len(password) < 6 {
		errs = append(errs, errors.New("password muy corta (min 6 chars)"))
	}

	if edad < 0 {
		errs = append(errs, errors.New("edad no puede ser negativa"))
	}

	return errs
}

func operacionConCleanup() (err error) {
	// Named return para modificar en defer
	fmt.Println("Iniciando operación...")

	defer func() {
		fmt.Println("Limpieza...")
		// Podríamos modificar err aquí si la limpieza falla
	}()

	// Simular operación
	return errors.New("operación falló")
}

/*
RESUMEN:

INTERFACE ERROR:
type error interface {
    Error() string
}

CREAR ERRORES:
- errors.New("mensaje")
- fmt.Errorf("formato %v", valor)
- fmt.Errorf("contexto: %w", err) // wrapping

SENTINEL ERRORS:
var ErrNotFound = errors.New("not found")

VERIFICAR ERRORES:
- err == ErrNotFound (comparación directa, solo sin wrap)
- errors.Is(err, ErrNotFound) (funciona con wrap)
- errors.As(err, &target) (extrae tipo específico)

WRAPPING:
- fmt.Errorf("contexto: %w", err)
- Implementar Unwrap() error

BUENAS PRÁCTICAS:

1. Siempre verificar errores
2. Agregar contexto al propagar: fmt.Errorf("ctx: %w", err)
3. Usar sentinel errors para errores conocidos
4. Usar errors.Is/As en lugar de comparación directa
5. Retornar error como último valor
6. No ignorar errores con _
7. Early return al encontrar errores
*/

/*
SUMMARY - ERRORS:

ERROR INTERFACE:
- type error interface { Error() string }
- Cualquier tipo que implemente Error() es un error

CREAR ERRORES:
- errors.New("mensaje") para errores simples
- fmt.Errorf("formato %v", val) con formato
- Tipos personalizados: type MiError struct { Codigo int }

SENTINEL ERRORS:
- var ErrNotFound = errors.New("not found")
- Errores predefinidos para comparación
- Stdlib: fs.ErrNotExist, fs.ErrPermission, sql.ErrNoRows

ERROR WRAPPING (Go 1.13+):
- fmt.Errorf("contexto: %w", err) agrega contexto
- Implementar Unwrap() error en custom errors
- errors.Join(err1, err2) combina múltiples (Go 1.20+)

VERIFICAR ERRORES:
- errors.Is(err, target) verifica a través de la cadena wrap
- errors.As(err, &target) extrae tipo específico
- Preferir Is/As sobre comparación directa

PATRÓN BÁSICO:
- if err != nil { return fmt.Errorf("ctx: %w", err) }
- Early return al encontrar errores
- Error como último valor de retorno

VALIDACIÓN:
- Retornar []error para múltiples errores de validación
- errors.Join para combinar en un solo error

ANTI-PATRONES:
- Ignorar errores con _
- Comparar err.Error() como string (frágil)
- Panic para errores esperados
*/
