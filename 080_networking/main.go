// Package main - Chapter 080: Networking
// Programacion de redes de bajo nivel: TCP, UDP, Unix sockets, TLS, mTLS,
// DNS lookups, net/netip, protocolos custom y patrones de produccion.
package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/netip"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== LOW-LEVEL NETWORK PROGRAMMING EN GO ===")

	// ============================================
	// 1. NET PACKAGE OVERVIEW
	// ============================================
	fmt.Println("\n--- 1. net Package Overview ---")
	fmt.Println(`
El paquete "net" es la base de toda la programacion de red en Go.
Provee una interfaz portable para I/O de red incluyendo TCP/IP,
UDP, domain name resolution y Unix domain sockets.

PROTOCOLOS SOPORTADOS:

  TCP  - Transmission Control Protocol (confiable, orientado a conexion)
  UDP  - User Datagram Protocol (no confiable, sin conexion)
  IP   - Internet Protocol (raw sockets)
  Unix - Unix domain sockets (comunicacion local entre procesos)

FUNCIONES PRINCIPALES:

  net.Listen(network, address)     -> net.Listener (TCP server)
  net.Dial(network, address)       -> net.Conn (TCP/UDP client)
  net.ListenPacket(network, addr)  -> net.PacketConn (UDP server)
  net.ResolveXXXAddr(...)          -> Resuelve direcciones especificas

REDES VALIDAS:

  "tcp", "tcp4", "tcp6"     - TCP (IPv4+6, solo IPv4, solo IPv6)
  "udp", "udp4", "udp6"     - UDP
  "ip",  "ip4",  "ip6"      - IP raw sockets
  "unix", "unixgram", "unixpacket" - Unix domain sockets

FORMATO DE DIRECCIONES:

  "host:port"       -> "localhost:8080", "192.168.1.1:443"
  ":port"           -> ":8080" (escuchar en todas las interfaces)
  "host%zone:port"  -> IPv6 con zone ID
  "[::1]:8080"      -> IPv6 literal con puerto

MODELO I/O:

Go usa goroutines + I/O bloqueante (internamente es epoll/kqueue).
Cada goroutine bloquea en Read/Write, pero el runtime multiplexa
eficientemente sobre pocos OS threads. NO necesitas event loops,
select/poll, o callbacks como en C/Java/Node.`)

	// ============================================
	// 2. TCP SERVER
	// ============================================
	fmt.Println("\n--- 2. TCP Server ---")
	os.Stdout.WriteString(`
TCP SERVER BASICO:

func runTCPServer(addr string) error {
    // Listen crea un socket, bind y listen
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("listen: %w", err)
    }
    defer listener.Close()

    fmt.Printf("TCP server listening on %s\n", addr)

    for {
        // Accept bloquea hasta recibir una conexion
        conn, err := listener.Accept()
        if err != nil {
            // Si el listener fue cerrado, Accept retorna error
            return fmt.Errorf("accept: %w", err)
        }

        // conn es net.Conn - interfaz que implementa io.Reader + io.Writer
        fmt.Printf("New connection from %s\n", conn.RemoteAddr())

        // Leer datos
        buf := make([]byte, 1024)
        n, err := conn.Read(buf)
        if err != nil {
            conn.Close()
            continue
        }
        fmt.Printf("Received: %s\n", buf[:n])

        // Escribir respuesta
        _, err = conn.Write([]byte("Hello from server\n"))
        if err != nil {
            conn.Close()
            continue
        }

        conn.Close()
    }
}

DETALLE DE net.Listen:
1. Crea un socket (syscall socket)
2. Setea SO_REUSEADDR automaticamente
3. Bind al address (syscall bind)
4. Empieza a escuchar (syscall listen)
5. Retorna net.Listener

DETALLE DE Accept:
1. Bloquea hasta que llega una conexion (el runtime usa epoll/kqueue)
2. Retorna net.Conn que representa la conexion aceptada
3. net.Conn tiene RemoteAddr() y LocalAddr() para identificar endpoints

net.Conn IMPLEMENTA:
- io.Reader     (Read)
- io.Writer     (Write)
- io.Closer     (Close)
- LocalAddr()   net.Addr
- RemoteAddr()  net.Addr
- SetDeadline(t time.Time) error
- SetReadDeadline(t time.Time) error
- SetWriteDeadline(t time.Time) error
`)

	// ============================================
	// 3. TCP CLIENT
	// ============================================
	fmt.Println("\n--- 3. TCP Client ---")
	os.Stdout.WriteString(`
TCP CLIENT BASICO:

func connectTCP(addr string) error {
    // Dial resuelve DNS + connect
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return fmt.Errorf("dial: %w", err)
    }
    defer conn.Close()

    // Enviar datos
    _, err = conn.Write([]byte("Hello from client\n"))
    if err != nil {
        return fmt.Errorf("write: %w", err)
    }

    // Leer respuesta
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        return fmt.Errorf("read: %w", err)
    }
    fmt.Printf("Server says: %s\n", buf[:n])

    return nil
}

CON TIMEOUT (net.DialTimeout):

conn, err := net.DialTimeout("tcp", "example.com:80", 5*time.Second)
if err != nil {
    // Timeout o error de conexion
}

CON CONTEXT (net.Dialer):

dialer := &net.Dialer{
    Timeout:   5 * time.Second,   // Timeout total de conexion
    KeepAlive: 30 * time.Second,  // Intervalo TCP keep-alive
    LocalAddr: nil,               // Bind a interfaz especifica
}

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

conn, err := dialer.DialContext(ctx, "tcp", "example.com:443")
if err != nil {
    // Context cancelado, timeout, o error de red
}

DIFERENCIAS:

  net.Dial          -> Simple, sin timeout (peligroso en produccion)
  net.DialTimeout   -> Con timeout fijo
  net.Dialer        -> Configurable: timeout, keepalive, local addr, context
  Dialer.DialContext -> Cancel-safe con context

ERRORES COMUNES:

  *net.OpError            -> Error de operacion de red
  *net.DNSError           -> Error de resolucion DNS
  *net.AddrError          -> Direccion mal formada
  syscall.ECONNREFUSED    -> Puerto cerrado / servicio no disponible
  syscall.ETIMEDOUT       -> Timeout de conexion
  io.EOF                  -> Conexion cerrada por el peer

DETECCION DE ERROR:

var dnsErr *net.DNSError
if errors.As(err, &dnsErr) {
    if dnsErr.IsNotFound {
        fmt.Println("Host no encontrado")
    }
    if dnsErr.IsTemporary {
        fmt.Println("Error DNS temporal, reintentar")
    }
}

var opErr *net.OpError
if errors.As(err, &opErr) {
    fmt.Printf("Op: %s, Net: %s, Addr: %s\n",
        opErr.Op, opErr.Net, opErr.Addr)
    if opErr.Timeout() {
        fmt.Println("Timeout de red")
    }
}
`)

	// ============================================
	// 4. CONCURRENT TCP SERVER
	// ============================================
	fmt.Println("\n--- 4. Concurrent TCP Server ---")
	os.Stdout.WriteString(`
GOROUTINE PER CONNECTION:

func runConcurrentServer(addr string) error {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("accept error: %v", err)
            continue
        }
        // Cada conexion se maneja en su propia goroutine
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    remoteAddr := conn.RemoteAddr().String()
    log.Printf("Client connected: %s", remoteAddr)

    buf := make([]byte, 4096)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            if err != io.EOF {
                log.Printf("Read error from %s: %v", remoteAddr, err)
            }
            return
        }
        // Echo back
        if _, err := conn.Write(buf[:n]); err != nil {
            log.Printf("Write error to %s: %v", remoteAddr, err)
            return
        }
    }
}

CON LIMITE DE CONEXIONES (semaforo):

func runLimitedServer(addr string, maxConns int) error {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        return err
    }
    defer listener.Close()

    sem := make(chan struct{}, maxConns)

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        sem <- struct{}{} // Bloquea si ya hay maxConns activas
        go func() {
            defer func() { <-sem }()
            handleConnection(conn)
        }()
    }
}

PATRON: SERVER CON SHUTDOWN GRACEFUL:

type Server struct {
    listener net.Listener
    quit     chan struct{}
    wg       sync.WaitGroup
}

func NewServer(addr string) (*Server, error) {
    l, err := net.Listen("tcp", addr)
    if err != nil {
        return nil, err
    }
    return &Server{
        listener: l,
        quit:     make(chan struct{}),
    }, nil
}

func (s *Server) Start() {
    s.wg.Add(1)
    go s.acceptLoop()
}

func (s *Server) acceptLoop() {
    defer s.wg.Done()
    for {
        conn, err := s.listener.Accept()
        if err != nil {
            select {
            case <-s.quit:
                return // Shutdown solicitado
            default:
                log.Printf("accept error: %v", err)
                continue
            }
        }
        s.wg.Add(1)
        go func() {
            defer s.wg.Done()
            s.handleConn(conn)
        }()
    }
}

func (s *Server) handleConn(conn net.Conn) {
    defer conn.Close()
    // ... manejo de conexion
}

func (s *Server) Stop() {
    close(s.quit)        // Senalar shutdown
    s.listener.Close()   // Desbloquear Accept
    s.wg.Wait()          // Esperar conexiones activas
}
`)

	// ============================================
	// 5. TCP WITH BUFIO
	// ============================================
	fmt.Println("\n--- 5. TCP con bufio (Line-Oriented Protocol) ---")
	os.Stdout.WriteString(`
bufio SOBRE net.Conn:

Los protocolos de texto como HTTP, SMTP, RESP (Redis) son
orientados a lineas. bufio.Scanner y bufio.Reader facilitan esto.

CON bufio.Scanner (SIMPLE):

func handleLineProtocol(conn net.Conn) {
    defer conn.Close()

    scanner := bufio.NewScanner(conn)
    // Limite: por defecto bufio.MaxScanTokenSize = 64KB
    // Para lineas mas largas:
    scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

    for scanner.Scan() {
        line := scanner.Text()
        if line == "QUIT" {
            conn.Write([]byte("BYE\n"))
            return
        }
        response := fmt.Sprintf("ECHO: %s\n", line)
        conn.Write([]byte(response))
    }

    if err := scanner.Err(); err != nil {
        log.Printf("Scanner error: %v", err)
    }
}

CON bufio.Reader (MAS CONTROL):

func handleWithBufioReader(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    for {
        // ReadString lee hasta el delimitador (inclusive)
        line, err := reader.ReadString('\n')
        if err != nil {
            return
        }

        line = strings.TrimSpace(line)

        // Escribir con buffer (mas eficiente para multiples writes)
        fmt.Fprintf(writer, "GOT: %s\r\n", line)
        writer.Flush() // IMPORTANTE: flush despues de cada respuesta
    }
}

CON bufio.ReadWriter (BIDIRECCIONAL):

func handleBidirectional(conn net.Conn) {
    defer conn.Close()

    rw := bufio.NewReadWriter(
        bufio.NewReader(conn),
        bufio.NewWriter(conn),
    )

    for {
        line, err := rw.ReadString('\n')
        if err != nil {
            return
        }

        fmt.Fprintf(rw, "REPLY: %s", line)
        rw.Flush()
    }
}

CUSTOM SCANNER (split function):

// Leer mensajes separados por \r\n\r\n (como HTTP)
func splitDoubleCRLF(data []byte, atEOF bool) (int, []byte, error) {
    if i := bytes.Index(data, []byte("\r\n\r\n")); i >= 0 {
        return i + 4, data[:i], nil
    }
    if atEOF && len(data) > 0 {
        return len(data), data, nil
    }
    return 0, nil, nil // Necesita mas datos
}

scanner := bufio.NewScanner(conn)
scanner.Split(splitDoubleCRLF)
`)

	// ============================================
	// 6. UDP SERVER / CLIENT
	// ============================================
	fmt.Println("\n--- 6. UDP Server/Client ---")
	os.Stdout.WriteString(`
UDP SERVER (net.ListenPacket):

func runUDPServer(addr string) error {
    // ListenPacket crea socket UDP
    pc, err := net.ListenPacket("udp", addr)
    if err != nil {
        return err
    }
    defer pc.Close()

    buf := make([]byte, 65535) // Max UDP payload
    for {
        // ReadFrom retorna datos + address del sender
        n, remoteAddr, err := pc.ReadFrom(buf)
        if err != nil {
            return err
        }

        msg := string(buf[:n])
        fmt.Printf("UDP from %s: %s\n", remoteAddr, msg)

        // Responder al sender
        response := fmt.Sprintf("ACK: %s", msg)
        _, err = pc.WriteTo([]byte(response), remoteAddr)
        if err != nil {
            log.Printf("WriteTo error: %v", err)
        }
    }
}

UDP CLIENT (net.DialUDP):

func sendUDP(serverAddr string, msg string) error {
    addr, err := net.ResolveUDPAddr("udp", serverAddr)
    if err != nil {
        return err
    }

    // DialUDP crea un "connected" UDP socket
    // Solo puede enviar/recibir de esta address
    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        return err
    }
    defer conn.Close()

    // Write en vez de WriteTo (ya esta "conectado")
    _, err = conn.Write([]byte(msg))
    if err != nil {
        return err
    }

    // Leer respuesta
    buf := make([]byte, 65535)
    conn.SetReadDeadline(time.Now().Add(5 * time.Second))
    n, err := conn.Read(buf)
    if err != nil {
        return err
    }
    fmt.Printf("Server response: %s\n", buf[:n])

    return nil
}

UDP CLIENT SIMPLE (net.Dial):

conn, err := net.Dial("udp", "localhost:9999")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

conn.Write([]byte("ping"))

TCP vs UDP:

  TCP                           UDP
  ---------------------------   ---------------------------
  Confiable (ACK, retransmit)   Best-effort (puede perder)
  Orientado a conexion          Sin conexion
  Ordenado                      Sin orden garantizado
  Flow control                  Sin flow control
  Overhead mayor                Overhead minimo
  HTTP, SSH, DB                 DNS, gaming, video, VoIP
`)

	// ============================================
	// 7. UNIX DOMAIN SOCKETS
	// ============================================
	fmt.Println("\n--- 7. Unix Domain Sockets ---")
	os.Stdout.WriteString(`
UNIX DOMAIN SOCKETS:

Comunicacion entre procesos en la misma maquina, sin pasar por
el stack TCP/IP. Mas rapido y eficiente que TCP localhost.

TIPOS:
- "unix"       -> Stream (como TCP, orientado a conexion)
- "unixgram"   -> Datagram (como UDP, mensajes individuales)
- "unixpacket" -> Sequenced packets (poco usado)

SERVER:

func runUnixServer(socketPath string) error {
    // Remover socket viejo si existe
    os.Remove(socketPath)

    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        return err
    }
    defer listener.Close()
    defer os.Remove(socketPath) // Cleanup

    // Setear permisos del socket file
    os.Chmod(socketPath, 0660)

    for {
        conn, err := listener.Accept()
        if err != nil {
            return err
        }
        go handleUnixConn(conn)
    }
}

CLIENT:

func connectUnix(socketPath string) error {
    conn, err := net.Dial("unix", socketPath)
    if err != nil {
        return err
    }
    defer conn.Close()

    conn.Write([]byte("hello via unix socket"))

    buf := make([]byte, 1024)
    n, _ := conn.Read(buf)
    fmt.Printf("Response: %s\n", buf[:n])
    return nil
}

USOS COMUNES:
- Docker daemon      -> /var/run/docker.sock
- MySQL              -> /var/run/mysqld/mysqld.sock
- PostgreSQL         -> /var/run/postgresql/.s.PGSQL.5432
- Nginx + PHP-FPM    -> /var/run/php-fpm.sock
- systemd journal    -> /run/systemd/journal/socket

ABSTRACT NAMESPACE (Linux):

// Prefijo @ indica abstract namespace (no crea archivo)
listener, err := net.Listen("unix", "@myapp.sock")
// No necesita cleanup, desaparece cuando el proceso termina

VENTAJAS SOBRE TCP LOCALHOST:
- ~2x mas rapido (no pasa por stack TCP)
- Seguridad via filesystem permissions
- No consume puertos
- Soporta file descriptor passing (SCM_RIGHTS)
`)

	// ============================================
	// 8. NET INTERFACES
	// ============================================
	fmt.Println("\n--- 8. net.Listener y net.Conn Interfaces ---")
	os.Stdout.WriteString(`
INTERFACES CORE DEL PAQUETE NET:

type Listener interface {
    Accept() (Conn, error)   // Espera y retorna nueva conexion
    Close() error            // Cierra el listener
    Addr() Addr              // Retorna la address del listener
}

type Conn interface {
    Read(b []byte) (n int, err error)
    Write(b []byte) (n int, err error)
    Close() error
    LocalAddr() Addr
    RemoteAddr() Addr
    SetDeadline(t time.Time) error
    SetReadDeadline(t time.Time) error
    SetWriteDeadline(t time.Time) error
}

type Addr interface {
    Network() string  // "tcp", "udp", "unix"
    String() string   // "192.168.1.1:8080"
}

type PacketConn interface {
    ReadFrom(p []byte) (n int, addr Addr, err error)
    WriteTo(p []byte, addr Addr) (n int, err error)
    Close() error
    LocalAddr() Addr
    SetDeadline(t time.Time) error
    SetReadDeadline(t time.Time) error
    SetWriteDeadline(t time.Time) error
}

POR QUE IMPORTAN:

1. Polimorfismo: Codigo que acepta net.Conn funciona con TCP,
   Unix sockets, TLS, mock connections, etc.

2. Testing: Puedes crear mock Conn para tests sin red real.

3. Wrapping: tls.Conn envuelve net.Conn transparentemente.

IMPLEMENTACIONES DE net.Conn:

  *net.TCPConn     -> Conexion TCP (tiene metodos extra)
  *net.UDPConn     -> Conexion UDP
  *net.UnixConn    -> Unix domain socket
  *net.IPConn      -> Raw IP connection
  *tls.Conn        -> TLS sobre cualquier net.Conn

METODOS EXTRA DE *net.TCPConn:

conn.(*net.TCPConn).SetKeepAlive(true)
conn.(*net.TCPConn).SetKeepAlivePeriod(30 * time.Second)
conn.(*net.TCPConn).SetNoDelay(true)   // Disable Nagle
conn.(*net.TCPConn).SetLinger(0)       // Reset on close
conn.(*net.TCPConn).CloseRead()        // Half-close read side
conn.(*net.TCPConn).CloseWrite()       // Half-close write side

METODOS EXTRA DE *net.UDPConn:

conn.(*net.UDPConn).ReadFromUDP(buf)  // Retorna *net.UDPAddr
conn.(*net.UDPConn).WriteToUDP(buf, addr)
conn.(*net.UDPConn).ReadMsgUDP(buf, oob)  // Con out-of-band data

PATRON: ACEPTAR CUALQUIER TRANSPORT:

type Handler interface {
    Handle(conn net.Conn) error
}

func Serve(l net.Listener, h Handler) error {
    for {
        conn, err := l.Accept()
        if err != nil {
            return err
        }
        go h.Handle(conn)
    }
}

// Funciona con TCP, Unix, TLS - cualquier Listener
`)

	// ============================================
	// 9. DEADLINES / TIMEOUTS
	// ============================================
	fmt.Println("\n--- 9. Deadlines y Timeouts en Connections ---")
	os.Stdout.WriteString(`
DEADLINES EN net.Conn:

Los deadlines son ABSOLUTOS (time.Time), no duraciones.
Aplican a la SIGUIENTE operacion Read/Write.

// SetDeadline aplica a Read Y Write
conn.SetDeadline(time.Now().Add(30 * time.Second))

// SetReadDeadline solo aplica a Read
conn.SetReadDeadline(time.Now().Add(10 * time.Second))

// SetWriteDeadline solo aplica a Write
conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

// Remover deadline (sin timeout)
conn.SetDeadline(time.Time{})

DETECCION DE TIMEOUT:

n, err := conn.Read(buf)
if err != nil {
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        // Timeout! La conexion sigue viva
        fmt.Println("Read timeout, conexion sigue abierta")
    } else {
        // Error real, cerrar conexion
        conn.Close()
    }
}

PATRON: IDLE TIMEOUT (server):

func handleWithIdleTimeout(conn net.Conn, idleTimeout time.Duration) {
    defer conn.Close()

    buf := make([]byte, 4096)
    for {
        // Resetear deadline antes de cada Read
        conn.SetReadDeadline(time.Now().Add(idleTimeout))
        n, err := conn.Read(buf)
        if err != nil {
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                fmt.Println("Client idle, disconnecting")
            }
            return
        }
        // Resetear deadline antes de Write
        conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
        conn.Write(buf[:n])
    }
}

PATRON: OPERATION TIMEOUT CON CONTEXT:

func readWithContext(ctx context.Context, conn net.Conn, buf []byte) (int, error) {
    // Si el context tiene deadline, aplicarlo a la conexion
    if deadline, ok := ctx.Deadline(); ok {
        conn.SetReadDeadline(deadline)
        defer conn.SetReadDeadline(time.Time{})
    }

    // Channel para el resultado
    type result struct {
        n   int
        err error
    }
    ch := make(chan result, 1)

    go func() {
        n, err := conn.Read(buf)
        ch <- result{n, err}
    }()

    select {
    case <-ctx.Done():
        conn.SetReadDeadline(time.Now()) // Forzar timeout
        <-ch // Esperar que Read retorne
        return 0, ctx.Err()
    case r := <-ch:
        return r.n, r.err
    }
}

ADVERTENCIA IMPORTANTE:

- Deadline es ABSOLUTO, no relativo. Si seteas una vez, no se
  resetea automaticamente. Debes re-setear antes de cada operacion.
- Un deadline expirado NO cierra la conexion. Solo hace que la
  siguiente Read/Write retorne error. La conexion sigue usable
  si seteas un nuevo deadline.
- El zero value time.Time{} REMUEVE el deadline.
`)

	// ============================================
	// 10. DNS LOOKUPS
	// ============================================
	fmt.Println("\n--- 10. DNS Lookups ---")
	os.Stdout.WriteString(`
LOOKUPS BASICOS:

// Resolver hostname a IPs
addrs, err := net.LookupHost("google.com")
// addrs = ["142.250.185.206", "2607:f8b0:4004:800::200e"]

// Resolver IP a hostnames (reverse DNS)
names, err := net.LookupAddr("8.8.8.8")
// names = ["dns.google."]

// Resolver a net.IP (mas control)
ips, err := net.LookupIP("google.com")
for _, ip := range ips {
    if ip.To4() != nil {
        fmt.Printf("IPv4: %s\n", ip)
    } else {
        fmt.Printf("IPv6: %s\n", ip)
    }
}

// Lookup CNAME
cname, err := net.LookupCNAME("www.google.com")
// cname = "www.google.com."

// Lookup MX records
mxs, err := net.LookupMX("google.com")
for _, mx := range mxs {
    fmt.Printf("MX: %s (priority %d)\n", mx.Host, mx.Pref)
}

// Lookup NS records
nss, err := net.LookupNS("google.com")

// Lookup TXT records
txts, err := net.LookupTXT("google.com")

// Lookup SRV records
_, addrs, err := net.LookupSRV("xmpp-server", "tcp", "google.com")

CUSTOM RESOLVER:

resolver := &net.Resolver{
    PreferGo: true, // Usar resolver Go puro (no CGO)
    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
        // Custom DNS server (e.g., Cloudflare)
        d := net.Dialer{Timeout: 5 * time.Second}
        return d.DialContext(ctx, "udp", "1.1.1.1:53")
    },
}

addrs, err := resolver.LookupHost(context.Background(), "example.com")

PreferGo VS CGO RESOLVER:

  PreferGo: true  -> Resolver DNS escrito en Go puro
                     No depende de libc, mas portable
                     Soporta /etc/hosts y /etc/resolv.conf

  PreferGo: false -> Usa CGO y libc (getaddrinfo)
                     Soporta NSS (LDAP, mDNS, etc)
                     Es el default en la mayoria de plataformas

FORCE GO RESOLVER (build time):

  GODEBUG=netdns=go   -> Forzar Go resolver
  GODEBUG=netdns=cgo  -> Forzar CGO resolver
  GODEBUG=netdns=go+2 -> Go resolver con debug logging
`)

	fmt.Println("--- Demo: DNS Lookups ---")
	demoDNSLookup()

	// ============================================
	// 11. IP ADDRESS HANDLING
	// ============================================
	fmt.Println("\n--- 11. IP Address Handling ---")
	os.Stdout.WriteString(`
net.IP:

// Parsear IP
ip := net.ParseIP("192.168.1.1")
if ip == nil {
    fmt.Println("IP invalida")
}

// IPv4 o IPv6?
if ip.To4() != nil {
    fmt.Println("IPv4")
} else {
    fmt.Println("IPv6")
}

// IPs especiales
ip.IsLoopback()        // 127.0.0.0/8 o ::1
ip.IsPrivate()         // 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
ip.IsMulticast()       // 224.0.0.0/4 o ff00::/8
ip.IsGlobalUnicast()   // IP publica routable
ip.IsUnspecified()     // 0.0.0.0 o ::
ip.IsLinkLocalUnicast()   // 169.254.0.0/16 o fe80::/10
ip.IsLinkLocalMulticast() // 224.0.0.0/24 o ff02::/16

// Comparar IPs
ip1 := net.ParseIP("192.168.1.1")
ip2 := net.ParseIP("192.168.1.1")
fmt.Println(ip1.Equal(ip2)) // true

net.IPNet (CIDR):

// Parsear CIDR
_, network, err := net.ParseCIDR("192.168.1.0/24")
// network.IP = 192.168.1.0
// network.Mask = ffffff00 (255.255.255.0)

// Verificar si IP esta en la red
ip := net.ParseIP("192.168.1.42")
fmt.Println(network.Contains(ip)) // true

ip2 := net.ParseIP("10.0.0.1")
fmt.Println(network.Contains(ip2)) // false

// Crear IPNet manualmente
network := &net.IPNet{
    IP:   net.IPv4(10, 0, 0, 0),
    Mask: net.CIDRMask(8, 32), // /8
}

// Tamano del mask
ones, bits := network.Mask.Size()
fmt.Printf("/%d (de %d bits)\n", ones, bits)

CIDR PARSING:

func parseCIDRList(cidrs []string) ([]*net.IPNet, error) {
    networks := make([]*net.IPNet, 0, len(cidrs))
    for _, cidr := range cidrs {
        _, network, err := net.ParseCIDR(cidr)
        if err != nil {
            return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
        }
        networks = append(networks, network)
    }
    return networks, nil
}

func isAllowed(ip net.IP, allowList []*net.IPNet) bool {
    for _, network := range allowList {
        if network.Contains(ip) {
            return true
        }
    }
    return false
}
`)

	fmt.Println("--- Demo: IP Address Handling ---")
	demoIPHandling()

	// ============================================
	// 12. CONNECTION POOLING
	// ============================================
	fmt.Println("\n--- 12. Connection Pooling Patterns ---")
	os.Stdout.WriteString(`
CONNECTION POOL:

Go no tiene un connection pool generico en la stdlib.
Debes implementarlo o usar database/sql (que si tiene pool).

POOL BASICO CON CHANNEL:

type ConnPool struct {
    mu       sync.Mutex
    conns    chan net.Conn
    factory  func() (net.Conn, error)
    maxConns int
    active   int
}

func NewConnPool(maxConns int, factory func() (net.Conn, error)) *ConnPool {
    return &ConnPool{
        conns:    make(chan net.Conn, maxConns),
        factory:  factory,
        maxConns: maxConns,
    }
}

func (p *ConnPool) Get(ctx context.Context) (net.Conn, error) {
    // Intentar obtener una conexion existente
    select {
    case conn := <-p.conns:
        if isConnAlive(conn) {
            return conn, nil
        }
        conn.Close()
        p.mu.Lock()
        p.active--
        p.mu.Unlock()
    default:
    }

    // Crear nueva si no excede el maximo
    p.mu.Lock()
    if p.active >= p.maxConns {
        p.mu.Unlock()
        // Esperar a que se libere una conexion
        select {
        case conn := <-p.conns:
            return conn, nil
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    p.active++
    p.mu.Unlock()

    conn, err := p.factory()
    if err != nil {
        p.mu.Lock()
        p.active--
        p.mu.Unlock()
        return nil, err
    }
    return conn, nil
}

func (p *ConnPool) Put(conn net.Conn) {
    select {
    case p.conns <- conn:
        // Devuelta al pool
    default:
        // Pool lleno, cerrar
        conn.Close()
        p.mu.Lock()
        p.active--
        p.mu.Unlock()
    }
}

func (p *ConnPool) Close() {
    close(p.conns)
    for conn := range p.conns {
        conn.Close()
    }
}

func isConnAlive(conn net.Conn) bool {
    conn.SetReadDeadline(time.Now())
    buf := make([]byte, 1)
    _, err := conn.Read(buf)
    conn.SetReadDeadline(time.Time{})
    if err == nil {
        return true
    }
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        return true // Timeout = no data, pero conexion viva
    }
    return false // Conexion muerta
}

USO:

pool := NewConnPool(10, func() (net.Conn, error) {
    return net.DialTimeout("tcp", "redis:6379", 5*time.Second)
})
defer pool.Close()

conn, err := pool.Get(ctx)
if err != nil {
    log.Fatal(err)
}
defer pool.Put(conn) // Devolver al pool, no Close()

POOL CON HEALTH CHECK:

type HealthCheckPool struct {
    *ConnPool
    checkInterval time.Duration
}

func (p *HealthCheckPool) StartHealthCheck() {
    go func() {
        ticker := time.NewTicker(p.checkInterval)
        defer ticker.Stop()
        for range ticker.C {
            p.evictBroken()
        }
    }()
}
`)

	// ============================================
	// 13. PROTOCOL IMPLEMENTATION
	// ============================================
	fmt.Println("\n--- 13. Protocol Implementation (Simple Redis-like) ---")
	os.Stdout.WriteString(`
PROTOCOLO CUSTOM TIPO REDIS (RESP-like):

Formato:
  Comando:   COMMAND arg1 arg2\r\n
  Respuesta: +OK\r\n          (simple string)
             -ERR message\r\n  (error)
             :42\r\n           (integer)
             $5\r\nHello\r\n  (bulk string, prefixed con largo)

IMPLEMENTACION SERVER:

type KVStore struct {
    mu   sync.RWMutex
    data map[string]string
}

func NewKVStore() *KVStore {
    return &KVStore{data: make(map[string]string)}
}

func (kv *KVStore) HandleConn(conn net.Conn) {
    defer conn.Close()
    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    for {
        conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        line, err := reader.ReadString('\n')
        if err != nil {
            return
        }

        line = strings.TrimSpace(line)
        parts := strings.Fields(line)
        if len(parts) == 0 {
            continue
        }

        cmd := strings.ToUpper(parts[0])
        args := parts[1:]

        var response string
        switch cmd {
        case "SET":
            if len(args) != 2 {
                response = "-ERR wrong number of arguments\r\n"
            } else {
                kv.mu.Lock()
                kv.data[args[0]] = args[1]
                kv.mu.Unlock()
                response = "+OK\r\n"
            }
        case "GET":
            if len(args) != 1 {
                response = "-ERR wrong number of arguments\r\n"
            } else {
                kv.mu.RLock()
                val, ok := kv.data[args[0]]
                kv.mu.RUnlock()
                if !ok {
                    response = "$-1\r\n" // nil
                } else {
                    response = fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
                }
            }
        case "DEL":
            if len(args) != 1 {
                response = "-ERR wrong number of arguments\r\n"
            } else {
                kv.mu.Lock()
                _, existed := kv.data[args[0]]
                delete(kv.data, args[0])
                kv.mu.Unlock()
                if existed {
                    response = ":1\r\n"
                } else {
                    response = ":0\r\n"
                }
            }
        case "PING":
            response = "+PONG\r\n"
        case "QUIT":
            writer.WriteString("+BYE\r\n")
            writer.Flush()
            return
        default:
            response = fmt.Sprintf("-ERR unknown command '%s'\r\n", cmd)
        }

        writer.WriteString(response)
        writer.Flush()
    }
}

IMPLEMENTACION CLIENT:

type KVClient struct {
    conn   net.Conn
    reader *bufio.Reader
}

func NewKVClient(addr string) (*KVClient, error) {
    conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
    if err != nil {
        return nil, err
    }
    return &KVClient{
        conn:   conn,
        reader: bufio.NewReader(conn),
    }, nil
}

func (c *KVClient) Do(cmd string) (string, error) {
    _, err := fmt.Fprintf(c.conn, "%s\r\n", cmd)
    if err != nil {
        return "", err
    }

    line, err := c.reader.ReadString('\n')
    if err != nil {
        return "", err
    }
    line = strings.TrimRight(line, "\r\n")

    switch {
    case strings.HasPrefix(line, "+"):
        return line[1:], nil
    case strings.HasPrefix(line, "-"):
        return "", fmt.Errorf("server error: %s", line[1:])
    case strings.HasPrefix(line, ":"):
        return line[1:], nil
    case strings.HasPrefix(line, "$"):
        // Bulk string
        if line == "$-1" {
            return "(nil)", nil
        }
        data, _ := c.reader.ReadString('\n')
        return strings.TrimRight(data, "\r\n"), nil
    }
    return line, nil
}

func (c *KVClient) Close() error {
    c.Do("QUIT")
    return c.conn.Close()
}
`)

	// ============================================
	// 14. TLS CONNECTIONS
	// ============================================
	fmt.Println("\n--- 14. TLS Connections (crypto/tls) ---")
	os.Stdout.WriteString(`
TLS SERVER:

func runTLSServer(addr, certFile, keyFile string) error {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return fmt.Errorf("load cert: %w", err)
    }

    config := &tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion:   tls.VersionTLS12,
        // CipherSuites se seleccionan automaticamente en Go 1.17+
    }

    listener, err := tls.Listen("tcp", addr, config)
    if err != nil {
        return fmt.Errorf("listen: %w", err)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go handleTLSConn(conn)
    }
}

func handleTLSConn(conn net.Conn) {
    defer conn.Close()

    // Type assertion para acceder a info TLS
    tlsConn, ok := conn.(*tls.Conn)
    if !ok {
        return
    }

    // El handshake ocurre automaticamente en el primer Read/Write
    // Pero podemos forzarlo para verificar antes:
    if err := tlsConn.Handshake(); err != nil {
        log.Printf("TLS handshake error: %v", err)
        return
    }

    state := tlsConn.ConnectionState()
    fmt.Printf("TLS version: %d\n", state.Version)
    fmt.Printf("Cipher suite: %s\n",
        tls.CipherSuiteName(state.CipherSuite))
    fmt.Printf("Server name: %s\n", state.ServerName)

    // Ahora leer/escribir normalmente
    buf := make([]byte, 4096)
    n, _ := conn.Read(buf)
    conn.Write(buf[:n])
}

TLS CLIENT:

func connectTLS(addr string) error {
    config := &tls.Config{
        MinVersion: tls.VersionTLS12,
        // ServerName se infiere del addr si es hostname
        // Para IP o mismatch:
        // ServerName: "expected.hostname.com",

        // InsecureSkipVerify: true, // NUNCA en produccion
    }

    conn, err := tls.Dial("tcp", addr, config)
    if err != nil {
        return fmt.Errorf("dial: %w", err)
    }
    defer conn.Close()

    // Verificar el certificado del server
    state := conn.ConnectionState()
    for _, cert := range state.PeerCertificates {
        fmt.Printf("Subject: %s\n", cert.Subject)
        fmt.Printf("Issuer: %s\n", cert.Issuer)
        fmt.Printf("Valid: %s - %s\n", cert.NotBefore, cert.NotAfter)
    }

    conn.Write([]byte("Hello TLS"))
    return nil
}

TLS CON DIALER (recomendado):

dialer := &tls.Dialer{
    NetDialer: &net.Dialer{
        Timeout:   5 * time.Second,
        KeepAlive: 30 * time.Second,
    },
    Config: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
}

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

conn, err := dialer.DialContext(ctx, "tcp", "example.com:443")

UPGRADE TCP A TLS (STARTTLS pattern):

// Primero conectar TCP normal
tcpConn, _ := net.Dial("tcp", "mail.example.com:587")

// Negociar STARTTLS...
// Luego upgrade a TLS:
tlsConn := tls.Client(tcpConn, &tls.Config{
    ServerName: "mail.example.com",
})
err := tlsConn.Handshake()

tls.Config OPCIONES IMPORTANTES:

config := &tls.Config{
    MinVersion: tls.VersionTLS12,
    MaxVersion: tls.VersionTLS13,

    // Curvas ellipticas preferidas
    CurvePreferences: []tls.CurveID{
        tls.X25519, tls.CurveP256,
    },

    // Solo para TLS 1.2 (TLS 1.3 las selecciona automaticamente)
    CipherSuites: []uint16{
        tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
        tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
        tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
    },

    // ALPN (Application-Layer Protocol Negotiation)
    NextProtos: []string{"h2", "http/1.1"},
}
`)

	// ============================================
	// 15. MUTUAL TLS (mTLS)
	// ============================================
	fmt.Println("\n--- 15. Mutual TLS (mTLS) ---")
	os.Stdout.WriteString(`
mTLS: Tanto el server como el client presentan certificados.
El server verifica la identidad del client y viceversa.

ESCENARIOS:
- Service mesh (Istio, Linkerd)
- Microservicios internos
- API B2B
- Zero-trust networks

mTLS SERVER:

func runMTLSServer(addr, certFile, keyFile, caFile string) error {
    // Cargar certificado del server
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return err
    }

    // Cargar CA que firmo los certificados de los clients
    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return err
    }
    caPool := x509.NewCertPool()
    caPool.AppendCertsFromPEM(caCert)

    config := &tls.Config{
        Certificates: []tls.Certificate{cert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    caPool,
        MinVersion:   tls.VersionTLS12,
    }

    listener, err := tls.Listen("tcp", addr, config)
    if err != nil {
        return err
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go handleMTLSConn(conn)
    }
}

func handleMTLSConn(conn net.Conn) {
    defer conn.Close()

    tlsConn := conn.(*tls.Conn)
    if err := tlsConn.Handshake(); err != nil {
        log.Printf("mTLS handshake failed: %v", err)
        return
    }

    state := tlsConn.ConnectionState()
    if len(state.PeerCertificates) > 0 {
        clientCert := state.PeerCertificates[0]
        fmt.Printf("Client: %s\n", clientCert.Subject.CommonName)
        fmt.Printf("Org: %v\n", clientCert.Subject.Organization)
    }

    // Continuar con la comunicacion...
}

mTLS CLIENT:

func connectMTLS(addr, certFile, keyFile, caFile string) error {
    // Certificado del client
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return err
    }

    // CA del server
    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return err
    }
    caPool := x509.NewCertPool()
    caPool.AppendCertsFromPEM(caCert)

    config := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caPool,
        MinVersion:   tls.VersionTLS12,
    }

    conn, err := tls.Dial("tcp", addr, config)
    if err != nil {
        return fmt.Errorf("mTLS dial: %w", err)
    }
    defer conn.Close()

    conn.Write([]byte("Hello from authenticated client"))
    return nil
}

NIVELES DE ClientAuth:

  tls.NoClientCert               -> No pide cert al client
  tls.RequestClientCert          -> Pide pero no requiere
  tls.RequireAnyClientCert       -> Requiere cert (no verifica CA)
  tls.VerifyClientCertIfGiven    -> Verifica si lo da (opcional)
  tls.RequireAndVerifyClientCert -> Requiere + verifica CA (mTLS)

AUTORIZACION BASADA EN CERT:

func authorizeClient(state tls.ConnectionState) error {
    if len(state.PeerCertificates) == 0 {
        return errors.New("no client certificate")
    }

    cert := state.PeerCertificates[0]

    // Verificar Organization
    allowedOrgs := map[string]bool{"MyCompany": true}
    for _, org := range cert.Subject.Organization {
        if allowedOrgs[org] {
            return nil
        }
    }

    // Verificar Common Name
    allowedCNs := map[string]bool{
        "service-a": true,
        "service-b": true,
    }
    if allowedCNs[cert.Subject.CommonName] {
        return nil
    }

    return fmt.Errorf("unauthorized client: %s", cert.Subject.CommonName)
}
`)

	// ============================================
	// 16. CERTIFICATE GENERATION FOR DEV
	// ============================================
	fmt.Println("\n--- 16. Certificate Generation for Development ---")
	os.Stdout.WriteString(`
GENERAR CERTIFICADOS EN GO (sin openssl):

func generateSelfSignedCert() (tls.Certificate, error) {
    // Generar key pair
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    if err != nil {
        return tls.Certificate{}, err
    }

    // Template del certificado
    template := &x509.Certificate{
        SerialNumber: big.NewInt(1),
        Subject: pkix.Name{
            Organization: []string{"Dev Corp"},
            CommonName:   "localhost",
        },
        NotBefore: time.Now(),
        NotAfter:  time.Now().Add(365 * 24 * time.Hour),

        KeyUsage:    x509.KeyUsageDigitalSignature,
        ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

        // SANs (Subject Alternative Names)
        DNSNames:    []string{"localhost", "*.localhost"},
        IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
    }

    // Self-sign
    certDER, err := x509.CreateCertificate(
        rand.Reader, template, template, &privateKey.PublicKey, privateKey,
    )
    if err != nil {
        return tls.Certificate{}, err
    }

    // Encode to PEM
    certPEM := pem.EncodeToMemory(&pem.Block{
        Type: "CERTIFICATE", Bytes: certDER,
    })

    keyDER, _ := x509.MarshalECPrivateKey(privateKey)
    keyPEM := pem.EncodeToMemory(&pem.Block{
        Type: "EC PRIVATE KEY", Bytes: keyDER,
    })

    return tls.X509KeyPair(certPEM, keyPEM)
}

GENERAR CA + SERVER CERT + CLIENT CERT:

func generateCA() (*x509.Certificate, *ecdsa.PrivateKey, error) {
    key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

    template := &x509.Certificate{
        SerialNumber: big.NewInt(1),
        Subject:      pkix.Name{CommonName: "Dev CA"},
        NotBefore:    time.Now(),
        NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
        IsCA:         true,
        KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
        BasicConstraintsValid: true,
        MaxPathLen:   1,
    }

    certDER, _ := x509.CreateCertificate(
        rand.Reader, template, template, &key.PublicKey, key,
    )
    cert, _ := x509.ParseCertificate(certDER)
    return cert, key, nil
}

func generateSignedCert(
    ca *x509.Certificate, caKey *ecdsa.PrivateKey,
    cn string, isServer bool,
) (tls.Certificate, []byte, error) {
    key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

    usage := x509.ExtKeyUsageClientAuth
    if isServer {
        usage = x509.ExtKeyUsageServerAuth
    }

    template := &x509.Certificate{
        SerialNumber: big.NewInt(time.Now().UnixNano()),
        Subject:      pkix.Name{CommonName: cn},
        NotBefore:    time.Now(),
        NotAfter:     time.Now().Add(365 * 24 * time.Hour),
        KeyUsage:     x509.KeyUsageDigitalSignature,
        ExtKeyUsage:  []x509.ExtKeyUsage{usage},
        DNSNames:     []string{"localhost"},
        IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
    }

    certDER, _ := x509.CreateCertificate(
        rand.Reader, template, ca, &key.PublicKey, caKey,
    )

    certPEM := pem.EncodeToMemory(&pem.Block{
        Type: "CERTIFICATE", Bytes: certDER,
    })
    keyDER, _ := x509.MarshalECPrivateKey(key)
    keyPEM := pem.EncodeToMemory(&pem.Block{
        Type: "EC PRIVATE KEY", Bytes: keyDER,
    })

    tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
    return tlsCert, certPEM, err
}

CON OPENSSL (alternativa CLI):

# Generar CA
openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 \
    -keyout ca-key.pem -out ca-cert.pem -days 3650 -nodes \
    -subj "/CN=Dev CA"

# Generar server cert
openssl req -newkey ec -pkeyopt ec_paramgen_curve:P-256 \
    -keyout server-key.pem -out server.csr -nodes \
    -subj "/CN=localhost"
openssl x509 -req -in server.csr -CA ca-cert.pem -CAkey ca-key.pem \
    -CAcreateserial -out server-cert.pem -days 365

# Generar client cert
openssl req -newkey ec -pkeyopt ec_paramgen_curve:P-256 \
    -keyout client-key.pem -out client.csr -nodes \
    -subj "/CN=my-client"
openssl x509 -req -in client.csr -CA ca-cert.pem -CAkey ca-key.pem \
    -CAcreateserial -out client-cert.pem -days 365

HERRAMIENTA RECOMENDADA: mkcert

# Instalar
brew install mkcert  # macOS
mkcert -install      # Instala CA local

# Generar certificados
mkcert localhost 127.0.0.1 ::1
# Crea: localhost+2.pem y localhost+2-key.pem
`)

	// ============================================
	// 17. NET/NETIP PACKAGE
	// ============================================
	fmt.Println("\n--- 17. net/netip Package (Go 1.18+) ---")
	fmt.Println(`
net/netip es la version moderna de manejo de IPs.
Introducido en Go 1.18, reemplaza net.IP en muchos casos.

VENTAJAS SOBRE net.IP:
- Comparable (se puede usar como map key)
- Inmutable (value type, no slice)
- Mas eficiente en memoria (no heap allocation para IPv4)
- Distingue IPv4 de IPv6 nativamente
- Incluye AddrPort (IP + port) y Prefix (CIDR)

TIPOS PRINCIPALES:

  netip.Addr      -> Una IP address (IPv4 o IPv6)
  netip.AddrPort  -> IP address + port
  netip.Prefix    -> IP network (CIDR)

Comparacion:

  net.IP       -> []byte (slice, not comparable, puede ser nil)
  netip.Addr   -> struct (value type, comparable, zero value valido)`)

	fmt.Println("\n--- Demo: netip ---")
	demoNetip()

	// ============================================
	// 18. KEEP-ALIVE CONFIGURATION
	// ============================================
	fmt.Println("\n--- 18. Keep-Alive Configuration ---")
	os.Stdout.WriteString(`
TCP KEEP-ALIVE:

Mecanismo para detectar conexiones muertas (peer crasheado,
red cortada, NAT timeout).

Sin keep-alive, una conexion TCP puede parecer "viva" indefinidamente
aunque el peer ya no exista.

COMO FUNCIONA:
1. Despues de inactividad (idle time), enviar probe vacio
2. Si no hay ACK, reintentar N veces cada X segundos
3. Si no hay respuesta, declarar conexion muerta

EN GO (net.TCPConn):

conn, _ := net.Dial("tcp", "server:8080")
tcpConn := conn.(*net.TCPConn)

// Habilitar keep-alive
tcpConn.SetKeepAlive(true)

// Periodo entre probes (despues de idle)
tcpConn.SetKeepAlivePeriod(30 * time.Second)

EN GO 1.24+ (KeepAliveConfig):

// Go 1.24 agrega control granular:
dialer := &net.Dialer{
    KeepAliveConfig: net.KeepAliveConfig{
        Enable:   true,
        Idle:     15 * time.Second, // Idle antes del primer probe
        Interval: 5 * time.Second,  // Intervalo entre probes
        Count:    3,                 // Numero de probes fallidos
    },
}

conn, err := dialer.Dial("tcp", "server:8080")

EN net.Dialer (antes de 1.24):

dialer := &net.Dialer{
    KeepAlive: 30 * time.Second, // Habilita y configura periodo
    // KeepAlive < 0 -> Deshabilitado
    // KeepAlive == 0 -> Default (~15s)
}

VALORES TIPICOS EN PRODUCCION:

  Web server:          15-30s idle, 5s interval, 3 probes
  Database connection: 30-60s idle, 10s interval, 5 probes
  Long-lived stream:   60s idle, 15s interval, 5 probes
  Behind NAT/LB:       < NAT timeout (tipicamente < 300s)

AWS/GCP LOAD BALANCER TIMEOUTS:
- AWS ALB idle timeout: 60s (default)
- AWS NLB: 350s
- GCP: 600s
- Keep-alive period debe ser MENOR que estos timeouts

CONFIGURACION A NIVEL OS:

Linux:
  /proc/sys/net/ipv4/tcp_keepalive_time   = 7200  (idle, 2 horas!)
  /proc/sys/net/ipv4/tcp_keepalive_intvl  = 75    (interval)
  /proc/sys/net/ipv4/tcp_keepalive_probes = 9     (count)

macOS:
  sysctl net.inet.tcp.keepidle    = 7200000 (ms)
  sysctl net.inet.tcp.keepintvl   = 75000   (ms)
  sysctl net.inet.tcp.keepcnt     = 8

Go OVERRIDES los defaults del OS con sus propios valores
cuando usas net.Dialer con KeepAlive > 0.
`)

	// ============================================
	// 19. SO_REUSEPORT PATTERNS
	// ============================================
	fmt.Println("\n--- 19. SO_REUSEPORT Patterns ---")
	os.Stdout.WriteString(`
SO_REUSEADDR vs SO_REUSEPORT:

SO_REUSEADDR (default en Go):
- Permite bind a un address en TIME_WAIT
- Go lo setea automaticamente en net.Listen
- Resuelve "address already in use" despues de restart

SO_REUSEPORT:
- Permite MULTIPLES procesos bind al MISMO puerto
- El kernel distribuye conexiones entre ellos
- Util para zero-downtime deploys y multi-process servers
- NO seteado por default en Go

ACTIVAR SO_REUSEPORT:

import "golang.org/x/sys/unix"

config := &net.ListenConfig{
    Control: func(network, address string, c syscall.RawConn) error {
        var opErr error
        err := c.Control(func(fd uintptr) {
            opErr = unix.SetsockoptInt(
                int(fd),
                unix.SOL_SOCKET,
                unix.SO_REUSEPORT,
                1,
            )
        })
        if err != nil {
            return err
        }
        return opErr
    },
}

listener, err := config.Listen(context.Background(), "tcp", ":8080")

ZERO-DOWNTIME DEPLOY PATTERN:

1. Nuevo proceso arranca con SO_REUSEPORT en el mismo puerto
2. Ambos procesos aceptan conexiones (kernel balancea)
3. Viejo proceso deja de aceptar nuevas conexiones
4. Viejo proceso espera que conexiones activas terminen
5. Viejo proceso se detiene

// Proceso nuevo
listener, _ := reusePortListen(":8080")
go acceptLoop(listener)

// Senalar al viejo proceso que drene
signal.Notify(sigCh, syscall.SIGTERM)

ALTERNATIVA: SOCKET PASSING (systemd / supervisord)

En Linux, el parent process puede pasar el file descriptor
del socket al child process:

// Parent: pasar FD
cmd := exec.Command("./new-binary")
cmd.ExtraFiles = []*os.File{listenerFile}
cmd.Start()

// Child: heredar FD
f := os.NewFile(3, "listener") // FD 3 (stdin=0, stdout=1, stderr=2)
listener, err := net.FileListener(f)

LIBRERIA: go-reuseport

import "github.com/libp2p/go-reuseport"

listener, err := reuseport.Listen("tcp", ":8080")
`)

	// ============================================
	// 20. GRACEFUL CONNECTION DRAINING
	// ============================================
	fmt.Println("\n--- 20. Graceful Connection Draining ---")
	os.Stdout.WriteString(`
CONCEPTO:

Cuando un servidor necesita apagarse (deploy, scaling down),
debe:
1. Dejar de aceptar NUEVAS conexiones
2. Esperar que conexiones ACTIVAS terminen (o timeout)
3. Cerrar todo y salir

PATRON COMPLETO:

type GracefulServer struct {
    listener  net.Listener
    quit      chan struct{}
    wg        sync.WaitGroup
    mu        sync.Mutex
    conns     map[net.Conn]struct{}
    drainTime time.Duration
}

func NewGracefulServer(addr string, drain time.Duration) (*GracefulServer, error) {
    l, err := net.Listen("tcp", addr)
    if err != nil {
        return nil, err
    }
    return &GracefulServer{
        listener:  l,
        quit:      make(chan struct{}),
        conns:     make(map[net.Conn]struct{}),
        drainTime: drain,
    }, nil
}

func (s *GracefulServer) Serve() error {
    for {
        conn, err := s.listener.Accept()
        if err != nil {
            select {
            case <-s.quit:
                return nil // Shutdown
            default:
                log.Printf("accept: %v", err)
                continue
            }
        }

        s.trackConn(conn, true)
        s.wg.Add(1)
        go func() {
            defer s.wg.Done()
            defer s.trackConn(conn, false)
            s.handleConn(conn)
        }()
    }
}

func (s *GracefulServer) trackConn(conn net.Conn, add bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if add {
        s.conns[conn] = struct{}{}
    } else {
        delete(s.conns, conn)
    }
}

func (s *GracefulServer) handleConn(conn net.Conn) {
    defer conn.Close()
    // ... logica de la conexion
}

func (s *GracefulServer) Shutdown() {
    // 1. Senalar que estamos cerrando
    close(s.quit)

    // 2. Dejar de aceptar nuevas conexiones
    s.listener.Close()

    // 3. Esperar con timeout
    done := make(chan struct{})
    go func() {
        s.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        fmt.Println("All connections drained gracefully")
    case <-time.After(s.drainTime):
        fmt.Println("Drain timeout, force closing remaining connections")
        s.mu.Lock()
        for conn := range s.conns {
            conn.Close()
        }
        s.mu.Unlock()
    }
}

USO CON SIGNALS:

func main() {
    server, _ := NewGracefulServer(":8080", 30*time.Second)

    go server.Serve()

    // Esperar signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
    <-sigCh

    fmt.Println("Shutting down...")
    server.Shutdown()
    fmt.Println("Done")
}

DRAIN CON HEALTH CHECK:

El load balancer necesita saber que el server esta drenando:

type ServerHealth int32

const (
    Healthy  ServerHealth = iota
    Draining
    Stopped
)

func (s *GracefulServer) Shutdown() {
    // Primero marcar como draining (health check retorna 503)
    atomic.StoreInt32(&s.health, int32(Draining))

    // Esperar que el LB deje de enviar trafico
    time.Sleep(5 * time.Second)

    // Ahora cerrar listener
    s.listener.Close()

    // Esperar conexiones activas
    s.wg.Wait()

    atomic.StoreInt32(&s.health, int32(Stopped))
}

// Health endpoint (para readiness probe K8s)
func (s *GracefulServer) HealthHandler(w http.ResponseWriter, r *http.Request) {
    health := ServerHealth(atomic.LoadInt32(&s.health))
    switch health {
    case Healthy:
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    case Draining:
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("draining"))
    case Stopped:
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("stopped"))
    }
}
`)

	// ============================================
	// WORKING DEMO: TCP ECHO SERVER
	// ============================================
	fmt.Println("\n--- DEMO: TCP Echo Server ---")
	demoTCPEchoServer()

	// ============================================
	// DEMO: TLS IN-MEMORY
	// ============================================
	fmt.Println("\n--- DEMO: TLS Server/Client (in-memory certs) ---")
	demoTLSConnection()

	// ============================================
	// RESUMEN
	// ============================================
	fmt.Println(`
=== RESUMEN CAPITULO 66: LOW-LEVEL NETWORK PROGRAMMING ===

NET PACKAGE:
- net.Listen      -> TCP/Unix server (Accept loop)
- net.Dial        -> TCP/UDP/Unix client
- net.ListenPacket-> UDP server (ReadFrom/WriteTo)
- net.Conn        -> Interfaz universal (Read/Write/Close/Deadlines)
- net.Listener    -> Interfaz de server (Accept/Close/Addr)

TCP:
- Server: Listen + Accept loop + goroutine per connection
- Client: Dial/DialTimeout/Dialer.DialContext
- Bufio: Scanner o Reader para protocolos de texto
- Concurrent: Semaphore para limitar conexiones

UDP:
- Server: ListenPacket + ReadFrom/WriteTo loop
- Client: DialUDP o Dial("udp", addr)
- Sin conexion, sin garantia de entrega

UNIX SOCKETS:
- Stream ("unix") o Datagram ("unixgram")
- Mas rapido que TCP localhost
- Seguridad via filesystem permissions

DEADLINES:
- SetDeadline/SetReadDeadline/SetWriteDeadline
- Absolutos (time.Time), no relativos
- Timeout no cierra conexion, solo falla la operacion

DNS:
- LookupHost, LookupAddr, LookupIP
- LookupMX, LookupNS, LookupTXT, LookupSRV
- Custom Resolver con PreferGo y Dial custom

IP HANDLING:
- net.IP + net.IPNet para IPv4/IPv6 y CIDR
- netip.Addr + netip.AddrPort + netip.Prefix (Go 1.18+)
- netip es comparable, inmutable y mas eficiente

CONNECTION POOLING:
- Channel-based pool con factory function
- Health check con read deadline trick
- database/sql tiene pool integrado

TLS:
- tls.Listen / tls.Dial para server/client
- tls.Config para cipher suites, versiones, ALPN
- mTLS: RequireAndVerifyClientCert + ClientCAs
- Generar certs en Go con crypto/x509

KEEP-ALIVE:
- Detectar conexiones muertas (peer crash, NAT timeout)
- SetKeepAlive + SetKeepAlivePeriod
- Go 1.24+: KeepAliveConfig con Idle/Interval/Count

SO_REUSEPORT:
- Multiples procesos en el mismo puerto
- Zero-downtime deploys
- ListenConfig.Control para setsockopt

GRACEFUL SHUTDOWN:
- Close listener (stop accepting)
- Wait for active connections (with timeout)
- Force close remaining after drain timeout
- Health check: Healthy -> Draining -> Stopped`)
}

