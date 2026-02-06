// Package main - Chapter 004: Conditionals
// Aprenderás if/else, el patrón para simular operador ternario,
// if con inicialización, y el patrón de early return.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func main() {
	fmt.Println("=== CONDICIONALES EN GO ===")
	// ============================================
	// IF BÁSICO
	// ============================================
	fmt.Println("--- If Básico ---")

	x := 10

	if x > 5 {
		fmt.Println("x es mayor que 5")
	}

	// Los paréntesis son opcionales (y no idiomáticos)
	if (x > 5) { // funciona pero no es el estilo Go
		fmt.Println("Con paréntesis (no idiomático)")
	}

	// Las llaves son OBLIGATORIAS
	// if x > 5 fmt.Println("error") // NO COMPILA

	// ============================================
	// IF-ELSE
	// ============================================
	fmt.Println("\n--- If-Else ---")

	edad := 17

	if edad >= 18 {
		fmt.Println("Eres mayor de edad")
	} else {
		fmt.Println("Eres menor de edad")
	}

	// ============================================
	// IF-ELSE IF-ELSE
	// ============================================
	fmt.Println("\n--- If-Else If-Else ---")

	nota := 85

	if nota >= 90 {
		fmt.Println("Calificación: A")
	} else if nota >= 80 {
		fmt.Println("Calificación: B")
	} else if nota >= 70 {
		fmt.Println("Calificación: C")
	} else if nota >= 60 {
		fmt.Println("Calificación: D")
	} else {
		fmt.Println("Calificación: F")
	}

	// ============================================
	// IF CON INICIALIZACIÓN (MUY IMPORTANTE)
	// ============================================
	fmt.Println("\n--- If con Inicialización ---")

	// Puedes declarar variables en el if, su scope es solo el if/else
	if valor := calcular(10); valor > 50 {
		fmt.Printf("Valor %d es mayor que 50\n", valor)
	} else {
		fmt.Printf("Valor %d es menor o igual a 50\n", valor)
	}
	// valor no existe aquí - fuera de scope

	// Muy común con errores
	if err := procesarDatos(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Muy común con el patrón "comma ok"
	mapa := map[string]int{"a": 1, "b": 2}
	if valor, existe := mapa["c"]; existe {
		fmt.Printf("Valor encontrado: %d\n", valor)
	} else {
		fmt.Println("Clave 'c' no existe en el mapa")
	}

	// ============================================
	// PATRÓN TERNARIO (No existe en Go)
	// ============================================
	fmt.Println("\n--- Patrón Ternario ---")

	// En otros lenguajes: result = condition ? valor1 : valor2
	// En Go, usamos if-else

	// Forma 1: Variable previa
	activo := true
	var estado string
	if activo {
		estado = "Activo"
	} else {
		estado = "Inactivo"
	}
	fmt.Printf("Estado: %s\n", estado)

	// Forma 2: Función helper (para casos muy usados)
	maximo := max(10, 20)
	fmt.Printf("Máximo de 10 y 20: %d\n", maximo)

	// Forma 3: Función inline (no recomendada generalmente)
	resultado := func() string {
		if activo {
			return "Sí"
		}
		return "No"
	}()
	fmt.Printf("¿Activo? %s\n", resultado)

	// ============================================
	// EARLY RETURN (PATRÓN MUY IMPORTANTE)
	// ============================================
	fmt.Println("\n--- Early Return ---")

	// MAL: código anidado difícil de leer
	fmt.Println("Ejemplo MAL (anidado):")
	resultadoMal := procesarUsuarioMal("admin", 25, true)
	fmt.Printf("Resultado: %s\n", resultadoMal)

	// BIEN: early return, código plano
	fmt.Println("\nEjemplo BIEN (early return):")
	resultadoBien := procesarUsuarioBien("admin", 25, true)
	fmt.Printf("Resultado: %s\n", resultadoBien)

	// ============================================
	// COMPARACIONES COMUNES
	// ============================================
	fmt.Println("\n--- Comparaciones Comunes ---")

	// Comparar con nil
	var ptr *int
	if ptr == nil {
		fmt.Println("Puntero es nil")
	}

	// Comparar strings
	nombre := "Gopher"
	if nombre == "" {
		fmt.Println("Nombre vacío")
	} else {
		fmt.Printf("Nombre: %s\n", nombre)
	}

	// Comparar slices con nil (NO con otro slice)
	var slice []int
	if slice == nil {
		fmt.Println("Slice es nil")
	}
	// if slice == []int{} // NO COMPILA - slices no son comparables

	// Comparar interfaces con nil
	var err error
	if err == nil {
		fmt.Println("No hay error")
	}

	// ============================================
	// BOOLEAN EXPRESSIONS
	// ============================================
	fmt.Println("\n--- Expresiones Booleanas ---")

	a, b, c := true, false, true

	// AND - todos deben ser true
	if a && b && c {
		fmt.Println("Todos son true")
	}

	// OR - al menos uno true
	if a || b || c {
		fmt.Println("Al menos uno es true")
	}

	// Combinaciones
	if (a && c) || b {
		fmt.Println("(a AND c) OR b es true")
	}

	// Negación
	if !b {
		fmt.Println("b es false")
	}

	// ============================================
	// IF CON MÚLTIPLES CONDICIONES
	// ============================================
	fmt.Println("\n--- Múltiples Condiciones ---")

	usuario := "admin"
	password := "secret123"
	intentos := 2

	if usuario == "admin" && password == "secret123" && intentos < 3 {
		fmt.Println("Login exitoso")
	}

	// Para condiciones complejas, es mejor separarlas
	credencialesValidas := usuario == "admin" && password == "secret123"
	intentosPermitidos := intentos < 3

	if credencialesValidas && intentosPermitidos {
		fmt.Println("Login exitoso (código más legible)")
	}

	// ============================================
	// ERRORES COMUNES
	// ============================================
	fmt.Println("\n--- Errores Comunes ---")

	// 1. Asignación vs comparación (Go no permite esto, es seguro)
	// if x = 5 { } // NO COMPILA - necesita ==

	// 2. Variables shadowing en if con inicialización
	valor := 100
	if valor := calcular(5); valor > 20 {
		fmt.Printf("Dentro del if: valor = %d\n", valor) // valor = 50
	}
	fmt.Printf("Fuera del if: valor = %d\n", valor) // valor = 100 (original)

	// ============================================
	// EJEMPLO PRÁCTICO: VALIDACIÓN
	// ============================================
	fmt.Println("\n--- Ejemplo: Validación de Formulario ---")

	email := "usuario@ejemplo.com"
	pass := "abc123"
	edad2 := 25

	errores := validarFormulario(email, pass, edad2)
	if len(errores) > 0 {
		fmt.Println("Errores de validación:")
		for _, e := range errores {
			fmt.Printf("  - %s\n", e)
		}
	} else {
		fmt.Println("Formulario válido")
	}
}

