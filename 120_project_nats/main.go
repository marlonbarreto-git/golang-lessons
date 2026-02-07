// Package main - Chapter 120: NATS - Cloud Native Messaging System in Go
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 120: NATS - CLOUD NATIVE MESSAGING SYSTEM IN GO ===")

	// ============================================
	// WHAT IS NATS?
	// ============================================
	fmt.Println(`
==============================
WHAT IS NATS?
==============================

NATS is a high-performance, cloud-native messaging system
written in Go. It provides:

  - Pub/Sub messaging (fire and forget)
  - Request/Reply patterns
  - Queue groups for load balancing
  - JetStream for persistence and exactly-once delivery
  - Subject-based addressing with wildcards
  - At-most-once (core) and at-least-once (JetStream) delivery

WHO USES IT:
  - Synadia (creators, cloud offering)
  - Mastercard (event-driven architecture)
  - Netlify (edge deployment coordination)
  - Clarifai (ML pipeline messaging)
  - Used in CNCF projects, IoT, edge computing

WHY GO:
  NATS was one of the first major Go projects. Go's goroutine
  model maps perfectly to handling thousands of concurrent
  connections. The server is a single ~15MB binary that can
  handle millions of messages per second.`)

	// ============================================
	// ARCHITECTURE OVERVIEW
	// ============================================
	fmt.Println(`
==============================
ARCHITECTURE OVERVIEW
==============================

  +------------------+     +------------------+
  |   Publisher       |     |   Subscriber     |
  |   (client)        |     |   (client)        |
  +--------+---------+     +--------+---------+
           |                         ^
           | PUB foo "hello"         | MSG foo "hello"
           v                         |
  +--------+-------------------------+---------+
  |                 NATS Server                 |
  |                                             |
  |  +-----------+  +-----------+  +---------+ |
  |  | Subject   |  | Client    |  | Route   | |
  |  | Trie      |  | Manager   |  | Manager | |
  |  +-----------+  +-----------+  +---------+ |
  |                                             |
  |  +------------------+  +-----------------+ |
  |  | JetStream Engine |  | Auth/TLS        | |
  |  | (persistence)    |  | (security)      | |
  |  +------------------+  +-----------------+ |
  +---------------------------------------------+

KEY CONCEPTS:
  - Subjects: hierarchical topic names (e.g., "orders.us.new")
  - Wildcards: * (single token), > (multi-token)
  - Queue Groups: competing consumers for load balancing
  - Leaf Nodes: extend clusters to edge locations
  - JetStream: built-in persistence with streams and consumers`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
==============================
KEY GO PATTERNS IN NATS
==============================

1. GOROUTINE-PER-CONNECTION
   Each client connection gets its own read/write goroutines.
   Channels coordinate between them.

2. ZERO-ALLOCATION PARSING
   The NATS protocol parser is designed to minimize allocations.
   It uses byte slices and avoids string conversions in hot paths.

3. SUBJECT TRIE
   A radix/trie structure for fast subject matching with
   wildcard support. O(k) lookup where k is subject depth.

4. INTEREST-BASED ROUTING
   Messages only flow to nodes with active subscriptions.
   No flooding, no wasted bandwidth.

5. SINGLE-THREADED CORE
   Despite Go's concurrency, the server core uses a
   single-goroutine event loop per client for ordering
   guarantees, with channels for coordination.

6. LEAFNODE ARCHITECTURE
   Edge nodes connect to hub clusters, extending the
   subject space transparently.`)

	// ============================================
	// DEMO: SIMPLIFIED NATS SERVER
	// ============================================
	fmt.Println(`
==============================
DEMO: SIMPLIFIED NATS SERVER
==============================

This demo implements core NATS concepts:
  - Subject-based pub/sub
  - Wildcard subscriptions (* and >)
  - Queue groups for load balancing
  - Request/Reply pattern
  - JetStream-like persistence
  - Subject trie for fast matching`)

	fmt.Println("\n--- Subject Trie ---")
	demoSubjectTrie()

	fmt.Println("\n--- Pub/Sub Messaging ---")
	demoPubSub()

	fmt.Println("\n--- Queue Groups ---")
	demoQueueGroups()

	fmt.Println("\n--- Request/Reply ---")
	demoRequestReply()

	fmt.Println("\n--- JetStream (Simplified) ---")
	demoJetStream()

	fmt.Println("\n--- Full Mini-NATS Server ---")
	demoMiniNATS()

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
==============================
PERFORMANCE TECHNIQUES
==============================

1. PROTOCOL PARSING
   NATS uses a simple text protocol (PUB, SUB, MSG):
   - Hand-written parser, no reflection
   - Reads directly from net.Conn buffers
   - Minimizes copies between read and dispatch

2. WRITE COALESCING
   - Batches multiple messages into single TCP writes
   - Uses bufio.Writer with flush intervals
   - Dramatically reduces syscall overhead

3. FAST SUBJECT MATCHING
   - Trie-based subject tree
   - Cached results for common subjects
   - Bloom filter for quick negative matches

4. CONNECTION MANAGEMENT
   - Efficient readloop/writeloop goroutine pairs
   - Minimal per-connection memory overhead
   - Fast client pruning on slow consumers

5. CLUSTER ROUTING
   - Interest-graph propagation
   - Route messages only where needed
   - Avoid duplicate delivery across routes`)

	// ============================================
	// GO PHILOSOPHY IN NATS
	// ============================================
	fmt.Println(`
==============================
GO PHILOSOPHY IN NATS
==============================

SIMPLICITY IS THE ULTIMATE SOPHISTICATION:
  NATS embodies Go's philosophy of simplicity.
  The entire server is a single binary. The protocol
  is human-readable text. Configuration is minimal.

DO LESS, ENABLE MORE:
  Core NATS provides simple pub/sub. JetStream adds
  persistence when needed. Users compose features
  rather than getting a monolithic system.

SHARE MEMORY BY COMMUNICATING:
  The NATS server uses channels extensively for
  internal coordination between goroutines handling
  different clients and routes.

ERRORS ARE VALUES:
  NATS propagates errors cleanly through the stack,
  with specific error types for different failure modes
  (slow consumer, authorization, protocol error).

COMPOSITION:
  Server components (auth, TLS, JetStream, clustering)
  are composed through interfaces, enabling clean
  separation and testing.`)

	fmt.Println("\n=== END OF CHAPTER 120 ===")
}

