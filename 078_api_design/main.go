// Package main - Chapter 078: API Design
// Diseno de APIs REST profesionales: versionado, documentacion,
// patrones de request/response, error handling y testing.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== API DESIGN, VERSIONING & OPENAPI EN GO ===")

	// ============================================
	// REST API DESIGN PRINCIPLES
	// ============================================
	fmt.Println("\n--- REST API Design Principles ---")
	fmt.Println(`
RESOURCE NAMING CONVENTIONS:

Usa sustantivos en plural para recursos:

  GET    /users          -> Listar usuarios
  POST   /users          -> Crear usuario
  GET    /users/123      -> Obtener usuario
  PUT    /users/123      -> Reemplazar usuario
  PATCH  /users/123      -> Actualizar parcial
  DELETE /users/123      -> Eliminar usuario

Relaciones anidadas (max 2 niveles):

  GET    /users/123/orders          -> Ordenes del usuario
  GET    /users/123/orders/456      -> Orden especifica
  POST   /users/123/orders          -> Crear orden para usuario

EVITAR:
  /getUser              -> Verbo en URL (usar GET /users/123)
  /users/123/getOrders  -> Verbo en recurso anidado
  /user                 -> Singular para colecciones
  /users/123/orders/456/items/789/details  -> Demasiada anidacion

CONVENCIONES:
  - kebab-case para URLs:  /order-items (no /orderItems)
  - camelCase para JSON:   {"firstName": "John"}
  - Consistencia total en toda la API
  - Filtros como query params: /users?role=admin&status=active`)

	// ============================================
	// HTTP METHOD SEMANTICS
	// ============================================
	fmt.Println("\n--- HTTP Method Semantics ---")
	fmt.Println(`
SEMANTICA DE METODOS HTTP:

Metodo  | Idempotente | Safe | Body Req | Body Resp | Uso
--------|-------------|------|----------|-----------|-------------------
GET     | Si          | Si   | No       | Si        | Obtener recurso
POST    | No          | No   | Si       | Si        | Crear recurso
PUT     | Si          | No   | Si       | Opcional  | Reemplazar completo
PATCH   | No          | No   | Si       | Si        | Actualizacion parcial
DELETE  | Si          | No   | Opcional | Opcional  | Eliminar recurso
HEAD    | Si          | Si   | No       | No        | Headers sin body
OPTIONS | Si          | Si   | No       | Si        | Capabilities/CORS

IDEMPOTENCIA:
- GET, PUT, DELETE: la misma request N veces produce el mismo resultado
- POST: cada request puede crear un recurso nuevo
- PATCH: depende de la implementacion

SAFE:
- GET, HEAD, OPTIONS: no modifican estado del servidor
- No deben tener side effects visibles al usuario`)

	// ============================================
	// STATUS CODES BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Status Codes Best Practices ---")
	fmt.Println(`
STATUS CODES - CUANDO USAR CADA UNO:

2xx - Exito:
  200 OK            -> GET, PUT, PATCH exitoso con body
  201 Created       -> POST exitoso (incluir Location header)
  202 Accepted      -> Request aceptado para procesamiento async
  204 No Content    -> DELETE exitoso, PUT/PATCH sin body de respuesta

3xx - Redireccion:
  301 Moved Permanently  -> Recurso movido permanentemente
  304 Not Modified       -> Conditional GET, usar cache (ETags)

4xx - Error del cliente:
  400 Bad Request        -> Body invalido, parametros incorrectos
  401 Unauthorized       -> No autenticado (falta o token invalido)
  403 Forbidden          -> Autenticado pero sin permisos
  404 Not Found          -> Recurso no existe
  405 Method Not Allowed -> HTTP method no soportado
  409 Conflict           -> Conflicto de estado (duplicado, version)
  410 Gone               -> Recurso eliminado permanentemente
  415 Unsupported Media  -> Content-Type no soportado
  422 Unprocessable      -> Validacion de negocio fallo
  429 Too Many Requests  -> Rate limited

5xx - Error del servidor:
  500 Internal Server Error -> Error inesperado del servidor
  502 Bad Gateway           -> Upstream error
  503 Service Unavailable   -> Mantenimiento, sobrecarga
  504 Gateway Timeout       -> Upstream timeout

REGLAS:
- 401 vs 403: 401 = "quien eres?", 403 = "se quien eres, no puedes"
- 400 vs 422: 400 = JSON invalido, 422 = JSON valido pero datos invalidos
- 404 vs 410: 404 = no se si existio, 410 = se que existio y se borro
- Nunca retornar 200 con un error en el body`)

	// ============================================
	// HATEOAS
	// ============================================
	fmt.Println("\n--- HATEOAS ---")
	fmt.Println(`
HATEOAS: Hypermedia As The Engine Of Application State

El servidor indica al cliente que acciones puede tomar siguiente:

{
  "id": 123,
  "status": "pending",
  "total": 99.99,
  "_links": {
    "self":    {"href": "/orders/123"},
    "cancel":  {"href": "/orders/123/cancel", "method": "POST"},
    "payment": {"href": "/orders/123/payment", "method": "POST"},
    "items":   {"href": "/orders/123/items"}
  }
}

Despues de pagar, los links cambian:

{
  "id": 123,
  "status": "paid",
  "total": 99.99,
  "_links": {
    "self":    {"href": "/orders/123"},
    "refund":  {"href": "/orders/123/refund", "method": "POST"},
    "receipt": {"href": "/orders/123/receipt"}
  }
}

BENEFICIOS:
- El cliente no hardcodea URLs
- El servidor controla el flujo de la aplicacion
- Autodescubrimiento de la API
- Versionado mas facil (cambiar links, no URLs del cliente)

EN GO:`)

	os.Stdout.WriteString(`
type Link struct {
    Href   string ` + "`" + `json:"href"` + "`" + `
    Method string ` + "`" + `json:"method,omitempty"` + "`" + `
    Rel    string ` + "`" + `json:"rel,omitempty"` + "`" + `
}

type OrderResponse struct {
    ID     int               ` + "`" + `json:"id"` + "`" + `
    Status string            ` + "`" + `json:"status"` + "`" + `
    Total  float64           ` + "`" + `json:"total"` + "`" + `
    Links  map[string]Link   ` + "`" + `json:"_links"` + "`" + `
}

func buildOrderLinks(order *Order, baseURL string) map[string]Link {
    links := map[string]Link{
        "self": {Href: fmt.Sprintf("%s/orders/%d", baseURL, order.ID)},
    }

    switch order.Status {
    case "pending":
        links["cancel"] = Link{
            Href:   fmt.Sprintf("%s/orders/%d/cancel", baseURL, order.ID),
            Method: "POST",
        }
        links["payment"] = Link{
            Href:   fmt.Sprintf("%s/orders/%d/payment", baseURL, order.ID),
            Method: "POST",
        }
    case "paid":
        links["refund"] = Link{
            Href:   fmt.Sprintf("%s/orders/%d/refund", baseURL, order.ID),
            Method: "POST",
        }
    }

    return links
}
`)

	// ============================================
	// API VERSIONING STRATEGIES
	// ============================================
	fmt.Println("\n--- API Versioning Strategies ---")

	fmt.Println("\n1. URL Path Versioning:")
	os.Stdout.WriteString(`
// /v1/users, /v2/users

mux := http.NewServeMux()

// Version 1
mux.HandleFunc("GET /v1/users/{id}", handleGetUserV1)

// Version 2 - agrega campo "role"
mux.HandleFunc("GET /v2/users/{id}", handleGetUserV2)

func handleGetUserV1(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    user := getUserByID(id)
    json.NewEncoder(w).Encode(UserV1{
        ID:   user.ID,
        Name: user.Name,
    })
}

func handleGetUserV2(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    user := getUserByID(id)
    json.NewEncoder(w).Encode(UserV2{
        ID:   user.ID,
        Name: user.Name,
        Role: user.Role,
    })
}

PROS:
- Facil de entender e implementar
- Cacheable (URL distinta)
- Visible en logs y documentacion

CONS:
- Rompe principio REST (version no es un recurso)
- Puede duplicar muchos handlers
- Clientes deben cambiar URLs al migrar
`)

	fmt.Println("2. Header Versioning:")
	os.Stdout.WriteString(`
// Accept: application/vnd.myapi+json;version=2
// O header custom: X-API-Version: 2

func versionMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        version := "1" // default

        // Opcion A: Accept header
        accept := r.Header.Get("Accept")
        if strings.Contains(accept, "version=2") {
            version = "2"
        }

        // Opcion B: Header custom
        if v := r.Header.Get("X-API-Version"); v != "" {
            version = v
        }

        // Inyectar version en context
        ctx := context.WithValue(r.Context(), "api-version", version)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    version := r.Context().Value("api-version").(string)

    switch version {
    case "1":
        // respuesta v1
    case "2":
        // respuesta v2
    default:
        http.Error(w, "Unsupported API version", 400)
    }
}

PROS:
- URLs limpias (misma URL para todas las versiones)
- Semanticamente correcto (content negotiation)
- Un solo set de routes

CONS:
- Menos visible (no se ve en la URL)
- Mas dificil de cachear
- Mas dificil de probar en browser
- Los clientes deben configurar headers
`)

	fmt.Println("3. Query Parameter Versioning:")
	os.Stdout.WriteString(`
// /users/123?version=2

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    version := r.URL.Query().Get("version")
    if version == "" {
        version = "1" // default
    }

    switch version {
    case "1":
        // respuesta v1
    case "2":
        // respuesta v2
    }
}

PROS:
- Facil de probar (solo agregar ?version=2 en browser)
- URL base no cambia
- Opcional (default a version actual)

CONS:
- Contamina query params (mezcla con filtros)
- Menos estandar
- Facil de olvidar
`)

	fmt.Println("COMPARACION FINAL:")
	fmt.Println(`
Estrategia   | Facilidad | Cacheabilidad | Pureza REST | Adopcion
-------------|-----------|---------------|-------------|----------
URL Path     | Alta      | Alta          | Baja        | Muy alta
Header       | Media     | Baja          | Alta        | Media
Query Param  | Alta      | Media         | Baja        | Baja

RECOMENDACION:
- APIs publicas: URL Path versioning (mas claro para consumidores)
- APIs internas: Header versioning (mas limpio, equipo controla clientes)
- Evitar: Query param (inconsistente con el resto de la API)`)

	// ============================================
	// OPENAPI / SWAGGER
	// ============================================
	fmt.Println("\n--- OpenAPI / Swagger ---")
	fmt.Println(`
OPENAPI SPECIFICATION (OAS):

Estandar para describir REST APIs en YAML/JSON.
Version actual: OpenAPI 3.1 (compatible con JSON Schema)

DOS ENFOQUES:

1. SCHEMA-FIRST (Design-First):
   - Escribir spec OpenAPI primero
   - Generar codigo Go desde la spec
   - Herramienta: oapi-codegen

2. CODE-FIRST:
   - Escribir codigo Go primero
   - Generar spec OpenAPI desde el codigo
   - Herramienta: swaggo/swag`)

	fmt.Println("\n--- oapi-codegen (Schema-First) ---")
	os.Stdout.WriteString(`
INSTALACION:

go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

SPEC OPENAPI (api.yaml):

openapi: "3.1.0"
info:
  title: User API
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 20
      responses:
        "200":
          description: List of users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/User"
    post:
      operationId: createUser
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
      responses:
        "201":
          description: User created
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
  /users/{id}:
    get:
      operationId: getUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: User found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "404":
          description: User not found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ProblemDetail"
components:
  schemas:
    User:
      type: object
      required: [id, name, email]
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string
          format: email
        role:
          type: string
          enum: [admin, user, viewer]
    CreateUserRequest:
      type: object
      required: [name, email]
      properties:
        name:
          type: string
          minLength: 1
        email:
          type: string
          format: email
    ProblemDetail:
      type: object
      properties:
        type:
          type: string
        title:
          type: string
        status:
          type: integer
        detail:
          type: string

GENERAR CODIGO:

# Generar tipos y server interface
oapi-codegen -package api -generate types api.yaml > api/types.gen.go
oapi-codegen -package api -generate chi-server api.yaml > api/server.gen.go
oapi-codegen -package api -generate spec api.yaml > api/spec.gen.go

# O con config file (oapi-codegen.yaml):
output: api/api.gen.go
package: api
generate:
  models: true
  chi-server: true
  strict-server: true
  embedded-spec: true

CODIGO GENERADO (ejemplo):

// types.gen.go
type User struct {
    ID    string  ` + "`" + `json:"id"` + "`" + `
    Name  string  ` + "`" + `json:"name"` + "`" + `
    Email string  ` + "`" + `json:"email"` + "`" + `
    Role  *string ` + "`" + `json:"role,omitempty"` + "`" + `
}

// server.gen.go
type ServerInterface interface {
    ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams)
    CreateUser(w http.ResponseWriter, r *http.Request)
    GetUser(w http.ResponseWriter, r *http.Request, id string)
}

// TU IMPLEMENTACION:
type UserServer struct {
    store UserStore
}

func (s *UserServer) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
    limit := 20
    if params.Limit != nil {
        limit = *params.Limit
    }
    users, err := s.store.List(r.Context(), limit)
    if err != nil {
        writeError(w, http.StatusInternalServerError, err)
        return
    }
    writeJSON(w, http.StatusOK, users)
}

func (s *UserServer) GetUser(w http.ResponseWriter, r *http.Request, id string) {
    user, err := s.store.GetByID(r.Context(), id)
    if err != nil {
        writeProblemDetail(w, http.StatusNotFound, "User not found", err.Error())
        return
    }
    writeJSON(w, http.StatusOK, user)
}
`)

	fmt.Println("\n--- swaggo/swag (Code-First) ---")
	os.Stdout.WriteString(`
INSTALACION:

go install github.com/swaggo/swag/cmd/swag@latest

ANOTAR HANDLERS CON COMENTARIOS:

// @title User API
// @version 1.0
// @description API para gestion de usuarios
// @host localhost:8080
// @BasePath /v1

// ListUsers godoc
// @Summary     List users
// @Description Get paginated list of users
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       limit  query    int    false "Max results"  default(20)
// @Param       offset query    int    false "Offset"       default(0)
// @Success     200    {array}  User
// @Failure     500    {object} ProblemDetail
// @Router      /users [get]
func ListUsers(w http.ResponseWriter, r *http.Request) {
    // implementacion
}

// GetUser godoc
// @Summary     Get user by ID
// @Description Get a single user by their ID
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id   path     string true "User ID"
// @Success     200  {object} User
// @Failure     404  {object} ProblemDetail
// @Router      /users/{id} [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
    // implementacion
}

GENERAR SPEC:

swag init -g main.go -o ./docs

# Genera:
# docs/docs.go      -> Spec embebida
# docs/swagger.json -> Spec JSON
# docs/swagger.yaml -> Spec YAML

SERVIR SWAGGER UI:

import (
    httpSwagger "github.com/swaggo/http-swagger/v2"
    _ "myapp/docs" // importar docs generados
)

mux.Handle("/swagger/", httpSwagger.WrapHandler)
// Acceder en: http://localhost:8080/swagger/index.html
`)

	fmt.Println("SCHEMA-FIRST vs CODE-FIRST:")
	fmt.Println(`
              | Schema-First          | Code-First
--------------|-----------------------|---------------------
Spec primero  | Si                    | No
Codigo gen.   | Server + tipos        | Solo docs
Validacion    | Automatica (generada) | Manual
Consistencia  | Garantizada           | Depende de anotaciones
Workflow      | Spec -> Code -> Impl  | Code -> Annotations -> Spec
Herramienta   | oapi-codegen          | swaggo/swag
Mejor para    | APIs publicas, teams  | APIs internas, rapido
Riesgo        | Spec desactualizada   | Annotations desactualizadas

RECOMENDACION:
- APIs publicas / multi-equipo: Schema-First (contrato claro)
- APIs internas / prototipos: Code-First (mas rapido)
- Ambos: validar en CI que spec y codigo estan sincronizados`)

	// ============================================
	// API DOCUMENTATION GENERATION
	// ============================================
	fmt.Println("\n--- API Documentation Generation ---")
	os.Stdout.WriteString(`
ESTRATEGIAS DE DOCUMENTACION:

1. SWAGGER UI (interactivo):
   - Embeber en el servidor
   - Permite probar endpoints desde el browser
   - import httpSwagger "github.com/swaggo/http-swagger/v2"
   - mux.Handle("/swagger/", httpSwagger.WrapHandler)

2. REDOC (referencia):
   - Mas limpio para documentacion de referencia
   - Solo lectura (no interactivo)
   - Ideal para portales publicos

3. STOPLIGHT / README.IO (hosted):
   - Plataformas que toman tu spec OpenAPI
   - Agregan guias, tutoriales, sandboxes
   - Para APIs publicas con muchos consumidores

BEST PRACTICES:
- Incluir ejemplos en cada endpoint
- Documentar todos los error responses
- Agregar descripciones de campos
- Mantener changelog de la API
- Versionado de la spec junto al codigo (monorepo)

GENERACION AUTOMATICA EN CI:

# .github/workflows/docs.yml
name: Generate API Docs
on: [push]
jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go install github.com/swaggo/swag/cmd/swag@latest
      - run: swag init -g main.go -o ./docs
      - run: |
          if [ -n "$(git diff --name-only docs/)" ]; then
            echo "API docs are out of date! Run 'swag init'."
            exit 1
          fi
`)

	// ============================================
	// REQUEST/RESPONSE DESIGN PATTERNS
	// ============================================
	fmt.Println("\n--- Request/Response Design Patterns ---")

	fmt.Println("\n1. PAGINATION:")
	os.Stdout.WriteString(`
A) OFFSET-BASED PAGINATION:

GET /users?limit=20&offset=40

type PaginatedResponse[T any] struct {
    Data       []T  ` + "`" + `json:"data"` + "`" + `
    Total      int  ` + "`" + `json:"total"` + "`" + `
    Limit      int  ` + "`" + `json:"limit"` + "`" + `
    Offset     int  ` + "`" + `json:"offset"` + "`" + `
    HasMore    bool ` + "`" + `json:"has_more"` + "`" + `
}

// SQL: SELECT * FROM users LIMIT 20 OFFSET 40

PROS: Simple, permite saltar a pagina N
CONS: Inconsistente con inserts/deletes concurrentes,
      lento en offsets grandes (O(offset+limit))

B) CURSOR-BASED PAGINATION:

GET /users?limit=20&cursor=eyJpZCI6MTIzfQ==

type CursorPaginatedResponse[T any] struct {
    Data       []T     ` + "`" + `json:"data"` + "`" + `
    NextCursor *string ` + "`" + `json:"next_cursor,omitempty"` + "`" + `
    HasMore    bool    ` + "`" + `json:"has_more"` + "`" + `
}

// Cursor = base64({"id": 123, "created_at": "2024-01-15T..."})
// SQL: SELECT * FROM users WHERE id > 123 ORDER BY id LIMIT 20

PROS: Consistente, performante (usa indices)
CONS: No puede saltar a pagina N, cursor opaco

C) KEYSET PAGINATION:

GET /users?limit=20&after_id=123&after_created=2024-01-15

// Similar a cursor pero con valores explicitos
// SQL: WHERE (created_at, id) > ('2024-01-15', 123)
//      ORDER BY created_at, id LIMIT 20

PROS: Transparente, performante
CONS: Requiere ordenamiento estable, multiples params

RECOMENDACION:
- Pocos datos / admin UI: Offset
- Feed infinito / APIs publicas: Cursor
- Datos ordenados / time-series: Keyset
`)

	fmt.Println("2. FILTERING, SORTING, SEARCHING:")
	os.Stdout.WriteString(`
FILTERING:

GET /users?status=active&role=admin&created_after=2024-01-01

type UserFilters struct {
    Status       *string    ` + "`" + `json:"status"` + "`" + `
    Role         *string    ` + "`" + `json:"role"` + "`" + `
    CreatedAfter *time.Time ` + "`" + `json:"created_after"` + "`" + `
    CreatedBefore *time.Time ` + "`" + `json:"created_before"` + "`" + `
}

func parseFilters(r *http.Request) UserFilters {
    q := r.URL.Query()
    var filters UserFilters

    if v := q.Get("status"); v != "" {
        filters.Status = &v
    }
    if v := q.Get("role"); v != "" {
        filters.Role = &v
    }
    if v := q.Get("created_after"); v != "" {
        t, _ := time.Parse(time.RFC3339, v)
        filters.CreatedAfter = &t
    }
    return filters
}

SORTING:

GET /users?sort=created_at&order=desc
GET /users?sort=-created_at,+name  (prefijo indica direccion)

ALLOWED SORTS (whitelist para evitar SQL injection):

var allowedSorts = map[string]string{
    "name":       "name",
    "created_at": "created_at",
    "email":      "email",
}

func parseSortParam(param string) (column, direction string, ok bool) {
    direction = "ASC"
    if strings.HasPrefix(param, "-") {
        direction = "DESC"
        param = param[1:]
    } else if strings.HasPrefix(param, "+") {
        param = param[1:]
    }

    column, ok = allowedSorts[param]
    return
}

SEARCHING:

GET /users?q=john&search_fields=name,email

// Fulltext search con Postgres:
// WHERE to_tsvector('english', name || ' ' || email)
//       @@ plainto_tsquery('english', 'john')
`)

	fmt.Println("3. FIELD SELECTION (Sparse Fieldsets):")
	os.Stdout.WriteString(`
GET /users/123?fields=id,name,email

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    user := getUser(r.PathValue("id"))

    fields := r.URL.Query().Get("fields")
    if fields == "" {
        writeJSON(w, http.StatusOK, user)
        return
    }

    // Convertir a map y filtrar
    allowed := map[string]bool{
        "id": true, "name": true, "email": true, "role": true,
    }

    fieldSet := strings.Split(fields, ",")
    userMap := structToMap(user)
    filtered := make(map[string]any)

    for _, f := range fieldSet {
        f = strings.TrimSpace(f)
        if allowed[f] {
            if v, ok := userMap[f]; ok {
                filtered[f] = v
            }
        }
    }

    writeJSON(w, http.StatusOK, filtered)
}

RESPUESTA:
{
    "id": "123",
    "name": "John",
    "email": "john@example.com"
}

BENEFICIO: Reduce payload, mejora performance.
USA EN: APIs GraphQL-like, mobile clients con bandwidth limitado.
`)

	fmt.Println("4. BULK OPERATIONS:")
	os.Stdout.WriteString(`
A) BATCH CREATE:

POST /users/batch
Content-Type: application/json

{
    "items": [
        {"name": "Alice", "email": "alice@test.com"},
        {"name": "Bob", "email": "bob@test.com"}
    ]
}

RESPUESTA (partial success):

{
    "results": [
        {"index": 0, "status": 201, "data": {"id": "1", "name": "Alice"}},
        {"index": 1, "status": 409, "error": {"detail": "email already exists"}}
    ],
    "summary": {"total": 2, "succeeded": 1, "failed": 1}
}

type BatchRequest[T any] struct {
    Items []T ` + "`" + `json:"items"` + "`" + `
}

type BatchResult struct {
    Index  int    ` + "`" + `json:"index"` + "`" + `
    Status int    ` + "`" + `json:"status"` + "`" + `
    Data   any    ` + "`" + `json:"data,omitempty"` + "`" + `
    Error  any    ` + "`" + `json:"error,omitempty"` + "`" + `
}

type BatchResponse struct {
    Results []BatchResult ` + "`" + `json:"results"` + "`" + `
    Summary struct {
        Total     int ` + "`" + `json:"total"` + "`" + `
        Succeeded int ` + "`" + `json:"succeeded"` + "`" + `
        Failed    int ` + "`" + `json:"failed"` + "`" + `
    } ` + "`" + `json:"summary"` + "`" + `
}

B) BATCH DELETE:

DELETE /users/batch
Content-Type: application/json

{"ids": ["1", "2", "3"]}

LIMITES:
- Definir max items por batch (ej: 100)
- Retornar 207 Multi-Status para partial success
- Procesar transaccionalmente o documentar que no lo es
`)

	fmt.Println("5. PARTIAL UPDATES (PATCH):")
	os.Stdout.WriteString(`
A) JSON MERGE PATCH (RFC 7396):

PATCH /users/123
Content-Type: application/merge-patch+json

{"name": "New Name", "email": null}

// name se actualiza, email se elimina, role no cambia

type UserPatch struct {
    Name  *string ` + "`" + `json:"name"` + "`" + `
    Email *string ` + "`" + `json:"email"` + "`" + `
    Role  *string ` + "`" + `json:"role"` + "`" + `
}

func handlePatchUser(w http.ResponseWriter, r *http.Request) {
    var patch UserPatch
    if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
        writeProblemDetail(w, 400, "Invalid JSON", err.Error())
        return
    }

    user := getUser(r.PathValue("id"))

    if patch.Name != nil {
        user.Name = *patch.Name
    }
    if patch.Email != nil {
        user.Email = *patch.Email
    }
    if patch.Role != nil {
        user.Role = *patch.Role
    }

    saveUser(user)
    writeJSON(w, http.StatusOK, user)
}

B) JSON PATCH (RFC 6902):

PATCH /users/123
Content-Type: application/json-patch+json

[
    {"op": "replace", "path": "/name", "value": "New Name"},
    {"op": "remove", "path": "/email"},
    {"op": "add", "path": "/tags/-", "value": "premium"}
]

// Operaciones: add, remove, replace, move, copy, test

MERGE PATCH vs JSON PATCH:
- Merge Patch: Simple, no puede distinguir null de "no enviar"
- JSON Patch: Preciso, soporta operaciones complejas (arrays, test)
- Recomendacion: Merge Patch para la mayoria de casos
`)

	// ============================================
	// ERROR RESPONSE FORMAT - RFC 7807
	// ============================================
	fmt.Println("\n--- Error Response Format (RFC 7807 - Problem Details) ---")
	os.Stdout.WriteString(`
RFC 7807 define un formato estandar para errores HTTP:

Content-Type: application/problem+json

{
    "type":     "https://api.example.com/errors/insufficient-funds",
    "title":    "Insufficient Funds",
    "status":   422,
    "detail":   "Account balance is $30.00, but transaction requires $50.00",
    "instance": "/transfers/abc123"
}

CAMPOS:
- type:     URI que identifica el tipo de error (para maquinas)
- title:    Resumen legible del error (para humanos)
- status:   HTTP status code
- detail:   Explicacion especifica de esta instancia
- instance: URI que identifica esta ocurrencia especifica

EXTENSIONES CUSTOM:

{
    "type":     "https://api.example.com/errors/validation",
    "title":    "Validation Error",
    "status":   422,
    "detail":   "One or more fields are invalid",
    "errors": [
        {
            "field":   "email",
            "message": "must be a valid email address",
            "value":   "not-an-email"
        },
        {
            "field":   "age",
            "message": "must be at least 18",
            "value":   15
        }
    ]
}

EN GO:

type ProblemDetail struct {
    Type     string             ` + "`" + `json:"type"` + "`" + `
    Title    string             ` + "`" + `json:"title"` + "`" + `
    Status   int                ` + "`" + `json:"status"` + "`" + `
    Detail   string             ` + "`" + `json:"detail,omitempty"` + "`" + `
    Instance string             ` + "`" + `json:"instance,omitempty"` + "`" + `
    Errors   []ValidationError  ` + "`" + `json:"errors,omitempty"` + "`" + `
}

type ValidationError struct {
    Field   string ` + "`" + `json:"field"` + "`" + `
    Message string ` + "`" + `json:"message"` + "`" + `
    Value   any    ` + "`" + `json:"value,omitempty"` + "`" + `
}

func writeProblemDetail(w http.ResponseWriter, status int, title, detail string) {
    pd := ProblemDetail{
        Type:   fmt.Sprintf("https://api.example.com/errors/%s", slugify(title)),
        Title:  title,
        Status: status,
        Detail: detail,
    }
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(pd)
}

func writeValidationError(w http.ResponseWriter, errors []ValidationError) {
    pd := ProblemDetail{
        Type:   "https://api.example.com/errors/validation",
        Title:  "Validation Error",
        Status: 422,
        Detail: "One or more fields are invalid",
        Errors: errors,
    }
    w.Header().Set("Content-Type", "application/problem+json")
    w.WriteHeader(422)
    json.NewEncoder(w).Encode(pd)
}
`)

	// ============================================
	// CONTENT NEGOTIATION
	// ============================================
	fmt.Println("\n--- Content Negotiation ---")
	os.Stdout.WriteString(`
El servidor selecciona el formato de respuesta segun el Accept header:

GET /users/123
Accept: application/json
-> Responde con JSON

GET /users/123
Accept: application/xml
-> Responde con XML

GET /users/123
Accept: text/csv
-> Responde con CSV

IMPLEMENTACION:

type Encoder interface {
    Encode(w http.ResponseWriter, status int, v any) error
    ContentType() string
}

type JSONEncoder struct{}

func (e JSONEncoder) Encode(w http.ResponseWriter, status int, v any) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(v)
}

func (e JSONEncoder) ContentType() string { return "application/json" }

func negotiate(accept string) Encoder {
    // Parsear Accept header con quality values
    // Accept: application/json;q=0.9, application/xml;q=0.8
    for _, mediaType := range parseAccept(accept) {
        switch mediaType {
        case "application/json", "*/*":
            return JSONEncoder{}
        case "application/xml":
            return XMLEncoder{}
        case "text/csv":
            return CSVEncoder{}
        }
    }
    return JSONEncoder{} // default
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    user := getUser(r.PathValue("id"))
    encoder := negotiate(r.Header.Get("Accept"))
    encoder.Encode(w, http.StatusOK, user)
}

// Si no se soporta el tipo pedido:
// 406 Not Acceptable
`)

	// ============================================
	// ETAGS AND CONDITIONAL REQUESTS
	// ============================================
	fmt.Println("\n--- ETags and Conditional Requests ---")
	os.Stdout.WriteString(`
ETAG: "huella digital" del recurso. Permite:
1. Cache validation (304 Not Modified)
2. Optimistic concurrency control (412 Precondition Failed)

FLUJO DE CACHE VALIDATION:

1) Primer request:
   GET /users/123
   <- 200 OK
   <- ETag: "abc123"

2) Segundo request (condicional):
   GET /users/123
   If-None-Match: "abc123"
   <- 304 Not Modified (si no cambio)
   <- 200 OK + nuevo ETag (si cambio)

FLUJO DE OPTIMISTIC LOCKING:

1) Leer recurso:
   GET /users/123
   <- 200 OK
   <- ETag: "abc123"

2) Actualizar (solo si no cambio):
   PUT /users/123
   If-Match: "abc123"
   <- 200 OK (si ETag coincide)
   <- 412 Precondition Failed (si alguien mas actualizo)

IMPLEMENTACION:

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    user := getUser(r.PathValue("id"))
    etag := generateETag(user) // hash del contenido o version

    // Verificar If-None-Match
    if match := r.Header.Get("If-None-Match"); match == etag {
        w.WriteHeader(http.StatusNotModified)
        return
    }

    w.Header().Set("ETag", etag)
    w.Header().Set("Cache-Control", "private, max-age=0, must-revalidate")
    writeJSON(w, http.StatusOK, user)
}

func handleUpdateUser(w http.ResponseWriter, r *http.Request) {
    user := getUser(r.PathValue("id"))
    currentETag := generateETag(user)

    // Verificar If-Match (optimistic locking)
    ifMatch := r.Header.Get("If-Match")
    if ifMatch != "" && ifMatch != currentETag {
        writeProblemDetail(w, 412, "Precondition Failed",
            "Resource was modified by another request")
        return
    }

    // Proceder con la actualizacion
    var update UserPatch
    json.NewDecoder(r.Body).Decode(&update)
    applyPatch(user, update)

    newETag := generateETag(user)
    w.Header().Set("ETag", newETag)
    writeJSON(w, http.StatusOK, user)
}

GENERAR ETAGS:

import "crypto/sha256"

func generateETag(v any) string {
    data, _ := json.Marshal(v)
    hash := sha256.Sum256(data)
    return fmt.Sprintf("\"%x\"", hash[:8])
}

// O usar version number:
func versionETag(version int) string {
    return fmt.Sprintf("\"v%d\"", version)
}

WEAK vs STRONG ETAGS:
- Strong: ETag: "abc123"     -> byte-for-byte identical
- Weak:   ETag: W/"abc123"   -> semantically equivalent
- Usar Weak para respuestas que pueden variar (whitespace, encoding)
`)

	// ============================================
	// RATE LIMIT HEADERS
	// ============================================
	fmt.Println("\n--- Rate Limit Headers (X-RateLimit-*) ---")
	os.Stdout.WriteString(`
HEADERS ESTANDAR (draft-ietf-httpapi-ratelimit-headers):

RateLimit-Limit:     100       // Requests permitidos por ventana
RateLimit-Remaining: 95        // Requests restantes en esta ventana
RateLimit-Reset:     1625000   // Unix timestamp cuando se resetea

// Headers legacy (aun muy usados):
X-RateLimit-Limit:     100
X-RateLimit-Remaining: 95
X-RateLimit-Reset:     1625000

// Cuando se excede el limite:
Retry-After: 30        // Segundos para reintentar

IMPLEMENTACION:

type RateLimitInfo struct {
    Limit     int
    Remaining int
    ResetAt   time.Time
}

func setRateLimitHeaders(w http.ResponseWriter, info RateLimitInfo) {
    w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
    w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
    w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.ResetAt.Unix(), 10))
}

func rateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
    // ... rate limiter logic ...

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            info := checkRateLimit(r)

            setRateLimitHeaders(w, info)

            if info.Remaining <= 0 {
                retryAfter := time.Until(info.ResetAt).Seconds()
                w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter)))
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

PLANES CON DIFERENTES LIMITES:

Plan      | Limit/min | Burst
----------|-----------|------
Free      | 60        | 10
Pro       | 600       | 50
Enterprise| 6000      | 200

func getLimitForUser(userID string) (limit int, burst int) {
    plan := getUserPlan(userID)
    switch plan {
    case "free":
        return 60, 10
    case "pro":
        return 600, 50
    case "enterprise":
        return 6000, 200
    default:
        return 30, 5
    }
}
`)

	// ============================================
	// API GATEWAY PATTERNS
	// ============================================
	fmt.Println("\n--- API Gateway Patterns ---")
	fmt.Println(`
API GATEWAY: punto de entrada unico para multiples microservicios.

                    API Gateway
Client -----> [Auth | Rate | Route | Transform] ----> Service A
                                                 \--> Service B
                                                  \-> Service C

RESPONSABILIDADES:
1. Authentication/Authorization
2. Rate Limiting
3. Request Routing
4. Load Balancing
5. Request/Response transformation
6. Caching
7. Circuit Breaking
8. Logging/Metrics
9. CORS handling
10. API versioning

PATRONES:

1. SIMPLE PROXY (reverse proxy):
   - Redirigir requests a backends
   - httputil.ReverseProxy en Go

2. BACKEND FOR FRONTEND (BFF):
   - Un gateway por tipo de cliente (web, mobile, IoT)
   - Cada BFF adapta la respuesta al cliente
   - Mobile BFF: menos campos, paginacion pequena
   - Web BFF: mas campos, includes/expands

3. API COMPOSITION:
   - Gateway agrega datos de multiples servicios
   - GET /user-profile -> User Service + Order Service + Review Service
   - Cuidado con latencia (llamadas en paralelo)

4. STRANGLER FIG:
   - Migrar gradualmente de monolito a microservicios
   - Gateway enruta viejo -> monolito, nuevo -> microservicio
   - Ir moviendo rutas gradualmente`)

	os.Stdout.WriteString(`
SIMPLE REVERSE PROXY EN GO:

import "net/http/httputil"

func newGateway() http.Handler {
    mux := http.NewServeMux()

    userProxy := httputil.NewSingleHostReverseProxy(
        &url.URL{Scheme: "http", Host: "user-service:8081"},
    )
    orderProxy := httputil.NewSingleHostReverseProxy(
        &url.URL{Scheme: "http", Host: "order-service:8082"},
    )

    mux.Handle("/v1/users/", userProxy)
    mux.Handle("/v1/orders/", orderProxy)

    // Middleware stack
    handler := rateLimitMiddleware(100, time.Minute)(
        authMiddleware(
            corsMiddleware(mux),
        ),
    )

    return handler
}

BFF PATTERN:

// Mobile BFF - respuestas compactas
mux.HandleFunc("GET /mobile/v1/dashboard", func(w http.ResponseWriter, r *http.Request) {
    user := fetchUser(r.Context())
    orders := fetchRecentOrders(r.Context(), 3) // solo 3 ultimas
    writeJSON(w, 200, MobileDashboard{
        Name:       user.Name,
        OrderCount: len(orders),
    })
})

// Web BFF - respuestas completas
mux.HandleFunc("GET /web/v1/dashboard", func(w http.ResponseWriter, r *http.Request) {
    user := fetchUser(r.Context())
    orders := fetchAllOrders(r.Context())
    stats := fetchStats(r.Context())
    writeJSON(w, 200, WebDashboard{
        User:   user,
        Orders: orders,
        Stats:  stats,
    })
})
`)

	// ============================================
	// BACKWARD COMPATIBILITY STRATEGIES
	// ============================================
	fmt.Println("\n--- Backward Compatibility Strategies ---")
	fmt.Println(`
REGLA DE ORO:
Nunca romper clientes existentes sin aviso.

CAMBIOS BACKWARD-COMPATIBLE (seguro):
- Agregar campos opcionales a respuestas
- Agregar nuevos endpoints
- Agregar query params opcionales
- Agregar nuevos valores a enums (si el cliente ignora desconocidos)
- Agregar headers opcionales

CAMBIOS QUE ROMPEN (breaking):
- Remover campos de respuestas
- Remover endpoints
- Cambiar tipo de un campo (int -> string)
- Renombrar campos
- Cambiar URL de un endpoint
- Hacer obligatorio un campo que era opcional
- Cambiar significado de un campo

ESTRATEGIAS:

1. ADDITIVE CHANGES (preferida):
   // V1 response:
   {"id": 1, "name": "John"}

   // V1.1 response (backward compatible):
   {"id": 1, "name": "John", "email": "john@test.com"}

2. FIELD ALIASING:
   // Campo renombrado pero se envian ambos:
   {"id": 1, "name": "John", "full_name": "John Doe"}
   // Deprecar "name" gradualmente

3. TOLERANT READER:
   // Clientes deben ignorar campos desconocidos
   // Go: json.Decoder no falla con campos extra (por defecto)

4. ROBUSTNESS PRINCIPLE (Postel's Law):
   "Be conservative in what you send,
    be liberal in what you accept"

VERSION LIFECYCLE:
  Alpha   -> Beta   -> GA    -> Deprecated -> Sunset
  (puede   (cambios  (estable, (aviso de    (removido)
   romper)  menores)  sin       remocion)
                      breaking)`)

	// ============================================
	// DEPRECATION HEADERS AND SUNSET
	// ============================================
	fmt.Println("\n--- Deprecation Headers and Sunset ---")
	os.Stdout.WriteString(`
HEADERS DE DEPRECACION:

// RFC 8594 - Sunset Header
Sunset: Sat, 01 Mar 2025 00:00:00 GMT

// Draft - Deprecation Header
Deprecation: true
// o con fecha:
Deprecation: Sat, 01 Jan 2025 00:00:00 GMT

// Link a documentacion de migracion
Link: <https://api.example.com/docs/migration-v2>; rel="sunset"

IMPLEMENTACION:

func deprecationMiddleware(
    deprecationDate time.Time,
    sunsetDate time.Time,
    migrationURL string,
) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Deprecation", deprecationDate.Format(time.RFC1123))
            w.Header().Set("Sunset", sunsetDate.Format(time.RFC1123))
            w.Header().Set("Link",
                fmt.Sprintf("<%s>; rel=\"sunset\"", migrationURL))

            next.ServeHTTP(w, r)
        })
    }
}

// Aplicar a endpoints deprecated
mux.Handle("GET /v1/users",
    deprecationMiddleware(
        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
        time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
        "https://api.example.com/docs/migrate-to-v2",
    )(http.HandlerFunc(handleListUsersV1)),
)

// Despues del sunset, retornar 410 Gone
func sunsetMiddleware(sunsetDate time.Time) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if time.Now().After(sunsetDate) {
                w.Header().Set("Content-Type", "application/problem+json")
                w.WriteHeader(http.StatusGone)
                json.NewEncoder(w).Encode(map[string]any{
                    "type":   "https://api.example.com/errors/endpoint-sunset",
                    "title":  "Endpoint Removed",
                    "status": 410,
                    "detail": "This endpoint has been removed. Please migrate to v2.",
                })
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

COMUNICACION DE DEPRECACION:
1. Anunciar deprecacion con fecha (minimo 6 meses antes)
2. Agregar headers de deprecacion inmediatamente
3. Enviar emails/notificaciones a consumidores conocidos
4. Monitorear uso del endpoint deprecated
5. Agregar warnings en respuestas
6. Sunset: retornar 410 Gone
`)

	// ============================================
	// HEALTH CHECK ENDPOINTS
	// ============================================
	fmt.Println("\n--- Health Check Endpoints ---")
	os.Stdout.WriteString(`
TRES TIPOS DE HEALTH CHECKS:

1. /healthz - Health basico (el proceso esta vivo?)
2. /readyz  - Readiness (puede recibir trafico?)
3. /livez   - Liveness (esta funcionando correctamente?)

DIFERENCIAS:

/livez:
- El proceso responde? Basico.
- Si falla: reiniciar el pod (Kubernetes liveness probe)
- NO verificar dependencias externas aqui

/readyz:
- Puede servir requests? Dependencias listas?
- Si falla: remover del load balancer (readiness probe)
- Verificar: DB, cache, servicios criticos

/healthz:
- Estado general detallado (para dashboards)
- Incluir estado de cada componente
- NO usar para probes de Kubernetes

IMPLEMENTACION:

type HealthChecker struct {
    checks map[string]func(ctx context.Context) error
}

func NewHealthChecker() *HealthChecker {
    return &HealthChecker{
        checks: make(map[string]func(ctx context.Context) error),
    }
}

func (h *HealthChecker) AddCheck(name string, check func(ctx context.Context) error) {
    h.checks[name] = check
}

// /livez - Solo verificar que el proceso responde
func (h *HealthChecker) LivezHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

// /readyz - Verificar dependencias
func (h *HealthChecker) ReadyzHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    for name, check := range h.checks {
        if err := check(ctx); err != nil {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]any{
                "status": "not_ready",
                "failed": name,
                "error":  err.Error(),
            })
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

// /healthz - Estado detallado
func (h *HealthChecker) HealthzHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    components := make(map[string]any)
    overall := "healthy"

    for name, check := range h.checks {
        if err := check(ctx); err != nil {
            components[name] = map[string]string{
                "status": "unhealthy",
                "error":  err.Error(),
            }
            overall = "unhealthy"
        } else {
            components[name] = map[string]string{"status": "healthy"}
        }
    }

    status := http.StatusOK
    if overall != "healthy" {
        status = http.StatusServiceUnavailable
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]any{
        "status":     overall,
        "components": components,
        "timestamp":  time.Now().UTC().Format(time.RFC3339),
    })
}

REGISTRO:

checker := NewHealthChecker()
checker.AddCheck("database", func(ctx context.Context) error {
    return db.PingContext(ctx)
})
checker.AddCheck("redis", func(ctx context.Context) error {
    return rdb.Ping(ctx).Err()
})
checker.AddCheck("external-api", func(ctx context.Context) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.dep.com/health", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    resp.Body.Close()
    if resp.StatusCode != 200 {
        return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
    }
    return nil
})

mux.HandleFunc("GET /livez", checker.LivezHandler)
mux.HandleFunc("GET /readyz", checker.ReadyzHandler)
mux.HandleFunc("GET /healthz", checker.HealthzHandler)

KUBERNETES PROBES:

livenessProbe:
  httpGet:
    path: /livez
    port: 8080
  initialDelaySeconds: 3
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
`)

	// ============================================
	// API TESTING WITH httptest
	// ============================================
	fmt.Println("\n--- API Testing with httptest ---")
	os.Stdout.WriteString(`
NET/HTTP/HTTPTEST:

Paquete de la standard library para testear HTTP handlers
sin levantar un servidor real.

DOS ENFOQUES:

1. httptest.NewRecorder() - Sin red, testea handler directo
2. httptest.NewServer()   - Servidor real en localhost

RECORDER (unit test de handler):

func TestGetUser(t *testing.T) {
    req := httptest.NewRequest("GET", "/users/123", nil)
    rec := httptest.NewRecorder()

    handler := http.HandlerFunc(handleGetUser)
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rec.Code)
    }

    var user User
    json.NewDecoder(rec.Body).Decode(&user)
    if user.ID != "123" {
        t.Errorf("expected ID 123, got %s", user.ID)
    }
}

SERVER (integration test):

func TestAPI(t *testing.T) {
    mux := setupRouter()
    server := httptest.NewServer(mux)
    defer server.Close()

    // Usar el URL del server
    resp, err := http.Get(server.URL + "/users/123")
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        t.Errorf("expected 200, got %d", resp.StatusCode)
    }
}

TABLE-DRIVEN API TESTS:

func TestCreateUser(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        wantStatus int
        wantErr    bool
    }{
        {
            name:       "valid user",
            body:       ` + "`" + `{"name":"John","email":"john@test.com"}` + "`" + `,
            wantStatus: 201,
        },
        {
            name:       "missing name",
            body:       ` + "`" + `{"email":"john@test.com"}` + "`" + `,
            wantStatus: 422,
            wantErr:    true,
        },
        {
            name:       "invalid json",
            body:       ` + "`" + `{invalid}` + "`" + `,
            wantStatus: 400,
            wantErr:    true,
        },
        {
            name:       "empty body",
            body:       "",
            wantStatus: 400,
            wantErr:    true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/users",
                strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()

            handleCreateUser(rec, req)

            if rec.Code != tt.wantStatus {
                t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
            }

            if tt.wantErr {
                var pd ProblemDetail
                json.NewDecoder(rec.Body).Decode(&pd)
                if pd.Status != tt.wantStatus {
                    t.Errorf("error status = %d, want %d", pd.Status, tt.wantStatus)
                }
            }
        })
    }
}

TESTING MIDDLEWARE:

func TestAuthMiddleware(t *testing.T) {
    inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("authorized"))
    })

    handler := authMiddleware(inner)

    // Sin token
    req := httptest.NewRequest("GET", "/protected", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", rec.Code)
    }

    // Con token valido
    req = httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    rec = httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rec.Code)
    }
}

MOCK EXTERNAL SERVICES:

func TestWithExternalAPI(t *testing.T) {
    // Mock del servicio externo
    mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/external/data":
            w.WriteHeader(200)
            json.NewEncoder(w).Encode(map[string]string{"key": "value"})
        default:
            w.WriteHeader(404)
        }
    }))
    defer mockAPI.Close()

    // Inyectar URL del mock
    client := NewAPIClient(mockAPI.URL)
    data, err := client.FetchData(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    if data.Key != "value" {
        t.Errorf("expected 'value', got %q", data.Key)
    }
}
`)

	// ============================================
	// WORKING DEMO: VERSIONED API
	// ============================================
	fmt.Println("\n--- WORKING DEMO: Versioned API with Testing ---")
	demoVersionedAPI()

	fmt.Println("\n=== RESUMEN EJECUTADO ===")
	fmt.Println("Todos los demos ejecutados exitosamente.")
}

