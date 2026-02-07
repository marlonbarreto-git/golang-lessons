// Package main - Chapter 075: Design Patterns
// Patrones de diseno idiomaticos en Go: functional options, builder,
// factory, singleton, strategy, observer, decorator, adapter,
// repository, unit of work, dependency injection, configuration,
// graceful shutdown, health checks y table-driven factories.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ============================================
// FUNCTIONAL OPTIONS PATTERN (Dave Cheney)
// ============================================

type LogLevel int

const (
	LogInfo LogLevel = iota
	LogDebug
	LogWarn
	LogError
)

type Logger interface {
	Printf(format string, args ...any)
}

type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, args ...any) {
	fmt.Printf("[default-logger] "+format+"\n", args...)
}

type ServerConfig struct {
	host         string
	port         int
	timeout      time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	maxRetries   int
	logger       Logger
	logLevel     LogLevel
	tlsEnabled   bool
	certFile     string
	keyFile      string
}

type ServerOption func(*ServerConfig)

func WithHost(host string) ServerOption {
	return func(c *ServerConfig) {
		c.host = host
	}
}

func WithPort(port int) ServerOption {
	return func(c *ServerConfig) {
		c.port = port
	}
}

func WithTimeout(d time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.timeout = d
	}
}

func WithReadTimeout(d time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.readTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.writeTimeout = d
	}
}

func WithMaxRetries(n int) ServerOption {
	return func(c *ServerConfig) {
		c.maxRetries = n
	}
}

func WithLogger(l Logger) ServerOption {
	return func(c *ServerConfig) {
		c.logger = l
	}
}

func WithLogLevel(level LogLevel) ServerOption {
	return func(c *ServerConfig) {
		c.logLevel = level
	}
}

func WithTLS(certFile, keyFile string) ServerOption {
	return func(c *ServerConfig) {
		c.tlsEnabled = true
		c.certFile = certFile
		c.keyFile = keyFile
	}
}

type Server struct {
	config ServerConfig
}

func NewServer(opts ...ServerOption) *Server {
	cfg := ServerConfig{
		host:         "localhost",
		port:         8080,
		timeout:      30 * time.Second,
		readTimeout:  10 * time.Second,
		writeTimeout: 10 * time.Second,
		maxRetries:   3,
		logger:       &defaultLogger{},
		logLevel:     LogInfo,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Server{config: cfg}
}

func (s *Server) Info() string {
	return fmt.Sprintf("Server{host:%s, port:%d, timeout:%s, retries:%d, tls:%v}",
		s.config.host, s.config.port, s.config.timeout,
		s.config.maxRetries, s.config.tlsEnabled)
}

// ============================================
// BUILDER PATTERN
// ============================================

type QueryBuilder struct {
	table      string
	conditions []string
	orderBy    string
	limit      int
	offset     int
	fields     []string
	joins      []string
}

func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:  table,
		fields: []string{"*"},
	}
}

func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.fields = fields
	return qb
}

func (qb *QueryBuilder) Where(condition string) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition)
	return qb
}

func (qb *QueryBuilder) Join(join string) *QueryBuilder {
	qb.joins = append(qb.joins, join)
	return qb
}

func (qb *QueryBuilder) OrderByField(field string) *QueryBuilder {
	qb.orderBy = field
	return qb
}

func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.limit = n
	return qb
}

func (qb *QueryBuilder) Offset(n int) *QueryBuilder {
	qb.offset = n
	return qb
}

func (qb *QueryBuilder) Build() string {
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(qb.fields, ", "), qb.table)

	for _, j := range qb.joins {
		query += " JOIN " + j
	}

	if len(qb.conditions) > 0 {
		query += " WHERE " + strings.Join(qb.conditions, " AND ")
	}

	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	if qb.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", qb.offset)
	}

	return query
}

// ============================================
// FACTORY PATTERN
// ============================================

type NotificationType string

const (
	NotificationEmail NotificationType = "email"
	NotificationSMS   NotificationType = "sms"
	NotificationPush  NotificationType = "push"
	NotificationSlack NotificationType = "slack"
)

type Notification interface {
	Send(to, message string) error
	Type() string
}

type EmailNotification struct{}

func (e *EmailNotification) Send(to, message string) error {
	fmt.Printf("    [EMAIL] To: %s -> %s\n", to, message)
	return nil
}
func (e *EmailNotification) Type() string { return "email" }

type SMSNotification struct{}

func (s *SMSNotification) Send(to, message string) error {
	fmt.Printf("    [SMS] To: %s -> %s\n", to, message)
	return nil
}
func (s *SMSNotification) Type() string { return "sms" }

type PushNotification struct{}

func (p *PushNotification) Send(to, message string) error {
	fmt.Printf("    [PUSH] To: %s -> %s\n", to, message)
	return nil
}
func (p *PushNotification) Type() string { return "push" }

type SlackNotification struct{}

func (s *SlackNotification) Send(to, message string) error {
	fmt.Printf("    [SLACK] To: %s -> %s\n", to, message)
	return nil
}
func (s *SlackNotification) Type() string { return "slack" }

func NewNotification(ntype NotificationType) (Notification, error) {
	switch ntype {
	case NotificationEmail:
		return &EmailNotification{}, nil
	case NotificationSMS:
		return &SMSNotification{}, nil
	case NotificationPush:
		return &PushNotification{}, nil
	case NotificationSlack:
		return &SlackNotification{}, nil
	default:
		return nil, fmt.Errorf("unknown notification type: %s", ntype)
	}
}

// ============================================
// SINGLETON PATTERN (sync.Once)
// ============================================

type DBConnection struct {
	host string
	port int
}

func (db *DBConnection) String() string {
	return fmt.Sprintf("DBConnection{host:%s, port:%d}", db.host, db.port)
}

var (
	dbInstance *DBConnection
	dbOnce     sync.Once
)

func GetDB() *DBConnection {
	dbOnce.Do(func() {
		fmt.Println("    Initializing database connection (only once)...")
		dbInstance = &DBConnection{
			host: "localhost",
			port: 5432,
		}
	})
	return dbInstance
}

// ============================================
// STRATEGY PATTERN
// ============================================

type CompressionStrategy interface {
	Compress(data []byte) ([]byte, error)
	Name() string
}

type GzipCompression struct{}

func (g *GzipCompression) Compress(data []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("[gzip:%d bytes]", len(data))), nil
}
func (g *GzipCompression) Name() string { return "gzip" }

type ZstdCompression struct{}

func (z *ZstdCompression) Compress(data []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("[zstd:%d bytes]", len(data))), nil
}
func (z *ZstdCompression) Name() string { return "zstd" }

