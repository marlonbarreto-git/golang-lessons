// Package main - Chapter 051: Hash Functions
// Non-cryptographic hashing in Go: hash.Hash interface,
// fnv, crc32, crc64, adler32, and maphash for fast hashing.
package main

import (
	"fmt"
	"hash"
	"hash/adler32"
	"hash/crc32"
	"hash/crc64"
	"hash/fnv"
	"hash/maphash"
	"os"
)

func main() {
	fmt.Println("=== HASH FUNCTIONS IN GO ===")

	// ============================================
	// HASH.HASH INTERFACE
	// ============================================
	fmt.Println("\n--- hash.Hash Interface ---")
	fmt.Println(`
type Hash interface {
    io.Writer                // Write(p []byte) (n int, err error)
    Sum(b []byte) []byte     // Append current hash to b
    Reset()                  // Reset to initial state
    Size() int               // Number of bytes Sum returns
    BlockSize() int          // Block size of the hash
}

type Hash32 interface {
    Hash
    Sum32() uint32
}

type Hash64 interface {
    Hash
    Sum64() uint64
}

ALL HASH FUNCTIONS SATISFY io.Writer:
- Write data incrementally
- Call Sum() when done
- Thread-safe for reading after final Write`)

	// ============================================
	// FNV HASH
	// ============================================
	fmt.Println("\n--- FNV (Fowler-Noll-Vo) ---")

	data := []byte("Hello, Go hashing!")

	h32 := fnv.New32()
	h32.Write(data)
	fmt.Printf("FNV-1  32-bit: %d\n", h32.Sum32())

	h32a := fnv.New32a()
	h32a.Write(data)
	fmt.Printf("FNV-1a 32-bit: %d\n", h32a.Sum32())

	h64 := fnv.New64()
	h64.Write(data)
	fmt.Printf("FNV-1  64-bit: %d\n", h64.Sum64())

	h64a := fnv.New64a()
	h64a.Write(data)
	fmt.Printf("FNV-1a 64-bit: %d\n", h64a.Sum64())

	h128 := fnv.New128()
	h128.Write(data)
	fmt.Printf("FNV-1  128-bit: %x\n", h128.Sum(nil))

	h128a := fnv.New128a()
	h128a.Write(data)
	fmt.Printf("FNV-1a 128-bit: %x\n", h128a.Sum(nil))

	fmt.Println(`
FNV VARIANTS:
fnv.New32()   / fnv.New32a()   - 32-bit
fnv.New64()   / fnv.New64a()   - 64-bit
fnv.New128()  / fnv.New128a()  - 128-bit

FNV-1a is generally preferred (better distribution)

USE CASES:
- Hash maps / hash tables
- Checksums for data integrity
- Bloom filters
- Consistent hashing
- Cache keys`)

	// ============================================
	// CRC32
	// ============================================
	fmt.Println("\n--- CRC32 ---")

	ieee := crc32.ChecksumIEEE(data)
	fmt.Printf("CRC32 IEEE: %d (0x%08x)\n", ieee, ieee)

	castagnoli := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	castagnoli.Write(data)
	fmt.Printf("CRC32 Castagnoli: %d (0x%08x)\n", castagnoli.Sum32(), castagnoli.Sum32())

	koopman := crc32.New(crc32.MakeTable(crc32.Koopman))
	koopman.Write(data)
	fmt.Printf("CRC32 Koopman: %d (0x%08x)\n", koopman.Sum32(), koopman.Sum32())

	fmt.Println(`
CRC32 POLYNOMIALS:
crc32.IEEE       - Most common (Ethernet, ZIP, PNG)
crc32.Castagnoli - Better error detection, hardware accelerated (SSE4.2)
crc32.Koopman    - Good for short messages

FUNCTIONS:
crc32.ChecksumIEEE(data)         - Quick one-shot
crc32.New(table)                 - Streaming
crc32.MakeTable(polynomial)      - Custom table
crc32.Update(crc, table, data)   - Update existing CRC

HARDWARE ACCELERATION:
Castagnoli uses CPU CRC32C instruction when available
(Intel SSE4.2, ARM CRC32)`)

	// ============================================
	// CRC64
	// ============================================
	fmt.Println("\n--- CRC64 ---")

	isoTable := crc64.MakeTable(crc64.ISO)
	ecmaTable := crc64.MakeTable(crc64.ECMA)

	isoHash := crc64.New(isoTable)
	isoHash.Write(data)
	fmt.Printf("CRC64 ISO: %d\n", isoHash.Sum64())

	ecmaHash := crc64.New(ecmaTable)
	ecmaHash.Write(data)
	fmt.Printf("CRC64 ECMA: %d\n", ecmaHash.Sum64())

	fmt.Println(`
CRC64 POLYNOMIALS:
crc64.ISO  - ISO 3309
crc64.ECMA - ECMA-182

USE CASES:
- Large file checksums
- Database checksums
- Distributed systems (etcd uses CRC64)`)

	// ============================================
	// ADLER32
	// ============================================
	fmt.Println("\n--- Adler32 ---")

	a32 := adler32.Checksum(data)
	fmt.Printf("Adler32: %d (0x%08x)\n", a32, a32)

	ah := adler32.New()
	ah.Write([]byte("Hello, "))
	ah.Write([]byte("Go hashing!"))
	fmt.Printf("Adler32 streaming: %d\n", ah.Sum32())

	fmt.Println(`
ADLER32:
- Faster than CRC32 but weaker error detection
- Used in zlib/deflate compression
- Simple: A = 1 + sum of bytes, B = sum of all A values
- 32-bit: upper 16 = B, lower 16 = A

FUNCTIONS:
adler32.Checksum(data) - One-shot
adler32.New()          - Streaming`)

	// ============================================
	// MAPHASH (Go 1.14+)
	// ============================================
	fmt.Println("\n--- maphash (Go 1.14+) ---")

	var mh maphash.Hash
	mh.WriteString("Hello")
	hash1 := mh.Sum64()
	mh.Reset()
	mh.WriteString("Hello")
	hash2 := mh.Sum64()
	fmt.Printf("Same seed, same input: %d == %d -> %v\n", hash1, hash2, hash1 == hash2)

	var mh2 maphash.Hash
	mh2.SetSeed(mh.Seed())
	mh2.WriteString("Hello")
	hash3 := mh2.Sum64()
	fmt.Printf("Shared seed: %d == %d -> %v\n", hash1, hash3, hash1 == hash3)

	fmt.Println(`
MAPHASH:
- Random seed per Hash instance
- Same seed + same data = same hash (within process)
- Different process runs = different hashes (random seed)
- NOT suitable for persistence or cross-process comparison

FUNCTIONS:
maphash.Bytes(seed, data)   - One-shot (Go 1.19+)
maphash.String(seed, data)  - One-shot for strings (Go 1.19+)

USE CASES:
- Building custom hash maps
- Probabilistic data structures (Bloom filters, HyperLogLog)
- Random bucketing within a process`)

	// ============================================
	// INCREMENTAL HASHING
	// ============================================
	fmt.Println("\n--- Incremental Hashing ---")

	hasher := fnv.New64a()
	chunks := []string{"Hello", ", ", "World", "!"}
	for _, chunk := range chunks {
		hasher.Write([]byte(chunk))
	}
	incrementalHash := hasher.Sum64()

	hasher.Reset()
	hasher.Write([]byte("Hello, World!"))
	oneShot := hasher.Sum64()

	fmt.Printf("Incremental: %d\n", incrementalHash)
	fmt.Printf("One-shot:    %d\n", oneShot)
	fmt.Printf("Equal: %v\n", incrementalHash == oneShot)

	fmt.Println(`
INCREMENTAL HASHING PATTERN:
h := fnv.New64a()
for chunk := range dataSource {
    h.Write(chunk)
}
result := h.Sum64()

USEFUL FOR:
- Streaming data (files too large for memory)
- Network data
- io.Copy(hasher, reader)`)

	// ============================================
	// HASHING FILES
	// ============================================
	fmt.Println("\n--- Hashing Files ---")
	os.Stdout.WriteString(`
HASH A FILE:
func hashFile(path string) (uint32, error) {
    f, err := os.Open(path)
    if err != nil {
        return 0, err
    }
    defer f.Close()

    h := crc32.NewIEEE()
    if _, err := io.Copy(h, f); err != nil {
        return 0, err
    }
    return h.Sum32(), nil
}

MULTI-HASH (compute multiple hashes in one pass):
func multiHash(r io.Reader) (uint32, uint64) {
    crc := crc32.NewIEEE()
    fnvH := fnv.New64a()
    w := io.MultiWriter(crc, fnvH)
    io.Copy(w, r)
    return crc.Sum32(), fnvH.Sum64()
}
`)

	// ============================================
	// CHOOSING A HASH
	// ============================================
	fmt.Println("\n--- Choosing a Hash Function ---")
	fmt.Println(`
DECISION GUIDE:

Need crypto security?     -> crypto/sha256, crypto/sha512
Need speed, good distrib? -> fnv.New64a() or maphash
Need error detection?     -> crc32 (Castagnoli for HW accel)
Need compression compat?  -> adler32 (zlib) or crc32 (gzip)
Need 128-bit?             -> fnv.New128a()
Need process-local only?  -> maphash (fastest)
Need persistence?         -> fnv or crc32 (deterministic)

PERFORMANCE (approximate, smaller = faster):
maphash   ~1 ns/op   (fastest, random seed)
crc32-C   ~2 ns/op   (HW accelerated)
fnv-1a    ~3 ns/op   (good general purpose)
adler32   ~3 ns/op   (compression)
crc32-IEEE ~5 ns/op  (standard)

NON-CRYPTO vs CRYPTO:
Non-crypto: fast, NOT collision resistant
Crypto: slower, collision resistant, preimage resistant`)

	// ============================================
	// CONSISTENT HASHING
	// ============================================
	fmt.Println("\n--- Consistent Hashing Pattern ---")

	consistentHash := func(key string, nodes []string) string {
		h := fnv.New32a()
		h.Write([]byte(key))
		idx := int(h.Sum32()) % len(nodes)
		if idx < 0 {
			idx = -idx
		}
		return nodes[idx]
	}

	nodes := []string{"server-1", "server-2", "server-3"}
	keys := []string{"user:1001", "user:1002", "user:1003", "user:1004"}
	for _, key := range keys {
		node := consistentHash(key, nodes)
		fmt.Printf("Key %q -> %s\n", key, node)
	}

	// ============================================
	// GENERIC HASH HELPER
	// ============================================
	fmt.Println("\n--- Generic Hash Helper ---")

	hashWith := func(h hash.Hash, data []byte) []byte {
		h.Reset()
		h.Write(data)
		return h.Sum(nil)
	}

	testData := []byte("test data")
	fmt.Printf("FNV-1a 64: %x\n", hashWith(fnv.New64a(), testData))
	fmt.Printf("CRC32 IEEE: %x\n", hashWith(crc32.NewIEEE(), testData))
	fmt.Printf("Adler32: %x\n", hashWith(adler32.New(), testData))
}

/*
SUMMARY - HASH FUNCTIONS:

INTERFACE:
hash.Hash     - Write, Sum, Reset, Size, BlockSize
hash.Hash32   - + Sum32()
hash.Hash64   - + Sum64()

PACKAGES:
hash/fnv      - FNV-1 and FNV-1a (32, 64, 128 bit)
hash/crc32    - CRC32 (IEEE, Castagnoli, Koopman)
hash/crc64    - CRC64 (ISO, ECMA)
hash/adler32  - Adler-32 (fast, weak)
hash/maphash  - Process-local random hashing

PATTERNS:
io.Copy(hasher, reader)          - Hash a file/stream
io.MultiWriter(hash1, hash2)     - Multiple hashes in one pass
hasher.Write(data); hasher.Sum64() - Incremental hashing

CHOOSE:
Speed + distribution -> fnv.New64a()
Error detection     -> crc32 Castagnoli
Compression         -> adler32
Process-local       -> maphash
Crypto security     -> crypto/sha256
*/
