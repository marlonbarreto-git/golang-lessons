// Package main - Chapter 018: Type System
// Aprenderás type alias, type definitions, type assertions,
// y cómo crear tipos personalizados poderosos.
package main

import (
	"fmt"
	"strconv"
)

func main() {
	fmt.Println("=== SISTEMA DE TIPOS EN GO ===")

	// ============================================
	// TYPE DEFINITIONS (Wrapper Types)
	// ============================================
	fmt.Println("\n--- Type Definitions ---")

	// Son tipos DISTINTOS aunque basados en el mismo tipo subyacente
	var tempC Celsius = 100
	var tempF Fahrenheit = 212

	// Esto NO compila - tipos diferentes
	// tempC = tempF // error: cannot use tempF (type Fahrenheit) as type Celsius

	// Necesitas conversión explícita
	tempC = Celsius(tempF)
	fmt.Printf("tempC: %.1f, tempF: %.1f\n", tempC, tempF)

	// Puedes agregar métodos a type definitions
	fmt.Printf("100°C en Fahrenheit: %.1f°F\n", Celsius(100).ToFahrenheit())
	fmt.Printf("212°F en Celsius: %.1f°C\n", Fahrenheit(212).ToCelsius())

	// ============================================
	// TYPE ALIAS
	// ============================================
	fmt.Println("\n--- Type Alias ---")

	// Type alias: dos nombres para el MISMO tipo
	type ID = uint64    // ID es exactamente igual a uint64
	type Texto = string // Texto es exactamente igual a string

	var id ID = 123
	var numId uint64 = id // NO necesita conversión
	fmt.Printf("id: %d, numId: %d\n", id, numId)

	// NO puedes agregar métodos a type alias
	// (porque es el mismo tipo que el original)

	// byte y rune son type aliases
	var byteVar byte = 'A' // byte es alias de uint8
	var runeVar rune = '世' // rune es alias de int32
	fmt.Printf("byte: %d, rune: %d\n", byteVar, runeVar)

	// ============================================
	// DIFERENCIA: DEFINITION VS ALIAS
	// ============================================
	fmt.Println("\n--- Definition vs Alias ---")
	fmt.Println(`
TYPE DEFINITION (type X T):
- Crea un NUEVO tipo
- Necesita conversión explícita
- Puede tener métodos propios
- Útil para type safety y semántica

TYPE ALIAS (type X = T):
- Dos nombres para el MISMO tipo
- NO necesita conversión
- NO puede tener métodos propios
- Útil para refactoring gradual`)
	// ============================================
	// TIPOS PERSONALIZADOS ÚTILES
	// ============================================
	fmt.Println("\n--- Tipos Personalizados Útiles ---")

	// 1. Money type para evitar errores con centavos
	type Money int64 // centavos

	precio := Money(1999) // $19.99
	cantidad := 3
	total := precio * Money(cantidad)
	fmt.Printf("Total: $%.2f\n", float64(total)/100)

	// 2. Email type con validación
	email, err := NewEmail("usuario@ejemplo.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Email válido: %s\n", email)
	}

	// 3. Status type para estados
	type Status string
	const (
		StatusPending  Status = "pending"
		StatusActive   Status = "active"
		StatusInactive Status = "inactive"
	)

	var estado Status = StatusActive
	fmt.Printf("Estado: %s\n", estado)

	// Usar Direccion con iota
	fmt.Printf("Dirección Norte: %s\n", Norte)

	// ============================================
	// TYPE ASSERTION
	// ============================================
	fmt.Println("\n--- Type Assertion ---")

	var valor any = "hola mundo"

	// Assertion directa (puede causar panic)
	str := valor.(string)
	fmt.Printf("Assertion directa: %s\n", str)

	// Assertion segura con ok
	if num, ok := valor.(int); ok {
		fmt.Printf("Es int: %d\n", num)
	} else {
		fmt.Println("No es int")
	}

	// Múltiples assertions
	checkType(42)
	checkType("texto")
	checkType(3.14)
	checkType([]int{1, 2, 3})

	// ============================================
	// TYPE SWITCH
	// ============================================
	fmt.Println("\n--- Type Switch ---")

	valores := []any{42, "hola", 3.14, true, nil, []int{1, 2}}

	for _, v := range valores {
		processValue(v)
	}

	// ============================================
	// UNDERLYING TYPE
	// ============================================
	fmt.Println("\n--- Underlying Type ---")

	// El tipo subyacente es importante para conversiones
	type MiInt int
	type OtroInt int

	var a MiInt = 10
	var b OtroInt = 20

	// No puedes asignar directamente
	// a = b // error

	// Pero puedes convertir al tipo subyacente común
	resultado := int(a) + int(b)
	fmt.Printf("int(MiInt) + int(OtroInt) = %d\n", resultado)

	// O convertir entre ellos
	a = MiInt(b)
	fmt.Printf("MiInt(OtroInt): %d\n", a)

	// ============================================
	// TIPOS COMPUESTOS PERSONALIZADOS
	// ============================================
	fmt.Println("\n--- Tipos Compuestos ---")

	// Slice personalizado
	nums := IntSlice{5, 2, 8, 1, 9}
	fmt.Printf("Original: %v\n", nums)
	fmt.Printf("Sum: %d\n", nums.Sum())
	fmt.Printf("Max: %d\n", nums.Max())

	// Map personalizado
	headers := StringMap{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	fmt.Printf("Headers: %v\n", headers)
	fmt.Printf("Content-Type: %s\n", headers.Get("Content-Type"))
	fmt.Printf("Missing: %s\n", headers.Get("Missing"))

	// ============================================
	// FUNCIÓN TYPE
	// ============================================
	fmt.Println("\n--- Function Types ---")

	// Definir un tipo función
	type Handler func(string) string
	type Middleware func(Handler) Handler

	// Handler base
	hello := func(name string) string {
		return "Hello, " + name
	}

	// Middleware que agrega logging
	withLogging := func(h Handler) Handler {
		return func(s string) string {
			fmt.Printf("[LOG] Llamando con: %s\n", s)
			result := h(s)
			fmt.Printf("[LOG] Resultado: %s\n", result)
			return result
		}
	}

	// Componer
	handler := withLogging(hello)
	handler("Gopher")

	// ============================================
	// STRUCT TYPES ANÓNIMOS
	// ============================================
	fmt.Println("\n--- Struct Types ---")

	// Type definition de struct
	type Persona struct {
		Nombre string
		Edad   int
	}

	// Struct anónimo (diferente tipo aunque mismo layout)
	anonimo := struct {
		Nombre string
		Edad   int
	}{"Alice", 30}

	persona := Persona{"Alice", 30}

	// Son tipos diferentes pero convertibles si tienen misma estructura
	persona = Persona(anonimo)
	fmt.Printf("Convertido: %+v\n", persona)

	// ============================================
	// GENERIC TYPE ALIAS (Go 1.24+)
	// ============================================
	fmt.Println("\n--- Generic Type Alias (Go 1.24+) ---")

	// Go 1.24 permite type alias con parámetros de tipo
	type Pair[T any] = struct {
		First, Second T
	}

	intPair := Pair[int]{First: 1, Second: 2}
	strPair := Pair[string]{First: "hello", Second: "world"}

	fmt.Printf("Int pair: %+v\n", intPair)
	fmt.Printf("String pair: %+v\n", strPair)

	// ============================================
	// PATTERNS COMUNES
	// ============================================
	fmt.Println("\n--- Patterns Comunes ---")

	// 1. Newtype pattern para type safety
	type UserID int64
	type OrderID int64

	var userId UserID = 123
	var orderId OrderID = 456

	// Esto NO compila - previene mezclar IDs
	// processOrder(userId) // error si espera OrderID

	_ = userId
	_ = orderId

	// 2. Const con tipo para enums
	fmt.Printf("Norte: %d, Sur: %d\n", Norte, Sur)
}

