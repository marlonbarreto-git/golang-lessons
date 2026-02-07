// Package main - Chapter 055: Crypto
// Criptografia en Go: hashing, HMAC, cifrado simetrico (AES-GCM),
// cifrado asimetrico (RSA, ECDSA, Ed25519), certificados x509,
// TLS, derivacion de claves (PBKDF2, Argon2), firmas digitales,
// comparacion segura y cumplimiento FIPS 140-3.
package main

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== CRYPTO PACKAGE DEEP DIVE ===")

	// ============================================
	// 1. CRYPTO/RAND - SECURE RANDOM BYTES
	// ============================================
	fmt.Println("\n--- 1. crypto/rand - Secure Random Bytes ---")
	os.Stdout.WriteString(`
CRYPTO/RAND vs MATH/RAND:

  crypto/rand:
  - Usa fuente de entropia del sistema operativo
  - /dev/urandom en Linux, CryptGenRandom en Windows
  - Criptograficamente seguro (CSPRNG)
  - Mas lento, pero seguro para tokens, claves, IVs

  math/rand:
  - Pseudoaleatorio deterministico (PRNG)
  - NO seguro para criptografia
  - Rapido, bueno para simulaciones, tests, juegos
  - NUNCA usar para seguridad

FUNCIONES PRINCIPALES:

  rand.Read(b []byte) (int, error)  // Llenar slice con bytes aleatorios
  rand.Int(rand io.Reader, max *big.Int) (*big.Int, error)  // Entero aleatorio [0, max)
  rand.Prime(rand io.Reader, bits int) (*big.Int, error)    // Primo aleatorio de N bits
  rand.Reader                        // io.Reader global de entropia

USOS COMUNES:

  // Token de 32 bytes (256 bits)
  token := make([]byte, 32)
  _, err := rand.Read(token)

  // Numero aleatorio seguro en rango [0, 100)
  n, err := rand.Int(rand.Reader, big.NewInt(100))

  // Generar primo de 2048 bits
  prime, err := rand.Prime(rand.Reader, 2048)
`)

	fmt.Println("  Generando bytes aleatorios seguros:")
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Println("  32 bytes aleatorios (hex):", hex.EncodeToString(randomBytes))
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	fmt.Println("  Numero aleatorio seguro [0, 1000000):", n)

	// ============================================
	// 2. HASHING - SHA256, SHA512, MD5
	// ============================================
	fmt.Println("\n--- 2. Hashing - SHA256, SHA512, MD5 ---")
	os.Stdout.WriteString(`
HASHING = Funcion unidireccional que produce digest de tamano fijo.

ALGORITMOS DISPONIBLES:

  crypto/md5     -> 128 bits (16 bytes)  - INSEGURO, solo checksums
  crypto/sha1    -> 160 bits (20 bytes)  - INSEGURO, solo checksums
  crypto/sha256  -> 256 bits (32 bytes)  - Recomendado, uso general
  crypto/sha512  -> 512 bits (64 bytes)  - Mayor seguridad

DOS FORMAS DE USAR:

  // 1. Simple - Todo en memoria
  hash := sha256.Sum256([]byte("Hello"))

  // 2. Streaming - Para archivos grandes
  h := sha256.New()
  h.Write([]byte("Hello "))
  h.Write([]byte("World"))
  sum := h.Sum(nil)

PROPIEDADES:
  - Determinista: mismo input = mismo output siempre
  - Irreversible: no se puede obtener el input del hash
  - Avalancha: un bit de cambio cambia ~50%% del hash
  - Resistente a colisiones: dificil encontrar dos inputs con mismo hash
`)

	data := []byte("Hello, Go Crypto!")

	md5Sum := md5.Sum(data)
	fmt.Println("  MD5:    ", hex.EncodeToString(md5Sum[:]))

	sha256Sum := sha256.Sum256(data)
	fmt.Println("  SHA-256:", hex.EncodeToString(sha256Sum[:]))

	sha512Sum := sha512.Sum512(data)
	fmt.Println("  SHA-512:", hex.EncodeToString(sha512Sum[:]))

	fmt.Println("\n  Hashing con streaming (sha256.New):")
	h := sha256.New()
	h.Write([]byte("Hello, "))
	h.Write([]byte("Go Crypto!"))
	streamHash := h.Sum(nil)
	fmt.Println("  SHA-256 streaming:", hex.EncodeToString(streamHash))
	fmt.Println("  Ambos son iguales:", hex.EncodeToString(sha256Sum[:]) == hex.EncodeToString(streamHash))

	// ============================================
	// 3. HMAC - MESSAGE AUTHENTICATION CODE
	// ============================================
	fmt.Println("\n--- 3. HMAC - Message Authentication Code ---")
	os.Stdout.WriteString(`
HMAC = Hash-based Message Authentication Code

Combina un hash con una clave secreta para:
  1. Verificar integridad del mensaje (no fue alterado)
  2. Verificar autenticidad (viene de quien tiene la clave)

NO proporciona:
  - Cifrado (el mensaje sigue visible)
  - No-repudio (ambas partes tienen la clave)

USOS COMUNES:
  - Firmar cookies de sesion
  - Verificar webhooks (GitHub, Stripe)
  - API authentication (AWS Signature V4)
  - JWT con algoritmo HS256

FUNCIONES:

  hmac.New(hash func() hash.Hash, key []byte) hash.Hash
  hmac.Equal(mac1, mac2 []byte) bool  // Comparacion constant-time

IMPORTANTE: Siempre usar hmac.Equal() para comparar MACs.
  - strings.Compare o == son vulnerables a timing attacks
  - hmac.Equal usa crypto/subtle internamente
`)

	key := []byte("mi-clave-secreta-de-32-bytes!!!!!")
	message := []byte("Transferir $1000 a cuenta 12345")

	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	signature := mac.Sum(nil)
	fmt.Println("  Mensaje:", string(message))
	fmt.Println("  HMAC-SHA256:", hex.EncodeToString(signature))

	mac2 := hmac.New(sha256.New, key)
	mac2.Write(message)
	expectedMAC := mac2.Sum(nil)
	fmt.Println("  Verificacion valida:", hmac.Equal(signature, expectedMAC))

	tampered := []byte("Transferir $9999 a cuenta 12345")
	mac3 := hmac.New(sha256.New, key)
	mac3.Write(tampered)
	tamperedMAC := mac3.Sum(nil)
	fmt.Println("  Verificacion mensaje alterado:", hmac.Equal(signature, tamperedMAC))

	// ============================================
	// 4. AES ENCRYPTION - SYMMETRIC CIPHER
	// ============================================
	fmt.Println("\n--- 4. AES Encryption - Symmetric Cipher ---")
	os.Stdout.WriteString(`
AES (Advanced Encryption Standard):
  - Cifrado simetrico de bloque (128 bits = 16 bytes)
  - Misma clave para cifrar y descifrar
  - Tamanos de clave: 128, 192 o 256 bits

MODOS DE OPERACION:

  GCM (Galois/Counter Mode) - RECOMENDADO:
  - Authenticated encryption (AEAD)
  - Proporciona confidencialidad + integridad + autenticidad
  - Detecta modificaciones del ciphertext
  - Requiere nonce unico de 12 bytes por mensaje
  - Estandar de la industria (TLS 1.3, IPsec)

  CBC (Cipher Block Chaining) - Legacy:
  - Solo confidencialidad (no autenticidad)
  - Requiere IV aleatorio + padding (PKCS7)
  - Vulnerable a padding oracle attacks si no se usa con HMAC
  - Usar solo para compatibilidad legacy

  CTR (Counter Mode):
  - Convierte block cipher en stream cipher
  - Solo confidencialidad, necesita MAC separado
  - Paralelizable, no necesita padding

REGLAS DE ORO:
  1. SIEMPRE usar GCM para cifrado nuevo
  2. NUNCA reusar un nonce con la misma clave
  3. La clave debe ser generada con crypto/rand
  4. Guardar nonce + ciphertext juntos (nonce es publico)
`)

	fmt.Println("  Demo: AES-256-GCM Encrypt/Decrypt")

	aesKey := make([]byte, 32)
	rand.Read(aesKey)

	plaintext := []byte("Datos secretos que quiero proteger")

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		fmt.Println("  Error creando cipher:", err)
		return
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("  Error creando GCM:", err)
		return
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("  Error generando nonce:", err)
		return
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	fmt.Println("  Plaintext: ", string(plaintext))
	fmt.Println("  Ciphertext (hex):", hex.EncodeToString(ciphertext[:40]), "...")
	fmt.Println("  Ciphertext length:", len(ciphertext), "bytes (nonce + data + tag)")

	extractedNonce := ciphertext[:aesGCM.NonceSize()]
	actualCiphertext := ciphertext[aesGCM.NonceSize():]
	decrypted, err := aesGCM.Open(nil, extractedNonce, actualCiphertext, nil)
	if err != nil {
		fmt.Println("  Error descifrando:", err)
		return
	}
	fmt.Println("  Decrypted: ", string(decrypted))
	fmt.Println("  Match:", string(plaintext) == string(decrypted))

	// ============================================
	// 5. RSA - ASYMMETRIC ENCRYPTION
	// ============================================
	fmt.Println("\n--- 5. RSA - Asymmetric Encryption ---")
	os.Stdout.WriteString(`
RSA (Rivest-Shamir-Adleman):
  - Par de claves: publica (cifrar/verificar) + privada (descifrar/firmar)
  - Tamanos recomendados: 2048 bits (minimo), 3072 o 4096 bits
  - Mas lento que AES, usado para intercambio de claves y firmas

OPERACIONES:

  rsa.GenerateKey(random, bits) (*rsa.PrivateKey, error)

  CIFRADO (OAEP - recomendado):
    rsa.EncryptOAEP(hash, random, pub, msg, label) ([]byte, error)
    rsa.DecryptOAEP(hash, random, priv, ciphertext, label) ([]byte, error)

  FIRMA (PSS - recomendado):
    rsa.SignPSS(random, priv, hash, hashed, opts) ([]byte, error)
    rsa.VerifyPSS(pub, hash, hashed, sig, opts) error

  LEGACY (evitar en codigo nuevo):
    rsa.EncryptPKCS1v15  -> Vulnerable a Bleichenbacher attack
    rsa.SignPKCS1v15      -> Menos seguro que PSS

LIMITACION:
  - Solo puede cifrar datos menores al tamano de clave menos padding
  - 2048-bit RSA con OAEP-SHA256: max ~190 bytes de plaintext
  - Para datos grandes: cifrar con AES, luego cifrar la clave AES con RSA
`)

	fmt.Println("  Demo: RSA Key Generation, Encrypt/Decrypt, Sign/Verify")

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("  Error generando clave RSA:", err)
		return
	}
	rsaPub := &rsaKey.PublicKey

	fmt.Println("  RSA key generated: 2048 bits")
	fmt.Println("  Public key exponent:", rsaPub.E)

	rsaPlaintext := []byte("Mensaje secreto con RSA")
	rsaCiphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, rsaPlaintext, nil)
	if err != nil {
		fmt.Println("  Error cifrando:", err)
		return
	}
	fmt.Println("  RSA Encrypted (hex):", hex.EncodeToString(rsaCiphertext[:30]), "...")

	rsaDecrypted, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, rsaKey, rsaCiphertext, nil)
	if err != nil {
		fmt.Println("  Error descifrando:", err)
		return
	}
	fmt.Println("  RSA Decrypted:", string(rsaDecrypted))

	msgHash := sha256.Sum256([]byte("Documento importante para firmar"))
	rsaSig, err := rsa.SignPSS(rand.Reader, rsaKey, crypto.SHA256, msgHash[:], nil)
	if err != nil {
		fmt.Println("  Error firmando:", err)
		return
	}
	fmt.Println("  RSA-PSS Signature (hex):", hex.EncodeToString(rsaSig[:30]), "...")

	err = rsa.VerifyPSS(rsaPub, crypto.SHA256, msgHash[:], rsaSig, nil)
	fmt.Println("  RSA-PSS Signature valid:", err == nil)

	// ============================================
	// 6. ECDSA - ELLIPTIC CURVE DIGITAL SIGNATURES
	// ============================================
	fmt.Println("\n--- 6. ECDSA - Elliptic Curve Digital Signatures ---")
	os.Stdout.WriteString(`
ECDSA (Elliptic Curve Digital Signature Algorithm):
  - Firmas digitales basadas en curvas elipticas
  - Claves mas pequenas que RSA para seguridad equivalente
  - 256-bit ECDSA ~ 3072-bit RSA en seguridad

CURVAS DISPONIBLES:

  elliptic.P224()  -> 224 bits (NIST P-224)
  elliptic.P256()  -> 256 bits (NIST P-256) - Recomendada, soporte hardware
  elliptic.P384()  -> 384 bits (NIST P-384)
  elliptic.P521()  -> 521 bits (NIST P-521)

FUNCIONES:

  ecdsa.GenerateKey(curve, random) (*ecdsa.PrivateKey, error)
  ecdsa.SignASN1(random, priv, hash) ([]byte, error)    // Go 1.15+
  ecdsa.VerifyASN1(pub, hash, sig) bool                 // Go 1.15+

USOS:
  - TLS certificates (la mayoria de sitios web)
  - Bitcoin y Ethereum (secp256k1, no P-256)
  - JWT con algoritmo ES256
  - Code signing
`)

	fmt.Println("  Demo: ECDSA Sign/Verify con P-256")

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("  Error generando clave ECDSA:", err)
		return
	}

	ecMsg := sha256.Sum256([]byte("Transaccion: enviar 1 BTC"))
	ecSig, err := ecdsa.SignASN1(rand.Reader, ecKey, ecMsg[:])
	if err != nil {
		fmt.Println("  Error firmando:", err)
		return
	}

	fmt.Println("  ECDSA P-256 key generated")
	fmt.Println("  Signature (hex):", hex.EncodeToString(ecSig[:30]), "...")
	fmt.Println("  Signature length:", len(ecSig), "bytes (vs RSA 256 bytes)")

	valid := ecdsa.VerifyASN1(&ecKey.PublicKey, ecMsg[:], ecSig)
	fmt.Println("  Signature valid:", valid)

	wrongMsg := sha256.Sum256([]byte("Transaccion: enviar 1000 BTC"))
	wrongValid := ecdsa.VerifyASN1(&ecKey.PublicKey, wrongMsg[:], ecSig)
	fmt.Println("  Wrong message verification:", wrongValid)

	// ============================================
	// 7. ED25519 - MODERN SIGNING
	// ============================================
	fmt.Println("\n--- 7. Ed25519 - Modern Signing ---")
	os.Stdout.WriteString(`
Ed25519 (Edwards-curve Digital Signature Algorithm):
  - Curva eliptica Curve25519 en forma Edwards
  - Firma mas rapida y simple que ECDSA
  - Clave fija: 32 bytes (publica), 64 bytes (privada)
  - Firma: 64 bytes (siempre)
  - Determinista: no necesita fuente de aleatoriedad para firmar

VENTAJAS SOBRE ECDSA:
  - No necesita random para firmar (immune a bad RNG)
  - Mas rapido en la mayoria de plataformas
  - Resistente a side-channel attacks por diseno
  - API mas simple

FUNCIONES:

  ed25519.GenerateKey(rand) (pub, priv, error)
  ed25519.Sign(priv, message) []byte
  ed25519.Verify(pub, message, sig) bool

USOS:
  - SSH keys (ssh-ed25519) - recomendado por OpenSSH
  - WireGuard VPN
  - Tor, Signal Protocol
  - DNSSEC, minisign
`)

	fmt.Println("  Demo: Ed25519 Sign/Verify")

	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Println("  Error generando clave Ed25519:", err)
		return
	}

	edMessage := []byte("Mensaje firmado con Ed25519")
	edSignature := ed25519.Sign(edPriv, edMessage)

	fmt.Println("  Ed25519 public key (hex):", hex.EncodeToString(edPub))
	fmt.Println("  Signature (hex):", hex.EncodeToString(edSignature))
	fmt.Println("  Signature length:", len(edSignature), "bytes (siempre 64)")

	edValid := ed25519.Verify(edPub, edMessage, edSignature)
	fmt.Println("  Signature valid:", edValid)

	edTampered := []byte("Mensaje ALTERADO con Ed25519")
	edTamperedValid := ed25519.Verify(edPub, edTampered, edSignature)
	fmt.Println("  Tampered message valid:", edTamperedValid)

	// ============================================
	// 8. X509 CERTIFICATES
	// ============================================
	fmt.Println("\n--- 8. x509 Certificates ---")
	os.Stdout.WriteString(`
CERTIFICADOS X.509:

Un certificado X.509 vincula una clave publica con una identidad.
Es la base de TLS/HTTPS, code signing y autenticacion mutua.

ESTRUCTURA:
  - Subject: identidad del propietario (CN, O, OU, etc.)
  - Issuer: quien firmo el certificado (CA)
  - Validity: Not Before / Not After
  - Public Key: clave publica del subject
  - Extensions: SAN, Key Usage, Basic Constraints
  - Signature: firma del issuer sobre todo lo anterior

TIPOS:
  - Root CA: auto-firmado, raiz de confianza
  - Intermediate CA: firmado por root, firma certificados finales
  - Leaf/End-entity: certificado de servidor o cliente

FORMATOS:
  - PEM: Base64 con headers (-----BEGIN CERTIFICATE-----)
  - DER: Binario (ASN.1 DER encoding)

FUNCIONES PRINCIPALES:

  x509.CreateCertificate(rand, template, parent, pub, priv) ([]byte, error)
  x509.ParseCertificate(asn1Data) (*x509.Certificate, error)
  x509.ParseCertificatePEM(pemData) (*x509.Certificate, error)

  // Pools de certificados confiables
  x509.NewCertPool()
  pool.AppendCertsFromPEM(pemData) bool
  x509.SystemCertPool() (*x509.CertPool, error)

VERIFICACION:

  cert.Verify(x509.VerifyOptions{
      Roots:         rootPool,
      Intermediates: intermediatePool,
      DNSNames:      []string{"example.com"},
  })
`)

	fmt.Println("  Demo: Crear certificado auto-firmado (self-signed)")

	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("  Error generando clave:", err)
		return
	}

	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Go Crypto Course"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "*.local"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &certKey.PublicKey, certKey)
	if err != nil {
		fmt.Println("  Error creando certificado:", err)
		return
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	fmt.Println("  Certificado PEM creado:")
	pemStr := string(certPEM)
	lines := strings.Split(pemStr, "\n")
	for i, line := range lines {
		if i < 3 || i >= len(lines)-3 {
			fmt.Println("   ", line)
		} else if i == 3 {
			fmt.Println("    ... (contenido base64) ...")
		}
	}

	parsedCert, err := x509.ParseCertificate(certDER)
	if err != nil {
		fmt.Println("  Error parsing certificado:", err)
		return
	}
	fmt.Println("  Subject:", parsedCert.Subject.CommonName)
	fmt.Println("  Issuer:", parsedCert.Issuer.CommonName)
	fmt.Println("  Not Before:", parsedCert.NotBefore.Format(time.RFC3339))
	fmt.Println("  Not After:", parsedCert.NotAfter.Format(time.RFC3339))
	fmt.Println("  DNS Names:", parsedCert.DNSNames)
	fmt.Println("  Is CA:", parsedCert.IsCA)

	// ============================================
	// 9. TLS CONFIGURATION
	// ============================================
	fmt.Println("\n--- 9. TLS Configuration ---")
	os.Stdout.WriteString(`
TLS (Transport Layer Security):

crypto/tls provee implementacion completa de TLS 1.2 y 1.3.

CONFIGURACION BASICA SERVIDOR:

  cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
  tlsConfig := &tls.Config{
      Certificates: []tls.Certificate{cert},
      MinVersion:   tls.VersionTLS12,
  }
  listener, err := tls.Listen("tcp", ":443", tlsConfig)

CONFIGURACION SEGURA RECOMENDADA:

  tlsConfig := &tls.Config{
      MinVersion:               tls.VersionTLS12,
      PreferServerCipherSuites: true,  // Deprecated en Go 1.17, ignorado
      CurvePreferences: []tls.CurveID{
          tls.X25519,    // Mas rapido
          tls.CurveP256, // Amplio soporte
      },
      CipherSuites: []uint16{
          // TLS 1.3 ciphers (siempre habilitados, no configurables)
          // TLS 1.2 ciphers recomendados:
          tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
          tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
          tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
          tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
      },
  }

mTLS (Mutual TLS):

El servidor TAMBIEN verifica el certificado del cliente.

  // Servidor con mTLS
  clientCAs := x509.NewCertPool()
  clientCAs.AppendCertsFromPEM(clientCACert)

  tlsConfig := &tls.Config{
      Certificates: []tls.Certificate{serverCert},
      ClientAuth:   tls.RequireAndVerifyClientCert,
      ClientCAs:    clientCAs,
      MinVersion:   tls.VersionTLS12,
  }

  // Cliente con certificado
  clientCert, _ := tls.LoadX509KeyPair("client.crt", "client.key")

  tlsConfig := &tls.Config{
      Certificates: []tls.Certificate{clientCert},
      RootCAs:      serverCAs,
  }

TLS 1.3 MEJORAS:
  - 1-RTT handshake (vs 2-RTT en TLS 1.2)
  - 0-RTT resumption (early data)
  - Solo cipher suites AEAD
  - Encrypted SNI (en desarrollo)
  - Diffie-Hellman efimero obligatorio
  - Remocion de algoritmos legacy (RSA key exchange, CBC, etc.)
`)

	// ============================================
	// 10. KEY DERIVATION FUNCTIONS
	// ============================================
	fmt.Println("\n--- 10. Key Derivation Functions ---")
	os.Stdout.WriteString(`
KDF = Key Derivation Function

Convierten passwords (baja entropia) en claves criptograficas (alta entropia).

ALGORITMOS (de mas nuevo a mas viejo):

  Argon2 (2015) - RECOMENDADO para passwords:
    - Ganador del Password Hashing Competition
    - Argon2id: resistente a side-channel y GPU attacks
    - Configurable: memoria, iteraciones, paralelismo
    - golang.org/x/crypto/argon2

  scrypt (2009) - Alternativa solida:
    - Memory-hard (costoso en GPU/ASIC)
    - Usado por Litecoin, Tarsnap
    - golang.org/x/crypto/scrypt

  PBKDF2 (2000) - Compatibilidad:
    - NIST aprobado
    - Simple, amplio soporte
    - NO memory-hard (vulnerable a GPU attacks)
    - crypto/pbkdf2 (Go 1.24+) o golang.org/x/crypto/pbkdf2

  bcrypt (1999) - Password hashing:
    - Ampliamente usado, probado por decadas
    - Cost factor ajustable
    - golang.org/x/crypto/bcrypt

EJEMPLO ARGON2:

  import "golang.org/x/crypto/argon2"

  salt := make([]byte, 16)
  rand.Read(salt)

  // Argon2id: time=1, memory=64MB, threads=4, keyLen=32
  key := argon2.IDKey([]byte("password"), salt, 1, 64*1024, 4, 32)

EJEMPLO BCRYPT:

  import "golang.org/x/crypto/bcrypt"

  // Hash
  hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

  // Verify
  err = bcrypt.CompareHashAndPassword(hash, []byte("password"))

PARAMETROS RECOMENDADOS (2024+):

  Argon2id:  time=3, memory=64MB, threads=4
  scrypt:    N=32768, r=8, p=1
  bcrypt:    cost=12
  PBKDF2:    iterations=600000 (con SHA-256)
`)

	// ============================================
	// 11. DIGITAL SIGNATURES WORKFLOW
	// ============================================
	fmt.Println("\n--- 11. Digital Signatures Workflow ---")
	os.Stdout.WriteString(`
FLUJO COMPLETO DE FIRMA DIGITAL:

1. GENERACION DE CLAVES:
   - Generar par (publica, privada) con crypto/rand
   - Almacenar privada de forma segura (HSM, vault, KMS)
   - Distribuir publica (certificados, JWKS, etc.)

2. FIRMA:
   sender:
     hash = SHA256(message)
     signature = Sign(privateKey, hash)
     send(message, signature)

3. VERIFICACION:
   receiver:
     hash = SHA256(message)
     valid = Verify(publicKey, hash, signature)
     if valid -> mensaje autentico e integro

ALGORITMOS RECOMENDADOS POR CASO:

  Web/TLS:        ECDSA P-256 (amplio soporte) o Ed25519
  SSH:             Ed25519
  JWT:             ES256 (ECDSA P-256) o EdDSA
  Code signing:    Ed25519 o RSA-4096
  Legacy systems:  RSA-2048 con PSS padding
  Blockchain:      ECDSA secp256k1 (Bitcoin/Ethereum)

COMPARATIVA:

  Algoritmo    Clave Publica  Firma     Velocidad
  Ed25519      32 bytes       64 bytes  Rapido
  ECDSA P-256  64 bytes       ~72 bytes Medio
  RSA-2048     256 bytes      256 bytes Lento
  RSA-4096     512 bytes      512 bytes Muy lento
`)

	fmt.Println("  Demo: Flujo completo de firma con Ed25519")

	senderPub, senderPriv, _ := ed25519.GenerateKey(rand.Reader)
	document := []byte("Contrato: Yo acepto los terminos y condiciones...")

	docSignature := ed25519.Sign(senderPriv, document)
	fmt.Println("  Documento firmado, signature:", hex.EncodeToString(docSignature[:20]), "...")

	isValid := ed25519.Verify(senderPub, document, docSignature)
	fmt.Println("  Verificacion por receptor:", isValid)

	alteredDoc := []byte("Contrato: Yo NO acepto los terminos y condiciones...")
	isValidAltered := ed25519.Verify(senderPub, alteredDoc, docSignature)
	fmt.Println("  Documento alterado valido:", isValidAltered)

	// ============================================
	// 12. CERTIFICATE CHAIN VALIDATION
	// ============================================
	fmt.Println("\n--- 12. Certificate Chain Validation ---")
	os.Stdout.WriteString(`
CADENA DE CONFIANZA:

  Root CA (auto-firmado, en trust store del OS)
    |
    +-> Intermediate CA (firmado por Root)
          |
          +-> Leaf Certificate (firmado por Intermediate)
                |
                +-> Tu servidor web

VERIFICACION:

  1. El cliente recibe leaf + intermediates del servidor
  2. El cliente tiene root CAs en su trust store
  3. Verifica firma de cada certificado en la cadena
  4. Verifica que el leaf matchea el hostname (SAN)
  5. Verifica que ninguno este expirado o revocado

EN GO:

  // Verificar un certificado
  roots := x509.NewCertPool()
  roots.AppendCertsFromPEM(rootCACert)

  intermediates := x509.NewCertPool()
  intermediates.AppendCertsFromPEM(intermediateCert)

  opts := x509.VerifyOptions{
      DNSNames:      []string{"example.com"},
      Roots:         roots,
      Intermediates: intermediates,
      CurrentTime:   time.Now(),
  }

  chains, err := leafCert.Verify(opts)
  if err != nil {
      // Certificado invalido
  }

  // Usando system roots
  systemRoots, err := x509.SystemCertPool()

ERRORES COMUNES:
  - x509: certificate signed by unknown authority -> falta CA en pool
  - x509: certificate has expired or is not yet valid -> verificar fechas
  - x509: certificate is valid for X, not Y -> hostname mismatch
  - x509: certificate relies on legacy Common Name field -> usar SAN
`)

	fmt.Println("  Demo: Crear y verificar cadena de certificados")

	rootKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	rootSerial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	rootTemplate := x509.Certificate{
		SerialNumber:          rootSerial,
		Subject:               pkix.Name{Organization: []string{"Go Course CA"}, CommonName: "Root CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}
	rootCertDER, _ := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	rootCert, _ := x509.ParseCertificate(rootCertDER)

	leafKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	leafSerial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	leafTemplate := x509.Certificate{
		SerialNumber: leafSerial,
		Subject:      pkix.Name{Organization: []string{"My Server"}, CommonName: "myserver.local"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"myserver.local", "localhost"},
	}
	leafCertDER, _ := x509.CreateCertificate(rand.Reader, &leafTemplate, rootCert, &leafKey.PublicKey, rootKey)
	leafCert, _ := x509.ParseCertificate(leafCertDER)

	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCert)

	chains, err := leafCert.Verify(x509.VerifyOptions{
		Roots: rootPool,
	})
	if err != nil {
		fmt.Println("  Verificacion fallida:", err)
	} else {
		fmt.Println("  Certificado verificado exitosamente!")
		fmt.Println("  Cadena de confianza:", len(chains), "cadena(s) encontrada(s)")
		for i, chain := range chains {
			fmt.Println("  Cadena", i+1, ":")
			for j, cert := range chain {
				fmt.Println("    ", j, "->", cert.Subject.CommonName, "(issuer:", cert.Issuer.CommonName+")")
			}
		}
	}

	wrongPool := x509.NewCertPool()
	_, err = leafCert.Verify(x509.VerifyOptions{
		Roots: wrongPool,
	})
	fmt.Println("  Verificacion con hostname incorrecto:", err != nil, "(esperado)")

	// ============================================
	// 13. SECURE COMPARISON
	// ============================================
	fmt.Println("\n--- 13. Secure Comparison (crypto/subtle) ---")
	os.Stdout.WriteString(`
TIMING ATTACKS:

Comparaciones normales (==, bytes.Equal) terminan al primer byte diferente.
Un atacante puede medir el tiempo de respuesta para descubrir bytes correctos.

  secret := "correct-token"
  input  := "correct-tXXXX"  // Tarda mas que "XXXXXXXXXXXXX"

SOLUCION: crypto/subtle

  subtle.ConstantTimeCompare(x, y []byte) int
  - Retorna 1 si iguales, 0 si diferentes
  - SIEMPRE toma el mismo tiempo (independiente de donde difieren)
  - AMBOS slices deben tener el mismo largo

  subtle.ConstantTimeEq(x, y int32) int
  - Compara enteros en tiempo constante

  subtle.ConstantTimeSelect(v, x, y int) int
  - Si v==1 retorna x, si v==0 retorna y (sin branching)

  subtle.ConstantTimeCopy(v int, x, y []byte)
  - Si v==1 copia y a x, si v==0 no hace nada

DONDE USAR:
  - Verificacion de tokens/API keys
  - Comparacion de MACs/HMACs
  - Cualquier secreto comparado con input del usuario

IMPORTANTE:
  hmac.Equal() ya usa ConstantTimeCompare internamente.
  Para passwords: usar bcrypt.CompareHashAndPassword (incluye timing safety).
`)

	fmt.Println("  Demo: crypto/subtle.ConstantTimeCompare")

	secret := []byte("mi-api-key-secreta-12345")
	correctInput := []byte("mi-api-key-secreta-12345")
	wrongInput := []byte("mi-api-key-secreta-99999")
	shortInput := []byte("short")

	fmt.Println("  Correct input match:", subtle.ConstantTimeCompare(secret, correctInput) == 1)
	fmt.Println("  Wrong input match:  ", subtle.ConstantTimeCompare(secret, wrongInput) == 1)
	fmt.Println("  Short input match:  ", subtle.ConstantTimeCompare(secret, shortInput) == 1)

	fmt.Println("\n  ConstantTimeEq:")
	fmt.Println("  42 == 42:", subtle.ConstantTimeEq(42, 42) == 1)
	fmt.Println("  42 == 99:", subtle.ConstantTimeEq(42, 99) == 1)

	fmt.Println("\n  ConstantTimeSelect:")
	fmt.Println("  Select(1, 10, 20):", subtle.ConstantTimeSelect(1, 10, 20))
	fmt.Println("  Select(0, 10, 20):", subtle.ConstantTimeSelect(0, 10, 20))

	// ============================================
	// 14. FIPS 140-3 COMPLIANCE
	// ============================================
	fmt.Println("\n--- 14. FIPS 140-3 Compliance ---")
	os.Stdout.WriteString(`
FIPS 140-3 = Federal Information Processing Standard

Estandar del gobierno de EE.UU. para modulos criptograficos.
Requerido por agencias federales y muchas empresas reguladas.

GO Y FIPS:

Go 1.24+ introdujo soporte nativo para FIPS 140-3:

  GOEXPERIMENT=boringcrypto (Go 1.19+):
  - Reemplaza crypto/* con BoringSSL (certificado FIPS)
  - Transparente para el codigo Go existente
  - Solo algoritmos FIPS aprobados disponibles

  Go 1.24 FIPS 140-3:
  - crypto/tls/fipsonly: Restringe TLS a algoritmos FIPS
  - Nuevo build tag: goexperiment.boringcrypto
  - Mecanismo de compliance mas robusto

ALGORITMOS APROBADOS POR FIPS:

  Cifrado:     AES (128, 192, 256) en GCM, CBC, CTR
  Hashing:     SHA-2 (SHA-256, SHA-384, SHA-512), SHA-3
  MAC:         HMAC con SHA-2
  Firmas:      RSA (2048+), ECDSA (P-256, P-384, P-521)
  KDF:         PBKDF2, HKDF, KBKDF
  RNG:         DRBG basado en AES-CTR o HMAC

NO APROBADOS POR FIPS:
  - MD5, SHA-1 (solo para usos no criptograficos)
  - DES, 3DES (retirado en 2023)
  - RSA < 2048 bits
  - Ed25519 (aun no aprobado, en proceso)
  - ChaCha20-Poly1305 (no FIPS, pero si seguro)

HABILITANDO FIPS EN GO:

  // Build con BoringCrypto
  GOEXPERIMENT=boringcrypto go build .

  // Verificar que BoringCrypto esta activo
  import "crypto/boring"
  fmt.Println(boring.Enabled())  // true si FIPS activo

  // Restringir TLS a FIPS
  import _ "crypto/tls/fipsonly"
  // Ahora solo cipher suites FIPS estan disponibles

VERIFICANDO COMPLIANCE:

  // En runtime
  import "crypto/boring"

  if !boring.Enabled() {
      log.Fatal("FIPS mode not enabled")
  }

  // En TLS config
  tlsConfig := &tls.Config{
      MinVersion: tls.VersionTLS12,
      // Con fipsonly, automaticamente solo usa cipher suites FIPS
  }
`)

	// ============================================
	// 15. PRACTICAL PATTERNS
	// ============================================
	fmt.Println("\n--- 15. Practical Crypto Patterns ---")
	os.Stdout.WriteString(`
PATRON: ENVELOPE ENCRYPTION (datos grandes)

  1. Generar DEK (Data Encryption Key) aleatorio con crypto/rand
  2. Cifrar datos con DEK usando AES-GCM
  3. Cifrar DEK con KEK (Key Encryption Key) usando RSA o KMS
  4. Almacenar: encrypted_DEK + encrypted_data

  // Ventajas:
  // - Datos cifrados con AES (rapido)
  // - Solo DEK cifrado con RSA (lento pero pequeno)
  // - Rotar KEK sin re-cifrar datos
  // - Cada registro tiene su propia DEK

PATRON: HASH + SALT PARA PASSWORDS

  // NUNCA almacenar passwords en texto plano
  // NUNCA usar hash sin salt (vulnerable a rainbow tables)
  // NUNCA usar MD5/SHA para passwords (muy rapido = inseguro)

  // CORRECTO: usar bcrypt, argon2 o scrypt
  // Estos incluyen salt automaticamente

PATRON: NONCE MANAGEMENT

  // GCM nonce de 12 bytes = max 2^32 mensajes por clave
  // Opciones:
  // 1. Counter: simple pero requiere estado persistente
  // 2. Random: facil, pero limite de ~2^32 mensajes por clave
  // 3. XChaCha20-Poly1305: nonce de 24 bytes, limite 2^64

  // REGLA: Rotar clave antes de 2^32 cifrados con GCM

PATRON: CONSTANT-TIME TOKEN VALIDATION

  func ValidateToken(provided, stored string) bool {
      // Convertir a []byte para usar ConstantTimeCompare
      // Hashear primero si largos pueden variar
      providedHash := sha256.Sum256([]byte(provided))
      storedHash := sha256.Sum256([]byte(stored))
      return subtle.ConstantTimeCompare(providedHash[:], storedHash[:]) == 1
  }

PATRON: HYBRID ENCRYPTION

  // Combinar asimetrico + simetrico:
  // 1. Generar clave AES aleatoria (32 bytes)
  // 2. Cifrar datos con AES-GCM (rapido)
  // 3. Cifrar clave AES con RSA-OAEP del receptor
  // 4. Enviar: RSA_encrypted_key + AES_encrypted_data

  type HybridMessage struct {
      EncryptedKey  []byte // RSA-OAEP encrypted AES key
      Nonce         []byte // GCM nonce
      Ciphertext    []byte // AES-GCM encrypted data
  }
`)

	fmt.Println("  Demo: Hybrid Encryption (RSA + AES-GCM)")

	recipientKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	recipientPub := &recipientKey.PublicKey

	sessionKey := make([]byte, 32)
	rand.Read(sessionKey)

	largeData := []byte("Datos grandes que necesitan cifrado eficiente... " + strings.Repeat("Lorem ipsum dolor sit amet. ", 10))

	encryptedSessionKey, _ := rsa.EncryptOAEP(sha256.New(), rand.Reader, recipientPub, sessionKey, nil)

	hybridBlock, _ := aes.NewCipher(sessionKey)
	hybridGCM, _ := cipher.NewGCM(hybridBlock)
	hybridNonce := make([]byte, hybridGCM.NonceSize())
	io.ReadFull(rand.Reader, hybridNonce)
	encryptedData := hybridGCM.Seal(nil, hybridNonce, largeData, nil)

	fmt.Println("  Original data size:", len(largeData), "bytes")
	fmt.Println("  Encrypted session key:", len(encryptedSessionKey), "bytes")
	fmt.Println("  Encrypted data:", len(encryptedData), "bytes")
	fmt.Println("  Nonce:", len(hybridNonce), "bytes")

	decSessionKey, _ := rsa.DecryptOAEP(sha256.New(), rand.Reader, recipientKey, encryptedSessionKey, nil)
	decBlock, _ := aes.NewCipher(decSessionKey)
	decGCM, _ := cipher.NewGCM(decBlock)
	decData, err := decGCM.Open(nil, hybridNonce, encryptedData, nil)
	if err != nil {
		fmt.Println("  Error decrypting:", err)
	} else {
		fmt.Println("  Decrypted matches original:", string(largeData) == string(decData))
	}

	// ============================================
	// 16. PEM ENCODING/DECODING
	// ============================================
	fmt.Println("\n--- 16. PEM Encoding/Decoding ---")
	os.Stdout.WriteString(`
PEM (Privacy Enhanced Mail):

Formato estandar para almacenar claves y certificados como texto.

  -----BEGIN <TYPE>-----
  <base64 encoded DER data>
  -----END <TYPE>-----

TIPOS COMUNES:

  "CERTIFICATE"          -> x509 certificado
  "PRIVATE KEY"          -> PKCS#8 private key
  "RSA PRIVATE KEY"      -> PKCS#1 RSA private key
  "EC PRIVATE KEY"       -> SEC 1 ECDSA private key
  "PUBLIC KEY"           -> PKIX public key
  "CERTIFICATE REQUEST"  -> CSR (Certificate Signing Request)

FUNCIONES:

  // Encode
  pem.Encode(out io.Writer, block *pem.Block) error
  pem.EncodeToMemory(block *pem.Block) []byte

  // Decode
  pem.Decode(data []byte) (block *pem.Block, rest []byte)

  // Marshaling claves
  x509.MarshalPKCS8PrivateKey(key any) ([]byte, error)   // Recomendado
  x509.MarshalECPrivateKey(key *ecdsa.PrivateKey) ([]byte, error)
  x509.MarshalPKIXPublicKey(pub any) ([]byte, error)

  // Parsing claves
  x509.ParsePKCS8PrivateKey(der []byte) (any, error)
  x509.ParsePKIXPublicKey(der []byte) (any, error)
`)

	fmt.Println("  Demo: Serializar/Deserializar claves Ed25519 como PEM")

	pemPub, pemPriv, _ := ed25519.GenerateKey(rand.Reader)

	privDER, _ := x509.MarshalPKCS8PrivateKey(pemPriv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})

	pubDER, _ := x509.MarshalPKIXPublicKey(pemPub)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})

	fmt.Println("  Private Key PEM:")
	for _, line := range strings.Split(string(privPEM), "\n") {
		if line != "" {
			if strings.HasPrefix(line, "-----") {
				fmt.Println("   ", line)
			} else {
				fmt.Println("    (base64 data)")
				break
			}
		}
	}
	fmt.Println("    -----END PRIVATE KEY-----")

	fmt.Println("  Public Key PEM:")
	for _, line := range strings.Split(string(pubPEM), "\n") {
		if line != "" {
			if strings.HasPrefix(line, "-----") {
				fmt.Println("   ", line)
			} else {
				fmt.Println("    (base64 data)")
				break
			}
		}
	}
	fmt.Println("    -----END PUBLIC KEY-----")

	pemBlock, _ := pem.Decode(privPEM)
	parsedPriv, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		fmt.Println("  Error parsing:", err)
	} else {
		edParsedPriv := parsedPriv.(ed25519.PrivateKey)
		testSig := ed25519.Sign(edParsedPriv, []byte("test"))
		fmt.Println("  Parsed key works:", ed25519.Verify(pemPub, []byte("test"), testSig))
	}

	// ============================================
	// 17. CRYPTO BEST PRACTICES
	// ============================================
	fmt.Println("\n--- 17. Crypto Best Practices ---")
	os.Stdout.WriteString(`
REGLAS DE ORO:

1. NO implementes tu propia criptografia
   - Usa las librerias estandar de Go (crypto/*)
   - Usa golang.org/x/crypto para lo que no esta en stdlib

2. Usa algoritmos modernos:
   - Cifrado simetrico: AES-256-GCM o ChaCha20-Poly1305
   - Hashing: SHA-256 o SHA-512
   - Passwords: Argon2id o bcrypt
   - Firmas: Ed25519 o ECDSA P-256
   - Key exchange: X25519 (ECDH)
   - KDF: HKDF (para derivar subclaves)

3. Genera claves con crypto/rand, NUNCA con math/rand

4. NUNCA reutilices nonces/IVs con la misma clave

5. Usa AEAD (AES-GCM, ChaCha20-Poly1305) para cifrado
   - Cifrar sin autenticar es peligroso

6. Compara secretos con crypto/subtle.ConstantTimeCompare

7. Rota claves periodicamente

8. Almacena claves privadas de forma segura:
   - Variables de entorno (minimo)
   - Secret managers (Vault, AWS KMS, GCP KMS)
   - HSMs para produccion critica

9. Cifra datos sensibles at-rest y in-transit

10. Valida certificados correctamente:
    - No deshabilites la verificacion TLS en produccion
    - Verifica hostname, expiracion y cadena de confianza

ERRORES COMUNES:

  // MAL: ECB mode (no usar NUNCA)
  block, _ := aes.NewCipher(key)
  block.Encrypt(dst, src)  // ECB = cada bloque independiente = inseguro

  // MAL: Nonce hardcoded
  nonce := []byte("0000000000000")  // NUNCA hacer esto

  // MAL: MD5/SHA1 para seguridad
  hash := md5.Sum(password)  // NO para passwords

  // MAL: Desactivar TLS verification
  &tls.Config{InsecureSkipVerify: true}  // SOLO para testing

  // MAL: RSA < 2048 bits
  rsa.GenerateKey(rand.Reader, 1024)  // Insuficiente

  // BIEN: Patron seguro
  key := make([]byte, 32)
  rand.Read(key)
  block, _ := aes.NewCipher(key)
  gcm, _ := cipher.NewGCM(block)
  nonce := make([]byte, gcm.NonceSize())
  io.ReadFull(rand.Reader, nonce)
  ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
`)

	fmt.Println("\n  RESUMEN DE ALGORITMOS RECOMENDADOS:")
	fmt.Println("  +------------------+-----------------------------+")
	fmt.Println("  | Uso              | Algoritmo                   |")
	fmt.Println("  +------------------+-----------------------------+")
	fmt.Println("  | Cifrado datos    | AES-256-GCM                 |")
	fmt.Println("  | Passwords        | Argon2id / bcrypt           |")
	fmt.Println("  | Firmas           | Ed25519 / ECDSA P-256       |")
	fmt.Println("  | Hashing          | SHA-256 / SHA-512           |")
	fmt.Println("  | HMAC             | HMAC-SHA256                 |")
	fmt.Println("  | Key exchange     | X25519 (ECDH)               |")
	fmt.Println("  | TLS              | TLS 1.3                     |")
	fmt.Println("  | Random           | crypto/rand                 |")
	fmt.Println("  | Key derivation   | HKDF / Argon2id             |")
	fmt.Println("  | Token comparison | crypto/subtle               |")
	fmt.Println("  +------------------+-----------------------------+")
}

