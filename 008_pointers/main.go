// Package main - Chapter 008: Pointers
// Los punteros permiten compartir memoria y modificar datos eficientemente.
// Aprenderás direcciones de memoria, dereferencia, nil, y cuándo usarlos.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== PUNTEROS EN GO ===")

	// ============================================
	// CONCEPTOS BÁSICOS
	// ============================================
	fmt.Println("\n--- Conceptos Básicos ---")

	// Una variable tiene un valor y una dirección en memoria
	x := 42
	fmt.Printf("x valor: %d\n", x)
	fmt.Printf("x dirección: %p\n", &x) // & obtiene la dirección

	// Un puntero almacena una dirección de memoria
	var p *int // puntero a int (zero value es nil)
	p = &x     // p ahora apunta a x

	fmt.Printf("p (dirección que guarda): %p\n", p)
	fmt.Printf("*p (valor en esa dirección): %d\n", *p) // * dereferencia

	// Modificar a través del puntero
	*p = 100
	fmt.Printf("Después de *p = 100: x = %d\n", x) // x cambió

	// ============================================
	// DECLARACIÓN DE PUNTEROS
	// ============================================
	fmt.Println("\n--- Declaración de Punteros ---")

	// Forma 1: var
	var ptr1 *int // nil pointer

	// Forma 2: new() - aloca memoria y retorna puntero
	ptr2 := new(int) // apunta a un int con zero value (0)
	*ptr2 = 50

	// Forma 3: dirección de variable existente
	valor := 30
	ptr3 := &valor

	fmt.Printf("ptr1 (nil): %v\n", ptr1)
	fmt.Printf("ptr2 (new): %p, valor: %d\n", ptr2, *ptr2)
	fmt.Printf("ptr3 (&var): %p, valor: %d\n", ptr3, *ptr3)

	// ============================================
	// NIL POINTERS
	// ============================================
	fmt.Println("\n--- Nil Pointers ---")

	var nilPtr *int
	fmt.Printf("nilPtr == nil: %t\n", nilPtr == nil)

	// PELIGRO: dereferenciar nil causa panic
	// fmt.Println(*nilPtr) // panic: runtime error

	// Siempre verificar antes de dereferenciar
	if nilPtr != nil {
		fmt.Println(*nilPtr)
	} else {
		fmt.Println("nilPtr es nil, no se puede dereferenciar")
	}

	// ============================================
	// PASO POR VALOR VS REFERENCIA
	// ============================================
	fmt.Println("\n--- Paso por Valor vs Referencia ---")

	// Go SIEMPRE pasa por valor
	// Pero el valor puede ser un puntero (referencia)

	numero := 10
	fmt.Printf("Original: %d\n", numero)

	// Paso por valor - la función recibe una copia
	incrementarValor(numero)
	fmt.Printf("Después de incrementarValor: %d (sin cambios)\n", numero)

	// Paso por puntero - la función puede modificar
	incrementarPuntero(&numero)
	fmt.Printf("Después de incrementarPuntero: %d (modificado)\n", numero)

	// ============================================
	// PUNTEROS Y STRUCTS
	// ============================================
	fmt.Println("\n--- Punteros y Structs ---")

	// Struct por valor
	persona1 := Persona{Nombre: "Alice", Edad: 30}
	cumplirValor(persona1)
	fmt.Printf("Después de cumplirValor: %+v (sin cambios)\n", persona1)

	// Struct por puntero
	cumplirPuntero(&persona1)
	fmt.Printf("Después de cumplirPuntero: %+v (modificado)\n", persona1)

	// Acceso a campos con punteros - Go hace dereferencia automática
	pPersona := &persona1
	fmt.Printf("pPersona.Nombre: %s\n", pPersona.Nombre) // equivale a (*pPersona).Nombre

	// ============================================
	// PUNTEROS A PUNTEROS
	// ============================================
	fmt.Println("\n--- Punteros a Punteros ---")

	a := 5
	pa := &a   // puntero a int
	ppa := &pa // puntero a puntero a int (**)

	fmt.Printf("a = %d\n", a)
	fmt.Printf("*pa = %d\n", *pa)
	fmt.Printf("**ppa = %d\n", **ppa)

	**ppa = 999
	fmt.Printf("Después de **ppa = 999: a = %d\n", a)

	// ============================================
	// PUNTEROS Y SLICES
	// ============================================
	fmt.Println("\n--- Punteros y Slices ---")

	// Los slices ya contienen un puntero al array subyacente
	// Pasar un slice es eficiente (no copia los datos)
	slice := []int{1, 2, 3}
	modificarSlice(slice)
	fmt.Printf("Después de modificarSlice: %v\n", slice) // modificado

	// Pero append puede crear un nuevo array
	slice2 := []int{1, 2, 3}
	agregarElemento(slice2)
	fmt.Printf("Después de agregarElemento: %v\n", slice2) // NO modificado

	// Para modificar el slice mismo, usar puntero
	agregarElementoPuntero(&slice2)
	fmt.Printf("Después de agregarElementoPuntero: %v\n", slice2) // modificado

	// ============================================
	// PUNTEROS Y MAPS
	// ============================================
	fmt.Println("\n--- Punteros y Maps ---")

	// Los maps ya son referencias, no necesitan punteros
	mapa := map[string]int{"a": 1}
	modificarMapa(mapa)
	fmt.Printf("Después de modificarMapa: %v\n", mapa) // modificado

	// ============================================
	// PUNTERO A FUNCIÓN
	// ============================================
	fmt.Println("\n--- Puntero a Función ---")

	// Las funciones ya son referencias
	fn := saludar
	fn("Gopher")

	// Pero puedes tener puntero a variable de función
	var fnPtr *func(string)
	fnPtr = &fn
	(*fnPtr)("Mundo")

	// ============================================
	// PUNTEROS COMO RECEPTORES DE MÉTODOS
	// ============================================
	fmt.Println("\n--- Punteros como Receptores ---")

	// Receptor por valor - no modifica el original
	c1 := Contador{valor: 0}
	c1.IncrementarValor()
	fmt.Printf("Después de IncrementarValor: %d\n", c1.valor)

	// Receptor por puntero - modifica el original
	c1.IncrementarPuntero()
	fmt.Printf("Después de IncrementarPuntero: %d\n", c1.valor)

	// Go convierte automáticamente para llamar métodos
	// c1.IncrementarPuntero() es igual a (&c1).IncrementarPuntero()

	// ============================================
	// CUÁNDO USAR PUNTEROS
	// ============================================
	fmt.Println("\n--- Cuándo Usar Punteros ---")
	fmt.Println(`
USAR PUNTEROS cuando:
1. Necesitas modificar el valor original
2. El struct es grande (evitar copias costosas)
3. Necesitas representar "ausencia de valor" (nil)
4. Métodos que modifican el receptor
5. Consistencia (si algunos métodos son punteros, todos deben serlo)

NO USAR PUNTEROS cuando:
1. Tipos pequeños (int, bool, etc.) - copiar es barato
2. Slices, maps, channels - ya son referencias
3. Funciones - ya son referencias
4. Cuando la inmutabilidad es deseada
5. Structs pequeños (< 64 bytes aproximadamente)`)
	// ============================================
	// ESCAPE ANALYSIS
	// ============================================
	fmt.Println("\n--- Escape Analysis ---")
	fmt.Println(`
Go decide automáticamente si una variable va en stack o heap:
- Variables locales que no "escapan" van en stack (más rápido)
- Variables que "escapan" (ej: retornadas por puntero) van en heap

Para ver escape analysis:
  go build -gcflags="-m" main.go`)
	// Esta variable "escapa" al heap porque se retorna un puntero
	ptr := crearNumero(42)
	fmt.Printf("Número creado: %d\n", *ptr)

	// ============================================
	// EJEMPLO PRÁCTICO: OPCIONAL
	// ============================================
	fmt.Println("\n--- Ejemplo: Valores Opcionales ---")

	// Config sin valores opcionales
	config1 := Config{Puerto: 8080}
	imprimirConfig(config1)

	// Config con valores opcionales
	timeout := 30
	debug := true
	config2 := Config{
		Puerto:  3000,
		Timeout: &timeout,
		Debug:   &debug,
	}
	imprimirConfig(config2)
}