type SnappyCompression struct{}

func (s *SnappyCompression) Compress(data []byte) ([]byte, error) {
	return []byte(fmt.Sprintf("[snappy:%d bytes]", len(data))), nil
}
func (s *SnappyCompression) Name() string { return "snappy" }

type FileProcessor struct {
	strategy CompressionStrategy
}

func NewFileProcessor(strategy CompressionStrategy) *FileProcessor {
	return &FileProcessor{strategy: strategy}
}

func (fp *FileProcessor) SetStrategy(strategy CompressionStrategy) {
	fp.strategy = strategy
}

func (fp *FileProcessor) Process(data []byte) ([]byte, error) {
	return fp.strategy.Compress(data)
}

// ============================================
// OBSERVER PATTERN (channels)
// ============================================

type Event struct {
	Type    string
	Payload any
}

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

func (eb *EventBus) Subscribe(eventType string) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, 16)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	return ch
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	for _, ch := range eb.subscribers[event.Type] {
		select {
		case ch <- event:
		default:
		}
	}
}

func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, subs := range eb.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
}

// ============================================
// DECORATOR PATTERN (middleware)
// ============================================

type HandlerFunc func(ctx context.Context, req string) (string, error)

func WithLogging(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, req string) (string, error) {
		start := time.Now()
		fmt.Printf("    [LOG] request=%s\n", req)
		resp, err := next(ctx, req)
		fmt.Printf("    [LOG] response=%s duration=%s err=%v\n", resp, time.Since(start), err)
		return resp, err
	}
}

func WithRetry(maxRetries int, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, req string) (string, error) {
		var resp string
		var err error
		for i := 0; i <= maxRetries; i++ {
			resp, err = next(ctx, req)
			if err == nil {
				return resp, nil
			}
			fmt.Printf("    [RETRY] attempt %d/%d failed: %v\n", i+1, maxRetries+1, err)
		}
		return resp, err
	}
}

func WithTimeoutDecorator(d time.Duration, next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, req string) (string, error) {
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()

		type result struct {
			resp string
			err  error
		}

		ch := make(chan result, 1)
		go func() {
			r, e := next(ctx, req)
			ch <- result{r, e}
		}()

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case res := <-ch:
			return res.resp, res.err
		}
	}
}

// ============================================
// ADAPTER PATTERN
// ============================================

type OldPaymentGateway struct{}

func (o *OldPaymentGateway) MakePayment(amount float64, currency string, cardNumber string) (string, error) {
	return fmt.Sprintf("OLD-TXN-%s-%.2f", currency, amount), nil
}

type PaymentRequest struct {
	Amount      float64
	Currency    string
	CardToken   string
	Description string
}

type PaymentResult struct {
	TransactionID string
	Status        string
}

type PaymentProcessor interface {
	ProcessPayment(req PaymentRequest) (*PaymentResult, error)
}

type PaymentAdapter struct {
	legacy *OldPaymentGateway
}

func NewPaymentAdapter(legacy *OldPaymentGateway) PaymentProcessor {
	return &PaymentAdapter{legacy: legacy}
}

func (a *PaymentAdapter) ProcessPayment(req PaymentRequest) (*PaymentResult, error) {
	txnID, err := a.legacy.MakePayment(req.Amount, req.Currency, req.CardToken)
	if err != nil {
		return nil, err
	}
	return &PaymentResult{
		TransactionID: txnID,
		Status:        "completed",
	}, nil
}

// ============================================
// REPOSITORY PATTERN
// ============================================

type User struct {
	ID    string
	Name  string
	Email string
}

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*User),
	}
}

func (r *InMemoryUserRepository) FindByID(_ context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return u, nil
}

func (r *InMemoryUserRepository) FindByEmail(_ context.Context, email string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, fmt.Errorf("user not found with email: %s", email)
}

func (r *InMemoryUserRepository) FindAll(_ context.Context) ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*User, 0, len(r.users))
	for _, u := range r.users {
		result = append(result, u)
	}
	return result, nil
}

func (r *InMemoryUserRepository) Create(_ context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; exists {
		return fmt.Errorf("user already exists: %s", user.ID)
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) Update(_ context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[id]; !exists {
		return fmt.Errorf("user not found: %s", id)
	}
	delete(r.users, id)
	return nil
}

// ============================================
// UNIT OF WORK PATTERN
// ============================================

type Operation struct {
	Type   string
	Entity string
	Data   any
}

type UnitOfWork struct {
	mu         sync.Mutex
	operations []Operation
	committed  bool
}

func NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{}
}

func (uow *UnitOfWork) RegisterNew(entity string, data any) {
	uow.mu.Lock()
	defer uow.mu.Unlock()
	uow.operations = append(uow.operations, Operation{
		Type:   "INSERT",
		Entity: entity,
		Data:   data,
	})
}

func (uow *UnitOfWork) RegisterDirty(entity string, data any) {
	uow.mu.Lock()
	defer uow.mu.Unlock()
	uow.operations = append(uow.operations, Operation{
		Type:   "UPDATE",
		Entity: entity,
		Data:   data,
	})
}

func (uow *UnitOfWork) RegisterDeleted(entity string, data any) {
	uow.mu.Lock()
	defer uow.mu.Unlock()
	uow.operations = append(uow.operations, Operation{
		Type:   "DELETE",
		Entity: entity,
		Data:   data,
	})
}

func (uow *UnitOfWork) Commit() error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.committed {
		return fmt.Errorf("unit of work already committed")
	}

	fmt.Printf("    Committing %d operations:\n", len(uow.operations))
	for i, op := range uow.operations {
		fmt.Printf("      [%d] %s %s -> %v\n", i+1, op.Type, op.Entity, op.Data)
	}

	uow.committed = true
	return nil
}

func (uow *UnitOfWork) Rollback() {
	uow.mu.Lock()
	defer uow.mu.Unlock()
	uow.operations = nil
	fmt.Println("    Rolled back all operations")
}

// ============================================
// DEPENDENCY INJECTION (without frameworks)
// ============================================

type Mailer interface {
	SendMail(to, subject, body string) error
}

type ConsoleMailer struct{}

func (m *ConsoleMailer) SendMail(to, subject, body string) error {
	fmt.Printf("    [MAIL] To:%s Subject:%s Body:%s\n", to, subject, body)
	return nil
}

type UserService struct {
	repo   UserRepository
	mailer Mailer
	logger Logger
}

func NewUserService(repo UserRepository, mailer Mailer, logger Logger) *UserService {
	return &UserService{
		repo:   repo,
		mailer: mailer,
		logger: logger,
	}
}

