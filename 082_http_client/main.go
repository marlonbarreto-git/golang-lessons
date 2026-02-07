// Package main - Chapter 082: HTTP Client
// Domina el cliente HTTP en Go: configuración avanzada, connection pooling,
// retries, circuit breaker, middleware, testing y mejores prácticas.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== HTTP CLIENT MASTERY EN GO ===")

	// ============================================
	// DEFAULT CLIENT - POR QUÉ NO USARLO
	// ============================================
	fmt.Println("\n--- http.DefaultClient y por qué NO usarlo en producción ---")
	fmt.Println(`
http.DefaultClient es el cliente HTTP global de Go:

var DefaultClient = &Client{}

PROBLEMA: NO tiene timeout configurado.
Si el servidor no responde, tu goroutine se bloquea PARA SIEMPRE.

// PELIGROSO - No usar en producción ❌
resp, err := http.Get("https://api.example.com/data")
resp, err := http.Post("https://api.example.com/data", "application/json", body)
resp, err := http.PostForm("https://api.example.com/form", url.Values{})

Estas funciones usan http.DefaultClient internamente.

OTROS PROBLEMAS:
- No tiene límites de conexiones
- No configura timeouts de TLS
- No controla keep-alive
- Compartido por todo el programa (race conditions si se modifica)
- No permite personalizar redirect policy

REGLA: SIEMPRE crear tu propio http.Client.`)

	// ============================================
	// CUSTOM CLIENT CON TIMEOUTS
	// ============================================
	fmt.Println("\n--- Custom http.Client con Timeouts ---")
	fmt.Println(`
ANATOMÍA DE UN TIMEOUT HTTP:

       DNS        TCP        TLS        Request     Response
     Lookup     Connect    Handshake    Headers      Body
    |---------|---------|------------|-----------|------------|
    <- - - - - - - DialContext - - ->
                          <-TLSHandshake->
                                     <-ResponseHeader->
                                         <-ExpectContinue->
    <- - - - - - - - - - - client.Timeout - - - - - - - - ->

CONFIGURACIÓN RECOMENDADA:

client := &http.Client{
    Timeout: 30 * time.Second,  // Timeout total (incluye body)
    Transport: &http.Transport{
        // DNS + TCP connect
        DialContext: (&net.Dialer{
            Timeout:   5 * time.Second,  // Timeout de conexión
            KeepAlive: 30 * time.Second, // Intervalo keep-alive
        }).DialContext,

        // TLS
        TLSHandshakeTimeout: 5 * time.Second,

        // Response
        ResponseHeaderTimeout: 10 * time.Second,  // Esperar headers
        ExpectContinueTimeout: 1 * time.Second,    // Esperar 100-continue

        // Idle
        IdleConnTimeout: 90 * time.Second,
    },
}

TIMEOUT TOTAL vs PER-REQUEST:

// client.Timeout: Timeout total para toda la operación
// Incluye: dial + TLS + request + response headers + response body
client := &http.Client{Timeout: 30 * time.Second}

// Per-request timeout con context (más flexible):
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
// Si el context expira, la request se cancela`)

	// ============================================
	// TRANSPORT CONFIGURATION
	// ============================================
	fmt.Println("\n--- http.Transport Configuration ---")
	fmt.Println(`
http.Transport controla la capa de transporte HTTP.

transport := &http.Transport{
    // === CONNECTION POOL ===

    // Máximo de conexiones idle en el pool (todas los hosts)
    MaxIdleConns: 100,  // Default: 100

    // Máximo de conexiones idle por host
    MaxIdleConnsPerHost: 10,  // Default: 2 (¡muy bajo!)

    // Máximo de conexiones totales por host (idle + activas)
    // 0 = sin límite
    MaxConnsPerHost: 20,

    // Tiempo máximo que una conexión idle permanece en el pool
    IdleConnTimeout: 90 * time.Second,  // Default: 90s

    // === TIMEOUTS ===

    // Timeout para TLS handshake
    TLSHandshakeTimeout: 10 * time.Second,  // Default: 10s

    // Timeout para esperar response headers
    ResponseHeaderTimeout: 10 * time.Second,  // Default: 0 (sin límite)

    // Timeout para esperar 100 Continue del servidor
    ExpectContinueTimeout: 1 * time.Second,  // Default: 1s

    // === BEHAVIOR ===

    // Deshabilitar HTTP keep-alive (una conexión por request)
    DisableKeepAlives: false,  // Default: false

    // Deshabilitar compresión (Accept-Encoding: gzip)
    DisableCompression: false,  // Default: false

    // Forzar HTTP/2 con h2c (sin TLS)
    ForceAttemptHTTP2: true,  // Default: true en Go 1.13+

    // === PROXY ===

    // Usar proxy del entorno (HTTP_PROXY, HTTPS_PROXY, NO_PROXY)
    Proxy: http.ProxyFromEnvironment,

    // Proxy fijo
    // Proxy: http.ProxyURL(proxyURL),
}

POR QUÉ MaxIdleConnsPerHost ES CRÍTICO:

Default es 2, lo que significa que si haces 100 requests concurrentes
al mismo host, solo 2 conexiones se reutilizan. Las otras 98 se crean
y destruyen cada vez.

Para microservicios que llaman mucho a un API:
MaxIdleConnsPerHost: 100  // Ajustar según carga`)

	// ============================================
	// CONNECTION POOLING Y KEEP-ALIVE
	// ============================================
	fmt.Println("\n--- Connection Pooling y Keep-Alive ---")
	fmt.Println(`
Go reutiliza conexiones TCP automáticamente si:
1. Lees el response body COMPLETAMENTE
2. Cierras el response body
3. DisableKeepAlives es false (default)

// CORRECTO - Conexión se reutiliza ✅
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()

// Leer todo el body (necesario para reutilizar conexión)
body, err := io.ReadAll(resp.Body)

// INCORRECTO - Conexión NO se reutiliza ❌
resp, err := client.Do(req)
if err != nil {
    return err
}
// Olvidaste cerrar el body
// Olvidaste leer el body

// SI NO TE INTERESA EL BODY:
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
io.Copy(io.Discard, resp.Body)  // Drenar body para reutilizar conexión

VERIFICAR POOL:

transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 100,
    MaxConnsPerHost:     100,
    IdleConnTimeout:     90 * time.Second,
}

client := &http.Client{Transport: transport}

// Después de N requests, el pool tiene conexiones listas
// Puedes monitorear con transport.TLSClientConfig para debug`)

	// ============================================
	// HTTP/2 SUPPORT
	// ============================================
	fmt.Println("\n--- HTTP/2 Support ---")
	fmt.Println(`
HTTP/2 se habilita AUTOMÁTICAMENTE cuando:
1. Usas TLS (https://)
2. No has personalizado TLSClientConfig de forma incompatible
3. ForceAttemptHTTP2 es true (default desde Go 1.13)

// HTTP/2 automático con TLS ✅
client := &http.Client{}
resp, _ := client.Get("https://http2.golang.org/reqinfo")

// Forzar HTTP/2 sin TLS (h2c):
import "golang.org/x/net/http2"

transport := &http2.Transport{
    AllowHTTP: true,
    DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
        var d net.Dialer
        return d.DialContext(ctx, network, addr)
    },
}
client := &http.Client{Transport: transport}

// Verificar protocolo usado:
resp, _ := client.Get("https://example.com")
fmt.Println(resp.Proto)  // "HTTP/2.0"

VENTAJAS HTTP/2:
- Multiplexing: múltiples requests en una conexión TCP
- Header compression (HPACK)
- Server push
- Binary framing (más eficiente que HTTP/1.1 text)
- Stream prioritization`)

	// ============================================
	// MAKING REQUESTS
	// ============================================
	fmt.Println("\n--- Making Requests: Get, Post, PostForm, Do ---")
	fmt.Println(`
FUNCIONES DE CONVENIENCIA (sobre un client personalizado):

client := &http.Client{Timeout: 30 * time.Second}

// GET
resp, err := client.Get("https://api.example.com/users")

// POST con JSON
body := strings.NewReader(` + "`" + `{"name":"Alice"}` + "`" + `)
resp, err := client.Post("https://api.example.com/users", "application/json", body)

// POST Form
resp, err := client.PostForm("https://api.example.com/login", url.Values{
    "username": {"alice"},
    "password": {"secret123"},
})

// HEAD
resp, err := client.Head("https://api.example.com/health")

DO (máximo control):

req, err := http.NewRequest("PUT", "https://api.example.com/users/1", body)
if err != nil {
    return err
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer token123")

resp, err := client.Do(req)

CUÁNDO USAR CADA UNO:
- Get/Post/PostForm: requests simples, prototipos
- Do: requests con headers custom, métodos distintos, contexts`)

	// ============================================
	// REQUEST CREATION WITH CONTEXT
	// ============================================
	fmt.Println("\n--- http.NewRequestWithContext ---")
	os.Stdout.WriteString(`
SIEMPRE usar NewRequestWithContext en producción:

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/data", nil)
if err != nil {
    return fmt.Errorf("creating request: %w", err)
}

resp, err := client.Do(req)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("request timed out")
    }
    if ctx.Err() == context.Canceled {
        return fmt.Errorf("request canceled")
    }
    return fmt.Errorf("request failed: %w", err)
}
defer resp.Body.Close()

CON CONTEXT DE HTTP HANDLER:

func handler(w http.ResponseWriter, r *http.Request) {
    // Reutilizar context del request entrante
    ctx := r.Context()

    req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.example.com", nil)
    resp, err := client.Do(req)
    // Si el cliente HTTP cierra la conexión, ctx se cancela automáticamente
}

CANCELACIÓN MANUAL:

ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(2 * time.Second)
    cancel()  // Cancelar request después de 2 segundos
}()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
// err contendrá context.Canceled si se canceló` + "\n")

	// ============================================
	// HEADERS, QUERY PARAMS, COOKIES
	// ============================================
	fmt.Println("\n--- Setting Headers, Query Params, Cookies ---")
	fmt.Println(`
HEADERS:

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

// Set reemplaza el valor
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)
req.Header.Set("Accept", "application/json")
req.Header.Set("User-Agent", "MyApp/1.0")

// Add agrega (puede haber múltiples valores)
req.Header.Add("Accept-Encoding", "gzip")
req.Header.Add("Accept-Encoding", "deflate")

// Get
contentType := req.Header.Get("Content-Type")

// Del
req.Header.Del("User-Agent")

QUERY PARAMS:

// Opción 1: Construir URL manualmente
baseURL := "https://api.example.com/search"
params := url.Values{}
params.Set("q", "golang http client")
params.Set("page", "1")
params.Set("per_page", "20")
params.Set("sort", "relevance")

fullURL := baseURL + "?" + params.Encode()
// https://api.example.com/search?page=1&per_page=20&q=golang+http+client&sort=relevance

// Opción 2: Parsear y modificar URL
u, _ := url.Parse("https://api.example.com/search")
q := u.Query()
q.Set("q", "golang")
q.Set("page", "1")
u.RawQuery = q.Encode()

req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)

COOKIES:

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

// Agregar cookie individual
req.AddCookie(&http.Cookie{
    Name:  "session_id",
    Value: "abc123",
})

// Cookie jar para manejo automático de cookies
import "net/http/cookiejar"

jar, _ := cookiejar.New(nil)
client := &http.Client{
    Jar:     jar,
    Timeout: 30 * time.Second,
}
// Las cookies se envían y reciben automáticamente

BASIC AUTH:

req.SetBasicAuth("username", "password")
// Equivale a: req.Header.Set("Authorization", "Basic dXNlcm5hbWU6cGFzc3dvcmQ=")`)

	// ============================================
	// READING RESPONSES
	// ============================================
	fmt.Println("\n--- Reading Responses (Always Close Body) ---")
	os.Stdout.WriteString(`
PATRÓN CORRECTO:

resp, err := client.Do(req)
if err != nil {
    return fmt.Errorf("request failed: %w", err)
}
defer resp.Body.Close()  // SIEMPRE cerrar el body

// Verificar status code
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)
    return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
}

// Leer body
body, err := io.ReadAll(resp.Body)
if err != nil {
    return fmt.Errorf("reading body: %w", err)
}

LIMITAR TAMAÑO DEL BODY:

// Prevenir ataques de body gigante
maxBytes := int64(10 * 1024 * 1024)  // 10 MB
limitedReader := io.LimitReader(resp.Body, maxBytes)
body, err := io.ReadAll(limitedReader)

LEER HEADERS DE RESPUESTA:

contentType := resp.Header.Get("Content-Type")
rateLimit := resp.Header.Get("X-RateLimit-Remaining")
etag := resp.Header.Get("ETag")

// Todas las headers
for key, values := range resp.Header {
    fmt.Printf("%s: %s\n", key, strings.Join(values, ", "))
}

RESPONSE METADATA:

fmt.Println(resp.StatusCode)     // 200
fmt.Println(resp.Status)         // "200 OK"
fmt.Println(resp.Proto)          // "HTTP/2.0"
fmt.Println(resp.ContentLength)  // -1 si chunked
fmt.Println(resp.Uncompressed)   // true si se descomprimió

COOKIES DE RESPUESTA:

for _, cookie := range resp.Cookies() {
    fmt.Printf("Cookie: %s = %s\n", cookie.Name, cookie.Value)
}
`)

	// ============================================
	// JSON REQUEST/RESPONSE PATTERNS
	// ============================================
	fmt.Println("\n--- JSON Request/Response Patterns ---")
	os.Stdout.WriteString(`
ENVIAR JSON:

type CreateUserRequest struct {
    Name  string ` + "`" + `json:"name"` + "`" + `
    Email string ` + "`" + `json:"email"` + "`" + `
}

func CreateUser(ctx context.Context, client *http.Client, user CreateUserRequest) error {
    body, err := json.Marshal(user)
    if err != nil {
        return fmt.Errorf("marshaling: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", "https://api.example.com/users", bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("sending request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("status %d: %s", resp.StatusCode, body)
    }

    return nil
}

RECIBIR JSON:

type User struct {
    ID    string ` + "`" + `json:"id"` + "`" + `
    Name  string ` + "`" + `json:"name"` + "`" + `
    Email string ` + "`" + `json:"email"` + "`" + `
}

func GetUser(ctx context.Context, client *http.Client, id string) (*User, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/users/"+id, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("status: %d", resp.StatusCode)
    }

    var user User
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, fmt.Errorf("decoding: %w", err)
    }

    return &user, nil
}

json.NewDecoder vs json.Unmarshal:

// Decoder: streaming, no carga todo en memoria
json.NewDecoder(resp.Body).Decode(&result)

// Unmarshal: todo en memoria, útil si necesitas el []byte raw
body, _ := io.ReadAll(resp.Body)
json.Unmarshal(body, &result)

// Regla: Decoder para HTTP responses, Unmarshal si necesitas el raw body

GENERIC HTTP HELPER:

func doJSON[T any](ctx context.Context, client *http.Client, method, url string, reqBody any) (*T, error) {
    var bodyReader io.Reader
    if reqBody != nil {
        data, err := json.Marshal(reqBody)
        if err != nil {
            return nil, err
        }
        bodyReader = bytes.NewReader(data)
    }

    req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
    }

    var result T
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// Uso:
// user, err := doJSON[User](ctx, client, "GET", "/users/1", nil)
// created, err := doJSON[User](ctx, client, "POST", "/users", newUser)
`)

	// ============================================
	// FORM DATA Y MULTIPART UPLOADS
	// ============================================
	fmt.Println("\n--- Form Data y Multipart Uploads ---")
	fmt.Println(`
URL-ENCODED FORM:

data := url.Values{
    "username": {"alice"},
    "password": {"secret123"},
}

req, _ := http.NewRequestWithContext(ctx, "POST", url,
    strings.NewReader(data.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

resp, err := client.Do(req)

MULTIPART FORM (File Upload):

func UploadFile(ctx context.Context, client *http.Client, url string, fileContent []byte, filename string) error {
    var buf bytes.Buffer
    writer := multipart.NewWriter(&buf)

    // Agregar archivo
    part, err := writer.CreateFormFile("file", filename)
    if err != nil {
        return err
    }
    part.Write(fileContent)

    // Agregar campos extra
    writer.WriteField("description", "My file upload")
    writer.WriteField("tags", "go,http")

    writer.Close()

    req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

MULTIPART CON ARCHIVO GRANDE (streaming):

func UploadLargeFile(ctx context.Context, client *http.Client, url string, file io.Reader, filename string) error {
    pr, pw := io.Pipe()

    writer := multipart.NewWriter(pw)

    go func() {
        defer pw.Close()
        part, _ := writer.CreateFormFile("file", filename)
        io.Copy(part, file)
        writer.Close()
    }()

    req, _ := http.NewRequestWithContext(ctx, "POST", url, pr)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}`)

	// ============================================
	// FILE DOWNLOADS WITH PROGRESS
	// ============================================
	fmt.Println("\n--- File Downloads with Progress ---")
	os.Stdout.WriteString(`
DOWNLOAD CON PROGRESS:

type ProgressReader struct {
    reader     io.Reader
    total      int64
    downloaded int64
    onProgress func(downloaded, total int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
    n, err := pr.reader.Read(p)
    pr.downloaded += int64(n)
    if pr.onProgress != nil {
        pr.onProgress(pr.downloaded, pr.total)
    }
    return n, err
}

func DownloadFile(ctx context.Context, client *http.Client, url, destPath string) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    file, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer file.Close()

    progress := &ProgressReader{
        reader: resp.Body,
        total:  resp.ContentLength,
        onProgress: func(downloaded, total int64) {
            if total > 0 {
                pct := float64(downloaded) / float64(total) * 100
                fmt.Printf("\rDownloading: %.1f%%", pct)
            }
        },
    }

    _, err = io.Copy(file, progress)
    fmt.Println()
    return err
}

DOWNLOAD CON RESUME:

func ResumeDownload(ctx context.Context, client *http.Client, url, destPath string) error {
    // Verificar si archivo parcial existe
    var startByte int64
    if info, err := os.Stat(destPath); err == nil {
        startByte = info.Size()
    }

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    if startByte > 0 {
        req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
    }

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 206 = Partial Content, 200 = servidor no soporta Range
    flag := os.O_CREATE | os.O_WRONLY
    if resp.StatusCode == http.StatusPartialContent {
        flag |= os.O_APPEND
    } else {
        flag |= os.O_TRUNC
        startByte = 0
    }

    file, err := os.OpenFile(destPath, flag, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    _, err = io.Copy(file, resp.Body)
    return err
}
`)

	// ============================================
	// RETRY WITH EXPONENTIAL BACKOFF
	// ============================================
	fmt.Println("\n--- Retry Patterns con Exponential Backoff ---")
	os.Stdout.WriteString(`
RETRY BÁSICO:

func DoWithRetry(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
    var resp *http.Response
    var err error

    for attempt := 0; attempt <= maxRetries; attempt++ {
        // Clonar request (body puede haberse consumido)
        clonedReq := req.Clone(req.Context())
        if req.Body != nil {
            // Necesitas un body que se pueda releer
            // Usar bytes.NewReader o GetBody
        }

        resp, err = client.Do(clonedReq)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }

        if resp != nil {
            io.Copy(io.Discard, resp.Body)
            resp.Body.Close()
        }

        if attempt < maxRetries {
            // Exponential backoff: 1s, 2s, 4s, 8s...
            backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
            time.Sleep(backoff)
        }
    }

    return resp, fmt.Errorf("max retries exceeded: %w", err)
}

RETRY CON JITTER:

func backoffWithJitter(attempt int) time.Duration {
    base := time.Duration(math.Pow(2, float64(attempt))) * time.Second
    jitter := time.Duration(rand.Int63n(int64(base / 2)))
    return base + jitter
}

RETRY INTELIGENTE (solo para errores recuperables):

func isRetryable(resp *http.Response, err error) bool {
    if err != nil {
        // Network errors son retryable
        return true
    }
    // Solo reintentar errores del servidor
    switch resp.StatusCode {
    case http.StatusTooManyRequests,       // 429
         http.StatusServiceUnavailable,    // 503
         http.StatusGatewayTimeout,        // 504
         http.StatusBadGateway:            // 502
        return true
    }
    return false
}

// Respetar Retry-After header
func getRetryAfter(resp *http.Response) time.Duration {
    if ra := resp.Header.Get("Retry-After"); ra != "" {
        if seconds, err := strconv.Atoi(ra); err == nil {
            return time.Duration(seconds) * time.Second
        }
        if t, err := http.ParseTime(ra); err == nil {
            return time.Until(t)
        }
    }
    return 0
}

BODY REPLAYABLE:

func newRetryableRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
    req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }

    // GetBody permite al client releer el body en redirects y retries
    req.GetBody = func() (io.ReadCloser, error) {
        return io.NopCloser(bytes.NewReader(body)), nil
    }

    return req, nil
}` + "\n")

	// ============================================
	// CIRCUIT BREAKER
	// ============================================
	fmt.Println("\n--- Circuit Breaker con HTTP Client ---")
	fmt.Println(`
Estados:  CLOSED -> OPEN -> HALF-OPEN -> CLOSED

CLOSED:    Requests pasan normalmente
           Si fallos >= threshold -> OPEN

OPEN:      Requests fallan inmediatamente
           Después de timeout -> HALF-OPEN

HALF-OPEN: Permite 1 request de prueba
           Si éxito -> CLOSED
           Si fallo -> OPEN

IMPLEMENTACIÓN:

type CircuitBreaker struct {
    mu           sync.Mutex
    state        string  // "closed", "open", "half-open"
    failures     int
    threshold    int
    timeout      time.Duration
    lastFailure  time.Time
    successCount int
    halfOpenMax  int
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:       "closed",
        threshold:   threshold,
        timeout:     timeout,
        halfOpenMax: 1,
    }
}

func (cb *CircuitBreaker) Allow() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case "closed":
        return true
    case "open":
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = "half-open"
            cb.successCount = 0
            return true
        }
        return false
    case "half-open":
        return cb.successCount < cb.halfOpenMax
    }
    return false
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case "half-open":
        cb.successCount++
        if cb.successCount >= cb.halfOpenMax {
            cb.state = "closed"
            cb.failures = 0
        }
    case "closed":
        cb.failures = 0
    }
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures++
    cb.lastFailure = time.Now()

    if cb.failures >= cb.threshold {
        cb.state = "open"
    }
}

USO CON HTTP CLIENT:

func DoWithCircuitBreaker(cb *CircuitBreaker, client *http.Client, req *http.Request) (*http.Response, error) {
    if !cb.Allow() {
        return nil, errors.New("circuit breaker is open")
    }

    resp, err := client.Do(req)
    if err != nil || resp.StatusCode >= 500 {
        cb.RecordFailure()
        return resp, err
    }

    cb.RecordSuccess()
    return resp, nil
}

LIBRERÍA sony/gobreaker:

import "github.com/sony/gobreaker/v2"

cb := gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
    Name:        "api-client",
    MaxRequests: 3,                    // Requests en half-open
    Interval:    10 * time.Second,     // Reset counter interval
    Timeout:     30 * time.Second,     // Open -> half-open
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.ConsecutiveFailures > 5
    },
})

resp, err := cb.Execute(func() (*http.Response, error) {
    return client.Do(req)
})`)

	// ============================================
	// CUSTOM ROUND TRIPPER (MIDDLEWARE)
	// ============================================
	fmt.Println("\n--- Request Middleware / Custom RoundTripper ---")
	fmt.Println(`
http.RoundTripper es la interfaz que ejecuta HTTP requests:

type RoundTripper interface {
    RoundTrip(*http.Request) (*http.Response, error)
}

LOGGING MIDDLEWARE:

type LoggingTransport struct {
    Transport http.RoundTripper
    Logger    *slog.Logger
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    start := time.Now()

    t.Logger.Info("HTTP request",
        "method", req.Method,
        "url", req.URL.String(),
    )

    resp, err := t.Transport.RoundTrip(req)
    if err != nil {
        t.Logger.Error("HTTP request failed",
            "method", req.Method,
            "url", req.URL.String(),
            "error", err,
            "duration", time.Since(start),
        )
        return nil, err
    }

    t.Logger.Info("HTTP response",
        "method", req.Method,
        "url", req.URL.String(),
        "status", resp.StatusCode,
        "duration", time.Since(start),
    )

    return resp, nil
}

AUTH MIDDLEWARE:

type AuthTransport struct {
    Token     string
    Transport http.RoundTripper
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    req = req.Clone(req.Context())  // No modificar request original
    req.Header.Set("Authorization", "Bearer "+t.Token)
    return t.Transport.RoundTrip(req)
}

HEADER MIDDLEWARE:

type HeaderTransport struct {
    Headers   map[string]string
    Transport http.RoundTripper
}

func (t *HeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    req = req.Clone(req.Context())
    for k, v := range t.Headers {
        req.Header.Set(k, v)
    }
    return t.Transport.RoundTrip(req)
}

ENCADENAR MIDDLEWARES:

client := &http.Client{
    Transport: &LoggingTransport{
        Logger: slog.Default(),
        Transport: &AuthTransport{
            Token: "my-token",
            Transport: &HeaderTransport{
                Headers: map[string]string{
                    "User-Agent": "MyApp/1.0",
                    "Accept":     "application/json",
                },
                Transport: http.DefaultTransport,
            },
        },
    },
}

RETRY MIDDLEWARE:

type RetryTransport struct {
    Transport  http.RoundTripper
    MaxRetries int
    Backoff    func(attempt int) time.Duration
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    var resp *http.Response
    var err error

    for attempt := 0; attempt <= t.MaxRetries; attempt++ {
        resp, err = t.Transport.RoundTrip(req)
        if err == nil && resp.StatusCode < 500 {
            return resp, nil
        }
        if resp != nil {
            io.Copy(io.Discard, resp.Body)
            resp.Body.Close()
        }
        if attempt < t.MaxRetries {
            select {
            case <-req.Context().Done():
                return nil, req.Context().Err()
            case <-time.After(t.Backoff(attempt)):
            }
        }
    }
    return resp, err
}`)

	// ============================================
	// CONTEXT INTEGRATION
	// ============================================
	fmt.Println("\n--- Context Integration (Cancellation, Per-Request Timeouts) ---")
	fmt.Println(`
PER-REQUEST TIMEOUT:

func fetchWithTimeout(client *http.Client, url string, timeout time.Duration) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

PARALLEL REQUESTS CON CANCELACIÓN:

func fetchAll(ctx context.Context, client *http.Client, urls []string) ([][]byte, error) {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    results := make([][]byte, len(urls))
    errs := make([]error, len(urls))
    var wg sync.WaitGroup

    for i, u := range urls {
        wg.Add(1)
        go func(idx int, url string) {
            defer wg.Done()
            req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
            if err != nil {
                errs[idx] = err
                cancel()  // Cancelar otras requests
                return
            }
            resp, err := client.Do(req)
            if err != nil {
                errs[idx] = err
                cancel()
                return
            }
            defer resp.Body.Close()
            results[idx], errs[idx] = io.ReadAll(resp.Body)
        }(i, u)
    }

    wg.Wait()

    for _, err := range errs {
        if err != nil {
            return nil, err
        }
    }
    return results, nil
}

CONTEXT VALUES PARA TRACING:

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey{}, id)
}

type TracingTransport struct {
    Transport http.RoundTripper
}

func (t *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    req = req.Clone(req.Context())
    if id, ok := req.Context().Value(requestIDKey{}).(string); ok {
        req.Header.Set("X-Request-ID", id)
    }
    return t.Transport.RoundTrip(req)
}`)

	// ============================================
	// TESTING HTTP CLIENTS
	// ============================================
	fmt.Println("\n--- Testing HTTP Clients (httptest.NewServer) ---")
	os.Stdout.WriteString(`
TEST SERVER BÁSICO:

func TestGetUser(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verificar request
        if r.Method != "GET" {
            t.Errorf("expected GET, got %s", r.Method)
        }
        if r.URL.Path != "/users/123" {
            t.Errorf("expected /users/123, got %s", r.URL.Path)
        }
        if r.Header.Get("Authorization") != "Bearer test-token" {
            t.Error("missing auth header")
        }

        // Enviar response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(User{ID: "123", Name: "Alice"})
    }))
    defer server.Close()

    client := server.Client()
    user, err := GetUser(context.Background(), client, server.URL, "123")

    require.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}

SIMULAR ERRORES:

func TestGetUser_ServerError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("internal error"))
    }))
    defer server.Close()

    _, err := GetUser(context.Background(), server.Client(), server.URL, "123")
    assert.Error(t, err)
}

SIMULAR TIMEOUTS:

func TestGetUser_Timeout(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(5 * time.Second)  // Simular servidor lento
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    _, err := GetUser(ctx, server.Client(), server.URL, "123")
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}

TEST TLS:

server := httptest.NewTLSServer(handler)
defer server.Close()
client := server.Client()  // Ya configurado con TLS del test server

RECORDER (sin levantar servidor):

func TestHandler(t *testing.T) {
    recorder := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/users/123", nil)

    handler.ServeHTTP(recorder, req)

    assert.Equal(t, http.StatusOK, recorder.Code)
    assert.Contains(t, recorder.Body.String(), "Alice")
}

MOCK DINÁMICO:

func TestRetries(t *testing.T) {
    attempts := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 3 {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("success"))
    }))
    defer server.Close()

    // Test que el retry funciona
    resp, err := DoWithRetry(server.Client(), server.URL, 5)
    assert.NoError(t, err)
    assert.Equal(t, "success", resp)
    assert.Equal(t, 3, attempts)
}
`)

	// ============================================
	// RATE LIMITING OUTBOUND
	// ============================================
	fmt.Println("\n--- Rate Limiting Outbound Requests ---")
	os.Stdout.WriteString(`
USANDO golang.org/x/time/rate:

import "golang.org/x/time/rate"

type RateLimitedTransport struct {
    Transport http.RoundTripper
    Limiter   *rate.Limiter
}

func (t *RateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Esperar hasta que haya un token disponible
    if err := t.Limiter.Wait(req.Context()); err != nil {
        return nil, fmt.Errorf("rate limit: %w", err)
    }
    return t.Transport.RoundTrip(req)
}

// Uso: 10 requests por segundo, burst de 5
client := &http.Client{
    Transport: &RateLimitedTransport{
        Transport: http.DefaultTransport,
        Limiter:   rate.NewLimiter(rate.Limit(10), 5),
    },
}

SEMÁFORO PARA CONCURRENCIA:

type ConcurrencyLimitTransport struct {
    Transport http.RoundTripper
    sem       chan struct{}
}

func NewConcurrencyLimitTransport(maxConcurrent int, transport http.RoundTripper) *ConcurrencyLimitTransport {
    return &ConcurrencyLimitTransport{
        Transport: transport,
        sem:       make(chan struct{}, maxConcurrent),
    }
}

func (t *ConcurrencyLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    select {
    case t.sem <- struct{}{}:
        defer func() { <-t.sem }()
        return t.Transport.RoundTrip(req)
    case <-req.Context().Done():
        return nil, req.Context().Err()
    }
}

// Máximo 20 requests concurrentes
client := &http.Client{
    Transport: NewConcurrencyLimitTransport(20, http.DefaultTransport),
}

PER-HOST RATE LIMITING:

type PerHostRateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func (rl *PerHostRateLimiter) GetLimiter(host string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    if limiter, ok := rl.limiters[host]; ok {
        return limiter
    }
    limiter := rate.NewLimiter(rl.rate, rl.burst)
    rl.limiters[host] = limiter
    return limiter
}` + "\n")

	// ============================================
	// PROXY CONFIGURATION
	// ============================================
	fmt.Println("\n--- Proxy Configuration ---")
	fmt.Println(`
PROXY DESDE ENTORNO:

// Usa HTTP_PROXY, HTTPS_PROXY, NO_PROXY
transport := &http.Transport{
    Proxy: http.ProxyFromEnvironment,
}

// Variables de entorno:
// HTTP_PROXY=http://proxy.company.com:8080
// HTTPS_PROXY=http://proxy.company.com:8443
// NO_PROXY=localhost,127.0.0.1,.internal.com

PROXY FIJO:

proxyURL, _ := url.Parse("http://proxy.company.com:8080")

transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
}

client := &http.Client{Transport: transport}

PROXY CON AUTENTICACIÓN:

proxyURL, _ := url.Parse("http://user:password@proxy.company.com:8080")

transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
}

SOCKS5 PROXY:

import "golang.org/x/net/proxy"

dialer, _ := proxy.SOCKS5("tcp", "localhost:1080", nil, proxy.Direct)

transport := &http.Transport{
    DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
        return dialer.Dial(network, addr)
    },
}

PROXY DINÁMICO:

transport := &http.Transport{
    Proxy: func(req *http.Request) (*url.URL, error) {
        if req.URL.Host == "api.internal.com" {
            return nil, nil  // Sin proxy
        }
        return url.Parse("http://proxy.company.com:8080")
    },
}`)

	// ============================================
	// TLS CONFIGURATION
	// ============================================
	fmt.Println("\n--- TLS Configuration ---")
	fmt.Println(`
CUSTOM CA (certificado corporativo):

import "crypto/tls"
import "crypto/x509"

// Cargar CA certificate
caCert, _ := os.ReadFile("/path/to/ca-cert.pem")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        RootCAs: caCertPool,
    },
}

client := &http.Client{Transport: transport}

MUTUAL TLS (mTLS):

cert, _ := tls.LoadX509KeyPair("client-cert.pem", "client-key.pem")

caCert, _ := os.ReadFile("ca-cert.pem")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
    },
}

SKIP VERIFY (solo desarrollo):

transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: true,  // ⚠️ NUNCA en producción
    },
}

// Mejor: agregar un flag
if os.Getenv("ENV") == "development" {
    transport.TLSClientConfig = &tls.Config{
        InsecureSkipVerify: true,
    }
}

TLS VERSION MÍNIMA:

transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
        MaxVersion: tls.VersionTLS13,
    },
}

CUSTOM CIPHER SUITES:

transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        },
    },
}`)

	// ============================================
	// BEST PRACTICES CHECKLIST
	// ============================================
	fmt.Println("\n--- HTTP Client Best Practices Checklist ---")
	fmt.Println(`
1. NUNCA usar http.DefaultClient en producción
2. SIEMPRE configurar Timeout en el client
3. SIEMPRE cerrar resp.Body (defer resp.Body.Close())
4. SIEMPRE drenar resp.Body para reutilizar conexiones
5. Usar NewRequestWithContext, NUNCA NewRequest
6. Configurar MaxIdleConnsPerHost según carga
7. Reutilizar el mismo http.Client (es goroutine-safe)
8. Implementar retries con exponential backoff + jitter
9. Agregar circuit breaker para servicios externos
10. Rate limit las requests salientes
11. Limitar tamaño de response body (io.LimitReader)
12. Configurar TLS mínimo 1.2
13. Usar context para cancelación y timeouts per-request
14. Loggear requests/responses con RoundTripper middleware
15. Verificar status code antes de leer body
16. Usar json.NewDecoder para responses (no ReadAll + Unmarshal)
17. No modificar requests después de pasarlas a Do()
18. Clone() requests si necesitas modificarlas en middleware
19. Configurar User-Agent descriptivo
20. Manejar redirects explícitamente si es necesario

REDIRECT POLICY:

client := &http.Client{
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        if len(via) >= 10 {
            return errors.New("too many redirects")
        }
        // Copiar headers del request original
        for key, values := range via[0].Header {
            req.Header[key] = values
        }
        return nil
    },
}

// Deshabilitar redirects:
client := &http.Client{
    CheckRedirect: func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    },
}

SINGLETON PATTERN:

var (
    httpClient     *http.Client
    httpClientOnce sync.Once
)

func GetHTTPClient() *http.Client {
    httpClientOnce.Do(func() {
        httpClient = &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 100,
                IdleConnTimeout:     90 * time.Second,
                TLSHandshakeTimeout: 10 * time.Second,
            },
        }
    })
    return httpClient
}`)

	// ============================================
	// POPULAR LIBRARIES
	// ============================================
	fmt.Println("\n--- Popular Libraries: resty, req, gentleman ---")
	os.Stdout.WriteString(`
RESTY (go-resty/resty):

go get github.com/go-resty/resty/v2

import "github.com/go-resty/resty/v2"

client := resty.New().
    SetTimeout(30 * time.Second).
    SetRetryCount(3).
    SetRetryWaitTime(1 * time.Second).
    SetRetryMaxWaitTime(10 * time.Second).
    SetHeader("Accept", "application/json").
    SetAuthToken("my-token")

// GET
resp, err := client.R().
    SetQueryParams(map[string]string{
        "page": "1",
        "size": "20",
    }).
    SetResult(&UserListResponse{}).
    Get("https://api.example.com/users")

users := resp.Result().(*UserListResponse)

// POST
resp, err := client.R().
    SetBody(CreateUserRequest{Name: "Alice", Email: "alice@example.com"}).
    SetResult(&User{}).
    Post("https://api.example.com/users")

// Upload
resp, err := client.R().
    SetFile("file", "/path/to/file.pdf").
    SetFormData(map[string]string{"description": "My file"}).
    Post("https://api.example.com/upload")

// Hooks
client.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
    fmt.Printf("Request: %s %s\n", r.Method, r.URL)
    return nil
})

client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
    fmt.Printf("Response: %d %s\n", r.StatusCode(), r.Time())
    return nil
})

REQ (imroc/req):

go get github.com/imroc/req/v3

import "github.com/imroc/req/v3"

client := req.C().
    SetTimeout(30 * time.Second).
    SetCommonRetryCount(3).
    SetCommonBearerAuthToken("my-token").
    EnableDumpAll()  // Debug

// GET
var user User
resp, err := client.R().
    SetSuccessResult(&user).
    Get("https://api.example.com/users/1")

// POST
resp, err := client.R().
    SetBody(&CreateUserRequest{Name: "Alice"}).
    SetSuccessResult(&user).
    Post("https://api.example.com/users")

// Download con progress
client.R().
    SetOutputFile("download.zip").
    SetDownloadCallbackWithInterval(func(info req.DownloadInfo) {
        fmt.Printf("%.2f%%\n", float64(info.DownloadedSize)/float64(info.Response.ContentLength)*100)
    }, 200*time.Millisecond).
    Get("https://example.com/large-file.zip")

GENTLEMAN:

go get gopkg.in/h2non/gentleman.v2

import "gopkg.in/h2non/gentleman.v2"

client := gentleman.New()
client.URL("https://api.example.com")
client.SetHeader("Authorization", "Bearer token")
client.Use(retry.New(retry.ConstantBackoff))

req := client.Request()
req.Path("/users")
req.SetQuery("page", "1")

resp, err := req.Send()

CUÁNDO USAR CADA UNO:
- net/http: máximo control, sin dependencias externas
- resty: API fluida, retries, lo más popular
- req: debug avanzado, dump requests, download progress
- gentleman: plugin system, middleware chains
`)

	// ============================================
	// DEMO EJECUTABLE
	// ============================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("DEMOS EJECUTABLES")
	fmt.Println(strings.Repeat("=", 60))

	demoBasicRequests()
	demoJSONRoundTrip()
	demoCustomRoundTripper()
	demoContextCancellation()
	demoHTTPTest()
	demoRetryPattern()
	demoCircuitBreaker()
	demoRateLimiting()
	demoMultipartUpload()
}

