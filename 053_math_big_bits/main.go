// Package main - Chapter 053: Math Big, Bits & Rand v2
// Aritmetica de precision arbitraria con math/big (Int, Float, Rat),
// manipulacion de bits a nivel de palabra con math/bits,
// y la nueva API de numeros aleatorios math/rand/v2 (Go 1.22+).
package main

import (
	"fmt"
	"math/big"
	"math/bits"
	"math/rand/v2"
	"os"
	"time"
)

func main() {
	fmt.Println("=== MATH/BIG, MATH/BITS & MATH/RAND/V2 ===")

	// ============================================
	// 1. MATH/BIG - BIG.INT (ENTEROS GIGANTES)
	// ============================================
	fmt.Println("\n--- 1. math/big - big.Int ---")
	os.Stdout.WriteString(`
BIG.INT - ENTEROS DE PRECISION ARBITRARIA:

  big.NewInt(x int64) *big.Int          // Crear desde int64
  new(big.Int).SetString("123", 10)     // Crear desde string en base 10
  z.Add(x, y)                           // z = x + y
  z.Sub(x, y)                           // z = x - y
  z.Mul(x, y)                           // z = x * y
  z.Div(x, y)                           // z = x / y (truncado)
  z.Mod(x, y)                           // z = x mod y
  z.Exp(x, y, m)                        // z = x^y mod m (si m != nil)
  z.GCD(x, y, a, b)                     // z = GCD(a, b), ax + by = z
  z.ProbablyPrime(n)                    // Test de primalidad Miller-Rabin
  z.ModInverse(g, n)                    // z = g^(-1) mod n
  x.Cmp(y)                              // -1, 0, +1

USOS COMUNES:
  - Criptografia (RSA, claves grandes)
  - Calculos financieros exactos
  - Factoriales y combinatorias enormes
  - Numeros de Fibonacci arbitrarios
`)

	a := big.NewInt(100)
	b := big.NewInt(37)

	sum := new(big.Int).Add(a, b)
	fmt.Printf("  100 + 37 = %s\n", sum)

	product := new(big.Int).Mul(a, b)
	fmt.Printf("  100 * 37 = %s\n", product)

	gcd := new(big.Int).GCD(nil, nil, big.NewInt(48), big.NewInt(18))
	fmt.Printf("  GCD(48, 18) = %s\n", gcd)

	factorial := computeFactorial(50)
	fmt.Printf("  50! = %s\n", factorial)

	modInv := new(big.Int).ModInverse(big.NewInt(3), big.NewInt(11))
	fmt.Printf("  3^(-1) mod 11 = %s (porque 3*4=12=1 mod 11)\n", modInv)

	prime := big.NewInt(104729)
	fmt.Printf("  %d es primo? %v\n", prime, prime.ProbablyPrime(20))
	composite := big.NewInt(104730)
	fmt.Printf("  %d es primo? %v\n", composite, composite.ProbablyPrime(20))

	// Exponenciacion modular: 7^256 mod 13
	base := big.NewInt(7)
	exp := big.NewInt(256)
	mod := big.NewInt(13)
	result := new(big.Int).Exp(base, exp, mod)
	fmt.Printf("  7^256 mod 13 = %s\n", result)

	// ============================================
	// 2. MATH/BIG - BIG.FLOAT (FLOTANTES PRECISION)
	// ============================================
	fmt.Println("\n--- 2. math/big - big.Float ---")
	os.Stdout.WriteString(`
BIG.FLOAT - FLOTANTES DE PRECISION ARBITRARIA:

  big.NewFloat(x float64)               // Crear desde float64
  f.SetPrec(prec uint)                  // Establecer precision en bits
  f.SetString("3.14159265358979323846") // Desde string
  f.Add(x, y), f.Sub(x, y)             // Aritmetica
  f.Mul(x, y), f.Quo(x, y)             // Multiplicacion, Division
  f.Cmp(y)                              // Comparacion
  f.Text('f', prec)                     // Formato decimal
  f.Prec()                              // Precision actual en bits

DIFERENCIA CON FLOAT64:
  float64: ~15-17 digitos significativos (IEEE 754)
  big.Float: precision configurable (256, 512, 1024+ bits)
`)

	pi, _, _ := new(big.Float).SetPrec(256).Parse("3.14159265358979323846264338327950288419716939937510", 10)
	fmt.Printf("  Pi con 256 bits: %s\n", pi.Text('f', 50))

	piLow := new(big.Float).SetPrec(32).SetFloat64(3.14159265358979323846)
	fmt.Printf("  Pi con 32 bits:  %s\n", piLow.Text('f', 50))

	one := new(big.Float).SetPrec(256).SetFloat64(1.0)
	three := new(big.Float).SetPrec(256).SetFloat64(3.0)
	third := new(big.Float).SetPrec(256).Quo(one, three)
	fmt.Printf("  1/3 con 256 bits: %s\n", third.Text('f', 60))

	x := new(big.Float).SetPrec(128).SetFloat64(0.1)
	y := new(big.Float).SetPrec(128).SetFloat64(0.2)
	z := new(big.Float).SetPrec(128).Add(x, y)
	expected := new(big.Float).SetPrec(128).SetFloat64(0.3)
	fmt.Printf("  0.1 + 0.2 = %s\n", z.Text('f', 20))
	fmt.Printf("  Igual a 0.3? %v (Cmp=%d)\n", z.Cmp(expected) == 0, z.Cmp(expected))

	// ============================================
	// 3. MATH/BIG - BIG.RAT (RACIONALES EXACTOS)
	// ============================================
	fmt.Println("\n--- 3. math/big - big.Rat ---")
	os.Stdout.WriteString(`
BIG.RAT - NUMEROS RACIONALES EXACTOS:

  big.NewRat(a, b int64)                // a/b
  r.SetString("355/113")               // Desde string
  r.Add(x, y), r.Sub(x, y)            // Aritmetica exacta
  r.Mul(x, y), r.Quo(x, y)            // Sin perdida de precision
  r.Num(), r.Denom()                   // Numerador, Denominador
  r.FloatString(prec int)              // Representacion decimal
  r.RatString()                        // Representacion "a/b"
  r.IsInt()                            // Es entero?
`)

	r1 := big.NewRat(1, 3)
	r2 := big.NewRat(1, 6)
	rSum := new(big.Rat).Add(r1, r2)
	fmt.Printf("  1/3 + 1/6 = %s\n", rSum.RatString())
	fmt.Printf("  Decimal: %s\n", rSum.FloatString(10))

	piApprox := new(big.Rat).SetFrac64(355, 113)
	fmt.Printf("  355/113 = %s (aproximacion de Pi)\n", piApprox.FloatString(15))

	r3 := big.NewRat(7, 4)
	r4 := big.NewRat(3, 8)
	rMul := new(big.Rat).Mul(r3, r4)
	fmt.Printf("  7/4 * 3/8 = %s = %s\n", rMul.RatString(), rMul.FloatString(6))
	fmt.Printf("  Es entero? %v\n", rMul.IsInt())

	rInt := big.NewRat(10, 5)
	fmt.Printf("  10/5 es entero? %v (valor: %s)\n", rInt.IsInt(), rInt.RatString())

	// ============================================
	// 4. MATH/BITS - OPERACIONES A NIVEL DE BITS
	// ============================================
	fmt.Println("\n--- 4. math/bits ---")
	os.Stdout.WriteString(`
MATH/BITS - FUNCIONES OPTIMIZADAS DE MANIPULACION DE BITS:

  bits.OnesCount(x uint)               // Contar bits en 1 (popcount)
  bits.Len(x uint)                     // Longitud en bits (posicion del bit mas alto)
  bits.TrailingZeros(x uint)           // Ceros a la derecha
  bits.LeadingZeros(x uint)            // Ceros a la izquierda
  bits.RotateLeft(x uint, k int)       // Rotacion circular a la izquierda
  bits.ReverseBytes(x uint)            // Invertir orden de bytes
  bits.Reverse(x uint)                 // Invertir orden de bits
  bits.Add(x, y, carry uint)           // Suma con carry
  bits.Mul(x, y uint)                  // Multiplicacion (hi, lo)
  bits.Sub(x, y, borrow uint)          // Resta con borrow
  bits.Div(hi, lo, d uint)             // Division (q, r)

Todas tienen variantes: 8, 16, 32, 64 (OnesCount8, OnesCount16, etc.)
`)

	val := uint(0b10110110)
	fmt.Printf("  Valor: %08b (decimal: %d)\n", val, val)
	fmt.Printf("  OnesCount:     %d (bits en 1)\n", bits.OnesCount(val))
	fmt.Printf("  Len:           %d (longitud en bits)\n", bits.Len(val))
	fmt.Printf("  TrailingZeros: %d\n", bits.TrailingZeros(val))
	fmt.Printf("  LeadingZeros:  %d (para uint de %d bits)\n",
		bits.LeadingZeros(val), bits.UintSize)

	val32 := uint32(0b10110110_00000000_11111111_00001010)
	fmt.Printf("\n  Valor 32-bit: %032b\n", val32)
	reversed := bits.ReverseBytes32(val32)
	fmt.Printf("  ReverseBytes:  %032b\n", reversed)
	bitReversed := bits.Reverse32(val32)
	fmt.Printf("  Reverse bits:  %032b\n", bitReversed)

	rotated := bits.RotateLeft32(val32, 8)
	fmt.Printf("  RotateLeft(8): %032b\n", rotated)

	hi, lo := bits.Mul64(0xFFFFFFFF, 0xFFFFFFFF)
	fmt.Printf("\n  Mul64(0xFFFFFFFF, 0xFFFFFFFF) = hi:%d lo:%d\n", hi, lo)

	sumVal, carry := bits.Add64(^uint64(0), 1, 0)
	fmt.Printf("  Add64(MaxUint64, 1, 0) = sum:%d carry:%d (overflow!)\n", sumVal, carry)

	// ============================================
	// 5. MATH/RAND/V2 - NUEVA API (GO 1.22+)
	// ============================================
	fmt.Println("\n--- 5. math/rand/v2 (Go 1.22+) ---")
	os.Stdout.WriteString(`
MATH/RAND/V2 - NUEVA API DE ALEATORIEDAD (Go 1.22+):

  Cambios vs math/rand:
  - Fuente global usa ChaCha8 (criptograficamente segura por defecto)
  - No necesitas Seed() - inicializado automaticamente
  - Funcion generica N[T]() para cualquier tipo numerico
  - IntN(n) reemplaza Intn(n) (mejor nombre)
  - Nuevas fuentes: ChaCha8, PCG

  rand.IntN(n int)                     // [0, n) - entero aleatorio
  rand.Int64()                         // int64 aleatorio
  rand.Float64()                       // [0.0, 1.0) float64
  rand.N[T](n T)                       // Generico: cualquier entero/duracion
  rand.Shuffle(n, swap func(i, j int)) // Barajar
  rand.NewChaCha8(seed [32]byte)       // Fuente ChaCha8 determinista
  rand.NewPCG(seed1, seed2 uint64)     // Fuente PCG determinista
`)

	fmt.Println("  Numeros aleatorios (fuente global ChaCha8):")
	for i := 0; i < 5; i++ {
		fmt.Printf("    IntN(100): %d\n", rand.IntN(100))
	}

	fmt.Printf("\n  Float64: %.6f\n", rand.Float64())
	fmt.Printf("  Int64:   %d\n", rand.Int64())

	fmt.Println("\n  Funcion generica N[T]():")
	fmt.Printf("    N[int](1000):          %d\n", rand.N[int](1000))
	fmt.Printf("    N[int32](50):          %d\n", rand.N[int32](50))
	fmt.Printf("    N[time.Duration](5s):  %v\n", rand.N[time.Duration](5*time.Second))

	words := []string{"Go", "Rust", "Python", "Java", "TypeScript"}
	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
	fmt.Printf("\n  Shuffle: %v\n", words)

	fmt.Println("\n  Fuente determinista con PCG:")
	src := rand.NewPCG(42, 99)
	rng := rand.New(src)
	fmt.Print("    PCG(42,99) -> ")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", rng.IntN(100))
	}
	fmt.Println()

	src2 := rand.NewPCG(42, 99)
	rng2 := rand.New(src2)
	fmt.Print("    PCG(42,99) -> ")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", rng2.IntN(100))
	}
	fmt.Println("(misma semilla = misma secuencia)")

	fmt.Println("\n  Fuente determinista con ChaCha8:")
	var seed [32]byte
	seed[0] = 1
	seed[1] = 2
	chachaRng := rand.New(rand.NewChaCha8(seed))
	fmt.Print("    ChaCha8 -> ")
	for i := 0; i < 5; i++ {
		fmt.Printf("%d ", chachaRng.IntN(100))
	}
	fmt.Println()

	// ============================================
	// 6. EJEMPLO PRACTICO: FIBONACCI GRANDE
	// ============================================
	fmt.Println("\n--- 6. Ejemplo: Fibonacci con big.Int ---")

	fib100 := fibonacci(100)
	fmt.Printf("  Fibonacci(100) = %s\n", fib100)

	fib500 := fibonacci(500)
	fmt.Printf("  Fibonacci(500) = %s\n", fib500)

	// ============================================
	// 7. EJEMPLO PRACTICO: CONTEO DE BITS EN RED
	// ============================================
	fmt.Println("\n--- 7. Ejemplo: Analisis de mascara de red ---")

	mask := uint32(0xFFFFFF00) // /24
	ones := bits.OnesCount32(mask)
	fmt.Printf("  Mascara: %08x -> /%d\n", mask, ones)

	mask2 := uint32(0xFFFFF800) // /21
	ones2 := bits.OnesCount32(mask2)
	fmt.Printf("  Mascara: %08x -> /%d\n", mask2, ones2)

	isPowerOf2 := func(n uint) bool {
		return n > 0 && bits.OnesCount(n) == 1
	}
	for _, v := range []uint{1, 2, 3, 4, 16, 31, 32, 64, 100, 128} {
		if isPowerOf2(v) {
			fmt.Printf("  %d es potencia de 2\n", v)
		}
	}

	fmt.Println("\n=== FIN CAPITULO 053 ===")
}

