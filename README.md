# Go Course: From Zero to Ninja 2026

Complete Go (Golang) course updated to 2026 with **132 chapters** organized in **19 parts**, covering everything from Hello World to Go Assembly, WASM, and major Go projects like Docker and Kubernetes. Includes best practices, design patterns, the entire standard library, golang.org/x/ packages, and the latest language features (Go 1.23, 1.24, 1.25, 1.26).

**Author:** Marlon Barreto (mbarretot@hotmail.com)

## Prerequisites

- Basic programming knowledge
- Go 1.24+ installed ([download](https://go.dev/dl/))
- Code editor (VS Code with Go extension, GoLand, or Neovim)

## Course Structure

### Part I: Language Basics

| # | Topic | Description |
|---|-------|-------------|
| 001 | [Hello World](./001_hello_world/) | First program, go run, go build, basic structure |
| 002 | [Variables & Constants](./002_variables_constants/) | Data types, declarations, zero values, inference |
| 003 | [Operators](./003_operators/) | Arithmetic, logical, comparison, bitwise |
| 004 | [Conditionals](./004_conditionals/) | if/else, ternary pattern, early return |
| 005 | [Loops](./005_loops/) | for, range, break, continue, labels |
| 006 | [Switch](./006_switch/) | switch/case, type switch, fallthrough |
| 007 | [Functions](./007_functions/) | Parameters, returns, variadic, closures, defer |
| 008 | [Pointers](./008_pointers/) | Memory addresses, references, nil |
| 009 | [Strings](./009_strings/) | Strings, runes, bytes, unicode, strings package |
| 010 | [Collections](./010_collections/) | Arrays, slices, maps, common operations |

### Part II: Data Structures Deep Dive

| # | Topic | Description |
|---|-------|-------------|
| 011 | [Slices Deep Dive](./011_slices_deep_dive/) | Internals, capacity growth, memory layout, subslice gotchas |
| 012 | [Maps Deep Dive](./012_maps_deep_dive/) | Internals, Swiss Tables (Go 1.24), iteration, nil vs empty |
| 013 | [Sort Package](./013_sort_package/) | sort.Slice, sort.Interface, binary search, custom comparators |
| 014 | [Container Types](./014_container_types/) | container/heap, container/list, container/ring |
| 015 | [Slices, Maps & Cmp](./015_slices_maps_cmp/) | Modern slices/maps/cmp packages (Go 1.21+) |

### Part III: Types and OOP

| # | Topic | Description |
|---|-------|-------------|
| 016 | [Structs](./016_structs/) | Definition, fields, embedding, methods |
| 017 | [Interfaces](./017_interfaces/) | Definition, implicit implementation, composition |
| 018 | [Type System](./018_type_system/) | Alias, wrappers, type assertions, type switch |
| 019 | [Enums & Iota](./019_enums_iota/) | iota, enum patterns, String() |
| 020 | [Generics](./020_generics/) | Type parameters, constraints, when to use |
| 021 | [Method Sets & Receivers](./021_method_sets_receivers/) | Value vs pointer receivers, interface satisfaction rules |
| 022 | [Interface Internals](./022_interface_internals/) | iface/eface, nil interfaces, comparable constraint |

### Part IV: Error Handling

| # | Topic | Description |
|---|-------|-------------|
| 023 | [Errors](./023_errors/) | error interface, errors package, wrapping |
| 024 | [Panic & Recover](./024_panic_recover/) | When to use panic, recover, defer |
| 025 | [Error Patterns](./025_error_patterns/) | Sentinel errors, custom errors, best practices |

### Part V: Code Organization

| # | Topic | Description |
|---|-------|-------------|
| 026 | [Packages](./026_packages/) | Visibility, init(), internal, vendor |
| 027 | [Init & Blank Identifier](./027_init_blank_identifier/) | init() execution order, _ patterns, side-effect imports |
| 028 | [Modules](./028_modules/) | go.mod, versioning, dependencies, workspaces |
| 029 | [Project Patterns](./029_project_patterns/) | Standard layout, clean architecture, DDD |

### Part VI: Concurrency

| # | Topic | Description |
|---|-------|-------------|
| 030 | [Goroutines](./030_goroutines/) | Creation, scheduling, WaitGroup, WaitGroup.Go() |
| 031 | [Channels](./031_channels/) | Buffered, unbuffered, directionality |
| 032 | [Select](./032_select/) | Multiplexing, timeouts, default |
| 033 | [Context](./033_context/) | Cancellation, timeouts, values, propagation |
| 034 | [Sync Package](./034_sync/) | Mutex, RWMutex, Once, Pool, Map, Cond |
| 035 | [Concurrency Patterns](./035_concurrency_patterns/) | Worker pools, fan-in/out, pipelines |

### Part VII: Testing

| # | Topic | Description |
|---|-------|-------------|
| 036 | [Testing](./036_testing/) | testing package, table-driven tests, subtests |
| 037 | [Benchmarks](./037_benchmarks/) | Benchmarking, b.Loop(), profiling |
| 038 | [Mocking](./038_mocking/) | Interfaces, testify, gomock |
| 039 | [Fuzzing](./039_fuzzing/) | Fuzz testing, corpus, coverage |

### Part VIII: Standard Library Essentials

| # | Topic | Description |
|---|-------|-------------|
| 040 | [I/O Patterns](./040_io_patterns/) | io.Reader/Writer, bufio, streaming, fs |
| 041 | [Strings & Strconv](./041_strings_strconv/) | strconv deep dive, strings.Builder, Cut, Fields |
| 042 | [Bytes Package](./042_bytes_package/) | bytes.Buffer, bytes.Reader, zero-copy techniques |
| 043 | [Time](./043_time/) | Parsing, formatting, timezones, monotonic clock |
| 044 | [Regexp](./044_regexp/) | Patterns, named groups, RE2, performance |
| 045 | [Encoding](./045_encoding/) | JSON, XML, gob, CSV, binary, base64, hex |
| 046 | [Templates](./046_templates/) | text/template, html/template, FuncMap, composition |
| 047 | [OS, Exec & Processes](./047_os_exec_processes/) | os/exec, signals, environment, process management |
| 048 | [Flag Package](./048_flag_package/) | Custom flags, FlagSet, subcommands, env fallback |
| 049 | [Path & Filepath](./049_path_filepath/) | path vs filepath, Walk, WalkDir, Glob |
| 050 | [Unicode & UTF-8](./050_unicode_utf8/) | utf8, utf16, rune classification, encoding/decoding |
| 051 | [Hash Functions](./051_hash_functions/) | fnv, crc32, crc64, adler32, maphash |
| 052 | [Structured Logging](./052_log_slog/) | log/slog deep dive, handlers, groups, performance |

### Part IX: Standard Library Advanced

| # | Topic | Description |
|---|-------|-------------|
| 053 | [Math Big & Bits](./053_math_big_bits/) | big.Int, big.Float, math/bits, rand/v2 |
| 054 | [Archive & Compress](./054_archive_compress/) | tar, zip, gzip, zlib, flate |
| 055 | [Crypto](./055_crypto/) | AES, RSA, ECDSA, hashing, TLS, certificates |
| 056 | [Net Mail & SMTP](./056_net_mail_smtp/) | Email parsing, sending, net/rpc |
| 057 | [Text Processing](./057_text_processing/) | text/scanner, tabwriter, mime/multipart |
| 058 | [Go AST Parser](./058_go_ast_parser/) | go/ast, go/parser, go/token, building linters |
| 059 | [Advanced Testing](./059_testing_advanced/) | testify, gomock, integration tests |
| 060 | [Testing Stdlib](./060_testing_stdlib/) | testing/iotest, fstest, golden files, fixtures |
| 061 | [Runtime & Debug](./061_runtime_debug/) | runtime/debug, metrics, trace, binary inspection |

### Part X: Modern Go Features

| # | Topic | Description |
|---|-------|-------------|
| 062 | [Iterators](./062_iterators/) | Range over functions (Go 1.23+), iter package |
| 063 | [go:embed](./063_embed/) | Embedding files, assets, templates in binary |
| 064 | [Build Tags](./064_build_tags/) | Conditional compilation, cross-platform |
| 065 | [Go 1.24 Features](./065_go124_features/) | Swiss Tables, generic aliases, FIPS 140-3 |
| 066 | [Go 1.25 Features](./066_go125_features/) | WaitGroup.Go(), JSON v2, synctest |
| 067 | [Go 1.26 Features](./067_go126_features/) | SIMD experimental, crypto/hpke |

### Part XI: Advanced Language

| # | Topic | Description |
|---|-------|-------------|
| 068 | [Reflection](./068_reflection/) | reflect package, use cases, performance |
| 069 | [Code Generation](./069_code_generation/) | go generate, stringer, mockgen, sqlc |
| 070 | [Unsafe](./070_unsafe/) | unsafe.Pointer, uintptr, valid use cases |
| 071 | [CGO](./071_cgo/) | Calling C from Go, callbacks, performance |
| 072 | [Go Assembly](./072_go_assembly/) | Plan 9 assembly, SIMD, optimization |
| 073 | [Plugin & Expvar](./073_plugin_expvar/) | Go plugins, expvar monitoring, /debug/vars |
| 074 | [WebAssembly](./074_wasm/) | WASM, syscall/js, WASI, TinyGo, wazero |

### Part XII: Design Patterns & Architecture

| # | Topic | Description |
|---|-------|-------------|
| 075 | [Design Patterns](./075_design_patterns/) | Functional options, DI, graceful shutdown |
| 076 | [Resilience](./076_resilience/) | Rate limiting, circuit breaker, bulkhead |
| 077 | [Event-Driven](./077_event_driven/) | CQRS, event sourcing, saga, outbox |
| 078 | [API Design](./078_api_design/) | Versioning, OpenAPI, pagination, ETags |
| 079 | [Security](./079_security/) | OWASP, input validation, secure headers |

### Part XIII: Web & Networking

| # | Topic | Description |
|---|-------|-------------|
| 080 | [Networking](./080_networking/) | TCP/UDP, TLS/mTLS, DNS, net package |
| 081 | [Web Backend](./081_web_backend/) | net/http, chi, fiber, middleware, REST |
| 082 | [HTTP Client](./082_http_client/) | Timeouts, pooling, retries, middleware |
| 083 | [WebSockets](./083_websockets/) | Real-time communication, gorilla/websocket |
| 084 | [gRPC](./084_grpc/) | Protocol Buffers, services, streaming |
| 085 | [GraphQL](./085_graphql/) | gqlgen, schema design, subscriptions |
| 086 | [Authentication](./086_auth/) | JWT, OAuth2, OIDC, password hashing |
| 087 | [CLI Applications](./087_cli_applications/) | cobra, bubbletea, TUI |
| 088 | [Desktop Apps](./088_desktop_apps/) | fyne, wails, native GUI |

### Part XIV: Data, Databases & Messaging

| # | Topic | Description |
|---|-------|-------------|
| 089 | [Databases](./089_databases/) | database/sql, pgx, gorm, migrations |
| 090 | [Redis & Caching](./090_redis_caching/) | go-redis, cache patterns, invalidation |
| 091 | [Message Queues](./091_message_queues/) | Kafka, NATS, RabbitMQ, event-driven |
| 092 | [Data Processing](./092_data_processing/) | gonum, arrow, parquet, ETL |
| 093 | [Images & Audio](./093_images_audio/) | image stdlib, gocv, bild, beep, oto |

### Part XV: golang.org/x/ Ecosystem

| # | Topic | Description |
|---|-------|-------------|
| 094 | [x/sync](./094_x_sync/) | errgroup, singleflight, semaphore |
| 095 | [x/net & x/crypto](./095_x_net_crypto/) | http2, html parsing, ssh, bcrypt, argon2 |
| 096 | [x/text](./096_x_text/) | transforms, language tags, collation, normalization |
| 097 | [x/oauth2 & x/term](./097_x_oauth2_term/) | OAuth2 flows, terminal I/O, raw mode |
| 098 | [x/sys & x/mod](./098_x_sys_mod/) | Platform syscalls, semver, modfile |
| 099 | [x/tools](./099_x_tools/) | go/analysis, custom analyzers, goimports |

### Part XVI: Production & Deployment

| # | Topic | Description |
|---|-------|-------------|
| 100 | [Performance](./100_performance/) | Profiling, escape analysis, optimization |
| 101 | [Observability](./101_observability/) | Logging, metrics, tracing, OpenTelemetry |
| 102 | [Debugging](./102_debugging_diagnostics/) | Delve, race detector, memory leaks, pprof |
| 103 | [CI/CD & Tooling](./103_cicd_tooling/) | Makefile, GitHub Actions, goreleaser |
| 104 | [Deploy](./104_deploy/) | Docker, Kubernetes, cloud-native |
| 105 | [AI & ML](./105_ai_ml/) | gonum, gorgonia, langchaingo, ONNX |
| 106 | [Final Project](./106_proyecto_final/) | Complete integrator project |

### Part XVII: Major Go Projects

Each chapter studies a real-world Go project: architecture, patterns, optimizations, Go philosophy, and includes a simplified working version.

| # | Project | Description |
|---|---------|-------------|
| 107 | [Docker](./107_project_docker/) | Container runtime, namespaces, cgroups, layered FS |
| 108 | [Kubernetes](./108_project_kubernetes/) | Controller pattern, reconciliation loops, informers |
| 109 | [etcd](./109_project_etcd/) | Raft consensus, distributed KV, WAL |
| 110 | [Terraform](./110_project_terraform/) | DAG dependency resolution, IaC, plugin architecture |
| 111 | [Prometheus](./111_project_prometheus/) | Pull-based metrics, time-series DB, PromQL |
| 112 | [Grafana](./112_project_grafana/) | Plugin architecture, dashboards, data sources |
| 113 | [Caddy](./113_project_caddy/) | Automatic HTTPS, module system, clean API |
| 114 | [Hugo](./114_project_hugo/) | Content pipeline, template engine, static site generation |
| 115 | [Traefik](./115_project_traefik/) | Dynamic reverse proxy, provider pattern, load balancing |
| 116 | [CoreDNS](./116_project_coredns/) | Plugin chain, DNS server, middleware composition |
| 117 | [Consul](./117_project_consul/) | Gossip protocol, service discovery, health checking |
| 118 | [Vault](./118_project_vault/) | Secrets management, seal/unseal, encryption |
| 119 | [CockroachDB](./119_project_cockroachdb/) | Distributed SQL, range partitioning, transactions |
| 120 | [NATS](./120_project_nats/) | Pub/sub messaging, request/reply, queue groups |
| 121 | [Minio](./121_project_minio/) | S3-compatible object storage, erasure coding |
| 122 | [Gitea](./122_project_gitea/) | Full-stack Go web app, Git service |
| 123 | [Badger/BoltDB](./123_project_badger_boltdb/) | Embedded KV: LSM tree vs B+ tree |
| 124 | [Cilium](./124_project_cilium/) | eBPF networking, kernel-level packet processing |
| 127 | [Ollama](./127_project_ollama/) | Local LLM serving, model management, streaming API |
| 128 | [TiDB](./128_project_tidb/) | Distributed SQL, Raft consensus, MySQL compatible |
| 129 | [Istio](./129_project_istio/) | Service mesh control plane, xDS, mTLS, traffic routing |
| 130 | [Milvus](./130_project_milvus/) | Vector database, similarity search, HNSW index, AI/RAG |

### Part XVIII: Advanced Testing & Modern Features

| # | Topic | Description |
|---|-------|-------------|
| 131 | [Property Testing](./131_property_testing/) | Property-based testing, testing/quick, generators, shrinking |
| 132 | [JSON v2](./132_json_v2/) | encoding/json/v2, strict mode, omitzero, migration from v1 |

### Part XIX: Philosophy & Mastery

| # | Topic | Description |
|---|-------|-------------|
| 125 | [Go Proverbs](./125_go_proverbs_mistakes/) | Go philosophy, common mistakes, idiomatic patterns |
| 126 | [Mastery Path](./126_go_mastery_path/) | Career paths, community, contributing to Go, resources |

## How to Use This Course

1. **Sequential**: Follow chapters in order for maximum learning
2. **By topic**: Jump to specific topics as needed
3. **Practice**: Each chapter has progressive examples

## Running the Examples

```bash
# Navigate to chapter
cd 001_hello_world

# Run directly
go run main.go

# Or compile and run
go build -o program
./program
```

## Code Conventions

- **Early return**: Always prefer early returns over nesting
- **Error handling**: Handle errors immediately after the call
- **Naming**: Short and descriptive names (idiomatic Go)
- **Comments**: Only when code is not self-explanatory

## Go Versions Covered

| Version | Release | Key Features |
|---------|---------|-------------|
| Go 1.23 | Aug 2024 | Range over func, unique package |
| Go 1.24 | Feb 2025 | Swiss Tables, generic aliases, FIPS 140-3 |
| Go 1.25 | Aug 2025 | WaitGroup.Go(), JSON v2, synctest |
| Go 1.26 | Feb 2026 | SIMD experimental, crypto/hpke |

## Coverage

- **Language**: Variables, types, functions, pointers, structs, interfaces, generics, errors, concurrency
- **Standard Library**: 100+ packages covered (io, fmt, strings, bytes, strconv, time, regexp, encoding, crypto, net, os, hash, math, archive, compress, text, mime, testing, runtime, debug, log/slog, container, sort, slices, maps, cmp, flag, path, filepath, unicode, plugin, expvar, go/ast)
- **golang.org/x/**: sync, net, crypto, text, oauth2, term, sys, mod, tools
- **Web**: HTTP server/client, gRPC, WebSockets, GraphQL, REST APIs
- **Data**: databases, Redis, Kafka, NATS, RabbitMQ, CSV, JSON, XML, Parquet
- **Apps**: CLI, desktop, AI/ML, images, audio, data processing, WASM
- **Architecture**: Design patterns, DI, CQRS, event sourcing, resilience
- **Security**: OWASP, JWT, OAuth2, OIDC, crypto, TLS/mTLS
- **Production**: Performance, profiling, observability, deploy, CI/CD
- **Major Projects**: Docker, Kubernetes, etcd, Terraform, Prometheus, Grafana, Caddy, Hugo, Traefik, CoreDNS, Consul, Vault, CockroachDB, NATS, Minio, Gitea, Badger/BoltDB, Cilium

## Additional Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Go Blog](https://go.dev/blog)
- [Go Playground](https://go.dev/play)

## License

MIT License - Free for educational and commercial use.
