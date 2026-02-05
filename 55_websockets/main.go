// Package main - Capítulo 55: WebSockets
// WebSockets permiten comunicación bidireccional en tiempo real.
// Go tiene excelente soporte con gorilla/websocket y nhooyr/websocket.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== WEBSOCKETS EN GO ===")

	// ============================================
	// CONCEPTO
	// ============================================
	fmt.Println("\n--- Concepto ---")
	fmt.Println(`
WebSocket es un protocolo de comunicación bidireccional sobre TCP.

HTTP vs WebSocket:
- HTTP: Request-Response (cliente inicia)
- WebSocket: Full-duplex (ambos pueden enviar)

CASOS DE USO:
- Chat en tiempo real
- Notificaciones push
- Gaming multiplayer
- Colaboración en vivo (Google Docs)
- Streaming de datos (trading, IoT)
- Live updates (dashboards)

HANDSHAKE:
1. Cliente envía HTTP Upgrade request
2. Servidor acepta con 101 Switching Protocols
3. Conexión WebSocket establecida`)
	// ============================================
	// GORILLA/WEBSOCKET
	// ============================================
	fmt.Println("\n--- gorilla/websocket ---")
	fmt.Println(`
INSTALACIÓN:
go get github.com/gorilla/websocket

SERVIDOR BÁSICO:

import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true  // Permitir todos los origins (dev only!)
    },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("upgrade error:", err)
        return
    }
    defer conn.Close()

    for {
        // Leer mensaje
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            return
        }

        // Echo: enviar de vuelta
        if err := conn.WriteMessage(messageType, p); err != nil {
            log.Println("write error:", err)
            return
        }
    }
}

func main() {
    http.HandleFunc("/ws", wsHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}`)
	// ============================================
	// NHOOYR/WEBSOCKET
	// ============================================
	fmt.Println("\n--- nhooyr/websocket (alternativa moderna) ---")
	fmt.Println(`
INSTALACIÓN:
go get nhooyr.io/websocket

VENTAJAS:
- API más moderna y limpia
- Mejor soporte para context
- io.Reader/Writer estándar
- Menos allocaciones

SERVIDOR:

import "nhooyr.io/websocket"

func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
        OriginPatterns: []string{"*"},
    })
    if err != nil {
        return
    }
    defer conn.Close(websocket.StatusNormalClosure, "")

    ctx := r.Context()
    for {
        // Leer
        typ, data, err := conn.Read(ctx)
        if err != nil {
            return
        }

        // Echo
        err = conn.Write(ctx, typ, data)
        if err != nil {
            return
        }
    }
}`)
	// ============================================
	// CLIENTE WEBSOCKET
	// ============================================
	fmt.Println("\n--- Cliente WebSocket ---")
	fmt.Println(`
GORILLA CLIENTE:

import "github.com/gorilla/websocket"

func connectWS() {
    url := "ws://localhost:8080/ws"

    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Fatal("dial error:", err)
    }
    defer conn.Close()

    // Enviar mensaje
    err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
    if err != nil {
        log.Println("write error:", err)
        return
    }

    // Leer respuesta
    _, message, err := conn.ReadMessage()
    if err != nil {
        log.Println("read error:", err)
        return
    }
    fmt.Println("received:", string(message))
}

CON HEADERS:

header := http.Header{}
header.Add("Authorization", "Bearer "+token)

conn, _, err := websocket.DefaultDialer.Dial(url, header)`)
	// ============================================
	// JSON MESSAGING
	// ============================================
	fmt.Println("\n--- JSON Messaging ---")
	fmt.Println(`
MENSAJES ESTRUCTURADOS:

type Message struct {
    Type    string          ` + "`" + `json:"type"` + "`" + `
    Payload json.RawMessage ` + "`" + `json:"payload"` + "`" + `
}

type ChatMessage struct {
    User    string ` + "`" + `json:"user"` + "`" + `
    Content string ` + "`" + `json:"content"` + "`" + `
    Time    int64  ` + "`" + `json:"time"` + "`" + `
}

// Leer JSON
func readJSON(conn *websocket.Conn, v any) error {
    return conn.ReadJSON(v)
}

// Escribir JSON
func writeJSON(conn *websocket.Conn, v any) error {
    return conn.WriteJSON(v)
}

// Handler con tipos
func handleMessage(conn *websocket.Conn, msg Message) {
    switch msg.Type {
    case "chat":
        var chat ChatMessage
        json.Unmarshal(msg.Payload, &chat)
        // Procesar chat
    case "typing":
        // Indicador de escritura
    case "ping":
        conn.WriteJSON(Message{Type: "pong"})
    }
}`)
	// ============================================
	// CHAT ROOM (HUB PATTERN)
	// ============================================
	fmt.Println("\n--- Chat Room (Hub Pattern) ---")
	fmt.Println(`
ARQUITECTURA DE CHAT:

type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte, 256),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

CLIENT READ PUMP:

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            return
        }
        c.hub.broadcast <- message
    }
}

CLIENT WRITE PUMP:

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            c.conn.WriteMessage(websocket.TextMessage, message)

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}`)
	// ============================================
	// PING/PONG Y KEEPALIVE
	// ============================================
	fmt.Println("\n--- Ping/Pong y Keepalive ---")
	fmt.Println(`
MANTENER CONEXIÓN VIVA:

const (
    writeWait  = 10 * time.Second
    pongWait   = 60 * time.Second
    pingPeriod = (pongWait * 9) / 10
)

// En el servidor
func (c *Client) readPump() {
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })
    // ...
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := c.conn.WriteControl(
                websocket.PingMessage,
                []byte{},
                time.Now().Add(writeWait),
            ); err != nil {
                return
            }
        // ...
        }
    }
}

DETECTAR DESCONEXIÓN:
- Ping timeout indica conexión muerta
- Permite cleanup de recursos
- Evita goroutine leaks`)
	// ============================================
	// ROOMS Y CHANNELS
	// ============================================
	fmt.Println("\n--- Rooms y Channels ---")
	fmt.Println(`
MÚLTIPLES SALAS:

type Hub struct {
    rooms map[string]map[*Client]bool
    mu    sync.RWMutex
}

func (h *Hub) JoinRoom(client *Client, roomID string) {
    h.mu.Lock()
    defer h.mu.Unlock()

    if h.rooms[roomID] == nil {
        h.rooms[roomID] = make(map[*Client]bool)
    }
    h.rooms[roomID][client] = true
    client.rooms[roomID] = true
}

func (h *Hub) LeaveRoom(client *Client, roomID string) {
    h.mu.Lock()
    defer h.mu.Unlock()

    delete(h.rooms[roomID], client)
    delete(client.rooms, roomID)

    if len(h.rooms[roomID]) == 0 {
        delete(h.rooms, roomID)
    }
}

func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    for client := range h.rooms[roomID] {
        select {
        case client.send <- message:
        default:
            // Buffer lleno, desconectar cliente
        }
    }
}`)
	// ============================================
	// ESCALABILIDAD
	// ============================================
	fmt.Println("\n--- Escalabilidad ---")
	fmt.Println(`
MÚLTIPLES SERVIDORES:

┌─────────────────────────────────────────────┐
│              Load Balancer                   │
│         (sticky sessions o Redis)            │
└─────────────────┬───────────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
┌───▼───┐    ┌───▼───┐    ┌───▼───┐
│ Go 1  │    │ Go 2  │    │ Go 3  │
└───┬───┘    └───┬───┘    └───┬───┘
    │            │            │
    └────────────┼────────────┘
                 │
         ┌──────▼──────┐
         │    Redis    │
         │   Pub/Sub   │
         └─────────────┘

REDIS PUB/SUB:

import "github.com/redis/go-redis/v9"

type Hub struct {
    local  map[*Client]bool
    redis  *redis.Client
    pubsub *redis.PubSub
}

func (h *Hub) Broadcast(channel string, msg []byte) {
    // Publicar a Redis
    h.redis.Publish(ctx, channel, msg)
}

func (h *Hub) Subscribe(channel string) {
    h.pubsub = h.redis.Subscribe(ctx, channel)

    go func() {
        for msg := range h.pubsub.Channel() {
            // Enviar a clientes locales
            for client := range h.local {
                client.send <- []byte(msg.Payload)
            }
        }
    }()
}`)
	// ============================================
	// AUTENTICACIÓN
	// ============================================
	fmt.Println("\n--- Autenticación ---")
	fmt.Println(`
EN EL UPGRADE:

func wsHandler(w http.ResponseWriter, r *http.Request) {
    // Verificar token en query param
    token := r.URL.Query().Get("token")
    userID, err := validateToken(token)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }

    client := &Client{
        conn:   conn,
        userID: userID,
    }
    // ...
}

EN HEADERS:

func wsHandler(w http.ResponseWriter, r *http.Request) {
    auth := r.Header.Get("Authorization")
    token := strings.TrimPrefix(auth, "Bearer ")

    userID, err := validateToken(token)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // ...
}

PRIMER MENSAJE:

// Cliente envía auth como primer mensaje
func (c *Client) authenticate() error {
    c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

    var authMsg AuthMessage
    if err := c.conn.ReadJSON(&authMsg); err != nil {
        return err
    }

    userID, err := validateToken(authMsg.Token)
    if err != nil {
        return err
    }

    c.userID = userID
    c.conn.SetReadDeadline(time.Time{})  // Reset deadline
    return nil
}`)
	// ============================================
	// RATE LIMITING
	// ============================================
	fmt.Println("\n--- Rate Limiting ---")
	fmt.Println(`
LIMITAR MENSAJES POR CLIENTE:

type Client struct {
    conn     *websocket.Conn
    limiter  *rate.Limiter
    msgCount int
}

func NewClient(conn *websocket.Conn) *Client {
    return &Client{
        conn:    conn,
        limiter: rate.NewLimiter(10, 20),  // 10 msg/s, burst 20
    }
}

func (c *Client) readPump() {
    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            return
        }

        // Rate limit
        if !c.limiter.Allow() {
            c.conn.WriteJSON(map[string]string{
                "error": "rate limit exceeded",
            })
            continue
        }

        c.hub.broadcast <- message
    }
}

LIMITAR TAMAÑO:

conn.SetReadLimit(4096)  // Max 4KB por mensaje`)
	// ============================================
	// COMPRESIÓN
	// ============================================
	fmt.Println("\n--- Compresión ---")
	fmt.Println(`
GORILLA CON COMPRESIÓN:

var upgrader = websocket.Upgrader{
    EnableCompression: true,
}

// Por mensaje
conn.EnableWriteCompression(true)
conn.SetCompressionLevel(flate.BestSpeed)

NHOOYR CON COMPRESIÓN:

conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
    CompressionMode: websocket.CompressionContextTakeover,
})

CUÁNDO USAR:
- Mensajes de texto grandes
- Alta frecuencia de mensajes
- Ancho de banda limitado

CUÁNDO NO USAR:
- Mensajes pequeños (overhead > ahorro)
- CPU limitada
- Ya comprimidos (imágenes, video)`)
	// ============================================
	// TESTING
	// ============================================
	fmt.Println("\n--- Testing ---")
	fmt.Println(`
TEST CON HTTPTEST:

func TestWebSocket(t *testing.T) {
    // Crear servidor de prueba
    hub := NewHub()
    go hub.Run()

    server := httptest.NewServer(http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            serveWs(hub, w, r)
        },
    ))
    defer server.Close()

    // Convertir http:// a ws://
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

    // Conectar cliente
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    require.NoError(t, err)
    defer conn.Close()

    // Enviar mensaje
    err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
    require.NoError(t, err)

    // Leer respuesta
    _, msg, err := conn.ReadMessage()
    require.NoError(t, err)
    assert.Equal(t, "hello", string(msg))
}`)}