// ============================================
// SUBJECT TRIE
// ============================================

type SubjectTrie struct {
	mu   sync.RWMutex
	root *trieNode
}

type trieNode struct {
	children map[string]*trieNode
	subs     []*Subscription
}

type Subscription struct {
	ID      int
	Subject string
	Queue   string
	Handler func(subject string, data []byte)
}

func NewSubjectTrie() *SubjectTrie {
	return &SubjectTrie{
		root: &trieNode{children: make(map[string]*trieNode)},
	}
}

func (t *SubjectTrie) Insert(sub *Subscription) {
	t.mu.Lock()
	defer t.mu.Unlock()

	tokens := strings.Split(sub.Subject, ".")
	node := t.root

	for _, token := range tokens {
		if node.children == nil {
			node.children = make(map[string]*trieNode)
		}
		child, ok := node.children[token]
		if !ok {
			child = &trieNode{children: make(map[string]*trieNode)}
			node.children[token] = child
		}
		node = child
	}
	node.subs = append(node.subs, sub)
}

func (t *SubjectTrie) Match(subject string) []*Subscription {
	t.mu.RLock()
	defer t.mu.RUnlock()

	tokens := strings.Split(subject, ".")
	var results []*Subscription
	t.matchRecursive(t.root, tokens, 0, &results)
	return results
}

func (t *SubjectTrie) matchRecursive(node *trieNode, tokens []string, depth int, results *[]*Subscription) {
	if depth == len(tokens) {
		*results = append(*results, node.subs...)
		return
	}

	token := tokens[depth]

	if child, ok := node.children[token]; ok {
		t.matchRecursive(child, tokens, depth+1, results)
	}

	if child, ok := node.children["*"]; ok {
		t.matchRecursive(child, tokens, depth+1, results)
	}

	if child, ok := node.children[">"]; ok {
		*results = append(*results, child.subs...)
	}
}

func (t *SubjectTrie) Remove(subID int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.removeRecursive(t.root, subID)
}