// ============================================
// IMPLEMENTACIONES
// ============================================

// Métodos para Celsius/Fahrenheit
type Celsius float64
type Fahrenheit float64

func (c Celsius) ToFahrenheit() Fahrenheit {
	return Fahrenheit(c*9/5 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
	return Celsius((f - 32) * 5 / 9)
}

// Email type con validación
type Email string

func NewEmail(s string) (Email, error) {
	// Validación simple
	if len(s) < 3 || !contains(s, "@") {
		return "", fmt.Errorf("email inválido: %s", s)
	}
	return Email(s), nil
}

func (e Email) Domain() string {
	for i := len(e) - 1; i >= 0; i-- {
		if e[i] == '@' {
			return string(e[i+1:])
		}
	}
	return ""
}

func contains(s string, char string) bool {
	for _, c := range s {
		if string(c) == char {
			return true
		}
	}
	return false
}

// Funciones de type assertion/switch
func checkType(v any) {
	switch v.(type) {
	case int:
		fmt.Printf("%v es int\n", v)
	case string:
		fmt.Printf("%v es string\n", v)
	case float64:
		fmt.Printf("%v es float64\n", v)
	default:
		fmt.Printf("%v es %T\n", v, v)
	}
}

func processValue(v any) {
	switch val := v.(type) {
	case int:
		fmt.Printf("int: %d * 2 = %d\n", val, val*2)
	case string:
		fmt.Printf("string: %q (len=%d)\n", val, len(val))
	case float64:
		fmt.Printf("float64: %.2f\n", val)
	case bool:
		fmt.Printf("bool: %t\n", val)
	case nil:
		fmt.Println("nil")
	default:
		fmt.Printf("otro: %v (%T)\n", val, val)
	}
}

// Métodos para tipos slice y map
type IntSlice []int

func (s IntSlice) Sum() int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func (s IntSlice) Max() int {
	if len(s) == 0 {
		return 0
	}
	max := s[0]
	for _, v := range s[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

type StringMap map[string]string

func (m StringMap) Get(key string) string {
	if val, ok := m[key]; ok {
		return val
	}
	return "(not found)"
}

// Stringer para Direccion
type Direccion int

const (
	Norte Direccion = iota
	Sur
	Este
	Oeste
)

func (d Direccion) String() string {
	return [...]string{"Norte", "Sur", "Este", "Oeste"}[d]
}

// Conversión segura
func SafeInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i, true
		}
	}
	return 0, false
}

/*
RESUMEN:

TYPE DEFINITION (type X T):
- Nuevo tipo con tipo subyacente T
- Requiere conversión explícita
- Puede tener métodos
- Útil para: type safety, semántica, métodos

TYPE ALIAS (type X = T):
- Mismo tipo, diferente nombre
- NO requiere conversión
- NO puede tener métodos
- Útil para: refactoring, compatibility

TYPE ASSERTION:
v := i.(T)       // panic si falla
v, ok := i.(T)   // seguro

TYPE SWITCH:
switch v := i.(type) { ... }

BUENAS PRÁCTICAS:

1. Usa type definitions para IDs (UserID, OrderID) - previene mezclar
2. Usa type definitions con métodos para domain types
3. Usa type alias para refactoring gradual
4. Siempre usa assertion segura (con ok)
5. Implementa Stringer para mejor debugging
6. Los tipos función son útiles para callbacks y middleware
*/