// --- DEMOS EJECUTABLES ---

func demoBasicRequests() {
	fmt.Println("\n--- Demo: Basic Requests ---")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Method: %s, Path: %s, UA: %s", r.Method, r.URL.Path, r.Header.Get("User-Agent"))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(server.URL + "/hello")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body: %s\n", body)

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/api", strings.NewReader("data"))
	req.Header.Set("User-Agent", "GoNinja/1.0")
	resp2, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("POST Response: %s\n", body2)
}

func demoJSONRoundTrip() {
	fmt.Println("\n--- Demo: JSON Request/Response ---")

	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var user User
			json.NewDecoder(r.Body).Decode(&user)
			user.ID = 42
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(User{ID: 1, Name: "Alice", Email: "alice@example.com"})
		}
	}))
	defer server.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	ctx := context.Background()

	newUser := User{Name: "Bob", Email: "bob@example.com"}
	data, _ := json.Marshal(newUser)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/users", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var created User
	json.NewDecoder(resp.Body).Decode(&created)
	fmt.Printf("Created user: ID=%d, Name=%s, Email=%s\n", created.ID, created.Name, created.Email)

	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/users/1", nil)
	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	var fetched User
	json.NewDecoder(resp2.Body).Decode(&fetched)
	fmt.Printf("Fetched user: ID=%d, Name=%s, Email=%s\n", fetched.ID, fetched.Name, fetched.Email)
}

