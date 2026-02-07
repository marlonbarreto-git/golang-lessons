// Package main - Chapter 091: Message Queues
// Los message queues son fundamentales en sistemas distribuidos. Aprenderás
// Kafka, NATS, RabbitMQ, patrones producer/consumer y garantías de entrega.
package main

import (
	"os"
	"fmt"
	"time"
)

func main() {
	fmt.Println("=== MESSAGE QUEUES EN GO ===")

	// ============================================
	// KAFKA CON SEGMENTIO/KAFKA-GO
	// ============================================
	fmt.Println("\n--- Kafka con segmentio/kafka-go ---")
	os.Stdout.WriteString(`
INSTALACIÓN:
go get github.com/segmentio/kafka-go

CARACTERÍSTICAS:
- Pure Go (no requiere librdkafka)
- API simple e idiomática
- Soporte para consumer groups
- Balanceo automático

PRODUCER BÁSICO:

import "github.com/segmentio/kafka-go"

func NewKafkaWriter(brokers []string, topic string) *kafka.Writer {
    return &kafka.Writer{
        Addr:     kafka.TCP(brokers...),
        Topic:    topic,
        Balancer: &kafka.LeastBytes{},

        // Performance tuning
        BatchSize:    100,
        BatchTimeout: 10 * time.Millisecond,
        Compression:  kafka.Snappy,

        // Garantías de entrega
        RequiredAcks: kafka.RequireAll,  // -1: all in-sync replicas
        Async:        false,              // Sync writes
    }
}

func ProduceMessage(ctx context.Context, w *kafka.Writer, msg Message) error {
    return w.WriteMessages(ctx, kafka.Message{
        Key:   []byte(msg.ID),
        Value: msg.Payload,
        Headers: []kafka.Header{
            {Key: "type", Value: []byte(msg.Type)},
            {Key: "timestamp", Value: []byte(time.Now().String())},
        },
    })
}

// Batch produce
func ProduceBatch(ctx context.Context, w *kafka.Writer, msgs []Message) error {
    kafkaMsgs := make([]kafka.Message, len(msgs))
    for i, msg := range msgs {
        kafkaMsgs[i] = kafka.Message{
            Key:   []byte(msg.ID),
            Value: msg.Payload,
        }
    }
    return w.WriteMessages(ctx, kafkaMsgs...)
}

CONSUMER BÁSICO:

func NewKafkaReader(brokers []string, topic, groupID string) *kafka.Reader {
    return kafka.NewReader(kafka.ReaderConfig{
        Brokers: brokers,
        Topic:   topic,
        GroupID: groupID,

        // Configuración
        MinBytes: 10e3,  // 10KB
        MaxBytes: 10e6,  // 10MB
        MaxWait:  500 * time.Millisecond,

        // Auto commit
        CommitInterval: time.Second,

        // Partition assignment
        GroupBalancers: []kafka.GroupBalancer{
            kafka.RangeGroupBalancer{},
        },

        // Start from
        StartOffset: kafka.LastOffset,  // or kafka.FirstOffset
    })
}

func ConsumeMessages(ctx context.Context, r *kafka.Reader) {
    for {
        msg, err := r.ReadMessage(ctx)
        if err != nil {
            if err == context.Canceled {
                break
            }
            log.Printf("Error reading: %v", err)
            continue
        }

        log.Printf("Topic: %s, Partition: %d, Offset: %d\n",
            msg.Topic, msg.Partition, msg.Offset)
        log.Printf("Key: %s, Value: %s\n", msg.Key, msg.Value)

        // Procesar mensaje
        if err := processMessage(msg); err != nil {
            log.Printf("Error processing: %v", err)
            // Manejo de error: DLQ, retry, etc
        }
    }
}

CONSUMER GROUP CON COMMIT MANUAL:

r := kafka.NewReader(kafka.ReaderConfig{
    Brokers:  brokers,
    Topic:    topic,
    GroupID:  groupID,

    // Commit manual
    CommitInterval: 0,  // Disable auto-commit
})

for {
    msg, err := r.FetchMessage(ctx)
    if err != nil {
        break
    }

    // Procesar
    if err := processMessage(msg); err != nil {
        log.Printf("Error: %v", err)
        continue  // No hacer commit
    }

    // Commit manual solo si procesado correctamente
    if err := r.CommitMessages(ctx, msg); err != nil {
        log.Printf("Commit error: %v", err)
    }
}

ADMIN OPERATIONS:

import "github.com/segmentio/kafka-go"

// Crear topic
conn, _ := kafka.Dial("tcp", "localhost:9092")
defer conn.Close()

topicConfig := kafka.TopicConfig{
    Topic:             "my-topic",
    NumPartitions:     10,
    ReplicationFactor: 3,
    ConfigEntries: []kafka.ConfigEntry{
        {ConfigName: "retention.ms", ConfigValue: "604800000"},  // 7 days
    },
}
conn.CreateTopics(topicConfig)

// Listar topics
partitions, _ := conn.ReadPartitions()
topics := make(map[string]struct{})
for _, p := range partitions {
    topics[p.Topic] = struct{}{}
}
`)

	// ============================================
	// NATS
	// ============================================
	fmt.Println("\n--- NATS con nats.io/nats.go ---")
	os.Stdout.WriteString(`
INSTALACIÓN:
go get github.com/nats-io/nats.go

CARACTERÍSTICAS:
- Ultra rápido y ligero
- Pub/Sub simple
- Request/Reply pattern
- JetStream para persistencia

CONEXIÓN BÁSICA:

import "github.com/nats-io/nats.go"

func ConnectNATS(url string) (*nats.Conn, error) {
    return nats.Connect(url,
        nats.Name("my-service"),
        nats.MaxReconnects(10),
        nats.ReconnectWait(2*time.Second),
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            log.Printf("Disconnected: %v", err)
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            log.Printf("Reconnected to %s", nc.ConnectedUrl())
        }),
    )
}

PUB/SUB:

// Publisher
func Publish(nc *nats.Conn, subject string, data []byte) error {
    return nc.Publish(subject, data)
}

// Subscriber
func Subscribe(nc *nats.Conn, subject string) {
    nc.Subscribe(subject, func(msg *nats.Msg) {
        log.Printf("Received on %s: %s", msg.Subject, string(msg.Data))
        processMessage(msg.Data)
    })
}

// Queue groups (load balancing)
nc.QueueSubscribe("tasks", "workers", func(msg *nats.Msg) {
    // Solo un worker en el grupo procesa el mensaje
    processTask(msg.Data)
})

REQUEST/REPLY PATTERN:

// Responder (servidor)
nc.Subscribe("service.request", func(msg *nats.Msg) {
    result := processRequest(msg.Data)
    msg.Respond(result)
})

// Requester (cliente)
msg, err := nc.Request("service.request", []byte("data"), 2*time.Second)
if err != nil {
    log.Printf("Request failed: %v", err)
    return
}
log.Printf("Response: %s", msg.Data)

JETSTREAM (PERSISTENCIA):

import "github.com/nats-io/nats.go"

func SetupJetStream(nc *nats.Conn) (nats.JetStreamContext, error) {
    js, err := nc.JetStream()
    if err != nil {
        return nil, err
    }

    // Crear stream
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "ORDERS",
        Subjects: []string{"orders.*"},
        Storage:  nats.FileStorage,
        MaxAge:   7 * 24 * time.Hour,
        Replicas: 3,
    })

    return js, err
}

// Publish con ACK
func JSPublish(js nats.JetStreamContext, subject string, data []byte) error {
    ack, err := js.Publish(subject, data)
    if err != nil {
        return err
    }
    log.Printf("Published to stream %s, seq %d", ack.Stream, ack.Sequence)
    return nil
}

// Subscribe con ACK manual
func JSSubscribe(js nats.JetStreamContext, subject string) {
    js.Subscribe(subject, func(msg *nats.Msg) {
        log.Printf("Received: %s", string(msg.Data))

        if err := processMessage(msg.Data); err != nil {
            msg.Nak()  // Negative acknowledgment
            return
        }

        msg.Ack()  // Acknowledge
    }, nats.Durable("my-consumer"), nats.ManualAck())
}

// Pull consumer (control de flujo)
sub, _ := js.PullSubscribe("orders.*", "order-processor")
for {
    msgs, _ := sub.Fetch(10, nats.MaxWait(5*time.Second))
    for _, msg := range msgs {
        processMessage(msg.Data)
        msg.Ack()
    }
}

KEY-VALUE STORE:

kv, _ := js.CreateKeyValue(&nats.KeyValueConfig{
    Bucket: "configs",
    TTL:    time.Hour,
})

// Put
kv.Put("feature.enabled", []byte("true"))

// Get
entry, _ := kv.Get("feature.enabled")
fmt.Println(string(entry.Value()))

// Watch changes
watcher, _ := kv.Watch("feature.*")
defer watcher.Stop()

for entry := range watcher.Updates() {
    fmt.Printf("Key %s changed to %s\n", entry.Key(), string(entry.Value()))
}
`)

	// ============================================
	// RABBITMQ
	// ============================================
	fmt.Println("\n--- RabbitMQ con streadway/amqp ---")
	os.Stdout.WriteString(`
INSTALACIÓN:
go get github.com/streadway/amqp

CARACTERÍSTICAS:
- AMQP protocol
- Exchanges, queues, bindings
- Message routing avanzado
- Múltiples patrones

CONEXIÓN Y SETUP:

import "github.com/streadway/amqp"

func ConnectRabbitMQ(url string) (*amqp.Connection, *amqp.Channel, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, nil, err
    }

    ch, err := conn.Channel()
    if err != nil {
        conn.Close()
        return nil, nil, err
    }

    return conn, ch, nil
}

// Declarar exchange
err := ch.ExchangeDeclare(
    "logs",    // name
    "fanout",  // type: direct, fanout, topic, headers
    true,      // durable
    false,     // auto-deleted
    false,     // internal
    false,     // no-wait
    nil,       // arguments
)

// Declarar queue
q, err := ch.QueueDeclare(
    "tasks",  // name
    true,     // durable
    false,    // delete when unused
    false,    // exclusive
    false,    // no-wait
    nil,      // arguments
)

// Bind queue a exchange
err = ch.QueueBind(
    q.Name,        // queue name
    "task.high",   // routing key
    "tasks",       // exchange
    false,
    nil,
)

PRODUCER:

func Publish(ch *amqp.Channel, exchange, routingKey string, body []byte) error {
    return ch.Publish(
        exchange,
        routingKey,
        false,  // mandatory
        false,  // immediate
        amqp.Publishing{
            ContentType:  "application/json",
            Body:         body,
            DeliveryMode: amqp.Persistent,  // Persist to disk
            Timestamp:    time.Now(),
            MessageId:    uuid.NewString(),
        },
    )
}

// Publisher confirms (garantía de entrega)
func PublishWithConfirm(ch *amqp.Channel, exchange, key string, body []byte) error {
    // Enable confirms
    if err := ch.Confirm(false); err != nil {
        return err
    }

    confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

    err := ch.Publish(exchange, key, false, false, amqp.Publishing{
        Body: body,
    })
    if err != nil {
        return err
    }

    // Esperar confirmación
    select {
    case confirm := <-confirms:
        if !confirm.Ack {
            return fmt.Errorf("message not confirmed")
        }
    case <-time.After(5 * time.Second):
        return fmt.Errorf("timeout waiting for confirmation")
    }

    return nil
}

CONSUMER:

func Consume(ch *amqp.Channel, queueName string) error {
    // Set QoS (prefetch)
    err := ch.Qos(
        10,    // prefetch count
        0,     // prefetch size
        false, // global
    )
    if err != nil {
        return err
    }

    msgs, err := ch.Consume(
        queueName,
        "",     // consumer tag
        false,  // auto-ack (manual ack)
        false,  // exclusive
        false,  // no-local
        false,  // no-wait
        nil,    // args
    )
    if err != nil {
        return err
    }

    for msg := range msgs {
        log.Printf("Received: %s", msg.Body)

        if err := processMessage(msg.Body); err != nil {
            log.Printf("Error: %v", err)
            msg.Nack(false, true)  // Requeue
            continue
        }

        msg.Ack(false)  // Acknowledge
    }

    return nil
}

PATRONES DE ROUTING:

1. DIRECT (routing key exacta):

ch.ExchangeDeclare("direct_logs", "direct", true, false, false, false, nil)
ch.Publish("direct_logs", "error", false, false, amqp.Publishing{Body: []byte("Error!")})

2. FANOUT (broadcast):

ch.ExchangeDeclare("logs", "fanout", true, false, false, false, nil)
ch.Publish("logs", "", false, false, amqp.Publishing{Body: []byte("Log message")})

3. TOPIC (pattern matching):

ch.ExchangeDeclare("topic_logs", "topic", true, false, false, false, nil)
ch.QueueBind(queue, "user.*.created", "topic_logs", false, nil)  // * = one word
ch.QueueBind(queue, "order.#", "topic_logs", false, nil)         // # = zero or more

4. HEADERS (metadata routing):

ch.ExchangeDeclare("headers_ex", "headers", true, false, false, false, nil)
ch.QueueBind(queue, "", "headers_ex", false, amqp.Table{
    "x-match": "all",
    "format":  "json",
    "type":    "report",
})

DEAD LETTER QUEUE:

// Declarar DLQ
dlq, _ := ch.QueueDeclare("tasks.dlq", true, false, false, false, nil)

// Queue principal con DLX
q, _ := ch.QueueDeclare(
    "tasks",
    true,
    false,
    false,
    false,
    amqp.Table{
        "x-dead-letter-exchange":    "dlx",
        "x-dead-letter-routing-key": "tasks.dlq",
        "x-message-ttl":             30000,  // 30s
    },
)
`)

	// ============================================
	// PRODUCER/CONSUMER PATTERNS
	// ============================================
	fmt.Println("\n--- Producer/Consumer Patterns ---")
	os.Stdout.WriteString(`
1. WORK QUEUE (Competing Consumers):

// Producer
for i := 0; i < 100; i++ {
    producer.Send(ctx, Task{ID: i})
}

// Multiple consumers
for w := 0; w < 5; w++ {
    go func(workerID int) {
        for msg := range consumer.Messages() {
            log.Printf("Worker %d processing %s", workerID, msg.ID)
            processTask(msg)
            consumer.Ack(msg)
        }
    }(w)
}

2. PUB/SUB (Broadcasting):

// Publisher
publisher.Publish(ctx, "notifications", Event{Type: "user.created"})

// Multiple subscribers
subscriber1.Subscribe(ctx, "notifications", handleNotification)
subscriber2.Subscribe(ctx, "notifications", handleNotification)
// Ambos reciben el mismo mensaje

3. REQUEST/REPLY (RPC):

// Client
func CallRPC(ctx context.Context, req Request) (Response, error) {
    replyTo := uuid.NewString()

    // Crear queue temporal para respuesta
    ch.QueueDeclare(replyTo, false, true, true, false, nil)

    // Subscribe a respuestas
    msgs, _ := ch.Consume(replyTo, "", true, false, false, false, nil)

    // Publicar request
    ch.Publish("", "rpc_queue", false, false, amqp.Publishing{
        CorrelationId: replyTo,
        ReplyTo:       replyTo,
        Body:          marshal(req),
    })

    // Esperar respuesta
    select {
    case msg := <-msgs:
        var resp Response
        unmarshal(msg.Body, &resp)
        return resp, nil
    case <-ctx.Done():
        return Response{}, ctx.Err()
    }
}

// Server
msgs, _ := ch.Consume("rpc_queue", "", false, false, false, false, nil)
for msg := range msgs {
    result := processRPC(msg.Body)

    ch.Publish("", msg.ReplyTo, false, false, amqp.Publishing{
        CorrelationId: msg.CorrelationId,
        Body:          marshal(result),
    })

    msg.Ack(false)
}

4. PRIORITY QUEUE:

// Declarar con prioridad
q, _ := ch.QueueDeclare("tasks", true, false, false, false, amqp.Table{
    "x-max-priority": 10,
})

// Publicar con prioridad
ch.Publish("", "tasks", false, false, amqp.Publishing{
    Priority: 9,  // 0-10
    Body:     []byte("high priority task"),
})

5. DELAYED/SCHEDULED MESSAGES:

// RabbitMQ delayed message plugin
ch.ExchangeDeclare("delayed", "x-delayed-message", true, false, false, false, amqp.Table{
    "x-delayed-type": "direct",
})

ch.Publish("delayed", "routing", false, false, amqp.Publishing{
    Headers: amqp.Table{
        "x-delay": 5000,  // 5 seconds
    },
    Body: []byte("delayed message"),
})
`)

	// ============================================
	// EXACTLY-ONCE DELIVERY
	// ============================================
	fmt.Println("\n--- Exactly-Once Delivery ---")
	os.Stdout.WriteString(`
PROBLEMA:
- At-most-once: puede perder mensajes
- At-least-once: puede duplicar mensajes
- Exactly-once: difícil de implementar

ESTRATEGIAS:

1. IDEMPOTENCIA (Recomendado):

type MessageProcessor struct {
    processed sync.Map  // or Redis set
}

func (p *MessageProcessor) Process(msg Message) error {
    // Check if already processed
    if _, loaded := p.processed.LoadOrStore(msg.ID, true); loaded {
        log.Printf("Message %s already processed, skipping", msg.ID)
        return nil
    }

    // Process message
    return processMessage(msg)
}

// Con Redis
func ProcessIdempotent(ctx context.Context, msg Message) error {
    key := "processed:" + msg.ID

    // Try to set (NX = only if not exists)
    set, err := redis.SetNX(ctx, key, "1", 24*time.Hour).Result()
    if err != nil {
        return err
    }

    if !set {
        log.Printf("Already processed: %s", msg.ID)
        return nil
    }

    return processMessage(msg)
}

2. TRANSACTIONAL OUTBOX PATTERN:

// Guardar mensaje Y estado en misma transacción
tx, _ := db.Begin()

// Insertar evento en outbox
query := "INSERT INTO outbox (id, aggregate_id, event_type, payload, created_at) VALUES ($1, $2, $3, $4, $5)"
_, err := tx.Exec(query, event.ID, event.AggregateID, event.Type, event.Payload, time.Now())

// Actualizar estado
_, err = tx.Exec("UPDATE orders SET status = $1 WHERE id = $2", "confirmed", orderID)

tx.Commit()

// Background worker publica desde outbox
func PublishFromOutbox(ctx context.Context) {
    for {
        events := loadUnpublishedEvents()
        for _, event := range events {
            if err := publisher.Publish(ctx, event); err != nil {
                continue
            }
            markAsPublished(event.ID)
        }
        time.Sleep(time.Second)
    }
}

3. KAFKA TRANSACTIONS (Exactly-once semantics):

// Producer con transacciones
writer := &kafka.Writer{
    Addr:  kafka.TCP("localhost:9092"),
    Topic: "orders",

    // Enable idempotence
    Idempotent: true,

    // Transactional ID
    RequiredAcks: kafka.RequireAll,
}

// Consume-Transform-Produce con transactions
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "input",
    GroupID: "processor",

    // Read committed only
    IsolationLevel: kafka.ReadCommitted,
})

for {
    msg, _ := reader.FetchMessage(ctx)

    // Begin transaction
    txWriter := writer.BeginTxn()

    // Process and produce
    result := process(msg)
    txWriter.WriteMessages(ctx, result)

    // Commit consumer offset within transaction
    txWriter.SendOffsetsToTransaction(ctx, map[string][]*kafka.TopicPartition{
        msg.Topic: {{Topic: msg.Topic, Partition: msg.Partition, Offset: msg.Offset + 1}},
    })

    // Commit transaction
    txWriter.CommitTxn(ctx)
}
`)

	// ============================================
	// ERROR HANDLING Y RETRY
	// ============================================
	fmt.Println("\n--- Error Handling y Retry ---")
	os.Stdout.WriteString(`
RETRY CON BACKOFF:

import "github.com/cenkalti/backoff/v4"

func ProcessWithRetry(ctx context.Context, msg Message) error {
    operation := func() error {
        return processMessage(msg)
    }

    b := backoff.NewExponentialBackOff()
    b.MaxElapsedTime = 5 * time.Minute

    return backoff.Retry(operation, backoff.WithContext(b, ctx))
}

DEAD LETTER QUEUE:

func ConsumeWithDLQ(ctx context.Context) {
    for msg := range consumer.Messages() {
        err := processMessage(msg)

        if err != nil {
            // Increment retry count
            retries := getRetryCount(msg)
            if retries >= maxRetries {
                // Send to DLQ
                dlq.Send(ctx, msg)
                consumer.Ack(msg)
                log.Printf("Sent to DLQ: %s", msg.ID)
            } else {
                // Nack and requeue with delay
                setRetryCount(msg, retries+1)
                consumer.Nack(msg)
            }
        } else {
            consumer.Ack(msg)
        }
    }
}

CIRCUIT BREAKER:

import "github.com/sony/gobreaker"

cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "message-processor",
    MaxRequests: 3,
    Timeout:     10 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.ConsecutiveFailures > 5
    },
})

func ProcessWithCircuitBreaker(msg Message) error {
    _, err := cb.Execute(func() (any, error) {
        return nil, processMessage(msg)
    })
    return err
}
`)

	// ============================================
	// MONITORING Y OBSERVABILITY
	// ============================================
	fmt.Println("\n--- Monitoring y Observability ---")
	fmt.Println(`
MÉTRICAS IMPORTANTES:

1. Producer metrics:
   - Messages produced/sec
   - Produce latency
   - Produce errors
   - Queue depth (backlog)

2. Consumer metrics:
   - Messages consumed/sec
   - Consumer lag
   - Processing time
   - Error rate
   - Redelivery rate

PROMETHEUS METRICS:

import "github.com/prometheus/client_golang/prometheus"

var (
    messagesProduced = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "messages_produced_total",
            Help: "Total messages produced",
        },
        []string{"topic"},
    )

    messagesConsumed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "messages_consumed_total",
            Help: "Total messages consumed",
        },
        []string{"topic", "status"},
    )

    processingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "message_processing_duration_seconds",
            Help:    "Message processing duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"topic"},
    )

    consumerLag = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "consumer_lag",
            Help: "Consumer lag in messages",
        },
        []string{"topic", "partition"},
    )
)

func ProcessWithMetrics(ctx context.Context, msg Message) error {
    start := time.Now()

    err := processMessage(msg)

    duration := time.Since(start).Seconds()
    processingDuration.WithLabelValues(msg.Topic).Observe(duration)

    status := "success"
    if err != nil {
        status = "error"
    }
    messagesConsumed.WithLabelValues(msg.Topic, status).Inc()

    return err
}

TRACING:

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func ProduceWithTracing(ctx context.Context, msg Message) error {
    tracer := otel.Tracer("message-queue")
    ctx, span := tracer.Start(ctx, "produce_message")
    defer span.End()

    span.SetAttributes(
        attribute.String("topic", msg.Topic),
        attribute.String("message.id", msg.ID),
    )

    err := producer.Send(ctx, msg)
    if err != nil {
        span.RecordError(err)
    }

    return err
}`)
	// Demostración conceptual
	demonstrateMessageQueues()
}

