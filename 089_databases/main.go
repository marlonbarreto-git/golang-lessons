// Package main - Chapter 089: Databases
// Go tiene excelente soporte para bases de datos via database/sql
// y drivers específicos. Aprenderás SQL, ORMs y mejores prácticas.
package main

import (
	"os"
	"context"
	"fmt"
)

func main() {
	fmt.Println("=== BASES DE DATOS EN GO ===")

	// ============================================
	// DATABASE/SQL OVERVIEW
	// ============================================
	fmt.Println("\n--- database/sql Overview ---")
	fmt.Println(`
El paquete database/sql provee interfaz genérica para SQL DBs.

DRIVERS POPULARES:
- PostgreSQL: github.com/lib/pq, github.com/jackc/pgx
- MySQL: github.com/go-sql-driver/mysql
- SQLite: github.com/mattn/go-sqlite3 (CGO)
         modernc.org/sqlite (pure Go)
- SQL Server: github.com/microsoft/go-mssqldb

IMPORTAR DRIVER:
import (
    "database/sql"
    _ "github.com/lib/pq"  // blank import para registrar
)`)
	// ============================================
	// CONEXIÓN
	// ============================================
	fmt.Println("\n--- Conexión ---")
	fmt.Println(`
// Abrir conexión
db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Verificar conexión
if err := db.Ping(); err != nil {
    log.Fatal(err)
}

// Configurar pool de conexiones
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(5 * time.Minute)

CONNECTION STRINGS:

PostgreSQL:
postgres://user:pass@host:5432/db?sslmode=disable

MySQL:
user:pass@tcp(host:3306)/db?parseTime=true

SQLite:
file:test.db?cache=shared&mode=rwc`)
	// ============================================
	// QUERIES
	// ============================================
	fmt.Println("\n--- Queries ---")
	os.Stdout.WriteString(`
// Query que retorna filas
rows, err := db.Query("SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return err
    }
    fmt.Printf("User: %d - %s\n", id, name)
}

// IMPORTANTE: Verificar error después del loop
if err := rows.Err(); err != nil {
    return err
}

// Query para una sola fila
var name string
err := db.QueryRow("SELECT name FROM users WHERE id = $1", 1).Scan(&name)
if err == sql.ErrNoRows {
    fmt.Println("No user found")
} else if err != nil {
    return err
}
`)

	// ============================================
	// EXEC (INSERT, UPDATE, DELETE)
	// ============================================
	fmt.Println("\n--- Exec ---")
	os.Stdout.WriteString(`
// INSERT
result, err := db.Exec(
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "Alice", "alice@example.com",
)
if err != nil {
    return err
}

// Obtener ID insertado (PostgreSQL)
// Con db.QueryRow y RETURNING
var id int
err = db.QueryRow(
    "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
    "Bob", "bob@example.com",
).Scan(&id)

// Filas afectadas
rowsAffected, _ := result.RowsAffected()
fmt.Printf("Rows affected: %d\n", rowsAffected)

// UPDATE
result, err = db.Exec(
    "UPDATE users SET name = $1 WHERE id = $2",
    "New Name", 1,
)

// DELETE
result, err = db.Exec("DELETE FROM users WHERE id = $1", 1)
`)

	// ============================================
	// PREPARED STATEMENTS
	// ============================================
	fmt.Println("\n--- Prepared Statements ---")
	fmt.Println(`
// Preparar statement (reutilizable)
stmt, err := db.Prepare("SELECT name FROM users WHERE id = $1")
if err != nil {
    return err
}
defer stmt.Close()

// Usar múltiples veces
for _, id := range []int{1, 2, 3} {
    var name string
    err := stmt.QueryRow(id).Scan(&name)
    if err != nil {
        continue
    }
    fmt.Println(name)
}

BENEFICIOS:
- Mejor rendimiento para queries repetidas
- Previene SQL injection
- El servidor parsea el query una sola vez`)
	// ============================================
	// TRANSACCIONES
	// ============================================
	fmt.Println("\n--- Transacciones ---")
	fmt.Println(`
// Iniciar transacción
tx, err := db.Begin()
if err != nil {
    return err
}
// Rollback automático si no se hace commit
defer tx.Rollback()

// Operaciones dentro de la transacción
_, err = tx.Exec("INSERT INTO orders (user_id, total) VALUES ($1, $2)", 1, 100)
if err != nil {
    return err  // Rollback ocurre en defer
}

_, err = tx.Exec("UPDATE inventory SET stock = stock - 1 WHERE id = $1", 1)
if err != nil {
    return err
}

// Commit
if err := tx.Commit(); err != nil {
    return err
}

// Con context
tx, err = db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
})`)
	// ============================================
	// CONTEXT
	// ============================================
	fmt.Println("\n--- Context y Timeouts ---")
	fmt.Println(`
// Siempre usar versiones con context en producción
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// QueryContext
rows, err := db.QueryContext(ctx, "SELECT * FROM users")

// QueryRowContext
var name string
err = db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = $1", 1).Scan(&name)

// ExecContext
result, err := db.ExecContext(ctx, "DELETE FROM old_logs WHERE created_at < $1", cutoff)

// PrepareContext
stmt, err := db.PrepareContext(ctx, "SELECT * FROM users WHERE id = $1")

// BeginTx
tx, err := db.BeginTx(ctx, nil)`)
	// ============================================
	// NULL VALUES
	// ============================================
	fmt.Println("\n--- NULL Values ---")
	fmt.Println(`
// Tipos nullable de database/sql
var name sql.NullString
var age sql.NullInt64
var salary sql.NullFloat64
var active sql.NullBool
var birthday sql.NullTime

err := db.QueryRow("SELECT name, age FROM users WHERE id = $1", 1).Scan(&name, &age)

if name.Valid {
    fmt.Println(name.String)
} else {
    fmt.Println("Name is NULL")
}

// Insertar NULL
_, err = db.Exec("INSERT INTO users (name, age) VALUES ($1, $2)",
    sql.NullString{String: "Alice", Valid: true},
    sql.NullInt64{Valid: false},  // NULL
)

// Alternativa: usar punteros
var name *string
err = row.Scan(&name)
if name != nil {
    fmt.Println(*name)
}`)
	// ============================================
	// PGX (POSTGRES NATIVO)
	// ============================================
	fmt.Println("\n--- pgx (PostgreSQL Nativo) ---")
	fmt.Println(`
import "github.com/jackc/pgx/v5"
import "github.com/jackc/pgx/v5/pgxpool"

// Pool de conexiones (recomendado para producción)
pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:5432/db")
if err != nil {
    return err
}
defer pool.Close()

// Queries
rows, err := pool.Query(ctx, "SELECT id, name FROM users")
for rows.Next() {
    var id int
    var name string
    rows.Scan(&id, &name)
}

// Batch queries
batch := &pgx.Batch{}
batch.Queue("INSERT INTO users (name) VALUES ($1)", "Alice")
batch.Queue("INSERT INTO users (name) VALUES ($1)", "Bob")
results := pool.SendBatch(ctx, batch)
defer results.Close()

// COPY para bulk insert
copyCount, err := pool.CopyFrom(
    ctx,
    pgx.Identifier{"users"},
    []string{"name", "email"},
    pgx.CopyFromRows(rows),
)

VENTAJAS DE PGX:
- Más rápido que lib/pq
- Soporte nativo para tipos PostgreSQL
- LISTEN/NOTIFY
- COPY protocol
- Mejor manejo de errores`)
	// ============================================
	// GORM (ORM)
	// ============================================
	fmt.Println("\n--- GORM (ORM) ---")
	fmt.Println(`
import "gorm.io/gorm"
import "gorm.io/driver/postgres"

// Modelo
type User struct {
    gorm.Model
    Name  string
    Email string ` + "`" + `gorm:"uniqueIndex"` + "`" + `
    Age   int
}

// Conectar
dsn := "host=localhost user=user password=pass dbname=db"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// Auto-migración
db.AutoMigrate(&User{})

// CRUD
// Create
user := User{Name: "Alice", Email: "alice@example.com"}
db.Create(&user)

// Read
var user User
db.First(&user, 1)  // Por ID
db.First(&user, "email = ?", "alice@example.com")

// Update
db.Model(&user).Update("Name", "New Name")
db.Model(&user).Updates(User{Name: "New", Age: 25})

// Delete
db.Delete(&user, 1)

// Queries complejas
var users []User
db.Where("age > ?", 18).Order("name").Limit(10).Find(&users)

// Preload (eager loading)
db.Preload("Orders").Find(&users)`)
	// ============================================
	// SQLX
	// ============================================
	fmt.Println("\n--- sqlx ---")
	fmt.Println(`
import "github.com/jmoiron/sqlx"

type User struct {
    ID    int    ` + "`" + `db:"id"` + "`" + `
    Name  string ` + "`" + `db:"name"` + "`" + `
    Email string ` + "`" + `db:"email"` + "`" + `
}

db, err := sqlx.Connect("postgres", "postgres://user:pass@localhost/db")

// Get (una fila)
var user User
err = db.Get(&user, "SELECT * FROM users WHERE id = $1", 1)

// Select (múltiples filas)
var users []User
err = db.Select(&users, "SELECT * FROM users WHERE age > $1", 18)

// Named queries
_, err = db.NamedExec(
    "INSERT INTO users (name, email) VALUES (:name, :email)",
    map[string]interface{}{"name": "Alice", "email": "alice@example.com"},
)

// StructScan en loop
rows, _ := db.Queryx("SELECT * FROM users")
for rows.Next() {
    var user User
    rows.StructScan(&user)
}

VENTAJAS DE SQLX:
- Extensión de database/sql
- Scan a structs automático
- Named queries
- In-clause expansion`)
	// ============================================
	// MIGRACIONES
	// ============================================
	fmt.Println("\n--- Migraciones ---")
	fmt.Println(`
HERRAMIENTAS POPULARES:

1. golang-migrate
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

   migrate create -ext sql -dir migrations -seq create_users
   migrate -path migrations -database "postgres://..." up
   migrate -path migrations -database "postgres://..." down 1

2. goose
   go install github.com/pressly/goose/v3/cmd/goose@latest

   goose -dir migrations postgres "postgres://..." up
   goose create add_users sql

3. atlas (moderno)
   atlas schema apply -u "postgres://..." --dev-url "docker://postgres"

ESTRUCTURA DE ARCHIVOS:
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_add_email.up.sql
└── 000002_add_email.down.sql`)
	// ============================================
	// BUENAS PRÁCTICAS
	// ============================================
	fmt.Println("\n--- Buenas Prácticas ---")
	fmt.Println(`
1. SIEMPRE usar context con timeout
2. SIEMPRE cerrar rows y statements
3. Configurar pool de conexiones apropiadamente
4. Usar prepared statements para queries repetidas
5. Usar transacciones para operaciones atómicas
6. Verificar rows.Err() después del loop
7. Usar índices apropiados en la DB
8. Manejar sql.ErrNoRows explícitamente
9. No exponer detalles de DB en errores de API
10. Usar migraciones versionadas`)
	// ============================================
	// EJEMPLO COMPLETO
	// ============================================
	fmt.Println("\n--- Ejemplo: Repository Pattern ---")
	fmt.Println(`
type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id int) (*User, error) {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    user := &User{}
    err := r.db.QueryRowContext(ctx,
        "SELECT id, name, email FROM users WHERE id = $1", id,
    ).Scan(&user.ID, &user.Name, &user.Email)

    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("querying user: %w", err)
    }
    return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    err := r.db.QueryRowContext(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        user.Name, user.Email,
    ).Scan(&user.ID)

    if err != nil {
        return fmt.Errorf("inserting user: %w", err)
    }
    return nil
}`)}

