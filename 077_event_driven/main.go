// Package main - Chapter 077: Event-Driven Architecture
// Arquitectura orientada a eventos, CQRS, Event Sourcing, Sagas, Outbox pattern,
// idempotencia, versionado de eventos y testing de sistemas event-driven.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== EVENT-DRIVEN ARCHITECTURE, CQRS & EVENT SOURCING EN GO ===")

	// ============================================
	// EVENT-DRIVEN ARCHITECTURE: CONCEPTOS
	// ============================================
	fmt.Println("\n--- Event-Driven Architecture: Conceptos ---")
	fmt.Print(`
CONCEPTO:
La arquitectura orientada a eventos (EDA) es un patron donde los
componentes se comunican a traves de eventos asincronos en lugar
de llamadas directas (request/response).

COMPONENTES PRINCIPALES:
- Producer/Publisher: Genera y emite eventos
- Consumer/Subscriber: Escucha y reacciona a eventos
- Event Broker/Bus: Transporta eventos entre producers y consumers
- Event: Mensaje inmutable que describe algo que ocurrio

VENTAJAS:
- Desacoplamiento:   Los servicios no se conocen entre si
- Escalabilidad:     Consumers se escalan independientemente
- Resiliencia:       Fallos aislados, retry automatico
- Auditoria:         Los eventos son un log natural
- Flexibilidad:      Agregar consumers sin modificar producers

DESVENTAJAS:
- Complejidad:       Debugging distribuido, eventual consistency
- Ordering:          Garantizar orden de eventos es dificil
- Idempotencia:      Consumers deben manejar duplicados
- Monitoring:        Necesitas observabilidad end-to-end

PATRONES DE COMUNICACION:

1. Event Notification:
   - "Algo paso" (minima informacion)
   - Consumer consulta al producer si necesita mas datos
   - Ejemplo: OrderCreated { order_id: "123" }

2. Event-Carried State Transfer:
   - El evento lleva toda la informacion necesaria
   - Consumer no necesita callback al producer
   - Ejemplo: OrderCreated { order_id: "123", items: [...], total: 99.99 }

3. Domain Events:
   - Eventos que representan hechos del dominio de negocio
   - Nombrados en pasado: OrderPlaced, PaymentReceived, ItemShipped
`)

	// ============================================
	// DOMAIN EVENTS
	// ============================================
	fmt.Println("\n--- Domain Events ---")
	fmt.Print(`
DEFINICION:
Un Domain Event es un registro inmutable de algo que ocurrio
en el dominio de negocio. Siempre nombrado en tiempo pasado.

CONVENCIONES DE NAMING:
- Verbo en pasado + Sustantivo: OrderPlaced, UserRegistered
- Especifico, no generico: PaymentFailed (no Error)
- Refleja lenguaje del negocio (Ubiquitous Language)

ESTRUCTURA BASE:
Todo evento debe tener:
- ID unico (para deduplicacion)
- Tipo del evento
- Timestamp
- Aggregate ID (entidad que origino el evento)
- Version (para schema evolution)
- Payload (datos del evento)
- Metadata (correlation ID, causation ID, user ID)
`)

	fmt.Println("--- Demo: Domain Events ---")
	demoDomainEvents()

	// ============================================
	// EVENT BUS (IN-MEMORY CON CHANNELS)
	// ============================================
	fmt.Println("\n--- Event Bus: In-Memory con Channels ---")
	fmt.Print(`
CONCEPTO:
Un Event Bus desacopla publishers de subscribers.
En Go, lo implementamos con channels y goroutines.

RESPONSABILIDADES:
- Registrar subscribers por tipo de evento
- Publicar eventos a todos los subscribers interesados
- Manejar errores de subscribers sin afectar al publisher
- Soportar wildcards o filtros

PATRONES:
1. Fan-out:  Un evento va a TODOS los subscribers
2. Fan-in:   Multiples producers envian a un solo channel
3. Competing: Solo UN consumer procesa cada evento (work queue)
`)

	fmt.Println("--- Demo: Event Bus con Publisher/Subscriber ---")
	demoEventBus()

	// ============================================
	// CQRS - COMMAND QUERY RESPONSIBILITY SEGREGATION
	// ============================================
	fmt.Println("\n--- CQRS: Command Query Responsibility Segregation ---")
	fmt.Print(`
CONCEPTO:
CQRS separa las operaciones de escritura (Commands) de las
operaciones de lectura (Queries) en modelos independientes.

SIN CQRS (modelo tradicional):
  Service -> Repository -> Database
  (reads y writes usan el mismo modelo)

CON CQRS:
  Command -> Command Handler -> Write Model -> Write DB
  Query   -> Query Handler   -> Read Model  -> Read DB

CUANDO USAR CQRS:
- Cargas de lectura y escritura muy diferentes
- Read model necesita denormalizacion o agregaciones
- Quieres escalar reads y writes independientemente
- Dominio complejo con validaciones de escritura

CUANDO NO USAR:
- CRUDs simples sin logica de negocio
- Aplicaciones pequenas con poco trafico
- El equipo no tiene experiencia con el patron

COMMAND:
- Expresa una intencion de cambiar el estado
- Nombrado en imperativo: CreateOrder, CancelPayment
- Puede ser rechazado (validacion)
- Retorna void o error (nunca datos)

QUERY:
- Solicita informacion sin efectos secundarios
- Nombrado como pregunta: GetOrderByID, ListActiveUsers
- Nunca modifica estado
- Optimizado para lectura
`)

	fmt.Println("--- Demo: CQRS Completo ---")
	demoCQRS()

	// ============================================
	// EVENTUAL CONSISTENCY
	// ============================================
	fmt.Println("\n--- Eventual Consistency ---")
	os.Stdout.WriteString(`
CONCEPTO:
En CQRS, el read model se actualiza de forma asincrona.
Hay un delay entre la escritura y su reflejo en lecturas.

TIMELINE:

  t=0: Command -> Write Model actualizado
  t=1: Evento emitido a projections
  t=2: Read Model actualizado (eventual consistency)

ESTRATEGIAS PARA MANEJAR:

1. Optimistic UI:
   - Frontend asume exito y actualiza localmente
   - Si falla, revierte

2. Read-your-writes:
   - Despues de escribir, leer del write model (no del read model)
   - Solo para el usuario que escribio

3. Causal consistency:
   - Usar version numbers para detectar stale reads
   - Si read version < expected, retry

4. Polling/SSE:
   - Frontend hace polling hasta que read model refleja el cambio
   - O usa Server-Sent Events para push

PATRON READ-YOUR-WRITES:

type OrderService struct {
    writeRepo WriteRepository
    readRepo  ReadRepository
}

func (s *OrderService) GetOrder(ctx context.Context, id string, afterVersion int) (*OrderView, error) {
    view, err := s.readRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Si la projection no ha alcanzado la version esperada,
    // leer directamente del write model
    if view.Version < afterVersion {
        order, err := s.writeRepo.GetByID(ctx, id)
        if err != nil {
            return nil, err
        }
        return toOrderView(order), nil
    }

    return view, nil
}
`)

	// ============================================
	// EVENT SOURCING
	// ============================================
	fmt.Println("\n--- Event Sourcing ---")
	fmt.Print(`
CONCEPTO:
En lugar de almacenar el estado actual de una entidad,
almacenamos TODOS los eventos que llevaron a ese estado.

ESTADO TRADICIONAL:
  Account { id: "1", balance: 150 }

EVENT SOURCING:
  AccountCreated  { id: "1", owner: "Alice" }        -> balance: 0
  MoneyDeposited  { id: "1", amount: 200 }            -> balance: 200
  MoneyWithdrawn  { id: "1", amount: 50 }             -> balance: 150

VENTAJAS:
- Auditoria completa (todo cambio queda registrado)
- Time travel (reconstruir estado en cualquier punto)
- Debugging (reproducir exactamente que paso)
- Diferentes proyecciones del mismo stream de eventos
- Deshacer operaciones (compensating events)

DESVENTAJAS:
- Complejidad de implementacion
- Event store especializado
- Schema evolution es critico
- Snapshots necesarios para performance
- Eventual consistency

COMPONENTES:

1. Event Store:     Almacena eventos ordenados por aggregate
2. Aggregate Root:  Entidad que aplica eventos y valida comandos
3. Projections:     Vistas materializadas del stream de eventos
4. Snapshots:       Estado pre-calculado para evitar replay completo
`)

	fmt.Println("--- Demo: Event Sourcing ---")
	demoEventSourcing()

	// ============================================
	// EVENT STORE INTERFACE
	// ============================================
	fmt.Println("\n--- Event Store Interface ---")
	os.Stdout.WriteString(`
INTERFACE PARA EVENT STORE:

type EventStore interface {
    // Append agrega eventos al stream de un aggregate
    // expectedVersion previene conflictos de concurrencia (optimistic locking)
    Append(ctx context.Context, aggregateID string, events []Event, expectedVersion int) error

    // Load retorna todos los eventos de un aggregate
    Load(ctx context.Context, aggregateID string) ([]Event, error)

    // LoadFrom retorna eventos desde una version especifica
    LoadFrom(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error)

    // LoadByType retorna eventos de un tipo especifico (para projections globales)
    LoadByType(ctx context.Context, eventType string) ([]Event, error)
}

IMPLEMENTACIONES COMUNES:
- PostgreSQL:   Tabla con aggregate_id, version, event_type, data, timestamp
- EventStoreDB: Base de datos especializada para event sourcing
- DynamoDB:     Partition key = aggregate_id, sort key = version
- Kafka:        Topic por aggregate type, partition por aggregate_id

SCHEMA SQL PARA EVENT STORE:

CREATE TABLE events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id  UUID NOT NULL,
    aggregate_type VARCHAR(100) NOT NULL,
    event_type    VARCHAR(100) NOT NULL,
    version       INT NOT NULL,
    data          JSONB NOT NULL,
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(aggregate_id, version)
);

CREATE INDEX idx_events_aggregate ON events(aggregate_id, version);
CREATE INDEX idx_events_type ON events(event_type, created_at);

OPTIMISTIC CONCURRENCY:

INSERT INTO events (aggregate_id, aggregate_type, event_type, version, data)
VALUES ($1, $2, $3, $4, $5)
-- Si la version ya existe, hay un conflicto de concurrencia
-- El UNIQUE constraint lanza error y el caller debe reintentar
`)

	// ============================================
	// AGGREGATE ROOT PATTERN
	// ============================================
	fmt.Println("\n--- Aggregate Root Pattern ---")
	os.Stdout.WriteString(`
CONCEPTO:
El Aggregate Root es la entidad principal que:
- Valida comandos contra reglas de negocio
- Genera eventos cuando un comando es valido
- Reconstruye su estado aplicando eventos

ESTRUCTURA:

type AggregateRoot struct {
    id             string
    version        int
    uncommitted    []Event
}

func (a *AggregateRoot) GetID() string       { return a.id }
func (a *AggregateRoot) GetVersion() int      { return a.version }
func (a *AggregateRoot) GetUncommitted() []Event { return a.uncommitted }
func (a *AggregateRoot) ClearUncommitted()    { a.uncommitted = nil }

func (a *AggregateRoot) Apply(event Event) {
    a.version++
    a.uncommitted = append(a.uncommitted, event)
}

EJEMPLO - BANK ACCOUNT AGGREGATE:

type BankAccount struct {
    AggregateRoot
    owner   string
    balance float64
    closed  bool
}

// Comando: Depositar dinero
func (a *BankAccount) Deposit(amount float64) error {
    if a.closed {
        return errors.New("account is closed")
    }
    if amount <= 0 {
        return errors.New("amount must be positive")
    }

    a.Apply(MoneyDeposited{
        AccountID: a.id,
        Amount:    amount,
        Timestamp: time.Now(),
    })
    return nil
}

// Aplicar evento al estado (usado durante replay)
func (a *BankAccount) OnMoneyDeposited(e MoneyDeposited) {
    a.balance += e.Amount
}

PATRON CUANDO/THEN:
- CUANDO llega un comando -> Validar reglas de negocio
- SI es valido -> Generar evento(s)
- Aplicar evento al estado interno
- Retornar eventos uncommitted para persistir
`)

	// ============================================
	// PROJECTIONS
	// ============================================
	fmt.Println("\n--- Projections ---")
	os.Stdout.WriteString(`
CONCEPTO:
Las Projections son vistas materializadas que se construyen
procesando el stream de eventos. Optimizadas para queries.

UN MISMO STREAM, MULTIPLES VISTAS:

Events: OrderPlaced, ItemAdded, PaymentReceived, OrderShipped

Projection 1 - Order Summary:
  { order_id, status, total, item_count }

Projection 2 - Revenue Report:
  { date, total_revenue, order_count, avg_order_value }

Projection 3 - Inventory:
  { product_id, quantity_sold, quantity_remaining }

INTERFACE:

type Projection interface {
    // Handle procesa un evento y actualiza la vista
    Handle(ctx context.Context, event Event) error

    // Reset borra la vista para reconstruirla desde cero
    Reset(ctx context.Context) error
}

type ProjectionManager struct {
    projections map[string][]Projection // eventType -> projections
    position    int64                   // last processed event position
}

func (pm *ProjectionManager) Process(ctx context.Context, event Event) error {
    handlers, ok := pm.projections[event.Type]
    if !ok {
        return nil // No hay projections para este tipo
    }

    for _, p := range handlers {
        if err := p.Handle(ctx, event); err != nil {
            return fmt.Errorf("projection error for %s: %w", event.Type, err)
        }
    }

    pm.position = event.Position
    return nil
}

REBUILD DE PROJECTIONS:
Una ventaja clave del event sourcing es poder reconstruir
cualquier projection desde cero:

func RebuildProjection(ctx context.Context, store EventStore, proj Projection) error {
    // 1. Resetear la projection
    if err := proj.Reset(ctx); err != nil {
        return err
    }

    // 2. Replay todos los eventos
    events, err := store.LoadAll(ctx)
    if err != nil {
        return err
    }

    for _, event := range events {
        if err := proj.Handle(ctx, event); err != nil {
            return err
        }
    }

    return nil
}
`)

	// ============================================
	// SNAPSHOTS
	// ============================================
	fmt.Println("\n--- Snapshots ---")
	os.Stdout.WriteString(`
CONCEPTO:
Un Snapshot guarda el estado de un aggregate en un punto
determinado para evitar replay de todos los eventos.

SIN SNAPSHOT (1000 eventos):
  Load 1000 events -> Apply one by one -> Current state

CON SNAPSHOT (cada 100 eventos):
  Load snapshot@v900 -> Load 100 events -> Apply -> Current state

INTERFACE:

type SnapshotStore interface {
    Save(ctx context.Context, aggregateID string, version int, state []byte) error
    Load(ctx context.Context, aggregateID string) (version int, state []byte, err error)
}

ESTRATEGIA DE SNAPSHOT:

type SnapshotStrategy interface {
    ShouldSnapshot(aggregate Aggregate) bool
}

// Snapshot cada N eventos
type EveryNEvents struct {
    n int
}

func (s EveryNEvents) ShouldSnapshot(a Aggregate) bool {
    return a.GetVersion() %% s.n == 0
}

LOAD CON SNAPSHOT:

func LoadAggregate(
    ctx context.Context,
    store EventStore,
    snapshots SnapshotStore,
    id string,
) (*BankAccount, error) {
    account := &BankAccount{}

    // 1. Intentar cargar snapshot
    version, data, err := snapshots.Load(ctx, id)
    if err == nil {
        json.Unmarshal(data, account)
    }

    // 2. Cargar eventos desde la version del snapshot
    events, err := store.LoadFrom(ctx, id, version+1)
    if err != nil {
        return nil, err
    }

    // 3. Aplicar eventos restantes
    for _, event := range events {
        account.ApplyEvent(event)
    }

    return account, nil
}

SAVE CON SNAPSHOT:

func SaveAggregate(
    ctx context.Context,
    store EventStore,
    snapshots SnapshotStore,
    account *BankAccount,
    strategy SnapshotStrategy,
) error {
    events := account.GetUncommitted()

    // Persistir eventos
    if err := store.Append(ctx, account.GetID(), events, account.GetVersion()-len(events)); err != nil {
        return err
    }

    // Snapshot si corresponde
    if strategy.ShouldSnapshot(account) {
        data, _ := json.Marshal(account)
        snapshots.Save(ctx, account.GetID(), account.GetVersion(), data)
    }

    account.ClearUncommitted()
    return nil
}
`)

	// ============================================
	// REPLAYING EVENTS
	// ============================================
	fmt.Println("\n--- Replaying Events ---")
	fmt.Print(`
CONCEPTO:
Replay es reconstruir el estado procesando eventos desde el inicio
(o desde un snapshot). Es fundamental para:
- Reconstruir projections
- Debugging (ver estado en un punto en el tiempo)
- Migrar a nuevos read models
- Time travel queries

TIPOS DE REPLAY:

1. Aggregate Replay:
   Reconstruir una sola entidad desde sus eventos

2. Projection Replay:
   Reconstruir una vista completa desde todos los eventos

3. Temporal Query:
   Reconstruir estado hasta un punto en el tiempo

4. Selective Replay:
   Solo ciertos tipos de eventos o ciertos aggregates
`)

	fmt.Println("--- Demo: Replay de Eventos ---")
	demoReplayEvents()

	// ============================================
	// SAGA PATTERN
	// ============================================
	fmt.Println("\n--- Saga Pattern ---")
	os.Stdout.WriteString(`
CONCEPTO:
Las Sagas manejan transacciones distribuidas que abarcan
multiples servicios. Si un paso falla, ejecutan compensaciones.

EJEMPLO: Crear una orden
1. Reservar inventario
2. Cobrar pago
3. Enviar confirmacion

Si el paso 2 falla -> Compensar paso 1 (liberar inventario)

DOS ESTILOS:

1. ORCHESTRATION (coordinador central):

   Saga Orchestrator
        |
        +--> InventoryService.Reserve()
        |         |
        |    (si falla) -> ABORT
        |
        +--> PaymentService.Charge()
        |         |
        |    (si falla) -> InventoryService.Release() -> ABORT
        |
        +--> NotificationService.Send()
              |
         (si falla) -> PaymentService.Refund()
                     -> InventoryService.Release()
                     -> ABORT

   Ventajas: Flujo claro, facil de debuggear, control centralizado
   Desventajas: Orchestrator es single point of failure, acoplamiento

2. CHOREOGRAPHY (eventos entre servicios):

   OrderService --OrderCreated--> InventoryService
                                       |
                              InventoryReserved
                                       |
                                  PaymentService
                                       |
                                PaymentCharged
                                       |
                              NotificationService
                                       |
                              ConfirmationSent

   Si PaymentFailed:
   PaymentService --PaymentFailed--> InventoryService (release)

   Ventajas: Desacoplado, resiliente, sin SPOF
   Desventajas: Flujo dificil de seguir, debugging complejo

IMPLEMENTACION ORCHESTRATOR:

type SagaStep struct {
    Name       string
    Execute    func(ctx context.Context, data any) error
    Compensate func(ctx context.Context, data any) error
}

type Saga struct {
    Name  string
    Steps []SagaStep
}

func (s *Saga) Run(ctx context.Context, data any) error {
    var completedSteps []SagaStep

    for _, step := range s.Steps {
        if err := step.Execute(ctx, data); err != nil {
            // Compensar en orden inverso
            for i := len(completedSteps) - 1; i >= 0; i-- {
                if compErr := completedSteps[i].Compensate(ctx, data); compErr != nil {
                    // Log compensacion fallida para intervencion manual
                    log.Printf("SAGA %s: compensacion fallida en %s: %v",
                        s.Name, completedSteps[i].Name, compErr)
                }
            }
            return fmt.Errorf("saga %s failed at step %s: %w", s.Name, step.Name, err)
        }
        completedSteps = append(completedSteps, step)
    }

    return nil
}

SAGA STATE MACHINE:

type SagaState string

const (
    SagaStarted      SagaState = "STARTED"
    SagaInProgress   SagaState = "IN_PROGRESS"
    SagaCompleted    SagaState = "COMPLETED"
    SagaFailed       SagaState = "FAILED"
    SagaCompensating SagaState = "COMPENSATING"
    SagaCompensated  SagaState = "COMPENSATED"
)

type SagaInstance struct {
    ID             string
    SagaType       string
    State          SagaState
    CurrentStep    int
    Data           json.RawMessage
    CompletedSteps []string
    Errors         []string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
`)

	// ============================================
	// OUTBOX PATTERN
	// ============================================
	fmt.Println("\n--- Outbox Pattern ---")
	os.Stdout.WriteString(`
CONCEPTO:
El Outbox Pattern garantiza que los cambios en la base de datos
y la publicacion de eventos sean atomicos (sin dual-write problem).

PROBLEMA (Dual Write):
1. Save to DB        -> OK
2. Publish to Kafka  -> FALLA!
   El evento se pierde, DB y message broker estan inconsistentes.

SOLUCION (Outbox):
1. En una MISMA transaccion:
   - Save entity to DB
   - Save event to outbox table
2. Un proceso separado (Outbox Relay) lee la outbox table
   y publica los eventos al message broker
3. Marca eventos como publicados

SCHEMA:

CREATE TABLE outbox (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id   UUID NOT NULL,
    event_type  VARCHAR(100) NOT NULL,
    payload     JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    published   BOOLEAN NOT NULL DEFAULT false,
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_unpublished ON outbox(created_at) WHERE NOT published;

TRANSACCION ATOMICA:

func (r *OrderRepo) CreateOrder(ctx context.Context, order Order) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. Guardar la orden
    _, err = tx.ExecContext(ctx,
        "INSERT INTO orders (id, customer_id, total) VALUES ($1, $2, $3)",
        order.ID, order.CustomerID, order.Total)
    if err != nil {
        return err
    }

    // 2. Guardar evento en outbox (misma transaccion)
    eventPayload, _ := json.Marshal(OrderCreatedEvent{
        OrderID:    order.ID,
        CustomerID: order.CustomerID,
        Total:      order.Total,
    })

    _, err = tx.ExecContext(ctx,
        "INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload) VALUES ($1, $2, $3, $4)",
        "Order", order.ID, "OrderCreated", eventPayload)
    if err != nil {
        return err
    }

    return tx.Commit()
}

OUTBOX RELAY (Polling):

func (r *OutboxRelay) Run(ctx context.Context) error {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            r.publishPending(ctx)
        }
    }
}

func (r *OutboxRelay) publishPending(ctx context.Context) error {
    events, err := r.repo.GetUnpublished(ctx, 100)
    if err != nil {
        return err
    }

    for _, event := range events {
        if err := r.publisher.Publish(ctx, event); err != nil {
            return err // Retry en siguiente tick
        }
        r.repo.MarkPublished(ctx, event.ID)
    }
    return nil
}

ALTERNATIVA: Change Data Capture (CDC)
Herramientas como Debezium leen el WAL de PostgreSQL y
publican cambios automaticamente a Kafka. Mas robusto que polling.
`)

	// ============================================
	// IDEMPOTENCY IN EVENT HANDLERS
	// ============================================
	fmt.Println("\n--- Idempotencia en Event Handlers ---")
	os.Stdout.WriteString(`
CONCEPTO:
Los eventos pueden entregarse mas de una vez (at-least-once delivery).
Los handlers deben ser idempotentes: procesar el mismo evento N veces
produce el mismo resultado que procesarlo 1 vez.

ESTRATEGIAS:

1. IDEMPOTENCY KEY (deduplication table):

CREATE TABLE processed_events (
    event_id   UUID PRIMARY KEY,
    handler    VARCHAR(100) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

func (h *OrderHandler) Handle(ctx context.Context, event Event) error {
    // Verificar si ya se proceso
    exists, err := h.repo.EventProcessed(ctx, event.ID, "OrderHandler")
    if err != nil {
        return err
    }
    if exists {
        return nil // Ya procesado, skip
    }

    // Procesar
    if err := h.processOrder(ctx, event); err != nil {
        return err
    }

    // Marcar como procesado (misma transaccion si es posible)
    return h.repo.MarkProcessed(ctx, event.ID, "OrderHandler")
}

2. NATURAL IDEMPOTENCY:
   Disenar operaciones que sean naturalmente idempotentes:
   - SET balance = 150     (idempotente)
   - ADD balance + 50      (NO idempotente)
   - SET status = 'paid'   (idempotente)
   - INSERT IF NOT EXISTS  (idempotente)

3. CONDITIONAL UPDATES:

UPDATE orders
SET status = 'shipped', version = version + 1
WHERE id = $1 AND version = $2
-- Si version no coincide, otro handler ya lo actualizo

4. CONSUMER OFFSET TRACKING:
   Kafka y otros brokers trackean el offset del consumer.
   Si el consumer falla, re-procesa desde el ultimo offset committeado.
`)

	// ============================================
	// EVENT VERSIONING & SCHEMA EVOLUTION
	// ============================================
	fmt.Println("\n--- Event Versioning y Schema Evolution ---")
	os.Stdout.WriteString(`
CONCEPTO:
Los eventos son inmutables. Cuando la estructura cambia,
necesitas estrategias de versionado.

ESTRATEGIAS:

1. WEAK SCHEMA (tolerante):
   - Ignorar campos desconocidos (json.Decoder con DisallowUnknownFields=false)
   - Usar valores default para campos nuevos

type OrderCreatedV1 struct {
    OrderID string
    Total   float64
}

type OrderCreatedV2 struct {
    OrderID  string
    Total    float64
    Currency string // Nuevo campo, default "USD"
}

func deserializeOrderCreated(data []byte) OrderCreatedV2 {
    var event OrderCreatedV2
    json.Unmarshal(data, &event)
    if event.Currency == "" {
        event.Currency = "USD" // Default para eventos v1
    }
    return event
}

2. UPCASTING (transformar al leer):

type EventUpcaster interface {
    CanUpcast(eventType string, version int) bool
    Upcast(data []byte, fromVersion int) ([]byte, int, error)
}

type OrderCreatedUpcaster struct{}

func (u OrderCreatedUpcaster) Upcast(data []byte, fromVersion int) ([]byte, int, error) {
    switch fromVersion {
    case 1:
        var v1 OrderCreatedV1
        json.Unmarshal(data, &v1)
        v2 := OrderCreatedV2{
            OrderID:  v1.OrderID,
            Total:    v1.Total,
            Currency: "USD",
        }
        result, _ := json.Marshal(v2)
        return result, 2, nil
    default:
        return data, fromVersion, nil
    }
}

3. NEW EVENT TYPE:
   Cuando el cambio es muy grande, crear un nuevo tipo:
   - OrderCreated     (v1, deprecated)
   - OrderPlaced      (nuevo, reemplaza OrderCreated)
   - Mantener handlers para ambos durante la transicion

4. EVENT COPY TRANSFORMATION:
   Para migraciones grandes, crear un nuevo stream
   transformando eventos del stream original:

   OldStream -> Transform -> NewStream

REGLAS DE COMPATIBILIDAD:
- NUNCA eliminar campos de eventos existentes
- NUNCA cambiar el tipo de un campo
- Agregar campos con defaults es SAFE
- Renombrar campos requiere upcasting
- Mantener backward compatibility siempre
`)

	// ============================================
	// LIBRARIES: WATERMILL
	// ============================================
	fmt.Println("\n--- Watermill (ThreeDotsLabs) ---")
	os.Stdout.WriteString(`
WATERMILL:
Framework de Go para trabajar con mensajes (events, commands).
Creado por ThreeDotsLabs.

go get github.com/ThreeDotsLabs/watermill

CARACTERISTICAS:
- Pub/Sub abstraction sobre Kafka, NATS, RabbitMQ, Google Pub/Sub
- Middlewares (retry, throttle, poison queue, correlation ID)
- Router para manejar multiples handlers
- CQRS module built-in
- Event sourcing support

PUBLISHER/SUBSCRIBER:

import (
    "github.com/ThreeDotsLabs/watermill"
    "github.com/ThreeDotsLabs/watermill/message"
    "github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
)

// Publisher
publisher, err := kafka.NewPublisher(
    kafka.PublisherConfig{
        Brokers:   []string{"localhost:9092"},
        Marshaler: kafka.DefaultMarshaler{},
    },
    watermill.NewStdLogger(false, false),
)

msg := message.NewMessage(watermill.NewUUID(), payload)
msg.Metadata.Set("correlation_id", correlationID)

publisher.Publish("order.events", msg)

// Subscriber
subscriber, err := kafka.NewSubscriber(
    kafka.SubscriberConfig{
        Brokers:       []string{"localhost:9092"},
        Unmarshaler:   kafka.DefaultMarshaler{},
        ConsumerGroup: "order-processor",
    },
    watermill.NewStdLogger(false, false),
)

messages, err := subscriber.Subscribe(ctx, "order.events")
for msg := range messages {
    // Procesar mensaje
    fmt.Println(string(msg.Payload))
    msg.Ack() // o msg.Nack() para retry
}

ROUTER:

router, err := message.NewRouter(
    message.RouterConfig{},
    watermill.NewStdLogger(false, false),
)

// Middlewares
router.AddMiddleware(
    middleware.Retry{MaxRetries: 3, WaitTime: time.Second}.Middleware,
    middleware.CorrelationID,
    middleware.Recoverer,
    middleware.NewThrottle(100, time.Second).Middleware,
)

// Handler: consume de un topic, publica a otro
router.AddHandler(
    "order-to-invoice",     // handler name
    "order.events",         // subscribe topic
    subscriber,
    "invoice.commands",     // publish topic
    publisher,
    func(msg *message.Message) ([]*message.Message, error) {
        var event OrderCreated
        json.Unmarshal(msg.Payload, &event)

        cmd := CreateInvoice{OrderID: event.OrderID, Amount: event.Total}
        payload, _ := json.Marshal(cmd)

        return []*message.Message{
            message.NewMessage(watermill.NewUUID(), payload),
        }, nil
    },
)

// No-publisher handler (solo consume)
router.AddNoPublisherHandler(
    "order-logger",
    "order.events",
    subscriber,
    func(msg *message.Message) error {
        log.Println("Order event:", string(msg.Payload))
        msg.Ack()
        return nil
    },
)

router.Run(ctx)

CQRS CON WATERMILL:

import "github.com/ThreeDotsLabs/watermill/components/cqrs"

// Command Bus
commandBus, err := cqrs.NewCommandBusWithConfig(publisher, cqrs.CommandBusConfig{
    GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
        return "commands." + params.CommandName, nil
    },
    Marshaler: cqrs.JSONMarshaler{},
})

// Event Bus
eventBus, err := cqrs.NewEventBusWithConfig(publisher, cqrs.EventBusConfig{
    GeneratePublishTopic: func(params cqrs.EventBusGeneratePublishTopicParams) (string, error) {
        return "events." + params.EventName, nil
    },
    Marshaler: cqrs.JSONMarshaler{},
})

// Command Handler
type CreateOrderHandler struct {}

func (h CreateOrderHandler) HandlerName() string { return "CreateOrderHandler" }

func (h CreateOrderHandler) NewCommand() any { return &CreateOrderCommand{} }

func (h CreateOrderHandler) Handle(ctx context.Context, cmd any) error {
    c := cmd.(*CreateOrderCommand)
    // Logica de negocio...
    return eventBus.Publish(ctx, OrderCreatedEvent{OrderID: c.OrderID})
}
`)

	// ============================================
	// INTEGRATION: KAFKA, NATS, RABBITMQ
	// ============================================
	fmt.Println("\n--- Integracion con Kafka, NATS, RabbitMQ ---")
	os.Stdout.WriteString(`
KAFKA:

go get github.com/ThreeDotsLabs/watermill-kafka/v3
go get github.com/segmentio/kafka-go

// Watermill + Kafka
publisher, _ := kafka.NewPublisher(kafka.PublisherConfig{
    Brokers:   []string{"localhost:9092"},
    Marshaler: kafka.DefaultMarshaler{},
}, logger)

subscriber, _ := kafka.NewSubscriber(kafka.SubscriberConfig{
    Brokers:       []string{"localhost:9092"},
    ConsumerGroup: "my-group",
    Unmarshaler:   kafka.DefaultMarshaler{},
}, logger)

KAFKA PARA EVENT SOURCING:
- Topic por aggregate type (ej: "accounts", "orders")
- Partition key = aggregate ID (garantiza orden por aggregate)
- Retention indefinido (log compaction para snapshots)
- Consumer groups para projections independientes

NATS:

go get github.com/ThreeDotsLabs/watermill-nats/v2
go get github.com/nats-io/nats.go

// NATS JetStream (persistent)
publisher, _ := nats.NewPublisher(nats.PublisherConfig{
    URL:       "nats://localhost:4222",
    Marshaler: nats.JSONMarshaler{},
}, logger)

subscriber, _ := nats.NewSubscriber(nats.SubscriberConfig{
    URL:            "nats://localhost:4222",
    QueueGroupPrefix: "my-service",
    Unmarshaler:    nats.JSONMarshaler{},
    JetStream: nats.JetStreamConfig{
        AutoProvision: true,
        DurablePrefix: "my-service",
    },
}, logger)

NATS PARA EVENTOS:
- Subjects con wildcards: "orders.>" para todos los eventos de orders
- JetStream para persistencia y at-least-once delivery
- Key-Value store para snapshots
- Muy bajo latencia (microsegundos)

RABBITMQ:

go get github.com/ThreeDotsLabs/watermill-amqp/v3

// RabbitMQ
publisher, _ := amqp.NewPublisher(amqp.NewDurableQueueConfig("amqp://localhost:5672"), logger)

subscriber, _ := amqp.NewSubscriber(amqp.NewDurableQueueConfig("amqp://localhost:5672"), logger)

RABBITMQ PARA EVENTOS:
- Exchange type "topic" para routing por event type
- Dead letter queues para poison messages
- Prefetch count para controlar backpressure
- Quorum queues para alta disponibilidad

COMPARACION:

| Feature          | Kafka        | NATS         | RabbitMQ     |
|------------------|-------------|-------------|--------------|
| Throughput       | Muy alto    | Muy alto    | Alto         |
| Latencia         | Baja        | Muy baja    | Baja         |
| Persistencia     | Si (default)| JetStream   | Si (durable) |
| Ordering         | Por partition| Por subject | Por queue    |
| Event Sourcing   | Excelente   | Bueno       | Posible      |
| Complejidad ops  | Alta        | Baja        | Media        |
| Consumer groups  | Nativo      | Queue groups| Bindings     |
`)

	// ============================================
	// TESTING EVENT-DRIVEN SYSTEMS
	// ============================================
	fmt.Println("\n--- Testing Event-Driven Systems ---")
	os.Stdout.WriteString(`
ESTRATEGIAS:

1. UNIT TESTING AGGREGATES:
   Patron Given/When/Then:

func TestBankAccount_Deposit(t *testing.T) {
    // Given: una cuenta existente con balance 100
    account := NewBankAccount("acc-1", "Alice")
    account.Deposit(100)
    account.ClearUncommitted()

    // When: depositar 50
    err := account.Deposit(50)

    // Then: sin error, evento generado, balance correcto
    assert.NoError(t, err)
    assert.Equal(t, 150.0, account.Balance())

    events := account.GetUncommitted()
    assert.Len(t, events, 1)
    assert.Equal(t, "MoneyDeposited", events[0].Type)
}

func TestBankAccount_Withdraw_InsufficientFunds(t *testing.T) {
    account := NewBankAccount("acc-1", "Alice")
    account.Deposit(50)
    account.ClearUncommitted()

    err := account.Withdraw(100)

    assert.ErrorContains(t, err, "insufficient funds")
    assert.Empty(t, account.GetUncommitted())
}

2. TESTING EVENT HANDLERS:

func TestOrderProjection_Handle(t *testing.T) {
    projection := NewOrderProjection(NewInMemoryStore())

    // Simular stream de eventos
    events := []Event{
        {Type: "OrderCreated", Data: OrderCreated{ID: "1", CustomerID: "c1"}},
        {Type: "ItemAdded", Data: ItemAdded{OrderID: "1", Product: "Widget", Qty: 2}},
        {Type: "OrderPaid", Data: OrderPaid{OrderID: "1", Amount: 49.99}},
    }

    for _, e := range events {
        err := projection.Handle(context.Background(), e)
        assert.NoError(t, err)
    }

    view := projection.GetOrder("1")
    assert.Equal(t, "paid", view.Status)
    assert.Equal(t, 49.99, view.Total)
    assert.Equal(t, 1, view.ItemCount)
}

3. TESTING SAGAS:

func TestOrderSaga_PaymentFails_CompensatesInventory(t *testing.T) {
    inventory := &MockInventoryService{}
    payment := &MockPaymentService{ShouldFail: true}

    saga := NewOrderSaga(inventory, payment)

    err := saga.Run(context.Background(), OrderData{ID: "1", Amount: 100})

    assert.Error(t, err)
    assert.True(t, inventory.ReleaseCalled, "debe compensar inventario")
    assert.False(t, payment.RefundCalled, "no debe refundar si pago fallo")
}

4. INTEGRATION TESTING CON IN-MEMORY BUS:

func TestOrderFlow_Integration(t *testing.T) {
    bus := NewInMemoryEventBus()
    orderRepo := NewInMemoryOrderRepo()
    invoiceRepo := NewInMemoryInvoiceRepo()

    // Registrar handlers
    bus.Subscribe("OrderCreated", NewInvoiceHandler(invoiceRepo))
    bus.Subscribe("OrderCreated", NewNotificationHandler())

    // Ejecutar command
    handler := NewCreateOrderHandler(orderRepo, bus)
    err := handler.Handle(ctx, CreateOrderCommand{CustomerID: "c1", Total: 99.99})
    assert.NoError(t, err)

    // Verificar side effects
    time.Sleep(10 * time.Millisecond) // esperar async handlers
    invoices := invoiceRepo.FindByCustomer("c1")
    assert.Len(t, invoices, 1)
}

5. CONTRACT TESTING:
   Verificar que los eventos cumplen el schema esperado.

func TestOrderCreatedEvent_Schema(t *testing.T) {
    event := OrderCreatedEvent{
        OrderID:    "123",
        CustomerID: "c1",
        Total:      99.99,
        Currency:   "USD",
        CreatedAt:  time.Now(),
    }

    data, err := json.Marshal(event)
    assert.NoError(t, err)

    // Verificar que deserializa correctamente
    var decoded OrderCreatedEvent
    err = json.Unmarshal(data, &decoded)
    assert.NoError(t, err)
    assert.Equal(t, event.OrderID, decoded.OrderID)
}

6. TESTCONTAINERS:

import "github.com/testcontainers/testcontainers-go"

func TestWithKafka(t *testing.T) {
    ctx := context.Background()

    kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "confluentinc/cp-kafka:7.5.0",
            ExposedPorts: []string{"9092/tcp"},
        },
        Started: true,
    })
    defer kafkaC.Terminate(ctx)

    // Usar kafkaC.Host() y kafkaC.MappedPort() para conectar
}
`)

	// ============================================
	// RESUMEN
	// ============================================
	fmt.Println("\n--- Resumen ---")
	fmt.Println("Todos los demos completados exitosamente.")
}

