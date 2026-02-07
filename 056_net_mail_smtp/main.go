// Package main - Chapter 056: Net Mail, SMTP & RPC
// Parsing y envio de email con net/mail y net/smtp,
// y patrones RPC con net/rpc y net/rpc/jsonrpc.
package main

import (
	"bytes"
	"fmt"
	"net"
	"net/mail"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== NET/MAIL, NET/SMTP & NET/RPC ===")

	// ============================================
	// 1. NET/MAIL - PARSING DE DIRECCIONES EMAIL
	// ============================================
	fmt.Println("\n--- 1. net/mail - Parsing de direcciones ---")
	os.Stdout.WriteString(`
NET/MAIL - PARSING DE EMAIL:

  mail.ParseAddress(addr string)           // Parsear una direccion
  mail.ParseAddressList(list string)       // Parsear lista separada por comas
  mail.ReadMessage(r io.Reader)            // Leer mensaje completo (headers + body)
  mail.ParseDate(date string)              // Parsear fecha RFC 2822
  mail.Header                               // Map de cabeceras del mensaje

  mail.Address{Name, Address}:
    Name:    "Juan Perez"
    Address: "juan@example.com"
    String() -> "Juan Perez <juan@example.com>"
`)

	addr, err := mail.ParseAddress("Juan Perez <juan@example.com>")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Nombre:    %s\n", addr.Name)
		fmt.Printf("  Direccion: %s\n", addr.Address)
		fmt.Printf("  Formateado: %s\n", addr.String())
	}

	addr2, _ := mail.ParseAddress("simple@example.com")
	fmt.Printf("  Sin nombre: %q -> Address=%s Name=%q\n", "simple@example.com", addr2.Address, addr2.Name)

	addrs, err := mail.ParseAddressList("Ana <ana@mail.com>, Bob <bob@mail.com>, carol@mail.com")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Lista de direcciones (%d):\n", len(addrs))
		for _, a := range addrs {
			fmt.Printf("    -> %s (%s)\n", a.Address, a.Name)
		}
	}

	customAddr := &mail.Address{Name: "Soporte Tecnico", Address: "soporte@empresa.com"}
	fmt.Printf("  Construida: %s\n", customAddr.String())

	// ============================================
	// 2. NET/MAIL - LEER MENSAJES EMAIL
	// ============================================
	fmt.Println("\n--- 2. net/mail - Leer mensajes ---")
	os.Stdout.WriteString(`
FORMATO DE EMAIL (RFC 5322):

  From: remitente
  To: destinatario(s)
  Subject: asunto
  Date: fecha RFC 2822
  Content-Type: tipo MIME
                                          <- linea en blanco
  Cuerpo del mensaje...

  msg.Header.Get("Subject")              // Obtener cabecera
  msg.Header.Date()                      // Parsear fecha
  msg.Header.AddressList("To")           // Parsear lista de destinatarios
  io.ReadAll(msg.Body)                   // Leer cuerpo
`)

	rawEmail := `From: Ana Garcia <ana@example.com>
To: bob@example.com, Carlos <carlos@example.com>
Cc: admin@example.com
Subject: Reunion de equipo
Date: Mon, 15 Jan 2025 10:30:00 -0500
Content-Type: text/plain; charset=utf-8
Message-ID: <abc123@example.com>

Hola equipo,

La reunion de planificacion sera el martes a las 10am.
Por favor confirmen su asistencia.

Saludos,
Ana`

	msg, err := mail.ReadMessage(strings.NewReader(rawEmail))
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  From:    %s\n", msg.Header.Get("From"))
		fmt.Printf("  Subject: %s\n", msg.Header.Get("Subject"))
		fmt.Printf("  Msg-ID:  %s\n", msg.Header.Get("Message-ID"))

		date, _ := msg.Header.Date()
		fmt.Printf("  Date:    %s\n", date.Format(time.RFC3339))

		toList, _ := msg.Header.AddressList("To")
		fmt.Printf("  To (%d destinatarios):\n", len(toList))
		for _, t := range toList {
			fmt.Printf("    - %s (%s)\n", t.Address, t.Name)
		}

		body := new(bytes.Buffer)
		body.ReadFrom(msg.Body)
		lines := strings.Split(strings.TrimSpace(body.String()), "\n")
		fmt.Printf("  Body (%d lineas):\n", len(lines))
		for _, line := range lines[:3] {
			fmt.Printf("    | %s\n", line)
		}
		fmt.Println("    | ...")
	}

	// ============================================
	// 3. NET/MAIL - PARSEAR FECHAS RFC 2822
	// ============================================
	fmt.Println("\n--- 3. net/mail - Parsear fechas ---")

	dateFormats := []string{
		"Mon, 15 Jan 2025 10:30:00 -0500",
		"15 Jan 2025 10:30:00 +0000",
		"Mon, 15 Jan 2025 10:30:00 PST",
	}
	for _, d := range dateFormats {
		t, err := mail.ParseDate(d)
		if err != nil {
			fmt.Printf("  %q -> Error: %v\n", d, err)
		} else {
			fmt.Printf("  %q\n    -> %s\n", d, t.Format(time.RFC3339))
		}
	}

	// ============================================
	// 4. NET/SMTP - ENVIO DE EMAIL (CONCEPTOS)
	// ============================================
	fmt.Println("\n--- 4. net/smtp - Conceptos de envio ---")
	os.Stdout.WriteString(`
NET/SMTP - ENVIO DE EMAIL:

  smtp.SendMail(addr, auth, from, to, msg)  // Envio simple
  smtp.PlainAuth(identity, user, pass, host) // Autenticacion PLAIN
  smtp.CRAMMD5Auth(user, secret)             // Autenticacion CRAM-MD5

  smtp.NewClient(conn, host)                 // Cliente SMTP manual
  c.Hello(localName)                         // EHLO/HELO
  c.StartTLS(config)                         // Iniciar TLS
  c.Auth(auth)                               // Autenticar
  c.Mail(from)                               // MAIL FROM
  c.Rcpt(to)                                 // RCPT TO
  c.Data()                                   // Iniciar cuerpo
  c.Quit()                                   // Cerrar conexion

EJEMPLO DE USO (no ejecutable sin servidor SMTP):

  auth := smtp.PlainAuth("", "user@gmail.com", "password", "smtp.gmail.com")
  to := []string{"destino@example.com"}
  msg := []byte("To: destino@example.com\r\n" +
      "Subject: Test\r\n" +
      "\r\n" +
      "Cuerpo del mensaje.\r\n")
  err := smtp.SendMail("smtp.gmail.com:587", auth, "user@gmail.com", to, msg)

FLUJO SMTP:
  1. Conectar al servidor (puerto 25/465/587)
  2. EHLO (presentarse)
  3. STARTTLS (cifrar conexion)
  4. AUTH (autenticarse)
  5. MAIL FROM (remitente)
  6. RCPT TO (destinatarios)
  7. DATA (contenido del mensaje)
  8. QUIT (cerrar)
`)

	fmt.Println("  Construyendo mensaje SMTP valido:")
	smtpMsg := buildSMTPMessage(
		"sender@example.com",
		[]string{"dest1@example.com", "dest2@example.com"},
		"Actualizacion del proyecto",
		"Hola,\n\nEl deploy se completo exitosamente.\n\nSaludos.",
	)
	lines := strings.Split(smtpMsg, "\r\n")
	for _, l := range lines {
		fmt.Printf("    %s\n", l)
	}

	// ============================================
	// 5. NET/RPC - SERVIDOR Y CLIENTE RPC
	// ============================================
	fmt.Println("\n--- 5. net/rpc - Remote Procedure Call ---")
	os.Stdout.WriteString(`
NET/RPC (legacy pero educativo):

  SERVIDOR:
  rpc.Register(service)                    // Registrar servicio
  rpc.RegisterName(name, service)          // Con nombre custom
  rpc.ServeConn(conn)                      // Servir una conexion
  rpc.HandleHTTP()                         // Servir via HTTP

  CLIENTE:
  rpc.Dial(network, address)               // Conectar
  rpc.DialHTTP(network, address)           // Conectar via HTTP
  client.Call(method, args, reply)         // Llamada sincrona
  client.Go(method, args, reply, done)     // Llamada asincrona

  REGLAS PARA METODOS RPC:
  - El tipo del servicio debe ser exportado
  - El metodo debe ser exportado
  - Dos argumentos: args (entrada) y reply (salida, puntero)
  - Retorna error
  - Firma: func (t *T) Method(args ArgsType, reply *ReplyType) error
`)

	fmt.Println("  Iniciando servidor RPC...")

	calculator := new(Calculator)
	server := rpc.NewServer()
	server.Register(calculator)

	clientConn, serverConn := net.Pipe()

	go server.ServeConn(serverConn)

	client := rpc.NewClient(clientConn)

	var result int
	err = client.Call("Calculator.Add", &CalcArgs{A: 15, B: 27}, &result)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Calculator.Add(15, 27) = %d\n", result)
	}

	err = client.Call("Calculator.Multiply", &CalcArgs{A: 6, B: 7}, &result)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Calculator.Multiply(6, 7) = %d\n", result)
	}

	err = client.Call("Calculator.Divide", &CalcArgs{A: 100, B: 3}, &result)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Calculator.Divide(100, 3) = %d\n", result)
	}

	err = client.Call("Calculator.Divide", &CalcArgs{A: 10, B: 0}, &result)
	if err != nil {
		fmt.Printf("  Calculator.Divide(10, 0) -> Error: %v\n", err)
	}

	client.Close()

	// ============================================
	// 6. NET/RPC - LLAMADAS ASINCRONAS
	// ============================================
	fmt.Println("\n--- 6. net/rpc - Llamadas asincronas ---")

	clientConn2, serverConn2 := net.Pipe()
	go server.ServeConn(serverConn2)
	client2 := rpc.NewClient(clientConn2)

	var r1, r2, r3 int
	call1 := client2.Go("Calculator.Add", &CalcArgs{A: 100, B: 200}, &r1, nil)
	call2 := client2.Go("Calculator.Multiply", &CalcArgs{A: 11, B: 11}, &r2, nil)
	call3 := client2.Go("Calculator.Add", &CalcArgs{A: 999, B: 1}, &r3, nil)

	<-call1.Done
	<-call2.Done
	<-call3.Done

	fmt.Printf("  Async Add(100,200) = %d\n", r1)
	fmt.Printf("  Async Multiply(11,11) = %d\n", r2)
	fmt.Printf("  Async Add(999,1) = %d\n", r3)

	client2.Close()

	// ============================================
	// 7. NET/RPC/JSONRPC
	// ============================================
	fmt.Println("\n--- 7. net/rpc/jsonrpc ---")
	os.Stdout.WriteString(`
NET/RPC/JSONRPC:

  SERVIDOR:
  jsonrpc.ServeConn(conn)                  // Servir con codec JSON

  CLIENTE:
  jsonrpc.Dial(network, address)           // Conectar con codec JSON
  jsonrpc.NewClient(conn)                  // Cliente desde conexion

  Ventajas sobre gob encoding:
  - Interoperable con otros lenguajes
  - Facil de depurar (texto legible)
  - Formato estandar JSON-RPC 1.0
`)

	clientConn3, serverConn3 := net.Pipe()
	go server.ServeCodec(jsonrpc.NewServerCodec(serverConn3))
	jsonClient := jsonrpc.NewClient(clientConn3)

	var jsonResult int
	err = jsonClient.Call("Calculator.Add", &CalcArgs{A: 42, B: 58}, &jsonResult)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  JSON-RPC Calculator.Add(42, 58) = %d\n", jsonResult)
	}

	err = jsonClient.Call("Calculator.Multiply", &CalcArgs{A: 12, B: 12}, &jsonResult)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  JSON-RPC Calculator.Multiply(12, 12) = %d\n", jsonResult)
	}

	jsonClient.Close()

	// ============================================
	// 8. NET/RPC - SERVICIO CON ESTADO
	// ============================================
	fmt.Println("\n--- 8. Servicio RPC con estado ---")

	kvStore := &KeyValueStore{data: make(map[string]string)}
	server2 := rpc.NewServer()
	server2.Register(kvStore)

	clientConn4, serverConn4 := net.Pipe()
	go server2.ServeConn(serverConn4)
	kvClient := rpc.NewClient(clientConn4)

	var ok bool
	kvClient.Call("KeyValueStore.Set", &KVSetArgs{Key: "nombre", Value: "Go"}, &ok)
	kvClient.Call("KeyValueStore.Set", &KVSetArgs{Key: "version", Value: "1.22"}, &ok)
	kvClient.Call("KeyValueStore.Set", &KVSetArgs{Key: "autor", Value: "Google"}, &ok)
	fmt.Printf("  Set 3 claves: ok=%v\n", ok)

	var val string
	kvClient.Call("KeyValueStore.Get", "nombre", &val)
	fmt.Printf("  Get('nombre') = %s\n", val)

	kvClient.Call("KeyValueStore.Get", "version", &val)
	fmt.Printf("  Get('version') = %s\n", val)

	var keys []string
	kvClient.Call("KeyValueStore.Keys", struct{}{}, &keys)
	fmt.Printf("  Keys() = %v\n", keys)

	kvClient.Close()

	fmt.Println("\n=== FIN CAPITULO 056 ===")
}

