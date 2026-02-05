// Package main - Capitulo 67: GraphQL in Go
// GraphQL es un lenguaje de consulta para APIs que permite a los clientes
// solicitar exactamente los datos que necesitan. Go tiene excelente soporte
// con gqlgen (code-first) y graphql-go (schema-first).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== GRAPHQL IN GO ===")

	// ============================================
	// GRAPHQL vs REST
	// ============================================
	fmt.Println("\n--- GraphQL vs REST ---")
	fmt.Println(`
GraphQL es un lenguaje de consulta y runtime para APIs, creado por
Facebook en 2012 y liberado como open-source en 2015.

REST vs GraphQL:

REST:
- Multiples endpoints (/users, /users/1/posts, /users/1/posts/1/comments)
- Over-fetching: recibes campos que no necesitas
- Under-fetching: necesitas multiples requests para datos relacionados
- Versionado por URL (/v1/users, /v2/users)
- Documentacion separada (Swagger/OpenAPI)

GraphQL:
- UN solo endpoint (/graphql)
- Pides EXACTAMENTE lo que necesitas
- UN request para datos anidados/relacionados
- Sin versionado (evolucion del schema)
- Schema ES la documentacion (introspection)

EJEMPLO COMPARATIVO:

REST - 3 requests:
GET /users/1           -> { id, name, email, phone, address, ... }
GET /users/1/posts     -> [{ id, title, body, createdAt, ... }, ...]
GET /posts/42/comments -> [{ id, text, author, ... }, ...]

GraphQL - 1 request:
POST /graphql
{
  user(id: 1) {
    name
    posts(last: 5) {
      title
      comments(first: 3) {
        text
        author { name }
      }
    }
  }
}

CUANDO USAR GRAPHQL:
- APIs con clientes diversos (web, mobile, TV)
- Datos altamente relacionados (grafos)
- Equipos frontend y backend independientes
- Necesidad de evitar over/under-fetching

CUANDO USAR REST:
- APIs simples con recursos bien definidos
- File upload/download como operacion principal
- Caching HTTP agresivo (CDN)
- Microservicios con contratos simples
- APIs publicas con rate limiting por endpoint`)

	// ============================================
	// CONCEPTOS FUNDAMENTALES
	// ============================================
	fmt.Println("\n--- Conceptos Fundamentales de GraphQL ---")
	fmt.Println(`
SCHEMA DEFINITION LANGUAGE (SDL):

# Types
type User {
  id: ID!
  name: String!
  email: String!
  age: Int
  posts: [Post!]!
  createdAt: DateTime!
}

type Post {
  id: ID!
  title: String!
  body: String!
  author: User!
  comments: [Comment!]!
  tags: [String!]
}

type Comment {
  id: ID!
  text: String!
  author: User!
  post: Post!
}

# Queries (lectura)
type Query {
  user(id: ID!): User
  users(first: Int, after: String): UserConnection!
  post(id: ID!): Post
  search(query: String!): [SearchResult!]!
}

# Mutations (escritura)
type Mutation {
  createUser(input: CreateUserInput!): User!
  updateUser(id: ID!, input: UpdateUserInput!): User!
  deleteUser(id: ID!): Boolean!
  createPost(input: CreatePostInput!): Post!
}

# Subscriptions (tiempo real)
type Subscription {
  postCreated: Post!
  commentAdded(postId: ID!): Comment!
}

# Inputs
input CreateUserInput {
  name: String!
  email: String!
  age: Int
}

input UpdateUserInput {
  name: String
  email: String
  age: Int
}

# Enums
enum Role {
  ADMIN
  USER
  MODERATOR
}

# Interfaces
interface Node {
  id: ID!
}

# Unions
union SearchResult = User | Post | Comment

# Custom Scalars
scalar DateTime
scalar Upload

TIPOS ESCALARES BUILT-IN:
- Int:     Entero con signo de 32 bits
- Float:   Punto flotante de doble precision
- String:  Secuencia de caracteres UTF-8
- Boolean: true o false
- ID:      Identificador unico (serializado como String)

MODIFICADORES:
- String    -> nullable
- String!   -> non-null
- [String]  -> lista nullable de strings nullables
- [String!] -> lista nullable de strings non-null
- [String!]!-> lista non-null de strings non-null`)

	// ============================================
	// GQLGEN - SETUP
	// ============================================
	fmt.Println("\n--- gqlgen: Code-First GraphQL en Go ---")
	fmt.Println(`
gqlgen es la libreria de GraphQL mas popular para Go.
Genera codigo Go a partir del schema GraphQL.

FILOSOFIA: Schema-first con code generation.
- Defines el schema GraphQL
- gqlgen genera types, resolvers e interfaces
- Tu implementas la logica de negocio

INSTALACION:

# Inicializar modulo
go mod init github.com/yourname/myapp

# Instalar gqlgen
go get github.com/99designs/gqlgen

# Agregar herramienta al proyecto
printf '//go:build tools\npackage tools\nimport _ "github.com/99designs/gqlgen"' > tools.go

# Inicializar proyecto gqlgen
go run github.com/99designs/gqlgen init

ESTRUCTURA GENERADA:

myapp/
  gqlgen.yml          # Configuracion de gqlgen
  graph/
    generated.go       # Codigo generado (NO EDITAR)
    model/
      models_gen.go    # Modelos generados
    resolver.go        # Resolver raiz (editar)
    schema.graphqls    # Schema GraphQL (editar)
    schema.resolvers.go # Resolvers generados (implementar)
  server.go            # Entry point

CONFIGURACION (gqlgen.yml):

schema:
  - graph/*.graphqls

exec:
  filename: graph/generated.go
  package: graph

model:
  filename: graph/model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: graph
  package: graph

autobind:
  - "github.com/yourname/myapp/graph/model"

models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32
  Int:
    model:
      - github.com/99designs/gqlgen/graphql.Int
      - github.com/99designs/gqlgen/graphql.Int64
      - github.com/99designs/gqlgen/graphql.Int32

REGENERAR CODIGO:

go run github.com/99designs/gqlgen generate`)

	// ============================================
	// GQLGEN - SCHEMA Y RESOLVERS
	// ============================================
	fmt.Println("\n--- gqlgen: Schema y Resolvers ---")
	os.Stdout.WriteString(`
SCHEMA (graph/schema.graphqls):

type Todo {
  id: ID!
  text: String!
  done: Boolean!
  user: User!
}

type User {
  id: ID!
  name: String!
}

type Query {
  todos: [Todo!]!
  todo(id: ID!): Todo
}

type Mutation {
  createTodo(input: NewTodo!): Todo!
  updateTodo(id: ID!, input: UpdateTodo!): Todo!
}

input NewTodo {
  text: String!
  userId: String!
}

input UpdateTodo {
  text: String
  done: Boolean
}

RESOLVER RAIZ (graph/resolver.go):

package graph

import "github.com/yourname/myapp/graph/model"

type Resolver struct {
    todos  []*model.Todo
    nextID int
}

func NewResolver() *Resolver {
    return &Resolver{
        todos:  make([]*model.Todo, 0),
        nextID: 1,
    }
}

IMPLEMENTACION DE RESOLVERS (graph/schema.resolvers.go):

package graph

import (
    "context"
    "fmt"
    "github.com/yourname/myapp/graph/model"
)

// Query resolvers

func (r *queryResolver) Todos(ctx context.Context) ([]*model.Todo, error) {
    return r.todos, nil
}

func (r *queryResolver) Todo(ctx context.Context, id string) (*model.Todo, error) {
    for _, todo := range r.todos {
        if todo.ID == id {
            return todo, nil
        }
    }
    return nil, fmt.Errorf("todo not found: %s", id)
}

// Mutation resolvers

func (r *mutationResolver) CreateTodo(
    ctx context.Context, input model.NewTodo,
) (*model.Todo, error) {
    todo := &model.Todo{
        ID:   fmt.Sprintf("todo-%d", r.nextID),
        Text: input.Text,
        Done: false,
        User: &model.User{
            ID:   input.UserID,
            Name: "User " + input.UserID,
        },
    }
    r.nextID++
    r.todos = append(r.todos, todo)
    return todo, nil
}

func (r *mutationResolver) UpdateTodo(
    ctx context.Context, id string, input model.UpdateTodo,
) (*model.Todo, error) {
    for _, todo := range r.todos {
        if todo.ID == id {
            if input.Text != nil {
                todo.Text = *input.Text
            }
            if input.Done != nil {
                todo.Done = *input.Done
            }
            return todo, nil
        }
    }
    return nil, fmt.Errorf("todo not found: %s", id)
}

// Resolver type bindings

func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() QueryResolver       { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

SERVER (server.go):

package main

import (
    "log"
    "net/http"
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
    "github.com/yourname/myapp/graph"
)

func main() {
    resolver := graph.NewResolver()
    srv := handler.NewDefaultServer(
        graph.NewExecutableSchema(graph.Config{
            Resolvers: resolver,
        }),
    )

    http.Handle("/", playground.Handler("GraphQL Playground", "/query"))
    http.Handle("/query", srv)

    log.Println("Server running at http://localhost:8080/")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
`)

	// ============================================
	// GQLGEN - DATALOADERS
	// ============================================
	fmt.Println("\n--- gqlgen: Dataloaders (N+1 Prevention) ---")
	os.Stdout.WriteString(`
EL PROBLEMA N+1:

query {
  todos {        # 1 query: SELECT * FROM todos
    user {       # N queries: SELECT * FROM users WHERE id = ?
      name       # (una por cada todo!)
    }
  }
}

Si hay 100 todos, se ejecutan 101 queries SQL.

SOLUCION: Dataloaders agrupan y batch-ean las llamadas.

INSTALACION:

go get github.com/vektah/dataloaden
go get github.com/vikstrous/dataloadgen

USANDO dataloadgen:

import "github.com/vikstrous/dataloadgen"

// 1. Definir el loader
type UserLoader struct {
    loader *dataloadgen.Loader[string, *model.User]
}

func NewUserLoader(db *sql.DB) *UserLoader {
    return &UserLoader{
        loader: dataloadgen.NewLoader(
            func(ctx context.Context, keys []string) ([]*model.User, []error) {
                // UNA sola query para todos los keys
                // SELECT * FROM users WHERE id IN (?, ?, ?)
                users, err := db.GetUsersByIDs(ctx, keys)
                if err != nil {
                    errs := make([]error, len(keys))
                    for i := range errs {
                        errs[i] = err
                    }
                    return nil, errs
                }

                // Mapear resultados en el orden de keys
                userMap := make(map[string]*model.User)
                for _, u := range users {
                    userMap[u.ID] = u
                }

                result := make([]*model.User, len(keys))
                errs := make([]error, len(keys))
                for i, key := range keys {
                    if u, ok := userMap[key]; ok {
                        result[i] = u
                    } else {
                        errs[i] = fmt.Errorf("user %s not found", key)
                    }
                }
                return result, errs
            },
            // Esperar 2ms para acumular requests del mismo "tick"
            dataloadgen.WithWait(2*time.Millisecond),
        ),
    }
}

func (l *UserLoader) Load(ctx context.Context, id string) (*model.User, error) {
    return l.loader.Load(ctx, id)
}

// 2. Inyectar via middleware (per-request)
func DataloaderMiddleware(db *sql.DB, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        loaders := &Loaders{
            UserLoader: NewUserLoader(db),
        }
        ctx := context.WithValue(r.Context(), loadersKey, loaders)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

type Loaders struct {
    UserLoader *UserLoader
}

type contextKey string
const loadersKey = contextKey("dataloaders")

func GetLoaders(ctx context.Context) *Loaders {
    return ctx.Value(loadersKey).(*Loaders)
}

// 3. Usar en resolver
func (r *todoResolver) User(ctx context.Context, obj *model.Todo) (*model.User, error) {
    return GetLoaders(ctx).UserLoader.Load(ctx, obj.UserID)
}

RESULTADO: En vez de 101 queries, ahora son 2:
1. SELECT * FROM todos
2. SELECT * FROM users WHERE id IN ('1','2','3',...)

IMPORTANTE:
- Los dataloaders deben ser PER-REQUEST (no compartidos entre requests)
- Usar middleware para inyectarlos en el context
- El "wait" time (2ms) permite agrupar loads del mismo resolver cycle
`)

	// ============================================
	// GQLGEN - CUSTOM SCALARS
	// ============================================
	fmt.Println("\n--- gqlgen: Custom Scalars ---")
	os.Stdout.WriteString(`
Los escalares custom permiten tipos propios en el schema.

SCHEMA:

scalar DateTime
scalar Upload
scalar JSON
scalar UUID

type Event {
  id: UUID!
  name: String!
  startAt: DateTime!
  metadata: JSON
}

IMPLEMENTACION DE DateTime:

package model

import (
    "fmt"
    "io"
    "strconv"
    "time"
    "github.com/99designs/gqlgen/graphql"
)

func MarshalDateTime(t time.Time) graphql.Marshaler {
    return graphql.WriterFunc(func(w io.Writer) {
        io.WriteString(w, strconv.Quote(t.Format(time.RFC3339)))
    })
}

func UnmarshalDateTime(v interface{}) (time.Time, error) {
    switch v := v.(type) {
    case string:
        return time.Parse(time.RFC3339, v)
    case int64:
        return time.Unix(v, 0), nil
    default:
        return time.Time{}, fmt.Errorf("invalid DateTime: %v", v)
    }
}

CONFIGURACION EN gqlgen.yml:

models:
  DateTime:
    model:
      - github.com/yourname/myapp/graph/model.DateTime

O usando autobind con la funcion Marshal/Unmarshal.

IMPLEMENTACION DE JSON scalar:

package model

import (
    "encoding/json"
    "fmt"
    "io"
    "github.com/99designs/gqlgen/graphql"
)

type JSON map[string]interface{}

func MarshalJSON(val JSON) graphql.Marshaler {
    return graphql.WriterFunc(func(w io.Writer) {
        err := json.NewEncoder(w).Encode(val)
        if err != nil {
            panic(err)
        }
    })
}

func UnmarshalJSON(v interface{}) (JSON, error) {
    switch v := v.(type) {
    case map[string]interface{}:
        return JSON(v), nil
    case string:
        var result JSON
        err := json.Unmarshal([]byte(v), &result)
        return result, err
    default:
        return nil, fmt.Errorf("invalid JSON: %T", v)
    }
}
`)

	// ============================================
	// GQLGEN - AUTHENTICATION MIDDLEWARE
	// ============================================
	fmt.Println("\n--- gqlgen: Authentication Middleware ---")
	os.Stdout.WriteString(`
AUTENTICACION EN GRAPHQL:
No hay un estandar fijo. Se suele pasar el token via HTTP headers.

MIDDLEWARE DE AUTH:

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            // Permitir requests sin auth (queries publicas)
            next.ServeHTTP(w, r)
            return
        }

        // Validar token
        token = strings.TrimPrefix(token, "Bearer ")
        user, err := validateToken(token)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        // Inyectar usuario en context
        ctx := context.WithValue(r.Context(), userCtxKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

type userContextKey string
const userCtxKey = userContextKey("user")

func UserFromContext(ctx context.Context) (*model.User, bool) {
    user, ok := ctx.Value(userCtxKey).(*model.User)
    return user, ok
}

PROTEGER RESOLVERS:

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
    user, ok := UserFromContext(ctx)
    if !ok {
        return nil, fmt.Errorf("access denied: not authenticated")
    }
    return user, nil
}

DIRECTIVAS DE AUTORIZACION:

# Schema
directive @auth(requires: Role = USER) on FIELD_DEFINITION

type Query {
    publicData: String!
    userData: String! @auth
    adminData: String! @auth(requires: ADMIN)
}

# Implementacion en gqlgen
func AuthDirective(
    ctx context.Context, obj interface{}, next graphql.Resolver, requires model.Role,
) (interface{}, error) {
    user, ok := UserFromContext(ctx)
    if !ok {
        return nil, fmt.Errorf("access denied")
    }

    if requires == model.RoleAdmin && user.Role != model.RoleAdmin {
        return nil, fmt.Errorf("insufficient permissions: requires %s", requires)
    }

    return next(ctx)
}

# Configurar en server.go
srv := handler.NewDefaultServer(
    graph.NewExecutableSchema(graph.Config{
        Resolvers: resolver,
        Directives: graph.DirectiveRoot{
            Auth: AuthDirective,
        },
    }),
)

USO EN HTTP:
http.Handle("/query", AuthMiddleware(srv))
`)

	// ============================================
	// GQLGEN - ERROR HANDLING
	// ============================================
	fmt.Println("\n--- gqlgen: Error Handling ---")
	os.Stdout.WriteString(`
ERRORES EN GRAPHQL:
GraphQL tiene su propio formato de errores en la respuesta.

{
  "data": { "user": null },
  "errors": [
    {
      "message": "user not found",
      "path": ["user"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ]
}

NOTA: GraphQL puede retornar data parcial + errores.

ERRORES BASICOS EN gqlgen:

// Error simple
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
    user, err := r.db.GetUser(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user not found: %s", id)
    }
    return user, nil
}

ERRORES CON EXTENSIONS (codigos custom):

import "github.com/vektah/gqlparser/v2/gqlerror"

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
    user, err := r.db.GetUser(ctx, id)
    if err != nil {
        return nil, &gqlerror.Error{
            Message: "User not found",
            Extensions: map[string]interface{}{
                "code":   "NOT_FOUND",
                "userId": id,
            },
        }
    }
    return user, nil
}

ERROR PRESENTER (personalizar TODOS los errores):

srv := handler.NewDefaultServer(...)

srv.SetErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
    err := graphql.DefaultErrorPresenter(ctx, e)

    // Ocultar errores internos en produccion
    var gqlErr *gqlerror.Error
    if errors.As(e, &gqlErr) {
        return gqlErr // Errores GraphQL: mostrar tal cual
    }

    // Errores internos: mensaje generico
    err.Message = "internal server error"
    err.Extensions = map[string]interface{}{
        "code": "INTERNAL_ERROR",
    }
    return err
})

RECOVER PANICS:

srv.SetRecoverFunc(func(ctx context.Context, err interface{}) error {
    log.Printf("PANIC in GraphQL resolver: %v", err)
    return fmt.Errorf("internal server error")
})

FIELD ERRORS (errores por campo):

import "github.com/99designs/gqlgen/graphql"

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
    users, err := r.db.ListUsers(ctx)
    if err != nil {
        // Agregar error al campo sin detener la resolucion
        graphql.AddError(ctx, &gqlerror.Error{
            Message: "partial data: some users failed to load",
            Extensions: map[string]interface{}{
                "code": "PARTIAL_ERROR",
            },
        })
        return users, nil // Retorna datos parciales
    }
    return users, nil
}

VALIDACION DE INPUTS:

func (r *mutationResolver) CreateUser(
    ctx context.Context, input model.CreateUserInput,
) (*model.User, error) {
    var validationErrors []string

    if len(input.Name) < 2 {
        validationErrors = append(validationErrors, "name must be at least 2 chars")
    }
    if !strings.Contains(input.Email, "@") {
        validationErrors = append(validationErrors, "invalid email format")
    }

    if len(validationErrors) > 0 {
        return nil, &gqlerror.Error{
            Message: "validation failed",
            Extensions: map[string]interface{}{
                "code":   "VALIDATION_ERROR",
                "errors": validationErrors,
            },
        }
    }

    return r.db.CreateUser(ctx, input)
}
`)

	// ============================================
	// GQLGEN - PAGINATION
	// ============================================
	fmt.Println("\n--- gqlgen: Cursor-Based Pagination ---")
	os.Stdout.WriteString(`
GraphQL recomienda pagination basada en cursores (Relay spec).

SCHEMA:

type Query {
  users(first: Int, after: String, last: Int, before: String): UserConnection!
}

type UserConnection {
  edges: [UserEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type UserEdge {
  cursor: String!
  node: User!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

IMPLEMENTACION:

import "encoding/base64"

func encodeCursor(id int) string {
    return base64.StdEncoding.EncodeToString(
        []byte(fmt.Sprintf("cursor:%d", id)),
    )
}

func decodeCursor(cursor string) (int, error) {
    b, err := base64.StdEncoding.DecodeString(cursor)
    if err != nil {
        return 0, err
    }
    var id int
    _, err = fmt.Sscanf(string(b), "cursor:%d", &id)
    return id, err
}

func (r *queryResolver) Users(
    ctx context.Context, first *int, after *string, last *int, before *string,
) (*model.UserConnection, error) {
    limit := 20 // default
    if first != nil {
        limit = *first
    }

    var afterID int
    if after != nil {
        var err error
        afterID, err = decodeCursor(*after)
        if err != nil {
            return nil, fmt.Errorf("invalid cursor: %w", err)
        }
    }

    // Query: SELECT * FROM users WHERE id > afterID ORDER BY id LIMIT limit+1
    users, err := r.db.ListUsers(ctx, afterID, limit+1) // +1 para hasNextPage
    if err != nil {
        return nil, err
    }

    hasNextPage := len(users) > limit
    if hasNextPage {
        users = users[:limit]
    }

    edges := make([]*model.UserEdge, len(users))
    for i, u := range users {
        edges[i] = &model.UserEdge{
            Cursor: encodeCursor(u.ID),
            Node:   u,
        }
    }

    var startCursor, endCursor *string
    if len(edges) > 0 {
        startCursor = &edges[0].Cursor
        endCursor = &edges[len(edges)-1].Cursor
    }

    totalCount, _ := r.db.CountUsers(ctx)

    return &model.UserConnection{
        Edges: edges,
        PageInfo: &model.PageInfo{
            HasNextPage:     hasNextPage,
            HasPreviousPage: afterID > 0,
            StartCursor:     startCursor,
            EndCursor:       endCursor,
        },
        TotalCount: totalCount,
    }, nil
}

USO DESDE CLIENTE:

# Primera pagina
query {
  users(first: 10) {
    edges {
      cursor
      node { id name email }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}

# Siguiente pagina
query {
  users(first: 10, after: "Y3Vyc29yOjEw") {
    edges {
      cursor
      node { id name email }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`)

	// ============================================
	// GQLGEN - SUBSCRIPTIONS
	// ============================================
	fmt.Println("\n--- gqlgen: Subscriptions (WebSockets) ---")
	os.Stdout.WriteString(`
Subscriptions permiten push de datos en tiempo real via WebSocket.

SCHEMA:

type Subscription {
  messagePosted(roomId: ID!): Message!
  userJoined(roomId: ID!): User!
}

type Message {
  id: ID!
  text: String!
  author: User!
  createdAt: DateTime!
}

IMPLEMENTACION DEL RESOLVER:

type Resolver struct {
    // Channels por room para broadcast
    mu          sync.Mutex
    subscribers map[string][]chan *model.Message
}

func (r *subscriptionResolver) MessagePosted(
    ctx context.Context, roomID string,
) (<-chan *model.Message, error) {
    // Crear channel para este subscriber
    ch := make(chan *model.Message, 1)

    r.mu.Lock()
    r.subscribers[roomID] = append(r.subscribers[roomID], ch)
    r.mu.Unlock()

    // Cleanup cuando el cliente se desconecta
    go func() {
        <-ctx.Done()
        r.mu.Lock()
        defer r.mu.Unlock()

        subs := r.subscribers[roomID]
        for i, sub := range subs {
            if sub == ch {
                r.subscribers[roomID] = append(subs[:i], subs[i+1:]...)
                break
            }
        }
        close(ch)
    }()

    return ch, nil
}

// Broadcast: llamar cuando se crea un mensaje
func (r *Resolver) BroadcastMessage(roomID string, msg *model.Message) {
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, ch := range r.subscribers[roomID] {
        // Non-blocking send
        select {
        case ch <- msg:
        default:
            // Subscriber lento, skip
        }
    }
}

CONFIGURACION DEL TRANSPORT:

import (
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/handler/transport"
    "github.com/gorilla/websocket"
)

srv := handler.New(graph.NewExecutableSchema(graph.Config{
    Resolvers: resolver,
}))

// Agregar transports
srv.AddTransport(transport.Options{})
srv.AddTransport(transport.GET{})
srv.AddTransport(transport.POST{})

// WebSocket para subscriptions
srv.AddTransport(&transport.Websocket{
    Upgrader: websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true // Configurar para produccion
        },
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    },
    KeepAlivePingInterval: 10 * time.Second,
    InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
        // Autenticacion en WebSocket connection
        token := initPayload.Authorization()
        if token != "" {
            user, err := validateToken(token)
            if err != nil {
                return ctx, nil, fmt.Errorf("invalid token")
            }
            ctx = context.WithValue(ctx, userCtxKey, user)
        }
        return ctx, nil, nil
    },
})

CLIENTE JavaScript:

import { createClient } from 'graphql-ws';

const client = createClient({
  url: 'ws://localhost:8080/query',
  connectionParams: {
    authorization: 'Bearer <token>',
  },
});

client.subscribe(
  {
    query: ` + "`" + `subscription { messagePosted(roomId: "room1") { id text author { name } } }` + "`" + `,
  },
  {
    next: (data) => console.log('New message:', data),
    error: (err) => console.error('Error:', err),
    complete: () => console.log('Subscription complete'),
  },
);
`)

	// ============================================
	// GQLGEN - FILE UPLOADS
	// ============================================
	fmt.Println("\n--- gqlgen: File Uploads ---")
	fmt.Println(`
GraphQL soporta file uploads via multipart form data.

SCHEMA:

scalar Upload

type Mutation {
  uploadFile(file: Upload!): File!
  uploadFiles(files: [Upload!]!): [File!]!
}

type File {
  id: ID!
  name: String!
  size: Int!
  contentType: String!
  url: String!
}

CONFIGURACION:

srv := handler.NewDefaultServer(...)

// Habilitar multipart uploads
srv.AddTransport(transport.MultipartForm{
    MaxMemory:  32 << 20, // 32 MB
    MaxUploadSize: 50 << 20, // 50 MB max por archivo
})

RESOLVER:

import "github.com/99designs/gqlgen/graphql"

func (r *mutationResolver) UploadFile(
    ctx context.Context, file graphql.Upload,
) (*model.File, error) {
    // file.File     -> io.Reader
    // file.Filename -> nombre original
    // file.Size     -> tamano en bytes
    // file.ContentType -> MIME type

    // Guardar archivo
    dst, err := os.Create(filepath.Join("uploads", file.Filename))
    if err != nil {
        return nil, fmt.Errorf("failed to create file: %w", err)
    }
    defer dst.Close()

    written, err := io.Copy(dst, file.File)
    if err != nil {
        return nil, fmt.Errorf("failed to write file: %w", err)
    }

    return &model.File{
        ID:          uuid.New().String(),
        Name:        file.Filename,
        Size:        int(written),
        ContentType: file.ContentType,
        URL:         "/uploads/" + file.Filename,
    }, nil
}

// Multiple files
func (r *mutationResolver) UploadFiles(
    ctx context.Context, files []*graphql.Upload,
) ([]*model.File, error) {
    var result []*model.File
    for _, file := range files {
        f, err := r.UploadFile(ctx, *file)
        if err != nil {
            return nil, err
        }
        result = append(result, f)
    }
    return result, nil
}

CLIENTE (curl):

curl -X POST http://localhost:8080/query \
  -F operations='{ "query": "mutation($file: Upload!) { uploadFile(file: $file) { id name size } }", "variables": { "file": null } }' \
  -F map='{ "0": ["variables.file"] }' \
  -F 0=@myfile.pdf`)

	// ============================================
	// GRAPHQL-GO LIBRARY
	// ============================================
	fmt.Println("\n--- graphql-go/graphql: Alternativa Schema-Runtime ---")
	os.Stdout.WriteString(`
graphql-go/graphql es una alternativa que define el schema en Go
directamente (sin archivos .graphqls ni code generation).

INSTALACION:

go get github.com/graphql-go/graphql
go get github.com/graphql-go/handler

EJEMPLO COMPLETO:

package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "github.com/graphql-go/graphql"
    gqlhandler "github.com/graphql-go/handler"
)

// Definir types programaticamente
var userType = graphql.NewObject(graphql.ObjectConfig{
    Name: "User",
    Fields: graphql.Fields{
        "id":    &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
        "name":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
        "email": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
    },
})

var queryType = graphql.NewObject(graphql.ObjectConfig{
    Name: "Query",
    Fields: graphql.Fields{
        "user": &graphql.Field{
            Type: userType,
            Args: graphql.FieldConfigArgument{
                "id": &graphql.ArgumentConfig{
                    Type: graphql.NewNonNull(graphql.ID),
                },
            },
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                id := p.Args["id"].(string)
                // Buscar usuario por ID
                return map[string]interface{}{
                    "id":    id,
                    "name":  "John Doe",
                    "email": "john@example.com",
                }, nil
            },
        },
        "users": &graphql.Field{
            Type: graphql.NewList(userType),
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return []map[string]interface{}{
                    {"id": "1", "name": "Alice", "email": "alice@example.com"},
                    {"id": "2", "name": "Bob", "email": "bob@example.com"},
                }, nil
            },
        },
    },
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
    Name: "Mutation",
    Fields: graphql.Fields{
        "createUser": &graphql.Field{
            Type: userType,
            Args: graphql.FieldConfigArgument{
                "name":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                "email": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
            },
            Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                return map[string]interface{}{
                    "id":    "new-id",
                    "name":  p.Args["name"],
                    "email": p.Args["email"],
                }, nil
            },
        },
    },
})

func main() {
    schema, err := graphql.NewSchema(graphql.SchemaConfig{
        Query:    queryType,
        Mutation: mutationType,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Ejecutar query programaticamente
    result := graphql.Do(graphql.Params{
        Schema:        schema,
        RequestString: ` + "`" + `{ users { id name email } }` + "`" + `,
    })
    json.NewEncoder(os.Stdout).Encode(result)

    // O servir via HTTP
    h := gqlhandler.New(&gqlhandler.Config{
        Schema:     &schema,
        Pretty:     true,
        GraphiQL:   true,
        Playground: true,
    })

    http.Handle("/graphql", h)
    log.Println("Server at http://localhost:8080/graphql")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

COMPARACION gqlgen vs graphql-go:

| Aspecto           | gqlgen              | graphql-go          |
|-------------------|---------------------|---------------------|
| Approach          | Schema-first + gen  | Code-first          |
| Type Safety       | Compile-time        | Runtime             |
| Performance       | Excelente           | Buena               |
| Subscriptions     | Built-in            | Manual              |
| Dataloaders       | Plugin ecosystem    | Manual              |
| Learning Curve    | Medio               | Bajo                |
| Community         | Grande              | Media               |
| Stars (GitHub)    | 10k+                | 9.5k+               |

RECOMENDACION:
- Proyectos nuevos: gqlgen (type safety, performance, ecosystem)
- Prototipos rapidos: graphql-go (sin code gen, mas flexible)
- APIs grandes: gqlgen (mejor tooling, dataloaders, federation)
`)

	// ============================================
	// SCHEMA DESIGN BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Schema Design Best Practices ---")
	fmt.Println(`
1. NAMING CONVENTIONS:

# Types: PascalCase
type UserProfile { ... }

# Fields: camelCase
type User {
  firstName: String!
  lastName: String!
  createdAt: DateTime!
}

# Enums: SCREAMING_SNAKE_CASE
enum OrderStatus {
  PENDING
  PROCESSING
  SHIPPED
  DELIVERED
  CANCELLED
}

# Inputs: sufijo "Input"
input CreateUserInput { ... }
input UpdateUserInput { ... }
input UserFilterInput { ... }

# Payloads: sufijo "Payload"
type CreateUserPayload {
  user: User!
  errors: [UserError!]!
}

2. MUTATIONS - RETURN PAYLOADS, NOT TYPES:

# MAL
type Mutation {
  createUser(name: String!, email: String!): User!
}

# BIEN - Payload con errores tipados
type Mutation {
  createUser(input: CreateUserInput!): CreateUserPayload!
}

type CreateUserPayload {
  user: User
  userErrors: [UserError!]!
}

type UserError {
  field: [String!]
  message: String!
  code: UserErrorCode!
}

enum UserErrorCode {
  INVALID_EMAIL
  DUPLICATE_EMAIL
  NAME_TOO_SHORT
}

3. NULLABLE vs NON-NULL:

# Regla general:
# - Query return types: nullable (permite errores parciales)
# - Mutation return types: non-null
# - Input fields obligatorios: non-null
# - Input fields opcionales: nullable

type Query {
  user(id: ID!): User           # Nullable: puede no existir
  users: [User!]!               # Lista non-null, items non-null
}

type Mutation {
  createUser(input: CreateUserInput!): CreateUserPayload!  # Non-null
}

4. NODE INTERFACE (Global Object Identification):

interface Node {
  id: ID!
}

type User implements Node {
  id: ID!
  name: String!
}

type Query {
  node(id: ID!): Node
}

5. DEPRECATION:

type User {
  name: String! @deprecated(reason: "Use firstName and lastName")
  firstName: String!
  lastName: String!
}

6. DESCRIPTION (documentacion en schema):

"""
Represents a user in the system.
Users can create posts and comment on other users' posts.
"""
type User {
  "Unique identifier"
  id: ID!

  "Display name visible to other users"
  name: String!
}

7. EVITAR OVER-NESTING:

# MAL - Demasiado anidado
query {
  user(id: 1) {
    company {
      department {
        team {
          members {
            profile { ... }
          }
        }
      }
    }
  }
}

# BIEN - Acceso directo con filtros
query {
  teamMembers(teamId: "abc") {
    id
    name
    profile { ... }
  }
}`)

	// ============================================
	// FEDERATION
	// ============================================
	fmt.Println("\n--- Apollo Federation en Go ---")
	os.Stdout.WriteString(`
Federation permite componer multiples GraphQL services en un supergraph.

CONCEPTO:

                    +------------------+
                    |   Apollo Gateway  |  <- Unico endpoint para clientes
                    +------------------+
                   /         |          \
          +--------+   +---------+   +----------+
          | Users  |   |  Posts  |   | Comments |
          | Service|   | Service |   | Service  |
          +--------+   +---------+   +----------+

Cada servicio es dueno de sus types y puede extender types de otros.

SOPORTE EN gqlgen:

# gqlgen.yml
federation:
  filename: graph/federation.go
  package: graph
  version: 2

SCHEMA DEL USER SERVICE:

# Marcar como entity con @key
type User @key(fields: "id") {
  id: ID!
  name: String!
  email: String!
}

type Query {
  user(id: ID!): User
  users: [User!]!
}

SCHEMA DEL POST SERVICE:

# Extender User de otro servicio
type User @key(fields: "id") {
  id: ID!
  posts: [Post!]!   # Agregar campo posts
}

type Post @key(fields: "id") {
  id: ID!
  title: String!
  body: String!
  author: User!     # Referencia a User (resuelto por gateway)
}

type Query {
  post(id: ID!): Post
  posts: [Post!]!
}

ENTITY RESOLVER (generado automaticamente, tu implementas):

// User service
func (r *entityResolver) FindUserByID(
    ctx context.Context, id string,
) (*model.User, error) {
    return r.db.GetUser(ctx, id)
}

// Post service - Resolver para User.posts
func (r *userResolver) Posts(
    ctx context.Context, obj *model.User,
) ([]*model.Post, error) {
    return r.db.GetPostsByUserID(ctx, obj.ID)
}

DIRECTIVAS DE FEDERATION v2:

extend schema
  @link(url: "https://specs.apollo.dev/federation/v2.0",
        import: ["@key", "@shareable", "@provides", "@requires", "@external"])

type Product @key(fields: "id") {
  id: ID!
  name: String!
  price: Float! @shareable   # Puede ser resuelto por multiples servicios
}

type Review @key(fields: "id") {
  id: ID!
  body: String!
  product: Product! @provides(fields: "name")
  author: User!
}

type User @key(fields: "id") {
  id: ID! @external
  reviews: [Review!]! @requires(fields: "id")
}

GATEWAY CON APOLLO ROUTER:

# router.yaml
supergraph:
  introspection: true
  listen: 0.0.0.0:4000

subgraphs:
  users:
    routing_url: http://users-service:8080/query
  posts:
    routing_url: http://posts-service:8081/query
  comments:
    routing_url: http://comments-service:8082/query

# Ejecutar
rover supergraph compose --config supergraph.yaml > supergraph.graphql
router --supergraph supergraph.graphql
`)

	// ============================================
	// GRAPHQL CLIENTS IN GO
	// ============================================
	fmt.Println("\n--- GraphQL Clients in Go ---")
	os.Stdout.WriteString(`
Cuando tu servicio Go necesita CONSUMIR una API GraphQL.

1. machinebox/graphql - Cliente simple y ligero:

go get github.com/machinebox/graphql

client := graphql.NewClient("https://api.example.com/graphql")

// Query
req := graphql.NewRequest(` + "`" + `
    query ($id: ID!) {
        user(id: $id) {
            id
            name
            email
        }
    }
` + "`" + `)

req.Var("id", "123")
req.Header.Set("Authorization", "Bearer "+token)

var resp struct {
    User struct {
        ID    string ` + "`" + `json:"id"` + "`" + `
        Name  string ` + "`" + `json:"name"` + "`" + `
        Email string ` + "`" + `json:"email"` + "`" + `
    }
}

if err := client.Run(ctx, req, &resp); err != nil {
    log.Fatal(err)
}

fmt.Println("User:", resp.User.Name)


2. shurcooL/graphql - Type-safe con struct tags:

go get github.com/shurcooL/graphql

client := graphql.NewClient("https://api.example.com/graphql", nil)

// Query via struct tags
var query struct {
    User struct {
        ID    graphql.ID
        Name  graphql.String
        Email graphql.String
    } ` + "`" + `graphql:"user(id: $id)"` + "`" + `
}

variables := map[string]interface{}{
    "id": graphql.ID("123"),
}

err := client.Query(ctx, &query, variables)
if err != nil {
    log.Fatal(err)
}

fmt.Println("User:", query.User.Name)

// Mutation
var mutation struct {
    CreateUser struct {
        ID   graphql.ID
        Name graphql.String
    } ` + "`" + `graphql:"createUser(input: $input)"` + "`" + `
}

type CreateUserInput struct {
    Name  graphql.String ` + "`" + `json:"name"` + "`" + `
    Email graphql.String ` + "`" + `json:"email"` + "`" + `
}

mutationVars := map[string]interface{}{
    "input": CreateUserInput{
        Name:  "Alice",
        Email: "alice@example.com",
    },
}

err = client.Mutate(ctx, &mutation, mutationVars)


3. hasura/go-graphql-client - Fork activo de shurcooL:

go get github.com/hasura/go-graphql-client

// Soporta subscriptions
client := graphql.NewSubscriptionClient("wss://api.example.com/graphql").
    WithConnectionParams(map[string]interface{}{
        "authorization": "Bearer " + token,
    })

var sub struct {
    MessagePosted struct {
        ID   graphql.ID
        Text graphql.String
    } ` + "`" + `graphql:"messagePosted(roomId: $roomId)"` + "`" + `
}

variables := map[string]interface{}{
    "roomId": graphql.ID("room1"),
}

subID, err := client.Subscribe(&sub, variables, func(data []byte, err error) error {
    if err != nil {
        return err
    }
    fmt.Println("New message:", sub.MessagePosted.Text)
    return nil
})

// Ejecutar (bloquea)
go client.Run()

// Cancelar despues
client.Unsubscribe(subID)
client.Close()


COMPARACION DE CLIENTES:

| Cliente              | Type-safe | Subscriptions | Maintainance |
|----------------------|-----------|---------------|--------------|
| machinebox/graphql   | No        | No            | Baja         |
| shurcooL/graphql     | Si        | No            | Baja         |
| hasura/go-graphql    | Si        | Si            | Activa       |
| Khan/genqlient       | Si (gen)  | No            | Activa       |

RECOMENDACION:
- Rapido y simple: machinebox/graphql
- Type-safe con subscriptions: hasura/go-graphql-client
- Type-safe con code gen: Khan/genqlient
`)

	// ============================================
	// PERFORMANCE CONSIDERATIONS
	// ============================================
	fmt.Println("\n--- Performance: Query Complexity & Depth Limiting ---")
	os.Stdout.WriteString(`
GraphQL permite queries arbitrariamente complejas.
Sin proteccion, un cliente puede tumbar el servidor.

ATAQUES COMUNES:

# 1. Query depth attack
query {
  user(id: 1) {
    friends {
      friends {
        friends {
          friends {
            friends { ... }  # Infinita profundidad
          }
        }
      }
    }
  }
}

# 2. Query complexity/width attack
query {
  users(first: 1000) {
    posts(first: 1000) {
      comments(first: 1000) {
        author {
          posts(first: 1000) { ... }
        }
      }
    }
  }
}

# 3. Alias attack (bypass rate limiting)
query {
  a1: user(id: 1) { name }
  a2: user(id: 2) { name }
  a3: user(id: 3) { name }
  # ... 1000 aliases
}

PROTECCIONES EN gqlgen:

import (
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/handler/extension"
)

srv := handler.NewDefaultServer(...)

// 1. QUERY DEPTH LIMITING
// Rechaza queries con profundidad > N
srv.Use(extension.FixedComplexityLimit(100))

// 2. COMPLEXITY CUSTOM POR CAMPO

// En gqlgen.yml, habilitar complexity:
// complexity:
//   filename: graph/complexity.go
//   package: graph

// Configurar complexity functions:
srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{
    Resolvers: resolver,
    Complexity: graph.ComplexityRoot{
        Query: struct {
            Users func(childComplexity int, first *int, after *string) int
        }{
            Users: func(childComplexity int, first *int, after *string) int {
                limit := 20
                if first != nil {
                    limit = *first
                }
                return limit * childComplexity
            },
        },
    },
}))

srv.Use(extension.FixedComplexityLimit(1000))

// 3. AUTOMATIC PERSISTED QUERIES (APQ)
// El cliente envia un hash en vez de la query completa.
// Reduce bandwidth y permite whitelist de queries.

import "github.com/99designs/gqlgen/graphql/handler/extension"

srv.Use(extension.AutomaticPersistedQuery{
    Cache: lru.New[string](100), // github.com/hashicorp/golang-lru/v2
})

// 4. QUERY TIMEOUT
srv.Use(extension.FixedComplexityLimit(500))

// Timeout por request via context
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

http.Handle("/query", TimeoutMiddleware(30*time.Second)(srv))

// 5. MAX ALIAS LIMIT (custom extension)
type AliasLimit struct {
    MaxAliases int
}

func (a *AliasLimit) Validate(schema *ast.Schema, doc *ast.QueryDocument) gqlerror.List {
    count := 0
    for _, op := range doc.Operations {
        count += countAliases(op.SelectionSet)
    }
    if count > a.MaxAliases {
        return gqlerror.List{{
            Message: fmt.Sprintf("query has %d aliases, max allowed is %d", count, a.MaxAliases),
        }}
    }
    return nil
}

PRODUCCION CHECKLIST:

- [ ] Query depth limit (max 10-15 levels)
- [ ] Query complexity limit (max 500-1000 points)
- [ ] Request timeout (15-30 seconds)
- [ ] Max request body size (1MB)
- [ ] Rate limiting por cliente
- [ ] Persisted queries en produccion
- [ ] Deshabilitar introspection en produccion
- [ ] Logging de queries lentas
- [ ] Metricas (latency, errors, complexity)
`)

	// ============================================
	// TESTING GRAPHQL
	// ============================================
	fmt.Println("\n--- Testing GraphQL Endpoints ---")
	os.Stdout.WriteString(`
TESTING EN gqlgen:

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/99designs/gqlgen/client"
    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yourname/myapp/graph"
)

// Usando gqlgen test client
func TestCreateUser(t *testing.T) {
    resolver := graph.NewResolver()
    srv := handler.NewDefaultServer(
        graph.NewExecutableSchema(graph.Config{Resolvers: resolver}),
    )

    c := client.New(srv)

    var resp struct {
        CreateUser struct {
            ID   string
            Name string
        }
    }

    c.MustPost(` + "`" + `mutation {
        createUser(input: { name: "Alice", email: "alice@example.com" }) {
            id
            name
        }
    }` + "`" + `, &resp)

    assert.Equal(t, "Alice", resp.CreateUser.Name)
    assert.NotEmpty(t, resp.CreateUser.ID)
}

// Con variables
func TestGetUser(t *testing.T) {
    c := setupTestClient(t)

    var resp struct {
        User struct {
            ID    string
            Name  string
            Email string
        }
    }

    c.MustPost(` + "`" + `query GetUser($id: ID!) {
        user(id: $id) {
            id
            name
            email
        }
    }` + "`" + `, &resp, client.Var("id", "user-1"))

    assert.Equal(t, "user-1", resp.User.ID)
}

// Testing errores
func TestGetUser_NotFound(t *testing.T) {
    c := setupTestClient(t)

    var resp struct {
        User *struct {
            ID string
        }
    }

    err := c.Post(` + "`" + `query { user(id: "nonexistent") { id } }` + "`" + `, &resp)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}

// Con autenticacion
func TestProtectedQuery(t *testing.T) {
    c := setupTestClient(t)

    var resp struct {
        Me struct {
            ID   string
            Name string
        }
    }

    c.MustPost(` + "`" + `query { me { id name } }` + "`" + `, &resp,
        client.AddHeader("Authorization", "Bearer valid-token"),
    )

    assert.NotEmpty(t, resp.Me.ID)
}

// Testing con httptest (HTTP completo)
func TestGraphQLHTTP(t *testing.T) {
    resolver := graph.NewResolver()
    srv := handler.NewDefaultServer(
        graph.NewExecutableSchema(graph.Config{Resolvers: resolver}),
    )

    ts := httptest.NewServer(srv)
    defer ts.Close()

    body := map[string]interface{}{
        "query": ` + "`" + `{ users { id name } }` + "`" + `,
    }
    jsonBody, _ := json.Marshal(body)

    resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(jsonBody))
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var result struct {
        Data struct {
            Users []struct {
                ID   string ` + "`" + `json:"id"` + "`" + `
                Name string ` + "`" + `json:"name"` + "`" + `
            }
        }
    }

    json.NewDecoder(resp.Body).Decode(&result)
    assert.NotEmpty(t, result.Data.Users)
}

// Testing subscriptions
func TestSubscription(t *testing.T) {
    resolver := graph.NewResolver()
    srv := handler.New(
        graph.NewExecutableSchema(graph.Config{Resolvers: resolver}),
    )
    srv.AddTransport(transport.Websocket{})
    srv.AddTransport(transport.POST{})

    c := client.New(srv)

    // Subscribe
    sub := c.Websocket(` + "`" + `subscription {
        messagePosted(roomId: "room1") {
            id
            text
        }
    }` + "`" + `)
    defer sub.Close()

    // Trigger event
    go func() {
        time.Sleep(100 * time.Millisecond)
        resolver.BroadcastMessage("room1", &model.Message{
            ID:   "1",
            Text: "Hello!",
        })
    }()

    var msg struct {
        MessagePosted struct {
            ID   string
            Text string
        }
    }
    err := sub.Next(&msg)
    require.NoError(t, err)
    assert.Equal(t, "Hello!", msg.MessagePosted.Text)
}

ESTRUCTURA DE TESTS:

graph/
  resolver_test.go      # Tests de logica de resolvers
  integration_test.go   # Tests HTTP end-to-end
  subscription_test.go  # Tests de subscriptions
  testutil_test.go      # Helpers (setupTestClient, fixtures)
`)

	// ============================================
	// WORKING DEMO
	// ============================================
	fmt.Println("\n--- Demo: Mini GraphQL Engine ---")
	demoGraphQLEngine()

	fmt.Println("\n--- Demo: GraphQL HTTP Server Test ---")
	demoGraphQLHTTPTest()

	fmt.Println("\n--- Demo: Query Complexity Calculator ---")
	demoQueryComplexity()

	fmt.Println("\n--- Demo: Cursor-Based Pagination ---")
	demoCursorPagination()

	fmt.Println("\n--- Demo: Dataloader Pattern ---")
	demoDataloader()

	// ============================================
	// RESUMEN FINAL
	// ============================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== RESUMEN CAPITULO 67: GRAPHQL IN GO ===")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(`
