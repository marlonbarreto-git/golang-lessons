// Package main - Chapter 042: Bytes Package Deep Dive
// bytes.Buffer, bytes.Reader, comparison functions, manipulation functions,
// relationship to strings package, and zero-copy techniques.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

func main() {
	fmt.Println("=== BYTES PACKAGE DEEP DIVE ===")

	// ============================================
	// 1. BYTES.BUFFER: THE SWISS ARMY KNIFE
	// ============================================
	fmt.Println("\n--- 1. bytes.Buffer: Read/Write Buffer ---")
	fmt.Println(`
bytes.Buffer is a variable-size buffer of bytes with Read and Write methods.
It implements: io.Reader, io.Writer, io.ByteScanner, io.RuneScanner,
               io.ReaderFrom, io.WriterTo, fmt.Stringer

CREATING:
  var buf bytes.Buffer                 // zero value is ready to use
  buf := bytes.NewBuffer([]byte{...})  // from existing slice
  buf := bytes.NewBufferString("...")  // from string

WRITING:
  buf.Write(p []byte)         -> write bytes
  buf.WriteString(s)          -> write string
  buf.WriteByte(c)            -> write single byte
  buf.WriteRune(r)            -> write single rune (UTF-8)
  buf.ReadFrom(r io.Reader)   -> read all from reader

READING:
  buf.Read(p []byte)          -> read into slice
  buf.ReadByte()              -> read one byte
  buf.ReadRune()              -> read one rune
  buf.ReadString(delim)       -> read until delim
  buf.ReadBytes(delim)        -> same but returns []byte
  buf.WriteTo(w io.Writer)    -> write all to writer

OTHER:
  buf.Bytes()     -> underlying byte slice (no copy!)
  buf.String()    -> string representation
  buf.Len()       -> unread bytes
  buf.Cap()       -> capacity
  buf.Grow(n)     -> pre-allocate n bytes
  buf.Reset()     -> clear buffer
  buf.Truncate(n) -> keep only first n unread bytes
  buf.Next(n)     -> return next n bytes as slice`)

	var buf bytes.Buffer
	buf.WriteString("Hello")
	buf.WriteByte(' ')
	buf.WriteString("World")
	buf.WriteRune('!')
	fmt.Printf("\n  Buffer: %q (len=%d, cap=%d)\n", buf.String(), buf.Len(), buf.Cap())

	// Read from buffer
	p := make([]byte, 5)
	n, _ := buf.Read(p)
	fmt.Printf("  Read %d bytes: %q, remaining: %q\n", n, string(p[:n]), buf.String())

	// ReadString
	buf.Reset()
	buf.WriteString("line1\nline2\nline3")
	line, _ := buf.ReadString('\n')
	fmt.Printf("  ReadString('\\n'): %q, remaining: %q\n", line, buf.String())

	// ============================================
	// 2. BYTES.BUFFER AS IO.WRITER
	// ============================================
	fmt.Println("\n--- 2. bytes.Buffer as io.Writer ---")
	fmt.Println(`
Since Buffer implements io.Writer, it can capture output from
any function that writes to an io.Writer.`)

	var capture bytes.Buffer
	fmt.Fprintf(&capture, "Name: %s, Age: %d", "Gopher", 10)
	fmt.Printf("\n  Captured fmt output: %q\n", capture.String())

	// JSON encoding to buffer
	capture.Reset()
	data := map[string]any{"name": "Go", "version": 1.22}
	json.NewEncoder(&capture).Encode(data)
	fmt.Printf("  JSON to buffer: %s", capture.String())

	// ============================================
	// 3. BYTES.BUFFER AS IO.READER
	// ============================================
	fmt.Println("\n--- 3. bytes.Buffer as io.Reader ---")
	fmt.Println(`
Buffer also implements io.Reader, useful for passing byte content
to functions expecting a reader.`)

	input := bytes.NewBufferString(`{"status":"ok","code":200}`)
	var result map[string]any
	json.NewDecoder(input).Decode(&result)
	fmt.Printf("\n  Decoded JSON from buffer: %v\n", result)

	// ReadFrom: copy from reader to buffer
	var dst bytes.Buffer
	src := strings.NewReader("content from reader")
	n64, _ := dst.ReadFrom(src)
	fmt.Printf("  ReadFrom: copied %d bytes, content=%q\n", n64, dst.String())

	// ============================================
	// 4. BYTES.NEWBUFFER AND NEWBUFFERSTRING
	// ============================================
	fmt.Println("\n--- 4. bytes.NewBuffer and NewBufferString ---")
	fmt.Println(`
  bytes.NewBuffer(buf []byte)    -> buffer initialized with buf
  bytes.NewBufferString(s)       -> buffer initialized with string

The provided data becomes the initial READABLE content.
For write-only usage, prefer var buf bytes.Buffer (zero value).`)

	fromBytes := bytes.NewBuffer([]byte{72, 101, 108, 108, 111})
	fmt.Printf("\n  NewBuffer from bytes: %q\n", fromBytes.String())

	fromString := bytes.NewBufferString("Pre-loaded content")
	fmt.Printf("  NewBufferString: %q (len=%d)\n", fromString.String(), fromString.Len())

	// Using NewBuffer for parsing
	rawData := bytes.NewBuffer([]byte("first\nsecond\nthird"))
	for {
		line, err := rawData.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		if line != "" {
			fmt.Printf("  Line: %q\n", line)
		}
		if err != nil {
			break
		}
	}

	// ============================================
	// 5. BYTES.READER
	// ============================================
	fmt.Println("\n--- 5. bytes.Reader ---")
	fmt.Println(`
bytes.Reader implements io.Reader, io.ReaderAt, io.Seeker for a []byte.
Unlike Buffer, Reader is read-only and supports seeking.

  bytes.NewReader(b []byte) *Reader
  r.Read(p)         -> io.Reader
  r.ReadAt(p, off)  -> io.ReaderAt (concurrent safe)
  r.Seek(offset, whence) -> io.Seeker
  r.Len()           -> unread bytes
  r.Size()          -> original length
  r.Reset(b)        -> reuse with new data`)

	reader := bytes.NewReader([]byte("ABCDEFGHIJ"))
	fmt.Printf("\n  Size: %d, Len: %d\n", reader.Size(), reader.Len())

	chunk := make([]byte, 3)
	reader.Read(chunk)
	fmt.Printf("  Read 3: %q, Len: %d\n", string(chunk), reader.Len())

	// Seek to beginning
	reader.Seek(0, io.SeekStart)
	fmt.Printf("  After Seek(0, Start), Len: %d\n", reader.Len())

	// ReadAt (random access, doesn't affect position)
	at := make([]byte, 3)
	reader.ReadAt(at, 5)
	fmt.Printf("  ReadAt(offset=5): %q\n", string(at))
	fmt.Printf("  Len after ReadAt (unchanged): %d\n", reader.Len())

	// Reset with new data
	reader.Reset([]byte("NEW DATA"))
	all, _ := io.ReadAll(reader)
	fmt.Printf("  After Reset: %q\n", string(all))

	// ============================================
	// 6. BYTES COMPARISON FUNCTIONS
	// ============================================
	fmt.Println("\n--- 6. Bytes Comparison Functions ---")
	fmt.Println(`
  bytes.Compare(a, b)   -> -1, 0, 1 (lexicographic)
  bytes.Equal(a, b)     -> bool (same content)
  bytes.EqualFold(a, b) -> bool (case-insensitive UTF-8)`)

	a := []byte("apple")
	b := []byte("banana")
	c := []byte("apple")
	d := []byte("APPLE")

	fmt.Printf("\n  Compare(\"apple\", \"banana\") = %d\n", bytes.Compare(a, b))
	fmt.Printf("  Compare(\"banana\", \"apple\") = %d\n", bytes.Compare(b, a))
	fmt.Printf("  Compare(\"apple\", \"apple\")  = %d\n", bytes.Compare(a, c))
	fmt.Printf("  Equal(\"apple\", \"apple\")    = %v\n", bytes.Equal(a, c))
	fmt.Printf("  Equal(\"apple\", \"APPLE\")    = %v\n", bytes.Equal(a, d))
	fmt.Printf("  EqualFold(\"apple\", \"APPLE\") = %v\n", bytes.EqualFold(a, d))

	// ============================================
	// 7. BYTES SEARCH FUNCTIONS
	// ============================================
	fmt.Println("\n--- 7. Bytes Search Functions ---")
	fmt.Println(`
  bytes.Contains(b, sub)        -> bool
  bytes.ContainsAny(b, chars)   -> bool (any byte from chars)
  bytes.ContainsRune(b, r)      -> bool
  bytes.Count(b, sep)           -> int (non-overlapping)
  bytes.Index(b, sep)           -> int (first, -1 if not found)
  bytes.LastIndex(b, sep)       -> int (last)
  bytes.HasPrefix(b, prefix)    -> bool
  bytes.HasSuffix(b, suffix)    -> bool`)

	text := []byte("Hello, Go World! Go is great. Go go go!")
	fmt.Printf("\n  Text: %q\n", string(text))
	fmt.Printf("  Contains(\"Go\"):      %v\n", bytes.Contains(text, []byte("Go")))
	fmt.Printf("  Contains(\"Python\"):  %v\n", bytes.Contains(text, []byte("Python")))
	fmt.Printf("  Count(\"Go\"):         %d\n", bytes.Count(text, []byte("Go")))
	fmt.Printf("  Count(\"go\"):         %d\n", bytes.Count(text, []byte("go")))
	fmt.Printf("  Index(\"Go\"):         %d\n", bytes.Index(text, []byte("Go")))
	fmt.Printf("  LastIndex(\"Go\"):     %d\n", bytes.LastIndex(text, []byte("Go")))
	fmt.Printf("  HasPrefix(\"Hello\"): %v\n", bytes.HasPrefix(text, []byte("Hello")))
	fmt.Printf("  HasSuffix(\"go!\"):   %v\n", bytes.HasSuffix(text, []byte("go!")))

	// ============================================
	// 8. BYTES MANIPULATION FUNCTIONS
	// ============================================
	fmt.Println("\n--- 8. Bytes Manipulation Functions ---")
	os.Stdout.WriteString(`
  bytes.Join(slices, sep)     -> concatenate with separator
  bytes.Split(b, sep)         -> split into slices
  bytes.SplitN(b, sep, n)     -> split into at most n slices
  bytes.SplitAfter(b, sep)    -> split keeping separator
  bytes.Replace(b, old, new, n) -> replace (n=-1 for all)
  bytes.ReplaceAll(b, old, new) -> replace all occurrences
  bytes.Map(f, b)             -> apply function to each rune
  bytes.Repeat(b, count)      -> repeat n times
  bytes.ToUpper(b) / ToLower(b) / ToTitle(b)
  bytes.Title(b)              -> deprecated, use cases.Title
  bytes.TrimSpace(b)          -> trim whitespace
  bytes.Trim(b, cutset)       -> trim specific chars
  bytes.TrimLeft / TrimRight / TrimPrefix / TrimSuffix
`)

	// Join
	parts := [][]byte{[]byte("one"), []byte("two"), []byte("three")}
	joined := bytes.Join(parts, []byte(", "))
	fmt.Printf("  Join: %q\n", string(joined))

	// Split
	csvData := []byte("alice,bob,charlie,dave")
	fields := bytes.Split(csvData, []byte(","))
	fmt.Printf("  Split: ")
	for i, f := range fields {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%q", string(f))
	}
	fmt.Println()

	// SplitN
	limited := bytes.SplitN(csvData, []byte(","), 2)
	fmt.Printf("  SplitN(n=2): %q, %q\n", string(limited[0]), string(limited[1]))

	// Replace
	original := []byte("foo bar foo baz foo")
	replaced := bytes.Replace(original, []byte("foo"), []byte("qux"), 2)
	fmt.Printf("  Replace(\"foo\"->\"qux\", n=2): %q\n", string(replaced))

	replacedAll := bytes.ReplaceAll(original, []byte("foo"), []byte("X"))
	fmt.Printf("  ReplaceAll(\"foo\"->\"X\"): %q\n", string(replacedAll))

	// Map
	mapped := bytes.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		return unicode.ToUpper(r)
	}, []byte("hello world"))
	fmt.Printf("  Map (upper+underscores): %q\n", string(mapped))

	// Repeat
	fmt.Printf("  Repeat(\"Go!\", 3): %q\n", string(bytes.Repeat([]byte("Go! "), 3)))

	// Trim functions
	padded := []byte("  \t  Hello, World!  \n  ")
	fmt.Printf("  TrimSpace: %q\n", string(bytes.TrimSpace(padded)))
	fmt.Printf("  TrimPrefix: %q\n", string(bytes.TrimPrefix([]byte("Hello World"), []byte("Hello "))))
	fmt.Printf("  TrimSuffix: %q\n", string(bytes.TrimSuffix([]byte("file.go"), []byte(".go"))))

	// ============================================
	// 9. BYTES VS STRINGS: MIRROR API
	// ============================================
	fmt.Println("\n--- 9. Bytes vs Strings: Mirror API ---")
	fmt.Println(`
The bytes package mirrors the strings package almost exactly.
The key difference: bytes works with []byte, strings works with string.

WHEN TO USE EACH:

  strings package:
    - Working with text that stays as string
    - Text processing, formatting
    - When the result will be a string

  bytes package:
    - Working with binary data
    - I/O operations (network, files)
    - When you need mutability
    - When avoiding string<->[]byte conversions
    - Performance-critical paths

EXAMPLE EQUIVALENTS:
  strings.Contains(s, sub)   <->  bytes.Contains(b, sub)
  strings.Split(s, sep)      <->  bytes.Split(b, sep)
  strings.Replace(s,o,n,c)   <->  bytes.Replace(b,o,n,c)
  strings.TrimSpace(s)       <->  bytes.TrimSpace(b)
  strings.NewReader(s)       <->  bytes.NewReader(b)
  strings.Builder            <->  bytes.Buffer`)

	// Same operation, different types
	strResult := strings.ToUpper("hello")
	bytResult := bytes.ToUpper([]byte("hello"))
	fmt.Printf("\n  strings.ToUpper: %q\n", strResult)
	fmt.Printf("  bytes.ToUpper:   %q\n", string(bytResult))

	// ============================================
	// 10. ZERO-COPY TECHNIQUES
	// ============================================
	fmt.Println("\n--- 10. Zero-Copy Techniques ---")
	fmt.Println(`
ZERO-COPY means avoiding unnecessary memory copies.

KEY TECHNIQUES:

1. Buffer.Bytes() returns the underlying slice (no copy):
   data := buf.Bytes()  // shares memory with buffer

2. bytes.NewReader wraps existing []byte (no copy):
   r := bytes.NewReader(data)  // data is NOT copied

3. Use Buffer.WriteTo instead of String() + Write:
   buf.WriteTo(w)  // writes directly, no intermediate string

4. Avoid string([]byte) conversions in hot paths:
   // BAD: s := string(data); strings.Contains(s, "x")
   // GOOD: bytes.Contains(data, []byte("x"))

CAUTION: Zero-copy means shared memory!
   data := buf.Bytes()
   buf.Reset()  // data is now invalid!`)

	// Demonstrate Bytes() sharing memory
	var zeroBuf bytes.Buffer
	zeroBuf.WriteString("original")
	shared := zeroBuf.Bytes()
	fmt.Printf("\n  Before Reset: buf=%q, shared=%q\n", zeroBuf.String(), string(shared))

	// Clone if you need to keep the data
	cloned := bytes.Clone(zeroBuf.Bytes())
	zeroBuf.Reset()
	zeroBuf.WriteString("modified")
	fmt.Printf("  After Reset: buf=%q, cloned=%q\n", zeroBuf.String(), string(cloned))

	// WriteTo for zero-copy output
	var outBuf bytes.Buffer
	outBuf.WriteString("  Direct write via WriteTo: ")
	outBuf.WriteTo(os.Stdout)
	fmt.Println("done!")

	// Avoid conversions in search
	rawBytes := []byte("The quick brown fox jumps over the lazy dog")
	target := []byte("fox")
	if idx := bytes.Index(rawBytes, target); idx >= 0 {
		fmt.Printf("  Found %q at index %d (no string conversion)\n", string(target), idx)
	}

	// ============================================
	// 11. BYTES.CLONE (GO 1.20+)
	// ============================================
	fmt.Println("\n--- 11. bytes.Clone (Go 1.20+) ---")
	fmt.Println(`
bytes.Clone returns a copy of b[:len(b)].
The result may have additional unused capacity.
Clone(nil) returns nil.

Before Go 1.20, you had to write:
  cp := make([]byte, len(b))
  copy(cp, b)`)

	orig := []byte("shared data")
	clone := bytes.Clone(orig)
	orig[0] = 'X'
	fmt.Printf("\n  Original after modify: %q\n", string(orig))
	fmt.Printf("  Clone (independent):   %q\n", string(clone))

	// Clone nil
	var nilSlice []byte
	clonedNil := bytes.Clone(nilSlice)
	fmt.Printf("  Clone(nil) == nil: %v\n", clonedNil == nil)

	// ============================================
	// 12. PRACTICAL EXAMPLES
	// ============================================
	fmt.Println("\n--- 12. Practical Examples ---")

	// Protocol message builder
	fmt.Println("\n  Protocol Message Builder:")
	msg := buildProtocolMessage("HELLO", map[string]string{
		"version": "1.0",
		"client":  "go-client",
	}, []byte("Hello, Server!"))
	fmt.Printf("  Message:\n%s\n", msg)

	// Byte stream tokenizer
	fmt.Println("  Stream Tokenizer (split on null bytes):")
	stream := []byte("token1\x00token2\x00token3\x00")
	tokens := bytes.Split(bytes.TrimRight(stream, "\x00"), []byte{0x00})
	for i, t := range tokens {
		fmt.Printf("    Token %d: %q\n", i, string(t))
	}

	// Efficient line counter
	content := []byte("line 1\nline 2\nline 3\nline 4\n")
	lineCount := bytes.Count(content, []byte("\n"))
	fmt.Printf("\n  Line count of %q: %d\n", string(content), lineCount)

	// Buffer pool pattern (conceptual)
	fmt.Println("\n  Buffer Reuse Pattern:")
	var reusable bytes.Buffer
	for i := 0; i < 3; i++ {
		reusable.Reset()
		fmt.Fprintf(&reusable, "iteration %d: data", i)
		fmt.Printf("    %s\n", reusable.String())
	}

	fmt.Println("\n=== End of Chapter 042 ===")
}

