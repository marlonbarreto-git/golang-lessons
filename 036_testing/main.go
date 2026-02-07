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

// StringReverse para testing
func StringReverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ============================================
// TESTING GUIDE - COMPREHENSIVE REFERENCE
// ============================================

/*
SUBTESTS (t.Run):

func TestAdd(t *testing.T) {
    t.Run("positive numbers", func(t *testing.T) {
        if got := Add(2, 3); got != 5 {
            t.Errorf("Add(2,3) = %d, want 5", got)
        }
    })
    t.Run("negative numbers", func(t *testing.T) {
        if got := Add(-1, -1); got != -2 {
            t.Errorf("Add(-1,-1) = %d, want -2", got)
        }
    })
    t.Run("zero", func(t *testing.T) {
        if got := Add(0, 0); got != 0 {
            t.Errorf("Add(0,0) = %d, want 0", got)
        }
    })
}

Run specific subtest:
  go test -run TestAdd/positive_numbers

---

TABLE-DRIVEN TESTS:

func TestIsPrime(t *testing.T) {
    tests := []struct {
        name string
        n    int
        want bool
    }{
        {"negative", -1, false},
        {"zero", 0, false},
        {"one", 1, false},
        {"two", 2, true},
        {"three", 3, true},
        {"four", 4, false},
        {"large prime", 997, true},
        {"large non-prime", 1000, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := IsPrime(tt.n); got != tt.want {
                t.Errorf("IsPrime(%d) = %v, want %v", tt.n, got, tt.want)
            }
        })
    }
}

---

TEST HELPERS:

func TestUserCreation(t *testing.T) {
    user := createTestUser(t, "Alice", "alice@test.com")
    if user.Name != "Alice" {
        t.Errorf("expected Alice, got %s", user.Name)
    }
}

func createTestUser(t *testing.T, name, email string) *User {
    t.Helper() // marks this as helper - errors report caller's line
    user := NewUser(1, name, email)
    if user == nil {
        t.Fatal("NewUser returned nil")
    }
    return user
}

// Multiple levels of helpers
func assertEqual(t *testing.T, got, want int) {
    t.Helper()
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

---

t.Cleanup:

func TestWithTempFile(t *testing.T) {
    f, err := os.CreateTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() {
        f.Close()
        os.Remove(f.Name())
    })
    // Cleanup runs after test completes, even on failure
    // Multiple Cleanup calls run in LIFO order

    f.WriteString("test data")
    // ... use the file
}

func TestWithDB(t *testing.T) {
    db := setupTestDB(t) // helper that registers cleanup
    // db is automatically cleaned up after test
    _ = db
}

func setupTestDB(t *testing.T) string {
    t.Helper()
    dbPath := t.TempDir() + "/test.db"
    // t.TempDir() auto-cleans up!
    return dbPath
}

---

t.Parallel:

func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input int
        want  int
    }{
        {"case1", 2, 4},
        {"case2", 3, 9},
        {"case3", 4, 16},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // run subtests in parallel
            got := tt.input * tt.input
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}

// IMPORTANT: t.Parallel() pauses the test until its parent
// test function returns, then runs in parallel with other
// parallel subtests. The parent runs non-parallel subtests
// sequentially first.

---

testing.Short():

func TestExpensiveOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping expensive test in short mode")
    }
    // Run expensive test only in full mode
    // Enable short mode: go test -short
}

func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    // ... integration test logic
}

---

TEST FIXTURES PATTERN:

// testdata/ directory is ignored by go build
// Use it for test fixtures:
//
// mypackage/
// ├── parser.go
// ├── parser_test.go
// └── testdata/
//     ├── input1.json
//     ├── expected1.json
//     └── complex_case.json

func TestParser(t *testing.T) {
    input, err := os.ReadFile("testdata/input1.json")
    if err != nil {
        t.Fatal(err)
    }
    expected, err := os.ReadFile("testdata/expected1.json")
    if err != nil {
        t.Fatal(err)
    }

    got := Parse(input)
    if !bytes.Equal(got, expected) {
        t.Errorf("Parse mismatch:\ngot:  %s\nwant: %s", got, expected)
    }
}

---

GOLDEN FILES PATTERN:

// Golden files auto-update expected output with -update flag

var update = flag.Bool("update", false, "update golden files")

func TestOutput(t *testing.T) {
    got := generateOutput()
    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {
        os.MkdirAll("testdata", 0755)
        os.WriteFile(golden, got, 0644)
        return
    }

    expected, err := os.ReadFile(golden)
    if err != nil {
        t.Fatal(err)
    }
    if !bytes.Equal(got, expected) {
        t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, expected)
    }
}

// Usage:
//   go test -run TestOutput -update   # regenerate golden files
//   go test -run TestOutput           # verify against golden files

---

t.TempDir():

func TestFileProcessing(t *testing.T) {
    // t.TempDir() creates a temp dir that is automatically
    // removed when the test and all its subtests complete
    dir := t.TempDir()
    path := filepath.Join(dir, "test.txt")
    os.WriteFile(path, []byte("content"), 0644)
    // ... test using the file
    // No cleanup needed!
}

---

TESTING.T METHODS REFERENCE:

t.Log(args...)        // log a message (shown with -v)
t.Logf(format, ...)   // formatted log
t.Error(args...)      // log error, continue test
t.Errorf(format, ...) // formatted error, continue
t.Fatal(args...)      // log error, stop test NOW
t.Fatalf(format, ...) // formatted fatal
t.Fail()              // mark failed, continue
t.FailNow()           // mark failed, stop NOW
t.Skip(args...)       // skip this test
t.Skipf(format, ...)  // formatted skip
t.Run(name, func)     // run subtest
t.Parallel()          // mark as parallel
t.Helper()            // mark as test helper
t.Cleanup(func)       // register cleanup
t.TempDir()           // get auto-cleaned temp dir
t.Setenv(key, val)    // set env var (auto restored)
t.Name()              // current test name
t.Deadline()          // test timeout deadline
*/

