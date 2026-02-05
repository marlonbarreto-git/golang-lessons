// Package main - Capítulo 06: Switch
// Switch en Go es más poderoso que en otros lenguajes.
// No requiere break, soporta type switch, y más.
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	fmt.Println("=== SWITCH EN GO ===")

	// ============================================
	// SWITCH BÁSICO
	// ============================================
	fmt.Println("\n--- Switch Básico ---")

	dia := 3
	switch dia {
	case 1:
		fmt.Println("Lunes")
	case 2:
		fmt.Println("Martes")
	case 3:
		fmt.Println("Miércoles")
	case 4:
		fmt.Println("Jueves")
	case 5:
		fmt.Println("Viernes")
	case 6:
		fmt.Println("Sábado")
	case 7:
		fmt.Println("Domingo")
	default:
		fmt.Println("Día inválido")
	}

	// NO hay break implícito - Go lo hace automáticamente
	// Esto es diferente a C/Java donde necesitas break

	// ============================================
	// MÚLTIPLES VALORES POR CASE
	// ============================================
	fmt.Println("\n--- Múltiples Valores por Case ---")

	diaSemana := "sábado"
	switch diaSemana {
	case "lunes", "martes", "miércoles", "jueves", "viernes":
		fmt.Println("Día laboral")
	case "sábado", "domingo":
		fmt.Println("Fin de semana")
	default:
		fmt.Println("Día desconocido")
	}

	// ============================================
	// SWITCH CON INICIALIZACIÓN
	// ============================================
	fmt.Println("\n--- Switch con Inicialización ---")

	// Similar a if, puedes inicializar una variable
	switch hora := time.Now().Hour(); {
	case hora < 12:
		fmt.Printf("Buenos días (son las %d)\n", hora)
	case hora < 18:
		fmt.Printf("Buenas tardes (son las %d)\n", hora)
	default:
		fmt.Printf("Buenas noches (son las %d)\n", hora)
	}

	// ============================================
	// SWITCH SIN EXPRESIÓN (como if-else)
	// ============================================
	fmt.Println("\n--- Switch sin Expresión ---")

	nota := 85

	// Esto es equivalente a switch true { }
	switch {
	case nota >= 90:
		fmt.Println("Excelente")
	case nota >= 80:
		fmt.Println("Muy bien")
	case nota >= 70:
		fmt.Println("Bien")
	case nota >= 60:
		fmt.Println("Suficiente")
	default:
		fmt.Println("Insuficiente")
	}

	// ============================================
	// FALLTHROUGH
	// ============================================
	fmt.Println("\n--- Fallthrough ---")

	// fallthrough continúa al siguiente case (sin evaluar condición)
	numero := 5
	switch numero {
	case 5:
		fmt.Println("Es 5")
		fallthrough
	case 4:
		fmt.Println("También ejecuta este (es 4 o vino de fallthrough)")
		fallthrough
	case 3:
		fmt.Println("Y este también")
		// No hay fallthrough aquí, se detiene
	case 2:
		fmt.Println("Este NO se ejecuta")
	}

	// NOTA: fallthrough es raro de usar, generalmente es mejor
	// usar múltiples valores en el case

	// ============================================
	// TYPE SWITCH
	// ============================================
	fmt.Println("\n--- Type Switch ---")

	// Type switch permite verificar el tipo de una interface{}
	valores := []any{42, "hola", 3.14, true, []int{1, 2}, nil}

	for _, v := range valores {
		describirTipo(v)
	}

	// ============================================
	// TYPE SWITCH CON VARIABLE
	// ============================================
	fmt.Println("\n--- Type Switch con Variable ---")

	var x any = "texto"

	switch v := x.(type) {
	case int:
		fmt.Printf("int: %d * 2 = %d\n", v, v*2)
	case string:
		fmt.Printf("string: len(%q) = %d\n", v, len(v))
	case bool:
		fmt.Printf("bool: !%t = %t\n", v, !v)
	default:
		fmt.Printf("tipo desconocido: %T\n", v)
	}

	// ============================================
	// SWITCH CON STRUCTS
	// ============================================
	fmt.Println("\n--- Switch con Structs ---")

	type Punto struct {
		X, Y int
	}

	punto := Punto{0, 0}
	switch punto {
	case Punto{0, 0}:
		fmt.Println("Origen")
	case Punto{1, 1}:
		fmt.Println("Uno-Uno")
	default:
		fmt.Printf("Punto: %+v\n", punto)
	}

	// ============================================
	// SWITCH CON OS/RUNTIME
	// ============================================
	fmt.Println("\n--- Switch con Runtime ---")

	switch os := runtime.GOOS; os {
	case "darwin":
		fmt.Println("macOS")
	case "linux":
		fmt.Println("Linux")
	case "windows":
		fmt.Println("Windows")
	default:
		fmt.Printf("Otro OS: %s\n", os)
	}

	// ============================================
	// CASOS PRÁCTICOS
	// ============================================
	fmt.Println("\n--- Casos Prácticos ---")

	// 1. Mapeo de códigos HTTP
	httpCode := 404
	fmt.Printf("HTTP %d: %s\n", httpCode, httpStatus(httpCode))

	// 2. Categorizar números
	for _, n := range []int{-5, 0, 7, 15, 100} {
		fmt.Printf("%d es %s\n", n, categorizar(n))
	}

	// 3. Parser de comandos
	comandos := []string{"start", "stop", "restart", "status", "unknown"}
	for _, cmd := range comandos {
		ejecutarComando(cmd)
	}

	// ============================================
	// BREAK EN SWITCH
	// ============================================
	fmt.Println("\n--- Break en Switch ---")

	// break en switch sale del switch (no de un bucle externo)
	valor := 2
	switch valor {
	case 1:
		fmt.Println("Uno")
	case 2:
		if valor%2 == 0 {
			fmt.Println("Es par, saliendo temprano")
			break // sale del switch
		}
		fmt.Println("Esto no se ejecuta")
	case 3:
		fmt.Println("Tres")
	}

	// Para salir de un bucle desde un switch, usa labels
