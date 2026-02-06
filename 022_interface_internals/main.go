// Package main - Chapter 022: Interface Internals
// Deep dive into how interfaces work under the hood: memory layout,
// nil interfaces, type assertions, dynamic dispatch, and best practices.
package main

import (
	"fmt"
	"os"
)

// ============================================
// TYPES FOR EXAMPLES
// ============================================

type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return d.Name + " says woof!"
}

func (d Dog) String() string {
	return "Dog{" + d.Name + "}"
}

type Cat struct {
	Name string
}

func (c Cat) Speak() string {
	return c.Name + " says meow!"
}

func (c Cat) String() string {
	return "Cat{" + c.Name + "}"
}

type Fish struct {
	Name string
}

func (f Fish) String() string {
	return "Fish{" + f.Name + "}"
}

// ============================================
// INTERFACES
// ============================================

type Speaker interface {
	Speak() string
}

type Stringer interface {
	String() string
}

type SpeakStringer interface {
	Speak() string
	String() string
}

// ============================================
// ERROR TYPE FOR NIL DEMO
// ============================================

type MyError struct {
	Code    int
	Message string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

func findUser(id int) error {
	var err *MyError
	if id <= 0 {
		err = &MyError{Code: 404, Message: "user not found"}
	}
	// BUG: always returns non-nil interface!
	return err
}

func findUserFixed(id int) error {
	if id <= 0 {
		return &MyError{Code: 404, Message: "user not found"}
	}
	return nil
}

// ============================================
// SHAPE HIERARCHY FOR TYPE SWITCHES
// ============================================

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

type CircleShape struct {
	Radius float64
}

func (c CircleShape) Area() float64      { return 3.14159 * c.Radius * c.Radius }
func (c CircleShape) Perimeter() float64 { return 2 * 3.14159 * c.Radius }

type Triangle struct {
	A, B, C float64
	Height  float64
}

func (t Triangle) Area() float64      { return 0.5 * t.A * t.Height }
func (t Triangle) Perimeter() float64 { return t.A + t.B + t.C }

func describeShape(s Shape) string {
	switch v := s.(type) {
	case Rectangle:
		return fmt.Sprintf("Rectangle %.0fx%.0f", v.Width, v.Height)
	case CircleShape:
		return fmt.Sprintf("Circle r=%.0f", v.Radius)
	case Triangle:
		return fmt.Sprintf("Triangle sides=%.0f,%.0f,%.0f", v.A, v.B, v.C)
	default:
		return "Unknown shape"
	}
}

// ============================================
// INTERFACE COMPOSITION
// ============================================

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}

type Closer interface {
	Close() error
}

type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

