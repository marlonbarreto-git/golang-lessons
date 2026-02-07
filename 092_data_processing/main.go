// Package main - Chapter 092: Data Processing
// Go es excelente para ETL, pipelines de datos, y análisis.
// Aprenderás gonum, formatos de datos, y patrones de procesamiento.
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func main() {
	fmt.Println("=== PROCESAMIENTO DE DATOS EN GO ===")

	// ============================================
	// CSV
	// ============================================
	fmt.Println("\n--- CSV ---")

	// Escribir CSV
	csvFile, err := os.Create("/tmp/data.csv")
	if err == nil {
		writer := csv.NewWriter(csvFile)

		// Escribir header
		writer.Write([]string{"id", "name", "value"})

		// Escribir datos
		data := [][]string{
			{"1", "Alice", "100"},
			{"2", "Bob", "200"},
			{"3", "Charlie", "300"},
		}
		for _, row := range data {
			writer.Write(row)
		}

		writer.Flush()
		csvFile.Close()
		fmt.Println("CSV escrito: /tmp/data.csv")
	}

	// Leer CSV
	readFile, err := os.Open("/tmp/data.csv")
	if err == nil {
		reader := csv.NewReader(readFile)
		records, _ := reader.ReadAll()
		fmt.Printf("CSV leído: %d filas\n", len(records))
		for _, record := range records {
			fmt.Printf("  %v\n", record)
		}
		readFile.Close()
	}

	// ============================================
	// JSON LINES (JSONL)
	// ============================================
	fmt.Println("\n--- JSON Lines ---")

	type Record struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Value float64 `json:"value"`
	}

	// Escribir JSONL
	jsonlFile, err := os.Create("/tmp/data.jsonl")
	if err == nil {
		encoder := json.NewEncoder(jsonlFile)
		records := []Record{
			{ID: 1, Name: "Alice", Value: 100.5},
			{ID: 2, Name: "Bob", Value: 200.3},
			{ID: 3, Name: "Charlie", Value: 300.7},
		}
		for _, r := range records {
			encoder.Encode(r)
		}
		jsonlFile.Close()
		fmt.Println("JSONL escrito: /tmp/data.jsonl")
	}

	// Leer JSONL
	readJsonl, err := os.Open("/tmp/data.jsonl")
	if err == nil {
		decoder := json.NewDecoder(readJsonl)
		for decoder.More() {
			var r Record
			if err := decoder.Decode(&r); err == nil {
				fmt.Printf("  %+v\n", r)
			}
		}
		readJsonl.Close()
	}

	// ============================================
	// GONUM
	// ============================================
	fmt.Println("\n--- Gonum (Estadísticas) ---")
	os.Stdout.WriteString(`
import (
    "gonum.org/v1/gonum/stat"
    "gonum.org/v1/gonum/floats"
)

// Estadísticas básicas
data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

mean := stat.Mean(data, nil)
variance := stat.Variance(data, nil)
stddev := stat.StdDev(data, nil)
median := stat.Quantile(0.5, stat.Empirical, data, nil)

fmt.Printf("Mean: %.2f, StdDev: %.2f, Median: %.2f\n", mean, stddev, median)

// Percentiles
p25 := stat.Quantile(0.25, stat.Empirical, data, nil)
p75 := stat.Quantile(0.75, stat.Empirical, data, nil)
p99 := stat.Quantile(0.99, stat.Empirical, data, nil)

// Correlación
x := []float64{1, 2, 3, 4, 5}
y := []float64{2, 4, 5, 8, 10}
corr := stat.Correlation(x, y, nil)

// Regresión lineal
alpha, beta := stat.LinearRegression(x, y, nil, false)
// y = alpha + beta*x

// Operaciones con slices
floats.Add(dst, s)       // dst += s
floats.Scale(c, dst)     // dst *= c
floats.Dot(s1, s2)       // producto punto
floats.Sum(s)            // suma
floats.Max(s)            // máximo
floats.Min(s)            // mínimo
`)

	// ============================================
	// MATRICES
	// ============================================
	fmt.Println("\n--- Gonum Matrices ---")
	fmt.Println(`
import "gonum.org/v1/gonum/mat"

// Crear matriz
data := []float64{1, 2, 3, 4, 5, 6}
m := mat.NewDense(2, 3, data)  // 2x3

// Acceso
val := m.At(0, 1)      // Obtener elemento
m.Set(0, 1, 10)        // Establecer elemento
rows, cols := m.Dims() // Dimensiones

// Operaciones
var result mat.Dense
result.Add(m1, m2)        // Suma
result.Sub(m1, m2)        // Resta
result.Mul(m1, m2)        // Multiplicación
result.Scale(2.0, m)      // Escalar
result.Apply(func(i, j int, v float64) float64 {
    return v * 2
}, m)                     // Aplicar función

// Transpuesta
result.T()

// Determinante
det := mat.Det(m)

// Inversa
var inv mat.Dense
inv.Inverse(m)

// SVD
var svd mat.SVD
svd.Factorize(m, mat.SVDFull)

// Eigenvalues
var eig mat.Eigen
eig.Factorize(m, mat.EigenRight)`)
	// ============================================
	// APACHE ARROW
	// ============================================
	fmt.Println("\n--- Apache Arrow ---")
	fmt.Println(`
import "github.com/apache/arrow/go/v14/arrow"
import "github.com/apache/arrow/go/v14/arrow/array"
import "github.com/apache/arrow/go/v14/arrow/memory"

// Pool de memoria
pool := memory.NewGoAllocator()

// Crear array de int64
builder := array.NewInt64Builder(pool)
defer builder.Release()

builder.Append(1)
builder.Append(2)
builder.Append(3)
builder.AppendNull()
builder.Append(5)

arr := builder.NewArray()
defer arr.Release()

// Acceso
for i := 0; i < arr.Len(); i++ {
    if arr.IsNull(i) {
        fmt.Println("null")
    } else {
        fmt.Println(arr.Value(i))
    }
}

// Schema y Record Batch
schema := arrow.NewSchema(
    []arrow.Field{
        {Name: "id", Type: arrow.PrimitiveTypes.Int64},
        {Name: "name", Type: arrow.BinaryTypes.String},
        {Name: "value", Type: arrow.PrimitiveTypes.Float64},
    },
    nil,
)

// Crear record batch
cols := []arrow.Array{idArray, nameArray, valueArray}
batch := array.NewRecord(schema, cols, -1)
defer batch.Release()

// Parquet
import "github.com/apache/arrow/go/v14/parquet/pqarrow"

// Leer Parquet
reader, _ := pqarrow.NewFileReader(file, props, pool)
table, _ := reader.ReadTable(ctx)

// Escribir Parquet
writer, _ := pqarrow.NewFileWriter(schema, file, props, writerProps)
writer.WriteTable(table, chunkSize)
writer.Close()`)
	// ============================================
	// PIPELINE PATTERN
	// ============================================
	fmt.Println("\n--- Pipeline de Datos ---")

	// Ejemplo simple de pipeline
	source := func() <-chan map[string]string {
		out := make(chan map[string]string)
		go func() {
			defer close(out)
			data := []map[string]string{
				{"name": "Alice", "value": "100"},
				{"name": "Bob", "value": "200"},
				{"name": "Charlie", "value": "300"},
			}
			for _, d := range data {
				out <- d
			}
		}()
		return out
	}

	transform := func(in <-chan map[string]string) <-chan map[string]any {
		out := make(chan map[string]any)
		go func() {
			defer close(out)
			for record := range in {
				val, _ := strconv.Atoi(record["value"])
				out <- map[string]any{
					"name":   record["name"],
					"value":  val,
					"doubled": val * 2,
				}
			}
		}()
		return out
	}

	sink := func(in <-chan map[string]any) {
		for record := range in {
			fmt.Printf("  Processed: %v\n", record)
		}
	}

	// Ejecutar pipeline
	fmt.Println("Pipeline ejecutándose:")
	sink(transform(source()))

	// ============================================
	// ETL PATTERNS
	// ============================================
	fmt.Println("\n--- Patrones ETL ---")
	fmt.Println(`
// Extract
type Extractor interface {
    Extract(ctx context.Context) (<-chan Record, error)
}

// Transform
type Transformer interface {
    Transform(ctx context.Context, in <-chan Record) <-chan Record
}

// Load
type Loader interface {
    Load(ctx context.Context, in <-chan Record) error
}

// Pipeline
type Pipeline struct {
    extractor   Extractor
    transformers []Transformer
    loader      Loader
}

func (p *Pipeline) Run(ctx context.Context) error {
    records, err := p.extractor.Extract(ctx)
    if err != nil {
        return err
    }

    for _, t := range p.transformers {
        records = t.Transform(ctx, records)
    }

    return p.loader.Load(ctx, records)
}

// Ejemplo de transformador
type FilterTransformer struct {
    predicate func(Record) bool
}

func (t *FilterTransformer) Transform(ctx context.Context, in <-chan Record) <-chan Record {
    out := make(chan Record)
    go func() {
        defer close(out)
        for r := range in {
            if t.predicate(r) {
                select {
                case out <- r:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return out
}`)
	// ============================================
	// STREAMING AGGREGATIONS
	// ============================================
	fmt.Println("\n--- Agregaciones en Streaming ---")
	fmt.Println(`
// Running statistics sin guardar todos los datos
type RunningStats struct {
    count int64
    mean  float64
    m2    float64  // Para varianza
    min   float64
    max   float64
}

func (s *RunningStats) Add(value float64) {
    s.count++
    delta := value - s.mean
    s.mean += delta / float64(s.count)
    delta2 := value - s.mean
    s.m2 += delta * delta2

    if s.count == 1 || value < s.min {
        s.min = value
    }
    if s.count == 1 || value > s.max {
        s.max = value
    }
}

func (s *RunningStats) Variance() float64 {
    if s.count < 2 {
        return 0
    }
    return s.m2 / float64(s.count-1)
}

// HyperLogLog para count distinct aproximado
// BloomFilter para membership testing
// CountMinSketch para frecuencias aproximadas`)}