// Suppress unused import warnings
var _ = io.ReadFull
var _ = strings.Repeat
var _ = time.Now
var _ = big.NewInt
var _ = elliptic.P256
var _ = pem.Encode
var _ = pkix.Name{}
var _ = md5.Sum
var _ = sha512.Sum512
var _ = hex.EncodeToString
var _ = os.Stdout

/*
RESUMEN CHAPTER 74 - CRYPTO PACKAGE DEEP DIVE:

crypto/rand:
- CSPRNG del sistema operativo, usar siempre para claves/nonces/tokens
- rand.Read, rand.Int, rand.Prime, rand.Reader

HASHING (SHA-256, SHA-512, MD5):
- Funciones unidireccionales de digest fijo
- SHA-256 para uso general, MD5/SHA-1 solo checksums
- Dos APIs: Sum256 (simple) y New+Write+Sum (streaming)

HMAC:
- Hash + clave secreta para autenticacion de mensajes
- Siempre comparar con hmac.Equal (constant-time)
- Usado en webhooks, JWT HS256, AWS signatures

AES (cifrado simetrico):
- AES-256-GCM: cifrado autenticado (AEAD), recomendado
- Nonce de 12 bytes, NUNCA reutilizar con misma clave
- CBC/CTR solo para compatibilidad legacy

RSA (cifrado asimetrico):
- Par de claves publica/privada, 2048+ bits
- OAEP para cifrado, PSS para firmas
- Limitado por tamano de clave, usar con envelope encryption

ECDSA (firmas con curvas elipticas):
- P-256 recomendada, claves mas pequenas que RSA
- SignASN1/VerifyASN1 para API moderna

Ed25519 (firmas modernas):
- 32 bytes pub, 64 bytes priv, 64 bytes firma
- Determinista, rapido, resistente a side-channels
- Preferido para SSH, VPN, protocolos modernos

x509 (certificados):
- Vincula clave publica con identidad
- Root CA -> Intermediate CA -> Leaf certificate
- PEM (texto) y DER (binario) formatos
- Verificacion con pools de certificados confiables

TLS:
- tls.Config para servidor y cliente
- MinVersion TLS 1.2, preferir TLS 1.3
- mTLS para autenticacion mutua

Key Derivation (KDF):
- Argon2id: recomendado para passwords (2015)
- bcrypt: probado por decadas, amplio soporte
- PBKDF2: NIST aprobado, compatibilidad
- scrypt: memory-hard alternativa

crypto/subtle:
- ConstantTimeCompare para evitar timing attacks
- Usar en validacion de tokens, API keys, MACs

FIPS 140-3:
- GOEXPERIMENT=boringcrypto para compliance
- Go 1.24+ con soporte nativo
- Solo algoritmos aprobados (AES, SHA-2, RSA 2048+, ECDSA)

Patrones practicos:
- Envelope encryption para datos grandes
- Hybrid encryption (RSA + AES-GCM)
- Nonce management con limites de GCM
- PEM encoding para serializar claves
*/

/* SUMMARY - CHAPTER 055: Cryptography
Topics covered in this chapter:

• crypto/rand: cryptographically secure random number generation with Reader
• Hashing: SHA-256, SHA-512 for integrity, never for passwords
• HMAC: message authentication codes for verifying data authenticity
• AES-GCM encryption: authenticated encryption with associated data (AEAD)
• Nonce management: unique nonces per encryption, limits for GCM rekey
• RSA encryption and signatures: OAEP for encryption, PSS for signatures, envelope encryption
• ECDSA: elliptic curve signatures with P-256, smaller keys than RSA
• Ed25519: modern deterministic signatures, fast and side-channel resistant
• x509 certificates: public key infrastructure, certificate chains, PEM/DER formats
• TLS configuration: tls.Config, MinVersion TLS 1.2, mTLS for mutual authentication
• Key derivation: Argon2id (recommended), bcrypt, PBKDF2, scrypt for password hashing
• crypto/subtle: ConstantTimeCompare for timing attack prevention
• FIPS 140-3: boringcrypto for compliance, approved algorithms only
• Best practices: envelope encryption, hybrid encryption, proper nonce management, PEM encoding
*/