// ============================================
// DEMO: DNS LOOKUPS
// ============================================

func demoDNSLookup() {
	addrs, err := net.LookupHost("localhost")
	if err != nil {
		fmt.Printf("  LookupHost error: %v\n", err)
	} else {
		fmt.Printf("  LookupHost(\"localhost\"): %v\n", addrs)
	}

	names, err := net.LookupAddr("127.0.0.1")
	if err != nil {
		fmt.Printf("  LookupAddr error: %v\n", err)
	} else {
		fmt.Printf("  LookupAddr(\"127.0.0.1\"): %v\n", names)
	}

	resolver := &net.Resolver{PreferGo: true}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ips, err := resolver.LookupHost(ctx, "localhost")
	if err != nil {
		fmt.Printf("  Custom Resolver error: %v\n", err)
	} else {
		fmt.Printf("  Custom Resolver(\"localhost\"): %v\n", ips)
	}
}

// ============================================
// DEMO: IP ADDRESS HANDLING
// ============================================

func demoIPHandling() {
	ip := net.ParseIP("192.168.1.100")
	fmt.Printf("  IP: %s\n", ip)
	fmt.Printf("  IsPrivate: %v\n", ip.IsPrivate())
	fmt.Printf("  IsLoopback: %v\n", ip.IsLoopback())
	fmt.Printf("  IsGlobalUnicast: %v\n", ip.IsGlobalUnicast())

	_, network, _ := net.ParseCIDR("10.0.0.0/8")
	fmt.Printf("  CIDR 10.0.0.0/8 contains 10.1.2.3: %v\n", network.Contains(net.ParseIP("10.1.2.3")))
	fmt.Printf("  CIDR 10.0.0.0/8 contains 192.168.1.1: %v\n", network.Contains(net.ParseIP("192.168.1.1")))

	ones, bits := network.Mask.Size()
	fmt.Printf("  Mask: /%d (of %d bits)\n", ones, bits)
}

