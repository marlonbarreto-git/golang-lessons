// Package main - Chapter 044: Regexp
// Expresiones regulares en Go: compilacion, matching, captura,
// reemplazo, RE2 engine, patrones comunes y performance.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	fmt.Println("=== REGEXP PACKAGE DEEP DIVE EN GO ===")

	// ============================================
	// 1. REGEXP BASICS
	// ============================================
	fmt.Println("\n--- 1. regexp Basics ---")
	fmt.Println(`
Go usa el motor RE2 para expresiones regulares.
RE2 garantiza tiempo lineal O(n) - no hay backtracking catastrofico.

DOS FORMAS DE COMPILAR:

  regexp.Compile(pattern)      -> (*Regexp, error)
  regexp.MustCompile(pattern)  -> *Regexp (panic si invalido)

REGLA:
  - MustCompile para patrones constantes conocidos en compile time
  - Compile para patrones que vienen del usuario (input dinamico)

PRECOMPILACION (package-level var):
  var emailRe = regexp.MustCompile(` + "`" + `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` + "`" + `)

  // Compilar regexp es costoso. Hacerlo una vez y reutilizar.
  // regexp.Regexp es safe para uso concurrente.`)

	re := regexp.MustCompile(`\d+`)
	fmt.Printf("\n  Pattern: %s\n", re.String())
	fmt.Printf("  Match '123': %v\n", re.MatchString("abc123def"))
	fmt.Printf("  Match 'abc': %v\n", re.MatchString("abc"))

	_, err := regexp.Compile(`[invalid`)
	fmt.Printf("  Compile error: %v\n", err)

	// ============================================
	// 2. PATTERN SYNTAX
	// ============================================
	fmt.Println("\n--- 2. Pattern Syntax ---")
	fmt.Println(`
RE2 SYNTAX (subconjunto de PCRE):

CARACTERES:
  .          Cualquier caracter (excepto \n por defecto)
  \d         Digito [0-9]
  \D         No digito
  \w         Word char [0-9A-Za-z_]
  \W         No word char
  \s         Whitespace [\t\n\f\r ]
  \S         No whitespace
  \b         Word boundary
  \B         No word boundary

CUANTIFICADORES:
  *          0 o mas (greedy)
  +          1 o mas (greedy)
  ?          0 o 1 (greedy)
  {n}        Exactamente n
  {n,}       n o mas
  {n,m}      Entre n y m
  *?         0 o mas (lazy/non-greedy)
  +?         1 o mas (lazy)
  ??         0 o 1 (lazy)

GRUPOS Y ALTERNACION:
  (abc)      Grupo de captura
  (?:abc)    Grupo sin captura
  (?P<name>) Grupo con nombre
  a|b        Alternacion (a o b)

CLASES DE CARACTERES:
  [abc]      Cualquiera de a, b, c
  [a-z]      Rango a-z
  [^abc]     Cualquiera excepto a, b, c
  [[:alpha:]] Clase POSIX (letras)
  [[:digit:]] Clase POSIX (digitos)

ANCLAS:
  ^          Inicio de linea
  $          Final de linea
  \A         Inicio de string (no multilinea)
  \z         Final de string (no multilinea)

FLAGS (en el patron):
  (?i)       Case insensitive
  (?m)       Multilinea (^ y $ match \n)
  (?s)       . tambien matchea \n
  (?U)       Ungreedy (lazy por defecto)

NO SOPORTADO EN RE2:
  - Lookahead (?=...) y lookbehind (?<=...)
  - Backreferences \1, \2
  - Possessive quantifiers ++, *+
  - Atomic groups (?>...)`)

	patterns := []struct {
		pattern, input string
	}{
		{`\d{3}-\d{4}`, "Tel: 555-1234"},
		{`(?i)hello`, "Hello World"},
		{`^Go\w+`, "Golang is great"},
		{`\b\w{5}\b`, "The quick brown fox jumps"},
		{`[aeiou]`, "Hello"},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		match := re.FindString(p.input)
		fmt.Printf("  %-25s in %-30s -> %q\n", p.pattern, p.input, match)
	}

	// ============================================
	// 3. MATCHSTRING - SIMPLE MATCHING
	// ============================================
	fmt.Println("\n--- 3. MatchString - Simple Matching ---")
	os.Stdout.WriteString(`
FUNCIONES DE MATCHING:

  re.MatchString(s)         -> bool (hay match en s?)
  re.Match([]byte)          -> bool
  regexp.MatchString(pat,s) -> (bool, error) (sin precompilar)

NOTA: MatchString verifica si hay match EN ALGUNA PARTE del string.
Para match completo, usar ^ y $:

  re := regexp.MustCompile(` + "`" + `\d+` + "`" + `)
  re.MatchString("abc123") -> true  (match parcial)

  re := regexp.MustCompile(` + "`" + `^\d+$` + "`" + `)
  re.MatchString("abc123") -> false (requiere match completo)
`)

	emailRe := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	emails := []string{
		"user@example.com",
		"user.name+tag@domain.co",
		"invalid@",
		"@domain.com",
		"user@.com",
		"valid_user@sub.domain.org",
	}
	for _, email := range emails {
		fmt.Printf("  %-30s valid: %v\n", email, emailRe.MatchString(email))
	}

	// ============================================
	// 4. FIND FUNCTIONS
	// ============================================
	fmt.Println("\n--- 4. Find Functions ---")
	os.Stdout.WriteString(`
FAMILIA DE FUNCIONES FIND:

  FindString(s)            -> string (primer match)
  FindAllString(s, n)      -> []string (todos los matches, n=-1 para todos)
  FindStringIndex(s)       -> []int{start, end}
  FindAllStringIndex(s, n) -> [][]int

  FindStringSubmatch(s)         -> []string (match + subgrupos)
  FindAllStringSubmatch(s, n)   -> [][]string
  FindStringSubmatchIndex(s)    -> []int

VARIANTES:
  - Sin "String": trabajan con []byte
  - Con "All": retornan todos los matches
  - Con "Submatch": incluyen grupos de captura
  - Con "Index": retornan posiciones en vez de strings

PARAMETRO n:
  n > 0  -> Maximo n matches
  n == -1 -> Todos los matches
`)

	text := "Precios: $10.50, $23.99, $5.00 y $199.95"
	priceRe := regexp.MustCompile(`\$\d+\.\d{2}`)

	fmt.Printf("  Text: %q\n", text)
	fmt.Printf("  FindString:       %q\n", priceRe.FindString(text))
	fmt.Printf("  FindAllString(-1): %v\n", priceRe.FindAllString(text, -1))
	fmt.Printf("  FindAllString(2):  %v\n", priceRe.FindAllString(text, 2))

	idx := priceRe.FindStringIndex(text)
	fmt.Printf("  FindStringIndex:   %v (text[%d:%d] = %q)\n",
		idx, idx[0], idx[1], text[idx[0]:idx[1]])

	allIdx := priceRe.FindAllStringIndex(text, -1)
	fmt.Printf("  FindAllStringIndex: %v\n", allIdx)

	// ============================================
	// 5. NAMED CAPTURE GROUPS
	// ============================================
	fmt.Println("\n--- 5. Named Capture Groups ---")
	os.Stdout.WriteString(`
GRUPOS CON NOMBRE - (?P<name>pattern):

  re := regexp.MustCompile(` + "`" + `(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})` + "`" + `)
  match := re.FindStringSubmatch("2024-03-15")
  // match[0] = "2024-03-15" (match completo)
  // match[1] = "2024"       (grupo "year")
  // match[2] = "03"         (grupo "month")
  // match[3] = "15"         (grupo "day")

  re.SubexpNames() -> ["", "year", "month", "day"]
  // Indice 0 es siempre "" (match completo)

PATRON PARA EXTRAER A MAP:

  func extractNamedGroups(re *regexp.Regexp, s string) map[string]string {
      match := re.FindStringSubmatch(s)
      if match == nil {
          return nil
      }
      result := make(map[string]string)
      for i, name := range re.SubexpNames() {
          if i != 0 && name != "" {
              result[name] = match[i]
          }
      }
      return result
  }
`)

	dateRe := regexp.MustCompile(`(?P<year>\d{4})-(?P<month>\d{2})-(?P<day>\d{2})`)
	match := dateRe.FindStringSubmatch("Event on 2024-03-15 was great")

	fmt.Printf("  Pattern: %s\n", dateRe.String())
	fmt.Printf("  SubexpNames: %v\n", dateRe.SubexpNames())

	if match != nil {
		fmt.Printf("  Full match: %q\n", match[0])
		for i, name := range dateRe.SubexpNames() {
			if i != 0 && name != "" {
				fmt.Printf("  Group %q (index %d): %q\n", name, i, match[i])
			}
		}
	}

	fmt.Println("\n  [Extract to map]")
	groups := extractNamedGroups(dateRe, "Date: 2024-12-25")
	for k, v := range groups {
		fmt.Printf("  %s = %s\n", k, v)
	}

	logRe := regexp.MustCompile(`(?P<level>\w+)\s+\[(?P<timestamp>[^\]]+)\]\s+(?P<message>.+)`)
	logGroups := extractNamedGroups(logRe, "ERROR [2024-03-15 14:30:00] connection timeout")
	fmt.Println("\n  [Log parsing]")
	for k, v := range logGroups {
		fmt.Printf("  %s = %q\n", k, v)
	}

	// ============================================
	// 6. REPLACE FUNCTIONS
	// ============================================
	fmt.Println("\n--- 6. Replace Functions ---")
	os.Stdout.WriteString(`
FUNCIONES DE REEMPLAZO:

  re.ReplaceAllString(src, repl)          -> string
  re.ReplaceAllLiteralString(src, repl)   -> string (sin expansion de $)
  re.ReplaceAllStringFunc(src, fn)        -> string (funcion custom)

EXPANSION EN repl:
  $0 o ${0}    -> Match completo
  $1 o ${1}    -> Primer grupo
  $name o ${name} -> Grupo con nombre

NOTA: ReplaceAllLiteralString NO expande $1, ${name}, etc.
`)

	src := "John Smith, age 30; Jane Doe, age 25"
	nameRe := regexp.MustCompile(`(\w+) (\w+)`)
	reversed := nameRe.ReplaceAllString(src, "${2}, ${1}")
	fmt.Printf("  Original:  %q\n", src)
	fmt.Printf("  Reversed:  %q\n", reversed)

	sensitiveRe := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	masked := sensitiveRe.ReplaceAllString(
		"SSN: 123-45-6789 and 987-65-4321",
		"***-**-****")
	fmt.Printf("  Masked:    %q\n", masked)

	literal := sensitiveRe.ReplaceAllLiteralString(
		"SSN: 123-45-6789", "$REDACTED")
	fmt.Printf("  Literal:   %q\n", literal)

	fmt.Println("\n  [ReplaceAllStringFunc]")
	wordRe := regexp.MustCompile(`\b[a-z]+\b`)
	titled := wordRe.ReplaceAllStringFunc(
		"hello world from go",
		func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		})
	fmt.Printf("  Title case: %q\n", titled)

	// Named group replacement
	dateStr := "Today is 2024-03-15 and tomorrow is 2024-03-16"
	dateReplace := dateRe.ReplaceAllString(dateStr, "${day}/${month}/${year}")
	fmt.Printf("  Date format: %q\n", dateReplace)

	// ============================================
	// 7. SPLIT FUNCTIONS
	// ============================================
	fmt.Println("\n--- 7. Split Functions ---")
	os.Stdout.WriteString(`
SPLIT CON REGEXP:

  re.Split(s, n)   -> []string
    n > 0  : maximo n substrings
    n == -1: todos
    n == 0 : retorna nil

Similar a strings.Split pero con regexp como separador.
`)

	splitRe := regexp.MustCompile(`[,;\s]+`)
	parts := splitRe.Split("apple, banana; cherry  date,fig", -1)
	fmt.Printf("  Split by [,;\\s]+: %v\n", parts)

	csvRe := regexp.MustCompile(`,\s*`)
	csv := csvRe.Split("name, age,  city,   country", -1)
	fmt.Printf("  CSV split:       %v\n", csv)

	lineRe := regexp.MustCompile(`\r?\n`)
	lines := lineRe.Split("line1\nline2\r\nline3\nline4", -1)
	fmt.Printf("  Line split:      %v\n", lines)

	limitRe := regexp.MustCompile(`\s+`)
	limited := limitRe.Split("one two three four five", 3)
	fmt.Printf("  Split(s, 3):     %v\n", limited)

	// ============================================
	// 8. PRECOMPILED REGEXP PATTERNS
	// ============================================
	fmt.Println("\n--- 8. Precompiled Regexp (Package-Level) ---")
	os.Stdout.WriteString(`
PATRON RECOMENDADO - Precompilar a nivel de paquete:

var (
    emailRegex   = regexp.MustCompile(` + "`" + `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` + "`" + `)
    phoneRegex   = regexp.MustCompile(` + "`" + `^\+?(\d{1,3})?[-.\s]?\(?\d{1,4}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,9}$` + "`" + `)
    urlRegex     = regexp.MustCompile(` + "`" + `^https?://[^\s/$.?#].[^\s]*$` + "`" + `)
    ipv4Regex    = regexp.MustCompile(` + "`" + `^(\d{1,3}\.){3}\d{1,3}$` + "`" + `)
    slugRegex    = regexp.MustCompile(` + "`" + `^[a-z0-9]+(-[a-z0-9]+)*$` + "`" + `)
    uuidRegex    = regexp.MustCompile(` + "`" + `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$` + "`" + `)
    hexColorRegex = regexp.MustCompile(` + "`" + `^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$` + "`" + `)
)

VENTAJAS:
  - Se compila una sola vez al iniciar el programa
  - Panic en init si el patron es invalido (bug detectado temprano)
  - Safe para uso concurrente desde multiples goroutines
  - Sin overhead de compilacion en runtime

ANTI-PATRON:
  func validate(email string) bool {
      re := regexp.MustCompile(` + "`" + `...` + "`" + `) // MAL: recompila cada llamada
      return re.MatchString(email)
  }
`)

	// ============================================
	// 9. COMMON PATTERNS
	// ============================================
	fmt.Println("\n--- 9. Common Patterns ---")

	demoCommonPatterns()

	// ============================================
	// 10. RE2 ENGINE - LINEAR TIME GUARANTEE
	// ============================================
	fmt.Println("\n--- 10. RE2 Engine - Linear Time Guarantee ---")
	fmt.Println(`
Go usa RE2 (Regular Expression 2) de Russ Cox / Google.

BACKTRACKING vs RE2:

  PCRE/Java/Python (backtracking):
  - Patron: (a+)+b
  - Input:  "aaaaaaaaaaaaaaaaac"
  - Resultado: CATASTROFICO - O(2^n) tiempo
  - Puede freezar el programa (ReDoS attack)

  Go RE2 (automata-based):
  - Mismo patron, mismo input
  - Resultado: O(n) siempre
  - No hay backtracking catastrofico

TRADEOFF:
  - RE2 NO soporta: lookahead, lookbehind, backreferences
  - RE2 SI garantiza: tiempo lineal, predecible, safe

IMPLICACION DE SEGURIDAD:
  - Go es immune a ReDoS (Regular Expression Denial of Service)
  - Seguro aceptar patrones de usuarios (con Compile, no MustCompile)
  - Otros lenguajes necesitan timeouts o sanitizacion de patrones`)

	fmt.Println("\n  [Demo: RE2 linear time]")
	dangerousPattern := regexp.MustCompile(`(a+)+b`)
	evilInput := strings.Repeat("a", 30) + "c"
	matched := dangerousPattern.MatchString(evilInput)
	fmt.Printf("  Pattern (a+)+b with 30 a's + c: matched=%v (instant in Go)\n", matched)
	fmt.Println("  In PCRE/Java/Python, this pattern would take exponential time!")

	// ============================================
	// 11. REGEXP vs STRINGS PACKAGE
	// ============================================
	fmt.Println("\n--- 11. regexp vs strings Package ---")
	os.Stdout.WriteString(`
CUANDO USAR strings (mas rapido):
  - Buscar/reemplazar texto literal
  - Split por delimitador fijo
  - Prefix/suffix check
  - Contains/Count

  strings.Contains(s, "error")       // mas rapido que re.MatchString
  strings.ReplaceAll(s, "old", "new") // mas rapido que re.ReplaceAll
  strings.Split(s, ",")              // mas rapido que re.Split

CUANDO USAR regexp:
  - Patrones variables (digitos, rangos, opcionales)
  - Captura de subgrupos
  - Validacion compleja (email, URL, etc.)
  - Reemplazo con backreferences ($1, $name)
  - Multiples delimitadores

REGLA: Si puedes resolverlo con strings, hazlo.
regexp es 5-100x mas lento que operaciones de strings equivalentes.
`)

	input := "Error: connection failed at 10:30:45"

	fmt.Println("  [Comparison: Extract time from log]")
	fmt.Printf("  Input: %q\n", input)

	if strings.Contains(input, "Error") {
		fmt.Println("  strings.Contains 'Error': true")
	}

	timeRe := regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2})`)
	timeMatch := timeRe.FindStringSubmatch(input)
	if timeMatch != nil {
		fmt.Printf("  regexp time extract: %s:%s:%s\n", timeMatch[1], timeMatch[2], timeMatch[3])
	}

	// ============================================
	// 12. ADVANCED PATTERNS
	// ============================================
	fmt.Println("\n--- 12. Advanced Patterns ---")

	demoAdvancedPatterns()

	// ============================================
	// 13. PERFORMANCE TIPS
	// ============================================
	fmt.Println("\n--- 13. Performance Tips ---")
	os.Stdout.WriteString(`
