// Package main - Chapter 026: Packages
// Los paquetes son la unidad de organización de código en Go.
// Aprenderás a crear, estructurar y usar paquetes efectivamente.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== PAQUETES EN GO ===")

	// ============================================
	// ESTRUCTURA, DECLARACIÓN E IMPORTS
	// ============================================
	fmt.Println(`
Los paquetes son la unidad fundamental de organización en Go.
Cada archivo .go pertenece a un paquete.

ESTRUCTURA BÁSICA:
myproject/
├── go.mod
├── main.go           # package main
├── internal/         # paquetes privados
│   └── config/
│       └── config.go # package config
├── pkg/              # paquetes públicos (opcional)
│   └── utils/
│       └── utils.go  # package utils
└── cmd/              # múltiples ejecutables
    ├── server/
    │   └── main.go   # package main
    └── cli/
        └── main.go   # package main

---

DECLARACIÓN DE PAQUETE:

// Siempre primera línea del archivo
package mypackage

// El nombre del paquete:
// - Debe coincidir con el directorio (excepto main)
// - Minúsculas, sin guiones bajos ni mixedCaps
// - Corto pero descriptivo

---

IMPORTS:

// Import simple
import "fmt"

// Import múltiple
import (
    "fmt"
    "strings"
)

// Import con alias
import (
    "fmt"
    str "strings"  // alias
    _ "image/png"  // blank import (solo init)
    . "math"       // dot import (evitar)
)

// Import de paquete local
import "myproject/internal/config"

---

EXPORTACIÓN:

// Exportado (público) - empieza con mayúscula
func PublicFunction() {}
type PublicType struct {
    PublicField string
}
const PublicConst = 42
var PublicVar = "hello"

// No exportado (privado) - empieza con minúscula
func privateFunction() {}
type privateType struct {
    privateField string
}

---

INIT FUNCTION:

package mypackage

func init() {
    // Se ejecuta automáticamente al importar el paquete
    // Antes de main()
    // Puede haber múltiples init() en un archivo
    // Se ejecutan en orden de declaración
}

Orden de ejecución:
1. Variables de nivel de paquete se inicializan
2. init() de paquetes importados (en orden de import)
3. init() del paquete actual
4. main() si es el paquete principal

---

DOCUMENTACIÓN:

// Package mypackage provides utilities for...
//
// Example usage:
//
//     result := mypackage.DoSomething()
//
package mypackage

// DoSomething performs an important operation.
// It returns the result of the operation.
//
// Example:
//
//     result := DoSomething()
//     fmt.Println(result)
//
func DoSomething() string {
    return "done"
}

Ver documentación:
  go doc mypackage
  go doc mypackage.DoSomething
  godoc -http=:6060  # servidor web

---

INTERNAL PACKAGES:

myproject/
├── internal/
│   └── secret/      # Solo visible para myproject y sub-paquetes
│       └── secret.go
└── pkg/
    └── public/      # Visible para cualquiera
        └── public.go

Los paquetes en internal/ solo pueden ser importados por
paquetes en el mismo árbol de directorios.

---

VENDOR:

myproject/
├── vendor/          # Dependencias copiadas localmente
│   └── github.com/
│       └── some/
│           └── lib/
└── go.mod

go mod vendor  # Crea directorio vendor

Útil para:
- Builds reproducibles
- Entornos sin acceso a internet
- Control total de dependencias

---

BLANK IMPORTS:

import (
    _ "image/png"  // Registrar decoder PNG
    _ "net/http/pprof"  // Registrar endpoints de profiling
    _ "github.com/lib/pq"  // Registrar driver PostgreSQL
)

Usado para efectos secundarios de init().

---

CONVENCIONES DE NOMBRES:

✓ BIEN:
  - http (no HTTP)
  - url (no URL)
  - id (no ID excepto en tipos exportados: UserID)
  - json (no JSON)

✗ MAL:
  - my_package (usar mypackage)
  - myPackage (usar mypackage)
  - util (demasiado genérico)
  - common (demasiado genérico)

---

ORGANIZACIÓN RECOMENDADA:

Proyecto pequeño:
myproject/
├── go.mod
├── main.go
└── helpers.go

Proyecto mediano:
myproject/
├── go.mod
├── main.go
├── server/
│   └── server.go
├── handler/
│   └── handler.go
└── model/
    └── model.go

Proyecto grande:
myproject/
├── go.mod
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── cli/
│       └── main.go
├── internal/
│   ├── server/
│   ├── handler/
│   └── model/
├── pkg/
│   └── client/
└── api/
    └── openapi.yaml

---

EVITAR:

1. Paquetes circulares (A importa B, B importa A)
   → Extraer a un tercer paquete

2. Paquetes gigantes
   → Dividir por funcionalidad

3. Nombres genéricos (util, common, misc)
   → Nombres específicos (stringutil, httputil)

4. Demasiados paquetes pequeños
   → Agrupar funcionalidad relacionada

5. Acoplar paquetes innecesariamente
   → Interfaces en el consumidor`)
	// ============================================
	// EJEMPLO DE ESTRUCTURA DE PAQUETE
	// ============================================
	fmt.Println("\n--- Ejemplo de Estructura ---")

	fmt.Println(`
// user/user.go
package user

type User struct {
    ID    int
    Name  string
    Email string
}

func New(name, email string) *User {
    return &User{Name: name, Email: email}
}

func (u *User) Validate() error {
    if u.Name == "" {
        return errors.New("name is required")
    }
    return nil
}

// user/repository.go
package user

type Repository interface {
    Find(id int) (*User, error)
    Save(user *User) error
}

// user/service.go
package user

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetUser(id int) (*User, error) {
    return s.repo.Find(id)
}

---

// main.go
package main

import (
    "myproject/user"
    "myproject/postgres"
)

func main() {
    repo := postgres.NewUserRepository(db)
    svc := user.NewService(repo)

    u, err := svc.GetUser(1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(u.Name)
}`)}