func (t *SubjectTrie) removeRecursive(node *trieNode, subID int) {
	filtered := node.subs[:0]
	for _, sub := range node.subs {
		if sub.ID != subID {
			filtered = append(filtered, sub)
		}
	}
	node.subs = filtered

	for _, child := range node.children {
		t.removeRecursive(child, subID)
	}
}

func demoSubjectTrie() {
	trie := NewSubjectTrie()

	trie.Insert(&Subscription{ID: 1, Subject: "orders.us.new"})
	trie.Insert(&Subscription{ID: 2, Subject: "orders.*.new"})
	trie.Insert(&Subscription{ID: 3, Subject: "orders.>"})
	trie.Insert(&Subscription{ID: 4, Subject: "orders.eu.cancel"})

	tests := []string{
		"orders.us.new",
		"orders.eu.new",
		"orders.eu.cancel",
		"orders.us.update",
	}

	for _, subject := range tests {
		matches := trie.Match(subject)
		ids := make([]int, len(matches))
		for i, m := range matches {
			ids[i] = m.ID
		}
		fmt.Printf("  Match '%s': sub IDs %v\n", subject, ids)
	}
}

// ============================================
// PUB/SUB MESSAGING
// ============================================

type NATSServer struct {
	mu      sync.RWMutex
	trie    *SubjectTrie
	nextSub int
	clients map[string]*NATSClient
}

type NATSClient struct {
	ID       string
	Server   *NATSServer
	Inbox    chan Message
	subIDs   []int
}

type Message struct {
	Subject string
	Reply   string
	Data    []byte
}

func NewNATSServer() *NATSServer {
	return &NATSServer{
		trie:    NewSubjectTrie(),
		clients: make(map[string]*NATSClient),
	}
}

func (s *NATSServer) Connect(clientID string) *NATSClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := &NATSClient{
		ID:     clientID,
		Server: s,
		Inbox:  make(chan Message, 100),
	}
	s.clients[clientID] = c
	return c
}

func (c *NATSClient) Subscribe(subject string) int {
	c.Server.mu.Lock()
	c.Server.nextSub++
	subID := c.Server.nextSub
	c.Server.mu.Unlock()

	inbox := c.Inbox
	sub := &Subscription{
		ID:      subID,
		Subject: subject,
		Handler: func(subj string, data []byte) {
			select {
			case inbox <- Message{Subject: subj, Data: data}:
			default:
			}
		},
	}

	c.Server.trie.Insert(sub)
	c.subIDs = append(c.subIDs, subID)
	return subID
}

func (c *NATSClient) SubscribeQueue(subject, queue string) int {
	c.Server.mu.Lock()
	c.Server.nextSub++
	subID := c.Server.nextSub
	c.Server.mu.Unlock()

	inbox := c.Inbox
	sub := &Subscription{
		ID:      subID,
		Subject: subject,
		Queue:   queue,
		Handler: func(subj string, data []byte) {
			select {
			case inbox <- Message{Subject: subj, Data: data}:
			default:
			}
		},
	}

	c.Server.trie.Insert(sub)
	c.subIDs = append(c.subIDs, subID)
	return subID
}

func (c *NATSClient) Publish(subject string, data []byte) {
	subs := c.Server.trie.Match(subject)
	if len(subs) == 0 {
		return
	}

	queues := make(map[string][]*Subscription)
	var direct []*Subscription
	for _, sub := range subs {
		if sub.Queue != "" {
			queues[sub.Queue] = append(queues[sub.Queue], sub)
		} else {
			direct = append(direct, sub)
		}
	}

	for _, sub := range direct {
		sub.Handler(subject, data)
	}

	for _, qsubs := range queues {
		idx := time.Now().UnixNano() % int64(len(qsubs))
		qsubs[idx].Handler(subject, data)
	}
}

