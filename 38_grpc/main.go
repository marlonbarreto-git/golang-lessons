// Package main - Capítulo 38: gRPC
// gRPC es un framework RPC de alto rendimiento de Google.
// Usa Protocol Buffers para serialización y HTTP/2 para transporte.
package main

import (
	"os"
	"fmt"
)

func main() {
	fmt.Println("=== gRPC EN GO ===")

	// ============================================
	// INTRODUCCIÓN
	// ============================================
	fmt.Println("\n--- Introducción ---")
	fmt.Println(`
gRPC: Google Remote Procedure Call

CARACTERÍSTICAS:
- Protocol Buffers para serialización (eficiente, tipado)
- HTTP/2 para transporte (multiplexing, streaming)
- Generación de código cliente/servidor
- Soporte para múltiples lenguajes
- Streaming bidireccional
- Interceptors (middleware)

INSTALACIÓN:
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

DEPENDENCIAS:
go get google.golang.org/grpc
go get google.golang.org/protobuf`)
	// ============================================
	// PROTOCOL BUFFERS
	// ============================================
	fmt.Println("\n--- Protocol Buffers ---")
	fmt.Println(`
// user.proto
syntax = "proto3";

package user;
option go_package = "github.com/myapp/proto/user";

// Mensaje
message User {
    int64 id = 1;
    string name = 2;
    string email = 3;
    repeated string roles = 4;  // array
    optional string phone = 5;  // opcional (Go 1.20+)
}

// Servicio
service UserService {
    // Unary RPC
    rpc GetUser(GetUserRequest) returns (User);
    rpc CreateUser(CreateUserRequest) returns (User);

    // Server streaming
    rpc ListUsers(ListUsersRequest) returns (stream User);

    // Client streaming
    rpc UploadUsers(stream User) returns (UploadResponse);

    // Bidirectional streaming
    rpc Chat(stream Message) returns (stream Message);
}

message GetUserRequest {
    int64 id = 1;
}

message CreateUserRequest {
    string name = 1;
    string email = 2;
}

message ListUsersRequest {
    int32 page_size = 1;
    string page_token = 2;
}

message UploadResponse {
    int32 count = 1;
}

message Message {
    string content = 1;
    int64 timestamp = 2;
}

GENERAR CÓDIGO:
protoc --go_out=. --go-grpc_out=. proto/*.proto`)
	// ============================================
	// SERVIDOR
	// ============================================
	fmt.Println("\n--- Servidor ---")
	os.Stdout.WriteString(`
package main

import (
    "context"
    "log"
    "net"

    "google.golang.org/grpc"
    pb "github.com/myapp/proto/user"
)

type userServer struct {
    pb.UnimplementedUserServiceServer
    users map[int64]*pb.User
}

func (s *userServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    user, ok := s.users[req.Id]
    if !ok {
        return nil, status.Errorf(codes.NotFound, "user %d not found", req.Id)
    }
    return user, nil
}

func (s *userServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
    user := &pb.User{
        Id:    int64(len(s.users) + 1),
        Name:  req.Name,
        Email: req.Email,
    }
    s.users[user.Id] = user
    return user, nil
}

// Server streaming
func (s *userServer) ListUsers(req *pb.ListUsersRequest, stream pb.UserService_ListUsersServer) error {
    for _, user := range s.users {
        if err := stream.Send(user); err != nil {
            return err
        }
    }
    return nil
}

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    pb.RegisterUserServiceServer(grpcServer, &userServer{
        users: make(map[int64]*pb.User),
    })

    log.Println("gRPC server listening on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
`)

	// ============================================
	// CLIENTE
	// ============================================
	fmt.Println("\n--- Cliente ---")
	os.Stdout.WriteString(`
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/myapp/proto/user"
)

func main() {
    // Conectar
    conn, err := grpc.NewClient("localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewUserServiceClient(conn)

    // Context con timeout
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    // Unary RPC
    user, err := client.GetUser(ctx, &pb.GetUserRequest{Id: 1})
    if err != nil {
        log.Fatalf("GetUser failed: %v", err)
    }
    log.Printf("User: %v", user)

    // Create user
    newUser, err := client.CreateUser(ctx, &pb.CreateUserRequest{
        Name:  "Alice",
        Email: "alice@example.com",
    })
    if err != nil {
        log.Fatalf("CreateUser failed: %v", err)
    }
    log.Printf("Created: %v", newUser)

    // Server streaming
    stream, err := client.ListUsers(ctx, &pb.ListUsersRequest{})
    if err != nil {
        log.Fatalf("ListUsers failed: %v", err)
    }
    for {
        user, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatalf("stream error: %v", err)
        }
        log.Printf("User: %v", user)
    }
}
`)

	// ============================================
	// STREAMING
	// ============================================
	fmt.Println("\n--- Streaming ---")
	fmt.Println(`
TIPOS DE STREAMING:

1. SERVER STREAMING:
   Cliente envía un request, servidor envía stream de responses.
   rpc ListUsers(Request) returns (stream User);

2. CLIENT STREAMING:
   Cliente envía stream de requests, servidor envía un response.
   rpc UploadUsers(stream User) returns (Response);

3. BIDIRECTIONAL STREAMING:
   Ambos lados pueden enviar mensajes independientemente.
   rpc Chat(stream Message) returns (stream Message);

// Client streaming - servidor
func (s *server) UploadUsers(stream pb.UserService_UploadUsersServer) error {
    var count int32
    for {
        user, err := stream.Recv()
        if err == io.EOF {
            return stream.SendAndClose(&pb.UploadResponse{Count: count})
        }
        if err != nil {
            return err
        }
        // Procesar user
        count++
    }
}

// Bidirectional streaming - servidor
func (s *server) Chat(stream pb.ChatService_ChatServer) error {
    for {
        msg, err := stream.Recv()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        // Procesar y responder
        response := &pb.Message{Content: "Got: " + msg.Content}
        if err := stream.Send(response); err != nil {
            return err
        }
    }
}`)
	// ============================================
	// INTERCEPTORS
	// ============================================
	fmt.Println("\n--- Interceptors (Middleware) ---")
	os.Stdout.WriteString(`
// Unary interceptor
func loggingInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    start := time.Now()
    log.Printf("Request: %s", info.FullMethod)

    resp, err := handler(ctx, req)

    log.Printf("Response: %s took %v", info.FullMethod, time.Since(start))
    return resp, err
}

// Stream interceptor
func streamInterceptor(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {
    log.Printf("Stream: %s", info.FullMethod)
    return handler(srv, ss)
}

// Registrar interceptors
server := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
    grpc.StreamInterceptor(streamInterceptor),
)

// Múltiples interceptors (chain)
import "google.golang.org/grpc/middleware"

server := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        loggingInterceptor,
        authInterceptor,
        recoveryInterceptor,
    ),
)

// Client interceptors
conn, err := grpc.NewClient("localhost:50051",
    grpc.WithUnaryInterceptor(clientLoggingInterceptor),
)
`)

	// ============================================
	// ERROR HANDLING
	// ============================================
	fmt.Println("\n--- Error Handling ---")
	os.Stdout.WriteString(`
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// Crear error con código
err := status.Errorf(codes.NotFound, "user %d not found", id)
err := status.Error(codes.InvalidArgument, "email is required")

// Con detalles
st := status.New(codes.InvalidArgument, "validation failed")
st, _ = st.WithDetails(&errdetails.BadRequest{
    FieldViolations: []*errdetails.BadRequest_FieldViolation{
        {Field: "email", Description: "invalid format"},
    },
})
return nil, st.Err()

// En el cliente
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        log.Printf("Code: %v, Message: %s", st.Code(), st.Message())
        for _, detail := range st.Details() {
            // Procesar detalles
        }
    }
}

CÓDIGOS COMUNES:
codes.OK              - Éxito
codes.InvalidArgument - Argumentos inválidos
codes.NotFound        - Recurso no encontrado
codes.AlreadyExists   - Recurso ya existe
codes.PermissionDenied - Sin permisos
codes.Unauthenticated - No autenticado
codes.Internal        - Error interno
codes.Unavailable     - Servicio no disponible
codes.DeadlineExceeded - Timeout
`)

	// ============================================
	// METADATA
	// ============================================
	fmt.Println("\n--- Metadata ---")
	fmt.Println(`
import "google.golang.org/grpc/metadata"

// Cliente: enviar metadata
md := metadata.Pairs(
    "authorization", "Bearer token123",
    "x-request-id", "abc-123",
)
ctx := metadata.NewOutgoingContext(ctx, md)
resp, err := client.GetUser(ctx, req)

// Servidor: leer metadata
md, ok := metadata.FromIncomingContext(ctx)
if ok {
    if tokens := md.Get("authorization"); len(tokens) > 0 {
        token := tokens[0]
        // Validar token
    }
}

// Servidor: enviar metadata en response
header := metadata.Pairs("x-response-id", "xyz-789")
grpc.SendHeader(ctx, header)

trailer := metadata.Pairs("x-processing-time", "100ms")
grpc.SetTrailer(ctx, trailer)

// Cliente: recibir metadata
var header, trailer metadata.MD
resp, err := client.GetUser(ctx, req,
    grpc.Header(&header),
    grpc.Trailer(&trailer),
)`)
	// ============================================
	// TLS
	// ============================================
	fmt.Println("\n--- TLS ---")
	fmt.Println(`
import "google.golang.org/grpc/credentials"

// Servidor con TLS
creds, err := credentials.NewServerTLSFromFile("cert.pem", "key.pem")
if err != nil {
    log.Fatal(err)
}
server := grpc.NewServer(grpc.Creds(creds))

// Cliente con TLS
creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
if err != nil {
    log.Fatal(err)
}
conn, err := grpc.NewClient("localhost:50051",
    grpc.WithTransportCredentials(creds),
)

// Mutual TLS
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    ClientCAs:    certPool,
    ClientAuth:   tls.RequireAndVerifyClientCert,
}
creds := credentials.NewTLS(tlsConfig)`)
	// ============================================
	// GRPC-GATEWAY
	// ============================================
	fmt.Println("\n--- gRPC-Gateway (REST) ---")
	fmt.Println(`
// Exponer gRPC como REST API

// En el .proto:
import "google/api/annotations.proto";

service UserService {
    rpc GetUser(GetUserRequest) returns (User) {
        option (google.api.http) = {
            get: "/v1/users/{id}"
        };
    }
    rpc CreateUser(CreateUserRequest) returns (User) {
        option (google.api.http) = {
            post: "/v1/users"
            body: "*"
        };
    }
}

// Generar:
protoc --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
       proto/*.proto

// En el servidor:
mux := runtime.NewServeMux()
err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)

http.ListenAndServe(":8080", mux)

// Ahora soporta:
// GET  /v1/users/123
// POST /v1/users {"name": "Alice", "email": "..."}`)
	// ============================================
	// HEALTH CHECKS
	// ============================================
	fmt.Println("\n--- Health Checks ---")
	fmt.Println(`
import "google.golang.org/grpc/health"
import "google.golang.org/grpc/health/grpc_health_v1"

// Servidor
healthServer := health.NewServer()
grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

// Establecer estado
healthServer.SetServingStatus("myservice", grpc_health_v1.HealthCheckResponse_SERVING)

// Cliente
healthClient := grpc_health_v1.NewHealthClient(conn)
resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
    Service: "myservice",
})
if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
    // Servicio disponible
}`)}

/*
RESUMEN DE gRPC:

COMPONENTES:
- Protocol Buffers (.proto)
- protoc (compilador)
- protoc-gen-go, protoc-gen-go-grpc

TIPOS DE RPC:
1. Unary: request -> response
2. Server streaming: request -> stream responses
3. Client streaming: stream requests -> response
4. Bidirectional: stream -> stream

SERVIDOR:
grpc.NewServer()
pb.RegisterXxxServer(server, impl)
server.Serve(listener)

CLIENTE:
grpc.NewClient(addr, opts...)
pb.NewXxxClient(conn)
client.Method(ctx, req)

INTERCEPTORS:
grpc.UnaryInterceptor(fn)
grpc.StreamInterceptor(fn)
grpc.ChainUnaryInterceptor(fn1, fn2)

ERRORES:
status.Errorf(codes.NotFound, "msg")
status.FromError(err)

METADATA:
metadata.NewOutgoingContext(ctx, md)
metadata.FromIncomingContext(ctx)

TLS:
credentials.NewServerTLSFromFile()
credentials.NewClientTLSFromFile()

EXTRAS:
- gRPC-Gateway: REST API
- Health checks
- Reflection para debugging
*/