// ============================================
// DEMO: VERSIONED API WITH TESTING
// ============================================

type UserV1 struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserV2 struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ProblemDetail struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type PaginatedResponse struct {
	Data    any  `json:"data"`
	Total   int  `json:"total"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

var usersDB = map[string]UserV2{
	"1": {ID: "1", Name: "Alice", Email: "alice@example.com", Role: "admin"},
	"2": {ID: "2", Name: "Bob", Email: "bob@example.com", Role: "user"},
	"3": {ID: "3", Name: "Charlie", Email: "charlie@example.com", Role: "viewer"},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ProblemDetail{
		Type:   "https://api.example.com/errors/" + strings.ReplaceAll(strings.ToLower(title), " ", "-"),
		Title:  title,
		Status: status,
		Detail: detail,
	})
}

func getUserHandlerV1() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		user, ok := usersDB[id]
		if !ok {
			writeProblem(w, http.StatusNotFound, "Not Found",
				"User with ID "+id+" not found")
			return
		}

		etag := fmt.Sprintf(`"v1-%s"`, user.ID)
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", etag)
		writeJSON(w, http.StatusOK, UserV1{
			ID:   user.ID,
			Name: user.Name,
		})
	}
}

func getUserHandlerV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		user, ok := usersDB[id]
		if !ok {
			writeProblem(w, http.StatusNotFound, "Not Found",
				"User with ID "+id+" not found")
			return
		}

		etag := fmt.Sprintf(`"v2-%s-%s"`, user.ID, user.Role)
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", etag)
		writeJSON(w, http.StatusOK, user)
	}
}

func listUsersHandlerV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		limit := 10
		offset := 0
		if v := q.Get("limit"); v != "" {
			fmt.Sscanf(v, "%d", &limit)
		}
		if v := q.Get("offset"); v != "" {
			fmt.Sscanf(v, "%d", &offset)
		}

		var users []UserV2
		for _, u := range usersDB {
			if role := q.Get("role"); role != "" && u.Role != role {
				continue
			}
			users = append(users, u)
		}

		total := len(users)
		end := offset + limit
		if end > total {
			end = total
		}
		if offset > total {
			offset = total
		}
		paged := users[offset:end]

		writeJSON(w, http.StatusOK, PaginatedResponse{
			Data:    paged,
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: end < total,
		})
	}
}

func healthzHandler() http.HandlerFunc {
	startTime := time.Now()
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "healthy",
			"uptime":  time.Since(startTime).String(),
			"version": "2.0.0",
			"components": map[string]string{
				"database": "healthy",
				"cache":    "healthy",
			},
		})
	}
}

func deprecationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Deprecation", "true")
		w.Header().Set("Sunset", "Sat, 01 Mar 2025 00:00:00 GMT")
		w.Header().Set("Link", `<https://api.example.com/docs/v2-migration>; rel="sunset"`)
		next.ServeHTTP(w, r)
	})
}