func (s *UserService) Register(ctx context.Context, name, email string) (*User, error) {
	user := &User{
		ID:    fmt.Sprintf("usr_%d", time.Now().UnixNano()),
		Name:  name,
		Email: email,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	s.logger.Printf("user registered: %s", user.ID)

	if err := s.mailer.SendMail(email, "Welcome!", "Welcome to the platform, "+name); err != nil {
		s.logger.Printf("failed to send welcome email: %v", err)
	}

	return user, nil
}

// ============================================
// CONFIGURATION MANAGEMENT
// ============================================

type AppConfig struct {
	Server   ServerConf
	Database DatabaseConf
	Redis    RedisConf
	Log      LogConf
}

type ServerConf struct {
	Host string
	Port int
}

type DatabaseConf struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConf struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type LogConf struct {
	Level  string
	Format string
}

func LoadConfigFromEnv() *AppConfig {
	getEnv := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}

	return &AppConfig{
		Server: ServerConf{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: 8080,
		},
		Database: DatabaseConf{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     5432,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "myapp"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConf{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: 6379,
		},
		Log: LogConf{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// ============================================
// GRACEFUL SHUTDOWN PATTERN
// ============================================

type GracefulServer struct {
	httpServer *http.Server
	logger     *log.Logger
	cleanups   []func()
	mu         sync.Mutex
}

func NewGracefulServer(addr string, handler http.Handler) *GracefulServer {
	return &GracefulServer{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: log.New(os.Stdout, "[server] ", log.LstdFlags),
	}
}

func (gs *GracefulServer) RegisterCleanup(fn func()) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.cleanups = append(gs.cleanups, fn)
}

func (gs *GracefulServer) Start(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		gs.logger.Printf("listening on %s", gs.httpServer.Addr)
		if err := gs.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		gs.logger.Printf("received signal: %v", sig)
	case <-ctx.Done():
		gs.logger.Printf("context cancelled")
	}

	return gs.shutdown()
}

func (gs *GracefulServer) shutdown() error {
	gs.logger.Println("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := gs.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	gs.mu.Lock()
	cleanups := gs.cleanups
	gs.mu.Unlock()

	for i := len(cleanups) - 1; i >= 0; i-- {
		cleanups[i]()
	}

	gs.logger.Println("shutdown complete")
	return nil
}

// ============================================
// HEALTH CHECK PATTERN
// ============================================

type HealthStatus string

const (
	HealthUp   HealthStatus = "up"
	HealthDown HealthStatus = "down"
)

type HealthCheck struct {
	Name   string
	Status HealthStatus
	Detail string
}

type HealthChecker interface {
	Check(ctx context.Context) HealthCheck
}

type DatabaseHealthChecker struct {
	name      string
	connected bool
}

func (d *DatabaseHealthChecker) Check(_ context.Context) HealthCheck {
	if d.connected {
		return HealthCheck{Name: d.name, Status: HealthUp, Detail: "connection pool active"}
	}
	return HealthCheck{Name: d.name, Status: HealthDown, Detail: "cannot reach database"}
}

type RedisHealthChecker struct {
	name      string
	connected bool
}

func (r *RedisHealthChecker) Check(_ context.Context) HealthCheck {
	if r.connected {
		return HealthCheck{Name: r.name, Status: HealthUp, Detail: "PONG received"}
	}
	return HealthCheck{Name: r.name, Status: HealthDown, Detail: "connection refused"}
}

type HealthEndpoint struct {
	checkers []HealthChecker
}

func NewHealthEndpoint(checkers ...HealthChecker) *HealthEndpoint {
	return &HealthEndpoint{checkers: checkers}
}

func (h *HealthEndpoint) Liveness() HealthCheck {
	return HealthCheck{Name: "liveness", Status: HealthUp, Detail: "process is running"}
}

func (h *HealthEndpoint) Readiness(ctx context.Context) (HealthStatus, []HealthCheck) {
	results := make([]HealthCheck, 0, len(h.checkers))
	overallStatus := HealthUp

	for _, checker := range h.checkers {
		check := checker.Check(ctx)
		results = append(results, check)
		if check.Status == HealthDown {
			overallStatus = HealthDown
		}
	}

	return overallStatus, results
}

// ============================================
// TABLE-DRIVEN FACTORY PATTERN
// ============================================

type Serializer interface {
	Serialize(data any) ([]byte, error)
	ContentType() string
}

type JSONSerializer struct{}

func (j *JSONSerializer) Serialize(data any) ([]byte, error) {
	return []byte(fmt.Sprintf(`{"data":"%v"}`, data)), nil
}
func (j *JSONSerializer) ContentType() string { return "application/json" }

type XMLSerializer struct{}

func (x *XMLSerializer) Serialize(data any) ([]byte, error) {
	return []byte(fmt.Sprintf(`<data>%v</data>`, data)), nil
}
func (x *XMLSerializer) ContentType() string { return "application/xml" }

type YAMLSerializer struct{}

func (y *YAMLSerializer) Serialize(data any) ([]byte, error) {
	return []byte(fmt.Sprintf("data: %v\n", data)), nil
}
func (y *YAMLSerializer) ContentType() string { return "application/yaml" }

var serializerRegistry = map[string]func() Serializer{
	"json": func() Serializer { return &JSONSerializer{} },
	"xml":  func() Serializer { return &XMLSerializer{} },
	"yaml": func() Serializer { return &YAMLSerializer{} },
}

func NewSerializer(format string) (Serializer, error) {
	factory, ok := serializerRegistry[format]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
	return factory(), nil
}

func RegisterSerializer(format string, factory func() Serializer) {
	serializerRegistry[format] = factory
}

// ============================================
// MAIN
// ============================================

func main() {
	fmt.Println("=== DESIGN PATTERNS EN GO ===")

	// ============================================
	// 1. FUNCTIONAL OPTIONS PATTERN
	// ============================================
	fmt.Println("\n--- 1. Functional Options Pattern (Dave Cheney) ---")
	fmt.Println(`
PROBLEMA:
Constructores con muchos parametros opcionales.

// Mal - demasiados parametros, muchos con zero values
func NewServer(host string, port int, timeout time.Duration,
    readTimeout time.Duration, writeTimeout time.Duration,
    maxRetries int, logger Logger, ...) *Server

// Mal - config struct obliga a crear struct siempre
func NewServer(cfg ServerConfig) *Server

SOLUCION: Functional Options
Cada opcion es una funcion que modifica la configuracion.

type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.timeout = d
    }
}

func NewServer(opts ...Option) *Server {
    cfg := defaultConfig()
    for _, opt := range opts {
        opt(&cfg)
    }
    return &Server{config: cfg}
}`)

	fmt.Println("\n  DEMO - Functional Options:")

	srv1 := NewServer()
	fmt.Printf("    Default: %s\n", srv1.Info())

	srv2 := NewServer(
		WithHost("0.0.0.0"),
		WithPort(9090),
		WithTimeout(60*time.Second),
		WithMaxRetries(5),
		WithLogLevel(LogDebug),
	)
	fmt.Printf("    Custom:  %s\n", srv2.Info())

	srv3 := NewServer(
		WithPort(443),
		WithTLS("/etc/ssl/cert.pem", "/etc/ssl/key.pem"),
	)
	fmt.Printf("    TLS:     %s\n", srv3.Info())

	os.Stdout.WriteString(`
REAL-WORLD EXAMPLES:

// gRPC
conn, err := grpc.Dial(
    "localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithBlock(),
    grpc.WithTimeout(5 * time.Second),
)

// HTTP Client
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:    100,
        IdleConnTimeout: 90 * time.Second,
    },
}

// Database (pgx)
config, _ := pgxpool.ParseConfig(dsn)
config.MaxConns = 25
config.MinConns = 5

// Zap Logger
logger, _ := zap.NewProduction(
    zap.AddCaller(),
    zap.AddStacktrace(zap.ErrorLevel),
    zap.Fields(zap.String("service", "api")),
)

VARIANTES:

// Con validacion
type Option func(*Config) error

func WithPort(port int) Option {
    return func(c *Config) error {
        if port < 1 || port > 65535 {
            return fmt.Errorf("invalid port: %d", port)
        }
        c.port = port
        return nil
    }
}

// Con interface (para testear)
type Option interface {
    apply(*Config)
}

type optionFunc func(*Config)
func (f optionFunc) apply(c *Config) { f(c) }
`)

	// ============================================
	// 2. BUILDER PATTERN
	// ============================================
	fmt.Println("\n--- 2. Builder Pattern ---")
	fmt.Println(`
Builder encadena metodos para construir objetos complejos paso a paso.
Cada metodo retorna el builder (*QueryBuilder) para method chaining.
El metodo Build() produce el resultado final.`)

	fmt.Println("\n  DEMO - Query Builder:")
	query := NewQueryBuilder("users").
		Select("id", "name", "email").
		Join("orders ON orders.user_id = users.id").
		Where("users.active = true").
		Where("users.age > 18").
		OrderByField("users.name ASC").
		Limit(10).
		Offset(20).
		Build()
	fmt.Printf("    %s\n", query)

	query2 := NewQueryBuilder("products").
		Select("*").
		Where("price > 100").
		Limit(5).
		Build()
	fmt.Printf("    %s\n", query2)

	fmt.Println(`
BUILDER vs FUNCTIONAL OPTIONS:

Builder:
- Para construir objetos inmutables paso a paso
- Cuando el orden de construccion importa
- Build() puede validar el estado completo
- Ejemplo: query builders, request builders

Functional Options:
- Para configurar servicios/servers con defaults
- Cuando el orden NO importa
- Cada opcion es independiente
- Ejemplo: NewServer, grpc.Dial, zap.New`)

	// ============================================
	// 3. FACTORY PATTERN
	// ============================================
	fmt.Println("\n--- 3. Factory Pattern ---")
	fmt.Println(`
Factory crea objetos sin exponer la logica de creacion.
En Go, simplemente usamos funciones New*() que retornan interfaces.`)

	fmt.Println("\n  DEMO - Notification Factory:")
	for _, ntype := range []NotificationType{NotificationEmail, NotificationSMS, NotificationPush, NotificationSlack} {
		notif, err := NewNotification(ntype)
		if err != nil {
			fmt.Printf("    Error: %v\n", err)
			continue
		}
		_ = notif.Send("user@example.com", "Hello from "+notif.Type())
	}

	_, err := NewNotification("telegram")
	fmt.Printf("    Unknown type error: %v\n", err)

	fmt.Println(`
ABSTRACT FACTORY:

type StorageFactory interface {
    CreateBlobStorage() BlobStorage
    CreateQueueStorage() QueueStorage
    CreateTableStorage() TableStorage
}

type AWSFactory struct{}
func (f *AWSFactory) CreateBlobStorage() BlobStorage { return &S3Storage{} }
func (f *AWSFactory) CreateQueueStorage() QueueStorage { return &SQSStorage{} }

type GCPFactory struct{}
func (f *GCPFactory) CreateBlobStorage() BlobStorage { return &GCSStorage{} }
func (f *GCPFactory) CreateQueueStorage() QueueStorage { return &PubSubStorage{} }`)

	// ============================================
	// 4. SINGLETON PATTERN
	// ============================================
	fmt.Println("\n--- 4. Singleton Pattern (sync.Once) ---")
	fmt.Println(`
sync.Once garantiza que la inicializacion ocurre EXACTAMENTE una vez,
incluso con multiples goroutines accediendo concurrentemente.

var (
    instance *DBConnection
    once     sync.Once
)

func GetDB() *DBConnection {
    once.Do(func() {
        instance = &DBConnection{...}
    })
    return instance
}`)

	fmt.Println("\n  DEMO - Singleton:")
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			db := GetDB()
			fmt.Printf("    Goroutine %d got: %s\n", id, db)
		}(i)
	}
	wg.Wait()

	fmt.Println(`
ALTERNATIVA MEJOR - package-level init:

// db/db.go
package db

var conn *Connection

func init() {
    conn = connect(os.Getenv("DATABASE_URL"))
}

func Conn() *Connection { return conn }

CUANDO USAR sync.Once vs init():
- init(): Inicializacion al arrancar el programa
- sync.Once: Lazy initialization (solo cuando se necesita)
- sync.Once es mejor para testing (se puede controlar cuando inicia)`)

	// ============================================
	// 5. STRATEGY PATTERN
	// ============================================
	fmt.Println("\n--- 5. Strategy Pattern ---")
	fmt.Println(`
Define una familia de algoritmos intercambiables via interface.
El cliente elige la estrategia en runtime.

type Strategy interface {
    Execute(data []byte) ([]byte, error)
}`)

	fmt.Println("\n  DEMO - Compression Strategies:")
	data := []byte("Hello, this is some data to compress")

	strategies := []CompressionStrategy{
		&GzipCompression{},
		&ZstdCompression{},
		&SnappyCompression{},
	}

	processor := NewFileProcessor(strategies[0])
	for _, s := range strategies {
		processor.SetStrategy(s)
		result, _ := processor.Process(data)
		fmt.Printf("    %s -> %s\n", s.Name(), string(result))
	}

	fmt.Println(`
STRATEGY CON FUNCIONES (mas idomiatico en Go):

type CompressFunc func([]byte) ([]byte, error)

func ProcessFile(data []byte, compress CompressFunc) ([]byte, error) {
    return compress(data)
}

// Uso:
ProcessFile(data, gzipCompress)
ProcessFile(data, zstdCompress)`)

	// ============================================
	// 6. OBSERVER PATTERN
	// ============================================
	fmt.Println("\n--- 6. Observer Pattern (channels) ---")
	fmt.Println(`
En Go, el Observer pattern se implementa naturalmente con channels.
Los subscribers reciben events via channels, no callbacks.

type EventBus struct {
    subscribers map[string][]chan Event
}`)

	fmt.Println("\n  DEMO - Event Bus:")
	bus := NewEventBus()

	userCreated := bus.Subscribe("user.created")
	orderCreated := bus.Subscribe("user.created")

	var obWg sync.WaitGroup
	obWg.Add(2)
	go func() {
		defer obWg.Done()
		ev := <-userCreated
		fmt.Printf("    Subscriber-1 got: %s -> %v\n", ev.Type, ev.Payload)
	}()
	go func() {
		defer obWg.Done()
		ev := <-orderCreated
		fmt.Printf("    Subscriber-2 got: %s -> %v\n", ev.Type, ev.Payload)
	}()

	bus.Publish(Event{Type: "user.created", Payload: "user-123"})
	obWg.Wait()
	bus.Close()

	fmt.Println(`
OBSERVER CON CONTEXT (cancelable):

func (eb *EventBus) SubscribeWithContext(ctx context.Context, eventType string) <-chan Event {
    ch := make(chan Event, 16)
    eb.subscribe(eventType, ch)

    go func() {
        <-ctx.Done()
        eb.unsubscribe(eventType, ch)
        close(ch)
    }()

    return ch
}`)

	// ============================================
	// 7. DECORATOR PATTERN (middleware)
	// ============================================
	fmt.Println("\n--- 7. Decorator Pattern (Middleware) ---")
	os.Stdout.WriteString(`
En Go, el patron Decorator es el patron Middleware.
Funciones que wrappean otras funciones anadiendo comportamiento.

type Middleware func(HandlerFunc) HandlerFunc

func WithLogging(next HandlerFunc) HandlerFunc {
    return func(ctx context.Context, req string) (string, error) {
        log.Printf("request: %s", req)
        resp, err := next(ctx, req)
        log.Printf("response: %s", resp)
        return resp, err
    }
}
`)

	fmt.Println("\n  DEMO - Stacking decorators:")
	baseHandler := func(_ context.Context, req string) (string, error) {
		return "processed:" + req, nil
	}

	decorated := WithLogging(
		WithTimeoutDecorator(5*time.Second,
			WithRetry(2, baseHandler),
		),
	)

	resp, _ := decorated(context.Background(), "hello")
	fmt.Printf("    Final response: %s\n", resp)

	os.Stdout.WriteString(`
HTTP MIDDLEWARE (el ejemplo mas comun):

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "unauthorized", 401)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Composicion:
handler := LoggingMiddleware(AuthMiddleware(myHandler))

// Con chi/mux:
r := chi.NewRouter()
r.Use(LoggingMiddleware)
r.Use(AuthMiddleware)
r.Use(CORSMiddleware)
`)

	// ============================================
	// 8. ADAPTER PATTERN
	// ============================================
	fmt.Println("\n--- 8. Adapter Pattern ---")
	fmt.Println(`
Convierte la interface de una clase existente en otra que el cliente espera.
Util para integrar librerias legacy o third-party.`)

	fmt.Println("\n  DEMO - Payment Adapter:")
	legacyGateway := &OldPaymentGateway{}
	payProcessor := NewPaymentAdapter(legacyGateway)

	result, _ := payProcessor.ProcessPayment(PaymentRequest{
		Amount:      99.99,
		Currency:    "USD",
		CardToken:   "tok_visa_4242",
		Description: "Premium plan",
	})
	fmt.Printf("    Transaction: %s, Status: %s\n", result.TransactionID, result.Status)

	fmt.Println(`
ADAPTER CON io.Reader/Writer:

// Adaptar net.Conn a io.Reader
type ConnReader struct {
    conn net.Conn
}

func (cr *ConnReader) Read(p []byte) (int, error) {
    return cr.conn.Read(p)
}

// La stdlib usa adapters constantemente:
// strings.NewReader(s) -> io.Reader
// bytes.NewBuffer(b)   -> io.ReadWriter
// bufio.NewReader(r)   -> io.Reader con buffer`)

	// ============================================
	// 9. REPOSITORY PATTERN
	// ============================================
	fmt.Println("\n--- 9. Repository Pattern ---")
	fmt.Println(`
Abstrae el acceso a datos detras de una interface.
Permite cambiar la implementacion (SQL, NoSQL, in-memory) sin
afectar la logica de negocio.

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    FindAll(ctx context.Context) ([]*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}`)

	fmt.Println("\n  DEMO - In-Memory Repository:")
	repo := NewInMemoryUserRepository()
	ctx := context.Background()

	_ = repo.Create(ctx, &User{ID: "1", Name: "Alice", Email: "alice@example.com"})
	_ = repo.Create(ctx, &User{ID: "2", Name: "Bob", Email: "bob@example.com"})

	user, _ := repo.FindByID(ctx, "1")
	fmt.Printf("    FindByID(1): %+v\n", user)

	user, _ = repo.FindByEmail(ctx, "bob@example.com")
	fmt.Printf("    FindByEmail(bob@...): %+v\n", user)

	all, _ := repo.FindAll(ctx)
	fmt.Printf("    FindAll: %d users\n", len(all))

	fmt.Println(`
IMPLEMENTACIONES:

// PostgreSQL
type PostgresUserRepo struct {
    db *pgxpool.Pool
}

func (r *PostgresUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    var u User
    err := r.db.QueryRow(ctx,
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&u.ID, &u.Name, &u.Email)
    return &u, err
}

// MongoDB
type MongoUserRepo struct {
    collection *mongo.Collection
}

func (r *MongoUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    var u User
    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
    return &u, err
}

// Redis (cache layer)
type CachedUserRepo struct {
    primary UserRepository
    cache   *redis.Client
    ttl     time.Duration
}`)

	// ============================================
	// 10. UNIT OF WORK PATTERN
	// ============================================
	fmt.Println("\n--- 10. Unit of Work Pattern ---")
	fmt.Println(`
Agrupa multiples operaciones de base de datos en una transaccion.
Si una falla, todas se revierten.

type UnitOfWork struct {
    operations []Operation
}

func (uow *UnitOfWork) RegisterNew(entity string, data any)
func (uow *UnitOfWork) RegisterDirty(entity string, data any)
func (uow *UnitOfWork) RegisterDeleted(entity string, data any)
func (uow *UnitOfWork) Commit() error
func (uow *UnitOfWork) Rollback()`)

	fmt.Println("\n  DEMO - Unit of Work:")
	uow := NewUnitOfWork()
	uow.RegisterNew("users", User{ID: "3", Name: "Charlie", Email: "charlie@test.com"})
	uow.RegisterDirty("users", User{ID: "1", Name: "Alice Updated", Email: "alice@test.com"})
	uow.RegisterDeleted("sessions", "session-expired-123")
	_ = uow.Commit()

	fmt.Println(`
CON BASE DE DATOS REAL:

type UnitOfWork struct {
    tx *sql.Tx
}

func NewUnitOfWork(db *sql.DB) (*UnitOfWork, error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    return &UnitOfWork{tx: tx}, nil
}

func (uow *UnitOfWork) Users() UserRepository {
    return &PostgresUserRepo{db: uow.tx}  // Usa la transaccion
}

func (uow *UnitOfWork) Orders() OrderRepository {
    return &PostgresOrderRepo{db: uow.tx}
}

func (uow *UnitOfWork) Commit() error { return uow.tx.Commit() }
func (uow *UnitOfWork) Rollback() error { return uow.tx.Rollback() }

// Uso:
uow, _ := NewUnitOfWork(db)
defer uow.Rollback()

user := &User{Name: "Alice"}
uow.Users().Create(ctx, user)

order := &Order{UserID: user.ID, Total: 99.99}
uow.Orders().Create(ctx, order)

uow.Commit()  // Todo o nada`)

	// ============================================
	// 11. DEPENDENCY INJECTION
	// ============================================
	fmt.Println("\n--- 11. Dependency Injection (sin frameworks) ---")
	fmt.Println(`
En Go, DI se hace via constructor injection.
NO necesitas frameworks como Spring. Pasa dependencias como interfaces.

CONSTRUCTOR INJECTION (el patron principal):

type UserService struct {
    repo   UserRepository
    mailer Mailer
    logger Logger
}

func NewUserService(repo UserRepository, mailer Mailer, logger Logger) *UserService {
    return &UserService{repo: repo, mailer: mailer, logger: logger}
}`)

	fmt.Println("\n  DEMO - Constructor Injection:")
	diRepo := NewInMemoryUserRepository()
	mailer := &ConsoleMailer{}
	logger := &defaultLogger{}

	userSvc := NewUserService(diRepo, mailer, logger)
	newUser, _ := userSvc.Register(ctx, "Diana", "diana@example.com")
	fmt.Printf("    Registered user: %s (%s)\n", newUser.Name, newUser.Email)

	fmt.Println(`
INTERFACE-BASED DI:

// En tests, inyectas mocks:
type MockUserRepo struct {
    users map[string]*User
}
func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    return m.users[id], nil
}

// En test:
mockRepo := &MockUserRepo{users: map[string]*User{"1": {Name: "Test"}}}
svc := NewUserService(mockRepo, &MockMailer{}, &MockLogger{})

WIRE (Google) - Code generation DI:

// wire.go (build tag: wireinject)
//go:build wireinject

func InitializeApp() (*App, error) {
    wire.Build(
        NewDatabase,
        NewUserRepository,
        NewUserService,
        NewHTTPHandler,
        NewApp,
    )
    return nil, nil
}

// wire genera wire_gen.go con:
func InitializeApp() (*App, error) {
    db, err := NewDatabase()
    if err != nil { return nil, err }
    repo := NewUserRepository(db)
    svc := NewUserService(repo)
    handler := NewHTTPHandler(svc)
    app := NewApp(handler)
    return app, nil
}

// Instalar: go install github.com/google/wire/cmd/wire@latest
// Ejecutar: wire ./...

FX (Uber) - Runtime DI container:

func main() {
    fx.New(
        fx.Provide(
            NewDatabase,
            NewUserRepository,
            NewUserService,
            NewHTTPHandler,
        ),
        fx.Invoke(func(h *HTTPHandler) {
            // h tiene todas sus dependencias inyectadas
        }),
        fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
            return &fxevent.ZapLogger{Logger: log}
        }),
    ).Run()
}

// fx resuelve el grafo de dependencias automaticamente
// Maneja lifecycle (OnStart/OnStop)
// Ideal para aplicaciones grandes

CUANDO USAR CADA UNO:
- Manual (constructor): Proyectos pequenos/medianos (recomendado)
- Wire: Proyectos grandes, quieres compile-time safety
- Fx: Proyectos enterprise, necesitas lifecycle management`)

	// ============================================
	// 12. CONFIGURATION MANAGEMENT
	// ============================================
	fmt.Println("\n--- 12. Configuration Management ---")
	fmt.Println(`
PATRON 1 - Environment Variables (12-Factor App):

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
}

func LoadConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Host: getEnv("SERVER_HOST", "0.0.0.0"),
            Port: getEnvInt("SERVER_PORT", 8080),
        },
        Database: DatabaseConfig{
            Host: getEnv("DB_HOST", "localhost"),
            Port: getEnvInt("DB_PORT", 5432),
        },
    }
}`)

	fmt.Println("\n  DEMO - Config from Environment:")
	cfg := LoadConfigFromEnv()
	fmt.Printf("    Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("    DB: %s:%d/%s\n", cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("    Log: level=%s format=%s\n", cfg.Log.Level, cfg.Log.Format)

	os.Stdout.WriteString(`
PATRON 2 - Config Structs con tags:

// Con envconfig (kelseyhightower)
type Config struct {
    Port     int           'envconfig:"PORT" default:"8080"'
    Host     string        'envconfig:"HOST" default:"0.0.0.0"'
    DBUrl    string        'envconfig:"DATABASE_URL" required:"true"'
    Debug    bool          'envconfig:"DEBUG" default:"false"'
    Timeout  time.Duration 'envconfig:"TIMEOUT" default:"30s"'
}

func Load() (*Config, error) {
    var cfg Config
    err := envconfig.Process("APP", &cfg)
    return &cfg, err
}

// Con cleanenv:
type Config struct {
    Port    int    'yaml:"port" env:"PORT" env-default:"8080"'
    Host    string 'yaml:"host" env:"HOST" env-default:"0.0.0.0"'
    DBUrl   string 'yaml:"db_url" env:"DATABASE_URL" env-required:"true"'
}

func Load() (*Config, error) {
    var cfg Config
    err := cleanenv.ReadConfig("config.yml", &cfg)  // YAML + env overrides
    return &cfg, err
}

PATRON 3 - Viper (spf13):

func InitConfig() {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("/etc/myapp/")

    // Defaults
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.host", "0.0.0.0")

    // Environment variables override
    viper.AutomaticEnv()
    viper.SetEnvPrefix("APP")
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        log.Printf("no config file found: %v", err)
    }

    // Watch for changes (hot reload)
    viper.WatchConfig()
    viper.OnConfigChange(func(e fsnotify.Event) {
        log.Println("config changed:", e.Name)
    })
}

// Acceso:
port := viper.GetInt("server.port")
host := viper.GetString("server.host")

// Unmarshal a struct:
var cfg Config
viper.Unmarshal(&cfg)

PATRON 4 - Validacion de config:

type Config struct {
    Port     int    'validate:"required,min=1,max=65535"'
    Host     string 'validate:"required,hostname"'
    LogLevel string 'validate:"required,oneof=debug info warn error"'
}

func (c *Config) Validate() error {
    validate := validator.New()
    return validate.Struct(c)
}

RECOMENDACION:
- Pequeno: env vars + struct manual (lo que mostramos arriba)
- Mediano: envconfig o cleanenv
- Grande: Viper + YAML + env overrides
`)

	// ============================================
	// 13. GRACEFUL SHUTDOWN
	// ============================================
	fmt.Println("\n--- 13. Graceful Shutdown Pattern ---")
	os.Stdout.WriteString(`
IMPLEMENTACION COMPLETA:

func main() {
    // 1. Signal handling
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // 2. Inicializar dependencias
    db := connectDB()
    cache := connectRedis()
    queue := connectRabbitMQ()

    // 3. Crear server con handler
    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/api/", apiHandler)

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // 4. Start server en goroutine
    go func() {
        log.Printf("server listening on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    // 5. Esperar signal
    <-ctx.Done()
    log.Println("shutdown signal received")

    // 6. Graceful shutdown con timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 30*time.Second)
    defer shutdownCancel()

    // 7. Dejar de aceptar nuevas conexiones
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("server shutdown error: %v", err)
    }

    // 8. Drenar conexiones activas (Shutdown ya lo hace)
    // 9. Cerrar dependencias (en orden inverso)
    queue.Close()
    cache.Close()
    db.Close()

    log.Println("shutdown complete")
}

CON MULTIPLES SERVERS (gRPC + HTTP):

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    g, gCtx := errgroup.WithContext(ctx)

    // HTTP server
    httpSrv := &http.Server{Addr: ":8080"}
    g.Go(func() error {
        return httpSrv.ListenAndServe()
    })
    g.Go(func() error {
        <-gCtx.Done()
        return httpSrv.Shutdown(context.Background())
    })

    // gRPC server
    grpcSrv := grpc.NewServer()
    g.Go(func() error {
        lis, _ := net.Listen("tcp", ":50051")
        return grpcSrv.Serve(lis)
    })
    g.Go(func() error {
        <-gCtx.Done()
        grpcSrv.GracefulStop()
        return nil
    })

    if err := g.Wait(); err != nil {
        log.Printf("exit reason: %v", err)
    }
}

CON CLEANUP FUNCTIONS:
`)

	fmt.Println("\n  DEMO - Graceful Server (estructura):")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "OK")
	})

	gs := NewGracefulServer(":0", mux)
	gs.RegisterCleanup(func() {
		fmt.Println("    Cleanup: closing database connection")
	})
	gs.RegisterCleanup(func() {
		fmt.Println("    Cleanup: flushing metrics")
	})
	fmt.Printf("    GracefulServer created with %d cleanup functions\n", len(gs.cleanups))

	fmt.Println(`
CONTEXT PROPAGATION:

// El context del request se cancela cuando el server hace shutdown.
// Handlers deben respetar ctx.Done():

func processHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    result := make(chan string, 1)
    go func() {
        // Long operation
        time.Sleep(5 * time.Second)
        result <- "done"
    }()

    select {
    case <-ctx.Done():
        // Client disconnected o server shutting down
        http.Error(w, "cancelled", http.StatusServiceUnavailable)
    case res := <-result:
        fmt.Fprint(w, res)
    }
}`)

	// ============================================
	// 14. HEALTH CHECK PATTERN
	// ============================================
	fmt.Println("\n--- 14. Health Check Pattern ---")
	fmt.Println(`
Kubernetes y load balancers necesitan dos endpoints:

LIVENESS (/healthz):
- "Esta vivo el proceso?"
- Si falla -> Kubernetes reinicia el pod
- Debe ser RAPIDO y SIMPLE (no check deps)

READINESS (/readyz):
- "Puede recibir trafico?"
- Si falla -> se saca del load balancer (no recibe trafico)
- Verifica dependencias (DB, Redis, etc.)`)

	fmt.Println("\n  DEMO - Health Checks:")
	dbChecker := &DatabaseHealthChecker{name: "postgres", connected: true}
	redisChecker := &RedisHealthChecker{name: "redis", connected: true}

	health := NewHealthEndpoint(dbChecker, redisChecker)

	liveness := health.Liveness()
	fmt.Printf("    Liveness:  status=%s detail=%s\n", liveness.Status, liveness.Detail)

	status, checks := health.Readiness(ctx)
	fmt.Printf("    Readiness: status=%s\n", status)
	for _, c := range checks {
		fmt.Printf("      - %s: %s (%s)\n", c.Name, c.Status, c.Detail)
	}

	redisChecker.connected = false
	status, checks = health.Readiness(ctx)
	fmt.Printf("    Readiness (redis down): status=%s\n", status)
	for _, c := range checks {
		fmt.Printf("      - %s: %s (%s)\n", c.Name, c.Status, c.Detail)
	}

	fmt.Println(`
HTTP HANDLERS:

func (h *HealthEndpoint) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(h.Liveness())
    })

    mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
        status, checks := h.Readiness(r.Context())
        code := http.StatusOK
        if status == HealthDown {
            code = http.StatusServiceUnavailable
        }
        w.WriteHeader(code)
        json.NewEncoder(w).Encode(map[string]any{
            "status": status,
            "checks": checks,
        })
    })
}

KUBERNETES MANIFEST:

livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  failureThreshold: 3`)

	// ============================================
	// 15. TABLE-DRIVEN FACTORY
	// ============================================
	fmt.Println("\n--- 15. Table-Driven Factory Pattern ---")
	os.Stdout.WriteString(`
Combina Factory + table-driven design.
Un map de string -> constructor function.
Extensible en runtime via Register().

var registry = map[string]func() Serializer{
    "json": func() Serializer { return &JSONSerializer{} },
    "xml":  func() Serializer { return &XMLSerializer{} },
}

func New(format string) (Serializer, error) {
    factory, ok := registry[format]
    if !ok {
        return nil, fmt.Errorf("unsupported: %s", format)
    }
    return factory(), nil
}

func Register(format string, factory func() Serializer) {
    registry[format] = factory
}
`)

	fmt.Println("\n  DEMO - Table-Driven Factory:")
	for _, format := range []string{"json", "xml", "yaml"} {
		s, _ := NewSerializer(format)
		data, _ := s.Serialize("hello world")
		fmt.Printf("    %s (%s): %s\n", format, s.ContentType(), string(data))
	}

	RegisterSerializer("csv", func() Serializer {
		return &csvSerializer{}
	})
	csvSer, _ := NewSerializer("csv")
	csvData, _ := csvSer.Serialize("hello world")
	fmt.Printf("    csv (%s): %s\n", csvSer.ContentType(), string(csvData))

	_, serErr := NewSerializer("protobuf")
	fmt.Printf("    protobuf error: %v\n", serErr)

	os.Stdout.WriteString(`
PATRON APLICADO A HANDLERS:

var handlers = map[string]http.HandlerFunc{
    "GET /users":     listUsers,
    "POST /users":    createUser,
    "GET /users/{id}": getUser,
}

PATRON APLICADO A VALIDATORS:

var validators = map[string]func(string) error{
    "email": validateEmail,
    "phone": validatePhone,
    "url":   validateURL,
}

func Validate(fieldType, value string) error {
    v, ok := validators[fieldType]
    if !ok {
        return fmt.Errorf("no validator for: %s", fieldType)
    }
    return v(value)
}
`)

	// Suppress unused import warnings
	_ = strings.NewReader
	_ = io.Discard
	_ = http.StatusOK
	_ = log.Println
	_ = time.Now
	_ = os.Getenv
	_ = signal.Notify
	_ = syscall.SIGINT
}

