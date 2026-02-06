// Package main - Chapter 072: Go Assembly
// Go usa una variante de Plan 9 assembly. Aprenderás la sintaxis básica,
// cómo escribir funciones en asm, y cuándo tiene sentido usarlo.
package main

import (
	"fmt"
	"unsafe"
)

func main() {
	fmt.Println("=== GO ASSEMBLY ===")

	fmt.Println(`
Go Assembly es un assembly pseudo-portátil basado en Plan 9.
El compilador de Go traduce este assembly a código máquina nativo.

ADVERTENCIA: El assembly es difícil de mantener, propenso a errores,
y rara vez necesario. Solo úsalo cuando:
1. Necesites instrucciones SIMD específicas
2. Tengas código crítico verificado con benchmarks
3. Necesites acceso a instrucciones no expuestas por Go
4. Estés implementando crypto con timing constant

En la MAYORÍA de casos, Go optimiza mejor de lo que escribirías a mano.`)
	// ============================================
	// LLAMAR FUNCIONES EN ASSEMBLY
	// ============================================
	fmt.Println("\n--- Llamar Funciones Assembly ---")

	// Estas funciones están implementadas en asm_amd64.s
	// Para este ejemplo, mostramos las versiones Go
	resultado := AddAsm(10, 20)
	fmt.Printf("AddAsm(10, 20) = %d\n", resultado)

	resultado2 := MultiplyAsm(7, 8)
	fmt.Printf("MultiplyAsm(7, 8) = %d\n", resultado2)

	// ============================================
	// ESTRUCTURA DE ARCHIVO ASM
	// ============================================
	fmt.Println("\n--- Estructura de Archivo .s ---")
	fmt.Println(`
Los archivos assembly en Go usan extensión .s
Nombre: nombre_GOOS_GOARCH.s o nombre_GOARCH.s

Ejemplo: math_amd64.s

// +build amd64

#include "textflag.h"

// func Add(a, b int64) int64
TEXT ·Add(SB), NOSPLIT, $0-24
    MOVQ a+0(FP), AX     // primer argumento
    ADDQ b+8(FP), AX     // sumar segundo argumento
    MOVQ AX, ret+16(FP)  // retornar resultado
    RET`)
	// ============================================
	// REGISTROS AMD64
	// ============================================
	fmt.Println("\n--- Registros AMD64 ---")
	fmt.Println(`
Registros de propósito general (64-bit):
  AX, BX, CX, DX - registros tradicionales
  SI, DI         - source/destination index
  BP             - base pointer (frame pointer)
  SP             - stack pointer
  R8-R15         - registros adicionales x86-64

Registros SIMD:
  X0-X15         - SSE (128-bit)
  Y0-Y15         - AVX (256-bit)
  Z0-Z31         - AVX-512 (512-bit)

Pseudo-registros Go:
  FP             - frame pointer (argumentos)
  PC             - program counter
  SB             - static base (símbolos globales)
  SP             - stack pointer (local vars)`)
	// ============================================
	// CONVENCIÓN DE LLAMADA
	// ============================================
	fmt.Println("\n--- Convención de Llamada ---")
	fmt.Println(`
Go 1.17+ usa convención de llamada basada en registros:
- Argumentos en registros: AX, BX, CX, DI, SI, R8-R11
- Retornos en registros: AX, BX, CX, DI, SI, R8-R11
- Stack para argumentos que no caben

Frame layout (pre-1.17 / ABI0):
  +----------------+
  | return values  | ← ret+N(FP)
  | arguments      | ← arg+0(FP)
  | return address |
  | saved BP       |
  | local vars     | ← -N(SP)
  +----------------+`)
	// ============================================
	// SINTAXIS BÁSICA
	// ============================================
	fmt.Println("\n--- Sintaxis Básica ---")
	fmt.Println(`
Instrucciones comunes:
  MOVQ src, dst       // move quad (64-bit)
  MOVL src, dst       // move long (32-bit)
  MOVW src, dst       // move word (16-bit)
  MOVB src, dst       // move byte (8-bit)

  ADDQ src, dst       // add quad
  SUBQ src, dst       // subtract quad
  IMULQ src, dst      // signed multiply

  CMPQ a, b           // compare
  JEQ label           // jump if equal
  JNE label           // jump if not equal
  JL label            // jump if less
  JG label            // jump if greater

  CALL symbol(SB)     // call function
  RET                 // return

Sufijos de tamaño:
  Q = quad (64-bit)
  L = long (32-bit)
  W = word (16-bit)
  B = byte (8-bit)

Direccionamiento:
  $123               // immediate value
  AX                 // register direct
  (AX)               // register indirect
  8(AX)              // displacement
  symbol(SB)         // static base
  name+0(FP)         // frame pointer`)
	// ============================================
	// FLAGS DE TEXTO
	// ============================================
	fmt.Println("\n--- Flags de TEXT ---")
	fmt.Println(`
TEXT symbol(SB), flags, $framesize-argsize

Flags comunes:
  NOSPLIT    - no insertar preámbulo de stack split
  NOFRAME    - sin frame pointer (implica NOSPLIT)
  WRAPPER    - es un wrapper, no aparece en stack traces
  NEEDCTXT   - necesita closure context

Tamaños:
  $framesize - espacio para variables locales
  -argsize   - tamaño de argumentos + retornos

Ejemplo:
  TEXT ·Add(SB), NOSPLIT, $0-24
  // 0 bytes de frame local
  // 24 bytes de args: int64 + int64 + int64`)
	// ============================================
	// EJEMPLO: SUMA SIMD
	// ============================================
	fmt.Println("\n--- Ejemplo SIMD ---")
	fmt.Println(`
// Sumar arrays de float32 usando SSE
// func AddFloat32(a, b, result []float32, n int)
TEXT ·AddFloat32(SB), NOSPLIT, $0
    MOVQ a_data+0(FP), SI    // ptr a
    MOVQ b_data+24(FP), DI   // ptr b
    MOVQ result_data+48(FP), DX // ptr result
    MOVQ n+72(FP), CX        // count

loop:
    CMPQ CX, $4
    JL scalar                 // menos de 4, usar escalar

    MOVUPS (SI), X0          // cargar 4 floats de a
    MOVUPS (DI), X1          // cargar 4 floats de b
    ADDPS X1, X0             // sumar
    MOVUPS X0, (DX)          // guardar resultado

    ADDQ $16, SI             // avanzar punteros
    ADDQ $16, DI
    ADDQ $16, DX
    SUBQ $4, CX              // decrementar contador
    JMP loop

scalar:
    // Procesar elementos restantes uno por uno
    ...
    RET`)
	// ============================================
	// INSTRUCCIONES DE BAJO NIVEL
	// ============================================
	fmt.Println("\n--- Instrucciones de Bajo Nivel ---")
	fmt.Println(`
INSTRUCCIONES ARITMÉTICAS:
  ADDQ src, dst      // dst = dst + src (64-bit)
  SUBQ src, dst      // dst = dst - src
  IMULQ src, dst     // dst = dst * src (signed)
  MULQ src           // DX:AX = AX * src (unsigned)
  IDIVQ src          // AX = DX:AX / src, DX = remainder
  NEGQ dst           // dst = -dst
  INCQ dst           // dst++
  DECQ dst           // dst--

INSTRUCCIONES LÓGICAS:
  ANDQ src, dst      // dst = dst & src
  ORQ src, dst       // dst = dst | src
  XORQ src, dst      // dst = dst ^ src
  NOTQ dst           // dst = ~dst
  SHLQ n, dst        // dst = dst << n (shift left)
  SHRQ n, dst        // dst = dst >> n (shift right, unsigned)
  SARQ n, dst        // dst = dst >> n (shift right, signed)
  ROLQ n, dst        // rotate left
  RORQ n, dst        // rotate right

INSTRUCCIONES DE COMPARACIÓN Y SALTO:
  CMPQ a, b          // set flags based on a - b
  TESTQ a, b         // set flags based on a & b
  JEQ label          // jump if equal (ZF=1)
  JNE label          // jump if not equal (ZF=0)
  JL label           // jump if less (signed)
  JLE label          // jump if less or equal
  JG label           // jump if greater (signed)
  JGE label          // jump if greater or equal
  JB label           // jump if below (unsigned)
  JBE label          // jump if below or equal
  JA label           // jump if above (unsigned)
  JAE label          // jump if above or equal
  JZ label           // jump if zero
  JNZ label          // jump if not zero
  JMP label          // unconditional jump

INSTRUCCIONES DE MEMORIA:
  MOVQ src, dst      // move 64-bit
  LEAQ src, dst      // load effective address
  XCHGQ a, b         // exchange values
  CMPXCHGQ src, dst  // compare and exchange (atomic)
  BSWAPQ dst         // byte swap (endianness)

INSTRUCCIONES DE STACK:
  PUSHQ src          // push to stack
  POPQ dst           // pop from stack
  CALL symbol        // call function
  RET                // return

INSTRUCCIONES DE BITS:
  BSFQ src, dst      // bit scan forward (find first 1)
  BSRQ src, dst      // bit scan reverse (find last 1)
  POPCNTQ src, dst   // population count (count 1s)
  LZCNTQ src, dst    // leading zero count
  TZCNTQ src, dst    // trailing zero count

INSTRUCCIONES CONDICIONALES (CMOV):
  CMOVQEQ src, dst   // conditional move if equal
  CMOVQNE src, dst   // conditional move if not equal
  CMOVQLT src, dst   // conditional move if less than
  CMOVQGT src, dst   // conditional move if greater than`)
	// ============================================
	// INSTRUCCIONES SIMD DETALLADAS
	// ============================================
	fmt.Println("\n--- Instrucciones SIMD Detalladas ---")
	fmt.Println(`
SSE (128-bit, registros X0-X15):

Carga y almacenamiento:
  MOVUPS src, dst    // move unaligned packed single (4 x float32)
  MOVAPS src, dst    // move aligned packed single
  MOVUPD src, dst    // move unaligned packed double (2 x float64)
  MOVAPD src, dst    // move aligned packed double
  MOVDQU src, dst    // move unaligned packed integers
  MOVDQA src, dst    // move aligned packed integers

Aritmética flotante:
  ADDPS X1, X0       // X0 = X0 + X1 (4 x float32)
  SUBPS X1, X0       // X0 = X0 - X1
  MULPS X1, X0       // X0 = X0 * X1
  DIVPS X1, X0       // X0 = X0 / X1
  SQRTPS X1, X0      // X0 = sqrt(X1)
  MAXPS X1, X0       // X0 = max(X0, X1)
  MINPS X1, X0       // X0 = min(X0, X1)

Aritmética entera:
  PADDD X1, X0       // packed add dword (4 x int32)
  PADDQ X1, X0       // packed add qword (2 x int64)
  PSUBD X1, X0       // packed subtract dword
  PMULLD X1, X0      // packed multiply dword

Comparación:
  CMPPS X1, X0, imm  // compare packed single
  PCMPEQD X1, X0     // compare equal dword
  PCMPGTD X1, X0     // compare greater than dword

Shuffle y reorganización:
  SHUFPS X1, X0, imm // shuffle packed single
  PSHUFD X1, X0, imm // shuffle packed dword
  PUNPCKLWD X1, X0   // unpack low words
  PUNPCKHWD X1, X0   // unpack high words

AVX (256-bit, registros Y0-Y15):
  VADDPS Y1, Y0, Y2  // Y2 = Y0 + Y1 (8 x float32)
  VMULPS Y1, Y0, Y2  // Y2 = Y0 * Y1
  VFMADD213PS Y2, Y1, Y0  // fused multiply-add

AVX-512 (512-bit, registros Z0-Z31):
  VADDPS Z1, Z0, Z2  // Z2 = Z0 + Z1 (16 x float32)
  VSCALEFPS Z1, Z0, Z2 // scale by power of 2`)
	// ============================================
	// INSTRUCCIONES ATÓMICAS
	// ============================================
	fmt.Println("\n--- Instrucciones Atómicas ---")
	fmt.Println(`
Go usa estas para sync/atomic:

LOCK prefix (hace la instrucción atómica):
  LOCK ADDQ $1, (AX)     // atomic increment
  LOCK XADDQ BX, (AX)    // atomic fetch-and-add
  LOCK CMPXCHGQ CX, (AX) // compare-and-swap
  LOCK XCHGQ BX, (AX)    // atomic exchange

Memory barriers:
  MFENCE               // full memory fence
  SFENCE               // store fence
  LFENCE               // load fence

Ejemplo de atomic.AddInt64:
TEXT ·AddInt64(SB), NOSPLIT, $0-24
    MOVQ ptr+0(FP), BX
    MOVQ delta+8(FP), AX
    LOCK
    XADDQ AX, (BX)
    ADDQ delta+8(FP), AX
    MOVQ AX, ret+16(FP)
    RET

Ejemplo de atomic.CompareAndSwapInt64:
TEXT ·CompareAndSwapInt64(SB), NOSPLIT, $0-25
    MOVQ ptr+0(FP), BX
    MOVQ old+8(FP), AX
    MOVQ new+16(FP), CX
    LOCK
    CMPXCHGQ CX, (BX)
    SETEQ ret+24(FP)
    RET`)
	// ============================================
	// INSTRUCCIONES CRYPTO
	// ============================================
	fmt.Println("\n--- Instrucciones Crypto (AES-NI) ---")
	fmt.Println(`
Intel AES-NI (usadas por crypto/aes):

  AESENC X1, X0       // one round of AES encryption
  AESENCLAST X1, X0   // final round of AES encryption
  AESDEC X1, X0       // one round of AES decryption
  AESDECLAST X1, X0   // final round of AES decryption
  AESKEYGENASSIST imm, X1, X0 // key expansion

Intel SHA extensions:
  SHA1RNDS4 imm, X1, X0  // SHA1 round
  SHA256RNDS2 X1, X0     // SHA256 round

CLMUL (usado para GCM):
  PCLMULQDQ imm, X1, X0  // carryless multiply

Ejemplo de AES block encrypt:
TEXT ·encryptBlockAsm(SB), NOSPLIT, $0
    MOVUPS in+0(FP), X0
    MOVUPS key+16(FP), X1
    PXOR X1, X0
    // rounds 1-9
    AESENC X2, X0
    // ... más rounds ...
    AESENCLAST X11, X0
    MOVUPS X0, out+32(FP)
    RET`)
	// ============================================
	// SIMD EXPERIMENTAL (Go 1.26+)
	// ============================================
	fmt.Println("\n--- SIMD Package (Go 1.26+) ---")
	fmt.Println(`
Go 1.26 introduce simd/archsimd (experimental):

import "simd/archsimd"

// GOEXPERIMENT=simd go build

// Tipos de vector:
Int8x16, Int16x8, Int32x4, Int64x2
Float32x4, Float64x2
// Y versiones de 256 y 512 bits

// Operaciones:
var a, b archsimd.Int32x4
result := a.Add(b)
result = a.Mul(b)
result = a.And(b)
result = a.ShiftLeft(n)`)
	// ============================================
	// GENERAR ASSEMBLY DESDE GO
	// ============================================
	fmt.Println("\n--- Generar Assembly ---")
	fmt.Println(`
Ver el assembly generado por Go:

go build -gcflags="-S" main.go 2>&1 | head -100

O con objdump:
go build -o main main.go
go tool objdump -s "main.Add" main

O generar archivo .s:
go build -gcflags="-S" main.go 2> main.s`)
	// ============================================
	// EJEMPLO PRÁCTICO
	// ============================================
	fmt.Println("\n--- Ejemplo Práctico: Popcnt ---")

	// Population count (contar bits en 1)
	// Go puede optimizar esto, pero es buen ejemplo
	n := uint64(0b10110110_11001010_10101010_11110000)
	fmt.Printf("Número: %b\n", n)
	fmt.Printf("Bits en 1 (Go): %d\n", PopcountGo(n))

	// ============================================
	// VERIFICAR OPTIMIZACIONES
	// ============================================
	fmt.Println("\n--- Verificar Optimizaciones ---")
	fmt.Println(`
Antes de escribir assembly, verifica qué hace Go:

1. go build -gcflags="-m" main.go
   → Ver escape analysis

2. go build -gcflags="-S" main.go
   → Ver assembly generado

3. go test -bench=. -benchmem
   → Medir performance

4. go tool pprof
   → Profiling detallado

A menudo, reorganizar el código Go es más efectivo
que escribir assembly.`)
	// ============================================
	// UNSAFE Y ASSEMBLY
	// ============================================
	fmt.Println("\n--- Unsafe para Interop ---")

	// A veces necesitas unsafe para estructuras de datos
	// que se pasan a assembly
	type Vector4 struct {
		X, Y, Z, W float32
	}

	v := Vector4{1.0, 2.0, 3.0, 4.0}
	ptr := unsafe.Pointer(&v)
	fmt.Printf("Vector: %+v\n", v)
	fmt.Printf("Pointer: %p\n", ptr)
	fmt.Printf("Size: %d bytes (alineado para SIMD)\n", unsafe.Sizeof(v))
}

