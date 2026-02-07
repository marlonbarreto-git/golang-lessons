// Package main - Chapter 132: encoding/json/v2
// Exploring the new JSON v2 package: strict defaults, omitzero,
// inline tags, performance improvements, and migration from v1.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== ENCODING/JSON/V2 ===")

	// ============================================
	// WHY JSON V2?
	// ============================================
	fmt.Println("\n--- Why JSON v2? ---")
	fmt.Println(`
The original encoding/json package (v1) has been part of Go since
Go 1.0 (2012). Over the years, several design issues became apparent
that could not be fixed without breaking backward compatibility.

PROBLEMS WITH encoding/json (V1):

1. CASE-INSENSITIVE MATCHING
   V1 matches JSON keys to struct fields case-insensitively:
     {"NAME": "Alice"} matches struct field Name
   This is surprising and can cause subtle bugs.

2. DUPLICATE KEYS SILENTLY ACCEPTED
   V1 silently accepts duplicate keys in JSON objects:
     {"name": "Alice", "name": "Bob"}  -> uses "Bob" (last wins)
   This hides data errors and violates RFC 7159 recommendations.

3. OMITEMPTY INCONSISTENCIES
   V1's omitempty behavior is confusing:
   - Zero-value time.Time is NOT omitted (not empty)
   - Empty map IS omitted
   - Pointer to zero value is NOT omitted
   Different types have different "empty" definitions.

4. NO WAY TO OMIT ZERO VALUES GENERICALLY
   V1 has no tag to say "omit if zero value" for all types.
   omitempty means different things for different types.

5. POOR PERFORMANCE
   V1 uses reflection heavily and allocates frequently.
   Many third-party libraries (easyjson, jsoniter) exist
   solely because of v1's performance issues.

6. LIMITED CUSTOMIZATION
   V1's Marshal/Unmarshal interfaces are all-or-nothing.
   You can't customize just the encoding of one field.

7. UNEXPORTED FIELD HANDLING
   V1 silently ignores unexported fields, which can lead
   to data loss during round-trips.

SOLUTION: encoding/json/v2 (also called "json v2")
Available as an experiment starting in Go 1.24+.`)

	// ============================================
	// KEY DIFFERENCES
	// ============================================
	fmt.Println("\n--- Key Differences: V1 vs V2 ---")
	fmt.Println(`
+-------------------------+------------------+------------------+
| Feature                 | V1 (json)        | V2 (json/v2)     |
+-------------------------+------------------+------------------+
| Case sensitivity        | Insensitive      | Sensitive         |
| Duplicate keys          | Last wins        | Error (rejected)  |
| omitempty               | Type-dependent   | Consistent        |
| omitzero                | Not available    | Available         |
| inline tag              | Not available    | Available         |
| Performance             | Slower           | Significantly     |
|                         |                  | faster            |
| String format options   | Limited          | Extensive         |
| Streaming               | Decoder only     | Full Encoder/     |
|                         |                  | Decoder           |
| Error messages          | Generic          | Detailed with     |
|                         |                  | JSON path         |
| Default behavior        | Lenient          | Strict            |
+-------------------------+------------------+------------------+

V2 is STRICT BY DEFAULT. This means:
  - Case-sensitive field matching
  - Duplicate keys are rejected
  - Unknown fields cause errors during unmarshal
  - More predictable behavior overall`)

	// ============================================
	// STRICT BY DEFAULT
	// ============================================
	fmt.Println("\n--- Strict by Default ---")
	fmt.Println(`
V2 CASE SENSITIVITY:
  type User struct {
      Name string
      Age  int
  }

  V1 behavior (lenient):
    {"name": "Alice"}     -> User{Name: "Alice"}  (matches!)
    {"NAME": "Alice"}     -> User{Name: "Alice"}  (matches!)
    {"NaMe": "Alice"}     -> User{Name: "Alice"}  (matches!)

  V2 behavior (strict):
    {"Name": "Alice"}     -> User{Name: "Alice"}  (matches!)
    {"name": "Alice"}     -> ERROR: unknown field  (no match!)
    {"NAME": "Alice"}     -> ERROR: unknown field  (no match!)

  To match v1 behavior in v2, use the json tag:
    type User struct {
        Name string ` + "`" + `json:"name"` + "`" + `
        Age  int    ` + "`" + `json:"age"` + "`" + `
    }

V2 DUPLICATE KEYS:
  Input: {"name": "Alice", "name": "Bob"}

  V1: silently uses "Bob" (last value wins)
  V2: returns an error (duplicate key rejected)

V2 UNKNOWN FIELDS:
  type User struct {
      Name string ` + "`" + `json:"name"` + "`" + `
  }
  Input: {"name": "Alice", "email": "alice@test.com"}

  V1: silently ignores "email"
  V2: returns an error (unknown field)
  V2 with option: can be configured to ignore unknown fields`)

	// ============================================
	// OMITZERO VS OMITEMPTY
	// ============================================
	fmt.Println("\n--- omitzero vs omitempty ---")
	os.Stdout.WriteString(`
V1 OMITEMPTY PROBLEMS:
  type Event struct {
      Name    string    ` + "`" + `json:"name,omitempty"` + "`" + `
      Time    time.Time ` + "`" + `json:"time,omitempty"` + "`" + `
      Count   int       ` + "`" + `json:"count,omitempty"` + "`" + `
      Enabled *bool     ` + "`" + `json:"enabled,omitempty"` + "`" + `
  }

  V1 with zero values:
    {"time": "0001-01-01T00:00:00Z", "enabled": null}
    - "name" omitted (empty string)
    - "time" NOT omitted (time.Time zero is not "empty" in v1)
    - "count" omitted (0 is "empty" for int)
    - "enabled" NOT omitted (nil pointer is not "empty" in v1... wait, actually it IS omitted in v1)

  The rules for what counts as "empty" vary by type in V1:
    string: ""  -> omitted
    int: 0      -> omitted
    bool: false -> omitted
    pointer: nil -> omitted
    time.Time: zero -> NOT omitted (it's a struct!)
    slice: nil -> omitted, but [] -> NOT omitted

V2 OMITZERO - CONSISTENT ZERO-VALUE OMISSION:
  type Event struct {
      Name    string    ` + "`" + `json:"name,omitzero"` + "`" + `
      Time    time.Time ` + "`" + `json:"time,omitzero"` + "`" + `
      Count   int       ` + "`" + `json:"count,omitzero"` + "`" + `
      Enabled *bool     ` + "`" + `json:"enabled,omitzero"` + "`" + `
  }

  V2 with zero values:
    {}
    ALL zero-value fields are omitted. Simple, consistent.

  omitzero uses the Go zero value (== zero) for all types:
    string: ""         -> omitted
    int: 0             -> omitted
    bool: false        -> omitted
    pointer: nil       -> omitted
    time.Time: zero    -> omitted (it IS the zero value!)
    slice: nil         -> omitted
    slice: []          -> NOT omitted (not zero value)
    map: nil           -> omitted

V2 ALSO SUPPORTS OMITEMPTY (for backward compat):
  In V2, omitempty keeps the V1 behavior for compatibility.
  Use omitzero for the new, consistent behavior.
  You can even combine them: ` + "`" + `json:"field,omitempty,omitzero"` + "`" + `
`)

	// ============================================
	// INLINE TAG
	// ============================================
	fmt.Println("\n--- The inline Tag ---")
	fmt.Println(`
V2 introduces the "inline" tag for embedding fields directly
into the parent JSON object.

WITHOUT INLINE:
  type Address struct {
      City    string ` + "`" + `json:"city"` + "`" + `
      Country string ` + "`" + `json:"country"` + "`" + `
  }
  type User struct {
      Name    string  ` + "`" + `json:"name"` + "`" + `
      Address Address ` + "`" + `json:"address"` + "`" + `
  }
  -> {"name":"Alice","address":{"city":"Lima","country":"PE"}}

WITH INLINE:
  type User struct {
      Name    string  ` + "`" + `json:"name"` + "`" + `
      Address Address ` + "`" + `json:",inline"` + "`" + `
  }
  -> {"name":"Alice","city":"Lima","country":"PE"}

The inline tag "flattens" the nested struct into the parent.

USE CASES:
  1. Flatten nested configuration
  2. Merge metadata with data
  3. API responses with embedded pagination
  4. Compose types without nesting in JSON

INLINE WITH MAP:
  type Document struct {
      ID    string         ` + "`" + `json:"id"` + "`" + `
      Extra map[string]any ` + "`" + `json:",inline"` + "`" + `
  }
  Input: {"id":"123","author":"Bob","version":2}
  -> Document{ID:"123", Extra:map[author:Bob version:2]}

  This replaces the common V1 pattern of unmarshaling to
  map[string]any and then manually extracting fields.`)

	// ============================================
	// STRING FORMAT OPTIONS
	// ============================================
	fmt.Println("\n--- String Format Options ---")
	fmt.Println(`
V2 provides rich formatting options for fields:

NUMERIC FORMATS:
  type Record struct {
      ID      int    ` + "`" + `json:"id,format:string"` + "`" + `
      Flags   uint8  ` + "`" + `json:"flags,format:string"` + "`" + `
  }
  -> {"id":"42","flags":"255"}

TIME FORMATS:
  type Event struct {
      Created time.Time ` + "`" + `json:"created,format:RFC3339"` + "`" + `
      Updated time.Time ` + "`" + `json:"updated,format:UnixSec"` + "`" + `
  }
  -> {"created":"2024-01-15T10:30:00Z","updated":1705312200}

  Available time formats:
    RFC3339        - "2024-01-15T10:30:00Z" (default)
    RFC3339Nano    - "2024-01-15T10:30:00.000000000Z"
    UnixSec        - 1705312200 (seconds since epoch)
    UnixMilli      - 1705312200000 (milliseconds)
    UnixMicro      - 1705312200000000 (microseconds)
    UnixNano       - 1705312200000000000 (nanoseconds)

BYTE SLICE FORMATS:
  type Data struct {
      Payload []byte ` + "`" + `json:"payload,format:base64"` + "`" + `
      Raw     []byte ` + "`" + `json:"raw,format:hex"` + "`" + `
  }

  Formats for []byte:
    base64   - standard base64 encoding (default)
    base64url - URL-safe base64
    base32   - base32 encoding
    base32hex - base32 with hex alphabet
    hex      - hexadecimal encoding
    array    - JSON array of numbers [72,101,108]`)

	// ============================================
	// NEW MARSHAL/UNMARSHAL INTERFACES
	// ============================================
	fmt.Println("\n--- New Marshal/Unmarshal Interfaces ---")
	os.Stdout.WriteString(`
V1 INTERFACES:
  type Marshaler interface {
      MarshalJSON() ([]byte, error)
  }
  type Unmarshaler interface {
      UnmarshalJSON([]byte) error
  }

V2 NEW INTERFACES (more powerful):
  type MarshalerV2 interface {
      MarshalJSONV2(enc *jsontext.Encoder, opts Options) error
  }
  type UnmarshalerV2 interface {
      UnmarshalJSONV2(dec *jsontext.Decoder, opts Options) error
  }

WHY V2 INTERFACES ARE BETTER:

1. ACCESS TO ENCODER/DECODER
   Instead of working with raw []byte, you get a streaming
   encoder/decoder. This is more efficient for large objects.

2. OPTIONS PROPAGATION
   Options are passed through the entire marshal/unmarshal
   chain. Custom types can respect global options.

3. STREAMING SUPPORT
   V2 interfaces work with streaming JSON, not buffered bytes.

EXAMPLE V2 INTERFACE:
  type Duration struct {
      time.Duration
  }

  func (d Duration) MarshalJSONV2(enc *jsontext.Encoder, opts Options) error {
      return enc.WriteToken(jsontext.String(d.String()))
  }

  func (d *Duration) UnmarshalJSONV2(dec *jsontext.Decoder, opts Options) error {
      tok, err := dec.ReadToken()
      if err != nil {
          return err
      }
      dur, err := time.ParseDuration(tok.String())
      if err != nil {
          return err
      }
      d.Duration = dur
      return nil
  }

V1 INTERFACES STILL WORK in V2 for backward compatibility.
V2 interfaces take precedence if both are implemented.
`)

	// ============================================
	// jsontext PACKAGE
	// ============================================
	fmt.Println("\n--- jsontext.Encoder and jsontext.Decoder ---")
	fmt.Println(`
V2 introduces the jsontext package for low-level JSON handling.

jsontext.Encoder:
  - Writes JSON tokens to an io.Writer
  - Validates JSON structure automatically
  - More efficient than building []byte manually

  var buf bytes.Buffer
  enc := jsontext.NewEncoder(&buf)
  enc.WriteToken(jsontext.ObjectStart)
  enc.WriteToken(jsontext.String("name"))
  enc.WriteToken(jsontext.String("Alice"))
  enc.WriteToken(jsontext.String("age"))
  enc.WriteToken(jsontext.Int(30))
  enc.WriteToken(jsontext.ObjectEnd)
  // buf: {"name":"Alice","age":30}

jsontext.Decoder:
  - Reads JSON tokens from an io.Reader
  - Validates JSON structure automatically
  - Reports detailed errors with byte offsets

  dec := jsontext.NewDecoder(reader)
  for {
      tok, err := dec.ReadToken()
      if err == io.EOF {
          break
      }
      switch tok.Kind() {
      case '"':
          fmt.Println("String:", tok.String())
      case '0':
          fmt.Println("Number:", tok.Int())
      case '{':
          fmt.Println("Object start")
      case '}':
          fmt.Println("Object end")
      }
  }

jsontext.Value:
  - Represents raw JSON as a byte slice
  - Can be used for delayed parsing
  - Validates that the content is valid JSON

  raw := jsontext.Value(` + "`" + `{"key":"value"}` + "`" + `)
  if !raw.IsValid() {
      log.Fatal("invalid JSON")
  }`)

	// ============================================
	// OPTIONS AND CODERS
	// ============================================
	fmt.Println("\n--- Options and Coders ---")
	fmt.Println(`
V2 uses a composable Options system for configuration:

COMMON OPTIONS:
  json.DefaultOptionsV2()          // strict v2 defaults
  json.DefaultOptionsV1()          // lenient v1-compatible defaults

  jsontext.WithIndent("  ")        // pretty-print with indentation
  jsontext.AllowDuplicateNames(true) // allow duplicate keys
  jsontext.AllowInvalidUTF8(true)  // allow invalid UTF-8 strings

  json.WithUnmarshalers(...)       // custom unmarshal functions
  json.WithMarshalers(...)         // custom marshal functions
  json.StringifyNumbers(true)      // marshal numbers as strings
  json.RejectUnknownMembers(true)  // error on unknown JSON fields
  json.DiscardUnknownMembers(true) // silently ignore unknown fields

COMPOSING OPTIONS:
  opts := json.JoinOptions(
      json.DefaultOptionsV2(),
      jsontext.WithIndent("  "),
      json.StringifyNumbers(true),
  )

  data, err := json.Marshal(value, opts)
  err = json.Unmarshal(data, &result, opts)

CUSTOM MARSHALERS (function-level):
  opts := json.JoinOptions(
      json.WithMarshalers(
          json.MarshalFuncV2(func(enc *jsontext.Encoder,
              t time.Time, opts json.Options) error {
              return enc.WriteToken(
                  jsontext.String(t.Format("2006-01-02")))
          }),
      ),
  )

  This lets you customize encoding for specific types
  WITHOUT modifying the types themselves.`)

	// ============================================
	// PERFORMANCE IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Performance Improvements ---")
	fmt.Println(`
V2 is significantly faster than V1 across most benchmarks:

TYPICAL IMPROVEMENTS:
  +---------------------------+--------+--------+----------+
  | Operation                 | V1     | V2     | Speedup  |
  +---------------------------+--------+--------+----------+
  | Marshal small struct      | 500ns  | 200ns  | 2.5x     |
  | Marshal large struct      | 5us    | 1.5us  | 3.3x     |
  | Unmarshal small struct    | 800ns  | 300ns  | 2.7x     |
  | Unmarshal large struct    | 8us    | 2.5us  | 3.2x     |
  | Marshal []int (1000)      | 15us   | 5us    | 3.0x     |
  | Marshal map[string]any    | 3us    | 1us    | 3.0x     |
  +---------------------------+--------+--------+----------+
  (approximate; actual results depend on data and hardware)

WHY V2 IS FASTER:

1. REDUCED ALLOCATIONS
   V2 pools and reuses internal buffers.
   V1 allocates frequently during reflection.

2. BETTER REFLECTION CACHING
   V2 caches type information more aggressively.
   Type analysis is done once and reused.

3. STREAMING ARCHITECTURE
   V2's encoder/decoder work with streaming tokens,
   avoiding intermediate []byte allocations.

4. OPTIMIZED STRING HANDLING
   V2 uses optimized routines for JSON string
   escaping and unescaping.

5. INLINED FAST PATHS
   Common types (string, int, bool) have hand-optimized
   code paths that avoid generic reflection.

BENCHMARK YOURSELF:
  func BenchmarkMarshalV1(b *testing.B) {
      v := MyStruct{...}
      for b.Loop() {
          json.Marshal(v)
      }
  }

  func BenchmarkMarshalV2(b *testing.B) {
      v := MyStruct{...}
      for b.Loop() {
          jsonv2.Marshal(v)
      }
  }`)

	// ============================================
	// MIGRATION GUIDE V1 TO V2
	// ============================================
	fmt.Println("\n--- Migration Guide: V1 to V2 ---")
	os.Stdout.WriteString(`
STEP 1: USE V1-COMPATIBLE DEFAULTS
  Start migration by using V1-compatible defaults in V2:

  import "encoding/json/v2"

  // Use v1 defaults to start - same behavior as v1
  opts := json.DefaultOptionsV1()
  data, err := json.Marshal(value, opts)
  err = json.Unmarshal(data, &target, opts)

  This gives you V2's performance with V1's behavior.

STEP 2: GRADUALLY TIGHTEN
  Switch to V2 defaults one option at a time:

  opts := json.JoinOptions(
      json.DefaultOptionsV1(),
      // Start enabling V2 strictness one by one:
      // jsontext.AllowDuplicateNames(false),  // step 2a
      // json.RejectUnknownMembers(true),      // step 2b
  )

STEP 3: SWITCH TO V2 DEFAULTS
  Once all code is compatible:

  opts := json.DefaultOptionsV2()

STEP 4: ADOPT V2 FEATURES
  Use new features like omitzero, inline, format options.

IMPORT PATH:
  V1: import "encoding/json"
  V2: import "encoding/json/v2"

  Both can coexist in the same project during migration.

API CHANGES:
  V1: json.Marshal(v)            -> []byte, error
  V2: json.Marshal(v, opts...)   -> []byte, error

  V1: json.Unmarshal(data, &v)            -> error
  V2: json.Unmarshal(data, &v, opts...)   -> error

  V1: json.NewEncoder(w).Encode(v)
  V2: json.MarshalWrite(w, v, opts...)

  V1: json.NewDecoder(r).Decode(&v)
  V2: json.UnmarshalRead(r, &v, opts...)

TAG COMPATIBILITY:
  V1 tags work in V2:
    ` + "`" + `json:"name"` + "`" + `           - field name
    ` + "`" + `json:"name,omitempty"` + "`" + `  - omit if empty (v1 behavior)
    ` + "`" + `json:"-"` + "`" + `              - skip field

  V2 new tags:
    ` + "`" + `json:",omitzero"` + "`" + `      - omit if zero value
    ` + "`" + `json:",inline"` + "`" + `        - flatten into parent
    ` + "`" + `json:",format:RFC3339"` + "`" + ` - time format
`)

	// ============================================
	// V1 VS V2 BEHAVIOR EXAMPLES
	// ============================================
	fmt.Println("\n--- V1 vs V2 Behavior Examples ---")
	fmt.Println(`
EXAMPLE 1: Case Sensitivity
  type Config struct {
      Host string
      Port int
  }

  JSON: {"host": "localhost", "port": 8080}

  V1: Config{Host: "localhost", Port: 8080}  (works, case-insensitive)
  V2: error: unknown field "host"            (strict, case-sensitive)
  V2 fix: add json tags ` + "`" + `json:"host"` + "`" + ` and ` + "`" + `json:"port"` + "`" + `

EXAMPLE 2: Duplicate Keys
  JSON: {"name": "Alice", "age": 30, "name": "Bob"}

  V1: uses "Bob" (last wins, silently)
  V2: error: duplicate key "name"

EXAMPLE 3: Unknown Fields
  type User struct {
      Name string ` + "`" + `json:"name"` + "`" + `
  }
  JSON: {"name": "Alice", "email": "alice@test.com"}

  V1: User{Name: "Alice"}  (ignores "email")
  V2: error: unknown field "email"
  V2 with option:
    json.Unmarshal(data, &u, json.DiscardUnknownMembers(true))

EXAMPLE 4: Empty Slice vs Nil Slice
  type Data struct {
      Items []string ` + "`" + `json:"items"` + "`" + `
  }

  V1 (nil slice):  {"items": null}
  V2 (nil slice):  {}            (omitted with omitzero)
  V1 (empty slice): {"items": []}
  V2 (empty slice): {"items": []}

EXAMPLE 5: time.Time with omitempty
  type Event struct {
      Name string    ` + "`" + `json:"name"` + "`" + `
      At   time.Time ` + "`" + `json:"at,omitempty"` + "`" + `
  }

  V1 (zero time): {"name":"test","at":"0001-01-01T00:00:00Z"}
  (time.Time zero is NOT empty in v1!)

  V2 with omitzero:
  type Event struct {
      Name string    ` + "`" + `json:"name"` + "`" + `
      At   time.Time ` + "`" + `json:"at,omitzero"` + "`" + `
  }
  V2 (zero time): {"name":"test"}
  (zero time IS omitted with omitzero!)`)

	// ============================================
	// ADVANCED: CUSTOM MARSHALERS
	// ============================================
	fmt.Println("\n--- Advanced: Custom Type Marshalers ---")
	os.Stdout.WriteString(`
V2 lets you register custom marshal/unmarshal functions
for specific types WITHOUT modifying the types:

EXAMPLE: Custom time format globally
  opts := json.JoinOptions(
      json.WithMarshalers(json.NewMarshalers(
          json.MarshalFuncV2(func(
              enc *jsontext.Encoder,
              t time.Time,
              opts json.Options,
          ) error {
              s := t.Format("2006-01-02")
              return enc.WriteToken(jsontext.String(s))
          }),
      )),
      json.WithUnmarshalers(json.NewUnmarshalers(
          json.UnmarshalFuncV2(func(
              dec *jsontext.Decoder,
              t *time.Time,
              opts json.Options,
          ) error {
              tok, err := dec.ReadToken()
              if err != nil {
                  return err
              }
              parsed, err := time.Parse("2006-01-02", tok.String())
              if err != nil {
                  return err
              }
              *t = parsed
              return nil
          }),
      )),
  )

  type Event struct {
      Name string    ` + "`" + `json:"name"` + "`" + `
      Date time.Time ` + "`" + `json:"date"` + "`" + `
  }

  e := Event{Name: "Launch", Date: time.Now()}
  data, _ := json.Marshal(e, opts)
  // {"name":"Launch","date":"2024-06-15"}

This is much cleaner than V1 where you had to create
wrapper types or implement MarshalJSON on every struct.
`)

	// ============================================
	// ERROR HANDLING IMPROVEMENTS
	// ============================================
	fmt.Println("\n--- Error Handling Improvements ---")
	os.Stdout.WriteString(`
V2 provides much better error messages with JSON paths:

V1 ERRORS (vague):
  json: cannot unmarshal string into Go struct field User.Age of type int

V2 ERRORS (precise):
  json: cannot unmarshal JSON string into Go int at path $.users[2].age

V2 ERROR TYPES:
  *json.SemanticError:
    - Offset: byte offset in JSON where error occurred
    - JSONPointer: JSON path (e.g., "/users/2/age")
    - GoType: the Go type that failed
    - Err: underlying error

  USING ERROR DETAILS:
    var target MyStruct
    err := json.Unmarshal(data, &target)
    if err != nil {
        var semErr *json.SemanticError
        if errors.As(err, &semErr) {
            fmt.Printf("Error at %s: %v\n",
                semErr.JSONPointer, semErr.Err)
            fmt.Printf("Go type: %v\n", semErr.GoType)
        }
    }

V1 SYNTAX ERRORS:
  json: invalid character 'x' looking for beginning of value

V2 SYNTAX ERRORS:
  jsontext: invalid character 'x' at offset 42 within object
  (includes byte offset for precise location)`)

	// ============================================
	// STREAMING JSON
	// ============================================
	fmt.Println("\n--- Streaming JSON ---")
	fmt.Println(`
V2 has first-class streaming support:

MARSHAL TO WRITER:
  V1:
    enc := json.NewEncoder(w)
    enc.Encode(value)

  V2:
    json.MarshalWrite(w, value, opts...)

UNMARSHAL FROM READER:
  V1:
    dec := json.NewDecoder(r)
    dec.Decode(&value)

  V2:
    json.UnmarshalRead(r, &value, opts...)

PROCESSING LARGE JSON ARRAYS:
  // V2 streaming approach for large arrays
  dec := jsontext.NewDecoder(reader)

  // Read array start
  tok, _ := dec.ReadToken() // '['

  for dec.PeekKind() != ']' {
      var item MyStruct
      json.UnmarshalRead(dec, &item)
      process(item)
  }

  // Read array end
  tok, _ = dec.ReadToken() // ']'

BENEFITS OF STREAMING:
  - Constant memory usage regardless of JSON size
  - Can process items as they arrive
  - Works with io.Reader/io.Writer directly
  - No need to buffer entire JSON in memory`)

	// ============================================
	// PRACTICAL MIGRATION PATTERNS
	// ============================================
	fmt.Println("\n--- Practical Migration Patterns ---")
	os.Stdout.WriteString(`
PATTERN 1: Wrapper for gradual migration
  package jsonutil

  import (
      jsonv2 "encoding/json/v2"
  )

  var defaultOpts = jsonv2.DefaultOptionsV1()

  func Marshal(v any) ([]byte, error) {
      return jsonv2.Marshal(v, defaultOpts)
  }

  func Unmarshal(data []byte, v any) error {
      return jsonv2.Unmarshal(data, v, defaultOpts)
  }

PATTERN 2: Struct tag audit
  Before migration, audit all JSON-tagged structs:

  // V1: works without tags (case-insensitive)
  type Config struct {
      Host     string    // matches "host", "HOST", "Host"
      Port     int       // matches "port", "PORT", "Port"
      Timeout  int       // matches "timeout", etc.
  }

  // V2: add explicit tags for API compatibility
  type Config struct {
      Host     string ` + "`" + `json:"host"` + "`" + `
      Port     int    ` + "`" + `json:"port"` + "`" + `
      Timeout  int    ` + "`" + `json:"timeout"` + "`" + `
  }

PATTERN 3: Replace omitempty with omitzero
  // V1
  type Response struct {
      Data    any       ` + "`" + `json:"data,omitempty"` + "`" + `
      Error   string    ` + "`" + `json:"error,omitempty"` + "`" + `
      Created time.Time ` + "`" + `json:"created,omitempty"` + "`" + `
  }

  // V2 - consistent zero-value omission
  type Response struct {
      Data    any       ` + "`" + `json:"data,omitzero"` + "`" + `
      Error   string    ` + "`" + `json:"error,omitzero"` + "`" + `
      Created time.Time ` + "`" + `json:"created,omitzero"` + "`" + `
  }

PATTERN 4: Use inline for API wrappers
  // V1: manual flatten/unflatten
  type APIResponse struct {
      Status  string          ` + "`" + `json:"status"` + "`" + `
      Message string          ` + "`" + `json:"message"` + "`" + `
      Data    json.RawMessage ` + "`" + `json:"data"` + "`" + `
  }

  // V2: inline for automatic flatten
  type APIResponse[T any] struct {
      Status  string ` + "`" + `json:"status"` + "`" + `
      Message string ` + "`" + `json:"message"` + "`" + `
      Data    T      ` + "`" + `json:",inline"` + "`" + `
  }
`)

	// ============================================
	// GOTCHAS AND TIPS
	// ============================================
	fmt.Println("\n--- Gotchas and Tips ---")
	fmt.Println(`
1. IMPORT PATH
   V2 is at "encoding/json/v2", NOT "encoding/json"
   Both can coexist:
     import (
         jsonv1 "encoding/json"
         jsonv2 "encoding/json/v2"
     )

2. JSON TAGS ARE MORE IMPORTANT IN V2
   Without tags, V2 uses exact Go field names (case-sensitive).
   Always add json tags for external APIs.

3. TEST YOUR UNMARSHAL PATHS
   V2 is strict by default. Existing JSON payloads from external
   systems may have duplicate keys or case mismatches.
   Test with real data before switching.

4. OPTIONS ARE COMPOSABLE
   Use JoinOptions to build option sets:
   opts := json.JoinOptions(opt1, opt2, opt3)
   Later options override earlier ones for the same setting.

5. V2 INTERFACES TAKE PRIORITY
   If a type implements both MarshalJSON (v1) and
   MarshalJSONV2 (v2), the V2 interface is used.

6. RAW JSON HANDLING
   V1: json.RawMessage
   V2: jsontext.Value
   Both work, but jsontext.Value has validation.

7. NUMBER HANDLING
   V2 preserves number precision better than V1.
   V1 converts all numbers to float64 for any/interface{}.
   V2 can preserve exact numeric values.

8. BACKWARD COMPATIBILITY
   V2 with DefaultOptionsV1() should behave identically to V1.
   Use this as a safe starting point for migration.`)

	// ============================================
	// USING V2 IN TESTS
	// ============================================
	fmt.Println("\n--- Using V2 in Tests ---")
	os.Stdout.WriteString(`
TESTING JSON COMPATIBILITY:
  func TestJSONV1V2Compat(t *testing.T) {
      type User struct {
          Name string ` + "`" + `json:"name"` + "`" + `
          Age  int    ` + "`" + `json:"age"` + "`" + `
      }

      u := User{Name: "Alice", Age: 30}

      // Marshal with v1
      v1Data, _ := jsonv1.Marshal(u)

      // Marshal with v2 (v1 defaults)
      v2Data, _ := jsonv2.Marshal(u, jsonv2.DefaultOptionsV1())

      // Should produce identical output
      if !bytes.Equal(v1Data, v2Data) {
          t.Errorf("v1: %s\nv2: %s", v1Data, v2Data)
      }
  }

TESTING STRICT MODE:
  func TestStrictUnmarshal(t *testing.T) {
      data := []byte(` + "`" + `{"name":"Alice","unknown":"field"}` + "`" + `)

      var u User
      err := jsonv2.Unmarshal(data, &u) // v2 strict defaults
      if err == nil {
          t.Error("expected error for unknown field")
      }
  }

GOLDEN FILE TESTS:
  Compare V1 and V2 output for all your API responses.
  This ensures migration doesn't break existing clients.`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")
	fmt.Println(`
KEY TAKEAWAYS:

1.  encoding/json/v2 fixes long-standing V1 design issues
2.  V2 is STRICT by default (case-sensitive, no duplicates)
3.  omitzero provides consistent zero-value omission
4.  inline tag flattens nested structs in JSON output
5.  format tag controls time, number, and byte encoding
6.  New MarshalerV2/UnmarshalerV2 interfaces for streaming
7.  jsontext package for low-level token reading/writing
8.  Composable Options system replaces Encoder/Decoder config
9.  Significantly faster than V1 (2-3x in most benchmarks)
10. Migrate gradually: start with DefaultOptionsV1(), tighten over time
11. Both V1 and V2 can coexist during migration
12. Always add json tags to structs for API compatibility`)
}

/* SUMMARY - ENCODING/JSON/V2:
encoding/json/v2 is the next generation JSON package for Go.
It is strict by default: case-sensitive field matching, duplicate
key rejection, and unknown field errors. New features: omitzero
for consistent zero-value omission, inline tag for flattening
nested structs, format tag for time/number/byte formatting.
jsontext package provides low-level streaming Encoder/Decoder.
Composable Options system for configuration. 2-3x faster than V1.
Migrate gradually using DefaultOptionsV1() for V1 compatibility.
New MarshalerV2/UnmarshalerV2 interfaces with streaming support.
*/
