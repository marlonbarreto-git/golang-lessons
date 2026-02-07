// Package main - Chapter 059: Advanced Testing
// Testing avanzado en Go con testify, gomock, integration tests,
// test fixtures, golden files y test containers.
package main

import (
	"os"
	"fmt"
)

func main() {
	fmt.Println("=== ADVANCED TESTING EN GO ===")

	// ============================================
	// TESTIFY
	// ============================================
	fmt.Println("\n--- Testify Framework ---")
	os.Stdout.WriteString(`
INSTALACIÓN:
go get github.com/stretchr/testify

Testify tiene 4 paquetes principales:
1. assert   - Assertions (continúa en fallos)
2. require  - Assertions (para en primer fallo)
3. suite    - Test suites
4. mock     - Mocking framework

ASSERT VS REQUIRE:

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestUser_Assert(t *testing.T) {
    user := GetUser("123")

    // Continúa ejecutando después de fallo
    assert.NotNil(t, user)
    assert.Equal(t, "Alice", user.Name)
    assert.Equal(t, "alice@example.com", user.Email)
    // Todos los asserts se ejecutan
}

func TestUser_Require(t *testing.T) {
    user := GetUser("123")

    // Para en primer fallo
    require.NotNil(t, user)  // Si falla, test termina aquí
    require.Equal(t, "Alice", user.Name)
    require.Equal(t, "alice@example.com", user.Email)
}

ASSERTIONS COMUNES:

// Equality
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.EqualValues(t, expected, actual)  // Ignora tipos

// Nil
assert.Nil(t, object)
assert.NotNil(t, object)

// Boolean
assert.True(t, value)
assert.False(t, value)

// Errors
assert.NoError(t, err)
assert.Error(t, err)
assert.ErrorIs(t, err, target)
assert.ErrorAs(t, err, &target)
assert.ErrorContains(t, err, "substring")

// Collections
assert.Len(t, list, 5)
assert.Empty(t, list)
assert.NotEmpty(t, list)
assert.Contains(t, list, element)
assert.NotContains(t, list, element)
assert.ElementsMatch(t, expected, actual)  // Same elements, any order

// Types
assert.IsType(t, expectedType, object)
assert.Implements(t, (*Interface)(nil), object)

// Numbers
assert.Greater(t, 5, 3)
assert.GreaterOrEqual(t, 5, 5)
assert.Less(t, 3, 5)
assert.LessOrEqual(t, 3, 3)
assert.InDelta(t, 3.14, actual, 0.01)  // Float comparison

// Strings
assert.Contains(t, "hello world", "world")
assert.NotContains(t, "hello world", "foo")
assert.Regexp(t, "^[a-z]+$", "hello")

// Panic
assert.Panics(t, func() { panic("boom") })
assert.NotPanics(t, func() { /* safe code */ })

// Files
assert.FileExists(t, "/path/to/file")
assert.DirExists(t, "/path/to/dir")

CUSTOM MESSAGES:

assert.Equal(t, 42, result, "Result should be 42, but got %d", result)

EVENTUALLY/CONDITION:

import "time"

// Retry hasta que pase o timeout
assert.Eventually(t, func() bool {
    return isServiceReady()
}, 5*time.Second, 100*time.Millisecond, "Service should be ready")

// Verificar condición sin retry
assert.Condition(t, func() bool {
    return user.Age >= 18
}, "User must be adult")
`)

	// ============================================
	// TEST SUITES
	// ============================================
	fmt.Println("\n--- Test Suites con testify ---")
	os.Stdout.WriteString(`
TEST SUITE:

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
    suite.Suite
    db     *sql.DB
    repo   *UserRepository
}

// SetupSuite: Antes de toda la suite
func (s *UserTestSuite) SetupSuite() {
    s.db = setupTestDB()
    s.repo = NewUserRepository(s.db)
}

// TearDownSuite: Después de toda la suite
func (s *UserTestSuite) TearDownSuite() {
    s.db.Close()
}

// SetupTest: Antes de cada test
func (s *UserTestSuite) SetupTest() {
    clearDatabase(s.db)
    seedTestData(s.db)
}

// TearDownTest: Después de cada test
func (s *UserTestSuite) TearDownTest() {
    // Cleanup
}

// BeforeTest: Antes de test específico
func (s *UserTestSuite) BeforeTest(suiteName, testName string) {
    fmt.Printf("Running %s.%s\n", suiteName, testName)
}

// AfterTest: Después de test específico
func (s *UserTestSuite) AfterTest(suiteName, testName string) {
    fmt.Printf("Finished %s.%s\n", suiteName, testName)
}

// Tests (deben empezar con Test)
func (s *UserTestSuite) TestCreateUser() {
    user := &User{Name: "Alice", Email: "alice@example.com"}

    err := s.repo.Create(user)

    s.NoError(err)
    s.NotEmpty(user.ID)
    s.Equal("Alice", user.Name)
}

func (s *UserTestSuite) TestGetUser() {
    user, err := s.repo.GetByID("123")

    s.NoError(err)
    s.NotNil(user)
    s.Equal("Alice", user.Name)
}

func (s *UserTestSuite) TestDeleteUser() {
    err := s.repo.Delete("123")

    s.NoError(err)

    _, err = s.repo.GetByID("123")
    s.Error(err)
}

// Ejecutar suite
func TestUserTestSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}

SUBTESTS EN SUITE:

func (s *UserTestSuite) TestUserValidation() {
    tests := []struct {
        name    string
        user    *User
        wantErr bool
    }{
        {
            name:    "valid user",
            user:    &User{Name: "Alice", Email: "alice@example.com"},
            wantErr: false,
        },
        {
            name:    "missing email",
            user:    &User{Name: "Bob"},
            wantErr: true,
        },
        {
            name:    "invalid email",
            user:    &User{Name: "Charlie", Email: "invalid"},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        s.Run(tt.name, func() {
            err := s.repo.Create(tt.user)

            if tt.wantErr {
                s.Error(err)
            } else {
                s.NoError(err)
            }
        })
    }
}
`)

	// ============================================
	// TESTIFY MOCK
	// ============================================
	fmt.Println("\n--- Testify Mock ---")
	fmt.Println(`
CREAR MOCK:

import "github.com/stretchr/testify/mock"

type MockUserService struct {
    mock.Mock
}

func (m *MockUserService) GetUser(id string) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserService) CreateUser(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserService) UpdateUser(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserService) DeleteUser(id string) error {
    args := m.Called(id)
    return args.Error(0)
}

USAR MOCK:

func TestUserHandler_GetUser(t *testing.T) {
    mockService := new(MockUserService)
    handler := NewUserHandler(mockService)

    expectedUser := &User{ID: "123", Name: "Alice"}

    // Setup expectations
    mockService.On("GetUser", "123").Return(expectedUser, nil)

    // Test
    user, err := handler.GetUserByID("123")

    // Asserts
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)

    // Verify mock was called
    mockService.AssertExpectations(t)
}

MATCHERS:

// Cualquier valor
mockService.On("GetUser", mock.Anything).Return(user, nil)

// Tipo específico
mockService.On("CreateUser", mock.AnythingOfType("*User")).Return(nil)

// Matcher custom
mockService.On("CreateUser", mock.MatchedBy(func(u *User) bool {
    return u.Email != ""
})).Return(nil)

MULTIPLE CALLS:

// Primera llamada retorna error, segunda OK
mockService.On("GetUser", "123").Return(nil, errors.New("not found")).Once()
mockService.On("GetUser", "123").Return(user, nil).Once()

// 3 veces
mockService.On("GetUser", mock.Anything).Return(user, nil).Times(3)

// Sin límite (default)
mockService.On("GetUser", mock.Anything).Return(user, nil)

RUN FUNCTION:

mockService.On("CreateUser", mock.AnythingOfType("*User")).Run(func(args mock.Arguments) {
    user := args.Get(0).(*User)
    user.ID = "generated-id"
    user.CreatedAt = time.Now()
}).Return(nil)

MAYBE:

// No es error si no se llama
mockService.On("GetUser", "123").Return(user, nil).Maybe()

UNSET:

mockService.On("GetUser", "123").Return(user, nil)
mockService.On("GetUser", "123").Unset()  // Remove expectation`)
	// ============================================
	// GOMOCK
	// ============================================
	fmt.Println("\n--- gomock (go.uber.org/mock) ---")
	os.Stdout.WriteString(`
GENERAR MOCK:

//go:generate mockgen -source=user_service.go -destination=mock_user_service.go -package=service

O:

//go:generate mockgen -destination=mocks/mock_database.go -package=mocks github.com/myapp/database Database

USAR GOMOCK:

import (
    "testing"
    "go.uber.org/mock/gomock"
)

func TestUserHandler(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockService := NewMockUserService(ctrl)
    handler := NewUserHandler(mockService)

    expectedUser := &User{ID: "123", Name: "Alice"}

    mockService.EXPECT().
        GetUser("123").
        Return(expectedUser, nil)

    user, err := handler.GetUserByID("123")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "Alice" {
        t.Errorf("expected Alice, got %s", user.Name)
    }
}

MATCHERS:

// Any
mockService.EXPECT().GetUser(gomock.Any()).Return(user, nil)

// Eq (explicit)
mockService.EXPECT().GetUser(gomock.Eq("123")).Return(user, nil)

// Not
mockService.EXPECT().GetUser(gomock.Not("456")).Return(user, nil)

// Nil
mockService.EXPECT().CreateUser(gomock.Nil()).Return(errors.New("user is nil"))

// AssignableToTypeOf
mockService.EXPECT().CreateUser(gomock.AssignableToTypeOf(&User{})).Return(nil)

ORDER:

gomock.InOrder(
    mockService.EXPECT().CreateUser(user1).Return(nil),
    mockService.EXPECT().CreateUser(user2).Return(nil),
    mockService.EXPECT().CreateUser(user3).Return(nil),
)

TIMES:

mockService.EXPECT().GetUser("123").Return(user, nil).Times(3)
mockService.EXPECT().GetUser("123").Return(user, nil).MinTimes(1)
mockService.EXPECT().GetUser("123").Return(user, nil).MaxTimes(5)
mockService.EXPECT().GetUser("123").Return(user, nil).AnyTimes()

DO:

mockService.EXPECT().
    CreateUser(gomock.Any()).
    Do(func(u *User) {
        fmt.Printf("Creating user: %s\n", u.Name)
    }).
    Return(nil)

DOANDRETURN:

mockService.EXPECT().
    GetUser(gomock.Any()).
    DoAndReturn(func(id string) (*User, error) {
        return &User{ID: id, Name: "Generated"}, nil
    })
`)

	// ============================================
	// INTEGRATION TESTING
	// ============================================
	fmt.Println("\n--- Integration Testing ---")
	os.Stdout.WriteString(`
ESTRUCTURA:

project/
├── internal/
│   ├── user/
│   │   ├── handler.go
│   │   ├── handler_test.go      # Unit tests
│   │   └── repository.go
├── tests/
│   └── integration/
│       ├── user_test.go         # Integration tests
│       ├── order_test.go
│       └── testhelpers/
│           ├── database.go
│           └── server.go

BUILD TAG:

// +build integration

package integration

import "testing"

func TestUserAPI(t *testing.T) {
    // Integration test
}

EJECUTAR:

go test -tags=integration ./tests/integration/...

O separar:

go test -short ./...           # Solo unit tests
go test -run Integration ./... # Solo integration

TEST HELPERS:

package testhelpers

import (
    "database/sql"
    "testing"
)

func SetupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgres://localhost/testdb")
    if err != nil {
        t.Fatalf("Failed to connect to DB: %v", err)
    }

    // Run migrations
    runMigrations(db)

    t.Cleanup(func() {
        db.Close()
    })

    return db
}

func ClearDatabase(t *testing.T, db *sql.DB) {
    tables := []string{"orders", "users"}
    for _, table := range tables {
        _, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
        if err != nil {
            t.Fatalf("Failed to clear table %s: %v", table, err)
        }
    }
}

func SeedTestData(t *testing.T, db *sql.DB) {
    _, err := db.Exec(` + "`" + `
        INSERT INTO users (id, name, email) VALUES
        ('1', 'Alice', 'alice@example.com'),
        ('2', 'Bob', 'bob@example.com')
    ` + "`" + `)
    if err != nil {
        t.Fatalf("Failed to seed data: %v", err)
    }
}

HTTP INTEGRATION TEST:

func TestUserAPI(t *testing.T) {
    // Setup
    db := testhelpers.SetupTestDB(t)
    server := testhelpers.NewTestServer(db)
    defer server.Close()

    client := server.Client()

    // Test GET
    resp, err := client.Get(server.URL + "/api/users/1")
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, resp.StatusCode)

    var user User
    json.NewDecoder(resp.Body).Decode(&user)
    assert.Equal(t, "Alice", user.Name)

    // Test POST
    newUser := User{Name: "Charlie", Email: "charlie@example.com"}
    body, _ := json.Marshal(newUser)

    resp, err = client.Post(server.URL+"/api/users", "application/json", bytes.NewBuffer(body))
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
}

httptest.Server HELPER:

func NewTestServer(db *sql.DB) *httptest.Server {
    router := setupRouter(db)
    return httptest.NewServer(router)
}
`)

	// ============================================
	// TEST FIXTURES
	// ============================================
	fmt.Println("\n--- Test Fixtures ---")
	os.Stdout.WriteString(`
FIXTURE BUILDERS:

package fixtures

type UserBuilder struct {
    user *User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: &User{
            ID:    uuid.NewString(),
            Name:  "Test User",
            Email: "test@example.com",
            Age:   25,
        },
    }
}

func (b *UserBuilder) WithID(id string) *UserBuilder {
    b.user.ID = id
    return b
}

func (b *UserBuilder) WithName(name string) *UserBuilder {
    b.user.Name = name
    return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) WithAge(age int) *UserBuilder {
    b.user.Age = age
    return b
}

func (b *UserBuilder) Build() *User {
    return b.user
}

// Uso
user := fixtures.NewUserBuilder().
    WithName("Alice").
    WithEmail("alice@example.com").
    WithAge(30).
    Build()

FIXTURE FILES (JSON):

// testdata/users/alice.json
{
  "id": "123",
  "name": "Alice",
  "email": "alice@example.com"
}

func LoadUserFixture(t *testing.T, name string) *User {
    data, err := os.ReadFile("testdata/users/" + name + ".json")
    require.NoError(t, err)

    var user User
    err = json.Unmarshal(data, &user)
    require.NoError(t, err)

    return &user
}

// Uso
alice := LoadUserFixture(t, "alice")

FACTORY PATTERN:

type UserFactory struct {
    counter int
}

func (f *UserFactory) Create() *User {
    f.counter++
    return &User{
        ID:    fmt.Sprintf("user-%d", f.counter),
        Name:  fmt.Sprintf("User %d", f.counter),
        Email: fmt.Sprintf("user%d@example.com", f.counter),
    }
}

func (f *UserFactory) CreateMany(n int) []*User {
    users := make([]*User, n)
    for i := 0; i < n; i++ {
        users[i] = f.Create()
    }
    return users
}

// Uso
factory := &UserFactory{}
users := factory.CreateMany(10)
`)

	// ============================================
	// GOLDEN FILES
	// ============================================
	fmt.Println("\n--- Golden Files ---")
	fmt.Println(`
Golden files almacenan output esperado para comparación.

ESTRUCTURA:

tests/
└── testdata/
    └── golden/
        ├── user_output.golden
        ├── report_output.golden
        └── api_response.golden

GOLDEN FILE TEST:

import (
    "flag"
    "testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestGenerateReport(t *testing.T) {
    report := GenerateReport(data)

    goldenFile := "testdata/golden/report_output.golden"

    if *update {
        // Actualizar golden file
        err := os.WriteFile(goldenFile, []byte(report), 0644)
        require.NoError(t, err)
        return
    }

    // Comparar con golden file
    expected, err := os.ReadFile(goldenFile)
    require.NoError(t, err)

    assert.Equal(t, string(expected), report)
}

EJECUTAR:

go test                    # Comparar con golden
go test -update           # Actualizar golden files

JSON GOLDEN:

func TestAPIResponse(t *testing.T) {
    resp := callAPI()

    goldenFile := "testdata/golden/api_response.golden"

    actual, _ := json.MarshalIndent(resp, "", "  ")

    if *update {
        os.WriteFile(goldenFile, actual, 0644)
        return
    }

    expected, _ := os.ReadFile(goldenFile)
    assert.JSONEq(t, string(expected), string(actual))
}

SNAPSHOT TESTING:

import "github.com/bradleyjkemp/cupaloy"

func TestSnapshot(t *testing.T) {
    output := GenerateOutput(input)

    // Crea snapshot automáticamente
    cupaloy.SnapshotT(t, output)
}

// Actualizar:
go test . -update`)
	// ============================================
	// TEST CONTAINERS
	// ============================================
	fmt.Println("\n--- Test Containers ---")
	os.Stdout.WriteString(`
INSTALACIÓN:
go get github.com/testcontainers/testcontainers-go

Test Containers levanta contenedores reales para testing.

POSTGRES CONTAINER:

import (
    "context"
    "testing"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgresContainer(t *testing.T) *sql.DB {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)

    // Cleanup
    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    // Get connection string
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
    db, err := sql.Open("postgres", connStr)
    require.NoError(t, err)

    return db
}

REDIS CONTAINER:

func SetupRedisContainer(t *testing.T) *redis.Client {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "redis:7-alpine",
        ExposedPorts: []string{"6379/tcp"},
        WaitingFor:   wait.ForLog("Ready to accept connections"),
    }

    container, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })

    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "6379")

    client := redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%s", host, port.Port()),
    })

    return client
}

DOCKER COMPOSE:

func SetupStack(t *testing.T) {
    compose := testcontainers.NewLocalDockerCompose(
        []string{"docker-compose.test.yml"},
        "test",
    )

    err := compose.
        WithCommand([]string{"up", "-d"}).
        Invoke()
    require.NoError(t, err)

    t.Cleanup(func() {
        compose.Down()
    })
}

KAFKA CONTAINER:

func SetupKafkaContainer(t *testing.T) string {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "confluentinc/cp-kafka:latest",
        ExposedPorts: []string{"9093/tcp"},
        Env: map[string]string{
            "KAFKA_BROKER_ID": "1",
            "KAFKA_ZOOKEEPER_CONNECT": "zookeeper:2181",
            // ... más config
        },
    }

    container, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })

    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "9093")

    return fmt.Sprintf("%s:%s", host, port.Port())
}
`)

	// ============================================
	// COVERAGE ANALYSIS
	// ============================================
	fmt.Println("\n--- Coverage Analysis ---")
	fmt.Println(`
EJECUTAR CON COVERAGE:

go test -cover ./...
go test -coverprofile=coverage.out ./...

VER COVERAGE:

go tool cover -html=coverage.out

O en terminal:

go tool cover -func=coverage.out

OUTPUT:

github.com/myapp/user/handler.go:15:    GetUser         100.0%
github.com/myapp/user/handler.go:25:    CreateUser      85.7%
github.com/myapp/user/handler.go:40:    UpdateUser      66.7%
total:                                  (statements)    84.5%

COVERAGE POR PAQUETE:

go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

POR LÍNEA:

go test -covermode=count -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

MODE:

-covermode=set    # Binary: ejecutado o no
-covermode=count  # Cuenta ejecuciones
-covermode=atomic # Thread-safe count

CI/CD INTEGRATION:

# GitHub Actions
- name: Test with coverage
  run: go test -race -covermode=atomic -coverprofile=coverage.out ./...

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out

EXCLUDE FROM COVERAGE:

// Pragma para excluir líneas
func generated() {
    // coverage:ignore
    panic("auto-generated code")
}`)
	// Demostración conceptual
	demonstrateAdvancedTesting()
}