// ============================================
// DEMO: NETIP
// ============================================

func demoNetip() {
	addr, err := netip.ParseAddr("192.168.1.1")
	if err != nil {
		fmt.Printf("  ParseAddr error: %v\n", err)
		return
	}
	fmt.Printf("  Addr: %s\n", addr)
	fmt.Printf("  Is4: %v, Is6: %v\n", addr.Is4(), addr.Is6())
	fmt.Printf("  IsPrivate: %v\n", addr.IsPrivate())
	fmt.Printf("  IsLoopback: %v\n", addr.IsLoopback())

	ap := netip.AddrPortFrom(addr, 8080)
	fmt.Printf("  AddrPort: %s\n", ap)
	fmt.Printf("  Addr: %s, Port: %d\n", ap.Addr(), ap.Port())

	prefix, _ := netip.ParsePrefix("10.0.0.0/8")
	fmt.Printf("  Prefix: %s\n", prefix)

	testAddr, _ := netip.ParseAddr("10.1.2.3")
	fmt.Printf("  10.0.0.0/8 contains 10.1.2.3: %v\n", prefix.Contains(testAddr))

	testAddr2, _ := netip.ParseAddr("192.168.1.1")
	fmt.Printf("  10.0.0.0/8 contains 192.168.1.1: %v\n", prefix.Contains(testAddr2))

	addr1, _ := netip.ParseAddr("192.168.1.1")
	addr2, _ := netip.ParseAddr("192.168.1.1")
	fmt.Printf("  Comparable: %s == %s -> %v\n", addr1, addr2, addr1 == addr2)

	m := map[netip.Addr]string{
		addr1: "my-server",
	}
	fmt.Printf("  Map lookup: %s -> %q\n", addr2, m[addr2])
}