func computeFactorial(n int64) *big.Int {
	result := big.NewInt(1)
	for i := int64(2); i <= n; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result
}

func fibonacci(n int) *big.Int {
	if n <= 1 {
		return big.NewInt(int64(n))
	}
	a := big.NewInt(0)
	b := big.NewInt(1)
	for i := 2; i <= n; i++ {
		a.Add(a, b)
		a, b = b, a
	}
	return b
}

/*
SUMMARY - MATH/BIG, MATH/BITS & MATH/RAND/V2:

MATH/BIG - ARBITRARY PRECISION ARITHMETIC:
- big.Int for integers of unlimited size (factorials, crypto, Fibonacci)
- big.Float for configurable precision floating point (256+ bits)
- big.Rat for exact rational arithmetic (no precision loss)
- Operations: Add, Sub, Mul, Div, Mod, Exp, GCD, ModInverse, ProbablyPrime

MATH/BITS - OPTIMIZED BIT MANIPULATION:
- OnesCount (popcount), Len, TrailingZeros, LeadingZeros
- RotateLeft, ReverseBytes, Reverse for bit-level transformations
- Add64, Mul64, Sub64, Div64 for extended arithmetic with carry/borrow
- All functions have 8, 16, 32, 64 bit variants

MATH/RAND/V2 - NEW RANDOM API (Go 1.22+):
- Global source uses ChaCha8 (cryptographically secure by default)
- No Seed() needed - auto-initialized
- Generic N[T]() for any numeric type including time.Duration
- IntN replaces Intn, deterministic sources: PCG and ChaCha8
- rand.Shuffle for randomizing slices

PRACTICAL EXAMPLES:
- Computing large factorials (50!)
- Fibonacci with arbitrary precision (Fib(500))
- Network mask analysis using bit counting
- Power-of-2 detection with OnesCount
*/