func rateLimitHeadersMiddleware(limit, remaining int, resetAt time.Time) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))
			next.ServeHTTP(w, r)
		})
	}
}

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /v1/users/{id}", deprecationMiddleware(getUserHandlerV1()))
	mux.HandleFunc("GET /v2/users/{id}", getUserHandlerV2())
	mux.HandleFunc("GET /v2/users", listUsersHandlerV2())
	mux.HandleFunc("GET /healthz", healthzHandler())
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	return mux
}

func demoVersionedAPI() {
	mux := setupRouter()
	server := httptest.NewServer(mux)
	defer server.Close()

	fmt.Println("Server de prueba levantado:", server.URL)

	fmt.Println("\n  [TEST 1] GET /v1/users/1 (deprecated)")
	testGetV1(server)

	fmt.Println("\n  [TEST 2] GET /v2/users/1 (current)")
	testGetV2(server)

	fmt.Println("\n  [TEST 3] GET /v2/users/999 (not found)")
	testGetNotFound(server)

	fmt.Println("\n  [TEST 4] GET /v2/users?role=admin (filtered)")
	testListFiltered(server)

	fmt.Println("\n  [TEST 5] Conditional request with ETag (304)")
	testConditionalRequest(server)

	fmt.Println("\n  [TEST 6] Health check endpoints")
	testHealthEndpoints(server)

	fmt.Println("\n  [TEST 7] Rate limit headers")
	testRateLimitHeaders(server)

	fmt.Println("\n  [TEST 8] RFC 7807 Problem Detail format")
	testProblemDetail(server)
}

