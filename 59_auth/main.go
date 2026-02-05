// Package main - Capítulo 59: Authentication & Authorization
// Auth es crucial en cualquier aplicación. Aprenderás JWT, OAuth2,
// OIDC, session management, password hashing y RBAC.
package main

import (
	"os"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== AUTHENTICATION & AUTHORIZATION EN GO ===")

	// ============================================
	// JWT (JSON WEB TOKENS)
	// ============================================
	fmt.Println("\n--- JWT con golang-jwt/jwt ---")
	os.Stdout.WriteString(`
INSTALACIÓN:
go get github.com/golang-jwt/jwt/v5

ESTRUCTURA JWT:
header.payload.signature

Header: {"alg": "HS256", "typ": "JWT"}
Payload: {"sub": "user123", "exp": 1234567890}
Signature: HMACSHA256(base64(header) + "." + base64(payload), secret)

CREAR JWT:

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID   string   ` + "`" + `json:"user_id"` + "`" + `
    Email    string   ` + "`" + `json:"email"` + "`" + `
    Roles    []string ` + "`" + `json:"roles"` + "`" + `
    jwt.RegisteredClaims
}

func GenerateJWT(userID, email string, roles []string, secret []byte) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        Roles:  roles,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "myapp",
            Subject:   userID,
            ID:        uuid.NewString(),  // jti (JWT ID)
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secret)
}

VALIDAR JWT:

func ValidateJWT(tokenString string, secret []byte) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
        // Verificar algoritmo
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secret, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

JWT MIDDLEWARE:

func JWTAuthMiddleware(secret []byte) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            // Format: "Bearer <token>"
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
                return
            }

            tokenString := parts[1]

            // Validate token
            claims, err := ValidateJWT(tokenString, secret)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Add claims to context
            ctx := context.WithValue(r.Context(), "claims", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Usar en handler
func GetProfile(w http.ResponseWriter, r *http.Request) {
    claims := r.Context().Value("claims").(*Claims)
    json.NewEncoder(w).Encode(map[string]any{
        "user_id": claims.UserID,
        "email":   claims.Email,
        "roles":   claims.Roles,
    })
}

JWT CON RS256 (RSA):

import "crypto/rsa"
import "crypto/x509"

// Generar keys
// openssl genrsa -out private.pem 2048
// openssl rsa -in private.pem -pubout -out public.pem

func LoadRSAKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
    privateKeyData, _ := os.ReadFile("private.pem")
    privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)

    publicKeyData, _ := os.ReadFile("public.pem")
    publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)

    return privateKey, publicKey, nil
}

// Sign
token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
tokenString, _ := token.SignedString(privateKey)

// Verify
token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
    return publicKey, nil
})

REFRESH TOKENS:

type TokenPair struct {
    AccessToken  string ` + "`" + `json:"access_token"` + "`" + `
    RefreshToken string ` + "`" + `json:"refresh_token"` + "`" + `
    ExpiresIn    int    ` + "`" + `json:"expires_in"` + "`" + `
}

func GenerateTokenPair(userID string) (*TokenPair, error) {
    // Access token (corta duración)
    accessClaims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
        },
    }
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, _ := accessToken.SignedString(jwtSecret)

    // Refresh token (larga duración)
    refreshClaims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
        },
    }
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, _ := refreshToken.SignedString(refreshSecret)

    // Guardar refresh token en DB
    saveRefreshToken(userID, refreshTokenString)

    return &TokenPair{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
        ExpiresIn:    900,  // 15 minutes
    }, nil
}

func RefreshAccessToken(refreshToken string) (*TokenPair, error) {
    // Validar refresh token
    claims, err := ValidateJWT(refreshToken, refreshSecret)
    if err != nil {
        return nil, err
    }

    // Verificar que existe en DB
    if !isValidRefreshToken(claims.UserID, refreshToken) {
        return nil, errors.New("invalid refresh token")
    }

    // Generar nuevo par
    return GenerateTokenPair(claims.UserID)
}
`)

	// ============================================
	// OAUTH2
	// ============================================
	fmt.Println("\n--- OAuth2 con golang.org/x/oauth2 ---")
	fmt.Println(`
INSTALACIÓN:
go get golang.org/x/oauth2

OAUTH2 FLOWS:

1. AUTHORIZATION CODE (más común):

import "golang.org/x/oauth2"
import "golang.org/x/oauth2/google"

var oauth2Config = &oauth2.Config{
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    RedirectURL:  "http://localhost:8080/callback",
    Scopes: []string{
        "https://www.googleapis.com/auth/userinfo.email",
        "https://www.googleapis.com/auth/userinfo.profile",
    },
    Endpoint: google.Endpoint,
}

// Redirect to OAuth provider
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    state := generateRandomState()  // CSRF protection
    saveState(state)  // Save in session or Redis

    url := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Handle callback
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
    // Verify state
    state := r.FormValue("state")
    if !isValidState(state) {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }

    // Exchange code for token
    code := r.FormValue("code")
    token, err := oauth2Config.Exchange(context.Background(), code)
    if err != nil {
        http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
        return
    }

    // Use token to get user info
    client := oauth2Config.Client(context.Background(), token)
    resp, _ := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    defer resp.Body.Close()

    var userInfo struct {
        Email string ` + "`" + `json:"email"` + "`" + `
        Name  string ` + "`" + `json:"name"` + "`" + `
    }
    json.NewDecoder(resp.Body).Decode(&userInfo)

    // Create session
    createSession(w, userInfo.Email)
}

2. CLIENT CREDENTIALS (server-to-server):

import "golang.org/x/oauth2/clientcredentials"

config := &clientcredentials.Config{
    ClientID:     "client-id",
    ClientSecret: "client-secret",
    TokenURL:     "https://auth.example.com/token",
    Scopes:       []string{"api.read", "api.write"},
}

token, _ := config.Token(context.Background())
client := config.Client(context.Background())

// Use client for API calls
resp, _ := client.Get("https://api.example.com/data")

3. PASSWORD GRANT (no recomendado):

config := &oauth2.Config{
    ClientID:     "client-id",
    ClientSecret: "client-secret",
    Endpoint: oauth2.Endpoint{
        TokenURL: "https://auth.example.com/token",
    },
}

token, err := config.PasswordCredentialsToken(ctx, username, password)

GITHUB OAUTH:

import "golang.org/x/oauth2/github"

githubConfig := &oauth2.Config{
    ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
    ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
    RedirectURL:  "http://localhost:8080/callback",
    Scopes:       []string{"user:email"},
    Endpoint:     github.Endpoint,
}

CUSTOM OAUTH2 SERVER:

import "github.com/go-oauth2/oauth2/v4"

// Ver capítulo 36_web_backend para implementación completa`)
	// ============================================
	// OIDC (OPENID CONNECT)
	// ============================================
	fmt.Println("\n--- OpenID Connect ---")
	fmt.Println(`
OIDC = OAuth2 + ID Token (JWT)

INSTALACIÓN:
go get github.com/coreos/go-oidc/v3/oidc

OIDC CLIENT:

import "github.com/coreos/go-oidc/v3/oidc"

provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
if err != nil {
    return err
}

oidcConfig := &oidc.Config{
    ClientID: os.Getenv("CLIENT_ID"),
}

verifier := provider.Verifier(oidcConfig)

oauth2Config := &oauth2.Config{
    ClientID:     os.Getenv("CLIENT_ID"),
    ClientSecret: os.Getenv("CLIENT_SECRET"),
    RedirectURL:  "http://localhost:8080/callback",
    Endpoint:     provider.Endpoint(),
    Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
}

// Login flow (similar a OAuth2)
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    state := randomString()
    nonce := randomString()  // OIDC nonce for ID token

    http.SetCookie(w, &http.Cookie{
        Name:     "nonce",
        Value:    nonce,
        MaxAge:   300,
        HttpOnly: true,
        Secure:   true,
    })

    url := oauth2Config.AuthCodeURL(state,
        oidc.Nonce(nonce),
        oauth2.AccessTypeOffline,
    )
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback with ID token verification
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
    code := r.FormValue("code")

    // Exchange
    oauth2Token, err := oauth2Config.Exchange(ctx, code)
    if err != nil {
        http.Error(w, "Failed to exchange token", 500)
        return
    }

    // Extract ID Token
    rawIDToken, ok := oauth2Token.Extra("id_token").(string)
    if !ok {
        http.Error(w, "No id_token", 500)
        return
    }

    // Verify ID Token
    idToken, err := verifier.Verify(ctx, rawIDToken)
    if err != nil {
        http.Error(w, "Invalid ID token", 401)
        return
    }

    // Verify nonce
    nonceCookie, _ := r.Cookie("nonce")
    if idToken.Nonce != nonceCookie.Value {
        http.Error(w, "Invalid nonce", 401)
        return
    }

    // Extract claims
    var claims struct {
        Email         string ` + "`" + `json:"email"` + "`" + `
        EmailVerified bool   ` + "`" + `json:"email_verified"` + "`" + `
        Name          string ` + "`" + `json:"name"` + "`" + `
        Picture       string ` + "`" + `json:"picture"` + "`" + `
    }
    idToken.Claims(&claims)

    // Create session
    createSession(w, claims.Email)
}`)
	// ============================================
	// SESSION MANAGEMENT
	// ============================================
	fmt.Println("\n--- Session Management ---")
	fmt.Println(`
GORILLA SESSIONS:

go get github.com/gorilla/sessions

import "github.com/gorilla/sessions"

var store = sessions.NewCookieStore([]byte("secret-key"))

// Cookie-based session
func SetSession(w http.ResponseWriter, r *http.Request, userID string) error {
    session, _ := store.Get(r, "session-name")
    session.Values["user_id"] = userID
    session.Values["authenticated"] = true
    session.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   3600,  // 1 hour
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    }
    return session.Save(r, w)
}

func GetSession(r *http.Request) (string, bool) {
    session, _ := store.Get(r, "session-name")
    userID, ok := session.Values["user_id"].(string)
    if !ok {
        return "", false
    }
    authenticated, _ := session.Values["authenticated"].(bool)
    return userID, authenticated
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
    session, _ := store.Get(r, "session-name")
    session.Options.MaxAge = -1
    return session.Save(r, w)
}

SESSION MIDDLEWARE:

func SessionAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID, authenticated := GetSession(r)
        if !authenticated {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "user_id", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

REDIS SESSION STORE:

import "github.com/rbcervilla/redisstore/v9"

store, err := redisstore.NewRedisStore(ctx, redisClient)
if err != nil {
    return err
}

store.Options = &sessions.Options{
    Path:     "/",
    MaxAge:   86400 * 7,  // 7 days
    HttpOnly: true,
    Secure:   true,
}

// Uso igual que cookie store
session, _ := store.Get(r, "session-name")
session.Values["user_id"] = userID
session.Save(r, w)

SESSION TOKEN (Custom):

type Session struct {
    ID        string
    UserID    string
    CreatedAt time.Time
    ExpiresAt time.Time
    IPAddress string
    UserAgent string
}

func CreateSession(userID, ip, ua string) (*Session, error) {
    session := &Session{
        ID:        uuid.NewString(),
        UserID:    userID,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
        IPAddress: ip,
        UserAgent: ua,
    }

    // Save to Redis
    data, _ := json.Marshal(session)
    redis.Set(ctx, "session:"+session.ID, data, 24*time.Hour)

    return session, nil
}

func ValidateSession(sessionID, ip, ua string) (*Session, error) {
    data, err := redis.Get(ctx, "session:"+sessionID).Bytes()
    if err != nil {
        return nil, errors.New("invalid session")
    }

    var session Session
    json.Unmarshal(data, &session)

    // Validate IP and UserAgent (optional)
    if session.IPAddress != ip || session.UserAgent != ua {
        return nil, errors.New("session fingerprint mismatch")
    }

    if time.Now().After(session.ExpiresAt) {
        return nil, errors.New("session expired")
    }

    return &session, nil
}

func InvalidateSession(sessionID string) error {
    return redis.Del(ctx, "session:"+sessionID).Err()
}

// Invalidar todas las sesiones del usuario
func InvalidateUserSessions(userID string) error {
    pattern := "session:*"
    iter := redis.Scan(ctx, 0, pattern, 0).Iterator()

    for iter.Next(ctx) {
        key := iter.Val()
        data, _ := redis.Get(ctx, key).Bytes()

        var session Session
        json.Unmarshal(data, &session)

        if session.UserID == userID {
            redis.Del(ctx, key)
        }
    }

    return iter.Err()
}`)
	// ============================================
	// PASSWORD HASHING
	// ============================================
	fmt.Println("\n--- Password Hashing ---")
	os.Stdout.WriteString(`
BCRYPT (Recomendado para la mayoría):

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    // Cost: 10-12 para producción
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash), err
}

func VerifyPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

// Custom cost
hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

ARGON2 (Más seguro, más lento):

import "golang.org/x/crypto/argon2"
import "crypto/rand"

type Argon2Params struct {
    Memory      uint32
    Iterations  uint32
    Parallelism uint8
    SaltLength  uint32
    KeyLength   uint32
}

var defaultParams = &Argon2Params{
    Memory:      64 * 1024,  // 64 MB
    Iterations:  3,
    Parallelism: 2,
    SaltLength:  16,
    KeyLength:   32,
}

func HashPasswordArgon2(password string) (string, error) {
    // Generate salt
    salt := make([]byte, defaultParams.SaltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", err
    }

    // Hash
    hash := argon2.IDKey(
        []byte(password),
        salt,
        defaultParams.Iterations,
        defaultParams.Memory,
        defaultParams.Parallelism,
        defaultParams.KeyLength,
    )

    // Encode: $argon2id$v=19$m=65536,t=3,p=2$salt$hash
    b64Salt := base64.RawStdEncoding.EncodeToString(salt)
    b64Hash := base64.RawStdEncoding.EncodeToString(hash)

    encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
        argon2.Version, defaultParams.Memory, defaultParams.Iterations,
        defaultParams.Parallelism, b64Salt, b64Hash)

    return encoded, nil
}

func VerifyPasswordArgon2(encoded, password string) (bool, error) {
    // Parse encoded hash
    parts := strings.Split(encoded, "$")
    if len(parts) != 6 {
        return false, errors.New("invalid hash format")
    }

    var params Argon2Params
    _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
        &params.Memory, &params.Iterations, &params.Parallelism)
    if err != nil {
        return false, err
    }

    salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
    hash, _ := base64.RawStdEncoding.DecodeString(parts[5])

    params.SaltLength = uint32(len(salt))
    params.KeyLength = uint32(len(hash))

    // Hash provided password
    newHash := argon2.IDKey(
        []byte(password),
        salt,
        params.Iterations,
        params.Memory,
        params.Parallelism,
        params.KeyLength,
    )

    // Constant time comparison
    return subtle.ConstantTimeCompare(hash, newHash) == 1, nil
}

SCRYPT (Alternativa):

import "golang.org/x/crypto/scrypt"

func HashPasswordScrypt(password string) (string, error) {
    salt := make([]byte, 16)
    rand.Read(salt)

    hash, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
    if err != nil {
        return "", err
    }

    return base64.StdEncoding.EncodeToString(salt) + "$" +
           base64.StdEncoding.EncodeToString(hash), nil
}
`)

	// ============================================
	// RBAC (ROLE-BASED ACCESS CONTROL)
	// ============================================
	fmt.Println("\n--- RBAC Patterns ---")
	fmt.Println(`
SIMPLE RBAC:

type Role string

const (
    RoleUser  Role = "user"
    RoleAdmin Role = "admin"
    RoleGuest Role = "guest"
)

type User struct {
    ID    string
    Email string
    Roles []Role
}

func (u *User) HasRole(role Role) bool {
    for _, r := range u.Roles {
        if r == role {
            return true
        }
    }
    return false
}

func (u *User) HasAnyRole(roles ...Role) bool {
    for _, role := range roles {
        if u.HasRole(role) {
            return true
        }
    }
    return false
}

// RBAC Middleware
func RequireRole(role Role) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user").(*User)
            if !user.HasRole(role) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Uso
r.Handle("/admin", RequireRole(RoleAdmin)(adminHandler))

RBAC CON PERMISOS:

type Permission string

const (
    PermissionRead   Permission = "read"
    PermissionWrite  Permission = "write"
    PermissionDelete Permission = "delete"
)

type Role struct {
    Name        string
    Permissions []Permission
}

var roles = map[string]*Role{
    "user": {
        Name:        "user",
        Permissions: []Permission{PermissionRead},
    },
    "admin": {
        Name: "admin",
        Permissions: []Permission{
            PermissionRead,
            PermissionWrite,
            PermissionDelete,
        },
    },
}

func (u *User) HasPermission(perm Permission) bool {
    for _, roleName := range u.Roles {
        role, exists := roles[string(roleName)]
        if !exists {
            continue
        }
        for _, p := range role.Permissions {
            if p == perm {
                return true
            }
        }
    }
    return false
}

// Permission Middleware
func RequirePermission(perm Permission) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user").(*User)
            if !user.HasPermission(perm) {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

CASBIN (RBAC/ABAC Framework):

go get github.com/casbin/casbin/v2

import "github.com/casbin/casbin/v2"

// model.conf
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act

// policy.csv
p, admin, /api/*, *
p, user, /api/users/:id, read
g, alice, admin
g, bob, user

// Code
enforcer, _ := casbin.NewEnforcer("model.conf", "policy.csv")

// Check permission
allowed, _ := enforcer.Enforce("alice", "/api/users", "delete")
if allowed {
    // Allow access
}

// RBAC Middleware
func CasbinMiddleware(e *casbin.Enforcer) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := getUserFromContext(r.Context())
            path := r.URL.Path
            method := r.Method

            allowed, _ := e.Enforce(user, path, method)
            if !allowed {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}`)
	// ============================================
	// AUTH MIDDLEWARE PATTERNS
	// ============================================
	fmt.Println("\n--- Auth Middleware Patterns ---")
	fmt.Println(`
COMPOSABLE MIDDLEWARE:

func Authenticated(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := getAuthenticatedUser(r)
        if user == nil {
            http.Error(w, "Unauthorized", 401)
            return
        }
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user").(*User)
            if !user.HasRole(role) {
                http.Error(w, "Forbidden", 403)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Chain
handler := Authenticated(
    RequireRole("admin")(
        http.HandlerFunc(adminHandler),
    ),
)

GIN MIDDLEWARE:

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        claims, err := ValidateJWT(token, secret)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("roles", claims.Roles)
        c.Next()
    }
}

func RoleMiddleware(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        roles, _ := c.Get("roles")
        userRoles := roles.([]string)

        hasRole := false
        for _, r := range userRoles {
            if r == role {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
            return
        }

        c.Next()
    }
}

// Uso
r.Use(AuthMiddleware())
admin := r.Group("/admin")
admin.Use(RoleMiddleware("admin"))
admin.GET("/users", listUsers)

ECHO MIDDLEWARE:

func JWTMiddleware(secret []byte) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            auth := c.Request().Header.Get("Authorization")
            if auth == "" {
                return echo.NewHTTPError(401, "Missing token")
            }

            token := strings.TrimPrefix(auth, "Bearer ")
            claims, err := ValidateJWT(token, secret)
            if err != nil {
                return echo.NewHTTPError(401, "Invalid token")
            }

            c.Set("user", claims)
            return next(c)
        }
    }
}

// Uso
e.Use(JWTMiddleware(secret))`)
	// Demostración conceptual
	demonstrateAuth()
}

