// Package main - Chapter 043: Time
// Manejo avanzado de tiempo en Go: parsing, formatting, timezones,
// monotonic clock, timers, tickers, unix timestamps y patrones.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== TIME PACKAGE DEEP DIVE EN GO ===")

	// ============================================
	// 1. TIME.TIME FUNDAMENTALS
	// ============================================
	fmt.Println("\n--- 1. time.Time Fundamentals ---")
	fmt.Println(`
time.Time representa un instante en el tiempo con precision de nanosegundos.
Internamente contiene:
  - wall clock (calendario): para mostrar/comparar fechas
  - monotonic clock: para medir duraciones (no afectado por NTP)
  - Location: timezone asociado

CREAR TIME VALUES:

  time.Now()                           -> Hora actual (wall + monotonic)
  time.Date(2024,1,15,10,30,0,0,loc)  -> Fecha especifica
  time.Unix(secs, nsecs)               -> Desde Unix timestamp
  time.UnixMilli(ms)                   -> Desde Unix milisegundos
  time.UnixMicro(us)                   -> Desde Unix microsegundos

ZERO VALUE:

  var t time.Time
  t.IsZero() == true
  // January 1, year 1, 00:00:00 UTC`)

	now := time.Now()
	fmt.Printf("\n  time.Now(): %v\n", now)
	fmt.Printf("  Year: %d, Month: %s, Day: %d\n", now.Year(), now.Month(), now.Day())
	fmt.Printf("  Hour: %d, Minute: %d, Second: %d\n", now.Hour(), now.Minute(), now.Second())
	fmt.Printf("  Weekday: %s\n", now.Weekday())
	fmt.Printf("  YearDay: %d\n", now.YearDay())

	specific := time.Date(2024, time.March, 15, 14, 30, 0, 0, time.UTC)
	fmt.Printf("  time.Date: %v\n", specific)

	var zero time.Time
	fmt.Printf("  Zero value: %v, IsZero: %v\n", zero, zero.IsZero())

	// ============================================
	// 2. TIME.DURATION
	// ============================================
	fmt.Println("\n--- 2. time.Duration ---")
	os.Stdout.WriteString(`
time.Duration es un int64 que representa nanosegundos.

CONSTANTES:
  time.Nanosecond   = 1
  time.Microsecond  = 1000 * Nanosecond
  time.Millisecond  = 1000 * Microsecond
  time.Second       = 1000 * Millisecond
  time.Minute       = 60 * Second
  time.Hour         = 60 * Minute

CREAR DURACIONES:
  5 * time.Second              -> 5s
  time.Duration(500) * time.Millisecond -> 500ms
  time.ParseDuration("1h30m")  -> 1h30m0s
  time.Since(t)                -> Duracion desde t hasta ahora
  time.Until(t)                -> Duracion desde ahora hasta t

METODOS:
  d.Hours()        -> float64 (horas)
  d.Minutes()      -> float64 (minutos)
  d.Seconds()      -> float64 (segundos)
  d.Milliseconds() -> int64
  d.Nanoseconds()  -> int64
  d.String()       -> "1h30m0s"
  d.Truncate(m)    -> Truncar a multiplo de m
  d.Round(m)       -> Redondear a multiplo de m

PELIGRO - No hay time.Day:
  // NO existe time.Day porque no todos los dias tienen 24h (DST)
  // Usar 24 * time.Hour solo si no te importa DST
`)

	d1 := 2*time.Hour + 30*time.Minute + 15*time.Second
	fmt.Printf("  Duration: %v\n", d1)
	fmt.Printf("  Hours: %.2f, Minutes: %.2f, Seconds: %.2f\n",
		d1.Hours(), d1.Minutes(), d1.Seconds())
	fmt.Printf("  Milliseconds: %d\n", d1.Milliseconds())

	d2, _ := time.ParseDuration("1h30m45s500ms")
	fmt.Printf("  ParseDuration: %v\n", d2)

	fmt.Printf("  Truncate to minute: %v\n", d1.Truncate(time.Minute))
	fmt.Printf("  Round to minute: %v\n", d1.Round(time.Minute))

	// ============================================
	// 3. REFERENCE TIME AND FORMATTING
	// ============================================
	fmt.Println("\n--- 3. Reference Time and Formatting ---")
	fmt.Println(`
Go usa un REFERENCE TIME para formatos (no strftime como otros lenguajes):

  Mon Jan 2 15:04:05 MST 2006
  01/02 03:04:05PM '06 -0700

Cada componente tiene un valor especifico:
  Month:    January / Jan / 01 / 1
  Day:      02 / 2 / _2
  Hour:     15 (24h) / 03 o 3 (12h)
  Minute:   04 / 4
  Second:   05 / 5
  Year:     2006 / 06
  Timezone: MST / -0700 / -07:00 / Z0700 / Z07:00
  AM/PM:    PM
  Weekday:  Monday / Mon

TRUCO PARA RECORDAR:
  1  2  3  4  5  6  7
  Mon Jan 2 15:04:05 2006 -0700
  (numeros del 1 al 7 en orden)

CONSTANTES PREDEFINIDAS:
  time.RFC3339      = "2006-01-02T15:04:05Z07:00"
  time.RFC3339Nano  = "2006-01-02T15:04:05.999999999Z07:00"
  time.RFC1123      = "Mon, 02 Jan 2006 15:04:05 MST"
  time.RFC822       = "02 Jan 06 15:04 MST"
  time.Kitchen      = "3:04PM"
  time.DateTime     = "2006-01-02 15:04:05"   (Go 1.20+)
  time.DateOnly     = "2006-01-02"             (Go 1.20+)
  time.TimeOnly     = "15:04:05"               (Go 1.20+)`)

	t := time.Date(2024, time.December, 25, 14, 30, 45, 0, time.UTC)
	fmt.Printf("\n  RFC3339:      %s\n", t.Format(time.RFC3339))
	fmt.Printf("  RFC1123:      %s\n", t.Format(time.RFC1123))
	fmt.Printf("  Kitchen:      %s\n", t.Format(time.Kitchen))
	fmt.Printf("  DateTime:     %s\n", t.Format(time.DateTime))
	fmt.Printf("  DateOnly:     %s\n", t.Format(time.DateOnly))
	fmt.Printf("  TimeOnly:     %s\n", t.Format(time.TimeOnly))
	fmt.Printf("  Custom:       %s\n", t.Format("02/01/2006 3:04 PM"))
	fmt.Printf("  ISO week:     %s\n", t.Format("2006-W01"))

	// ============================================
	// 4. PARSING TIME
	// ============================================
	fmt.Println("\n--- 4. Parsing Time ---")
	os.Stdout.WriteString(`
PARSING - Convertir string a time.Time:

  time.Parse(layout, value)             -> time.Time (asume UTC si no hay TZ)
  time.ParseInLocation(layout, val, loc) -> time.Time (asume loc si no hay TZ)

ERRORES COMUNES:
  - Usar "01-02-2006" cuando el input es "15-03-2024" (mes fuera de rango)
  - Olvidar que Parse asume UTC cuando no hay timezone en el string
  - Confundir formato americano (MM/DD) con europeo (DD/MM)
`)

	formats := []struct {
		layout, value string
	}{
		{time.RFC3339, "2024-03-15T14:30:00Z"},
		{time.RFC3339, "2024-03-15T14:30:00-05:00"},
		{time.DateTime, "2024-03-15 14:30:00"},
		{time.DateOnly, "2024-03-15"},
		{"02/01/2006", "15/03/2024"},
		{"Jan 2, 2006", "Mar 15, 2024"},
		{"2006-01-02 3:04 PM", "2024-03-15 2:30 PM"},
	}

	for _, f := range formats {
		parsed, err := time.Parse(f.layout, f.value)
		if err != nil {
			fmt.Printf("  PARSE ERROR: layout=%q value=%q -> %v\n", f.layout, f.value, err)
		} else {
			fmt.Printf("  layout=%-25q value=%-30q -> %v\n", f.layout, f.value, parsed)
		}
	}

	bogota, _ := time.LoadLocation("America/Bogota")
	parsedLocal, _ := time.ParseInLocation(time.DateTime, "2024-03-15 14:30:00", bogota)
	fmt.Printf("\n  ParseInLocation (Bogota): %v\n", parsedLocal)

	// ============================================
	// 5. TIMEZONE HANDLING
	// ============================================
	fmt.Println("\n--- 5. Timezone Handling ---")
	os.Stdout.WriteString(`
LOCATION - Representa una timezone:

  time.UTC                                 -> UTC
  time.Local                               -> Timezone del sistema
  time.LoadLocation("America/New_York")    -> IANA timezone
  time.FixedZone("COT", -5*60*60)         -> Offset fijo

CONVERTIR ENTRE TIMEZONES:
  t.In(loc)           -> Misma instancia, diferente representacion
  t.UTC()             -> Convertir a UTC
  t.Local()           -> Convertir a timezone local

IMPORTANTE:
  t.In(loc) NO cambia el instante en el tiempo, solo la representacion.
  t1.Equal(t1.UTC()) es SIEMPRE true.

IANA TIMEZONE DATABASE:
  Go incluye su propia copia del tzdata.
  Import "time/tzdata" para embeber en el binario
  (util cuando el sistema no tiene tzdata instalado).
`)

	utcTime := time.Date(2024, time.July, 4, 18, 0, 0, 0, time.UTC)
	fmt.Printf("  UTC:        %v\n", utcTime)

	nyLoc, _ := time.LoadLocation("America/New_York")
	tokyoLoc, _ := time.LoadLocation("Asia/Tokyo")
	londonLoc, _ := time.LoadLocation("Europe/London")
	bogotaLoc, _ := time.LoadLocation("America/Bogota")

	fmt.Printf("  New York:   %v\n", utcTime.In(nyLoc))
	fmt.Printf("  Tokyo:      %v\n", utcTime.In(tokyoLoc))
	fmt.Printf("  London:     %v\n", utcTime.In(londonLoc))
	fmt.Printf("  Bogota:     %v\n", utcTime.In(bogotaLoc))

	customZone := time.FixedZone("MY-TZ", 5*60*60+30*60)
	fmt.Printf("  Custom +5:30: %v\n", utcTime.In(customZone))

	fmt.Printf("\n  Equal check: UTC == Bogota? %v (same instant)\n",
		utcTime.Equal(utcTime.In(bogotaLoc)))

	_, offset := utcTime.In(bogotaLoc).Zone()
	fmt.Printf("  Bogota offset: %d seconds (%d hours)\n", offset, offset/3600)

	// ============================================
	// 6. MONOTONIC CLOCK
	// ============================================
	fmt.Println("\n--- 6. Monotonic Clock (Go 1.9+) ---")
	fmt.Println(`
Go's time.Now() captura DOS relojes:

  1. WALL CLOCK: Fecha/hora del calendario
     - Puede retroceder (ajustes NTP, DST)
     - Se usa para: Display, comparacion de fechas

  2. MONOTONIC CLOCK: Contador que solo avanza
     - NUNCA retrocede
     - Se usa para: Medir duraciones, timeouts

REGLAS:
  - time.Now() tiene ambos (wall + monotonic)
  - time.Date() solo tiene wall clock
  - time.Parse() solo tiene wall clock
  - t.Sub(u) usa monotonic si ambos lo tienen
  - t.Add(d) preserva monotonic
  - t.Round(), t.Truncate() eliminan monotonic
  - t.In(), t.Local(), t.UTC() eliminan monotonic

IMPLICACION:
  start := time.Now()          // wall + monotonic
  // ... operacion ...
  elapsed := time.Since(start) // usa monotonic (preciso)

  // vs
  start := time.Date(...)      // solo wall
  elapsed := time.Since(start) // usa wall (puede ser negativo si NTP ajusta)`)

	start := time.Now()
	fmt.Printf("\n  time.Now() string: %v\n", start)
	fmt.Printf("  (Nota: el 'm=+...' al final es el monotonic reading)\n")

	time.Sleep(10 * time.Millisecond)
	elapsed := time.Since(start)
	fmt.Printf("  time.Since(start): %v (uses monotonic)\n", elapsed)

	stripped := start.Round(0)
	fmt.Printf("  After Round(0):    %v (monotonic stripped)\n", stripped)

	// ============================================
	// 7. UNIX TIMESTAMPS
	// ============================================
	fmt.Println("\n--- 7. Unix Timestamps ---")
	os.Stdout.WriteString(`
UNIX TIMESTAMP: Segundos desde January 1, 1970 00:00:00 UTC

  OBTENER:
    t.Unix()      -> int64 (seconds)
    t.UnixMilli() -> int64 (milliseconds)
    t.UnixMicro() -> int64 (microseconds)
    t.UnixNano()  -> int64 (nanoseconds)

  CREAR DESDE TIMESTAMP:
    time.Unix(sec, nsec)  -> time.Time
    time.UnixMilli(ms)    -> time.Time
    time.UnixMicro(us)    -> time.Time

CUIDADO CON UnixNano():
  - int64 overflow ocurre en el ano 2262
  - Para timestamps lejanos, usar Unix() + nsec
  - Zero time (year 1) retorna undefined para UnixNano()
`)

	now2 := time.Now()
	fmt.Printf("  Unix (seconds):      %d\n", now2.Unix())
	fmt.Printf("  UnixMilli:           %d\n", now2.UnixMilli())
	fmt.Printf("  UnixMicro:           %d\n", now2.UnixMicro())
	fmt.Printf("  UnixNano:            %d\n", now2.UnixNano())

	fromUnix := time.Unix(1700000000, 0)
	fmt.Printf("\n  time.Unix(1700000000, 0): %v\n", fromUnix.UTC())

	fromMilli := time.UnixMilli(1700000000000)
	fmt.Printf("  time.UnixMilli(1700000000000): %v\n", fromMilli.UTC())

	fmt.Printf("  Both equal? %v\n", fromUnix.Equal(fromMilli))

	// ============================================
	// 8. TIME ARITHMETIC
	// ============================================
	fmt.Println("\n--- 8. Time Arithmetic ---")
	os.Stdout.WriteString(`
OPERACIONES CON TIEMPO:

  t.Add(d)             -> time.Time (sumar duracion)
  t.Sub(u)             -> time.Duration (diferencia entre dos tiempos)
  t.AddDate(y, m, d)   -> time.Time (sumar anos, meses, dias)
  time.Since(t)        -> time.Duration (equivale a time.Now().Sub(t))
  time.Until(t)        -> time.Duration (equivale a t.Sub(time.Now()))

COMPARACIONES:
  t.Before(u)          -> bool
  t.After(u)           -> bool
  t.Equal(u)           -> bool (normaliza timezone)
  t.Compare(u)         -> int (-1, 0, +1) (Go 1.20+)

IMPORTANTE:
  - Usar t.Equal(u) en vez de t == u
  - t == u falla si tienen diferente Location o monotonic reading
  - t.Equal(u) compara el instante real
`)

	base := time.Date(2024, time.January, 15, 10, 0, 0, 0, time.UTC)
	fmt.Printf("  Base:              %v\n", base)
	fmt.Printf("  +2h30m:            %v\n", base.Add(2*time.Hour+30*time.Minute))
	fmt.Printf("  +1 year, 2 months: %v\n", base.AddDate(1, 2, 0))
	fmt.Printf("  +30 days:          %v\n", base.AddDate(0, 0, 30))

	t1 := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	diff := t1.Sub(t2)
	fmt.Printf("\n  Mar 1 - Jan 1 = %v (%d days)\n", diff, int(diff.Hours()/24))

	fmt.Printf("  Jan 15 Before Mar 1? %v\n", base.Before(t1))
	fmt.Printf("  Jan 15 After Jan 1?  %v\n", base.After(t2))

	// Equal vs ==
	utc := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	bogota2, _ := time.LoadLocation("America/Bogota")
	col := utc.In(bogota2)
	fmt.Printf("\n  UTC:    %v\n", utc)
	fmt.Printf("  Bogota: %v\n", col)
	fmt.Printf("  t == u:      %v (WRONG way to compare)\n", utc == col)
	fmt.Printf("  t.Equal(u):  %v (CORRECT way to compare)\n", utc.Equal(col))

	// ============================================
	// 9. DAYLIGHT SAVING TIME GOTCHAS
	// ============================================
	fmt.Println("\n--- 9. Daylight Saving Time Gotchas ---")
	fmt.Println(`
DST puede causar bugs sutiles:

1. DIAS QUE NO TIENEN 24 HORAS:
   - En spring forward: un dia tiene 23h
   - En fall back: un dia tiene 25h
   - 24 * time.Hour != 1 dia calendario

2. HORAS QUE NO EXISTEN:
   - Spring forward: 2:00 AM salta a 3:00 AM
   - Las 2:30 AM no existe ese dia

3. HORAS AMBIGUAS:
   - Fall back: 1:30 AM ocurre DOS veces
   - Go elige la primera ocurrencia

SOLUCION: Usar AddDate para aritmetica de calendario:
   t.AddDate(0, 0, 1) -> Exactamente manana a la misma hora
   t.Add(24*time.Hour) -> 24h despues (puede ser diferente hora en DST)`)

	nyLoc2, _ := time.LoadLocation("America/New_York")

	springForward := time.Date(2024, time.March, 10, 1, 0, 0, 0, nyLoc2)
	fmt.Printf("\n  Before spring forward: %v\n", springForward)
	fmt.Printf("  +24h:                  %v\n", springForward.Add(24*time.Hour))
	fmt.Printf("  +1 day (AddDate):      %v\n", springForward.AddDate(0, 0, 1))

	plus24h := springForward.Add(24 * time.Hour)
	plus1day := springForward.AddDate(0, 0, 1)
	fmt.Printf("  Same result? %v (24h and 1 day differ during DST)\n",
		plus24h.Equal(plus1day))

	// ============================================
	// 10. TIMERS AND TICKERS
	// ============================================
	fmt.Println("\n--- 10. Timers and Tickers ---")
	os.Stdout.WriteString(`
TIMER - Se dispara UNA sola vez despues de la duracion:

  timer := time.NewTimer(5 * time.Second)
  <-timer.C    // Bloquea hasta que se dispara

  timer.Stop()  // Cancelar (retorna true si estaba activo)
  timer.Reset(d) // Reiniciar con nueva duracion

  // Shortcut (no cancelable):
  <-time.After(5 * time.Second)

TICKER - Se dispara REPETIDAMENTE a intervalos regulares:

  ticker := time.NewTicker(1 * time.Second)
  defer ticker.Stop()   // SIEMPRE hacer Stop()

  for t := range ticker.C {
      fmt.Println("Tick at", t)
  }

TIMER vs TICKER:
  Timer  -> Una accion despues de un delay (timeout, debounce)
  Ticker -> Accion periodica (polling, heartbeat, metrics)

GOTCHAS:
  - Ticker.Stop() no cierra el canal. Usar select con done channel.
  - time.After() crea un timer que no se puede cancelar (memory leak en loops)
  - En Go 1.23+: Stop() no necesita drenar el canal
`)

	fmt.Println("  [Timer demo]")
	timer := time.NewTimer(50 * time.Millisecond)
	timerStart := time.Now()
	<-timer.C
	fmt.Printf("  Timer fired after: %v\n", time.Since(timerStart).Round(time.Millisecond))

	fmt.Println("\n  [Ticker demo - 3 ticks]")
	ticker := time.NewTicker(30 * time.Millisecond)
	tickCount := 0
	tickStart := time.Now()
	for range ticker.C {
		tickCount++
		fmt.Printf("  Tick %d at %v\n", tickCount, time.Since(tickStart).Round(time.Millisecond))
		if tickCount >= 3 {
			ticker.Stop()
			break
		}
	}

	fmt.Println("\n  [Timer.Stop demo]")
	timer2 := time.NewTimer(1 * time.Second)
	stopped := timer2.Stop()
	fmt.Printf("  Timer stopped before firing: %v\n", stopped)

	// ============================================
	// 11. TIME.AFTERFUNC
	// ============================================
	fmt.Println("\n--- 11. time.AfterFunc ---")
	os.Stdout.WriteString(`
time.AfterFunc ejecuta una funcion en una goroutine despues del delay:

  timer := time.AfterFunc(5*time.Second, func() {
      fmt.Println("Executed after 5s")
  })

  timer.Stop()   // Cancelar
  timer.Reset(d) // Reprogramar

CASOS DE USO:
  - Scheduled cleanup tasks
  - Delayed notifications
  - Debouncing (reset on each event)
  - Session expiration
`)

	var wg sync.WaitGroup
	wg.Add(1)

	afterStart := time.Now()
	time.AfterFunc(50*time.Millisecond, func() {
		fmt.Printf("  AfterFunc executed after: %v\n",
			time.Since(afterStart).Round(time.Millisecond))
		wg.Done()
	})
	wg.Wait()

	fmt.Println("\n  [Debounce pattern with AfterFunc]")
	os.Stdout.WriteString(`
func debounce(delay time.Duration, fn func()) func() {
    var timer *time.Timer
    var mu sync.Mutex

    return func() {
        mu.Lock()
        defer mu.Unlock()

        if timer != nil {
            timer.Stop()
        }
        timer = time.AfterFunc(delay, fn)
    }
}

// Uso:
debouncedSave := debounce(500*time.Millisecond, func() {
    fmt.Println("Saving...")
})
debouncedSave() // Reseteado
debouncedSave() // Reseteado
debouncedSave() // Solo este se ejecuta (despues de 500ms)
`)

	// ============================================
	// 12. TIME.SINCE AND TIME.UNTIL
	// ============================================
	fmt.Println("\n--- 12. time.Since and time.Until ---")
	os.Stdout.WriteString(`
SHORTCUTS para medir tiempo:

  time.Since(t) == time.Now().Sub(t)   // Cuanto paso desde t
  time.Until(t) == t.Sub(time.Now())   // Cuanto falta para t

PATRON COMUN - Medir duracion de operaciones:

  start := time.Now()
  doSomething()
  fmt.Printf("Took %v\n", time.Since(start))

PATRON - Timeout check:

  deadline := time.Now().Add(5 * time.Second)
  for {
      if time.Until(deadline) <= 0 {
          return ErrTimeout
      }
      // ... try operation ...
  }
`)

	measureStart := time.Now()
	time.Sleep(25 * time.Millisecond)
	fmt.Printf("  time.Since(start): %v\n", time.Since(measureStart).Round(time.Millisecond))

	future := time.Now().Add(1 * time.Hour)
	fmt.Printf("  time.Until(+1h):   %v\n", time.Until(future).Round(time.Second))

	// ============================================
	// 13. JSON MARSHALING OF TIME
	// ============================================
	fmt.Println("\n--- 13. JSON Marshaling of Time ---")
	os.Stdout.WriteString(`
time.Time implementa json.Marshaler/Unmarshaler:
  - Formato por defecto: RFC3339Nano
  - "2024-03-15T14:30:00Z"
  - "2024-03-15T14:30:00.123456789-05:00"

CUSTOM FORMAT - Necesitas un wrapper type:

type DateOnly time.Time

func (d DateOnly) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Time(d).Format("2006-01-02"))
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    t, err := time.Parse("2006-01-02", s)
    if err != nil {
        return err
    }
    *d = DateOnly(t)
    return nil
}
`)

	type Event struct {
		Name      string    `json:"name"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}

	event := Event{
		Name:      "Go Conference",
		StartTime: time.Date(2024, 11, 15, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2024, 11, 15, 17, 0, 0, 0, time.UTC),
	}

	jsonBytes, _ := json.MarshalIndent(event, "  ", "  ")
	fmt.Printf("  Marshal:\n  %s\n", string(jsonBytes))

	jsonStr := `{"name":"Workshop","start_time":"2024-12-01T10:00:00Z","end_time":"2024-12-01T16:00:00Z"}`
	var parsed Event
	json.Unmarshal([]byte(jsonStr), &parsed)
	fmt.Printf("\n  Unmarshal: %+v\n", parsed)

	// ============================================
	// 14. DATABASE TIME HANDLING
	// ============================================
	fmt.Println("\n--- 14. Database Time Handling ---")
	os.Stdout.WriteString(`
REGLAS PARA TIEMPO EN BASES DE DATOS:

1. SIEMPRE almacenar en UTC:
   - Evita ambiguedades por DST
   - Conversion a local timezone en la capa de presentacion

2. TIPOS DE COLUMNA:
   - TIMESTAMP WITH TIME ZONE (timestamptz): Almacena UTC, convierte automatico
   - TIMESTAMP WITHOUT TIME ZONE: Sin info de timezone (evitar)
   - DATE: Solo fecha, sin hora
   - BIGINT: Unix timestamp (portable pero sin funciones de fecha)

3. EN GO CON database/sql:
   - time.Time se mapea automaticamente a TIMESTAMP
   - Usar sql.NullTime para campos nullable
   - Verificar que el driver convierte a UTC

4. PATRON RECOMENDADO:

type Model struct {
    ID        int64
    CreatedAt time.Time     // NOT NULL DEFAULT now()
    UpdatedAt time.Time     // NOT NULL DEFAULT now()
    DeletedAt sql.NullTime  // NULL para soft delete
}

// Al insertar:
_, err := db.Exec(
    "INSERT INTO items (name, created_at) VALUES ($1, $2)",
    name, time.Now().UTC(),
)

// Al leer:
var createdAt time.Time
err := row.Scan(&createdAt)
// createdAt viene en UTC si el driver esta bien configurado

5. MIGRACION DE TIMEZONE:
   ALTER TABLE items
   ALTER COLUMN created_at TYPE timestamptz
   USING created_at AT TIME ZONE 'UTC';
`)

	// ============================================
	// 15. CLOCK ABSTRACTION FOR TESTING
	// ============================================
	fmt.Println("\n--- 15. Clock Abstraction for Testing ---")
	os.Stdout.WriteString(`
PROBLEMA: Testear codigo que depende de time.Now() es dificil.

SOLUCION: Abstraer el reloj con una interfaz:

type Clock interface {
    Now() time.Time
    Since(t time.Time) time.Duration
    After(d time.Duration) <-chan time.Time
    NewTimer(d time.Duration) *time.Timer
    NewTicker(d time.Duration) *time.Ticker
}

// Implementacion real:
type RealClock struct{}
func (RealClock) Now() time.Time { return time.Now() }
func (RealClock) Since(t time.Time) time.Duration { return time.Since(t) }
func (RealClock) After(d time.Duration) <-chan time.Time { return time.After(d) }
func (RealClock) NewTimer(d time.Duration) *time.Timer { return time.NewTimer(d) }
func (RealClock) NewTicker(d time.Duration) *time.Ticker { return time.NewTicker(d) }

// Mock para tests:
type MockClock struct {
    current time.Time
}
func (m *MockClock) Now() time.Time { return m.current }
func (m *MockClock) Advance(d time.Duration) { m.current = m.current.Add(d) }
func (m *MockClock) Since(t time.Time) time.Duration { return m.current.Sub(t) }

// Uso en produccion:
type TokenService struct {
    clock Clock
    ttl   time.Duration
}

func (s *TokenService) IsExpired(issuedAt time.Time) bool {
    return s.clock.Since(issuedAt) > s.ttl
}

// Test:
func TestTokenExpiration(t *testing.T) {
    mock := &MockClock{current: time.Date(2024,1,1,0,0,0,0,time.UTC)}
    svc := &TokenService{clock: mock, ttl: 1 * time.Hour}

    issuedAt := mock.Now()
    if svc.IsExpired(issuedAt) {
        t.Error("should not be expired yet")
    }

    mock.Advance(2 * time.Hour) // "avanzar" 2 horas
    if !svc.IsExpired(issuedAt) {
        t.Error("should be expired after 2h")
    }
}
`)

	demoClockAbstraction()

	// ============================================
	// 16. COMMON PITFALLS
	// ============================================
	fmt.Println("\n--- 16. Common Pitfalls ---")
	fmt.Println(`
ERRORES FRECUENTES CON TIME:

1. COMPARAR CON == EN VEZ DE Equal():
   t1 == t2       // INCORRECTO: compara Location pointer y monotonic
   t1.Equal(t2)   // CORRECTO: compara el instante real

2. USAR time.After EN LOOPS:
   for {
       select {
       case <-time.After(5*time.Second): // LEAK: crea timer cada iteracion
           // ...
       }
   }
   // SOLUCION: Usar time.NewTimer con Reset

3. SLEEP EN TESTS:
   time.Sleep(1*time.Second) // Hace los tests lentos y flaky
   // SOLUCION: Clock abstraction o channels

4. ASUMIR 24h = 1 DIA:
   tomorrow := now.Add(24 * time.Hour) // INCORRECTO en DST
   tomorrow := now.AddDate(0, 0, 1)    // CORRECTO

5. NO LLAMAR Ticker.Stop():
   ticker := time.NewTicker(time.Second)
   // Sin Stop() = goroutine leak

6. FORMAT/PARSE CON STRFTIME:
   t.Format("yyyy-MM-dd") // INCORRECTO (no es strftime)
   t.Format("2006-01-02") // CORRECTO (reference time)

7. ZERO TIME EN JSON:
   // time.Time zero value se serializa como "0001-01-01T00:00:00Z"
   // Usar *time.Time (pointer) para omitempty
   type Event struct {
       DeletedAt *time.Time ` + "`" + `json:"deleted_at,omitempty"` + "`" + `
   }

8. IGNORAR ERRORES DE Parse:
   t, _ := time.Parse(layout, value) // Si falla, t es zero value
   // SIEMPRE verificar el error`)

	// ============================================
	// 17. PERFORMANCE CONSIDERATIONS
	// ============================================
	fmt.Println("\n--- 17. Performance Considerations ---")
	os.Stdout.WriteString(`
TIPS DE PERFORMANCE:

1. PRECOMPUTAR LOCATIONS:
   // MALO: LoadLocation en cada request
   func handler(w http.ResponseWriter, r *http.Request) {
       loc, _ := time.LoadLocation("America/New_York") // I/O cada vez
   }

   // BUENO: Cargar una vez
   var nyLocation = mustLoadLocation("America/New_York")

2. EVITAR FORMAT/PARSE INNECESARIOS:
   // Si solo necesitas comparar, no formatees
   if t.Before(deadline) { ... }  // Rapido
   if t.Format(time.RFC3339) < deadline.Format(time.RFC3339) { ... } // Lento

3. USAR Unix() PARA COMPARACIONES RAPIDAS:
   // Comparar int64 es mas rapido que time.Time
   if t.Unix() > threshold.Unix() { ... }

4. POOL DE FORMATTERS (si formateas mucho):
   // time.Format asigna un buffer cada vez
   // Para hot paths, considerar buffer pool

5. TIMER/TICKER CLEANUP:
   // Siempre Stop() cuando ya no se necesitan
   // En Go 1.23+, GC limpia timers parados automaticamente

6. time.Now() ES BARATO:
   // Usa vDSO en Linux (sin syscall)
   // ~25ns por llamada, no te preocupes por llamarlo mucho
`)

	// ============================================
	// WORKING DEMO
	// ============================================
	fmt.Println("\n--- WORKING DEMO: Time Utilities ---")
	demoTimeUtilities()

	fmt.Println("\n=== RESUMEN EJECUTADO ===")
	fmt.Println("Todos los demos de time ejecutados exitosamente.")
}

