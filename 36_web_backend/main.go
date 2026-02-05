// Package main - Capítulo 36: Web Backend
// Go es excelente para backend web. Aprenderás net/http, routing,
// middleware, JSON, y frameworks populares como chi y fiber.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=== WEB BACKEND EN GO ===")

	// ============================================
	// NET/HTTP BÁSICO
	// ============================================
	fmt.Println("\n--- net/http Básico ---")
	fmt.Println(`
Go tiene un servidor HTTP de producción en la stdlib:

http.HandleFunc("/", handler)
http.ListenAndServe(":8080", nil)

Características:
- Concurrente por defecto (goroutine por request)
- HTTP/1.1 y HTTP/2
- TLS integrado
- Sin dependencias externas`)
	// Crear servidor con ejemplo completo
	mux := http.NewServeMux()

	// Handler básico
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/json", jsonHandler)

	// CRUD de usuarios (ejemplo)
	store := NewUserStore()
	userHandler := &UserHandler{store: store}
	mux.HandleFunc("/users", userHandler.HandleUsers)
	mux.HandleFunc("/users/", userHandler.HandleUserByID)

	// Middleware chain
	handler := LoggingMiddleware(
		RecoverMiddleware(
			CORSMiddleware(mux),
		),
	)

	// Servidor con configuración
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("\nServidor configurado en :8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /          - Home")
	fmt.Println("  GET  /health    - Health check")
	fmt.Println("  GET  /json      - Ejemplo JSON")
	fmt.Println("  GET  /users     - Listar usuarios")
	fmt.Println("  POST /users     - Crear usuario")
	fmt.Println("  GET  /users/:id - Obtener usuario")

	// ============================================
	// GRACEFUL SHUTDOWN
	// ============================================
	fmt.Println("\n--- Graceful Shutdown ---")
	fmt.Println(`
En producción, siempre implementar graceful shutdown:

1. Escuchar señales del sistema (SIGINT, SIGTERM)
2. Dejar de aceptar conexiones nuevas
3. Esperar que terminen las conexiones activas
4. Cerrar recursos (DB, caches, etc.)`)
	// Demostrar el patrón (sin ejecutar el servidor)
	go func() {
		// En producción, esto ejecutaría el servidor
		// if err := server.ListenAndServe(); err != http.ErrServerClosed {
		//     log.Fatalf("Server error: %v", err)
		// }
		_ = server // evitar unused
	}()

	// Signal handler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Simular shutdown sin esperar señal real
	go func() {
		time.Sleep(100 * time.Millisecond)
		quit <- syscall.SIGINT
	}()

	<-quit
	fmt.Println("\nRecibida señal de shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error en shutdown: %v", err)
	}
	fmt.Println("Servidor cerrado correctamente")

	// ============================================
	// FRAMEWORKS POPULARES
	// ============================================
	fmt.Println("\n--- Frameworks Populares ---")
	fmt.Println(`
CHI (github.com/go-chi/chi):
- Router compatible con net/http
- Middleware composable
- URL parameters
- Muy idiomático Go

r := chi.NewRouter()
r.Use(middleware.Logger)
r.Get("/users/{id}", getUser)
r.Post("/users", createUser)

---

FIBER (github.com/gofiber/fiber):
- Inspirado en Express.js
- Muy rápido (usa fasthttp)
- Sintaxis familiar para JS devs

app := fiber.New()
app.Get("/users/:id", getUser)
app.Post("/users", createUser)

---

GIN (github.com/gin-gonic/gin):
- Popular, muchos middlewares
- Validación integrada
- Más opinado

r := gin.Default()
r.GET("/users/:id", getUser)
r.POST("/users", createUser)

---

ECHO (github.com/labstack/echo):
- Alto rendimiento
- Extensible
- Buen balance

e := echo.New()
e.GET("/users/:id", getUser)
e.POST("/users", createUser)`)
	// ============================================
	// EJEMPLO CHI-LIKE ROUTER
	// ============================================
	fmt.Println("\n--- Ejemplo: Mini Router ---")

	router := NewRouter()
	router.Use(LoggingMiddleware)

	router.Get("/api/items", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"items": "list"})
	})

	router.Post("/api/items", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusCreated, map[string]string{"status": "created"})
	})

	fmt.Println("Mini router configurado")

	// ============================================
	// VALIDACIÓN Y BINDING
	// ============================================
	fmt.Println("\n--- Validación y Binding ---")
	fmt.Println(`
Paquetes populares para validación:

go-playground/validator:
  type User struct {
      Name  string 'validate:"required,min=3"'
      Email string 'validate:"required,email"'
      Age   int    'validate:"gte=0,lte=130"'
  }

  validate := validator.New()
  err := validate.Struct(user)

---

JSON binding típico:
  var user User
  if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
      http.Error(w, "Invalid JSON", http.StatusBadRequest)
      return
  }`)
	// ============================================
	// TESTING DE HANDLERS
	// ============================================
	fmt.Println("\n--- Testing de Handlers ---")
	os.Stdout.WriteString(`
Usar httptest para tests:

func TestHealthHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    healthHandler(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("got %d, want %d", w.Code, http.StatusOK)
    }
}
`)

	// ============================================
	// BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Best Practices ---")
	fmt.Println(`
1. SIEMPRE usar timeouts en http.Server
2. Implementar graceful shutdown
3. Usar context para propagación
4. Middleware para cross-cutting concerns
5. Separar handlers de business logic
6. Validar todo input
7. Usar structured logging (slog)
8. Métricas y health checks
9. Rate limiting
10. HTTPS en producción`)}