func demonstrateMessageQueues() {
	fmt.Printf("\n--- Demostración Conceptual ---\n")
	fmt.Printf("Características principales:\n")
	fmt.Printf("  Kafka: Alta throughput, persistencia, streaming\n")
	fmt.Printf("  NATS: Ultra rápido, ligero, request/reply\n")
	fmt.Printf("  RabbitMQ: Routing complejo, AMQP, flexibilidad\n")
	fmt.Printf("\nGarantías de entrega:\n")
	fmt.Printf("  At-most-once: rápido, puede perder mensajes\n")
	fmt.Printf("  At-least-once: duplicados posibles, usa idempotencia\n")
	fmt.Printf("  Exactly-once: complejo, usa transacciones o outbox\n")
	fmt.Printf("\nTTL recomendados:\n")
	fmt.Printf("  - Processed messages: %v\n", 24*time.Hour)
	fmt.Printf("  - Consumer locks: %v\n", 30*time.Second)
}

/*
RESUMEN DE MESSAGE QUEUES:

KAFKA:
- Librería: github.com/segmentio/kafka-go
- Uso: Event streaming, logs, high throughput
- Producer: kafka.Writer con RequiredAcks
- Consumer: kafka.Reader con consumer groups
- Exactamente-una-vez: transactions

NATS:
- Librería: github.com/nats-io/nats.go
- Uso: Microservicios, request/reply, pub/sub
- Core NATS: fire-and-forget
- JetStream: persistencia, ACK, replay
- Key-Value store incorporado

RABBITMQ:
- Librería: github.com/streadway/amqp
- Uso: Task queues, routing complejo
- Exchanges: direct, fanout, topic, headers
- Publisher confirms para garantías
- Dead Letter Queues built-in

PATRONES:
1. Work Queue: competing consumers
2. Pub/Sub: broadcasting
3. Request/Reply: RPC pattern
4. Priority Queue: priorización
5. Delayed: scheduled messages

EXACTLY-ONCE:
- Idempotencia (recomendado)
- Transactional Outbox
- Kafka Transactions
- Deduplicación con Redis

ERROR HANDLING:
- Retry con exponential backoff
- Dead Letter Queue
- Circuit breaker
- Max retry limit

MONITORING:
- Producer/consumer throughput
- Consumer lag
- Error rate
- Processing latency

MEJORES PRÁCTICAS:
1. Usar consumer groups para escalabilidad
2. Implementar idempotencia
3. Configurar DLQ para errores
4. Monitorear consumer lag
5. Usar ACK manual en producción
6. Configurar timeouts apropiados
7. Incluir correlation IDs
8. Versionar mensajes
*/

