// Package main - Capítulo 01: Hola Mundo
// Este es tu primer programa en Go. Aprenderás la estructura básica
// de un programa Go y cómo ejecutarlo.
package main

import (
	"fmt"
	"os"
	"runtime"
)

// main es el punto de entrada de todo programa Go ejecutable.
// El compilador busca esta función en el paquete "main" para iniciar la ejecución.
func main() {
	// fmt.Println imprime en stdout con un newline al final
	fmt.Println("¡Hola, Mundo!")
	fmt.Println("Hello, World!")
	fmt.Println("你好，世界！")

	// Información del sistema
	fmt.Printf("\n=== Información del Sistema ===\n")
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Arch: %s\n", runtime.GOARCH)
	fmt.Printf("CPUs: %d\n", runtime.NumCPU())

	// Argumentos de línea de comandos
	fmt.Printf("\n=== Argumentos ===\n")
	fmt.Printf("Programa: %s\n", os.Args[0])
	if len(os.Args) > 1 {
		fmt.Printf("Argumentos: %v\n", os.Args[1:])
	}

	// Diferencias entre Print, Printf, Println
	fmt.Printf("\n=== Funciones de Impresión ===\n")

	// Print: sin newline, sin formato
	fmt.Print("Print: sin newline")
	fmt.Print(" - continúa aquí\n")

	// Printf: con formato (verbos)
	nombre := "Gopher"
	edad := 15
	fmt.Printf("Printf: Hola %s, tienes %d años\n", nombre, edad)

	// Println: con newline automático, espacios entre argumentos
	fmt.Println("Println:", nombre, "tiene", edad, "años")

	// Verbos de formato comunes
	fmt.Printf("\n=== Verbos de Formato ===\n")
	fmt.Printf("%%v  - valor: %v\n", 42)
	fmt.Printf("%%+v - valor con campos: %+v\n", struct{ X int }{42})
	fmt.Printf("%%#v - sintaxis Go: %#v\n", "hello")
	fmt.Printf("%%T  - tipo: %T\n", 42)
	fmt.Printf("%%d  - entero decimal: %d\n", 42)
	fmt.Printf("%%b  - binario: %b\n", 42)
	fmt.Printf("%%o  - octal: %o\n", 42)
	fmt.Printf("%%x  - hexadecimal: %x\n", 42)
	fmt.Printf("%%f  - float: %f\n", 3.14159)
	fmt.Printf("%%.2f - float con precisión: %.2f\n", 3.14159)
	fmt.Printf("%%e  - notación científica: %e\n", 1234567.89)
	fmt.Printf("%%s  - string: %s\n", "texto")
	fmt.Printf("%%q  - string quoted: %q\n", "texto")
	fmt.Printf("%%p  - puntero: %p\n", &nombre)
	fmt.Printf("%%t  - boolean: %t\n", true)

	// Ancho y padding
	fmt.Printf("\n=== Ancho y Padding ===\n")
	fmt.Printf("|%10s|\n", "derecha")    // alineado a la derecha
	fmt.Printf("|%-10s|\n", "izquierda") // alineado a la izquierda
	fmt.Printf("|%010d|\n", 42)          // padding con ceros
}

/*
COMANDOS BÁSICOS DE GO:

1. go run main.go
   - Compila y ejecuta en un solo paso
   - Ideal para desarrollo rápido
   - No genera archivo ejecutable permanente

2. go build
   - Compila y genera un ejecutable
   - go build -o miapp main.go  (especificar nombre)
   - El ejecutable es nativo y no requiere Go instalado

3. go fmt main.go
   - Formatea el código según el estándar de Go
   - gofmt -w .  (formatear todos los archivos)

4. go vet main.go
   - Analiza el código en busca de errores comunes
   - Detecta problemas que el compilador no detecta

5. go mod init nombre-modulo
   - Inicializa un módulo Go (para dependencias)

ESTRUCTURA DE UN PROGRAMA GO:

1. package main
   - Todo archivo Go pertenece a un paquete
   - "main" es especial: indica un programa ejecutable

2. import
   - Importa otros paquetes
   - "fmt" es el paquete de formato (input/output)

3. func main()
   - Punto de entrada del programa
   - No recibe argumentos (usar os.Args)
   - No retorna valor (usar os.Exit(code))
*/
