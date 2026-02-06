// Package main - Chapter 106: Final Project
// Este capítulo integra todo lo aprendido en un proyecto completo:
// una API REST con base de datos, tests, observabilidad y deployment.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== PROYECTO FINAL: API REST COMPLETA ===")

	// ============================================
	// ESTRUCTURA DEL PROYECTO
	// ============================================
	fmt.Println("\n--- Estructura del Proyecto ---")
	fmt.Println(`
taskapi/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuración
│   ├── domain/
│   │   └── task.go           # Entidades de dominio
│   ├── handler/
│   │   ├── handler.go        # HTTP handlers
│   │   └── middleware.go     # Middlewares
│   ├── repository/
│   │   ├── repository.go     # Interface
│   │   └── postgres.go       # Implementación PostgreSQL
│   └── service/
│       └── task.go           # Lógica de negocio
├── migrations/
│   ├── 000001_create_tasks.up.sql
│   └── 000001_create_tasks.down.sql
├── test/
│   └── integration/
│       └── api_test.go       # Tests de integración
├── .github/
│   └── workflows/
│       └── ci.yaml           # CI/CD
├── Dockerfile
├── docker-compose.yaml
├── Makefile
├── go.mod
└── README.md`)
	// ============================================
	// DOMINIO
	// ============================================
	fmt.Println("\n--- internal/domain/task.go ---")
	fmt.Println(`
package domain

import (
    "errors"
    "time"
)

var (
    ErrTaskNotFound = errors.New("task not found")
    ErrInvalidTask  = errors.New("invalid task")
)

type TaskStatus string

const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusCompleted  TaskStatus = "completed"
)

type Task struct {
    ID          int64      ` + "`" + `json:"id"` + "`" + `
    Title       string     ` + "`" + `json:"title"` + "`" + `
    Description string     ` + "`" + `json:"description,omitempty"` + "`" + `
    Status      TaskStatus ` + "`" + `json:"status"` + "`" + `
    CreatedAt   time.Time  ` + "`" + `json:"created_at"` + "`" + `
    UpdatedAt   time.Time  ` + "`" + `json:"updated_at"` + "`" + `
}

func (t *Task) Validate() error {
    if t.Title == "" {
        return ErrInvalidTask
    }
    if t.Status == "" {
        t.Status = TaskStatusPending
    }
    return nil
}

type CreateTaskRequest struct {
    Title       string ` + "`" + `json:"title"` + "`" + `
    Description string ` + "`" + `json:"description"` + "`" + `
}

type UpdateTaskRequest struct {
    Title       *string     ` + "`" + `json:"title"` + "`" + `
    Description *string     ` + "`" + `json:"description"` + "`" + `
    Status      *TaskStatus ` + "`" + `json:"status"` + "`" + `
}`)
	// ============================================
	// REPOSITORY
	// ============================================
	fmt.Println("\n--- internal/repository/repository.go ---")
	fmt.Println(`
package repository

import (
    "context"
    "taskapi/internal/domain"
)

type TaskRepository interface {
    Create(ctx context.Context, task *domain.Task) error
    GetByID(ctx context.Context, id int64) (*domain.Task, error)
    List(ctx context.Context, limit, offset int) ([]*domain.Task, error)
    Update(ctx context.Context, task *domain.Task) error
    Delete(ctx context.Context, id int64) error
}`)
	fmt.Println("\n--- internal/repository/postgres.go ---")
	fmt.Println(`
package repository

import (
    "context"
    "database/sql"
    "taskapi/internal/domain"
)

type PostgresTaskRepository struct {
    db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
    return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, task *domain.Task) error {
    query := ` + "`" + `
        INSERT INTO tasks (title, description, status, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW())
        RETURNING id, created_at, updated_at
    ` + "`" + `
    return r.db.QueryRowContext(ctx, query,
        task.Title, task.Description, task.Status,
    ).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

func (r *PostgresTaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
    query := ` + "`" + `
        SELECT id, title, description, status, created_at, updated_at
        FROM tasks WHERE id = $1
    ` + "`" + `
    task := &domain.Task{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &task.ID, &task.Title, &task.Description,
        &task.Status, &task.CreatedAt, &task.UpdatedAt,
    )
    if err == sql.ErrNoRows {
        return nil, domain.ErrTaskNotFound
    }
    return task, err
}

func (r *PostgresTaskRepository) List(ctx context.Context, limit, offset int) ([]*domain.Task, error) {
    query := ` + "`" + `
        SELECT id, title, description, status, created_at, updated_at
        FROM tasks ORDER BY created_at DESC LIMIT $1 OFFSET $2
    ` + "`" + `
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tasks []*domain.Task
    for rows.Next() {
        task := &domain.Task{}
        if err := rows.Scan(
            &task.ID, &task.Title, &task.Description,
            &task.Status, &task.CreatedAt, &task.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        tasks = append(tasks, task)
    }
    return tasks, rows.Err()
}`)
	// ============================================
	// SERVICE
	// ============================================
	fmt.Println("\n--- internal/service/task.go ---")
	fmt.Println(`
package service

import (
    "context"
    "taskapi/internal/domain"
    "taskapi/internal/repository"
)

type TaskService struct {
    repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
    return &TaskService{repo: repo}
}

func (s *TaskService) CreateTask(ctx context.Context, req domain.CreateTaskRequest) (*domain.Task, error) {
    task := &domain.Task{
        Title:       req.Title,
        Description: req.Description,
        Status:      domain.TaskStatusPending,
    }

    if err := task.Validate(); err != nil {
        return nil, err
    }

    if err := s.repo.Create(ctx, task); err != nil {
        return nil, err
    }

    return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context, page, pageSize int) ([]*domain.Task, error) {
    offset := (page - 1) * pageSize
    return s.repo.List(ctx, pageSize, offset)
}

func (s *TaskService) UpdateTask(ctx context.Context, id int64, req domain.UpdateTaskRequest) (*domain.Task, error) {
    task, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    if req.Title != nil {
        task.Title = *req.Title
    }
    if req.Description != nil {
        task.Description = *req.Description
    }
    if req.Status != nil {
        task.Status = *req.Status
    }

    if err := task.Validate(); err != nil {
        return nil, err
    }

    if err := s.repo.Update(ctx, task); err != nil {
        return nil, err
    }

    return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
    return s.repo.Delete(ctx, id)
}`)
	// ============================================
	// HANDLER
	// ============================================
	fmt.Println("\n--- internal/handler/handler.go ---")
	fmt.Println(`
package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "taskapi/internal/domain"
    "taskapi/internal/service"
)

type TaskHandler struct {
    svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
    return &TaskHandler{svc: svc}
}

func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /tasks", h.ListTasks)
    mux.HandleFunc("POST /tasks", h.CreateTask)
    mux.HandleFunc("GET /tasks/{id}", h.GetTask)
    mux.HandleFunc("PUT /tasks/{id}", h.UpdateTask)
    mux.HandleFunc("DELETE /tasks/{id}", h.DeleteTask)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    var req domain.CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    task, err := h.svc.CreateTask(r.Context(), req)
    if err != nil {
        if errors.Is(err, domain.ErrInvalidTask) {
            writeError(w, http.StatusBadRequest, err.Error())
            return
        }
        writeError(w, http.StatusInternalServerError, "internal error")
        return
    }

    writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid id")
        return
    }

    task, err := h.svc.GetTask(r.Context(), id)
    if err != nil {
        if errors.Is(err, domain.ErrTaskNotFound) {
            writeError(w, http.StatusNotFound, err.Error())
            return
        }
        writeError(w, http.StatusInternalServerError, "internal error")
        return
    }

    writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
    page, _ := strconv.Atoi(r.URL.Query().Get("page"))
    if page < 1 {
        page = 1
    }
    pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }

    tasks, err := h.svc.ListTasks(r.Context(), page, pageSize)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "internal error")
        return
    }

    writeJSON(w, http.StatusOK, map[string]any{
        "tasks": tasks,
        "page":  page,
    })
}

func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, map[string]string{"error": message})
}`)
	// ============================================
	// MAIN
	// ============================================
	fmt.Println("\n--- cmd/server/main.go ---")
	fmt.Println(`
package main

import (
    "context"
    "database/sql"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    _ "github.com/lib/pq"

    "taskapi/internal/config"
    "taskapi/internal/handler"
    "taskapi/internal/repository"
    "taskapi/internal/service"
)

func main() {
    // Logger
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    // Config
    cfg, err := config.Load()
    if err != nil {
        slog.Error("Failed to load config", "error", err)
        os.Exit(1)
    }

    // Database
    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        slog.Error("Failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    // Dependencies
    repo := repository.NewPostgresTaskRepository(db)
    svc := service.NewTaskService(repo)
    h := handler.NewTaskHandler(svc)

    // Router
    mux := http.NewServeMux()
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    h.RegisterRoutes(mux)

    // Middlewares
    handler := handler.LoggingMiddleware(
        handler.RecoverMiddleware(mux),
    )

    // Server
    srv := &http.Server{
        Addr:         ":" + cfg.Port,
        Handler:      handler,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    // Start
    go func() {
        slog.Info("Starting server", "port", cfg.Port)
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            slog.Error("Server error", "error", err)
            os.Exit(1)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    slog.Info("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("Shutdown error", "error", err)
    }

    slog.Info("Server stopped")
}`)
	// ============================================
	// DOCKER & MAKEFILE
	// ============================================
	fmt.Println("\n--- Makefile ---")
	fmt.Println(`
.PHONY: build test run docker-build docker-run migrate

build:
	go build -o bin/server ./cmd/server

test:
	go test -v -race ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run ./cmd/server

docker-build:
	docker build -t taskapi:latest .

docker-run:
	docker-compose up -d

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

lint:
	golangci-lint run

.DEFAULT_GOAL := build`)
	fmt.Println("\n--- docker-compose.yaml ---")
	fmt.Println(`
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATABASE_URL=postgres://user:pass@db:5432/taskapi?sslmode=disable
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:15
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=pass
      - POSTGRES_DB=taskapi
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user -d taskapi"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  pgdata:`)
	// ============================================
	// RESUMEN
	// ============================================
	fmt.Println("\n--- Lo Que Aprendiste ---")
	fmt.Println(`
Este curso cubrió:

FUNDAMENTOS:
✓ Variables, tipos, operadores
✓ Condicionales, bucles, switch
✓ Funciones, punteros, structs
✓ Interfaces, generics
✓ Manejo de errores

CONCURRENCIA:
✓ Goroutines, channels, select
✓ Context, sync package
✓ Patrones de concurrencia

TESTING:
✓ Unit tests, table-driven tests
✓ Benchmarks, fuzzing
✓ Mocking

AVANZADO:
✓ Reflection, unsafe, CGO
✓ Go Assembly

WEB & DATOS:
✓ net/http, frameworks
✓ Bases de datos, gRPC
✓ CLI, Desktop apps

PRODUCCIÓN:
✓ Performance, observabilidad
✓ Docker, Kubernetes
✓ CI/CD, deployment

¡Ahora eres un Ninja de Go! 🥷`)}

/*
PROYECTO FINAL:
- Clean Architecture
- Repository Pattern
- Dependency Injection
- Graceful Shutdown
- Structured Logging
- Health Checks
- Docker + Compose
- Makefile

ESTE CURSO CUBRIÓ:
- Go 1.24, 1.25, 1.26 features
- Desde Hola Mundo hasta Go Assembly
- Best practices y clean code
- Performance y observabilidad
- Deployment en producción

¡Felicidades por completar el curso!
*/