/*
RESUMEN DE WEBSOCKETS:

LIBRERÍAS:
- gorilla/websocket: estándar de facto, battle-tested
- nhooyr/websocket: API moderna, context-aware

SERVIDOR BÁSICO:
var upgrader = websocket.Upgrader{}

func handler(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()

    for {
        _, msg, _ := conn.ReadMessage()
        conn.WriteMessage(websocket.TextMessage, msg)
    }
}

CLIENTE:
conn, _, _ := websocket.DefaultDialer.Dial(url, nil)
conn.WriteMessage(websocket.TextMessage, []byte("hi"))
_, msg, _ := conn.ReadMessage()

PATRONES:
1. Echo server (básico)
2. Chat con Hub (broadcast)
3. Rooms/Channels (múltiples grupos)
4. Pub/Sub con Redis (escalabilidad)

KEEPALIVE:
- Ping/Pong frames
- Read/Write deadlines
- Detectar desconexiones

SEGURIDAD:
- CheckOrigin en upgrader
- Autenticación (token, header, primer mensaje)
- Rate limiting
- Message size limits

ESCALABILIDAD:
- Múltiples instancias con sticky sessions
- Redis Pub/Sub para broadcast entre servidores
- Connection pooling

COMPRESIÓN:
- EnableCompression para mensajes grandes
- Trade-off CPU vs bandwidth
*/
