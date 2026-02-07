// Package main - Chapter 125: Go Proverbs and Common Mistakes
// Rob Pike's Go Proverbs distill the philosophy of Go.
// This chapter covers each proverb with practical examples,
// plus the most common mistakes Go developers make.
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

func main() {
	fmt.Println("=== GO PROVERBS AND COMMON MISTAKES ===")

	// ============================================
	// GO PROVERBS (Rob Pike, 2015)
	// https://go-proverbs.github.io
	// ============================================
	fmt.Println("\n==============================")
	fmt.Println("PART 1: GO PROVERBS")
	fmt.Println("==============================")

	// ============================================
	// 1. Don't communicate by sharing memory,
	//    share memory by communicating.
	// ============================================
	fmt.Println("\n--- Proverb 1: Share Memory by Communicating ---")

	demonstrateShareByCommunicating()

	fmt.Println(`
// MAL: compartir memoria con mutex
type CounterMutex struct {
    mu    sync.Mutex
    count int
}
func (c *CounterMutex) Inc() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}

// BIEN: comunicar via channels (cuando la logica lo amerita)
func counter(inc <-chan struct{}) <-chan int {
    ch := make(chan int)
    go func() {
        count := 0
        for range inc {
            count++
        }
        ch <- count
        close(ch)
    }()
    return ch
}

CUANDO USAR CADA UNO:
- Channels: flujo de datos, pipelines, fan-out/fan-in
- Mutex: proteger estado interno simple (cache, config)
No fuerces channels donde un mutex es mas simple.`)

	// ============================================
	// 2. Concurrency is not parallelism.
	// ============================================
	fmt.Println("\n--- Proverb 2: Concurrency is Not Parallelism ---")
	fmt.Println(`
CONCURRENCIA: estructurar un programa como composicion
de tareas independientes. Es un patron de DISENO.

PARALELISMO: ejecutar multiples tareas simultaneamente.
Es una propiedad de la EJECUCION.

EJEMPLO:
  Un web server que maneja 1000 requests con goroutines
  es CONCURRENTE. Si corre en 1 CPU, NO es paralelo
  pero sigue funcionando correctamente.

  Con GOMAXPROCS > 1, las goroutines pueden ejecutarse
  en paralelo, pero el programa no cambia.

IMPLICACION PRACTICA:
  - Disenla para concurrencia (goroutines, channels)
  - El paralelismo viene gratis del runtime/hardware
  - No asumas orden de ejecucion entre goroutines`)

	// ============================================
	// 3. Channels orchestrate; mutexes serialize.
	// ============================================
	fmt.Println("\n--- Proverb 3: Channels Orchestrate; Mutexes Serialize ---")
	fmt.Println(`
CHANNELS para orquestar flujo de trabajo:
  results := make(chan Result)
  for _, url := range urls {
      go func(u string) {
          results <- fetch(u)
      }(url)
  }

MUTEX para proteger datos compartidos:
  var mu sync.Mutex
  var cache = make(map[string]string)
  mu.Lock()
  cache[key] = value
  mu.Unlock()

REGLA: Si estas enviando datos entre goroutines -> channel
       Si estas protegiendo acceso a un recurso -> mutex`)

	// ============================================
	// 4. The bigger the interface, the weaker the abstraction.
	// ============================================
	fmt.Println("\n--- Proverb 4: The Bigger the Interface, the Weaker the Abstraction ---")

	demonstrateSmallInterfaces()

	fmt.Println(`
// MAL: interface gigante
type Repository interface {
    Create(ctx context.Context, user User) error
    Get(ctx context.Context, id string) (User, error)
    Update(ctx context.Context, user User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter Filter) ([]User, error)
    Count(ctx context.Context) (int, error)
    Search(ctx context.Context, q string) ([]User, error)
    Export(ctx context.Context, format string) ([]byte, error)
}
// Dificil de implementar, testear, y mockear

// BIEN: interfaces pequenas y composicion
type Reader interface {
    Get(ctx context.Context, id string) (User, error)
}
type Writer interface {
    Create(ctx context.Context, user User) error
}
type ReadWriter interface {
    Reader
    Writer
}

STDLIB EJEMPLOS:
  io.Reader     // 1 metodo: Read(p []byte) (n int, err error)
  io.Writer     // 1 metodo: Write(p []byte) (n int, err error)
  fmt.Stringer  // 1 metodo: String() string
  error         // 1 metodo: Error() string`)

	// ============================================
	// 5. Make the zero value useful.
	// ============================================
	fmt.Println("\n--- Proverb 5: Make the Zero Value Useful ---")

	demonstrateZeroValue()

	fmt.Println(`
STDLIB EJEMPLOS:
  var buf bytes.Buffer    // Listo para usar, no necesita init
  var mu sync.Mutex       // Listo para Lock/Unlock
  var wg sync.WaitGroup   // Listo para Add/Done/Wait
  var once sync.Once      // Listo para Do

// MAL: requiere constructor
type Cache struct {
    data map[string]string  // Zero value es nil -> panic en write
}

// BIEN: zero value funciona
type Cache struct {
    mu   sync.Mutex
    data map[string]string
}
func (c *Cache) Set(k, v string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.data == nil {
        c.data = make(map[string]string)
    }
    c.data[k] = v
}`)

	// ============================================
	// 6. interface{} says nothing.
	// ============================================
	fmt.Println("\n--- Proverb 6: interface{} (any) Says Nothing ---")
	fmt.Println(`
// MAL: pierde type safety
func Process(data any) any {
    // Que tipo es? Que puedo hacer con el?
    // Necesito type assertions en todas partes
}

// BIEN: tipos concretos o interfaces especificas
func Process(data []byte) (Result, error) {
    // Claro: recibe bytes, retorna Result
}

// BIEN: interface con contrato
func Process(r io.Reader) (Result, error) {
    // Claro: lee de cualquier reader
}

CUANDO any ES ACEPTABLE:
- Funciones de logging/debugging (fmt.Println)
- Serialization generica (json.Marshal)
- Contenedores genericos (pre-generics, ahora usar generics)

Con Go 1.18+ generics:
  // ANTES
  func Contains(slice []any, item any) bool
  // AHORA
  func Contains[T comparable](slice []T, item T) bool`)

	// ============================================
	// 7. Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.
	// ============================================
	fmt.Println("\n--- Proverb 7: Gofmt's Style ---")
	fmt.Println(`
PRINCIPIO: eliminar debates sobre formato.

  gofmt -w .           // Formatear todos los archivos
  goimports -w .       // gofmt + organizar imports

CONFIGURACION CI:
  # Verificar que todo esta formateado
  test -z "$(gofmt -l .)"

HERRAMIENTAS MODERNAS:
  golangci-lint run    // Lint comprehensivo
  go vet ./...         // Analisis estatico built-in
  staticcheck ./...    // Checks avanzados

REGLA: No configures estilo. Usa las herramientas.
Tabs, no spaces. Bracket en misma linea. Fin del debate.`)

	// ============================================
	// 8. A little copying is better than a little dependency.
	// ============================================
	fmt.Println("\n--- Proverb 8: A Little Copying is Better than a Little Dependency ---")
	fmt.Println(`
// MAL: importar un package enorme por una funcion
import "github.com/big-lib/utils"
utils.MinInt(a, b)

// BIEN: copiar la funcion trivial
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
// (Ahora min/max son builtins desde Go 1.21)

EVALUAR DEPENDENCIAS:
- Es trivial reimplementar? -> Copiar
- Es complejo y bien mantenido? -> Importar
- Cuantas dependencias transitivas trae?
- Esta activamente mantenido?
- Es de un autor/org confiable?

HERRAMIENTAS:
  go mod graph           // Ver arbol de dependencias
  go mod why <module>    // Por que necesito este modulo?
  go mod tidy            // Limpiar dependencias no usadas`)

	// ============================================
	// 9. Errors are values.
	// ============================================
	fmt.Println("\n--- Proverb 9: Errors Are Values ---")

	demonstrateErrorsAreValues()

	fmt.Println(`
LOS ERRORES SON VALORES, NO EXCEPCIONES:
- Se pueden programar, almacenar, pasar como argumentos
- Se pueden componer, envolver, inspeccionar
- No rompen el flujo de control como exceptions

// Error como valor permite patrones:
type errWriter struct {
    w   io.Writer
    err error
}
func (ew *errWriter) write(buf []byte) {
    if ew.err != nil { return }  // Skip si ya hubo error
    _, ew.err = ew.w.Write(buf)
}

// Uso: elimina if err != nil repetido
ew := &errWriter{w: os.Stdout}
ew.write(header)
ew.write(body)
ew.write(footer)
if ew.err != nil {
    return ew.err
}

PATTERN: bufio.Scanner usa el mismo principio
  scanner := bufio.NewScanner(file)
  for scanner.Scan() {  // Acumula error internamente
      process(scanner.Text())
  }
  if err := scanner.Err(); err != nil {  // Chequear al final
      log.Fatal(err)
  }`)

	// ============================================
	// 10. Don't just check errors, handle them gracefully.
	// ============================================
	fmt.Println("\n--- Proverb 10: Handle Errors Gracefully ---")
	os.Stdout.WriteString(`
// MAL: solo loggear y continuar
if err != nil {
    log.Println(err)
}
// El programa sigue en estado invalido!

// MAL: panic en error recuperable
if err != nil {
    panic(err)
}

// BIEN: retornar con contexto
if err != nil {
    return fmt.Errorf("fetching user %s: %w", id, err)
}

// BIEN: accion alternativa
data, err := cache.Get(key)
if err != nil {
    data, err = db.Get(key)  // Fallback
    if err != nil {
        return fmt.Errorf("getting data for %s: %w", key, err)
    }
}

// BIEN: degradar gracefully
config, err := loadConfig(path)
if err != nil {
    log.Printf("using defaults: %v", err)
    config = defaultConfig()
}
`)

	// ============================================
	// 11. Don't panic.
	// ============================================
	fmt.Println("\n--- Proverb 11: Don't Panic ---")
	fmt.Println(`
PANIC ES PARA:
- Errores de programacion (bug en TU codigo)
- Invariantes rotas que no deberian ocurrir
- Inicializacion que no puede fallar

PANIC NO ES PARA:
- Input invalido del usuario
- Archivos que no existen
- Conexiones de red que fallan
- Cualquier error esperado/recuperable

// OK: MustCompile (panic si regexp es invalido)
var re = regexp.MustCompile("^[a-z]+$")  // Constante, falla = bug

// OK: init que no puede continuar
func init() {
    if os.Getenv("REQUIRED_VAR") == "" {
        panic("REQUIRED_VAR must be set")
    }
}

// MAL:
func GetUser(id string) User {
    user, err := db.Find(id)
    if err != nil {
        panic(err)  // NO! Retorna error
    }
    return user
}`)

	// ============================================
	// PART 2: COMMON MISTAKES
	// ============================================
	fmt.Println("\n==============================")
	fmt.Println("PART 2: COMMON MISTAKES")
	fmt.Println("==============================")

	// ============================================
	// MISTAKE 1: Nil Map Write
	// ============================================
	fmt.Println("\n--- Mistake 1: Writing to Nil Map ---")

	demonstrateNilMap()

	fmt.Println(`
// PANIC: assignment to entry in nil map
var m map[string]int
m["key"] = 1  // PANIC!

// FIX: siempre inicializar
m := make(map[string]int)
m["key"] = 1  // OK

// NOTA: leer de nil map NO es panic
var m map[string]int
v := m["key"]  // v = 0, ok = false`)

	// ============================================
	// MISTAKE 2: Goroutine Leak
	// ============================================
	fmt.Println("\n--- Mistake 2: Goroutine Leaks ---")
	fmt.Println(`
// LEAK: goroutine bloqueada para siempre
func fetch(url string) {
    ch := make(chan string)
    go func() {
        resp := httpGet(url)
        ch <- resp  // Bloqueado si fetch() retorna antes de leer
    }()
    // Si hay timeout, ch nunca se lee -> goroutine leak
}

// FIX 1: buffered channel
ch := make(chan string, 1)  // No bloquea al escribir

// FIX 2: context cancelable
func fetch(ctx context.Context, url string) (string, error) {
    ch := make(chan string, 1)
    go func() {
        ch <- httpGet(url)
    }()
    select {
    case result := <-ch:
        return result, nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}`)

	// ============================================
	// MISTAKE 3: Slice Gotchas
	// ============================================
	fmt.Println("\n--- Mistake 3: Slice Gotchas ---")

	demonstrateSliceGotchas()

	fmt.Println(`
GOTCHA 1: Slices comparten memoria
  original := []int{1, 2, 3, 4, 5}
  sub := original[1:3]  // [2, 3]
  sub[0] = 99           // original es ahora [1, 99, 3, 4, 5]!

  // FIX: copiar si necesitas independencia
  sub := make([]int, 2)
  copy(sub, original[1:3])

GOTCHA 2: Append puede (o no) crear nuevo array
  s := make([]int, 0, 10)
  s1 := append(s, 1, 2, 3)
  s2 := append(s, 4, 5, 6)
  // s1 y s2 comparten backing array! s1 = [4, 5, 6]

  // FIX: usar full slice expression
  s1 := append(s[:0:0], 1, 2, 3)  // Fuerza copia

GOTCHA 3: Memory leak con subslice
  func getHeader(data []byte) []byte {
      return data[:10]  // Retiene referencia a TODO data
  }
  // FIX:
  func getHeader(data []byte) []byte {
      header := make([]byte, 10)
      copy(header, data[:10])
      return header  // data puede ser GC'd
  }`)

	// ============================================
	// MISTAKE 4: defer in Loop
	// ============================================
	fmt.Println("\n--- Mistake 4: Defer in Loop ---")
	fmt.Println(`
// MAL: defers se acumulan hasta que la funcion retorna
func processFiles(paths []string) error {
    for _, path := range paths {
        f, err := os.Open(path)
        if err != nil { return err }
        defer f.Close()  // No se cierra hasta que processFiles retorna!
        // Si hay 10000 archivos -> 10000 file descriptors abiertos
    }
    return nil
}

// FIX: extraer a funcion
func processFiles(paths []string) error {
    for _, path := range paths {
        if err := processFile(path); err != nil {
            return err
        }
    }
    return nil
}

func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil { return err }
    defer f.Close()  // Se cierra al retornar de processFile
    // ...
    return nil
}`)

	// ============================================
	// MISTAKE 5: Pointer to Loop Variable
	// ============================================
	fmt.Println("\n--- Mistake 5: Closure Over Loop Variable (pre Go 1.22) ---")
	fmt.Println(`
// BUG (pre Go 1.22): todas las closures ven el mismo valor
funcs := make([]func(), 0)
for _, v := range []string{"a", "b", "c"} {
    funcs = append(funcs, func() { fmt.Println(v) })
}
for _, f := range funcs { f() }
// Pre 1.22: imprime "c", "c", "c"
// Go 1.22+: imprime "a", "b", "c" (cada iteracion nueva variable)

// FIX universal (funciona en todas las versiones):
for _, v := range values {
    v := v  // Shadow con copia local
    go func() { use(v) }()
}

Go 1.22+ CAMBIO:
  Las variables de loop tienen scope PER-ITERATION.
  Ya no necesitas el shadow trick.
  Pero es bueno conocerlo por codigo legacy.`)

	// ============================================
	// MISTAKE 6: Not Handling HTTP Body
	// ============================================
	fmt.Println("\n--- Mistake 6: HTTP Response Body ---")
	fmt.Println(`
// MAL: body leak
resp, err := http.Get(url)
if err != nil { return err }
// Olvido resp.Body.Close() -> connection leak

// MAL: close antes de check error
resp, err := http.Get(url)
defer resp.Body.Close()  // PANIC si err != nil (resp es nil)

// BIEN:
resp, err := http.Get(url)
if err != nil { return err }
defer resp.Body.Close()

// BIEN: drenar body para reutilizar conexion
defer func() {
    io.Copy(io.Discard, resp.Body)
    resp.Body.Close()
}()

// Si no necesitas el body, igual hay que drenarlo
// para que la conexion vuelva al pool del transport.`)

	// ============================================
	// MISTAKE 7: String Concatenation in Loop
	// ============================================
	fmt.Println("\n--- Mistake 7: String Concatenation ---")

	demonstrateStringConcat()

	fmt.Println(`
// MAL: O(n^2) - cada + crea un nuevo string
result := ""
for _, s := range items {
    result += s  // Copia todo el string cada vez
}

// BIEN: O(n) - strings.Builder
var sb strings.Builder
sb.Grow(estimatedSize)  // Pre-allocate si sabes el tamano
for _, s := range items {
    sb.WriteString(s)
}
result := sb.String()

// TAMBIEN BIEN para join simple:
result := strings.Join(items, "")`)

	// ============================================
	// MISTAKE 8: Embedding Mutex
	// ============================================
	fmt.Println("\n--- Mistake 8: Embedding sync.Mutex ---")
	fmt.Println(`
// MAL: mutex embebido expone Lock/Unlock publicamente
type Cache struct {
    sync.Mutex  // Cache.Lock() y Cache.Unlock() son publicos!
    data map[string]string
}

// BIEN: mutex como campo privado
type Cache struct {
    mu   sync.Mutex  // No exportado
    data map[string]string
}

// La API del Cache no deberia exponer detalles de sincronizacion.
// Los usuarios del Cache no deberian hacer cache.Lock().`)

	// ============================================
	// MISTAKE 9: Using time.After in Loop
	// ============================================
	fmt.Println("\n--- Mistake 9: time.After in Select Loop ---")
	fmt.Println(`
// MAL: memory leak - time.After crea un timer cada iteracion
for {
    select {
    case msg := <-ch:
        process(msg)
    case <-time.After(5 * time.Second):  // LEAK: timer nunca se libera
        return
    }
}

// BIEN: reusar timer
timer := time.NewTimer(5 * time.Second)
defer timer.Stop()
for {
    select {
    case msg := <-ch:
        if !timer.Stop() {
            <-timer.C
        }
        timer.Reset(5 * time.Second)
        process(msg)
    case <-timer.C:
        return
    }
}`)

	// ============================================
	// MISTAKE 10: Ignoring Context
	// ============================================
	fmt.Println("\n--- Mistake 10: Ignoring Context ---")
	fmt.Println(`
// MAL: ignorar context
func GetUser(ctx context.Context, id string) (User, error) {
    // ctx nunca se usa -> no respeta cancelation ni timeout
    return db.Query("SELECT * FROM users WHERE id = ?", id)
}

// BIEN: pasar context a operaciones I/O
func GetUser(ctx context.Context, id string) (User, error) {
    row := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", id)
    // ...
}

// BIEN: chequear context en loops largos
func Process(ctx context.Context, items []Item) error {
    for _, item := range items {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        if err := processItem(item); err != nil {
            return err
        }
    }
    return nil
}

REGLA: Si una funcion recibe context.Context,
       debe usarlo o propagarlo.`)

	// ============================================
	// EFFECTIVE GO: KEY PRINCIPLES
	// ============================================
	fmt.Println("\n==============================")
	fmt.Println("PART 3: EFFECTIVE GO PRINCIPLES")
	fmt.Println("==============================")

	os.Stdout.WriteString(`
1. ACCEPT INTERFACES, RETURN STRUCTS
   func New(r io.Reader) *Parser  // Flexible input, concrete output

2. RETURN EARLY
   // MAL:
   if condition {
       // 50 lines
   }
   // BIEN:
   if !condition { return }
   // 50 lines

3. PACKAGE NAMING
   - Corto, lowercase, una palabra: http, fmt, json
   - NO: utilities, common, base, helpers
   - El nombre del package es parte del nombre exportado:
     http.Server (no http.HTTPServer)

4. ERROR WRAPPING
   return fmt.Errorf("creating user %s: %w", name, err)
   // Agrega contexto sin perder el error original

5. TABLE-DRIVEN TESTS
   tests := []struct{
       name  string
       input int
       want  int
   }{
       {"positive", 5, 25},
       {"zero", 0, 0},
       {"negative", -3, 9},
   }
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           got := Square(tt.input)
           if got != tt.want {
               t.Errorf("Square(%d) = %d, want %d", tt.input, got, tt.want)
           }
       })
   }

6. NAMING CONVENTIONS
   - MixedCaps (no underscores)
   - Acronyms en caps: HTTPServer, URL, ID (no HttpServer, Url, Id)
   - Getters sin "Get": user.Name() (no user.GetName())
   - Interfaces de 1 metodo: Reader, Writer, Closer, Stringer
   - Bool: isReady, hasPermission, canRetry

7. COMPOSITION OVER INHERITANCE
   Go no tiene herencia. Usa composicion:
   type Server struct {
       logger *slog.Logger
       store  Store  // Interface, no struct concreto
       config Config
   }

8. KEEP MAIN THIN
   func main() {
       if err := run(os.Args, os.Stdout); err != nil {
           fmt.Fprintf(os.Stderr, "error: %%v\n", err)
           os.Exit(1)
       }
   }
   // Testeable, configurable, limpio
`)

	fmt.Println("\n=== FIN CAPITULO 79: GO PROVERBS AND COMMON MISTAKES ===")
}

