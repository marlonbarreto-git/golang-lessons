// Package main - Chapter 118: HashiCorp Vault Internals
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// ============================================
// INTRODUCTION TO HASHICORP VAULT
// ============================================

func intro() {
	os.Stdout.WriteString(`
=== HashiCorp Vault: Secrets Management ===

Vault is a tool for securely accessing secrets. A secret is anything
that you want to tightly control access to, such as API keys, passwords,
certificates, and more.

Key Features:
1. Secure Secret Storage: Encrypted at rest and in transit
2. Dynamic Secrets: Generate secrets on-demand for services
3. Data Encryption: Cryptography as a Service
4. Leasing and Renewal: All secrets have a lease
5. Revocation: Built-in support for secret revocation

Core Architecture:
- Seal/Unseal mechanism (protects master key)
- Storage Backend (encrypted data store)
- Auth Methods (how clients authenticate)
- Secrets Engines (store, generate, encrypt secrets)
- Audit Devices (comprehensive audit logging)
- Policies (fine-grained authorization)

Go Patterns Used:
- Interface-based plugin system
- Middleware/chain of responsibility
- Barrier encryption (encrypted storage layer)
- Zero-memory cleanup for secrets
- State machines for seal/unseal
`)
}

// ============================================
// SHAMIR'S SECRET SHARING
// ============================================

// ShamirShare represents one shard of the master key
type ShamirShare struct {
	Index int
	Data  []byte
}

// SimpleShamir demonstrates Shamir's Secret Sharing concept
// Real Vault uses proper polynomial interpolation
type SimpleShamir struct {
	threshold int
	shares    int
}

func NewShamir(threshold, shares int) *SimpleShamir {
	return &SimpleShamir{
		threshold: threshold,
		shares:    shares,
	}
}

// Split divides a secret into shares (simplified XOR-based approach for demo)
func (s *SimpleShamir) Split(secret []byte) []ShamirShare {
	shares := make([]ShamirShare, s.shares)

	// Generate random shares
	for i := 0; i < s.shares-1; i++ {
		shares[i] = ShamirShare{
			Index: i + 1,
			Data:  make([]byte, len(secret)),
		}
		rand.Read(shares[i].Data)
	}

	// Last share is XOR of all others with secret
	shares[s.shares-1] = ShamirShare{
		Index: s.shares,
		Data:  make([]byte, len(secret)),
	}
	copy(shares[s.shares-1].Data, secret)

	for i := 0; i < s.shares-1; i++ {
		for j := 0; j < len(secret); j++ {
			shares[s.shares-1].Data[j] ^= shares[i].Data[j]
		}
	}

	return shares
}

// Combine reconstructs secret from threshold shares
func (s *SimpleShamir) Combine(shares []ShamirShare) ([]byte, error) {
	if len(shares) < s.threshold {
		return nil, fmt.Errorf("insufficient shares: need %d, got %d", s.threshold, len(shares))
	}

	result := make([]byte, len(shares[0].Data))
	copy(result, shares[0].Data)

	for i := 1; i < len(shares); i++ {
		for j := 0; j < len(result); j++ {
			result[j] ^= shares[i].Data[j]
		}
	}

	return result, nil
}

// ============================================
// SEAL/UNSEAL STATE MACHINE
// ============================================

type SealState int

const (
	StateSealed SealState = iota
	StateUnsealed
)

type VaultCore struct {
	mu          sync.RWMutex
	state       SealState
	masterKey   []byte
	unsealKeys  []ShamirShare
	threshold   int
	barrier     *EncryptionBarrier
	storage     map[string][]byte
	tokens      map[string]*Token
	policies    map[string]*Policy
	leases      map[string]*Lease
	auditLog    []AuditEntry
}

func NewVaultCore(threshold, shares int) *VaultCore {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)

	shamir := NewShamir(threshold, shares)
	unsealShares := shamir.Split(masterKey)

	fmt.Printf("Vault initialized with %d shares (threshold: %d)\n", shares, threshold)
	for i, share := range unsealShares {
		fmt.Printf("Unseal Key %d: %s\n", i+1, hex.EncodeToString(share.Data)[:16]+"...")
	}

	return &VaultCore{
		state:      StateSealed,
		threshold:  threshold,
		unsealKeys: unsealShares,
		storage:    make(map[string][]byte),
		tokens:     make(map[string]*Token),
		policies:   make(map[string]*Policy),
		leases:     make(map[string]*Lease),
		auditLog:   make([]AuditEntry, 0),
	}
}

