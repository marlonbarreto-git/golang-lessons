// Package main - Chapter 028: Modules
// Los módulos de Go gestionan dependencias y versionado.
// Introducidos en Go 1.11 y por defecto desde Go 1.16.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== MÓDULOS EN GO ===")

	// ============================================
	// MÓDULOS, GO.MOD Y COMANDOS
	// ============================================
	fmt.Println(`
Los módulos son la unidad de distribución de código en Go.
Un módulo es una colección de paquetes relacionados.

---

CREAR UN MÓDULO:

# Nuevo módulo
go mod init github.com/usuario/proyecto

# Esto crea go.mod:
module github.com/usuario/proyecto

go 1.26

---

ARCHIVO GO.MOD:

module github.com/usuario/proyecto

go 1.26

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
)

require (
    // indirect: dependencias de tus dependencias
    github.com/bytedance/sonic v1.9.1 // indirect
)

exclude (
    // Excluir versiones problemáticas
    github.com/broken/pkg v1.0.0
)

replace (
    // Reemplazar con versión local o fork
    github.com/original/pkg => ../my-fork
    github.com/original/pkg => github.com/myfork/pkg v1.0.0
)

retract (
    // Marcar versiones como retiradas (para mantenedores)
    v1.0.0 // Error crítico
    [v1.1.0, v1.2.0] // Rango de versiones
)

// Go 1.24+: Tool directives
tool golang.org/x/tools/cmd/stringer

---

ARCHIVO GO.SUM:

// Checksums criptográficos de dependencias
github.com/gin-gonic/gin v1.9.1 h1:4idEAncQnU5cB7BeOkP...
github.com/gin-gonic/gin v1.9.1/go.mod h1:RdK6Y52cP...

NUNCA editar manualmente. Git commit este archivo.

---

COMANDOS PRINCIPALES:

# Agregar dependencia
go get github.com/pkg/errors
go get github.com/pkg/errors@v0.9.1  # versión específica
go get github.com/pkg/errors@latest  # última versión

# Actualizar dependencias
go get -u ./...           # actualizar todas (minor/patch)
go get -u=patch ./...     # solo patches

# Limpiar dependencias no usadas
go mod tidy

# Verificar dependencias
go mod verify

# Descargar dependencias
go mod download

# Ver grafo de dependencias
go mod graph

# Ver por qué se necesita un módulo
go mod why github.com/some/module

# Crear vendor/
go mod vendor

# Editar go.mod
go mod edit -require github.com/pkg/errors@v0.9.1
go mod edit -replace old=new
go mod edit -droprequire github.com/old/dep

---

VERSIONADO SEMÁNTICO:

v1.2.3
│ │ │
│ │ └── PATCH: bug fixes compatibles
│ └──── MINOR: features compatibles
└────── MAJOR: cambios incompatibles

MAJOR v0.x.y: desarrollo, API inestable
MAJOR v1.x.y+: API estable, backward compatible

---

VERSIONES Y MÓDULOS:

v0 y v1: import "github.com/user/mod"
v2+:     import "github.com/user/mod/v2"

// Para v2+, el módulo necesita:
// 1. go.mod: module github.com/user/mod/v2
// 2. O usar /v2 subdirectory

---

PSEUDO-VERSIONES:

// Para commits sin tag
go get github.com/user/repo@abcd1234

// Formato: vX.Y.Z-yyyymmddhhmmss-abcdefabcdef
v0.0.0-20231015123456-abcdef123456

---

GO WORK (WORKSPACES):

# Para desarrollo multi-módulo local
go work init ./module1 ./module2

# Archivo go.work:
go 1.26

use (
    ./module1
    ./module2
)

replace github.com/user/module2 => ./module2

# Comandos
go work use ./new-module
go work sync

---

PRIVATE MODULES:

# Configurar acceso a repos privados
go env -w GOPRIVATE=github.com/mycompany/*
go env -w GONOPROXY=github.com/mycompany/*
go env -w GONOSUMDB=github.com/mycompany/*

# O en .gitconfig:
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/

---

PROXY Y CHECKSUM:

# Proxy por defecto
GOPROXY=https://proxy.golang.org,direct

# Sin proxy (acceso directo)
GOPROXY=direct

# Deshabilitar checksum DB
GONOSUMDB=*

# Checksum DB
GOSUMDB=sum.golang.org

---

TOOL DIRECTIVES (Go 1.24+):

// go.mod
module myproject

go 1.24

tool (
    golang.org/x/tools/cmd/stringer
    github.com/golangci/golangci-lint/cmd/golangci-lint
)

// Comandos
go get -tool golang.org/x/tools/cmd/stringer
go tool stringer -type=MyType
go install tool  // instalar todos los tools

---

BUENAS PRÁCTICAS:

1. Siempre hacer commit de go.mod y go.sum
2. go mod tidy antes de cada commit
3. Usar versiones específicas, no @latest en producción
4. Revisar actualizaciones de seguridad regularmente
5. Usar replace solo temporalmente en desarrollo
6. Documentar razones de exclude/replace
7. Para librerías: soportar últimas 2-3 versiones de Go`)
	// ============================================
	// EJEMPLO PRÁCTICO
	// ============================================
	fmt.Println("\n--- Ejemplo Práctico ---")

	fmt.Println(`
# Crear nuevo proyecto
mkdir myapi && cd myapi
go mod init github.com/user/myapi

# Agregar dependencias
go get github.com/gin-gonic/gin
go get github.com/lib/pq

# Ver go.mod resultante
cat go.mod
module github.com/user/myapi

go 1.26

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
)

# Limpiar
go mod tidy

# Para desarrollo con fork local
go mod edit -replace github.com/gin-gonic/gin=../my-gin-fork

# Quitar replace cuando termines
go mod edit -dropreplace github.com/gin-gonic/gin

# Actualizar una dependencia
go get -u github.com/gin-gonic/gin

# Ver todas las versiones disponibles
go list -m -versions github.com/gin-gonic/gin

# Downgrade si hay problemas
go get github.com/gin-gonic/gin@v1.8.0

# Verificar vulnerabilidades
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...`)}