// ============================================
// DEMO: CLOCK ABSTRACTION
// ============================================

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type MockClock struct {
	current time.Time
}

func (m *MockClock) Now() time.Time           { return m.current }
func (m *MockClock) Advance(d time.Duration)  { m.current = m.current.Add(d) }
func (m *MockClock) Set(t time.Time)          { m.current = t }

type SessionManager struct {
	clock   Clock
	timeout time.Duration
}

func (sm *SessionManager) IsExpired(lastActivity time.Time) bool {
	return sm.clock.Now().Sub(lastActivity) > sm.timeout
}

func demoClockAbstraction() {
	fmt.Println("\n  [Demo: Clock abstraction]")

	mock := &MockClock{current: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)}
	sm := &SessionManager{clock: mock, timeout: 30 * time.Minute}

	activity := mock.Now()
	fmt.Printf("  Session created at: %v\n", activity)
	fmt.Printf("  Expired? %v (just created)\n", sm.IsExpired(activity))

	mock.Advance(15 * time.Minute)
	fmt.Printf("  After 15 min - Expired? %v\n", sm.IsExpired(activity))

	mock.Advance(20 * time.Minute)
	fmt.Printf("  After 35 min - Expired? %v\n", sm.IsExpired(activity))
}

// ============================================
// DEMO: TIME UTILITIES
// ============================================