type LoggingRoundTripper struct {
	next http.RoundTripper
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	fmt.Printf("  -> %s %s\n", req.Method, req.URL.Path)

	resp, err := l.next.RoundTrip(req)
	if err != nil {
		fmt.Printf("  <- ERROR: %v (%s)\n", err, time.Since(start))
		return nil, err
	}

	fmt.Printf("  <- %d %s (%s)\n", resp.StatusCode, resp.Status, time.Since(start))
	return resp, nil
}

type AuthRoundTripper struct {
	token string
	next  http.RoundTripper
}

func (a *AuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+a.token)
	return a.next.RoundTrip(req)
}

func demoCustomRoundTripper() {
	fmt.Println("\n--- Demo: Custom RoundTripper (Middleware Chain) ---")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		fmt.Fprintf(w, "Auth: %s", auth)
	}))
	defer server.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &LoggingRoundTripper{
			next: &AuthRoundTripper{
				token: "secret-ninja-token",
				next:  http.DefaultTransport,
			},
		},
	}

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/protected", nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", body)
}

func demoContextCancellation() {
	fmt.Println("\n--- Demo: Context Cancellation ---")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Second):
			w.Write([]byte("done"))
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	client := &http.Client{Timeout: 30 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/slow", nil)
	start := time.Now()
	_, err := client.Do(req)

	fmt.Printf("Request completed in %s\n", time.Since(start).Round(time.Millisecond))
	if err != nil {
		fmt.Printf("Error (expected): %v\n", ctx.Err())
	}
}