/*
RESUMEN DE DATA:

FORMATOS:
encoding/csv - CSV
encoding/json - JSON, JSONL
apache/arrow/go - Arrow, Parquet

ESTADÍSTICAS (gonum):
stat.Mean(), stat.StdDev()
stat.Variance(), stat.Quantile()
stat.Correlation()
stat.LinearRegression()

MATRICES (gonum/mat):
mat.NewDense()
Add, Mul, Inverse, Det
SVD, Eigenvalues

PIPELINE PATTERN:
Source -> Transform -> Transform -> Sink
Channels para streaming

ETL:
Extract: leer de fuente
Transform: procesar
Load: escribir a destino

STREAMING:
- Running statistics
- Approximate algorithms
- Backpressure handling

BIBLIOTECAS:
- gonum: estadísticas, matrices
- arrow: columnar format
- gota: DataFrames (like pandas)
*/

/*
SUMMARY - CHAPTER 092: Data Processing

TOPIC: Data Processing and ETL in Go
- CSV processing: encoding/csv Reader/Writer, ReadAll for small files, streaming with Read() for large files
- JSON Lines (JSONL): One JSON object per line, use json.Encoder/Decoder for streaming, efficient for logs
- Gonum statistics: Mean(), StdDev(), Variance(), Quantile() for percentiles, Correlation(), LinearRegression()
- Gonum matrices: mat.NewDense() for matrix creation, operations (Add/Mul/Scale), Inverse/Det/SVD/Eigenvalues
- Apache Arrow: Columnar memory format, array builders (Int64Builder, StringBuilder), schema and record batches
- Parquet: Compressed columnar format, pqarrow.NewFileReader/Writer for efficient storage and queries
- Pipeline pattern: Source (generates data) -> Transform (processes) -> Sink (consumes), use channels for flow
- ETL architecture: Extractor interface (read data), Transformer interface (process), Loader interface (write)
- Composable transformers: FilterTransformer, MapTransformer, chain multiple transformers in pipeline
- Streaming aggregations: RunningStats for mean/variance without storing all data, Welford's algorithm for numerics
- Approximate algorithms: HyperLogLog (count distinct), Bloom Filter (membership), CountMinSketch (frequency)
- Pipeline coordination: Context for cancellation, buffered channels for backpressure, goroutines for parallelism
- Best practices: Use channels for streaming, handle context cancellation, implement proper cleanup with defer
*/
