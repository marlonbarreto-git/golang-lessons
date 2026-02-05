// Package main - Capítulo 12: Interfaces
// Las interfaces definen comportamiento, no datos.
// Aprenderás implementación implícita, composición, y patrones comunes.
package main

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
)

func main() {
	fmt.Println("=== INTERFACES EN GO ===")

	// ============================================
	// DEFINICIÓN BÁSICA
	// ============================================
	fmt.Println("\n--- Definición Básica ---")

	// PersonaSimple implementa Saludador porque tiene Saludar()
	// No declaramos que implementa nada

	p := PersonaSimple{Nombre: "Alice"}
	var s Saludador = p // funciona porque PersonaSimple implementa Saludador
	fmt.Println(s.Saludar())

	// ============================================
	// INTERFACES COMUNES
	// ============================================
	fmt.Println("\n--- Interfaces Comunes ---")

	// Stringer (fmt package)
	prod := Producto{Nombre: "Laptop", Precio: 999.99}
	fmt.Println(prod) // usa String() automáticamente

	// error interface
	err := miError("algo salió mal")
	fmt.Printf("Error: %v\n", err)

	// io.Reader / io.Writer
	reader := strings.NewReader("Hello, World!")
	buffer := make([]byte, 5)
	n, _ := reader.Read(buffer)
	fmt.Printf("Leído: %s (%d bytes)\n", buffer, n)

	// ============================================
	// INTERFACE VACÍA (any)
	// ============================================
	fmt.Println("\n--- Interface Vacía (any) ---")

	// interface{} o any puede contener cualquier valor
	var cualquiera any

	cualquiera = 42
	fmt.Printf("int: %v (tipo: %T)\n", cualquiera, cualquiera)

	cualquiera = "texto"
	fmt.Printf("string: %v (tipo: %T)\n", cualquiera, cualquiera)

	cualquiera = []int{1, 2, 3}
	fmt.Printf("slice: %v (tipo: %T)\n", cualquiera, cualquiera)

	// Función que acepta cualquier cosa
	imprimirCualquiera(42)
	imprimirCualquiera("hola")
	imprimirCualquiera([]string{"a", "b"})

	// ============================================
	// TYPE ASSERTION
	// ============================================
	fmt.Println("\n--- Type Assertion ---")

	var valor any = "texto"

	// Type assertion sin verificación (puede causar panic)
	str := valor.(string)
	fmt.Printf("Valor como string: %s\n", str)

	// Type assertion con verificación (segura)
	if num, ok := valor.(int); ok {
		fmt.Printf("Es int: %d\n", num)
	} else {
		fmt.Println("No es int")
	}

	// Type switch
	describir(42)
	describir("hola")
	describir(3.14)
	describir(true)

	// ============================================
	// COMPOSICIÓN DE INTERFACES
	// ============================================
	fmt.Println("\n--- Composición de Interfaces ---")

	// Las interfaces pueden componerse de otras interfaces
	// Ejemplo clásico: io.ReadWriter

	type Reader interface {
		Read(p []byte) (n int, err error)
	}

	type Writer interface {
		Write(p []byte) (n int, err error)
	}

	type ReadWriter interface {
		Reader
		Writer
	}

	// strings.Builder implementa Writer
	var writer io.Writer = &strings.Builder{}
	_, _ = writer.Write([]byte("Hola"))

	// ============================================
	// POLIMORFISMO
	// ============================================
	fmt.Println("\n--- Polimorfismo ---")

	// Diferentes tipos, misma interface
	formas := []Forma{
		Circulo{Radio: 5},
		RectanguloForma{Ancho: 4, Alto: 3},
		Triangulo{Base: 6, Altura: 4},
	}

	for _, forma := range formas {
		fmt.Printf("%s - Área: %.2f, Perímetro: %.2f\n",
			forma.Nombre(), forma.Area(), forma.Perimetro())
	}

	// ============================================
	// NIL INTERFACES
	// ============================================
	fmt.Println("\n--- Nil Interfaces ---")

	// Una interface es nil solo si tanto el tipo como el valor son nil
	var i interface{}
	fmt.Printf("i == nil: %t (tipo: %T)\n", i == nil, i)

	var ptr *int
	i = ptr
	fmt.Printf("i == nil después de asignar *int nil: %t (tipo: %T)\n",
		i == nil, i)
	// ¡i NO es nil! tiene tipo *int, aunque el valor sea nil

	// Esto es una trampa común en manejo de errores
	demostrarNilInterfaceTrap()

	// ============================================
	// INTERFACES PEQUEÑAS
	// ============================================
	fmt.Println("\n--- Interfaces Pequeñas (Go Idiomático) ---")
	fmt.Println(`
En Go, preferimos interfaces pequeñas:

type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Closer interface { Close() error }
type Stringer interface { String() string }
type Error interface { Error() string }

"The bigger the interface, the weaker the abstraction"
- Rob Pike`)
	// ============================================
	// VERIFICAR IMPLEMENTACIÓN EN COMPILE TIME
	// ============================================
	fmt.Println("\n--- Verificación en Compile Time ---")

	// Truco para verificar que un tipo implementa una interface
	var _ Forma = Circulo{}        // si Circulo no implementa Forma, no compila
	var _ Forma = RectanguloForma{} // lo mismo para Rectangulo

	fmt.Println("Circulo y Rectangulo implementan Forma ✓")

	// ============================================
	// SORT INTERFACE
	// ============================================
	fmt.Println("\n--- sort.Interface ---")

	personas := PersonasPorEdad{
		{Nombre: "Charlie", Edad: 25},
		{Nombre: "Alice", Edad: 30},
		{Nombre: "Bob", Edad: 20},
	}

	fmt.Println("Antes de ordenar:", personas)
	sort.Sort(personas)
	fmt.Println("Después de ordenar:", personas)

	// Go 1.8+ con sort.Slice (más simple)
	personas2 := []PersonaOrd{
		{Nombre: "Charlie", Edad: 25},
		{Nombre: "Alice", Edad: 30},
		{Nombre: "Bob", Edad: 20},
	}

	sort.Slice(personas2, func(i, j int) bool {
		return personas2[i].Nombre < personas2[j].Nombre
	})
	fmt.Println("Ordenado por nombre:", personas2)

	// ============================================
	// ACEPTAR INTERFACES, RETORNAR STRUCTS
	// ============================================
	fmt.Println("\n--- Aceptar Interfaces, Retornar Structs ---")
	fmt.Println(`
Principio de diseño en Go:

"Accept interfaces, return concrete types"

- Funciones deben ACEPTAR interfaces (flexibilidad)
- Funciones deben RETORNAR tipos concretos (claridad)

func ProcesarDatos(r io.Reader) *Resultado {
    // r puede ser cualquier Reader
    // Retornamos tipo concreto, no interface
}`)
	// ============================================
	// EJEMPLO: INYECCIÓN DE DEPENDENCIAS
	// ============================================
	fmt.Println("\n--- Inyección de Dependencias ---")

	// Definir interface para storage
	type UserStorage interface {
		GetUser(id int) (string, error)
		SaveUser(id int, name string) error
	}

	// Implementación real
	realDB := &RealDatabase{}

	// Implementación mock para tests
	mockDB := &MockDatabase{
		users: map[int]string{1: "TestUser"},
	}

	// El servicio acepta la interface, no la implementación
	service := UserService{storage: realDB}
	service.storage = mockDB // fácil de cambiar para tests

	nombre, _ := service.GetUserName(1)
	fmt.Printf("Usuario: %s\n", nombre)
}