func demoHTTPTest() {
	fmt.Println("\n--- Demo: httptest.NewServer ---")

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"something broke"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := server.Client()
	ctx := context.Background()

	req1, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/health", nil)
	resp1, err := client.Do(req1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp1.Body.Close()
	body1, _ := io.ReadAll(resp1.Body)
	fmt.Printf("Health check: %d - %s\n", resp1.StatusCode, body1)

	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/error", nil)
	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Error endpoint: %d - %s\n", resp2.StatusCode, body2)

	req3, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/missing", nil)
	resp3, err := client.Do(req3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp3.Body.Close()
	fmt.Printf("Not found: %d\n", resp3.StatusCode)

	fmt.Printf("Total server calls: %d\n", callCount)
}

func demoRetryPattern() {
	fmt.Println("\n--- Demo: Retry with Exponential Backoff ---")

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "attempt %d: unavailable", attempts)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "success on attempt %d", attempts)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 10 * time.Second}

	var lastBody string
	var lastStatus int

	for attempt := 0; attempt < 5; attempt++ {
		ctx := context.Background()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/flaky", nil)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("  Attempt %d: error %v\n", attempt+1, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastBody = string(body)
		lastStatus = resp.StatusCode

		fmt.Printf("  Attempt %d: status=%d body=%s\n", attempt+1, resp.StatusCode, body)

		if resp.StatusCode < 500 {
			break
		}

		backoff := time.Duration(math.Pow(2, float64(attempt))) * 10 * time.Millisecond
		time.Sleep(backoff)
	}

	fmt.Printf("Final result: status=%d body=%s\n", lastStatus, lastBody)
}