LIBRERIAS PRINCIPALES:
- gqlgen:              Schema-first + code gen (RECOMENDADA)
- graphql-go/graphql:  Code-first, sin generacion
- graph-gophers:       Alternativa con buena API

CLIENTES:
- machinebox/graphql:  Simple, sin type safety
- shurcooL/graphql:    Type-safe con struct tags
- hasura/go-graphql:   Type-safe + subscriptions
- Khan/genqlient:      Type-safe con code generation

GQLGEN WORKFLOW:
1. Definir schema (.graphqls)
2. Ejecutar: go run github.com/99designs/gqlgen generate
3. Implementar resolvers generados
4. Configurar server con transports

DATALOADERS:
- Problema: N+1 queries en campos relacionados
- Solucion: Batch + cache per-request
- Libreria: vikstrous/dataloadgen
- SIEMPRE crear dataloaders per-request (middleware)

PAGINATION:
- Cursor-based (Relay spec) es el estandar
- Usa base64 encoded cursors
- Retorna Connection + Edges + PageInfo

SUBSCRIPTIONS:
- WebSocket transport (graphql-ws protocol)
- Channels de Go mapean perfectamente
- Cleanup via context cancellation

SECURITY:
- Query depth limiting (10-15 levels max)
- Query complexity analysis (500-1000 points max)
- Persisted queries en produccion
- Disable introspection en produccion
- Rate limiting por cliente/IP
- Request timeout (15-30s)
- Max request body size

