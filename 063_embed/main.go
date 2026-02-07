// Package main - Chapter 063: go:embed
// Go 1.16+ permite embeber archivos y directorios directamente en el binario.
// Perfecto para assets, templates, configuración y distribución single-binary.
package main

import (
	"fmt"
	"strings"
)

// ============================================
// EJEMPLOS DE EMBED (comentados porque necesitan archivos)
// ============================================
//
// Para usar go:embed, necesitas los archivos reales.
// Estos son ejemplos de sintaxis:
//
// Embeber un solo archivo como string:
// //go:embed version.txt
// var versionString string
//
// Embeber un solo archivo como []byte:
// //go:embed config.json
// var configBytes []byte
//
// Embeber múltiples archivos en embed.FS:
// //go:embed templates/*
// var templatesFS embed.FS
//
// Embeber directorio completo:
// //go:embed static
// var staticFS embed.FS
//
// Embeber con patrones:
// //go:embed assets/*.css assets/*.js
// var assetsFS embed.FS

func main() {
	fmt.Println("=== GO:EMBED ===")

	// ============================================
	// SINTAXIS BÁSICA
	// ============================================
	fmt.Println("\n--- Sintaxis Básica ---")
	fmt.Println(`
DIRECTIVA go:embed:

// Embeber como string (un archivo)
//go:embed file.txt
var content string

// Embeber como []byte (un archivo)
//go:embed file.bin
var data []byte

// Embeber como embed.FS (uno o más archivos)
//go:embed file.txt
var fs embed.FS

//go:embed dir/*
var dirFS embed.FS

//go:embed *.html templates/*.tmpl
var multiFS embed.FS

REGLAS:
1. La directiva debe estar JUSTO antes de la variable
2. No puede haber líneas en blanco entre //go:embed y var
3. Solo funciona con variables a nivel de paquete
4. Los paths son relativos al archivo .go`)
	// ============================================
	// STRING Y []BYTE
	// ============================================
	fmt.Println("\n--- String y []byte ---")

	// Simular archivos embebidos (en producción serían reales)
	fmt.Printf("Versión (string): %s\n", strings.TrimSpace(mockVersion))
	fmt.Printf("Config (bytes): %d bytes\n", len(mockConfig))

	fmt.Println(`
CUÁNDO USAR STRING:
- Archivos de texto pequeños
- Templates simples
- Configuración estática
- SQL queries

CUÁNDO USAR []BYTE:
- Archivos binarios
- Imágenes, fuentes
- Archivos que necesitas procesar como bytes`)
	// ============================================
	// EMBED.FS
	// ============================================
	fmt.Println("\n--- embed.FS ---")
	fmt.Println(`
embed.FS implementa:
- fs.FS
- fs.ReadDirFS
- fs.ReadFileFS

MÉTODOS:
fs.Open(name string) (fs.File, error)
fs.ReadDir(name string) ([]fs.DirEntry, error)
fs.ReadFile(name string) ([]byte, error)

EJEMPLO:
//go:embed templates/*
var templates embed.FS

// Leer archivo
data, err := templates.ReadFile("templates/index.html")

// Listar directorio
entries, err := templates.ReadDir("templates")

// Abrir como fs.File
file, err := templates.Open("templates/style.css")`)
	// ============================================
	// PATRONES DE EMBED
	// ============================================
	fmt.Println("\n--- Patrones de Embed ---")
	fmt.Println(`
PATRONES SOPORTADOS:

// Un archivo específico
//go:embed config.yaml

// Todos los archivos en directorio (no recursivo)
//go:embed templates/*

// Recursivo
//go:embed templates

// Múltiples patrones
//go:embed *.html *.css *.js

// Múltiples líneas
//go:embed file1.txt
//go:embed file2.txt
//go:embed dir/*
var combined embed.FS

// Con subdirectorios
//go:embed static/css/* static/js/*

ARCHIVOS OCULTOS:
- Por defecto, archivos que empiezan con . o _ se ignoran
- Para incluirlos, usar "all:" prefix:
//go:embed all:templates`)
	// ============================================
	// USO CON HTTP
	// ============================================
	fmt.Println("\n--- Uso con HTTP Server ---")
	fmt.Println(`
SERVIR ARCHIVOS ESTÁTICOS:

//go:embed static
var staticFiles embed.FS

func main() {
    // Opción 1: Servir directamente
    http.Handle("/static/", http.FileServer(http.FS(staticFiles)))

    // Opción 2: Con StripPrefix si el embed tiene subdirectorio
    subFS, _ := fs.Sub(staticFiles, "static")
    http.Handle("/", http.FileServer(http.FS(subFS)))

    http.ListenAndServe(":8080", nil)
}

CON CHI/GORILLA:
r := chi.NewRouter()
r.Handle("/static/*", http.StripPrefix("/static/",
    http.FileServer(http.FS(staticFS))))`)
	// ============================================
	// USO CON TEMPLATES
	// ============================================
	fmt.Println("\n--- Uso con Templates ---")
	fmt.Println(`
HTML TEMPLATES:

//go:embed templates/*.html
var templateFS embed.FS

func loadTemplates() (*template.Template, error) {
    return template.ParseFS(templateFS, "templates/*.html")
}

// Uso
tmpl, err := loadTemplates()
if err != nil {
    log.Fatal(err)
}

tmpl.ExecuteTemplate(w, "index.html", data)

TEXT TEMPLATES:
//go:embed queries/*.sql
var sqlFS embed.FS

func loadQuery(name string) (string, error) {
    data, err := sqlFS.ReadFile("queries/" + name)
    return string(data), err
}`)
	// ============================================
	// SUB-FILESYSTEMS
	// ============================================
	fmt.Println("\n--- Sub-Filesystems ---")
	fmt.Println(`
fs.Sub crea un sub-filesystem:

//go:embed web/static
var webFS embed.FS

// Acceso normal: webFS.ReadFile("web/static/style.css")

// Con Sub: eliminar prefijo
subFS, err := fs.Sub(webFS, "web/static")
// Ahora: subFS.ReadFile("style.css")

ÚTIL PARA:
- Servir archivos sin prefijo de directorio
- Crear múltiples vistas del mismo embed
- Simplificar paths`)
	// ============================================
	// WALKING EMBED.FS
	// ============================================
	fmt.Println("\n--- Walking embed.FS ---")
	fmt.Println(`
LISTAR ARCHIVOS:

//go:embed assets
var assetsFS embed.FS

func listFiles() error {
    return fs.WalkDir(assetsFS, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() {
            fmt.Println(path)
        }
        return nil
    })
}

BUSCAR ARCHIVOS POR EXTENSIÓN:

func findByExt(fsys fs.FS, ext string) ([]string, error) {
    var files []string
    err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() && strings.HasSuffix(path, ext) {
            files = append(files, path)
        }
        return nil
    })
    return files, err
}`)
	// ============================================
	// CASOS DE USO COMUNES
	// ============================================
	fmt.Println("\n--- Casos de Uso Comunes ---")
	fmt.Println(`
1. SINGLE BINARY DEPLOYMENT:
//go:embed migrations/*.sql
var migrations embed.FS

//go:embed config/defaults.yaml
var defaultConfig []byte

2. WEB SERVER CON ASSETS:
//go:embed dist/*
var frontendBuild embed.FS

3. CLI CON TEMPLATES:
//go:embed templates/*.tmpl
var codeTemplates embed.FS

4. VERSIÓN DEL BUILD:
//go:embed VERSION
var version string

5. CERTIFICADOS/KEYS (desarrollo):
//go:embed certs/dev.crt certs/dev.key
var devCerts embed.FS

6. DATOS DE PRUEBA:
//go:embed testdata/*.json
var testFixtures embed.FS

7. DOCUMENTACIÓN EMBEBIDA:
//go:embed docs/*.md
var documentation embed.FS

8. TRADUCCIONES/I18N:
//go:embed locales/*.json
var translations embed.FS`)
	// ============================================
	// DESARROLLO VS PRODUCCIÓN
	// ============================================
	fmt.Println("\n--- Desarrollo vs Producción ---")
	fmt.Println(`
PATRÓN: Hot reload en desarrollo, embed en producción

// embed_prod.go
//go:build !dev

//go:embed static
var staticFS embed.FS

func getStaticFS() fs.FS {
    return staticFS
}

// embed_dev.go
//go:build dev

func getStaticFS() fs.FS {
    return os.DirFS("static")  // Lee del disco
}

BUILD:
go build                    # Usa embed
go build -tags dev          # Lee del disco (hot reload)`)
	// ============================================
	// ERRORES COMUNES
	// ============================================
	fmt.Println("\n--- Errores Comunes ---")
	fmt.Println(`
1. LÍNEA EN BLANCO:
❌ //go:embed file.txt

   var data string

✓  //go:embed file.txt
   var data string

2. VARIABLE LOCAL:
❌ func foo() {
       //go:embed file.txt
       var data string  // Error: solo nivel de paquete
   }

3. ARCHIVO NO EXISTE:
❌ //go:embed noexiste.txt  // Error en compile time

4. DIRECTORIO VACÍO:
❌ //go:embed emptydir/*    // Error si está vacío

5. PATH ABSOLUTO:
❌ //go:embed /etc/config   // Error: debe ser relativo

6. ESCAPAR DEL MÓDULO:
❌ //go:embed ../otrodir    // Error: no puede salir del módulo`)
	// ============================================
	// SEGURIDAD
	// ============================================
	fmt.Println("\n--- Consideraciones de Seguridad ---")
	fmt.Println(`
⚠️  CUIDADO CON LO QUE EMBEDES:

1. NO embeber secretos en producción:
   - API keys
   - Passwords
   - Private keys

2. El binario es legible:
   - strings mybinary | grep "secret"
   - Cualquiera puede extraer archivos embebidos

3. OCULTAR no es PROTEGER:
   - Embed no es encriptación
   - Solo conveniencia de distribución

ALTERNATIVAS PARA SECRETOS:
- Variables de entorno
- Vault/Secret Manager
- Encrypted config files (decrypt at runtime)`)
	// Demostración con mock data
	demonstrateEmbed()
}

