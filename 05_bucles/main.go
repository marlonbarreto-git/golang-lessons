// Package main - Capítulo 05: Bucles
// En Go solo existe 'for', pero es muy versátil.
// Aprenderás todas sus formas: tradicional, while, infinito, range.
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== BUCLES EN GO ===")

	// ============================================
	// FOR TRADICIONAL (estilo C)
	// ============================================
	fmt.Println("\n--- For Tradicional ---")

	// for inicialización; condición; post { }
	for i := 0; i < 5; i++ {
		fmt.Printf("i = %d\n", i)
	}

	// Múltiples variables
	fmt.Println("\nMúltiples variables:")
	for i, j := 0, 10; i < j; i, j = i+1, j-1 {
		fmt.Printf("i=%d, j=%d\n", i, j)
	}

	// Nota: no hay ++i ni i++ como expresión, solo statement
	// for i := 0; i < 5; i += 2 { } // incremento de 2

	// ============================================
	// FOR COMO WHILE
	// ============================================
	fmt.Println("\n--- For como While ---")

	// Omitiendo inicialización y post
	contador := 0
	for contador < 5 {
		fmt.Printf("contador = %d\n", contador)
		contador++
	}

	// ============================================
	// FOR INFINITO
	// ============================================
	fmt.Println("\n--- For Infinito (con break) ---")

	iteraciones := 0
	for {
		iteraciones++
		fmt.Printf("Iteración %d\n", iteraciones)
		if iteraciones >= 3 {
			break // sale del bucle
		}
	}

	// ============================================
	// BREAK Y CONTINUE
	// ============================================
	fmt.Println("\n--- Break y Continue ---")

	fmt.Println("Imprimiendo solo impares del 0 al 9:")
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			continue // salta a la siguiente iteración
		}
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	fmt.Println("\nParando al encontrar 5:")
	for i := 0; i < 10; i++ {
		if i == 5 {
			break // sale completamente del bucle
		}
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// ============================================
	// LABELS (ETIQUETAS)
	// ============================================
	fmt.Println("\n--- Labels para bucles anidados ---")

	// Las etiquetas permiten break/continue en bucles externos
Outer:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if i == 1 && j == 1 {
				fmt.Println("Saltando al bucle externo")
				continue Outer // continúa en el bucle externo
			}
			fmt.Printf("i=%d, j=%d\n", i, j)
		}
	}

	fmt.Println("\nSaliendo de bucles anidados:")
