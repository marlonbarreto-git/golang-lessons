// Package main - Chapter 105: AI and Machine Learning
// Go tiene un ecosistema creciente para AI/ML, desde
// computación numérica hasta integración con LLMs.
package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== AI Y MACHINE LEARNING EN GO ===")

	// ============================================
	// ECOSISTEMA
	// ============================================
	fmt.Println("\n--- Ecosistema AI/ML en Go ---")
	fmt.Println(`
COMPUTACIÓN NUMÉRICA:
- gonum (gonum.org/v1/gonum)
  Matrices, estadísticas, optimización, plots

MACHINE LEARNING:
- gorgonia (gorgonia.org)
  Neural networks, autodiff, GPU support

INFERENCE:
- onnxruntime-go
  Ejecutar modelos ONNX
- go-tensorflow
  TensorFlow bindings (deprecated)

LLMs:
- langchaingo (github.com/tmc/langchaingo)
  LLM applications, RAG, agents
- ollama/api (para modelos locales)
- openai-go (API de OpenAI)

VECTORES:
- qdrant-go, pinecone-go, weaviate-client-go
  Bases de datos vectoriales`)
	// ============================================
	// GONUM
	// ============================================
	fmt.Println("\n--- Gonum ---")
	fmt.Println(`
import (
    "gonum.org/v1/gonum/mat"
    "gonum.org/v1/gonum/stat"
)

// Crear matriz
data := []float64{1, 2, 3, 4, 5, 6}
m := mat.NewDense(2, 3, data)

// Operaciones
var result mat.Dense
result.Mul(m1, m2)        // Multiplicación
result.Add(m1, m2)        // Suma
result.Scale(2.0, m)      // Escalar

// Determinante, inversa
det := mat.Det(m)
var inv mat.Dense
inv.Inverse(m)

// SVD
var svd mat.SVD
svd.Factorize(m, mat.SVDFull)
u := svd.UTo(nil)
s := svd.Values(nil)
v := svd.VTo(nil)

// Estadísticas
mean := stat.Mean(data, nil)
std := stat.StdDev(data, nil)
variance := stat.Variance(data, nil)

// Correlación
corr := stat.Correlation(x, y, nil)

// Regresión lineal
alpha, beta := stat.LinearRegression(x, y, nil, false)

// Distribuciones
import "gonum.org/v1/gonum/stat/distuv"

norm := distuv.Normal{Mu: 0, Sigma: 1}
sample := norm.Rand()
prob := norm.Prob(x)
cdf := norm.CDF(x)`)
	// ============================================
	// GORGONIA
	// ============================================
	fmt.Println("\n--- Gorgonia (Neural Networks) ---")
	fmt.Println(`
import (
    "gorgonia.org/gorgonia"
    "gorgonia.org/tensor"
)

// Crear grafo computacional
g := gorgonia.NewGraph()

// Tensores
x := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(100, 784), gorgonia.WithName("x"))
y := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(100, 10), gorgonia.WithName("y"))

// Pesos
w := gorgonia.NewMatrix(g, tensor.Float64, gorgonia.WithShape(784, 10),
    gorgonia.WithName("w"), gorgonia.WithInit(gorgonia.GlorotN(1.0)))
b := gorgonia.NewVector(g, tensor.Float64, gorgonia.WithShape(10),
    gorgonia.WithName("b"), gorgonia.WithInit(gorgonia.Zeroes()))

// Forward pass
pred := gorgonia.Must(gorgonia.Add(gorgonia.Must(gorgonia.Mul(x, w)), b))
pred = gorgonia.Must(gorgonia.SoftMax(pred))

// Loss
losses := gorgonia.Must(gorgonia.HadamardProd(y, gorgonia.Must(gorgonia.Log(pred))))
loss := gorgonia.Must(gorgonia.Mean(losses))
loss = gorgonia.Must(gorgonia.Neg(loss))

// Gradientes
gorgonia.Grad(loss, w, b)

// VM para ejecución
vm := gorgonia.NewTapeMachine(g)
defer vm.Close()

// Training loop
solver := gorgonia.NewAdamSolver(gorgonia.WithLearnRate(0.01))

for epoch := 0; epoch < 100; epoch++ {
    if err := vm.RunAll(); err != nil {
        log.Fatal(err)
    }
    solver.Step(gorgonia.NodesToValueGrads(gorgonia.Nodes{w, b}))
    vm.Reset()
}`)
	// ============================================
	// ONNX RUNTIME
	// ============================================
	fmt.Println("\n--- ONNX Runtime ---")
	fmt.Println(`
import "github.com/yalue/onnxruntime_go"

// Inicializar
onnxruntime_go.SetSharedLibraryPath("path/to/libonnxruntime.so")
err := onnxruntime_go.InitializeEnvironment()
defer onnxruntime_go.DestroyEnvironment()

// Cargar modelo
session, err := onnxruntime_go.NewAdvancedSession(
    "model.onnx",
    []string{"input"},
    []string{"output"},
    nil,
)
defer session.Destroy()

// Preparar input
inputShape := onnxruntime_go.NewShape(1, 3, 224, 224)
inputTensor, err := onnxruntime_go.NewTensor(inputShape, inputData)
defer inputTensor.Destroy()

// Inferencia
outputTensor, err := onnxruntime_go.NewEmptyTensor[float32](onnxruntime_go.NewShape(1, 1000))
defer outputTensor.Destroy()

err = session.Run()

// Obtener resultados
output := outputTensor.GetData()

CASOS DE USO:
- Modelos entrenados en Python (PyTorch, TensorFlow)
- Inferencia en producción con Go
- Edge deployment`)
	// ============================================
	// LANGCHAINGO
	// ============================================
	fmt.Println("\n--- LangChainGo (LLMs) ---")
	fmt.Println(`
import (
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/chains"
    "github.com/tmc/langchaingo/memory"
)

// Cliente OpenAI
llm, err := openai.New(
    openai.WithModel("gpt-4"),
    openai.WithToken(os.Getenv("OPENAI_API_KEY")),
)

// Llamada simple
response, err := llms.GenerateFromSinglePrompt(ctx, llm, "Explain Go in one sentence")

// Con streaming
llm.Call(ctx, prompt, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
    fmt.Print(string(chunk))
    return nil
}))

// Cadenas (Chains)
chain := chains.NewLLMChain(llm, prompts.NewPromptTemplate(
    "Translate the following to {{.language}}: {{.text}}",
    []string{"language", "text"},
))

result, err := chains.Call(ctx, chain, map[string]any{
    "language": "Spanish",
    "text":     "Hello, world!",
})

// Memoria para conversaciones
mem := memory.NewConversationBuffer()
conversationChain := chains.NewConversation(llm, mem)

response1, _ := chains.Run(ctx, conversationChain, "Hi, I'm Alice")
response2, _ := chains.Run(ctx, conversationChain, "What's my name?")
// El modelo recuerda que eres Alice

// RAG (Retrieval Augmented Generation)
import (
    "github.com/tmc/langchaingo/vectorstores/qdrant"
    "github.com/tmc/langchaingo/embeddings"
)

embedder, _ := embeddings.NewEmbedder(llm)
store, _ := qdrant.New(
    qdrant.WithURL("http://localhost:6333"),
    qdrant.WithEmbedder(embedder),
)

// Agregar documentos
store.AddDocuments(ctx, docs)

// Buscar documentos relevantes
retriever := vectorstores.ToRetriever(store, 5)
relevantDocs, _ := retriever.GetRelevantDocuments(ctx, query)

// Cadena RAG
qaChain := chains.NewRetrievalQAFromLLM(llm, retriever)
answer, _ := chains.Run(ctx, qaChain, "What is Go?")`)
	// ============================================
	// OLLAMA
	// ============================================
	fmt.Println("\n--- Ollama (Modelos Locales) ---")
	fmt.Println(`
import "github.com/ollama/ollama/api"

// Cliente
client, err := api.ClientFromEnvironment()

// Generar
req := &api.GenerateRequest{
    Model:  "llama2",
    Prompt: "Why is the sky blue?",
}

err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
    fmt.Print(resp.Response)
    return nil
})

// Chat
messages := []api.Message{
    {Role: "system", Content: "You are a helpful assistant."},
    {Role: "user", Content: "Hello!"},
}

req := &api.ChatRequest{
    Model:    "llama2",
    Messages: messages,
}

err = client.Chat(ctx, req, func(resp api.ChatResponse) error {
    fmt.Print(resp.Message.Content)
    return nil
})

// Embeddings
req := &api.EmbeddingRequest{
    Model:  "nomic-embed-text",
    Prompt: "Hello, world!",
}

resp, err := client.Embeddings(ctx, req)
embedding := resp.Embedding  // []float64`)
	// ============================================
	// VECTORES Y EMBEDDINGS
	// ============================================
	fmt.Println("\n--- Bases de Datos Vectoriales ---")
	fmt.Println(`
QDRANT:
import "github.com/qdrant/go-client/qdrant"

client, _ := qdrant.NewClient(&qdrant.Config{
    Host: "localhost",
    Port: 6334,
})

// Crear colección
client.CreateCollection(ctx, &qdrant.CreateCollection{
    CollectionName: "my_collection",
    VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
        Size:     384,
        Distance: qdrant.Distance_Cosine,
    }),
})

// Insertar vectores
client.Upsert(ctx, &qdrant.UpsertPoints{
    CollectionName: "my_collection",
    Points: []*qdrant.PointStruct{
        {
            Id:      qdrant.NewIDNum(1),
            Vectors: qdrant.NewVectors(embedding...),
            Payload: qdrant.NewValueMap(map[string]any{
                "text": "Hello world",
            }),
        },
    },
})

// Buscar
results, _ := client.Search(ctx, &qdrant.SearchPoints{
    CollectionName: "my_collection",
    Vector:         queryEmbedding,
    Limit:          5,
})

PINECONE:
import "github.com/pinecone-io/go-pinecone/pinecone"

pc, _ := pinecone.NewClient(pinecone.NewClientParams{
    ApiKey: os.Getenv("PINECONE_API_KEY"),
})

idx := pc.Index("my-index")
idx.Upsert(ctx, vectors)
results, _ := idx.Query(ctx, queryVector, 5)`)
	// ============================================
	// EJEMPLO: CLASIFICADOR SIMPLE
	// ============================================
	fmt.Println("\n--- Ejemplo: Clasificador Simple ---")
	fmt.Println(`
// Clasificador de sentimiento simple con embeddings

type SentimentClassifier struct {
    llm       llms.Model
    embedder  embeddings.Embedder
    positives [][]float32
    negatives [][]float32
}

func (c *SentimentClassifier) Train(positiveTexts, negativeTexts []string) error {
    ctx := context.Background()

    for _, text := range positiveTexts {
        emb, _ := c.embedder.EmbedQuery(ctx, text)
        c.positives = append(c.positives, emb)
    }

    for _, text := range negativeTexts {
        emb, _ := c.embedder.EmbedQuery(ctx, text)
        c.negatives = append(c.negatives, emb)
    }

    return nil
}

func (c *SentimentClassifier) Predict(text string) string {
    ctx := context.Background()
    emb, _ := c.embedder.EmbedQuery(ctx, text)

    posAvg := averageSimilarity(emb, c.positives)
    negAvg := averageSimilarity(emb, c.negatives)

    if posAvg > negAvg {
        return "positive"
    }
    return "negative"
}

func cosineSimilarity(a, b []float32) float32 {
    var dot, normA, normB float32
    for i := range a {
        dot += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    return dot / (sqrt(normA) * sqrt(normB))
}`)}