// demonstrateShareByCommunicating muestra el proverbio 1
func demonstrateShareByCommunicating() {
	// En vez de compartir datos con mutex, enviar datos por channel
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			for _, n := range nums {
				out <- n
			}
			close(out)
		}()
		return out
	}

	square := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			for n := range in {
				out <- n * n
			}
			close(out)
		}()
		return out
	}

	// Pipeline: gen -> square -> print
	for v := range square(gen(2, 3, 4)) {
		fmt.Printf("%d ", v)
	}
	fmt.Println()
}

// demonstrateSmallInterfaces muestra el proverbio 4
func demonstrateSmallInterfaces() {
	// io.Reader es una interface de 1 solo metodo
	// y es una de las mas poderosas de Go
	r := strings.NewReader("Hello from io.Reader")
	buf := make([]byte, 20)
	n, _ := r.Read(buf)
	fmt.Printf("Read %d bytes: %s\n", n, buf[:n])
}

// demonstrateZeroValue muestra el proverbio 5
func demonstrateZeroValue() {
	// sync.Mutex funciona sin inicializacion
	var mu sync.Mutex
	mu.Lock()
	fmt.Println("Mutex zero value funciona directamente")
	mu.Unlock()

	// sync.WaitGroup funciona sin inicializacion
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("WaitGroup zero value funciona directamente")
	}()
	wg.Wait()
}