/*
SUMMARY - CHAPTER 091: Message Queues

TOPIC: Asynchronous Messaging Systems
- Kafka (segmentio/kafka-go): High-throughput event streaming, partition-based distribution, consumer groups
- Kafka Producer: Writer with batching, compression (Snappy), RequiredAcks for delivery guarantees
- Kafka Consumer: Reader with consumer groups, auto-commit or manual commit, offset management
- NATS (nats-io/nats.go): Ultra-fast pub/sub, request/reply pattern, queue groups for load balancing
- NATS JetStream: Adds persistence, acknowledgments, replay capability, key-value store built-in
- RabbitMQ (streadway/amqp): AMQP protocol, exchanges (direct/fanout/topic/headers), flexible routing
- RabbitMQ patterns: Queue/exchange binding, routing keys, publisher confirms, consumer acknowledgments
- Producer/Consumer patterns: Work queues (competing consumers), Pub/Sub (broadcasting), Request/Reply (RPC)
- Exactly-once delivery: Idempotency (storing processed message IDs), Transactional Outbox pattern
- Kafka transactions: Idempotent producer, transactional writes, read committed isolation level
- Error handling: Exponential backoff retry, Dead Letter Queue (DLQ), circuit breaker, max retry limit
- Idempotency: Store processed message IDs in Redis/DB, LoadOrStore for deduplication, TTL cleanup
- Transactional Outbox: DB writes and event publishing in same transaction, background worker publishes
- Monitoring: Producer/consumer throughput, consumer lag, error rate, processing duration, redelivery rate
- Best practices: Consumer groups for scaling, manual ACK in production, correlation IDs, message versioning
*/