type SimpleCircuitBreaker struct {
	mu          sync.Mutex
	state       string
	failures    int
	threshold   int
	timeout     time.Duration
	lastFailure time.Time
}

func NewSimpleCircuitBreaker(threshold int, timeout time.Duration) *SimpleCircuitBreaker {
	return &SimpleCircuitBreaker{
		state:     "closed",
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *SimpleCircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case "closed":
		return true
	case "open":
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = "half-open"
			return true
		}
		return false
	case "half-open":
		return true
	}
	return false
}

func (cb *SimpleCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = "closed"
	cb.failures = 0
}

func (cb *SimpleCircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.state = "open"
	}
}

func (cb *SimpleCircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func demoCircuitBreaker() {
	fmt.Println("\n--- Demo: Circuit Breaker ---")

	requestNum := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestNum++
		if requestNum <= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("recovered"))
	}))
	defer server.Close()

	cb := NewSimpleCircuitBreaker(2, 50*time.Millisecond)
	client := &http.Client{Timeout: 5 * time.Second}

	for i := 1; i <= 6; i++ {
		if !cb.Allow() {
			fmt.Printf("  Request %d: BLOCKED by circuit breaker (state=%s)\n", i, cb.State())
			time.Sleep(60 * time.Millisecond)
			continue
		}

		ctx := context.Background()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		resp, err := client.Do(req)

		if err != nil || resp.StatusCode >= 500 {
			cb.RecordFailure()
			status := 0
			if resp != nil {
				status = resp.StatusCode
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			fmt.Printf("  Request %d: FAILED status=%d (state=%s, failures=%d)\n", i, status, cb.State(), cb.failures)
		} else {
			cb.RecordSuccess()
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("  Request %d: SUCCESS body=%s (state=%s)\n", i, body, cb.State())
		}
	}
}

type SimpleRateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
	stop     chan struct{}
}

func NewSimpleRateLimiter(rps int) *SimpleRateLimiter {
	rl := &SimpleRateLimiter{
		tokens:   make(chan struct{}, rps),
		interval: time.Second / time.Duration(rps),
		stop:     make(chan struct{}),
	}

	for range rps {
		rl.tokens <- struct{}{}
	}

	go func() {
		ticker := time.NewTicker(rl.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
				}
			case <-rl.stop:
				return
			}
		}
	}()

	return rl
}

func (rl *SimpleRateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *SimpleRateLimiter) Stop() {
	close(rl.stop)
}

func demoRateLimiting() {
	fmt.Println("\n--- Demo: Rate Limiting Outbound Requests ---")

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	limiter := NewSimpleRateLimiter(5)
	defer limiter.Stop()

	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()
	for i := 0; i < 5; i++ {
		ctx := context.Background()
		if err := limiter.Wait(ctx); err != nil {
			fmt.Printf("  Rate limited: %v\n", err)
			continue
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}

	elapsed := time.Since(start)
	fmt.Printf("Sent %d requests in %s (rate limited to 5 rps)\n", requestCount.Load(), elapsed.Round(time.Millisecond))
}

func demoMultipartUpload() {
	fmt.Println("\n--- Demo: Multipart Form Upload ---")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(32 << 20)

		file, header, err := r.FormFile("document")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error: %v", err)
			return
		}
		defer file.Close()

		content, _ := io.ReadAll(file)
		description := r.FormValue("description")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"filename":    header.Filename,
			"size":        len(content),
			"description": description,
			"content":     string(content),
		})
	}))
	defer server.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, _ := writer.CreateFormFile("document", "hello.txt")
	part.Write([]byte("Hello from Go HTTP Client Ninja!"))

	writer.WriteField("description", "Test upload from chapter 63")
	writer.Close()

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Upload result: filename=%v, size=%v, description=%v\n",
		result["filename"], result["size"], result["description"])
}

