// Package main - Chapter 021: Method Sets and Receivers
// Understanding value receivers vs pointer receivers, method sets,
// and how they affect interface satisfaction in Go.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// ============================================
// BASIC TYPES FOR EXAMPLES
// ============================================

type Point struct {
	X, Y float64
}

type Circle struct {
	Center Point
	Radius float64
}

// ============================================
// VALUE RECEIVERS
// ============================================

func (p Point) Distance(other Point) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (p Point) String() string {
	return fmt.Sprintf("(%g, %g)", p.X, p.Y)
}

// ============================================
// POINTER RECEIVERS
// ============================================

func (p *Point) Translate(dx, dy float64) {
	p.X += dx
	p.Y += dy
}

func (p *Point) Scale(factor float64) {
	p.X *= factor
	p.Y *= factor
}

// ============================================
// CIRCLE METHODS - CONSISTENCY RULE
// ============================================

func (c *Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c *Circle) Grow(factor float64) {
	c.Radius *= factor
}

func (c *Circle) Move(dx, dy float64) {
	c.Center.Translate(dx, dy)
}

func (c *Circle) String() string {
	return fmt.Sprintf("Circle{center=%s, radius=%g}", c.Center.String(), c.Radius)
}

// ============================================
// INTERFACES FOR METHOD SET DEMO
// ============================================

type Stringer interface {
	String() string
}

type Mover interface {
	Translate(dx, dy float64)
}

type StringMover interface {
	String() string
	Translate(dx, dy float64)
}

// ============================================
// NIL RECEIVER HANDLING
// ============================================

type List struct {
	Value int
	Next  *List
}

func (l *List) Len() int {
	if l == nil {
		return 0
	}
	return 1 + l.Next.Len()
}

func (l *List) String() string {
	if l == nil {
		return "nil"
	}
	parts := []string{}
	for node := l; node != nil; node = node.Next {
		parts = append(parts, fmt.Sprintf("%d", node.Value))
	}
	return "[" + strings.Join(parts, " -> ") + "]"
}

// ============================================
// EMBEDDING AND METHOD PROMOTION
// ============================================

type Animal struct {
	Name string
}

func (a Animal) Speak() string {
	return a.Name + " makes a sound"
}

func (a *Animal) Rename(name string) {
	a.Name = name
}

type Dog struct {
	Animal
	Breed string
}

func (d Dog) Fetch() string {
	return d.Name + " fetches the ball!"
}

// Overriding promoted method
type Cat struct {
	Animal
}

func (c Cat) Speak() string {
	return c.Name + " says meow"
}

// ============================================
// PROMOTED METHOD SETS - INTERFACE DEMO
// ============================================

type Speaker interface {
	Speak() string
}

type Renamer interface {
	Rename(name string)
}

type SpeakRenamer interface {
	Speak() string
	Rename(name string)
}

// ============================================
// PERFORMANCE: VALUE VS POINTER
// ============================================

type SmallStruct struct {
	X, Y int
}

type LargeStruct struct {
	Data [1024]byte
	Name string
	Tags [10]string
}

func (s SmallStruct) Sum() int {
	return s.X + s.Y
}

func (l *LargeStruct) FirstByte() byte {
	return l.Data[0]
}