FEDERATION:
- Apollo Federation v2 soportado en gqlgen
- @key para entities
- @external, @requires, @provides para campos
- Apollo Router como gateway

TESTING:
- gqlgen/client para unit tests
- httptest para integration tests
- client.Websocket() para subscription tests
- Mock resolvers con interfaces

PERFORMANCE TIPS:
- Dataloaders para TODOS los campos relacionados
- Complejidad por campo (no global)
- Field-level caching con Redis
- Query result caching con APQ
- Connection pooling en DB
- Monitoring: latency por resolver, query complexity`)
}

// ============================================
// DEMO: MINI GRAPHQL ENGINE
// ============================================

type Field struct {
	Name      string
	Arguments map[string]interface{}
	Children  []Field
}

type Schema struct {
	resolvers map[string]func(args map[string]interface{}) interface{}
}

func NewSchema() *Schema {
	return &Schema{
		resolvers: make(map[string]func(args map[string]interface{}) interface{}),
	}
}

func (s *Schema) AddResolver(name string, fn func(args map[string]interface{}) interface{}) {
	s.resolvers[name] = fn
}

func (s *Schema) Execute(fields []Field) map[string]interface{} {
	result := make(map[string]interface{})
	for _, f := range fields {
		resolver, ok := s.resolvers[f.Name]
		if !ok {
			result[f.Name] = nil
			continue
		}
		data := resolver(f.Arguments)
		if len(f.Children) > 0 {
			if m, ok := data.(map[string]interface{}); ok {
				filtered := make(map[string]interface{})
				for _, child := range f.Children {
					if val, exists := m[child.Name]; exists {
						filtered[child.Name] = val
					}
				}
				result[f.Name] = filtered
			} else {
				result[f.Name] = data
			}
		} else {
			result[f.Name] = data
		}
	}
	return result
}

func demoGraphQLEngine() {
	schema := NewSchema()

	users := map[string]map[string]interface{}{
		"1": {"id": "1", "name": "Alice", "email": "alice@go.dev", "age": 30},
		"2": {"id": "2", "name": "Bob", "email": "bob@go.dev", "age": 25},
		"3": {"id": "3", "name": "Charlie", "email": "charlie@go.dev", "age": 35},
	}

	schema.AddResolver("user", func(args map[string]interface{}) interface{} {
		id, _ := args["id"].(string)
		if user, ok := users[id]; ok {
			return user
		}
		return nil
	})

	schema.AddResolver("users", func(_ map[string]interface{}) interface{} {
		result := make([]interface{}, 0, len(users))
		for _, u := range users {
			result = append(result, u)
		}
		return result
	})

	fmt.Println("Query: user(id: \"1\") { name email }")
	result := schema.Execute([]Field{
		{
			Name:      "user",
			Arguments: map[string]interface{}{"id": "1"},
			Children: []Field{
				{Name: "name"},
				{Name: "email"},
			},
		},
	})
	printJSON(result)

	fmt.Println("\nQuery: user(id: \"2\") { name }")
	result = schema.Execute([]Field{
		{
			Name:      "user",
			Arguments: map[string]interface{}{"id": "2"},
			Children: []Field{
				{Name: "name"},
			},
		},
	})
	printJSON(result)

	fmt.Println("\nQuery: users { id name email }")
	result = schema.Execute([]Field{
		{
			Name: "users",
		},
	})
	printJSON(result)
}

// ============================================
// DEMO: GRAPHQL HTTP SERVER TEST
// ============================================

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []graphQLError `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string                 `json:"message"`
	Path    []string               `json:"path,omitempty"`
	Ext     map[string]interface{} `json:"extensions,omitempty"`
}

