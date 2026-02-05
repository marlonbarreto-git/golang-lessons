// Package main - Capítulo 58: Security Best Practices
// La seguridad es fundamental en cualquier aplicación. Aprenderás
// validación, prevención de vulnerabilidades y OWASP Top 10 para Go.
package main

import (
	"os"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== SECURITY BEST PRACTICES EN GO ===")

	// ============================================
	// INPUT VALIDATION
	// ============================================
	fmt.Println("\n--- Input Validation ---")
	os.Stdout.WriteString(`
REGLAS DE ORO:
1. Validar SIEMPRE en el servidor
2. Whitelist > Blacklist
3. Validar tipo, formato, rango
4. Sanitizar antes de usar

VALIDACIÓN CON go-playground/validator:

go get github.com/go-playground/validator/v10

import "github.com/go-playground/validator/v10"

type User struct {
    Email    string ` + "`" + `json:"email" validate:"required,email"` + "`" + `
    Username string ` + "`" + `json:"username" validate:"required,alphanum,min=3,max=20"` + "`" + `
    Age      int    ` + "`" + `json:"age" validate:"required,gte=0,lte=130"` + "`" + `
    Website  string ` + "`" + `json:"website" validate:"omitempty,url"` + "`" + `
    Password string ` + "`" + `json:"password" validate:"required,min=8,containsany=!@#$%"` + "`" + `
}

validate := validator.New()

func ValidateUser(user User) error {
    return validate.Struct(user)
}

// Handle errors
if err := validate.Struct(user); err != nil {
    for _, err := range err.(validator.ValidationErrors) {
        fmt.Printf("Field: %s, Tag: %s, Value: %v\n",
            err.Field(), err.Tag(), err.Value())
    }
}

CUSTOM VALIDATORS:

validate.RegisterValidation("notsqlkeyword", func(fl validator.FieldLevel) bool {
    sqlKeywords := []string{"SELECT", "DROP", "DELETE", "INSERT", "UPDATE"}
    value := strings.ToUpper(fl.Field().String())
    for _, keyword := range sqlKeywords {
        if strings.Contains(value, keyword) {
            return false
        }
    }
    return true
})

type Query struct {
    Search string ` + "`" + `validate:"required,notsqlkeyword"` + "`" + `
}

SANITIZACIÓN:

import "github.com/microcosm-cc/bluemonday"

// HTML sanitization
policy := bluemonday.UGCPolicy()  // User Generated Content
safe := policy.Sanitize("<script>alert('xss')</script>Hello")
// Output: "Hello"

// Strict policy (solo texto)
strict := bluemonday.StrictPolicy()
clean := strict.Sanitize("<p>Hello <b>World</b></p>")
// Output: "Hello World"

PATH TRAVERSAL PREVENTION:

import (
    "path/filepath"
    "strings"
)

func SafeFilePath(baseDir, userPath string) (string, error) {
    // Clean path
    cleanPath := filepath.Clean(userPath)

    // Check for ..
    if strings.Contains(cleanPath, "..") {
        return "", errors.New("invalid path")
    }

    // Build full path
    fullPath := filepath.Join(baseDir, cleanPath)

    // Verify it's still under baseDir
    if !strings.HasPrefix(fullPath, baseDir) {
        return "", errors.New("path outside base directory")
    }

    return fullPath, nil
}

EMAIL VALIDATION:

import "net/mail"

func ValidateEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}

URL VALIDATION:

import "net/url"

func ValidateURL(rawURL string) (*url.URL, error) {
    u, err := url.Parse(rawURL)
    if err != nil {
        return nil, err
    }

    // Verify scheme
    if u.Scheme != "http" && u.Scheme != "https" {
        return nil, errors.New("invalid scheme")
    }

    return u, nil
}
`)

	// ============================================
	// SQL INJECTION PREVENTION
	// ============================================
	fmt.Println("\n--- SQL Injection Prevention ---")
	os.Stdout.WriteString(`
NUNCA hacer esto:

// VULNERABLE ❌
query := "SELECT * FROM users WHERE email = '" + email + "'"
db.Query(query)

// Ataque: email = "' OR '1'='1"
// Query: SELECT * FROM users WHERE email = '' OR '1'='1'

SIEMPRE usar prepared statements:

// SEGURO ✅
query := "SELECT * FROM users WHERE email = $1"
db.Query(query, email)

// SEGURO ✅ con named parameters (sqlx)
query := "SELECT * FROM users WHERE email = :email AND active = :active"
db.NamedQuery(query, map[string]any{
    "email":  email,
    "active": true,
})

IN CLAUSES:

// VULNERABLE ❌
ids := strings.Join(userIDs, ",")
query := "SELECT * FROM users WHERE id IN (" + ids + ")"

// SEGURO ✅
query, args, _ := sqlx.In("SELECT * FROM users WHERE id IN (?)", userIDs)
query = db.Rebind(query)
db.Select(&users, query, args...)

DYNAMIC QUERIES (usar con cuidado):

// Validar nombres de columnas contra whitelist
func BuildOrderBy(column string) (string, error) {
    allowedColumns := map[string]bool{
        "name": true, "email": true, "created_at": true,
    }

    if !allowedColumns[column] {
        return "", errors.New("invalid column")
    }

    return fmt.Sprintf("ORDER BY %s", column), nil
}

// Usar con prepared statement
orderBy, _ := BuildOrderBy(userInput)
query := "SELECT * FROM users WHERE active = $1 " + orderBy
db.Query(query, true)

SQLC (Type-safe SQL):

-- queries.sql
-- name: GetUser :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users WHERE active = $1 ORDER BY created_at DESC;

// Generated code is type-safe
user, err := queries.GetUser(ctx, email)
users, err := queries.ListUsers(ctx, true)
`)

	// ============================================
	// XSS PREVENTION
	// ============================================
	fmt.Println("\n--- XSS Prevention ---")
	fmt.Println(`
TEMPLATE ESCAPING (html/template):

import "html/template"

// Auto-escaping
tmpl := template.Must(template.New("page").Parse(` + "`" + `
<h1>Hello {{.Name}}</h1>
<div>{{.Content}}</div>
` + "`" + `))

// User input: <script>alert('xss')</script>
// Output: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;

SAFE HTML (cuando necesitas HTML real):

tmpl := template.Must(template.New("page").Parse(` + "`" + `
<div>{{.Content}}</div>
` + "`" + `))

// Sanitizar primero
policy := bluemonday.UGCPolicy()
safeHTML := policy.Sanitize(userInput)

// Marcar como safe
tmpl.Execute(w, map[string]any{
    "Content": template.HTML(safeHTML),
})

JSON ESCAPING:

import "encoding/json"

// json.Marshal auto-escapa
data := map[string]string{
    "message": "<script>alert('xss')</script>",
}
json.NewEncoder(w).Encode(data)
// Output: {"message":"\u003cscript\u003ealert('xss')\u003c/script\u003e"}

JAVASCRIPT CONTEXT:

// VULNERABLE ❌
<script>
var data = "{{.UserInput}}";
</script>

// SEGURO ✅
<script>
var data = {{.UserInputJSON}};  // JSON-encoded
</script>

tmpl.Execute(w, map[string]any{
    "UserInputJSON": template.JS(jsonData),  // Safe for JS context
})

URL CONTEXT:

<a href="{{.URL}}">Link</a>

// Validar URL primero
u, err := url.Parse(userURL)
if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
    return errors.New("invalid URL")
}

// Template auto-escapa
tmpl.Execute(w, map[string]string{"URL": u.String()})

CONTENT SECURITY POLICY:

func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; "+
            "script-src 'self' 'unsafe-inline'; "+
            "style-src 'self' 'unsafe-inline'; "+
            "img-src 'self' data: https:; "+
            "font-src 'self'; "+
            "connect-src 'self'; "+
            "frame-ancestors 'none';")

        next.ServeHTTP(w, r)
    })
}`)
	// ============================================
	// CORS CONFIGURATION
	// ============================================
	fmt.Println("\n--- CORS Configuration ---")
	fmt.Println(`
USANDO rs/cors:

go get github.com/rs/cors

import "github.com/rs/cors"

// PERMISIVO (solo desarrollo) ⚠️
c := cors.New(cors.Options{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"*"},
})

// RESTRICTIVO (producción) ✅
c := cors.New(cors.Options{
    AllowedOrigins: []string{
        "https://app.example.com",
        "https://admin.example.com",
    },
    AllowedMethods: []string{
        http.MethodGet,
        http.MethodPost,
        http.MethodPut,
        http.MethodDelete,
    },
    AllowedHeaders: []string{
        "Accept",
        "Authorization",
        "Content-Type",
        "X-CSRF-Token",
    },
    ExposedHeaders: []string{
        "Link",
        "X-Total-Count",
    },
    AllowCredentials: true,
    MaxAge:          300,  // 5 minutes
})

handler := c.Handler(mux)

MANUAL CORS:

func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")

        // Validate origin
        allowedOrigins := []string{
            "https://app.example.com",
        }

        allowed := false
        for _, o := range allowedOrigins {
            if o == origin {
                allowed = true
                break
            }
        }

        if allowed {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Credentials", "true")
        }

        if r.Method == "OPTIONS" {
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
            w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}`)
	// ============================================
	// RATE LIMITING
	// ============================================
	fmt.Println("\n--- Rate Limiting ---")
	fmt.Println(`
USANDO golang.org/x/time/rate:

import "golang.org/x/time/rate"

// Global rate limiter
limiter := rate.NewLimiter(rate.Limit(10), 20)  // 10 req/s, burst 20

func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

PER-USER RATE LIMITING:

import "sync"

type RateLimiter struct {
    limiters sync.Map  // map[string]*rate.Limiter
}

func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
    limiter, exists := rl.limiters.Load(key)
    if !exists {
        limiter = rate.NewLimiter(rate.Limit(10), 20)
        rl.limiters.Store(key, limiter)
    }
    return limiter.(*rate.Limiter)
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := getUserID(r)
        limiter := rl.GetLimiter(userID)

        if !limiter.Allow() {
            w.Header().Set("Retry-After", "60")
            http.Error(w, "Rate limit exceeded", 429)
            return
        }

        next.ServeHTTP(w, r)
    })
}

SLIDING WINDOW (Redis):

func IsRateLimited(ctx context.Context, userID string, limit int, window time.Duration) bool {
    key := "ratelimit:" + userID
    now := time.Now().Unix()

    pipe := redis.Pipeline()
    pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(now-int64(window.Seconds()), 10))
    pipe.ZCard(ctx, key)
    pipe.ZAdd(ctx, key, &redis.Z{Score: float64(now), Member: now})
    pipe.Expire(ctx, key, window)

    cmds, _ := pipe.Exec(ctx)
    count := cmds[1].(*redis.IntCmd).Val()

    return count >= int64(limit)
}

DISTRIBUTED RATE LIMITING (Redis):

func TakeToken(ctx context.Context, key string, limit int, window time.Duration) bool {
    script := redis.NewScript(` + "`" + `
        local current = redis.call('incr', KEYS[1])
        if current == 1 then
            redis.call('expire', KEYS[1], ARGV[2])
        end
        return current <= tonumber(ARGV[1])
    ` + "`" + `)

    result, _ := script.Run(ctx, redis, []string{key}, limit, int(window.Seconds())).Bool()
    return result
}`)
	// ============================================
	// SECURE HEADERS
	// ============================================
	fmt.Println("\n--- Secure Headers ---")
	fmt.Println(`
USANDO unrolled/secure:

go get github.com/unrolled/secure

import "github.com/unrolled/secure"

secureMiddleware := secure.New(secure.Options{
    // HTTPS
    SSLRedirect:           true,
    SSLHost:               "example.com",
    SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
    STSSeconds:            31536000,  // 1 year
    STSIncludeSubdomains:  true,
    STSPreload:            true,

    // Headers
    FrameDeny:             true,  // X-Frame-Options: DENY
    ContentTypeNosniff:    true,  // X-Content-Type-Options: nosniff
    BrowserXssFilter:      true,  // X-XSS-Protection: 1; mode=block
    ReferrerPolicy:        "strict-origin-when-cross-origin",

    // CSP
    ContentSecurityPolicy: "default-src 'self'",

    // Other
    IsDevelopment: false,
})

handler := secureMiddleware.Handler(mux)

MANUAL SECURITY HEADERS:

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // HTTPS Strict Transport Security
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

        // Prevent clickjacking
        w.Header().Set("X-Frame-Options", "DENY")

        // XSS Protection
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")

        // Referrer Policy
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions Policy
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

        // Content Security Policy
        w.Header().Set("Content-Security-Policy",
            "default-src 'self'; "+
            "script-src 'self' 'unsafe-inline'; "+
            "style-src 'self' 'unsafe-inline'; "+
            "img-src 'self' data: https:; "+
            "font-src 'self'; "+
            "connect-src 'self'; "+
            "frame-ancestors 'none'; "+
            "base-uri 'self'; "+
            "form-action 'self'")

        next.ServeHTTP(w, r)
    })
}`)
	// ============================================
	// OWASP TOP 10 FOR GO
	// ============================================
	fmt.Println("\n--- OWASP Top 10 para Go ---")
	os.Stdout.WriteString(`
1. BROKEN ACCESS CONTROL:

// VULNERABLE ❌
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    user, _ := db.GetUser(userID)
    json.NewEncoder(w).Encode(user)
}

// SEGURO ✅
func GetUserProfile(w http.ResponseWriter, r *http.Request) {
    sessionUserID := getAuthenticatedUserID(r)
    requestedUserID := r.URL.Query().Get("id")

    // Authorization check
    if sessionUserID != requestedUserID && !isAdmin(sessionUserID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    user, _ := db.GetUser(requestedUserID)
    json.NewEncoder(w).Encode(user)
}

2. CRYPTOGRAPHIC FAILURES:

// NUNCA almacenar passwords en texto plano ❌
user.Password = password

// Usar bcrypt ✅
import "golang.org/x/crypto/bcrypt"

hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
user.Password = string(hashedPassword)

// Verificar
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
if err != nil {
    return errors.New("invalid password")
}

// Encriptar datos sensibles
import "crypto/aes"
import "crypto/cipher"

func Encrypt(plaintext, key []byte) ([]byte, error) {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)
    return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

3. INJECTION (SQL, NoSQL, OS):

// SQL Injection: usar prepared statements (ver arriba)

// Command Injection:
// VULNERABLE ❌
exec.Command("sh", "-c", "ls "+userInput).Run()

// SEGURO ✅
exec.Command("ls", filepath.Clean(userInput)).Run()

// NoSQL Injection (MongoDB):
// VULNERABLE ❌
filter := bson.M{"username": username, "password": password}

// SEGURO ✅
filter := bson.M{
    "username": bson.M{"$eq": username},
    "password": bson.M{"$eq": hashedPassword},
}

4. INSECURE DESIGN:

- Usar principio de menor privilegio
- Separación de ambientes (dev/staging/prod)
- Secrets en variables de entorno o vault
- Multi-factor authentication
- Rate limiting
- Audit logging

5. SECURITY MISCONFIGURATION:

// Deshabilitar stack traces en producción
if !isDev {
    gin.SetMode(gin.ReleaseMode)
}

// No exponer información sensible en errores
// VULNERABLE ❌
http.Error(w, err.Error(), 500)

// SEGURO ✅
log.Printf("Error: %v", err)
http.Error(w, "Internal Server Error", 500)

// Configuración segura de cookies
http.SetCookie(w, &http.Cookie{
    Name:     "session",
    Value:    sessionID,
    HttpOnly: true,      // No accesible desde JavaScript
    Secure:   true,      // Solo HTTPS
    SameSite: http.SameSiteStrictMode,
    MaxAge:   3600,
    Path:     "/",
})

6. VULNERABLE AND OUTDATED COMPONENTS:

// Usar dependabot o renovate
// Mantener dependencias actualizadas
go get -u ./...

// Escanear vulnerabilidades
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

7. IDENTIFICATION AND AUTHENTICATION FAILURES:

- Implementar MFA
- Rate limiting en login
- Password policies
- Session timeout
- Secure password reset

func LoginWithRateLimit(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")

    // Check rate limit
    if isRateLimited("login:" + username) {
        http.Error(w, "Too many attempts", 429)
        return
    }

    // Authenticate
    if !authenticate(username, password) {
        incrementFailedAttempts(username)
        http.Error(w, "Invalid credentials", 401)
        return
    }

    // Create session
    createSession(w, username)
}

8. SOFTWARE AND DATA INTEGRITY FAILURES:

- Verificar checksums de dependencias
- Code signing
- CI/CD pipeline seguro
- go.sum para integridad de módulos

9. SECURITY LOGGING AND MONITORING FAILURES:

import "log/slog"

// Structured logging
slog.Info("User login",
    slog.String("user_id", userID),
    slog.String("ip", r.RemoteAddr),
    slog.Time("timestamp", time.Now()),
)

// Log security events
func LogSecurityEvent(event string, details map[string]any) {
    slog.Warn("Security event",
        slog.String("event", event),
        slog.Any("details", details),
    )
}

10. SERVER-SIDE REQUEST FORGERY (SSRF):

// VULNERABLE ❌
url := r.FormValue("url")
http.Get(url)  // User puede acceder a internal IPs

// SEGURO ✅
func SafeHTTPClient() *http.Client {
    dialer := &net.Dialer{
        Timeout: 10 * time.Second,
    }

    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
                host, _, _ := net.SplitHostPort(addr)
                ip := net.ParseIP(host)

                // Block private IPs
                if ip != nil && (ip.IsPrivate() || ip.IsLoopback()) {
                    return nil, errors.New("access to private IP blocked")
                }

                return dialer.DialContext(ctx, network, addr)
            },
        },
    }
}
`)

	// ============================================
	// SECRET MANAGEMENT
	// ============================================
	fmt.Println("\n--- Secret Management ---")
	fmt.Println(`
VARIABLES DE ENTORNO:

import "github.com/joho/godotenv"

// Load .env in development only
if os.Getenv("ENV") != "production" {
    godotenv.Load()
}

dbPassword := os.Getenv("DB_PASSWORD")

HASHICORP VAULT:

go get github.com/hashicorp/vault/api

import "github.com/hashicorp/vault/api"

func GetSecret(path string) (string, error) {
    config := api.DefaultConfig()
    config.Address = os.Getenv("VAULT_ADDR")

    client, err := api.NewClient(config)
    if err != nil {
        return "", err
    }

    client.SetToken(os.Getenv("VAULT_TOKEN"))

    secret, err := client.Logical().Read(path)
    if err != nil {
        return "", err
    }

    return secret.Data["value"].(string), nil
}

AWS SECRETS MANAGER:

import "github.com/aws/aws-sdk-go/service/secretsmanager"

func GetAWSSecret(secretName string) (string, error) {
    svc := secretsmanager.New(session.New())

    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    }

    result, err := svc.GetSecretValue(input)
    if err != nil {
        return "", err
    }

    return *result.SecretString, nil
}

KUBERNETES SECRETS:

apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  password: <base64-encoded>

// Montar como variable de entorno
env:
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: db-secret
      key: password`)
	// Demostración conceptual
	demonstrateSecurity()
}

