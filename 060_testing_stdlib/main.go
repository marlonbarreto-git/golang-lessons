// Package main - Chapter 060: Testing Stdlib Utilities
// Utilidades de testing: testing/iotest para readers/writers,
// testing/fstest para filesystem virtual, property testing manual,
// y patrones de testdata y golden files.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing/fstest"
	"testing/iotest"
)

func main() {
	fmt.Println("=== TESTING STDLIB UTILITIES ===")

	// ============================================
	// 1. TESTING/IOTEST - READERS ESPECIALES
	// ============================================
	fmt.Println("\n--- 1. testing/iotest - Readers especiales ---")
	os.Stdout.WriteString(`
TESTING/IOTEST - READERS/WRITERS PARA TESTING:

  Readers que simulan comportamientos edge-case:

  iotest.HalfReader(r)                     // Lee solo la mitad de bytes pedidos
  iotest.OneByteReader(r)                  // Lee solo 1 byte a la vez
  iotest.DataErrReader(r)                  // Retorna error junto con ultimos datos
  iotest.ErrReader(err)                    // Siempre retorna error
  iotest.TimeoutReader(r)                  // Retorna timeout en segunda lectura

  Loggers para depuracion:
  iotest.NewReadLogger(prefix, r)          // Log cada lectura
  iotest.NewWriteLogger(prefix, r)         // Log cada escritura

  Validador:
  iotest.TestReader(r, content)            // Verificar que reader funciona bien
`)

	data := "Go es un lenguaje eficiente"

	fmt.Println("  HalfReader - lee la mitad de bytes pedidos:")
	halfR := iotest.HalfReader(strings.NewReader(data))
	buf := make([]byte, 10)
	for {
		n, err := halfR.Read(buf)
		if n > 0 {
			fmt.Printf("    Pedidos: %d, Leidos: %d, Datos: %q\n", len(buf), n, string(buf[:n]))
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("    Error: %v\n", err)
			break
		}
	}

	fmt.Println("\n  OneByteReader - lee 1 byte a la vez:")
	oneR := iotest.OneByteReader(strings.NewReader("Hello"))
	var collected []byte
	for {
		b := make([]byte, 100)
		n, err := oneR.Read(b)
		if n > 0 {
			collected = append(collected, b[:n]...)
		}
		if err == io.EOF {
			break
		}
	}
	fmt.Printf("    Resultado: %q (byte a byte)\n", string(collected))

	fmt.Println("\n  ErrReader - siempre retorna error:")
	customErr := errors.New("disco lleno")
	errR := iotest.ErrReader(customErr)
	_, err := errR.Read(make([]byte, 10))
	fmt.Printf("    Read -> error: %v\n", err)

	fmt.Println("\n  DataErrReader - error con ultimos datos:")
	dataErrR := iotest.DataErrReader(strings.NewReader("ABC"))
	for {
		b := make([]byte, 10)
		n, err := dataErrR.Read(b)
		fmt.Printf("    Read: n=%d data=%q err=%v\n", n, string(b[:n]), err)
		if err != nil {
			break
		}
	}

	// ============================================
	// 2. TESTING/IOTEST - LOGGERS
	// ============================================
	fmt.Println("\n--- 2. testing/iotest - Read/Write Loggers ---")

	fmt.Println("  ReadLogger (log a stderr, mostramos resultado):")
	var logBuf bytes.Buffer
	logR := iotest.NewReadLogger("READ", strings.NewReader("test data"))
	_ = logR
	readAll, _ := io.ReadAll(strings.NewReader("test data"))
	fmt.Printf("    Datos leidos: %q (%d bytes)\n", string(readAll), len(readAll))

	fmt.Println("\n  WriteLogger:")
	logW := iotest.NewWriteLogger("WRITE", &logBuf)
	logW.Write([]byte("hello"))
	logW.Write([]byte(" world"))
	fmt.Printf("    Buffer resultado: %q\n", logBuf.String())

	// ============================================
	// 3. TESTING/IOTEST - TESTEREADER
	// ============================================
	fmt.Println("\n--- 3. testing/iotest - TestReader ---")
	os.Stdout.WriteString(`
IOTEST.TESTREADER - VALIDAR IMPLEMENTACIONES DE IO.READER:

  iotest.TestReader(r io.Reader, content []byte) error

  Verifica que un reader:
  - Retorna los datos correctos
  - Maneja EOF correctamente
  - Funciona con buffers de distintos tamanos
  - No pierde datos en lecturas parciales
`)

	goodReader := strings.NewReader("hello world")
	err = iotest.TestReader(goodReader, []byte("hello world"))
	if err != nil {
		fmt.Printf("  strings.Reader: FALLO - %v\n", err)
	} else {
		fmt.Println("  strings.Reader: PASO todas las verificaciones")
	}

	err = iotest.TestReader(bytes.NewReader([]byte{1, 2, 3}), []byte{1, 2, 3})
	if err != nil {
		fmt.Printf("  bytes.Reader: FALLO - %v\n", err)
	} else {
		fmt.Println("  bytes.Reader: PASO todas las verificaciones")
	}

	customReader := &repeatReader{data: []byte("AB"), repeats: 3}
	err = iotest.TestReader(customReader, []byte("ABABAB"))
	if err != nil {
		fmt.Printf("  repeatReader: FALLO - %v\n", err)
	} else {
		fmt.Println("  repeatReader: PASO todas las verificaciones")
	}

	// ============================================
	// 4. TESTING/FSTEST - FILESYSTEM VIRTUAL
	// ============================================
	fmt.Println("\n--- 4. testing/fstest - MapFS ---")
	os.Stdout.WriteString(`
TESTING/FSTEST - FILESYSTEM EN MEMORIA:

  fstest.MapFS{}                           // Mapa de archivos virtuales
  fs["path"] = &fstest.MapFile{...}        // Agregar archivo
  fs.Open(name)                            // Abrir archivo
  fs.ReadFile(name)                        // Leer contenido
  fs.ReadDir(name)                         // Listar directorio
  fs.Glob(pattern)                         // Buscar por patron
  fs.Stat(name)                            // Info del archivo

  fstest.TestFS(fs, expectedFiles...)      // Validar implementacion de fs.FS

  MapFile campos:
    Data []byte, Mode fs.FileMode, ModTime time.Time, Sys interface{}
`)

	memFS := fstest.MapFS{
		"README.md":          {Data: []byte("# Mi Proyecto\n\nEjemplo de filesystem virtual.")},
		"src/main.go":        {Data: []byte("package main\n\nfunc main() {}\n")},
		"src/handler.go":     {Data: []byte("package main\n\nfunc handleRequest() {}\n")},
		"src/handler_test.go": {Data: []byte("package main\n\nfunc TestHandler(t *testing.T) {}\n")},
		"config/app.yaml":    {Data: []byte("port: 8080\nhost: localhost\n")},
		"config/db.yaml":     {Data: []byte("driver: postgres\nport: 5432\n")},
		"docs/api.md":        {Data: []byte("# API Documentation\n")},
	}

	fmt.Println("  Archivos en el filesystem virtual:")
	for name := range memFS {
		data, _ := memFS.ReadFile(name)
		fmt.Printf("    %-25s (%d bytes)\n", name, len(data))
	}

	fmt.Println("\n  Listar directorio 'src/':")
	entries, _ := memFS.ReadDir("src")
	for _, entry := range entries {
		info, _ := entry.Info()
		fmt.Printf("    %-20s size:%d dir:%v\n", entry.Name(), info.Size(), entry.IsDir())
	}

	fmt.Println("\n  Glob '*.md':")
	matches, _ := memFS.Glob("*.md")
	for _, m := range matches {
		fmt.Printf("    %s\n", m)
	}

	fmt.Println("\n  Glob 'src/*.go':")
	goFiles, _ := memFS.Glob("src/*.go")
	for _, m := range goFiles {
		fmt.Printf("    %s\n", m)
	}

	fmt.Println("\n  Leer config/app.yaml:")
	content, _ := memFS.ReadFile("config/app.yaml")
	fmt.Printf("    %s", string(content))

	fmt.Println("\n  Validar con TestFS:")
	err = fstest.TestFS(memFS,
		"README.md",
		"src/main.go",
		"config/app.yaml",
	)
	if err != nil {
		fmt.Printf("    FALLO: %v\n", err)
	} else {
		fmt.Println("    PASO: filesystem valido")
	}

	// ============================================
	// 5. PROPERTY TESTING MANUAL
	// ============================================
	fmt.Println("\n--- 5. Property-based testing (manual) ---")
	os.Stdout.WriteString(`
PROPERTY TESTING - VERIFICAR PROPIEDADES INVARIANTES:

  En lugar de test cases especificos, verificamos propiedades que
  SIEMPRE deben cumplirse para cualquier entrada.

  Ejemplos de propiedades:
  - sort(list) siempre esta ordenado
  - reverse(reverse(x)) == x
  - len(append(a, b)) == len(a) + len(b)
  - encode(decode(x)) == x

  testing/quick esta deprecated desde Go 1.22.
  Patron manual: generar entradas aleatorias y verificar propiedades.
`)

	fmt.Println("  Propiedad: reverse(reverse(s)) == s")
	testCases := []string{"", "a", "ab", "hello", "Go language", "racecar"}
	allPassed := true
	for _, tc := range testCases {
		result := reverseString(reverseString(tc))
		passed := result == tc
		if !passed {
			allPassed = false
		}
		fmt.Printf("    %q -> reverse(reverse) = %q  ok:%v\n", tc, result, passed)
	}
	fmt.Printf("  Resultado: %v\n", statusStr(allPassed))

	fmt.Println("\n  Propiedad: sort es idempotente (sort(sort(x)) == sort(x))")
	intSlices := [][]int{
		{5, 3, 1, 4, 2},
		{1},
		{},
		{3, 3, 3},
		{10, -5, 0, 7, -3},
	}
	allPassed = true
	for _, slice := range intSlices {
		sorted1 := sortCopy(slice)
		sorted2 := sortCopy(sorted1)
		equal := slicesEqual(sorted1, sorted2)
		if !equal {
			allPassed = false
		}
		fmt.Printf("    %v -> sorted: %v  idempotent:%v\n", slice, sorted1, equal)
	}
	fmt.Printf("  Resultado: %v\n", statusStr(allPassed))

	fmt.Println("\n  Propiedad: len(split(s, sep)) >= 1 para cualquier s y sep")
	allPassed = true
	splitCases := [][2]string{
		{"hello world", " "},
		{"a,b,c", ","},
		{"no-sep", "X"},
		{"", ","},
		{"aaa", "a"},
	}
	for _, tc := range splitCases {
		parts := strings.Split(tc[0], tc[1])
		passed := len(parts) >= 1
		if !passed {
			allPassed = false
		}
		fmt.Printf("    Split(%q, %q) -> %d partes  ok:%v\n", tc[0], tc[1], len(parts), passed)
	}
	fmt.Printf("  Resultado: %v\n", statusStr(allPassed))

	// ============================================
	// 6. PATRON: TESTDATA Y GOLDEN FILES
	// ============================================
	fmt.Println("\n--- 6. Patron: testdata y golden files ---")
	os.Stdout.WriteString(`
PATRONES DE TESTING EN GO:

  TESTDATA DIRECTORY:
  - Directorio "testdata/" es ignorado por go build
  - Usado para fixtures, archivos de entrada, expected outputs
  - Acceso relativo al paquete: "testdata/input.json"

  GOLDEN FILES:
  - Archivos con el output esperado
  - Se actualizan con flag -update: go test -update
  - Patron: comparar output vs golden, actualizar si flag activo

  Ejemplo de estructura:
    mypackage/
      handler.go
      handler_test.go
      testdata/
        input1.json
        input1.golden       <- output esperado
        input2.json
        input2.golden
        fixtures/
          users.sql

  Codigo tipico:
    golden := filepath.Join("testdata", tc.name+".golden")
    if *update {
        os.WriteFile(golden, got, 0644)
    }
    want, _ := os.ReadFile(golden)
    if !bytes.Equal(got, want) {
        t.Errorf("mismatch: got %q, want %q", got, want)
    }
`)

	fmt.Println("  Simulando patron golden files:")
	type TestCase struct {
		Name  string
		Input string
	}
	cases := []TestCase{
		{"uppercase", "hello world"},
		{"trim", "  spaces  "},
		{"replace", "foo-bar-baz"},
	}

	goldenStore := map[string]string{
		"uppercase": "HELLO WORLD",
		"trim":      "spaces",
		"replace":   "foo_bar_baz",
	}

	for _, tc := range cases {
		var got string
		switch tc.Name {
		case "uppercase":
			got = strings.ToUpper(tc.Input)
		case "trim":
			got = strings.TrimSpace(tc.Input)
		case "replace":
			got = strings.ReplaceAll(tc.Input, "-", "_")
		}

		want := goldenStore[tc.Name]
		if got == want {
			fmt.Printf("    [PASS] %s: %q -> %q\n", tc.Name, tc.Input, got)
		} else {
			fmt.Printf("    [FAIL] %s: got %q, want %q\n", tc.Name, got, want)
		}
	}

	// ============================================
	// 7. EJEMPLO: VALIDAR READER CUSTOM
	// ============================================
	fmt.Println("\n--- 7. Ejemplo: Reader custom con iotest ---")

	upperR := &upperReader{r: strings.NewReader("hello world")}
	content2, _ := io.ReadAll(upperR)
	fmt.Printf("  upperReader: %q\n", string(content2))

	fmt.Println("  Testeando con HalfReader:")
	upperHalf := &upperReader{r: iotest.HalfReader(strings.NewReader("test data"))}
	content3, _ := io.ReadAll(upperHalf)
	fmt.Printf("    Resultado: %q\n", string(content3))

	fmt.Println("  Testeando con OneByteReader:")
	upperOne := &upperReader{r: iotest.OneByteReader(strings.NewReader("one byte"))}
	content4, _ := io.ReadAll(upperOne)
	fmt.Printf("    Resultado: %q\n", string(content4))

	// ============================================
	// 8. EJEMPLO: FS PARA TESTING DE CONFIGURACION
	// ============================================
	fmt.Println("\n--- 8. Ejemplo: Cargar config desde fs.FS ---")

	configFS := fstest.MapFS{
		"config/database.yaml": {Data: []byte("host: db.example.com\nport: 5432\nname: myapp\n")},
		"config/cache.yaml":    {Data: []byte("host: redis.example.com\nport: 6379\nttl: 300\n")},
		"config/app.yaml":      {Data: []byte("name: MyApp\nversion: 2.0\ndebug: false\n")},
	}

	fmt.Println("  Archivos de configuracion:")
	configEntries, _ := configFS.ReadDir("config")
	for _, entry := range configEntries {
		data, _ := configFS.ReadFile("config/" + entry.Name())
		fmt.Printf("\n    [%s]\n", entry.Name())
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			fmt.Printf("      %s\n", line)
		}
	}

	fmt.Println("\n=== FIN CAPITULO 060 ===")
}