// Mock data para demostración
var mockVersion = "1.0.0"
var mockConfig = []byte(`{"app": "demo", "port": 8080}`)

func demonstrateEmbed() {
	fmt.Println("\n--- Demostración ---")

	// Simular embed.FS con estructura en memoria
	fmt.Println("Estructura típica de proyecto:")
	fmt.Println(`
myapp/
├── main.go
├── embed.go          # Declaraciones //go:embed
├── static/
│   ├── css/
│   │   └── style.css
│   ├── js/
│   │   └── app.js
│   └── index.html
├── templates/
│   ├── base.html
│   └── pages/
│       └── home.html
├── migrations/
│   ├── 001_init.sql
│   └── 002_users.sql
└── VERSION`)}

/*
RESUMEN DE GO:EMBED:

SINTAXIS:
//go:embed file.txt
var content string        // Un archivo como string

//go:embed file.bin
var data []byte           // Un archivo como bytes

//go:embed dir/* *.html
var fs embed.FS           // Filesystem embebido

MÉTODOS DE embed.FS:
- Open(name) (fs.File, error)
- ReadDir(name) ([]DirEntry, error)
- ReadFile(name) ([]byte, error)

PATRONES:
- archivo.txt          → archivo específico
- dir/*                → contenido del directorio
- dir                  → directorio recursivo
- *.html               → glob pattern
- all:dir              → incluir archivos ocultos

INTEGRACIÓN:
- http.FileServer(http.FS(embedFS))
- template.ParseFS(embedFS, pattern)
- fs.Sub(embedFS, "subdir")
- fs.WalkDir(embedFS, ".", walkFn)

CASOS DE USO:
1. Single binary deployment
2. Web server con assets
3. Templates embebidos
4. Migrations SQL
5. Configuración por defecto
6. Documentación

REGLAS:
1. Solo variables de paquete
2. No líneas en blanco entre directiva y var
3. Paths relativos al archivo .go
4. No puede escapar del módulo
5. Archivos deben existir en compile time

SEGURIDAD:
- No embeber secretos
- El contenido es extraíble del binario
- Usar para conveniencia, no protección
*/

/* SUMMARY - CHAPTER 063: Embed Files
Topics covered in this chapter:

• go:embed directive: embedding files and directories into Go binaries
• Embed types: string, []byte, embed.FS for different use cases
• Pattern syntax: specific files, directories, glob patterns, all: prefix for hidden files
• embed.FS methods: Open, ReadDir, ReadFile for accessing embedded content
• HTTP integration: serving embedded files with http.FileServer(http.FS(embedFS))
• Template integration: using template.ParseFS for embedded templates
• Filesystem operations: fs.Sub for subdirectories, fs.WalkDir for traversal
• Use cases: single binary deployment, web assets, templates, SQL migrations, default configs
• Rules and constraints: package-level vars only, no blank lines, relative paths, compile-time only
• Security considerations: embedded content is extractable, not for secrets
• Best practices: organizing embedded files, subdirectory structure, documentation
*/
