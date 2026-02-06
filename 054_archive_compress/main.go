// Package main - Chapter 054: Archive & Compress
// Creacion y extraccion de archivos tar/zip, compresion y descompresion
// con gzip, zlib, flate, y procesamiento de archivos en streaming.
package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== ARCHIVE & COMPRESS ===")

	// ============================================
	// 1. ARCHIVE/TAR - CREAR ARCHIVOS TAR
	// ============================================
	fmt.Println("\n--- 1. archive/tar - Crear archivo tar ---")
	os.Stdout.WriteString(`
ARCHIVE/TAR - FORMATO TAR:

  tar.NewWriter(w io.Writer)           // Crear escritor tar
  tw.WriteHeader(hdr *tar.Header)      // Escribir cabecera de archivo
  tw.Write(data []byte)                // Escribir contenido del archivo
  tw.Close()                           // Cerrar (flush final)

  tar.NewReader(r io.Reader)           // Crear lector tar
  tr.Next() (*tar.Header, error)       // Siguiente entrada (io.EOF al final)
  io.ReadAll(tr)                       // Leer contenido de la entrada

  tar.Header campos importantes:
    Name, Size, Mode, ModTime, Typeflag (TypeReg, TypeDir, TypeSymlink)
`)

	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	files := []struct {
		name    string
		content string
		mode    int64
	}{
		{"readme.txt", "Bienvenido al proyecto Go!", 0644},
		{"src/main.go", "package main\n\nfunc main() {}\n", 0644},
		{"src/util.go", "package main\n\nfunc helper() string { return \"ok\" }\n", 0644},
		{"config.json", `{"version": "1.0", "debug": false}`, 0600},
	}

	for _, f := range files {
		hdr := &tar.Header{
			Name:    f.name,
			Size:    int64(len(f.content)),
			Mode:    f.mode,
			ModTime: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			fmt.Printf("  Error en header: %v\n", err)
			continue
		}
		if _, err := tw.Write([]byte(f.content)); err != nil {
			fmt.Printf("  Error escribiendo: %v\n", err)
		}
		fmt.Printf("  Agregado: %-20s (%d bytes, mode: %o)\n", f.name, len(f.content), f.mode)
	}
	tw.Close()
	fmt.Printf("  Tamano total tar: %d bytes\n", tarBuf.Len())

	// ============================================
	// 2. ARCHIVE/TAR - LEER ARCHIVOS TAR
	// ============================================
	fmt.Println("\n--- 2. archive/tar - Leer archivo tar ---")

	tr := tar.NewReader(bytes.NewReader(tarBuf.Bytes()))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			break
		}
		content, _ := io.ReadAll(tr)
		fmt.Printf("  Archivo: %-20s Size:%-4d Mode:%o ModTime:%s\n",
			hdr.Name, hdr.Size, hdr.Mode, hdr.ModTime.Format("2006-01-02"))
		if len(content) < 60 {
			fmt.Printf("    Contenido: %s\n", strings.TrimSpace(string(content)))
		}
	}

	// ============================================
	// 3. ARCHIVE/ZIP - CREAR ARCHIVOS ZIP
	// ============================================
	fmt.Println("\n--- 3. archive/zip - Crear archivo zip ---")
	os.Stdout.WriteString(`
ARCHIVE/ZIP - FORMATO ZIP:

  zip.NewWriter(w io.Writer)           // Crear escritor zip
  zw.Create(name string)               // Crear archivo simple
  zw.CreateHeader(fh *zip.FileHeader)  // Crear con cabecera custom
  zw.Close()                           // Cerrar y escribir directorio central

  zip.NewReader(r io.ReaderAt, size)   // Leer zip desde ReaderAt
  zip.OpenReader(name string)          // Abrir zip desde disco
  zr.File                              // Lista de archivos
  f.Open()                             // Abrir un archivo para leer

  zip.FileHeader:
    Name, Method (Store/Deflate), Modified, ExternalAttrs
`)

	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	for _, f := range files {
		header := &zip.FileHeader{
			Name:     f.name,
			Method:   zip.Deflate,
			Modified: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
		}
		w, err := zw.CreateHeader(header)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		n, _ := w.Write([]byte(f.content))
		fmt.Printf("  Comprimido: %-20s (%d bytes originales)\n", f.name, n)
	}
	zw.Close()
	fmt.Printf("  Tamano total zip: %d bytes\n", zipBuf.Len())

	// ============================================
	// 4. ARCHIVE/ZIP - LEER ARCHIVOS ZIP
	// ============================================
	fmt.Println("\n--- 4. archive/zip - Leer archivo zip ---")

	zipReader, err := zip.NewReader(bytes.NewReader(zipBuf.Bytes()), int64(zipBuf.Len()))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		for _, f := range zipReader.File {
			rc, err := f.Open()
			if err != nil {
				fmt.Printf("  Error abriendo %s: %v\n", f.Name, err)
				continue
			}
			content, _ := io.ReadAll(rc)
			rc.Close()
			fmt.Printf("  Archivo: %-20s Comprimido:%-4d Original:%-4d Metodo:%s\n",
				f.Name, f.CompressedSize64, f.UncompressedSize64, methodName(f.Method))
			_ = content
		}
	}

	// ============================================
	// 5. COMPRESS/GZIP
	// ============================================
	fmt.Println("\n--- 5. compress/gzip ---")
	os.Stdout.WriteString(`
COMPRESS/GZIP:

  gzip.NewWriter(w io.Writer)          // Compresor gzip
  gzip.NewWriterLevel(w, level)        // Con nivel (BestSpeed..BestCompression)
  gw.Write(data)                       // Escribir datos comprimidos
  gw.Close()                           // Flush y cerrar (OBLIGATORIO)

  gzip.NewReader(r io.Reader)          // Descompresor gzip
  io.ReadAll(gr)                       // Leer datos descomprimidos
  gr.Close()                           // Cerrar

  gzip.Header: Name, Comment, OS, ModTime, Extra
`)

	original := strings.Repeat("Go es un lenguaje eficiente y concurrente. ", 100)
	fmt.Printf("  Texto original: %d bytes\n", len(original))

	var gzBuf bytes.Buffer
	gzWriter, _ := gzip.NewWriterLevel(&gzBuf, gzip.BestCompression)
	gzWriter.Name = "mensaje.txt"
	gzWriter.Comment = "Texto repetitivo para demo"
	gzWriter.Write([]byte(original))
	gzWriter.Close()
	fmt.Printf("  Comprimido gzip (BestCompression): %d bytes (%.1f%%)\n",
		gzBuf.Len(), float64(gzBuf.Len())/float64(len(original))*100)

	var gzFast bytes.Buffer
	gzFastWriter, _ := gzip.NewWriterLevel(&gzFast, gzip.BestSpeed)
	gzFastWriter.Write([]byte(original))
	gzFastWriter.Close()
	fmt.Printf("  Comprimido gzip (BestSpeed):       %d bytes (%.1f%%)\n",
		gzFast.Len(), float64(gzFast.Len())/float64(len(original))*100)

	gzReader, err := gzip.NewReader(bytes.NewReader(gzBuf.Bytes()))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		decompressed, _ := io.ReadAll(gzReader)
		gzReader.Close()
		fmt.Printf("  Descomprimido: %d bytes (nombre: %s)\n", len(decompressed), gzReader.Name)
		fmt.Printf("  Coincide con original: %v\n", string(decompressed) == original)
	}

	// ============================================
	// 6. COMPRESS/ZLIB
	// ============================================
	fmt.Println("\n--- 6. compress/zlib ---")
	os.Stdout.WriteString(`
COMPRESS/ZLIB (RFC 1950):

  zlib.NewWriter(w)                    // Compresor zlib
  zlib.NewWriterLevel(w, level)        // Con nivel
  zlib.NewReader(r)                    // Descompresor

Diferencia con gzip:
  - gzip = zlib + cabecera con nombre/fecha/OS
  - zlib es mas ligero (menos overhead)
  - Usado internamente por PNG, HTTP deflate, etc.
`)

	var zlibBuf bytes.Buffer
	zlibWriter, _ := zlib.NewWriterLevel(&zlibBuf, zlib.BestCompression)
	zlibWriter.Write([]byte(original))
	zlibWriter.Close()
	fmt.Printf("  Comprimido zlib: %d bytes (%.1f%%)\n",
		zlibBuf.Len(), float64(zlibBuf.Len())/float64(len(original))*100)

	zlibReader, _ := zlib.NewReader(bytes.NewReader(zlibBuf.Bytes()))
	zlibDecomp, _ := io.ReadAll(zlibReader)
	zlibReader.Close()
	fmt.Printf("  Descomprimido: %d bytes, coincide: %v\n", len(zlibDecomp), string(zlibDecomp) == original)

	// ============================================
	// 7. COMPRESS/FLATE (DEFLATE RAW)
	// ============================================
	fmt.Println("\n--- 7. compress/flate (Deflate raw) ---")
	os.Stdout.WriteString(`
COMPRESS/FLATE (RFC 1951):

  flate.NewWriter(w, level)            // Compresor deflate puro
  flate.NewReader(r)                   // Descompresor

  Es el algoritmo base de gzip y zlib.
  Niveles: NoCompression(0), BestSpeed(1) ... BestCompression(9)
           DefaultCompression(-1), HuffmanOnly(-2)
`)

	var flateBuf bytes.Buffer
	flateWriter, _ := flate.NewWriter(&flateBuf, flate.BestCompression)
	flateWriter.Write([]byte(original))
	flateWriter.Close()
	fmt.Printf("  Comprimido flate: %d bytes (%.1f%%)\n",
		flateBuf.Len(), float64(flateBuf.Len())/float64(len(original))*100)

	flateReader := flate.NewReader(bytes.NewReader(flateBuf.Bytes()))
	flateDecomp, _ := io.ReadAll(flateReader)
	flateReader.Close()
	fmt.Printf("  Descomprimido: %d bytes, coincide: %v\n", len(flateDecomp), string(flateDecomp) == original)

	fmt.Printf("\n  Comparacion de formatos para %d bytes:\n", len(original))
	fmt.Printf("    gzip (BestCompression): %4d bytes\n", gzBuf.Len())
	fmt.Printf("    zlib (BestCompression): %4d bytes\n", zlibBuf.Len())
	fmt.Printf("    flate(BestCompression): %4d bytes\n", flateBuf.Len())

	// ============================================
	// 8. EJEMPLO: TAR.GZ STREAMING
	// ============================================
	fmt.Println("\n--- 8. Ejemplo: Crear y leer tar.gz ---")

	var tarGzBuf bytes.Buffer
	gzw := gzip.NewWriter(&tarGzBuf)
	twGz := tar.NewWriter(gzw)

	tarFiles := map[string]string{
		"app/main.go":   "package main\n\nimport \"fmt\"\n\nfunc main() { fmt.Println(\"hello\") }\n",
		"app/go.mod":    "module myapp\n\ngo 1.22\n",
		"app/README.md": "# Mi Aplicacion\n\nEjemplo de tar.gz en Go.\n",
	}

	for name, content := range tarFiles {
		hdr := &tar.Header{
			Name:    name,
			Size:    int64(len(content)),
			Mode:    0644,
			ModTime: time.Now(),
		}
		twGz.WriteHeader(hdr)
		twGz.Write([]byte(content))
	}
	twGz.Close()
	gzw.Close()

	totalOriginal := 0
	for _, c := range tarFiles {
		totalOriginal += len(c)
	}
	fmt.Printf("  Creado tar.gz: %d bytes (original: %d bytes)\n", tarGzBuf.Len(), totalOriginal)

	gzr, _ := gzip.NewReader(bytes.NewReader(tarGzBuf.Bytes()))
	trGz := tar.NewReader(gzr)
	fmt.Println("  Contenido del tar.gz:")
	for {
		hdr, err := trGz.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		data, _ := io.ReadAll(trGz)
		fmt.Printf("    %s (%d bytes)\n", hdr.Name, len(data))
	}
	gzr.Close()

	// ============================================
	// 9. EJEMPLO: ZIP CON DIRECTORIOS
	// ============================================
	fmt.Println("\n--- 9. Ejemplo: ZIP con directorios ---")

	var zipBuf2 bytes.Buffer
	zw2 := zip.NewWriter(&zipBuf2)

	dirs := []string{"project/", "project/src/", "project/docs/"}
	for _, d := range dirs {
		header := &zip.FileHeader{Name: d}
		_, err := zw2.CreateHeader(header)
		if err != nil {
			fmt.Printf("  Error creando dir %s: %v\n", d, err)
		}
	}

	zipFiles := map[string]string{
		"project/src/app.go":   "package main\n",
		"project/docs/api.md":  "# API Documentation\n",
		"project/README.md":    "# Project\n",
	}
	for name, content := range zipFiles {
		w, _ := zw2.Create(name)
		w.Write([]byte(content))
	}
	zw2.Close()

	zipR2, _ := zip.NewReader(bytes.NewReader(zipBuf2.Bytes()), int64(zipBuf2.Len()))
	for _, f := range zipR2.File {
		kind := "FILE"
		if strings.HasSuffix(f.Name, "/") {
			kind = "DIR "
		}
		fmt.Printf("  [%s] %s\n", kind, f.Name)
	}

	fmt.Println("\n=== FIN CAPITULO 054 ===")
}

func methodName(method uint16) string {
	switch method {
	case zip.Store:
		return "Store"
	case zip.Deflate:
		return "Deflate"
	default:
		return fmt.Sprintf("Unknown(%d)", method)
	}
}