// Funciones auxiliares

func incrementarValor(n int) {
	n++ // modifica la copia local
}

func incrementarPuntero(n *int) {
	*n++ // modifica el valor original
}

type Persona struct {
	Nombre string
	Edad   int
}

func cumplirValor(p Persona) {
	p.Edad++
}

func cumplirPuntero(p *Persona) {
	p.Edad++
}

func modificarSlice(s []int) {
	s[0] = 999 // modifica el array subyacente
}

func agregarElemento(s []int) {
	s = append(s, 4) // crea potencialmente nuevo array
	_ = s            // el slice original no cambia
}

func agregarElementoPuntero(s *[]int) {
	*s = append(*s, 4) // modifica el slice original
}

func modificarMapa(m map[string]int) {
	m["b"] = 2
}

func saludar(nombre string) {
	fmt.Printf("¡Hola, %s!\n", nombre)
}

type Contador struct {
	valor int
}

func (c Contador) IncrementarValor() {
	c.valor++
}

func (c *Contador) IncrementarPuntero() {
	c.valor++
}

func crearNumero(n int) *int {
	resultado := n
	return &resultado // escapa al heap
}

type Config struct {
	Puerto  int
	Timeout *int
	Debug   *bool
}

func imprimirConfig(c Config) {
	fmt.Printf("Puerto: %d\n", c.Puerto)
	if c.Timeout != nil {
		fmt.Printf("Timeout: %d\n", *c.Timeout)
	} else {
		fmt.Println("Timeout: (default)")
	}
	if c.Debug != nil {
		fmt.Printf("Debug: %t\n", *c.Debug)
	} else {
		fmt.Println("Debug: (default)")
	}
}

