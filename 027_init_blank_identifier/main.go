// Package main - Chapter 027: Init Functions and Blank Identifier
// The init() function controls initialization order,
// and the blank identifier _ enables powerful patterns in Go.
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"unicode/utf8"
)

// ============================================
// PACKAGE-LEVEL VARIABLES (initialized first)
// ============================================

var (
	appName    = "MyApp"
	appVersion = computeVersion()
	startupLog []string
)

func computeVersion() string {
	return "1.0.0-" + "go"
}

// ============================================
// INIT FUNCTIONS
// ============================================

func init() {
	startupLog = append(startupLog, "init #1: basic setup")
}

func init() {
	startupLog = append(startupLog, "init #2: validation")
	if appName == "" {
		appName = "DefaultApp"
	}
}

func init() {
	startupLog = append(startupLog, "init #3: final prep")
}

// ============================================
// INTERFACES FOR BLANK IDENTIFIER CHECKS
// ============================================

type Validator interface {
	Validate() error
}

type Config struct {
	Host string
	Port int
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	return nil
}

// Compile-time interface check using blank identifier
var _ Validator = (*Config)(nil)

// ============================================
// DRIVER REGISTRY PATTERN
// ============================================

var drivers = make(map[string]func() string)

func registerDriver(name string, factory func() string) {
	drivers[name] = factory
}

func init() {
	registerDriver("postgres", func() string { return "PostgreSQL driver v5.0" })
	registerDriver("mysql", func() string { return "MySQL driver v8.0" })
	registerDriver("sqlite", func() string { return "SQLite driver v3.39" })
}

// ============================================
// SYNC.ONCE ALTERNATIVE TO INIT
// ============================================

var (
	instance     *Config
	instanceOnce sync.Once
)

func getConfig() *Config {
	instanceOnce.Do(func() {
		instance = &Config{Host: "localhost", Port: 8080}
	})
	return instance
}

