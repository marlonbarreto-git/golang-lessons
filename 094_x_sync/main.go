// Package main - Chapter 094: golang.org/x/sync
// The x/sync package provides advanced synchronization primitives:
// errgroup for parallel task execution, singleflight for call deduplication,
// and semaphore for weighted concurrency limiting.
package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("=== GOLANG.ORG/X/SYNC ===")

	fmt.Println(`
  golang.org/x/sync extends the standard sync package with:

  1. errgroup  - manage groups of goroutines with error propagation
  2. singleflight - deduplicate concurrent function calls
  3. semaphore - weighted semaphore for concurrency limiting

  Install: go get golang.org/x/sync`)

	// ============================================
	// ERRGROUP - CONCEPT AND API
	// ============================================
	fmt.Println("\n--- errgroup: Parallel Tasks with Error Handling ---")

	fmt.Println(`
  errgroup manages a group of goroutines working on subtasks
  of a common task. It handles:
  - Waiting for all goroutines to complete
  - Collecting the first error from any goroutine
  - Context cancellation when any goroutine fails

  API:
    g, ctx := errgroup.WithContext(ctx)  // create group with context
    g := new(errgroup.Group)             // create without context

    g.Go(func() error { ... })    // launch goroutine
    g.SetLimit(n)                 // limit concurrent goroutines
    g.TryGo(func() error { ... })// launch if under limit (returns bool)

    err := g.Wait()  // wait for all, return first error

  EXAMPLE - Parallel HTTP fetches:

    g, ctx := errgroup.WithContext(ctx)

    urls := []string{"https://a.com", "https://b.com", "https://c.com"}

    results := make([]string, len(urls))
    for i, url := range urls {
        g.Go(func() error {
            resp, err := http.Get(url)
            if err != nil { return err }
            defer resp.Body.Close()
            body, err := io.ReadAll(resp.Body)
            if err != nil { return err }
            results[i] = string(body)
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        log.Fatal(err)
    }
    // All results are populated`)

	// Working errgroup-like example using stdlib
	fmt.Println("\n  Working example (errgroup pattern with stdlib):")
	demonstrateErrGroupPattern()

	fmt.Println(`
  ERRGROUP WITH LIMIT:

    g := new(errgroup.Group)
    g.SetLimit(3)  // max 3 concurrent goroutines

    for _, item := range items {
        g.Go(func() error {
            return process(item)
        })
    }
    // At most 3 goroutines run concurrently
    err := g.Wait()

  ERRGROUP WITH TRYGO:

    g := new(errgroup.Group)
    g.SetLimit(5)

    for _, item := range items {
        if !g.TryGo(func() error {
            return process(item)
        }) {
            // Group is at limit, handle overflow
            log.Println("at capacity, skipping", item)
        }
    }

  CONTEXT CANCELLATION:

    g, ctx := errgroup.WithContext(parentCtx)

    g.Go(func() error {
        return longTask(ctx)  // use group's ctx
    })
    g.Go(func() error {
        return failingTask()  // returns error
    })

    // When failingTask returns error:
    // 1. ctx is cancelled (signals longTask to stop)
    // 2. Wait() returns failingTask's error`)

	// ============================================
	// SINGLEFLIGHT - CONCEPT AND API
	// ============================================
	fmt.Println("\n--- singleflight: Call Deduplication ---")

	fmt.Println(`
  singleflight ensures only ONE execution of a function
  is in-flight for a given key at a time. Duplicate callers
  wait for the in-flight call and receive its result.

  PROBLEM IT SOLVES - Cache stampede:
    1000 requests arrive for same uncached key
    Without singleflight: 1000 DB queries
    With singleflight: 1 DB query, 999 get shared result

  API:
    var g singleflight.Group

    // Do executes fn once per key, sharing result
    v, err, shared := g.Do(key, func() (any, error) {
        return fetchFromDB(key)
    })
    // shared=true means result was from another caller

    // DoChan is like Do but returns a channel
    ch := g.DoChan(key, func() (any, error) {
        return fetchFromDB(key)
    })
    result := <-ch  // result.Val, result.Err, result.Shared

    // Forget removes key, allowing next call to execute
    g.Forget(key)`)

	// Working singleflight-like example
	fmt.Println("\n  Working example (singleflight pattern with stdlib):")
	demonstrateSingleflightPattern()

	fmt.Println(`
  REAL-WORLD PATTERNS:

  PATTERN 1: Cache with singleflight

    type Cache struct {
        mu    sync.RWMutex
        data  map[string]any
        group singleflight.Group
    }

    func (c *Cache) Get(key string) (any, error) {
        // Check cache first
        c.mu.RLock()
        if v, ok := c.data[key]; ok {
            c.mu.RUnlock()
            return v, nil
        }
        c.mu.RUnlock()

        // Deduplicate concurrent cache misses
        v, err, _ := c.group.Do(key, func() (any, error) {
            result, err := fetchFromDB(key)
            if err != nil { return nil, err }
            c.mu.Lock()
            c.data[key] = result
            c.mu.Unlock()
            return result, nil
        })
        return v, err
    }

  PATTERN 2: Rate-limited API calls

    var apiGroup singleflight.Group

    func GetUserProfile(userID string) (*Profile, error) {
        v, err, _ := apiGroup.Do(userID, func() (any, error) {
            return api.FetchProfile(userID)
        })
        if err != nil { return nil, err }
        return v.(*Profile), nil
    }`)

	// ============================================
	// SEMAPHORE - CONCEPT AND API
	// ============================================
	fmt.Println("\n--- semaphore: Weighted Concurrency Limiting ---")

	fmt.Println(`
  A weighted semaphore limits concurrent access to a resource
  based on a total weight. Unlike sync.Mutex (binary: locked/unlocked),
  semaphore allows N concurrent accesses.

  API:
    sem := semaphore.NewWeighted(maxWeight)

    // Acquire weight (blocks until available or ctx cancelled)
    err := sem.Acquire(ctx, weight)

    // TryAcquire returns immediately
    ok := sem.TryAcquire(weight)

    // Release weight
    sem.Release(weight)

  EXAMPLE - Limit concurrent downloads:

    sem := semaphore.NewWeighted(10)  // max 10 concurrent

    for _, url := range urls {
        sem.Acquire(ctx, 1)
        go func() {
            defer sem.Release(1)
            download(url)
        }()
    }`)

	// Working semaphore-like example
	fmt.Println("\n  Working example (semaphore pattern with stdlib):")
	demonstrateSemaphorePattern()

	fmt.Println(`
  WEIGHTED SEMAPHORE PATTERNS:

  PATTERN 1: Different weights for different tasks

    sem := semaphore.NewWeighted(100)  // total capacity: 100

    // Small task takes weight 1
    sem.Acquire(ctx, 1)
    go func() {
        defer sem.Release(1)
        smallTask()
    }()

    // Large task takes weight 50
    sem.Acquire(ctx, 50)
    go func() {
        defer sem.Release(50)
        largeTask()  // limits concurrency with other tasks
    }()

  PATTERN 2: Worker pool with semaphore

    const maxWorkers = 5
    sem := semaphore.NewWeighted(maxWorkers)

    for _, job := range jobs {
        if err := sem.Acquire(ctx, 1); err != nil {
            break  // context cancelled
        }
        go func() {
            defer sem.Release(1)
            process(job)
        }()
    }
    // Wait for all workers
    sem.Acquire(ctx, maxWorkers)

  PATTERN 3: Memory-bounded processing

    const maxMemMB = 1024
    sem := semaphore.NewWeighted(maxMemMB)

    for _, file := range files {
        sizeMB := file.Size() / (1024 * 1024)
        sem.Acquire(ctx, sizeMB)
        go func() {
            defer sem.Release(sizeMB)
            processFile(file)
        }()
    }`)

	// ============================================
	// COMPARISON: WHEN TO USE WHAT
	// ============================================
	fmt.Println("\n--- When to Use Each Primitive ---")

	fmt.Println(`
  +----------------+-----------------------------+---------------------------+
  | Primitive      | Use When                    | Key Feature               |
  +----------------+-----------------------------+---------------------------+
  | errgroup       | Parallel subtasks that can  | Error propagation +       |
  |                | fail, need first error      | context cancellation      |
  +----------------+-----------------------------+---------------------------+
  | singleflight   | Multiple callers request    | Call deduplication,       |
  |                | same expensive operation    | prevents cache stampede   |
  +----------------+-----------------------------+---------------------------+
  | semaphore      | Limit concurrent access to  | Weighted, context-aware   |
  |                | a shared resource           | backpressure              |
  +----------------+-----------------------------+---------------------------+

  VS STDLIB:

  errgroup vs sync.WaitGroup:
    - WaitGroup: fire-and-forget, no error handling
    - errgroup: collects errors, cancels on failure

  singleflight vs sync.Once:
    - Once: runs exactly once forever
    - singleflight: deduplicates concurrent calls, can retry

  semaphore vs buffered channel:
    - Buffered chan: binary (slot or no slot)
    - Semaphore: weighted, context-aware`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  golang.org/x/sync ESSENTIALS:

  errgroup:
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(n)
    g.Go(func() error { ... })
    err := g.Wait()

  singleflight:
    var g singleflight.Group
    v, err, shared := g.Do(key, fn)
    ch := g.DoChan(key, fn)
    g.Forget(key)

  semaphore:
    sem := semaphore.NewWeighted(n)
    sem.Acquire(ctx, weight)
    sem.TryAcquire(weight)
    sem.Release(weight)

  Install: go get golang.org/x/sync
  Import paths:
    golang.org/x/sync/errgroup
    golang.org/x/sync/singleflight
    golang.org/x/sync/semaphore`)
}

// ============================================
// WORKING EXAMPLES WITH STDLIB
// ============================================

func demonstrateErrGroupPattern() {
	type result struct {
		index int
		value string
		err   error
	}

	tasks := []string{"fetch-users", "fetch-orders", "fetch-products"}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	results := make(chan result, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				results <- result{idx, "", ctx.Err()}
			default:
				time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
				results <- result{idx, fmt.Sprintf("%s: done", name), nil}
			}
		}(i, task)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r.err != nil {
			os.Stdout.WriteString(fmt.Sprintf("    Task %d error: %v\n", r.index, r.err))
		} else {
			os.Stdout.WriteString(fmt.Sprintf("    Task %d: %s\n", r.index, r.value))
		}
	}
}

func demonstrateSingleflightPattern() {
	var (
		mu       sync.Mutex
		inflight = make(map[string]chan struct{ val string; err error })
	)

	var callCount atomic.Int32

	fetch := func(key string) (string, error) {
		mu.Lock()
		if ch, ok := inflight[key]; ok {
			mu.Unlock()
			result := <-ch
			return result.val, result.err
		}
		ch := make(chan struct{ val string; err error }, 1)
		inflight[key] = ch
		mu.Unlock()

		callCount.Add(1)
		time.Sleep(10 * time.Millisecond)
		val := fmt.Sprintf("data-for-%s", key)

		result := struct{ val string; err error }{val, nil}
		ch <- result

		mu.Lock()
		delete(inflight, key)
		mu.Unlock()

		return result.val, result.err
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := fetch("user:123")
			if err == nil {
				_ = val
			}
		}()
	}
	wg.Wait()

	os.Stdout.WriteString(fmt.Sprintf("    5 concurrent requests, actual DB calls: %d (deduplicated!)\n", callCount.Load()))
}

func demonstrateSemaphorePattern() {
	maxConcurrent := 3
	sem := make(chan struct{}, maxConcurrent)

	var completed atomic.Int32
	var wg sync.WaitGroup

	tasks := 10
	for i := 0; i < tasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			time.Sleep(10 * time.Millisecond)
			completed.Add(1)
		}(i)
	}

	wg.Wait()
	os.Stdout.WriteString(fmt.Sprintf("    Processed %d tasks with max %d concurrent (semaphore pattern)\n",
		completed.Load(), maxConcurrent))
}

// Ensure error type is used
var _ = errors.New
