// Package main - Chapter 131: Property-Based Testing
// Understanding property-based testing, testing/quick, generators,
// and when to use property testing vs table-driven vs fuzzing.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== PROPERTY-BASED TESTING ===")

	// ============================================
	// WHAT IS PROPERTY-BASED TESTING?
	// ============================================
	fmt.Println("\n--- What is Property-Based Testing? ---")
	fmt.Println(`
PROPERTY-BASED TESTING is a testing technique where instead of
specifying exact input-output pairs (example-based testing), you
describe PROPERTIES that your code should satisfy for ALL inputs.

The testing framework then GENERATES random inputs and verifies
that the properties hold for every generated case.

EXAMPLE-BASED (traditional):
  assert add(2, 3) == 5
  assert add(0, 0) == 0
  assert add(-1, 1) == 0

PROPERTY-BASED:
  For all integers a, b:
    assert add(a, b) == add(b, a)          // commutativity
    assert add(a, 0) == a                  // identity
    assert add(a, add(b, c)) == add(add(a, b), c) // associativity

KEY INSIGHT: You test the BEHAVIOR and INVARIANTS of your code,
not specific examples. This catches edge cases you never thought of.`)

	// ============================================
	// COMPARISON: EXAMPLE vs PROPERTY vs FUZZ
	// ============================================
	fmt.Println("\n--- Testing Approaches Compared ---")
	fmt.Println(`
EXAMPLE-BASED TESTING (Table-Driven):
  - You choose specific inputs and expected outputs
  - Deterministic and reproducible
  - Tests what you thought of
  - Best for: specific business logic, known edge cases

PROPERTY-BASED TESTING:
  - Framework generates random inputs
  - You define properties/invariants that must hold
  - Tests things you didn't think of
  - Reproducible (seed-based randomness)
  - Best for: algorithms, data structures, serialization

FUZZ TESTING (go test -fuzz):
  - Mutates seed inputs to find crashes/panics
  - Coverage-guided (tries to reach new code paths)
  - Focuses on finding crashes, not logical errors
  - Built into Go since 1.18
  - Best for: parsers, validators, security-sensitive code

WHEN TO USE EACH:
+------------------+------------------+------------------+
| Table-Driven     | Property-Based   | Fuzzing          |
+------------------+------------------+------------------+
| Known I/O pairs  | Math properties  | Crash detection  |
| Business rules   | Encode/decode    | Parser testing   |
| Error cases      | Idempotency      | Malformed input  |
| Regression tests | Round-trips      | Security testing |
+------------------+------------------+------------------+`)

	// ============================================
	// THE testing/quick PACKAGE
	// ============================================
	fmt.Println("\n--- The testing/quick Package ---")
	fmt.Println(`
Go's standard library includes testing/quick for property-based testing.

PACKAGE: testing/quick

KEY FUNCTIONS:
  quick.Check(f, config)      - Verify a property holds
  quick.CheckEqual(f, g, cfg) - Verify two functions are equivalent

HOW IT WORKS:
  1. You define a function that returns bool
  2. quick.Check generates random inputs for the function parameters
  3. It calls your function many times with different random inputs
  4. If the function ever returns false, the test fails
  5. It reports the input that caused the failure

BASIC USAGE:
  import (
      "testing"
      "testing/quick"
  )

  func TestAddCommutative(t *testing.T) {
      f := func(a, b int) bool {
          return add(a, b) == add(b, a)
      }
      if err := quick.Check(f, nil); err != nil {
          t.Error(err)
      }
  }

The function f receives randomly generated values for each parameter.
quick.Check will call f up to 100 times by default.`)

	// ============================================
	// CONFIGURATION
	// ============================================
	fmt.Println("\n--- quick.Config ---")
	fmt.Println(`
You can customize the behavior with quick.Config:

  config := &quick.Config{
      MaxCount:      1000,  // Number of iterations (default: 100)
      MaxCountScale: 10.0,  // Multiplier for MaxCount
      Rand:          rng,   // Custom *rand.Rand source
      Values:        gen,   // Custom value generator function
  }

EXAMPLE WITH CONFIG:
  func TestSortIdempotent(t *testing.T) {
      config := &quick.Config{MaxCount: 500}
      f := func(xs []int) bool {
          sorted := sortCopy(xs)
          sortedAgain := sortCopy(sorted)
          return reflect.DeepEqual(sorted, sortedAgain)
      }
      if err := quick.Check(f, config); err != nil {
          t.Error(err)
      }
  }

MAXCOUNT: How many random inputs to test.
  - Default 100 is often sufficient
  - Increase for critical properties
  - Higher values = slower tests but more confidence`)

	// ============================================
	// SUPPORTED TYPES
	// ============================================
	fmt.Println("\n--- Supported Types for Generation ---")
	fmt.Println(`
testing/quick can automatically generate random values for:

BASIC TYPES:
  bool, int, int8, int16, int32, int64,
  uint, uint8, uint16, uint32, uint64,
  float32, float64, complex64, complex128,
  string, byte, rune

COMPOSITE TYPES:
  []T        - slices of any supported type
  [N]T       - arrays of any supported type
  map[K]V    - maps with supported key/value types
  struct{..} - structs with exported supported fields

EXAMPLE - Struct Generation:
  type User struct {
      Name  string
      Age   int
      Score float64
  }

  func TestUserProperty(t *testing.T) {
      f := func(u User) bool {
          // quick will generate random User values
          return u.Age >= 0 || u.Age < 0  // always true for demo
      }
      quick.Check(f, nil)
  }

NOTE: Random strings may contain any UTF-8 rune.
      Random ints span the full range of the type.`)

	// ============================================
	// COMMON PROPERTIES
	// ============================================
	fmt.Println("\n--- Common Properties to Test ---")
	fmt.Println(`
1. COMMUTATIVITY
   f(a, b) == f(b, a)
   Examples: addition, multiplication, set union, max/min

   func TestMaxCommutative(t *testing.T) {
       f := func(a, b int) bool {
           return max(a, b) == max(b, a)
       }
       quick.Check(f, nil)
   }

2. ASSOCIATIVITY
   f(f(a, b), c) == f(a, f(b, c))
   Examples: addition, string concat, list append

   func TestAddAssociative(t *testing.T) {
       f := func(a, b, c int64) bool {
           // Use int64 to check overflow behavior
           return (a + b) + c == a + (b + c)
       }
       quick.Check(f, nil)
   }

3. IDENTITY ELEMENT
   f(a, identity) == a
   Examples: add(a, 0) == a, mul(a, 1) == a, concat(s, "") == s

4. IDEMPOTENCY
   f(f(x)) == f(x)
   Examples: sort, abs, toLower, deduplicate, normalize

   func TestSortIdempotent(t *testing.T) {
       f := func(xs []int) bool {
           s1 := sortCopy(xs)
           s2 := sortCopy(s1)
           return reflect.DeepEqual(s1, s2)
       }
       quick.Check(f, nil)
   }

5. ROUND-TRIP (ENCODE/DECODE)
   decode(encode(x)) == x
   Examples: JSON, base64, compress/decompress, serialize

6. INVARIANTS
   Property that must always hold after an operation
   Examples: sorted output is same length, map size after insert`)

	// ============================================
	// ROUND-TRIP TESTING
	// ============================================
	fmt.Println("\n--- Round-Trip Testing (Encode/Decode) ---")
	os.Stdout.WriteString(`
ROUND-TRIP is one of the most powerful property-based patterns:
  decode(encode(x)) == x

JSON ROUND-TRIP:
  type Config struct {
      Host string
      Port int
      TLS  bool
  }

  func TestJSONRoundTrip(t *testing.T) {
      f := func(c Config) bool {
          data, err := json.Marshal(c)
          if err != nil {
              return false
          }
          var decoded Config
          if err := json.Unmarshal(data, &decoded); err != nil {
              return false
          }
          return reflect.DeepEqual(c, decoded)
      }
      if err := quick.Check(f, nil); err != nil {
          t.Error(err)
      }
  }

BASE64 ROUND-TRIP:
  func TestBase64RoundTrip(t *testing.T) {
      f := func(data []byte) bool {
          encoded := base64.StdEncoding.EncodeToString(data)
          decoded, err := base64.StdEncoding.DecodeString(encoded)
          if err != nil {
              return false
          }
          return bytes.Equal(data, decoded)
      }
      quick.Check(f, nil)
  }

HEX ROUND-TRIP:
  func TestHexRoundTrip(t *testing.T) {
      f := func(data []byte) bool {
          encoded := hex.EncodeToString(data)
          decoded, err := hex.DecodeString(encoded)
          if err != nil {
              return false
          }
          return bytes.Equal(data, decoded)
      }
      quick.Check(f, nil)
  }
`)

	// ============================================
	// quick.CheckEqual
	// ============================================
	fmt.Println("\n--- quick.CheckEqual ---")
	fmt.Println(`
quick.CheckEqual verifies that two functions produce identical
output for the same random inputs. Great for:
  - Verifying optimized version matches naive version
  - Comparing old and new implementations
  - Testing refactored code

USAGE:
  func TestCheckEqual(t *testing.T) {
      // Naive implementation
      naive := func(n int) int {
          sum := 0
          for i := 1; i <= n; i++ {
              sum += i
          }
          return sum
      }
      // Optimized implementation
      optimized := func(n int) int {
          if n < 0 {
              return 0
          }
          return n * (n + 1) / 2
      }
      // Verify they produce the same results
      if err := quick.CheckEqual(naive, optimized, nil); err != nil {
          t.Error(err)
      }
  }

NOTE: Both functions must have the same signature.
The framework generates matching inputs for both.`)

	// ============================================
	// CUSTOM GENERATORS
	// ============================================
	fmt.Println("\n--- Custom Generators ---")
	fmt.Println(`
For constrained inputs, implement the quick.Generator interface:

  type Generator interface {
      Generate(rand *rand.Rand, size int) reflect.Value
  }

EXAMPLE - Positive integers only:
  type PositiveInt int

  func (PositiveInt) Generate(rand *rand.Rand, size int) reflect.Value {
      return reflect.ValueOf(PositiveInt(rand.Intn(10000) + 1))
  }

  func TestSqrtPositive(t *testing.T) {
      f := func(n PositiveInt) bool {
          result := math.Sqrt(float64(n))
          return result >= 0 && !math.IsNaN(result)
      }
      quick.Check(f, nil)
  }

EXAMPLE - Valid email-like strings:
  type Email string

  func (Email) Generate(rand *rand.Rand, size int) reflect.Value {
      users := []string{"alice", "bob", "charlie", "test"}
      domains := []string{"example.com", "test.org", "mail.io"}
      user := users[rand.Intn(len(users))]
      domain := domains[rand.Intn(len(domains))]
      return reflect.ValueOf(Email(user + "@" + domain))
  }

EXAMPLE - Bounded slice:
  type SmallSlice []int

  func (SmallSlice) Generate(rand *rand.Rand, size int) reflect.Value {
      length := rand.Intn(10) // max 10 elements
      s := make(SmallSlice, length)
      for i := range s {
          s[i] = rand.Intn(100) - 50 // range [-50, 49]
      }
      return reflect.ValueOf(s)
  }`)

	// ============================================
	// VALUES GENERATOR
	// ============================================
	fmt.Println("\n--- Config.Values Generator ---")
	os.Stdout.WriteString(`
Alternative to the Generator interface: use Config.Values
to provide a custom generation function:

  func TestWithCustomValues(t *testing.T) {
      config := &quick.Config{
          Values: func(values []reflect.Value, rand *rand.Rand) {
              // Generate a string of length 1-20
              length := rand.Intn(20) + 1
              b := make([]byte, length)
              for i := range b {
                  b[i] = byte('a' + rand.Intn(26))
              }
              values[0] = reflect.ValueOf(string(b))
          },
      }

      f := func(s string) bool {
          return len(s) > 0 && len(s) <= 20
      }
      if err := quick.Check(f, config); err != nil {
          t.Error(err)
      }
  }

Config.Values gives you full control over ALL parameters at once.
The values slice has one entry per function parameter.
`)

	// ============================================
	// SHRINKING
	// ============================================
	fmt.Println("\n--- Shrinking ---")
	os.Stdout.WriteString(`
SHRINKING is the process of finding the SMALLEST failing input.

When a property fails, the framework tries to reduce the input
while still keeping the failure. This makes debugging much easier.

EXAMPLE:
  Property: "all elements in sorted slice are positive"
  First failing input: [42, -7, 100, 55, -3, 88]
  After shrinking:     [-1]      <-- minimal failing case!

Go's testing/quick has LIMITED shrinking support.
It reports the first failing input but does not shrink.

WORKAROUND - Manual investigation:
  if err := quick.Check(f, nil); err != nil {
      // err contains the failing input
      t.Logf("Failing input: %v", err)
      // You can then manually test smaller inputs
  }

For full shrinking support, consider external libraries:

pgregory.net/rapid:
  - Full shrinking support
  - Better generators
  - More flexible property definitions
  - Recommended for serious property testing
`)

	// ============================================
	// pgregory.net/rapid PATTERNS
	// ============================================
	fmt.Println("\n--- pgregory.net/rapid Patterns ---")
	fmt.Println(`
rapid is the most popular Go property-testing library.
It provides powerful generators and automatic shrinking.

NOTE: rapid is an external package. Below are patterns shown
for documentation purposes. Use "go get pgregory.net/rapid".

BASIC PATTERN:
  import "pgregory.net/rapid"

  func TestReverse(t *testing.T) {
      rapid.Check(t, func(t *rapid.T) {
          xs := rapid.SliceOf(rapid.Int()).Draw(t, "xs")
          rev := reverse(xs)
          revrev := reverse(rev)
          if !reflect.DeepEqual(xs, revrev) {
              t.Fatal("reverse(reverse(xs)) != xs")
          }
      })
  }

GENERATORS (rapid.* functions):
  rapid.Int()                  // any int
  rapid.IntRange(0, 100)       // bounded int
  rapid.Float64()              // any float64
  rapid.String()               // any string
  rapid.StringN(1, 50, -1)     // string with length 1-50
  rapid.SliceOf(rapid.Int())   // []int
  rapid.SliceOfN(gen, 1, 10)   // slice with 1-10 elements
  rapid.MapOf(keyGen, valGen)  // map generator
  rapid.Bool()                 // random bool
  rapid.Byte()                 // random byte

CUSTOM GENERATORS:
  userGen := rapid.Custom(func(t *rapid.T) User {
      return User{
          Name:  rapid.String().Draw(t, "name"),
          Age:   rapid.IntRange(0, 150).Draw(t, "age"),
          Email: rapid.StringMatching("[a-z]+@[a-z]+.com").Draw(t, "email"),
      }
  })

SHRINKING WITH rapid:
  When a test fails, rapid automatically shrinks the input:
  - Numeric values shrink toward 0
  - Strings shrink toward ""
  - Slices shrink toward []
  - Custom types shrink based on their generator

  Example failure output:
    [rapid] draw xs: [-1]
    Failed after 15 tests and 8 shrink steps`)

	// ============================================
	// PRACTICAL EXAMPLE: SORT PROPERTIES
	// ============================================
	fmt.Println("\n--- Practical: Sort Properties ---")
	fmt.Println(`
Testing a sort function with properties demonstrates the power
of property-based testing. A sort should satisfy:

  1. IDEMPOTENCY:  sort(sort(xs)) == sort(xs)
  2. LENGTH:       len(sort(xs)) == len(xs)
  3. ORDERED:      sort(xs)[i] <= sort(xs)[i+1]
  4. PERMUTATION:  sort(xs) contains same elements as xs
  5. MIN/MAX:      sort(xs)[0] == min(xs), sort(xs)[n-1] == max(xs)

FULL TEST:
  func TestSortProperties(t *testing.T) {
      // Property 1: Idempotent
      t.Run("idempotent", func(t *testing.T) {
          f := func(xs []int) bool {
              s1 := sortCopy(xs)
              s2 := sortCopy(s1)
              return reflect.DeepEqual(s1, s2)
          }
          if err := quick.Check(f, nil); err != nil {
              t.Error(err)
          }
      })

      // Property 2: Preserves length
      t.Run("preserves_length", func(t *testing.T) {
          f := func(xs []int) bool {
              return len(sortCopy(xs)) == len(xs)
          }
          if err := quick.Check(f, nil); err != nil {
              t.Error(err)
          }
      })

      // Property 3: Output is ordered
      t.Run("ordered", func(t *testing.T) {
          f := func(xs []int) bool {
              sorted := sortCopy(xs)
              for i := 1; i < len(sorted); i++ {
                  if sorted[i-1] > sorted[i] {
                      return false
                  }
              }
              return true
          }
          if err := quick.Check(f, nil); err != nil {
              t.Error(err)
          }
      })

      // Property 4: Same elements (permutation)
      t.Run("permutation", func(t *testing.T) {
          f := func(xs []int) bool {
              sorted := sortCopy(xs)
              counts := make(map[int]int)
              for _, x := range xs {
                  counts[x]++
              }
              for _, x := range sorted {
                  counts[x]--
              }
              for _, c := range counts {
                  if c != 0 {
                      return false
                  }
              }
              return true
          }
          if err := quick.Check(f, nil); err != nil {
              t.Error(err)
          }
      })
  }`)

	// ============================================
	// PRACTICAL EXAMPLE: MATH PROPERTIES
	// ============================================
	fmt.Println("\n--- Practical: Math Properties ---")
	fmt.Println(`
ABSOLUTE VALUE PROPERTIES:
  func TestAbsProperties(t *testing.T) {
      // abs(x) >= 0 for all x
      t.Run("non_negative", func(t *testing.T) {
          f := func(x float64) bool {
              return math.Abs(x) >= 0
          }
          quick.Check(f, nil)
      })

      // abs(x) == abs(-x)
      t.Run("symmetric", func(t *testing.T) {
          f := func(x float64) bool {
              return math.Abs(x) == math.Abs(-x)
          }
          quick.Check(f, nil)
      })

      // abs(x * y) == abs(x) * abs(y)
      t.Run("multiplicative", func(t *testing.T) {
          f := func(x, y float64) bool {
              if math.IsInf(x, 0) || math.IsInf(y, 0) {
                  return true // skip infinities
              }
              if math.IsNaN(x) || math.IsNaN(y) {
                  return true // skip NaN
              }
              lhs := math.Abs(x * y)
              rhs := math.Abs(x) * math.Abs(y)
              return math.Abs(lhs-rhs) < 1e-10
          }
          quick.Check(f, nil)
      })

      // abs(x) == 0 iff x == 0
      t.Run("zero", func(t *testing.T) {
          f := func(x float64) bool {
              if math.IsNaN(x) {
                  return true
              }
              return (math.Abs(x) == 0) == (x == 0)
          }
          quick.Check(f, nil)
      })
  }

STRING PROPERTIES:
  func TestStringProperties(t *testing.T) {
      // len(a + b) == len(a) + len(b) for bytes
      t.Run("concat_length", func(t *testing.T) {
          f := func(a, b string) bool {
              return len(a+b) == len(a)+len(b)
          }
          quick.Check(f, nil)
      })

      // strings.ToUpper is idempotent
      t.Run("upper_idempotent", func(t *testing.T) {
          f := func(s string) bool {
              u1 := strings.ToUpper(s)
              u2 := strings.ToUpper(u1)
              return u1 == u2
          }
          quick.Check(f, nil)
      })
  }`)

	// ============================================
	// PRACTICAL EXAMPLE: DATA STRUCTURE
	// ============================================
	fmt.Println("\n--- Practical: Data Structure Properties ---")
	fmt.Println(`
MAP/SET PROPERTIES:
  func TestMapProperties(t *testing.T) {
      // Insert then lookup should find the value
      t.Run("insert_lookup", func(t *testing.T) {
          f := func(key string, value int) bool {
              m := make(map[string]int)
              m[key] = value
              got, ok := m[key]
              return ok && got == value
          }
          quick.Check(f, nil)
      })

      // Delete then lookup should not find it
      t.Run("delete_lookup", func(t *testing.T) {
          f := func(key string, value int) bool {
              m := make(map[string]int)
              m[key] = value
              delete(m, key)
              _, ok := m[key]
              return !ok
          }
          quick.Check(f, nil)
      })

      // Size after n unique inserts should be n
      // (with high probability for random strings)
      t.Run("size_after_insert", func(t *testing.T) {
          config := &quick.Config{MaxCount: 50}
          f := func(keys []string) bool {
              m := make(map[string]bool)
              for _, k := range keys {
                  m[k] = true
              }
              // Count unique keys
              unique := make(map[string]bool)
              for _, k := range keys {
                  unique[k] = true
              }
              return len(m) == len(unique)
          }
          quick.Check(f, config)
      })
  }

STACK PROPERTIES:
  // Push then Pop returns same element (LIFO)
  // Push increases size by 1
  // Pop decreases size by 1
  // Empty stack has size 0`)

	// ============================================
	// INTEGRATION WITH go test
	// ============================================
	fmt.Println("\n--- Integration with go test ---")
	os.Stdout.WriteString(`
Property tests are normal Go tests that use testing/quick:

FILE: mypackage/mypackage_prop_test.go
  package mypackage

  import (
      "testing"
      "testing/quick"
  )

  // Property tests use the same TestXxx naming convention
  func TestAdd_Commutative(t *testing.T) {
      f := func(a, b int) bool {
          return Add(a, b) == Add(b, a)
      }
      if err := quick.Check(f, nil); err != nil {
          t.Error(err)
      }
  }

  func TestAdd_Identity(t *testing.T) {
      f := func(a int) bool {
          return Add(a, 0) == a
      }
      if err := quick.Check(f, nil); err != nil {
          t.Error(err)
      }
  }

RUNNING:
  go test -v ./mypackage/     # run all tests including property
  go test -run Property ./...  # run only property tests (if named)
  go test -count=5 ./...       # run 5 times for more confidence

COMBINING WITH TABLE-DRIVEN:
  func TestAdd(t *testing.T) {
      // Table-driven for known cases
      t.Run("examples", func(t *testing.T) {
          tests := []struct{ a, b, want int }{
              {1, 2, 3},
              {0, 0, 0},
              {-1, 1, 0},
          }
          for _, tt := range tests {
              got := Add(tt.a, tt.b)
              if got != tt.want {
                  t.Errorf("Add(%d,%d) = %d, want %d",
                      tt.a, tt.b, got, tt.want)
              }
          }
      })

      // Property-based for general invariants
      t.Run("commutative", func(t *testing.T) {
          f := func(a, b int) bool {
              return Add(a, b) == Add(b, a)
          }
          quick.Check(f, nil)
      })
  }
`)

	// ============================================
	// COMPARING WITH FUZZING
	// ============================================
	fmt.Println("\n--- Property Testing vs Fuzzing ---")
	os.Stdout.WriteString(`
PROPERTY TESTING:                    FUZZING (go test -fuzz):
  - Tests logical properties           - Tests for crashes/panics
  - You define what "correct" means    - "Not crashing" is the property
  - Random but not coverage-guided     - Coverage-guided mutation
  - Runs in normal test time           - Can run for hours
  - Good for: algorithms, encoding     - Good for: parsers, untrusted input
  - testing/quick or rapid             - Built into go test since 1.18

FUZZING EXAMPLE FOR COMPARISON:
  func FuzzParseJSON(f *testing.F) {
      f.Add([]byte(`+"`"+`{"key":"value"}`+"`"+`))
      f.Add([]byte(`+"`"+`[1,2,3]`+"`"+`))
      f.Fuzz(func(t *testing.T, data []byte) {
          var v interface{}
          // Just checking it doesn't panic
          json.Unmarshal(data, &v)
      })
  }

PROPERTY TEST EQUIVALENT:
  func TestJSONRoundTrip(t *testing.T) {
      f := func(m map[string]int) bool {
          data, err := json.Marshal(m)
          if err != nil {
              return false
          }
          var result map[string]int
          err = json.Unmarshal(data, &result)
          if err != nil {
              return false
          }
          return reflect.DeepEqual(m, result)
      }
      quick.Check(f, nil)
  }

USE BOTH TOGETHER:
  - Property tests: verify correctness invariants
  - Fuzz tests: find crashes and edge cases
  - Table-driven: document known scenarios
`)

	// ============================================
	// BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Best Practices ---")
	fmt.Println(`
1. START WITH SIMPLE PROPERTIES
   Begin with obvious properties before complex ones:
   - Length preservation
   - Type preservation
   - Idempotency
   - Round-trip

2. HANDLE EDGE CASES IN PROPERTIES
   Random generation includes edge cases (0, -1, empty, nil):
   f := func(xs []int) bool {
       if len(xs) == 0 {
           return true // empty case is trivially true
       }
       // ... test property for non-empty
   }

3. USE APPROPRIATE ITERATION COUNT
   - Default 100 is fine for most cases
   - Increase for cryptographic or security properties
   - Decrease if property is expensive to check

4. NAME YOUR PROPERTY TESTS CLEARLY
   - TestSort_Idempotent
   - TestEncode_RoundTrip
   - TestAdd_Commutative
   - TestHash_Deterministic

5. COMBINE WITH EXAMPLE TESTS
   Property tests complement, not replace, example tests.
   Use both for maximum confidence.

6. BE CAREFUL WITH FLOATING POINT
   Use epsilon comparisons, not exact equality:
   math.Abs(a - b) < 1e-10

7. CONSIDER PERFORMANCE
   Each property test runs the function 100+ times.
   Keep the property check fast.

8. DOCUMENT THE PROPERTY
   The test name or a short comment should describe
   what invariant is being verified.`)

	// ============================================
	// COMMON PITFALLS
	// ============================================
	fmt.Println("\n--- Common Pitfalls ---")
	fmt.Println(`
1. TESTING IMPLEMENTATION, NOT PROPERTY
   BAD:  func(x int) bool { return myFunc(x) == x*2+1 }
   GOOD: func(a, b int) bool { return myFunc(a+b) == myFunc(a)+myFunc(b) }
   The first just reimplements the function. The second tests a property.

2. TAUTOLOGICAL PROPERTIES
   BAD:  func(x int) bool { return sort(x) == sort(x) }
   This is always true and tests nothing useful.

3. IGNORING GENERATOR BIAS
   Random generators may not cover your domain well.
   Use custom generators for constrained inputs.

4. OVERFLOW IN ARITHMETIC PROPERTIES
   Be careful with integer overflow:
   func(a, b int) bool { return a+b == b+a }
   This works even with overflow since overflow is deterministic.
   But: func(a, b int64) bool { return a+b-b == a }
   This can fail due to overflow!

5. FORGETTING NaN IN FLOAT PROPERTIES
   NaN != NaN in IEEE 754. Always handle NaN explicitly:
   if math.IsNaN(x) { return true }  // skip NaN

6. TOO MANY SKIPPED CASES
   If your property skips most inputs, you're not testing much.
   Use custom generators instead of filtering.`)

	// ============================================
	// REAL-WORLD USE CASES
	// ============================================
	fmt.Println("\n--- Real-World Use Cases ---")
	fmt.Println(`
SERIALIZATION LIBRARIES:
  - JSON, Protocol Buffers, MessagePack
  - Test: marshal/unmarshal round-trip
  - Test: compatibility between versions

COMPRESSION:
  - gzip, zlib, snappy
  - Test: compress/decompress round-trip
  - Test: compressed size <= original + overhead

CRYPTOGRAPHY:
  - Encrypt/decrypt round-trip
  - Hash determinism (same input = same output)
  - Signature verify after sign

DATABASE OPERATIONS:
  - Insert then select returns same data
  - Update then select shows new data
  - Delete then select returns nothing

HTTP ROUTING:
  - Route generation/parsing round-trip
  - URL encoding/decoding round-trip

STATE MACHINES:
  - All transitions leave system in valid state
  - Reset always returns to initial state

PARSERS:
  - Print(Parse(input)) == input (for valid inputs)
  - Parse never panics (defensive property)`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")
	fmt.Println(`
KEY TAKEAWAYS:

1. Property-based testing verifies INVARIANTS, not examples
2. testing/quick is built into Go stdlib
3. quick.Check generates random inputs for your function
4. quick.CheckEqual compares two implementations
5. Common properties: commutativity, idempotency, round-trip
6. Custom generators via quick.Generator interface
7. Combine with table-driven tests and fuzzing
8. pgregory.net/rapid for advanced features (shrinking)
9. Property tests are regular Go tests - run with go test
10. Start simple, add properties as you discover invariants`)
}

/* SUMMARY - PROPERTY-BASED TESTING:
Property-based testing verifies code invariants with random inputs
instead of specific examples. Go's testing/quick provides quick.Check
and quick.CheckEqual. Common properties: commutativity, associativity,
idempotency, identity, round-trip (encode/decode). Custom generators
via the Generator interface constrain random inputs. Shrinking in
testing/quick is limited; pgregory.net/rapid offers full shrinking.
Combine property tests with table-driven tests (known cases) and
fuzz testing (crash detection). Run with standard go test tooling.
*/
