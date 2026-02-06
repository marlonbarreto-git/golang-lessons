package main

import (
	"errors"
	"fmt"
	"testing"
)

// ============================================
// TEST BÁSICO
// ============================================

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	expected := 5

	if result != expected {
		t.Errorf("Add(2, 3) = %d; want %d", result, expected)
	}
}

// ============================================
// TABLE-DRIVEN TESTS (PATRÓN MÁS COMÚN)
// ============================================

func TestAddTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
		{"mixed signs", -2, 3, 1},
		{"zeros", 0, 0, 0},
		{"with zero", 5, 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ============================================
// TEST CON ERROR
// ============================================

func TestDivide(t *testing.T) {
	tests := []struct {
		name      string
		a, b      float64
		expected  float64
		expectErr bool
	}{
		{"normal division", 10, 2, 5, false},
		{"division by zero", 10, 0, 0, true},
		{"float division", 7, 2, 3.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Divide(tt.a, tt.b)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Divide(%v, %v) = %v; want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ============================================
// TEST CON SUBTESTS
// ============================================

func TestIsPrime(t *testing.T) {
	t.Run("primes", func(t *testing.T) {
		primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23}
		for _, p := range primes {
			if !IsPrime(p) {
				t.Errorf("IsPrime(%d) = false; want true", p)
			}
		}
	})

	t.Run("non-primes", func(t *testing.T) {
		nonPrimes := []int{0, 1, 4, 6, 8, 9, 10, 12}
		for _, n := range nonPrimes {
			if IsPrime(n) {
				t.Errorf("IsPrime(%d) = true; want false", n)
			}
		}
	})

	t.Run("negative numbers", func(t *testing.T) {
		for _, n := range []int{-1, -5, -10} {
			if IsPrime(n) {
				t.Errorf("IsPrime(%d) = true; want false", n)
			}
		}
	})
}

// ============================================
// TEST HELPER FUNCTIONS
// ============================================

func TestWithHelper(t *testing.T) {
	// Helper para comparar usuarios
	assertUserEqual := func(t testing.TB, got, want *User) {
		t.Helper() // marca como helper para mejor stack traces
		if got.ID != want.ID {
			t.Errorf("ID = %d; want %d", got.ID, want.ID)
		}
		if got.Name != want.Name {
			t.Errorf("Name = %s; want %s", got.Name, want.Name)
		}
		if got.Email != want.Email {
			t.Errorf("Email = %s; want %s", got.Email, want.Email)
		}
	}

	t.Run("create user", func(t *testing.T) {
		got := NewUser(1, "Alice", "alice@example.com")
		want := &User{ID: 1, Name: "Alice", Email: "alice@example.com", IsActive: true}
		assertUserEqual(t, got, want)
	})
}

// ============================================
// TEST SETUP Y TEARDOWN
// ============================================

func TestWithSetup(t *testing.T) {
	// Setup
	t.Log("Setting up test...")

	// Cleanup se ejecuta al final del test
	t.Cleanup(func() {
		t.Log("Cleaning up...")
	})

	// Test code
	result := Add(1, 1)
	if result != 2 {
		t.Error("setup test failed")
	}
}

// ============================================
// TEST SKIP
// ============================================

func TestSkipExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Test largo aquí...
	t.Log("Running long test...")
}

// ============================================
// TEST PARALLEL
// ============================================

func TestParallel(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"factorial of 0", 0, 1},
		{"factorial of 1", 1, 1},
		{"factorial of 5", 5, 120},
		{"factorial of 10", 10, 3628800},
	}

	for _, tt := range tests {
		tt := tt // capturar variable para closure
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // ejecutar en paralelo

			result := Factorial(tt.input)
			if result != tt.expected {
				t.Errorf("Factorial(%d) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================
// MOCKING CON INTERFACES
// ============================================

// Mock del repository
type MockUserRepository struct {
	users     map[int]*User
	saveError error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[int]*User),
	}
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
	user.ID = len(m.users) + 1
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) SetSaveError(err error) {
	m.saveError = err
}

func TestUserService(t *testing.T) {
	t.Run("get existing user", func(t *testing.T) {
		repo := NewMockUserRepository()
		repo.users[1] = &User{ID: 1, Name: "Alice"}
		service := NewUserService(repo)

		user, err := service.GetUser(1)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("Name = %s; want Alice", user.Name)
		}
	})

	t.Run("get non-existing user", func(t *testing.T) {
		repo := NewMockUserRepository()
		service := NewUserService(repo)

		_, err := service.GetUser(999)

		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("create user success", func(t *testing.T) {
		repo := NewMockUserRepository()
		service := NewUserService(repo)

		user, err := service.CreateUser("Bob", "bob@example.com")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Name != "Bob" {
			t.Errorf("Name = %s; want Bob", user.Name)
		}
	})

	t.Run("create user error", func(t *testing.T) {
		repo := NewMockUserRepository()
		repo.SetSaveError(errors.New("database error"))
		service := NewUserService(repo)

		_, err := service.CreateUser("Bob", "bob@example.com")

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// ============================================
// BENCHMARK
// ============================================

func BenchmarkFibonacci10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fibonacci(10)
	}
}

func BenchmarkFibonacci20(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fibonacci(20)
	}
}

// Benchmark con b.Loop() (Go 1.24+)
func BenchmarkAdd(b *testing.B) {
	for b.Loop() {
		Add(2, 3)
	}
}

// Benchmark con sub-benchmarks
func BenchmarkIsPrime(b *testing.B) {
	cases := []struct {
		name string
		n    int
	}{
		{"small", 17},
		{"medium", 7919},
		{"large", 104729},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				IsPrime(tc.n)
			}
		})
	}
}

