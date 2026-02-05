// Package main - Capítulo 14: Enumeradores
// Go no tiene enums nativos, pero usa iota y constantes tipadas.
// Aprenderás patrones de enum, String(), y mejores prácticas.
package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	fmt.Println("=== ENUMERADORES EN GO ===")

	// ============================================
	// IOTA BÁSICO
	// ============================================
	fmt.Println("\n--- Iota Básico ---")

	// iota es un contador que empieza en 0 y se incrementa
	const (
		Lunes = iota // 0
		Martes       // 1
		Miercoles    // 2
		Jueves       // 3
		Viernes      // 4
		Sabado       // 5
		Domingo      // 6
	)

	fmt.Printf("Lunes=%d, Martes=%d, Miercoles=%d\n", Lunes, Martes, Miercoles)
	fmt.Printf("Jueves=%d, Viernes=%d, Sabado=%d, Domingo=%d\n", Jueves, Viernes, Sabado, Domingo)

	// iota se reinicia en cada const block
	const (
		A = iota // 0
		B        // 1
	)
	const (
		C = iota // 0 (reinicia)
		D        // 1
	)
	fmt.Printf("A=%d, B=%d, C=%d, D=%d\n", A, B, C, D)

	// ============================================
	// ENUM TIPADO (PATRÓN RECOMENDADO)
	// ============================================
	fmt.Println("\n--- Enum Tipado ---")

	var estado Estado = EstadoActivo
	fmt.Printf("Estado: %d\n", estado)

	// Esto previene usar int directamente (type safety)
	// estado = 5 // compila pero no es un estado válido

	// ============================================
	// STRINGER (PARA DEBUGGING)
	// ============================================
	fmt.Println("\n--- Stringer ---")

	// Implementar String() para mejor output
	fmt.Printf("EstadoPendiente: %s\n", EstadoPendiente)
	fmt.Printf("EstadoActivo: %s\n", EstadoActivo)
	fmt.Printf("EstadoInactivo: %s\n", EstadoInactivo)

	// fmt usa String() automáticamente
	fmt.Println("Estado actual:", estado)

	// ============================================
	// IOTA CON EXPRESIONES
	// ============================================
	fmt.Println("\n--- Iota con Expresiones ---")

	// Tamaños de archivo
	const (
		_          = iota             // ignorar 0
		KB float64 = 1 << (10 * iota) // 1 << 10 = 1024
		MB                            // 1 << 20 = 1048576
		GB                            // 1 << 30
		TB                            // 1 << 40
	)

	fmt.Printf("KB=%.0f, MB=%.0f, GB=%.0f, TB=%.0f\n", KB, MB, GB, TB)

	tamanioArchivo := 1.5 * GB
	fmt.Printf("1.5 GB = %.0f bytes\n", tamanioArchivo)

	// Permisos estilo Unix (bit flags)
	const (
		Ejecutar = 1 << iota // 001
		Escribir             // 010
		Leer                 // 100
	)

	permisos := Leer | Escribir // 110
	fmt.Printf("Permisos: %03b\n", permisos)
	fmt.Printf("Puede leer: %t\n", permisos&Leer != 0)
	fmt.Printf("Puede escribir: %t\n", permisos&Escribir != 0)
	fmt.Printf("Puede ejecutar: %t\n", permisos&Ejecutar != 0)

	// ============================================
	// ENUM CON STRING VALUES
	// ============================================
	fmt.Println("\n--- Enum con String Values ---")

	// Para enums que necesitan ser strings (JSON, DB, etc.)
	type Role string

	const (
		RoleAdmin  Role = "admin"
		RoleUser   Role = "user"
		RoleGuest  Role = "guest"
	)

	var rol Role = RoleAdmin
	fmt.Printf("Rol: %s\n", rol)

	// Fácil de serializar a JSON
	usuario := struct {
		Nombre string `json:"nombre"`
		Rol    Role   `json:"rol"`
	}{
		Nombre: "Alice",
		Rol:    RoleAdmin,
	}

	jsonBytes, _ := json.Marshal(usuario)
	fmt.Printf("JSON: %s\n", jsonBytes)

	// ============================================
	// VALIDACIÓN DE ENUM
	// ============================================
	fmt.Println("\n--- Validación de Enum ---")

	// Función para validar
	if IsValidEstado(EstadoActivo) {
		fmt.Println("EstadoActivo es válido")
	}

	// Cuidado: Go no previene valores inválidos
	estadoInvalido := Estado(99)
	if !IsValidEstado(estadoInvalido) {
		fmt.Printf("Estado %d no es válido\n", estadoInvalido)
	}

	// ============================================
	// ENUM EXHAUSTIVO (patrón switch)
	// ============================================
	fmt.Println("\n--- Enum Exhaustivo ---")

	// El compilador NO verifica exhaustividad de switch
	// Usa linters como exhaustive para verificar
	procesarEstado(EstadoPendiente)
	procesarEstado(EstadoActivo)

	// ============================================
	// ENUM CON MÉTODOS
	// ============================================
	fmt.Println("\n--- Enum con Métodos ---")

	fmt.Printf("¿Activo está finalizado? %t\n", EstadoActivo.EsFinal())
	fmt.Printf("¿Eliminado está finalizado? %t\n", EstadoEliminado.EsFinal())

	// Transiciones de estado
	siguiente := EstadoPendiente.Siguiente()
	fmt.Printf("Siguiente de Pendiente: %s\n", siguiente)

	// ============================================
	// BITS FLAGS PATTERN
	// ============================================
	fmt.Println("\n--- Bit Flags Pattern ---")

	// Combinar permisos
	permisosUsuario := PermisoLeer | PermisoEscribir
	fmt.Printf("Permisos usuario: %04b\n", permisosUsuario)

	// Verificar permisos
	fmt.Printf("Puede leer: %t\n", permisosUsuario.Tiene(PermisoLeer))
	fmt.Printf("Puede eliminar: %t\n", permisosUsuario.Tiene(PermisoEliminar))

	// Agregar permiso
	permisosUsuario = permisosUsuario.Agregar(PermisoEjecutar)
	fmt.Printf("Después de agregar Ejecutar: %04b\n", permisosUsuario)

	// Quitar permiso
	permisosUsuario = permisosUsuario.Quitar(PermisoEscribir)
	fmt.Printf("Después de quitar Escribir: %04b\n", permisosUsuario)

	// Toggle
	permisosUsuario = permisosUsuario.Toggle(PermisoLeer)
	fmt.Printf("Después de toggle Leer: %04b\n", permisosUsuario)

	// ============================================
	// SENTINEL VALUES
	// ============================================
	fmt.Println("\n--- Sentinel Values ---")

	type Color int

	const (
		ColorDesconocido Color = iota // 0 como valor "inválido" o "no establecido"
		ColorRojo
		ColorVerde
		ColorAzul
	)

	var color Color // zero value es ColorDesconocido
	if color == ColorDesconocido {
		fmt.Println("Color no establecido")
	}

	color = ColorRojo
	fmt.Printf("Color: %d\n", color)

	// ============================================
	// EJEMPLO COMPLETO: MÁQUINA DE ESTADOS
	// ============================================
	fmt.Println("\n--- Ejemplo: Máquina de Estados ---")

	orderStatusNames := map[OrderStatus]string{
		OrderCreated:   "Creado",
		OrderPaid:      "Pagado",
		OrderShipped:   "Enviado",
		OrderDelivered: "Entregado",
		OrderCancelled: "Cancelado",
	}

	validTransitions := map[OrderStatus][]OrderStatus{
		OrderCreated:   {OrderPaid, OrderCancelled},
		OrderPaid:      {OrderShipped, OrderCancelled},
		OrderShipped:   {OrderDelivered},
		OrderDelivered: {},
		OrderCancelled: {},
	}

	orden := OrderCreated
	fmt.Printf("Estado inicial: %s\n", orderStatusNames[orden])

	// Intentar transición válida
	if canTransitionOrder(orden, OrderPaid, validTransitions) {
		orden = OrderPaid
		fmt.Printf("Transición a: %s\n", orderStatusNames[orden])
	}

	// Intentar transición inválida
	if !canTransitionOrder(orden, OrderDelivered, validTransitions) {
		fmt.Println("No se puede ir directamente de Pagado a Entregado")
	}
}