TIPS DE PERFORMANCE:

1. PRECOMPILAR SIEMPRE:
   var re = regexp.MustCompile(...)  // Una vez
   // NO: regexp.MustCompile(...) dentro de funciones

2. USAR strings CUANDO POSIBLE:
   strings.Contains > regexp para texto literal
   strings.HasPrefix > regexp con ^
   strings.HasSuffix > regexp con $

3. ANCLAR PATRONES:
   ^pattern$ es mas rapido que pattern
   El motor no necesita buscar en cada posicion

4. EVITAR .* CUANDO POSIBLE:
   .*foo.* es lento (muchas posiciones)
   Mejor: strings.Contains(s, "foo")

5. USAR (?:...) EN VEZ DE (...):
   Grupos sin captura son mas rapidos
   Solo capturar lo que necesitas

6. FindString vs MatchString:
   Si solo necesitas bool, usa MatchString
   FindString hace mas trabajo (extrae el match)

7. BYTE OPERATIONS:
   Match([]byte) es mas rapido que MatchString(string)
   Evita conversion string <-> []byte

8. BENCHMARK:
   func BenchmarkRegexp(b *testing.B) {
       re := regexp.MustCompile(` + "`" + `\d{3}-\d{4}` + "`" + `)
       input := "Phone: 555-1234"
       for i := 0; i < b.N; i++ {
           re.MatchString(input)
       }
   }
