// Package main - Capítulo 11: Structs
// Los structs son la forma de crear tipos compuestos en Go.
// Aprenderás definición, campos, métodos, embedding y tags.
package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func main() {
	fmt.Println("=== STRUCTS EN GO ===")

	// ============================================
	// DEFINICIÓN BÁSICA
	// ============================================
	fmt.Println("\n--- Definición Básica ---")

	// Crear instancias
	var p1 Persona // zero values: "", 0, ""
	fmt.Printf("p1 (zero): %+v\n", p1)

	p2 := Persona{
		Nombre: "Alice",
		Edad:   30,
		Email:  "alice@example.com",
	}
	fmt.Printf("p2: %+v\n", p2)

	// Sin nombres de campo (orden importa, no recomendado)
	p3 := Persona{"Bob", 25, "bob@example.com"}
	fmt.Printf("p3: %+v\n", p3)

	// Parcialmente inicializado
	p4 := Persona{Nombre: "Carol"} // Edad=0, Email=""
	fmt.Printf("p4: %+v\n", p4)

	// ============================================
	// ACCESO A CAMPOS
	// ============================================
	fmt.Println("\n--- Acceso a Campos ---")

	fmt.Printf("Nombre: %s\n", p2.Nombre)
	fmt.Printf("Edad: %d\n", p2.Edad)

	// Modificar campos
	p2.Edad = 31
	fmt.Printf("Nueva edad: %d\n", p2.Edad)

	// Con punteros - dereferencia automática
	ptr := &p2
	fmt.Printf("Via puntero: %s\n", ptr.Nombre) // equivale a (*ptr).Nombre

	// ============================================
	// STRUCTS ANÓNIMOS
	// ============================================
	fmt.Println("\n--- Structs Anónimos ---")

	// Útil para datos temporales o one-off
	config := struct {
		Host string
		Port int
	}{
		Host: "localhost",
		Port: 8080,
	}
	fmt.Printf("Config: %+v\n", config)

	// Común en tests
	testCases := []struct {
		input    int
		expected int
	}{
		{1, 2},
		{2, 4},
		{3, 6},
	}
	for _, tc := range testCases {
		fmt.Printf("Input: %d, Expected: %d\n", tc.input, tc.expected)
	}

	// ============================================
	// MÉTODOS
	// ============================================
	fmt.Println("\n--- Métodos ---")

	// Los métodos se definen fuera del struct
	rect := Rectangulo{Ancho: 10, Alto: 5}
	fmt.Printf("Rectángulo: %+v\n", rect)
	fmt.Printf("Área: %.2f\n", rect.Area())
	fmt.Printf("Perímetro: %.2f\n", rect.Perimetro())

	// Método con receptor puntero (puede modificar)
	rect.Escalar(2)
	fmt.Printf("Después de Escalar(2): %+v\n", rect)

	// ============================================
	// RECEPTOR VALOR VS PUNTERO
	// ============================================
	fmt.Println("\n--- Receptor Valor vs Puntero ---")
	fmt.Println(`
Usar receptor VALOR (*T) cuando:
- El struct es pequeño
- El método no modifica el struct
- Quieres inmutabilidad

Usar receptor PUNTERO (*T) cuando:
- El método necesita modificar el struct
- El struct es grande (evitar copias)
- Consistencia (si uno es puntero, todos lo son)`)
	// ============================================
	// EMBEDDING (Composición)
	// ============================================
	fmt.Println("\n--- Embedding (Composición) ---")

	type PersonaEmb struct {
		Nombre string
		Edad   int
		Email  string
	}

	type Direccion struct {
		Calle  string
		Ciudad string
		Pais   string
	}

	type Empleado struct {
		PersonaEmb // embedding - no tiene nombre de campo
		Direccion  // otro embedding
		Puesto     string
		Salario    float64
	}

	emp := Empleado{
		PersonaEmb: PersonaEmb{
			Nombre: "Diana",
			Edad:   28,
			Email:  "diana@empresa.com",
		},
		Direccion: Direccion{
			Calle:  "Calle Principal 123",
			Ciudad: "Madrid",
			Pais:   "España",
		},
		Puesto:  "Developer",
		Salario: 50000,
	}

	// Acceso directo a campos del embedded struct
	fmt.Printf("Nombre: %s\n", emp.Nombre)    // equivale a emp.PersonaEmb.Nombre
	fmt.Printf("Ciudad: %s\n", emp.Ciudad)    // equivale a emp.Direccion.Ciudad
	fmt.Printf("Puesto: %s\n", emp.Puesto)

	// También puedes acceder explícitamente
	fmt.Printf("Persona: %+v\n", emp.PersonaEmb)

	// Los métodos también se "heredan"
	// Si PersonaEmb tuviera métodos, Empleado los tendría

	// ============================================
	// COLISIÓN DE NOMBRES
	// ============================================
	fmt.Println("\n--- Colisión de Nombres ---")

	type A struct {
		Valor int
	}

	type B struct {
		Valor string
	}

	type C struct {
		A
		B
		Valor bool // campo propio tiene prioridad
	}

	c := C{
		A:     A{Valor: 42},
		B:     B{Valor: "texto"},
		Valor: true,
	}

	fmt.Printf("c.Valor: %v (bool, campo propio)\n", c.Valor)
	fmt.Printf("c.A.Valor: %v (int, de A)\n", c.A.Valor)
	fmt.Printf("c.B.Valor: %v (string, de B)\n", c.B.Valor)

	// ============================================
	// STRUCT TAGS
	// ============================================
	fmt.Println("\n--- Struct Tags ---")

	type Usuario struct {
		ID        int    `json:"id" db:"user_id"`
		Nombre    string `json:"nombre" db:"name"`
		Email     string `json:"email,omitempty" db:"email"`
		Password  string `json:"-" db:"password_hash"` // - omite en JSON
		CreatedAt string `json:"created_at,omitempty"`
	}

	user := Usuario{
		ID:     1,
		Nombre: "Alice",
		Email:  "alice@example.com",
	}

	// JSON usa los tags
	jsonBytes, _ := json.MarshalIndent(user, "", "  ")
	fmt.Printf("JSON:\n%s\n", jsonBytes)

	// Acceder a tags via reflection
	t := reflect.TypeOf(user)
	field, _ := t.FieldByName("Email")
	fmt.Printf("Tag json de Email: %s\n", field.Tag.Get("json"))
	fmt.Printf("Tag db de Email: %s\n", field.Tag.Get("db"))

	// ============================================
	// COMPARACIÓN DE STRUCTS
	// ============================================
	fmt.Println("\n--- Comparación de Structs ---")

	type Punto struct {
		X, Y int
	}

	punto1 := Punto{1, 2}
	punto2 := Punto{1, 2}
	punto3 := Punto{2, 3}

	fmt.Printf("punto1 == punto2: %t\n", punto1 == punto2)
	fmt.Printf("punto1 == punto3: %t\n", punto1 == punto3)

	// NOTA: Structs con campos no comparables (slices, maps, functions)
	// no se pueden comparar con ==
	// type NoComparable struct {
	//     Data []int
	// }
	// n1 == n2 // NO COMPILA

	// ============================================
	// CONSTRUCTORES
	// ============================================
	fmt.Println("\n--- Constructores (funciones New) ---")

	type Servidor struct {
		host    string
		port    int
		timeout int
	}

	// Por convención, funciones New* actúan como constructores
	newServer := func(host string, port int) *Servidor {
		return &Servidor{
			host:    host,
			port:    port,
			timeout: 30, // valor default
		}
	}

	srv := newServer("localhost", 8080)
	fmt.Printf("Servidor: %+v\n", srv)

	// ============================================
	// OPCIONES FUNCIONALES (patrón avanzado)
	// ============================================
	fmt.Println("\n--- Opciones Funcionales ---")

	srv2 := NewServidor("api.example.com",
		WithPort(443),
		WithTimeout(60),
		WithTLS(true),
	)
	fmt.Printf("Servidor con opciones: %+v\n", srv2)

	// ============================================
	// STRUCT VACÍO
	// ============================================
	fmt.Println("\n--- Struct Vacío ---")

	// struct{} no ocupa memoria (0 bytes)
	type Vacio struct{}
	fmt.Printf("Tamaño de struct{}: %d bytes\n", reflect.TypeOf(Vacio{}).Size())

	// Útil para sets
	set := map[string]struct{}{
		"a": {},
		"b": {},
	}
	if _, exists := set["a"]; exists {
		fmt.Println("'a' existe en el set")
	}

	// Útil para señalización en channels
	done := make(chan struct{})
	go func() {
		// trabajo...
		close(done) // señal de completado
	}()
	<-done

	// ============================================
	// COPIA DE STRUCTS
	// ============================================
	fmt.Println("\n--- Copia de Structs ---")

	originalP := PersonaEmb{Nombre: "Original", Edad: 30}
	copiaP := originalP // copia por valor

	copiaP.Nombre = "Copia"
	fmt.Printf("Original: %+v\n", originalP) // no afectado
	fmt.Printf("Copia: %+v\n", copiaP)

	// Cuidado con campos que son referencias (slices, maps, pointers)
	type ConSlice struct {
		Datos []int
	}

	orig := ConSlice{Datos: []int{1, 2, 3}}
	cop := orig // copia el header del slice, NO los datos

	cop.Datos[0] = 999
	fmt.Printf("Original con slice: %+v (¡modificado!)\n", orig)
	fmt.Printf("Copia con slice: %+v\n", cop)
}

