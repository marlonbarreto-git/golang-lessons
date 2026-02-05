// Package main - Capítulo 56: Redis y Caching
// Redis es el cache más popular para Go. Aprenderás patrones de
// caching, go-redis, y estrategias de invalidación.
package main

import (
	"os"
	"context"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== REDIS Y CACHING EN GO ===")

	// ============================================
	// GO-REDIS
	// ============================================
	fmt.Println("\n--- go-redis ---")
	fmt.Println(`
INSTALACIÓN:
go get github.com/redis/go-redis/v9

CONEXIÓN BÁSICA:

import "github.com/redis/go-redis/v9"

func NewRedisClient() *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",  // no password
        DB:       0,   // default DB
    })
}

// Verificar conexión
ctx := context.Background()
_, err := client.Ping(ctx).Result()

CONEXIÓN CON URL:

opt, _ := redis.ParseURL("redis://user:pass@localhost:6379/0")
client := redis.NewClient(opt)

CLUSTER:

client := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "localhost:7000",
        "localhost:7001",
        "localhost:7002",
    },
})

SENTINEL (HA):

client := redis.NewFailoverClient(&redis.FailoverOptions{
    MasterName:    "mymaster",
    SentinelAddrs: []string{"localhost:26379"},
})`)
	// ============================================
	// OPERACIONES BÁSICAS
	// ============================================
	fmt.Println("\n--- Operaciones Básicas ---")
	fmt.Println(`
STRINGS:

// Set con TTL
err := client.Set(ctx, "key", "value", time.Hour).Err()

// Get
val, err := client.Get(ctx, "key").Result()
if err == redis.Nil {
    // Key no existe
}

// SetNX (solo si no existe)
set, err := client.SetNX(ctx, "key", "value", time.Hour).Result()

// GetSet (get old, set new)
old, err := client.GetSet(ctx, "key", "newvalue").Result()

// Increment/Decrement
client.Incr(ctx, "counter")
client.IncrBy(ctx, "counter", 10)
client.Decr(ctx, "counter")

HASHES:

// Set field
client.HSet(ctx, "user:1", "name", "Alice", "age", "30")

// Get field
name, _ := client.HGet(ctx, "user:1", "name").Result()

// Get all
data, _ := client.HGetAll(ctx, "user:1").Result()

// Struct to hash
type User struct {
    Name string ` + "`" + `redis:"name"` + "`" + `
    Age  int    ` + "`" + `redis:"age"` + "`" + `
}

client.HSet(ctx, "user:1", User{Name: "Alice", Age: 30})

var user User
client.HGetAll(ctx, "user:1").Scan(&user)

LISTS:

client.LPush(ctx, "queue", "item1", "item2")
client.RPush(ctx, "queue", "item3")
item, _ := client.LPop(ctx, "queue").Result()
items, _ := client.LRange(ctx, "queue", 0, -1).Result()

SETS:

client.SAdd(ctx, "tags", "go", "redis", "cache")
members, _ := client.SMembers(ctx, "tags").Result()
isMember, _ := client.SIsMember(ctx, "tags", "go").Result()

SORTED SETS:

client.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "player1"})
topPlayers, _ := client.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()`)
	// ============================================
	// PATRONES DE CACHE
	// ============================================
	fmt.Println("\n--- Patrones de Cache ---")
	fmt.Println(`
1. CACHE-ASIDE (Lazy Loading):

func GetUser(ctx context.Context, id string) (*User, error) {
    // 1. Intentar cache
    key := "user:" + id
    cached, err := cache.Get(ctx, key).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // 2. Cache miss: cargar de DB
    user, err := db.FindUser(id)
    if err != nil {
        return nil, err
    }

    // 3. Guardar en cache
    data, _ := json.Marshal(user)
    cache.Set(ctx, key, data, time.Hour)

    return user, nil
}

2. WRITE-THROUGH:

func UpdateUser(ctx context.Context, user *User) error {
    // 1. Escribir a DB
    if err := db.UpdateUser(user); err != nil {
        return err
    }

    // 2. Actualizar cache
    key := "user:" + user.ID
    data, _ := json.Marshal(user)
    return cache.Set(ctx, key, data, time.Hour).Err()
}

3. WRITE-BEHIND (Async):

func UpdateUserAsync(ctx context.Context, user *User) error {
    // 1. Actualizar cache inmediatamente
    key := "user:" + user.ID
    data, _ := json.Marshal(user)
    cache.Set(ctx, key, data, time.Hour)

    // 2. Encolar escritura a DB
    return queue.Publish(ctx, "db-writes", data)
}`)
	// ============================================
	// CACHE INVALIDATION
	// ============================================
	fmt.Println("\n--- Cache Invalidation ---")
	os.Stdout.WriteString(`
"There are only two hard things in CS: cache invalidation and naming things."

1. TTL (Time-To-Live):

cache.Set(ctx, key, value, 5*time.Minute)  // Expira en 5 min

2. DELETE ON UPDATE:

func UpdateUser(ctx context.Context, user *User) error {
    err := db.UpdateUser(user)
    if err == nil {
        cache.Del(ctx, "user:"+user.ID)
    }
    return err
}

3. PATTERN DELETE:

// Eliminar todas las keys que matchean
keys, _ := cache.Keys(ctx, "user:*").Result()
if len(keys) > 0 {
    cache.Del(ctx, keys...)
}

// SCAN es mejor para producción (no bloquea)
var cursor uint64
for {
    keys, cursor, _ = cache.Scan(ctx, cursor, "user:*", 100).Result()
    if len(keys) > 0 {
        cache.Del(ctx, keys...)
    }
    if cursor == 0 {
        break
    }
}

4. TAGS/GROUPS:

// Guardar con tag
cache.SAdd(ctx, "tag:users", "user:1", "user:2")
cache.Set(ctx, "user:1", data, 0)

// Invalidar todo el tag
keys, _ := cache.SMembers(ctx, "tag:users").Result()
if len(keys) > 0 {
    cache.Del(ctx, keys...)
}
cache.Del(ctx, "tag:users")

5. VERSIONING:

// Key incluye versión
version := cache.Incr(ctx, "users:version").Val()
key := fmt.Sprintf("users:v%d:list", version)
`)

	// ============================================
	// CACHE STAMPEDE PREVENTION
	// ============================================
	fmt.Println("\n--- Cache Stampede Prevention ---")
	fmt.Println(`
Problema: Cuando el cache expira, muchos requests golpean la DB.

1. MUTEX/SINGLEFLIGHT:

import "golang.org/x/sync/singleflight"

var sf singleflight.Group

func GetUserSafe(ctx context.Context, id string) (*User, error) {
    key := "user:" + id

    // Intentar cache
    if cached, err := cache.Get(ctx, key).Result(); err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // Singleflight: solo una goroutine carga
    result, err, _ := sf.Do(key, func() (any, error) {
        user, err := db.FindUser(id)
        if err != nil {
            return nil, err
        }
        data, _ := json.Marshal(user)
        cache.Set(ctx, key, data, time.Hour)
        return user, nil
    })

    if err != nil {
        return nil, err
    }
    return result.(*User), nil
}

2. DISTRIBUTED LOCK:

func GetWithLock(ctx context.Context, key string) (string, error) {
    // Intentar cache
    val, err := cache.Get(ctx, key).Result()
    if err == nil {
        return val, nil
    }

    // Adquirir lock
    lockKey := "lock:" + key
    acquired, _ := cache.SetNX(ctx, lockKey, "1", 10*time.Second).Result()

    if acquired {
        defer cache.Del(ctx, lockKey)

        // Cargar datos
        data := loadFromDB(key)
        cache.Set(ctx, key, data, time.Hour)
        return data, nil
    }

    // Otro proceso tiene el lock, esperar
    time.Sleep(100 * time.Millisecond)
    return cache.Get(ctx, key).Result()
}

3. PROBABILISTIC EARLY EXPIRATION:

func GetWithEarlyRefresh(ctx context.Context, key string) (string, error) {
    val, ttl, _ := cache.GetEx(ctx, key, 0).Result()

    // Si TTL < 10% del original, refrescar proactivamente
    if ttl < 6*time.Minute && rand.Float64() < 0.1 {
        go refreshCache(key)
    }

    return val, nil
}`)
	// ============================================
	// PUB/SUB
	// ============================================
	fmt.Println("\n--- Pub/Sub ---")
	fmt.Println(`
PUBLISHER:

func PublishEvent(ctx context.Context, event Event) error {
    data, _ := json.Marshal(event)
    return cache.Publish(ctx, "events", data).Err()
}

SUBSCRIBER:

func Subscribe(ctx context.Context) {
    pubsub := cache.Subscribe(ctx, "events")
    defer pubsub.Close()

    ch := pubsub.Channel()
    for msg := range ch {
        var event Event
        json.Unmarshal([]byte(msg.Payload), &event)
        handleEvent(event)
    }
}

PATTERN SUBSCRIBE:

pubsub := cache.PSubscribe(ctx, "user:*")
for msg := range pubsub.Channel() {
    fmt.Println("Channel:", msg.Channel)  // user:123
    fmt.Println("Pattern:", msg.Pattern)  // user:*
    fmt.Println("Payload:", msg.Payload)
}`)
	// ============================================
	// PIPELINES Y TRANSACTIONS
	// ============================================
	fmt.Println("\n--- Pipelines y Transactions ---")
	fmt.Println(`
PIPELINE (batch commands):

pipe := cache.Pipeline()
incr := pipe.Incr(ctx, "counter")
pipe.Expire(ctx, "counter", time.Hour)
pipe.Set(ctx, "key", "value", 0)
_, err := pipe.Exec(ctx)

// Los resultados están disponibles después de Exec
fmt.Println(incr.Val())

TRANSACTION (MULTI/EXEC):

// Watch key y ejecutar transacción
err := cache.Watch(ctx, func(tx *redis.Tx) error {
    val, err := tx.Get(ctx, "key").Int()
    if err != nil {
        return err
    }

    _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.Set(ctx, "key", val+1, 0)
        return nil
    })
    return err
}, "key")

if err == redis.TxFailedErr {
    // Key fue modificada, reintentar
}`)
	// ============================================
	// RATE LIMITING CON REDIS
	// ============================================
	fmt.Println("\n--- Rate Limiting ---")
	os.Stdout.WriteString(`
SLIDING WINDOW:

func IsAllowed(ctx context.Context, userID string, limit int, window time.Duration) bool {
    key := fmt.Sprintf("ratelimit:%s", userID)
    now := time.Now().UnixNano()
    windowStart := now - int64(window)

    pipe := cache.Pipeline()

    // Eliminar requests viejos
    pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(windowStart))

    // Contar requests en ventana
    pipe.ZCard(ctx, key)

    // Agregar request actual
    pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

    // Establecer TTL
    pipe.Expire(ctx, key, window)

    cmds, _ := pipe.Exec(ctx)
    count := cmds[1].(*redis.IntCmd).Val()

    return count < int64(limit)
}

TOKEN BUCKET:

func TakeToken(ctx context.Context, key string, rate float64, capacity int) bool {
    script := redis.NewScript(` + "`" + `
        local tokens = tonumber(redis.call('get', KEYS[1]) or capacity)
        local lastRefill = tonumber(redis.call('get', KEYS[2]) or 0)
        local now = tonumber(ARGV[1])
        local rate = tonumber(ARGV[2])
        local capacity = tonumber(ARGV[3])

        local elapsed = now - lastRefill
        tokens = math.min(capacity, tokens + elapsed * rate)

        if tokens >= 1 then
            tokens = tokens - 1
            redis.call('set', KEYS[1], tokens)
            redis.call('set', KEYS[2], now)
            return 1
        end
        return 0
    ` + "`" + `)

    result, _ := script.Run(ctx, cache,
        []string{key + ":tokens", key + ":last"},
        time.Now().Unix(), rate, capacity,
    ).Int()

    return result == 1
}
`)

	// ============================================
	// SESIONES
	// ============================================
	fmt.Println("\n--- Session Store ---")
	fmt.Println(`
SESSION MANAGEMENT:

type Session struct {
    ID        string
    UserID    string
    Data      map[string]any
    ExpiresAt time.Time
}

func CreateSession(ctx context.Context, userID string) (*Session, error) {
    session := &Session{
        ID:        uuid.NewString(),
        UserID:    userID,
        Data:      make(map[string]any),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }

    data, _ := json.Marshal(session)
    err := cache.Set(ctx, "session:"+session.ID, data, 24*time.Hour).Err()
    return session, err
}

func GetSession(ctx context.Context, sessionID string) (*Session, error) {
    data, err := cache.Get(ctx, "session:"+sessionID).Bytes()
    if err != nil {
        return nil, err
    }

    var session Session
    json.Unmarshal(data, &session)
    return &session, nil
}

func DestroySession(ctx context.Context, sessionID string) error {
    return cache.Del(ctx, "session:"+sessionID).Err()
}`)
	// ============================================
	// CONNECTION POOLING
	// ============================================
	fmt.Println("\n--- Connection Pooling ---")
	os.Stdout.WriteString(`
CONFIGURACIÓN DE POOL:

client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",

    // Pool size
    PoolSize:     100,              // Conexiones máximas
    MinIdleConns: 10,               // Conexiones mínimas idle

    // Timeouts
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolTimeout:  4 * time.Second,

    // Idle connection cleanup
    ConnMaxIdleTime: 5 * time.Minute,
    ConnMaxLifetime: 0,  // No limit
})

MONITOREAR POOL:

stats := client.PoolStats()
fmt.Printf("Hits: %d\n", stats.Hits)
fmt.Printf("Misses: %d\n", stats.Misses)
fmt.Printf("Timeouts: %d\n", stats.Timeouts)
fmt.Printf("TotalConns: %d\n", stats.TotalConns)
fmt.Printf("IdleConns: %d\n", stats.IdleConns)
`)

	// Demostración conceptual
	demonstrateCaching()
}

