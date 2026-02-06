// Package main - Chapter 038: Mocking
// El mocking es esencial para escribir tests unitarios aislados.
// Go usa interfaces para facilitar el mocking sin frameworks externos.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== MOCKING EN GO ===")

	// ============================================
	// INTERFACES PARA MOCKING
	// ============================================
	fmt.Println("\n--- Interfaces para Mocking ---")
	fmt.Println(`
El secreto del mocking en Go son las interfaces.
Diseña tu código para depender de interfaces, no de implementaciones concretas.

// MAL: dependencia concreta (difícil de testear)
type UserService struct {
    db *sql.DB  // Dependencia concreta
}

// BIEN: dependencia de interfaz (fácil de mockear)
type UserRepository interface {
    FindByID(id int) (*User, error)
    Save(user *User) error
}

type UserService struct {
    repo UserRepository  // Dependencia de interfaz
}`)
	// ============================================
	// EJEMPLO: USER SERVICE CON MOCK
	// ============================================
	fmt.Println("\n--- Ejemplo: User Service ---")

	// Mock del repository
	mockRepo := &MockUserRepository{
		users: map[int]*User{
			1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
			2: {ID: 2, Name: "Bob", Email: "bob@example.com"},
		},
	}

	service := NewUserService(mockRepo)

	// Usar el servicio con mock
	user, err := service.GetUser(1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("User found: %+v\n", user)
	}

	// Crear usuario
	newUser, err := service.CreateUser("Charlie", "charlie@example.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("User created: %+v\n", newUser)
	}

	// ============================================
	// MOCK CON COMPORTAMIENTO CONFIGURABLE
	// ============================================
	fmt.Println("\n--- Mock Configurable ---")

	// Mock que puede simular errores
	failingRepo := &MockUserRepository{
		users:     make(map[int]*User),
		saveError: errors.New("database connection failed"),
	}

	failingService := NewUserService(failingRepo)
	_, err = failingService.CreateUser("Dave", "dave@example.com")
	fmt.Printf("Expected error: %v\n", err)

	// ============================================
	// HTTP CLIENT MOCK
	// ============================================
	fmt.Println("\n--- HTTP Client Mock ---")

	// Mock para cliente HTTP
	mockHTTP := &MockHTTPClient{
		responses: map[string]string{
			"https://api.example.com/users/1": `{"id":1,"name":"Alice"}`,
		},
	}

	apiClient := &APIClient{client: mockHTTP}
	data, err := apiClient.FetchUser(1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("API response: %s\n", data)
	}

	// ============================================
	// TIME MOCK
	// ============================================
	fmt.Println("\n--- Time Mock ---")

	// Mock para funciones que dependen del tiempo
	fixedTime := time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC)
	mockClock := &MockClock{now: fixedTime}

	scheduler := &Scheduler{clock: mockClock}
	nextRun := scheduler.NextRun(time.Hour)
	fmt.Printf("Fixed now: %v\n", fixedTime)
	fmt.Printf("Next run: %v\n", nextRun)

	// ============================================
	// MOCK CON ASSERTIONS
	// ============================================
	fmt.Println("\n--- Mock con Assertions ---")

	spyRepo := &SpyUserRepository{}
	spyService := NewUserService(spyRepo)

	_, _ = spyService.GetUser(1)
	_, _ = spyService.GetUser(2)
	_, _ = spyService.CreateUser("Test", "test@example.com")

	fmt.Printf("FindByID calls: %d\n", spyRepo.findByIDCalls)
	fmt.Printf("Save calls: %d\n", spyRepo.saveCalls)

	// ============================================
	// FAKE IMPLEMENTATION
	// ============================================
	fmt.Println("\n--- Fake Implementation ---")
	fmt.Println(`
Un Fake es una implementación funcional simplificada:
- InMemoryDatabase en lugar de PostgreSQL
- LocalFileStorage en lugar de S3
- Útil para tests de integración

type InMemoryCache struct {
    data map[string][]byte
    mu   sync.RWMutex
}

func (c *InMemoryCache) Get(key string) ([]byte, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    if v, ok := c.data[key]; ok {
        return v, nil
    }
    return nil, ErrNotFound
}

func (c *InMemoryCache) Set(key string, value []byte) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
    return nil
}`)
	// ============================================
	// TESTIFY/MOCK (FRAMEWORK POPULAR)
	// ============================================
	fmt.Println("\n--- Testify/Mock ---")
	fmt.Println(`
github.com/stretchr/testify/mock es popular para mocking:

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(id int) (*User, error) {
    args := m.Called(id)
    return args.Get(0).(*User), args.Error(1)
}

// En el test:
func TestGetUser(t *testing.T) {
    mockRepo := new(MockRepository)
    mockRepo.On("FindByID", 1).Return(&User{ID: 1, Name: "Alice"}, nil)

    service := NewUserService(mockRepo)
    user, err := service.GetUser(1)

    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
    mockRepo.AssertExpectations(t)
}`)
	// ============================================
	// MOCKGEN (GENERACIÓN AUTOMÁTICA)
	// ============================================
	fmt.Println("\n--- Mockgen ---")
	fmt.Println(`
go install github.com/golang/mock/mockgen@latest

Generar mocks automáticamente:

//go:generate mockgen -source=repository.go -destination=mock_repository.go -package=main

O por interfaz:
mockgen -package=mocks -destination=mocks/user_repo.go mypackage UserRepository

Uso en tests:
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockRepo := NewMockUserRepository(ctrl)
mockRepo.EXPECT().FindByID(1).Return(&User{ID: 1}, nil)`)
	// ============================================
	// PATTERNS
	// ============================================
	fmt.Println("\n--- Patrones de Mocking ---")
	fmt.Println(`
1. CONSTRUCTOR INJECTION:
   func NewService(repo Repository) *Service {
       return &Service{repo: repo}
   }

2. INTERFACE SEGREGATION:
   // Interfaces pequeñas son más fáciles de mockear
   type Reader interface { Read(id int) (*Item, error) }
   type Writer interface { Write(item *Item) error }

3. FUNCTIONAL OPTIONS:
   func NewClient(opts ...Option) *Client {
       c := &Client{httpClient: http.DefaultClient}
       for _, opt := range opts {
           opt(c)
       }
       return c
   }

   func WithHTTPClient(hc HTTPClient) Option {
       return func(c *Client) { c.httpClient = hc }
   }

4. CLOCK ABSTRACTION:
   type Clock interface { Now() time.Time }

   type RealClock struct{}
   func (RealClock) Now() time.Time { return time.Now() }

   type MockClock struct { now time.Time }
   func (m MockClock) Now() time.Time { return m.now }`)}