func demoGraphQLHTTPTest() {
	data := map[string]map[string]interface{}{
		"1": {"id": "1", "name": "Alice", "role": "ADMIN"},
		"2": {"id": "2", "name": "Bob", "role": "USER"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeGraphQLError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		query := strings.TrimSpace(req.Query)

		switch {
		case strings.Contains(query, "users"):
			users := make([]interface{}, 0)
			for _, u := range data {
				users = append(users, u)
			}
			resp := graphQLResponse{
				Data: map[string]interface{}{"users": users},
			}
			json.NewEncoder(w).Encode(resp)

		case strings.Contains(query, "user"):
			id, _ := req.Variables["id"].(string)
			if user, ok := data[id]; ok {
				resp := graphQLResponse{
					Data: map[string]interface{}{"user": user},
				}
				json.NewEncoder(w).Encode(resp)
			} else {
				resp := graphQLResponse{
					Data:   map[string]interface{}{"user": nil},
					Errors: []graphQLError{{Message: "user not found", Path: []string{"user"}}},
				}
				json.NewEncoder(w).Encode(resp)
			}

		default:
			writeGraphQLError(w, "unknown query", http.StatusOK)
		}
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	fmt.Println("Testing GraphQL HTTP endpoint...")

	body, _ := json.Marshal(graphQLRequest{
		Query: `{ users { id name role } }`,
	})
	resp, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	defer resp.Body.Close()

	var result graphQLResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("  Status: %d\n", resp.StatusCode)
	fmt.Print("  Response: ")
	printJSON(result)

	body, _ = json.Marshal(graphQLRequest{
		Query:     `query GetUser($id: ID!) { user(id: $id) { id name } }`,
		Variables: map[string]interface{}{"id": "1"},
	})
	resp2, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	defer resp2.Body.Close()

	var result2 graphQLResponse
	json.NewDecoder(resp2.Body).Decode(&result2)
	fmt.Printf("\n  GetUser(id=1) Status: %d\n", resp2.StatusCode)
	fmt.Print("  Response: ")
	printJSON(result2)

	body, _ = json.Marshal(graphQLRequest{
		Query:     `query GetUser($id: ID!) { user(id: $id) { id name } }`,
		Variables: map[string]interface{}{"id": "999"},
	})
	resp3, err := http.Post(ts.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	defer resp3.Body.Close()

	var result3 graphQLResponse
	json.NewDecoder(resp3.Body).Decode(&result3)
	fmt.Printf("\n  GetUser(id=999) Status: %d\n", resp3.StatusCode)
	fmt.Print("  Response: ")
	printJSON(result3)
}

func writeGraphQLError(w http.ResponseWriter, msg string, _ int) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graphQLResponse{
		Errors: []graphQLError{{Message: msg}},
	})
}

// ============================================
// DEMO: QUERY COMPLEXITY CALCULATOR
// ============================================

type QueryNode struct {
	Name       string
	Multiplier int
	Children   []QueryNode
}

func calculateComplexity(node QueryNode) int {
	if len(node.Children) == 0 {
		return 1
	}

	childComplexity := 0
	for _, child := range node.Children {
		childComplexity += calculateComplexity(child)
	}

	multiplier := node.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return multiplier * childComplexity
}

func calculateDepth(node QueryNode) int {
	if len(node.Children) == 0 {
		return 1
	}

	maxChildDepth := 0
	for _, child := range node.Children {
		d := calculateDepth(child)
		if d > maxChildDepth {
			maxChildDepth = d
		}
	}
	return 1 + maxChildDepth
}

func demoQueryComplexity() {
	simpleQuery := QueryNode{
		Name: "query",
		Children: []QueryNode{
			{Name: "user", Children: []QueryNode{
				{Name: "id"},
				{Name: "name"},
				{Name: "email"},
			}},
		},
	}

	fmt.Println("Simple query: { user { id name email } }")
	fmt.Printf("  Complexity: %d, Depth: %d\n",
		calculateComplexity(simpleQuery), calculateDepth(simpleQuery))

	nestedQuery := QueryNode{
		Name: "query",
		Children: []QueryNode{
			{Name: "users", Multiplier: 10, Children: []QueryNode{
				{Name: "name"},
				{Name: "posts", Multiplier: 20, Children: []QueryNode{
					{Name: "title"},
					{Name: "comments", Multiplier: 50, Children: []QueryNode{
						{Name: "text"},
						{Name: "author", Children: []QueryNode{
							{Name: "name"},
						}},
					}},
				}},
			}},
		},
	}

	fmt.Println("\nNested query: { users(first:10) { name posts(first:20) { title comments(first:50) { text author { name } } } } }")
	complexity := calculateComplexity(nestedQuery)
	depth := calculateDepth(nestedQuery)
	fmt.Printf("  Complexity: %d, Depth: %d\n", complexity, depth)

	maxComplexity := 1000
	maxDepth := 10
	fmt.Printf("  Limits: maxComplexity=%d, maxDepth=%d\n", maxComplexity, maxDepth)
	fmt.Printf("  Complexity OK: %v (complexity %d <= %d)\n", complexity <= maxComplexity, complexity, maxComplexity)
	fmt.Printf("  Depth OK: %v (depth %d <= %d)\n", depth <= maxDepth, depth, maxDepth)

	aliasAttack := QueryNode{
		Name: "query",
		Children: []QueryNode{
			{Name: "a1_user", Children: []QueryNode{{Name: "name"}}},
			{Name: "a2_user", Children: []QueryNode{{Name: "name"}}},
			{Name: "a3_user", Children: []QueryNode{{Name: "name"}}},
			{Name: "a4_user", Children: []QueryNode{{Name: "name"}}},
			{Name: "a5_user", Children: []QueryNode{{Name: "name"}}},
		},
	}

	fmt.Println("\nAlias attack: { a1: user { name } a2: user { name } ... }")
	fmt.Printf("  Aliases: %d, Max allowed: 10\n", len(aliasAttack.Children))
	fmt.Printf("  Alias count OK: %v\n", len(aliasAttack.Children) <= 10)
}

// ============================================
// DEMO: CURSOR-BASED PAGINATION
// ============================================

type PaginatedUser struct {
	ID    int
	Name  string
	Email string
}

type UserEdge struct {
	Cursor string
	Node   PaginatedUser
}

type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     string
	EndCursor       string
}