`)

	// ============================================
	// WORKING DEMO: TEXT PROCESSOR
	// ============================================
	fmt.Println("\n--- WORKING DEMO: Text Processor ---")
	demoTextProcessor()

	fmt.Println("\n=== RESUMEN EJECUTADO ===")
	fmt.Println("Todos los demos de regexp ejecutados exitosamente.")
}

// ============================================
// HELPERS
// ============================================

func extractNamedGroups(re *regexp.Regexp, s string) map[string]string {
	match := re.FindStringSubmatch(s)
	if match == nil {
		return nil
	}
	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

// ============================================
// DEMO: COMMON PATTERNS
// ============================================

var (
	emailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlPattern      = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	ipv4Pattern     = regexp.MustCompile(`^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$`)
	phonePattern    = regexp.MustCompile(`^\+?\d{1,3}[-.\s]?\(?\d{1,4}\)?[-.\s]?\d{1,4}[-.\s]?\d{1,9}$`)
	hexColorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	uuidPattern     = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	slugPattern     = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

func demoCommonPatterns() {
	type testCase struct {
		name    string
		re      *regexp.Regexp
		inputs  []string
		expects []bool
	}

	tests := []testCase{
		{
			name: "Email",
			re:   emailPattern,
			inputs: []string{
				"user@example.com",
				"name+tag@domain.co.uk",
				"invalid@",
				"@domain.com",
			},
			expects: []bool{true, true, false, false},
		},
		{
			name: "IPv4",
			re:   ipv4Pattern,
			inputs: []string{
				"192.168.1.1",
				"255.255.255.255",
				"0.0.0.0",
				"256.1.1.1",
				"1.2.3",
			},
			expects: []bool{true, true, true, false, false},
		},
		{
			name: "Hex Color",
			re:   hexColorPattern,
			inputs: []string{
				"#fff",
				"#FF5733",
				"#12345g",
				"123456",
			},
			expects: []bool{true, true, false, false},
		},
		{
			name: "UUID",
			re:   uuidPattern,
			inputs: []string{
				"550e8400-e29b-41d4-a716-446655440000",
				"invalid-uuid",
			},
			expects: []bool{true, false},
		},
		{
			name: "Slug",
			re:   slugPattern,
			inputs: []string{
				"hello-world",
				"my-blog-post-123",
				"Hello-World",
				"--invalid",
				"valid",
			},
			expects: []bool{true, true, false, false, true},
		},
	}

	for _, tc := range tests {
		fmt.Printf("  [%s]\n", tc.name)
		for i, input := range tc.inputs {
			result := tc.re.MatchString(input)
			status := "PASS"
			if result != tc.expects[i] {
				status = "FAIL"
			}
			fmt.Printf("    %-40s -> %-5v %s\n", input, result, status)
		}
	}
}

// ============================================
// DEMO: ADVANCED PATTERNS
// ============================================

func demoAdvancedPatterns() {
	fmt.Println("  1. Extract all URLs from text:")
	text := `Check out https://golang.org and http://example.com/path?q=1 for more info.
