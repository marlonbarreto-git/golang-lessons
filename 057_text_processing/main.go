// Package main - Chapter 057: Text Processing
// Tokenizacion con text/scanner, formateo tabular con text/tabwriter,
// tipos MIME con mime, y formularios multipart con mime/multipart.
package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"strings"
	"text/scanner"
	"text/tabwriter"
)

func main() {
	fmt.Println("=== TEXT PROCESSING ===")

	// ============================================
	// 1. TEXT/SCANNER - TOKENIZACION
	// ============================================
	fmt.Println("\n--- 1. text/scanner - Tokenizacion ---")
	os.Stdout.WriteString(`
TEXT/SCANNER - TOKENIZADOR CONFIGURABLE:

  scanner.Scanner{}                        // Crear scanner
  s.Init(r io.Reader)                      // Inicializar con fuente
  s.Scan() rune                            // Siguiente token (retorna tipo)
  s.TokenText() string                     // Texto del token actual
  s.Pos() scanner.Position                 // Posicion actual

  Tipos de token (rune):
    scanner.EOF        // Fin de entrada
    scanner.Ident      // Identificador
    scanner.Int        // Entero
    scanner.Float      // Flotante
    scanner.Char       // Caracter 'x'
    scanner.String     // String "x"
    scanner.RawString  // Raw string ` + "`" + `x` + "`" + `
    scanner.Comment    // Comentario

  s.Mode = bits para activar/desactivar tipos:
    scanner.ScanIdents | scanner.ScanFloats | scanner.ScanStrings
`)

	src := `package main

import "fmt"

func main() {
    x := 42
    pi := 3.14159
    name := "Go"
    fmt.Println(x, pi, name)
}`

	var s scanner.Scanner
	s.Init(strings.NewReader(src))
	s.Filename = "example.go"

	tokenCounts := map[string]int{}
	fmt.Println("  Tokens encontrados:")
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		typeName := tokenTypeName(tok)
		tokenCounts[typeName]++
		if len(tokenCounts) <= 12 {
			fmt.Printf("    [%-6s] %-15s (linea %d, col %d)\n",
				typeName, s.TokenText(), s.Pos().Line, s.Pos().Column)
		}
	}
	fmt.Println("    ...")
	fmt.Println("  Resumen de tokens:")
	for typ, count := range tokenCounts {
		fmt.Printf("    %-10s: %d\n", typ, count)
	}

	// Scanner con modo personalizado
	fmt.Println("\n  Scanner solo para numeros:")
	var numScanner scanner.Scanner
	numScanner.Init(strings.NewReader("La temp es 23.5 grados, humedad 87%, presion 1013"))
	numScanner.Mode = scanner.ScanInts | scanner.ScanFloats

	for tok := numScanner.Scan(); tok != scanner.EOF; tok = numScanner.Scan() {
		if tok == scanner.Int || tok == scanner.Float {
			fmt.Printf("    Numero encontrado: %s\n", numScanner.TokenText())
		}
	}

	// ============================================
	// 2. TEXT/SCANNER - TOKENIZADOR CUSTOM
	// ============================================
	fmt.Println("\n--- 2. text/scanner - Tokenizador de config ---")

	configSrc := `
server.host = "localhost"
server.port = 8080
database.name = "mydb"
database.pool_size = 25
debug = true
timeout = 30.5
`
	var configScanner scanner.Scanner
	configScanner.Init(strings.NewReader(configSrc))
	configScanner.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings

	config := parseConfig(&configScanner)
	fmt.Println("  Configuracion parseada:")
	for k, v := range config {
		fmt.Printf("    %s = %s\n", k, v)
	}

	// ============================================
	// 3. TEXT/TABWRITER - FORMATEO TABULAR
	// ============================================
	fmt.Println("\n--- 3. text/tabwriter - Formateo tabular ---")
	os.Stdout.WriteString(`
TEXT/TABWRITER - TABLAS ALINEADAS:

  tabwriter.NewWriter(w, minwidth, tabwidth, padding, padchar, flags)
  tw.Write([]byte(...))                    // Escribir filas con tabs
  tw.Flush()                               // Alinear y flush

  Parametros:
    minwidth: ancho minimo de celda
    tabwidth: ancho de tab para padding
    padding:  espacios extra entre columnas
    padchar:  caracter de relleno (' ' o '\t')
    flags:    AlignRight, Debug, etc.

  Las columnas se separan con '\t'
  Filas se separan con '\n'
`)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  LENGUAJE\tANO\tCREADOR\tTIPO")
	fmt.Fprintln(tw, "  --------\t---\t-------\t----")
	fmt.Fprintln(tw, "  Go\t2009\tRobert Griesemer, Rob Pike\tCompilado")
	fmt.Fprintln(tw, "  Rust\t2010\tGraydon Hoare\tCompilado")
	fmt.Fprintln(tw, "  Python\t1991\tGuido van Rossum\tInterpretado")
	fmt.Fprintln(tw, "  Java\t1995\tJames Gosling\tCompilado (JVM)")
	fmt.Fprintln(tw, "  TypeScript\t2012\tAnders Hejlsberg\tTranspilado")
	tw.Flush()

	// Tabwriter con padding y alineacion
	fmt.Println("\n  Tabla con separadores:")
	tw2 := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	fmt.Fprintln(tw2, "  Nombre\tEdad\tCiudad\tRol")
	fmt.Fprintln(tw2, "  Ana\t28\tMadrid\tBackend")
	fmt.Fprintln(tw2, "  Carlos\t35\tBogota\tDevOps")
	fmt.Fprintln(tw2, "  Maria\t42\tSantiago\tArquitecta")
	tw2.Flush()

	// Tabwriter para CLI help
	fmt.Println("\n  Formato estilo CLI help:")
	tw3 := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(tw3, "  -p, --port\tPuerto del servidor\t(default: 8080)")
	fmt.Fprintln(tw3, "  -h, --host\tHost de escucha\t(default: localhost)")
	fmt.Fprintln(tw3, "  -d, --debug\tModo debug\t(default: false)")
	fmt.Fprintln(tw3, "  -c, --config\tArchivo de configuracion\t(default: config.yaml)")
	fmt.Fprintln(tw3, "  -v, --verbose\tSalida detallada\t(default: false)")
	tw3.Flush()

	// Tabwriter a string buffer
	fmt.Println("\n  Tabwriter a buffer (para tests/logs):")
	var buf bytes.Buffer
	tw4 := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw4, "Metrica\tValor\tUnidad")
	fmt.Fprintln(tw4, "CPU\t85.3\t%")
	fmt.Fprintln(tw4, "Memoria\t4.2\tGB")
	fmt.Fprintln(tw4, "Disco\t120\tGB")
	tw4.Flush()
	fmt.Print("  " + strings.ReplaceAll(buf.String(), "\n", "\n  "))

	// ============================================
	// 4. MIME - TIPOS MIME
	// ============================================
	fmt.Println("\n--- 4. mime - Tipos MIME ---")
	os.Stdout.WriteString(`
MIME - IDENTIFICACION DE TIPOS DE CONTENIDO:

  mime.TypeByExtension(ext string)         // Tipo MIME por extension
  mime.ExtensionsByType(typ string)        // Extensiones por tipo MIME
  mime.FormatMediaType(t, params)          // Formatear tipo con parametros
  mime.ParseMediaType(v string)            // Parsear tipo y parametros
  mime.AddExtensionType(ext, typ)          // Registrar extension custom
`)

	extensions := []string{".html", ".json", ".png", ".pdf", ".go", ".js", ".css", ".xml", ".csv", ".mp4"}
	for _, ext := range extensions {
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "(no registrado)"
		}
		fmt.Printf("  %-6s -> %s\n", ext, mimeType)
	}

	exts, _ := mime.ExtensionsByType("text/html")
	fmt.Printf("\n  Extensiones para text/html: %v\n", exts)

	mediaType := mime.FormatMediaType("multipart/form-data", map[string]string{
		"boundary": "----FormBoundary123",
	})
	fmt.Printf("  FormatMediaType: %s\n", mediaType)

	parsedType, params, err := mime.ParseMediaType("text/html; charset=utf-8; boundary=something")
	if err == nil {
		fmt.Printf("  ParseMediaType: type=%s charset=%s boundary=%s\n",
			parsedType, params["charset"], params["boundary"])
	}

	// ============================================
	// 5. MIME/MULTIPART - ESCRITURA
	// ============================================
	fmt.Println("\n--- 5. mime/multipart - Crear formulario ---")
	os.Stdout.WriteString(`
MIME/MULTIPART - FORMULARIOS Y ARCHIVOS:

  ESCRITURA:
  multipart.NewWriter(w io.Writer)         // Crear escritor multipart
  mw.FormDataContentType()                 // Content-Type con boundary
  mw.WriteField(name, value)               // Campo de texto
  mw.CreateFormFile(fieldname, filename)   // Campo de archivo
  mw.CreateFormField(fieldname)            // Campo con io.Writer
  mw.CreatePart(header)                    // Parte custom
  mw.Boundary()                            // Obtener boundary
  mw.Close()                               // Cerrar (boundary final)

  LECTURA:
  multipart.NewReader(r, boundary)         // Crear lector
  mr.NextPart() (*Part, error)             // Siguiente parte
  mr.ReadForm(maxMemory)                   // Leer todo como Form
  part.FormName()                          // Nombre del campo
  part.FileName()                          // Nombre del archivo
`)

	var multipartBuf bytes.Buffer
	mw := multipart.NewWriter(&multipartBuf)

	mw.WriteField("username", "gopher")
	mw.WriteField("email", "gopher@golang.org")
	mw.WriteField("bio", "Soy un gopher apasionado por Go")

	fileWriter, _ := mw.CreateFormFile("avatar", "profile.png")
	fileWriter.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	fileWriter.Write(bytes.Repeat([]byte{0xFF}, 100))

	docWriter, _ := mw.CreateFormFile("resume", "cv.pdf")
	docWriter.Write([]byte("%PDF-1.4 fake content for demo"))

	contentType := mw.FormDataContentType()
	boundary := mw.Boundary()
	mw.Close()

	fmt.Printf("  Content-Type: %s\n", contentType)
	fmt.Printf("  Boundary: %s\n", boundary)
	fmt.Printf("  Tamano total: %d bytes\n", multipartBuf.Len())

	// ============================================
	// 6. MIME/MULTIPART - LECTURA
	// ============================================
	fmt.Println("\n--- 6. mime/multipart - Leer formulario ---")

	mr := multipart.NewReader(bytes.NewReader(multipartBuf.Bytes()), boundary)
	fmt.Println("  Leyendo partes:")
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			break
		}

		data, _ := io.ReadAll(part)
		if part.FileName() != "" {
			fmt.Printf("    [Archivo] campo=%s nombre=%s tamano=%d bytes\n",
				part.FormName(), part.FileName(), len(data))
		} else {
			fmt.Printf("    [Campo]   %s = %s\n", part.FormName(), string(data))
		}
		part.Close()
	}

	// ReadForm para leer todo de una vez
	fmt.Println("\n  Usando ReadForm():")
	mr2 := multipart.NewReader(bytes.NewReader(multipartBuf.Bytes()), boundary)
	form, err := mr2.ReadForm(10 << 20) // 10 MB max en memoria
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		for key, values := range form.Value {
			fmt.Printf("    Campo: %s = %v\n", key, values)
		}
		for key, files := range form.File {
			for _, fh := range files {
				fmt.Printf("    Archivo: campo=%s nombre=%s tamano=%d\n",
					key, fh.Filename, fh.Size)
			}
		}
		form.RemoveAll()
	}

	// ============================================
	// 7. EJEMPLO: GENERADOR DE REPORTES TABULARES
	// ============================================
	fmt.Println("\n--- 7. Ejemplo: Generador de reportes ---")

	type Record struct {
		Name   string
		Status string
		CPU    float64
		Mem    float64
	}

	records := []Record{
		{"api-server", "Running", 45.2, 512.0},
		{"worker-1", "Running", 78.9, 1024.0},
		{"worker-2", "Stopped", 0.0, 0.0},
		{"scheduler", "Running", 12.3, 256.0},
		{"cache", "Running", 5.1, 2048.0},
	}

	var reportBuf bytes.Buffer
	tw5 := tabwriter.NewWriter(&reportBuf, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw5, "SERVICE\tSTATUS\tCPU(%)\tMEM(MB)")
	fmt.Fprintln(tw5, "-------\t------\t------\t-------")
	for _, r := range records {
		fmt.Fprintf(tw5, "%s\t%s\t%.1f\t%.0f\n", r.Name, r.Status, r.CPU, r.Mem)
	}
	tw5.Flush()

	for _, line := range strings.Split(reportBuf.String(), "\n") {
		if line != "" {
			fmt.Printf("  %s\n", line)
		}
	}

	fmt.Println("\n=== FIN CAPITULO 057 ===")
}