// ============================================
// TYPES PARA DEMOS
// ============================================

type Event struct {
	ID            string
	Type          string
	AggregateID   string
	AggregateType string
	Version       int
	Timestamp     time.Time
	Data          any
	Metadata      map[string]string
}

type EventEnvelope struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	AggregateID   string          `json:"aggregate_id"`
	Version       int             `json:"version"`
	Timestamp     time.Time       `json:"timestamp"`
	Data          json.RawMessage `json:"data"`
	CorrelationID string          `json:"correlation_id"`
	CausationID   string          `json:"causation_id"`
}

// ============================================
// DEMO: DOMAIN EVENTS
// ============================================

type OrderCreated struct {
	OrderID    string    `json:"order_id"`
	CustomerID string    `json:"customer_id"`
	Items      []string  `json:"items"`
	Total      float64   `json:"total"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrderPaid struct {
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Method    string    `json:"method"`
	PaidAt    time.Time `json:"paid_at"`
}

type OrderShipped struct {
	OrderID    string    `json:"order_id"`
	TrackingNo string    `json:"tracking_no"`
	ShippedAt  time.Time `json:"shipped_at"`
}

func demoDomainEvents() {
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	events := []Event{
		{
			ID:          "evt-001",
			Type:        "OrderCreated",
			AggregateID: "order-123",
			Version:     1,
			Timestamp:   now,
			Data: OrderCreated{
				OrderID:    "order-123",
				CustomerID: "cust-456",
				Items:      []string{"widget-a", "widget-b"},
				Total:      149.99,
				CreatedAt:  now,
			},
			Metadata: map[string]string{
				"correlation_id": "req-789",
				"user_id":        "cust-456",
			},
		},
		{
			ID:          "evt-002",
			Type:        "OrderPaid",
			AggregateID: "order-123",
			Version:     2,
			Timestamp:   now.Add(5 * time.Minute),
			Data: OrderPaid{
				OrderID: "order-123",
				Amount:  149.99,
				Method:  "credit_card",
				PaidAt:  now.Add(5 * time.Minute),
			},
			Metadata: map[string]string{
				"correlation_id": "req-789",
			},
		},
		{
			ID:          "evt-003",
			Type:        "OrderShipped",
			AggregateID: "order-123",
			Version:     3,
			Timestamp:   now.Add(2 * time.Hour),
			Data: OrderShipped{
				OrderID:    "order-123",
				TrackingNo: "TRACK-001",
				ShippedAt:  now.Add(2 * time.Hour),
			},
			Metadata: map[string]string{
				"correlation_id": "req-789",
			},
		},
	}

	fmt.Println("Stream de eventos para order-123:")
	for _, e := range events {
		data, _ := json.Marshal(e.Data)
		fmt.Printf("  v%d [%s] %s -> %s\n", e.Version, e.Timestamp.Format("15:04:05"), e.Type, string(data))
	}

	envelope := EventEnvelope{
		ID:            events[0].ID,
		Type:          events[0].Type,
		AggregateID:   events[0].AggregateID,
		Version:       events[0].Version,
		Timestamp:     events[0].Timestamp,
		CorrelationID: events[0].Metadata["correlation_id"],
		CausationID:   "",
	}
	raw, _ := json.Marshal(events[0].Data)
	envelope.Data = raw

	envJSON, _ := json.MarshalIndent(envelope, "  ", "  ")
	fmt.Printf("\n  Event Envelope (JSON):\n  %s\n", string(envJSON))
}

// ============================================
// DEMO: EVENT BUS (IN-MEMORY)
// ============================================

type EventHandler func(event Event)

type InMemoryEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]EventHandler
	allHandlers []EventHandler
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		subscribers: make(map[string][]EventHandler),
	}
}

func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers[eventType] = append(bus.subscribers[eventType], handler)
}

func (bus *InMemoryEventBus) SubscribeAll(handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.allHandlers = append(bus.allHandlers, handler)
}

func (bus *InMemoryEventBus) Publish(event Event) {
	bus.mu.RLock()
	handlers := make([]EventHandler, 0)
	handlers = append(handlers, bus.subscribers[event.Type]...)
	handlers = append(handlers, bus.allHandlers...)
	bus.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

func (bus *InMemoryEventBus) PublishAsync(event Event) <-chan struct{} {
	done := make(chan struct{})

	bus.mu.RLock()
	handlers := make([]EventHandler, 0)
	handlers = append(handlers, bus.subscribers[event.Type]...)
	handlers = append(handlers, bus.allHandlers...)
	bus.mu.RUnlock()

	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for _, h := range handlers {
			wg.Add(1)
			go func(handler EventHandler) {
				defer wg.Done()
				handler(event)
			}(h)
		}
		wg.Wait()
	}()

	return done
}

func demoEventBus() {
	bus := NewInMemoryEventBus()

	var mu sync.Mutex
	var log []string

	bus.Subscribe("OrderCreated", func(e Event) {
		mu.Lock()
		log = append(log, fmt.Sprintf("  [InventoryService] Reservando stock para %s", e.AggregateID))
		mu.Unlock()
	})

	bus.Subscribe("OrderCreated", func(e Event) {
		mu.Lock()
		log = append(log, fmt.Sprintf("  [NotificationService] Email enviado para %s", e.AggregateID))
		mu.Unlock()
	})

	bus.Subscribe("OrderPaid", func(e Event) {
		mu.Lock()
		log = append(log, fmt.Sprintf("  [ShippingService] Preparando envio para %s", e.AggregateID))
		mu.Unlock()
	})

	bus.SubscribeAll(func(e Event) {
		mu.Lock()
		log = append(log, fmt.Sprintf("  [AuditLog] Evento registrado: %s (aggregate: %s)", e.Type, e.AggregateID))
		mu.Unlock()
	})

	fmt.Println("Publicando eventos sincronos:")

	bus.Publish(Event{
		ID:          "evt-1",
		Type:        "OrderCreated",
		AggregateID: "order-100",
		Timestamp:   time.Now(),
	})

	bus.Publish(Event{
		ID:          "evt-2",
		Type:        "OrderPaid",
		AggregateID: "order-100",
		Timestamp:   time.Now(),
	})

	for _, entry := range log {
		fmt.Println(entry)
	}

	log = nil

	fmt.Println("\nPublicando evento asincrono:")
	done := bus.PublishAsync(Event{
		ID:          "evt-3",
		Type:        "OrderCreated",
		AggregateID: "order-200",
		Timestamp:   time.Now(),
	})
	<-done

	mu.Lock()
	for _, entry := range log {
		fmt.Println(entry)
	}
	mu.Unlock()

	fmt.Printf("  Total subscribers para OrderCreated: %d\n", len(bus.subscribers["OrderCreated"]))
	fmt.Printf("  Total subscribers globales: %d\n", len(bus.allHandlers))
}

// ============================================
// DEMO: CQRS COMPLETO
// ============================================

type Command interface {
	CommandName() string
}

type Query interface {
	QueryName() string
}

type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}

type QueryHandler interface {
	Handle(ctx context.Context, query Query) (any, error)
}

type CreateOrderCommand struct {
	OrderID    string
	CustomerID string
	Items      []OrderItem
}

func (c CreateOrderCommand) CommandName() string { return "CreateOrder" }

type CancelOrderCommand struct {
	OrderID string
	Reason  string
}

func (c CancelOrderCommand) CommandName() string { return "CancelOrder" }

type GetOrderQuery struct {
	OrderID string
}

func (q GetOrderQuery) QueryName() string { return "GetOrder" }

type ListOrdersByCustomerQuery struct {
	CustomerID string
}

func (q ListOrdersByCustomerQuery) QueryName() string { return "ListOrdersByCustomer" }

type OrderItem struct {
	ProductID string
	Name      string
	Quantity  int
	Price     float64
}

type OrderWriteModel struct {
	ID         string
	CustomerID string
	Items      []OrderItem
	Status     string
	Total      float64
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OrderReadModel struct {
	ID            string
	CustomerID    string
	CustomerName  string
	Status        string
	ItemCount     int
	Total         float64
	ItemSummary   string
	LastUpdated   time.Time
}

type WriteRepository struct {
	mu     sync.RWMutex
	orders map[string]*OrderWriteModel
	events []Event
	bus    *InMemoryEventBus
}

func NewWriteRepository(bus *InMemoryEventBus) *WriteRepository {
	return &WriteRepository{
		orders: make(map[string]*OrderWriteModel),
		bus:    bus,
	}
}

func (r *WriteRepository) Save(order *OrderWriteModel) {
	r.mu.Lock()
	r.orders[order.ID] = order
	r.mu.Unlock()
}

func (r *WriteRepository) GetByID(id string) (*OrderWriteModel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[id]
	return o, ok
}

type ReadRepository struct {
	mu     sync.RWMutex
	orders map[string]*OrderReadModel
}

func NewReadRepository() *ReadRepository {
	return &ReadRepository{
		orders: make(map[string]*OrderReadModel),
	}
}

func (r *ReadRepository) Upsert(view *OrderReadModel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[view.ID] = view
}

func (r *ReadRepository) GetByID(id string) (*OrderReadModel, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.orders[id]
	return v, ok
}

func (r *ReadRepository) FindByCustomer(customerID string) []*OrderReadModel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*OrderReadModel
	for _, v := range r.orders {
		if v.CustomerID == customerID {
			result = append(result, v)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

type CreateOrderHandler struct {
	writeRepo *WriteRepository
	bus       *InMemoryEventBus
}

func (h *CreateOrderHandler) Handle(_ context.Context, cmd Command) error {
	c, ok := cmd.(CreateOrderCommand)
	if !ok {
		return errors.New("invalid command type")
	}

	if c.OrderID == "" || c.CustomerID == "" {
		return errors.New("order_id and customer_id are required")
	}

	if _, exists := h.writeRepo.GetByID(c.OrderID); exists {
		return errors.New("order already exists")
	}

	var total float64
	for _, item := range c.Items {
		total += float64(item.Quantity) * item.Price
	}

	order := &OrderWriteModel{
		ID:         c.OrderID,
		CustomerID: c.CustomerID,
		Items:      c.Items,
		Status:     "created",
		Total:      total,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	h.writeRepo.Save(order)

	h.bus.Publish(Event{
		ID:          fmt.Sprintf("evt-%s-created", c.OrderID),
		Type:        "OrderCreated",
		AggregateID: c.OrderID,
		Version:     1,
		Timestamp:   order.CreatedAt,
		Data: OrderCreated{
			OrderID:    c.OrderID,
			CustomerID: c.CustomerID,
			Items:      itemNames(c.Items),
			Total:      total,
			CreatedAt:  order.CreatedAt,
		},
	})

	return nil
}

type CancelOrderHandler struct {
	writeRepo *WriteRepository
	bus       *InMemoryEventBus
}

func (h *CancelOrderHandler) Handle(_ context.Context, cmd Command) error {
	c, ok := cmd.(CancelOrderCommand)
	if !ok {
		return errors.New("invalid command type")
	}

	order, exists := h.writeRepo.GetByID(c.OrderID)
	if !exists {
		return errors.New("order not found")
	}

	if order.Status == "cancelled" {
		return errors.New("order already cancelled")
	}

	if order.Status == "shipped" {
		return errors.New("cannot cancel shipped order")
	}

	order.Status = "cancelled"
	order.Version++
	order.UpdatedAt = time.Now()
	h.writeRepo.Save(order)

	h.bus.Publish(Event{
		ID:          fmt.Sprintf("evt-%s-cancelled", c.OrderID),
		Type:        "OrderCancelled",
		AggregateID: c.OrderID,
		Version:     order.Version,
		Timestamp:   order.UpdatedAt,
		Data: map[string]string{
			"order_id": c.OrderID,
			"reason":   c.Reason,
		},
	})

	return nil
}

type GetOrderHandler struct {
	readRepo *ReadRepository
}

func (h *GetOrderHandler) Handle(_ context.Context, query Query) (any, error) {
	q, ok := query.(GetOrderQuery)
	if !ok {
		return nil, errors.New("invalid query type")
	}

	view, exists := h.readRepo.GetByID(q.OrderID)
	if !exists {
		return nil, errors.New("order not found")
	}

	return view, nil
}

type ListOrdersByCustomerHandler struct {
	readRepo *ReadRepository
}

func (h *ListOrdersByCustomerHandler) Handle(_ context.Context, query Query) (any, error) {
	q, ok := query.(ListOrdersByCustomerQuery)
	if !ok {
		return nil, errors.New("invalid query type")
	}

	return h.readRepo.FindByCustomer(q.CustomerID), nil
}

type OrderProjectionHandler struct {
	readRepo *ReadRepository
}

func NewOrderProjectionHandler(readRepo *ReadRepository) *OrderProjectionHandler {
	return &OrderProjectionHandler{readRepo: readRepo}
}

func (p *OrderProjectionHandler) HandleOrderCreated(e Event) {
	data, ok := e.Data.(OrderCreated)
	if !ok {
		return
	}

	p.readRepo.Upsert(&OrderReadModel{
		ID:           data.OrderID,
		CustomerID:   data.CustomerID,
		CustomerName: "Customer " + data.CustomerID,
		Status:       "created",
		ItemCount:    len(data.Items),
		Total:        data.Total,
		ItemSummary:  strings.Join(data.Items, ", "),
		LastUpdated:  data.CreatedAt,
	})
}

func (p *OrderProjectionHandler) HandleOrderCancelled(e Event) {
	data, ok := e.Data.(map[string]string)
	if !ok {
		return
	}

	orderID := data["order_id"]
	view, exists := p.readRepo.GetByID(orderID)
	if !exists {
		return
	}

	view.Status = "cancelled"
	view.LastUpdated = e.Timestamp
	p.readRepo.Upsert(view)
}

type CommandBus struct {
	handlers map[string]CommandHandler
}

func NewCommandBus() *CommandBus {
	return &CommandBus{handlers: make(map[string]CommandHandler)}
}

func (b *CommandBus) Register(cmdName string, handler CommandHandler) {
	b.handlers[cmdName] = handler
}

func (b *CommandBus) Dispatch(ctx context.Context, cmd Command) error {
	handler, ok := b.handlers[cmd.CommandName()]
	if !ok {
		return fmt.Errorf("no handler registered for command: %s", cmd.CommandName())
	}
	return handler.Handle(ctx, cmd)
}

type QueryBus struct {
	handlers map[string]QueryHandler
}

func NewQueryBus() *QueryBus {
	return &QueryBus{handlers: make(map[string]QueryHandler)}
}

func (b *QueryBus) Register(queryName string, handler QueryHandler) {
	b.handlers[queryName] = handler
}

func (b *QueryBus) Dispatch(ctx context.Context, query Query) (any, error) {
	handler, ok := b.handlers[query.QueryName()]
	if !ok {
		return nil, fmt.Errorf("no handler registered for query: %s", query.QueryName())
	}
	return handler.Handle(ctx, query)
}

func demoCQRS() {
	eventBus := NewInMemoryEventBus()
	writeRepo := NewWriteRepository(eventBus)
	readRepo := NewReadRepository()

	projection := NewOrderProjectionHandler(readRepo)
	eventBus.Subscribe("OrderCreated", projection.HandleOrderCreated)
	eventBus.Subscribe("OrderCancelled", projection.HandleOrderCancelled)

	commandBus := NewCommandBus()
	commandBus.Register("CreateOrder", &CreateOrderHandler{writeRepo: writeRepo, bus: eventBus})
	commandBus.Register("CancelOrder", &CancelOrderHandler{writeRepo: writeRepo, bus: eventBus})

	queryBus := NewQueryBus()
	queryBus.Register("GetOrder", &GetOrderHandler{readRepo: readRepo})
	queryBus.Register("ListOrdersByCustomer", &ListOrdersByCustomerHandler{readRepo: readRepo})

	ctx := context.Background()

	fmt.Println("1) Crear ordenes via Command Bus:")
	err := commandBus.Dispatch(ctx, CreateOrderCommand{
		OrderID:    "ord-001",
		CustomerID: "cust-1",
		Items: []OrderItem{
			{ProductID: "p1", Name: "Go Book", Quantity: 1, Price: 39.99},
			{ProductID: "p2", Name: "Keyboard", Quantity: 1, Price: 89.99},
		},
	})
	fmt.Printf("   CreateOrder ord-001: err=%v\n", err)

	err = commandBus.Dispatch(ctx, CreateOrderCommand{
		OrderID:    "ord-002",
		CustomerID: "cust-1",
		Items: []OrderItem{
			{ProductID: "p3", Name: "Monitor", Quantity: 2, Price: 299.99},
		},
	})
	fmt.Printf("   CreateOrder ord-002: err=%v\n", err)

	fmt.Println("\n2) Query: obtener orden del Read Model:")
	result, err := queryBus.Dispatch(ctx, GetOrderQuery{OrderID: "ord-001"})
	if err == nil {
		view := result.(*OrderReadModel)
		fmt.Printf("   Order: id=%s, status=%s, items=%d, total=%.2f, summary=%q\n",
			view.ID, view.Status, view.ItemCount, view.Total, view.ItemSummary)
	}

	fmt.Println("\n3) Query: listar ordenes de cust-1:")
	result, err = queryBus.Dispatch(ctx, ListOrdersByCustomerQuery{CustomerID: "cust-1"})
	if err == nil {
		orders := result.([]*OrderReadModel)
		for _, v := range orders {
			fmt.Printf("   Order: id=%s, status=%s, total=%.2f\n", v.ID, v.Status, v.Total)
		}
	}

	fmt.Println("\n4) Command: cancelar orden:")
	err = commandBus.Dispatch(ctx, CancelOrderCommand{OrderID: "ord-002", Reason: "changed mind"})
	fmt.Printf("   CancelOrder ord-002: err=%v\n", err)

	result, _ = queryBus.Dispatch(ctx, GetOrderQuery{OrderID: "ord-002"})
	view := result.(*OrderReadModel)
	fmt.Printf("   After cancel: id=%s, status=%s\n", view.ID, view.Status)

	fmt.Println("\n5) Command: intentar duplicar orden:")
	err = commandBus.Dispatch(ctx, CreateOrderCommand{OrderID: "ord-001", CustomerID: "cust-1"})
	fmt.Printf("   CreateOrder ord-001 (duplicate): err=%v\n", err)

	fmt.Println("\n   Write model (state actual):")
	wo, _ := writeRepo.GetByID("ord-001")
	fmt.Printf("   order=%s, status=%s, version=%d, items=%d\n", wo.ID, wo.Status, wo.Version, len(wo.Items))

	fmt.Println("   Read model (optimizado para queries):")
	ro, _ := readRepo.GetByID("ord-001")
	fmt.Printf("   order=%s, customer_name=%q, item_summary=%q\n", ro.ID, ro.CustomerName, ro.ItemSummary)
}

// ============================================
// DEMO: EVENT SOURCING
// ============================================

type BankAccountAggregate struct {
	id          string
	owner       string
	balance     float64
	closed      bool
	version     int
	uncommitted []Event
}

func NewBankAccountAggregate(id, owner string) *BankAccountAggregate {
	a := &BankAccountAggregate{}
	a.apply(Event{
		Type:        "AccountOpened",
		AggregateID: id,
		Timestamp:   time.Now(),
		Data: map[string]string{
			"account_id": id,
			"owner":      owner,
		},
	})
	return a
}

func (a *BankAccountAggregate) Deposit(amount float64) error {
	if a.closed {
		return errors.New("account is closed")
	}
	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	a.apply(Event{
		Type:        "MoneyDeposited",
		AggregateID: a.id,
		Timestamp:   time.Now(),
		Data: map[string]any{
			"account_id": a.id,
			"amount":     amount,
		},
	})
	return nil
}

func (a *BankAccountAggregate) Withdraw(amount float64) error {
	if a.closed {
		return errors.New("account is closed")
	}
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if amount > a.balance {
		return errors.New("insufficient funds")
	}

	a.apply(Event{
		Type:        "MoneyWithdrawn",
		AggregateID: a.id,
		Timestamp:   time.Now(),
		Data: map[string]any{
			"account_id": a.id,
			"amount":     amount,
		},
	})
	return nil
}

func (a *BankAccountAggregate) Close() error {
	if a.closed {
		return errors.New("account already closed")
	}
	if a.balance != 0 {
		return errors.New("balance must be zero to close")
	}

	a.apply(Event{
		Type:        "AccountClosed",
		AggregateID: a.id,
		Timestamp:   time.Now(),
		Data: map[string]string{
			"account_id": a.id,
		},
	})
	return nil
}

func (a *BankAccountAggregate) apply(event Event) {
	a.version++
	event.Version = a.version
	a.when(event)
	a.uncommitted = append(a.uncommitted, event)
}

func (a *BankAccountAggregate) when(event Event) {
	switch event.Type {
	case "AccountOpened":
		data := event.Data.(map[string]string)
		a.id = data["account_id"]
		a.owner = data["owner"]
		a.balance = 0
		a.closed = false

	case "MoneyDeposited":
		data := event.Data.(map[string]any)
		a.balance += data["amount"].(float64)

	case "MoneyWithdrawn":
		data := event.Data.(map[string]any)
		a.balance -= data["amount"].(float64)

	case "AccountClosed":
		a.closed = true
	}
}

func (a *BankAccountAggregate) GetUncommitted() []Event {
	return a.uncommitted
}

func (a *BankAccountAggregate) ClearUncommitted() {
	a.uncommitted = nil
}

type InMemoryEventStore struct {
	mu     sync.RWMutex
	events map[string][]Event
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]Event),
	}
}

func (s *InMemoryEventStore) Append(aggregateID string, events []Event, expectedVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.events[aggregateID]
	currentVersion := len(existing)

	if currentVersion != expectedVersion {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", expectedVersion, currentVersion)
	}

	s.events[aggregateID] = append(existing, events...)
	return nil
}

func (s *InMemoryEventStore) Load(aggregateID string) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.events[aggregateID]
	result := make([]Event, len(events))
	copy(result, events)
	return result
}

func (s *InMemoryEventStore) LoadFrom(aggregateID string, fromVersion int) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := s.events[aggregateID]
	if fromVersion > len(all) {
		return nil
	}
	result := make([]Event, len(all)-fromVersion)
	copy(result, all[fromVersion:])
	return result
}

func (s *InMemoryEventStore) LoadAll() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var all []Event
	for _, events := range s.events {
		all = append(all, events...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.Before(all[j].Timestamp)
	})
	return all
}

func demoEventSourcing() {
	store := NewInMemoryEventStore()

	account := NewBankAccountAggregate("acc-001", "Alice")
	_ = account.Deposit(200)
	_ = account.Deposit(50)
	_ = account.Withdraw(30)

	uncommitted := account.GetUncommitted()
	err := store.Append("acc-001", uncommitted, 0)
	fmt.Printf("Persistir %d eventos: err=%v\n", len(uncommitted), err)
	account.ClearUncommitted()

	fmt.Printf("Estado actual: owner=%s, balance=%.2f, version=%d\n",
		account.owner, account.balance, account.version)

	fmt.Println("\nReconstruir desde event store:")
	events := store.Load("acc-001")
	reconstructed := &BankAccountAggregate{}
	for _, e := range events {
		reconstructed.version = e.Version
		reconstructed.when(e)
	}
	fmt.Printf("Reconstruido: owner=%s, balance=%.2f, version=%d\n",
		reconstructed.owner, reconstructed.balance, reconstructed.version)

	fmt.Printf("Estado identico: %v\n",
		account.balance == reconstructed.balance &&
			account.owner == reconstructed.owner &&
			account.version == reconstructed.version)

	_ = account.Withdraw(220)
	_ = account.Close()
	uncommitted = account.GetUncommitted()
	err = store.Append("acc-001", uncommitted, 4)
	fmt.Printf("\nPersistir %d eventos mas: err=%v\n", len(uncommitted), err)
	account.ClearUncommitted()
	fmt.Printf("Estado final: balance=%.2f, closed=%v, version=%d\n",
		account.balance, account.closed, account.version)

	fmt.Println("\nValidacion de reglas de negocio:")
	err = account.Deposit(100)
	fmt.Printf("  Deposit en cuenta cerrada: %v\n", err)

	account2 := NewBankAccountAggregate("acc-002", "Bob")
	_ = account2.Deposit(50)
	err = account2.Withdraw(100)
	fmt.Printf("  Withdraw sin fondos: %v\n", err)

	err = account2.Close()
	fmt.Printf("  Close con balance > 0: %v\n", err)

	fmt.Println("\nConcurrency check:")
	err = store.Append("acc-001", []Event{{Type: "test"}}, 0)
	fmt.Printf("  Append con version incorrecta: %v\n", err)
}

// ============================================
// DEMO: REPLAY DE EVENTOS
// ============================================

type BalanceProjection struct {
	balances map[string]float64
}

func NewBalanceProjection() *BalanceProjection {
	return &BalanceProjection{
		balances: make(map[string]float64),
	}
}

func (p *BalanceProjection) Handle(event Event) {
	switch event.Type {
	case "AccountOpened":
		data := event.Data.(map[string]string)
		p.balances[data["account_id"]] = 0

	case "MoneyDeposited":
		data := event.Data.(map[string]any)
		p.balances[data["account_id"].(string)] += data["amount"].(float64)

	case "MoneyWithdrawn":
		data := event.Data.(map[string]any)
		p.balances[data["account_id"].(string)] -= data["amount"].(float64)
	}
}

type AuditProjection struct {
	entries []string
}

func NewAuditProjection() *AuditProjection {
	return &AuditProjection{}
}

func (p *AuditProjection) Handle(event Event) {
	entry := fmt.Sprintf("[v%d] %s on %s", event.Version, event.Type, event.AggregateID)
	p.entries = append(p.entries, entry)
}

func demoReplayEvents() {
	store := NewInMemoryEventStore()

	acc1 := NewBankAccountAggregate("acc-A", "Alice")
	_ = acc1.Deposit(500)
	_ = acc1.Withdraw(100)
	_ = store.Append("acc-A", acc1.GetUncommitted(), 0)

	acc2 := NewBankAccountAggregate("acc-B", "Bob")
	_ = acc2.Deposit(300)
	_ = acc2.Deposit(200)
	_ = store.Append("acc-B", acc2.GetUncommitted(), 0)

	allEvents := store.LoadAll()
	fmt.Printf("Total eventos en el store: %d\n", len(allEvents))

	fmt.Println("\n1) Balance Projection (replay completo):")
	balanceProj := NewBalanceProjection()
	for _, e := range allEvents {
		balanceProj.Handle(e)
	}
	ids := make([]string, 0, len(balanceProj.balances))
	for id := range balanceProj.balances {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Printf("   %s: balance=%.2f\n", id, balanceProj.balances[id])
	}

	fmt.Println("\n2) Audit Projection (replay completo):")
	auditProj := NewAuditProjection()
	for _, e := range allEvents {
		auditProj.Handle(e)
	}
	for _, entry := range auditProj.entries {
		fmt.Printf("   %s\n", entry)
	}

	fmt.Println("\n3) Temporal Query (estado de acc-A hasta version 2):")
	accAEvents := store.Load("acc-A")
	tempAccount := &BankAccountAggregate{}
	for _, e := range accAEvents {
		if e.Version > 2 {
			break
		}
		tempAccount.version = e.Version
		tempAccount.when(e)
	}
	fmt.Printf("   acc-A at v2: owner=%s, balance=%.2f\n", tempAccount.owner, tempAccount.balance)

	fmt.Println("\n4) Rebuild projection (nueva vista del mismo stream):")
	fmt.Println("   Si necesitas una nueva vista (ej: revenue report),")
	fmt.Println("   simplemente replay todos los eventos con un nuevo handler.")
	fmt.Printf("   Eventos disponibles para replay: %d\n", len(allEvents))
}

// ============================================
// HELPERS
// ============================================

func itemNames(items []OrderItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
	}
	return names
}

var _ = reflect.TypeOf
var _ = os.Stdout
var _ = strings.Join
var _ = json.Marshal
var _ = sort.Strings
var _ = context.Background
var _ = errors.New
var _ = time.Now

/*
RESUMEN CAPITULO 68: EVENT-DRIVEN ARCHITECTURE, CQRS & EVENT SOURCING

EVENT-DRIVEN ARCHITECTURE:
- Comunicacion asincrona via eventos entre componentes
- Producer emite, Consumer reacciona, Bus transporta
- Desacoplamiento, escalabilidad, auditoria natural
- Event Notification vs Event-Carried State Transfer

DOMAIN EVENTS:
- Registro inmutable de algo que ocurrio
- Nombrados en pasado: OrderPlaced, PaymentReceived
- Estructura: ID, Type, AggregateID, Version, Timestamp, Data, Metadata
- Correlation ID para tracing end-to-end

EVENT BUS:
- Desacopla publishers de subscribers
- Fan-out (todos reciben), Competing (uno procesa)
- In-memory con channels para desarrollo/testing
- Produccion: Kafka, NATS, RabbitMQ

CQRS:
- Separa modelos de escritura (Commands) y lectura (Queries)
- Commands: imperativo (CreateOrder), puede fallar, retorna void/error
- Queries: GetOrderByID, sin side effects, optimizado para lectura
- Command Bus y Query Bus para dispatch
- Eventual consistency entre write y read model

EVENT SOURCING:
- Almacenar TODOS los eventos, no solo estado final
- Event Store con optimistic concurrency (version)
- Aggregate Root: valida, genera eventos, reconstruye estado
- Projections: vistas materializadas desde stream de eventos
- Snapshots: estado pre-calculado para evitar replay completo
- Replay: reconstruir projections, debugging, time travel

SAGA PATTERN:
- Transacciones distribuidas con compensaciones
- Orchestration: coordinador central, flujo claro
- Choreography: eventos entre servicios, mas desacoplado
- Saga State Machine: Started -> InProgress -> Completed/Compensated

OUTBOX PATTERN:
- Resuelve dual-write problem (DB + message broker)
- Evento en misma transaccion que el cambio
- Relay separado publica eventos pendientes
- Alternativa: CDC con Debezium

IDEMPOTENCIA:
- At-least-once delivery requiere handlers idempotentes
- Deduplication table con event_id
- Natural idempotency: SET vs ADD
- Conditional updates con version checks

EVENT VERSIONING:
- Eventos son inmutables, schema cambia
- Weak schema: ignorar campos desconocidos, defaults
- Upcasting: transformar al leer
- Nuevo event type para cambios grandes
- NUNCA eliminar/cambiar campos existentes

LIBRERIAS:
- Watermill (ThreeDotsLabs): Pub/Sub abstraction, Router, CQRS module
- watermill-kafka, watermill-nats, watermill-amqp
- sony/gobreaker para circuit breaker en sagas

BROKERS:
- Kafka:    Alto throughput, event sourcing excelente, ops complejo
- NATS:     Muy baja latencia, JetStream, simple ops
- RabbitMQ: Flexible routing, dead letter queues, quorum queues

TESTING:
- Given/When/Then para aggregates
- In-memory Event Bus para integration tests
- Contract testing para schemas de eventos
- Testcontainers para Kafka/NATS/RabbitMQ
*/

/* SUMMARY - EVENT-DRIVEN ARCHITECTURE, CQRS & EVENT SOURCING:

TOPIC: Arquitectura orientada a eventos y separación de responsabilidades

EVENT-DRIVEN ARCHITECTURE:
- Comunicación asíncrona via eventos (desacoplamiento, escalabilidad)
- Domain Events en pasado: OrderPlaced, PaymentReceived
- Event Bus: Fan-out (todos reciben) vs Competing (uno procesa)
- Brokers: Kafka (alto throughput), NATS (baja latencia), RabbitMQ (flexible)

CQRS (COMMAND QUERY RESPONSIBILITY SEGREGATION):
- Commands: imperativo (CreateOrder), puede fallar
- Queries: lectura sin side effects (GetOrderByID)
- Command Bus y Query Bus para dispatch
- Write Model vs Read Model (eventual consistency)

EVENT SOURCING:
- Almacenar TODOS los eventos, no solo estado actual
- Event Store con optimistic concurrency (versioning)
- Aggregate Root: valida comandos, genera eventos
- Projections: vistas materializadas desde stream de eventos
- Snapshots: evitar replay completo (cada N eventos)

SAGA PATTERN:
- Transacciones distribuidas con compensaciones
- Orchestration: coordinador central
- Choreography: eventos entre servicios (más desacoplado)

OUTBOX PATTERN:
- Resolver dual-write problem (DB + broker)
- Evento en misma transacción que el cambio
- Relay separado publica eventos pendientes

IDEMPOTENCIA & VERSIONING:
- Deduplication table con event_id
- Upcasting: transformar eventos al leer
- NUNCA eliminar/cambiar campos existentes

TESTING:
- In-memory Event Bus para integration tests
- Watermill para Pub/Sub abstraction
*/
