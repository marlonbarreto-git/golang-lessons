// Package main - Chapter 014: Container Types
package main

import (
	"container/heap"
	"container/list"
	"container/ring"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== CONTAINER TYPES ===")

	// ============================================
	// OVERVIEW
	// ============================================
	fmt.Println(`
--- Overview ---

Go's standard library provides three container types in the container/ package:
  container/heap - Priority queue (binary heap)
  container/list - Doubly linked list
  container/ring - Circular buffer (ring buffer)

These use interface{} (any) for values since they predate generics.
Wrap them with typed accessors for type safety.`)

	// ============================================
	// CONTAINER/LIST - DOUBLY LINKED LIST
	// ============================================
	fmt.Println(`
--- container/list - Doubly Linked List ---

A doubly linked list where each element has Prev and Next pointers.
Operations: PushBack, PushFront, InsertBefore, InsertAfter, Remove,
            MoveToFront, MoveToBack, MoveBefore, MoveAfter.
All operations are O(1) given an element pointer.`)

	l := list.New()

	fmt.Println("\n  Building the list:")
	l.PushBack("second")
	l.PushBack("third")
	l.PushFront("first")
	last := l.PushBack("fourth")
	printList("    After PushBack/PushFront", l)

	l.InsertBefore("between-1-2", l.Front().Next())
	printList("    After InsertBefore", l)

	l.InsertAfter("between-3-4", last.Prev())
	printList("    After InsertAfter", l)

	l.Remove(l.Front().Next())
	printList("    After Remove 2nd elem", l)

	l.MoveToFront(last)
	printList("    After MoveToFront(last)", l)

	l.MoveToBack(l.Front())
	printList("    After MoveToBack(front)", l)

	fmt.Println("\n  Iterating forward:")
	os.Stdout.WriteString("    ")
	for e := l.Front(); e != nil; e = e.Next() {
		os.Stdout.WriteString(fmt.Sprintf("%v ", e.Value))
	}
	fmt.Println()

	fmt.Println("  Iterating backward:")
	os.Stdout.WriteString("    ")
	for e := l.Back(); e != nil; e = e.Prev() {
		os.Stdout.WriteString(fmt.Sprintf("%v ", e.Value))
	}
	fmt.Println()

	os.Stdout.WriteString(fmt.Sprintf("  Length: %d\n", l.Len()))

	// ============================================
	// LIST USE CASES
	// ============================================
	fmt.Println(`
--- List Use Cases ---

  1. LRU Cache: Move accessed items to front, evict from back
  2. Ordered collections with frequent insert/remove at arbitrary positions
  3. Undo/redo stacks
  4. Queue/deque implementations

  When NOT to use list:
  - Random access needed (use slice)
  - Only append/pop from ends (use slice)
  - Cache locality matters (list nodes are scattered in memory)`)

	fmt.Println("  LRU Cache example:")
	lru := NewLRUCache(3)
	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("c", 3)
	fmt.Println("    Added a=1, b=2, c=3")

	if v, ok := lru.Get("a"); ok {
		os.Stdout.WriteString(fmt.Sprintf("    Get(a) = %v (moves to front)\n", v))
	}

	lru.Put("d", 4)
	fmt.Println("    Added d=4 (evicts 'b' - least recently used)")

	_, okB := lru.Get("b")
	_, okC := lru.Get("c")
	os.Stdout.WriteString(fmt.Sprintf("    Get(b) found=%t, Get(c) found=%t\n", okB, okC))

	// ============================================
	// CONTAINER/HEAP - PRIORITY QUEUE
	// ============================================
	fmt.Println(`
--- container/heap - Priority Queue ---

container/heap provides a min-heap on top of sort.Interface.
You implement 5 methods: Len, Less, Swap, Push, Pop.
The heap package calls these to maintain the heap invariant.

heap.Init   - Establishes heap ordering on existing data
heap.Push   - Adds element, maintains heap property
heap.Pop    - Removes and returns the minimum element
heap.Fix    - Re-establishes heap after modifying an element
heap.Remove - Removes element at index i`)

	pq := &PriorityQueue{}
	heap.Init(pq)

	heap.Push(pq, &PQItem{value: "low priority task", priority: 3})
	heap.Push(pq, &PQItem{value: "urgent task", priority: 1})
	heap.Push(pq, &PQItem{value: "medium task", priority: 2})
	heap.Push(pq, &PQItem{value: "critical task", priority: 0})

	fmt.Println("\n  Popping items (lowest priority number = highest priority):")
	for pq.Len() > 0 {
		item := heap.Pop(pq).(*PQItem)
		os.Stdout.WriteString(fmt.Sprintf("    priority=%d: %s\n", item.priority, item.value))
	}

	// ============================================
	// HEAP: MAX-HEAP AND CUSTOM ORDERING
	// ============================================
	fmt.Println(`
--- Max-Heap and Custom Ordering ---

Flip the Less method to get a max-heap.
The heap is always a min-heap based on your Less definition.`)

	maxH := &MaxIntHeap{5, 3, 8, 1, 9, 2}
	heap.Init(maxH)
	fmt.Println("\n  Max-heap (pop gives largest first):")
	os.Stdout.WriteString("    ")
	for maxH.Len() > 0 {
		os.Stdout.WriteString(fmt.Sprintf("%d ", heap.Pop(maxH)))
	}
	fmt.Println()

	// ============================================
	// HEAP: UPDATING PRIORITY
	// ============================================
	fmt.Println(`
--- Updating Priority (heap.Fix) ---

heap.Fix re-establishes the heap after modifying an element's priority.
This is O(log n) vs removing and re-adding which is also O(log n) but
with less allocation overhead.`)

	pq2 := &PriorityQueue{}
	heap.Init(pq2)

	items := []*PQItem{
		{value: "task-A", priority: 3},
		{value: "task-B", priority: 1},
		{value: "task-C", priority: 2},
	}
	for _, item := range items {
		heap.Push(pq2, item)
	}

	fmt.Println("\n  Before update:")
	for i, item := range *pq2 {
		os.Stdout.WriteString(fmt.Sprintf("    [%d] %s priority=%d\n", i, item.value, item.priority))
	}

	items[0].priority = 0
	heap.Fix(pq2, items[0].index)

	fmt.Println("  After changing task-A priority to 0:")
	for pq2.Len() > 0 {
		item := heap.Pop(pq2).(*PQItem)
		os.Stdout.WriteString(fmt.Sprintf("    %s priority=%d\n", item.value, item.priority))
	}

	// ============================================
	// CONTAINER/RING - CIRCULAR BUFFER
	// ============================================
	fmt.Println(`
--- container/ring - Circular Buffer ---

A ring is a circular doubly linked list with no beginning or end.
Each element points to the next and previous elements in a loop.

ring.New(n) - Create a ring with n elements
r.Value     - Get/set the current element's value
r.Next()    - Move to next element
r.Prev()    - Move to previous element
r.Move(n)   - Move n positions (positive=forward, negative=backward)
r.Len()     - Number of elements
r.Do(f)     - Call f for each element
r.Link(s)   - Connect ring r to ring s
r.Unlink(n) - Remove n elements starting from r.Next()`)

	r := ring.New(5)
	for i := 0; i < r.Len(); i++ {
		r.Value = i * 10
		r = r.Next()
	}

	fmt.Println("\n  Ring with 5 elements:")
	os.Stdout.WriteString("    Values: ")
	r.Do(func(v any) {
		os.Stdout.WriteString(fmt.Sprintf("%v ", v))
	})
	fmt.Println()
	os.Stdout.WriteString(fmt.Sprintf("    Length: %d\n", r.Len()))

	fmt.Println("\n  Navigation:")
	os.Stdout.WriteString(fmt.Sprintf("    Current:  %v\n", r.Value))
	os.Stdout.WriteString(fmt.Sprintf("    Next:     %v\n", r.Next().Value))
	os.Stdout.WriteString(fmt.Sprintf("    Prev:     %v\n", r.Prev().Value))
	os.Stdout.WriteString(fmt.Sprintf("    Move(2):  %v\n", r.Move(2).Value))
	os.Stdout.WriteString(fmt.Sprintf("    Move(-1): %v\n", r.Move(-1).Value))

	// ============================================
	// RING: LINK AND UNLINK
	// ============================================
	fmt.Println(`
--- Ring: Link and Unlink ---

r.Link(s) splices ring s into ring r (after r, before r.Next()).
r.Unlink(n) removes n elements starting from r.Next().`)

	r1 := ring.New(3)
	r2 := ring.New(2)

	for i := 0; i < 3; i++ {
		r1.Value = fmt.Sprintf("A%d", i)
		r1 = r1.Next()
	}
	for i := 0; i < 2; i++ {
		r2.Value = fmt.Sprintf("B%d", i)
		r2 = r2.Next()
	}

	os.Stdout.WriteString("\n  Ring 1: ")
	r1.Do(func(v any) { os.Stdout.WriteString(fmt.Sprintf("%v ", v)) })
	os.Stdout.WriteString("\n  Ring 2: ")
	r2.Do(func(v any) { os.Stdout.WriteString(fmt.Sprintf("%v ", v)) })

	r1.Link(r2)
	os.Stdout.WriteString("\n  After Link: ")
	r1.Do(func(v any) { os.Stdout.WriteString(fmt.Sprintf("%v ", v)) })
	os.Stdout.WriteString(fmt.Sprintf("(len=%d)\n", r1.Len()))

	removed := r1.Unlink(2)
	os.Stdout.WriteString("  After Unlink(2): ")
	r1.Do(func(v any) { os.Stdout.WriteString(fmt.Sprintf("%v ", v)) })
	os.Stdout.WriteString(fmt.Sprintf("(len=%d)\n", r1.Len()))
	os.Stdout.WriteString("  Removed ring: ")
	removed.Do(func(v any) { os.Stdout.WriteString(fmt.Sprintf("%v ", v)) })
	fmt.Println()

	// ============================================
	// RING USE CASES
	// ============================================
	fmt.Println(`
--- Ring Use Cases ---

  1. Circular buffers (logs, metrics with fixed window)
  2. Round-robin scheduling
  3. Rotating through a set of connections/servers
  4. Rolling averages / sliding windows`)

	fmt.Println("  Rolling average example (window of 5):")
	window := ring.New(5)
	values := []float64{10, 20, 30, 40, 50, 60, 70}

	for _, v := range values {
		window.Value = v
		window = window.Next()

		sum := 0.0
		count := 0
		window.Do(func(val any) {
			if val != nil {
				sum += val.(float64)
				count++
			}
		})
		if count > 0 {
			os.Stdout.WriteString(fmt.Sprintf("    Added %.0f -> average = %.1f (n=%d)\n", v, sum/float64(count), count))
		}
	}

	// ============================================
	// WHEN TO USE EACH CONTAINER
	// ============================================
	fmt.Println(`
--- When to Use Each Container ---

  Use SLICE when:
    - Random access needed
    - Mostly append/pop from end
    - Cache locality matters
    - Known or bounded size

  Use LIST when:
    - Frequent insert/remove at arbitrary positions
    - Need stable element references (pointers don't invalidate)
    - Implementing LRU cache, deque, undo stack

  Use HEAP when:
    - Need min/max element quickly
    - Priority queue, job scheduler
    - Top-K problems, merge K sorted lists

  Use RING when:
    - Fixed-size circular buffer
    - Round-robin scheduling
    - Sliding window calculations

  General advice:
    Slices cover 95% of use cases. Only reach for container types
    when slices genuinely don't fit the access pattern.`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println(`
--- Summary ---

  1. container/list: Doubly linked list with O(1) insert/remove
  2. container/heap: Binary min-heap for priority queues
  3. container/ring: Fixed-size circular buffer
  4. All use interface{}/any values (pre-generics)
  5. Wrap with typed methods for type safety
  6. Slices are better for most use cases
  7. heap requires implementing 5 methods (sort.Interface + Push/Pop)
  8. ring is useful for sliding windows and round-robin patterns`)
}

// ============================================
// HELPER TYPES AND FUNCTIONS
// ============================================

func printList(label string, l *list.List) {
	var parts []string
	for e := l.Front(); e != nil; e = e.Next() {
		parts = append(parts, fmt.Sprintf("%v", e.Value))
	}
	os.Stdout.WriteString(fmt.Sprintf("  %s: [%s]\n", label, strings.Join(parts, " -> ")))
}

// LRUCache is a simple LRU cache using container/list
type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

type lruEntry struct {
	key   string
	value any
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

func (c *LRUCache) Get(key string) (any, bool) {
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*lruEntry).value, true
	}
	return nil, false
}

func (c *LRUCache) Put(key string, value any) {
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*lruEntry).value = value
		return
	}
	if c.order.Len() >= c.capacity {
		oldest := c.order.Back()
		c.order.Remove(oldest)
		delete(c.items, oldest.Value.(*lruEntry).key)
	}
	elem := c.order.PushFront(&lruEntry{key: key, value: value})
	c.items[key] = elem
}

// PQItem represents an item in the priority queue
type PQItem struct {
	value    string
	priority int
	index    int
}

// PriorityQueue implements heap.Interface as a min-heap
type PriorityQueue []*PQItem

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool   { return pq[i].priority < pq[j].priority }
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*PQItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// MaxIntHeap implements heap.Interface as a max-heap for ints
type MaxIntHeap []int

func (h MaxIntHeap) Len() int            { return len(h) }
func (h MaxIntHeap) Less(i, j int) bool   { return h[i] > h[j] }
func (h MaxIntHeap) Swap(i, j int)        { h[i], h[j] = h[j], h[i] }

func (h *MaxIntHeap) Push(x any) {
	*h = append(*h, x.(int))
}

func (h *MaxIntHeap) Pop() any {
	old := *h
	n := len(old)
	val := old[n-1]
	*h = old[:n-1]
	return val
}