func main() {
	fmt.Println("=== INIT FUNCTIONS AND BLANK IDENTIFIER ===")

	// ============================================
	// INIT() EXECUTION ORDER
	// ============================================
	fmt.Println("\n--- init() Execution Order ---")

	fmt.Println("  Startup log (from init functions):")
	for i, entry := range startupLog {
		os.Stdout.WriteString(fmt.Sprintf("    %d. %s\n", i+1, entry))
	}
	os.Stdout.WriteString(fmt.Sprintf("  App: %s v%s\n", appName, appVersion))

	fmt.Println(`
  INITIALIZATION ORDER:

  1. Import dependencies (recursively, depth-first)
  2. Package-level variable declarations (in order)
  3. init() functions (in order of appearance)
  4. main() function (only in package main)

  WITHIN A PACKAGE:
  - Files processed in alphabetical order
  - Variables initialized in declaration order
  - Multiple init() per file: top to bottom
  - Multiple init() per package: file order

  ACROSS PACKAGES:
  - Dependency order (imported packages first)
  - If A imports B imports C:
    C init() -> B init() -> A init() -> main()`)

	// ============================================
	// INIT() PROPERTIES
	// ============================================
	fmt.Println("\n--- init() Properties ---")

	fmt.Println(`
  RULES:
  - No parameters, no return values: func init() { ... }
  - Cannot be called or referenced
  - Multiple init() per file allowed
  - Multiple init() per package allowed
  - Runs exactly once, before main()
  - Runs in a single goroutine (no concurrency issues)
  - Panics in init() are fatal (program exits)

  EXAMPLE - MULTIPLE INIT IN ONE FILE:

    // file: setup.go
    func init() {
        // First init - runs first
        log.Println("setup: phase 1")
    }

    func init() {
        // Second init - runs second
        log.Println("setup: phase 2")
    }`)

	// ============================================
	// COMMON INIT() PATTERNS
	// ============================================
	fmt.Println("\n--- Common init() Patterns ---")

	fmt.Println("  Registered drivers:")
	for name, factory := range drivers {
		os.Stdout.WriteString(fmt.Sprintf("    %s: %s\n", name, factory()))
	}

	os.Stdout.WriteString(`
  PATTERN 1: Register Drivers/Plugins

    // In database/sql:
    func init() {
        sql.Register("postgres", &PostgresDriver{})
    }

    // Usage:
    import _ "github.com/lib/pq"  // import for side effect
    db, err := sql.Open("postgres", connStr)

  PATTERN 2: Set Defaults

    var defaultTimeout time.Duration

    func init() {
        if t := os.Getenv("TIMEOUT"); t != "" {
            defaultTimeout, _ = time.ParseDuration(t)
        } else {
            defaultTimeout = 30 * time.Second
        }
    }

  PATTERN 3: Validate Configuration

    func init() {
        required := []string{"DATABASE_URL", "API_KEY"}
        for _, key := range required {
            if os.Getenv(key) == "" {
                log.Fatalf("missing required env: %s", key)
            }
        }
    }

  PATTERN 4: Compute Lookup Tables

    var isPrime [1000]bool

    func init() {
        for i := 2; i < len(isPrime); i++ {
            isPrime[i] = true
        }
        for i := 2; i*i < len(isPrime); i++ {
            if isPrime[i] {
                for j := i * i; j < len(isPrime); j += i {
                    isPrime[j] = false
                }
            }
        }
    }`)

	// ============================================
	// INIT() ANTI-PATTERNS
	// ============================================
	fmt.Println("\n--- init() Anti-Patterns ---")

	fmt.Println(`
  ANTI-PATTERN 1: Heavy initialization

    func init() {
        // BAD: slow, may fail, hard to test
        db, err := sql.Open("postgres", os.Getenv("DB_URL"))
        if err != nil { log.Fatal(err) }
        globalDB = db
    }

    // BETTER: explicit initialization
    func main() {
        db, err := setupDB()
        if err != nil { log.Fatal(err) }
        defer db.Close()
    }

  ANTI-PATTERN 2: Hidden side effects

    func init() {
        // BAD: importing this package changes global state
        http.HandleFunc("/health", healthHandler)
    }

    // BETTER: explicit registration
    func RegisterRoutes(mux *http.ServeMux) {
        mux.HandleFunc("/health", healthHandler)
    }

  ANTI-PATTERN 3: init() that depends on order

    // BAD: fragile, depends on import order
    // package a
    func init() { shared.Config.Debug = true }

    // package b
    func init() { shared.Config.Debug = false }

    // Which wins? Depends on import order!

  PREFER sync.Once for lazy initialization:`)

	cfg := getConfig()
	os.Stdout.WriteString(fmt.Sprintf("  Lazy config: %s:%d\n", cfg.Host, cfg.Port))

	fmt.Println(`
    var instance *Config
    var once sync.Once

    func GetConfig() *Config {
        once.Do(func() {
            instance = loadConfig()
        })
        return instance
    }

  BENEFITS over init():
  - Lazy: only runs when needed
  - Testable: can be reset in tests
  - Explicit: clear where initialization happens`)

	// ============================================
	// BLANK IDENTIFIER: IGNORING VALUES
	// ============================================
	fmt.Println("\n--- Blank Identifier: Ignoring Values ---")

	// Ignore return values
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	for _, v := range m {
		_ = v // ignore value (just demonstrating)
	}

	// Ignore index in range
	fruits := []string{"apple", "banana", "cherry"}
	fmt.Println("  Fruits (ignoring index):")
	for _, fruit := range fruits {
		fmt.Println("   ", fruit)
	}

	// Ignore one of multiple return values
	value, _ := m["a"]
	os.Stdout.WriteString(fmt.Sprintf("  map lookup (ignoring ok): %d\n", value))

	r := strings.NewReader("Hello, World!")
	buf := make([]byte, 5)
	_, err := r.Read(buf)
	if err != nil {
		fmt.Println("  Error:", err)
	}
	os.Stdout.WriteString(fmt.Sprintf("  Read (ignoring n): %s\n", buf))

	fmt.Println(`
  BLANK IDENTIFIER _ :
  - Discards a value
  - Tells compiler "I know this exists, I don't need it"
  - Required because Go forbids unused variables

  COMMON USES:

    // Ignore loop index
    for _, v := range slice { ... }

    // Ignore loop value
    for i := range slice { ... }  // preferred over: for i, _ := range

    // Ignore return value
    _, err := strconv.Atoi(s)

    // Ignore multiple returns
    _, _, err := net.SplitHostPort(addr)  // ERROR: can't use _ for err
    host, _, err := net.SplitHostPort(addr)  // ignore port`)

	// ============================================
	// BLANK IMPORT (SIDE EFFECTS)
	// ============================================
	fmt.Println("\n--- Blank Import (Side Effects) ---")

	fmt.Println(`
  BLANK IMPORT: import _ "package"

  Imports a package ONLY for its init() side effects.
  Does not make any of its exported names available.

  COMMON EXAMPLES:

    // Register database drivers
    import _ "github.com/lib/pq"           // PostgreSQL
    import _ "github.com/go-sql-driver/mysql"  // MySQL

    // Register image decoders
    import _ "image/png"   // registers PNG decoder
    import _ "image/jpeg"  // registers JPEG decoder
    import _ "image/gif"   // registers GIF decoder

    // Then use the generic API:
    img, format, err := image.Decode(reader)
    // format will be "png", "jpeg", or "gif"

    // Register net/http/pprof handlers
    import _ "net/http/pprof"   // adds /debug/pprof/ routes

    // Embed files
    import _ "embed"   // enables //go:embed directive

  HOW IT WORKS:
    import _ "image/png"
    // 1. Compiles the image/png package
    // 2. Runs its init() function:
    //    func init() {
    //        image.RegisterFormat("png", pngHeader, Decode, DecodeConfig)
    //    }
    // 3. Does NOT import any names`)

	// ============================================
	// INTERFACE COMPLIANCE CHECK
	// ============================================
	fmt.Println("\n--- Interface Compliance Check ---")

	cfg2 := &Config{Host: "example.com", Port: 443}
	if err := cfg2.Validate(); err != nil {
		fmt.Println("  Validation error:", err)
	} else {
		fmt.Println("  Config is valid")
	}

	fmt.Println(`
  COMPILE-TIME INTERFACE CHECK:

    // Ensure *Config implements Validator at compile time
    var _ Validator = (*Config)(nil)

    // How it works:
    // 1. (*Config)(nil) creates a nil *Config pointer
    // 2. Assigns it to Validator interface variable
    // 3. If *Config doesn't implement Validator -> compile error
    // 4. var _ discards the value (no runtime cost)

  REAL-WORLD EXAMPLES:

    // Ensure MyHandler implements http.Handler
    var _ http.Handler = (*MyHandler)(nil)

    // Ensure MyWriter implements io.Writer
    var _ io.Writer = (*MyWriter)(nil)

    // Ensure MyError implements error
    var _ error = (*MyError)(nil)

    // Can also check value types:
    var _ fmt.Stringer = MyType{}

  WHY DO THIS?
    - Catches missing methods immediately
    - Documents the intended interface
    - No runtime cost (compiled away)`)

	// ============================================
	// STRUCT FIELD PADDING / ALIGNMENT
	// ============================================
	fmt.Println("\n--- Blank Identifier in Structs ---")

	fmt.Println(`
  PREVENT UNKEYED STRUCT LITERALS:

    type Point struct {
        X int
        Y int
        _ struct{}  // force keyed literals
    }

    // This now fails:
    p := Point{1, 2}  // ERROR: too few values

    // Must use keyed literal:
    p := Point{X: 1, Y: 2}  // OK

  PREVENT COMPARISON:

    type NoCmp struct {
        Value int
        _     [0]func()  // funcs are not comparable
    }

    a := NoCmp{Value: 1}
    b := NoCmp{Value: 1}
    // a == b  // COMPILE ERROR: cannot compare

  PREVENT COPYING (with go vet):

    type NoCopy struct {
        _ noCopy
    }
    type noCopy struct{}
    func (*noCopy) Lock()   {}
    func (*noCopy) Unlock() {}
    // go vet warns if NoCopy is copied`)

	// ============================================
	// BLANK IDENTIFIER COMPREHENSIVE PATTERNS
	// ============================================
	fmt.Println("\n--- All Blank Identifier Patterns ---")

	// Pattern: type assertion check
	var v any = "hello"
	if s, ok := v.(string); ok {
		_ = s // use value or ignore
	}

	// Pattern: ensure exhaustive handling
	count := utf8.RuneCountInString("Hello, World!")
	os.Stdout.WriteString(fmt.Sprintf("  Rune count: %d\n", count))

	fmt.Println(`
  ALL BLANK IDENTIFIER USES:

  1. Ignore return value:     _, err := f()
  2. Ignore loop index:       for _, v := range s
  3. Blank import:            import _ "pkg"
  4. Interface check:         var _ I = (*T)(nil)
  5. Force keyed literals:    _ struct{} field
  6. Prevent comparison:      _ [0]func() field
  7. Prevent copy (go vet):   _ noCopy field`)

	// ============================================
	// INITIALIZATION ORDER DEMO
	// ============================================
	fmt.Println("\n--- Full Initialization Order ---")

	fmt.Println(`
  COMPLETE ORDER FOR A GO PROGRAM:

  1. Runtime initialization (memory, scheduler, GC)

  2. For each package (dependency order):
     a. Package-level var declarations
     b. init() functions (file order, then declaration order)

  3. main.main() executes

  DEPENDENCY EXAMPLE:

    main imports A, B
    A imports C
    B imports C, D

    Order:
    1. C (init vars, init funcs)   - deepest dependency
    2. D (init vars, init funcs)
    3. A (init vars, init funcs)   - depends on C
    4. B (init vars, init funcs)   - depends on C, D
    5. main (init vars, init funcs, then main())

    NOTE: C's init() runs only ONCE even though
    both A and B import C.

  CIRCULAR IMPORTS ARE FORBIDDEN:
    If A imports B and B imports A -> compile error
    Solution: extract shared types to a third package`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  INIT() FUNCTION:
  +------------------+----------------------------+
  | Can have many    | Multiple per file/package  |
  | No params/return | func init() { ... }        |
  | Runs once        | Before main()              |
  | Order            | Dependency -> vars -> init  |
  | Best for         | Register, defaults, tables |
  | Avoid for        | Heavy I/O, DB connections  |
  +------------------+----------------------------+

  BLANK IDENTIFIER:
  +------------------+----------------------------+
  | _                | Discard values             |
  | import _         | Side-effect imports        |
  | var _ I = (*T)() | Interface compliance check |
  | _ struct{}       | Force keyed literals       |
  | _ [0]func()      | Prevent comparison         |
  +------------------+----------------------------+

  PREFER sync.Once over init() for lazy, testable init.
  PREFER explicit setup over hidden init() side effects.`)
}