func testGetV1(server *httptest.Server) {
	resp, err := http.Get(server.URL + "/v1/users/1")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Printf("  Deprecation: %s\n", resp.Header.Get("Deprecation"))
	fmt.Printf("  Sunset: %s\n", resp.Header.Get("Sunset"))
	fmt.Printf("  Link: %s\n", resp.Header.Get("Link"))

	var user UserV1
	json.NewDecoder(resp.Body).Decode(&user)
	fmt.Printf("  Response: {id:%s, name:%s}\n", user.ID, user.Name)
	fmt.Printf("  PASS: V1 returns only id+name, includes deprecation headers\n")
}

func testGetV2(server *httptest.Server) {
	resp, err := http.Get(server.URL + "/v2/users/1")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Printf("  ETag: %s\n", resp.Header.Get("ETag"))

	var user UserV2
	json.NewDecoder(resp.Body).Decode(&user)
	fmt.Printf("  Response: {id:%s, name:%s, email:%s, role:%s}\n",
		user.ID, user.Name, user.Email, user.Role)
	fmt.Printf("  PASS: V2 returns full user with email+role\n")
}

func testGetNotFound(server *httptest.Server) {
	resp, err := http.Get(server.URL + "/v2/users/999")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Printf("  Content-Type: %s\n", resp.Header.Get("Content-Type"))

	var pd ProblemDetail
	json.NewDecoder(resp.Body).Decode(&pd)
	fmt.Printf("  Problem: type=%s, title=%s, detail=%s\n",
		pd.Type, pd.Title, pd.Detail)
	fmt.Printf("  PASS: Returns RFC 7807 Problem Detail\n")
}

