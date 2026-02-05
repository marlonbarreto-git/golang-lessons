// Package main - Capitulo 76: CI/CD, Makefile & Go Tooling
// Automatizacion de builds, testing, linting, releases y flujos CI/CD
// profesionales para proyectos Go.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== CI/CD, MAKEFILE & GO TOOLING ===")

	// ============================================
	// MAKEFILE PATTERNS FOR GO
	// ============================================
	fmt.Println("\n--- Makefile Patterns for Go ---")
	os.Stdout.WriteString(`
MAKEFILE PROFESIONAL PARA GO:

Un Makefile bien estructurado es la base de todo proyecto Go.
Define targets claros y consistentes para todo el equipo.

# Variables al inicio
BINARY_NAME=myapp
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"
GOFILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Phony targets (no generan archivos)
.PHONY: all build test lint fmt vet run clean docker release help

# Default target
all: lint test build

# ===== BUILD =====
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/server

build-all: ## Cross-compile para multiples plataformas
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/server
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/server

# ===== TEST =====
test: ## Ejecutar tests
	go test -race -count=1 ./...

test-verbose: ## Tests con output detallado
	go test -race -count=1 -v ./...

test-short: ## Solo tests rapidos
	go test -short -count=1 ./...

test-coverage: ## Tests con reporte de coverage
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Tests de integracion (requieren servicios externos)
	go test -race -tags=integration -count=1 -timeout=5m ./...

# ===== QUALITY =====
lint: ## Ejecutar linter
	golangci-lint run ./...

fmt: ## Formatear codigo
	gofmt -s -w $(GOFILES)
	goimports -w $(GOFILES)

fmt-check: ## Verificar formato sin modificar
	@test -z "$$(gofmt -l $(GOFILES))" || (echo "Archivos sin formatear:" && gofmt -l $(GOFILES) && exit 1)

vet: ## Analisis estatico basico
	go vet ./...

vuln: ## Escaneo de vulnerabilidades
	govulncheck ./...

# ===== RUN =====
run: ## Ejecutar la aplicacion
	go run ./cmd/server

run-dev: ## Ejecutar con hot reload (requiere air)
	air

# ===== CLEAN =====
clean: ## Limpiar artefactos
	rm -rf bin/ coverage.out coverage.html
	go clean -cache -testcache

# ===== DOCKER =====
docker-build: ## Construir imagen Docker
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

docker-push: ## Push imagen a registry
	docker push $(BINARY_NAME):$(VERSION)
	docker push $(BINARY_NAME):latest

# ===== RELEASE =====
release: ## Release con goreleaser
	goreleaser release --clean

release-snapshot: ## Release local sin publicar
	goreleaser release --snapshot --clean

# ===== TOOLS =====
tools: ## Instalar herramientas de desarrollo
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/air-verse/air@latest
	go install github.com/goreleaser/goreleaser/v2@latest

# ===== HELP =====
help: ## Mostrar esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

TIPS:
- Usa @echo para mensajes, @ silencia el comando
- ?= permite override: make build VERSION=2.0.0
- .PHONY evita conflictos con archivos del mismo nombre
- $(shell ...) ejecuta comandos del sistema
- Alinear con tabs, NO espacios (requisito de Make)
`)

	// ============================================
	// GITHUB ACTIONS FOR GO
	// ============================================
	fmt.Println("\n--- GitHub Actions for Go ---")
	fmt.Println(`
WORKFLOW BASICO (.github/workflows/ci.yml):

name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

permissions:
  contents: read

env:
  GO_VERSION: '1.23'

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.22', '1.23']
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: |
            ~/go/pkg/mod
            ~/.cache/go-build
          key: ${{ runner.os }}-go-${{ matrix.go-version }}-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-${{ matrix.go-version }}-

      - name: Download dependencies
        run: go mod download

      - name: Verify dependencies
        run: go mod verify

      - name: Run go vet
        run: go vet ./...

      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  vulnerability-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run govulncheck
        run: govulncheck ./...`)

	// ============================================
	// GITHUB ACTIONS - RELEASE WORKFLOW
	// ============================================
	fmt.Println("\n--- GitHub Actions: Release Workflow ---")
	fmt.Println(`
WORKFLOW DE RELEASE (.github/workflows/release.yml):

name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write
  packages: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  docker:
    runs-on: ubuntu-latest
    needs: release
    steps:
      - uses: actions/checkout@v4

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: |
            ghcr.io/${{ github.repository }}:${{ github.ref_name }}
            ghcr.io/${{ github.repository }}:latest`)

	// ============================================
	// GOLANGCI-LINT CONFIGURATION
	// ============================================
	fmt.Println("\n--- golangci-lint Configuration ---")
	fmt.Println(`
ARCHIVO .golangci.yml:

golangci-lint es EL linter estandar de Go. Ejecuta
decenas de linters en paralelo con una sola herramienta.

run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    # Errores comunes
    - errcheck        # Verifica que errores se manejan
    - govet           # Analisis similar a go vet pero mas
    - staticcheck     # Mega-linter (reemplaza golint)

    # Estilo y consistencia
    - gofmt           # Formato estandar
    - goimports       # Imports ordenados
    - revive          # Sucesor de golint
    - misspell        # Detecta typos en ingles

    # Bugs potenciales
    - bodyclose       # Verifica que http.Response.Body se cierra
    - noctx           # Verifica que requests usan context
    - sqlclosecheck   # Verifica que sql.Rows se cierra
    - rowserrcheck    # Verifica sql.Rows.Err()

    # Performance
    - prealloc        # Sugiere pre-allocar slices
    - gocritic        # Analisis de estilo avanzado

    # Seguridad
    - gosec           # Detecta problemas de seguridad

  disable:
    - wsl             # Demasiado opinionado sobre whitespace
    - gochecknoglobals # Globals son validos en muchos casos
    - exhaustruct     # Fuerza inicializar todos los campos

linters-settings:
  govet:
    enable-all: true
  revive:
    rules:
      - name: exported
        arguments:
          - "checkPrivateReceivers"
      - name: unexported-return
        disabled: true
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
  gosec:
    excludes:
      - G104  # Ignorar errors de audit (auditado manualmente)

issues:
  max-issues-per-linter: 50
  max-same-issues: 3
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - errcheck
    - path: cmd/
      linters:
        - gochecknoinits

INSTALAR:
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

EJECUTAR:
  golangci-lint run ./...                     # Todo el proyecto
  golangci-lint run --fix ./...               # Auto-fix
  golangci-lint run --new-from-rev=HEAD~1     # Solo cambios recientes
  golangci-lint linters                       # Listar linters activos`)

	// ============================================
	// PRE-COMMIT HOOKS
	// ============================================
	fmt.Println("\n--- Pre-commit Hooks ---")
	fmt.Println(`
PRE-COMMIT HOOKS PARA GO:

Opcion 1: Script manual (.git/hooks/pre-commit)

#!/bin/bash
set -e

echo "Running pre-commit checks..."

# Verificar formato
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "ERROR: Archivos sin formatear:"
    echo "$UNFORMATTED"
    echo "Ejecuta: gofmt -s -w ."
    exit 1
fi

# go vet
echo "Running go vet..."
go vet ./...

# Tests rapidos
echo "Running short tests..."
go test -short -count=1 ./...

# golangci-lint (solo archivos staged)
echo "Running golangci-lint..."
golangci-lint run --new-from-rev=HEAD ./...

echo "All pre-commit checks passed!"

# Hacerlo ejecutable:
chmod +x .git/hooks/pre-commit


Opcion 2: pre-commit framework (pre-commit.com)

# .pre-commit-config.yaml
repos:
  - repo: https://github.com/tekwizely/pre-commit-golang
    rev: v1.0.0-rc.1
    hooks:
      - id: go-fmt
      - id: go-vet
      - id: go-lint
      - id: go-build
      - id: go-test-mod
        args: [-short, -count=1]

  - repo: https://github.com/golangci/golangci-lint
    rev: v1.61.0
    hooks:
      - id: golangci-lint

INSTALAR:
  pip install pre-commit
  pre-commit install
  pre-commit run --all-files  # Verificar setup


Opcion 3: lefthook (alternativa rapida, escrita en Go)

# lefthook.yml
pre-commit:
  parallel: true
  commands:
    fmt-check:
      glob: "*.go"
      run: test -z "$(gofmt -l {staged_files})"
    vet:
      glob: "*.go"
      run: go vet ./...
    lint:
      glob: "*.go"
      run: golangci-lint run --new-from-rev=HEAD

pre-push:
  commands:
    test:
      run: go test -race -count=1 ./...

INSTALAR:
  go install github.com/evilmartians/lefthook@latest
  lefthook install`)

	// ============================================
	// GOVULNCHECK
	// ============================================
	fmt.Println("\n--- govulncheck: Vulnerability Scanning ---")
	fmt.Println(`
GOVULNCHECK - HERRAMIENTA OFICIAL DE VULNERABILIDADES:

A diferencia de otros scanners, govulncheck analiza el call graph
real de tu codigo. Solo reporta vulnerabilidades que TU CODIGO
realmente puede alcanzar.

INSTALAR:
  go install golang.org/x/vuln/cmd/govulncheck@latest

USAR:
  govulncheck ./...

OUTPUT DE EJEMPLO:
  Scanning your code and 47 packages across 12 dependent modules
  for known vulnerabilities...

  Vulnerability #1: GO-2024-2887
    HTTP/2 CONTINUATION flood in net/http
    More info: https://pkg.go.dev/vuln/GO-2024-2887
    Module: golang.org/x/net
    Found in: golang.org/x/net@v0.17.0
    Fixed in: golang.org/x/net@v0.23.0
    Example trace found:
      myapp/server.go:45 -> net/http.ListenAndServe
                          -> golang.org/x/net/http2.(*serverConn).serve

MODOS DE ANALISIS:
  govulncheck ./...              # Analiza source code
  govulncheck -mode=binary myapp # Analiza binario compilado
  govulncheck -json ./...        # Output JSON para CI
  govulncheck -show=verbose ./...# Detalles completos

EN CI (fail si hay vulns):
  govulncheck ./...
  # Exit code 0: sin vulnerabilidades
  # Exit code 3: vulnerabilidades encontradas

EXCLUIR FALSOS POSITIVOS:
  No hay forma de excluir directamente. Opciones:
  1. Actualizar dependencia (preferido)
  2. Reemplazar dependencia
  3. Aceptar riesgo y documentar`)

	// ============================================
	// GORELEASER
	// ============================================
	fmt.Println("\n--- GoReleaser: Cross-compilation & Packaging ---")
	os.Stdout.WriteString(`
GORELEASER - RELEASE AUTOMATIZADO:

Compila, empaqueta, firma y publica releases de Go
para multiples plataformas automaticamente.

INSTALAR:
  go install github.com/goreleaser/goreleaser/v2@latest

INICIALIZAR:
  goreleaser init    # Crea .goreleaser.yaml base

ARCHIVO .goreleaser.yaml:

version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: myapp
    main: ./cmd/server
    binary: myapp
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
    mod_timestamp: '{{ .CommitTimestamp }}'

archives:
  - id: default
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md

checksum:
  name_template: 'checksums.txt'
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

dockers:
  - image_templates:
      - "ghcr.io/myorg/myapp:{{ .Version }}-amd64"
    dockerfile: Dockerfile
    build_flag_templates:
      - "--platform=linux/amd64"
    goarch: amd64

  - image_templates:
      - "ghcr.io/myorg/myapp:{{ .Version }}-arm64"
    dockerfile: Dockerfile
    build_flag_templates:
      - "--platform=linux/arm64"
    goarch: arm64

docker_manifests:
  - name_template: "ghcr.io/myorg/myapp:{{ .Version }}"
    image_templates:
      - "ghcr.io/myorg/myapp:{{ .Version }}-amd64"
      - "ghcr.io/myorg/myapp:{{ .Version }}-arm64"

nfpms:
  - id: packages
    package_name: myapp
    vendor: MyOrg
    homepage: https://github.com/myorg/myapp
    maintainer: Team <team@myorg.com>
    description: My awesome Go application
    license: MIT
    formats:
      - deb
      - rpm
      - apk
    contents:
      - src: ./config.yaml.example
        dst: /etc/myapp/config.yaml
        type: config
      - src: ./myapp.service
        dst: /etc/systemd/system/myapp.service

brews:
  - repository:
      owner: myorg
      name: homebrew-tap
    homepage: https://github.com/myorg/myapp
    description: My awesome Go application

COMANDOS:
  goreleaser check                 # Validar config
  goreleaser release --snapshot    # Release local (sin publicar)
  goreleaser release --clean       # Release real
  goreleaser build --single-target # Solo plataforma actual

TAGS & VERSIONADO:
  git tag -a v1.0.0 -m "Release v1.0.0"
  git push origin v1.0.0
  # GitHub Actions detecta el tag y ejecuta GoReleaser
`)

	// ============================================
	// GO TOOL COVER
	// ============================================
	fmt.Println("\n--- go tool cover: Coverage Reports ---")
	os.Stdout.WriteString(`
COVERAGE EN GO - GUIA COMPLETA:

GENERAR COVERAGE:
  go test -coverprofile=coverage.out ./...

  # Con race detector
  go test -race -coverprofile=coverage.out -covermode=atomic ./...

VER COVERAGE:
  # Resumen en terminal
  go tool cover -func=coverage.out

  # Output de ejemplo:
  myapp/handler.go:15:    HandleCreate    100.0%
  myapp/handler.go:45:    HandleGet       85.7%
  myapp/service.go:20:    Process         92.3%
  total:                  (statements)    90.5%

  # Reporte HTML interactivo
  go tool cover -html=coverage.out -o coverage.html

COVERAGE POR PAQUETE:
  go test -coverprofile=coverage.out -coverpkg=./... ./...
  # -coverpkg incluye paquetes que no tienen tests propios

MERGE PROFILES (multiples test runs):
  # Go 1.20+ soporta merge nativo
  go tool covdata merge -i=dir1,dir2 -o=merged

COVERAGE EN CI:
  # Fail si coverage < umbral
  COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
  if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage ${COVERAGE}% is below 80% threshold"
    exit 1
  fi

INTEGRATION TEST COVERAGE (Go 1.20+):
  # Compilar binario con coverage
  go build -cover -o myapp-cover ./cmd/server

  # Ejecutar con GOCOVERDIR
  GOCOVERDIR=./coverdata ./myapp-cover &
  # ... ejecutar integration tests ...
  kill $!

  # Generar reporte
  go tool covdata textfmt -i=./coverdata -o=integration.out
  go tool cover -html=integration.out
`)

	// ============================================
	// PPROF IN CI
	// ============================================
	fmt.Println("\n--- go tool pprof in CI ---")
	os.Stdout.WriteString(`
PPROF EN CI - DETECTAR REGRESIONES DE PERFORMANCE:

BENCHMARK EN CI:
  # Ejecutar benchmarks y guardar resultado
  go test -bench=. -benchmem -count=5 ./... > bench-new.txt

  # Comparar con baseline
  go install golang.org/x/perf/cmd/benchstat@latest
  benchstat bench-old.txt bench-new.txt

OUTPUT DE BENCHSTAT:
  goos: linux
  goarch: amd64
  pkg: myapp/service
                  │ bench-old.txt │          bench-new.txt           │
                  │    sec/op     │   sec/op     vs base             │
  Process-8          1.234µ ± 2%    0.987µ ± 1%  -20.02% (p=0.001)
  ProcessBatch-8     45.67µ ± 3%    44.12µ ± 2%   ~ (p=0.089)

CI WORKFLOW PARA BENCHMARKS:

name: Benchmark

on:
  pull_request:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run benchmarks
        run: go test -bench=. -benchmem -count=5 ./... > bench.txt

      - name: Compare with main
        uses: benchmark-action/github-action-benchmark@v1
        with:
          tool: 'go'
          output-file-path: bench.txt
          alert-threshold: '150%'
          comment-on-alert: true
          fail-on-alert: true

PPROF PROFILE EN CI:
  # Generar CPU profile durante tests
  go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...

  # Analizar en texto (para CI logs)
  go tool pprof -top -cum cpu.prof
  go tool pprof -top mem.prof

  # Detectar memory leaks
  go tool pprof -inuse_space mem.prof
`)

	// ============================================
	// DEPENDABOT / RENOVATE
	// ============================================
	fmt.Println("\n--- Dependabot / Renovate ---")
	fmt.Println(`
DEPENDABOT PARA GO (.github/dependabot.yml):

version: 2
updates:
  # Go modules
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
      timezone: "America/New_York"
    open-pull-requests-limit: 10
    reviewers:
      - "myorg/backend-team"
    labels:
      - "dependencies"
      - "go"
    commit-message:
      prefix: "chore(deps):"
    groups:
      golang-x:
        patterns:
          - "golang.org/x/*"
      aws:
        patterns:
          - "github.com/aws/*"

  # GitHub Actions
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    labels:
      - "ci"


RENOVATE (alternativa mas potente):

// renovate.json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": [
    "config:recommended",
    ":pinAllExceptPeerDependencies"
  ],
  "go": {
    "enabled": true,
    "postUpdateOptions": ["gomodTidy"]
  },
  "packageRules": [
    {
      "matchManagers": ["gomod"],
      "matchUpdateTypes": ["minor", "patch"],
      "groupName": "go minor/patch updates",
      "automerge": true
    },
    {
      "matchManagers": ["gomod"],
      "matchUpdateTypes": ["major"],
      "groupName": "go major updates",
      "automerge": false
    },
    {
      "matchPackagePatterns": ["golang.org/x/.*"],
      "groupName": "golang.org/x packages"
    }
  ],
  "vulnerabilityAlerts": {
    "enabled": true,
    "labels": ["security"]
  }
}

DIFERENCIAS:
  Dependabot:
  - Nativo de GitHub, zero config
  - Funcionalidad limitada pero suficiente
  - Solo un ecosistema por directorio

  Renovate:
  - Mas configurable y potente
  - Automerge, grouping avanzado
  - Soporta mono-repos complejos
  - Disponible como GitHub App o self-hosted`)

	// ============================================
	// DOCKER MULTI-STAGE BUILDS (DETAILED)
	// ============================================
	fmt.Println("\n--- Docker Multi-Stage Builds ---")
	fmt.Println(`
DOCKERFILE MULTI-STAGE PROFESIONAL:

# ===== Stage 1: Dependencies =====
FROM golang:1.23-alpine AS deps

WORKDIR /app

# Solo copiar go.mod/go.sum para cachear deps
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ===== Stage 2: Build =====
FROM deps AS builder

COPY . .

# Tests en el build (opcional pero recomendado)
RUN go test -short ./...

# Build estatico
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/server \
    ./cmd/server

# ===== Stage 3: Runtime =====
FROM alpine:3.20 AS runtime

# Certificados SSL y timezone data
RUN apk --no-cache add ca-certificates tzdata

# Usuario no-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/config.yaml ./config.yaml

USER appuser

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]


VARIANTE SCRATCH (imagen minima, ~5MB):

FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/server ./cmd/server

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]

NOTA SOBRE SCRATCH:
- Sin shell (no puedes hacer exec/sh)
- Sin libc (necesitas CGO_ENABLED=0)
- Sin herramientas de debug
- Ideal para produccion, malo para troubleshooting


VARIANTE DISTROLESS (Google):

FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]

DISTROLESS vs SCRATCH vs ALPINE:
  scratch:     ~5MB   Sin nada, max seguridad, zero debug
  distroless:  ~7MB   CA certs, tzdata, nonroot user built-in
  alpine:      ~12MB  Shell, package manager, herramientas basicas

RECOMENDACION:
  Produccion: distroless (mejor balance seguridad/usabilidad)
  Max seguridad: scratch
  Desarrollo: alpine

DOCKER BUILD OPTIMIZATIONS:
  - Separar COPY go.mod de COPY . para cachear deps
  - Usar .dockerignore para excluir archivos innecesarios
  - Multi-platform builds con docker buildx`)

	// ============================================
	// SEMANTIC VERSIONING & RELEASE TAGGING
	// ============================================
	fmt.Println("\n--- Semantic Versioning & Release Tagging ---")
	fmt.Println(`
SEMANTIC VERSIONING (SemVer) EN GO:

FORMATO: vMAJOR.MINOR.PATCH
  v1.0.0  -> Release inicial estable
  v1.1.0  -> Nueva funcionalidad backward-compatible
  v1.1.1  -> Bug fix backward-compatible
  v2.0.0  -> Breaking changes

CUANDO INCREMENTAR:
  MAJOR: Cambios que rompen compatibilidad
    - Eliminar funciones/tipos publicos
    - Cambiar signatures de funciones
    - Cambiar comportamiento existente
    - En Go modules: v2+ requiere /v2 en import path

  MINOR: Nueva funcionalidad backward-compatible
    - Nuevas funciones, tipos, metodos
    - Nuevos campos opcionales en structs
    - Deprecation de funciones (sin eliminar)

  PATCH: Bug fixes backward-compatible
    - Correccion de bugs
    - Mejoras de performance
    - Actualizacion de dependencias (minor/patch)

PRE-RELEASE:
  v1.0.0-alpha.1    -> Desarrollo temprano
  v1.0.0-beta.1     -> Feature-complete, testing
  v1.0.0-rc.1       -> Release candidate

RELEASE WORKFLOW:
  # Crear tag anotado
  git tag -a v1.0.0 -m "Release v1.0.0: initial stable release"
  git push origin v1.0.0

  # Listar tags
  git tag -l 'v*'

  # Eliminar tag (si es necesario)
  git tag -d v1.0.0
  git push origin --delete v1.0.0

GO MODULES Y MAJOR VERSIONS:
  v0.x.x y v1.x.x: import path normal
    import "github.com/myorg/mylib"

  v2.x.x+: REQUIERE /v2 en module path
    module github.com/myorg/mylib/v2  (en go.mod)
    import "github.com/myorg/mylib/v2" (en codigo)

  Estrategias:
  - Major branch: rama v2 con go.mod actualizado
  - Major subdirectory: directorio v2/ en el repo`)

	// ============================================
	// CHANGELOG GENERATION
	// ============================================
	fmt.Println("\n--- Changelog Generation ---")
	fmt.Println(`
GENERACION AUTOMATICA DE CHANGELOGS:

Opcion 1: GoReleaser (incluido)
  GoReleaser genera changelog automaticamente basado en commits
  entre el tag anterior y el actual.

  changelog:
    sort: asc
    groups:
      - title: Features
        regexp: '^.*?feat(\([[:word:]]+\))??!?:.+$'
        order: 0
      - title: Bug Fixes
        regexp: '^.*?fix(\([[:word:]]+\))??!?:.+$'
        order: 1
      - title: Performance
        regexp: '^.*?perf(\([[:word:]]+\))??!?:.+$'
        order: 2
      - title: Others
        order: 999

Opcion 2: git-cliff (herramienta dedicada)
  INSTALAR:
    cargo install git-cliff  # o brew install git-cliff

  cliff.toml:
  [changelog]
  header = "# Changelog\n\n"
  body = """
  {% for group, commits in commits | group_by(attribute="group") %}
  ### {{ group | upper_first }}
  {% for commit in commits %}
  - {{ commit.message | upper_first }} ({{ commit.id | truncate(length=7) }})
  {% endfor %}
  {% endfor %}
  """

  [git]
  conventional_commits = true
  filter_unconventional = true
  commit_parsers = [
    { message = "^feat", group = "Features" },
    { message = "^fix", group = "Bug Fixes" },
    { message = "^perf", group = "Performance" },
    { message = "^refactor", group = "Refactoring" },
    { message = "^doc", group = "Documentation" },
    { message = "^test", group = "Testing" },
  ]

  EJECUTAR:
    git-cliff -o CHANGELOG.md           # Generar completo
    git-cliff --latest -o CHANGELOG.md  # Solo ultimo release
    git-cliff --unreleased              # Cambios sin release

CONVENTIONAL COMMITS (base para changelog automatico):
  feat: add user registration endpoint
  fix: resolve race condition in cache
  perf: optimize database query for orders
  refactor: extract validation logic to service
  docs: update API documentation
  test: add integration tests for auth
  chore: update dependencies
  ci: add benchmark workflow

  feat!: redesign API response format  (breaking change)
  feat(auth): add OAuth2 provider       (con scope)`)

	// ============================================
	// AIR - HOT RELOAD
	// ============================================
	fmt.Println("\n--- Air: Hot Reload for Development ---")
	fmt.Println(`
AIR - HOT RELOAD PARA DESARROLLO:

Recompila y reinicia automaticamente cuando cambias archivos Go.
Esencial para desarrollo web/API.

INSTALAR:
  go install github.com/air-verse/air@latest

INICIALIZAR:
  air init    # Crea .air.toml con config por defecto

ARCHIVO .air.toml:

root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  bin = "./tmp/main"
  full_bin = "./tmp/main --config=./config.dev.yaml"
  include_ext = ["go", "tpl", "tmpl", "html", "yaml"]
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "node_modules"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  delay = 500  # ms antes de rebuild
  stop_on_error = true
  send_interrupt = true
  kill_delay = 500  # ms para graceful shutdown

[log]
  time = false
  main_only = false

[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"

[misc]
  clean_on_exit = true

USAR:
  air                  # Usa .air.toml del directorio actual
  air -c .air.toml     # Config explicita
  air -- --port=3000   # Pasar args a tu binario

EN MAKEFILE:
  run-dev:
  	air

EN DOCKER (desarrollo):
  FROM golang:1.23-alpine
  RUN go install github.com/air-verse/air@latest
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  CMD ["air"]

TIPS:
  - Agrega tmp/ a .gitignore
  - Usa delay para evitar rebuilds excesivos
  - send_interrupt=true permite graceful shutdown
  - Excluye _test.go para evitar rebuilds en tests`)

	// ============================================
	// TASK (TASKFILE.DEV)
	// ============================================
	fmt.Println("\n--- Task (taskfile.dev): Makefile Alternative ---")
	fmt.Println(`
TASK - ALTERNATIVA MODERNA A MAKE:

Escrito en Go, usa YAML, cross-platform nativo.
No tiene las peculiaridades de Make (tabs, etc).

INSTALAR:
  go install github.com/go-task/task/v3/cmd/task@latest
  # o
  brew install go-task

ARCHIVO Taskfile.yml:

version: '3'

vars:
  BINARY_NAME: myapp
  VERSION:
    sh: git describe --tags --always --dirty
  COMMIT:
    sh: git rev-parse --short HEAD

env:
  CGO_ENABLED: '0'

tasks:
  default:
    cmds:
      - task: lint
      - task: test
      - task: build

  build:
    desc: Build the application
    sources:
      - ./**/*.go
      - go.mod
      - go.sum
    generates:
      - bin/{{.BINARY_NAME}}
    cmds:
      - go build -ldflags="-s -w -X main.version={{.VERSION}}"
        -o bin/{{.BINARY_NAME}} ./cmd/server

  test:
    desc: Run tests
    cmds:
      - go test -race -count=1 ./...

  test:coverage:
    desc: Run tests with coverage
    cmds:
      - go test -race -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out -o coverage.html

  lint:
    desc: Run linter
    cmds:
      - golangci-lint run ./...

  fmt:
    desc: Format code
    cmds:
      - gofmt -s -w .
      - goimports -w .

  vet:
    desc: Run go vet
    cmds:
      - go vet ./...

  run:
    desc: Run the application
    deps: [build]
    cmds:
      - ./bin/{{.BINARY_NAME}}

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -rf bin/ coverage.out coverage.html
      - go clean -cache -testcache

  docker:build:
    desc: Build Docker image
    cmds:
      - docker build -t {{.BINARY_NAME}}:{{.VERSION}} .

  tools:install:
    desc: Install development tools
    cmds:
      - go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
      - go install golang.org/x/tools/cmd/goimports@latest
      - go install golang.org/x/vuln/cmd/govulncheck@latest

  generate:
    desc: Run code generation
    sources:
      - ./**/*.go
    cmds:
      - go generate ./...

  ci:
    desc: Run full CI pipeline locally
    cmds:
      - task: fmt
      - task: vet
      - task: lint
      - task: test:coverage
      - task: build

EJECUTAR:
  task              # Ejecuta default task
  task build        # Ejecuta task especifico
  task --list       # Lista tasks disponibles
  task ci           # Ejecuta pipeline CI local

VENTAJAS SOBRE MAKE:
  - YAML en vez de Makefile syntax
  - Cross-platform (Windows, Linux, macOS)
  - Sources/generates: solo re-ejecuta si hay cambios
  - Deps: dependencias entre tasks
  - Variables con sh: para comandos dinamicos
  - Namespacing con : (task test:coverage)
  - No necesita tabs (comun fuente de errores en Make)`)

	// ============================================
	// COMPLETE CI PIPELINE EXAMPLE
	// ============================================
	fmt.Println("\n--- Complete CI Pipeline Example ---")
	os.Stdout.WriteString(`
PIPELINE CI COMPLETO - EJEMPLO REAL:

.github/workflows/ci.yml:

name: CI

on:
  push:
    branches: [main, develop]
  pull_request:

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true

permissions:
  contents: read
  pull-requests: write

env:
  GO_VERSION: '1.23'
  GOLANGCI_LINT_VERSION: 'v1.61'

jobs:
  # ===== Lint =====
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - uses: golangci/golangci-lint-action@v6
        with:
          version: ${{ env.GOLANGCI_LINT_VERSION }}

  # ===== Test =====
  test:
    runs-on: ubuntu-latest
    strategy:
      fail-fast: false
      matrix:
        go-version: ['1.22', '1.23']
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - run: go mod download
      - run: go mod verify
      - name: Unit tests
        run: go test -race -coverprofile=coverage.out -count=1 ./...
      - name: Integration tests
        run: go test -race -tags=integration -count=1 -timeout=5m ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379
      - name: Upload coverage
        if: matrix.go-version == '1.23'
        uses: codecov/codecov-action@v4

  # ===== Security =====
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
      - name: gosec
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out results.sarif ./...'
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif

  # ===== Build =====
  build:
    runs-on: ubuntu-latest
    needs: [lint, test, security]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      - name: Build
        run: |
          CGO_ENABLED=0 go build \
            -ldflags="-s -w -X main.version=${{ github.sha }}" \
            -o bin/server ./cmd/server
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: server
          path: bin/server

TIPS PARA CI EN GO:
- Usa go mod download separado para cachear dependencias
- go mod verify detecta dependencias corrompidas
- -count=1 evita cache de tests en CI
- -race detecta race conditions (solo en tests, no en prod builds)
- concurrency group cancela workflows previos del mismo PR
- fail-fast: false para que un Go version no cancele otros
- Services de GitHub Actions para bases de datos en integration tests
`)

	// ============================================
	// DOCKERIGNORE
	// ============================================
	fmt.Println("\n--- .dockerignore Best Practices ---")
	fmt.Println(`
ARCHIVO .dockerignore:

# Version control
.git
.gitignore

# IDE
.vscode
.idea
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Go
bin/
tmp/
vendor/  # si usas go mod download en Docker

# Test & coverage
*_test.go
coverage.out
coverage.html
*.prof
testdata/

# CI/CD
.github/
.gitlab-ci.yml
Makefile
Taskfile.yml
.goreleaser.yaml

# Docs
*.md
LICENSE
docs/

# Config local
.env
.env.local
*.local.yaml
docker-compose.override.yml

# Build artifacts
*.exe
*.dll
*.so
*.dylib

NOTA: Sin .dockerignore, COPY . . incluye TODO,
haciendo builds mas lentos e imagenes mas grandes.
Cada archivo copiado invalida la cache de Docker.`)

	// ============================================
	// SUMMARY: RECOMMENDED PROJECT SETUP
	// ============================================
	fmt.Println("\n--- Summary: Recommended Project Setup ---")
	fmt.Println(`
SETUP RECOMENDADO PARA UN PROYECTO GO PROFESIONAL:

Archivos en la raiz del proyecto:

  myproject/
  ├── .github/
  │   ├── workflows/
  │   │   ├── ci.yml              # Tests, lint, security
  │   │   └── release.yml         # GoReleaser on tag push
  │   └── dependabot.yml          # Dependency updates
  ├── cmd/
  │   └── server/
  │       └── main.go
  ├── internal/
  │   └── ...
  ├── .air.toml                   # Hot reload config
  ├── .dockerignore               # Docker build exclusions
  ├── .gitignore                  # Git exclusions
  ├── .golangci.yml               # Linter config
  ├── .goreleaser.yaml            # Release config
  ├── .pre-commit-config.yaml     # Pre-commit hooks (o lefthook.yml)
  ├── Dockerfile                  # Multi-stage build
  ├── docker-compose.yml          # Local dev environment
  ├── go.mod
  ├── go.sum
  ├── Makefile                    # (o Taskfile.yml)
  └── README.md

ORDEN DE IMPLEMENTACION SUGERIDO:
  1. go.mod + estructura de directorios
  2. Makefile/Taskfile con targets basicos
  3. .golangci.yml configurado
  4. Dockerfile multi-stage
  5. .dockerignore
  6. CI workflow (GitHub Actions)
  7. Pre-commit hooks
  8. .goreleaser.yaml (cuando necesites releases)
  9. dependabot.yml
  10. .air.toml (para desarrollo local)

FLUJO DE TRABAJO DIARIO:
  make run-dev     # Desarrollo con hot reload
  make test        # Tests locales
  make lint        # Verificar linting
  make ci          # Pipeline CI completo local
  git push         # CI remoto automatico

FLUJO DE RELEASE:
  git tag -a v1.0.0 -m "Release v1.0.0"
  git push origin v1.0.0
  # -> GitHub Actions ejecuta GoReleaser
  # -> Binarios, Docker images, changelog automaticos`)
}

// Capitulo 76: CI/CD, Makefile & Go Tooling
// - Makefile patterns: build, test, lint, fmt, vet, run, clean, docker, release
// - GitHub Actions: matrix testing, caching, release workflows
// - golangci-lint: configuracion completa (.golangci.yml)
// - Pre-commit hooks: scripts, pre-commit framework, lefthook
// - govulncheck: escaneo de vulnerabilidades oficial
// - GoReleaser: cross-compilation, packaging, Docker, Homebrew
// - go tool cover: coverage reports y umbrales en CI
// - go tool pprof: benchmarks y regresiones en CI
// - Dependabot/Renovate: actualizacion automatica de dependencias
// - Docker: multi-stage builds, scratch, distroless, .dockerignore
// - Semantic versioning: tags, Go module major versions
// - Changelog: GoReleaser, git-cliff, conventional commits
// - Air: hot reload para desarrollo
// - Task (taskfile.dev): alternativa moderna a Make
// - Pipeline CI completo de ejemplo con services
