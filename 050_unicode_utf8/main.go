// Package main - Chapter 050: Unicode, UTF-8, and UTF-16 Packages
// unicode/utf8: DecodeRune, EncodeRune, RuneCount, ValidString.
// unicode/utf16: Encode, Decode, IsSurrogate.
// unicode: IsLetter, IsDigit, IsSpace, rune categories and ranges.
package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== UNICODE, UTF-8, AND UTF-16 PACKAGES ===")

	// ============================================
	// 1. STRING ITERATION: BYTES VS RUNES
	// ============================================
	fmt.Println("\n--- 1. String Iteration: Bytes vs Runes ---")
	fmt.Println(`
In Go, a string is a read-only slice of bytes (usually UTF-8).
A rune is an alias for int32, representing a Unicode code point.

Key distinction:
  len(s)           -> number of BYTES
  utf8.RuneCountInString(s) -> number of RUNES (characters)
  for i, b := range []byte(s) -> iterate BYTES
  for i, r := range s         -> iterate RUNES (decodes UTF-8)

Multi-byte characters:
  ASCII (U+0000 to U+007F):  1 byte
  Latin (U+0080 to U+07FF):  2 bytes
  CJK   (U+0800 to U+FFFF):  3 bytes
  Emoji (U+10000+):           4 bytes`)

	text := "Hello, \u4e16\u754c! \U0001F600"
	fmt.Printf("\n  String: %s\n", text)
	fmt.Printf("  Bytes (len):  %d\n", len(text))
	fmt.Printf("  Runes (count): %d\n", utf8.RuneCountInString(text))

	// Byte iteration
	fmt.Printf("\n  Byte iteration: [")
	for i, b := range []byte(text) {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%02x", b)
	}
	fmt.Println("]")

	// Rune iteration
	fmt.Println("\n  Rune iteration:")
	for i, r := range text {
		fmt.Printf("    byte[%2d]: U+%04X '%c' (%d bytes)\n", i, r, r, utf8.RuneLen(r))
	}

	// ============================================
	// 2. UNICODE/UTF8: DECODERUNE
	// ============================================
	fmt.Println("\n--- 2. unicode/utf8: DecodeRune ---")
	fmt.Println(`
DecodeRune decodes the first UTF-8 encoding from a byte slice:
  r, size := utf8.DecodeRune(p []byte)
  r, size := utf8.DecodeRuneInString(s string)

DecodeLastRune decodes the last:
  r, size := utf8.DecodeLastRune(p []byte)
  r, size := utf8.DecodeLastRuneInString(s string)

If encoding is invalid, returns (RuneError, 1).`)

	sample := []byte("Go\u4e16\u754c")
	fmt.Printf("\n  Decoding %q byte by byte:\n", string(sample))
	for len(sample) > 0 {
		r, size := utf8.DecodeRune(sample)
		fmt.Printf("    rune=%c (U+%04X), size=%d bytes\n", r, r, size)
		sample = sample[size:]
	}

	// DecodeRuneInString
	str := "\U0001F4BB laptop"
	r, size := utf8.DecodeRuneInString(str)
	fmt.Printf("\n  First rune of %q: %c (U+%04X), %d bytes\n", str, r, r, size)

	// DecodeLastRune
	r, size = utf8.DecodeLastRuneInString(str)
	fmt.Printf("  Last rune of %q: %c (U+%04X), %d bytes\n", str, r, r, size)

	// Invalid UTF-8
	invalid := []byte{0xff, 0xfe, 0x41}
	r, size = utf8.DecodeRune(invalid)
	fmt.Printf("\n  Invalid bytes [0xff, 0xfe, 0x41]:\n")
	fmt.Printf("    rune=%c (U+%04X), size=%d (RuneError=%v)\n", r, r, size, r == utf8.RuneError)

	// ============================================
	// 3. UNICODE/UTF8: ENCODERUNE
	// ============================================
	fmt.Println("\n--- 3. unicode/utf8: EncodeRune ---")
	fmt.Println(`
EncodeRune writes a rune's UTF-8 encoding to a byte slice:
  n := utf8.EncodeRune(p []byte, r rune)
  p must be at least utf8.UTFMax (4) bytes

AppendRune appends the UTF-8 encoding (Go 1.18+):
  p = utf8.AppendRune(p []byte, r rune)`)

	runes := []rune{'A', '\u00e9', '\u4e16', '\U0001F600'}
	for _, r := range runes {
		buf := make([]byte, utf8.UTFMax)
		n := utf8.EncodeRune(buf, r)
		fmt.Printf("  '%c' (U+%04X) -> %d bytes: %v\n", r, r, n, buf[:n])
	}

	// AppendRune
	var encoded []byte
	for _, r := range runes {
		encoded = utf8.AppendRune(encoded, r)
	}
	fmt.Printf("\n  AppendRune result: %q (%d bytes)\n", string(encoded), len(encoded))

	// ============================================
	// 4. UNICODE/UTF8: COUNTING AND VALIDATION
	// ============================================
	fmt.Println("\n--- 4. unicode/utf8: Counting and Validation ---")
	fmt.Println(`
COUNTING:
  utf8.RuneCount(p)           -> rune count in []byte
  utf8.RuneCountInString(s)   -> rune count in string
  utf8.RuneLen(r)             -> bytes needed for rune

VALIDATION:
  utf8.Valid(p)               -> is valid UTF-8?
  utf8.ValidString(s)         -> is valid UTF-8 string?
  utf8.ValidRune(r)           -> is valid Unicode code point?
  utf8.FullRune(p)            -> does p begin with full rune?
  utf8.FullRuneInString(s)    -> same for string

CONSTANTS:
  utf8.RuneError = U+FFFD (replacement character)
  utf8.RuneSelf  = 0x80 (below this, byte is valid ASCII rune)
  utf8.UTFMax    = 4 (max bytes per rune)
  utf8.MaxRune   = U+10FFFF (max valid code point)`)

	testStrings := []string{
		"Hello",
		"caf\u00e9",
		"\u4f60\u597d",
		"\U0001F600\U0001F601\U0001F602",
	}
	for _, s := range testStrings {
		fmt.Printf("\n  %q:\n", s)
		fmt.Printf("    Bytes: %d, Runes: %d, Valid: %v\n",
			len(s), utf8.RuneCountInString(s), utf8.ValidString(s))
	}

	// RuneLen
	fmt.Println("\n  RuneLen:")
	for _, r := range []rune{'A', '\u00e9', '\u4e16', '\U0001F600'} {
		fmt.Printf("    '%c' (U+%04X): %d bytes\n", r, r, utf8.RuneLen(r))
	}

	// Valid / ValidRune
	fmt.Println("\n  Validation:")
	fmt.Printf("    Valid(\"hello\"): %v\n", utf8.Valid([]byte("hello")))
	fmt.Printf("    Valid([0xff, 0xfe]): %v\n", utf8.Valid([]byte{0xff, 0xfe}))
	fmt.Printf("    ValidRune('A'): %v\n", utf8.ValidRune('A'))
	fmt.Printf("    ValidRune(0x110000): %v\n", utf8.ValidRune(0x110000))
	fmt.Printf("    ValidRune(0xD800): %v (surrogate)\n", utf8.ValidRune(0xD800))

	// FullRune
	fmt.Println("\n  FullRune (partial UTF-8 detection):")
	partial := []byte{0xE4, 0xB8}
	fmt.Printf("    FullRune([0xE4, 0xB8]): %v (incomplete 3-byte seq)\n", utf8.FullRune(partial))
	full := []byte{0xE4, 0xB8, 0x96}
	fmt.Printf("    FullRune([0xE4, 0xB8, 0x96]): %v (complete: '%s')\n", utf8.FullRune(full), string(full))

	// ============================================
	// 5. UNICODE/UTF16 PACKAGE
	// ============================================
	fmt.Println("\n--- 5. unicode/utf16 Package ---")
	fmt.Println(`
UTF-16 encodes Unicode code points as sequences of 16-bit values.
Code points above U+FFFF use SURROGATE PAIRS (two uint16 values).

FUNCTIONS:
  utf16.Encode(s []rune)     -> []uint16  (runes to UTF-16)
  utf16.Decode(s []uint16)   -> []rune    (UTF-16 to runes)
  utf16.IsSurrogate(r rune)  -> bool      (is surrogate half?)
  utf16.EncodeRune(r rune)   -> (r1, r2 rune)  (to surrogate pair)
  utf16.DecodeRune(r1,r2)    -> rune      (from surrogate pair)

SURROGATES (U+D800 to U+DFFF):
  High surrogate: U+D800 to U+DBFF
  Low surrogate:  U+DC00 to U+DFFF
  Not valid Unicode code points by themselves!`)

	// Encode to UTF-16
	runeSlice := []rune{'H', 'e', 'l', 'l', 'o', ' ', '\u4e16', '\U0001F600'}
	encoded16 := utf16.Encode(runeSlice)
	fmt.Printf("\n  Runes: %v\n", runeSlice)
	fmt.Printf("  UTF-16: %v\n", encoded16)
	fmt.Printf("  Note: emoji U+1F600 becomes surrogate pair [0x%04X, 0x%04X]\n",
		encoded16[len(encoded16)-2], encoded16[len(encoded16)-1])

	// Decode from UTF-16
	decoded := utf16.Decode(encoded16)
	fmt.Printf("  Decoded back: %q\n", string(decoded))

	// Surrogate pairs
	fmt.Println("\n  Surrogate pairs:")
	emoji := '\U0001F600'
	r1, r2 := utf16.EncodeRune(emoji)
	fmt.Printf("    U+1F600 -> high=0x%04X, low=0x%04X\n", r1, r2)
	decoded1 := utf16.DecodeRune(r1, r2)
	fmt.Printf("    Decoded back: %c (U+%04X)\n", decoded1, decoded1)

	// IsSurrogate
	fmt.Println("\n  IsSurrogate:")
	fmt.Printf("    IsSurrogate(0xD800): %v (high surrogate)\n", utf16.IsSurrogate(0xD800))
	fmt.Printf("    IsSurrogate(0xDC00): %v (low surrogate)\n", utf16.IsSurrogate(0xDC00))
	fmt.Printf("    IsSurrogate('A'):    %v\n", utf16.IsSurrogate('A'))

	// ============================================
	// 6. UNICODE PACKAGE: CHARACTER CLASSIFICATION
	// ============================================
	fmt.Println("\n--- 6. unicode Package: Character Classification ---")
	os.Stdout.WriteString(`
The unicode package provides functions to classify runes.

CLASSIFICATION FUNCTIONS:
  unicode.IsLetter(r)    -> letter (any script)
  unicode.IsDigit(r)     -> decimal digit (any script)
  unicode.IsNumber(r)    -> number (includes fractions, roman numerals)
  unicode.IsSpace(r)     -> whitespace (space, tab, newline, etc)
  unicode.IsUpper(r)     -> uppercase letter
  unicode.IsLower(r)     -> lowercase letter
  unicode.IsTitle(r)     -> titlecase letter (e.g., Lj, Dz)
  unicode.IsPunct(r)     -> punctuation
  unicode.IsSymbol(r)    -> symbol (math, currency, etc)
  unicode.IsControl(r)   -> control character
  unicode.IsPrint(r)     -> printable (graphic + space)
  unicode.IsGraphic(r)   -> graphic (visible character)
  unicode.IsMark(r)      -> mark (combining character)

CASE CONVERSION:
  unicode.ToUpper(r)     -> uppercase rune
  unicode.ToLower(r)     -> lowercase rune
  unicode.ToTitle(r)     -> titlecase rune
  unicode.SimpleFold(r)  -> next case fold equivalent
`)

	testRunes := []rune{
		'A', 'z', '5', '\u00e9', '\u4e16', '\U0001F600',
		' ', '\t', '.', '$', '\u0300', '\n',
	}
	fmt.Println("  Character classification:")
	fmt.Printf("    %-6s %-8s %-7s %-6s %-6s %-6s %-6s %-6s %-7s\n",
		"Rune", "Letter", "Digit", "Space", "Upper", "Lower", "Punct", "Symb", "Print")
	for _, r := range testRunes {
		label := fmt.Sprintf("'%c'", r)
		if r == '\t' {
			label = "'\\t'"
		} else if r == '\n' {
			label = "'\\n'"
		}
		fmt.Printf("    %-6s %-8v %-7v %-6v %-6v %-6v %-6v %-6v %-7v\n",
			label,
			unicode.IsLetter(r),
			unicode.IsDigit(r),
			unicode.IsSpace(r),
			unicode.IsUpper(r),
			unicode.IsLower(r),
			unicode.IsPunct(r),
			unicode.IsSymbol(r),
			unicode.IsPrint(r),
		)
	}

	// Case conversion
	fmt.Println("\n  Case conversion:")
	fmt.Printf("    ToUpper('a') = '%c'\n", unicode.ToUpper('a'))
	fmt.Printf("    ToLower('Z') = '%c'\n", unicode.ToLower('Z'))
	fmt.Printf("    ToTitle('a') = '%c'\n", unicode.ToTitle('a'))

	// SimpleFold
	fmt.Println("\n  SimpleFold (case equivalents):")
	r0 := 'K'
	fmt.Printf("    Starting from '%c': ", r0)
	fold := unicode.SimpleFold(r0)
	fmt.Printf("'%c'", fold)
	for fold != r0 {
		fold = unicode.SimpleFold(fold)
		fmt.Printf(" -> '%c'", fold)
	}
	fmt.Println()

	// ============================================
	// 7. UNICODE RANGE TABLES
	// ============================================
	fmt.Println("\n--- 7. Unicode Range Tables ---")
	fmt.Println(`
The unicode package defines RangeTable for script/category checking:

SCRIPT TABLES:
  unicode.Latin, unicode.Greek, unicode.Cyrillic, unicode.Han,
  unicode.Arabic, unicode.Hiragana, unicode.Katakana, etc.

CATEGORY TABLES:
  unicode.Lu (uppercase letter), unicode.Ll (lowercase letter),
  unicode.Nd (decimal digit), unicode.Zs (space separator), etc.

CHECKING:
  unicode.Is(rangeTable, r) -> bool
  unicode.In(r, tables...)  -> bool (any table matches)`)

	scriptTests := []struct {
		r      rune
		label  string
	}{
		{'A', "Latin"},
		{'\u03B1', "Greek alpha"},
		{'\u0414', "Cyrillic De"},
		{'\u4e16', "Han (Chinese)"},
		{'\u3042', "Hiragana a"},
		{'\u0627', "Arabic Alef"},
	}

	fmt.Println("\n  Script detection:")
	for _, t := range scriptTests {
		scripts := []string{}
		if unicode.Is(unicode.Latin, t.r) {
			scripts = append(scripts, "Latin")
		}
		if unicode.Is(unicode.Greek, t.r) {
			scripts = append(scripts, "Greek")
		}
		if unicode.Is(unicode.Cyrillic, t.r) {
			scripts = append(scripts, "Cyrillic")
		}
		if unicode.Is(unicode.Han, t.r) {
			scripts = append(scripts, "Han")
		}
		if unicode.Is(unicode.Hiragana, t.r) {
			scripts = append(scripts, "Hiragana")
		}
		if unicode.Is(unicode.Arabic, t.r) {
			scripts = append(scripts, "Arabic")
		}
		fmt.Printf("    '%c' (%s): %s\n", t.r, t.label, strings.Join(scripts, ", "))
	}

	// unicode.In for multiple tables
	fmt.Println("\n  unicode.In (CJK detection):")
	cjkRunes := []rune{'\u4e16', '\u3042', '\u30A2', 'A'}
	for _, r := range cjkRunes {
		isCJK := unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana)
		fmt.Printf("    '%c' (U+%04X) is CJK: %v\n", r, r, isCJK)
	}

	// ============================================
	// 8. BOM HANDLING
	// ============================================
	fmt.Println("\n--- 8. BOM (Byte Order Mark) Handling ---")
	fmt.Println(`
BOM (Byte Order Mark) is U+FEFF at the start of a text stream.
In UTF-8, BOM is optional and often undesirable.

UTF-8 BOM:  [0xEF, 0xBB, 0xBF]
UTF-16 LE:  [0xFF, 0xFE]
UTF-16 BE:  [0xFE, 0xFF]

Go strings are UTF-8. BOM can cause issues with comparisons.`)

	// Detect and strip BOM
	withBOM := "\xef\xbb\xbfHello"
	noBOM := "Hello"
	fmt.Printf("\n  With BOM bytes: %v\n", []byte(withBOM)[:6])
	fmt.Printf("  Equal: %v (unexpected!)\n", withBOM == noBOM)

	stripped := stripBOM(withBOM)
	fmt.Printf("  After stripBOM: %q, equal: %v\n", stripped, stripped == noBOM)

	// ============================================
	// 9. PRACTICAL EXAMPLES
	// ============================================
	fmt.Println("\n--- 9. Practical Examples ---")

	// Validate identifier
	fmt.Println("\n  Validate Go identifier:")
	identifiers := []string{"myVar", "_private", "123abc", "hello-world", "cafeA", "\u4e16\u754c"}
	for _, id := range identifiers {
		valid := isValidIdentifier(id)
		fmt.Printf("    %q -> valid: %v\n", id, valid)
	}

	// Safe truncate (rune-aware)
	fmt.Println("\n  Safe truncate (rune-aware):")
	longStr := "Hello \u4e16\u754c! \U0001F600 Go is great"
	for _, maxRunes := range []int{5, 8, 12} {
		truncated := safeTruncate(longStr, maxRunes)
		fmt.Printf("    Truncate(%d runes): %q\n", maxRunes, truncated)
	}

	// Count characters by script
	fmt.Println("\n  Character analysis:")
	mixed := "Hello \u4f60\u597d 123 \u30a2 \U0001F600"
	analysis := analyzeString(mixed)
	fmt.Printf("    Input: %q\n", mixed)
	for category, count := range analysis {
		fmt.Printf("    %s: %d\n", category, count)
	}

	fmt.Println("\n=== End of Chapter 050 ===")
}

func stripBOM(s string) string {
	if strings.HasPrefix(s, "\xef\xbb\xbf") {
		return s[3:]
	}
	return s
}

func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

func safeTruncate(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := 0
	for i := range s {
		if runes >= maxRunes {
			return s[:i] + "..."
		}
		runes++
	}
	return s
}

func analyzeString(s string) map[string]int {
	counts := make(map[string]int)
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Latin, r):
			counts["Latin"]++
		case unicode.Is(unicode.Han, r):
			counts["Han (Chinese)"]++
		case unicode.Is(unicode.Katakana, r) || unicode.Is(unicode.Hiragana, r):
			counts["Japanese"]++
		case unicode.IsDigit(r):
			counts["Digit"]++
		case unicode.IsSpace(r):
			counts["Space"]++
		case unicode.IsSymbol(r):
			counts["Symbol"]++
		default:
			counts["Other"]++
		}
	}
	return counts
}