// ============================================
// DEMO: TCP ECHO SERVER
// ============================================

func demoTCPEchoServer() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Printf("  Listen error: %v\n", err)
		return
	}
	addr := listener.Addr().String()
	fmt.Printf("  Echo server listening on %s\n", addr)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(c net.Conn) {
				defer wg.Done()
				defer c.Close()
				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					line := scanner.Text()
					if line == "QUIT" {
						fmt.Fprintf(c, "BYE\n")
						return
					}
					fmt.Fprintf(c, "ECHO: %s\n", line)
				}
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("  Dial error: %v\n", err)
		listener.Close()
		return
	}

	reader := bufio.NewReader(conn)

	messages := []string{"Hello", "World", "Networking in Go", "QUIT"}
	for _, msg := range messages {
		fmt.Fprintf(conn, "%s\n", msg)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("  Read error: %v\n", err)
			break
		}
		fmt.Printf("  Sent: %-20s -> Received: %s", msg, response)
	}

	conn.Close()
	listener.Close()
	wg.Wait()
	fmt.Println("  Echo server stopped")
}

// ============================================
// DEMO: TLS CONNECTION
// ============================================

func demoTLSConnection() {
	cert, err := generateDevCert()
	if err != nil {
		fmt.Printf("  Certificate generation error: %v\n", err)
		return
	}
	fmt.Println("  Generated self-signed certificate")

	serverConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverConfig)
	if err != nil {
		fmt.Printf("  TLS Listen error: %v\n", err)
		return
	}
	addr := listener.Addr().String()
	fmt.Printf("  TLS server listening on %s\n", addr)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		tlsConn := conn.(*tls.Conn)
		if err := tlsConn.Handshake(); err != nil {
			fmt.Printf("  Server handshake error: %v\n", err)
			return
		}

		state := tlsConn.ConnectionState()
		fmt.Printf("  Server: TLS %s, cipher=%s\n",
			tlsVersionName(state.Version),
			tls.CipherSuiteName(state.CipherSuite))

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		response := fmt.Sprintf("TLS ECHO: %s", buf[:n])
		conn.Write([]byte(response))
	}()

	time.Sleep(50 * time.Millisecond)

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		fmt.Printf("  Parse cert error: %v\n", err)
		listener.Close()
		wg.Wait()
		return
	}
	certPool := x509.NewCertPool()
	certPool.AddCert(x509Cert)

	clientConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
		ServerName: "localhost",
	}

	conn, err := tls.Dial("tcp", addr, clientConfig)
	if err != nil {
		fmt.Printf("  TLS Dial error: %v\n", err)
		listener.Close()
		wg.Wait()
		return
	}

	state := conn.ConnectionState()
	fmt.Printf("  Client: TLS %s, cipher=%s\n",
		tlsVersionName(state.Version),
		tls.CipherSuiteName(state.CipherSuite))

	conn.Write([]byte("Hello TLS!"))

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("  Client read error: %v\n", err)
	} else {
		fmt.Printf("  Client received: %s\n", string(buf[:n]))
	}

	conn.Close()
	listener.Close()
	wg.Wait()
	fmt.Println("  TLS demo completed")
}