/*
SUMMARY - Chapter 036: Testing in Go

TESTING BASICS:
- Files must end in _test.go
- Functions must start with TestXxx(t *testing.T)
- Run with: go test, go test -v, go test -run Pattern
- Coverage: go test -cover, go test -coverprofile=c.out

TABLE-DRIVEN TESTS:
- Define test cases as slice of structs
- Use t.Run for named subtests
- Enables targeted test execution with -run flag

SUBTESTS (t.Run):
- Organize tests hierarchically
- Run specific subtests: go test -run TestX/subtest
- Enable parallel execution within a test

TEST HELPERS:
- Use t.Helper() to mark helper functions
- Errors report caller line instead of helper line
- Use t.Fatal for setup failures

CLEANUP AND FIXTURES:
- t.Cleanup registers cleanup functions (LIFO order)
- t.TempDir creates auto-cleaned temp directories
- t.Setenv sets and auto-restores env vars
- testdata/ directory for test fixtures

PARALLEL TESTING:
- t.Parallel marks subtests for parallel execution
- Useful for independent, slow tests
- Be careful with shared state

TESTING.SHORT:
- testing.Short() checks if -short flag is set
- Use t.Skip to skip expensive tests in short mode

GOLDEN FILES PATTERN:
- Store expected output in testdata/*.golden files
- Use -update flag to regenerate expected output
- Compare actual output against golden files

TESTING.T METHODS:
- t.Error/Errorf: report failure, continue
- t.Fatal/Fatalf: report failure, stop immediately
- t.Skip/Skipf: skip test with reason
- t.Log/Logf: log messages (visible with -v)
- t.Run: run subtests
- t.Parallel: enable parallel execution
- t.Name: get current test name
- t.Deadline: get test timeout
*/