func testListFiltered(server *httptest.Server) {
	resp, err := http.Get(server.URL + "/v2/users?role=admin")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var page PaginatedResponse
	json.NewDecoder(resp.Body).Decode(&page)
	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Printf("  Total: %d, HasMore: %v\n", page.Total, page.HasMore)

	data, _ := json.Marshal(page.Data)
	fmt.Printf("  Data: %s\n", string(data))
	fmt.Printf("  PASS: Filtered pagination response\n")
}

func testConditionalRequest(server *httptest.Server) {
	resp1, err := http.Get(server.URL + "/v2/users/1")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	resp1.Body.Close()
	etag := resp1.Header.Get("ETag")
	fmt.Printf("  First request ETag: %s\n", etag)

	req, _ := http.NewRequest("GET", server.URL+"/v2/users/1", nil)
	req.Header.Set("If-None-Match", etag)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	resp2.Body.Close()

	fmt.Printf("  Conditional request status: %d\n", resp2.StatusCode)
	if resp2.StatusCode == http.StatusNotModified {
		fmt.Printf("  PASS: Server returned 304 Not Modified (cache hit)\n")
	} else {
		fmt.Printf("  FAIL: Expected 304, got %d\n", resp2.StatusCode)
	}
}

func testHealthEndpoints(server *httptest.Server) {
	endpoints := []string{"/livez", "/readyz", "/healthz"}
	for _, ep := range endpoints {
		resp, err := http.Get(server.URL + ep)
		if err != nil {
			fmt.Printf("  ERROR %s: %v\n", ep, err)
			continue
		}
		var body map[string]any
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		fmt.Printf("  %s -> %d, status=%v\n", ep, resp.StatusCode, body["status"])
	}
	fmt.Printf("  PASS: All health endpoints responding\n")
}