func generateDevCert() (tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Dev"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "1.0"
	case tls.VersionTLS11:
		return "1.1"
	case tls.VersionTLS12:
		return "1.2"
	case tls.VersionTLS13:
		return "1.3"
	default:
		return "unknown"
	}
}

// Supress unused import warnings
var _ = strings.TrimSpace
var _ = os.Stdout
var _ = context.Background
var _ = bufio.NewReader
var _ = big.NewInt
var _ = elliptic.P256
var _ = pem.Encode
var _ = pkix.Name{}
var _ = x509.CreateCertificate

/*
SUMMARY - CHAPTER 080: LOW-LEVEL NETWORK PROGRAMMING

NET PACKAGE CORE:
- net.Listen/net.Dial for TCP servers and clients
- net.ListenPacket for UDP servers (ReadFrom/WriteTo)
- net.Conn interface: Read, Write, Close, deadlines (io.Reader+Writer)
- net.Listener interface: Accept, Close, Addr
- Supports tcp, udp, unix, ip with IPv4/IPv6 variants

TCP PATTERNS:
- Server: Listen + Accept loop + goroutine per connection
- Client: Dial, DialTimeout, Dialer.DialContext with context
- Concurrent server with semaphore channel for connection limits
- Graceful shutdown: close listener, wait active conns, force timeout
- bufio.Scanner/Reader for line-oriented text protocols

UDP AND UNIX SOCKETS:
- UDP: connectionless, best-effort, max 65535 byte payload
- Unix domain sockets: faster than TCP localhost, filesystem permissions
- Stream ("unix") and datagram ("unixgram") modes

DEADLINES AND TIMEOUTS:
- SetDeadline/SetReadDeadline/SetWriteDeadline are absolute (time.Time)
- Timeout does not close connection, only fails the current operation
- Idle timeout pattern: reset deadline before each Read

DNS AND IP HANDLING:
- LookupHost, LookupAddr, LookupIP, LookupMX, LookupNS, LookupTXT
- Custom Resolver with PreferGo and custom Dial function
- net.IP + net.IPNet for CIDR, netip.Addr/AddrPort/Prefix (Go 1.18+)

TLS AND mTLS:
- tls.Listen/tls.Dial for encrypted connections
- tls.Config: MinVersion, CipherSuites, ALPN, certificates
- mTLS: RequireAndVerifyClientCert + ClientCAs pool
- Certificate generation with crypto/x509 and crypto/ecdsa

PRODUCTION PATTERNS:
- Connection pooling with channel-based pool and health checks
- Custom protocol implementation (Redis-like RESP)
- TCP keep-alive configuration (Go 1.24+ KeepAliveConfig)
- SO_REUSEPORT for zero-downtime deploys
- Graceful connection draining with health check integration
*/