func (c *NATSClient) Request(subject string, data []byte, timeout time.Duration) (Message, bool) {
	replySubject := fmt.Sprintf("_INBOX.%s.%d", c.ID, time.Now().UnixNano())

	replyCh := make(chan Message, 1)

	c.Server.mu.Lock()
	c.Server.nextSub++
	subID := c.Server.nextSub
	c.Server.mu.Unlock()

	sub := &Subscription{
		ID:      subID,
		Subject: replySubject,
		Handler: func(subj string, data []byte) {
			select {
			case replyCh <- Message{Subject: subj, Data: data}:
			default:
			}
		},
	}
	c.Server.trie.Insert(sub)
	defer c.Server.trie.Remove(subID)

	subs := c.Server.trie.Match(subject)
	for _, s := range subs {
		s.Handler(subject, data)
	}

	select {
	case msg := <-replyCh:
		return msg, true
	case <-time.After(timeout):
		return Message{}, false
	}
}

func demoPubSub() {
	server := NewNATSServer()

	sub1 := server.Connect("subscriber-1")
	sub2 := server.Connect("subscriber-2")
	pub := server.Connect("publisher")

	sub1.Subscribe("events.user.>")
	sub2.Subscribe("events.user.login")

	pub.Publish("events.user.login", []byte(`{"user":"alice"}`))
	pub.Publish("events.user.logout", []byte(`{"user":"bob"}`))

	time.Sleep(10 * time.Millisecond)

	count1, count2 := 0, 0
	for {
		select {
		case msg := <-sub1.Inbox:
			count1++
			fmt.Printf("  Sub1 received: %s -> %s\n", msg.Subject, msg.Data)
		default:
			goto doneSub1
		}
	}
doneSub1:
	for {
		select {
		case msg := <-sub2.Inbox:
			count2++
			fmt.Printf("  Sub2 received: %s -> %s\n", msg.Subject, msg.Data)
		default:
			goto doneSub2
		}
	}
doneSub2:
	fmt.Printf("  Sub1 got %d msgs (wildcard >), Sub2 got %d msgs (exact)\n", count1, count2)
}

// ============================================
// QUEUE GROUPS
// ============================================

func demoQueueGroups() {
	server := NewNATSServer()

	workers := make([]*NATSClient, 3)
	for i := range workers {
		workers[i] = server.Connect(fmt.Sprintf("worker-%d", i))
		workers[i].SubscribeQueue("tasks.process", "worker-group")
	}

	pub := server.Connect("producer")
	for i := 0; i < 9; i++ {
		pub.Publish("tasks.process", []byte(fmt.Sprintf(`{"task":%d}`, i)))
	}

	time.Sleep(10 * time.Millisecond)

	for i, w := range workers {
		count := 0
		for {
			select {
			case <-w.Inbox:
				count++
			default:
				goto doneWorker
			}
		}
	doneWorker:
		fmt.Printf("  Worker-%d processed %d tasks\n", i, count)
	}
}

// ============================================
// REQUEST/REPLY
// ============================================

func demoRequestReply() {
	server := NewNATSServer()

	service := server.Connect("math-service")
	service.Subscribe("math.add")

	go func() {
		for msg := range service.Inbox {
			var req struct{ A, B int }
			json.Unmarshal(msg.Data, &req)
			result := req.A + req.B
			resp, _ := json.Marshal(map[string]int{"result": result})

			replySubs := server.trie.Match(msg.Reply)
			for _, s := range replySubs {
				s.Handler(msg.Reply, resp)
			}
		}
	}()

	client := server.Connect("client")

	reqData, _ := json.Marshal(map[string]int{"A": 40, "B": 2})

	subs := server.trie.Match("math.add")
	if len(subs) > 0 {
		replySubject := fmt.Sprintf("_INBOX.%s.%d", client.ID, time.Now().UnixNano())

		replyCh := make(chan Message, 1)
		server.mu.Lock()
		server.nextSub++
		replySubID := server.nextSub
		server.mu.Unlock()

		replySub := &Subscription{
			ID:      replySubID,
			Subject: replySubject,
			Handler: func(subj string, data []byte) {
				replyCh <- Message{Subject: subj, Data: data}
			},
		}
		server.trie.Insert(replySub)

		for _, s := range subs {
			s.Handler("math.add", reqData)
			service.Inbox <- Message{Subject: "math.add", Reply: replySubject, Data: reqData}
			break
		}

		select {
		case reply := <-replyCh:
			fmt.Printf("  Request: 40 + 2 = ? -> Response: %s\n", reply.Data)
		case <-time.After(time.Second):
			fmt.Println("  Request timed out")
		}

		server.trie.Remove(replySubID)
	}
}