/*
SUMMARY - Chapter 042: Bytes Package Deep Dive

BYTES.BUFFER:
- Variable-size buffer implementing io.Reader and io.Writer
- Zero value is ready to use
- Write: Write, WriteString, WriteByte, WriteRune, ReadFrom
- Read: Read, ReadByte, ReadRune, ReadString, ReadBytes, WriteTo
- Info: Bytes(), String(), Len(), Cap(), Grow(n), Reset()
- Bytes() returns underlying slice (no copy - shared memory!)

BYTES.READER:
- Read-only reader over []byte with seeking support
- Implements io.Reader, io.ReaderAt, io.Seeker
- NewReader(b), Read, ReadAt, Seek, Len, Size, Reset

COMPARISON FUNCTIONS:
- Compare(a, b): lexicographic comparison (-1, 0, 1)
- Equal(a, b): content equality
- EqualFold(a, b): case-insensitive UTF-8 comparison

SEARCH FUNCTIONS:
- Contains, ContainsAny, ContainsRune
- Count, Index, LastIndex
- HasPrefix, HasSuffix

MANIPULATION FUNCTIONS:
- Join, Split, SplitN, SplitAfter
- Replace, ReplaceAll, Map, Repeat
- ToUpper, ToLower, TrimSpace, Trim, TrimPrefix, TrimSuffix

BYTES VS STRINGS:
- bytes mirrors strings API for []byte
- Use bytes for binary data, I/O, mutability, performance
- Use strings for text processing staying as string

ZERO-COPY TECHNIQUES:
- Buffer.Bytes() shares memory (no copy)
- bytes.NewReader wraps existing []byte
- bytes.Clone (Go 1.20+) for safe independent copy
- Avoid string([]byte) in hot paths
*/

func buildProtocolMessage(command string, headers map[string]string, body []byte) string {
	var buf bytes.Buffer
	buf.Grow(256)

	buf.WriteString(command)
	buf.WriteByte('\n')

	for k, v := range headers {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v)
		buf.WriteByte('\n')
	}

	buf.WriteByte('\n')
	buf.Write(body)
	buf.WriteByte('\n')

	return buf.String()
}