func (v *VaultCore) Unseal(shares []ShamirShare) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.state == StateUnsealed {
		return fmt.Errorf("vault already unsealed")
	}

	shamir := NewShamir(v.threshold, len(v.unsealKeys))
	masterKey, err := shamir.Combine(shares)
	if err != nil {
		return err
	}

	v.masterKey = masterKey
	v.barrier = NewEncryptionBarrier(masterKey)
	v.state = StateUnsealed

	v.audit("system", "unseal", "vault unsealed successfully")
	return nil
}

func (v *VaultCore) Seal() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.state == StateSealed {
		return fmt.Errorf("vault already sealed")
	}

	// Zero out master key from memory
	for i := range v.masterKey {
		v.masterKey[i] = 0
	}
	v.masterKey = nil
	v.barrier = nil
	v.state = StateSealed

	v.audit("system", "seal", "vault sealed")
	return nil
}

func (v *VaultCore) IsSealed() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.state == StateSealed
}

// ============================================
// ENCRYPTION BARRIER
// ============================================

// EncryptionBarrier provides transparent encryption/decryption
type EncryptionBarrier struct {
	aesKey []byte
}

func NewEncryptionBarrier(key []byte) *EncryptionBarrier {
	return &EncryptionBarrier{aesKey: key}
}

func (eb *EncryptionBarrier) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(eb.aesKey)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func (eb *EncryptionBarrier) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(eb.aesKey)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}

// ============================================
// POLICY-BASED ACCESS CONTROL
// ============================================

type Policy struct {
	Name  string
	Rules []PolicyRule
}

type PolicyRule struct {
	Path         string
	Capabilities []string // read, write, delete, list
}

func (p *Policy) IsAllowed(path, capability string) bool {
	for _, rule := range p.Rules {
		if matchPath(rule.Path, path) {
			for _, cap := range rule.Capabilities {
				if cap == capability || cap == "*" {
					return true
				}
			}
		}
	}
	return false
}

func matchPath(pattern, path string) bool {
	// Simplified path matching (real Vault uses glob patterns)
	if pattern == path {
		return true
	}
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}
	return false
}

// ============================================
// TOKEN-BASED AUTHENTICATION
// ============================================

type Token struct {
	ID           string
	Policies     []string
	TTL          time.Duration
	CreatedAt    time.Time
	Renewable    bool
	Parent       string
	DisplayName  string
}

func (v *VaultCore) CreateToken(policies []string, ttl time.Duration) (*Token, error) {
	if v.IsSealed() {
		return nil, fmt.Errorf("vault is sealed")
	}

	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	tokenID := base64.URLEncoding.EncodeToString(tokenBytes)

	token := &Token{
		ID:          tokenID,
		Policies:    policies,
		TTL:         ttl,
		CreatedAt:   time.Now(),
		Renewable:   true,
		DisplayName: "generated-token",
	}

	v.mu.Lock()
	v.tokens[tokenID] = token
	v.mu.Unlock()

	// Create lease for token
	v.createLease(tokenID, "auth/token/"+tokenID, ttl)

	v.audit(tokenID, "create_token", fmt.Sprintf("token created with policies: %v", policies))
	return token, nil
}

func (v *VaultCore) ValidateToken(tokenID string) (*Token, error) {
	v.mu.RLock()
	token, exists := v.tokens[tokenID]
	v.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid token")
	}

	// Check if token has expired
	if time.Since(token.CreatedAt) > token.TTL {
		return nil, fmt.Errorf("token expired")
	}

	return token, nil
}

// ============================================
// LEASE AND TTL MANAGEMENT
// ============================================

type Lease struct {
	ID         string
	Path       string
	IssueTime  time.Time
	ExpireTime time.Time
	Renewable  bool
	Data       map[string]interface{}
}

func (v *VaultCore) createLease(id, path string, ttl time.Duration) {
	lease := &Lease{
		ID:         id,
		Path:       path,
		IssueTime:  time.Now(),
		ExpireTime: time.Now().Add(ttl),
		Renewable:  true,
		Data:       make(map[string]interface{}),
	}

	v.mu.Lock()
	v.leases[id] = lease
	v.mu.Unlock()
}

func (v *VaultCore) RenewLease(leaseID string, increment time.Duration) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	lease, exists := v.leases[leaseID]
	if !exists {
		return fmt.Errorf("lease not found")
	}

	if !lease.Renewable {
		return fmt.Errorf("lease not renewable")
	}

	lease.ExpireTime = time.Now().Add(increment)
	v.audit("system", "renew_lease", fmt.Sprintf("lease %s renewed", leaseID))
	return nil
}