// demonstrateNilMap muestra el error de nil map
func demonstrateNilMap() {
	// Leer de nil map es safe
	var m map[string]int
	v := m["key"]
	fmt.Printf("Leer nil map: %d (zero value, sin panic)\n", v)

	// Escribir requiere make
	m = make(map[string]int)
	m["key"] = 42
	fmt.Printf("Escribir map inicializado: %d\n", m["key"])
}

// demonstrateSliceGotchas muestra gotchas de slices
func demonstrateSliceGotchas() {
	// Slices comparten backing array
	original := []int{1, 2, 3, 4, 5}
	sub := original[1:3]
	fmt.Printf("Original: %v, Sub: %v\n", original, sub)

	sub[0] = 99
	fmt.Printf("Despues de modificar sub: Original=%v Sub=%v\n", original, sub)

	// Fix: copiar
	original2 := []int{1, 2, 3, 4, 5}
	safe := make([]int, 2)
	copy(safe, original2[1:3])
	safe[0] = 99
	fmt.Printf("Con copy: Original=%v Safe=%v (independientes)\n", original2, safe)
}

// demonstrateErrorsAreValues muestra errores como valores programables
func demonstrateErrorsAreValues() {
	// errWriter acumula errores
	ew := &errWriter{w: os.Stdout}
	ew.write([]byte("Hello "))
	ew.write([]byte("World"))
	ew.write([]byte("\n"))
	if ew.err != nil {
		fmt.Printf("Error: %v\n", ew.err)
	}
}