Loop:
	for i := 0; i < 5; i++ {
		switch i {
		case 3:
			fmt.Println("Encontrado 3, saliendo del bucle")
			break Loop // sale del bucle, no solo del switch
		default:
			fmt.Printf("Iteración %d\n", i)
		}
	}

	// ============================================
	// SWITCH VS IF-ELSE
	// ============================================
	fmt.Println("\n--- Cuándo Usar Switch vs If-Else ---")

	// Usa switch cuando:
	// 1. Comparas una variable contra múltiples valores constantes
	// 2. Necesitas type switch
	// 3. Quieres código más limpio que múltiples if-else

	// Usa if-else cuando:
	// 1. Las condiciones son complejas y variadas
	// 2. Solo tienes 2-3 opciones simples
	// 3. Necesitas evaluar diferentes variables en cada condición
}

func describirTipo(v any) {
	switch v.(type) {
	case int:
		fmt.Printf("%v es int\n", v)
	case string:
		fmt.Printf("%v es string\n", v)
	case float64:
		fmt.Printf("%v es float64\n", v)
	case bool:
		fmt.Printf("%v es bool\n", v)
	case []int:
		fmt.Printf("%v es []int\n", v)
	case nil:
		fmt.Println("valor es nil")
	default:
		fmt.Printf("%v es tipo desconocido (%T)\n", v, v)
	}
}

func httpStatus(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
	}
}

func categorizar(n int) string {
	switch {
	case n < 0:
		return "negativo"
	case n == 0:
		return "cero"
	case n < 10:
		return "pequeño"
	case n < 100:
		return "mediano"
	default:
		return "grande"
	}
}

func ejecutarComando(cmd string) {
	switch cmd {
	case "start":
		fmt.Println("Iniciando servicio...")
	case "stop":
		fmt.Println("Deteniendo servicio...")
	case "restart":
		fmt.Println("Reiniciando servicio...")
	case "status":
		fmt.Println("Estado: running")
	default:
		fmt.Printf("Comando desconocido: %s\n", cmd)
	}
}

/*
DIFERENCIAS CON OTROS LENGUAJES:

1. No necesitas break - Go lo hace implícito
2. fallthrough existe pero es explícito
3. Los cases no necesitan ser constantes
4. Puedes tener switch sin expresión (como if-else)
5. Type switch para verificar tipos de interface{}
6. Switch con inicialización (como if)

BUENAS PRÁCTICAS:

1. Prefiere switch sobre múltiples if-else cuando comparas
   una variable contra valores constantes
2. Usa switch sin expresión para rangos de valores
3. Evita fallthrough a menos que sea realmente necesario
4. Type switch es la forma idiomática de manejar interface{}
5. Siempre incluye default para manejar casos inesperados
*/