type csvSerializer struct{}

func (c *csvSerializer) Serialize(data any) ([]byte, error) {
	return []byte(fmt.Sprintf("data\n%v\n", data)), nil
}
func (c *csvSerializer) ContentType() string { return "text/csv" }

/*
RESUMEN DE DESIGN PATTERNS EN GO:

FUNCTIONAL OPTIONS (Dave Cheney):
- Patron mas importante e idiomatico en Go
- Resuelve constructores con muchos parametros opcionales
- type Option func(*Config) + WithXxx() functions
- Usado en grpc.Dial, zap.New, http libs

BUILDER:
- Construccion paso a paso con method chaining
- Build() produce el resultado final
- Ideal para queries, requests, configs complejas

FACTORY:
- Funcion New*() que retorna interface
- switch/map para seleccionar implementacion
- Abstract Factory para familias de objetos

SINGLETON (sync.Once):
- Inicializacion thread-safe una sola vez
- sync.Once.Do() garantiza atomicidad
- Preferir sobre init() para lazy initialization

STRATEGY:
- Interface con multiples implementaciones
- Cliente elige algoritmo en runtime
- En Go, tambien se puede hacer con funciones

OBSERVER (channels):
- EventBus con subscribers via channels
- Mas idiomatico que callbacks en Go
- Publicacion no-bloqueante con select/default

DECORATOR (middleware):
- func(Handler) Handler
- Componer: WithLogging(WithAuth(WithRetry(handler)))
- Patron fundamental en HTTP servers Go

ADAPTER:
- Convierte interface legacy a interface moderna
- Wrapper struct que implementa la nueva interface
- La stdlib lo usa extensivamente (io.Reader adapters)

REPOSITORY:
- Interface para acceso a datos
- Permite swap de implementacion (SQL/NoSQL/Memory)
- Facilita testing con mocks

UNIT OF WORK:
- Agrupa operaciones en transaccion
- Commit/Rollback atomico
- RegisterNew/RegisterDirty/RegisterDeleted

DEPENDENCY INJECTION:
- Constructor injection (pasar interfaces)
- NO necesitas frameworks en Go
- Wire (Google): compile-time DI via codegen
- Fx (Uber): runtime DI container

CONFIGURATION:
- 12-Factor: environment variables
- Config structs con defaults
- envconfig/cleanenv para struct tags
- Viper para proyectos grandes (YAML + env + watch)

GRACEFUL SHUTDOWN:
- signal.NotifyContext para capturar SIGINT/SIGTERM
- srv.Shutdown(ctx) drena conexiones activas
- Cerrar dependencias en orden inverso
- errgroup para multiples servers

HEALTH CHECKS:
- Liveness (/healthz): proceso vivo, sin check deps
- Readiness (/readyz): puede recibir trafico, check deps
- HealthChecker interface para cada dependencia

TABLE-DRIVEN FACTORY:
- map[string]func() Interface
- Register() para extensibilidad
- Combina Factory + table-driven testing style
*/

