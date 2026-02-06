// Package main - Chapter 013: Sort Package
package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
)

func main() {
	fmt.Println("=== SORT PACKAGE ===")

	// ============================================
	// BASIC SORTING
	// ============================================
	fmt.Println(`
--- Basic Sorting ---

The sort package provides sorting for built-in types and custom collections.
sort.Ints, sort.Strings, sort.Float64s sort in ascending order.
These sort IN PLACE (modify the original slice).`)

	ints := []int{5, 3, 8, 1, 9, 2, 7}
	fmt.Println("\n  Before sort.Ints:", ints)
	sort.Ints(ints)
	fmt.Println("  After  sort.Ints:", ints)

	strs := []string{"banana", "apple", "cherry", "date"}
	fmt.Println("\n  Before sort.Strings:", strs)
	sort.Strings(strs)
	fmt.Println("  After  sort.Strings:", strs)

	floats := []float64{3.14, 1.41, 2.72, 0.58}
	fmt.Println("\n  Before sort.Float64s:", floats)
	sort.Float64s(floats)
	fmt.Println("  After  sort.Float64s:", floats)

	// ============================================
	// sort.Slice AND sort.SliceStable
	// ============================================
	fmt.Println(`
--- sort.Slice and sort.SliceStable ---

sort.Slice(s, less) sorts any slice using a custom comparator.
sort.SliceStable preserves the original order of equal elements.

"Stable" means: if a == b, their relative order stays the same.`)

	nums := []int{5, 3, 8, 1, 9, 2}
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})
	fmt.Println("\n  sort.Slice ascending:", nums)

	sort.Slice(nums, func(i, j int) bool {
		return nums[i] > nums[j]
	})
	fmt.Println("  sort.Slice descending:", nums)

	type Task struct {
		Name     string
		Priority int
	}

	tasks := []Task{
		{"Write docs", 2},
		{"Fix bug", 1},
		{"Add feature", 2},
		{"Deploy", 1},
		{"Code review", 3},
	}

	fmt.Println("\n  Tasks before stable sort:")
	for _, t := range tasks {
		os.Stdout.WriteString(fmt.Sprintf("    %s (priority %d)\n", t.Name, t.Priority))
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return tasks[i].Priority < tasks[j].Priority
	})

	fmt.Println("\n  Tasks after SliceStable by priority:")
	fmt.Println("  (equal-priority tasks keep original order)")
	for _, t := range tasks {
		os.Stdout.WriteString(fmt.Sprintf("    %s (priority %d)\n", t.Name, t.Priority))
	}

	// ============================================
	// sort.Interface
	// ============================================
	fmt.Println(`
--- sort.Interface ---

For full control, implement the sort.Interface:
  type Interface interface {
      Len() int
      Less(i, j int) bool
      Swap(i, j int)
  }

This allows sorting any collection type.`)

	employees := ByAge{
		{"Alice", 30},
		{"Bob", 25},
		{"Carol", 35},
		{"Dave", 25},
	}

	fmt.Println("\n  Before sort:", employees)
	sort.Sort(employees)
	fmt.Println("  After  sort:", employees)

	sort.Sort(sort.Reverse(employees))
	fmt.Println("  Reversed:  ", employees)

	// ============================================
	// sort.Reverse
	// ============================================
	fmt.Println(`
--- sort.Reverse ---

sort.Reverse wraps a sort.Interface to invert the Less method.
It works with any type that implements sort.Interface.`)

	words := sort.StringSlice{"gamma", "alpha", "delta", "beta"}
	sort.Sort(sort.Reverse(words))
	fmt.Println("\n  Reverse-sorted strings:", []string(words))

	numbers := sort.IntSlice{5, 3, 8, 1, 9, 2}
	sort.Sort(sort.Reverse(numbers))
	fmt.Println("  Reverse-sorted ints:   ", []int(numbers))

	// ============================================
	// SORTING STRUCTS BY MULTIPLE FIELDS
	// ============================================
	fmt.Println(`
--- Sorting by Multiple Fields ---

Sort by primary field, then by secondary field for ties.
This is common for table-like data.`)

	type Student struct {
		Name  string
		Grade int
		Age   int
	}

	students := []Student{
		{"Alice", 90, 20},
		{"Bob", 85, 22},
		{"Carol", 90, 19},
		{"Dave", 85, 21},
		{"Eve", 90, 20},
	}

	sort.SliceStable(students, func(i, j int) bool {
		if students[i].Grade != students[j].Grade {
			return students[i].Grade > students[j].Grade
		}
		if students[i].Age != students[j].Age {
			return students[i].Age < students[j].Age
		}
		return students[i].Name < students[j].Name
	})

	fmt.Println("\n  Sorted by grade (desc), age (asc), name (asc):")
	for _, s := range students {
		os.Stdout.WriteString(fmt.Sprintf("    %-6s grade=%d age=%d\n", s.Name, s.Grade, s.Age))
	}

	// ============================================
	// sort.Search (BINARY SEARCH)
	// ============================================
	fmt.Println(`
--- sort.Search (Binary Search) ---

sort.Search(n, f) returns the smallest index i in [0, n) for which f(i) is true.
The slice MUST be sorted in the order that matches the predicate.
Returns n if no index satisfies the predicate.`)

	sorted := []int{1, 3, 5, 7, 9, 11, 13, 15}
	fmt.Println("\n  sorted:", sorted)

	target := 7
	idx := sort.SearchInts(sorted, target)
	os.Stdout.WriteString(fmt.Sprintf("  SearchInts(%d) -> index %d", target, idx))
	if idx < len(sorted) && sorted[idx] == target {
		fmt.Println(" (found!)")
	} else {
		fmt.Println(" (not found)")
	}

	target = 6
	idx = sort.SearchInts(sorted, target)
	os.Stdout.WriteString(fmt.Sprintf("  SearchInts(%d) -> index %d", target, idx))
	if idx < len(sorted) && sorted[idx] == target {
		fmt.Println(" (found!)")
	} else {
		os.Stdout.WriteString(fmt.Sprintf(" (not found, insertion point = %d)\n", idx))
	}

	idx = sort.Search(len(sorted), func(i int) bool {
		return sorted[i] >= 10
	})
	os.Stdout.WriteString(fmt.Sprintf("  First element >= 10: index %d, value %d\n", idx, sorted[idx]))

	// ============================================
	// sort.SliceIsSorted
	// ============================================
	fmt.Println(`
--- sort.SliceIsSorted ---

Check if a slice is already sorted according to a comparator.`)

	asc := []int{1, 2, 3, 4, 5}
	desc := []int{5, 4, 3, 2, 1}
	mixed := []int{1, 3, 2, 4, 5}

	checkSorted := func(name string, s []int) {
		result := sort.SliceIsSorted(s, func(i, j int) bool { return s[i] < s[j] })
		os.Stdout.WriteString(fmt.Sprintf("  %-8s %v -> sorted ascending: %t\n", name+":", s, result))
	}

	fmt.Println()
	checkSorted("asc", asc)
	checkSorted("desc", desc)
	checkSorted("mixed", mixed)

	// ============================================
	// STABLE vs UNSTABLE SORT
	// ============================================
	fmt.Println(`
--- Stable vs Unstable Sort ---

Unstable (sort.Slice): faster, no guarantee on equal-element order.
  Uses pattern-defeating quicksort (pdqsort).
Stable (sort.SliceStable): preserves relative order of equal elements.
  Uses merge sort variant. Slightly slower but deterministic.

Use stable sort when:
  - Sorting by one field while preserving a previous sort
  - Deterministic output matters (tests, user-facing data)
  - Multiple equal elements should stay in original order`)

	type Item struct {
		Name     string
		Category string
	}

	items := []Item{
		{"Laptop", "Electronics"},
		{"Shirt", "Clothing"},
		{"Phone", "Electronics"},
		{"Pants", "Clothing"},
		{"Tablet", "Electronics"},
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Category < items[j].Category
	})

	fmt.Println("\n  Stable sort by category (original order within category preserved):")
	for _, item := range items {
		os.Stdout.WriteString(fmt.Sprintf("    %-10s %s\n", item.Name, item.Category))
	}

	// ============================================
	// GENERIC SORTING WITH slices PACKAGE
	// ============================================
	fmt.Println(`
--- Generic Sorting with slices Package (Go 1.21+) ---

The slices package provides generic sorting functions that are
cleaner and type-safe compared to the sort package.
  slices.Sort       - sorts ordered types (int, string, float64...)
  slices.SortFunc   - sorts with custom comparison function
  slices.SortStableFunc - stable sort with custom comparison
  slices.IsSorted   - checks if sorted`)

	vals := []int{5, 3, 8, 1, 9, 2}
	slices.Sort(vals)
	fmt.Println("\n  slices.Sort:", vals)

	os.Stdout.WriteString(fmt.Sprintf("  slices.IsSorted: %t\n", slices.IsSorted(vals)))

	names := []string{"Charlie", "alice", "Bob", "dave"}
	slices.SortFunc(names, func(a, b string) int {
		return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
	})
	fmt.Println("\n  slices.SortFunc (case-insensitive):", names)

	type Record struct {
		Name  string
		Score int
	}
	records := []Record{
		{"Alice", 90},
		{"Bob", 85},
		{"Carol", 90},
		{"Dave", 85},
	}

	slices.SortStableFunc(records, func(a, b Record) int {
		return cmp.Compare(b.Score, a.Score)
	})

	fmt.Println("\n  slices.SortStableFunc (by score desc):")
	for _, r := range records {
		os.Stdout.WriteString(fmt.Sprintf("    %-6s score=%d\n", r.Name, r.Score))
	}

	// ============================================
	// BINARY SEARCH WITH slices PACKAGE
	// ============================================
	fmt.Println(`
--- Binary Search with slices Package ---

slices.BinarySearch is the generic replacement for sort.Search.
Returns (index, found) instead of just an index.`)

	haystack := []int{1, 3, 5, 7, 9, 11, 13}
	fmt.Println("\n  haystack:", haystack)

	idx2, found := slices.BinarySearch(haystack, 7)
	os.Stdout.WriteString(fmt.Sprintf("  BinarySearch(7):  index=%d, found=%t\n", idx2, found))

	idx2, found = slices.BinarySearch(haystack, 6)
	os.Stdout.WriteString(fmt.Sprintf("  BinarySearch(6):  index=%d, found=%t (insertion point)\n", idx2, found))

	type Person struct {
		Name string
		Age  int
	}

	people := []Person{
		{"Alice", 25},
		{"Bob", 30},
		{"Carol", 35},
	}

	idx3, found := slices.BinarySearchFunc(people, Person{Age: 30}, func(a, b Person) int {
		return cmp.Compare(a.Age, b.Age)
	})
	os.Stdout.WriteString(fmt.Sprintf("\n  BinarySearchFunc(age=30): index=%d, found=%t -> %s\n", idx3, found, people[idx3].Name))

	// ============================================
	// SORT vs SLICES COMPARISON
	// ============================================
	fmt.Println(`
--- sort vs slices Package Comparison ---

  sort package (classic):         slices package (modern, Go 1.21+):
  -------------------------       -----------------------------------
  sort.Ints(s)                    slices.Sort(s)
  sort.Slice(s, less)             slices.SortFunc(s, cmp)
  sort.SliceStable(s, less)       slices.SortStableFunc(s, cmp)
  sort.SliceIsSorted(s, less)     slices.IsSortedFunc(s, cmp)
  sort.SearchInts(s, v)           slices.BinarySearch(s, v)

  Key differences:
  - slices uses generics: type-safe, no interface boxing
  - slices comparators use cmp function (returns -1/0/+1) vs less (bool)
  - slices.BinarySearch returns (index, found) tuple
  - slices functions are ~15-20% faster due to no interface overhead

  Recommendation: prefer slices package for new code (Go 1.21+).`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println(`
--- Summary ---

  1. sort.Slice/sort.SliceStable: sort any slice with a comparator
  2. sort.Interface: Len/Less/Swap for custom collection types
  3. sort.Reverse wraps an Interface to sort in reverse order
  4. sort.Search does binary search (slice must be sorted first)
  5. Stable sort preserves relative order of equal elements
  6. slices.Sort/SortFunc: modern, generic, faster alternatives
  7. slices.BinarySearch returns (index, found) - cleaner API
  8. Prefer slices package for new Go 1.21+ code`)
}

// Employee is used to demonstrate sort.Interface
type Employee struct {
	Name string
	Age  int
}

// ByAge implements sort.Interface for []Employee
type ByAge []Employee

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
