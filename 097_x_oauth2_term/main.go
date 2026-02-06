// Package main - Chapter 097: golang.org/x/oauth2 and golang.org/x/term
// x/oauth2 implements OAuth 2.0 flows for authentication.
// x/term provides terminal handling utilities.
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== GOLANG.ORG/X/OAUTH2 AND GOLANG.ORG/X/TERM ===")

	// ============================================
	// OAUTH2 OVERVIEW
	// ============================================
	fmt.Println("\n--- golang.org/x/oauth2 Overview ---")

	fmt.Println(`
  OAuth 2.0 is the industry standard for authorization.
  x/oauth2 provides a complete implementation.

  Install: go get golang.org/x/oauth2

  OAUTH2 ROLES:
    Resource Owner  - The user
    Client          - Your application
    Auth Server     - Issues tokens (Google, GitHub, etc.)
    Resource Server - API that accepts tokens

  GRANT TYPES SUPPORTED:
    1. Authorization Code  - Web apps (most common)
    2. Authorization Code + PKCE - Mobile/SPA apps
    3. Client Credentials  - Server-to-server
    4. Device Flow         - IoT/CLI devices`)

	// ============================================
	// OAUTH2 CONFIG
	// ============================================
	fmt.Println("\n--- OAuth2 Configuration ---")

	fmt.Println(`
  BASIC CONFIG:

    import "golang.org/x/oauth2"

    config := &oauth2.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        Endpoint: oauth2.Endpoint{
            AuthURL:  "https://provider.com/auth",
            TokenURL: "https://provider.com/token",
        },
        RedirectURL: "http://localhost:8080/callback",
        Scopes:      []string{"openid", "profile", "email"},
    }

  PROVIDER-SPECIFIC CONFIGS:

    import (
        "golang.org/x/oauth2/google"
        "golang.org/x/oauth2/github"
        "golang.org/x/oauth2/facebook"
    )

    // Google
    config := &oauth2.Config{
        ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
        ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
        Endpoint:     google.Endpoint,
        Scopes:       []string{"openid", "profile", "email"},
        RedirectURL:  "http://localhost:8080/callback",
    }

    // GitHub
    config := &oauth2.Config{
        ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
        ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
        Endpoint:     github.Endpoint,
        Scopes:       []string{"user:email", "repo"},
    }`)

	// ============================================
	// AUTHORIZATION CODE FLOW
	// ============================================
	fmt.Println("\n--- Authorization Code Flow ---")

	os.Stdout.WriteString(`
  THE MOST COMMON OAUTH2 FLOW:

  1. Redirect user to auth provider:

    // Generate state for CSRF protection
    state := generateRandomString(32)

    url := config.AuthCodeURL(state,
        oauth2.AccessTypeOffline,  // request refresh token
    )
    http.Redirect(w, r, url, http.StatusFound)

  2. Handle callback (user returns with code):

    func callbackHandler(w http.ResponseWriter, r *http.Request) {
        // Verify state matches
        if r.URL.Query().Get("state") != expectedState {
            http.Error(w, "invalid state", 400)
            return
        }

        // Exchange code for token
        code := r.URL.Query().Get("code")
        token, err := config.Exchange(r.Context(), code)
        if err != nil { ... }

        // token.AccessToken  - for API calls
        // token.RefreshToken - for getting new access tokens
        // token.Expiry       - when access token expires
        // token.TokenType    - usually "Bearer"
    }

  3. Use token to make API calls:

    // Create an HTTP client with the token
    client := config.Client(ctx, token)

    // This client automatically:
    // - Adds Authorization: Bearer <token> header
    // - Refreshes token when expired (if refresh token exists)
    resp, err := client.Get("https://api.provider.com/user")

  COMPLETE WEB SERVER EXAMPLE:

    func main() {
        http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
            url := config.AuthCodeURL("random-state")
            http.Redirect(w, r, url, http.StatusFound)
        })

        http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
            code := r.URL.Query().Get("code")
            token, err := config.Exchange(r.Context(), code)
            if err != nil {
                http.Error(w, err.Error(), 500)
                return
            }

            client := config.Client(r.Context(), token)
            resp, err := client.Get("https://api.github.com/user")
            // ... use response
            fmt.Fprintf(w, "Logged in! Token expires: %s", token.Expiry)
        })

        http.ListenAndServe(":8080", nil)
    }
` + "\n")

	// ============================================
	// PKCE FLOW
	// ============================================
	fmt.Println("--- PKCE Flow (for public clients) ---")

	fmt.Println(`
  PKCE (Proof Key for Code Exchange) adds security for
  clients that can't keep secrets (mobile, SPA, CLI).

  FLOW:

    import "golang.org/x/oauth2"

    // 1. Generate code verifier (random string)
    verifier := oauth2.GenerateVerifier()

    // 2. Create auth URL with PKCE challenge
    url := config.AuthCodeURL("state",
        oauth2.S256ChallengeOption(verifier),
    )

    // 3. Exchange code with verifier
    token, err := config.Exchange(ctx, code,
        oauth2.VerifierOption(verifier),
    )

  WHY PKCE:
    - Authorization code can be intercepted
    - Without PKCE, interceptor can exchange the code
    - With PKCE, only the original client has the verifier
    - Recommended for ALL authorization code flows (RFC 9126)`)

	// ============================================
	// CLIENT CREDENTIALS FLOW
	// ============================================
	fmt.Println("\n--- Client Credentials Flow ---")

	fmt.Println(`
  For server-to-server communication (no user involved):

    import (
        "golang.org/x/oauth2"
        "golang.org/x/oauth2/clientcredentials"
    )

    config := &clientcredentials.Config{
        ClientID:     "service-account-id",
        ClientSecret: "service-account-secret",
        TokenURL:     "https://auth.example.com/token",
        Scopes:       []string{"api.read", "api.write"},
    }

    // Get an HTTP client (auto-refreshes token)
    client := config.Client(ctx)

    // Make API calls
    resp, err := client.Get("https://api.example.com/data")

  USE CASES:
    - Microservice-to-microservice auth
    - Background jobs accessing APIs
    - Service accounts in cloud providers`)

	// ============================================
	// DEVICE FLOW
	// ============================================
	fmt.Println("\n--- Device Authorization Flow ---")

	os.Stdout.WriteString(`
  For devices with limited input (smart TV, IoT, CLI):

    // 1. Request device code
    resp, err := config.DeviceAuth(ctx)
    // resp.UserCode       = "ABCD-1234"
    // resp.VerificationURI = "https://provider.com/device"

    // 2. Show user the code and URL
    fmt.Printf("Go to %s and enter code: %s\n",
        resp.VerificationURI, resp.UserCode)

    // 3. Poll for token (blocks until user authorizes)
    token, err := config.DeviceAccessToken(ctx, resp)

  USER EXPERIENCE:
    1. CLI shows: "Go to https://github.com/login/device"
    2. CLI shows: "Enter code: ABCD-1234"
    3. User opens URL on phone/computer
    4. User enters code and authorizes
    5. CLI receives token automatically`)

	// ============================================
	// TOKEN MANAGEMENT
	// ============================================
	fmt.Println("\n--- Token Management ---")

	fmt.Println(`
  TOKEN STRUCTURE:

    type Token struct {
        AccessToken  string
        TokenType    string    // "Bearer"
        RefreshToken string
        Expiry       time.Time
    }

    token.Valid()  // checks if not expired

  TOKEN SOURCE (auto-refresh):

    // TokenSource automatically refreshes expired tokens
    tokenSource := config.TokenSource(ctx, token)
    newToken, err := tokenSource.Token()

    // ReuseTokenSource caches valid tokens
    reusable := oauth2.ReuseTokenSource(token, tokenSource)

    // Create client with TokenSource
    client := oauth2.NewClient(ctx, reusable)

  CUSTOM TOKEN SOURCE:

    type myTokenSource struct { ... }

    func (s *myTokenSource) Token() (*oauth2.Token, error) {
        // Load token from database, file, etc.
        return loadTokenFromDB()
    }

  PERSISTING TOKENS:

    // Save token (e.g., to JSON file)
    data, _ := json.Marshal(token)
    os.WriteFile("token.json", data, 0600)

    // Load token
    data, _ := os.ReadFile("token.json")
    var token oauth2.Token
    json.Unmarshal(data, &token)

    // Create client with loaded token
    client := config.Client(ctx, &token)`)

	// ============================================
	// GOOGLE-SPECIFIC
	// ============================================
	fmt.Println("\n--- Google-Specific OAuth2 ---")

	fmt.Println(`
  Google provides extra utilities:

    import "golang.org/x/oauth2/google"

    // Service account from JSON key file
    data, _ := os.ReadFile("service-account.json")
    config, err := google.JWTConfigFromJSON(data,
        "https://www.googleapis.com/auth/cloud-platform",
    )
    client := config.Client(ctx)

    // Default credentials (auto-detects environment)
    // Works on GCE, GKE, Cloud Run, local dev
    creds, err := google.FindDefaultCredentials(ctx,
        "https://www.googleapis.com/auth/cloud-platform",
    )
    client := oauth2.NewClient(ctx, creds.TokenSource)

    // Application Default Credentials (ADC)
    // 1. Checks GOOGLE_APPLICATION_CREDENTIALS env var
    // 2. Checks well-known file location
    // 3. Checks metadata server (on GCE/GKE)`)

	// ============================================
	// GOLANG.ORG/X/TERM
	// ============================================
	fmt.Println("\n--- golang.org/x/term Overview ---")

	os.Stdout.WriteString(`
  x/term provides terminal handling utilities.

  Install: go get golang.org/x/term

  KEY FUNCTIONS:

    import "golang.org/x/term"

    // Check if file descriptor is a terminal
    if term.IsTerminal(int(os.Stdin.Fd())) {
        fmt.Println("Running in a terminal")
    } else {
        fmt.Println("Input is piped")
    }

    // Get terminal dimensions
    width, height, err := term.GetSize(int(os.Stdout.Fd()))
    fmt.Printf("Terminal: %dx%d\n", width, height)

    // Read password (no echo)
    fmt.Print("Password: ")
    password, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Printf("\nYou entered %d characters\n", len(password))
`)

	// Working terminal detection example
	fmt.Println("\n  Working example (terminal detection):")
	fd := int(os.Stdout.Fd())
	os.Stdout.WriteString(fmt.Sprintf("    stdout fd: %d\n", fd))
	// Note: we can't use term.IsTerminal without importing x/term,
	// but we can show the pattern
	fmt.Println("    (use term.IsTerminal(fd) to check if terminal)")

	fmt.Println(`
  RAW MODE:

    // Switch terminal to raw mode (no line buffering)
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    defer term.Restore(int(os.Stdin.Fd()), oldState)

    // In raw mode:
    // - No line buffering (each keypress is immediate)
    // - No echo (characters not printed)
    // - No signal processing (Ctrl+C doesn't send SIGINT)
    // Perfect for building TUI applications

  TERMINAL STATE:

    // Save terminal state
    state, err := term.GetState(int(os.Stdin.Fd()))

    // Restore terminal state (defer this!)
    term.Restore(int(os.Stdin.Fd()), state)

  TERMINAL (full-featured):

    // Create a terminal with prompt
    t := term.NewTerminal(readWriter, "prompt> ")

    // Read a line
    line, err := t.ReadLine()

    // Read password
    password, err := t.ReadPassword("Password: ")

    // Set terminal size (for word wrapping)
    t.SetSize(80, 24)

  USE CASES:
    - CLI password input (ReadPassword)
    - Detect piped input vs interactive (IsTerminal)
    - TUI applications (MakeRaw)
    - Responsive output (GetSize)
    - Interactive prompts (Terminal)`)

	// ============================================
	// BUILDING A CLI WITH OAUTH2 + TERM
	// ============================================
	fmt.Println("\n--- Pattern: CLI with OAuth2 + term ---")

	os.Stdout.WriteString(`
  COMPLETE CLI LOGIN PATTERN:

    func login() error {
        config := &oauth2.Config{
            ClientID: "cli-app",
            Endpoint: github.Endpoint,
            Scopes:   []string{"user:email"},
        }

        // Try device flow first (best UX)
        resp, err := config.DeviceAuth(ctx)
        if err != nil {
            return fallbackToLocalServer(config)
        }

        fmt.Printf("Open %s\n", resp.VerificationURI)
        fmt.Printf("Enter code: %s\n", resp.UserCode)

        token, err := config.DeviceAccessToken(ctx, resp)
        if err != nil { return err }

        // Save token securely
        return saveToken(token)
    }

    func getClient() *http.Client {
        token := loadToken()
        if token == nil {
            // Interactive login
            if !term.IsTerminal(int(os.Stdin.Fd())) {
                log.Fatal("not logged in, run: myapp login")
            }
            login()
            token = loadToken()
        }
        return config.Client(ctx, token)
    }
` + "\n")

	// ============================================
	// SECURITY BEST PRACTICES
	// ============================================
	fmt.Println("--- OAuth2 Security Best Practices ---")

	fmt.Println(`
  1. ALWAYS use PKCE (even for confidential clients)
  2. ALWAYS validate state parameter (CSRF protection)
  3. NEVER store tokens in localStorage (XSS vulnerable)
  4. Use short-lived access tokens + refresh tokens
  5. Store client secrets in environment variables
  6. Use HTTPS for all OAuth2 endpoints
  7. Validate redirect URIs strictly
  8. Request minimal scopes needed
  9. Implement token revocation on logout
  10. Use secure storage for tokens (keychain, encrypted file)`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  x/oauth2:
    Config            - OAuth2 client configuration
    AuthCodeURL       - Generate authorization URL
    Exchange          - Exchange code for token
    Client            - HTTP client with auto-refresh
    TokenSource       - Interface for providing tokens
    GenerateVerifier  - PKCE verifier generation

  Flows:
    Authorization Code - Web apps (+ PKCE)
    Client Credentials - Server-to-server
    Device Flow        - CLI/IoT devices

  Providers:
    google.Endpoint    - Google OAuth2
    github.Endpoint    - GitHub OAuth2

  x/term:
    IsTerminal(fd)     - Check if terminal
    GetSize(fd)        - Terminal dimensions
    ReadPassword(fd)   - Read without echo
    MakeRaw(fd)        - Enter raw mode
    Restore(fd, state) - Restore terminal state
    NewTerminal(rw)    - Full terminal emulation

  Install:
    go get golang.org/x/oauth2
    go get golang.org/x/term`)
}

// Ensure imports are used
var _ = strings.NewReader