/*
RESUMEN DE PAQUETES:

DECLARACIÓN:
package mypackage

IMPORTS:
import "fmt"
import "myproject/internal/config"
import alias "package"
import _ "package" // blank import
import . "package" // dot import (evitar)

EXPORTACIÓN:
- Mayúscula = público (exportado)
- Minúscula = privado (no exportado)

INIT:
func init() { }
- Ejecuta al importar
- Puede haber múltiples

ESTRUCTURA RECOMENDADA:
├── cmd/         # Ejecutables
├── internal/    # Paquetes privados
├── pkg/         # Paquetes públicos
└── go.mod

CONVENCIONES:
- Nombres cortos, minúsculas
- Sin guiones bajos
- Evitar nombres genéricos

EVITAR:
- Dependencias circulares
- Paquetes gigantes
- Nombres genéricos (util, common)
*/

/*
SUMMARY - CHAPTER 026: PACKAGES

PACKAGE BASICS:
- Packages are Go's fundamental unit of code organization
- Each .go file belongs to a package
- First line: package declaration
- Package names should be lowercase, no underscores

IMPORTS:
- Import single or multiple packages
- Import with alias, blank import for side effects
- Import local packages with module path

VISIBILITY:
- Uppercase = exported (public)
- Lowercase = unexported (private)
- Applies to functions, types, constants, variables

INITIALIZATION:
- init() functions run automatically on import
- Multiple init() allowed per file/package
- Execution order: imported packages → current package → main()

STRUCTURE:
- cmd/ for executables
- internal/ for private packages (import restricted)
- pkg/ for public reusable code
- vendor/ for local dependency copies

DOCUMENTATION:
- Comments above package/functions appear in godoc
- go doc to view documentation

BEST PRACTICES:
- Avoid circular dependencies
- Keep packages focused and cohesive
- Use descriptive names, avoid generic util/common
- Don't store context in structs
*/
