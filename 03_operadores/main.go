// Package main - Capítulo 03: Operadores
// Aprenderás todos los operadores disponibles en Go:
// aritméticos, de comparación, lógicos, bit a bit, y de asignación.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== OPERADORES EN GO ===")
	// ============================================
	// OPERADORES ARITMÉTICOS
	// ============================================
	fmt.Println("--- Operadores Aritméticos ---")

	a, b := 17, 5

	fmt.Printf("%d + %d = %d\n", a, b, a+b)   // Suma
	fmt.Printf("%d - %d = %d\n", a, b, a-b)   // Resta
	fmt.Printf("%d * %d = %d\n", a, b, a*b)   // Multiplicación
	fmt.Printf("%d / %d = %d\n", a, b, a/b)   // División entera
	fmt.Printf("%d %% %d = %d\n", a, b, a%b)  // Módulo (resto)

	// División con flotantes
	fa, fb := 17.0, 5.0
	fmt.Printf("%.1f / %.1f = %.2f\n", fa, fb, fa/fb)

	// Operadores unarios
	x := 5
	fmt.Printf("\nOperadores unarios:\n")
	fmt.Printf("+x = %d (positivo)\n", +x)
	fmt.Printf("-x = %d (negativo)\n", -x)

	// ============================================
	// INCREMENTO Y DECREMENTO
	// ============================================
	fmt.Println("\n--- Incremento y Decremento ---")

	counter := 10
	fmt.Printf("Valor inicial: %d\n", counter)

	counter++ // Incremento (NO es expresión, es statement)
	fmt.Printf("Después de ++: %d\n", counter)

	counter-- // Decremento
	fmt.Printf("Después de --: %d\n", counter)

	// NOTA: En Go, ++ y -- son statements, no expresiones
	// Esto NO compila:
	// y := x++        // ERROR
	// fmt.Println(x++) // ERROR
	// ++x             // ERROR (no existe pre-incremento)

	// ============================================
	// OPERADORES DE COMPARACIÓN
	// ============================================
	fmt.Println("\n--- Operadores de Comparación ---")

	p, q := 10, 20

	fmt.Printf("%d == %d: %t\n", p, q, p == q) // Igual
	fmt.Printf("%d != %d: %t\n", p, q, p != q) // Diferente
	fmt.Printf("%d < %d:  %t\n", p, q, p < q)  // Menor que
	fmt.Printf("%d > %d:  %t\n", p, q, p > q)  // Mayor que
	fmt.Printf("%d <= %d: %t\n", p, q, p <= q) // Menor o igual
	fmt.Printf("%d >= %d: %t\n", p, q, p >= q) // Mayor o igual

	// Comparación de strings (lexicográfica)
	s1, s2 := "apple", "banana"
	fmt.Printf("\n%q < %q: %t\n", s1, s2, s1 < s2)
	fmt.Printf("%q == %q: %t\n", "go", "go", "go" == "go")

	// ============================================
	// OPERADORES LÓGICOS
	// ============================================
	fmt.Println("\n--- Operadores Lógicos ---")

	verdadero, falso := true, false

	fmt.Printf("true && true:  %t\n", verdadero && verdadero)  // AND
	fmt.Printf("true && false: %t\n", verdadero && falso)
	fmt.Printf("true || false: %t\n", verdadero || falso)      // OR
	fmt.Printf("false || false: %t\n", falso || falso)
	fmt.Printf("!true:  %t\n", !verdadero)                     // NOT
	fmt.Printf("!false: %t\n", !falso)

	// Short-circuit evaluation (evaluación de cortocircuito)
	fmt.Println("\nShort-circuit evaluation:")
	// En AND, si el primero es false, no evalúa el segundo
	// En OR, si el primero es true, no evalúa el segundo

	result := falso && expensiveCheck() // expensiveCheck NO se ejecuta
	fmt.Printf("false && expensive(): %t (expensive no se llamó)\n", result)

	result = verdadero || expensiveCheck() // expensiveCheck NO se ejecuta
	fmt.Printf("true || expensive(): %t (expensive no se llamó)\n", result)

	// ============================================
	// OPERADORES BIT A BIT
	// ============================================
	fmt.Println("\n--- Operadores Bit a Bit ---")

	m, n := uint8(0b11001010), uint8(0b10101100) // 202 y 172 en decimal

	fmt.Printf("m = %08b (%d)\n", m, m)
	fmt.Printf("n = %08b (%d)\n", n, n)
	fmt.Printf("\n")
	fmt.Printf("m & n  = %08b (%d) AND\n", m&n, m&n)
	fmt.Printf("m | n  = %08b (%d) OR\n", m|n, m|n)
	fmt.Printf("m ^ n  = %08b (%d) XOR\n", m^n, m^n)
	fmt.Printf("^m     = %08b (%d) NOT (complemento)\n", ^m, ^m)
	fmt.Printf("m &^ n = %08b (%d) AND NOT (bit clear)\n", m&^n, m&^n)

	// Shift operators
	fmt.Println("\nShift operators:")
	v := uint8(0b00000001) // 1
	fmt.Printf("Original:  %08b (%d)\n", v, v)
	fmt.Printf("<< 1:      %08b (%d) - multiplicar por 2\n", v<<1, v<<1)
	fmt.Printf("<< 2:      %08b (%d) - multiplicar por 4\n", v<<2, v<<2)
	fmt.Printf("<< 4:      %08b (%d) - multiplicar por 16\n", v<<4, v<<4)

	w := uint8(0b10000000) // 128
	fmt.Printf("\nOriginal:  %08b (%d)\n", w, w)
	fmt.Printf(">> 1:      %08b (%d) - dividir por 2\n", w>>1, w>>1)
	fmt.Printf(">> 2:      %08b (%d) - dividir por 4\n", w>>2, w>>2)

	// ============================================
	// OPERADORES DE ASIGNACIÓN
	// ============================================
	fmt.Println("\n--- Operadores de Asignación ---")

	var num int

	num = 10
	fmt.Printf("num = 10:  %d\n", num)

	num += 5 // num = num + 5
	fmt.Printf("num += 5:  %d\n", num)

	num -= 3 // num = num - 3
	fmt.Printf("num -= 3:  %d\n", num)

	num *= 2 // num = num * 2
	fmt.Printf("num *= 2:  %d\n", num)

	num /= 4 // num = num / 4
	fmt.Printf("num /= 4:  %d\n", num)

	num %= 3 // num = num % 3
	fmt.Printf("num %%= 3:  %d\n", num)

	// Asignación con operadores bit a bit
	bits := uint8(0b11110000)
	fmt.Printf("\nbits inicial: %08b\n", bits)

	bits &= 0b10101010
	fmt.Printf("bits &= 0b10101010: %08b\n", bits)

	bits |= 0b00001111
	fmt.Printf("bits |= 0b00001111: %08b\n", bits)

	bits ^= 0b11111111
	fmt.Printf("bits ^= 0b11111111: %08b\n", bits)

	bits <<= 2
	fmt.Printf("bits <<= 2: %08b\n", bits)

	bits >>= 4
	fmt.Printf("bits >>= 4: %08b\n", bits)

	// ============================================
	// OPERADOR DE DIRECCIÓN Y DESREFERENCIA
	// ============================================
	fmt.Println("\n--- Operadores de Punteros ---")

	valor := 42
	puntero := &valor // & obtiene la dirección de memoria

	fmt.Printf("valor: %d\n", valor)
	fmt.Printf("&valor (dirección): %p\n", puntero)
	fmt.Printf("*puntero (desreferencia): %d\n", *puntero)

	*puntero = 100 // modifica el valor a través del puntero
	fmt.Printf("Después de *puntero = 100: valor = %d\n", valor)

	// ============================================
	// PRECEDENCIA DE OPERADORES
	// ============================================
	fmt.Println("\n--- Precedencia de Operadores ---")
	fmt.Println("De mayor a menor precedencia:")
	fmt.Println("1. * / % << >> & &^")
	fmt.Println("2. + - | ^")
	fmt.Println("3. == != < <= > >=")
	fmt.Println("4. &&")
	fmt.Println("5. ||")

	// Ejemplo de precedencia
	result1 := 2 + 3*4     // 3*4 primero = 14
	result2 := (2 + 3) * 4 // paréntesis primero = 20
	fmt.Printf("\n2 + 3 * 4 = %d\n", result1)
	fmt.Printf("(2 + 3) * 4 = %d\n", result2)

	// ============================================
	// CASOS ESPECIALES Y TRAMPAS
	// ============================================
	fmt.Println("\n--- Casos Especiales ---")

	// División por cero con enteros causa panic
	// fmt.Println(10 / 0) // panic: integer divide by zero

	// División por cero con flotantes da infinito
	var cero float64 = 0.0
	inf := 10.0 / cero
	fmt.Printf("10.0 / 0.0 = %v\n", inf)

	// NaN (Not a Number)
	nan := cero / cero
	fmt.Printf("0.0 / 0.0 = %v\n", nan)
	fmt.Printf("NaN == NaN: %t (siempre false)\n", nan == nan)

	// Overflow (silencioso, no hay panic)
	var overflow uint8 = 255
	overflow++
	fmt.Printf("uint8(255) + 1 = %d (overflow silencioso)\n", overflow)

	// Underflow
	var underflow uint8 = 0
	underflow--
	fmt.Printf("uint8(0) - 1 = %d (underflow silencioso)\n", underflow)
}

// Función costosa que no queremos ejecutar innecesariamente
func expensiveCheck() bool {
	fmt.Println("  [expensiveCheck fue llamada]")
	return true
}

/*
BUENAS PRÁCTICAS:

1. Usa paréntesis cuando la precedencia no sea obvia
2. Aprovecha short-circuit evaluation para evitar cálculos costosos
3. Ten cuidado con overflow/underflow - Go no lo detecta automáticamente
4. Prefiere operaciones de bits para flags y permisos
5. Usa constantes para valores de bits que representan opciones

OPERADORES QUE NO EXISTEN EN GO:
- ?: (operador ternario) - usa if/else
- ++x, --x (pre-incremento/decremento)
- x++ como expresión (solo statement)
*/
