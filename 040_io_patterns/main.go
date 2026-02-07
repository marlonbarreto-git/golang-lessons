// Package main - Chapter 040: I/O Patterns
// Go tiene un sistema de I/O basado en interfaces simples pero extremadamente poderosas.
// io.Reader e io.Writer son las interfaces mas importantes del lenguaje, y componer
// con ellas es la clave para escribir programas Go eficientes y elegantes.
package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

// ============================================
// CUSTOM READER: CountingReader
// ============================================

type CountingReader struct {
	reader    io.Reader
	bytesRead int64
}

func NewCountingReader(r io.Reader) *CountingReader {
	return &CountingReader{reader: r}
}

func (cr *CountingReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	cr.bytesRead += int64(n)
	return n, err
}

func (cr *CountingReader) BytesRead() int64 {
	return cr.bytesRead
}

// ============================================
// CUSTOM WRITER: UppercaseWriter
// ============================================

type UppercaseWriter struct {
	writer io.Writer
}

func NewUppercaseWriter(w io.Writer) *UppercaseWriter {
	return &UppercaseWriter{writer: w}
}

func (uw *UppercaseWriter) Write(p []byte) (int, error) {
	upper := bytes.Map(unicode.ToUpper, p)
	return uw.writer.Write(upper)
}

// ============================================
// CUSTOM READER: FibonacciReader
// ============================================

type FibonacciReader struct {
	a, b  int
	count int
	max   int
	buf   []byte
}

func NewFibonacciReader(n int) *FibonacciReader {
	return &FibonacciReader{a: 0, b: 1, max: n}
}

func (fr *FibonacciReader) Read(p []byte) (int, error) {
	if fr.count >= fr.max {
		return 0, io.EOF
	}

	if len(fr.buf) == 0 {
		if fr.count > 0 {
			fr.buf = []byte(fmt.Sprintf("%d\n", fr.a))
		} else {
			fr.buf = []byte(fmt.Sprintf("%d\n", fr.a))
		}
		fr.a, fr.b = fr.b, fr.a+fr.b
		fr.count++
	}

	n := copy(p, fr.buf)
	fr.buf = fr.buf[n:]
	return n, nil
}

// ============================================
// CUSTOM WRITER: LineCountWriter
// ============================================

type LineCountWriter struct {
	writer io.Writer
	lines  int
}

func NewLineCountWriter(w io.Writer) *LineCountWriter {
	return &LineCountWriter{writer: w}
}

func (lw *LineCountWriter) Write(p []byte) (int, error) {
	for _, b := range p {
		if b == '\n' {
			lw.lines++
		}
	}
	return lw.writer.Write(p)
}

func (lw *LineCountWriter) Lines() int {
	return lw.lines
}

// ============================================
// HELPER: JSON streaming demo types
// ============================================

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