func demonstrateSecurity() {
	fmt.Printf("\n--- Demostración Conceptual ---\n")
	fmt.Printf("OWASP Top 10 Checklist:\n")
	fmt.Printf("  ✓ Access control verificado\n")
	fmt.Printf("  ✓ Passwords hasheados (bcrypt)\n")
	fmt.Printf("  ✓ Prepared statements (SQL)\n")
	fmt.Printf("  ✓ Input validation\n")
	fmt.Printf("  ✓ Security headers\n")
	fmt.Printf("  ✓ Rate limiting\n")
	fmt.Printf("  ✓ CORS configurado\n")
	fmt.Printf("  ✓ Secrets en vault\n")
	fmt.Printf("  ✓ Security logging\n")
	fmt.Printf("  ✓ SSRF prevention\n")
	fmt.Printf("\nSession timeout recomendado: %v\n", 15*time.Minute)
}

/*
RESUMEN DE SECURITY BEST PRACTICES:

INPUT VALIDATION:
- go-playground/validator para validación
- bluemonday para sanitización HTML
- Whitelist sobre blacklist
- Validar tipo, formato, rango

SQL INJECTION:
- SIEMPRE usar prepared statements
- sqlx.In para IN clauses
- Validar nombres de columnas
- sqlc para type-safety

XSS PREVENTION:
- html/template auto-escapa
- bluemonday para sanitizar HTML
- CSP headers
- JSON encoding automático

CORS:
- rs/cors librería
- Whitelist origins
- AllowCredentials con cuidado
- Validar origin en backend

RATE LIMITING:
- golang.org/x/time/rate
- Per-user limiters
- Sliding window con Redis
- Retry-After header

SECURE HEADERS:
- HSTS (Strict-Transport-Security)
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- Content-Security-Policy
- Referrer-Policy

OWASP TOP 10:
1. Access Control: verificar authorization
2. Crypto: bcrypt, no texto plano
3. Injection: prepared statements
4. Design: least privilege
5. Misconfiguration: no exponer errores
6. Components: govulncheck
7. Auth: MFA, rate limiting
8. Integrity: go.sum, checksums
9. Logging: structured logs
10. SSRF: block private IPs

SECRET MANAGEMENT:
- Variables de entorno (.env)
- HashiCorp Vault
- AWS Secrets Manager
- Kubernetes Secrets

MEJORES PRÁCTICAS:
1. Validar en servidor siempre
2. Nunca confiar en input del usuario
3. Principio de menor privilegio
4. Defense in depth (múltiples capas)
5. Fail securely
6. Security by design, not afterthought
7. Mantener dependencias actualizadas
8. Auditar y monitorear
9. Encrypt data in transit y at rest
10. Regular security testing
*/