func demonstrateAuth() {
	fmt.Printf("\n--- Demostración Conceptual ---\n")
	fmt.Printf("Métodos de autenticación:\n")
	fmt.Printf("  JWT: Stateless, escalable, bueno para APIs\n")
	fmt.Printf("  Sessions: Stateful, fácil invalidación\n")
	fmt.Printf("  OAuth2: Delegación, login social\n")
	fmt.Printf("  OIDC: OAuth2 + identity, SSO\n")
	fmt.Printf("\nPassword hashing:\n")
	fmt.Printf("  bcrypt: Balance seguridad/performance\n")
	fmt.Printf("  argon2: Más seguro, más CPU/RAM\n")
	fmt.Printf("  scrypt: Alternativa\n")
	fmt.Printf("\nTTLs recomendados:\n")
	fmt.Printf("  - Access token: %v\n", 15*time.Minute)
	fmt.Printf("  - Refresh token: %v\n", 7*24*time.Hour)
	fmt.Printf("  - Session: %v\n", 24*time.Hour)
}

/*
RESUMEN DE AUTHENTICATION & AUTHORIZATION:

JWT:
- golang-jwt/jwt/v5
- Stateless, escalable
- HS256 (HMAC) o RS256 (RSA)
- Access + Refresh tokens
- Claims: user_id, roles, exp, iat

OAUTH2:
- golang.org/x/oauth2
- Authorization Code (web apps)
- Client Credentials (server-to-server)
- Implicit/Password (deprecated)
- State para CSRF protection

OIDC:
- OAuth2 + ID Token (JWT)
- coreos/go-oidc
- Nonce para replay protection
- UserInfo endpoint

SESSION MANAGEMENT:
- gorilla/sessions (cookie/Redis)
- HttpOnly, Secure, SameSite
- Session timeout
- Invalidación por usuario
- Fingerprinting (IP, UserAgent)

PASSWORD HASHING:
- bcrypt: balanced (cost 10-12)
- argon2: más seguro (64MB, 3 iter)
- NUNCA texto plano
- Constant time comparison

RBAC:
- Roles: user, admin, guest
- Permissions: read, write, delete
- Middleware por role/permission
- Casbin para RBAC/ABAC complejo

MIDDLEWARE PATTERNS:
- Authenticated: verificar token/session
- RequireRole: verificar role
- RequirePermission: verificar permission
- Composable: chain múltiples

MEJORES PRÁCTICAS:
1. JWT para APIs, sessions para web
2. Short-lived access tokens
3. Refresh token rotation
4. HTTPS obligatorio
5. CSRF protection con state/nonce
6. Rate limiting en login
7. MFA para cuentas sensibles
8. Audit logging
9. Secure password policies
10. Regular token rotation

SEGURIDAD:
- Validar algoritmo JWT (alg)
- Verificar exp, iat, nbf
- State/nonce anti-CSRF
- HttpOnly cookies
- SameSite strict/lax
- Invalidar sesiones al cambiar password
*/