// ============================================
// JETSTREAM (SIMPLIFIED PERSISTENCE)
// ============================================

type StreamConfig struct {
	Name       string
	Subjects   []string
	MaxMsgs    int
	MaxAge     time.Duration
	Retention  string
}

type Stream struct {
	mu       sync.RWMutex
	Config   StreamConfig
	Messages []StoredMessage
	nextSeq  uint64
}

type StoredMessage struct {
	Sequence  uint64
	Subject   string
	Data      []byte
	Timestamp time.Time
}

type Consumer struct {
	mu          sync.Mutex
	Name        string
	Stream      *Stream
	AckPolicy   string
	DeliverSeq  uint64
	AckPending  map[uint64]bool
	MaxInflight int
}

type JetStreamManager struct {
	mu      sync.RWMutex
	streams map[string]*Stream
}

func NewJetStreamManager() *JetStreamManager {
	return &JetStreamManager{
		streams: make(map[string]*Stream),
	}
}

func (js *JetStreamManager) AddStream(config StreamConfig) *Stream {
	js.mu.Lock()
	defer js.mu.Unlock()

	s := &Stream{
		Config:  config,
		nextSeq: 1,
	}
	js.streams[config.Name] = s
	return s
}

func (s *Stream) Publish(subject string, data []byte) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	seq := s.nextSeq
	s.nextSeq++

	msg := StoredMessage{
		Sequence:  seq,
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now(),
	}
	s.Messages = append(s.Messages, msg)

	if s.Config.MaxMsgs > 0 && len(s.Messages) > s.Config.MaxMsgs {
		s.Messages = s.Messages[1:]
	}

	return seq
}

func (s *Stream) CreateConsumer(name string, ackPolicy string) *Consumer {
	return &Consumer{
		Name:        name,
		Stream:      s,
		AckPolicy:   ackPolicy,
		DeliverSeq:  1,
		AckPending:  make(map[uint64]bool),
		MaxInflight: 10,
	}
}

func (c *Consumer) Fetch(count int) []StoredMessage {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Stream.mu.RLock()
	defer c.Stream.mu.RUnlock()

	var msgs []StoredMessage
	for _, msg := range c.Stream.Messages {
		if msg.Sequence >= c.DeliverSeq && len(msgs) < count {
			if len(c.AckPending) < c.MaxInflight {
				msgs = append(msgs, msg)
				c.AckPending[msg.Sequence] = true
				c.DeliverSeq = msg.Sequence + 1
			}
		}
	}
	return msgs
}

func (c *Consumer) Ack(seq uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.AckPending, seq)
}

func demoJetStream() {
	js := NewJetStreamManager()

	stream := js.AddStream(StreamConfig{
		Name:     "ORDERS",
		Subjects: []string{"orders.>"},
		MaxMsgs:  1000,
	})

	for i := 0; i < 5; i++ {
		data, _ := json.Marshal(map[string]interface{}{
			"order_id": i + 1,
			"item":     fmt.Sprintf("product-%d", i+1),
			"quantity": i + 1,
		})
		seq := stream.Publish("orders.new", data)
		fmt.Printf("  Published order to seq %d\n", seq)
	}

	consumer := stream.CreateConsumer("order-processor", "explicit")

	msgs := consumer.Fetch(3)
	fmt.Printf("  Fetched %d messages\n", len(msgs))
	for _, msg := range msgs {
		fmt.Printf("    Seq %d: %s\n", msg.Sequence, msg.Data)
		consumer.Ack(msg.Sequence)
	}

	msgs2 := consumer.Fetch(5)
	fmt.Printf("  Fetched %d more messages\n", len(msgs2))
	for _, msg := range msgs2 {
		fmt.Printf("    Seq %d: %s\n", msg.Sequence, msg.Data)
		consumer.Ack(msg.Sequence)
	}
}

// ============================================
// FULL MINI-NATS SERVER
// ============================================

type MiniNATS struct {
	server    *NATSServer
	jetstream *JetStreamManager
	stats     ServerStats
	mu        sync.RWMutex
}

type ServerStats struct {
	MsgsSent     int64
	MsgsReceived int64
	BytesSent    int64
	BytesReceived int64
	Connections  int
	Subscriptions int
}

