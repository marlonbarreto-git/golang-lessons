// Package main - Chapter 039: Fuzzing
// Go 1.18+ incluye fuzzing nativo en el paquete testing.
// El fuzzing encuentra bugs generando inputs aleatorios automáticamente.
package main

import (
	"os"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== FUZZING EN GO ===")

	os.Stdout.WriteString(`
Fuzzing es una técnica de testing que genera inputs aleatorios
para encontrar bugs, panics, y edge cases inesperados.

Go 1.18+ incluye fuzzing nativo en el paquete testing.

ANATOMÍA DE UN FUZZ TEST:

func FuzzMyFunction(f *testing.F) {
    // 1. Seed corpus: ejemplos iniciales
    f.Add("hello")
    f.Add("")
    f.Add("special\x00chars")

    // 2. Fuzz function
    f.Fuzz(func(t *testing.T, input string) {
        // Probar la función con input generado
        result := MyFunction(input)

        // Verificar propiedades (no valores específicos)
        if result < 0 {
            t.Errorf("got negative result for input %q", input)
        }
    })
}

EJECUTAR FUZZING:

# Fuzzear por 30 segundos
go test -fuzz=FuzzMyFunction -fuzztime=30s

# Fuzzear hasta encontrar un fallo
go test -fuzz=FuzzMyFunction

# Ejecutar solo con corpus existente (no genera nuevos)
go test -run=FuzzMyFunction

CORPUS:
- Los casos que causan fallos se guardan en testdata/fuzz/<FuzzName>
- Estos casos se ejecutan como tests normales
- Git commit los casos de fallo para regression testing
`)

	// ============================================
	// FUNCIONES PARA FUZZEAR
	// ============================================
	fmt.Println("\n--- Funciones para Fuzzear ---")

	// Ejemplo 1: Parser de enteros
	fmt.Println("\n1. ParseInt:")
	fmt.Printf("   ParseInt('123') = %d\n", ParseInt("123"))
	fmt.Printf("   ParseInt('-456') = %d\n", ParseInt("-456"))
	fmt.Printf("   ParseInt('abc') = %d\n", ParseInt("abc"))

	// Ejemplo 2: Reverse string
	fmt.Println("\n2. Reverse:")
	fmt.Printf("   Reverse('hello') = %s\n", Reverse("hello"))
	fmt.Printf("   Reverse('世界') = %s\n", Reverse("世界"))

	// Ejemplo 3: Validar email
	fmt.Println("\n3. ValidateEmail:")
	fmt.Printf("   ValidateEmail('test@example.com') = %v\n", ValidateEmail("test@example.com"))
	fmt.Printf("   ValidateEmail('invalid') = %v\n", ValidateEmail("invalid"))

	// Ejemplo 4: URL parser
	fmt.Println("\n4. ParseURL:")
	url, err := ParseURL("https://example.com:8080/path?query=value")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Scheme: %s, Host: %s, Port: %s\n", url.Scheme, url.Host, url.Port)
	}

	// ============================================
	// PROPIEDADES A VERIFICAR
	// ============================================
	fmt.Println("\n--- Propiedades para Fuzzing ---")
	os.Stdout.WriteString(`
En fuzzing, verificamos PROPIEDADES, no valores específicos:

1. NO PANICS:
   defer func() {
       if r := recover(); r != nil {
           t.Errorf("panic: %v", r)
       }
   }()
   MyFunction(input)

2. ROUND-TRIP (encode/decode):
   encoded := Encode(input)
   decoded := Decode(encoded)
   if decoded != input {
       t.Errorf("round-trip failed")
   }

3. IDEMPOTENCIA:
   result1 := Process(input)
   result2 := Process(Process(input))
   // Si Process es idempotente, result1 == result2

4. INVARIANTES:
   result := Reverse(Reverse(input))
   if result != input {
       t.Errorf("double reverse should equal original")
   }

5. NO OUT OF BOUNDS:
   result := GetElement(slice, index)
   // No debería causar panic por index fuera de rango

6. COMPARACIÓN CON IMPLEMENTACIÓN DE REFERENCIA:
   result := MySort(input)
   expected := sort.Strings(input)
   if !reflect.DeepEqual(result, expected) {
       t.Errorf("results differ")
   }
`)

	// ============================================
	// TIPOS DE INPUT SOPORTADOS
	// ============================================
	fmt.Println("\n--- Tipos de Input ---")
	fmt.Println(`
Tipos soportados para f.Add() y f.Fuzz():
- string
- []byte
- int, int8, int16, int32, int64
- uint, uint8, uint16, uint32, uint64
- float32, float64
- bool
- rune

Múltiples argumentos:
f.Add("hello", 42, true)
f.Fuzz(func(t *testing.T, s string, n int, b bool) {
    // ...
})`)
	// ============================================
	// CORPUS MANAGEMENT
	// ============================================
	fmt.Println("\n--- Gestión del Corpus ---")
	fmt.Println(`
ESTRUCTURA:
testdata/
└── fuzz/
    └── FuzzMyFunction/
        ├── seed1           # Casos seed
        ├── seed2
        └── 5d41402abc4b2...# Casos encontrados por fuzzer

AGREGAR SEEDS MANUALMENTE:
1. Crear archivo en testdata/fuzz/FuzzName/
2. El contenido es el input serializado

LIMPIAR CORPUS:
go clean -fuzzcache

EJECUTAR SOLO CORPUS (sin generar nuevos):
go test -run=FuzzMyFunction  # Sin -fuzz flag`)
	// ============================================
	// FUZZING CON ESTRUCTURAS COMPLEJAS
	// ============================================
	fmt.Println("\n--- Fuzzing con Estructuras Complejas ---")
	os.Stdout.WriteString(`
Para tipos complejos, usa []byte y deserializa:

func FuzzJSONParser(f *testing.F) {
    f.Add([]byte("{}"))
    f.Add([]byte("{\"key\": \"value\"}"))

    f.Fuzz(func(t *testing.T, data []byte) {
        var result map[string]any
        err := json.Unmarshal(data, &result)

        // Si parsea sin error, debería serializar de vuelta
        if err == nil {
            _, err := json.Marshal(result)
            if err != nil {
                t.Errorf("failed to re-marshal: %v", err)
            }
        }
    })
}` + "\n")

	// ============================================
	// COVERAGE-GUIDED FUZZING
	// ============================================
	fmt.Println("\n--- Coverage-Guided Fuzzing ---")
	fmt.Println(`
El fuzzer de Go es coverage-guided:
- Instrumenta el código para tracking de cobertura
- Genera inputs que exploran nuevos paths
- Muta inputs existentes (bitflips, inserciones, etc.)
- Prioriza inputs que aumentan cobertura

ESTRATEGIAS DEL FUZZER:
- Bit flipping
- Byte shuffling
- Arithmetic mutations (+1, -1, *2, etc.)
- Known interesting values (0, -1, MAX_INT, etc.)
- Dictionary-based mutations
- Cross-over entre inputs`)
	// ============================================
	// BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Best Practices ---")
	fmt.Println(`
1. SEED CORPUS DIVERSO:
   - Edge cases conocidos
   - Inputs válidos e inválidos
   - Caracteres especiales, Unicode
   - Valores límite (0, -1, MAX)

2. FUZZ PARSERS Y DESERIALIZERS:
   - JSON, XML, Protocol Buffers
   - Cualquier input de usuario

3. FUZZ CRYPTO Y ENCODING:
   - Base64, URL encoding
   - Encryption/decryption round-trips

4. NO ASUMIR NADA DEL INPUT:
   - Puede ser nil, vacío, gigante
   - Caracteres inválidos, UTF-8 malformado

5. VERIFICAR PROPIEDADES, NO VALORES:
   - Round-trip equality
   - Invariantes matemáticas
   - No panics

6. COMMIT FAILING CASES:
   - Git add testdata/fuzz/*
   - Sirven como regression tests`)}