/*
RESUMEN DE HTTP CLIENT MASTERY:

DEFAULT CLIENT:
- http.DefaultClient NO tiene timeout
- NUNCA usarlo en producción
- SIEMPRE crear tu propio http.Client

CUSTOM CLIENT:
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   100,
        MaxConnsPerHost:       100,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:  10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}

CONNECTION POOLING:
- Cerrar resp.Body SIEMPRE (defer resp.Body.Close())
- Drenar body si no lo necesitas (io.Copy(io.Discard, resp.Body))
- MaxIdleConnsPerHost default es 2 (muy bajo para alta carga)
- Reutilizar el mismo http.Client (es goroutine-safe)

HTTP/2:
- Automático con TLS (https://)
- ForceAttemptHTTP2: true (default Go 1.13+)
- Multiplexing, header compression, binary framing

REQUESTS:
- GET/POST/PostForm: funciones de conveniencia
- Do: máximo control (headers, methods, context)
- NewRequestWithContext: SIEMPRE usar en producción

CONTEXT:
- Per-request timeouts con context.WithTimeout
- Cancelación con context.WithCancel
- Propagación de context desde HTTP handlers
- Parallel requests con context compartido

JSON PATTERNS:
- json.Marshal + bytes.NewReader para enviar
- json.NewDecoder(resp.Body).Decode para recibir
- Generic helper con doJSON[T]

MULTIPART:
- multipart.NewWriter para crear form data
- CreateFormFile para archivos
- WriteField para campos
- io.Pipe para streaming de archivos grandes

FILE DOWNLOADS:
- io.Copy para streaming
- ProgressReader para mostrar progreso
- Range header para resume

RETRY:
- Exponential backoff: 2^attempt segundos
- Jitter para evitar thundering herd
- Solo reintentar errores recuperables (5xx, network)
- Respetar Retry-After header
- Body replayable con GetBody

CIRCUIT BREAKER:
- CLOSED -> OPEN -> HALF-OPEN -> CLOSED
- Protege contra servicios caídos
- sony/gobreaker para implementación robusta

MIDDLEWARE (RoundTripper):
- Logging, Auth, Headers, Retry, Rate Limiting
- Encadenar con composición
- Clone() requests antes de modificar

TESTING:
- httptest.NewServer para mock servers
- httptest.NewTLSServer para TLS
- httptest.NewRecorder para unit tests
- server.Client() pre-configurado

RATE LIMITING:
- golang.org/x/time/rate para token bucket
- Semáforo para limitar concurrencia
- Per-host rate limiting

PROXY:
- http.ProxyFromEnvironment (HTTP_PROXY, HTTPS_PROXY)
- http.ProxyURL para proxy fijo
- SOCKS5 con golang.org/x/net/proxy

TLS:
- Custom CA: x509.NewCertPool + AppendCertsFromPEM
- mTLS: tls.LoadX509KeyPair
- InsecureSkipVerify solo en desarrollo
- MinVersion: tls.VersionTLS12

LIBRERÍAS:
- net/http: estándar, máximo control
- resty: API fluida, retries, lo más popular
- req: debug avanzado, downloads con progress
- gentleman: plugin system extensible

CHECKLIST DE PRODUCCIÓN:
 1. Custom client con Timeout
 2. Transport con MaxIdleConnsPerHost apropiado
 3. TLS 1.2 mínimo
 4. Cerrar y drenar resp.Body siempre
 5. Context para cancelación per-request
 6. Retries con exponential backoff + jitter
 7. Circuit breaker para servicios externos
 8. Rate limiting outbound
 9. Logging con RoundTripper middleware
10. Limitar tamaño de response body
11. Manejar redirects explícitamente
12. User-Agent descriptivo
*/