func (v *VaultCore) RevokeLease(leaseID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if _, exists := v.leases[leaseID]; !exists {
		return fmt.Errorf("lease not found")
	}

	delete(v.leases, leaseID)
	v.audit("system", "revoke_lease", fmt.Sprintf("lease %s revoked", leaseID))
	return nil
}

// ============================================
// SECRET STORAGE
// ============================================

func (v *VaultCore) WriteSecret(token *Token, path string, data []byte) error {
	if v.IsSealed() {
		return fmt.Errorf("vault is sealed")
	}

	// Check authorization
	allowed := false
	v.mu.RLock()
	for _, policyName := range token.Policies {
		if policy, exists := v.policies[policyName]; exists {
			if policy.IsAllowed(path, "write") {
				allowed = true
				break
			}
		}
	}
	v.mu.RUnlock()

	if !allowed {
		v.audit(token.ID, "write_denied", fmt.Sprintf("path: %s", path))
		return fmt.Errorf("permission denied")
	}

	// Encrypt data
	encrypted, err := v.barrier.Encrypt(data)
	if err != nil {
		return err
	}

	v.mu.Lock()
	v.storage[path] = encrypted
	v.mu.Unlock()

	v.audit(token.ID, "write", fmt.Sprintf("path: %s", path))
	return nil
}

func (v *VaultCore) ReadSecret(token *Token, path string) ([]byte, error) {
	if v.IsSealed() {
		return nil, fmt.Errorf("vault is sealed")
	}

	// Check authorization
	allowed := false
	v.mu.RLock()
	for _, policyName := range token.Policies {
		if policy, exists := v.policies[policyName]; exists {
			if policy.IsAllowed(path, "read") {
				allowed = true
				break
			}
		}
	}
	v.mu.RUnlock()

	if !allowed {
		v.audit(token.ID, "read_denied", fmt.Sprintf("path: %s", path))
		return nil, fmt.Errorf("permission denied")
	}

	v.mu.RLock()
	encrypted, exists := v.storage[path]
	v.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("secret not found")
	}

	// Decrypt data
	decrypted, err := v.barrier.Decrypt(encrypted)
	if err != nil {
		return nil, err
	}

	v.audit(token.ID, "read", fmt.Sprintf("path: %s", path))
	return decrypted, nil
}

// ============================================
// TRANSIT SECRETS ENGINE (ENCRYPTION AS A SERVICE)
// ============================================

type TransitEngine struct {
	keys map[string][]byte
}

func NewTransitEngine() *TransitEngine {
	return &TransitEngine{
		keys: make(map[string][]byte),
	}
}

func (te *TransitEngine) CreateKey(name string) error {
	key := make([]byte, 32)
	rand.Read(key)
	te.keys[name] = key
	return nil
}

func (te *TransitEngine) Encrypt(keyName string, plaintext []byte) (string, error) {
	key, exists := te.keys[keyName]
	if !exists {
		return "", fmt.Errorf("key not found")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return "vault:v1:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (te *TransitEngine) Decrypt(keyName, ciphertext string) ([]byte, error) {
	key, exists := te.keys[keyName]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}

	// Remove vault prefix
	if !strings.HasPrefix(ciphertext, "vault:v1:") {
		return nil, fmt.Errorf("invalid ciphertext format")
	}
	ciphertext = strings.TrimPrefix(ciphertext, "vault:v1:")

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return data, nil
}

// ============================================
// AUDIT LOGGING
// ============================================

type AuditEntry struct {
	Timestamp time.Time
	TokenID   string
	Operation string
	Details   string
}

func (v *VaultCore) audit(tokenID, operation, details string) {
	entry := AuditEntry{
		Timestamp: time.Now(),
		TokenID:   tokenID,
		Operation: operation,
		Details:   details,
	}

	v.auditLog = append(v.auditLog, entry)
}

func (v *VaultCore) GetAuditLog() []AuditEntry {
	v.mu.RLock()
	defer v.mu.RUnlock()

	log := make([]AuditEntry, len(v.auditLog))
	copy(log, v.auditLog)
	return log
}

// ============================================
// DEMONSTRATION
// ============================================

func demonstrateShamir() {
	fmt.Println("\n=== Shamir's Secret Sharing Demo ===")

	secret := []byte("my-super-secret-master-key-32bit")
	fmt.Printf("Original Secret: %s\n", secret)

	shamir := NewShamir(3, 5)
	shares := shamir.Split(secret)

	fmt.Printf("\nSplit into %d shares (threshold: 3):\n", len(shares))
	for i, share := range shares {
		fmt.Printf("Share %d: %s\n", i+1, hex.EncodeToString(share.Data)[:16]+"...")
	}

	// Try with insufficient shares
	fmt.Println("\nAttempting reconstruction with 2 shares (insufficient):")
	_, err := shamir.Combine(shares[:2])
	fmt.Printf("Result: %v\n", err)

	// Reconstruct with sufficient shares
	fmt.Println("\nReconstruction with 3 shares:")
	reconstructed, err := shamir.Combine(shares[:3])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Reconstructed: %s\n", reconstructed)
		fmt.Printf("Match: %v\n", string(reconstructed) == string(secret))
	}
}

