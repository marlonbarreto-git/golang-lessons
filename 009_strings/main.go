// Package main - Chapter 009: Strings
// Los strings en Go son secuencias de bytes inmutables.
// Aprenderás sobre UTF-8, runes, bytes y el paquete strings.
package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== STRINGS EN GO ===")

	// ============================================
	// STRINGS BÁSICOS
	// ============================================
	fmt.Println("\n--- Strings Básicos ---")

	// Strings son secuencias de bytes inmutables
	s1 := "Hola, Mundo"
	s2 := "Hello, 世界" // soporta Unicode nativo

	fmt.Printf("s1: %s (len: %d bytes)\n", s1, len(s1))
	fmt.Printf("s2: %s (len: %d bytes)\n", s2, len(s2))

	// len() devuelve bytes, no caracteres
	// "世界" = 6 bytes (3 bytes por caracter chino en UTF-8)

	// Contar caracteres reales (runes)
	fmt.Printf("s2 runes: %d\n", utf8.RuneCountInString(s2))

	// ============================================
	// STRING LITERALS
	// ============================================
	fmt.Println("\n--- String Literals ---")

	// Interpreted strings (comillas dobles)
	// Procesan secuencias de escape
	interpreted := "Línea 1\nLínea 2\tTabulado"
	fmt.Println("Interpreted:")
	fmt.Println(interpreted)

	// Raw strings (backticks)
	// No procesan escapes, preservan todo
	raw := `Línea 1\nNo es escape
	Esta es línea 2
	Preserva espacios y tabs`
	fmt.Println("\nRaw:")
	fmt.Println(raw)

	// Secuencias de escape comunes
	fmt.Println("\n--- Secuencias de Escape ---")
	fmt.Println("\\n = newline")
	fmt.Println("\\t = tab")
	fmt.Println("\\r = carriage return")
	fmt.Println("\\\\ = backslash")
	fmt.Println("\\\" = comilla doble")
	fmt.Printf("\\x41 = %s (hex byte)\n", "\x41")
	fmt.Printf("\\u4e16 = %s (unicode 16-bit)\n", "\u4e16")
	fmt.Printf("\\U0001F600 = %s (unicode 32-bit)\n", "\U0001F600")

	// ============================================
	// BYTES VS RUNES
	// ============================================
	fmt.Println("\n--- Bytes vs Runes ---")

	texto := "Go语言"

	// Iterar como bytes
	fmt.Println("Como bytes:")
	for i := 0; i < len(texto); i++ {
		fmt.Printf("  [%d] byte: %d, char: %c\n", i, texto[i], texto[i])
	}

	// Iterar como runes (range hace conversión automática)
	fmt.Println("\nComo runes:")
	for i, r := range texto {
		fmt.Printf("  [%d] rune: %d, char: %c, bytes: %d\n",
			i, r, r, utf8.RuneLen(r))
	}

	// Conversiones
	fmt.Println("\n--- Conversiones ---")

	str := "Hello"
	bytes := []byte(str)     // string a []byte
	runes := []rune(str)     // string a []rune
	strBack := string(bytes) // []byte a string
	runeStr := string(runes) // []rune a string

	fmt.Printf("string: %s\n", str)
	fmt.Printf("[]byte: %v\n", bytes)
	fmt.Printf("[]rune: %v\n", runes)
	fmt.Printf("back to string: %s, %s\n", strBack, runeStr)

	// ============================================
	// STRINGS SON INMUTABLES
	// ============================================
	fmt.Println("\n--- Inmutabilidad ---")

	original := "Hello"
	// original[0] = 'h' // NO COMPILA: cannot assign

	// Para modificar, convertir a []byte o []rune
	modificable := []byte(original)
	modificable[0] = 'h'
	modificado := string(modificable)
	fmt.Printf("Original: %s\n", original)
	fmt.Printf("Modificado: %s\n", modificado)

	// ============================================
	// PAQUETE STRINGS
	// ============================================
	fmt.Println("\n--- Paquete strings ---")

	s := "  Hello, World!  "

	// Búsqueda
	fmt.Printf("Contains 'World': %t\n", strings.Contains(s, "World"))
	fmt.Printf("HasPrefix '  He': %t\n", strings.HasPrefix(s, "  He"))
	fmt.Printf("HasSuffix '!  ': %t\n", strings.HasSuffix(s, "!  "))
	fmt.Printf("Index 'World': %d\n", strings.Index(s, "World"))
	fmt.Printf("LastIndex 'o': %d\n", strings.LastIndex(s, "o"))
	fmt.Printf("Count 'l': %d\n", strings.Count(s, "l"))

	// Transformación
	fmt.Printf("ToUpper: %s\n", strings.ToUpper(s))
	fmt.Printf("ToLower: %s\n", strings.ToLower(s))
	fmt.Printf("TrimSpace: '%s'\n", strings.TrimSpace(s))
	fmt.Printf("Trim '! ': '%s'\n", strings.Trim(s, "! "))
	fmt.Printf("Replace: %s\n", strings.Replace(s, "World", "Gopher", 1))
	fmt.Printf("ReplaceAll: %s\n", strings.ReplaceAll(s, "l", "L"))

	// Title case (depreciado, usar cases package para Unicode correcto)
	// fmt.Printf("Title: %s\n", strings.Title(s)) // depreciado

	// Split y Join
	fmt.Println("\n--- Split y Join ---")

	csv := "apple,banana,cherry,date"
	partes := strings.Split(csv, ",")
	fmt.Printf("Split: %v\n", partes)

	unido := strings.Join(partes, " - ")
	fmt.Printf("Join: %s\n", unido)

	// SplitN limita número de partes
	limited := strings.SplitN(csv, ",", 2)
	fmt.Printf("SplitN(2): %v\n", limited)

	// Fields (split por espacios)
	frase := "uno   dos    tres"
	palabras := strings.Fields(frase)
	fmt.Printf("Fields: %v\n", palabras)

	// ============================================
	// STRINGS.BUILDER (eficiente para concatenación)
	// ============================================
	fmt.Println("\n--- strings.Builder ---")

	// MAL: concatenación con + (crea muchos strings temporales)
	// resultado := ""
	// for i := 0; i < 1000; i++ { resultado += "x" }

	// BIEN: usar strings.Builder
	var builder strings.Builder
	for i := 0; i < 5; i++ {
		builder.WriteString("Go")
		builder.WriteByte(' ')
	}
	resultado := builder.String()
	fmt.Printf("Builder result: %s\n", resultado)

	// Builder con capacidad inicial (mejor performance)
	builder2 := strings.Builder{}
	builder2.Grow(100) // pre-alocar espacio
	builder2.WriteString("Hello")
	builder2.WriteRune(' ')
	builder2.WriteString("世界")
	fmt.Printf("Builder2: %s\n", builder2.String())

	// ============================================
	// STRINGS.READER
	// ============================================
	fmt.Println("\n--- strings.Reader ---")

	reader := strings.NewReader("Hello, World!")
	buffer := make([]byte, 5)

	n, _ := reader.Read(buffer)
	fmt.Printf("Read %d bytes: %s\n", n, buffer)
	fmt.Printf("Remaining: %d bytes\n", reader.Len())

	// ============================================
	// COMPARACIÓN
	// ============================================
	fmt.Println("\n--- Comparación ---")

	a := "apple"
	b := "banana"
	c := "Apple"

	fmt.Printf("%s == %s: %t\n", a, a, a == a)
	fmt.Printf("%s < %s: %t\n", a, b, a < b) // lexicográfica

	// Comparación case-insensitive
	fmt.Printf("EqualFold(%s, %s): %t\n", a, c, strings.EqualFold(a, c))

	// Compare (-1, 0, 1)
	fmt.Printf("Compare(%s, %s): %d\n", a, b, strings.Compare(a, b))

	// ============================================
	// UNICODE
	// ============================================
	fmt.Println("\n--- Unicode ---")

	emoji := "Hello 👋 World 🌍"
	fmt.Printf("String: %s\n", emoji)
	fmt.Printf("Bytes: %d\n", len(emoji))
	fmt.Printf("Runes: %d\n", utf8.RuneCountInString(emoji))

	// Validar UTF-8
	valid := "Hello"
	invalid := string([]byte{0xff, 0xfe})
	fmt.Printf("Valid UTF-8 '%s': %t\n", valid, utf8.ValidString(valid))
	fmt.Printf("Valid UTF-8 (invalid): %t\n", utf8.ValidString(invalid))

	// Propiedades Unicode
	for _, r := range "Hello 世界 123" {
		props := []string{}
		if unicode.IsLetter(r) {
			props = append(props, "letter")
		}
		if unicode.IsDigit(r) {
			props = append(props, "digit")
		}
		if unicode.IsSpace(r) {
			props = append(props, "space")
		}
		if unicode.Is(unicode.Han, r) {
			props = append(props, "Han (Chinese)")
		}
		if len(props) > 0 {
			fmt.Printf("  %c: %v\n", r, props)
		}
	}

	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrones Comunes ---")

	// 1. Verificar string vacío
	emptyStr := ""
	if emptyStr == "" {
		fmt.Println("String vacío (usar == \"\")")
	}
	// Alternativa: len(s) == 0

	// 2. Substring seguro
	longStr := "Hello, World!"
	start, end := 0, 5
	if end <= len(longStr) {
		fmt.Printf("Substring [%d:%d]: %s\n", start, end, longStr[start:end])
	}

	// 3. Formatear strings
	nombre := "Gopher"
	edad := 15
	formatted := fmt.Sprintf("Hola %s, tienes %d años", nombre, edad)
	fmt.Printf("Sprintf: %s\n", formatted)

	// 4. Multiline string con formato
	template := `
Nombre: %s
Edad: %d
Email: %s
`
	fmt.Printf(template, "Alice", 30, "alice@example.com")

	// 5. Repetir string
	separador := strings.Repeat("-", 20)
	fmt.Println(separador)

	// 6. Padding
	padded := fmt.Sprintf("|%-10s|%10s|", "left", "right")
	fmt.Println(padded)

	// 7. Truncar con ellipsis
	largo := "Este es un texto muy largo que necesita truncarse"
	maxLen := 20
	if len(largo) > maxLen {
		truncado := largo[:maxLen-3] + "..."
		fmt.Printf("Truncado: %s\n", truncado)
	}
}

/*
RESUMEN:

CONCEPTOS CLAVE:
- Strings son []byte inmutables (UTF-8 encoded)
- byte = uint8, rune = int32 (Unicode code point)
- len() devuelve bytes, no caracteres
- range sobre string itera runes

PAQUETES IMPORTANTES:
- strings: manipulación general
- strconv: conversión a/desde otros tipos
- unicode: propiedades de caracteres
- unicode/utf8: operaciones UTF-8

BUENAS PRÁCTICAS:

1. Usa strings.Builder para concatenación en loops
2. Usa range para iterar caracteres Unicode
3. Verifica largo en bytes vs runes según el caso
4. Usa strings.EqualFold para comparación case-insensitive
5. Prefiere string literals sobre concatenación
6. Raw strings (``) para regex, paths, y texto multilínea
*/