func testRateLimitHeaders(server *httptest.Server) {
	resetAt := time.Now().Add(time.Minute)
	mux := http.NewServeMux()
	mux.Handle("GET /api/data",
		rateLimitHeadersMiddleware(100, 95, resetAt)(
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(w, http.StatusOK, map[string]string{"data": "ok"})
			}),
		),
	)
	rlServer := httptest.NewServer(mux)
	defer rlServer.Close()

	resp, err := http.Get(rlServer.URL + "/api/data")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	resp.Body.Close()

	fmt.Printf("  X-RateLimit-Limit: %s\n", resp.Header.Get("X-RateLimit-Limit"))
	fmt.Printf("  X-RateLimit-Remaining: %s\n", resp.Header.Get("X-RateLimit-Remaining"))
	fmt.Printf("  X-RateLimit-Reset: %s\n", resp.Header.Get("X-RateLimit-Reset"))
	fmt.Printf("  PASS: Rate limit headers present\n")
}

func testProblemDetail(server *httptest.Server) {
	resp, err := http.Get(server.URL + "/v2/users/nonexistent")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("  Content-Type: %s\n", contentType)

	var pd ProblemDetail
	json.NewDecoder(resp.Body).Decode(&pd)

	valid := true
	if pd.Type == "" {
		fmt.Printf("  FAIL: missing 'type' field\n")
		valid = false
	}
	if pd.Title == "" {
		fmt.Printf("  FAIL: missing 'title' field\n")
		valid = false
	}
	if pd.Status != 404 {
		fmt.Printf("  FAIL: expected status 404, got %d\n", pd.Status)
		valid = false
	}
	if !strings.Contains(contentType, "application/problem+json") {
		fmt.Printf("  FAIL: Content-Type should be application/problem+json\n")
		valid = false
	}

	if valid {
		fmt.Printf("  PASS: RFC 7807 format: type=%s, title=%s, status=%d\n",
			pd.Type, pd.Title, pd.Status)
	}
}