func NewMiniNATS() *MiniNATS {
	return &MiniNATS{
		server:    NewNATSServer(),
		jetstream: NewJetStreamManager(),
	}
}

func (n *MiniNATS) Connect(id string) *NATSClient {
	n.mu.Lock()
	n.stats.Connections++
	n.mu.Unlock()
	return n.server.Connect(id)
}

func (n *MiniNATS) Publish(client *NATSClient, subject string, data []byte) {
	n.mu.Lock()
	n.stats.MsgsSent++
	n.stats.BytesSent += int64(len(data))
	n.mu.Unlock()
	client.Publish(subject, data)
}

func (n *MiniNATS) Stats() ServerStats {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.stats
}

func demoMiniNATS() {
	nats := NewMiniNATS()

	stream := nats.jetstream.AddStream(StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"events.>"},
		MaxMsgs:  100,
	})

	pub := nats.Connect("event-publisher")
	sub1 := nats.Connect("event-logger")
	sub2 := nats.Connect("event-processor")

	sub1.Subscribe("events.>")
	sub2.Subscribe("events.user.*")

	events := []struct {
		subject string
		data    map[string]interface{}
	}{
		{"events.user.login", map[string]interface{}{"user": "alice", "ip": "10.0.0.1"}},
		{"events.user.purchase", map[string]interface{}{"user": "alice", "amount": 99.99}},
		{"events.system.health", map[string]interface{}{"cpu": 45.2, "mem": 62.1}},
		{"events.user.logout", map[string]interface{}{"user": "alice"}},
	}

	for _, evt := range events {
		data, _ := json.Marshal(evt.data)
		nats.Publish(pub, evt.subject, data)
		stream.Publish(evt.subject, data)
	}

	time.Sleep(10 * time.Millisecond)

	count1, count2 := 0, 0
	for {
		select {
		case <-sub1.Inbox:
			count1++
		default:
			goto d1
		}
	}
d1:
	for {
		select {
		case <-sub2.Inbox:
			count2++
		default:
			goto d2
		}
	}
d2:
	fmt.Printf("  Logger (events.>) received: %d messages\n", count1)
	fmt.Printf("  Processor (events.user.*) received: %d messages\n", count2)

	consumer := stream.CreateConsumer("replay", "explicit")
	replayMsgs := consumer.Fetch(10)
	fmt.Printf("  JetStream replay: %d messages available\n", len(replayMsgs))
	for _, msg := range replayMsgs {
		fmt.Printf("    [seq=%d] %s: %s\n", msg.Sequence, msg.Subject, msg.Data)
		consumer.Ack(msg.Sequence)
	}

	stats := nats.Stats()
	fmt.Printf("  Server stats: connections=%d, msgs_sent=%d, bytes=%d\n",
		stats.Connections, stats.MsgsSent, stats.BytesSent)
}

/*
SUMMARY - NATS: CLOUD NATIVE MESSAGING SYSTEM IN GO:

WHAT IS NATS:
- High-performance cloud-native messaging system written in Go
- Pub/Sub, Request/Reply, Queue Groups, JetStream persistence
- Subject-based addressing with wildcard support (* and >)
- Single ~15MB binary handling millions of messages per second

ARCHITECTURE:
- Subject Trie for fast O(k) topic matching with wildcards
- Goroutine-per-connection with read/write loop pairs
- JetStream engine for durable streams and consumers
- Interest-based routing: messages flow only where subscribed

KEY GO PATTERNS DEMONSTRATED:
- Subject Trie: radix tree for efficient wildcard matching
- Pub/Sub: direct subscribers and wildcard subscriptions
- Queue Groups: competing consumers for load balancing
- Request/Reply: synchronous RPC over async messaging
- JetStream: persistent streams with consumer ack tracking

PERFORMANCE TECHNIQUES:
- Zero-allocation protocol parsing with byte slices
- Write coalescing: batch multiple messages into single TCP writes
- Cached subject matching results with bloom filters
- Minimal per-connection memory overhead
- Interest-graph propagation for cluster routing

GO PHILOSOPHY:
- Simplicity: human-readable text protocol, minimal configuration
- Channels for internal coordination between goroutines
- Specific error types for different failure modes
- Composition: core pub/sub + optional JetStream persistence
*/
