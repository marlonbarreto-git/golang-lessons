# Curso Go: De Cero a Ninja 2026

Curso completo de Go (Golang) actualizado a 2026 con 79 capítulos, cubriendo desde Hello World hasta Go Assembly y WASM, con las mejores prácticas, patrones de diseño, toda la standard library, y las últimas novedades del lenguaje (Go 1.23, 1.24, 1.25, 1.26).

**Autor:** Marlon Barreto (mbarretot@hotmail.com)

## Requisitos Previos

- Conocimientos básicos de programación
- Go 1.24+ instalado ([descargar](https://go.dev/dl/))
- Editor de código (VS Code con extensión Go, GoLand, o Neovim)

## Estructura del Curso

### Parte I: Fundamentos

| # | Tema | Descripción |
|---|------|-------------|
| 01 | [Hola Mundo](./01_hola_mundo/) | Primer programa, go run, go build, estructura básica |
| 02 | [Variables y Constantes](./02_variables_constantes/) | Tipos de datos, declaración, zero values, inferencia |
| 03 | [Operadores](./03_operadores/) | Aritméticos, lógicos, comparación, bit a bit |
| 04 | [Condicionales](./04_condicionales/) | if/else, operador ternario (patrón), early return |
| 05 | [Bucles](./05_bucles/) | for, range, break, continue, labels |
| 06 | [Switch](./06_switch/) | switch/case, type switch, fallthrough |
| 07 | [Colecciones](./07_colecciones/) | Arrays, slices, maps, operaciones comunes |
| 08 | [Strings](./08_strings/) | Strings, runes, bytes, unicode, strings package |
| 09 | [Funciones](./09_funciones/) | Parámetros, retornos, variadic, closures, defer |
| 10 | [Punteros](./10_punteros/) | Direcciones de memoria, referencias, nil |

### Parte II: Tipos y Estructuras

| # | Tema | Descripción |
|---|------|-------------|
| 11 | [Structs](./11_structs/) | Definición, campos, embedding, métodos |
| 12 | [Interfaces](./12_interfaces/) | Definición, implementación implícita, composición |
| 13 | [Type System](./13_type_system/) | Alias, wrappers, type assertions, type switch |
| 14 | [Enumeradores](./14_enumeradores/) | iota, patrones de enum, String() |
| 15 | [Generics](./15_generics/) | Type parameters, constraints, cuando usar |

### Parte III: Manejo de Errores

| # | Tema | Descripción |
|---|------|-------------|
| 16 | [Errors](./16_errors/) | error interface, errors package, wrapping |
| 17 | [Panic y Recover](./17_panic_recover/) | Cuándo usar panic, recover, defer |
| 18 | [Error Handling Patterns](./18_error_patterns/) | Sentinel errors, custom errors, best practices |

### Parte IV: Concurrencia

| # | Tema | Descripción |
|---|------|-------------|
| 19 | [Goroutines](./19_goroutines/) | Creación, scheduling, WaitGroup, WaitGroup.Go() |
| 20 | [Channels](./20_channels/) | Buffered, unbuffered, direccionalidad |
| 21 | [Select](./21_select/) | Multiplexación, timeouts, default |
| 22 | [Context](./22_context/) | Cancelación, timeouts, valores, propagación |
| 23 | [Sync Package](./23_sync/) | Mutex, RWMutex, Once, Pool, Map, Cond |
| 24 | [Patrones de Concurrencia](./24_concurrency_patterns/) | Worker pools, fan-in/out, pipelines |

### Parte V: Testing

| # | Tema | Descripción |
|---|------|-------------|
| 25 | [Unit Testing](./25_testing/) | testing package, table-driven tests, subtests |
| 26 | [Benchmarks](./26_benchmarks/) | Benchmarking, b.Loop(), profiling |
| 27 | [Mocking](./27_mocking/) | Interfaces, testify, gomock |
| 28 | [Fuzzing](./28_fuzzing/) | Fuzz testing, corpus, coverage |

### Parte VI: Organización de Código

| # | Tema | Descripción |
|---|------|-------------|
| 29 | [Paquetes](./29_packages/) | Visibilidad, init(), internal, vendor |
| 30 | [Módulos](./30_modules/) | go.mod, versioning, dependencias, workspaces |
| 31 | [Patrones de Proyecto](./31_project_patterns/) | Standard layout, clean architecture, DDD |

### Parte VII: Avanzado

| # | Tema | Descripción |
|---|------|-------------|
| 32 | [Reflection](./32_reflection/) | reflect package, casos de uso, performance |
| 33 | [Unsafe](./33_unsafe/) | unsafe.Pointer, uintptr, casos válidos |
| 34 | [CGO](./34_cgo/) | Llamar C desde Go, callbacks, performance |
| 35 | [Go Assembly](./35_go_assembly/) | Plan 9 assembly, SIMD, optimización |

### Parte VIII: Aplicaciones Prácticas

| # | Tema | Descripción |
|---|------|-------------|
| 36 | [Web Backend](./36_web_backend/) | net/http, chi, fiber, middleware, REST |
| 37 | [Bases de Datos](./37_databases/) | database/sql, pgx, gorm, migrations |
| 38 | [gRPC](./38_grpc/) | Protocol Buffers, servicios, streaming |
| 39 | [CLI Applications](./39_cli/) | cobra, bubbletea, TUI |
| 40 | [Desktop Apps](./40_desktop/) | fyne, wails, GUI nativa |

### Parte IX: Especialización

| # | Tema | Descripción |
|---|------|-------------|
| 41 | [AI y ML](./41_ai_ml/) | gonum, gorgonia, langchaingo, ONNX |
| 42 | [Procesamiento de Imágenes](./42_images/) | image stdlib, gocv, bild |
| 43 | [Procesamiento de Audio](./43_audio/) | beep, oto, go-audio |
| 44 | [Data Processing](./44_data/) | gonum, arrow, parquet, ETL |

### Parte X: Go Moderno (2024-2026)

| # | Tema | Descripción |
|---|------|-------------|
| 45 | [Go 1.24 Features](./45_go124/) | Swiss Tables, generic aliases, FIPS 140-3 |
| 46 | [Go 1.25 Features](./46_go125/) | WaitGroup.Go(), JSON v2, synctest |
| 47 | [Go 1.26 Features](./47_go126/) | SIMD, crypto/hpke, new() expressions |

### Parte XI: Producción

| # | Tema | Descripción |
|---|------|-------------|
| 48 | [Performance](./48_performance/) | Profiling, escape analysis, optimización |
| 49 | [Observabilidad](./49_observability/) | Logging, metrics, tracing, OpenTelemetry |
| 50 | [Deploy](./50_deploy/) | Docker, Kubernetes, cloud-native |

### Parte XII: Go Avanzado 2026

| # | Tema | Descripción |
|---|------|-------------|
| 52 | [Iterators](./52_iterators/) | Range over functions (Go 1.23+), iter package |
| 53 | [go:embed](./53_embed/) | Embedding files, assets, templates en binario |
| 54 | [Build Tags](./54_build_tags/) | Conditional compilation, cross-platform |
| 55 | [WebSockets](./55_websockets/) | Real-time communication, gorilla/websocket |
| 56 | [Redis y Caching](./56_redis_caching/) | go-redis, patrones de cache, invalidación |
| 57 | [Message Queues](./57_message_queues/) | Kafka, NATS, RabbitMQ, event-driven |
| 58 | [Security](./58_security/) | OWASP, input validation, secure headers |
| 59 | [Auth](./59_auth/) | JWT, OAuth2, OIDC, password hashing |
| 60 | [Code Generation](./60_code_generation/) | go generate, stringer, mockgen, sqlc |
| 61 | [Testing Avanzado](./61_testing_advanced/) | testify, gomock, integration tests |

### Parte XIII: Sistemas y Arquitectura

| # | Tema | Descripción |
|---|------|-------------|
| 62 | [I/O Patterns](./62_io_patterns/) | io.Reader/Writer, bufio, streaming, fs |
| 63 | [HTTP Client](./63_http_client/) | Timeouts, pooling, retries, middleware |
| 64 | [Design Patterns](./64_design_patterns/) | Functional options, DI, graceful shutdown |
| 65 | [Resilience](./65_resilience/) | Rate limiting, circuit breaker, bulkhead |
| 66 | [Networking](./66_networking/) | TCP/UDP, TLS/mTLS, DNS, net package |
| 67 | [GraphQL](./67_graphql/) | gqlgen, schema design, subscriptions |
| 68 | [Event-Driven](./68_event_driven/) | CQRS, event sourcing, saga, outbox |
| 69 | [API Design](./69_api_design/) | Versioning, OpenAPI, pagination, ETags |

### Parte XIV: Standard Library Mastery

| # | Tema | Descripción |
|---|------|-------------|
| 70 | [Time](./70_time/) | Parsing, formatting, timezones, monotonic clock, DST |
| 71 | [Regexp](./71_regexp/) | Patrones, named groups, RE2, performance |
| 72 | [Encoding](./72_encoding/) | JSON, XML, gob, CSV, binary, base64, hex |
| 73 | [Templates](./73_templates/) | text/template, html/template, FuncMap, composición |
| 74 | [Crypto](./74_crypto/) | AES, RSA, ECDSA, hashing, TLS, certificados |
| 75 | [WASM](./75_wasm/) | WebAssembly, syscall/js, WASI, TinyGo, wazero |

### Parte XV: Tooling y Filosofía

| # | Tema | Descripción |
|---|------|-------------|
| 76 | [CI/CD & Tooling](./76_cicd_tooling/) | Makefile, GitHub Actions, goreleaser, govulncheck |
| 77 | [Debugging](./77_debugging_diagnostics/) | Delve, race detector, memory leaks, pprof |
| 78 | [OS & Processes](./78_os_exec_processes/) | os/exec, signals, environment, process management |
| 79 | [Go Proverbs](./79_go_proverbs_mistakes/) | Filosofía Go, errores comunes, patrones idiomáticos |

### Proyecto Final

| # | Tema | Descripción |
|---|------|-------------|
| 51 | [Proyecto Integrador](./51_proyecto_final/) | Aplicación completa aplicando todo el curso |

## Cómo Usar Este Curso

1. **Secuencial**: Sigue los capítulos en orden para máximo aprendizaje
2. **Por tema**: Salta a temas específicos según tu necesidad
3. **Práctica**: Cada capítulo tiene ejercicios progresivos

## Ejecutar los Ejemplos

```bash
# Navegar al capítulo
cd 01_hola_mundo

# Ejecutar directamente
go run main.go

# O compilar y ejecutar
go build -o programa
./programa
```

## Convenciones del Código

- **Early return**: Siempre preferir retornos tempranos sobre anidación
- **Error handling**: Manejar errores inmediatamente después de la llamada
- **Naming**: Nombres cortos y descriptivos (Go idiomático)
- **Comments**: Solo cuando el código no es auto-explicativo

## Recursos Adicionales

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Go Blog](https://go.dev/blog)
- [Go Playground](https://go.dev/play)

## Versiones de Go Cubiertas

| Versión | Release | Novedades Principales |
|---------|---------|----------------------|
| Go 1.23 | Aug 2024 | Range over func, unique package |
| Go 1.24 | Feb 2025 | Swiss Tables, generic aliases, FIPS 140-3 |
| Go 1.25 | Aug 2025 | WaitGroup.Go(), JSON v2, synctest |
| Go 1.26 | Feb 2026 | SIMD experimental, crypto/hpke |

## Licencia

MIT License - Libre para uso educativo y comercial.
