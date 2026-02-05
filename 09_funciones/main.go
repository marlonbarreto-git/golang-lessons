// Package main - Capítulo 09: Funciones
// Las funciones son ciudadanos de primera clase en Go.
// Aprenderás declaración, parámetros, retornos, closures, y defer.
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== FUNCIONES EN GO ===")

	// ============================================
	// FUNCIONES BÁSICAS
	// ============================================
	fmt.Println("\n--- Funciones Básicas ---")

	saludo := saludar("Gopher")
	fmt.Println(saludo)

	resultado := sumar(5, 3)
	fmt.Printf("5 + 3 = %d\n", resultado)

	// ============================================
	// MÚLTIPLES VALORES DE RETORNO
	// ============================================
	fmt.Println("\n--- Múltiples Retornos ---")

	cociente, resto := dividir(17, 5)
	fmt.Printf("17 / 5 = %d, resto = %d\n", cociente, resto)

	// Ignorar un valor con _
	soloResto, _ := dividir(17, 5)
	fmt.Printf("Solo cociente: %d\n", soloResto)

	// ============================================
	// RETORNOS NOMBRADOS
	// ============================================
	fmt.Println("\n--- Retornos Nombrados ---")

	area, perimetro := dimensionesRectangulo(5, 3)
	fmt.Printf("Rectángulo 5x3: área=%d, perímetro=%d\n", area, perimetro)

	// ============================================
	// PARÁMETROS VARIÁDICOS
	// ============================================
	fmt.Println("\n--- Parámetros Variádicos ---")

	fmt.Printf("sumaVariadica(): %d\n", sumaVariadica())
	fmt.Printf("sumaVariadica(1): %d\n", sumaVariadica(1))
	fmt.Printf("sumaVariadica(1,2,3): %d\n", sumaVariadica(1, 2, 3))
	fmt.Printf("sumaVariadica(1,2,3,4,5): %d\n", sumaVariadica(1, 2, 3, 4, 5))

	// Pasar slice como variadic
	nums := []int{10, 20, 30}
	fmt.Printf("sumaVariadica(slice...): %d\n", sumaVariadica(nums...))

	// printf es variádico
	fmt.Printf("Printf es %s: %d + %d = %d\n", "variádico", 1, 2, 3)

	// ============================================
	// FUNCIONES COMO VALORES
	// ============================================
	fmt.Println("\n--- Funciones como Valores ---")

	// Las funciones son first-class citizens
	var operacion func(int, int) int
	operacion = sumar
	fmt.Printf("operacion(5, 3) = %d\n", operacion(5, 3))

	operacion = multiplicar
	fmt.Printf("operacion(5, 3) = %d\n", operacion(5, 3))

	// Funciones como parámetros
	resultadoOp := aplicarOperacion(10, 5, sumar)
	fmt.Printf("aplicarOperacion(10, 5, sumar) = %d\n", resultadoOp)

	resultadoOp = aplicarOperacion(10, 5, multiplicar)
	fmt.Printf("aplicarOperacion(10, 5, multiplicar) = %d\n", resultadoOp)

	// ============================================
	// FUNCIONES ANÓNIMAS
	// ============================================
	fmt.Println("\n--- Funciones Anónimas ---")

	// Función anónima asignada a variable
	doble := func(n int) int {
		return n * 2
	}
	fmt.Printf("doble(5) = %d\n", doble(5))

	// Función anónima ejecutada inmediatamente (IIFE)
	resultado2 := func(a, b int) int {
		return a + b
	}(3, 4)
	fmt.Printf("IIFE(3, 4) = %d\n", resultado2)

	// ============================================
	// CLOSURES
	// ============================================
	fmt.Println("\n--- Closures ---")

	// Los closures capturan variables del entorno
	contador := crearContador()
	fmt.Printf("contador(): %d\n", contador())
	fmt.Printf("contador(): %d\n", contador())
	fmt.Printf("contador(): %d\n", contador())

	// Cada closure tiene su propia copia
	contador2 := crearContador()
	fmt.Printf("contador2(): %d (nuevo contador)\n", contador2())

	// Closure con parámetro inicial
	desde10 := crearContadorDesde(10)
	fmt.Printf("desde10(): %d\n", desde10())
	fmt.Printf("desde10(): %d\n", desde10())

	// ============================================
	// DEFER
	// ============================================
	fmt.Println("\n--- Defer ---")

	// defer ejecuta al final de la función (LIFO)
	demostrarDefer()

	// defer común: cerrar recursos
	contenido := leerArchivo("config.txt")
	fmt.Printf("Contenido: %s\n", contenido)

	// defer evalúa argumentos inmediatamente
	demostrarDeferEvaluacion()

	// ============================================
	// FUNCIONES QUE RETORNAN FUNCIONES
	// ============================================
	fmt.Println("\n--- Funciones que Retornan Funciones ---")

	multiplicarPor := crearMultiplicador(3)
	fmt.Printf("multiplicarPor3(5) = %d\n", multiplicarPor(5))
	fmt.Printf("multiplicarPor3(10) = %d\n", multiplicarPor(10))

	// Factory de operaciones
	suma := crearOperacion("+")
	resta := crearOperacion("-")
	fmt.Printf("suma(10, 3) = %d\n", suma(10, 3))
	fmt.Printf("resta(10, 3) = %d\n", resta(10, 3))

	// ============================================
	// RECURSIÓN
	// ============================================
	fmt.Println("\n--- Recursión ---")

	fmt.Printf("factorial(5) = %d\n", factorial(5))
	fmt.Printf("fibonacci(10) = %d\n", fibonacci(10))

	// ============================================
	// MÉTODOS (preview - se profundiza en structs)
	// ============================================
	fmt.Println("\n--- Métodos (preview) ---")

	rect := Rectangulo{Ancho: 5, Alto: 3}
	fmt.Printf("Rectángulo: %+v\n", rect)
	fmt.Printf("Área: %d\n", rect.Area())
	fmt.Printf("Perímetro: %d\n", rect.Perimetro())

	// ============================================
	// FUNCIONES GENÉRICAS (Go 1.18+)
	// ============================================
	fmt.Println("\n--- Funciones Genéricas ---")

	fmt.Printf("minimo(5, 3) = %d\n", minimo(5, 3))
	fmt.Printf("minimo(5.5, 3.2) = %.1f\n", minimo(5.5, 3.2))
	fmt.Printf("minimo(\"banana\", \"apple\") = %s\n", minimo("banana", "apple"))

	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrones Comunes ---")

	// 1. Función con error (el patrón más común en Go)
	valor, err := dividirSeguro(10, 0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Resultado: %d\n", valor)
	}

	// 2. Opciones funcionales
	srv := NuevoServidor(
		ConPuerto(8080),
		ConTimeout(30),
	)
	fmt.Printf("Servidor: %+v\n", srv)

	// 3. Middleware/decoradores
	operacionLenta := medirTiempo(func() int {
		time.Sleep(10 * time.Millisecond)
		return 42
	})
	res := operacionLenta()
	fmt.Printf("Resultado: %d\n", res)
}