Search:
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i*j == 6 {
				fmt.Printf("Encontrado: i=%d, j=%d\n", i, j)
				break Search // sale de ambos bucles
			}
		}
	}

	// ============================================
	// FOR RANGE - SLICES Y ARRAYS
	// ============================================
	fmt.Println("\n--- For Range: Slices ---")

	numeros := []int{10, 20, 30, 40, 50}

	// Índice y valor
	fmt.Println("Índice y valor:")
	for idx, val := range numeros {
		fmt.Printf("índice=%d, valor=%d\n", idx, val)
	}

	// Solo índice
	fmt.Println("\nSolo índice:")
	for idx := range numeros {
		fmt.Printf("índice=%d\n", idx)
	}

	// Solo valor (ignorar índice con _)
	fmt.Println("\nSolo valor:")
	for _, val := range numeros {
		fmt.Printf("valor=%d\n", val)
	}

	// Ignorar ambos (solo para efectos secundarios)
	count := 0
	for range numeros {
		count++
	}
	fmt.Printf("Total elementos: %d\n", count)

	// ============================================
	// FOR RANGE - MAPS
	// ============================================
	fmt.Println("\n--- For Range: Maps ---")

	paises := map[string]string{
		"MX": "México",
		"AR": "Argentina",
		"CO": "Colombia",
		"ES": "España",
	}

	// Clave y valor
	for codigo, nombre := range paises {
		fmt.Printf("%s: %s\n", codigo, nombre)
	}

	// NOTA: El orden de iteración en maps NO está garantizado

	// Solo claves
	fmt.Println("\nSolo claves:")
	for codigo := range paises {
		fmt.Printf("%s\n", codigo)
	}

	// ============================================
	// FOR RANGE - STRINGS
	// ============================================
	fmt.Println("\n--- For Range: Strings ---")

	texto := "Hola 世界"

	// Range sobre string itera sobre runes (caracteres Unicode)
	fmt.Println("Iterando sobre runes:")
	for idx, runeVal := range texto {
		fmt.Printf("índice=%d, rune=%c (valor: %d)\n", idx, runeVal, runeVal)
	}

	// Nota: el índice es el byte offset, no el índice del caracter
	// Los caracteres multibyte (como 世) ocupan múltiples bytes

	// Si necesitas iterar sobre bytes:
	fmt.Println("\nIterando sobre bytes:")
	for i := 0; i < len(texto); i++ {
		fmt.Printf("byte[%d] = %d (%c)\n", i, texto[i], texto[i])
	}

	// ============================================
	// FOR RANGE - CHANNELS
	// ============================================
	fmt.Println("\n--- For Range: Channels ---")

	ch := make(chan int)

	// Goroutine que envía datos
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i
			time.Sleep(100 * time.Millisecond)
		}
		close(ch) // Importante: cerrar el channel
	}()

	// Range sobre channel itera hasta que se cierre
	for valor := range ch {
		fmt.Printf("Recibido: %d\n", valor)
	}

	// ============================================
	// FOR RANGE - INTEGERS (Go 1.22+)
	// ============================================
	fmt.Println("\n--- For Range: Integers (Go 1.22+) ---")

	// Nueva sintaxis en Go 1.22
	for i := range 5 {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// Equivalente a: for i := 0; i < 5; i++

	// ============================================
	// FOR RANGE - FUNCIONES (Go 1.23+)
	// ============================================
	fmt.Println("\n--- For Range: Funciones Iteradoras (Go 1.23+) ---")

	// Go 1.23 introdujo range over functions
	// La función debe aceptar un yield callback

	for v := range countdown(5) {
		fmt.Printf("%d ", v)
	}
	fmt.Println("¡Despegue!")

	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrones Comunes ---")

	// 1. Iterar hacia atrás
	fmt.Println("Iteración inversa:")
	arr := []string{"a", "b", "c", "d"}
	for i := len(arr) - 1; i >= 0; i-- {
		fmt.Printf("%s ", arr[i])
	}
	fmt.Println()

	// 2. Iterar con paso
	fmt.Println("\nCada 2 elementos:")
	for i := 0; i < len(arr); i += 2 {
		fmt.Printf("%s ", arr[i])
	}
	fmt.Println()

	// 3. Iterar pares de elementos
	fmt.Println("\nPares (actual, siguiente):")
	for i := 0; i < len(arr)-1; i++ {
		fmt.Printf("(%s, %s) ", arr[i], arr[i+1])
	}
	fmt.Println()

	// 4. Enumerar con índice personalizado
	fmt.Println("\nEnumeración desde 1:")
	for idx, val := range arr {
		fmt.Printf("%d. %s\n", idx+1, val)
	}

	// ============================================
	// MODIFICAR DURANTE ITERACIÓN
	// ============================================
	fmt.Println("\n--- Cuidado al Modificar ---")

	// SAFE: Modificar valores de slice existentes
	nums := []int{1, 2, 3, 4, 5}
	for i := range nums {
		nums[i] *= 2
	}
	fmt.Printf("Duplicados: %v\n", nums)

	// CUIDADO: El valor en range es una copia
	for _, v := range nums {
		v *= 2 // Esto NO modifica el slice original
		_ = v
	}
	fmt.Printf("Sin cambios: %v\n", nums)

	// UNSAFE: No agregar/eliminar elementos durante range
	// Esto puede causar comportamiento indefinido
}

// Función iteradora (Go 1.23+)
// yield retorna false si el llamador quiere parar
func countdown(start int) func(yield func(int) bool) {
	return func(yield func(int) bool) {
		for i := start; i > 0; i-- {
			if !yield(i) {
				return
			}
		}
	}
}

/*
RESUMEN DE FORMAS DE FOR:

1. Tradicional:     for i := 0; i < n; i++ { }
2. Como while:      for condición { }
3. Infinito:        for { }
4. Range slice:     for i, v := range slice { }
5. Range map:       for k, v := range mapa { }
6. Range string:    for i, r := range str { }  (runes)
7. Range channel:   for v := range ch { }
8. Range int:       for i := range n { }  (Go 1.22+)
9. Range func:      for v := range iterFunc() { }  (Go 1.23+)

BUENAS PRÁCTICAS:

1. Usa range siempre que sea posible (más idiomático)
2. Usa _ para ignorar valores que no necesitas
3. Ten cuidado con modificar colecciones durante iteración
4. Usa labels solo cuando sea realmente necesario
5. Prefiere funciones iteradoras para lógica de iteración compleja
6. Recuerda que range sobre string itera runes, no bytes
*/