// ============================================
// FUNCIONES PARA FUZZEAR
// ============================================

func ParseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

type URL struct {
	Scheme string
	Host   string
	Port   string
	Path   string
	Query  string
}

func ParseURL(raw string) (*URL, error) {
	if raw == "" {
		return nil, errors.New("empty URL")
	}

	url := &URL{}

	// Scheme
	if idx := strings.Index(raw, "://"); idx > 0 {
		url.Scheme = raw[:idx]
		raw = raw[idx+3:]
	} else {
		return nil, errors.New("missing scheme")
	}

	// Query
	if idx := strings.Index(raw, "?"); idx >= 0 {
		url.Query = raw[idx+1:]
		raw = raw[:idx]
	}

	// Path
	if idx := strings.Index(raw, "/"); idx >= 0 {
		url.Path = raw[idx:]
		raw = raw[:idx]
	}

	// Port
	if idx := strings.LastIndex(raw, ":"); idx >= 0 {
		url.Port = raw[idx+1:]
		raw = raw[:idx]
	}

	// Host
	url.Host = raw

	return url, nil
}

func IsValidUTF8(s string) bool {
	return utf8.ValidString(s)
}

func SafeDivide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

/*
RESUMEN DE FUZZING:

ESTRUCTURA:
func FuzzXxx(f *testing.F) {
    f.Add(seedValue)           // Seed corpus
    f.Fuzz(func(t *testing.T, input Type) {
        // Verificar propiedades
    })
}

COMANDOS:
go test -fuzz=FuzzXxx           # Fuzzear hasta fallo
go test -fuzz=FuzzXxx -fuzztime=1m  # Por 1 minuto
go test -run=FuzzXxx            # Solo corpus existente
go clean -fuzzcache             # Limpiar cache

TIPOS SOPORTADOS:
string, []byte, int*, uint*, float*, bool, rune

QUÉ FUZZEAR:
- Parsers (JSON, XML, custom formats)
- Deserializers
- Validators
- Encoders/Decoders
- Input de usuario

PROPIEDADES A VERIFICAR:
1. No panics
2. Round-trip equality
3. Idempotencia
4. Invariantes
5. Comparación con referencia

CORPUS:
- Seed: f.Add() o testdata/fuzz/FuzzName/
- Fallos se guardan automáticamente
- Commit fallos para regression

BUENAS PRÁCTICAS:
1. Seeds diversos y edge cases
2. Verificar propiedades, no valores
3. Commit failing cases
4. No asumir nada del input
5. Fuzzear boundaries del sistema
*/