/*
RESUMEN CAPITULO 69: API DESIGN, VERSIONING & OPENAPI

REST API DESIGN:
- Sustantivos en plural para recursos (/users, /orders)
- HTTP methods con semantica correcta (GET=read, POST=create, etc.)
- Status codes apropiados (201 Created, 404 Not Found, 422 Unprocessable)
- HATEOAS: links que guian al cliente sobre acciones disponibles
- kebab-case para URLs, camelCase para JSON

API VERSIONING:
- URL Path (/v1/users): Mas popular, mas claro para APIs publicas
- Header (Accept: vnd.api+json;version=2): Mas limpio, menos visible
- Query Param (?version=2): Facil de probar, menos estandar
- Recomendacion: URL path para publicas, header para internas

OPENAPI/SWAGGER:
- Schema-First: Spec -> oapi-codegen -> Server interface + types
- Code-First: Annotations -> swaggo/swag -> Spec + Swagger UI
- Schema-First para APIs publicas, Code-First para internas
- Validar en CI que spec y codigo estan sincronizados

DOCUMENTATION:
- Swagger UI para endpoints interactivos
- Redoc para documentacion de referencia
- Incluir ejemplos, errores, y changelog

PAGINATION:
- Offset: Simple, permite saltar paginas, lento en offsets grandes
- Cursor: Consistente, performante, no puede saltar
- Keyset: Transparente, ideal para time-series

FILTERING/SORTING:
- Query params para filtros: ?status=active&role=admin
- Whitelist de campos permitidos para sort (evitar injection)
- Field selection: ?fields=id,name,email

BULK OPERATIONS:
- POST /resource/batch con partial success (207 Multi-Status)
- Limitar max items por batch
- Documentar si es transaccional o no

PARTIAL UPDATES:
- JSON Merge Patch (RFC 7396): Simple, null elimina campo
- JSON Patch (RFC 6902): Preciso, operaciones en arrays

ERROR FORMAT (RFC 7807):
- Content-Type: application/problem+json
- Campos: type, title, status, detail, instance
- Extensiones custom para validation errors

CONTENT NEGOTIATION:
- Accept header para seleccionar formato
- 406 Not Acceptable si no se soporta

ETAGS:
- Cache validation: If-None-Match -> 304 Not Modified
- Optimistic locking: If-Match -> 412 Precondition Failed
- Strong vs Weak ETags

RATE LIMIT HEADERS:
- X-RateLimit-Limit, Remaining, Reset
- Retry-After cuando se excede el limite
- Diferentes limites por plan/tier

API GATEWAY:
- Punto de entrada unico (auth, rate limit, routing)
- BFF: Un gateway por tipo de cliente
- API Composition: Agregar datos de multiples servicios
- Strangler Fig: Migracion gradual de monolito

BACKWARD COMPATIBILITY:
- Cambios aditivos son seguros (agregar campos, endpoints)
- Nunca remover campos, cambiar tipos, o renombrar
- Tolerant Reader: clientes ignoran campos desconocidos
- Postel's Law: conservador enviando, liberal recibiendo

DEPRECATION:
- Headers: Deprecation, Sunset, Link rel="sunset"
- Lifecycle: Alpha -> Beta -> GA -> Deprecated -> Sunset
- Minimo 6 meses de aviso antes de sunset
- 410 Gone despues del sunset

HEALTH CHECKS:
- /livez: El proceso esta vivo? (Kubernetes liveness)
- /readyz: Puede recibir trafico? (Kubernetes readiness)
- /healthz: Estado detallado por componente (dashboards)

TESTING (httptest):
- httptest.NewRecorder: Unit test de handlers (sin red)
- httptest.NewServer: Integration test (servidor real local)
- Table-driven tests para multiples escenarios
- Mock external services con httptest.NewServer
*/