func main() {
	fmt.Println("=== I/O PATTERNS EN GO ===")

	// ============================================
	// 1. io.Reader INTERFACE
	// ============================================
	fmt.Println("\n--- 1. io.Reader Interface ---")
	fmt.Println(`
io.Reader es LA interfaz mas importante de Go:

type Reader interface {
    Read(p []byte) (n int, err error)
}

REGLAS DEL CONTRATO:
1. Read lee hasta len(p) bytes en p
2. Retorna n (0 <= n <= len(p)) y posible error
3. Cuando no hay mas datos, retorna n=0, err=io.EOF
4. Puede retornar n>0 junto con io.EOF (datos finales)
5. No debe retornar n=0, err=nil (excepto len(p)==0)

QUIEN IMPLEMENTA io.Reader:
- os.File              (archivos)
- strings.Reader       (strings como reader)
- bytes.Reader         ([]byte como reader)
- bytes.Buffer         (buffer de bytes)
- bufio.Reader         (buffered reader)
- net.Conn             (conexiones de red)
- http.Request.Body    (body de HTTP)
- compress/gzip.Reader (descompresion)
- crypto/cipher.StreamReader
- io.LimitReader       (limitar bytes)
- io.SectionReader     (leer seccion)
- io.TeeReader         (bifurcar lectura)
- io.MultiReader       (concatenar readers)
- io.PipeReader        (pipe sincrono)`)

	// Demo: strings.Reader
	fmt.Println("Demo - strings.Reader:")
	sr := strings.NewReader("Hello, Go I/O!")
	buf := make([]byte, 5)
	for {
		n, err := sr.Read(buf)
		if n > 0 {
			fmt.Printf("  Read %d bytes: %q\n", n, buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	// Demo: bytes.Reader
	fmt.Println("\nDemo - bytes.Reader:")
	br := bytes.NewReader([]byte{0x48, 0x65, 0x6C, 0x6C, 0x6F})
	fmt.Printf("  Size: %d, Remaining: %d\n", br.Size(), br.Len())
	data, _ := io.ReadAll(br)
	fmt.Printf("  Content: %s\n", data)

	// ============================================
	// 2. io.Writer INTERFACE
	// ============================================
	fmt.Println("\n--- 2. io.Writer Interface ---")
	fmt.Println(`
type Writer interface {
    Write(p []byte) (n int, err error)
}

REGLAS DEL CONTRATO:
1. Write escribe len(p) bytes de p al flujo
2. Retorna n (0 <= n <= len(p)) y posible error
3. Si n < len(p), DEBE retornar un error non-nil
4. Write NO debe modificar el slice p
5. Write NO debe retener p despues de retornar

QUIEN IMPLEMENTA io.Writer:
- os.File               (archivos)
- os.Stdout/Stderr      (salida estandar)
- bytes.Buffer          (buffer de bytes)
- bufio.Writer          (buffered writer)
- strings.Builder       (construir strings)
- net.Conn              (conexiones de red)
- http.ResponseWriter   (respuestas HTTP)
- compress/gzip.Writer  (compresion)
- crypto/cipher.StreamWriter
- io.MultiWriter        (escribir a multiples)
- io.PipeWriter         (pipe sincrono)
- io.Discard            (descarta todo)`)

	// Demo: bytes.Buffer as Writer
	fmt.Println("Demo - bytes.Buffer como Writer:")
	var bbuf bytes.Buffer
	bbuf.WriteString("Hello ")
	bbuf.Write([]byte("World"))
	bbuf.WriteByte('!')
	fmt.Printf("  Buffer: %q (len=%d, cap=%d)\n", bbuf.String(), bbuf.Len(), bbuf.Cap())

	// Demo: strings.Builder
	fmt.Println("\nDemo - strings.Builder:")
	var sb strings.Builder
	sb.WriteString("Go ")
	sb.WriteString("I/O ")
	sb.WriteString("Patterns")
	fmt.Printf("  Result: %q (len=%d)\n", sb.String(), sb.Len())

	// Demo: io.Discard
	fmt.Println("\nDemo - io.Discard:")
	n, err := io.Copy(io.Discard, strings.NewReader("discarded data"))
	fmt.Printf("  Discarded %d bytes, err=%v\n", n, err)

	// ============================================
	// 3. INTERFACES COMPUESTAS
	// ============================================
	fmt.Println("\n--- 3. Interfaces Compuestas ---")
	fmt.Println(`
Go combina las interfaces basicas en interfaces compuestas:

type Closer interface {
    Close() error
}

type Seeker interface {
    Seek(offset int64, whence int) (int64, error)
    // whence: io.SeekStart(0), io.SeekCurrent(1), io.SeekEnd(2)
}

type ReadWriter interface {
    Reader
    Writer
}

type ReadCloser interface {
    Reader
    Closer
}

type WriteCloser interface {
    Writer
    Closer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

type ReadSeeker interface {
    Reader
    Seeker
}

type WriteSeeker interface {
    Writer
    Seeker
}

type ReadWriteSeeker interface {
    Reader
    Writer
    Seeker
}

// Interfaces adicionales:
type ReaderFrom interface {
    ReadFrom(r Reader) (n int64, err error)
}

type WriterTo interface {
    WriteTo(w Writer) (n int64, err error)
}

type ReaderAt interface {
    ReadAt(p []byte, off int64) (n int, err error)
}

type WriterAt interface {
    WriteAt(p []byte, off int64) (n int, err error)
}

type ByteReader interface {
    ReadByte() (byte, error)
}

type ByteWriter interface {
    WriteByte(c byte) error
}

type RuneReader interface {
    ReadRune() (r rune, size int, err error)
}

type StringWriter interface {
    WriteString(s string) (n int, err error)
}

EJEMPLO - Seeker:
r := strings.NewReader("Hello, World!")
r.Seek(7, io.SeekStart)  // Ir a posicion 7
// Leer desde ahí -> "World!"`)

	// Demo: Seeker
	fmt.Println("Demo - Seeker:")
	seeker := strings.NewReader("Hello, World!")
	seeker.Seek(7, io.SeekStart)
	rest, _ := io.ReadAll(seeker)
	fmt.Printf("  Desde posicion 7: %q\n", rest)

	seeker.Seek(-6, io.SeekEnd)
	rest, _ = io.ReadAll(seeker)
	fmt.Printf("  Ultimos 6 bytes: %q\n", rest)

	// ============================================
	// 4. io.ReadAll, io.Copy, io.CopyN, io.CopyBuffer
	// ============================================
	fmt.Println("\n--- 4. io.ReadAll, io.Copy, io.CopyN, io.CopyBuffer ---")
	fmt.Println(`
FUNCIONES DE UTILIDAD:

// Lee TODO el contenido de un Reader
func ReadAll(r Reader) ([]byte, error)
// CUIDADO: Carga todo en memoria. No usar con datos grandes.

// Copia de src a dst hasta EOF
func Copy(dst Writer, src Reader) (written int64, err error)
// Usa un buffer interno de 32KB

// Copia exactamente n bytes
func CopyN(dst Writer, src Reader, n int64) (written int64, err error)

// Copia con buffer personalizado
func CopyBuffer(dst Writer, src Reader, buf []byte) (written int64, err error)
// Permite controlar el tamano del buffer

// OPTIMIZACIONES INTERNAS DE io.Copy:
// Si src implementa WriterTo -> usa src.WriteTo(dst)
// Si dst implementa ReaderFrom -> usa dst.ReadFrom(src)
// Si no, usa buffer intermedio`)

	// Demo: io.ReadAll
	fmt.Println("Demo - io.ReadAll:")
	allData, _ := io.ReadAll(strings.NewReader("Complete content here"))
	fmt.Printf("  ReadAll: %q\n", allData)

	// Demo: io.Copy
	fmt.Println("\nDemo - io.Copy:")
	var copyBuf bytes.Buffer
	written, _ := io.Copy(&copyBuf, strings.NewReader("Copied data"))
	fmt.Printf("  Copied %d bytes: %q\n", written, copyBuf.String())

	// Demo: io.CopyN
	fmt.Println("\nDemo - io.CopyN:")
	var copyNBuf bytes.Buffer
	written, _ = io.CopyN(&copyNBuf, strings.NewReader("Only first 5 bytes"), 5)
	fmt.Printf("  CopyN %d bytes: %q\n", written, copyNBuf.String())

	// Demo: io.CopyBuffer
	fmt.Println("\nDemo - io.CopyBuffer:")
	var copyBufDst bytes.Buffer
	customBuf := make([]byte, 4)
	written, _ = io.CopyBuffer(&copyBufDst, strings.NewReader("Buffered copy"), customBuf)
	fmt.Printf("  CopyBuffer %d bytes: %q\n", written, copyBufDst.String())

	// ============================================
	// 5. io.TeeReader, io.LimitReader, io.SectionReader
	// ============================================
	fmt.Println("\n--- 5. TeeReader, LimitReader, SectionReader ---")
	fmt.Println(`
TeeReader - Bifurcar lectura (como el comando tee):
func TeeReader(r Reader, w Writer) Reader
// Todo lo que se lee de r tambien se escribe en w

LimitReader - Limitar bytes leidos:
func LimitReader(r Reader, n int64) Reader
// Retorna Reader que lee maximo n bytes

SectionReader - Leer seccion especifica:
func NewSectionReader(r ReaderAt, off int64, n int64) *SectionReader
// Lee n bytes desde offset off`)

	// Demo: TeeReader
	fmt.Println("Demo - TeeReader:")
	var teeLog bytes.Buffer
	teeReader := io.TeeReader(strings.NewReader("TeeReader data"), &teeLog)
	primary, _ := io.ReadAll(teeReader)
	fmt.Printf("  Primary read: %q\n", primary)
	fmt.Printf("  Tee copy:     %q\n", teeLog.String())

	// Demo: TeeReader para calcular hash mientras lees
	fmt.Println("\nDemo - TeeReader para hash:")
	originalData := "Calculate hash while reading"
	hasher := sha256.New()
	hashReader := io.TeeReader(strings.NewReader(originalData), hasher)
	content, _ := io.ReadAll(hashReader)
	fmt.Printf("  Content: %q\n", content)
	fmt.Printf("  SHA-256: %x\n", hasher.Sum(nil))

	// Demo: LimitReader
	fmt.Println("\nDemo - LimitReader:")
	limited := io.LimitReader(strings.NewReader("Only read first 10 bytes of this long string"), 10)
	limitedData, _ := io.ReadAll(limited)
	fmt.Printf("  Limited: %q\n", limitedData)

	// Demo: SectionReader
	fmt.Println("\nDemo - SectionReader:")
	fullReader := strings.NewReader("AAABBBCCCDDDEEE")
	section := io.NewSectionReader(fullReader, 3, 6)
	sectionData, _ := io.ReadAll(section)
	fmt.Printf("  Section [3:9]: %q\n", sectionData)
	fmt.Printf("  Section size: %d\n", section.Size())

	// ============================================
	// 6. io.MultiReader, io.MultiWriter
	// ============================================
	fmt.Println("\n--- 6. MultiReader, MultiWriter ---")
	fmt.Println(`
MultiReader - Concatenar readers secuencialmente:
func MultiReader(readers ...Reader) Reader
// Lee de cada reader en orden hasta EOF de todos

MultiWriter - Escribir a multiples writers simultaneamente:
func MultiWriter(writers ...Writer) Writer
// Cada Write() se envia a TODOS los writers`)

	// Demo: MultiReader
	fmt.Println("Demo - MultiReader:")
	header := strings.NewReader("[HEADER]")
	body := strings.NewReader("[BODY]")
	footer := strings.NewReader("[FOOTER]")
	combined := io.MultiReader(header, body, footer)
	allContent, _ := io.ReadAll(combined)
	fmt.Printf("  Combined: %q\n", allContent)

	// Demo: MultiWriter
	fmt.Println("\nDemo - MultiWriter:")
	var out1, out2, out3 bytes.Buffer
	multi := io.MultiWriter(&out1, &out2, &out3)
	multi.Write([]byte("Written to all three"))
	fmt.Printf("  out1: %q\n", out1.String())
	fmt.Printf("  out2: %q\n", out2.String())
	fmt.Printf("  out3: %q\n", out3.String())

	// Demo: MultiWriter para logging + file
	fmt.Println("\nDemo - MultiWriter (stdout + buffer):")
	var logBuf bytes.Buffer
	fmt.Print("  ")
	logWriter := io.MultiWriter(os.Stdout, &logBuf)
	fmt.Fprint(logWriter, "Log to stdout AND buffer\n")
	fmt.Printf("  Buffer captured: %q\n", logBuf.String())

	// ============================================
	// 7. io.Pipe
	// ============================================
	fmt.Println("\n--- 7. io.Pipe ---")
	fmt.Println(`
Pipe crea un par sincronizado de Reader/Writer:
func Pipe() (*PipeReader, *PipeWriter)

CARACTERISTICAS:
- Sincronico: Write bloquea hasta que Read consume los datos
- Sin buffer interno (a diferencia de channels)
- Thread-safe: puede usarse entre goroutines
- PipeWriter.Close() causa EOF en PipeReader
- PipeReader.Close() causa ErrClosedPipe en PipeWriter
- CloseWithError() permite pasar error personalizado

CASO DE USO:
- Conectar un Writer con un Reader sin buffer intermedio
- Streaming entre goroutines
- Adaptar APIs (cuando uno produce Writer y otro espera Reader)`)

	// Demo: io.Pipe
	fmt.Println("Demo - io.Pipe:")
	pr, pw := io.Pipe()

	var pipeWg sync.WaitGroup
	pipeWg.Add(1)

	go func() {
		defer pipeWg.Done()
		defer pw.Close()
		for i := 1; i <= 3; i++ {
			msg := fmt.Sprintf("Message %d\n", i)
			pw.Write([]byte(msg))
		}
	}()

	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		fmt.Printf("  Pipe received: %q\n", scanner.Text())
	}
	pipeWg.Wait()

	// Demo: Pipe con JSON encoding/decoding
	fmt.Println("\nDemo - Pipe con JSON:")
	pr2, pw2 := io.Pipe()
	var pipeWg2 sync.WaitGroup
	pipeWg2.Add(1)

	go func() {
		defer pipeWg2.Done()
		defer pw2.Close()
		enc := json.NewEncoder(pw2)
		enc.Encode(LogEntry{Level: "info", Message: "pipe started"})
		enc.Encode(LogEntry{Level: "warn", Message: "low memory", Code: 42})
	}()

	dec := json.NewDecoder(pr2)
	for dec.More() {
		var entry LogEntry
		if err := dec.Decode(&entry); err != nil {
			break
		}
		fmt.Printf("  JSON via pipe: level=%s msg=%q code=%d\n", entry.Level, entry.Message, entry.Code)
	}
	pr2.Close()
	pipeWg2.Wait()

	// ============================================
	// 8. io.NopCloser
	// ============================================
	fmt.Println("\n--- 8. io.NopCloser ---")
	fmt.Println(`
NopCloser envuelve un Reader y le agrega un Close() que no hace nada:
func NopCloser(r Reader) ReadCloser

CASO DE USO:
- Cuando necesitas un ReadCloser pero solo tienes un Reader
- Muy comun en tests y mocking de HTTP bodies
- http.Request.Body es io.ReadCloser

EJEMPLO:
body := io.NopCloser(strings.NewReader("request body"))
req.Body = body  // http.Request.Body necesita ReadCloser`)

	// Demo: NopCloser
	fmt.Println("Demo - io.NopCloser:")
	rc := io.NopCloser(strings.NewReader("ReadCloser data"))
	nopData, _ := io.ReadAll(rc)
	err = rc.Close() // no-op, siempre nil
	fmt.Printf("  Data: %q, Close error: %v\n", nopData, err)

	// ============================================
	// 9. CUSTOM READER/WRITER IMPLEMENTATIONS
	// ============================================
	fmt.Println("\n--- 9. Custom Reader/Writer ---")
	fmt.Println(`
CREAR TU PROPIO READER:

type MyReader struct {
    data []byte
    pos  int
}

func (r *MyReader) Read(p []byte) (int, error) {
    if r.pos >= len(r.data) {
        return 0, io.EOF
    }
    n := copy(p, r.data[r.pos:])
    r.pos += n
    return n, nil
}

CREAR TU PROPIO WRITER:

type MyWriter struct {
    buf []byte
}

func (w *MyWriter) Write(p []byte) (int, error) {
    w.buf = append(w.buf, p...)
    return len(p), nil
}`)

	// Demo: CountingReader
	fmt.Println("Demo - CountingReader (custom):")
	cr := NewCountingReader(strings.NewReader("Counting these bytes"))
	io.ReadAll(cr)
	fmt.Printf("  Total bytes read: %d\n", cr.BytesRead())

	// Demo: UppercaseWriter
	fmt.Println("\nDemo - UppercaseWriter (custom):")
	var upperBuf bytes.Buffer
	uw := NewUppercaseWriter(&upperBuf)
	uw.Write([]byte("hello world"))
	fmt.Printf("  Uppercased: %q\n", upperBuf.String())

	// Demo: FibonacciReader
	fmt.Println("\nDemo - FibonacciReader (custom):")
	fib := NewFibonacciReader(8)
	fibData, _ := io.ReadAll(fib)
	fmt.Printf("  Fibonacci: %s", fibData)

	// Demo: LineCountWriter
	fmt.Println("\nDemo - LineCountWriter (custom):")
	var lcBuf bytes.Buffer
	lc := NewLineCountWriter(&lcBuf)
	fmt.Fprint(lc, "Line 1\nLine 2\nLine 3\n")
	fmt.Printf("  Content: %q\n", lcBuf.String())
	fmt.Printf("  Lines counted: %d\n", lc.Lines())

	// ============================================
	// 10. bufio PACKAGE
	// ============================================
	fmt.Println("\n--- 10. bufio Package ---")
	fmt.Println(`
bufio proporciona I/O con buffer para mejorar rendimiento:

BUFIO.READER:
- Envuelve un io.Reader con buffer (default 4096 bytes)
- Reduce syscalls al leer en bloques
- Metodos adicionales: ReadString, ReadLine, ReadBytes, Peek

br := bufio.NewReader(file)      // buffer default 4096
br := bufio.NewReaderSize(file, 8192)  // buffer custom

BUFIO.WRITER:
- Envuelve un io.Writer con buffer
- Acumula escrituras y las envia en bloque
- IMPORTANTE: Siempre llamar Flush() al terminar

bw := bufio.NewWriter(file)
bw := bufio.NewWriterSize(file, 8192)
bw.Write(data)
bw.Flush()  // OBLIGATORIO

BUFIO.SCANNER:
- Lee input token por token (default: linea por linea)
- Mas simple que ReadLine para la mayoria de casos

scanner := bufio.NewScanner(reader)
for scanner.Scan() {
    line := scanner.Text()  // string sin newline
    // o scanner.Bytes() para []byte
}
if err := scanner.Err(); err != nil { ... }`)

	// Demo: bufio.Reader
	fmt.Println("Demo - bufio.Reader:")
	bufReader := bufio.NewReader(strings.NewReader("Line 1\nLine 2\nLine 3"))
	for {
		line, err := bufReader.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		if line != "" {
			fmt.Printf("  Read line: %q\n", line)
		}
		if err != nil {
			break
		}
	}

	// Demo: bufio.Reader.Peek
	fmt.Println("\nDemo - bufio.Reader.Peek:")
	peekReader := bufio.NewReader(strings.NewReader("Peek at this"))
	peeked, _ := peekReader.Peek(4)
	fmt.Printf("  Peeked: %q\n", peeked)
	full, _ := io.ReadAll(peekReader)
	fmt.Printf("  Full (still available): %q\n", full)

	// Demo: bufio.Writer
	fmt.Println("\nDemo - bufio.Writer:")
	var bwBuf bytes.Buffer
	bw := bufio.NewWriter(&bwBuf)
	bw.WriteString("Buffered ")
	bw.WriteString("output ")
	bw.WriteString("here")
	fmt.Printf("  Before Flush - buffer has %d bytes buffered, dest has %d\n", bw.Buffered(), bwBuf.Len())
	bw.Flush()
	fmt.Printf("  After Flush  - buffer has %d bytes buffered, dest has %q\n", bw.Buffered(), bwBuf.String())

	// Demo: bufio.Scanner
	fmt.Println("\nDemo - bufio.Scanner:")
	input := "apple\nbanana\ncherry\ndate"
	lineScanner := bufio.NewScanner(strings.NewReader(input))
	for lineScanner.Scan() {
		fmt.Printf("  Scanned: %q\n", lineScanner.Text())
	}

	// ============================================
	// 11. bufio.Scanner CUSTOM SPLIT FUNCTIONS
	// ============================================
	fmt.Println("\n--- 11. bufio.Scanner Custom Split ---")
	fmt.Println(`
SPLIT FUNCTIONS PREDEFINIDAS:
- bufio.ScanLines  (default) - divide por \n
- bufio.ScanWords  - divide por espacios
- bufio.ScanBytes  - byte por byte
- bufio.ScanRunes  - rune por rune (Unicode)

CUSTOM SPLIT FUNCTION:
type SplitFunc func(data []byte, atEOF bool) (advance int, token []byte, err error)

Retorna:
- advance: cuantos bytes avanzar en el input
- token: el token encontrado (nil si necesita mas datos)
- err: error (nil para continuar, ErrFinalToken para terminar)`)

	// Demo: ScanWords
	fmt.Println("Demo - bufio.ScanWords:")
	wordScanner := bufio.NewScanner(strings.NewReader("Go is awesome and fast"))
	wordScanner.Split(bufio.ScanWords)
	words := []string{}
	for wordScanner.Scan() {
		words = append(words, wordScanner.Text())
	}
	fmt.Printf("  Words: %v\n", words)

	// Demo: ScanRunes
	fmt.Println("\nDemo - bufio.ScanRunes:")
	runeScanner := bufio.NewScanner(strings.NewReader("Go!"))
	runeScanner.Split(bufio.ScanRunes)
	runes := []string{}
	for runeScanner.Scan() {
		runes = append(runes, runeScanner.Text())
	}
	fmt.Printf("  Runes: %v\n", runes)

	// Demo: Custom split by comma
	fmt.Println("\nDemo - Custom SplitFunc (split by comma):")
	csvLine := "apple,banana,,cherry,date"
	commaScanner := bufio.NewScanner(strings.NewReader(csvLine))
	commaScanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, ','); i >= 0 {
			return i + 1, data[:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	for commaScanner.Scan() {
		fmt.Printf("  Token: %q\n", commaScanner.Text())
	}

	// Demo: Scanner with larger buffer
	fmt.Println("\nDemo - Scanner con buffer grande:")
	fmt.Println(`
// Default max token size es 64KB. Para lineas mas grandes:
scanner := bufio.NewScanner(reader)
scanner.Buffer(make([]byte, 1024*1024), 1024*1024)  // 1MB max
for scanner.Scan() {
    // procesar lineas de hasta 1MB
}`)

	// ============================================
	// 12. bytes.Buffer, bytes.Reader, strings.Reader, strings.Builder
	// ============================================
	fmt.Println("\n--- 12. bytes.Buffer, bytes.Reader, strings.Reader, strings.Builder ---")
	fmt.Println(`
BYTES.BUFFER - Buffer de lectura/escritura:
- Implementa io.Reader, io.Writer, io.ByteReader, io.ByteWriter, io.RuneReader
- Crece automaticamente
- ReadFrom y WriteTo para eficiencia

var buf bytes.Buffer           // zero-value es usable
buf := bytes.NewBuffer(data)   // con datos iniciales
buf := bytes.NewBufferString("hello")

BYTES.READER - Reader sobre []byte:
- Implementa io.Reader, io.ReaderAt, io.Seeker, io.ByteReader, io.RuneReader
- Inmutable (no crece)
- Ideal para leer []byte como stream

r := bytes.NewReader(data)

STRINGS.READER - Reader sobre string:
- Igual que bytes.Reader pero para strings
- No copia la string (eficiente)

r := strings.NewReader("hello")

STRINGS.BUILDER - Construir strings eficientemente:
- Implementa io.Writer
- Mas eficiente que concatenar strings
- No se puede leer (solo escribir y obtener resultado)
- No se puede copiar despues de escribir

var sb strings.Builder
sb.WriteString("hello")
sb.WriteByte(' ')
sb.WriteRune('!')
result := sb.String()
sb.Reset()`)

	// Demo: bytes.Buffer full API
	fmt.Println("Demo - bytes.Buffer API completa:")
	var demoBuf bytes.Buffer
	demoBuf.WriteString("Hello")
	demoBuf.WriteByte(' ')
	demoBuf.Write([]byte("World"))
	demoBuf.WriteRune('!')
	fmt.Printf("  String: %q\n", demoBuf.String())
	fmt.Printf("  Bytes:  %v\n", demoBuf.Bytes())
	fmt.Printf("  Len: %d, Cap: %d\n", demoBuf.Len(), demoBuf.Cap())

	firstByte, _ := demoBuf.ReadByte()
	fmt.Printf("  ReadByte: %q, Remaining: %q\n", firstByte, demoBuf.String())

	readStr, _ := demoBuf.ReadString(' ')
	fmt.Printf("  ReadString(' '): %q, Remaining: %q\n", readStr, demoBuf.String())

	demoBuf.Reset()
	fmt.Printf("  After Reset: %q (len=%d)\n", demoBuf.String(), demoBuf.Len())

	// Demo: bytes.Buffer ReadFrom
	fmt.Println("\nDemo - bytes.Buffer.ReadFrom:")
	var rfBuf bytes.Buffer
	rfBuf.ReadFrom(strings.NewReader("Efficient transfer via ReadFrom"))
	fmt.Printf("  ReadFrom result: %q\n", rfBuf.String())

	// ============================================
	// 13. os.File as Reader/Writer
	// ============================================
	fmt.Println("\n--- 13. os.File como Reader/Writer ---")
	fmt.Println(`
os.File implementa:
- io.Reader, io.Writer
- io.ReaderAt, io.WriterAt
- io.Seeker, io.Closer
- io.ReadWriteCloser, io.ReadWriteSeeker

ABRIR ARCHIVOS:
f, err := os.Open("file.txt")       // Solo lectura (O_RDONLY)
f, err := os.Create("file.txt")     // Crear/truncar para escritura
f, err := os.OpenFile("file.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

FLAGS DE APERTURA:
os.O_RDONLY   // Solo lectura
os.O_WRONLY   // Solo escritura
os.O_RDWR     // Lectura y escritura
os.O_APPEND   // Append al final
os.O_CREATE   // Crear si no existe
os.O_EXCL     // Error si ya existe (con O_CREATE)
os.O_TRUNC    // Truncar al abrir
os.O_SYNC     // I/O sincronico

PERMISOS (mode):
0644  // Owner: rw, Group: r, Others: r  (archivos)
0755  // Owner: rwx, Group: rx, Others: rx (directorios, ejecutables)
0600  // Owner: rw (archivos sensibles)
0700  // Owner: rwx (directorios privados)

PATRON BASICO:
f, err := os.Open("data.txt")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

// Leer todo
data, err := io.ReadAll(f)

// O copiar a stdout
io.Copy(os.Stdout, f)

// O usar scanner
scanner := bufio.NewScanner(f)
for scanner.Scan() {
    fmt.Println(scanner.Text())
}`)

	// Demo: Write and read a temp file
	fmt.Println("Demo - os.File write/read:")
	tmpFile, err := os.CreateTemp("", "go-io-demo-*.txt")
	if err != nil {
		log.Fatal(err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	tmpFile.WriteString("Line 1\n")
	tmpFile.WriteString("Line 2\n")
	tmpFile.WriteString("Line 3\n")
	tmpFile.Close()

	readFile, _ := os.Open(tmpName)
	fileContent, _ := io.ReadAll(readFile)
	readFile.Close()
	fmt.Printf("  File content: %q\n", fileContent)

	// Demo: File info
	fmt.Println("\nDemo - File info:")
	info, _ := os.Stat(tmpName)
	fmt.Printf("  Name: %s\n", info.Name())
	fmt.Printf("  Size: %d bytes\n", info.Size())
	fmt.Printf("  Mode: %s\n", info.Mode())
	fmt.Printf("  IsDir: %v\n", info.IsDir())

	// Demo: os.ReadFile / os.WriteFile (convenience)
	fmt.Println("\nDemo - os.ReadFile / os.WriteFile:")
	fmt.Println(`
// Leer archivo completo (Go 1.16+)
data, err := os.ReadFile("file.txt")

// Escribir archivo completo (Go 1.16+)
err := os.WriteFile("file.txt", []byte("content"), 0644)

// NOTA: Estos son convenientes pero NO permiten control granular
// Para archivos grandes, usar Open/Create con io.Copy`)

	// ============================================
	// 14. FILE OPERATIONS
	// ============================================
	fmt.Println("\n--- 14. File Operations ---")
	fmt.Println(`
CREAR:
os.Create("file.txt")           // Crear/truncar, permisos 0666
os.OpenFile("f.txt", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)  // Solo si no existe

RENOMBRAR/MOVER:
os.Rename("old.txt", "new.txt")
os.Rename("file.txt", "/other/path/file.txt")  // Mover

ELIMINAR:
os.Remove("file.txt")          // Eliminar archivo o directorio vacio
os.RemoveAll("directory/")     // Eliminar recursivamente

COPIAR ARCHIVO (no hay funcion built-in):
func CopyFile(src, dst string) error {
    source, err := os.Open(src)
    if err != nil {
        return err
    }
    defer source.Close()

    destination, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destination.Close()

    _, err = io.Copy(destination, source)
    return err
}

PERMISOS:
os.Chmod("file.txt", 0644)
os.Chown("file.txt", uid, gid)

SYMLINKS:
os.Symlink("target", "link")
target, err := os.Readlink("link")

TRUNCAR:
os.Truncate("file.txt", 100)  // Truncar a 100 bytes
f.Truncate(0)                 // Vaciar archivo abierto

FILE INFO:
info, err := os.Stat("file.txt")    // Sigue symlinks
info, err := os.Lstat("file.txt")   // No sigue symlinks
info.Name()      // nombre base
info.Size()      // tamano en bytes
info.Mode()      // permisos
info.ModTime()   // ultima modificacion
info.IsDir()     // es directorio?

EXISTS:
_, err := os.Stat("file.txt")
if os.IsNotExist(err) {
    // no existe
}
if errors.Is(err, fs.ErrNotExist) {  // Go 1.16+ preferred
    // no existe
}`)

	// ============================================
	// 15. DIRECTORY OPERATIONS
	// ============================================
	fmt.Println("\n--- 15. Directory Operations ---")
	fmt.Println(`
CREAR DIRECTORIOS:
os.Mkdir("dir", 0755)              // Un solo nivel
os.MkdirAll("a/b/c/d", 0755)      // Todos los niveles (como mkdir -p)

LEER DIRECTORIO:
entries, err := os.ReadDir(".")     // Go 1.16+ (reemplaza ioutil.ReadDir)
for _, entry := range entries {
    fmt.Println(entry.Name(), entry.IsDir(), entry.Type())
    info, _ := entry.Info()  // Lazy: solo si necesitas
}

WALK DIRECTORY (recursivo):
// Go 1.16+ filepath.WalkDir (mas eficiente que filepath.Walk)
err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
    if err != nil {
        return err  // o return nil para continuar
    }
    if d.IsDir() && d.Name() == ".git" {
        return filepath.SkipDir  // Saltar directorio
    }
    fmt.Println(path)
    return nil
})

WORKING DIRECTORY:
dir, err := os.Getwd()    // Obtener
err := os.Chdir("/path")  // Cambiar`)

	// Demo: Directory operations
	fmt.Println("Demo - Directory operations:")
	tmpDir, _ := os.MkdirTemp("", "go-io-dir-*")
	defer os.RemoveAll(tmpDir)

	os.MkdirAll(filepath.Join(tmpDir, "a", "b", "c"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("one"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("two"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "a", "nested.txt"), []byte("nested"), 0644)

	entries, _ := os.ReadDir(tmpDir)
	fmt.Printf("  Directory %s:\n", filepath.Base(tmpDir))
	for _, e := range entries {
		fmt.Printf("    %s (dir=%v)\n", e.Name(), e.IsDir())
	}

	// Demo: WalkDir
	fmt.Println("\nDemo - filepath.WalkDir:")
	filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(tmpDir, path)
		if rel == "." {
			return nil
		}
		prefix := ""
		if d.IsDir() {
			prefix = "[DIR] "
		}
		fmt.Printf("    %s%s\n", prefix, rel)
		return nil
	})

	// ============================================
	// 16. TEMP FILES AND DIRECTORIES
	// ============================================
	fmt.Println("\n--- 16. Temp Files y Directories ---")
	fmt.Println(`
ARCHIVOS TEMPORALES:
f, err := os.CreateTemp("", "prefix-*.txt")    // Dir vacio = os.TempDir()
f, err := os.CreateTemp("/tmp", "myapp-*.log")
defer os.Remove(f.Name())
defer f.Close()

DIRECTORIOS TEMPORALES:
dir, err := os.MkdirTemp("", "myapp-*")
defer os.RemoveAll(dir)

PATRON:
* en el patron se reemplaza con string aleatorio
"prefix-*"       -> "prefix-123456789"
"*.txt"          -> "123456789.txt"
"prefix-*-suffix" -> "prefix-123456789-suffix"

TEMP DIR DEL SISTEMA:
os.TempDir()  // "/tmp" en Linux/Mac, "TEMP" en Windows`)

	// Demo: Temp files
	fmt.Println("Demo - Temp file:")
	tmp, _ := os.CreateTemp("", "demo-*.txt")
	tmp.WriteString("temporary content")
	fmt.Printf("  Temp file: %s\n", tmp.Name())
	tmp.Close()
	os.Remove(tmp.Name())

	// ============================================
	// 17. FILE LOCKING PATTERNS
	// ============================================
	fmt.Println("\n--- 17. File Locking Patterns ---")
	fmt.Println(`
Go no tiene file locking en la stdlib (excepto syscall).
Patrones comunes:

PATRON 1 - Lock file:
func AcquireLock(path string) (*os.File, error) {
    f, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
    if err != nil {
        return nil, fmt.Errorf("lock already held: ...", err)
    }
    fmt.Fprintf(f, pid, os.Getpid())
    return f, nil
}

func ReleaseLock(f *os.File) error {
    name := f.Name()
    f.Close()
    return os.Remove(name)
}

PATRON 2 - Escritura atomica (rename):
func AtomicWrite(filename string, data []byte) error {
    tmp, err := os.CreateTemp(filepath.Dir(filename), ".tmp-*")
    if err != nil {
        return err
    }
    defer func() {
        tmp.Close()
        os.Remove(tmp.Name())  // Cleanup si rename falla
    }()

    if _, err := tmp.Write(data); err != nil {
        return err
    }
    if err := tmp.Sync(); err != nil {  // Fsync para durabilidad
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }
    return os.Rename(tmp.Name(), filename)  // Atomico en mismo FS
}

PATRON 3 - sync.Mutex para acceso concurrente en mismo proceso:
type SafeFile struct {
    mu   sync.Mutex
    path string
}

func (sf *SafeFile) Write(data []byte) error {
    sf.mu.Lock()
    defer sf.mu.Unlock()

    f, err := os.OpenFile(sf.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = f.Write(data)
    return err
}

PATRON 4 - syscall.Flock (Unix):
import "syscall"

f, _ := os.Open("file.txt")
syscall.Flock(int(f.Fd()), syscall.LOCK_EX)   // Exclusive lock
defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN) // Unlock`)

	// ============================================
	// 18. io/fs PACKAGE
	// ============================================
	fmt.Println("\n--- 18. io/fs Package (Go 1.16+) ---")
	fmt.Println(`
io/fs define interfaces abstractas para filesystems:

INTERFACES PRINCIPALES:

type FS interface {
    Open(name string) (File, error)
}

type File interface {
    Stat() (FileInfo, error)
    Read([]byte) (int, error)
    Close() error
}

type ReadDirFS interface {
    FS
    ReadDir(name string) ([]DirEntry, error)
}

type ReadFileFS interface {
    FS
    ReadFile(name string) ([]byte, error)
}

type StatFS interface {
    FS
    Stat(name string) (FileInfo, error)
}

type SubFS interface {
    FS
    Sub(dir string) (FS, error)
}

type GlobFS interface {
    FS
    Glob(pattern string) ([]string, error)
}

FUNCIONES HELPER:
fs.ReadFile(fsys, "file.txt")
fs.ReadDir(fsys, ".")
fs.Stat(fsys, "file.txt")
fs.Sub(fsys, "subdir")
fs.Glob(fsys, "*.go")
fs.WalkDir(fsys, ".", walkFn)

IMPLEMENTACIONES:
- os.DirFS("/path")       // Filesystem real restringido a path
- embed.FS               // Archivos embebidos
- zip.Reader             // Archivos ZIP como FS
- testing/fstest.MapFS   // FS en memoria para tests

USO DE os.DirFS:
fsys := os.DirFS("/etc")
data, err := fs.ReadFile(fsys, "hostname")

TESTING CON fstest.MapFS:
import "testing/fstest"

testFS := fstest.MapFS{
    "hello.txt":       {Data: []byte("hello")},
    "dir/world.txt":   {Data: []byte("world")},
}

data, _ := fs.ReadFile(testFS, "hello.txt")
// data = []byte("hello")

// Validar que tu FS cumple el contrato:
err := fstest.TestFS(myFS, "expected-file.txt")`)

	// Demo: os.DirFS
	fmt.Println("Demo - os.DirFS:")
	tmpDirFS, _ := os.MkdirTemp("", "go-fs-demo-*")
	defer os.RemoveAll(tmpDirFS)
	os.WriteFile(filepath.Join(tmpDirFS, "hello.txt"), []byte("Hello from fs!"), 0644)
	os.WriteFile(filepath.Join(tmpDirFS, "world.txt"), []byte("World!"), 0644)
	os.MkdirAll(filepath.Join(tmpDirFS, "sub"), 0755)
	os.WriteFile(filepath.Join(tmpDirFS, "sub", "nested.txt"), []byte("Nested!"), 0644)

	fsys := os.DirFS(tmpDirFS)

	fsData, _ := fs.ReadFile(fsys, "hello.txt")
	fmt.Printf("  fs.ReadFile: %q\n", fsData)

	fmt.Println("  fs.WalkDir:")
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		kind := "FILE"
		if d.IsDir() {
			kind = "DIR "
		}
		fmt.Printf("    [%s] %s\n", kind, path)
		return nil
	})

	// Demo: Sub filesystem
	fmt.Println("\n  fs.Sub:")
	subFS, _ := fs.Sub(fsys, "sub")
	subData, _ := fs.ReadFile(subFS, "nested.txt")
	fmt.Printf("    Sub read: %q\n", subData)

	// ============================================
	// 19. STREAMING PATTERNS
	// ============================================
	fmt.Println("\n--- 19. Streaming Patterns ---")
	fmt.Println(`
PROCESAR ARCHIVOS GRANDES SIN CARGAR EN MEMORIA:

PATRON 1 - Linea por linea:
func ProcessLargeFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    scanner.Buffer(make([]byte, 1024*1024), 1024*1024)  // 1MB buffer
    for scanner.Scan() {
        processLine(scanner.Text())
    }
    return scanner.Err()
}

PATRON 2 - Chunks de tamano fijo:
func ProcessInChunks(r io.Reader, chunkSize int, process func([]byte) error) error {
    buf := make([]byte, chunkSize)
    for {
        n, err := io.ReadFull(r, buf)
        if n > 0 {
            if err := process(buf[:n]); err != nil {
                return err
            }
        }
        if err == io.EOF || err == io.ErrUnexpectedEOF {
            return nil
        }
        if err != nil {
            return err
        }
    }
}

PATRON 3 - io.Copy para transferir:
func StreamFile(w http.ResponseWriter, path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(w, f)  // Nunca carga todo en memoria
    return err
}

PATRON 4 - Pipeline de transformacion:
func CompressAndHash(src io.Reader, dst io.Writer) (hash []byte, err error) {
    h := sha256.New()
    tee := io.TeeReader(src, h)           // Bifurcar para hash
    gzw := gzip.NewWriter(dst)            // Comprimir al destino
    defer gzw.Close()
    if _, err := io.Copy(gzw, tee); err != nil {
        return nil, err
    }
    return h.Sum(nil), nil
}`)

	// Demo: Streaming pipeline
	fmt.Println("Demo - Streaming pipeline (compress + hash):")
	inputData := "This is data that will be compressed and hashed simultaneously using io.TeeReader and gzip.Writer to demonstrate streaming composition."
	var compressed bytes.Buffer
	h := sha256.New()
	tee := io.TeeReader(strings.NewReader(inputData), h)
	gzw := gzip.NewWriter(&compressed)
	io.Copy(gzw, tee)
	gzw.Close()
	fmt.Printf("  Original size:    %d bytes\n", len(inputData))
	fmt.Printf("  Compressed size:  %d bytes\n", compressed.Len())
	fmt.Printf("  SHA-256:          %x\n", h.Sum(nil))

	// Demo: Decompress to verify
	gzr, _ := gzip.NewReader(&compressed)
	decompressed, _ := io.ReadAll(gzr)
	gzr.Close()
	fmt.Printf("  Decompressed:     %q\n", decompressed)

	// Demo: Streaming chunk processing
	fmt.Println("\nDemo - Chunk processing:")
	chunkData := strings.NewReader("AABBCCDDEE")
	chunkBuf := make([]byte, 4)
	chunkNum := 0
	for {
		n, err := chunkData.Read(chunkBuf)
		if n > 0 {
			chunkNum++
			fmt.Printf("  Chunk %d: %q (%d bytes)\n", chunkNum, chunkBuf[:n], n)
		}
		if err == io.EOF {
			break
		}
	}

	// ============================================
	// 20. JSON STREAMING
	// ============================================
	fmt.Println("\n--- 20. JSON Streaming con Decoder/Encoder ---")
	fmt.Println(`
json.Decoder lee JSON de un io.Reader (streaming):
dec := json.NewDecoder(reader)
for dec.More() {
    var v MyType
    if err := dec.Decode(&v); err != nil {
        break
    }
    process(v)
}

json.Encoder escribe JSON a un io.Writer (streaming):
enc := json.NewEncoder(writer)
enc.SetIndent("", "  ")
enc.Encode(value)

VENTAJAS SOBRE json.Marshal/Unmarshal:
1. No carga todo en memoria
2. Puede procesar NDJSON (newline-delimited JSON)
3. Puede decodificar arrays JSON grandes elemento por elemento
4. Streaming a/desde archivos, conexiones, etc.

DECODIFICAR ARRAY GRANDE:
dec := json.NewDecoder(reader)
t, _ := dec.Token()  // Leer '['
for dec.More() {
    var item Item
    dec.Decode(&item)
    process(item)
}
t, _ = dec.Token()   // Leer ']'

NDJSON (un JSON por linea):
scanner := bufio.NewScanner(reader)
for scanner.Scan() {
    var entry Entry
    json.Unmarshal(scanner.Bytes(), &entry)
    process(entry)
}`)

	// Demo: JSON Encoder
	fmt.Println("Demo - json.Encoder:")
	var jsonBuf bytes.Buffer
	enc := json.NewEncoder(&jsonBuf)
	enc.Encode(LogEntry{Level: "info", Message: "server started"})
	enc.Encode(LogEntry{Level: "error", Message: "connection failed", Code: 500})
	enc.Encode(LogEntry{Level: "debug", Message: "request processed"})
	fmt.Printf("  Encoded NDJSON:\n%s", jsonBuf.String())

	// Demo: JSON Decoder
	fmt.Println("Demo - json.Decoder:")
	jdec := json.NewDecoder(strings.NewReader(jsonBuf.String()))
	for jdec.More() {
		var entry LogEntry
		if err := jdec.Decode(&entry); err != nil {
			break
		}
		fmt.Printf("  Decoded: [%s] %s (code=%d)\n", entry.Level, entry.Message, entry.Code)
	}

	// Demo: Decode JSON array streaming
	fmt.Println("\nDemo - Streaming JSON array:")
	jsonArray := `[{"level":"info","message":"one"},{"level":"warn","message":"two"},{"level":"error","message":"three"}]`
	arrayDec := json.NewDecoder(strings.NewReader(jsonArray))

	token, _ := arrayDec.Token()
	fmt.Printf("  Opening token: %v\n", token)
	for arrayDec.More() {
		var entry LogEntry
		arrayDec.Decode(&entry)
		fmt.Printf("  Array element: [%s] %s\n", entry.Level, entry.Message)
	}
	token, _ = arrayDec.Token()
	fmt.Printf("  Closing token: %v\n", token)

	// ============================================
	// 21. CSV READING/WRITING
	// ============================================
	fmt.Println("\n--- 21. CSV Reading/Writing ---")
	fmt.Println(`
csv.Reader lee CSV de un io.Reader:
r := csv.NewReader(reader)
r.Comma = ','          // Delimitador (default ',')
r.Comment = '#'        // Caracter de comentario
r.FieldsPerRecord = -1 // -1 = variable, 0 = primer record define, N = exacto
r.LazyQuotes = true    // Ser permisivo con quotes
r.TrimLeadingSpace = true

// Leer todo
records, err := r.ReadAll()

// Leer linea por linea (streaming)
for {
    record, err := r.Read()
    if err == io.EOF { break }
    if err != nil { log.Fatal(err) }
    process(record)
}

csv.Writer escribe CSV a un io.Writer:
w := csv.NewWriter(writer)
w.Comma = ';'  // Delimitador custom
w.Write([]string{"Name", "Age", "City"})
w.Write([]string{"Alice", "30", "NYC"})
w.Flush()
err := w.Error()`)

	// Demo: CSV Write
	fmt.Println("Demo - CSV Write:")
	var csvBuf bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuf)
	csvWriter.Write([]string{"Name", "Language", "Level"})
	csvWriter.Write([]string{"Alice", "Go", "Expert"})
	csvWriter.Write([]string{"Bob", "Rust", "Intermediate"})
	csvWriter.Write([]string{"Charlie", "Python", "Beginner"})
	csvWriter.Flush()
	fmt.Printf("  CSV output:\n%s", csvBuf.String())

	// Demo: CSV Read
	fmt.Println("Demo - CSV Read (streaming):")
	csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
	csvHeader, _ := csvReader.Read()
	fmt.Printf("  Header: %v\n", csvHeader)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  Record: %v\n", record)
	}

	// Demo: CSV with different delimiter
	fmt.Println("\nDemo - CSV con delimitador custom (tab):")
	var tsvBuf bytes.Buffer
	tsvWriter := csv.NewWriter(&tsvBuf)
	tsvWriter.Comma = '\t'
	tsvWriter.Write([]string{"ID", "Name", "Score"})
	tsvWriter.Write([]string{"1", "Alice", "95.5"})
	tsvWriter.Write([]string{"2", "Bob", "87.3"})
	tsvWriter.Flush()
	fmt.Printf("  TSV output:\n%s", tsvBuf.String())

	// ============================================
	// 22. COMPOSITION PATTERNS
	// ============================================
	fmt.Println("\n--- 22. Composition Patterns ---")
	fmt.Println(`
La composicion de Readers/Writers es el patron mas poderoso de Go I/O.
Encadenar transformaciones como un pipeline Unix.

PATRON 1 - Reader chain:
// Leer archivo -> descomprimir -> decodificar
f, _ := os.Open("data.json.gz")
defer f.Close()
gzr, _ := gzip.NewReader(f)
defer gzr.Close()
dec := json.NewDecoder(gzr)

PATRON 2 - Writer chain:
// Codificar -> comprimir -> escribir a archivo
f, _ := os.Create("data.json.gz")
defer f.Close()
gzw := gzip.NewWriter(f)
defer gzw.Close()
enc := json.NewEncoder(gzw)

PATRON 3 - Transform pipeline:
// Reader -> CountBytes -> Decompress -> Decode
counting := NewCountingReader(file)
gzr, _ := gzip.NewReader(counting)
dec := json.NewDecoder(gzr)
// ... despues de procesar:
// counting.BytesRead() tiene los bytes raw leidos

PATRON 4 - Fan-out writer:
// Escribir a archivo + calcular hash + contar bytes
f, _ := os.Create("output.dat")
h := sha256.New()
counter := NewCountingWriter(f)
multi := io.MultiWriter(counter, h)

PATRON 5 - Buffered I/O wrapping:
// Siempre buffer para archivos
f, _ := os.Open("large.txt")
br := bufio.NewReaderSize(f, 64*1024)  // 64KB buffer
scanner := bufio.NewScanner(br)

PATRON 6 - Adaptar con NopCloser:
// Cuando necesitas ReadCloser pero tienes Reader
body := io.NopCloser(strings.NewReader("data"))

PATRON 7 - Pipe para conectar Writer con Reader:
pr, pw := io.Pipe()
go func() {
    enc := json.NewEncoder(pw)
    enc.Encode(data)
    pw.Close()
}()
dec := json.NewDecoder(pr)
dec.Decode(&result)`)

	// Demo: Full composition pipeline
	fmt.Println("Demo - Full pipeline (encode -> compress -> hash -> count):")
	type Record struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	records := []Record{
		{1, "Alpha"}, {2, "Beta"}, {3, "Gamma"},
		{4, "Delta"}, {5, "Epsilon"},
	}

	var pipelineBuf bytes.Buffer
	pipelineHash := sha256.New()
	pipelineCounter := NewLineCountWriter(io.MultiWriter(&pipelineBuf, pipelineHash))
	gzWriter := gzip.NewWriter(pipelineCounter)
	jsonEnc := json.NewEncoder(gzWriter)

	for _, r := range records {
		jsonEnc.Encode(r)
	}
	gzWriter.Close()

	fmt.Printf("  Records encoded:   %d\n", len(records))
	fmt.Printf("  Compressed size:   %d bytes\n", pipelineBuf.Len())
	fmt.Printf("  SHA-256 of gzip:   %x\n", pipelineHash.Sum(nil))

	// Decompress and read back
	gzReader, _ := gzip.NewReader(&pipelineBuf)
	jsonDec := json.NewDecoder(gzReader)
	fmt.Println("  Decoded back:")
	for jsonDec.More() {
		var r Record
		jsonDec.Decode(&r)
		fmt.Printf("    %+v\n", r)
	}
	gzReader.Close()

	// Demo: Reader composition
	fmt.Println("\nDemo - Reader composition (limit + counting):")
	longText := strings.NewReader("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	limitedR := io.LimitReader(longText, 10)
	countingR := NewCountingReader(limitedR)
	result, _ := io.ReadAll(countingR)
	fmt.Printf("  Read: %q (%d bytes via counting)\n", result, countingR.BytesRead())

	fmt.Println(`
=== RESUMEN DE I/O PATTERNS ===

INTERFACES FUNDAMENTALES:
  io.Reader    - Read(p []byte) (n int, err error)
  io.Writer    - Write(p []byte) (n int, err error)
  io.Closer    - Close() error
  io.Seeker    - Seek(offset int64, whence int) (int64, error)

COMBINACIONES:
  io.ReadCloser, io.WriteCloser, io.ReadWriteCloser
  io.ReadSeeker, io.WriteSeeker, io.ReadWriteSeeker
  io.ReaderAt, io.WriterAt, io.ReaderFrom, io.WriterTo

UTILIDADES io:
  io.ReadAll          - Leer todo (cuidado con memoria)
  io.Copy/CopyN       - Copiar entre Reader y Writer
  io.TeeReader        - Bifurcar lectura
  io.LimitReader       - Limitar bytes
  io.SectionReader     - Leer seccion
  io.MultiReader       - Concatenar readers
  io.MultiWriter       - Fan-out a multiples writers
  io.Pipe             - Conectar Writer con Reader
  io.NopCloser        - Agregar Close() no-op

IMPLEMENTACIONES CLAVE:
  bytes.Buffer         - Buffer de lectura/escritura
  bytes.Reader         - Reader sobre []byte
  strings.Reader       - Reader sobre string
  strings.Builder      - Construir strings eficientemente
  os.File             - Archivos del filesystem

BUFIO:
  bufio.Reader         - Buffered reader (Peek, ReadString, ReadLine)
  bufio.Writer         - Buffered writer (SIEMPRE Flush!)
  bufio.Scanner        - Leer tokens (Lines, Words, custom split)

FILESYSTEM:
  os.Open/Create/OpenFile  - Abrir archivos
  os.ReadFile/WriteFile    - Convenience functions
  os.ReadDir               - Leer directorio
  filepath.WalkDir         - Recorrer recursivo
  os.DirFS + io/fs         - Filesystem abstracto

STREAMING:
  json.Encoder/Decoder  - JSON streaming
  csv.Reader/Writer     - CSV streaming
  gzip.Reader/Writer    - Compresion/descompresion

PRINCIPIOS:
  1. Acepta interfaces, retorna structs concretos
  2. io.Reader e io.Writer son las interfaces mas compuestas de Go
  3. Usa composicion para crear pipelines de transformacion
  4. Nunca olvides defer Close() y bufio.Writer.Flush()
  5. Para archivos grandes, SIEMPRE streaming (nunca ReadAll)
  6. io.Copy es tu mejor amigo para transferencias eficientes`)
}

/*
SUMMARY - Chapter 040: I/O Patterns

FUNDAMENTAL INTERFACES:
- io.Reader: Read(p []byte) (n int, err error)
- io.Writer: Write(p []byte) (n int, err error)
- io.Closer, io.Seeker, io.ReaderAt, io.WriterAt
- Combinations: ReadCloser, WriteCloser, ReadWriteCloser, ReadSeeker

IO UTILITIES:
- io.ReadAll: read everything (careful with memory)
- io.Copy/CopyN: efficient transfer between Reader and Writer
- io.TeeReader: bifurcate reading to a Writer
- io.LimitReader: limit bytes read
- io.MultiReader: concatenate multiple readers
- io.MultiWriter: fan-out to multiple writers
- io.Pipe: connect a Writer to a Reader
- io.NopCloser: add no-op Close to a Reader

KEY IMPLEMENTATIONS:
- bytes.Buffer: read/write buffer
- bytes.Reader: read-only over []byte
- strings.Reader: read-only over string
- strings.Builder: efficient string building
- os.File: filesystem files

BUFIO:
- bufio.Reader: buffered reader (Peek, ReadString, ReadLine)
- bufio.Writer: buffered writer (always Flush!)
- bufio.Scanner: token-based reading (lines, words, custom split)

FILESYSTEM:
- os.Open/Create/OpenFile for file access
- os.ReadFile/WriteFile for convenience
- os.ReadDir for directory listing
- filepath.WalkDir for recursive traversal
- os.DirFS + io/fs for filesystem abstraction

STREAMING:
- json.Encoder/Decoder for JSON streaming
- csv.Reader/Writer for CSV streaming
- gzip.Reader/Writer for compression/decompression
- Custom readers for composition and transformation

PRINCIPLES:
- Accept interfaces, return concrete types
- Use composition to build transformation pipelines
- Always defer Close() and bufio.Writer.Flush()
- For large files, always use streaming (never ReadAll)
- io.Copy is the most efficient transfer mechanism
*/
