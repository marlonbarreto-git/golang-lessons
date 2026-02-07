// Package main - Chapter 041: Strings & Strconv Deep Dive
// Advanced string manipulation with strings.Builder, strings.Replacer, strings.NewReader,
// strings.Map, strings.Fields, strings.Cut, and full strconv package for conversions.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func main() {
	fmt.Println("=== STRINGS & STRCONV DEEP DIVE ===")

	// ============================================
	// 1. STRCONV: STRING TO NUMBER CONVERSIONS
	// ============================================
	fmt.Println("\n--- 1. strconv: String to Number Conversions ---")
	fmt.Println(`
strconv provides conversions between strings and basic Go types.

MAIN FUNCTIONS:
  Atoi(s)            -> int, error     (string to int, base 10)
  Itoa(i)            -> string         (int to string, base 10)
  ParseBool(s)       -> bool, error    (accepts 1, t, T, TRUE, true, 0, f, F, FALSE, false)
  ParseInt(s, base, bits)  -> int64, error
  ParseUint(s, base, bits) -> uint64, error
  ParseFloat(s, bits)      -> float64, error

base: 0 = auto-detect (0x, 0o, 0b prefixes), 2-36
bits: 0, 8, 16, 32, or 64 (size for range check)`)

	// Atoi / Itoa
	fmt.Println("\n  Atoi / Itoa:")
	n, err := strconv.Atoi("42")
	fmt.Printf("    Atoi(\"42\") = %d, err = %v\n", n, err)

	n, err = strconv.Atoi("not_a_number")
	fmt.Printf("    Atoi(\"not_a_number\") = %d, err = %v\n", n, err)

	s := strconv.Itoa(255)
	fmt.Printf("    Itoa(255) = %q\n", s)

	// ParseBool
	fmt.Println("\n  ParseBool:")
	for _, input := range []string{"1", "t", "TRUE", "true", "0", "f", "FALSE", "false", "yes"} {
		b, err := strconv.ParseBool(input)
		if err != nil {
			fmt.Printf("    ParseBool(%q) = error: %v\n", input, err)
		} else {
			fmt.Printf("    ParseBool(%q) = %v\n", input, b)
		}
	}

	// ParseInt with different bases
	fmt.Println("\n  ParseInt with bases:")
	i64, _ := strconv.ParseInt("FF", 16, 64)
	fmt.Printf("    ParseInt(\"FF\", 16, 64) = %d\n", i64)

	i64, _ = strconv.ParseInt("11111111", 2, 64)
	fmt.Printf("    ParseInt(\"11111111\", 2, 64) = %d\n", i64)

	i64, _ = strconv.ParseInt("377", 8, 64)
	fmt.Printf("    ParseInt(\"377\", 8, 64) = %d\n", i64)

	// Auto-detect base with 0
	i64, _ = strconv.ParseInt("0xFF", 0, 64)
	fmt.Printf("    ParseInt(\"0xFF\", 0, 64) = %d (auto hex)\n", i64)

	i64, _ = strconv.ParseInt("0b11111111", 0, 64)
	fmt.Printf("    ParseInt(\"0b11111111\", 0, 64) = %d (auto bin)\n", i64)

	i64, _ = strconv.ParseInt("0o377", 0, 64)
	fmt.Printf("    ParseInt(\"0o377\", 0, 64) = %d (auto oct)\n", i64)

	// ParseUint
	fmt.Println("\n  ParseUint:")
	u64, _ := strconv.ParseUint("18446744073709551615", 10, 64)
	fmt.Printf("    ParseUint max uint64 = %d\n", u64)

	// ParseFloat
	fmt.Println("\n  ParseFloat:")
	f64, _ := strconv.ParseFloat("3.14159265358979", 64)
	fmt.Printf("    ParseFloat(\"3.14159265358979\", 64) = %f\n", f64)

	f64, _ = strconv.ParseFloat("1.23e10", 64)
	fmt.Printf("    ParseFloat(\"1.23e10\", 64) = %g\n", f64)

	f64, _ = strconv.ParseFloat("NaN", 64)
	fmt.Printf("    ParseFloat(\"NaN\", 64) = %v\n", f64)

	f64, _ = strconv.ParseFloat("+Inf", 64)
	fmt.Printf("    ParseFloat(\"+Inf\", 64) = %v\n", f64)

	// ============================================
	// 2. STRCONV: NUMBER TO STRING (FORMAT)
	// ============================================
	fmt.Println("\n--- 2. strconv: Number to String (Format) ---")
	os.Stdout.WriteString(`
FORMAT FUNCTIONS:
  FormatBool(b)               -> "true" or "false"
  FormatInt(i, base)          -> string in given base
  FormatUint(u, base)         -> string in given base
  FormatFloat(f, fmt, prec, bits) -> string
    fmt: 'f' (decimal), 'e' (exponent), 'g' (auto), 'b' (binary exp)
    prec: precision (-1 = smallest needed)
    bits: 32 or 64
`)

	fmt.Println("  FormatBool:")
	fmt.Printf("    FormatBool(true) = %q\n", strconv.FormatBool(true))
	fmt.Printf("    FormatBool(false) = %q\n", strconv.FormatBool(false))

	fmt.Println("\n  FormatInt (different bases):")
	fmt.Printf("    FormatInt(255, 10) = %q\n", strconv.FormatInt(255, 10))
	fmt.Printf("    FormatInt(255, 16) = %q\n", strconv.FormatInt(255, 16))
	fmt.Printf("    FormatInt(255, 2)  = %q\n", strconv.FormatInt(255, 2))
	fmt.Printf("    FormatInt(255, 8)  = %q\n", strconv.FormatInt(255, 8))
	fmt.Printf("    FormatInt(255, 36) = %q\n", strconv.FormatInt(255, 36))

	fmt.Println("\n  FormatFloat:")
	pi := 3.14159265358979
	fmt.Printf("    FormatFloat(pi, 'f', 2, 64)  = %q\n", strconv.FormatFloat(pi, 'f', 2, 64))
	fmt.Printf("    FormatFloat(pi, 'f', 10, 64) = %q\n", strconv.FormatFloat(pi, 'f', 10, 64))
	fmt.Printf("    FormatFloat(pi, 'e', 4, 64)  = %q\n", strconv.FormatFloat(pi, 'e', 4, 64))
	fmt.Printf("    FormatFloat(pi, 'g', -1, 64) = %q\n", strconv.FormatFloat(pi, 'g', -1, 64))

	// ============================================
	// 3. STRCONV: APPEND FUNCTIONS
	// ============================================
	fmt.Println("\n--- 3. strconv: Append Functions ---")
	fmt.Println(`
Append functions write directly to a []byte, avoiding allocations.
Useful for building output in performance-critical code.

  AppendBool(dst, bool)
  AppendInt(dst, int64, base)
  AppendUint(dst, uint64, base)
  AppendFloat(dst, float64, fmt, prec, bits)
  AppendQuote(dst, string)`)

	buf := []byte("value: ")
	buf = strconv.AppendInt(buf, 42, 10)
	fmt.Printf("\n  AppendInt: %q\n", string(buf))

	buf2 := []byte("pi=")
	buf2 = strconv.AppendFloat(buf2, 3.14, 'f', 2, 64)
	fmt.Printf("  AppendFloat: %q\n", string(buf2))

	buf3 := []byte("flag=")
	buf3 = strconv.AppendBool(buf3, true)
	fmt.Printf("  AppendBool: %q\n", string(buf3))

	buf4 := make([]byte, 0, 64)
	buf4 = append(buf4, "items: "...)
	for i := 0; i < 5; i++ {
		if i > 0 {
			buf4 = append(buf4, ',')
		}
		buf4 = strconv.AppendInt(buf4, int64(i*10), 10)
	}
	fmt.Printf("  Building list: %q\n", string(buf4))

	// ============================================
	// 4. STRCONV: QUOTE / UNQUOTE
	// ============================================
	fmt.Println("\n--- 4. strconv: Quote / Unquote ---")
	fmt.Println(`
Quote functions add Go string literal quoting:
  Quote(s)           -> Go double-quoted string
  QuoteToASCII(s)    -> Same but non-ASCII escaped
  QuoteRune(r)       -> Go single-quoted rune literal
  Unquote(s)         -> Reverses Quote
  UnquoteChar(s, q)  -> Decode first char of quoted string`)

	fmt.Printf("\n  Quote(\"hello\\tworld\") = %s\n", strconv.Quote("hello\tworld"))
	fmt.Printf("  Quote(\"cafe\\u0301\") = %s\n", strconv.Quote("caf\u00e9"))
	fmt.Printf("  QuoteToASCII(\"cafe\\u0301\") = %s\n", strconv.QuoteToASCII("caf\u00e9"))
	fmt.Printf("  QuoteRune('A') = %s\n", strconv.QuoteRune('A'))
	fmt.Printf("  QuoteRune('\\n') = %s\n", strconv.QuoteRune('\n'))

	unquoted, _ := strconv.Unquote(`"hello\tworld"`)
	fmt.Printf("  Unquote(`\"hello\\tworld\"`) = %q\n", unquoted)

	// ============================================
	// 5. STRCONV: ERROR HANDLING
	// ============================================
	fmt.Println("\n--- 5. strconv: Error Handling ---")
	fmt.Println(`
strconv errors implement *strconv.NumError:
  type NumError struct {
      Func string  // the function that failed
      Num  string  // the input string
      Err  error   // the reason (ErrRange, ErrSyntax)
  }

Sentinel errors:
  strconv.ErrRange  -> value out of range
  strconv.ErrSyntax -> invalid syntax`)

	_, err = strconv.Atoi("abc")
	if numErr, ok := err.(*strconv.NumError); ok {
		fmt.Printf("\n  NumError: Func=%q, Num=%q, Err=%v\n", numErr.Func, numErr.Num, numErr.Err)
		fmt.Printf("  Is ErrSyntax: %v\n", numErr.Err == strconv.ErrSyntax)
	}

	_, err = strconv.ParseInt("999999999999999999999", 10, 64)
	if numErr, ok := err.(*strconv.NumError); ok {
		fmt.Printf("  Range error: Func=%q, Err=%v\n", numErr.Func, numErr.Err)
		fmt.Printf("  Is ErrRange: %v\n", numErr.Err == strconv.ErrRange)
	}

	// ============================================
	// 6. STRINGS.BUILDER: EFFICIENT CONCATENATION
	// ============================================
	fmt.Println("\n--- 6. strings.Builder: Efficient Concatenation ---")
	fmt.Println(`
strings.Builder minimizes memory copying when building strings.
It implements io.Writer and has methods:
  WriteString(s)  -> append a string
  WriteByte(b)    -> append a byte
  WriteRune(r)    -> append a rune
  Write(p)        -> implement io.Writer
  String()        -> get the result
  Len()           -> current length
  Cap()           -> current capacity
  Reset()         -> clear the builder
  Grow(n)         -> pre-allocate n bytes

A Builder must NOT be copied after first use.`)

	var builder strings.Builder
	builder.Grow(100)
	for i := 0; i < 10; i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString("item_")
		builder.WriteString(strconv.Itoa(i))
	}
	fmt.Printf("\n  Builder result: %s\n", builder.String())
	fmt.Printf("  Builder len=%d, cap=%d\n", builder.Len(), builder.Cap())

	// ============================================
	// 7. PERFORMANCE: CONCAT VS BUILDER
	// ============================================
	fmt.Println("\n--- 7. Performance: String Concat vs Builder ---")

	const iterations = 10000

	// Naive concatenation
	start := time.Now()
	result := ""
	for i := 0; i < iterations; i++ {
		result += "x"
	}
	naiveTime := time.Since(start)

	// strings.Builder
	start = time.Now()
	var b strings.Builder
	for i := 0; i < iterations; i++ {
		b.WriteString("x")
	}
	_ = b.String()
	builderTime := time.Since(start)

	// strings.Builder with Grow
	start = time.Now()
	var bg strings.Builder
	bg.Grow(iterations)
	for i := 0; i < iterations; i++ {
		bg.WriteString("x")
	}
	_ = bg.String()
	builderGrowTime := time.Since(start)

	fmt.Printf("\n  Naive concat (%d iterations): %v\n", iterations, naiveTime)
	fmt.Printf("  strings.Builder:              %v\n", builderTime)
	fmt.Printf("  strings.Builder with Grow:    %v\n", builderGrowTime)
	fmt.Println("  (Builder is significantly faster due to fewer allocations)")

	// ============================================
	// 8. STRINGS.REPLACER
	// ============================================
	fmt.Println("\n--- 8. strings.Replacer ---")
	fmt.Println(`
strings.NewReplacer creates a reusable replacer from old/new pairs.
It is safe for concurrent use and more efficient than chained Replace calls.

  r := strings.NewReplacer(old1, new1, old2, new2, ...)
  r.Replace(s)       -> returns replaced string
  r.WriteString(w,s) -> writes replaced string to io.Writer`)

	htmlEscaper := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	input := `<script>alert("xss") & 'evil'</script>`
	escaped := htmlEscaper.Replace(input)
	fmt.Printf("\n  Original: %s\n", input)
	fmt.Printf("  Escaped:  %s\n", escaped)

	envReplacer := strings.NewReplacer(
		"${HOME}", "/home/user",
		"${USER}", "gopher",
		"${PORT}", "8080",
	)
	template := "Server at ${HOME}/app by ${USER} on port ${PORT}"
	fmt.Printf("\n  Template: %s\n", template)
	fmt.Printf("  Expanded: %s\n", envReplacer.Replace(template))

	// WriteString to stdout
	fmt.Print("  WriteString: ")
	envReplacer.WriteString(os.Stdout, "Hello ${USER}!\n")

	// ============================================
	// 9. STRINGS.NEWREADER
	// ============================================
	fmt.Println("\n--- 9. strings.NewReader ---")
	fmt.Println(`
strings.NewReader creates an io.Reader from a string.
It implements: io.Reader, io.ReaderAt, io.Seeker, io.WriterTo, io.ByteScanner, io.RuneScanner.

Useful for passing strings to functions expecting io.Reader.`)

	reader := strings.NewReader("Hello, Go World!")
	fmt.Printf("\n  Length: %d\n", reader.Len())
	fmt.Printf("  Size: %d\n", reader.Size())

	buf5 := make([]byte, 5)
	n2, _ := reader.Read(buf5)
	fmt.Printf("  Read %d bytes: %q\n", n2, string(buf5[:n2]))
	fmt.Printf("  Remaining: %d\n", reader.Len())

	// Seek and read again
	reader.Seek(0, 0)
	n2, _ = reader.Read(buf5)
	fmt.Printf("  After Seek(0,0), Read: %q\n", string(buf5[:n2]))

	// ReadAt (random access)
	n2, _ = reader.ReadAt(buf5, 7)
	fmt.Printf("  ReadAt(offset=7): %q\n", string(buf5[:n2]))

	// ============================================
	// 10. STRINGS.MAP
	// ============================================
	fmt.Println("\n--- 10. strings.Map ---")
	fmt.Println(`
strings.Map applies a mapping function to each rune.
  func Map(mapping func(rune) rune, s string) string

Return -1 to remove the rune from output.`)

	// ROT13
	rot13 := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return 'a' + (r-'a'+13)%26
		case r >= 'A' && r <= 'Z':
			return 'A' + (r-'A'+13)%26
		default:
			return r
		}
	}, "Hello, World!")
	fmt.Printf("\n  ROT13(\"Hello, World!\") = %q\n", rot13)

	// Remove non-printable characters
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, "Hello\x00World\x01!\x02")
	fmt.Printf("  Remove non-printable: %q\n", cleaned)

	// To upper (manual)
	upper := strings.Map(unicode.ToUpper, "hello gopher")
	fmt.Printf("  ToUpper via Map: %q\n", upper)

	// ============================================
	// 11. STRINGS.FIELDS AND FIELDSFUNCTION
	// ============================================
	fmt.Println("\n--- 11. strings.Fields and FieldsFunc ---")
	fmt.Println(`
strings.Fields splits on whitespace (space, tab, newline, etc).
Unlike Split(" "), it handles multiple consecutive whitespace.

strings.FieldsFunc splits on runes where the function returns true.`)

	text := "  hello   world  \t  go  \n  lang  "
	fields := strings.Fields(text)
	fmt.Printf("\n  Input: %q\n", text)
	fmt.Printf("  Fields: %v (count: %d)\n", fields, len(fields))

	csvLine := "name;age;city;country"
	parts := strings.FieldsFunc(csvLine, func(r rune) bool {
		return r == ';'
	})
	fmt.Printf("\n  FieldsFunc on %q: %v\n", csvLine, parts)

	// Split on multiple delimiters
	mixed := "hello,world;foo:bar"
	multiParts := strings.FieldsFunc(mixed, func(r rune) bool {
		return r == ',' || r == ';' || r == ':'
	})
	fmt.Printf("  FieldsFunc multiple delims: %v\n", multiParts)

	// ============================================
	// 12. STRINGS.CUT (GO 1.18+)
	// ============================================
	fmt.Println("\n--- 12. strings.Cut (Go 1.18+) ---")
	fmt.Println(`
strings.Cut slices s around the first instance of sep:
  func Cut(s, sep string) (before, after string, found bool)

It replaces many uses of Index + slicing. Cleaner and safer.`)

	before, after, found := strings.Cut("user@example.com", "@")
	fmt.Printf("\n  Cut(\"user@example.com\", \"@\"):\n")
	fmt.Printf("    before=%q, after=%q, found=%v\n", before, after, found)

	before, after, found = strings.Cut("KEY=VALUE=extra", "=")
	fmt.Printf("  Cut(\"KEY=VALUE=extra\", \"=\"):\n")
	fmt.Printf("    before=%q, after=%q, found=%v\n", before, after, found)

	before, after, found = strings.Cut("no-separator-here", "@")
	fmt.Printf("  Cut(\"no-separator-here\", \"@\"):\n")
	fmt.Printf("    before=%q, after=%q, found=%v\n", before, after, found)

	// Practical: Parse HTTP header
	header := "Content-Type: application/json; charset=utf-8"
	name, value, _ := strings.Cut(header, ": ")
	fmt.Printf("\n  HTTP Header: name=%q, value=%q\n", name, value)

	// ============================================
	// 13. STRINGS.CUTPREFIX / CUTSUFFIX (GO 1.20+)
	// ============================================
	fmt.Println("\n--- 13. strings.CutPrefix / CutSuffix (Go 1.20+) ---")
	fmt.Println(`
  CutPrefix(s, prefix) (after string, found bool)
  CutSuffix(s, suffix) (before string, found bool)

Replaces the pattern: if strings.HasPrefix(s, p) { s = s[len(p):] }`)

	if after, ok := strings.CutPrefix("/api/v2/users", "/api/v2"); ok {
		fmt.Printf("\n  CutPrefix(\"/api/v2/users\", \"/api/v2\"): %q\n", after)
	}

	if after, ok := strings.CutPrefix("/api/v1/items", "/api/v2"); !ok {
		fmt.Printf("  CutPrefix(\"/api/v1/items\", \"/api/v2\"): not found, original=%q\n", after)
	}

	if before, ok := strings.CutSuffix("document.pdf", ".pdf"); ok {
		fmt.Printf("  CutSuffix(\"document.pdf\", \".pdf\"): %q\n", before)
	}

	if before, ok := strings.CutSuffix("image.png", ".pdf"); !ok {
		fmt.Printf("  CutSuffix(\"image.png\", \".pdf\"): not found, original=%q\n", before)
	}

	// Practical: Parse file extensions
	files := []string{"main.go", "README.md", "Makefile", "test.go", "config.yaml"}
	fmt.Println("\n  Go files:")
	for _, f := range files {
		if name, ok := strings.CutSuffix(f, ".go"); ok {
			fmt.Printf("    %s (base: %q)\n", f, name)
		}
	}

	// ============================================
	// 14. BASE CONVERSIONS WITH STRCONV
	// ============================================
	fmt.Println("\n--- 14. Base Conversions with strconv ---")
	os.Stdout.WriteString(`
Common pattern: convert between different number bases.

  FormatInt(n, base)      -> number to string in any base (2-36)
  ParseInt(s, base, bits) -> string in any base to int64
  FormatUint(n, base)     -> unsigned version
`)

	num := int64(255)
	fmt.Printf("  Decimal %d in different bases:\n", num)
	fmt.Printf("    Binary:  %s\n", strconv.FormatInt(num, 2))
	fmt.Printf("    Octal:   %s\n", strconv.FormatInt(num, 8))
	fmt.Printf("    Hex:     %s\n", strconv.FormatInt(num, 16))
	fmt.Printf("    Base-36: %s\n", strconv.FormatInt(num, 36))

	// Convert between bases
	hexStr := "1A3F"
	parsed, _ := strconv.ParseInt(hexStr, 16, 64)
	fmt.Printf("\n  Hex %q -> decimal: %d\n", hexStr, parsed)
	fmt.Printf("  Hex %q -> binary:  %s\n", hexStr, strconv.FormatInt(parsed, 2))
	fmt.Printf("  Hex %q -> octal:   %s\n", hexStr, strconv.FormatInt(parsed, 8))

	// ============================================
	// 15. PRACTICAL EXAMPLES
	// ============================================
	fmt.Println("\n--- 15. Practical Examples ---")

	// CSV-like parser using Cut
	fmt.Println("\n  CSV Parser using Cut:")
	record := "John,30,New York,Engineer"
	var fieldList []string
	remaining := record
	for {
		field, rest, found := strings.Cut(remaining, ",")
		fieldList = append(fieldList, field)
		if !found {
			break
		}
		remaining = rest
	}
	fmt.Printf("    Input: %q\n", record)
	fmt.Printf("    Fields: %v\n", fieldList)

	// URL query parser
	fmt.Println("\n  URL Query Parser:")
	query := "name=gopher&age=10&city=gotham&debug=true"
	params := make(map[string]string)
	for _, pair := range strings.Split(query, "&") {
		key, val, _ := strings.Cut(pair, "=")
		params[key] = val
	}
	fmt.Printf("    Query: %q\n", query)
	for k, v := range params {
		fmt.Printf("    %s = %s\n", k, v)
	}

	// Number formatting utility
	fmt.Println("\n  Format file sizes:")
	sizes := []int64{0, 1023, 1024, 1048576, 1073741824, 1099511627776}
	for _, size := range sizes {
		fmt.Printf("    %d bytes = %s\n", size, formatSize(size))
	}

	fmt.Println("\n=== End of Chapter 041 ===")
}