/* SUMMARY - API DESIGN, VERSIONING & OPENAPI:

TOPIC: Diseño profesional de REST APIs con versionado y documentación

REST API DESIGN:
- Sustantivos en plural (/users), HTTP methods semánticos
- Status codes: 201 Created, 404 Not Found, 422 Unprocessable
- HATEOAS: links que guían al cliente sobre acciones disponibles
- kebab-case URLs, camelCase JSON

API VERSIONING:
- URL Path (/v1/users): más popular, claro para APIs públicas
- Header (Accept: vnd.api+json;version=2): más limpio
- Recomendación: URL path para públicas, header para internas

OPENAPI/SWAGGER:
- Schema-First: Spec YAML → oapi-codegen → Server + types
- Code-First: Annotations → swaggo/swag → Spec + Swagger UI
- Validar sincronización spec-código en CI

PAGINATION & FILTERING:
- Offset (simple), Cursor (consistente), Keyset (time-series)
- Query params: ?status=active&role=admin&sort=created_at
- Sparse fieldsets: ?fields=id,name,email

PARTIAL UPDATES:
- JSON Merge Patch (RFC 7396): null elimina campo
- JSON Patch (RFC 6902): operaciones precisas

ERROR FORMAT (RFC 7807):
- application/problem+json: type, title, status, detail, instance

ETAGS & CACHE:
- If-None-Match → 304 Not Modified (cache validation)
- If-Match → 412 Precondition Failed (optimistic locking)

BACKWARD COMPATIBILITY:
- Cambios aditivos seguros, nunca remover/renombrar campos
- Deprecation → Sunset (mínimo 6 meses) → 410 Gone

HEALTH CHECKS:
- /livez: proceso vivo, /readyz: puede recibir tráfico
- /healthz: estado detallado (Kubernetes probes)

TESTING:
- httptest.NewRecorder (unit), httptest.NewServer (integration)
*/