func demonstrateAdvancedTesting() {
	fmt.Printf("\n--- Demostración Conceptual ---\n")
	fmt.Printf("Frameworks de testing:\n")
	fmt.Printf("  testify: Assertions y suites\n")
	fmt.Printf("  gomock: Mocking robusto\n")
	fmt.Printf("  testcontainers: Contenedores reales\n")
	fmt.Printf("\nPatrones:\n")
	fmt.Printf("  - Unit tests: Mocks y asserts\n")
	fmt.Printf("  - Integration tests: DB y HTTP reales\n")
	fmt.Printf("  - Test fixtures: Builders y factories\n")
	fmt.Printf("  - Golden files: Snapshot testing\n")
	fmt.Printf("\nMejores prácticas:\n")
	fmt.Printf("  - Coverage > 80%%\n")
	fmt.Printf("  - Test suites para setup/teardown\n")
	fmt.Printf("  - Integration tests con tags\n")
	fmt.Printf("  - Test containers para deps reales\n")
}

/*
RESUMEN DE ADVANCED TESTING:

TESTIFY:
- assert: Continúa en fallos
- require: Para en primer fallo
- suite: Test suites con setup/teardown
- mock: Mocking framework
- github.com/stretchr/testify

ASSERT VS REQUIRE:
assert.Equal(t, expected, actual)  // Continúa
require.Equal(t, expected, actual) // Para

TEST SUITES:
- SetupSuite/TearDownSuite: Una vez
- SetupTest/TearDownTest: Cada test
- BeforeTest/AfterTest: Por test con nombre
- suite.Run(t, new(MyTestSuite))

TESTIFY MOCK:
- mock.Mock embedding
- m.Called(args...) en métodos
- On() para expectations
- AssertExpectations(t) para verificar

GOMOCK:
- Generado con mockgen
- EXPECT() para expectations
- Matchers: Any, Eq, Not, Nil
- Times, MinTimes, MaxTimes
- Do, DoAndReturn

INTEGRATION TESTING:
- Build tags: // +build integration
- Test helpers para setup
- httptest.Server para HTTP
- testcontainers para deps

TEST FIXTURES:
- Builder pattern
- Factory pattern
- JSON fixtures en testdata/
- Repeatable y mantenible

GOLDEN FILES:
- Snapshot testing
- -update flag para actualizar
- testdata/golden/
- Útil para output complejo

TEST CONTAINERS:
- Contenedores Docker reales
- testcontainers-go
- Postgres, Redis, Kafka, etc
- Cleanup automático
- CI-friendly

COVERAGE:
- go test -cover
- go test -coverprofile=coverage.out
- go tool cover -html=coverage.out
- Objetivo: >80%
- Covermode: set, count, atomic

MEJORES PRÁCTICAS:
1. Unit tests con mocks
2. Integration tests separados (tags)
3. Test suites para setup complejo
4. Builders para fixtures
5. Golden files para snapshots
6. Test containers para deps reales
7. Coverage >80%
8. Subtests con t.Run()
9. Table-driven tests
10. Parallel tests con t.Parallel()

ESTRUCTURA:
project/
├── internal/
│   └── user/
│       ├── handler.go
│       ├── handler_test.go        # Unit
│       └── repository_test.go     # Unit
├── tests/
│   ├── integration/
│   │   └── user_test.go          # Integration
│   └── testdata/
│       ├── fixtures/
│       └── golden/

COMANDOS:
go test ./...                      # Todos
go test -short ./...               # Solo unit
go test -tags=integration ./...    # Solo integration
go test -cover ./...               # Con coverage
go test -race ./...                # Race detector
go test -v ./...                   # Verbose
go test -run TestName ./...        # Test específico
go test -update                    # Actualizar goldens
*/

/* SUMMARY - CHAPTER 059: Advanced Testing
Topics covered in this chapter:

• testify framework: assert and require packages for expressive assertions
• Test suites: suite.Suite interface for setup/teardown and shared state
• Mocking with testify: mock.Mock for interface mocking, On(), Return(), AssertExpectations()
• gomock framework: generating mocks with mockgen, EXPECT() syntax, AnyTimes(), Times()
• Integration testing: testing.Short() flag, build tags for separating test types
• Test fixtures: builder pattern for test data, testdata directory conventions
• Golden files: storing expected output, -update flag for regenerating
• Test containers: testcontainers-go for real databases and dependencies
• Coverage: -cover flag, -coverprofile, >80% target, covermode options
• Test organization: unit tests alongside code, integration tests in tests/ directory
• Best practices: table-driven tests, subtests with t.Run(), t.Parallel() for concurrency
• Commands: go test flags for filtering, coverage, race detection, verbose output
*/