// Persona struct for basic examples
type Persona struct {
	Nombre string
	Edad   int
	Email  string
}

// Métodos de Rectangulo
type Rectangulo struct {
	Ancho, Alto float64
}

func (r Rectangulo) Area() float64 {
	return r.Ancho * r.Alto
}

func (r Rectangulo) Perimetro() float64 {
	return 2 * (r.Ancho + r.Alto)
}

func (r *Rectangulo) Escalar(factor float64) {
	r.Ancho *= factor
	r.Alto *= factor
}

// Servidor con opciones funcionales
type ServidorConfig struct {
	Host    string
	Port    int
	Timeout int
	TLS     bool
}

type ServerOption func(*ServidorConfig)

func WithPort(port int) ServerOption {
	return func(s *ServidorConfig) {
		s.Port = port
	}
}

func WithTimeout(timeout int) ServerOption {
	return func(s *ServidorConfig) {
		s.Timeout = timeout
	}
}

func WithTLS(tls bool) ServerOption {
	return func(s *ServidorConfig) {
		s.TLS = tls
	}
}

func NewServidor(host string, opts ...ServerOption) *ServidorConfig {
	srv := &ServidorConfig{
		Host:    host,
		Port:    80,
		Timeout: 30,
		TLS:     false,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv
}

/*
RESUMEN:

DEFINICIÓN:
type NombreStruct struct {
    Campo1 Tipo1
    Campo2 Tipo2
}

CARACTERÍSTICAS:
- Zero value: todos los campos con sus zero values
- Comparables si todos los campos son comparables
- Se copian por valor (cuidado con campos referencia)

MÉTODOS:
func (r Receptor) Metodo() { }  // receptor valor
func (r *Receptor) Metodo() { } // receptor puntero

EMBEDDING:
- Composición sobre herencia
- Campos y métodos se "promueven"
- Colisiones se resuelven explícitamente

STRUCT TAGS:
`nombre:"valor" otro:"valor2"`
- Usados por encoding/json, database drivers, validators, etc.

BUENAS PRÁCTICAS:

1. Usa nombres descriptivos para campos
2. Agrupa campos relacionados
3. Documenta structs públicos
4. Usa constructores (New*) para inicialización compleja
5. Sé consistente con receptores (todos valor o todos puntero)
6. Usa embedding para composición, no para "herencia"
7. Cuidado con copiar structs con campos referencia
*/