// Funciones básicas
func saludar(nombre string) string {
	return "¡Hola, " + nombre + "!"
}

func sumar(a, b int) int {
	return a + b
}

func multiplicar(a, b int) int {
	return a * b
}

// Múltiples retornos
func dividir(dividendo, divisor int) (int, int) {
	cociente := dividendo / divisor
	resto := dividendo % divisor
	return cociente, resto
}

// Retornos nombrados
func dimensionesRectangulo(ancho, alto int) (area, perimetro int) {
	area = ancho * alto
	perimetro = 2 * (ancho + alto)
	return // "naked return" - retorna las variables nombradas
}

// Variádica
func sumaVariadica(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Función como parámetro
func aplicarOperacion(a, b int, op func(int, int) int) int {
	return op(a, b)
}

// Closures
func crearContador() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func crearContadorDesde(inicio int) func() int {
	return func() int {
		inicio++
		return inicio
	}
}

// Funciones que retornan funciones
func crearMultiplicador(factor int) func(int) int {
	return func(n int) int {
		return n * factor
	}
}

func crearOperacion(op string) func(int, int) int {
	switch op {
	case "+":
		return func(a, b int) int { return a + b }
	case "-":
		return func(a, b int) int { return a - b }
	case "*":
		return func(a, b int) int { return a * b }
	default:
		return func(a, b int) int { return 0 }
	}
}

// Defer
func demostrarDefer() {
	fmt.Println("Inicio")
	defer fmt.Println("Defer 1 (último en ejecutar)")
	defer fmt.Println("Defer 2 (segundo)")
	defer fmt.Println("Defer 3 (primero)")
	fmt.Println("Fin (antes de defers)")
}

func leerArchivo(nombre string) string {
	fmt.Printf("Abriendo archivo: %s\n", nombre)
	defer fmt.Println("Cerrando archivo (defer)")

	return "contenido del archivo"
}

func demostrarDeferEvaluacion() {
	x := 10
	defer fmt.Printf("Defer con x=%d (evaluado al momento del defer)\n", x)
	x = 20
	fmt.Printf("x actual: %d\n", x)
}

// Recursión
func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// Métodos (preview)
type Rectangulo struct {
	Ancho, Alto int
}

func (r Rectangulo) Area() int {
	return r.Ancho * r.Alto
}

func (r Rectangulo) Perimetro() int {
	return 2 * (r.Ancho + r.Alto)
}

// Genéricos
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

func minimo[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Patrón error
func dividirSeguro(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("división por cero")
	}
	return a / b, nil
}

// Opciones funcionales
type Servidor struct {
	Puerto  int
	Timeout int
}

type OpcionServidor func(*Servidor)

func ConPuerto(puerto int) OpcionServidor {
	return func(s *Servidor) {
		s.Puerto = puerto
	}
}

func ConTimeout(timeout int) OpcionServidor {
	return func(s *Servidor) {
		s.Timeout = timeout
	}
}

func NuevoServidor(opciones ...OpcionServidor) *Servidor {
	srv := &Servidor{
		Puerto:  80,  // default
		Timeout: 10,  // default
	}
	for _, opt := range opciones {
		opt(srv)
	}
	return srv
}

// Middleware/decorador
func medirTiempo(fn func() int) func() int {
	return func() int {
		inicio := time.Now()
		resultado := fn()
		fmt.Printf("Tiempo: %v\n", time.Since(inicio))
		return resultado
	}
}

/*
RESUMEN:

SINTAXIS:
func nombre(params) retorno { }
func nombre(params) (retorno1, retorno2) { }
func nombre(params) (nombre1, nombre2 tipo) { }
func (receiver Tipo) Metodo(params) retorno { }

CARACTERÍSTICAS:
- First-class citizens (pueden asignarse, pasarse, retornarse)
- Múltiples valores de retorno
- Retornos nombrados (usar con moderación)
- Parámetros variádicos (...)
- Closures (capturan variables del entorno)
- defer (LIFO, para cleanup)

BUENAS PRÁCTICAS:

1. El último parámetro de retorno debe ser error (si hay error)
2. Usa retornos nombrados solo cuando mejoren claridad
3. defer para cerrar recursos (archivos, conexiones, mutex)
4. Closures para encapsular estado
5. Opciones funcionales para configuración flexible
6. Early return para código más legible
*/