/*
SUMMARY - Chapter 041: Strings & Strconv Deep Dive

STRCONV PARSING:
- Atoi/Itoa: string<->int conversion
- ParseBool, ParseInt, ParseUint, ParseFloat: typed parsing
- ParseInt supports bases 2-36, auto-detect with base 0
- Errors implement *strconv.NumError (ErrRange, ErrSyntax)

STRCONV FORMATTING:
- FormatBool, FormatInt, FormatUint, FormatFloat: typed formatting
- FormatFloat: 'f' decimal, 'e' exponent, 'g' auto
- AppendInt/AppendFloat/AppendBool: zero-alloc append to []byte

STRCONV QUOTING:
- Quote/QuoteToASCII: Go string literal quoting
- QuoteRune: single rune quoting
- Unquote: reverse of Quote

STRINGS.BUILDER:
- Efficient string concatenation (implements io.Writer)
- WriteString, WriteByte, WriteRune, Write methods
- Grow(n) pre-allocates capacity
- Must NOT be copied after first use

STRINGS.REPLACER:
- NewReplacer creates reusable replacer from old/new pairs
- Safe for concurrent use
- WriteString writes to io.Writer directly

STRINGS.MAP:
- Apply function to each rune
- Return -1 to remove rune from output

STRINGS.FIELDS / FIELDSFUNC:
- Fields splits on whitespace (handles multiple consecutive)
- FieldsFunc splits on custom rune predicate

STRINGS.CUT (GO 1.18+):
- Cut(s, sep) -> before, after, found
- CutPrefix/CutSuffix (Go 1.20+)
- Cleaner alternative to Index + slicing
*/

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	var b strings.Builder
	switch {
	case bytes >= TB:
		b.WriteString(strconv.FormatFloat(float64(bytes)/float64(TB), 'f', 2, 64))
		b.WriteString(" TB")
	case bytes >= GB:
		b.WriteString(strconv.FormatFloat(float64(bytes)/float64(GB), 'f', 2, 64))
		b.WriteString(" GB")
	case bytes >= MB:
		b.WriteString(strconv.FormatFloat(float64(bytes)/float64(MB), 'f', 2, 64))
		b.WriteString(" MB")
	case bytes >= KB:
		b.WriteString(strconv.FormatFloat(float64(bytes)/float64(KB), 'f', 2, 64))
		b.WriteString(" KB")
	default:
		b.WriteString(strconv.FormatInt(bytes, 10))
		b.WriteString(" B")
	}
	return b.String()
}