// ============================================
// IMPLEMENTACIONES
// ============================================

// Estado con String
type Estado int

const (
	EstadoPendiente Estado = iota
	EstadoActivo
	EstadoInactivo
	EstadoEliminado
	estadoMax // sentinel para validación
)

func (e Estado) String() string {
	nombres := [...]string{
		"Pendiente",
		"Activo",
		"Inactivo",
		"Eliminado",
	}
	if e < 0 || e >= estadoMax {
		return fmt.Sprintf("Estado(%d)", e)
	}
	return nombres[e]
}

func (e Estado) EsFinal() bool {
	return e == EstadoInactivo || e == EstadoEliminado
}

func (e Estado) Siguiente() Estado {
	switch e {
	case EstadoPendiente:
		return EstadoActivo
	case EstadoActivo:
		return EstadoInactivo
	default:
		return e
	}
}

func IsValidEstado(e Estado) bool {
	return e >= EstadoPendiente && e < estadoMax
}

func procesarEstado(e Estado) {
	switch e {
	case EstadoPendiente:
		fmt.Println("Procesando estado pendiente...")
	case EstadoActivo:
		fmt.Println("Procesando estado activo...")
	case EstadoInactivo:
		fmt.Println("Procesando estado inactivo...")
	case EstadoEliminado:
		fmt.Println("Procesando estado eliminado...")
	default:
		fmt.Printf("Estado desconocido: %d\n", e)
	}
}