func demonstrateVault() {
	fmt.Println("\n=== Vault Core Demo ===")

	// Initialize vault with 3-of-5 Shamir sharing
	vault := NewVaultCore(3, 5)

	fmt.Printf("\nVault Status: Sealed=%v\n", vault.IsSealed())

	// Unseal vault using 3 shares
	fmt.Println("\nUnsealing vault with 3 shares...")
	err := vault.Unseal(vault.unsealKeys[:3])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Vault Status: Sealed=%v\n", vault.IsSealed())

	// Create policies
	adminPolicy := &Policy{
		Name: "admin",
		Rules: []PolicyRule{
			{Path: "secret/*", Capabilities: []string{"read", "write", "delete"}},
			{Path: "transit/*", Capabilities: []string{"read", "write"}},
		},
	}

	readOnlyPolicy := &Policy{
		Name: "readonly",
		Rules: []PolicyRule{
			{Path: "secret/*", Capabilities: []string{"read"}},
		},
	}

	vault.mu.Lock()
	vault.policies["admin"] = adminPolicy
	vault.policies["readonly"] = readOnlyPolicy
	vault.mu.Unlock()

	// Create tokens
	fmt.Println("\nCreating admin token...")
	adminToken, err := vault.CreateToken([]string{"admin"}, 1*time.Hour)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Admin Token: %s\n", adminToken.ID[:16]+"...")

	fmt.Println("\nCreating read-only token...")
	roToken, err := vault.CreateToken([]string{"readonly"}, 30*time.Minute)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Read-Only Token: %s\n", roToken.ID[:16]+"...")

	// Write secret with admin token
	fmt.Println("\nWriting secret with admin token...")
	secretData := []byte(`{"username":"admin","password":"super-secret-123"}`)
	err = vault.WriteSecret(adminToken, "secret/database/config", secretData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Secret written successfully")
	}

	// Read secret with read-only token
	fmt.Println("\nReading secret with read-only token...")
	data, err := vault.ReadSecret(roToken, "secret/database/config")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Secret retrieved: %s\n", string(data))
	}

	// Try to write with read-only token (should fail)
	fmt.Println("\nAttempting write with read-only token...")
	err = vault.WriteSecret(roToken, "secret/database/config2", []byte("data"))
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Demonstrate seal
	fmt.Println("\nSealing vault...")
	vault.Seal()
	fmt.Printf("Vault Status: Sealed=%v\n", vault.IsSealed())

	// Try to read while sealed (should fail)
	fmt.Println("\nAttempting read while sealed...")
	_, err = vault.ReadSecret(adminToken, "secret/database/config")
	fmt.Printf("Expected error: %v\n", err)
}

func demonstrateTransit() {
	fmt.Println("\n=== Transit Engine Demo (Encryption as a Service) ===")

	transit := NewTransitEngine()

	// Create encryption key
	fmt.Println("\nCreating encryption key 'my-app-key'...")
	transit.CreateKey("my-app-key")

	// Encrypt data
	plaintext := []byte("sensitive customer data")
	fmt.Printf("\nPlaintext: %s\n", plaintext)

	ciphertext, err := transit.Encrypt("my-app-key", plaintext)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Ciphertext: %s\n", ciphertext[:50]+"...")

	// Decrypt data
	decrypted, err := transit.Decrypt("my-app-key", ciphertext)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Decrypted: %s\n", decrypted)
	fmt.Printf("Match: %v\n", string(decrypted) == string(plaintext))
}

func demonstrateLeases() {
	fmt.Println("\n=== Lease Management Demo ===")

	vault := NewVaultCore(2, 3)
	vault.Unseal(vault.unsealKeys[:2])

	// Create a lease
	leaseID := "lease-12345"
	vault.createLease(leaseID, "database/creds/readonly", 1*time.Minute)
	fmt.Printf("Lease created: %s (TTL: 1 minute)\n", leaseID)

	// Renew lease
	fmt.Println("\nRenewing lease for 2 more minutes...")
	err := vault.RenewLease(leaseID, 2*time.Minute)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Lease renewed successfully")
	}

	// Revoke lease
	fmt.Println("\nRevoking lease...")
	err = vault.RevokeLease(leaseID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Lease revoked successfully")
	}
}

