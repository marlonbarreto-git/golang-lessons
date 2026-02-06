// Package main - Chapter 036: Testing
// Go tiene soporte integrado para testing. Aprenderás unit tests,
// table-driven tests, subtests, helpers, y mejores prácticas.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== TESTING EN GO ===")
	fmt.Println(`
Este capítulo se demuestra mejor con archivos _test.go.
Revisa el archivo main_test.go en este directorio.

COMANDOS DE TESTING:
  go test                    # ejecutar tests del paquete actual
  go test ./...              # ejecutar tests de todos los paquetes
  go test -v                 # verbose output
  go test -run TestNombre    # ejecutar tests que matchean el patrón
  go test -cover             # mostrar coverage
  go test -coverprofile=c.out && go tool cover -html=c.out  # coverage visual
  go test -race              # detectar race conditions
  go test -bench .           # ejecutar benchmarks
  go test -count=1           # deshabilitar cache

ESTRUCTURA DE ARCHIVOS:
  mypackage/
    calculator.go      # código
    calculator_test.go # tests

CONVENCIONES:
  - Archivos: *_test.go
  - Funciones: TestXxx(t *testing.T)
  - El nombre debe empezar con mayúscula después de Test`)

	// Demostrar las funciones que vamos a testear
	fmt.Println("\n--- Funciones a Testear ---")
	fmt.Printf("Add(2, 3) = %d\n", Add(2, 3))
	fmt.Printf("Multiply(4, 5) = %d\n", Multiply(4, 5))
	fmt.Printf("Divide(10, 2) = %.2f, err = nil\n", func() float64 { v, _ := Divide(10, 2); return v }())
	fmt.Printf("IsPrime(7) = %t\n", IsPrime(7))
	fmt.Printf("Factorial(5) = %d\n", Factorial(5))
}

// ============================================
// FUNCIONES A TESTEAR
// ============================================

func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("división por cero")
	}
	return a / b, nil
}

func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// Fibonacci para benchmarks
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

// Usuario para tests con structs
type User struct {
	ID       int
	Name     string
	Email    string
	IsActive bool
}

func NewUser(id int, name, email string) *User {
	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		IsActive: true,
	}
}

func (u *User) Deactivate() {
	u.IsActive = false
}

// Interface para mocking
type UserRepository interface {
	FindByID(id int) (*User, error)
	Save(user *User) error
}

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
	user := NewUser(0, name, email)
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}