// ============================================
// IMPLEMENTACIONES
// ============================================

// Saludador interface
type Saludador interface {
	Saludar() string
}

// Saludador implementation
type PersonaSimple struct {
	Nombre string
}

func (p PersonaSimple) Saludar() string {
	return fmt.Sprintf("¡Hola, soy %s!", p.Nombre)
}

// Stringer implementation
type Producto struct {
	Nombre string
	Precio float64
}

func (p Producto) String() string {
	return fmt.Sprintf("%s ($%.2f)", p.Nombre, p.Precio)
}

// error implementation
type miError string

func (e miError) Error() string {
	return string(e)
}

// Interface vacía helper
func imprimirCualquiera(v any) {
	fmt.Printf("Valor: %v, Tipo: %T\n", v, v)
}

// Type switch
func describir(v any) {
	switch val := v.(type) {
	case int:
		fmt.Printf("int: %d (el doble es %d)\n", val, val*2)
	case string:
		fmt.Printf("string: %q (longitud %d)\n", val, len(val))
	case float64:
		fmt.Printf("float64: %.2f\n", val)
	case bool:
		fmt.Printf("bool: %t\n", val)
	default:
		fmt.Printf("tipo desconocido: %T\n", val)
	}
}

// Formas - polimorfismo
type Forma interface {
	Area() float64
	Perimetro() float64
	Nombre() string
}