// ============================================
// FUZZING (Go 1.18+)
// ============================================

func FuzzAdd(f *testing.F) {
	// Seed corpus
	f.Add(1, 2)
	f.Add(0, 0)
	f.Add(-1, 1)
	f.Add(100, -100)

	f.Fuzz(func(t *testing.T, a, b int) {
		result := Add(a, b)
		// Propiedad: Add es conmutativa
		if result != Add(b, a) {
			t.Errorf("Add is not commutative: Add(%d, %d) != Add(%d, %d)", a, b, b, a)
		}
		// Propiedad: identidad con 0
		if Add(a, 0) != a {
			t.Errorf("Add(%d, 0) != %d", a, a)
		}
	})
}

// ============================================
// EXAMPLE TESTS (documentación ejecutable)
// ============================================

func ExampleAdd() {
	result := Add(2, 3)
	fmt.Println(result)
	// Output: 5
}

func ExampleFactorial() {
	fmt.Println(Factorial(0))
	fmt.Println(Factorial(1))
	fmt.Println(Factorial(5))
	// Output:
	// 1
	// 1
	// 120
}

/*
RESUMEN:

TIPOS DE TEST:
- Unit tests: TestXxx(t *testing.T)
- Benchmarks: BenchmarkXxx(b *testing.B)
- Fuzz tests: FuzzXxx(f *testing.F)
- Examples: ExampleXxx() con // Output: comment

MÉTODOS DE testing.T:
- t.Error/Errorf: reportar error, continuar
- t.Fatal/Fatalf: reportar error, terminar test
- t.Log/Logf: logging (solo con -v)
- t.Skip/Skipf: saltar test
- t.Run: subtests
- t.Parallel: ejecutar en paralelo
- t.Helper: marcar función como helper
- t.Cleanup: registrar función de limpieza

MÉTODOS DE testing.B:
- b.N: número de iteraciones
- b.Loop(): nuevo loop style (Go 1.24+)
- b.ResetTimer: reiniciar timer
- b.StopTimer/StartTimer: pausar timer
- b.ReportAllocs: reportar allocations

BUENAS PRÁCTICAS:

1. Table-driven tests para múltiples casos
2. t.Run() para subtests legibles
3. t.Helper() para funciones auxiliares
4. t.Parallel() para tests independientes
5. Interfaces para mocking
6. t.Cleanup() en lugar de defer
7. Nombres descriptivos para casos de test
*/