// ============================================
// FUNCIONES
// ============================================

// AddAsm simula una función assembly
// En producción, estaría en un archivo .s
func AddAsm(a, b int64) int64 {
	return a + b
}

// MultiplyAsm simula una función assembly
func MultiplyAsm(a, b int64) int64 {
	return a * b
}

// PopcountGo cuenta bits en 1 (versión Go)
func PopcountGo(n uint64) int {
	count := 0
	for n != 0 {
		count += int(n & 1)
		n >>= 1
	}
	return count
}

/*
ARCHIVOS DE EJEMPLO ASSEMBLY:

=== add_amd64.s ===

#include "textflag.h"

// func AddAsm(a, b int64) int64
TEXT ·AddAsm(SB), NOSPLIT, $0-24
    MOVQ a+0(FP), AX
    ADDQ b+8(FP), AX
    MOVQ AX, ret+16(FP)
    RET

=== multiply_amd64.s ===

#include "textflag.h"

// func MultiplyAsm(a, b int64) int64
TEXT ·MultiplyAsm(SB), NOSPLIT, $0-24
    MOVQ a+0(FP), AX
    IMULQ b+8(FP), AX
    MOVQ AX, ret+16(FP)
    RET

=== popcount_amd64.s ===

#include "textflag.h"

// func PopcountAsm(n uint64) int
// Usa instrucción POPCNT (requiere CPU con SSE4.2)
TEXT ·PopcountAsm(SB), NOSPLIT, $0-16
    MOVQ n+0(FP), AX
    POPCNTQ AX, AX
    MOVQ AX, ret+8(FP)
    RET

=== memcpy_amd64.s (optimizado SIMD) ===

#include "textflag.h"

// func MemcpyAsm(dst, src []byte, n int)
TEXT ·MemcpyAsm(SB), NOSPLIT, $0
    MOVQ dst_data+0(FP), DI    // destination
    MOVQ src_data+24(FP), SI   // source
    MOVQ n+48(FP), CX          // count

    // Si n >= 64, usar AVX
    CMPQ CX, $64
    JL small

avx_loop:
    CMPQ CX, $64
    JL sse_loop
    VMOVDQU (SI), Y0           // load 32 bytes
    VMOVDQU 32(SI), Y1         // load next 32 bytes
    VMOVDQU Y0, (DI)           // store 32 bytes
    VMOVDQU Y1, 32(DI)         // store next 32 bytes
    ADDQ $64, SI
    ADDQ $64, DI
    SUBQ $64, CX
    JMP avx_loop

sse_loop:
    CMPQ CX, $16
    JL small
    MOVDQU (SI), X0
    MOVDQU X0, (DI)
    ADDQ $16, SI
    ADDQ $16, DI
    SUBQ $16, CX
    JMP sse_loop

small:
    // Copiar byte por byte
    CMPQ CX, $0
    JE done
    MOVB (SI), AX
    MOVB AX, (DI)
    INCQ SI
    INCQ DI
    DECQ CX
    JMP small

done:
    VZEROUPPER                 // limpiar estado AVX
    RET

=== memset_amd64.s ===

#include "textflag.h"

// func MemsetAsm(dst []byte, val byte, n int)
TEXT ·MemsetAsm(SB), NOSPLIT, $0
    MOVQ dst_data+0(FP), DI
    MOVBQZX val+24(FP), AX     // value (zero-extended)
    MOVQ n+32(FP), CX

    // Expandir byte a 8 bytes
    MOVQ $0x0101010101010101, DX
    IMULQ AX, DX               // DX = val repeated 8 times

    // Si n >= 32, usar SIMD
    CMPQ CX, $32
    JL small_set

    // Crear vector con valor repetido
    MOVQ DX, X0
    PUNPCKLQDQ X0, X0          // X0 = 16 bytes del mismo valor

simd_set:
    CMPQ CX, $16
    JL small_set
    MOVDQU X0, (DI)
    ADDQ $16, DI
    SUBQ $16, CX
    JMP simd_set

small_set:
    CMPQ CX, $8
    JL tiny_set
    MOVQ DX, (DI)
    ADDQ $8, DI
    SUBQ $8, CX
    JMP small_set

tiny_set:
    CMPQ CX, $0
    JE set_done
    MOVB AX, (DI)
    INCQ DI
    DECQ CX
    JMP tiny_set

set_done:
    RET

=== dot_product_amd64.s (SSE) ===

#include "textflag.h"

// func DotProductAsm(a, b []float32) float32
TEXT ·DotProductAsm(SB), NOSPLIT, $0-52
    MOVQ a_data+0(FP), SI
    MOVQ a_len+8(FP), CX
    MOVQ b_data+24(FP), DI

    XORPS X0, X0               // acumulador = 0

    CMPQ CX, $4
    JL scalar_dot

simd_dot:
    CMPQ CX, $4
    JL scalar_dot
    MOVUPS (SI), X1            // cargar 4 floats de a
    MOVUPS (DI), X2            // cargar 4 floats de b
    MULPS X2, X1               // multiplicar elemento por elemento
    ADDPS X1, X0               // acumular
    ADDQ $16, SI
    ADDQ $16, DI
    SUBQ $4, CX
    JMP simd_dot

scalar_dot:
    CMPQ CX, $0
    JE finish_dot
    MOVSS (SI), X1
    MULSS (DI), X1
    ADDSS X1, X0
    ADDQ $4, SI
    ADDQ $4, DI
    DECQ CX
    JMP scalar_dot

finish_dot:
    // Horizontal add: X0 = {a, b, c, d} → a+b+c+d
    MOVAPS X0, X1
    SHUFPS $0x4E, X0, X1       // X1 = {c, d, a, b}
    ADDPS X1, X0               // X0 = {a+c, b+d, ...}
    MOVAPS X0, X1
    SHUFPS $0xB1, X0, X1       // X1 = {b+d, a+c, ...}
    ADDSS X1, X0               // X0[0] = a+b+c+d

    MOVSS X0, ret+48(FP)
    RET

RESUMEN DE INSTRUCCIONES:

CATEGORÍA          | INSTRUCCIONES PRINCIPALES
-------------------|------------------------------------------
Aritmética         | ADD, SUB, IMUL, IDIV, INC, DEC, NEG
Lógica             | AND, OR, XOR, NOT, SHL, SHR, SAR, ROL
Comparación        | CMP, TEST
Saltos             | JMP, JEQ, JNE, JL, JG, JZ, JNZ, CALL, RET
Memoria            | MOV, LEA, PUSH, POP, XCHG
Bits               | BSF, BSR, POPCNT, LZCNT, TZCNT
Atómicas           | LOCK prefix, XADD, CMPXCHG, XCHG
SSE (128-bit)      | MOVUPS, ADDPS, MULPS, SHUFPS, PCMPEQD
AVX (256-bit)      | VMOVDQU, VADDPS, VMULPS, VFMADD213PS
AVX-512 (512-bit)  | VADDPS (Z regs), VSCALEFPS
Crypto             | AESENC, AESDEC, SHA256RNDS2, PCLMULQDQ
Condicionales      | CMOVQEQ, CMOVQNE, CMOVQLT, CMOVQGT

CUÁNDO USAR ASSEMBLY:
- Instrucciones específicas (SIMD, crypto, atomic)
- Código crítico verificado con benchmarks
- Acceso a hardware especial
- RARAMENTE - el compilador Go es muy bueno

HERRAMIENTAS:
- go build -gcflags="-S" → ver assembly generado
- go tool objdump → desensamblar binario
- go build -race → detectar races
- go test -bench → benchmarks

RECURSOS:
- Go Assembly Guide: https://go.dev/doc/asm
- Plan 9 Assembly: http://doc.cat-v.org/plan_9/4th_edition/papers/asm
- x86-64 Reference: https://www.felixcloutier.com/x86/
- Intel Intrinsics Guide: https://www.intel.com/content/www/us/en/docs/intrinsics-guide
*/