/*
RESUMEN DE MÓDULOS:

CREAR:
go mod init module/path

ARCHIVOS:
- go.mod: define módulo y dependencias
- go.sum: checksums (siempre commit)

COMANDOS:
go get pkg@version   # agregar/actualizar
go mod tidy          # limpiar
go mod download      # descargar
go mod verify        # verificar
go mod graph         # dependencias
go mod vendor        # crear vendor/

VERSIONADO:
v1.2.3 = MAJOR.MINOR.PATCH
v2+ requiere /v2 en import path

DIRECTIVAS:
require  - dependencias
exclude  - excluir versiones
replace  - reemplazar módulos
retract  - retirar versiones (mantenedores)
tool     - herramientas de desarrollo (Go 1.24+)

PROXY:
GOPROXY=https://proxy.golang.org,direct

PRIVADOS:
GOPRIVATE=github.com/mycompany/*

WORKSPACES:
go work init ./mod1 ./mod2
*/

/*
SUMMARY - CHAPTER 028: MODULES

MODULE BASICS:
- Modules are the unit of code distribution in Go
- Created with go mod init <module-path>
- go.mod defines module and dependencies
- go.sum stores cryptographic checksums (always commit)

VERSIONING:
- Semantic versioning: vMAJOR.MINOR.PATCH
- v0 = development, unstable API
- v1+ = stable API, backward compatible
- v2+ requires /v2 in import path and module declaration

COMMANDS:
- go get pkg@version: add or update dependency
- go mod tidy: clean unused dependencies
- go mod download: download dependencies
- go mod verify: verify checksums
- go mod vendor: create local vendor directory

GO.MOD DIRECTIVES:
- require: declare dependencies
- exclude: block specific versions
- replace: use local or forked versions
- retract: mark versions as withdrawn
- tool: development tools (Go 1.24+)

PRIVATE MODULES:
- GOPRIVATE: skip proxy for private repos
- Configure git access via SSH or tokens

WORKSPACES (Go 1.18+):
- go work init: create workspace for multi-module development
- Useful for local development across modules

BEST PRACTICES:
- Always commit go.mod and go.sum
- Run go mod tidy before commits
- Use specific versions in production
- Review security updates regularly
*/
