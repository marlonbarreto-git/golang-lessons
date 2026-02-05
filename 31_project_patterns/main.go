// Package main - Capítulo 31: Patrones de Proyecto
// Aprenderás estructuras y patrones comunes para proyectos Go
// en producción, desde microservicios hasta CLI tools.
package main

import (
	"os"
	"fmt"
)

func main() {
	fmt.Println("=== PATRONES DE PROYECTO ===")

	fmt.Println(`
ESTRUCTURAS DE PROYECTO COMUNES EN GO

---

1. FLAT STRUCTURE (Proyectos Pequeños):

myapp/
├── go.mod
├── main.go
├── handler.go
├── model.go
├── db.go
└── README.md

Ideal para: scripts, herramientas simples, prototipos.

---

2. LAYERED STRUCTURE (Proyectos Medianos):

myapi/
├── go.mod
├── main.go
├── config/
│   └── config.go
├── handler/
│   ├── user.go
│   └── product.go
├── service/
│   ├── user.go
│   └── product.go
├── repository/
│   ├── user.go
│   └── product.go
├── model/
│   ├── user.go
│   └── product.go
└── middleware/
    ├── auth.go
    └── logging.go

Ideal para: APIs REST, servicios web medianos.

---

3. STANDARD PROJECT LAYOUT (Proyectos Grandes):

myproject/
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── worker/
│       └── main.go
├── internal/
│   ├── config/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   └── model/
├── pkg/
│   └── client/
├── api/
│   ├── openapi.yaml
│   └── proto/
├── web/
│   └── static/
├── scripts/
│   └── setup.sh
├── deployments/
│   ├── docker/
│   └── kubernetes/
├── docs/
├── test/
│   └── integration/
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md

Directorios:
- cmd/: Entry points (package main)
- internal/: Código privado del proyecto
- pkg/: Código público reutilizable
- api/: OpenAPI, Protocol Buffers, GraphQL
- web/: Assets web, templates
- scripts/: Scripts de build, setup
- deployments/: Configs de deployment
- docs/: Documentación
- test/: Tests de integración, e2e

---

4. DOMAIN-DRIVEN DESIGN (DDD):

myapp/
├── cmd/
│   └── api/
├── internal/
│   ├── domain/
│   │   ├── user/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   └── order/
│   │       ├── entity.go
│   │       ├── repository.go
│   │       └── service.go
│   ├── application/
│   │   ├── user/
│   │   │   └── commands.go
│   │   └── order/
│   │       └── commands.go
│   ├── infrastructure/
│   │   ├── postgres/
│   │   │   ├── user_repo.go
│   │   │   └── order_repo.go
│   │   └── http/
│   │       └── handlers.go
│   └── interfaces/
│       └── api/
└── go.mod

Ideal para: sistemas complejos con reglas de negocio elaboradas.

---

5. CLEAN ARCHITECTURE:

myapp/
├── cmd/
│   └── api/
├── internal/
│   ├── entity/           # Entidades de negocio
│   │   └── user.go
│   ├── usecase/          # Casos de uso
│   │   └── user/
│   │       ├── interface.go
│   │       └── interactor.go
│   ├── interface/        # Adaptadores
│   │   ├── controller/
│   │   │   └── user.go
│   │   └── repository/
│   │       └── user.go
│   └── infrastructure/   # Frameworks/drivers
│       ├── database/
│       └── web/
└── go.mod

Regla de dependencias: de afuera hacia adentro.

---

6. HEXAGONAL ARCHITECTURE:

myapp/
├── cmd/
│   └── api/
├── internal/
│   ├── core/
│   │   ├── domain/       # Entidades
│   │   ├── ports/        # Interfaces
│   │   └── services/     # Lógica de negocio
│   └── adapters/
│       ├── driven/       # Secundarios (DB, APIs externas)
│       │   └── postgres/
│       └── driving/      # Primarios (HTTP, gRPC)
│           └── http/
└── go.mod

---

PATRONES COMUNES:

1. DEPENDENCY INJECTION:

// En vez de crear dependencias dentro:
type Service struct {
    repo *PostgresRepository  // MAL
}

// Inyectar interfaces:
type Service struct {
    repo Repository  // BIEN
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

2. REPOSITORY PATTERN:

type UserRepository interface {
    Find(id int) (*User, error)
    Save(user *User) error
    Delete(id int) error
}

// Implementación separada de interfaz
type PostgresUserRepo struct {
    db *sql.DB
}

func (r *PostgresUserRepo) Find(id int) (*User, error) {
    // ...
}

3. SERVICE LAYER:

type UserService struct {
    repo   UserRepository
    cache  Cache
    events EventPublisher
}

func (s *UserService) CreateUser(cmd CreateUserCmd) (*User, error) {
    // Validación
    // Lógica de negocio
    // Persistencia
    // Eventos
}

4. OPTIONS PATTERN:

type ServerOption func(*Server)

func WithPort(port int) ServerOption {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) {
        s.timeout = d
    }
}

func NewServer(opts ...ServerOption) *Server {
    s := &Server{port: 8080}  // defaults
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Uso
srv := NewServer(
    WithPort(9000),
    WithTimeout(30*time.Second),
)

5. FACTORY PATTERN:

type Database interface {
    Query(q string) ([]Row, error)
}

func NewDatabase(driver string) (Database, error) {
    switch driver {
    case "postgres":
        return NewPostgres()
    case "mysql":
        return NewMySQL()
    default:
        return nil, errors.New("unknown driver")
    }
}

6. SINGLETON (cuando necesario):

var (
    instance *Config
    once     sync.Once
)

func GetConfig() *Config {
    once.Do(func() {
        instance = loadConfig()
    })
    return instance
}

---

CONFIGURACIÓN:

// config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Port    int           ` + "`" + `env:"PORT" envDefault:"8080"` + "`" + `
    Timeout time.Duration ` + "`" + `env:"TIMEOUT" envDefault:"30s"` + "`" + `
}

// Cargar desde env
func Load() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}

---

ERROR HANDLING:

// errors/errors.go
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrValidation    = errors.New("validation failed")
)

type AppError struct {
    Code    string
    Message string
    Err     error
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

---

LOGGING:

// Usar slog (Go 1.21+)
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("request processed",
    "method", "GET",
    "path", "/users",
    "duration", time.Since(start),
)`)
	fmt.Println("\n--- Ejemplo Completo: API REST ---")

	os.Stdout.WriteString(`
// cmd/api/main.go
package main

func main() {
    cfg := config.MustLoad()
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    db := database.MustConnect(cfg.Database)
    defer db.Close()

    repo := postgres.NewUserRepository(db)
    svc := service.NewUserService(repo)
    handler := http.NewUserHandler(svc)

    router := chi.NewRouter()
    router.Use(middleware.Logger)
    router.Use(middleware.Recoverer)

    router.Route("/api/v1", func(r chi.Router) {
        r.Get("/users", handler.List)
        r.Post("/users", handler.Create)
        r.Get("/users/{id}", handler.Get)
    })

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
        Handler: router,
    }

    go func() {
        logger.Info("starting server", "port", cfg.Server.Port)
        if err := srv.ListenAndServe(); err != nil {
            logger.Error("server error", "error", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}
`)
}

/*
RESUMEN:

ESTRUCTURAS:
1. Flat: proyectos simples
2. Layered: APIs medianas
3. Standard Layout: proyectos grandes
4. DDD: dominio complejo
5. Clean/Hexagonal: máximo desacoplamiento

DIRECTORIOS CLAVE:
- cmd/: entry points
- internal/: código privado
- pkg/: código público
- api/: especificaciones API

PATRONES:
- Dependency Injection
- Repository
- Service Layer
- Options Pattern
- Factory

BUENAS PRÁCTICAS:
1. Interfaces en el consumidor
2. Inyectar dependencias
3. Separar config de código
4. Structured logging (slog)
5. Graceful shutdown
6. Health checks
*/