func tokenTypeName(tok rune) string {
	switch tok {
	case scanner.EOF:
		return "EOF"
	case scanner.Ident:
		return "Ident"
	case scanner.Int:
		return "Int"
	case scanner.Float:
		return "Float"
	case scanner.Char:
		return "Char"
	case scanner.String:
		return "String"
	case scanner.RawString:
		return "Raw"
	case scanner.Comment:
		return "Cmt"
	default:
		return fmt.Sprintf("Op(%c)", tok)
	}
}

func parseConfig(s *scanner.Scanner) map[string]string {
	config := make(map[string]string)
	for {
		tok := s.Scan()
		if tok == scanner.EOF {
			break
		}
		if tok != scanner.Ident {
			continue
		}

		key := s.TokenText()
		for {
			tok = s.Scan()
			if tok == scanner.EOF {
				break
			}
			if s.TokenText() == "." {
				tok = s.Scan()
				key += "." + s.TokenText()
				continue
			}
			break
		}

		if s.TokenText() != "=" {
			continue
		}

		tok = s.Scan()
		if tok == scanner.EOF {
			break
		}
		val := s.TokenText()
		if tok == scanner.String {
			val = strings.Trim(val, `"`)
		}
		config[key] = val
	}
	return config
}

/*
SUMMARY - TEXT PROCESSING:

TEXT/SCANNER - CONFIGURABLE TOKENIZER:
- scanner.Scanner with Init(reader) and Scan() loop returning token types
- Token types: Ident, Int, Float, Char, String, RawString, Comment, EOF
- s.Mode bitmask controls which token types to recognize
- s.TokenText() returns current token, s.Pos() returns position
- Custom config parser example using dotted keys (server.host)

TEXT/TABWRITER - ALIGNED TABULAR OUTPUT:
- tabwriter.NewWriter(w, minwidth, tabwidth, padding, padchar, flags)
- Columns separated by \t, rows by \n, call Flush() to align
- tabwriter.Debug flag adds column separators
- Useful for CLI help text, reports, and formatted tables
- Can write to any io.Writer including bytes.Buffer

MIME - CONTENT TYPE IDENTIFICATION:
- mime.TypeByExtension and ExtensionsByType for MIME lookups
- mime.FormatMediaType and ParseMediaType for Content-Type headers
- mime.AddExtensionType to register custom extensions

MIME/MULTIPART - FORM DATA AND FILE UPLOADS:
- multipart.NewWriter for creating multipart form data
- WriteField for text fields, CreateFormFile for file uploads
- FormDataContentType returns the full Content-Type with boundary
- multipart.NewReader with NextPart() loop or ReadForm() to parse
- Part.FormName() and Part.FileName() identify field vs file
*/