type errWriter struct {
	w   *os.File
	err error
}

func (ew *errWriter) write(buf []byte) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.w.Write(buf)
}

/*
SUMMARY - GO PROVERBS AND COMMON MISTAKES:

GO PROVERBS (ROB PIKE):
- Share memory by communicating (channels for data flow, mutex for state)
- Concurrency is not parallelism (design vs execution property)
- Channels orchestrate; mutexes serialize
- The bigger the interface, the weaker the abstraction (prefer 1-method interfaces)
- Make the zero value useful (bytes.Buffer, sync.Mutex work without init)
- interface{}/any says nothing (prefer specific types or interfaces)
- Gofmt's style is no one's favorite, yet everyone's favorite
- A little copying is better than a little dependency
- Errors are values (programmable, composable, inspectable)
- Don't just check errors, handle them gracefully (wrap with context)
- Don't panic (only for programmer bugs, not expected errors)

COMMON MISTAKES:
- Writing to nil map (always initialize with make)
- Goroutine leaks (use buffered channels or context cancellation)
- Slice shared backing array (copy for independence)
- Defer in loop (extract to separate function)
- Closure over loop variable (fixed in Go 1.22, shadow trick for older)
- HTTP body not closed (always defer resp.Body.Close after error check)
- String concatenation in loop (use strings.Builder instead)
- Embedding sync.Mutex (keep as private field, don't export Lock/Unlock)
- time.After in select loop (reuse timer to prevent memory leak)
- Ignoring context (always propagate to I/O operations)

EFFECTIVE GO PRINCIPLES:
- Accept interfaces, return structs
- Return early to reduce nesting
- Short lowercase package names (no utilities/common/helpers)
- Table-driven tests for comprehensive coverage
- Composition over inheritance
- Keep main thin (delegate to testable run function)
*/

// demonstrateStringConcat muestra el rendimiento de strings.Builder
func demonstrateStringConcat() {
	items := []string{"Go", " ", "is", " ", "awesome"}

	// Eficiente
	var sb strings.Builder
	for _, s := range items {
		sb.WriteString(s)
	}
	fmt.Printf("Builder: %s\n", sb.String())

	// Mas simple para join
	fmt.Printf("Join: %s\n", strings.Join(items, ""))
}