/* SUMMARY - DESIGN PATTERNS EN GO:

TOPIC: Patrones de diseño idiomáticos para Go

FUNCTIONAL OPTIONS:
- Construcción flexible con defaults (type Option func(*Config))
- Usado por grpc.Dial, zap.New y bibliotecas estándar

BUILDER:
- Method chaining para construir objetos complejos paso a paso
- Ideal para query builders, request builders

FACTORY & SINGLETON:
- Factory: funciones New*() que retornan interfaces
- Singleton: sync.Once para inicialización thread-safe única

STRATEGY & OBSERVER:
- Strategy: interfaces para algoritmos intercambiables
- Observer: channels para pub/sub (más idiomático que callbacks)

DECORATOR (MIDDLEWARE):
- func(Handler) Handler para composición de comportamiento
- Stack de middlewares en HTTP servers

ADAPTER & REPOSITORY:
- Adapter: convertir interfaces legacy a nuevas
- Repository: abstraer acceso a datos (SQL/NoSQL/Memory)

DEPENDENCY INJECTION:
- Constructor injection sin frameworks
- Wire (codegen) o Fx (runtime) para proyectos grandes

CONFIGURATION & SHUTDOWN:
- 12-Factor: env vars, Viper para proyectos complejos
- Graceful shutdown con signal.NotifyContext y srv.Shutdown()

HEALTH CHECKS:
- /livez: proceso vivo (Kubernetes liveness)
- /readyz: puede recibir tráfico (Kubernetes readiness)
*/
