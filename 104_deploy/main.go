// Package main - Chapter 104: Deploy
// Aprenderás a empaquetar, distribuir y desplegar
// aplicaciones Go en producción.
package main

import (
	"os"
	"fmt"
)

func main() {
	fmt.Println("=== DEPLOYMENT EN GO ===")

	// ============================================
	// BUILD
	// ============================================
	fmt.Println("\n--- Build ---")
	fmt.Println(`
BUILD BÁSICO:
go build -o myapp main.go
go build -o myapp ./cmd/server

BUILD FLAGS:
# Optimizar tamaño
go build -ldflags="-s -w" -o myapp

# Inyectar versión
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)" -o myapp

# Static linking (Linux)
CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags "-static"' -o myapp

CROSS-COMPILATION:
GOOS=linux GOARCH=amd64 go build -o myapp-linux
GOOS=darwin GOARCH=arm64 go build -o myapp-darwin
GOOS=windows GOARCH=amd64 go build -o myapp.exe

# Lista de targets
go tool dist list

GOAMD64 (optimización CPU):
GOAMD64=v1  # Compatible con todos los x86-64
GOAMD64=v2  # SSE3, SSSE3, SSE4.1, SSE4.2
GOAMD64=v3  # AVX, AVX2, BMI1, BMI2, FMA (recomendado moderno)
GOAMD64=v4  # AVX-512`)
	// ============================================
	// DOCKER
	// ============================================
	fmt.Println("\n--- Docker ---")
	fmt.Println(`
MULTI-STAGE DOCKERFILE:
# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/server .

# Non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080
CMD ["./server"]

DISTROLESS (más seguro):
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/server /
CMD ["/server"]

SCRATCH (mínimo):
FROM scratch
COPY --from=builder /app/server /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/server"]

BUILD & RUN:
docker build -t myapp:latest .
docker run -p 8080:8080 myapp:latest

DOCKER COMPOSE:
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://db:5432/mydb
    depends_on:
      - db
  db:
    image: postgres:15
    volumes:
      - pgdata:/var/lib/postgresql/data`)
	// ============================================
	// KUBERNETES
	// ============================================
	fmt.Println("\n--- Kubernetes ---")
	fmt.Println(`
DEPLOYMENT:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myregistry/myapp:v1.0.0
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: myapp-secrets
              key: database-url
        - name: GOMAXPROCS
          valueFrom:
            resourceFieldRef:
              resource: limits.cpu

SERVICE:
apiVersion: v1
kind: Service
metadata:
  name: myapp
spec:
  selector:
    app: myapp
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP

CONFIGMAP:
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  config.yaml: |
    server:
      port: 8080
    log:
      level: info

HPA (Autoscaling):
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: myapp
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70`)
	// ============================================
	// SYSTEMD
	// ============================================
	fmt.Println("\n--- Systemd (Linux) ---")
	fmt.Println(`
/etc/systemd/system/myapp.service:

[Unit]
Description=My Go Application
After=network.target

[Service]
Type=simple
User=myapp
Group=myapp
WorkingDirectory=/opt/myapp
ExecStart=/opt/myapp/server
Restart=always
RestartSec=5

# Environment
Environment=PORT=8080
Environment=LOG_LEVEL=info
EnvironmentFile=/etc/myapp/env

# Resource limits
LimitNOFILE=65536
MemoryMax=512M
CPUQuota=200%

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/myapp

[Install]
WantedBy=multi-user.target

COMANDOS:
sudo systemctl daemon-reload
sudo systemctl enable myapp
sudo systemctl start myapp
sudo systemctl status myapp
sudo journalctl -u myapp -f`)
	// ============================================
	// GORELEASER
	// ============================================
	fmt.Println("\n--- GoReleaser ---")
	fmt.Println(`
INSTALACIÓN:
go install github.com/goreleaser/goreleaser@latest

.goreleaser.yaml:
project_name: myapp

builds:
  - id: server
    main: ./cmd/server
    binary: myapp
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}

archives:
  - id: default
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

dockers:
  - image_templates:
      - "myregistry/{{ .ProjectName }}:{{ .Version }}"
      - "myregistry/{{ .ProjectName }}:latest"
    dockerfile: Dockerfile
    build_flag_templates:
      - "--platform=linux/amd64"

nfpms:
  - formats:
      - deb
      - rpm
    maintainer: You <you@example.com>
    description: My application
    license: MIT

brews:
  - tap:
      owner: myorg
      name: homebrew-tap
    homepage: https://github.com/myorg/myapp
    description: My application

COMANDOS:
goreleaser check      # Validar config
goreleaser build      # Solo build local
goreleaser release    # Release completo
goreleaser release --snapshot  # Sin publicar`)
	// ============================================
	// CI/CD
	// ============================================
	fmt.Println("\n--- CI/CD ---")
	fmt.Println(`
GITHUB ACTIONS:
.github/workflows/release.yaml:

name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Run tests
        run: go test -v ./...

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

TEST WORKFLOW:
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'

      - name: Lint
        uses: golangci/golangci-lint-action@v4

      - name: Test
        run: go test -race -coverprofile=coverage.txt ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3`)
	// ============================================
	// GRACEFUL SHUTDOWN
	// ============================================
	fmt.Println("\n--- Graceful Shutdown ---")
	os.Stdout.WriteString(`
func main() {
    srv := &http.Server{
        Addr:    ":8080",
        Handler: handler,
    }

    // Servidor en goroutine
    go func() {
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Esperar señal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down...")

    // Timeout para shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }

    // Cleanup adicional
    db.Close()
    cache.Close()

    log.Println("Server stopped")
}
`)

	// ============================================
	// CONFIGURACIÓN
	// ============================================
	fmt.Println("\n--- Configuración ---")
	fmt.Println(`
12-FACTOR APP:
1. Codebase: Un repo, muchos deploys
2. Dependencies: Declarar explícitamente
3. Config: Variables de entorno
4. Backing services: Tratar como recursos
5. Build/Release/Run: Separar etapas
6. Processes: Stateless
7. Port binding: Exportar vía puerto
8. Concurrency: Escalar vía procesos
9. Disposability: Startup/shutdown rápido
10. Dev/Prod parity: Mantener similar
11. Logs: Tratar como streams
12. Admin: Tareas como procesos one-off

GESTIÓN DE CONFIG:
type Config struct {
    Port        int           ` + "`" + `env:"PORT" envDefault:"8080"` + "`" + `
    DatabaseURL string        ` + "`" + `env:"DATABASE_URL,required"` + "`" + `
    LogLevel    string        ` + "`" + `env:"LOG_LEVEL" envDefault:"info"` + "`" + `
    Timeout     time.Duration ` + "`" + `env:"TIMEOUT" envDefault:"30s"` + "`" + `
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}`)}

/*
RESUMEN DE DEPLOYMENT:

BUILD:
go build -ldflags="-s -w" -o app
CGO_ENABLED=0 para static
GOOS/GOARCH para cross-compile

DOCKER:
- Multi-stage builds
- Alpine/Distroless/Scratch
- Non-root user

KUBERNETES:
- Deployment + Service
- Liveness/Readiness probes
- Resource limits
- ConfigMaps/Secrets
- HPA para autoscaling

SYSTEMD:
- Service file
- Restart policies
- Resource limits
- Security hardening

GORELEASER:
- Builds multiplataforma
- Docker images
- Homebrew, deb, rpm

CI/CD:
- GitHub Actions
- Tests + Lint
- Automated releases

GRACEFUL SHUTDOWN:
- signal.Notify
- srv.Shutdown(ctx)
- Cleanup resources

12-FACTOR:
- Config en env vars
- Stateless processes
- Logs como streams
*/