func main() {
	fmt.Println("=== INTERFACE INTERNALS ===")

	// ============================================
	// INTERFACE MEMORY LAYOUT
	// ============================================
	fmt.Println("\n--- Interface Memory Layout ---")

	fmt.Println(`
  Go has two internal interface representations:

  1. iface - for interfaces WITH methods:
     +----------+----------+
     | tab *itab| data *T  |
     +----------+----------+
     itab contains:
       - Type information (concrete type)
       - Method table (function pointers)
       - Hash for fast type assertion

  2. eface - for empty interface (any/interface{}):
     +----------+----------+
     | _type    | data *T  |
     +----------+----------+
     Only stores type info and data pointer.
     No method table needed.

  Both are 16 bytes (two pointers) on 64-bit systems.

  IMPORTANT:
  - interface value = (type, value) PAIR
  - An interface is nil ONLY when BOTH type and value are nil
  - This is the source of the "nil interface" gotcha`)

	// ============================================
	// IFACE VS EFACE IN ACTION
	// ============================================
	fmt.Println("\n--- iface vs eface ---")

	var sp Speaker = Dog{Name: "Rex"}
	var a any = Dog{Name: "Buddy"}

	os.Stdout.WriteString(fmt.Sprintf("  Speaker (iface): type=%T, value=%v\n", sp, sp))
	os.Stdout.WriteString(fmt.Sprintf("  any (eface):     type=%T, value=%v\n", a, a))

	fmt.Println(`
  Speaker interface stores:
    itab -> {concrete type: Dog, methods: [Speak]}
    data -> pointer to Dog{"Rex"}

  any/interface{} stores:
    _type -> {concrete type: Dog}
    data  -> pointer to Dog{"Buddy"}`)

	// ============================================
	// NIL INTERFACE VS INTERFACE HOLDING NIL
	// ============================================
	fmt.Println("\n--- Nil Interface vs Interface Holding Nil ---")

	// Case 1: Truly nil interface
	var err1 error
	fmt.Println("  Nil interface:")
	os.Stdout.WriteString(fmt.Sprintf("    err1 == nil: %t\n", err1 == nil))
	os.Stdout.WriteString(fmt.Sprintf("    type: %T, value: %v\n", err1, err1))

	// Case 2: Interface holding nil concrete value
	var myErr *MyError // nil pointer
	var err2 error = myErr
	fmt.Println("\n  Interface holding nil pointer:")
	os.Stdout.WriteString(fmt.Sprintf("    myErr == nil: %t\n", myErr == nil))
	os.Stdout.WriteString(fmt.Sprintf("    err2 == nil:  %t  <-- SURPRISE!\n", err2 == nil))
	os.Stdout.WriteString(fmt.Sprintf("    type: %T, value: %v\n", err2, err2))

	fmt.Println(`
  WHY err2 != nil?

  Memory layout of err2:
    +-------------------+-------------------+
    | type = *MyError   | value = nil       |
    +-------------------+-------------------+

  For interface == nil, BOTH type and value must be nil.
  err2 has type = *MyError, so it's NOT nil!

  This is Go's BILLION-DOLLAR MISTAKE:
    var err *MyError   // nil pointer
    return err         // returns non-nil error interface!`)

	// Demonstrate the bug
	fmt.Println("\n  Demonstrating the bug:")
	err := findUser(1) // valid ID
	if err != nil {
		fmt.Println("    findUser(1): got unexpected error!", err)
	}

	errFixed := findUserFixed(1)
	if errFixed != nil {
		fmt.Println("    findUserFixed(1): got error")
	} else {
		fmt.Println("    findUserFixed(1): nil (correct!)")
	}

	fmt.Println(`
  FIX: Always return the interface type directly:

    // BAD - always returns non-nil error
    func findUser(id int) error {
        var err *MyError
        if id <= 0 { err = &MyError{...} }
        return err  // type is *MyError, even if nil!
    }

    // GOOD - return nil explicitly
    func findUser(id int) error {
        if id <= 0 {
            return &MyError{...}
        }
        return nil  // truly nil interface
    }`)

	// ============================================
	// TYPE ASSERTIONS
	// ============================================
	fmt.Println("\n--- Type Assertions ---")

	var animal Speaker = Dog{Name: "Rex"}

	// Single-value assertion (panics on failure)
	dog := animal.(Dog)
	fmt.Println("  Assertion succeeded:", dog.Name)

	// Two-value assertion (safe, returns ok bool)
	cat, ok := animal.(Cat)
	os.Stdout.WriteString(fmt.Sprintf("  animal.(Cat): value=%v, ok=%t\n", cat, ok))

	dog2, ok := animal.(Dog)
	os.Stdout.WriteString(fmt.Sprintf("  animal.(Dog): value=%v, ok=%t\n", dog2, ok))

	fmt.Println(`
  TYPE ASSERTION SYNTAX:

    value := iface.(ConcreteType)       // panics if wrong
    value, ok := iface.(ConcreteType)   // safe, ok=false if wrong

  TYPE ASSERTION TO INTERFACE:

    stringer, ok := iface.(Stringer)    // checks if implements Stringer

  PERFORMANCE:
    - Single type assertion: very fast (hash comparison)
    - Interface assertion: slightly slower (method set check)
    - Both are O(1) operations`)

	// Assert to interface
	stringer, ok := animal.(Stringer)
	if ok {
		fmt.Println("  Dog also implements Stringer:", stringer.String())
	}

	// Fish does not implement Speaker
	var s Stringer = Fish{Name: "Nemo"}
	_, ok = s.(Speaker)
	os.Stdout.WriteString(fmt.Sprintf("  Fish implements Speaker: %t\n", ok))

	// ============================================
	// TYPE SWITCHES
	// ============================================
	fmt.Println("\n--- Type Switches ---")

	shapes := []Shape{
		Rectangle{Width: 10, Height: 5},
		CircleShape{Radius: 7},
		Triangle{A: 3, B: 4, C: 5, Height: 4},
	}

	for _, s := range shapes {
		os.Stdout.WriteString(fmt.Sprintf("  %s: area=%.2f, perimeter=%.2f\n",
			describeShape(s), s.Area(), s.Perimeter()))
	}

	fmt.Println(`
  TYPE SWITCH SYNTAX:

    switch v := iface.(type) {
    case Type1:
        // v is Type1
    case Type2:
        // v is Type2
    case Type3, Type4:
        // v is the interface type (not narrowed)
    default:
        // no match
    }

  RULES:
  - Can only be used with interface values
  - Each case can list multiple types (comma-separated)
  - Default handles unmatched types
  - Variable v gets the concrete type in each case`)

	// ============================================
	// TYPE SWITCH ON any
	// ============================================
	fmt.Println("\n--- Type Switch on any ---")

	values := []any{42, "hello", 3.14, true, []int{1, 2, 3}, nil}

	for _, v := range values {
		describeAny(v)
	}

	// ============================================
	// DYNAMIC DISPATCH OVERHEAD
	// ============================================
	fmt.Println("\n--- Dynamic Dispatch ---")

	fmt.Println(`
  Interface method calls use DYNAMIC DISPATCH:

    // Direct call - compiler knows exact method
    dog.Speak()     // static dispatch, can be inlined

    // Interface call - resolved at runtime
    var s Speaker = dog
    s.Speak()       // dynamic dispatch via itab

  COST:
    - Interface call: ~2ns overhead (vtable lookup)
    - Direct call: can be inlined (0ns overhead)
    - For hot loops, consider avoiding interface calls
    - For most code, the overhead is negligible

  The compiler CAN devirtualize interface calls when it
  can prove the concrete type at compile time.`)

	// ============================================
	// COMPARABLE CONSTRAINT
	// ============================================
	fmt.Println("\n--- Comparable Types and Interfaces ---")

	fmt.Println(`
  COMPARABLE (Go 1.18+):

  The 'comparable' constraint allows == and !=:
    func Contains[T comparable](s []T, v T) bool {
        for _, x := range s {
            if x == v { return true }
        }
        return false
    }

  Types that satisfy comparable:
    - bool, numeric types, string
    - Pointers, channels
    - Structs where ALL fields are comparable
    - Arrays where element type is comparable

  Types that do NOT satisfy comparable:
    - Slices
    - Maps
    - Functions
    - Structs containing non-comparable fields

  INTERFACE COMPARISON:
    Two interface values are equal if:
    1. Both are nil, OR
    2. Both have identical dynamic type AND
       the dynamic values are equal

    If the dynamic type is NOT comparable
    (e.g., contains slice), comparing panics at runtime!`)

	// ============================================
	// INTERFACE COMPOSITION PATTERNS
	// ============================================
	fmt.Println("\n--- Interface Composition ---")

	fmt.Println(`
  COMPOSING INTERFACES:

    type Reader interface { Read([]byte) (int, error) }
    type Writer interface { Write([]byte) (int, error) }
    type Closer interface { Close() error }

    // Composed interfaces
    type ReadWriter interface {
        Reader
        Writer
    }

    type ReadWriteCloser interface {
        Reader
        Writer
        Closer
    }

  STDLIB EXAMPLES:
    io.Reader, io.Writer, io.Closer
    io.ReadWriter, io.ReadCloser, io.WriteCloser
    io.ReadWriteCloser

  RULE: Prefer small, focused interfaces
  that can be composed as needed.`)

	// ============================================
	// INTERFACE POLLUTION ANTI-PATTERN
	// ============================================
	fmt.Println("\n--- Interface Pollution (Anti-Pattern) ---")

	fmt.Println(`
  INTERFACE POLLUTION: defining interfaces you don't need.

  BAD - interface defined alongside implementation:

    type UserRepository interface {
        GetUser(id int) (*User, error)
        CreateUser(u *User) error
        UpdateUser(u *User) error
        DeleteUser(id int) error
    }

    type userRepo struct { db *sql.DB }
    func (r *userRepo) GetUser(id int) (*User, error) { ... }
    // ... all methods

  WHY IT'S BAD:
  - Only one implementation -> interface is useless overhead
  - Defined where implemented, not where consumed
  - Makes code harder to read

  GOOD - "Accept interfaces, return structs":

    // In the repository package - return concrete type
    type UserRepo struct { db *sql.DB }
    func NewUserRepo(db *sql.DB) *UserRepo { ... }
    func (r *UserRepo) GetUser(id int) (*User, error) { ... }

    // In the consumer package - define minimal interface
    type UserGetter interface {
        GetUser(id int) (*User, error)
    }

    func HandleGetUser(repo UserGetter) http.HandlerFunc {
        // Only declares what it NEEDS
    }

  Go's implicit interface satisfaction makes this powerful:
  *UserRepo automatically satisfies UserGetter!`)

	// ============================================
	// ACCEPT INTERFACES, RETURN STRUCTS
	// ============================================
	fmt.Println("\n--- Accept Interfaces, Return Structs ---")

	fmt.Println(`
  This Go proverb means:

  FUNCTION PARAMETERS: Accept interfaces
    func Process(r io.Reader) error {
        // Works with *os.File, *bytes.Buffer, net.Conn, etc.
    }

  RETURN VALUES: Return concrete types
    func NewBuffer() *bytes.Buffer {
        return &bytes.Buffer{}
    }

  WHY?
  - Interfaces as params: maximum flexibility for callers
  - Concrete returns: no unnecessary abstraction
  - Callers can use the concrete type or assign to interface
  - Adding methods to return type is not breaking change
  - Adding methods to interface param IS breaking change

  EXCEPTION: Return interface when you need to hide implementation
    func Open(name string) (io.ReadCloser, error)
    // Hides whether it's file, network, in-memory, etc.`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  INTERFACE INTERNALS CHECKLIST:

  1. Interface = (type, value) pair - 16 bytes
  2. iface for methods, eface for empty interface (any)
  3. nil interface: BOTH type AND value are nil
  4. Interface holding nil != nil interface (common bug!)
  5. Type assertions: use comma-ok to avoid panics
  6. Type switches: cleanly handle multiple concrete types
  7. Dynamic dispatch: ~2ns overhead, usually negligible
  8. comparable: not all types support == in interfaces
  9. Compose small interfaces, don't define large ones
  10. Accept interfaces, return structs`)
}

func describeAny(v any) {
	switch v := v.(type) {
	case int:
		os.Stdout.WriteString(fmt.Sprintf("  int: %d\n", v))
	case string:
		os.Stdout.WriteString(fmt.Sprintf("  string: %q (len=%d)\n", v, len(v)))
	case float64:
		os.Stdout.WriteString(fmt.Sprintf("  float64: %.2f\n", v))
	case bool:
		os.Stdout.WriteString(fmt.Sprintf("  bool: %t\n", v))
	case []int:
		os.Stdout.WriteString(fmt.Sprintf("  []int: %v (len=%d)\n", v, len(v)))
	case nil:
		fmt.Println("  nil value")
	default:
		os.Stdout.WriteString(fmt.Sprintf("  unknown: %T = %v\n", v, v))
	}
}