func demonstrateCaching() {
	ctx := context.Background()
	fmt.Printf("\n--- Demostración Conceptual ---\n")
	fmt.Printf("Context: %v\n", ctx)
	fmt.Printf("TTL recomendado para diferentes datos:\n")
	fmt.Printf("  - Sesiones: %v\n", 24*time.Hour)
	fmt.Printf("  - User profiles: %v\n", 15*time.Minute)
	fmt.Printf("  - API responses: %v\n", 5*time.Minute)
	fmt.Printf("  - Static config: %v\n", time.Hour)
}

/*
RESUMEN DE REDIS Y CACHING:

LIBRERÍA:
go get github.com/redis/go-redis/v9

CONEXIÓN:
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

TIPOS DE DATOS:
- Strings: Set/Get, Incr, SetNX
- Hashes: HSet/HGet/HGetAll
- Lists: LPush/RPush/LPop/RPop
- Sets: SAdd/SMembers/SIsMember
- Sorted Sets: ZAdd/ZRange/ZRevRange

PATRONES DE CACHE:
1. Cache-Aside (Lazy Loading)
2. Write-Through
3. Write-Behind (Async)

INVALIDACIÓN:
- TTL automático
- Delete on update
- Pattern delete con SCAN
- Tags/Groups
- Versioning

STAMPEDE PREVENTION:
- singleflight (una goroutine)
- Distributed locks (SetNX)
- Early probabilistic refresh

PUB/SUB:
client.Publish(ctx, channel, msg)
pubsub := client.Subscribe(ctx, channel)

PIPELINES:
pipe := client.Pipeline()
pipe.Set(...)
pipe.Get(...)
pipe.Exec(ctx)

RATE LIMITING:
- Sliding window (ZSet)
- Token bucket (Lua script)

POOL CONFIG:
PoolSize, MinIdleConns, Timeouts

MEJORES PRÁCTICAS:
1. Usar TTL siempre
2. Prevenir stampede
3. Monitorear hit rate
4. Usar pipelines para batch
5. Configurar pool adecuadamente
*/