// ============================================
// HANDLERS
// ============================================

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Welcome to Go Web Server!")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Message string    `json:"message"`
		Time    time.Time `json:"time"`
		Version string    `json:"version"`
	}{
		Message: "Hello, JSON!",
		Time:    time.Now(),
		Version: "1.0.0",
	}
	JSON(w, http.StatusOK, data)
}

// ============================================
// USER CRUD HANDLER
// ============================================

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserStore struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (s *UserStore) Create(user *User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user.ID = s.nextID
	s.nextID++
	s.users[user.ID] = user
}

func (s *UserStore) Get(id int) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	return user, ok
}

func (s *UserStore) List() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

type UserHandler struct {
	store *UserStore
}

func (h *UserHandler) HandleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listUsers(w, r)
	case http.MethodPost:
		h.createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	users := h.store.List()
	JSON(w, http.StatusOK, users)
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if user.Name == "" || user.Email == "" {
		JSONError(w, "Name and email required", http.StatusBadRequest)
		return
	}

	h.store.Create(&user)
	JSON(w, http.StatusCreated, user)
}

func (h *UserHandler) HandleUserByID(w http.ResponseWriter, r *http.Request) {
	// Extraer ID del path /users/123
	path := r.URL.Path
	idStr := path[len("/users/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		JSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, ok := h.store.Get(id)
		if !ok {
			JSONError(w, "User not found", http.StatusNotFound)
			return
		}
		JSON(w, http.StatusOK, user)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ============================================
// MIDDLEWARE
// ============================================

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================
// HELPERS
// ============================================

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func JSONError(w http.ResponseWriter, message string, status int) {
	JSON(w, status, map[string]string{"error": message})
}

func ReadJSON(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// ============================================
// MINI ROUTER
// ============================================

type Router struct {
	routes     map[string]map[string]http.HandlerFunc
	middleware []func(http.Handler) http.Handler
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

func (r *Router) Use(mw func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, mw)
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.addRoute(http.MethodGet, path, handler)
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.addRoute(http.MethodPost, path, handler)
}

func (r *Router) addRoute(method, path string, handler http.HandlerFunc) {
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]http.HandlerFunc)
	}
	r.routes[method][path] = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handlers, ok := r.routes[req.Method]; ok {
		if handler, ok := handlers[req.URL.Path]; ok {
			handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

/*
RESUMEN:

NET/HTTP:
- http.HandleFunc/Handle para registrar handlers
- http.ListenAndServe para iniciar servidor
- http.Server para configuración avanzada
- Siempre usar timeouts

HANDLERS:
func handler(w http.ResponseWriter, r *http.Request) {
    // w para escribir respuesta
    // r para leer request
}

MIDDLEWARE:
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w, r) {
        // before
        next.ServeHTTP(w, r)
        // after
    })
}

GRACEFUL SHUTDOWN:
signal.Notify(quit, SIGINT, SIGTERM)
<-quit
server.Shutdown(ctx)

FRAMEWORKS:
- chi: stdlib compatible, idiomático
- fiber: muy rápido, API Express-like
- gin: popular, muchos middlewares
- echo: buen balance

BUENAS PRÁCTICAS:
1. Timeouts siempre
2. Graceful shutdown
3. Context propagation
4. Input validation
5. Structured logging
6. Health checks
7. HTTPS
*/