type UserConnection struct {
	Edges      []UserEdge
	PageInfo   PageInfo
	TotalCount int
}

func encodeCursor(id int) string {
	raw := fmt.Sprintf("cursor:%d", id)
	encoded := make([]byte, 0, len(raw)*2)
	for _, b := range []byte(raw) {
		encoded = append(encoded, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"[b%64])
	}
	return string(encoded)
}

func decodeCursor(cursor string) int {
	var id int
	_, _ = fmt.Sscanf(cursor, "%d", &id)
	for i := 0; i < len(cursor); i++ {
		id = i
	}
	return id
}

func paginateUsers(allUsers []PaginatedUser, first int, afterIdx int) UserConnection {
	start := afterIdx
	end := start + first
	if end > len(allUsers) {
		end = len(allUsers)
	}

	users := allUsers[start:end]

	edges := make([]UserEdge, len(users))
	for i, u := range users {
		edges[i] = UserEdge{
			Cursor: fmt.Sprintf("cursor:%d", start+i),
			Node:   u,
		}
	}

	var startCursor, endCursor string
	if len(edges) > 0 {
		startCursor = edges[0].Cursor
		endCursor = edges[len(edges)-1].Cursor
	}

	return UserConnection{
		Edges: edges,
		PageInfo: PageInfo{
			HasNextPage:     end < len(allUsers),
			HasPreviousPage: start > 0,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: len(allUsers),
	}
}

func demoCursorPagination() {
	allUsers := make([]PaginatedUser, 15)
	for i := range allUsers {
		allUsers[i] = PaginatedUser{
			ID:    i + 1,
			Name:  fmt.Sprintf("User_%d", i+1),
			Email: fmt.Sprintf("user%d@example.com", i+1),
		}
	}

	fmt.Printf("Total users: %d\n\n", len(allUsers))

	fmt.Println("Page 1 (first: 5, after: start):")
	page1 := paginateUsers(allUsers, 5, 0)
	for _, e := range page1.Edges {
		fmt.Printf("  [%s] %s (%s)\n", e.Cursor, e.Node.Name, e.Node.Email)
	}
	fmt.Printf("  PageInfo: hasNext=%v, hasPrev=%v, endCursor=%s\n",
		page1.PageInfo.HasNextPage, page1.PageInfo.HasPreviousPage, page1.PageInfo.EndCursor)

	fmt.Println("\nPage 2 (first: 5, after: cursor:4):")
	page2 := paginateUsers(allUsers, 5, 5)
	for _, e := range page2.Edges {
		fmt.Printf("  [%s] %s (%s)\n", e.Cursor, e.Node.Name, e.Node.Email)
	}
	fmt.Printf("  PageInfo: hasNext=%v, hasPrev=%v, endCursor=%s\n",
		page2.PageInfo.HasNextPage, page2.PageInfo.HasPreviousPage, page2.PageInfo.EndCursor)

	fmt.Println("\nPage 3 (first: 5, after: cursor:9):")
	page3 := paginateUsers(allUsers, 5, 10)
	for _, e := range page3.Edges {
		fmt.Printf("  [%s] %s (%s)\n", e.Cursor, e.Node.Name, e.Node.Email)
	}
	fmt.Printf("  PageInfo: hasNext=%v, hasPrev=%v, endCursor=%s\n",
		page3.PageInfo.HasNextPage, page3.PageInfo.HasPreviousPage, page3.PageInfo.EndCursor)

	fmt.Printf("\n  TotalCount: %d\n", page3.TotalCount)
}

// ============================================
// DEMO: DATALOADER PATTERN
// ============================================

type DataLoader[K comparable, V any] struct {
	batchFn  func(keys []K) (map[K]V, error)
	cache    map[K]V
	pending  []K
	maxBatch int
}

func NewDataLoader[K comparable, V any](
	batchFn func(keys []K) (map[K]V, error),
	maxBatch int,
) *DataLoader[K, V] {
	return &DataLoader[K, V]{
		batchFn:  batchFn,
		cache:    make(map[K]V),
		pending:  make([]K, 0),
		maxBatch: maxBatch,
	}
}

func (d *DataLoader[K, V]) Queue(key K) {
	if _, ok := d.cache[key]; ok {
		return
	}
	for _, k := range d.pending {
		if k == key {
			return
		}
	}
	d.pending = append(d.pending, key)
}

func (d *DataLoader[K, V]) Dispatch() (int, error) {
	if len(d.pending) == 0 {
		return 0, nil
	}

	keys := make([]K, len(d.pending))
	copy(keys, d.pending)
	d.pending = d.pending[:0]

	results, err := d.batchFn(keys)
	if err != nil {
		return 0, err
	}

	for k, v := range results {
		d.cache[k] = v
	}

	return len(keys), nil
}

func (d *DataLoader[K, V]) Get(key K) (V, bool) {
	val, ok := d.cache[key]
	return val, ok
}

func demoDataloader() {
	batchCalls := 0

	type User struct {
		ID   string
		Name string
	}

	loader := NewDataLoader(
		func(keys []string) (map[string]User, error) {
			batchCalls++
			fmt.Printf("  BATCH CALL #%d: loading %d users: %v\n", batchCalls, len(keys), keys)

			result := make(map[string]User)
			for _, k := range keys {
				result[k] = User{ID: k, Name: "User_" + k}
			}
			return result, nil
		},
		100,
	)

	fmt.Println("Without dataloader: N+1 problem")
	fmt.Println("  10 todos -> 10 individual user queries (BAD)")
	fmt.Println("  SELECT * FROM users WHERE id = 1")
	fmt.Println("  SELECT * FROM users WHERE id = 2")
	fmt.Println("  SELECT * FROM users WHERE id = 3")
	fmt.Println("  ... (10 queries total)")

	fmt.Println("\nWith dataloader: Batched loading")

	todoUserIDs := []string{"1", "2", "3", "1", "2", "4", "5", "3", "1", "6"}

	fmt.Printf("  Queueing %d user loads from todo resolvers...\n", len(todoUserIDs))
	for _, id := range todoUserIDs {
		loader.Queue(id)
	}

	count, err := loader.Dispatch()
	if err != nil {
		fmt.Println("  ERROR:", err)
		return
	}

	fmt.Printf("  Dispatched: %d unique keys batched into 1 SQL query\n", count)
	fmt.Println("  SELECT * FROM users WHERE id IN (1, 2, 3, 4, 5, 6)")

	fmt.Println("\n  Retrieved users:")
	for _, id := range []string{"1", "2", "3", "4", "5", "6"} {
		if u, ok := loader.Get(id); ok {
			fmt.Printf("    User{ID:%s, Name:%s}\n", u.ID, u.Name)
		}
	}

	fmt.Println("\n  Second access (from cache, no SQL):")
	loader.Queue("1")
	loader.Queue("2")
	loader.Queue("3")
	count2, _ := loader.Dispatch()
	fmt.Printf("  Batch calls for cached keys: %d (all from cache)\n", count2)

	fmt.Printf("\n  Summary: %d loads, %d batch call(s), 0 N+1 queries\n",
		len(todoUserIDs), batchCalls)
	fmt.Println("  N+1 prevented!")
}

// ============================================
// HELPERS
// ============================================

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		fmt.Println("ERROR marshaling JSON:", err)
		return
	}
	fmt.Println("  " + string(data))
}

// Suppress unused import warnings
var (
	_ = context.Background
	_ = errors.New
	_ = math.Min
	_ = os.Stdout
	_ = strings.TrimSpace
	_ = sync.Mutex{}
	_ = time.Now
)