func demoTimeUtilities() {
	now := time.Now()

	fmt.Println("  1. Start/End of day:")
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	fmt.Printf("     Start: %v\n", startOfDay.Format(time.DateTime))
	fmt.Printf("     End:   %v\n", endOfDay.Format(time.DateTime))

	fmt.Println("\n  2. Start of week (Monday):")
	weekday := now.Weekday()
	daysSinceMonday := (int(weekday) - int(time.Monday) + 7) % 7
	startOfWeek := now.AddDate(0, 0, -daysSinceMonday)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(),
		0, 0, 0, 0, startOfWeek.Location())
	fmt.Printf("     Start of week: %v (%s)\n",
		startOfWeek.Format(time.DateOnly), startOfWeek.Weekday())

	fmt.Println("\n  3. Days between dates:")
	d1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	days := int(d2.Sub(d1).Hours() / 24)
	fmt.Printf("     Jan 1 to Dec 31, 2024: %d days\n", days)

	fmt.Println("\n  4. Is weekend?")
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, i)
		isWeekend := day.Weekday() == time.Saturday || day.Weekday() == time.Sunday
		fmt.Printf("     %s %s: weekend=%v\n",
			day.Format(time.DateOnly), day.Weekday(), isWeekend)
	}

	fmt.Println("\n  5. Human-readable relative time:")
	durations := []time.Duration{
		3 * time.Second,
		45 * time.Second,
		3 * time.Minute,
		2 * time.Hour,
		36 * time.Hour,
		15 * 24 * time.Hour,
		90 * 24 * time.Hour,
		400 * 24 * time.Hour,
	}
	for _, d := range durations {
		fmt.Printf("     %v ago -> %q\n", d, humanizeAge(d))
	}

	fmt.Println("\n  6. Multiple timezone display:")
	zones := []string{
		"America/New_York",
		"America/Bogota",
		"Europe/London",
		"Europe/Berlin",
		"Asia/Tokyo",
		"Australia/Sydney",
	}
	for _, zone := range zones {
		loc, err := time.LoadLocation(zone)
		if err != nil {
			continue
		}
		fmt.Printf("     %-25s %s\n", zone, now.In(loc).Format("15:04 MST"))
	}
}

func humanizeAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 30*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

/*
RESUMEN CAPITULO 70: TIME PACKAGE DEEP DIVE

time.Time:
- Representa un instante con precision de nanosegundos
- Contiene wall clock + monotonic clock + Location
- Zero value: January 1, year 1, 00:00:00 UTC
- Siempre usar Equal() para comparar, nunca ==

time.Duration:
- int64 representando nanosegundos
- Constantes: Nanosecond, Microsecond, Millisecond, Second, Minute, Hour
- No existe time.Day (DST hace que los dias no siempre sean 24h)
- ParseDuration("1h30m") para parsear strings

FORMATTING (Reference Time):
- Mon Jan 2 15:04:05 MST 2006 (numeros 1-7 en orden)
- Constantes: RFC3339, DateTime, DateOnly, TimeOnly, Kitchen
- NO usar strftime, Go tiene su propio sistema

PARSING:
- time.Parse(layout, value) -> asume UTC si no hay timezone
- time.ParseInLocation(layout, value, loc) -> asume loc
- Siempre verificar error de Parse

TIMEZONE:
- time.LoadLocation("America/New_York") -> IANA timezone
- t.In(loc) cambia representacion, no el instante
- t.Equal(t.UTC()) siempre es true
- Import "time/tzdata" para embeber en el binario

MONOTONIC CLOCK:
- time.Now() captura wall + monotonic
- Sub/Since/Until usan monotonic cuando esta disponible
- Round/Truncate/In eliminan monotonic

UNIX TIMESTAMPS:
- t.Unix(), t.UnixMilli(), t.UnixMicro(), t.UnixNano()
- time.Unix(sec, nsec), time.UnixMilli(ms)

ARITHMETIC:
- t.Add(duration), t.Sub(t2), t.AddDate(y, m, d)
- time.Since(t), time.Until(t)
- AddDate para calendario, Add para duraciones exactas

DST GOTCHAS:
- 24*time.Hour != 1 dia en DST transitions
- Usar AddDate(0, 0, 1) para "manana"
- Horas pueden no existir o ser ambiguas

TIMERS:
- time.NewTimer: se dispara una vez
- time.NewTicker: se dispara repetidamente (siempre Stop())
- time.AfterFunc: ejecuta funcion en goroutine
- Evitar time.After en loops (memory leak)

JSON:
- time.Time se serializa como RFC3339 por defecto
- Custom format requiere wrapper type con MarshalJSON/UnmarshalJSON
- Usar *time.Time para omitempty con zero values

DATABASE:
- Siempre almacenar en UTC
- Usar TIMESTAMP WITH TIME ZONE
- Convertir a local en capa de presentacion

TESTING:
- Clock interface para abstraer time.Now()
- MockClock con Advance() para tests deterministas
- Evitar time.Sleep en tests

PERFORMANCE:
- Precomputar Locations (no LoadLocation en cada request)
- time.Now() es barato (~25ns, usa vDSO)
- Siempre Stop() timers y tickers
*/