func main() {
	fmt.Println("=== METHOD SETS AND RECEIVERS ===")

	// ============================================
	// VALUE RECEIVERS VS POINTER RECEIVERS
	// ============================================
	fmt.Println("\n--- Value Receivers vs Pointer Receivers ---")

	p := Point{3, 4}
	origin := Point{0, 0}

	fmt.Println("Value receiver - does NOT modify original:")
	os.Stdout.WriteString(fmt.Sprintf("  p = %s, distance to origin = %.2f\n", p.String(), p.Distance(origin)))

	fmt.Println("\nPointer receiver - MODIFIES original:")
	fmt.Println("  Before:", p.String())
	p.Translate(1, 1)
	fmt.Println("  After Translate(1,1):", p.String())
	p.Scale(2)
	fmt.Println("  After Scale(2):", p.String())

	fmt.Println(`
RULES FOR CHOOSING RECEIVER TYPE:

  Use POINTER receiver (*T) when:
  - Method needs to MODIFY the receiver
  - Struct is large (avoids copy overhead)
  - Consistency: if any method uses pointer, all should

  Use VALUE receiver (T) when:
  - Method only READS data (no mutation)
  - Struct is small (comparable to a machine word)
  - You want the method to work on copies too
  - The type is a map, func, or chan (reference types)

  CONSISTENCY RULE:
  If ANY method of a type uses pointer receiver,
  ALL methods should use pointer receiver.
  This avoids confusion about method sets.`)

	// ============================================
	// METHOD SETS: T vs *T
	// ============================================
	fmt.Println("\n--- Method Sets: T vs *T ---")

	fmt.Println(`
METHOD SET RULES:

  Type T's method set:
    - Only methods with VALUE receiver (T)

  Type *T's method set:
    - Methods with VALUE receiver (T)
    - Methods with POINTER receiver (*T)
    - (superset of T's method set)

  This means:
    T  can satisfy interfaces requiring value methods only
    *T can satisfy interfaces requiring any combination`)

	// Demonstrate with Point
	var p2 Point = Point{1, 2}
	var pp *Point = &Point{3, 4}

	// Point (value) has: Distance, String (value receivers)
	// *Point (pointer) has: Distance, String, Translate, Scale (all)

	var s Stringer
	s = p2  // OK: Point has String() with value receiver
	fmt.Println("  Value as Stringer:", s.String())
	s = pp // OK: *Point has String() too
	fmt.Println("  Pointer as Stringer:", s.String())

	var m Mover
	// m = p2  // COMPILE ERROR: Point does not have Translate (pointer receiver)
	m = pp // OK: *Point has Translate
	m.Translate(10, 10)
	fmt.Println("  After move:", pp.String())

	_ = m

	fmt.Println(`
  var s Stringer = Point{1,2}   // OK - String() is value receiver
  var m Mover = Point{1,2}     // ERROR - Translate() is pointer receiver
  var m Mover = &Point{1,2}    // OK - *Point has all methods`)

	// ============================================
	// INTERFACE SATISFACTION RULES
	// ============================================
	fmt.Println("\n--- Interface Satisfaction in Detail ---")

	fmt.Println(`
  WHY can't T satisfy interfaces with pointer methods?

  Because T might be stored in a non-addressable location:

    // This is addressable - Go auto-takes address:
    p := Point{1, 2}
    p.Translate(1, 1)  // Go rewrites as (&p).Translate(1, 1)

    // But this is NOT addressable:
    Point{1, 2}.Translate(1, 1)  // COMPILE ERROR

    // Inside an interface, the concrete value may not be
    // addressable, so Go can't auto-take address.
    // Therefore T cannot satisfy pointer-receiver interfaces.

  PRACTICAL EXAMPLE:
    type Writer interface { Write([]byte) (int, error) }

    // bytes.Buffer has Write with pointer receiver:
    var w Writer = &bytes.Buffer{}   // OK
    var w Writer = bytes.Buffer{}    // ERROR`)

	// ============================================
	// AUTO-DEREFERENCING IN METHOD CALLS
	// ============================================
	fmt.Println("\n--- Auto-Dereferencing ---")

	p3 := Point{5, 5}
	pp3 := &p3

	// Go automatically takes address or dereferences
	p3.Translate(1, 1) // Go rewrites as (&p3).Translate(1, 1)
	os.Stdout.WriteString(fmt.Sprintf("  p3.Translate(1,1) = %s\n", p3.String()))

	d := pp3.Distance(origin) // Go rewrites as (*pp3).Distance(origin)
	os.Stdout.WriteString(fmt.Sprintf("  pp3.Distance(origin) = %.2f\n", d))

	fmt.Println(`
  AUTO-DEREFERENCING RULES:
  - Value variable calling pointer method: Go takes address (&v).Method()
  - Pointer variable calling value method: Go dereferences (*p).Method()
  - This is syntactic sugar for VARIABLES only
  - Does NOT apply to interface values`)

	// ============================================
	// NIL RECEIVER HANDLING
	// ============================================
	fmt.Println("\n--- Nil Receiver Handling ---")

	var list *List
	fmt.Println("  nil list length:", list.Len())
	fmt.Println("  nil list string:", list.String())

	list = &List{1, &List{2, &List{3, nil}}}
	fmt.Println("  list:", list.String())
	fmt.Println("  list length:", list.Len())

	fmt.Println(`
  Pointer receivers can be called on nil pointers!
  This is useful for recursive data structures.

  func (l *List) Len() int {
      if l == nil {
          return 0     // handle nil case
      }
      return 1 + l.Next.Len()
  }

  Value receivers CANNOT be called on nil:
    var p *Point
    p.Distance(origin)  // PANIC: nil pointer dereference
    p.Translate(1, 1)   // OK: pointer receiver handles nil`)

	// ============================================
	// EMBEDDING AND METHOD PROMOTION
	// ============================================
	fmt.Println("\n--- Embedding and Method Promotion ---")

	dog := Dog{
		Animal: Animal{Name: "Rex"},
		Breed:  "Labrador",
	}

	// Promoted methods from Animal
	fmt.Println("  dog.Speak():", dog.Speak())
	fmt.Println("  dog.Fetch():", dog.Fetch())

	// Pointer method is also promoted
	dog.Rename("Max")
	fmt.Println("  After Rename:", dog.Speak())

	// Method override
	cat := Cat{Animal: Animal{Name: "Whiskers"}}
	fmt.Println("  cat.Speak():", cat.Speak())
	fmt.Println("  cat.Animal.Speak():", cat.Animal.Speak())

	fmt.Println(`
  EMBEDDING METHOD PROMOTION RULES:

  Struct S embeds T:
    - S gets T's value receiver methods
    - S's method set includes T's value methods

  Struct S embeds *T:
    - S gets all of T's methods (value + pointer)
    - S's method set includes all T's methods

  struct S embeds T (value):
    S method set  = S own methods + T value methods
    *S method set = S own methods + all T methods

  struct S embeds *T (pointer):
    S method set  = S own methods + all T methods
    *S method set = S own methods + all T methods`)

	// ============================================
	// PROMOTED METHOD SETS AND INTERFACES
	// ============================================
	fmt.Println("\n--- Promoted Method Sets and Interfaces ---")

	var sp Speaker
	sp = dog // Dog has Speak() from Animal (value receiver)
	fmt.Println("  Dog as Speaker:", sp.Speak())

	sp = cat // Cat has its own Speak() (value receiver)
	fmt.Println("  Cat as Speaker:", sp.Speak())

	var rn Renamer
	// rn = dog  // ERROR: Dog embeds Animal (value), Rename is pointer receiver
	rn = &dog // OK: *Dog has all promoted methods
	rn.Rename("Buddy")
	fmt.Println("  Dog renamed:", dog.Speak())

	_ = rn

	var sr SpeakRenamer
	// sr = dog  // ERROR: Dog (value) doesn't have Rename in method set
	sr = &dog // OK: *Dog has both Speak and Rename
	fmt.Println("  *Dog as SpeakRenamer:", sr.Speak())

	_ = sr

	fmt.Println(`
  SUMMARY:
    Dog embeds Animal (value embedding):
      Dog  satisfies Speaker (value methods)
      Dog  does NOT satisfy Renamer (needs pointer method)
      *Dog satisfies both Speaker and Renamer`)

	// ============================================
	// PERFORMANCE IMPLICATIONS
	// ============================================
	fmt.Println("\n--- Performance Implications ---")

	small := SmallStruct{1, 2}
	large := LargeStruct{}
	large.Data[0] = 42

	fmt.Println("  Small struct sum:", small.Sum())
	os.Stdout.WriteString(fmt.Sprintf("  Large struct first byte: %d\n", large.FirstByte()))

	fmt.Println(`
  PERFORMANCE GUIDELINES:

  Value receiver copies the entire struct:
    - Small structs (1-3 fields): negligible cost
    - Large structs (arrays, many fields): expensive copy

  Pointer receiver passes 8 bytes (pointer size):
    - Always cheap regardless of struct size
    - But has indirection cost (pointer chase)

  BENCHMARKING RULE:
    Don't optimize prematurely. Choose based on:
    1. Does the method need to mutate? -> pointer
    2. Is the struct large (>256 bytes)? -> pointer
    3. Is consistency needed? -> match other methods
    4. Only optimize after benchmarking shows a problem`)

	// ============================================
	// COMMON PATTERNS
	// ============================================
	fmt.Println("\n--- Common Patterns ---")

	fmt.Println(`
  PATTERN 1: Fluent/Builder API (always pointer)

    type Builder struct { parts []string }

    func (b *Builder) Add(s string) *Builder {
        b.parts = append(b.parts, s)
        return b  // enables chaining
    }

    b := &Builder{}
    b.Add("hello").Add("world")

  PATTERN 2: Immutable types (always value)

    type Money struct {
        Amount   int64
        Currency string
    }

    func (m Money) Add(other Money) Money {
        return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}
    }

  PATTERN 3: Interface compliance check

    // Compile-time check that *MyType satisfies io.Writer
    var _ io.Writer = (*MyType)(nil)

  PATTERN 4: Method on function type

    type HandlerFunc func(w ResponseWriter, r *Request)

    func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
        f(w, r)  // value receiver on function type
    }`)

	// ============================================
	// SUMMARY TABLE
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  +------------------+--------------------+---------------------+
  | Feature          | Value Receiver (T) | Pointer Receiver(*T)|
  +------------------+--------------------+---------------------+
  | Can modify?      | No (works on copy) | Yes (modifies orig) |
  | Copy cost        | Copies full struct | Copies pointer (8B) |
  | Nil safe?        | No (panic on nil)  | Yes (can check nil) |
  | T method set     | Included           | NOT included        |
  | *T method set    | Included           | Included            |
  | T -> interface   | Only value methods | N/A                 |
  | *T -> interface  | All methods        | N/A                 |
  | Auto-deref?      | p.Method() works   | v.Method() works    |
  +------------------+--------------------+---------------------+

  GOLDEN RULES:
  1. If in doubt, use pointer receiver
  2. Be consistent across all methods of a type
  3. Small immutable types: value receiver is fine
  4. Method needs mutation: MUST use pointer receiver
  5. Check interface satisfaction at compile time`)
}