func calcular(n int) int {
	return n * 5
}

func procesarDatos() error {
	// Simula un error
	return errors.New("datos inválidos")
}

// MAL: código anidado
func procesarUsuarioMal(nombre string, edad int, activo bool) string {
	if nombre != "" {
		if edad >= 18 {
			if activo {
				return "Usuario válido"
			} else {
				return "Usuario inactivo"
			}
		} else {
			return "Usuario menor de edad"
		}
	} else {
		return "Nombre requerido"
	}
}

// BIEN: early return (Guard Clauses)
func procesarUsuarioBien(nombre string, edad int, activo bool) string {
	if nombre == "" {
		return "Nombre requerido"
	}

	if edad < 18 {
		return "Usuario menor de edad"
	}

	if !activo {
		return "Usuario inactivo"
	}

	return "Usuario válido"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func validarFormulario(email, password string, edad int) []string {
	var errores []string

	if email == "" {
		errores = append(errores, "Email requerido")
	}

	if len(password) < 6 {
		errores = append(errores, "Password debe tener al menos 6 caracteres")
	}

	if edad < 0 || edad > 150 {
		errores = append(errores, "Edad inválida")
	}

	return errores
}

// ============================================
// EJEMPLO CON OS.ARGS
// ============================================
func ejemploArgs() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: programa <número>")
		os.Exit(1)
	}

	num, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %s no es un número válido\n", os.Args[1])
		os.Exit(1)
	}

	if num%2 == 0 {
		fmt.Printf("%d es par\n", num)
	} else {
		fmt.Printf("%d es impar\n", num)
	}
}

/*
BUENAS PRÁCTICAS:

1. EARLY RETURN: Siempre preferir retornos tempranos sobre anidación
   - Maneja casos de error/inválidos primero
   - El código principal queda sin anidar

2. IF CON INICIALIZACIÓN: Usar cuando la variable solo se necesita en el if
   - Limita el scope de variables
   - Código más limpio

3. LEGIBILIDAD:
   - Extraer condiciones complejas a variables con nombres descriptivos
   - No tener más de 2-3 niveles de anidación

4. EVITAR:
   - Código muy anidado (pyramid of doom)
   - Comparaciones innecesarias (if b == true → if b)
   - else después de return (código muerto)

ANTI-PATRÓN:
if err != nil {
    return err
} else {
    // código
}

MEJOR:
if err != nil {
    return err
}
// código
*/