/*
RESUMEN:

OPERADORES:
& - obtener dirección (address of)
* - dereferenciar (value at)

ZERO VALUE:
*T es nil (no apunta a nada)

TIPOS QUE YA SON REFERENCIAS:
- slices (header con puntero al array)
- maps
- channels
- functions
- interfaces (internamente)

BUENAS PRÁCTICAS:

1. Siempre verificar nil antes de dereferenciar
2. No retornar punteros a variables locales... espera, ¡en Go sí puedes!
   (el escape analysis lo maneja)
3. Usa punteros para structs grandes
4. Usa punteros como receptores si el método modifica el struct
5. Sé consistente: si un método es puntero, todos deberían serlo
6. Para valores opcionales, considera usar punteros o tipos como sql.NullInt64
*/

/*
SUMMARY - POINTERS:

CONCEPTOS BÁSICOS:
- & obtiene la dirección de memoria de una variable
- * desreferencia un puntero (accede al valor)
- Zero value de *T es nil

DECLARACIÓN:
- var p *int (nil pointer)
- p := new(int) (aloca memoria, retorna puntero)
- p := &variable (dirección de variable existente)

PASO POR VALOR VS REFERENCIA:
- Go SIEMPRE pasa por valor
- Pasar puntero permite modificar el original
- Slices, maps, channels y funciones ya son referencias

PUNTEROS Y STRUCTS:
- Go hace dereferencia automática: ptr.Campo = (*ptr).Campo
- Receptores puntero modifican el struct original
- Receptores valor trabajan con una copia

CUÁNDO USAR PUNTEROS:
- Para modificar el valor original
- Structs grandes (evitar copias costosas)
- Representar ausencia de valor (nil)
- Consistencia en receptores de métodos

ESCAPE ANALYSIS:
- Go decide automáticamente stack vs heap
- Variables que "escapan" (retornadas por puntero) van al heap
- go build -gcflags="-m" para ver escape analysis

VALORES OPCIONALES:
- Usar *int, *bool para campos opcionales en structs
- nil indica "no establecido", diferente del zero value
*/