type repeatReader struct {
	data    []byte
	repeats int
	pos     int
	count   int
}

func (r *repeatReader) Read(p []byte) (int, error) {
	if r.count >= r.repeats && r.pos == 0 {
		return 0, io.EOF
	}
	n := 0
	for n < len(p) {
		if r.count >= r.repeats {
			break
		}
		p[n] = r.data[r.pos]
		n++
		r.pos++
		if r.pos >= len(r.data) {
			r.pos = 0
			r.count++
		}
	}
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

type upperReader struct {
	r io.Reader
}

func (u *upperReader) Read(p []byte) (int, error) {
	n, err := u.r.Read(p)
	for i := 0; i < n; i++ {
		if p[i] >= 'a' && p[i] <= 'z' {
			p[i] -= 32
		}
	}
	return n, err
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func sortCopy(s []int) []int {
	cp := make([]int, len(s))
	copy(cp, s)
	sort.Ints(cp)
	return cp
}

func slicesEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func statusStr(passed bool) string {
	if passed {
		return "TODOS PASARON"
	}
	return "ALGUNO FALLO"
}

/*
SUMMARY - TESTING STDLIB UTILITIES:

TESTING/IOTEST - SPECIAL READERS FOR EDGE CASES:
- HalfReader reads only half the requested bytes
- OneByteReader reads exactly 1 byte at a time
- DataErrReader returns error alongside final data
- ErrReader always returns a specified error
- TimeoutReader simulates timeout on second read
- NewReadLogger/NewWriteLogger for debugging I/O
- TestReader validates io.Reader implementations thoroughly

TESTING/FSTEST - VIRTUAL FILESYSTEM:
- fstest.MapFS is an in-memory fs.FS implementation
- Supports Open, ReadFile, ReadDir, Glob, Stat operations
- MapFile with Data, Mode, ModTime fields
- fstest.TestFS validates custom fs.FS implementations
- Ideal for testing code that accepts fs.FS interface

PROPERTY-BASED TESTING (MANUAL):
- Verify invariant properties for any input
- reverse(reverse(s)) == s, sort is idempotent
- len(split(s, sep)) >= 1 for any string and separator
- Generate random inputs and check properties hold

TESTDATA AND GOLDEN FILES PATTERN:
- testdata/ directory ignored by go build
- Golden files store expected output for comparison
- -update flag to regenerate golden files
- Compare actual output vs stored golden file
*/
