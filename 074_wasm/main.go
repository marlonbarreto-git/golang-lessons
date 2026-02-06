// Package main - Chapter 074: WebAssembly
// WebAssembly targets (browser via syscall/js, WASI via wasip1),
// building and running WASM modules, DOM manipulation, Go-JS interop,
// TinyGo for smaller binaries, wazero for embedding, limitations
// and performance considerations.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== GO AND WEBASSEMBLY (WASM) ===")

	// ============================================
	// 1. WHAT IS WEBASSEMBLY
	// ============================================
	fmt.Println("\n--- 1. What is WebAssembly ---")
	os.Stdout.WriteString(`
WEBASSEMBLY (WASM):

WebAssembly es un formato de instrucciones binarias para una maquina
virtual basada en stack. Disenado como target de compilacion portable
para lenguajes como C, C++, Rust y Go.

CARACTERISTICAS:
  - Binario compacto y eficiente
  - Ejecucion casi nativa en el navegador
  - Sandbox seguro (sin acceso directo al OS)
  - Complementa JavaScript, no lo reemplaza
  - Soportado por todos los navegadores modernos

POR QUE GO SOPORTA WASM:
  - Permite reutilizar logica de negocio Go en el browser
  - Computacion intensiva (crypto, compresion, parseo) mas rapida que JS
  - Compartir codigo entre backend Go y frontend WASM
  - Aplicaciones edge/serverless con sandboxing
  - Plugins extensibles en formato portable

DOS TARGETS EN GO:

  1. GOOS=js GOARCH=wasm
     - Target para navegadores web
     - Usa syscall/js para interactuar con JavaScript
     - Requiere wasm_exec.js como glue code
     - Binarios grandes (~2MB+ minimo)

  2. GOOS=wasip1 GOARCH=wasm (Go 1.21+)
     - Target WASI (WebAssembly System Interface)
     - Acceso a filesystem, stdin/stdout, args, env vars
     - Se ejecuta fuera del navegador (wasmtime, wasmer, wazero)
     - Portable entre runtimes WASI-compatible
     - No requiere JavaScript

HISTORIA:
  Go 1.11: Soporte experimental WASM (GOOS=js)
  Go 1.13: Mejoras de estabilidad
  Go 1.21: Soporte WASI (GOOS=wasip1)
  Go 1.24: Mejoras de performance y reduccion de binarios
`)

	// ============================================
	// 2. BUILDING FOR THE BROWSER
	// ============================================
	fmt.Println("\n--- 2. Building for the Browser (GOOS=js GOARCH=wasm) ---")
	os.Stdout.WriteString(`
COMPILAR GO PARA EL NAVEGADOR:

  GOOS=js GOARCH=wasm go build -o main.wasm ./cmd/app

El resultado es un archivo .wasm que el navegador puede ejecutar.

ARCHIVOS NECESARIOS:

1. wasm_exec.js - Glue code proporcionado por Go:
   cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .

2. wasm_exec_node.js - Para ejecutar en Node.js:
   cp "$(go env GOROOT)/misc/wasm/wasm_exec_node.js" .

3. HTML loader - Pagina que carga el WASM:

<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Go WASM App</title>
</head>
<body>
  <h1>Go WebAssembly</h1>
  <div id="app"></div>

  <script src="wasm_exec.js"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(
      fetch("main.wasm"), go.importObject
    ).then((result) => {
      go.run(result.instance);
    });
  </script>
</body>
</html>

SERVIR LOCALMENTE:

  # Necesitas un servidor HTTP (WASM requiere MIME type correcto)
  # Opcion 1: goexec
  goexec 'http.ListenAndServe(":8080", http.FileServer(http.Dir(".")))'

  # Opcion 2: Python
  python3 -m http.server 8080

  # Opcion 3: servidor Go custom con MIME type
  http.HandleFunc("/main.wasm", func(w http.ResponseWriter, r *http.Request) {
      w.Header().Set("Content-Type", "application/wasm")
      http.ServeFile(w, r, "main.wasm")
  })

EJECUTAR EN NODE.JS (sin navegador):
  GOOS=js GOARCH=wasm go build -o main.wasm .
  node wasm_exec_node.js main.wasm

NOTA SOBRE TAMANO:
  Un "Hello World" en Go WASM pesa ~2MB porque incluye el runtime
  completo de Go (garbage collector, goroutine scheduler, etc.).
  Ver seccion de TinyGo para binarios mas pequenos.
`)

	// ============================================
	// 3. SYSCALL/JS PACKAGE
	// ============================================
	fmt.Println("\n--- 3. syscall/js Package ---")
	os.Stdout.WriteString(`
SYSCALL/JS - INTEROPERABILIDAD GO <-> JAVASCRIPT:

Este paquete solo esta disponible cuando compilas con GOOS=js GOARCH=wasm.
Permite manipular el DOM, llamar funciones JS y exponer funciones Go.

TIPOS PRINCIPALES:

  js.Value    - Representa cualquier valor JavaScript
  js.Func     - Wrapper de funcion Go callable desde JS
  js.Global() - Acceso al objeto global (window en browser)
  js.Null()   - Valor null de JavaScript
  js.Undefined() - Valor undefined de JavaScript

OBTENER VALORES GLOBALES:

  // Equivalente a: window.document
  document := js.Global().Get("document")

  // Equivalente a: window.console
  console := js.Global().Get("console")

  // Equivalente a: window.location.href
  href := js.Global().Get("location").Get("href").String()

MANIPULAR EL DOM:

  // document.getElementById("app")
  app := document.Call("getElementById", "app")

  // app.innerHTML = "<h1>Hello from Go!</h1>"
  app.Set("innerHTML", "<h1>Hello from Go!</h1>")

  // Crear elemento
  div := document.Call("createElement", "div")
  div.Set("textContent", "Created by Go WASM")
  div.Get("style").Set("color", "blue")
  app.Call("appendChild", div)

  // Leer atributos
  className := div.Get("className").String()
  width := div.Get("offsetWidth").Int()

EVENTOS (js.FuncOf para callbacks):

  button := document.Call("getElementById", "myButton")
  callback := js.FuncOf(func(this js.Value, args []js.Value) any {
      event := args[0]
      fmt.Println("Button clicked!", event.Get("type").String())
      return nil
  })
  button.Call("addEventListener", "click", callback)
  // IMPORTANTE: llamar callback.Release() cuando ya no se necesite

LLAMAR FUNCIONES JAVASCRIPT DESDE GO:

  // console.log("Hello from Go!")
  js.Global().Get("console").Call("log", "Hello from Go!")

  // JSON.stringify(obj)
  jsonStr := js.Global().Get("JSON").Call("stringify", myObj).String()

  // setTimeout
  js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) any {
      fmt.Println("Timer fired!")
      return nil
  }), 1000)

EXPONER FUNCIONES GO EN JAVASCRIPT (Calling Go from JS):

  // Registrar funcion Go como global JS
  js.Global().Set("goAdd", js.FuncOf(func(this js.Value, args []js.Value) any {
      a := args[0].Int()
      b := args[1].Int()
      return a + b
  }))

  // Ahora en JavaScript puedes llamar:
  //   const result = goAdd(3, 4);  // returns 7

  // Registrar objeto con multiples funciones
  goModule := js.Global().Get("Object").New()
  goModule.Set("add", js.FuncOf(addFunc))
  goModule.Set("multiply", js.FuncOf(multiplyFunc))
  js.Global().Set("GoMath", goModule)

  // En JavaScript:
  //   GoMath.add(1, 2)
  //   GoMath.multiply(3, 4)

CONVERSIONES DE TIPOS:

  Go -> JS:
    int, float64        -> js.ValueOf(42)
    string              -> js.ValueOf("hello")
    bool                -> js.ValueOf(true)
    nil                 -> js.Null()
    map[string]any      -> js.ValueOf(map[string]any{"key": "value"})
    []any               -> js.ValueOf([]any{1, 2, 3})

  JS -> Go:
    value.Int()         -> int
    value.Float()       -> float64
    value.String()      -> string
    value.Bool()        -> bool
    value.IsNull()      -> bool
    value.IsUndefined() -> bool
    value.Length()       -> int (para arrays)
    value.Index(i)      -> js.Value (elemento de array)

MANTENER EL PROGRAMA VIVO:

  func main() {
      js.Global().Set("myFunc", js.FuncOf(myHandler))

      // Bloquear main para que el programa no termine
      // Opcion 1: channel bloqueante
      <-make(chan struct{})

      // Opcion 2: select vacio
      select {}
  }

FETCH API DESDE GO (patron con Promises):

  func fetchJSON(url string) (string, error) {
      ch := make(chan string, 1)
      errCh := make(chan error, 1)

      promise := js.Global().Call("fetch", url)

      thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
          return args[0].Call("json")
      })

      resolveFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
          result := js.Global().Get("JSON").Call("stringify", args[0])
          ch <- result.String()
          return nil
      })

      catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
          errCh <- fmt.Errorf("fetch error: %%s", args[0].Get("message").String())
          return nil
      })

      promise.Call("then", thenFunc).Call("then", resolveFunc).Call("catch", catchFunc)

      select {
      case result := <-ch:
          return result, nil
      case err := <-errCh:
          return "", err
      }
  }
`)

	// ============================================
	// 4. BUILDING FOR WASI
	// ============================================
	fmt.Println("\n--- 4. Building for WASI (GOOS=wasip1 GOARCH=wasm) ---")
	os.Stdout.WriteString(`
WASI (WebAssembly System Interface):

WASI define una interfaz estandar para que WASM interactue con el
sistema operativo: filesystem, stdin/stdout, args, env vars, clocks.

Go 1.21+ soporta WASI Preview 1 (wasip1).

COMPILAR:
  GOOS=wasip1 GOARCH=wasm go build -o main.wasm .

EJECUTAR CON DISTINTOS RUNTIMES:

  # wasmtime (Bytecode Alliance)
  wasmtime main.wasm
  wasmtime --dir=. main.wasm          # Acceso al directorio actual
  wasmtime --env KEY=VALUE main.wasm  # Variables de entorno

  # wasmer (Wasmer Inc)
  wasmer main.wasm
  wasmer --dir=. main.wasm

  # wazero (pure Go, ver seccion dedicada)
  wazero run main.wasm
  wazero run -mount=.:/ main.wasm

  # Node.js 20+ (soporte WASI experimental)
  node --experimental-wasi-unstable-preview1 run.mjs

DIFERENCIAS CON GOOS=js:

  GOOS=js:
  - Requiere JavaScript runtime (browser o Node.js)
  - Acceso al DOM y APIs del navegador
  - Necesita wasm_exec.js
  - Interoperabilidad bidireccional con JS

  GOOS=wasip1:
  - No necesita JavaScript
  - Acceso a filesystem, stdin/stdout, env
  - Portable entre runtimes WASI
  - Mas parecido a un programa "normal"
  - Mejor para CLI tools, serverless, plugins

QUE FUNCIONA EN WASIP1:
  - fmt.Println, os.Stdout, os.Stderr
  - os.Args, os.Getenv
  - Lectura/escritura de archivos (con permisos del runtime)
  - time.Now (clock_time_get)
  - Goroutines (limitado, single-threaded)
  - La mayoria de la stdlib que no requiere networking

QUE NO FUNCIONA EN WASIP1:
  - net/* (WASI Preview 1 no tiene sockets)
  - os/exec (no puede lanzar procesos)
  - os/signal (no hay senales en WASM)
  - CGO (no hay soporte C en WASM)
  - Plugin loading (no hay dlopen)

EJEMPLO COMPLETO WASI:

  package main

  import (
      "fmt"
      "os"
  )

  func main() {
      fmt.Println("Hello from WASI!")
      fmt.Println("Args:", os.Args)

      if val, ok := os.LookupEnv("MY_VAR"); ok {
          fmt.Println("MY_VAR:", val)
      }

      // Escribir archivo (requiere --dir flag en runtime)
      os.WriteFile("output.txt", []byte("Written from WASM!"), 0644)
  }

  // Build & run:
  // GOOS=wasip1 GOARCH=wasm go build -o main.wasm .
  // wasmtime --dir=. main.wasm --arg1 --arg2
`)

	// ============================================
	// 5. TINYGO FOR SMALLER BINARIES
	// ============================================
	fmt.Println("\n--- 5. TinyGo for Smaller Binaries ---")
	os.Stdout.WriteString(`
TINYGO - COMPILADOR GO ALTERNATIVO PARA WASM:

TinyGo es un compilador Go basado en LLVM optimizado para
targets pequenos: microcontroladores y WebAssembly.

INSTALAR:
  brew install tinygo     # macOS
  # o descargar de https://tinygo.org/getting-started/install/

COMPILAR PARA WASM (browser):
  tinygo build -o main.wasm -target wasm ./cmd/app

COMPILAR PARA WASI:
  tinygo build -o main.wasm -target wasi ./cmd/app

COMPARACION DE TAMANO:

  Programa         | Go standard | TinyGo   | Reduccion
  -----------------+-------------+----------+----------
  Hello World      | ~2.0 MB     | ~250 KB  | 87%%
  JSON parsing     | ~3.5 MB     | ~400 KB  | 89%%
  HTTP handler     | ~5.0 MB     | ~600 KB  | 88%%
  Crypto (sha256)  | ~3.0 MB     | ~350 KB  | 88%%
  Minimal WASI     | ~1.5 MB     | ~50 KB   | 97%%

OPTIMIZACIONES ADICIONALES:

  # Optimizar tamano (-opt=z)
  tinygo build -o main.wasm -target wasm -opt=z -no-debug .

  # Con wasm-opt (binaryen) para optimizar aun mas
  tinygo build -o main.wasm -target wasm -opt=z -no-debug .
  wasm-opt -Oz main.wasm -o main-opt.wasm

  # Resultado tipico: ~30-50KB para programas simples

LIMITACIONES DE TINYGO:
  - reflect parcial (no TypeOf en todos los casos)
  - cgo limitado
  - Algunos paquetes de stdlib no disponibles
  - Goroutines: si, pero scheduler cooperativo
  - GC: mas simple (mark-sweep conservativo)
  - Compilacion mas lenta que Go standard

CUANDO USAR TINYGO vs GO STANDARD:

  TinyGo:
  - Aplicaciones web donde el tamano importa
  - Edge computing con limites de memoria
  - Tiempo de carga critico (mobile, slow networks)
  - Funciones WASM simples (validacion, calculos)

  Go Standard:
  - Aplicaciones complejas que necesitan stdlib completa
  - Cuando el tamano no es critico (serverless, plugins)
  - Necesitas reflect, net/http u otros paquetes avanzados
  - Compatibilidad total garantizada

wasm_exec.js PARA TINYGO:

  TinyGo tiene su PROPIO wasm_exec.js (diferente al de Go):
  cp "$(tinygo env TINYGOROOT)/targets/wasm_exec.js" .

  NO mezclar el wasm_exec.js de Go standard con binarios TinyGo
  ni viceversa. Cada compilador necesita su propia version.

//export EN TINYGO:

  //go:build tinygo

  package main

  //export multiply
  func multiply(a, b int) int {
      return a * b
  }

  func main() {}

  // Compilar: tinygo build -o math.wasm -target wasm .
  // En JS: instance.exports.multiply(3, 4) -> 12
  // TinyGo permite exportar funciones directamente sin syscall/js
`)

	// ============================================
	// 6. WAZERO - PURE GO WASM RUNTIME
	// ============================================
	fmt.Println("\n--- 6. wazero: Pure Go WASM Runtime ---")
	os.Stdout.WriteString(`
WAZERO - RUNTIME WASM ESCRITO EN GO PURO:

wazero es un runtime WebAssembly sin dependencias CGO.
Permite ejecutar modulos WASM desde aplicaciones Go.

INSTALAR COMO CLI:
  go install github.com/tetratelabs/wazero/cmd/wazero@latest

INSTALAR COMO LIBRERIA:
  go get github.com/tetratelabs/wazero

CASO DE USO: EJECUTAR WASM DESDE GO:

  package main

  import (
      "context"
      "fmt"
      "os"

      "github.com/tetratelabs/wazero"
      "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
  )

  func main() {
      ctx := context.Background()

      // Crear runtime
      r := wazero.NewRuntime(ctx)
      defer r.Close(ctx)

      // Instanciar WASI (stdin, stdout, filesystem, etc.)
      wasi_snapshot_preview1.MustInstantiate(ctx, r)

      // Leer modulo WASM
      wasmBytes, _ := os.ReadFile("plugin.wasm")

      // Instanciar y ejecutar
      _, err := r.Instantiate(ctx, wasmBytes)
      if err != nil {
          fmt.Println("Error:", err)
      }
  }

EXPONER FUNCIONES HOST AL MODULO WASM:

  _, err := r.NewHostModuleBuilder("env").
      NewFunctionBuilder().
      WithFunc(func(ctx context.Context, x, y int32) int32 {
          return x + y
      }).
      Export("host_add").
      NewFunctionBuilder().
      WithFunc(func(ctx context.Context, ptr, len uint32) {
          mod := wazero.GetModuleFromContext(ctx)
          bytes, _ := mod.Memory().Read(ptr, len)
          fmt.Println("WASM says:", string(bytes))
      }).
      Export("host_log").
      Instantiate(ctx)

LLAMAR FUNCIONES EXPORTADAS POR EL MODULO WASM:

  mod, _ := r.Instantiate(ctx, wasmBytes)

  results, err := mod.ExportedFunction("calculate").Call(ctx, 10, 20)
  if err != nil {
      fmt.Println("Error:", err)
  }
  fmt.Println("Result:", results[0])

PLUGIN SYSTEM CON WAZERO:

  Paso 1: Escribir plugin en Go (WASI target)

    // plugin/main.go
    package main

    //export calculate
    func calculate(a, b int32) int32 {
        return a*a + b*b
    }

    func main() {}

  Paso 2: Compilar con TinyGo (exporta funciones)
    tinygo build -o plugin.wasm -target wasi -no-debug ./plugin/

  Paso 3: Cargar y ejecutar desde la app host (ver ejemplo anterior)

VENTAJAS DE WAZERO:
  - Zero dependencias CGO (pure Go)
  - Compilacion cruzada sin problemas
  - API Go idiomatica con context.Context
  - Soporte WASI completo
  - AOT compilation para performance
  - Bien mantenido (Tetrate)

ALTERNATIVAS:
  - wasmer-go: bindings a wasmer (requiere CGO)
  - wasmtime-go: bindings a wasmtime (requiere CGO)
  - wazero es preferido en Go por ser pure Go
`)

	// ============================================
	// 7. USE CASES
	// ============================================
	fmt.Println("\n--- 7. Use Cases ---")
	fmt.Println(`
CASOS DE USO REALES DE GO + WASM:

1. BROWSER APPLICATIONS:
   - Editores de texto/codigo con Go logic
   - Juegos 2D/3D con Go logic + Canvas/WebGL
   - Herramientas de procesamiento de imagenes client-side
   - Dashboards con computacion local intensiva

2. EDGE COMPUTING / SERVERLESS:
   - Cloudflare Workers (WASM support)
   - Fastly Compute (WASM nativo)
   - Fermyon Spin (WASI-based serverless)
   - Funciones con cold start sub-millisecond

3. PLUGIN SYSTEMS:
   - Envoy proxy: WASM filters para traffic management
   - OPA (Open Policy Agent): policies en WASM
   - Dapr: pluggable components
   - Extism: framework universal de plugins WASM

4. SANDBOXED EXECUTION:
   - Ejecutar codigo de terceros de forma segura
   - Validaciones custom configurables por tenant
   - Transformaciones de datos user-defined
   - Funciones serverless multi-tenant

5. PORTABILIDAD:
   - Compilar una vez, ejecutar en cualquier runtime
   - Distribuir binarios sin preocuparse por OS/arch
   - CI/CD tools portables

6. COMPARTIR LOGICA FRONTEND/BACKEND:
   - Validaciones ejecutadas en browser Y servidor
   - Parseo de formatos custom
   - Algoritmos de negocio compartidos
   - Calculos financieros/cientificos`)

	// ============================================
	// 8. LIMITATIONS
	// ============================================
	fmt.Println("\n--- 8. Limitations ---")
	os.Stdout.WriteString(`
LIMITACIONES DE GO WASM:

1. GOROUTINES (GOOS=js):
   - WebAssembly es single-threaded
   - Las goroutines se ejecutan cooperativamente (no preemptive)
   - El scheduler de Go simula concurrencia pero NO paralelismo
   - Goroutines que hacen busy-wait bloquean todo
   - Usa callbacks y channels, no busy loops

2. NETWORKING (GOOS=js):
   - net/* no funciona directamente
   - HTTP requests deben usar fetch API via syscall/js
   - WebSocket disponible via JS interop
   - No hay TCP/UDP sockets directos

3. NETWORKING (GOOS=wasip1):
   - WASI Preview 1 NO tiene soporte de sockets
   - No puedes hacer HTTP requests ni escuchar puertos
   - WASI Preview 2 agregara sockets (futuro)

4. FILESYSTEM:
   - GOOS=js: sin acceso al filesystem real
     (puedes usar memfs via JS, o IndexedDB)
   - GOOS=wasip1: acceso controlado por el runtime
     (--dir flag en wasmtime/wasmer)

5. TAMANO DE BINARIOS:
   - Go standard: ~2-5 MB por la inclusion del runtime
   - TinyGo reduce a ~50-600 KB pero con limitaciones
   - JavaScript bundled tipicamente < 500 KB
   - Impacto en tiempo de carga inicial

6. CGO:
   - No disponible en WASM
   - Librerias que dependen de CGO no compilan
   - Alternativa: reimplementar en pure Go

7. OS/EXEC:
   - No se pueden lanzar subprocesos
   - os/signal no disponible
   - No hay acceso a hardware directo

8. REFLECT:
   - Go standard: funciona completo
   - TinyGo: soporte parcial (algunos metodos fallan)

9. PERFORMANCE:
   - ~40-60%% del rendimiento nativo para computacion pura
   - Overhead en llamadas Go <-> JS (marshalling)
   - GC pauses pueden causar jank en animaciones
   - Memory limitada (tipicamente 4GB max en browsers)

10. DEBUGGING:
    - Source maps limitados
    - No hay debugger step-through como en Go nativo
    - fmt.Println va a la consola del browser (console.log)
    - Profiling limitado comparado con Go nativo (pprof)
`)

	// ============================================
	// 9. PERFORMANCE CONSIDERATIONS
	// ============================================
	fmt.Println("\n--- 9. Performance Considerations ---")
	os.Stdout.WriteString(`
PERFORMANCE DE GO WASM:

BENCHMARKS TIPICOS (vs Go nativo):

  Operacion          | Nativo | WASM  | Ratio
  -------------------+--------+-------+------
  Fibonacci(40)      | 500ms  | 800ms | 1.6x
  SHA256 (1MB)       | 2ms    | 5ms   | 2.5x
  JSON parse (100KB) | 1ms    | 3ms   | 3.0x
  Sort (1M ints)     | 80ms   | 150ms | 1.9x
  Regex match        | 10us   | 25us  | 2.5x

  Nota: los numeros son aproximados y varian segun browser/runtime.

OPTIMIZACIONES:

1. REDUCIR LLAMADAS JS <-> GO:
   - Cada llamada cruzada tiene overhead (~1-5us)
   - Batch operations: pasar arrays en vez de items individuales
   - Minimizar acceso al DOM desde Go

   // MAL: muchas llamadas pequenas
   for i := 0; i < 1000; i++ {
       document.Call("getElementById", fmt.Sprintf("item-%%d", i)).
           Set("textContent", values[i])
   }

   // MEJOR: una llamada con batch update
   js.Global().Call("batchUpdate", js.ValueOf(values))

2. USAR SharedArrayBuffer PARA DATOS GRANDES:
   - Compartir memoria entre Go WASM y JS
   - Evitar copiar datos en cada llamada
   - Requiere headers COOP/COEP en el servidor

3. COMPILAR CON OPTIMIZACIONES:
   // Go standard
   GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm .

   // TinyGo
   tinygo build -o main.wasm -target wasm -opt=2 -no-debug .

4. LAZY LOADING:
   - No cargar todo el WASM al inicio
   - Usar WebAssembly.compileStreaming para compilar mientras descarga
   - Cachear el modulo compilado en IndexedDB

5. MEMORY MANAGEMENT:
   - Go WASM empieza con ~16MB de heap
   - Crece automaticamente pero NUNCA decrece
   - Evitar allocations innecesarias en hot paths
   - Reutilizar buffers (sync.Pool funciona en WASM)

6. WEB WORKERS:
   - Ejecutar WASM en un Web Worker para no bloquear el UI
   - Comunicacion via postMessage
   - Cada Worker tiene su propia instancia WASM

CUANDO WASM ES MAS RAPIDO QUE JS:
  - Computacion numerica intensiva
  - Procesamiento de imagenes/audio/video
  - Criptografia (SHA, AES, etc.)
  - Parseo de formatos binarios
  - Compresion/descompresion

CUANDO JS ES SUFICIENTE:
  - Manipulacion del DOM
  - Animaciones CSS/JS
  - Logica de UI simple
  - HTTP requests y manejo de datos JSON
`)

	// ============================================
	// 10. COMPLETE BROWSER EXAMPLE
	// ============================================
	fmt.Println("\n--- 10. Complete Browser Example ---")
	os.Stdout.WriteString(`
EJEMPLO COMPLETO: CALCULADORA WASM EN EL NAVEGADOR

// main.go (compilar con GOOS=js GOARCH=wasm)
package main

import (
    "fmt"
    "math"
    "syscall/js"
)

func registerCallbacks() {
    js.Global().Set("goAdd", js.FuncOf(func(this js.Value, args []js.Value) any {
        if len(args) < 2 {
            return "Error: need 2 arguments"
        }
        return args[0].Float() + args[1].Float()
    }))

    js.Global().Set("goSqrt", js.FuncOf(func(this js.Value, args []js.Value) any {
        if len(args) < 1 {
            return "Error: need 1 argument"
        }
        return math.Sqrt(args[0].Float())
    }))

    js.Global().Set("goFibonacci", js.FuncOf(func(this js.Value, args []js.Value) any {
        n := args[0].Int()
        return fibonacci(n)
    }))

    js.Global().Set("goRenderList", js.FuncOf(func(this js.Value, args []js.Value) any {
        document := js.Global().Get("document")
        container := document.Call("getElementById", "list")
        container.Set("innerHTML", "")

        for i := 0; i < 10; i++ {
            li := document.Call("createElement", "li")
            li.Set("textContent", fmt.Sprintf("Item %%d: fib(%%d) = %%d", i, i, fibonacci(i)))
            container.Call("appendChild", li)
        }
        return nil
    }))

    fmt.Println("Go WASM functions registered!")
}

func fibonacci(n int) int {
    if n <= 1 { return n }
    a, b := 0, 1
    for i := 2; i <= n; i++ { a, b = b, a+b }
    return b
}

func main() {
    registerCallbacks()
    select {}
}

HTML:

<!DOCTYPE html>
<html>
<head><title>Go WASM Calculator</title></head>
<body>
  <h1>Go WASM Calculator</h1>

  <div>
    <input id="a" type="number" value="10">
    <input id="b" type="number" value="20">
    <button onclick="calculate()">Add (Go)</button>
    <span id="result"></span>
  </div>

  <button onclick="goRenderList()">Render Fibonacci List</button>
  <ul id="list"></ul>

  <script src="wasm_exec.js"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(
      fetch("main.wasm"), go.importObject
    ).then(result => go.run(result.instance));

    function calculate() {
      const a = parseFloat(document.getElementById("a").value);
      const b = parseFloat(document.getElementById("b").value);
      document.getElementById("result").textContent = "= " + goAdd(a, b);
    }
  </script>
</body>
</html>

BUILD & SERVE:
  GOOS=js GOARCH=wasm go build -o main.wasm .
  cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
  python3 -m http.server 8080
`)

	// ============================================
	// 11. COMPLETE WASI EXAMPLE
	// ============================================
	fmt.Println("\n--- 11. Complete WASI Example ---")
	os.Stdout.WriteString(`
EJEMPLO COMPLETO: HERRAMIENTA CLI COMO WASM (WASI)

// main.go (compilar con GOOS=wasip1 GOARCH=wasm)
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

type Config struct {
    Name    string   ` + "`" + `json:"name"` + "`" + `
    Version string   ` + "`" + `json:"version"` + "`" + `
    Tags    []string ` + "`" + `json:"tags"` + "`" + `
}

func main() {
    fmt.Println("=== WASI Go Tool ===")
    fmt.Println("Args:", os.Args)

    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "upper":
            if len(os.Args) > 2 {
                fmt.Println(strings.ToUpper(os.Args[2]))
            }
        case "config":
            cfg := Config{
                Name:    "my-wasi-app",
                Version: "1.0.0",
                Tags:    []string{"wasm", "wasi", "go"},
            }
            enc := json.NewEncoder(os.Stdout)
            enc.SetIndent("", "  ")
            enc.Encode(cfg)
        case "write":
            err := os.WriteFile("output.txt", []byte("Hello from WASI!\n"), 0644)
            if err != nil {
                fmt.Fprintln(os.Stderr, "Error:", err)
                os.Exit(1)
            }
            fmt.Println("File written successfully")
        default:
            fmt.Println("Unknown command:", os.Args[1])
        }
    }
}

BUILD & RUN:
  GOOS=wasip1 GOARCH=wasm go build -o tool.wasm .
  wasmtime tool.wasm config
  wasmtime tool.wasm upper "hello world"
  wasmtime --dir=. tool.wasm write
  wazero run tool.wasm config
`)

	// ============================================
	// 12. WORKING DEMO
	// ============================================
	fmt.Println("\n--- 12. Working Demo ---")
	fmt.Println("  Functions that work in any context (native or WASM):")

	fmt.Println("\n  Fibonacci sequence (iterative):")
	for i := 0; i <= 15; i++ {
		fmt.Printf("    fib(%2d) = %d\n", i, fibonacci(i))
	}

	fmt.Println("\n  Prime sieve (would work identically in WASM):")
	primes := sieveOfEratosthenes(100)
	fmt.Println("    Primes up to 100:", primes)

	fmt.Println("\n  Math computations (portable across native/WASM):")
	values := []float64{2, 16, 144, 1024, 65536}
	for _, v := range values {
		fmt.Printf("    sqrt(%.0f) = %.4f, log2(%.0f) = %.4f\n",
			v, math.Sqrt(v), v, math.Log2(v))
	}

	fmt.Println("\n  String processing (same code in browser or server):")
	words := []string{"webassembly", "golang", "browser", "wasi", "portable"}
	for _, w := range words {
		fmt.Printf("    %-12s -> upper: %-12s | reversed: %s\n",
			w, strings.ToUpper(w), reverseString(w))
	}

	fmt.Println("\n  Caesar cipher (encryption logic shareable via WASM):")
	original := "Hello WebAssembly from Go!"
	encrypted := caesarCipher(original, 13)
	decrypted := caesarCipher(encrypted, -13)
	fmt.Println("    Original: ", original)
	fmt.Println("    Encrypted:", encrypted)
	fmt.Println("    Decrypted:", decrypted)
	fmt.Println("    Match:    ", original == decrypted)

	// ============================================
	// 13. ADVANCED PATTERNS
	// ============================================
	fmt.Println("\n--- 13. Advanced Patterns ---")
	os.Stdout.WriteString(`
PATRON: PROMISE WRAPPER

Envolver funciones Go async como JavaScript Promises:

  func promiseWrapper(fn func() (any, error)) js.Value {
      handler := js.FuncOf(func(this js.Value, args []js.Value) any {
          resolve := args[0]
          reject := args[1]

          go func() {
              result, err := fn()
              if err != nil {
                  reject.Invoke(js.Global().Get("Error").New(err.Error()))
                  return
              }
              resolve.Invoke(js.ValueOf(result))
          }()

          return nil
      })

      return js.Global().Get("Promise").New(handler)
  }

  // En JavaScript:
  //   const data = await goFetchData("/api/items");

PATRON: TYPED ARRAY BRIDGE

Compartir datos binarios eficientemente entre Go y JS:

  // Go -> JS: pasar []byte como Uint8Array
  func goToJSBytes(data []byte) js.Value {
      arr := js.Global().Get("Uint8Array").New(len(data))
      js.CopyBytesToJS(arr, data)
      return arr
  }

  // JS -> Go: recibir Uint8Array como []byte
  func jsToGoBytes(arr js.Value) []byte {
      data := make([]byte, arr.Length())
      js.CopyBytesToGo(data, arr)
      return data
  }

PATRON: EVENT BUS CON CHANNELS

  type EventBus struct {
      events chan js.Value
  }

  func NewEventBus() *EventBus {
      eb := &EventBus{events: make(chan js.Value, 100)}
      go eb.processEvents()
      return eb
  }

  func (eb *EventBus) Handler() js.Func {
      return js.FuncOf(func(this js.Value, args []js.Value) any {
          if len(args) > 0 {
              eb.events <- args[0]
          }
          return nil
      })
  }

  func (eb *EventBus) processEvents() {
      for event := range eb.events {
          eventType := event.Get("type").String()
          fmt.Println("Processing event:", eventType)
      }
  }

PATRON: LIFECYCLE MANAGEMENT

  func main() {
      registerFunctions()

      // Senalar que Go esta listo
      js.Global().Get("document").Call("dispatchEvent",
          js.Global().Get("CustomEvent").New("go-ready"),
      )

      <-make(chan struct{})
  }

  // JS: document.addEventListener("go-ready", () => initializeApp());
`)

	// ============================================
	// 14. PROJECT STRUCTURE
	// ============================================
	fmt.Println("\n--- 14. Project Structure for WASM Apps ---")
	fmt.Println(`
ESTRUCTURA RECOMENDADA:

  myapp/
  ├── cmd/
  │   ├── server/         # Backend Go (sirve el WASM)
  │   │   └── main.go
  │   └── wasm/           # Frontend WASM
  │       └── main.go     # Punto de entrada WASM
  ├── internal/
  │   └── shared/         # Logica compartida (compila en ambos)
  │       ├── calc.go
  │       ├── validate.go
  │       └── calc_test.go
  ├── web/
  │   ├── index.html
  │   ├── wasm_exec.js    # Copiado de GOROOT
  │   ├── main.wasm       # Generado por build
  │   └── app.js          # JS complementario
  ├── Makefile
  └── go.mod

MAKEFILE:

  .PHONY: build-wasm serve clean

  build-wasm:
  	GOOS=js GOARCH=wasm go build -o web/main.wasm ./cmd/wasm/
  	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/

  serve: build-wasm
  	go run ./cmd/server/

  test:
  	go test ./internal/shared/...

  test-wasm:
  	GOOS=js GOARCH=wasm go test ./cmd/wasm/...

  clean:
  	rm -f web/main.wasm web/wasm_exec.js`)

	// ============================================
	// 15. SUMMARY
	// ============================================
	fmt.Println("\n--- 15. Summary ---")
	fmt.Println(`
  +------------------+-------------------------------------+
  | Concepto         | Detalle                             |
  +------------------+-------------------------------------+
  | Browser target   | GOOS=js GOARCH=wasm                 |
  | WASI target      | GOOS=wasip1 GOARCH=wasm (Go 1.21+) |
  | JS interop       | syscall/js (js.Value, js.FuncOf)    |
  | Glue code        | wasm_exec.js (from GOROOT/misc)     |
  | Keep alive       | select{} or <-make(chan struct{})    |
  | Binary size      | Go ~2MB, TinyGo ~50KB               |
  | Performance      | ~40-60% of native speed             |
  | Goroutines       | Cooperative, single-threaded        |
  | Networking       | Via JS fetch (browser), none (WASI) |
  | Filesystem       | None (browser), sandboxed (WASI)    |
  | WASM runtime     | wazero (pure Go, no CGO)            |
  | Plugin system    | wazero + TinyGo exports             |
  | DOM access       | js.Global().Get("document")         |
  | Go->JS calls     | value.Call("method", args...)        |
  | JS->Go calls     | js.Global().Set("name", js.FuncOf)  |
  | Smaller binaries | TinyGo with -opt=z -no-debug        |
  +------------------+-------------------------------------+`)
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func sieveOfEratosthenes(limit int) []int {
	sieve := make([]bool, limit+1)
	for i := 2; i <= limit; i++ {
		sieve[i] = true
	}
	for i := 2; i*i <= limit; i++ {
		if sieve[i] {
			for j := i * i; j <= limit; j += i {
				sieve[j] = false
			}
		}
	}
	var primes []int
	for i := 2; i <= limit; i++ {
		if sieve[i] {
			primes = append(primes, i)
		}
	}
	return primes
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func caesarCipher(text string, shift int) string {
	shift = ((shift % 26) + 26) % 26
	var result strings.Builder
	for _, ch := range text {
		switch {
		case ch >= 'a' && ch <= 'z':
			result.WriteRune('a' + (ch-'a'+rune(shift))%26)
		case ch >= 'A' && ch <= 'Z':
			result.WriteRune('A' + (ch-'A'+rune(shift))%26)
		default:
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// Suppress unused import warnings
var _ = math.Sqrt
var _ = os.Stdout
var _ = fmt.Sprintf

/*
RESUMEN CHAPTER 75 - GO AND WEBASSEMBLY (WASM):

WebAssembly:
- Formato binario portable para maquina virtual basada en stack
- Go soporta dos targets: browser (GOOS=js) y WASI (GOOS=wasip1)
- Permite reutilizar logica Go en browsers y edge computing

Browser target (GOOS=js GOARCH=wasm):
- Compila a .wasm ejecutable en navegadores
- Requiere wasm_exec.js como glue code entre Go runtime y JS
- syscall/js para interoperabilidad bidireccional Go <-> JavaScript
- js.Global(), js.Value, js.FuncOf para DOM y JS API access
- select{} para mantener el programa vivo

WASI target (GOOS=wasip1 GOARCH=wasm):
- Go 1.21+ soporta WASI Preview 1
- Ejecutable fuera del browser (wasmtime, wasmer, wazero)
- Acceso a filesystem, stdin/stdout, args, env vars
- Sin soporte de networking (WASI Preview 1 limitation)

syscall/js:
- js.Global().Get/Set para acceso a objetos globales JS
- js.FuncOf para crear callbacks Go callable desde JS
- js.CopyBytesToJS/js.CopyBytesToGo para datos binarios
- Conversiones automaticas de tipos Go <-> JS

TinyGo:
- Compilador alternativo basado en LLVM
- Binarios 87-97% mas pequenos que Go standard
- ~50KB vs ~2MB para programas simples
- Limitaciones en reflect, stdlib y GC
- Tiene su propio wasm_exec.js (no mezclar)

wazero:
- Runtime WASM pure Go (zero CGO dependencies)
- Permite ejecutar modulos WASM desde apps Go
- Ideal para plugin systems, sandboxing, edge computing
- API idiomatica con context.Context

Use cases:
- Browser apps, edge computing, plugin systems
- Sandboxed execution, shared frontend/backend logic
- Serverless con cold start sub-millisecond

Limitations:
- Single-threaded (goroutines cooperativas, no paralelas)
- Binarios grandes con Go standard
- Sin CGO, sin os/exec, sin signals
- Networking limitada segun target
- ~40-60% rendimiento nativo

Performance:
- Reducir llamadas Go <-> JS (batch operations)
- Usar Uint8Array para datos binarios
- Web Workers para no bloquear UI
- TinyGo para menor tamano y carga rapida
*/
