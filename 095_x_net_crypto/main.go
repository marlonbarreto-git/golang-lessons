// Package main - Chapter 095: golang.org/x/net and golang.org/x/crypto
// x/net extends networking with HTTP/2, HTML parsing, and more.
// x/crypto provides modern cryptographic algorithms beyond the stdlib.
package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== GOLANG.ORG/X/NET AND GOLANG.ORG/X/CRYPTO ===")

	// ============================================
	// GOLANG.ORG/X/NET OVERVIEW
	// ============================================
	fmt.Println("\n--- golang.org/x/net Overview ---")

	fmt.Println(`
  golang.org/x/net provides supplementary networking libraries:

  KEY PACKAGES:
    x/net/http2      - HTTP/2 protocol support
    x/net/html       - HTML5-compliant tokenizer and parser
    x/net/context    - (legacy, now in stdlib as context)
    x/net/websocket  - WebSocket protocol
    x/net/proxy      - SOCKS5 proxy support
    x/net/netutil    - Network utilities (LimitListener)

  Install: go get golang.org/x/net`)

	// ============================================
	// HTTP/2 WITH X/NET
	// ============================================
	fmt.Println("\n--- x/net/http2: HTTP/2 Protocol ---")

	fmt.Println(`
  Go's net/http already supports HTTP/2 for clients.
  x/net/http2 gives fine-grained control:

  SERVER SETUP:

    import "golang.org/x/net/http2"

    server := &http.Server{Addr: ":8443"}
    // Configure HTTP/2 on existing server
    http2.ConfigureServer(server, &http2.Server{
        MaxConcurrentStreams: 100,
        MaxReadFrameSize:    1 << 20,
    })

    // TLS is required for HTTP/2
    server.ListenAndServeTLS("cert.pem", "key.pem")

  CLIENT SETUP:

    transport := &http2.Transport{
        AllowHTTP:          false,
        DisableCompression: false,
        TLSClientConfig:    &tls.Config{},
    }

    client := &http.Client{Transport: transport}
    resp, err := client.Get("https://example.com")

  SERVER PUSH (HTTP/2 feature):

    func handler(w http.ResponseWriter, r *http.Request) {
        if pusher, ok := w.(http.Pusher); ok {
            // Push CSS before browser requests it
            pusher.Push("/styles.css", nil)
            pusher.Push("/script.js", nil)
        }
        fmt.Fprint(w, "<html>...</html>")
    }

  HTTP/2 FEATURES:
    - Multiplexing: multiple requests on one connection
    - Header compression (HPACK)
    - Server push
    - Stream prioritization
    - Binary framing (more efficient than HTTP/1.1 text)`)

	// ============================================
	// HTML PARSING WITH X/NET
	// ============================================
	fmt.Println("\n--- x/net/html: HTML5 Parser ---")

	os.Stdout.WriteString(`
  x/net/html implements an HTML5-compliant tokenizer and parser.
  Essential for web scraping and HTML manipulation.

  PARSING HTML:

    import "golang.org/x/net/html"

    doc, err := html.Parse(strings.NewReader(htmlStr))
    // doc is a *html.Node tree

  NODE TYPES:

    html.ErrorNode       // parse error
    html.TextNode        // text content
    html.DocumentNode    // root document
    html.ElementNode     // HTML element (<div>, <a>, etc.)
    html.CommentNode     // HTML comment
    html.DoctypeNode     // <!DOCTYPE>

  TRAVERSING THE TREE:

    var traverse func(*html.Node)
    traverse = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "a" {
            for _, attr := range n.Attr {
                if attr.Key == "href" {
                    fmt.Println("Link:", attr.Val)
                }
            }
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            traverse(c)
        }
    }
    traverse(doc)

  TOKENIZER (lower-level):

    tokenizer := html.NewTokenizer(reader)
    for {
        tt := tokenizer.Next()
        switch tt {
        case html.ErrorToken:
            return  // end of document
        case html.StartTagToken:
            tn, _ := tokenizer.TagName()
            fmt.Printf("Start tag: %s\n", tn)
        case html.TextToken:
            text := tokenizer.Text()
            fmt.Printf("Text: %s\n", text)
        }
    }

  RENDERING HTML:

    html.Render(writer, node)  // serialize node tree to HTML

  WEB SCRAPING EXAMPLE:

    resp, _ := http.Get("https://example.com")
    defer resp.Body.Close()
    doc, _ := html.Parse(resp.Body)

    var titles []string
    var findTitles func(*html.Node)
    findTitles = func(n *html.Node) {
        if n.Type == html.ElementNode && n.Data == "h1" {
            if n.FirstChild != nil {
                titles = append(titles, n.FirstChild.Data)
            }
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            findTitles(c)
        }
    }
    findTitles(doc)`)

	// ============================================
	// NETUTIL
	// ============================================
	fmt.Println("\n--- x/net/netutil: Network Utilities ---")

	fmt.Println(`
  LimitListener wraps a net.Listener to limit concurrent connections:

    import "golang.org/x/net/netutil"

    listener, err := net.Listen("tcp", ":8080")
    if err != nil { log.Fatal(err) }

    // Limit to 100 concurrent connections
    limited := netutil.LimitListener(listener, 100)

    http.Serve(limited, handler)
    // Connection 101 blocks until one of the 100 finishes

  USE CASES:
    - Prevent server overload
    - Protect downstream services
    - Graceful degradation under load`)

	// ============================================
	// GOLANG.ORG/X/CRYPTO OVERVIEW
	// ============================================
	fmt.Println("\n--- golang.org/x/crypto Overview ---")

	fmt.Println(`
  golang.org/x/crypto provides modern cryptographic algorithms
  not in the standard library:

  KEY PACKAGES:
    x/crypto/bcrypt            - Password hashing (recommended)
    x/crypto/argon2            - Password hashing (strongest)
    x/crypto/chacha20poly1305  - AEAD cipher
    x/crypto/ssh               - SSH client/server
    x/crypto/nacl/box          - Public-key authenticated encryption
    x/crypto/nacl/sign         - Public-key signatures
    x/crypto/scrypt            - Password-based key derivation
    x/crypto/curve25519        - Elliptic curve Diffie-Hellman

  Install: go get golang.org/x/crypto`)

	// ============================================
	// BCRYPT - PASSWORD HASHING
	// ============================================
	fmt.Println("\n--- bcrypt: Password Hashing ---")

	fmt.Println(`
  bcrypt is the RECOMMENDED password hashing algorithm.
  It's slow by design (prevents brute force).

  API:

    import "golang.org/x/crypto/bcrypt"

    // Hash a password
    hash, err := bcrypt.GenerateFromPassword(
        []byte("mypassword"),
        bcrypt.DefaultCost,  // cost=10
    )
    // hash is a []byte like: $2a$10$...

    // Verify password against hash
    err := bcrypt.CompareHashAndPassword(hash, []byte("mypassword"))
    if err != nil {
        // password doesn't match
    }

  COST PARAMETER:
    bcrypt.MinCost     = 4    // fast, insecure
    bcrypt.DefaultCost = 10   // good default
    bcrypt.MaxCost     = 31   // very slow

  CHOOSING COST:
    - Target ~250ms for hashing on your server
    - DefaultCost (10) is fine for most applications
    - Increase by 1 every ~2 years (doubles time)

  COMPLETE AUTH EXAMPLE:

    func RegisterUser(password string) (string, error) {
        hash, err := bcrypt.GenerateFromPassword(
            []byte(password), bcrypt.DefaultCost)
        if err != nil { return "", err }
        return string(hash), nil
    }

    func AuthenticateUser(hash, password string) bool {
        err := bcrypt.CompareHashAndPassword(
            []byte(hash), []byte(password))
        return err == nil
    }`)

	// ============================================
	// ARGON2 - MODERN PASSWORD HASHING
	// ============================================
	fmt.Println("\n--- argon2: Modern Password Hashing ---")

	fmt.Println(`
  argon2 is the winner of the Password Hashing Competition (2015).
  It's the STRONGEST password hash, resistant to GPU attacks.

  TWO VARIANTS:
    argon2id - recommended (hybrid, resistant to side-channels + GPU)
    argon2i  - data-independent (resistant to side-channel attacks)

  API:

    import "golang.org/x/crypto/argon2"

    // Generate a key from password
    salt := make([]byte, 16)
    rand.Read(salt)

    // argon2id: time=1, memory=64MB, threads=4, keyLen=32
    key := argon2.IDKey(
        []byte("password"),
        salt,
        1,         // time iterations
        64 * 1024, // memory in KB (64MB)
        4,         // parallelism (threads)
        32,        // key length
    )

  RECOMMENDED PARAMETERS:
    - time=1, memory=64MB, threads=4 (interactive login)
    - time=3, memory=256MB, threads=4 (sensitive data)

  argon2 vs bcrypt:
    - argon2: stronger, memory-hard (resists GPU/ASIC)
    - bcrypt: simpler API, well-established, widely supported
    - Both are good choices for password hashing`)

	// ============================================
	// CHACHA20-POLY1305
	// ============================================
	fmt.Println("\n--- chacha20poly1305: AEAD Cipher ---")

	fmt.Println(`
  ChaCha20-Poly1305 is an AEAD (Authenticated Encryption with
  Associated Data) cipher. Alternative to AES-GCM.

  API:

    import "golang.org/x/crypto/chacha20poly1305"

    // Create cipher with 256-bit key
    key := make([]byte, chacha20poly1305.KeySize)  // 32 bytes
    rand.Read(key)

    aead, err := chacha20poly1305.New(key)
    // or NewX for extended nonce (192-bit):
    aead, err := chacha20poly1305.NewX(key)

    // Encrypt
    nonce := make([]byte, aead.NonceSize())
    rand.Read(nonce)
    ciphertext := aead.Seal(nil, nonce, plaintext, additionalData)

    // Decrypt
    plaintext, err := aead.Open(nil, nonce, ciphertext, additionalData)

  WHEN TO USE:
    - ChaCha20: faster on CPUs without AES hardware (ARM, mobile)
    - AES-GCM: faster on CPUs with AES-NI (modern x86)
    - XChaCha20: when you need random nonces (larger nonce space)`)

	// ============================================
	// NaCl BOX AND SIGN
	// ============================================
	fmt.Println("\n--- nacl/box and nacl/sign ---")

	fmt.Println(`
  NaCl (Networking and Cryptography library) provides
  high-level, hard-to-misuse cryptographic APIs.

  nacl/box - Public-key authenticated encryption:

    import "golang.org/x/crypto/nacl/box"

    // Generate key pairs
    senderPub, senderPriv, _ := box.GenerateKey(rand.Reader)
    recipientPub, recipientPriv, _ := box.GenerateKey(rand.Reader)

    // Encrypt (sender encrypts for recipient)
    var nonce [24]byte
    rand.Read(nonce[:])
    encrypted := box.Seal(nil, message, &nonce, recipientPub, senderPriv)

    // Decrypt (recipient decrypts from sender)
    decrypted, ok := box.Open(nil, encrypted, &nonce, senderPub, recipientPriv)

  nacl/sign - Public-key signatures:

    import "golang.org/x/crypto/nacl/sign"

    pub, priv, _ := sign.GenerateKey(rand.Reader)

    // Sign
    signed := sign.Sign(nil, message, priv)

    // Verify
    verified, ok := sign.Open(nil, signed, pub)`)

	// ============================================
	// SSH - SECURE SHELL
	// ============================================
	fmt.Println("\n--- x/crypto/ssh: SSH Protocol ---")

	fmt.Println(`
  x/crypto/ssh implements the SSH protocol for building
  SSH clients and servers in Go.

  SSH CLIENT:

    import "golang.org/x/crypto/ssh"

    config := &ssh.ClientConfig{
        User: "admin",
        Auth: []ssh.AuthMethod{
            ssh.Password("secret"),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    // Connect
    client, err := ssh.Dial("tcp", "server:22", config)
    defer client.Close()

    // Run command
    session, err := client.NewSession()
    defer session.Close()
    output, err := session.CombinedOutput("ls -la")

  KEY-BASED AUTH:

    key, err := os.ReadFile(os.ExpandEnv("$HOME/.ssh/id_rsa"))
    signer, err := ssh.ParsePrivateKey(key)

    config := &ssh.ClientConfig{
        User: "admin",
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(signer),
        },
        HostKeyCallback: ssh.FixedHostKey(hostKey),
    }

  SSH SERVER:

    config := &ssh.ServerConfig{
        PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            if c.User() == "admin" && string(pass) == "secret" {
                return nil, nil
            }
            return nil, fmt.Errorf("denied")
        },
    }

    // Add host key
    private, _ := ssh.ParsePrivateKey(hostKeyBytes)
    config.AddHostKey(private)

    // Accept connections
    listener, _ := net.Listen("tcp", ":2222")
    conn, _ := listener.Accept()
    _, chans, reqs, _ := ssh.NewServerConn(conn, config)
    go ssh.DiscardRequests(reqs)

    for newChannel := range chans {
        channel, requests, _ := newChannel.Accept()
        // Handle channel and requests
    }`)

	// ============================================
	// WORKING CRYPTO EXAMPLE (STDLIB)
	// ============================================
	fmt.Println("\n--- Working Example: HMAC Signing (stdlib) ---")

	secret := make([]byte, 32)
	rand.Read(secret)

	message := "Hello, authenticated world!"
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	signature := mac.Sum(nil)
	os.Stdout.WriteString(fmt.Sprintf("  Message:   %s\n", message))
	os.Stdout.WriteString(fmt.Sprintf("  HMAC-SHA256: %s\n", hex.EncodeToString(signature)))

	// Verify
	mac2 := hmac.New(sha256.New, secret)
	mac2.Write([]byte(message))
	valid := hmac.Equal(signature, mac2.Sum(nil))
	os.Stdout.WriteString(fmt.Sprintf("  Signature valid: %t\n", valid))

	// Tampered message
	mac3 := hmac.New(sha256.New, secret)
	mac3.Write([]byte("tampered message"))
	tampered := hmac.Equal(signature, mac3.Sum(nil))
	os.Stdout.WriteString(fmt.Sprintf("  Tampered valid:  %t\n", tampered))

	// Random bytes example
	token := make([]byte, 16)
	rand.Read(token)
	os.Stdout.WriteString(fmt.Sprintf("  Random token: %s\n", hex.EncodeToString(token)))

	// ============================================
	// SECURITY BEST PRACTICES
	// ============================================
	fmt.Println("\n--- Crypto Best Practices ---")

	fmt.Println(`
  PASSWORDS:
    - Use bcrypt or argon2 (NEVER SHA256, MD5, etc.)
    - Always use unique salts (bcrypt does this automatically)
    - Use constant-time comparison (hmac.Equal or subtle.ConstantTimeCompare)

  ENCRYPTION:
    - Use AEAD ciphers (AES-GCM or ChaCha20-Poly1305)
    - NEVER reuse nonces with the same key
    - Use XChaCha20 if generating random nonces
    - Use crypto/rand for all randomness (NEVER math/rand)

  KEYS:
    - Generate with crypto/rand
    - Store securely (environment vars, KMS, vault)
    - Rotate periodically
    - Use proper key derivation (argon2, scrypt, HKDF)

  SSH:
    - NEVER use InsecureIgnoreHostKey() in production
    - Use ssh.FixedHostKey() or known_hosts file
    - Prefer key-based auth over passwords

  GENERAL:
    - Use x/crypto for modern algorithms
    - Use crypto/subtle for constant-time operations
    - Don't implement your own crypto`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  x/net:
    http2        - Fine-grained HTTP/2 control
    html         - HTML5 tokenizer/parser (web scraping)
    netutil      - LimitListener for connection limits

  x/crypto:
    bcrypt       - Password hashing (recommended, simple)
    argon2       - Password hashing (strongest, GPU-resistant)
    chacha20poly1305 - AEAD cipher (alternative to AES-GCM)
    nacl/box     - Public-key encryption
    nacl/sign    - Digital signatures
    ssh          - SSH client/server protocol

  Install:
    go get golang.org/x/net
    go get golang.org/x/crypto`)
}

// Ensure imports are used
var _ = strings.NewReader
var _ = os.Stdout