func buildSMTPMessage(from string, to []string, subject, body string) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("From: %s\r\n", from))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	buf.WriteString("\r\n")
	return buf.String()
}

// ============================================
// TIPOS Y SERVICIOS RPC
// ============================================

type CalcArgs struct {
	A, B int
}

type Calculator struct{}

func (c *Calculator) Add(args *CalcArgs, reply *int) error {
	*reply = args.A + args.B
	return nil
}

func (c *Calculator) Multiply(args *CalcArgs, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (c *Calculator) Divide(args *CalcArgs, reply *int) error {
	if args.B == 0 {
		return fmt.Errorf("division por cero")
	}
	*reply = args.A / args.B
	return nil
}

type KVSetArgs struct {
	Key   string
	Value string
}

type KeyValueStore struct {
	data map[string]string
}

func (kv *KeyValueStore) Set(args *KVSetArgs, reply *bool) error {
	kv.data[args.Key] = args.Value
	*reply = true
	return nil
}

func (kv *KeyValueStore) Get(key string, reply *string) error {
	val, ok := kv.data[key]
	if !ok {
		return fmt.Errorf("clave no encontrada: %s", key)
	}
	*reply = val
	return nil
}

func (kv *KeyValueStore) Keys(_ struct{}, reply *[]string) error {
	keys := make([]string, 0, len(kv.data))
	for k := range kv.data {
		keys = append(keys, k)
	}
	*reply = keys
	return nil
}

/*
SUMMARY - NET/MAIL, NET/SMTP & NET/RPC:

NET/MAIL - EMAIL PARSING:
- mail.ParseAddress and ParseAddressList for RFC 5322 addresses
- mail.ReadMessage parses headers + body from io.Reader
- mail.ParseDate for RFC 2822 date formats
- Header.Get, Header.Date, Header.AddressList for field access

NET/SMTP - EMAIL SENDING:
- smtp.SendMail for simple sending with PlainAuth or CRAMMD5Auth
- smtp.NewClient for manual SMTP flow: Hello, StartTLS, Auth, Mail, Rcpt, Data, Quit
- SMTP flow: connect -> EHLO -> STARTTLS -> AUTH -> MAIL FROM -> RCPT TO -> DATA -> QUIT

NET/RPC - REMOTE PROCEDURE CALLS:
- rpc.Register to expose service methods (exported type, exported method, args + *reply, returns error)
- rpc.NewServer and server.ServeConn for custom servers
- client.Call for synchronous, client.Go for asynchronous calls
- net.Pipe() for in-process testing without network

NET/RPC/JSONRPC:
- JSON codec for language-interoperable RPC
- jsonrpc.NewClient and jsonrpc.NewServerCodec
- Same API as gob-encoded RPC but human-readable

STATEFUL RPC SERVICE:
- Services can maintain internal state (maps, counters)
- KeyValueStore example with Set, Get, Keys methods
*/
