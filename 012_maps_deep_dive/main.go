// Package main - Chapter 012: Maps Deep Dive
package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

func main() {
	fmt.Println("=== MAPS DEEP DIVE ===")

	// ============================================
	// MAP BASICS AND CREATION
	// ============================================
	fmt.Println(`
--- Map Basics and Creation ---

A map is a hash table providing O(1) average-case lookups, inserts, and deletes.
Keys must be comparable (==, !=): strings, ints, floats, bools, pointers, structs
with all comparable fields. Slices, maps, and functions CANNOT be keys.`)

	var nilMap map[string]int
	emptyMap := map[string]int{}
	madeMap := make(map[string]int)
	sizedMap := make(map[string]int, 100)

	os.Stdout.WriteString(fmt.Sprintf("\n  var nilMap map[string]int     -> nil=%t, len=%d\n", nilMap == nil, len(nilMap)))
	os.Stdout.WriteString(fmt.Sprintf("  emptyMap := map[string]int{} -> nil=%t, len=%d\n", emptyMap == nil, len(emptyMap)))
	os.Stdout.WriteString(fmt.Sprintf("  madeMap := make(...)         -> nil=%t, len=%d\n", madeMap == nil, len(madeMap)))
	os.Stdout.WriteString(fmt.Sprintf("  sizedMap := make(..., 100)   -> nil=%t, len=%d\n", sizedMap == nil, len(sizedMap)))

	// ============================================
	// NIL MAP vs EMPTY MAP
	// ============================================
	fmt.Println(`
--- Nil Map vs Empty Map ---

A nil map can be READ from (returns zero value) but CANNOT be written to.
Writing to a nil map causes a runtime panic.
An empty map (or make'd map) can be both read and written.`)

	fmt.Println("\n  Reading from nil map:")
	v := nilMap["key"]
	os.Stdout.WriteString(fmt.Sprintf("    nilMap[\"key\"] = %d (zero value, no panic)\n", v))

	fmt.Println("\n  Writing to nil map would panic:")
	fmt.Println("    nilMap[\"key\"] = 1  // panic: assignment to entry in nil map")
	fmt.Println("\n  Always initialize before writing:")
	fmt.Println("    m := make(map[string]int)  // or map[string]int{}")

	// ============================================
	// THE COMMA-OK IDIOM
	// ============================================
	fmt.Println(`
--- The Comma-OK Idiom ---

Use v, ok := m[key] to distinguish "key absent" from "key has zero value".`)

	scores := map[string]int{"alice": 0, "bob": 95}

	score, ok := scores["alice"]
	os.Stdout.WriteString(fmt.Sprintf("\n  scores[\"alice\"] = %d, ok = %t (present with zero value)\n", score, ok))

	score, ok = scores["charlie"]
	os.Stdout.WriteString(fmt.Sprintf("  scores[\"charlie\"] = %d, ok = %t (absent)\n", score, ok))

	if val, exists := scores["bob"]; exists {
		os.Stdout.WriteString(fmt.Sprintf("  Bob's score: %d\n", val))
	}

	// ============================================
	// MAP OPERATIONS
	// ============================================
	fmt.Println(`
--- Map Operations ---`)

	m := map[string]int{"a": 1, "b": 2, "c": 3}
	fmt.Println("\n  m:", m)

	m["d"] = 4
	fmt.Println("  After m[\"d\"] = 4:", m)

	delete(m, "b")
	fmt.Println("  After delete(m, \"b\"):", m)

	delete(m, "nonexistent")
	fmt.Println("  delete(m, \"nonexistent\"): no panic, no-op")

	os.Stdout.WriteString(fmt.Sprintf("  len(m) = %d\n", len(m)))

	// ============================================
	// ITERATION ORDER IS RANDOM
	// ============================================
	fmt.Println(`
--- Iteration Order Is Random ---

Map iteration order is intentionally randomized by Go's runtime.
This is by design to prevent code from depending on insertion order.
If you need ordered iteration, sort the keys first.`)

	ordered := map[string]int{"banana": 2, "apple": 1, "cherry": 3, "date": 4}

	fmt.Println("\n  Three iterations over the same map:")
	for attempt := 0; attempt < 3; attempt++ {
		os.Stdout.WriteString(fmt.Sprintf("    Attempt %d: ", attempt+1))
		for k := range ordered {
			os.Stdout.WriteString(k + " ")
		}
		fmt.Println()
	}

	fmt.Println("\n  Sorted iteration:")
	keys := make([]string, 0, len(ordered))
	for k := range ordered {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	os.Stdout.WriteString("    ")
	for _, k := range keys {
		os.Stdout.WriteString(fmt.Sprintf("%s=%d ", k, ordered[k]))
	}
	fmt.Println()

	// ============================================
	// HASH TABLE INTERNALS
	// ============================================
	fmt.Println(`
--- Hash Table Internals ---

Go's map is a hash table with buckets.

Classic implementation (before Go 1.24):
  - Each bucket holds 8 key-value pairs
  - Overflow buckets chain when a bucket fills up
  - Load factor triggers growth at ~6.5 elements per bucket
  - Growth doubles the number of buckets
  - Evacuation is incremental (amortized during inserts/deletes)

Swiss Tables (Go 1.24+):
  - New implementation based on Google's Swiss Tables design
  - Uses SIMD-friendly control bytes for faster probing
  - Better cache locality with group-based probing
  - Reduced memory overhead per entry
  - Faster lookups, especially for misses
  - Transparent: same API, just faster under the hood`)

	// ============================================
	// MAP OF POINTERS vs MAP OF VALUES
	// ============================================
	fmt.Println(`
--- Map of Pointers vs Map of Values ---

You CANNOT take the address of a map value:
  &m["key"]  // compile error

This is because map values may move in memory during growth.
Use a map of pointers if you need to modify structs in place.`)

	type User struct {
		Name  string
		Score int
	}

	byValue := map[string]User{
		"alice": {Name: "Alice", Score: 10},
	}
	fmt.Println("\n  Map of values:")
	fmt.Println("   ", byValue)
	u := byValue["alice"]
	u.Score = 20
	byValue["alice"] = u
	fmt.Println("    After modify-copy-reassign:", byValue)

	byPointer := map[string]*User{
		"bob": {Name: "Bob", Score: 10},
	}
	fmt.Println("\n  Map of pointers:")
	fmt.Println("   ", byPointer["bob"])
	byPointer["bob"].Score = 20
	os.Stdout.WriteString(fmt.Sprintf("    After direct modify: %+v\n", byPointer["bob"]))

	// ============================================
	// MAPS ARE REFERENCE TYPES
	// ============================================
	fmt.Println(`
--- Maps Are Reference Types ---

A map variable holds a pointer to the underlying hash table.
Assigning a map to another variable creates a shared reference.
Both variables see the same data.`)

	original := map[string]int{"x": 1, "y": 2}
	alias := original
	alias["z"] = 3
	fmt.Println("\n  original:", original)
	fmt.Println("  alias:   ", alias)
	fmt.Println("  Adding to alias also appears in original!")

	fmt.Println("\n  To make a true copy, iterate and copy elements:")
	clone := make(map[string]int, len(original))
	for k, v := range original {
		clone[k] = v
	}
	clone["w"] = 4
	fmt.Println("  original:", original)
	fmt.Println("  clone:   ", clone)

	// ============================================
	// CONCURRENT ACCESS
	// ============================================
	fmt.Println(`
--- Concurrent Access ---

Maps are NOT safe for concurrent use. Concurrent reads are fine,
but concurrent read+write or write+write causes a fatal runtime error:
  "fatal error: concurrent map read and map write"

Options for concurrent access:
  1. sync.Mutex or sync.RWMutex (most common)
  2. sync.Map (specialized: few writes, many reads, or disjoint key sets)`)

	fmt.Println("\n  Example: sync.RWMutex protected map")
	safeMap := &SafeMap{data: make(map[string]int)}
	safeMap.Set("counter", 0)
	val := safeMap.Get("counter")
	os.Stdout.WriteString(fmt.Sprintf("    safeMap.Get(\"counter\") = %d\n", val))
	safeMap.Set("counter", 42)
	os.Stdout.WriteString(fmt.Sprintf("    safeMap.Get(\"counter\") = %d\n", safeMap.Get("counter")))

	fmt.Println("\n  Example: sync.Map (built-in concurrent map)")
	var sm sync.Map
	sm.Store("key1", "value1")
	sm.Store("key2", "value2")
	if v, ok := sm.Load("key1"); ok {
		os.Stdout.WriteString(fmt.Sprintf("    sm.Load(\"key1\") = %v\n", v))
	}
	sm.Range(func(key, value any) bool {
		os.Stdout.WriteString(fmt.Sprintf("    Range: %v -> %v\n", key, value))
		return true
	})

	// ============================================
	// MAP PATTERNS
	// ============================================
	fmt.Println(`
--- Common Map Patterns ---`)

	fmt.Println("\n  Set (unique elements):")
	set := make(map[string]struct{})
	words := []string{"go", "is", "great", "go", "is", "fun"}
	for _, w := range words {
		set[w] = struct{}{}
	}
	os.Stdout.WriteString(fmt.Sprintf("    Words: %v\n", words))
	os.Stdout.WriteString("    Unique: ")
	for w := range set {
		os.Stdout.WriteString(w + " ")
	}
	fmt.Println()
	fmt.Println("    Use struct{} as value: zero memory per entry")

	fmt.Println("\n  Counting / Frequency:")
	text := []string{"the", "cat", "sat", "on", "the", "mat", "the"}
	freq := make(map[string]int)
	for _, w := range text {
		freq[w]++
	}
	os.Stdout.WriteString(fmt.Sprintf("    Frequency: %v\n", freq))

	fmt.Println("\n  Grouping:")
	people := []struct {
		Name string
		Age  int
	}{
		{"Alice", 30}, {"Bob", 25}, {"Carol", 30}, {"Dave", 25},
	}
	grouped := make(map[int][]string)
	for _, p := range people {
		grouped[p.Age] = append(grouped[p.Age], p.Name)
	}
	os.Stdout.WriteString(fmt.Sprintf("    Grouped by age: %v\n", grouped))

	fmt.Println("\n  Default value with ok check:")
	defaults := map[string]string{}
	key := "theme"
	if val, ok := defaults[key]; ok {
		os.Stdout.WriteString(fmt.Sprintf("    %s = %s\n", key, val))
	} else {
		os.Stdout.WriteString(fmt.Sprintf("    %s not found, using default\n", key))
	}

	// ============================================
	// MAP WITH STRUCT KEYS
	// ============================================
	fmt.Println(`
--- Map with Struct Keys ---

Any comparable struct can be a map key. This enables multi-dimensional lookups.`)

	type Point struct {
		X, Y int
	}

	grid := map[Point]string{
		{0, 0}: "origin",
		{1, 0}: "right",
		{0, 1}: "up",
	}
	os.Stdout.WriteString(fmt.Sprintf("\n    grid[Point{0,0}] = %q\n", grid[Point{0, 0}]))
	os.Stdout.WriteString(fmt.Sprintf("    grid[Point{1,0}] = %q\n", grid[Point{1, 0}]))

	type CacheKey struct {
		UserID int
		Page   string
	}
	cache := map[CacheKey]string{
		{UserID: 1, Page: "/home"}:  "cached-home",
		{UserID: 1, Page: "/about"}: "cached-about",
	}
	os.Stdout.WriteString(fmt.Sprintf("    cache[{1, \"/home\"}] = %q\n", cache[CacheKey{1, "/home"}]))

	// ============================================
	// PERFORMANCE TIPS
	// ============================================
	fmt.Println(`
--- Performance Tips ---

  1. Pre-allocate: make(map[K]V, expectedSize) to avoid rehashing
  2. Use int keys over string keys when possible (faster hash)
  3. Small maps (<8 entries) may be faster as sorted slices
  4. For read-heavy concurrent access, consider sync.Map
  5. Maps never shrink: delete() frees values but not bucket memory
     To reclaim memory, create a new map and copy live entries
  6. Avoid pointers as keys unless necessary (slower hash, GC pressure)`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println(`
--- Summary ---

  1. Maps are hash tables: O(1) avg lookup, insert, delete
  2. nil map reads return zero values; writes panic
  3. Use comma-ok to distinguish absent keys from zero values
  4. Iteration order is randomized by design
  5. Maps are reference types: assignment shares the data
  6. NOT safe for concurrent writes: use sync.Mutex or sync.Map
  7. Swiss Tables (Go 1.24+) bring faster lookups transparently
  8. Common patterns: sets, counting, grouping, multi-key lookups`)
}

// SafeMap demonstrates a mutex-protected map
type SafeMap struct {
	mu   sync.RWMutex
	data map[string]int
}

func (sm *SafeMap) Get(key string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.data[key]
}

func (sm *SafeMap) Set(key string, value int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
}
