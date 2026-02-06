// Package main - Chapter 015: Slices, Maps, and Cmp Packages
package main

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
)

func main() {
	fmt.Println("=== SLICES, MAPS, AND CMP PACKAGES ===")

	fmt.Println(`
These packages (Go 1.21+) provide generic utility functions that replace
many common hand-written loops and patterns from earlier Go code.
They are type-safe, efficient, and idiomatic modern Go.`)

	// ============================================
	// slices: SEARCHING
	// ============================================
	fmt.Println(`
--- slices: Searching ---`)

	fruits := []string{"apple", "banana", "cherry", "date", "elderberry"}
	fmt.Println("\n  fruits:", fruits)

	os.Stdout.WriteString(fmt.Sprintf("  slices.Contains(fruits, \"cherry\"):     %t\n", slices.Contains(fruits, "cherry")))
	os.Stdout.WriteString(fmt.Sprintf("  slices.Contains(fruits, \"fig\"):         %t\n", slices.Contains(fruits, "fig")))

	os.Stdout.WriteString(fmt.Sprintf("  slices.Index(fruits, \"banana\"):         %d\n", slices.Index(fruits, "banana")))
	os.Stdout.WriteString(fmt.Sprintf("  slices.Index(fruits, \"fig\"):            %d  (not found)\n", slices.Index(fruits, "fig")))

	idx, found := slices.BinarySearch(fruits, "cherry")
	os.Stdout.WriteString(fmt.Sprintf("  slices.BinarySearch(fruits, \"cherry\"): idx=%d, found=%t\n", idx, found))

	idx, found = slices.BinarySearch(fruits, "coconut")
	os.Stdout.WriteString(fmt.Sprintf("  slices.BinarySearch(fruits, \"coconut\"): idx=%d, found=%t (insertion point)\n", idx, found))

	type Item struct {
		Name  string
		Price float64
	}
	items := []Item{
		{"Widget", 9.99},
		{"Gadget", 24.99},
		{"Doohickey", 4.99},
	}
	idx2 := slices.IndexFunc(items, func(item Item) bool {
		return item.Price > 20
	})
	os.Stdout.WriteString(fmt.Sprintf("\n  slices.IndexFunc(price > 20): idx=%d -> %s\n", idx2, items[idx2].Name))

	hasExpensive := slices.ContainsFunc(items, func(item Item) bool {
		return item.Price > 50
	})
	os.Stdout.WriteString(fmt.Sprintf("  slices.ContainsFunc(price > 50): %t\n", hasExpensive))

	// ============================================
	// slices: SORTING
	// ============================================
	fmt.Println(`
--- slices: Sorting ---`)

	nums := []int{5, 3, 8, 1, 9, 2, 7, 4, 6}
	slices.Sort(nums)
	fmt.Println("\n  slices.Sort:", nums)
	os.Stdout.WriteString(fmt.Sprintf("  slices.IsSorted: %t\n", slices.IsSorted(nums)))

	words := []string{"Banana", "apple", "Cherry", "date"}
	slices.SortFunc(words, func(a, b string) int {
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	fmt.Println("\n  slices.SortFunc (case-insensitive):", words)

	type Student struct {
		Name  string
		Grade int
	}
	students := []Student{
		{"Alice", 90}, {"Bob", 85}, {"Carol", 90}, {"Dave", 85},
	}
	slices.SortStableFunc(students, func(a, b Student) int {
		return cmp.Compare(b.Grade, a.Grade)
	})
	fmt.Println("\n  slices.SortStableFunc (by grade desc, stable):")
	for _, s := range students {
		os.Stdout.WriteString(fmt.Sprintf("    %-6s grade=%d\n", s.Name, s.Grade))
	}

	data := []int{3, 1, 4, 1, 5, 9, 2, 6}
	os.Stdout.WriteString(fmt.Sprintf("\n  slices.Min(%v): %d\n", data, slices.Min(data)))
	os.Stdout.WriteString(fmt.Sprintf("  slices.Max(%v): %d\n", data, slices.Max(data)))

	minItem := slices.MinFunc(items, func(a, b Item) int {
		return cmp.Compare(a.Price, b.Price)
	})
	os.Stdout.WriteString(fmt.Sprintf("  slices.MinFunc(by price): %s ($%.2f)\n", minItem.Name, minItem.Price))

	// ============================================
	// slices: COMPARISON
	// ============================================
	fmt.Println(`
--- slices: Comparison ---`)

	a := []int{1, 2, 3}
	b := []int{1, 2, 3}
	c := []int{1, 2, 4}
	d := []int{1, 2}

	os.Stdout.WriteString(fmt.Sprintf("\n  slices.Equal([1,2,3], [1,2,3]): %t\n", slices.Equal(a, b)))
	os.Stdout.WriteString(fmt.Sprintf("  slices.Equal([1,2,3], [1,2,4]): %t\n", slices.Equal(a, c)))
	os.Stdout.WriteString(fmt.Sprintf("  slices.Equal([1,2,3], [1,2]):   %t\n", slices.Equal(a, d)))

	result := slices.Compare(a, c)
	os.Stdout.WriteString(fmt.Sprintf("  slices.Compare([1,2,3], [1,2,4]): %d  (first is less)\n", result))

	result = slices.Compare(c, a)
	os.Stdout.WriteString(fmt.Sprintf("  slices.Compare([1,2,4], [1,2,3]): %d  (first is greater)\n", result))

	items1 := []Item{{"A", 1.0}, {"B", 2.0}}
	items2 := []Item{{"A", 1.0}, {"B", 2.0}}
	eq := slices.EqualFunc(items1, items2, func(x, y Item) bool {
		return x.Name == y.Name && x.Price == y.Price
	})
	os.Stdout.WriteString(fmt.Sprintf("\n  slices.EqualFunc (struct comparison): %t\n", eq))

	// ============================================
	// slices: TRANSFORMATION
	// ============================================
	fmt.Println(`
--- slices: Transformation ---`)

	fmt.Println("\n  slices.Clone:")
	orig := []int{1, 2, 3, 4, 5}
	clone := slices.Clone(orig)
	clone[0] = 999
	fmt.Println("    original:", orig)
	fmt.Println("    clone:   ", clone)

	fmt.Println("\n  slices.Compact (removes consecutive duplicates):")
	dupes := []int{1, 1, 2, 3, 3, 3, 2, 2, 1}
	compact := slices.Compact(slices.Clone(dupes))
	os.Stdout.WriteString(fmt.Sprintf("    %v -> %v\n", dupes, compact))

	fmt.Println("    For full dedup, sort first:")
	sorted := slices.Clone(dupes)
	slices.Sort(sorted)
	unique := slices.Compact(sorted)
	os.Stdout.WriteString(fmt.Sprintf("    sort + compact: %v -> %v\n", dupes, unique))

	fmt.Println("\n  slices.CompactFunc (custom equality):")
	names := []string{"Alice", "alice", "ALICE", "Bob", "bob"}
	compactNames := slices.CompactFunc(slices.Clone(names), func(a, b string) bool {
		return strings.EqualFold(a, b)
	})
	os.Stdout.WriteString(fmt.Sprintf("    %v -> %v\n", names, compactNames))

	fmt.Println("\n  slices.Clip (reduces capacity to length):")
	big := make([]int, 3, 100)
	big[0], big[1], big[2] = 1, 2, 3
	os.Stdout.WriteString(fmt.Sprintf("    Before: len=%d, cap=%d\n", len(big), cap(big)))
	clipped := slices.Clip(big)
	os.Stdout.WriteString(fmt.Sprintf("    After:  len=%d, cap=%d\n", len(clipped), cap(clipped)))

	fmt.Println("\n  slices.Grow (increases capacity without changing length):")
	small := []int{1, 2, 3}
	os.Stdout.WriteString(fmt.Sprintf("    Before: len=%d, cap=%d\n", len(small), cap(small)))
	grown := slices.Grow(small, 100)
	os.Stdout.WriteString(fmt.Sprintf("    After Grow(100): len=%d, cap=%d\n", len(grown), cap(grown)))

	fmt.Println("\n  slices.Reverse:")
	rev := []int{1, 2, 3, 4, 5}
	slices.Reverse(rev)
	fmt.Println("    Reversed:", rev)

	fmt.Println("\n  slices.Concat (Go 1.22+):")
	s1 := []int{1, 2, 3}
	s2 := []int{4, 5, 6}
	s3 := []int{7, 8, 9}
	combined := slices.Concat(s1, s2, s3)
	fmt.Println("    Concat:", combined)

	// ============================================
	// slices: DELETE AND INSERT
	// ============================================
	fmt.Println(`
--- slices: Delete and Insert ---`)

	fmt.Println("\n  slices.Delete(s, i, j) - removes elements s[i:j]:")
	del := []int{0, 1, 2, 3, 4, 5}
	del = slices.Delete(del, 2, 4)
	os.Stdout.WriteString(fmt.Sprintf("    Delete [0,1,2,3,4,5] at [2:4] -> %v\n", del))

	fmt.Println("\n  slices.DeleteFunc - removes elements matching predicate:")
	mixed := []int{1, -2, 3, -4, 5, -6}
	positives := slices.DeleteFunc(slices.Clone(mixed), func(n int) bool {
		return n < 0
	})
	os.Stdout.WriteString(fmt.Sprintf("    Remove negatives from %v -> %v\n", mixed, positives))

	fmt.Println("\n  slices.Insert(s, i, vals...) - inserts values at index:")
	ins := []int{1, 2, 5, 6}
	ins = slices.Insert(ins, 2, 3, 4)
	fmt.Println("    Insert 3,4 at index 2 in [1,2,5,6]:", ins)

	fmt.Println("\n  slices.Replace(s, i, j, vals...) - replaces s[i:j] with vals:")
	rep := []int{1, 2, 3, 4, 5}
	rep = slices.Replace(rep, 1, 4, 20, 30)
	fmt.Println("    Replace [1:4] with 20,30 in [1,2,3,4,5]:", rep)

	// ============================================
	// slices: COLLECT AND ALL/VALUES (Go 1.23+ iterators)
	// ============================================
	fmt.Println(`
--- slices: Iterator Functions (Go 1.23+) ---

Go 1.23 introduced iterator support. The slices package integrates
with iterators via slices.All, slices.Values, slices.Collect, etc.`)

	src := []string{"hello", "world", "go"}

	fmt.Println("\n  slices.All (index-value iterator):")
	for i, v := range slices.All(src) {
		os.Stdout.WriteString(fmt.Sprintf("    [%d] %s\n", i, v))
	}

	fmt.Println("\n  slices.Values (value-only iterator):")
	os.Stdout.WriteString("    ")
	for v := range slices.Values(src) {
		os.Stdout.WriteString(v + " ")
	}
	fmt.Println()

	fmt.Println("\n  slices.Backward (reverse iterator):")
	os.Stdout.WriteString("    ")
	for _, v := range slices.Backward(src) {
		os.Stdout.WriteString(v + " ")
	}
	fmt.Println()

	fmt.Println("\n  slices.Collect (iterator -> slice):")
	collected := slices.Collect(slices.Values(src))
	fmt.Println("    Collected:", collected)

	fmt.Println("\n  slices.Chunk (split into chunks):")
	longList := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for chunk := range slices.Chunk(longList, 3) {
		os.Stdout.WriteString(fmt.Sprintf("    %v\n", chunk))
	}

	// ============================================
	// maps: KEYS AND VALUES
	// ============================================
	fmt.Println(`
--- maps: Keys and Values ---`)

	scores := map[string]int{
		"alice": 95,
		"bob":   87,
		"carol": 92,
		"dave":  88,
	}
	fmt.Println("\n  scores:", scores)

	keys := slices.Sorted(maps.Keys(scores))
	fmt.Println("  maps.Keys (sorted):", keys)

	vals := slices.Collect(maps.Values(scores))
	slices.Sort(vals)
	fmt.Println("  maps.Values (sorted):", vals)

	// ============================================
	// maps: COMPARISON AND CLONING
	// ============================================
	fmt.Println(`
--- maps: Comparison and Cloning ---`)

	m1 := map[string]int{"a": 1, "b": 2, "c": 3}
	m2 := map[string]int{"a": 1, "b": 2, "c": 3}
	m3 := map[string]int{"a": 1, "b": 2, "c": 4}

	os.Stdout.WriteString(fmt.Sprintf("\n  maps.Equal(m1, m2): %t\n", maps.Equal(m1, m2)))
	os.Stdout.WriteString(fmt.Sprintf("  maps.Equal(m1, m3): %t\n", maps.Equal(m1, m3)))

	type Score struct {
		Value int
		Curve bool
	}
	sm1 := map[string]Score{"a": {90, true}, "b": {85, false}}
	sm2 := map[string]Score{"a": {90, true}, "b": {85, false}}
	eqFunc := maps.EqualFunc(sm1, sm2, func(a, b Score) bool {
		return a.Value == b.Value && a.Curve == b.Curve
	})
	os.Stdout.WriteString(fmt.Sprintf("  maps.EqualFunc (struct values): %t\n", eqFunc))

	fmt.Println("\n  maps.Clone:")
	original := map[string]int{"x": 1, "y": 2, "z": 3}
	cloneMap := maps.Clone(original)
	cloneMap["w"] = 4
	fmt.Println("    original:", original)
	fmt.Println("    clone:   ", cloneMap)

	// ============================================
	// maps: DELETE AND COPY
	// ============================================
	fmt.Println(`
--- maps: Delete and Copy ---`)

	fmt.Println("\n  maps.DeleteFunc (remove entries matching predicate):")
	inventory := map[string]int{
		"apples":  5,
		"bananas": 0,
		"oranges": 12,
		"grapes":  0,
	}
	fmt.Println("    Before:", inventory)
	maps.DeleteFunc(inventory, func(k string, v int) bool {
		return v == 0
	})
	fmt.Println("    After removing zero-stock:", inventory)

	fmt.Println("\n  maps.Copy (merge src into dst):")
	defaults := map[string]string{
		"host":    "localhost",
		"port":    "8080",
		"timeout": "30s",
	}
	overrides := map[string]string{
		"port": "9090",
		"debug": "true",
	}
	config := maps.Clone(defaults)
	maps.Copy(config, overrides)
	fmt.Println("    defaults: ", defaults)
	fmt.Println("    overrides:", overrides)
	fmt.Println("    merged:   ", config)

	// ============================================
	// maps: ITERATOR FUNCTIONS (Go 1.23+)
	// ============================================
	fmt.Println(`
--- maps: Iterator Functions (Go 1.23+) ---

maps.All returns an iterator over key-value pairs.
maps.Keys and maps.Values return iterators (not slices).
Use slices.Collect to materialize them.`)

	settings := map[string]string{
		"theme": "dark",
		"lang":  "en",
		"font":  "mono",
	}

	fmt.Println("\n  maps.All:")
	for k, v := range maps.All(settings) {
		os.Stdout.WriteString(fmt.Sprintf("    %s = %s\n", k, v))
	}

	fmt.Println("\n  Sorted iteration (maps.Keys + slices.Sorted):")
	for _, k := range slices.Sorted(maps.Keys(settings)) {
		os.Stdout.WriteString(fmt.Sprintf("    %s = %s\n", k, settings[k]))
	}

	fmt.Println("\n  maps.Insert (copy from iterator):")
	base := map[string]int{"a": 1, "b": 2}
	extra := map[string]int{"c": 3, "d": 4}
	maps.Insert(base, maps.All(extra))
	fmt.Println("    base after Insert:", base)

	fmt.Println("\n  maps.Collect (iterator -> map):")
	rebuilt := maps.Collect(maps.All(settings))
	fmt.Println("    Collected:", rebuilt)

	// ============================================
	// cmp: COMPARISON UTILITIES
	// ============================================
	fmt.Println(`
--- cmp: Comparison Utilities ---

The cmp package provides generic comparison functions for ordered types.
These are the building blocks for sorting and ordering operations.`)

	fmt.Println("\n  cmp.Compare (returns -1, 0, or +1):")
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Compare(1, 2):       %d\n", cmp.Compare(1, 2)))
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Compare(2, 2):       %d\n", cmp.Compare(2, 2)))
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Compare(3, 2):       %d\n", cmp.Compare(3, 2)))
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Compare(\"a\", \"b\"):   %d\n", cmp.Compare("a", "b")))

	fmt.Println("\n  cmp.Less:")
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Less(1, 2): %t\n", cmp.Less(1, 2)))
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Less(2, 2): %t\n", cmp.Less(2, 2)))
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Less(3, 2): %t\n", cmp.Less(3, 2)))

	fmt.Println("\n  cmp.Or (returns first non-zero value):")
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Or(0, 0, 3, 4):       %d\n", cmp.Or(0, 0, 3, 4)))
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Or(\"\", \"\", \"hello\"): %q\n", cmp.Or("", "", "hello")))
	os.Stdout.WriteString(fmt.Sprintf("    cmp.Or(0, 0, 0):          %d  (all zero, returns zero)\n", cmp.Or(0, 0, 0)))

	fmt.Println(`
  cmp.Or is great for default values:
    port := cmp.Or(os.Getenv("PORT"), "8080")
    timeout := cmp.Or(userTimeout, configTimeout, 30)`)

	fmt.Println("\n  cmp.Ordered constraint:")
	fmt.Println(`    type Ordered interface {
        ~int | ~int8 | ~int16 | ~int32 | ~int64 |
        ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
        ~float32 | ~float64 |
        ~string
    }`)
	fmt.Println("    Use as a generic constraint for comparable and orderable types.")

	// ============================================
	// PRACTICAL PATTERNS
	// ============================================
	fmt.Println(`
--- Practical Patterns ---`)

	fmt.Println("\n  Pattern: Multi-field struct sorting with cmp.Compare:")
	type Employee struct {
		Dept   string
		Name   string
		Salary int
	}
	emps := []Employee{
		{"Engineering", "Carol", 120000},
		{"Marketing", "Alice", 90000},
		{"Engineering", "Bob", 110000},
		{"Marketing", "Dave", 95000},
		{"Engineering", "Alice", 115000},
	}
	slices.SortFunc(emps, func(a, b Employee) int {
		if c := cmp.Compare(a.Dept, b.Dept); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Name, b.Name); c != 0 {
			return c
		}
		return cmp.Compare(a.Salary, b.Salary)
	})
	fmt.Println("  Sorted by dept, name, salary:")
	for _, e := range emps {
		os.Stdout.WriteString(fmt.Sprintf("    %-12s %-6s $%d\n", e.Dept, e.Name, e.Salary))
	}

	fmt.Println("\n  Pattern: Frequency count with maps + slices:")
	text := []string{"go", "is", "great", "go", "is", "fun", "go"}
	freq := make(map[string]int)
	for _, w := range text {
		freq[w]++
	}
	type WordCount struct {
		Word  string
		Count int
	}
	var wc []WordCount
	for w, c := range freq {
		wc = append(wc, WordCount{w, c})
	}
	slices.SortFunc(wc, func(a, b WordCount) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.Word, b.Word)
	})
	fmt.Println("  Word frequencies (desc):")
	for _, item := range wc {
		os.Stdout.WriteString(fmt.Sprintf("    %-8s %d\n", item.Word, item.Count))
	}

	fmt.Println("\n  Pattern: Set operations with maps:")
	setA := map[string]struct{}{"a": {}, "b": {}, "c": {}, "d": {}}
	setB := map[string]struct{}{"b": {}, "d": {}, "e": {}, "f": {}}

	intersection := make(map[string]struct{})
	for k := range setA {
		if _, ok := setB[k]; ok {
			intersection[k] = struct{}{}
		}
	}
	intKeys := slices.Sorted(maps.Keys(intersection))
	fmt.Println("    A = {a,b,c,d}, B = {b,d,e,f}")
	fmt.Println("    Intersection:", intKeys)

	union := maps.Clone(setA)
	maps.Copy(union, setB)
	unionKeys := slices.Sorted(maps.Keys(union))
	fmt.Println("    Union:       ", unionKeys)

	difference := maps.Clone(setA)
	maps.DeleteFunc(difference, func(k string, _ struct{}) bool {
		_, inB := setB[k]
		return inB
	})
	diffKeys := slices.Sorted(maps.Keys(difference))
	fmt.Println("    A - B:       ", diffKeys)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println(`
--- Summary ---

  slices package:
    Search:    Contains, Index, BinarySearch, ContainsFunc, IndexFunc
    Sort:      Sort, SortFunc, SortStableFunc, IsSorted, Min, Max
    Compare:   Equal, Compare, EqualFunc, CompareFunc
    Transform: Clone, Compact, Clip, Grow, Reverse, Concat
    Modify:    Delete, DeleteFunc, Insert, Replace
    Iterators: All, Values, Backward, Collect, Chunk (Go 1.23+)

  maps package:
    Access:    Keys, Values, All
    Compare:   Equal, EqualFunc
    Copy:      Clone, Copy, Insert, Collect
    Delete:    DeleteFunc

  cmp package:
    Compare, Less, Or, Ordered constraint

  These three packages together replace most hand-written loops for
  searching, sorting, comparing, and transforming collections in Go.`)
}