// ============================================
// DOMAIN TYPES
// ============================================

type User struct {
	ID    int
	Name  string
	Email string
}

// ============================================
// INTERFACES
// ============================================

type UserRepository interface {
	FindByID(id int) (*User, error)
	Save(user *User) error
}

type HTTPClient interface {
	Get(url string) ([]byte, error)
}

type Clock interface {
	Now() time.Time
}

// ============================================
// SERVICES
// ============================================

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) CreateUser(name, email string) (*User, error) {
	user := &User{Name: name, Email: email}
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

type APIClient struct {
	client HTTPClient
}

func (c *APIClient) FetchUser(id int) ([]byte, error) {
	url := fmt.Sprintf("https://api.example.com/users/%d", id)
	return c.client.Get(url)
}

type Scheduler struct {
	clock Clock
}

func (s *Scheduler) NextRun(interval time.Duration) time.Time {
	return s.clock.Now().Add(interval)
}

// ============================================
// MOCKS
// ============================================

type MockUserRepository struct {
	users     map[int]*User
	nextID    int
	saveError error
}

func (m *MockUserRepository) FindByID(id int) (*User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) Save(user *User) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.nextID++
	user.ID = m.nextID
	if m.users == nil {
		m.users = make(map[int]*User)
	}
	m.users[user.ID] = user
	return nil
}

type MockHTTPClient struct {
	responses map[string]string
	err       error
}

func (m *MockHTTPClient) Get(url string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	if resp, ok := m.responses[url]; ok {
		return []byte(resp), nil
	}
	return nil, errors.New("not found")
}

type MockClock struct {
	now time.Time
}

func (m *MockClock) Now() time.Time {
	return m.now
}

// Spy: mock que registra llamadas
type SpyUserRepository struct {
	users         map[int]*User
	findByIDCalls int
	saveCalls     int
}

func (s *SpyUserRepository) FindByID(id int) (*User, error) {
	s.findByIDCalls++
	if s.users == nil {
		return nil, errors.New("not found")
	}
	if user, ok := s.users[id]; ok {
		return user, nil
	}
	return nil, errors.New("not found")
}

func (s *SpyUserRepository) Save(user *User) error {
	s.saveCalls++
	return nil
}

// ContextMock para testing con context
type ContextMock struct {
	context.Context
	deadline time.Time
	done     chan struct{}
	err      error
}

func NewMockContext() *ContextMock {
	return &ContextMock{
		done: make(chan struct{}),
	}
}

func (m *ContextMock) Deadline() (time.Time, bool) {
	return m.deadline, !m.deadline.IsZero()
}

func (m *ContextMock) Done() <-chan struct{} {
	return m.done
}

func (m *ContextMock) Err() error {
	return m.err
}

func (m *ContextMock) Cancel() {
	close(m.done)
	m.err = context.Canceled
}

/*
RESUMEN DE MOCKING:

TIPOS DE TEST DOUBLES:
- Mock: objeto pre-programado con expectativas
- Stub: objeto con respuestas fijas
- Spy: mock que registra llamadas
- Fake: implementación funcional simplificada

PRINCIPIOS:
1. Diseñar para interfaces, no implementaciones
2. Inyectar dependencias
3. Interfaces pequeñas (segregación)
4. Abstraer dependencias externas (time, HTTP, DB)

HERRAMIENTAS:
- Manual: crear mocks manualmente (recomendado)
- testify/mock: framework popular
- gomock/mockgen: generación automática

BUENAS PRÁCTICAS:
1. Mockear solo en boundaries (DB, HTTP, tiempo)
2. No mockear todo - some things should be real
3. Tests deben probar comportamiento, no implementación
4. Spy para verificar interacciones críticas
5. Fake para tests de integración

PATRONES:
- Constructor injection
- Interface segregation
- Functional options
- Clock abstraction
*/