/* SUMMARY - HTTP CLIENT MASTERY IN GO:

TOPIC: Advanced HTTP client patterns for production systems
- http.DefaultClient has NO timeout - never use in production
- Custom http.Client with Timeout and Transport configuration mandatory
- Transport settings: MaxIdleConnsPerHost (default 2 is too low), DialContext timeout, TLSHandshakeTimeout
- Connection pooling: must close and drain resp.Body to reuse connections
- HTTP/2 automatic with TLS, multiplexing and binary framing
- Context for per-request timeouts and cancellation with NewRequestWithContext
- Retry pattern with exponential backoff and jitter for 5xx errors
- Circuit breaker states: CLOSED -> OPEN -> HALF-OPEN preventing cascade failures
- RoundTripper middleware for logging, auth, headers, rate limiting
- JSON patterns: json.Marshal for request, json.NewDecoder for response streaming
- Multipart form-data for file uploads with streaming support
- File downloads with progress tracking and resume capability using Range header
- Rate limiting with golang.org/x/time/rate token bucket
- Proxy configuration: http.ProxyFromEnvironment, http.ProxyURL, SOCKS5 support
- TLS: custom CA certificates, mutual TLS (mTLS), minimum version TLS 1.2
- Testing with httptest.NewServer for mock servers
- Popular libraries: resty (fluent API), req (debug features), gentleman (plugins)
- Body must be replayable for retries using GetBody function
- Always limit response body size with io.LimitReader to prevent attacks
- Respect Retry-After header from server responses
*/
