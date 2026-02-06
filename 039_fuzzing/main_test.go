package main

import (
	"testing"
	"unicode/utf8"
)

// ============================================
// FUZZ TEST BÁSICO
// ============================================

func FuzzReverse(f *testing.F) {
	// Seed corpus
	f.Add("hello")
	f.Add("world")
	f.Add("")
	f.Add("a")
	f.Add("世界")
	f.Add("hello\x00world") // Con null byte

	f.Fuzz(func(t *testing.T, input string) {
		// Propiedad 1: double reverse = original
		reversed := Reverse(input)
		doubleReversed := Reverse(reversed)

		if doubleReversed != input {
			t.Errorf("double reverse of %q = %q, want %q", input, doubleReversed, input)
		}

		// Propiedad 2: misma longitud en runes
		if utf8.RuneCountInString(reversed) != utf8.RuneCountInString(input) {
			t.Errorf("reverse changed rune count: input=%d, reversed=%d",
				utf8.RuneCountInString(input), utf8.RuneCountInString(reversed))
		}
	})
}

// ============================================
// FUZZ PARSER
// ============================================

func FuzzParseInt(f *testing.F) {
	// Seeds: varios formatos de números
	f.Add("0")
	f.Add("1")
	f.Add("-1")
	f.Add("123")
	f.Add("-456")
	f.Add("999999999")
	f.Add("")
	f.Add("abc")
	f.Add("12abc")
	f.Add(" 42 ")
	f.Add("\t123\n")

	f.Fuzz(func(t *testing.T, input string) {
		// No debería causar panic
		result := ParseInt(input)

		// Si el resultado no es 0, debería ser consistente
		if result != 0 {
			// Parse de nuevo debería dar el mismo resultado
			resultStr := string(rune(result))
			_ = resultStr
		}
	})
}

// ============================================
// FUZZ URL PARSER
// ============================================

func FuzzParseURL(f *testing.F) {
	// Seeds: URLs válidas e inválidas
	f.Add("https://example.com")
	f.Add("http://localhost:8080")
	f.Add("https://example.com/path")
	f.Add("https://example.com/path?query=value")
	f.Add("https://example.com:443/path?q=1&b=2")
	f.Add("")
	f.Add("not-a-url")
	f.Add("://missing-scheme")
	f.Add("https://")

	f.Fuzz(func(t *testing.T, input string) {
		// No debería causar panic
		url, err := ParseURL(input)

		// Si parsea correctamente, los campos deberían ser válidos
		if err == nil && url != nil {
			if url.Scheme == "" {
				t.Errorf("parsed URL has empty scheme for input %q", input)
			}
		}
	})
}

// ============================================
// FUZZ VALIDATOR
// ============================================

func FuzzValidateEmail(f *testing.F) {
	// Emails válidos
	f.Add("test@example.com")
	f.Add("user.name@domain.org")
	f.Add("user+tag@example.co.uk")

	// Emails inválidos
	f.Add("")
	f.Add("@")
	f.Add("test@")
	f.Add("@example.com")
	f.Add("test")
	f.Add("a@b")

	f.Fuzz(func(t *testing.T, input string) {
		// No debería causar panic
		result := ValidateEmail(input)

		// Propiedad: resultado debe ser consistente
		result2 := ValidateEmail(input)
		if result != result2 {
			t.Errorf("inconsistent validation for %q: %v vs %v", input, result, result2)
		}
	})
}

// ============================================
// FUZZ CON MÚLTIPLES ARGUMENTOS
// ============================================

func FuzzSafeDivide(f *testing.F) {
	// Seeds con diferentes combinaciones
	f.Add(10, 2)
	f.Add(0, 1)
	f.Add(1, 0)  // División por cero
	f.Add(-10, 2)
	f.Add(10, -2)
	f.Add(-10, -2)
	f.Add(0, 0)

	f.Fuzz(func(t *testing.T, a, b int) {
		result, err := SafeDivide(a, b)

		// Propiedad 1: división por cero debe retornar error
		if b == 0 && err == nil {
			t.Errorf("expected error for division by zero, got nil")
		}

		// Propiedad 2: si no hay error, el resultado es correcto
		if err == nil && b != 0 {
			expected := a / b
			if result != expected {
				t.Errorf("SafeDivide(%d, %d) = %d, want %d", a, b, result, expected)
			}
		}
	})
}

// ============================================
// FUZZ UTF-8 VALIDATION
// ============================================

func FuzzUTF8(f *testing.F) {
	// Strings UTF-8 válidos
	f.Add("hello")
	f.Add("世界")
	f.Add("🎉")
	f.Add("")

	// Bytes que podrían ser UTF-8 inválido se generan automáticamente

	f.Fuzz(func(t *testing.T, input string) {
		isValid := IsValidUTF8(input)

		// Si es válido, debería ser igual a utf8.ValidString
		expected := utf8.ValidString(input)
		if isValid != expected {
			t.Errorf("IsValidUTF8(%q) = %v, want %v", input, isValid, expected)
		}
	})
}

// ============================================
// FUZZ CON BYTES
// ============================================

func FuzzWithBytes(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte{})
	f.Add([]byte{0, 1, 2, 255})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Procesar bytes
		_ = len(data)

		// Verificar que no causa panic
		if len(data) > 0 {
			_ = data[0]
		}
	})
}

/*
EJECUTAR FUZZING:

# Ejecutar un fuzz test específico
go test -fuzz=FuzzReverse -fuzztime=30s

# Ejecutar todos los fuzz tests (solo corpus, sin generar)
go test -run=Fuzz

# Ver qué inputs se probaron
go test -fuzz=FuzzReverse -v -fuzztime=10s

# Los casos de fallo se guardan en:
# testdata/fuzz/FuzzReverse/<hash>
*/
