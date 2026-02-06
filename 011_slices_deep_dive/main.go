// Package main - Chapter 011: Slices Deep Dive
package main

import (
	"fmt"
	"os"
	"unsafe"
)

func main() {
	fmt.Println("=== SLICES DEEP DIVE ===")

	// ============================================
	// SLICE HEADER INTERNALS
	// ============================================
	fmt.Println(`
--- Slice Header Internals ---

A slice is a lightweight descriptor (header) with three fields:
  - Pointer: points to the first element of the underlying array
  - Length:  number of elements the slice currently holds
  - Cap:    number of elements in the underlying array from the slice start

The slice header is 24 bytes on 64-bit systems (3 x 8 bytes).`)

	s := []int{10, 20, 30, 40, 50}
	fmt.Println("\ns := []int{10, 20, 30, 40, 50}")
	os.Stdout.WriteString(fmt.Sprintf("  len=%d, cap=%d\n", len(s), cap(s)))
	os.Stdout.WriteString(fmt.Sprintf("  Slice header size: %d bytes\n", unsafe.Sizeof(s)))
	os.Stdout.WriteString(fmt.Sprintf("  Data pointer: %p\n", &s[0]))

	// ============================================
	// NIL vs EMPTY SLICES
	// ============================================
	fmt.Println(`
--- Nil vs Empty Slices ---

A nil slice has no underlying array (pointer is nil).
An empty slice has a backing array but zero length.
Both have len=0 and cap=0, but they differ in nil checks and JSON marshaling.`)

	var nilSlice []int
	emptySlice := []int{}
	makeSlice := make([]int, 0)

	os.Stdout.WriteString(fmt.Sprintf("\n  var nilSlice []int       -> nil=%t, len=%d, cap=%d\n", nilSlice == nil, len(nilSlice), cap(nilSlice)))
	os.Stdout.WriteString(fmt.Sprintf("  emptySlice := []int{}    -> nil=%t, len=%d, cap=%d\n", emptySlice == nil, len(emptySlice), cap(emptySlice)))
	os.Stdout.WriteString(fmt.Sprintf("  makeSlice := make([]int,0) -> nil=%t, len=%d, cap=%d\n", makeSlice == nil, len(makeSlice), cap(makeSlice)))

	fmt.Println(`
  Key difference: json.Marshal(nil) -> "null", json.Marshal([]int{}) -> "[]"
  Best practice: use var s []T for declarations (nil is fine for most uses).
  append() works on nil slices: append(nilSlice, 1) works perfectly.`)

	nilSlice = append(nilSlice, 1, 2, 3)
	os.Stdout.WriteString(fmt.Sprintf("\n  After append(nilSlice, 1, 2, 3): %v, len=%d, cap=%d\n", nilSlice, len(nilSlice), cap(nilSlice)))

	// ============================================
	// CAPACITY GROWTH ALGORITHM
	// ============================================
	fmt.Println(`
--- Capacity Growth Algorithm ---

When append() needs more space, Go allocates a new, larger array.
The growth strategy (as of Go 1.18+) is:
  - For cap < 256: double the capacity
  - For cap >= 256: grow by ~25% + 192 (smoothing formula)
This avoids excessive memory waste for large slices.`)

	fmt.Println("\nDemonstrating growth pattern:")
	var grow []int
	prevCap := 0
	for i := 0; i < 2048; i++ {
		grow = append(grow, i)
		if cap(grow) != prevCap {
			os.Stdout.WriteString(fmt.Sprintf("  len=%-5d cap=%-5d (grew from %d)\n", len(grow), cap(grow), prevCap))
			prevCap = cap(grow)
		}
	}

	// ============================================
	// APPEND MECHANICS
	// ============================================
	fmt.Println(`
--- Append Mechanics ---

append() returns a new slice header. If capacity is sufficient,
the same backing array is reused. If not, a new array is allocated
and all elements are copied.

CRITICAL: always capture the return value of append!
  s = append(s, elem)  // correct
  append(s, elem)      // BUG: result is discarded`)

	original := make([]int, 3, 5)
	copy(original, []int{1, 2, 3})
	fmt.Println("\noriginal := make([]int, 3, 5) with values [1, 2, 3]")
	os.Stdout.WriteString(fmt.Sprintf("  original: %v, len=%d, cap=%d, ptr=%p\n", original, len(original), cap(original), &original[0]))

	withRoom := append(original, 4)
	os.Stdout.WriteString(fmt.Sprintf("  append(original, 4): %v, ptr=%p (SAME array - had room)\n", withRoom, &withRoom[0]))

	full := []int{1, 2, 3}
	os.Stdout.WriteString(fmt.Sprintf("\n  full: %v, len=%d, cap=%d, ptr=%p\n", full, len(full), cap(full), &full[0]))
	grown := append(full, 4)
	os.Stdout.WriteString(fmt.Sprintf("  append(full, 4): %v, ptr=%p (NEW array - no room)\n", grown, &grown[0]))

	// ============================================
	// COPY SEMANTICS
	// ============================================
	fmt.Println(`
--- Copy Semantics ---

copy(dst, src) copies min(len(dst), len(src)) elements.
It returns the number of elements copied.
dst must be pre-allocated (copy doesn't grow the slice).`)

	src := []int{1, 2, 3, 4, 5}
	dst := make([]int, 3)
	n := copy(dst, src)
	os.Stdout.WriteString(fmt.Sprintf("\n  src: %v, dst (len 3): copied %d elements -> %v\n", src, n, dst))

	dst2 := make([]int, 7)
	n2 := copy(dst2, src)
	os.Stdout.WriteString(fmt.Sprintf("  src: %v, dst (len 7): copied %d elements -> %v\n", src, n2, dst2))

	overlap := []int{1, 2, 3, 4, 5}
	copy(overlap[1:], overlap[:4])
	os.Stdout.WriteString(fmt.Sprintf("  Overlapping copy (shift right): %v\n", overlap))

	// ============================================
	// SUBSLICE GOTCHAS - SHARED BACKING ARRAY
	// ============================================
	fmt.Println(`
--- Subslice Gotchas: Shared Backing Array ---

A subslice shares the same underlying array as the original.
Modifying elements through one affects the other.
This is a very common source of bugs!`)

	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	sub := data[3:6]
	fmt.Println("\n  data:", data)
	fmt.Println("  sub := data[3:6] ->", sub)

	sub[0] = 999
	fmt.Println("\n  After sub[0] = 999:")
	fmt.Println("  data:", data)
	fmt.Println("  sub: ", sub)
	fmt.Println("  data[3] was modified through sub!")

	// ============================================
	// MEMORY LEAKS FROM SUBSLICING
	// ============================================
	fmt.Println(`
--- Memory Leaks from Subslicing ---

If you subslice a large slice and only keep the subslice,
the entire backing array stays in memory (cannot be GC'd).

BAD:  small := bigSlice[5:10]     // keeps entire bigSlice in memory
GOOD: small := append([]int{}, bigSlice[5:10]...)  // copies to new array
GOOD: small := slices.Clone(bigSlice[5:10])         // Go 1.21+`)

	big := make([]int, 1_000_000)
	big[500_000] = 42

	leaky := big[500_000:500_001]
	os.Stdout.WriteString(fmt.Sprintf("\n  leaky := big[500000:500001] -> cap=%d (retains full array!)\n", cap(leaky)))

	safe := make([]int, 1)
	copy(safe, big[500_000:500_001])
	os.Stdout.WriteString(fmt.Sprintf("  safe := copy approach -> cap=%d (independent, small allocation)\n", cap(safe)))

	// ============================================
	// FULL SLICE EXPRESSION [low:high:max]
	// ============================================
	fmt.Println(`
--- Full Slice Expression [low:high:max] ---

The three-index slice expression s[low:high:max] controls capacity.
  - len = high - low
  - cap = max - low
This prevents a subslice from accidentally overwriting elements
beyond its intended range in the shared backing array.`)

	base := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	normal := base[2:5]
	limited := base[2:5:5]

	os.Stdout.WriteString(fmt.Sprintf("\n  base: %v\n", base))
	os.Stdout.WriteString(fmt.Sprintf("  normal := base[2:5]   -> %v, len=%d, cap=%d\n", normal, len(normal), cap(normal)))
	os.Stdout.WriteString(fmt.Sprintf("  limited := base[2:5:5] -> %v, len=%d, cap=%d\n", limited, len(limited), cap(limited)))

	fmt.Println(`
  With normal (cap=8), append reuses base's array:`)
	extended := append(normal, 99)
	os.Stdout.WriteString(fmt.Sprintf("    append(normal, 99): %v\n", extended))
	os.Stdout.WriteString(fmt.Sprintf("    base is now: %v  (base[5] was overwritten!)\n", base))

	base2 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	limited2 := base2[2:5:5]
	extended2 := append(limited2, 99)
	fmt.Println("\n  With limited (cap=3), append allocates new array:")
	os.Stdout.WriteString(fmt.Sprintf("    append(limited, 99): %v\n", extended2))
	os.Stdout.WriteString(fmt.Sprintf("    base2 is still: %v  (safe!)\n", base2))

	// ============================================
	// PRE-ALLOCATION BEST PRACTICES
	// ============================================
	fmt.Println(`
--- Pre-Allocation Best Practices ---

Pre-allocating slices avoids repeated allocations during growth.
Use make([]T, 0, expectedSize) when you know the approximate size.

Common patterns:`)

	fmt.Println("\n  Pattern 1: Known size - use make with length")
	ids := []int{1, 2, 3, 4, 5}
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = fmt.Sprintf("item-%d", id)
	}
	fmt.Println("   ", names)

	fmt.Println("\n  Pattern 2: Known max size - use make with 0 length, N capacity")
	source := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evens := make([]int, 0, len(source))
	for _, v := range source {
		if v%2 == 0 {
			evens = append(evens, v)
		}
	}
	fmt.Println("    Evens:", evens)

	fmt.Println("\n  Pattern 3: Unknown size - start with nil, let append grow")
	fmt.Println("    var result []int  // nil is fine, append handles growth")

	// ============================================
	// SLICE TRICKS AND IDIOMS
	// ============================================
	fmt.Println(`
--- Common Slice Tricks ---`)

	fmt.Println("\n  Delete element at index i (order preserved):")
	del := []int{0, 1, 2, 3, 4, 5}
	i := 2
	del = append(del[:i], del[i+1:]...)
	fmt.Println("    Delete index 2 from [0,1,2,3,4,5]:", del)

	fmt.Println("\n  Delete element at index i (order NOT preserved - faster):")
	del2 := []int{0, 1, 2, 3, 4, 5}
	j := 2
	del2[j] = del2[len(del2)-1]
	del2 = del2[:len(del2)-1]
	fmt.Println("    Delete index 2 from [0,1,2,3,4,5]:", del2)

	fmt.Println("\n  Insert element at index i:")
	ins := []int{0, 1, 3, 4, 5}
	k := 2
	ins = append(ins[:k+1], ins[k:]...)
	ins[k] = 2
	fmt.Println("    Insert 2 at index 2 in [0,1,3,4,5]:", ins)

	fmt.Println("\n  Reverse a slice:")
	rev := []int{1, 2, 3, 4, 5}
	for left, right := 0, len(rev)-1; left < right; left, right = left+1, right-1 {
		rev[left], rev[right] = rev[right], rev[left]
	}
	fmt.Println("    Reversed [1,2,3,4,5]:", rev)

	fmt.Println("\n  Filter in-place (zero allocation):")
	filter := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n3 := 0
	for _, v := range filter {
		if v%3 == 0 {
			filter[n3] = v
			n3++
		}
	}
	filter = filter[:n3]
	fmt.Println("    Multiples of 3 from [1..10]:", filter)

	// ============================================
	// PASSING SLICES TO FUNCTIONS
	// ============================================
	fmt.Println(`
--- Passing Slices to Functions ---

Slices are passed by value, but the VALUE is the header (ptr, len, cap).
So the function gets a copy of the header pointing to the SAME array.
  - Modifying elements: visible to caller (same array)
  - Appending: NOT visible to caller (header copy is local)
To reflect appends, return the slice or use a pointer to slice.`)

	arr := []int{1, 2, 3}
	modifyElements(arr)
	fmt.Println("\n  After modifyElements:", arr)

	arr2 := []int{1, 2, 3}
	appendToSlice(arr2)
	fmt.Println("  After appendToSlice (not visible):", arr2)

	arr3 := []int{1, 2, 3}
	arr3 = appendAndReturn(arr3)
	fmt.Println("  After appendAndReturn:", arr3)

	// ============================================
	// MULTIDIMENSIONAL SLICES
	// ============================================
	fmt.Println(`
--- Multidimensional Slices ---

Go doesn't have built-in 2D slices; use a slice of slices.
Each inner slice can have different lengths (jagged arrays).`)

	matrix := make([][]int, 3)
	for row := range matrix {
		matrix[row] = make([]int, 4)
		for col := range matrix[row] {
			matrix[row][col] = row*4 + col
		}
	}
	fmt.Println("\n  3x4 matrix:")
	for _, row := range matrix {
		os.Stdout.WriteString(fmt.Sprintf("    %v\n", row))
	}

	fmt.Println(`
  For performance-critical code, use a flat slice with manual indexing:
    flat := make([]int, rows*cols)
    elem := flat[row*cols + col]
  This gives better cache locality than slice-of-slices.`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println(`
--- Summary ---

  1. A slice = (pointer, length, capacity) - a view into an array
  2. nil slices work with append, len, cap - prefer var s []T
  3. Growth: 2x for small, ~1.25x for large (Go 1.18+)
  4. Subslices share backing arrays - mutations propagate
  5. Use [low:high:max] to limit capacity and prevent overwrites
  6. Copy data to avoid memory leaks from large backing arrays
  7. Pre-allocate with make([]T, 0, n) when size is known
  8. Pass and return slices to reflect appends in callers`)
}

func modifyElements(s []int) {
	if len(s) > 0 {
		s[0] = 999
	}
}

func appendToSlice(s []int) {
	s = append(s, 4)
	_ = s
}

func appendAndReturn(s []int) []int {
	return append(s, 4)
}