// Ejemplo de estructura User
type User struct {
	ID    int
	Name  string
	Email string
}

// Simular errores comunes
var (
	ErrUserNotFound = fmt.Errorf("user not found")
)

// Ejemplo de Repository interface
type UserRepository interface {
	FindByID(ctx context.Context, id int) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
}

// Este archivo es educativo - el código real de DB
// requiere una conexión activa a base de datos

/*
RESUMEN DE BASES DE DATOS:

DRIVERS:
- PostgreSQL: pgx (recomendado), lib/pq
- MySQL: go-sql-driver/mysql
- SQLite: modernc.org/sqlite (pure Go)

OPERACIONES BÁSICAS:
db.Query()     - múltiples filas
db.QueryRow()  - una fila
db.Exec()      - INSERT/UPDATE/DELETE
db.Prepare()   - prepared statements
db.Begin()     - transacciones

CONTEXT:
Usar *Context() versiones siempre:
db.QueryContext(ctx, ...)
db.ExecContext(ctx, ...)

POOL:
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5*time.Minute)

NULL:
sql.NullString, sql.NullInt64, etc.
O usar punteros: *string, *int

ORMs:
- GORM: completo, popular
- sqlx: extensión ligera de database/sql
- ent: Facebook, type-safe

MIGRACIONES:
golang-migrate, goose, atlas

BUENAS PRÁCTICAS:
1. Siempre context con timeout
2. Siempre cerrar recursos
3. Verificar rows.Err()
4. Repository pattern
5. Prepared statements
*/