Also see https://github.com/golang/go`
	urlExtract := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlExtract.FindAllString(text, -1)
	for _, u := range urls {
		fmt.Printf("     %s\n", u)
	}

	fmt.Println("\n  2. Parse key=value pairs:")
	kvRe := regexp.MustCompile(`(?P<key>\w+)=(?P<value>[^\s,]+)`)
	config := "host=localhost,port=5432,dbname=mydb,sslmode=disable"
	matches := kvRe.FindAllStringSubmatch(config, -1)
	for _, m := range matches {
		groups := make(map[string]string)
		for i, name := range kvRe.SubexpNames() {
			if i != 0 && name != "" {
				groups[name] = m[i]
			}
		}
		fmt.Printf("     %s => %s\n", groups["key"], groups["value"])
	}

	fmt.Println("\n  3. Sanitize HTML tags:")
	htmlRe := regexp.MustCompile(`<[^>]+>`)
	html := "<p>Hello <b>World</b></p><script>alert('xss')</script>"
	sanitized := htmlRe.ReplaceAllString(html, "")
	fmt.Printf("     Before: %q\n", html)
	fmt.Printf("     After:  %q\n", sanitized)

	fmt.Println("\n  4. Extract structured data (log parsing):")
	logRe := regexp.MustCompile(
		`(?P<ip>\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+-\s+-\s+` +
			`\[(?P<date>[^\]]+)\]\s+"(?P<method>\w+)\s+(?P<path>[^\s]+)\s+HTTP/[^"]+"\s+` +
			`(?P<status>\d{3})\s+(?P<size>\d+)`)

	logLine := `192.168.1.100 - - [15/Mar/2024:14:30:00 +0000] "GET /api/users HTTP/1.1" 200 1234`
	logMatch := logRe.FindStringSubmatch(logLine)
	if logMatch != nil {
		for i, name := range logRe.SubexpNames() {
			if i != 0 && name != "" {
				fmt.Printf("     %s: %s\n", name, logMatch[i])
			}
		}
	}

	fmt.Println("\n  5. Password strength validation:")
	password := "MyP@ssw0rd!"
	checks := []struct {
		name    string
		pattern string
	}{
		{"Has uppercase", `[A-Z]`},
		{"Has lowercase", `[a-z]`},
		{"Has digit", `[0-9]`},
		{"Has special char", `[!@#$%^&*(),.?":{}|<>]`},
		{"Min 8 chars", `.{8,}`},
	}

	fmt.Printf("     Password: %q\n", password)
	allPassed := true
	for _, c := range checks {
		re := regexp.MustCompile(c.pattern)
		passed := re.MatchString(password)
		if !passed {
			allPassed = false
		}
		fmt.Printf("     %-20s: %v\n", c.name, passed)
	}
	fmt.Printf("     Overall strong: %v\n", allPassed)
}