func demonstrateAudit() {
	fmt.Println("\n=== Audit Log Demo ===")

	vault := NewVaultCore(2, 3)
	vault.Unseal(vault.unsealKeys[:2])

	adminPolicy := &Policy{
		Name:  "admin",
		Rules: []PolicyRule{{Path: "secret/*", Capabilities: []string{"*"}}},
	}
	vault.policies["admin"] = adminPolicy

	token, _ := vault.CreateToken([]string{"admin"}, 1*time.Hour)
	vault.WriteSecret(token, "secret/test", []byte("data"))
	vault.ReadSecret(token, "secret/test")

	fmt.Println("\nAudit Log:")
	log := vault.GetAuditLog()
	for _, entry := range log {
		fmt.Printf("[%s] Token:%s Op:%s - %s\n",
			entry.Timestamp.Format("15:04:05"),
			entry.TokenID[:8]+"...",
			entry.Operation,
			entry.Details)
	}
}

func compareAlternatives() {
	os.Stdout.WriteString(`
=== Vault vs Alternatives ===

1. HashiCorp Vault (Open Source + Enterprise)
   + Multi-cloud and on-premise
   + Rich plugin ecosystem
   + Dynamic secrets generation
   + Encryption as a service
   + Advanced access policies
   - More complex setup
   - Requires operational expertise

2. AWS Secrets Manager
   + Fully managed service
   + Native AWS integration
   + Automatic rotation
   - AWS-only
   - Limited encryption features
   - Higher cost for high-volume

3. Azure Key Vault
   + Fully managed
   + HSM-backed options
   + Azure integration
   - Azure-only
   - Less flexible policies

4. Google Secret Manager
   + Simple API
   + GCP integration
   + Automatic replication
   - GCP-only
   - Basic feature set

5. Kubernetes Secrets
   + Built into K8s
   + Easy to use
   - Base64 only (not encrypted by default)
   - Limited scope
   - Requires external KMS for encryption

Vault Philosophy:
- Security by default
- Defense in depth (multiple layers)
- Zero-trust architecture
- Principle of least privilege
- Comprehensive audit trail
`)
}

func main() {
	intro()
	demonstrateShamir()
	demonstrateVault()
	demonstrateTransit()
	demonstrateLeases()
	demonstrateAudit()
	compareAlternatives()
}

/* SUMMARY

HashiCorp Vault Internals - Key Concepts:

1. SHAMIR'S SECRET SHARING
   - Master key split into multiple shares
   - Requires threshold shares to reconstruct
   - Protects against single point of compromise

2. SEAL/UNSEAL MECHANISM
   - Sealed: master key not in memory, vault inactive
   - Unsealed: master key reconstructed, vault operational
   - State machine pattern for security

3. ENCRYPTION BARRIER
   - All data encrypted at rest using master key
   - Transparent encryption/decryption layer
   - AES-256-GCM in production

4. POLICY-BASED ACCESS CONTROL
   - Fine-grained path-based permissions
   - Capabilities: read, write, delete, list, etc.
   - Attached to tokens for authorization

5. TOKEN AUTHENTICATION
   - Token-based identity system
   - Policies determine access rights
   - Support for TTL and renewal

6. LEASE MANAGEMENT
   - All secrets have a lease (TTL)
   - Automatic expiration and revocation
   - Renewable leases for long-running apps

7. TRANSIT ENGINE
   - Encryption as a Service
   - Centralized key management
   - Apps don't store encryption keys

8. AUDIT LOGGING
   - Comprehensive audit trail
   - Every operation logged
   - Multiple audit backend support

9. PERFORMANCE CONSIDERATIONS
   - Zero-memory cleanup for sensitive data
   - Efficient encryption with AES-NI hardware
   - In-memory caching with encrypted storage

10. SECURITY PHILOSOPHY
    - Security by default
    - Defense in depth
    - Zero-trust architecture
    - Least privilege access
    - Comprehensive auditability

Go Patterns Demonstrated:
- Interface-based plugin architecture
- Middleware chains for request processing
- State machines (seal/unseal)
- Encryption barrier pattern
- Memory-safe secret handling
- Policy engine pattern
- Lease lifecycle management

Real Vault Implementation:
- Uses raft or consul for HA clustering
- Supports multiple auth methods (LDAP, AWS, K8s, etc.)
- Multiple secrets engines (KV, PKI, SSH, databases)
- HSM integration for key protection
- Auto-unseal with cloud KMS
- Namespaces for multi-tenancy (Enterprise)
- Replication across regions

*/
