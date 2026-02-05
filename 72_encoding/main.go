// Package main - Capitulo 72: Encoding Packages
// Go ofrece paquetes de encoding para JSON, XML, CSV, Gob, Binary,
// Base64, Hex y mas. Aprende cuando usar cada formato y como
// implementar interfaces de marshaling personalizadas.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

// ============================================
// TYPES FOR DEMOS
// ============================================

type User struct {
	ID        int       `json:"id" xml:"id"`
	Name      string    `json:"name" xml:"name"`
	Email     string    `json:"email" xml:"email"`
	Age       int       `json:"age,omitempty" xml:"age,omitempty"`
	Active    bool      `json:"active" xml:"active"`
	CreatedAt time.Time `json:"created_at" xml:"created_at"`
}

type Address struct {
	Street  string `json:"street" xml:"street"`
	City    string `json:"city" xml:"city"`
	Country string `json:"country" xml:"country"`
	ZIP     string `json:"zip" xml:"zip,attr"`
}

type UserWithAddress struct {
	User    User    `json:"user" xml:"user"`
	Address Address `json:"address" xml:"address"`
}

// CustomTime implementa json.Marshaler/Unmarshaler
type CustomTime struct {
	time.Time
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
	formatted := ct.Time.Format("2006-01-02")
	return json.Marshal(formatted)
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

// LogEntry implementa encoding.TextMarshaler/TextUnmarshaler
type LogEntry struct {
	Level   string
	Message string
}

func (le LogEntry) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("[%s] %s", le.Level, le.Message)), nil
}

func (le *LogEntry) UnmarshalText(data []byte) error {
	s := string(data)
	if len(s) < 4 || s[0] != '[' {
		return fmt.Errorf("invalid log entry format")
	}
	closeBracket := strings.Index(s, "]")
	if closeBracket == -1 {
		return fmt.Errorf("missing closing bracket")
	}
	le.Level = s[1:closeBracket]
	if closeBracket+2 < len(s) {
		le.Message = s[closeBracket+2:]
	}
	return nil
}