// Permiso con métodos de bit flags
type Permiso uint8

const (
	PermisoLeer Permiso = 1 << iota
	PermisoEscribir
	PermisoEjecutar
	PermisoEliminar
)

func (p Permiso) Tiene(permiso Permiso) bool {
	return p&permiso != 0
}

func (p Permiso) Agregar(permiso Permiso) Permiso {
	return p | permiso
}

func (p Permiso) Quitar(permiso Permiso) Permiso {
	return p &^ permiso
}

func (p Permiso) Toggle(permiso Permiso) Permiso {
	return p ^ permiso
}

func (p Permiso) String() string {
	var perms []string
	if p.Tiene(PermisoLeer) {
		perms = append(perms, "Leer")
	}
	if p.Tiene(PermisoEscribir) {
		perms = append(perms, "Escribir")
	}
	if p.Tiene(PermisoEjecutar) {
		perms = append(perms, "Ejecutar")
	}
	if p.Tiene(PermisoEliminar) {
		perms = append(perms, "Eliminar")
	}
	if len(perms) == 0 {
		return "Sin permisos"
	}
	return fmt.Sprintf("[%v]", perms)
}

// OrderStatus for state machine
type OrderStatus int

const (
	OrderCreated OrderStatus = iota
	OrderPaid
	OrderShipped
	OrderDelivered
	OrderCancelled
)

// Máquina de estados para OrderStatus
func canTransitionOrder(current, target OrderStatus, validTransitions map[OrderStatus][]OrderStatus) bool {
	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}
	for _, s := range allowed {
		if s == target {
			return true
		}
	}
	return false
}

/*
RESUMEN:

IOTA:
- Contador que empieza en 0
- Se incrementa automáticamente
- Se reinicia en cada const block

PATRONES DE ENUM:

1. Simple:
   const ( A = iota; B; C )

2. Tipado (recomendado):
   type Estado int
   const ( Estado1 Estado = iota; ... )

3. Con expresiones:
   const ( KB = 1 << (10 * iota); MB; GB )

4. String values:
   type Role string
   const ( Admin Role = "admin"; ... )

5. Bit flags:
   const ( Read = 1 << iota; Write; Execute )

BUENAS PRÁCTICAS:

1. Siempre usa tipos personalizados para enums
2. Implementa String() para debugging
3. Agrega validación de valores
4. Usa _ = iota para ignorar el primer valor si 0 no es válido
5. Considera usar linter exhaustive para switches
6. Para bit flags, usa potencias de 2 (1 << iota)
7. Documenta los valores válidos
*/