// ============================================
// DEMO: TEXT PROCESSOR
// ============================================

func demoTextProcessor() {
	text := `
Contact us at support@example.com or sales@company.org.
Visit https://example.com/products for details.
US office: +1-555-123-4567
UK office: +44-20-7946-0958
Meeting on 2024-03-15 at 14:30.
Server IP: 192.168.1.100
Order #ORD-2024-00142 shipped.
`
	fmt.Printf("  Input text:\n%s\n", text)

	processors := []struct {
		name    string
		pattern string
	}{
		{"Emails", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`},
		{"URLs", `https?://[^\s]+`},
		{"Phone numbers", `\+?\d{1,3}-\d{2,3}-\d{3,4}-\d{4}`},
		{"Dates (YYYY-MM-DD)", `\d{4}-\d{2}-\d{2}`},
		{"Times (HH:MM)", `\d{2}:\d{2}`},
		{"IP addresses", `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`},
		{"Order numbers", `ORD-\d{4}-\d{5}`},
	}

	for _, p := range processors {
		re := regexp.MustCompile(p.pattern)
		found := re.FindAllString(text, -1)
		fmt.Printf("  %-25s found %d: %v\n", p.name, len(found), found)
	}

	fmt.Println("\n  [Redaction demo]")
	redactors := []struct {
		name        string
		pattern     string
		replacement string
	}{
		{"Emails", `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, "[EMAIL REDACTED]"},
		{"Phones", `\+?\d{1,3}-\d{2,3}-\d{3,4}-\d{4}`, "[PHONE REDACTED]"},
		{"IPs", `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`, "[IP REDACTED]"},
	}

	redacted := text
	for _, r := range redactors {
		re := regexp.MustCompile(r.pattern)
		redacted = re.ReplaceAllString(redacted, r.replacement)
	}
	fmt.Printf("  Redacted text:\n%s\n", redacted)

	fmt.Println("  [Word frequency with regexp]")
	wordRe := regexp.MustCompile(`\b[a-zA-Z]+\b`)
	words := wordRe.FindAllString(strings.ToLower(text), -1)
	freq := make(map[string]int)
	for _, w := range words {
		freq[w]++
	}
	fmt.Printf("  Total words: %d\n", len(words))
	fmt.Printf("  Unique words: %d\n", len(freq))
	fmt.Println("  Top words with count > 1:")
	for word, count := range freq {
		if count > 1 {
			fmt.Printf("    %q: %d\n", word, count)
		}
	}
}

/*
RESUMEN CAPITULO 71: REGEXP PACKAGE DEEP DIVE

COMPILACION:
- regexp.Compile(pattern) -> (*Regexp, error) para input dinamico
- regexp.MustCompile(pattern) -> *Regexp (panic) para constantes
- Precompilar como var a nivel de paquete (safe para concurrencia)

RE2 ENGINE:
- Garantia de tiempo lineal O(n) - sin backtracking catastrofico
- Immune a ReDoS (Regular Expression Denial of Service)
- Tradeoff: no soporta lookahead, lookbehind, backreferences
- Seguro aceptar patrones de usuarios

PATTERN SYNTAX:
- Caracteres: . \d \w \s \b
- Cuantificadores: * + ? {n} {n,m} (greedy) y *? +? ?? (lazy)
- Grupos: (capture) (?:non-capture) (?P<named>)
- Clases: [abc] [a-z] [^abc] [[:alpha:]]
- Flags: (?i) case insensitive, (?m) multiline, (?s) dotall

MATCHING:
- MatchString(s) -> bool (match parcial)
- ^pattern$ para match completo
- FindString -> primer match
- FindAllString(s, -1) -> todos los matches
- FindStringSubmatch -> match + grupos de captura

NAMED GROUPS:
- (?P<name>pattern) para captura con nombre
- SubexpNames() retorna nombres de los grupos
- Extraer a map con helper function

REPLACE:
- ReplaceAllString(src, repl) con $1, ${name}
- ReplaceAllLiteralString sin expansion
- ReplaceAllStringFunc para transformaciones custom

SPLIT:
- re.Split(s, n) como strings.Split pero con regexp
- n=-1 para todos, n>0 para limitar

COMMON PATTERNS:
- Email, URL, IPv4, phone, hex color, UUID, slug
- Siempre precompilar como package-level var

REGEXP vs STRINGS:
- strings es 5-100x mas rapido para operaciones literales
- Usar regexp solo cuando necesitas patrones
- strings.Contains > re.MatchString para texto fijo

PERFORMANCE:
- Precompilar siempre (regexp.MustCompile a nivel de paquete)
- Anclar patrones con ^ $ cuando sea posible
- Usar (?:...) en vez de (...) si no necesitas captura
- MatchString es mas rapido que FindString si solo necesitas bool
- Match([]byte) es mas rapido que MatchString(string)
*/

/* SUMMARY - CHAPTER 044: Regular Expressions
Topics covered in this chapter:

• Regexp compilation: Compile vs MustCompile, precompiling patterns at package level
• RE2 engine: linear time guarantee, no exponential backtracking, safe for user input
• Pattern syntax: character classes, quantifiers (greedy and lazy), anchors, word boundaries
• Matching operations: MatchString, FindString, FindAllString, FindStringSubmatch
• Capture groups: positional and named groups using (?P<name>pattern)
• SubexpNames: extracting group names and building maps from matches
• Replace operations: ReplaceAllString with $1/${name} expansion, ReplaceAllStringFunc
• Split functionality: regexp-based splitting with limit control
• Common patterns: email, URL, IPv4, phone, UUID, hex color, slug validation
• regexp vs strings: when to use each, performance characteristics (strings is 5-100x faster)
• Performance optimization: precompilation, anchoring, non-capturing groups, byte operations
• Best practices: package-level vars, avoid recompilation, prefer strings for literals
*/