// Shape - para demo de gob con interfaces
type Shape interface {
	Area() float64
	Name() string
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 { return math.Pi * c.Radius * c.Radius }
func (c Circle) Name() string  { return "Circle" }

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 { return r.Width * r.Height }
func (r Rectangle) Name() string  { return "Rectangle" }

// Packet - para demo de encoding/binary
type Packet struct {
	Version  uint8
	Type     uint8
	Length   uint16
	Sequence uint32
}

func main() {
	fmt.Println("=== ENCODING PACKAGES EN GO ===")

	// ============================================
	// 1. ENCODING/JSON DEEP DIVE
	// ============================================
	fmt.Println("\n--- 1. encoding/json Deep Dive ---")

	fmt.Println(`
ENCODING/JSON - EL PAQUETE MAS USADO:

Funciones principales:
  json.Marshal(v)           -> []byte, error  (Go -> JSON)
  json.MarshalIndent(v,...) -> []byte, error  (Go -> JSON con formato)
  json.Unmarshal(data, &v)  -> error          (JSON -> Go)

Tags de struct:
  json:"name"         -> Nombre del campo en JSON
  json:"name,omitempty" -> Omitir si valor es zero value
  json:"-"            -> Ignorar campo completamente
  json:",string"      -> Codificar numero/bool como string JSON`)

	user := User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Age:       30,
		Active:    true,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	// Marshal basico
	data, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("\njson.Marshal (compacto):")
	fmt.Println(string(data))

	// MarshalIndent
	prettyData, _ := json.MarshalIndent(user, "", "  ")
	fmt.Println("\njson.MarshalIndent:")
	fmt.Println(string(prettyData))

	// Unmarshal
	jsonStr := `{"id":2,"name":"Bob","email":"bob@example.com","active":false,"created_at":"2024-06-01T00:00:00Z"}`
	var user2 User
	_ = json.Unmarshal([]byte(jsonStr), &user2)
	fmt.Printf("\njson.Unmarshal: %+v\n", user2)

	// omitempty en accion
	userNoAge := User{ID: 3, Name: "Charlie", Email: "charlie@example.com", Active: true}
	noAgeJSON, _ := json.Marshal(userNoAge)
	fmt.Println("\nomitempty (Age=0 se omite):")
	fmt.Println(string(noAgeJSON))

	// ============================================
	// JSON: RawMessage para parsing diferido
	// ============================================
	fmt.Println("\n--- json.RawMessage ---")

	fmt.Println(`
json.RawMessage permite diferir el parsing de un campo JSON.
Util cuando el tipo depende de otro campo:`)

	type Event struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	eventJSON := `{"type":"user_created","payload":{"id":1,"name":"Alice"}}`
	var event Event
	_ = json.Unmarshal([]byte(eventJSON), &event)
	fmt.Printf("Event type: %s\n", event.Type)
	fmt.Printf("Raw payload: %s\n", string(event.Payload))

	// Ahora parsear el payload segun el tipo
	if event.Type == "user_created" {
		var u struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		_ = json.Unmarshal(event.Payload, &u)
		fmt.Printf("Parsed payload: id=%d, name=%s\n", u.ID, u.Name)
	}

	// ============================================
	// JSON: Number para precision numerica
	// ============================================
	fmt.Println("\n--- json.Number ---")

	os.Stdout.WriteString(`
json.Number preserva la precision numerica original.
Por defecto, JSON decodifica numeros como float64,
lo cual puede perder precision con enteros grandes:

  dec := json.NewDecoder(reader)
  dec.UseNumber()  // Habilitar json.Number

  var data map[string]interface{}
  dec.Decode(&data)
  // data["big_id"] es json.Number, no float64

  num := data["big_id"].(json.Number)
  id, _ := num.Int64()    // Convertir a int64
  f, _ := num.Float64()   // Convertir a float64
  s := num.String()       // Como string original
`)

	numberJSON := `{"big_id": 9007199254740993, "amount": 99.99}`
	dec := json.NewDecoder(strings.NewReader(numberJSON))
	dec.UseNumber()
	var numMap map[string]interface{}
	_ = dec.Decode(&numMap)

	bigID := numMap["big_id"].(json.Number)
	idInt, _ := bigID.Int64()
	fmt.Printf("json.Number: string=%s, int64=%d\n", bigID.String(), idInt)

	// ============================================
	// JSON: Custom Marshaler
	// ============================================
	fmt.Println("\n--- Custom JSON Marshaler ---")

	os.Stdout.WriteString(`
Implementa json.Marshaler/json.Unmarshaler para control total:

  type json.Marshaler interface {
      MarshalJSON() ([]byte, error)
  }

  type json.Unmarshaler interface {
      UnmarshalJSON([]byte) error
  }
`)

	type Birthday struct {
		Name string     `json:"name"`
		Date CustomTime `json:"date"`
	}

	birthday := Birthday{
		Name: "Alice",
		Date: CustomTime{time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)},
	}
	bData, _ := json.Marshal(birthday)
	fmt.Printf("Custom marshal: %s\n", string(bData))

	var parsed Birthday
	_ = json.Unmarshal(bData, &parsed)
	fmt.Printf("Custom unmarshal: name=%s, date=%s\n", parsed.Name, parsed.Date.Format("2006-01-02"))

	// ============================================
	// JSON: Streaming con Encoder/Decoder
	// ============================================
	fmt.Println("\n--- JSON Streaming ---")

	fmt.Println(`
json.NewEncoder/json.NewDecoder para streaming:

  // Escribir JSON a un writer (archivo, HTTP response, etc)
  encoder := json.NewEncoder(writer)
  encoder.SetIndent("", "  ")
  encoder.Encode(value)

  // Leer JSON de un reader (archivo, HTTP body, etc)
  decoder := json.NewDecoder(reader)
  decoder.DisallowUnknownFields()  // Estricto
  decoder.Decode(&value)`)

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(user)
	fmt.Printf("Encoded to buffer:\n%s", buf.String())

	// Decodificar multiples objetos
	multiJSON := `{"name":"Alice"}
{"name":"Bob"}
{"name":"Charlie"}`
	decoder := json.NewDecoder(strings.NewReader(multiJSON))
	fmt.Println("Streaming decode:")
	for decoder.More() {
		var item struct{ Name string }
		if err := decoder.Decode(&item); err != nil {
			break
		}
		fmt.Printf("  -> %s\n", item.Name)
	}

	// ============================================
	// JSON: Map y interface{}
	// ============================================
	fmt.Println("\n--- JSON con map[string]interface{} ---")

	os.Stdout.WriteString(`
Para JSON dinamico sin struct predefinido:

  var data map[string]interface{}
  json.Unmarshal(jsonBytes, &data)

  // Tipos de Go para valores JSON:
  // bool              -> JSON boolean
  // float64           -> JSON number
  // string            -> JSON string
  // []interface{}     -> JSON array
  // map[string]interface{} -> JSON object
  // nil               -> JSON null
`)

	dynamicJSON := `{"name":"Alice","scores":[95,87,92],"metadata":{"role":"admin"}}`
	var dynamic map[string]interface{}
	_ = json.Unmarshal([]byte(dynamicJSON), &dynamic)
	fmt.Printf("Dynamic: name=%v, scores=%v\n", dynamic["name"], dynamic["scores"])

	// ============================================
	// 2. ENCODING/XML
	// ============================================
	fmt.Println("\n--- 2. encoding/xml ---")

	fmt.Println(`
ENCODING/XML - PARA SOAP, RSS, CONFIG:

Tags de struct:
  xml:"name"           -> Elemento hijo con ese nombre
  xml:"name,attr"      -> Atributo del elemento
  xml:",chardata"      -> Contenido de texto del elemento
  xml:",cdata"         -> Seccion CDATA
  xml:",innerxml"      -> XML crudo interno
  xml:",comment"       -> Comentario XML
  xml:"a>b>c"          -> Elementos anidados a/b/c
  xml:"-"              -> Ignorar campo`)

	type Book struct {
		XMLName xml.Name `xml:"book"`
		ISBN    string   `xml:"isbn,attr"`
		Title   string   `xml:"title"`
		Author  string   `xml:"author"`
		Year    int      `xml:"year"`
		Price   float64  `xml:"price"`
	}

	book := Book{
		ISBN:   "978-0-13-468599-1",
		Title:  "The Go Programming Language",
		Author: "Donovan & Kernighan",
		Year:   2015,
		Price:  34.99,
	}

	xmlData, _ := xml.MarshalIndent(book, "", "  ")
	fmt.Println("\nxml.MarshalIndent:")
	fmt.Println(xml.Header + string(xmlData))

	// Unmarshal XML
	xmlInput := `<book isbn="978-0-12-345678-9"><title>Learning Go</title><author>Jon Bodner</author><year>2021</year><price>49.99</price></book>`
	var book2 Book
	_ = xml.Unmarshal([]byte(xmlInput), &book2)
	fmt.Printf("\nxml.Unmarshal: %s by %s (%d) ISBN:%s\n", book2.Title, book2.Author, book2.Year, book2.ISBN)

	// XML con namespaces
	fmt.Println("\n--- XML Namespaces ---")

	type Envelope struct {
		XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
		Body    struct {
			Message string `xml:"message"`
		} `xml:"Body"`
	}

	env := Envelope{}
	env.Body.Message = "Hello SOAP"
	envData, _ := xml.MarshalIndent(env, "", "  ")
	fmt.Println(string(envData))

	// XML streaming con Encoder/Decoder
	fmt.Println("\n--- XML Streaming ---")

	os.Stdout.WriteString(`
XML Streaming con xml.NewEncoder/xml.NewDecoder:

  // Encoder
  encoder := xml.NewEncoder(writer)
  encoder.Indent("", "  ")
  encoder.Encode(value)
  encoder.Flush()

  // Decoder - Token-based parsing
  decoder := xml.NewDecoder(reader)
  for {
      token, err := decoder.Token()
      if err == io.EOF { break }
      switch t := token.(type) {
      case xml.StartElement:
          fmt.Println("Start:", t.Name.Local)
      case xml.EndElement:
          fmt.Println("End:", t.Name.Local)
      case xml.CharData:
          fmt.Println("Text:", string(t))
      }
  }
`)

	// ============================================
	// 3. ENCODING/GOB
	// ============================================
	fmt.Println("\n--- 3. encoding/gob ---")

	os.Stdout.WriteString(`
ENCODING/GOB - SERIALIZACION BINARIA DE GO:

Caracteristicas:
  - Formato binario nativo de Go (no compatible con otros lenguajes)
  - Mas eficiente que JSON/XML para comunicacion Go-a-Go
  - Soporta tipos complejos: structs, maps, slices, interfaces
  - Requiere gob.Register() para interfaces

Casos de uso:
  - RPC entre servicios Go
  - Cache de objetos Go
  - Persistencia rapida

  // Codificar
  var buf bytes.Buffer
  enc := gob.NewEncoder(&buf)
  enc.Encode(value)

  // Decodificar
  dec := gob.NewDecoder(&buf)
  dec.Decode(&target)
`)

	// Gob basico con struct
	var gobBuf bytes.Buffer
	gobEnc := gob.NewEncoder(&gobBuf)
	_ = gobEnc.Encode(user)
	fmt.Printf("Gob encoded size: %d bytes (vs JSON: %d bytes)\n", gobBuf.Len(), len(data))

	var gobUser User
	gobDec := gob.NewDecoder(&gobBuf)
	_ = gobDec.Decode(&gobUser)
	fmt.Printf("Gob decoded: %+v\n", gobUser)

	// Gob con interfaces (requiere Register)
	fmt.Println("\n--- Gob con interfaces ---")

	gob.Register(Circle{})
	gob.Register(Rectangle{})

	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 10, Height: 3},
	}

	var shapeBuf bytes.Buffer
	shapeEnc := gob.NewEncoder(&shapeBuf)
	_ = shapeEnc.Encode(shapes)
	fmt.Printf("Encoded %d shapes (%d bytes)\n", len(shapes), shapeBuf.Len())

	var decodedShapes []Shape
	shapeDec := gob.NewDecoder(&shapeBuf)
	_ = shapeDec.Decode(&decodedShapes)
	for _, s := range decodedShapes {
		fmt.Printf("  %s: area=%.2f\n", s.Name(), s.Area())
	}

	// Gob con maps
	fmt.Println("\n--- Gob con maps ---")

	config := map[string]interface{}{
		"host":    "localhost",
		"port":    8080,
		"debug":   true,
		"workers": 4,
	}

	gob.Register(map[string]interface{}{})
	var configBuf bytes.Buffer
	_ = gob.NewEncoder(&configBuf).Encode(config)
	fmt.Printf("Config encoded: %d bytes\n", configBuf.Len())

	var decodedConfig map[string]interface{}
	_ = gob.NewDecoder(&configBuf).Decode(&decodedConfig)
	fmt.Printf("Config decoded: %v\n", decodedConfig)

	// ============================================
	// 4. ENCODING/CSV
	// ============================================
	fmt.Println("\n--- 4. encoding/csv ---")

	fmt.Println(`
ENCODING/CSV - COMMA-SEPARATED VALUES:

Reader:
  reader := csv.NewReader(r)
  reader.Comma = ';'           // Delimitador custom
  reader.Comment = '#'         // Lineas de comentario
  reader.LazyQuotes = true     // Quotes no estrictos
  reader.TrimLeadingSpace = true

  records, _ := reader.ReadAll()  // Leer todo
  record, _ := reader.Read()      // Leer una fila

Writer:
  writer := csv.NewWriter(w)
  writer.Comma = '\t'          // TSV (tab-separated)
  writer.Write(record)         // Escribir una fila
  writer.WriteAll(records)     // Escribir todo
  writer.Flush()`)

	// CSV Writer
	var csvBuf bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuf)
	csvWriter.Write([]string{"Name", "Email", "Age"})
	csvWriter.Write([]string{"Alice", "alice@example.com", "30"})
	csvWriter.Write([]string{"Bob", "bob@example.com", "25"})
	csvWriter.Write([]string{"Charlie", "charlie@example.com", "35"})
	csvWriter.Flush()

	fmt.Println("\nCSV output:")
	fmt.Print(csvBuf.String())

	// CSV Reader
	csvReader := csv.NewReader(strings.NewReader(csvBuf.String()))
	records, _ := csvReader.ReadAll()
	fmt.Printf("\nCSV parsed: %d records (including header)\n", len(records))
	for i, record := range records {
		if i == 0 {
			fmt.Printf("  Headers: %v\n", record)
		} else {
			fmt.Printf("  Row %d: name=%s, email=%s, age=%s\n", i, record[0], record[1], record[2])
		}
	}

	// CSV con delimitador custom
	fmt.Println("\n--- CSV con delimitador custom ---")

	tsvData := "Name\tAge\tCity\nAlice\t30\tNYC\nBob\t25\tLA"
	tsvReader := csv.NewReader(strings.NewReader(tsvData))
	tsvReader.Comma = '\t'
	tsvRecords, _ := tsvReader.ReadAll()
	fmt.Println("TSV parsed:")
	for _, r := range tsvRecords {
		fmt.Printf("  %v\n", r)
	}

	// Lazy read (row by row para archivos grandes)
	fmt.Println("\n--- CSV lazy read ---")

	largeCSV := "id,name\n1,Alice\n2,Bob\n3,Charlie"
	lazyReader := csv.NewReader(strings.NewReader(largeCSV))
	fmt.Println("Lazy read (row by row):")
	for {
		record, err := lazyReader.Read()
		if err != nil {
			break
		}
		fmt.Printf("  %v\n", record)
	}

	// ============================================
	// 5. ENCODING/BINARY
	// ============================================
	fmt.Println("\n--- 5. encoding/binary ---")

	os.Stdout.WriteString(`
ENCODING/BINARY - PARA PROTOCOLOS DE RED Y FORMATOS BINARIOS:

ByteOrder:
  binary.BigEndian      -> Network byte order (MSB first)
  binary.LittleEndian   -> Intel/AMD byte order (LSB first)

Funciones:
  binary.Write(w, order, data)   -> Escribir struct/valor
  binary.Read(r, order, &data)   -> Leer struct/valor

  binary.BigEndian.PutUint16(b, v)    -> Escribir uint16 en []byte
  binary.BigEndian.Uint16(b)          -> Leer uint16 de []byte
  binary.BigEndian.PutUint32(b, v)    -> Escribir uint32
  binary.BigEndian.Uint32(b)          -> Leer uint32
  binary.BigEndian.PutUint64(b, v)    -> Escribir uint64
  binary.BigEndian.Uint64(b)          -> Leer uint64

Caso de uso: protocolos TCP, formatos de archivo, IPC
`)

	// Binary encoding de un packet
	packet := Packet{
		Version:  1,
		Type:     3,
		Length:   256,
		Sequence: 42,
	}

	var binBuf bytes.Buffer
	_ = binary.Write(&binBuf, binary.BigEndian, packet)
	fmt.Printf("Packet binary: %d bytes -> %v\n", binBuf.Len(), binBuf.Bytes())

	// Decodificar
	var decoded Packet
	_ = binary.Read(bytes.NewReader(binBuf.Bytes()), binary.BigEndian, &decoded)
	fmt.Printf("Decoded packet: Version=%d, Type=%d, Length=%d, Seq=%d\n",
		decoded.Version, decoded.Type, decoded.Length, decoded.Sequence)

	// ByteOrder directo
	fmt.Println("\n--- ByteOrder directo ---")

	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b[0:4], 0xDEADBEEF)
	binary.BigEndian.PutUint32(b[4:8], 0xCAFEBABE)
	fmt.Printf("Bytes: % X\n", b)
	fmt.Printf("First uint32: 0x%X\n", binary.BigEndian.Uint32(b[0:4]))
	fmt.Printf("Second uint32: 0x%X\n", binary.BigEndian.Uint32(b[4:8]))

	// Tamanio de tipos
	fmt.Println("\n--- binary.Size ---")

	fmt.Printf("Size of Packet: %d bytes\n", binary.Size(packet))
	fmt.Printf("Size of uint16: %d bytes\n", binary.Size(uint16(0)))
	fmt.Printf("Size of float64: %d bytes\n", binary.Size(float64(0)))

	// ============================================
	// 6. ENCODING/BASE64
	// ============================================
	fmt.Println("\n--- 6. encoding/base64 ---")

	os.Stdout.WriteString(`
ENCODING/BASE64 - CODIFICACION BINARIO-A-TEXTO:

Encodings disponibles:
  base64.StdEncoding     -> Estandar (A-Z, a-z, 0-9, +, /)
  base64.URLEncoding     -> URL-safe (A-Z, a-z, 0-9, -, _)
  base64.RawStdEncoding  -> Sin padding (=)
  base64.RawURLEncoding  -> URL-safe sin padding

Funciones:
  encoding.EncodeToString(data)  -> string
  encoding.DecodeString(s)       -> []byte, error
  encoding.EncodedLen(n)         -> int (tamanio resultante)
  encoding.DecodedLen(n)         -> int (tamanio original max)

Casos de uso: JWT, email attachments, data URIs, API keys
`)

	original := "Hello, Go encoding! Special chars: +/= and unicode: cafe"
	originalBytes := []byte(original)

	// Standard encoding
	stdEncoded := base64.StdEncoding.EncodeToString(originalBytes)
	fmt.Printf("Original:     %s\n", original)
	fmt.Printf("StdEncoding:  %s\n", stdEncoded)

	// URL encoding
	urlEncoded := base64.URLEncoding.EncodeToString(originalBytes)
	fmt.Printf("URLEncoding:  %s\n", urlEncoded)

	// Raw (sin padding)
	rawEncoded := base64.RawStdEncoding.EncodeToString(originalBytes)
	fmt.Printf("RawEncoding:  %s\n", rawEncoded)

	// Decode
	decodedBytes, _ := base64.StdEncoding.DecodeString(stdEncoded)
	fmt.Printf("Decoded:      %s\n", string(decodedBytes))

	// Codificar datos binarios
	fmt.Println("\n--- Base64 con datos binarios ---")
	binaryData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46}
	b64 := base64.StdEncoding.EncodeToString(binaryData)
	fmt.Printf("Binary data base64: %s\n", b64)
	fmt.Printf("Encoded length: %d (original: %d bytes)\n",
		base64.StdEncoding.EncodedLen(len(binaryData)), len(binaryData))

	// ============================================
	// 7. ENCODING/HEX
	// ============================================
	fmt.Println("\n--- 7. encoding/hex ---")

	os.Stdout.WriteString(`
ENCODING/HEX - CODIFICACION HEXADECIMAL:

Funciones:
  hex.EncodeToString(data)  -> string ("48656c6c6f")
  hex.DecodeString(s)       -> []byte, error
  hex.Encode(dst, src)      -> int
  hex.Decode(dst, src)      -> int, error
  hex.Dump(data)            -> string (formato hexdump)
  hex.Dumper(w)             -> io.WriteCloser

Casos de uso: hashes, claves, debugging binario, colores
`)

	hexOriginal := []byte("Hello, Hex!")
	hexEncoded := hex.EncodeToString(hexOriginal)
	fmt.Printf("Original:    %s\n", string(hexOriginal))
	fmt.Printf("Hex encoded: %s\n", hexEncoded)

	hexDecoded, _ := hex.DecodeString(hexEncoded)
	fmt.Printf("Hex decoded: %s\n", string(hexDecoded))

	// Hex dump (formato debug)
	fmt.Println("\nhex.Dump:")
	fmt.Print(hex.Dump([]byte("Hello, World! This is a hex dump demo.")))

	// ============================================
	// 8. INTERFACES DE ENCODING
	// ============================================
	fmt.Println("\n--- 8. Interfaces de Encoding ---")

	os.Stdout.WriteString(`
INTERFACES DEL PAQUETE ENCODING:

encoding.TextMarshaler / TextUnmarshaler:
  MarshalText() ([]byte, error)
  UnmarshalText([]byte) error

  Usado por: json (como fallback), xml, csv
  Cuando un tipo implementa TextMarshaler, json lo usa
  como string en lugar del formato default.

encoding.BinaryMarshaler / BinaryUnmarshaler:
  MarshalBinary() ([]byte, error)
  UnmarshalBinary([]byte) error

  Usado por: gob, otros formatos binarios
  Tipos de stdlib que lo implementan: time.Time, net.IP

JERARQUIA DE MARSHALING EN JSON:
  1. json.Marshaler (MarshalJSON)     -> Prioridad maxima
  2. encoding.TextMarshaler            -> Fallback a texto
  3. Reflection-based marshaling       -> Default
`)

	// TextMarshaler demo
	entry := LogEntry{Level: "ERROR", Message: "connection timeout"}
	textData, _ := entry.MarshalText()
	fmt.Printf("TextMarshaler: %s\n", string(textData))

	// TextMarshaler en JSON (se usa como string)
	jsonEntry, _ := json.Marshal(entry)
	fmt.Printf("As JSON (TextMarshaler): %s\n", string(jsonEntry))

	// TextUnmarshaler
	var parsedEntry LogEntry
	_ = parsedEntry.UnmarshalText([]byte("[WARN] disk space low"))
	fmt.Printf("TextUnmarshaler: level=%s, msg=%s\n", parsedEntry.Level, parsedEntry.Message)

	// BinaryMarshaler demo con time.Time
	fmt.Println("\n--- BinaryMarshaler (time.Time) ---")
	now := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	timeBinary, _ := now.MarshalBinary()
	fmt.Printf("time.Time binary: %d bytes -> % X\n", len(timeBinary), timeBinary)

	var restoredTime time.Time
	_ = restoredTime.UnmarshalBinary(timeBinary)
	fmt.Printf("Restored time: %s\n", restoredTime.Format(time.RFC3339))

	// ============================================
	// 9. PERFORMANCE Y CUANDO USAR CADA FORMATO
	// ============================================
	fmt.Println("\n--- 9. Comparacion de Formatos ---")

	os.Stdout.WriteString(`
COMPARACION DE FORMATOS DE ENCODING:

Formato  | Legible | Compacto | Rapido | Interop | Caso de uso
---------|---------|----------|--------|---------|----------------------------
JSON     | Si      | Medio    | Medio  | Alta    | APIs REST, config, web
XML      | Si      | No       | Lento  | Alta    | SOAP, RSS, config legacy
CSV      | Si      | Si       | Rapido | Alta    | Datos tabulares, imports
Gob      | No      | Si       | Rapido | Solo Go | RPC Go, cache interno
Binary   | No      | Si       | Rapido | Custom  | Protocolos, formatos binarios
Base64   | Semi    | No       | Rapido | Alta    | Transport binario en texto
Hex      | Semi    | No       | Rapido | Alta    | Hashes, debugging

RECOMENDACIONES:
  - API publica       -> JSON
  - Integracion SOAP  -> XML
  - Import/export     -> CSV
  - Entre servicios Go -> Gob (o protobuf para multi-lenguaje)
  - Protocolo custom  -> binary
  - Embed binario     -> base64
  - Debug/hashes      -> hex
`)

	// Demo de tamanios comparativos
	complexUser := UserWithAddress{
		User: User{
			ID: 1, Name: "Alice Johnson", Email: "alice@example.com",
			Age: 30, Active: true,
			CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		Address: Address{
			Street: "123 Main St", City: "New York",
			Country: "US", ZIP: "10001",
		},
	}

	jsonBytes, _ := json.Marshal(complexUser)
	xmlBytes, _ := xml.Marshal(complexUser)

	var gobCompBuf bytes.Buffer
	_ = gob.NewEncoder(&gobCompBuf).Encode(complexUser)

	fmt.Println("Tamanio comparativo (mismo objeto):")
	fmt.Printf("  JSON: %d bytes\n", len(jsonBytes))
	fmt.Printf("  XML:  %d bytes\n", len(xmlBytes))
	fmt.Printf("  Gob:  %d bytes\n", gobCompBuf.Len())

	// ============================================
	// 10. PATRONES COMUNES
	// ============================================
	fmt.Println("\n--- 10. Patrones Comunes ---")

	// Patron: JSON con campos desconocidos
	os.Stdout.WriteString(`
PATRON: Preservar campos desconocidos en JSON:

  type Config struct {
      Name    string          ` + "`" + `json:"name"` + "`" + `
      Version int             ` + "`" + `json:"version"` + "`" + `
      Extra   json.RawMessage ` + "`" + `json:"-"` + "`" + `
  }

  // Unmarshal preservando extras
  var raw map[string]json.RawMessage
  json.Unmarshal(data, &raw)
  json.Unmarshal(raw["name"], &config.Name)
  config.Extra = data  // Preservar original
`)

	// Patron: Enum con JSON
	fmt.Println("\n--- Patron: Enum con JSON ---")

	type Status int
	const (
		StatusPending  Status = iota
		StatusActive
		StatusInactive
	)

	statusNames := map[Status]string{
		StatusPending:  "pending",
		StatusActive:   "active",
		StatusInactive: "inactive",
	}

	statusJSON, _ := json.Marshal(statusNames[StatusActive])
	fmt.Printf("Status as JSON string: %s\n", string(statusJSON))

	// Patron: Nested JSON con flatten
	fmt.Println("\n--- Patron: Struct embedding en JSON ---")

	type Timestamps struct {
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
		Timestamps
	}

	product := Product{
		ID:    1,
		Name:  "Widget",
		Price: 9.99,
		Timestamps: Timestamps{
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	productJSON, _ := json.MarshalIndent(product, "", "  ")
	fmt.Printf("Embedded struct (flat JSON):\n%s\n", string(productJSON))

	fmt.Println("\n=== FIN DEL CAPITULO 72 ===")
}

/*
RESUMEN CAPITULO 72: ENCODING PACKAGES

encoding/json:
- Marshal/Unmarshal para conversion Go <-> JSON
- Tags: json:"name,omitempty", json:"-"
- json.RawMessage para parsing diferido
- json.Number para precision numerica
- Custom Marshaler/Unmarshaler para control total
- Streaming con Encoder/Decoder
- map[string]interface{} para JSON dinamico

encoding/xml:
- Marshal/Unmarshal similar a JSON pero con tags XML
- Atributos con xml:"name,attr"
- Namespaces con xml.Name
- Streaming con token-based parsing

encoding/gob:
- Formato binario nativo de Go
- Mas compacto y rapido que JSON
- Requiere gob.Register() para interfaces
- Solo compatible entre programas Go

encoding/csv:
- Reader/Writer para datos tabulares
- Delimitador configurable (CSV, TSV, etc)
- Lazy read fila por fila para archivos grandes

encoding/binary:
- BigEndian/LittleEndian para byte order
- Read/Write para serializar structs a binario
- PutUint16/32/64 para escritura directa en []byte
- Ideal para protocolos de red

encoding/base64:
- StdEncoding, URLEncoding, Raw variants
- Para transportar binario como texto
- Usado en JWT, email, data URIs

encoding/hex:
- EncodeToString/DecodeString
- Dump para debug hexadecimal
- Usado para hashes, claves, debugging

INTERFACES DE ENCODING:
- encoding.TextMarshaler/TextUnmarshaler
- encoding.BinaryMarshaler/BinaryUnmarshaler
- JSON prioriza: MarshalJSON > MarshalText > reflection
- time.Time implementa BinaryMarshaler

REGLA GENERAL:
- JSON para APIs y config
- XML para SOAP y legacy
- CSV para datos tabulares
- Gob para Go-a-Go
- Binary para protocolos
- Base64 para embed binario en texto
- Hex para debugging y hashes
*/
