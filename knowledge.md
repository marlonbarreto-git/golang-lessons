# Knowledge - Curso Go de 0 a Ninja 2026

## Estado: COMPLETO (126 capítulos)
- **Fecha**: 2026-02-06
- **Autor**: Marlon Barreto <mbarretot@hotmail.com>

## Verificación
- `go build -o /dev/null ./...` → ALL PASS (126/126)
- `go vet ./...` → 4 warnings intencionales (002, 003, 024)
- Sin Claude co-author

## Estructura (126 capítulos, 18 partes)

### Part I: Language Basics (001-010)
001-hello_world, 002-variables_constants, 003-operators, 004-conditionals,
005-loops, 006-switch, 007-functions, 008-pointers, 009-strings, 010-collections

### Part II: Data Structures Deep Dive (011-015)
011-slices_deep_dive, 012-maps_deep_dive, 013-sort_package,
014-container_types, 015-slices_maps_cmp

### Part III: Types and OOP (016-022)
016-structs, 017-interfaces, 018-type_system, 019-enums_iota, 020-generics,
021-method_sets_receivers, 022-interface_internals

### Part IV: Error Handling (023-025)
023-errors, 024-panic_recover, 025-error_patterns

### Part V: Code Organization (026-029)
026-packages, 027-init_blank_identifier, 028-modules, 029-project_patterns

### Part VI: Concurrency (030-035)
030-goroutines, 031-channels, 032-select, 033-context, 034-sync,
035-concurrency_patterns

### Part VII: Testing (036-039)
036-testing, 037-benchmarks, 038-mocking, 039-fuzzing

### Part VIII: Standard Library Essentials (040-052)
040-io_patterns, 041-strings_strconv, 042-bytes_package, 043-time, 044-regexp,
045-encoding, 046-templates, 047-os_exec_processes, 048-flag_package,
049-path_filepath, 050-unicode_utf8, 051-hash_functions, 052-log_slog

### Part IX: Standard Library Advanced (053-061)
053-math_big_bits, 054-archive_compress, 055-crypto, 056-net_mail_smtp,
057-text_processing, 058-go_ast_parser, 059-testing_advanced,
060-testing_stdlib, 061-runtime_debug

### Part X: Modern Go Features (062-067)
062-iterators, 063-embed, 064-build_tags, 065-go124_features,
066-go125_features, 067-go126_features

### Part XI: Advanced Language (068-074)
068-reflection, 069-code_generation, 070-unsafe, 071-cgo, 072-go_assembly,
073-plugin_expvar, 074-wasm

### Part XII: Design Patterns & Architecture (075-079)
075-design_patterns, 076-resilience, 077-event_driven, 078-api_design, 079-security

### Part XIII: Web & Networking (080-088)
080-networking, 081-web_backend, 082-http_client, 083-websockets, 084-grpc,
085-graphql, 086-auth, 087-cli_applications, 088-desktop_apps

### Part XIV: Data, Databases & Messaging (089-093)
089-databases, 090-redis_caching, 091-message_queues, 092-data_processing,
093-images_audio

### Part XV: golang.org/x/ Ecosystem (094-099)
094-x_sync, 095-x_net_crypto, 096-x_text, 097-x_oauth2_term,
098-x_sys_mod, 099-x_tools

### Part XVI: Production & Deployment (100-106)
100-performance, 101-observability, 102-debugging_diagnostics, 103-cicd_tooling,
104-deploy, 105-ai_ml, 106-proyecto_final

### Part XVII: Major Go Projects (107-124)
107-docker, 108-kubernetes, 109-etcd, 110-terraform, 111-prometheus,
112-grafana, 113-caddy, 114-hugo, 115-traefik, 116-coredns, 117-consul,
118-vault, 119-cockroachdb, 120-nats, 121-minio, 122-gitea,
123-badger_boltdb, 124-cilium

### Part XVIII: Philosophy & Mastery (125-126)
125-go_proverbs_mistakes, 126-go_mastery_path

## Cobertura Total
- Core language: variables, types, functions, pointers, structs, interfaces, generics, errors
- Data structures: slices deep dive, maps deep dive, sort, container/heap/list/ring, slices/maps/cmp
- Method sets, receivers, interface internals
- Concurrency: goroutines, channels, select, context, sync, atomic, patterns
- Testing: unit, benchmarks, mocking, fuzzing, testify, gomock, integration, testing stdlib
- Stdlib essentials: fmt, strings, strconv, bytes, io, bufio, os, os/exec, flag, path/filepath,
  unicode/utf8, hash/fnv/crc32/maphash, log/slog, time, regexp, encoding/*, templates
- Stdlib advanced: math/big/bits, archive/compress, crypto/*, net/mail, text/scanner/tabwriter,
  go/ast/parser/token, testing/quick/iotest/fstest, runtime/debug
- Modern Go: iterators, embed, build tags, Go 1.24/1.25/1.26 features
- Advanced: reflection, code generation, unsafe, cgo, go assembly, plugin, expvar, wasm
- Web: HTTP server/client, gRPC, WebSockets, GraphQL, REST APIs, networking
- Data: databases, Redis, Kafka, NATS, RabbitMQ, CSV, JSON, XML, Parquet, images, audio
- Apps: CLI, desktop, AI/ML, data processing
- Architecture: design patterns, DI, CQRS, event sourcing, resilience, event-driven, API design
- Security: OWASP, JWT, OAuth2, OIDC, crypto, TLS/mTLS
- Production: performance, profiling, observability, deploy, CI/CD, debugging, diagnostics
- golang.org/x/: sync, net, crypto, text, oauth2, term, sys, mod, tools
- Major Go projects: Docker, Kubernetes, etcd, Terraform, Prometheus, Grafana, Caddy, Hugo,
  Traefik, CoreDNS, Consul, Vault, CockroachDB, NATS, Minio, Gitea, Badger/BoltDB, Cilium
- Philosophy: Go proverbs, common mistakes, mastery path, career paths, community
- Go versions: 1.23, 1.24, 1.25, 1.26