type Circulo struct {
	Radio float64
}

func (c Circulo) Area() float64      { return math.Pi * c.Radio * c.Radio }
func (c Circulo) Perimetro() float64 { return 2 * math.Pi * c.Radio }
func (c Circulo) Nombre() string     { return "Círculo" }

type RectanguloForma struct {
	Ancho, Alto float64
}

func (r RectanguloForma) Area() float64      { return r.Ancho * r.Alto }
func (r RectanguloForma) Perimetro() float64 { return 2 * (r.Ancho + r.Alto) }
func (r RectanguloForma) Nombre() string     { return "Rectángulo" }

type Triangulo struct {
	Base, Altura float64
}

func (t Triangulo) Area() float64 { return t.Base * t.Altura / 2 }
func (t Triangulo) Perimetro() float64 {
	// Asumiendo triángulo isósceles para simplificar
	lado := math.Sqrt(t.Altura*t.Altura + (t.Base/2)*(t.Base/2))
	return t.Base + 2*lado
}
func (t Triangulo) Nombre() string { return "Triángulo" }

// Nil interface trap
func demostrarNilInterfaceTrap() {
	var err error = getNilError()
	if err != nil {
		fmt.Println("¡err no es nil aunque el valor subyacente sí lo es!")
	}
}

func getNilError() *customError {
	return nil
}

type customError struct{}

func (e *customError) Error() string { return "error" }

// Sort interface
type PersonaOrd struct {
	Nombre string
	Edad   int
}

type PersonasPorEdad []PersonaOrd

func (p PersonasPorEdad) Len() int           { return len(p) }
func (p PersonasPorEdad) Less(i, j int) bool { return p[i].Edad < p[j].Edad }
func (p PersonasPorEdad) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Inyección de dependencias
type RealDatabase struct{}

func (db *RealDatabase) GetUser(id int) (string, error) {
	return "RealUser", nil
}

func (db *RealDatabase) SaveUser(id int, name string) error {
	return nil
}

type MockDatabase struct {
	users map[int]string
}

func (db *MockDatabase) GetUser(id int) (string, error) {
	if name, ok := db.users[id]; ok {
		return name, nil
	}
	return "", fmt.Errorf("user not found")
}

func (db *MockDatabase) SaveUser(id int, name string) error {
	db.users[id] = name
	return nil
}

type UserStorage interface {
	GetUser(id int) (string, error)
	SaveUser(id int, name string) error
}

type UserService struct {
	storage UserStorage
}

func (s *UserService) GetUserName(id int) (string, error) {
	return s.storage.GetUser(id)
}

/*
RESUMEN:

SINTAXIS:
type NombreInterface interface {
    Metodo1() TipoRetorno
    Metodo2(param Tipo) TipoRetorno
}

CARACTERÍSTICAS:
- Implementación IMPLÍCITA (no hay "implements")
- Una interface nil tiene tipo y valor nil
- interface{} / any puede contener cualquier valor
- Las interfaces pueden componerse

TYPE ASSERTION:
valor := i.(Tipo)         // panic si falla
valor, ok := i.(Tipo)     // seguro

TYPE SWITCH:
switch v := i.(type) {
case int: ...
case string: ...
}

INTERFACES ESTÁNDAR IMPORTANTES:
- fmt.Stringer: String() string
- error: Error() string
- io.Reader: Read([]byte) (int, error)
- io.Writer: Write([]byte) (int, error)
- io.Closer: Close() error
- sort.Interface: Len, Less, Swap

BUENAS PRÁCTICAS:

1. Interfaces pequeñas (1-3 métodos)
2. "Accept interfaces, return concrete types"
3. Define interfaces donde se usan, no donde se implementan
4. Usa interface{}/any solo cuando sea necesario
5. Verifica implementación en compile time: var _ Interface = Tipo{}
6. Cuidado con nil interface trap
*/
