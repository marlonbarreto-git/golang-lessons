// Package main - Chapter 034: Sync Package
// El paquete sync proporciona primitivas de sincronización de bajo nivel
// para coordinar goroutines de manera segura.
package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// TIPOS AUXILIARES (fuera de main)
// ============================================

type SafeCounter struct {
	mu    sync.Mutex
	count int
}

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

type ThreadSafeCache struct {
	mu    sync.RWMutex
	items map[string]any
}

func (c *ThreadSafeCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

func (c *ThreadSafeCache) Set(key string, val any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = val
}

type Database struct {
	name string
}

func main() {
	fmt.Println("=== PAQUETE SYNC ===")

	// ============================================
	// MUTEX (MUTUAL EXCLUSION)
	// ============================================
	fmt.Println("\n--- Mutex ---")

	counter := &SafeCounter{}

	increment := func(wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			counter.mu.Lock()
			counter.count++
			counter.mu.Unlock()
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go increment(&wg)
	}
	wg.Wait()

	fmt.Printf("Counter final (con Mutex): %d\n", counter.count)

	// ============================================
	// RWMUTEX (READ-WRITE MUTEX)
	// ============================================
	fmt.Println("\n--- RWMutex ---")

	cache := &Cache{data: make(map[string]string)}

	cache.mu.Lock()
	cache.data["key1"] = "value1"
	cache.data["key2"] = "value2"
	cache.mu.Unlock()

	var readWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		readWg.Add(1)
		go func(id int) {
			defer readWg.Done()
			cache.mu.RLock()
			defer cache.mu.RUnlock()
			fmt.Printf("Reader %d: key1=%s\n", id, cache.data["key1"])
		}(i)
	}
	readWg.Wait()

	// ============================================
	// WAITGROUP
	// ============================================
	fmt.Println("\n--- WaitGroup ---")

	var wg2 sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg2.Add(1)
		go func(n int) {
			defer wg2.Done()
			time.Sleep(time.Duration(n*10) * time.Millisecond)
			fmt.Printf("Worker %d completado\n", n)
		}(i)
	}
	wg2.Wait()
	fmt.Println("Todos los workers terminaron")

	// Go 1.25+: WaitGroup.Go()
	fmt.Println("\n--- WaitGroup.Go() (Go 1.25+) ---")

	var wg3 sync.WaitGroup
	for i := 1; i <= 3; i++ {
		n := i
		wg3.Go(func() {
			fmt.Printf("Go-worker %d ejecutado\n", n)
		})
	}
	wg3.Wait()

	// ============================================
	// ONCE
	// ============================================
	fmt.Println("\n--- Once ---")

	var once sync.Once
	initCounter := 0

	initialize := func() {
		initCounter++
		fmt.Println("Inicialización ejecutada")
	}

	var onceWg sync.WaitGroup
	for i := 0; i < 5; i++ {
		onceWg.Add(1)
		go func() {
			defer onceWg.Done()
			once.Do(initialize)
		}()
	}
	onceWg.Wait()

	fmt.Printf("Inicialización ejecutada %d vez(ces)\n", initCounter)

	// ============================================
	// POOL
	// ============================================
	fmt.Println("\n--- Pool ---")

	bufferPool := &sync.Pool{
		New: func() any {
			fmt.Println("Creando nuevo buffer")
			return make([]byte, 1024)
		},
	}

	buf1 := bufferPool.Get().([]byte)
	fmt.Printf("Buffer 1 len: %d\n", len(buf1))

	bufferPool.Put(buf1)

	buf2 := bufferPool.Get().([]byte)
	fmt.Printf("Buffer 2 len: %d\n", len(buf2))

	// ============================================
	// COND (CONDITION VARIABLE)
	// ============================================
	fmt.Println("\n--- Cond ---")

	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false

	go func() {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		ready = true
		cond.Broadcast()
		mu.Unlock()
	}()

	var condWg sync.WaitGroup
	for i := 0; i < 3; i++ {
		condWg.Add(1)
		go func(id int) {
			defer condWg.Done()
			mu.Lock()
			for !ready {
				cond.Wait()
			}
			fmt.Printf("Consumidor %d: ready!\n", id)
			mu.Unlock()
		}(i)
	}
	condWg.Wait()

	// ============================================
	// MAP (CONCURRENT MAP)
	// ============================================
	fmt.Println("\n--- sync.Map ---")

	var m sync.Map

	m.Store("key1", "value1")
	m.Store("key2", "value2")
	m.Store("key3", "value3")

	if val, ok := m.Load("key1"); ok {
		fmt.Printf("key1: %v\n", val)
	}

	actual, loaded := m.LoadOrStore("key4", "value4")
	fmt.Printf("key4: %v, loaded: %v\n", actual, loaded)

	m.Range(func(key, value any) bool {
		fmt.Printf("  %v: %v\n", key, value)
		return true
	})

	m.Delete("key3")

	val, loaded := m.LoadAndDelete("key2")
	fmt.Printf("Deleted key2: %v, was loaded: %v\n", val, loaded)

	// ============================================
	// ATOMIC OPERATIONS
	// ============================================
	fmt.Println("\n--- Atomic Operations ---")

	var atomicCounter atomic.Int64

	var atomicWg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		atomicWg.Add(1)
		go func() {
			defer atomicWg.Done()
			atomicCounter.Add(1)
		}()
	}
	atomicWg.Wait()

	fmt.Printf("Atomic counter: %d\n", atomicCounter.Load())

	var ptr atomic.Pointer[string]
	s := "hello"
	ptr.Store(&s)
	fmt.Printf("Atomic pointer: %s\n", *ptr.Load())

	var cas atomic.Int32
	cas.Store(100)
	swapped := cas.CompareAndSwap(100, 200)
	fmt.Printf("CAS swapped: %v, value: %d\n", swapped, cas.Load())

	// ============================================
	// ONCEFUNC / ONCEVALUE / ONCEVALUES (Go 1.21+)
	// ============================================
	fmt.Println("\n--- OnceFunc / OnceValue (Go 1.21+) ---")

	heavyInit := sync.OnceFunc(func() {
		fmt.Println("Heavy initialization done")
	})

	for i := 0; i < 3; i++ {
		heavyInit()
	}

	getConfig := sync.OnceValue(func() string {
		fmt.Println("Loading config...")
		return "production"
	})

	fmt.Printf("Config 1: %s\n", getConfig())
	fmt.Printf("Config 2: %s\n", getConfig())

	loadData := sync.OnceValues(func() ([]byte, error) {
		fmt.Println("Loading data from file...")
		return []byte("data"), nil
	})

	data1, err1 := loadData()
	data2, err2 := loadData()
	fmt.Printf("Data1: %s, err1: %v\n", data1, err1)
	fmt.Printf("Data2: %s, err2: %v\n", data2, err2)

	// ============================================
	// PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- Patrones Comunes ---")

	// Singleton con sync.Once
	var (
		dbInstance *Database
		dbOnce     sync.Once
	)

	getDB := func() *Database {
		dbOnce.Do(func() {
			fmt.Println("Conectando a la base de datos...")
			dbInstance = &Database{name: "production_db"}
		})
		return dbInstance
	}

	db1 := getDB()
	db2 := getDB()
	fmt.Printf("Same instance: %v\n", db1 == db2)

	// Thread-safe cache
	tsc := &ThreadSafeCache{items: make(map[string]any)}
	tsc.Set("user:1", "Alice")
	if val, ok := tsc.Get("user:1"); ok {
		fmt.Printf("Cache hit: %v\n", val)
	}
}

/*
RESUMEN DE SYNC:

MUTEX:
var mu sync.Mutex
mu.Lock() / mu.Unlock()

RWMUTEX:
var mu sync.RWMutex
mu.RLock() / mu.RUnlock()
mu.Lock() / mu.Unlock()

WAITGROUP:
var wg sync.WaitGroup
wg.Add(1) / wg.Done() / wg.Wait()
wg.Go(func() { ... }) // Go 1.25+

ONCE:
var once sync.Once
once.Do(initFunc)

POOL:
pool := &sync.Pool{New: func() any { ... }}
obj := pool.Get()
pool.Put(obj)

COND:
cond := sync.NewCond(&mu)
cond.Wait() / cond.Signal() / cond.Broadcast()

SYNC.MAP:
var m sync.Map
m.Store(k, v) / m.Load(k) / m.Delete(k)

ATOMIC:
var counter atomic.Int64
counter.Add(1) / counter.Load() / counter.Store(0)
*/
