// Package main - Chapter 096: golang.org/x/text
// x/text provides internationalization (i18n) and text processing:
// Unicode normalization, collation, language tags, encoding conversion,
// and locale-aware case mapping.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== GOLANG.ORG/X/TEXT ===")

	fmt.Println(`
  golang.org/x/text provides comprehensive text processing
  for internationalization (i18n):

  KEY PACKAGES:
    x/text/language    - BCP 47 language tags
    x/text/collation   - Locale-aware string sorting
    x/text/transform   - Streaming text transformations
    x/text/unicode/norm - Unicode normalization (NFC, NFD, etc.)
    x/text/encoding    - Character set conversion
    x/text/cases       - Locale-aware case mapping

  Install: go get golang.org/x/text`)

	// ============================================
	// LANGUAGE TAGS
	// ============================================
	fmt.Println("\n--- language: BCP 47 Language Tags ---")

	fmt.Println(`
  Language tags identify human languages (BCP 47 / RFC 5646):

  API:

    import "golang.org/x/text/language"

    // Parse a language tag
    tag, err := language.Parse("en-US")
    tag := language.Make("pt-BR")        // panics on error
    tag := language.MustParse("zh-Hans") // panics on error

    // Predefined tags
    language.English       // en
    language.Spanish       // es
    language.Portuguese    // pt
    language.Japanese      // ja
    language.Chinese       // zh

    // Components
    base, _ := tag.Base()     // language.English
    region, _ := tag.Region() // language.US
    script, _ := tag.Script() // language.Latin

  TAG FORMAT:
    language[-script][-region][-variant]

    en           - English
    en-US        - English (United States)
    en-GB        - English (United Kingdom)
    zh-Hans      - Chinese (Simplified)
    zh-Hant-TW   - Chinese (Traditional, Taiwan)
    pt-BR        - Portuguese (Brazil)
    sr-Latn      - Serbian (Latin script)

  LANGUAGE MATCHING:

    // Create a matcher with supported languages
    matcher := language.NewMatcher([]language.Tag{
        language.English,
        language.Spanish,
        language.Portuguese,
    })

    // Match user's preferred language
    tag, index, confidence := matcher.Match(
        language.Make("pt-BR"),    // user wants Brazilian Portuguese
    )
    // tag=pt, index=2, confidence=High
    // Matched Portuguese even though BR variant wasn't listed

  MATCHING CONFIDENCE:
    language.No        // no match
    language.Low       // weak match
    language.High      // good match (region/variant may differ)
    language.Exact     // exact match`)

	// ============================================
	// COLLATION - LOCALE-AWARE SORTING
	// ============================================
	fmt.Println("\n--- collation: Locale-Aware Sorting ---")

	// Demonstrate the problem with naive sorting
	words := []string{"banana", "apple", "Banana", "Apple", "cherry"}
	naive := make([]string, len(words))
	copy(naive, words)
	sort.Strings(naive)
	fmt.Println("  Naive sort (ASCII):", naive)

	// Case-insensitive sort with stdlib
	caseInsensitive := make([]string, len(words))
	copy(caseInsensitive, words)
	sort.Slice(caseInsensitive, func(i, j int) bool {
		return strings.ToLower(caseInsensitive[i]) < strings.ToLower(caseInsensitive[j])
	})
	fmt.Println("  Case-insensitive:  ", caseInsensitive)

	fmt.Println(`
  PROBLEM: sort.Strings uses byte comparison (ASCII order).
  This gives wrong results for many languages:

    ASCII:  "Z" < "a"  (uppercase before lowercase)
    German: "a" < "Z"  (case-insensitive by locale)
    Spanish: "ch" sorts after "c" (digraph)
    Swedish: "z" < "a-with-ring" (different alphabet order)

  SOLUTION: x/text/collation

    import "golang.org/x/text/collation"

    // Create a collator for a specific locale
    c := collation.New(language.Spanish)

    // Sort strings
    c.SortStrings(words)

    // Compare strings
    result := c.CompareString("a", "b")
    // -1 (a < b), 0 (equal), 1 (a > b)

  COLLATION OPTIONS:

    // Case-insensitive
    c := collation.New(language.English, collation.IgnoreCase)

    // Ignore diacritics (accents)
    c := collation.New(language.French, collation.IgnoreDiacritics)

    // Numeric sorting ("file2" before "file10")
    c := collation.New(language.English, collation.Numeric)

  EXAMPLE - Sorting names by locale:

    names := []string{"Oscar", "Andre", "Oscar", "Andres"}

    // English: Andre, Andres, Oscar, Oscar
    // Spanish: Andre, Andres, Oscar, Oscar (n-tilde after n)
    // The collator handles locale-specific ordering`)

	// ============================================
	// UNICODE NORMALIZATION
	// ============================================
	fmt.Println("\n--- unicode/norm: Unicode Normalization ---")

	// Demonstrate the problem
	e1 := "\u00e9"     // e with acute (single code point: LATIN SMALL LETTER E WITH ACUTE)
	e2 := "e\u0301"    // e + combining acute (two code points)
	fmt.Println("  Same visual character, different bytes:")
	os.Stdout.WriteString(fmt.Sprintf("    composed:   %q (len=%d, runes=%d)\n", e1, len(e1), utf8.RuneCountInString(e1)))
	os.Stdout.WriteString(fmt.Sprintf("    decomposed: %q (len=%d, runes=%d)\n", e2, len(e2), utf8.RuneCountInString(e2)))
	os.Stdout.WriteString(fmt.Sprintf("    equal (==): %t\n", e1 == e2))

	fmt.Println(`
  PROBLEM: The same character can have multiple Unicode
  representations. "e with accent" can be:
    U+00E9          (precomposed: single code point)
    U+0065 U+0301   (decomposed: 'e' + combining accent)

  These look identical but have different byte sequences!
  This breaks: string comparison, hashing, searching.

  NORMALIZATION FORMS:

    NFC  - Canonical Decomposition + Canonical Composition
           Preferred form. Composes characters where possible.
           "e" + accent -> single code point

    NFD  - Canonical Decomposition
           Decomposes all characters to base + combining.
           Single code point -> "e" + accent

    NFKC - Compatibility Decomposition + Canonical Composition
           Like NFC but also normalizes compatibility variants.
           Ligature "fi" -> "f" + "i"

    NFKD - Compatibility Decomposition
           Like NFD but also decomposes compatibility variants.

  API:

    import "golang.org/x/text/unicode/norm"

    // Normalize a string
    nfc := norm.NFC.String(input)
    nfd := norm.NFD.String(input)
    nfkc := norm.NFKC.String(input)

    // Check if already normalized
    if norm.NFC.IsNormalString(input) { ... }

    // Streaming normalization
    reader := norm.NFC.Reader(inputReader)
    writer := norm.NFC.Writer(outputWriter)

  WHEN TO NORMALIZE:

    - Database storage: normalize to NFC before storing
    - String comparison: normalize both sides first
    - Search/indexing: normalize at index time and query time
    - Hashing: normalize before hashing
    - File names: normalize for cross-platform compatibility

  RECOMMENDATION:
    Use NFC for storage and comparison (most compact).
    Use NFKC when you need compatibility normalization
    (e.g., search should match ligatures).`)

	// ============================================
	// ENCODING CONVERSION
	// ============================================
	fmt.Println("\n--- encoding: Character Set Conversion ---")

	fmt.Println(`
  Go uses UTF-8 internally. x/text/encoding converts
  between UTF-8 and other character encodings.

  SUPPORTED ENCODINGS:

    x/text/encoding/charmap    - ISO 8859-x, Windows-125x
    x/text/encoding/japanese   - Shift JIS, EUC-JP, ISO-2022-JP
    x/text/encoding/korean     - EUC-KR
    x/text/encoding/simplifiedchinese  - GBK, GB18030, HZ-GB2312
    x/text/encoding/traditionalchinese - Big5
    x/text/encoding/unicode    - UTF-16 (LE/BE), UTF-8 BOM

  CONVERTING TO UTF-8:

    import (
        "golang.org/x/text/encoding/japanese"
        "golang.org/x/text/transform"
        "io"
    )

    // Decode Shift JIS to UTF-8
    decoder := japanese.ShiftJIS.NewDecoder()
    utf8Reader := transform.NewReader(shiftJISReader, decoder)
    utf8Bytes, _ := io.ReadAll(utf8Reader)

    // Or decode a byte slice directly
    utf8Bytes, _, _ := transform.Bytes(decoder, shiftJISBytes)

  CONVERTING FROM UTF-8:

    encoder := japanese.ShiftJIS.NewEncoder()
    shiftJISReader := transform.NewReader(utf8Reader, encoder)

  DETECT ENCODING:

    // x/text doesn't include detection, but you can use:
    // - HTTP Content-Type header
    // - HTML <meta charset> tag
    // - BOM (Byte Order Mark) for UTF-16
    // - Third-party: github.com/saintfish/chardet

  REAL-WORLD EXAMPLE - Reading a CSV in Windows-1252:

    import "golang.org/x/text/encoding/charmap"

    file, _ := os.Open("data.csv")
    reader := transform.NewReader(file, charmap.Windows1252.NewDecoder())
    // reader now produces UTF-8 output
    csvReader := csv.NewReader(reader)`)

	// ============================================
	// TRANSFORM
	// ============================================
	fmt.Println("\n--- transform: Streaming Text Transformations ---")

	fmt.Println(`
  The transform package provides a framework for streaming
  text transformations that can be chained together.

  TRANSFORMER INTERFACE:

    type Transformer interface {
        Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error)
        Reset()
    }

  BUILT-IN TRANSFORMERS:

    // Chain multiple transformers
    t := transform.Chain(
        norm.NFD,                    // decompose
        runes.Remove(runes.In(unicode.Mn)),  // remove accents
        norm.NFC,                    // recompose
    )

    // Apply to a string
    result, _, _ := transform.String(t, "cafe avec creme")
    // Result: "cafe avec creme" (accents removed)

  USING WITH READERS/WRITERS:

    reader := transform.NewReader(input, transformer)
    writer := transform.NewWriter(output, transformer)

  RUNES PACKAGE:

    import "golang.org/x/text/runes"

    // Remove characters in a Unicode category
    removeAccents := runes.Remove(runes.In(unicode.Mn))

    // Map characters
    toASCII := runes.Map(func(r rune) rune {
        if r > 127 { return '?' }
        return r
    })

    // Replace characters
    replaceSpaces := runes.Map(func(r rune) rune {
        if unicode.IsSpace(r) { return '-' }
        return r
    })`)

	// ============================================
	// CASES - LOCALE-AWARE CASE MAPPING
	// ============================================
	fmt.Println("\n--- cases: Locale-Aware Case Mapping ---")

	// Demonstrate stdlib limitations
	fmt.Println("  stdlib strings.ToUpper limitations:")
	turkish := "istanbul"
	os.Stdout.WriteString(fmt.Sprintf("    strings.ToUpper(%q) = %q\n", turkish, strings.ToUpper(turkish)))
	fmt.Println("    (In Turkish, 'i' should become 'I-dot-above', not 'I')")

	fmt.Println(`
  PROBLEM: strings.ToUpper/ToLower doesn't handle locale rules.

  Turkish 'i':
    English: i -> I, I -> i
    Turkish: i -> I-with-dot, I -> dotless-i

  German sharp s:
    "strasse" titlecase = "Strasse"
    But "STRASSE" lowercase != "strasse" (could be "strasse")

  SOLUTION: x/text/cases

    import (
        "golang.org/x/text/cases"
        "golang.org/x/text/language"
    )

    // English
    en := cases.Upper(language.English)
    en.String("hello world")  // "HELLO WORLD"

    // Turkish
    tr := cases.Upper(language.Turkish)
    tr.String("istanbul")  // "ISTANBUL" (with dotted I)

    // Title case
    title := cases.Title(language.English)
    title.String("hello world")  // "Hello World"

    // Dutch title case
    nl := cases.Title(language.Dutch)
    nl.String("ijsselmeer")  // "IJsselmeer" (IJ is a digraph)

  CASE OPERATIONS:

    cases.Upper(tag)    // locale-aware uppercase
    cases.Lower(tag)    // locale-aware lowercase
    cases.Title(tag)    // locale-aware title case
    cases.Fold()        // case folding (for comparison)

  CASE FOLDING (for comparison):

    folder := cases.Fold()
    a := folder.String("Strasse")
    b := folder.String("STRASSE")
    // a == b  (case-insensitive comparison)`)

	// ============================================
	// WORKING EXAMPLES WITH STDLIB
	// ============================================
	fmt.Println("\n--- Working Examples (stdlib equivalents) ---")

	// Unicode category checking
	fmt.Println("\n  Unicode category checking:")
	samples := []rune{'A', 'a', '1', ' ', '\n', '\u00e9', '\u4e16'}
	for _, r := range samples {
		categories := []string{}
		if unicode.IsLetter(r) {
			categories = append(categories, "Letter")
		}
		if unicode.IsUpper(r) {
			categories = append(categories, "Upper")
		}
		if unicode.IsLower(r) {
			categories = append(categories, "Lower")
		}
		if unicode.IsDigit(r) {
			categories = append(categories, "Digit")
		}
		if unicode.IsSpace(r) {
			categories = append(categories, "Space")
		}
		os.Stdout.WriteString(fmt.Sprintf("    %U %q: %s\n", r, r, strings.Join(categories, ", ")))
	}

	// Simple accent removal (ASCII folding)
	fmt.Println("\n  Simple ASCII folding (stdlib):")
	accentMap := map[rune]rune{
		'\u00e0': 'a', '\u00e1': 'a', '\u00e2': 'a', '\u00e3': 'a', '\u00e4': 'a',
		'\u00e8': 'e', '\u00e9': 'e', '\u00ea': 'e', '\u00eb': 'e',
		'\u00f1': 'n', '\u00f6': 'o', '\u00fc': 'u',
	}
	input := "caf\u00e9 cr\u00e8me bru\u0308l\u00e9e"
	var folded strings.Builder
	for _, r := range input {
		if mapped, ok := accentMap[r]; ok {
			folded.WriteRune(mapped)
		} else if r < 128 {
			folded.WriteRune(r)
		} else {
			folded.WriteRune(r)
		}
	}
	os.Stdout.WriteString(fmt.Sprintf("    input:  %q\n", input))
	os.Stdout.WriteString(fmt.Sprintf("    folded: %q\n", folded.String()))
	fmt.Println("    (full accent removal requires x/text/transform + norm)")

	// ============================================
	// PRACTICAL PATTERNS
	// ============================================
	fmt.Println("\n--- Practical i18n Patterns ---")

	fmt.Println(`
  PATTERN 1: Normalize user input before storage

    func normalizeInput(s string) string {
        return norm.NFC.String(strings.TrimSpace(s))
    }

  PATTERN 2: Case-insensitive search across locales

    func containsI18n(text, query string) bool {
        fold := cases.Fold()
        return strings.Contains(
            fold.String(norm.NFC.String(text)),
            fold.String(norm.NFC.String(query)),
        )
    }

  PATTERN 3: Generate URL slugs

    func slugify(s string) string {
        t := transform.Chain(
            norm.NFD,
            runes.Remove(runes.In(unicode.Mn)),
            norm.NFC,
            cases.Lower(language.English),
        )
        slug, _, _ := transform.String(t, s)
        // Replace non-alphanumeric with hyphens
        // ... additional processing
        return slug
    }

  PATTERN 4: Content negotiation

    func negotiateLanguage(acceptHeader string) language.Tag {
        supported := []language.Tag{
            language.English,
            language.Spanish,
            language.Portuguese,
        }
        matcher := language.NewMatcher(supported)
        tags, _, _ := language.ParseAcceptLanguage(acceptHeader)
        tag, _, _ := matcher.Match(tags...)
        return tag
    }`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  golang.org/x/text ESSENTIALS:

  language:
    - BCP 47 language tags (en-US, pt-BR, zh-Hans)
    - Language matching for content negotiation
    - Base, region, script components

  collation:
    - Locale-aware string sorting
    - Case/diacritic insensitive options
    - Numeric sorting

  unicode/norm:
    - NFC (compose), NFD (decompose), NFKC, NFKD
    - Always normalize before comparing/hashing/storing
    - Streaming with Reader/Writer

  encoding:
    - Convert between UTF-8 and other encodings
    - Shift JIS, EUC-KR, GBK, Windows-1252, etc.

  cases:
    - Locale-aware Upper/Lower/Title case
    - Turkish, Dutch, German special rules
    - Case folding for comparison

  transform:
    - Chain transformers together
    - Streaming Reader/Writer adapters
    - Remove accents, map runes, normalize

  Install: go get golang.org/x/text`)
}