/*
RESUMEN DE AI/ML:

COMPUTACIÓN NUMÉRICA:
gonum - matrices, estadísticas, plots

NEURAL NETWORKS:
gorgonia - autodiff, GPU, training

INFERENCE:
onnxruntime-go - modelos ONNX
Ideal para modelos entrenados en Python

LLMs:
langchaingo - OpenAI, chains, RAG
ollama/api - modelos locales

VECTORES:
qdrant-go, pinecone-go
Embeddings + búsqueda semántica

FLUJO TÍPICO:
1. Entrenar modelo en Python
2. Exportar a ONNX
3. Inferencia en Go
4. O usar LLMs via API

CASOS DE USO:
- Inference en producción (Go servers)
- Pipelines de datos
- Aplicaciones LLM
- Embedding-based search
*/

/*
SUMMARY - CHAPTER 105: AI and Machine Learning

TOPIC: AI/ML Integration in Go
- Gonum: Numerical computing library for matrices (mat.NewDense), statistics (Mean/StdDev/Correlation), linear algebra
- Gonum capabilities: Matrix operations (Mul/Add/Scale), SVD decomposition, eigenvalues, distributions, linear regression
- Gorgonia: Neural network framework with computational graphs, automatic differentiation, GPU support
- Gorgonia training: Define graph with tensors/weights, forward pass, loss calculation, gradient computation, solver optimization
- ONNX Runtime: Execute models trained in Python (PyTorch/TensorFlow), load .onnx files, prepare tensors, run inference
- ONNX workflow: Train model in Python ecosystem, export to ONNX format, deploy inference in Go production servers
- LangChainGo: Framework for LLM applications, supports OpenAI/Ollama, chains for complex workflows, conversation memory
- LangChain patterns: Simple prompts, prompt templates, conversation chains with memory, streaming responses
- RAG (Retrieval Augmented Generation): Embeddings for documents, vector store (Qdrant/Pinecone), retriever for context
- Ollama: Run local models (Llama2, Mistral), generate text, chat interface, embeddings for semantic search
- Vector databases: Qdrant/Pinecone for similarity search, store embeddings with metadata, cosine similarity queries
- Embeddings: Convert text to vectors, semantic search, clustering, classification with similarity comparison
- Use cases: Production inference in Go servers, data pipelines, LLM-powered applications, semantic search systems
*/
