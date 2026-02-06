// Package main - Chapter 046: Templates
// Go tiene un potente sistema de templates para generar texto y HTML.
// Aprende la sintaxis, funciones custom, composicion, auto-escaping
// y patrones de uso en produccion.
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	texttemplate "text/template"
	"time"
)

// ============================================
// TYPES FOR DEMOS
// ============================================

type Person struct {
	Name    string
	Age     int
	Email   string
	Active  bool
	Friends []string
	Address Address
	Tags    map[string]string
}

type Address struct {
	Street string
	City   string
	State  string
	ZIP    string
}

type Product struct {
	Name     string
	Price    float64
	Category string
	InStock  bool
	Rating   float64
}

type PageData struct {
	Title    string
	Products []Product
	User     *Person
	Year     int
}

type EmailData struct {
	To       string
	Subject  string
	Body     string
	Items    []OrderItem
	Total    float64
	Currency string
}

type OrderItem struct {
	Name     string
	Quantity int
	Price    float64
}

func main() {
	fmt.Println("=== TEMPLATES EN GO ===")

	// ============================================
	// 1. TEXT/TEMPLATE BASICOS
	// ============================================
	fmt.Println("\n--- 1. text/template Basicos ---")

	os.Stdout.WriteString(`
TEMPLATES EN GO:

Flujo basico:
  1. Crear template:   t := template.New("name")
  2. Parsear:          t.Parse(templateString)
  3. Ejecutar:         t.Execute(writer, data)

Shortcut:
  t := template.Must(template.New("name").Parse(str))

Acceso a datos:
  {{.}}         -> El dato completo pasado a Execute
  {{.Field}}    -> Campo de un struct
  {{.Method}}   -> Metodo sin argumentos
  {{print .}}   -> Funcion built-in
`)

	// Template basico con struct
	tmpl := texttemplate.Must(texttemplate.New("hello").Parse(
		"Hello, {{.Name}}! You are {{.Age}} years old.\n"))

	person := Person{Name: "Alice", Age: 30}
	_ = tmpl.Execute(os.Stdout, person)

	// Template con string
	strTmpl := texttemplate.Must(texttemplate.New("str").Parse(
		"The value is: {{.}}\n"))
	_ = strTmpl.Execute(os.Stdout, "Go Templates")

	// Template con map
	mapTmpl := texttemplate.Must(texttemplate.New("map").Parse(
		"Name: {{.name}}, Role: {{.role}}\n"))
	_ = mapTmpl.Execute(os.Stdout, map[string]string{
		"name": "Bob",
		"role": "admin",
	})

	// ============================================
	// 2. ACCIONES Y CONTROL DE FLUJO
	// ============================================
	fmt.Println("\n--- 2. Acciones y Control de Flujo ---")

	os.Stdout.WriteString(`
ACCIONES DE TEMPLATE:

Condicionales:
  {{if .Condition}}...{{end}}
  {{if .Condition}}...{{else}}...{{end}}
  {{if .A}}...{{else if .B}}...{{else}}...{{end}}

Iteracion:
  {{range .Slice}}{{.}}{{end}}
  {{range $index, $elem := .Slice}}{{$index}}: {{$elem}}{{end}}
  {{range .Slice}}...{{else}}empty{{end}}

Variables:
  {{$var := .Field}}
  {{$var := "value"}}

With (scope):
  {{with .Field}}...{{end}}         -> Si no es zero value
  {{with .Field}}...{{else}}...{{end}}

Comentarios:
  {{/* esto es un comentario */}}
`)

	// if/else
	ifTmpl := texttemplate.Must(texttemplate.New("if").Parse(
		`{{if .Active}}User {{.Name}} is ACTIVE{{else}}User {{.Name}} is INACTIVE{{end}}
`))
	_ = ifTmpl.Execute(os.Stdout, Person{Name: "Alice", Active: true})
	_ = ifTmpl.Execute(os.Stdout, Person{Name: "Bob", Active: false})

	// range con slice
	fmt.Println("\nRange con slice:")
	rangeTmpl := texttemplate.Must(texttemplate.New("range").Parse(
		`Friends of {{.Name}}:
{{range .Friends}}  - {{.}}
{{end}}`))
	_ = rangeTmpl.Execute(os.Stdout, Person{
		Name:    "Alice",
		Friends: []string{"Bob", "Charlie", "Diana"},
	})

	// range con index
	fmt.Println("Range con index:")
	indexTmpl := texttemplate.Must(texttemplate.New("index").Parse(
		`{{range $i, $friend := .Friends}}  [{{$i}}] {{$friend}}
{{end}}`))
	_ = indexTmpl.Execute(os.Stdout, Person{
		Friends: []string{"Bob", "Charlie", "Diana"},
	})

	// range con else (vacio)
	emptyTmpl := texttemplate.Must(texttemplate.New("empty").Parse(
		`{{range .Friends}}  - {{.}}
{{else}}  (no friends)
{{end}}`))
	fmt.Println("Range con else (vacio):")
	_ = emptyTmpl.Execute(os.Stdout, Person{Friends: nil})

	// with (scope)
	fmt.Println("With (scope):")
	withTmpl := texttemplate.Must(texttemplate.New("with").Parse(
		`{{with .Address}}Address: {{.Street}}, {{.City}}, {{.State}} {{.ZIP}}{{else}}No address{{end}}
`))
	_ = withTmpl.Execute(os.Stdout, Person{
		Address: Address{Street: "123 Main St", City: "NYC", State: "NY", ZIP: "10001"},
	})
	_ = withTmpl.Execute(os.Stdout, Person{})

	// Variables
	fmt.Println("\nVariables:")
	varTmpl := texttemplate.Must(texttemplate.New("var").Parse(
		`{{$name := .Name}}{{$upper := printf "%s" $name}}Name is: {{$upper}}
`))
	_ = varTmpl.Execute(os.Stdout, Person{Name: "Alice"})

	// ============================================
	// 3. PIPELINES Y FUNCIONES BUILT-IN
	// ============================================
	fmt.Println("\n--- 3. Pipelines y Funciones Built-in ---")

	os.Stdout.WriteString(`
PIPELINES (similar a Unix pipes):

  {{.Field | funcA | funcB}}
  Equivale a: funcB(funcA(.Field))

FUNCIONES BUILT-IN:
  and, or, not       -> Logicas
  eq, ne, lt, le,    -> Comparacion
  gt, ge
  len                -> Longitud
  index              -> Acceso por indice: index .Slice 0
  print, printf,     -> Formato (aliases de fmt.Sprint*)
  println
  call               -> Llamar funcion: call .Func arg1 arg2
  html, js, urlquery -> Escaping
  slice              -> Sub-slice: slice .S 1 3
`)

	// Pipeline
	fmt.Println("Pipeline demo:")
	pipeTmpl := texttemplate.Must(texttemplate.New("pipe").Parse(
		`Length of name: {{.Name | len | printf "%d"}}
`))
	_ = pipeTmpl.Execute(os.Stdout, Person{Name: "Alice"})

	// Comparacion
	fmt.Println("Comparacion:")
	cmpTmpl := texttemplate.Must(texttemplate.New("cmp").Parse(
		`{{if gt .Age 18}}{{.Name}} is an adult (age {{.Age}}){{else}}{{.Name}} is a minor{{end}}
`))
	_ = cmpTmpl.Execute(os.Stdout, Person{Name: "Alice", Age: 30})
	_ = cmpTmpl.Execute(os.Stdout, Person{Name: "Charlie", Age: 15})

	// index
	fmt.Println("\nIndex:")
	idxTmpl := texttemplate.Must(texttemplate.New("idx").Parse(
		`First friend: {{index .Friends 0}}, Last: {{index .Friends 2}}
`))
	_ = idxTmpl.Execute(os.Stdout, Person{Friends: []string{"Bob", "Charlie", "Diana"}})

	// and / or / not
	fmt.Println("Logicas (and/or/not):")
	logicTmpl := texttemplate.Must(texttemplate.New("logic").Parse(
		`{{if and .Active (gt .Age 18)}}{{.Name}}: active adult{{else}}{{.Name}}: does not match{{end}}
`))
	_ = logicTmpl.Execute(os.Stdout, Person{Name: "Alice", Active: true, Age: 30})
	_ = logicTmpl.Execute(os.Stdout, Person{Name: "Bob", Active: false, Age: 25})

	// ============================================
	// 4. FUNCIONES CUSTOM (FuncMap)
	// ============================================
	fmt.Println("\n--- 4. Funciones Custom (FuncMap) ---")

	os.Stdout.WriteString(`
FUNCIONES CUSTOM CON FuncMap:

  funcMap := template.FuncMap{
      "upper":    strings.ToUpper,
      "lower":    strings.ToLower,
      "title":    strings.Title,
      "contains": strings.Contains,
      "add":      func(a, b int) int { return a + b },
      "formatDate": func(t time.Time) string {
          return t.Format("2006-01-02")
      },
  }

  tmpl := template.New("name").Funcs(funcMap).Parse(str)

IMPORTANTE: Funcs() debe llamarse ANTES de Parse()
`)

	funcMap := texttemplate.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"repeat": func(s string, n int) string {
			return strings.Repeat(s, n)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"currency": func(amount float64, curr string) string {
			return fmt.Sprintf("%s %.2f", curr, amount)
		},
		"join": strings.Join,
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
	}

	funcTmpl := texttemplate.Must(texttemplate.New("func").Funcs(funcMap).Parse(
		`Name: {{.Name | upper}}
Email: {{.Email | lower}}
Age + 5: {{add .Age 5}}
Friends: {{join .Friends ", "}}
Separator: {{repeat "-" 20}}
`))
	_ = funcTmpl.Execute(os.Stdout, Person{
		Name: "Alice", Email: "ALICE@EXAMPLE.COM", Age: 30,
		Friends: []string{"Bob", "Charlie"},
	})

	// ============================================
	// 5. COMPOSICION DE TEMPLATES
	// ============================================
	fmt.Println("\n--- 5. Composicion de Templates ---")

	os.Stdout.WriteString(`
COMPOSICION CON define/template/block:

define - Definir un template con nombre:
  {{define "header"}}...{{end}}

template - Invocar un template definido:
  {{template "header" .}}

block - Define + template (con default):
  {{block "sidebar" .}}default content{{end}}

Esto permite:
  - Layouts reutilizables
  - Componentes parciales
  - Override de secciones (block)
`)

	// define + template
	composed := texttemplate.Must(texttemplate.New("page").Parse(`
{{define "header"}}=== {{.Title}} ==={{end}}
{{define "footer"}}--- Page rendered in {{.Year}} ---{{end}}
{{template "header" .}}
Products:
{{range .Products}}  * {{.Name}} (${{printf "%.2f" .Price}})
{{end}}
{{template "footer" .}}`))

	pageData := PageData{
		Title: "Product Catalog",
		Year:  2024,
		Products: []Product{
			{Name: "Widget", Price: 9.99},
			{Name: "Gadget", Price: 24.99},
			{Name: "Doohickey", Price: 14.50},
		},
	}
	_ = composed.Execute(os.Stdout, pageData)

	// block (define con default que puede ser overridden)
	fmt.Println("\nBlock (con default):")
	blockTmpl := texttemplate.Must(texttemplate.New("base").Parse(
		`Title: {{.Title}}
{{block "content" .}}Default content here{{end}}
`))
	_ = blockTmpl.Execute(os.Stdout, pageData)

	// Nested templates con New + ParseFiles alternativo
	fmt.Println("Nested templates con New:")
	nested := texttemplate.New("root")
	texttemplate.Must(nested.New("greeting").Parse(`Hello, {{.}}!`))
	texttemplate.Must(nested.New("farewell").Parse(`Goodbye, {{.}}!`))
	texttemplate.Must(nested.Parse(
		`{{template "greeting" .Name}} ... {{template "farewell" .Name}}
`))
	_ = nested.Execute(os.Stdout, Person{Name: "Alice"})

	// ============================================
	// 6. HTML/TEMPLATE (AUTO-ESCAPING)
	// ============================================
	fmt.Println("\n--- 6. html/template (Auto-Escaping) ---")

	os.Stdout.WriteString(`
HTML/TEMPLATE - SEGURIDAD AUTOMATICA:

html/template es identico a text/template en API,
pero con auto-escaping segun contexto:

  Contexto HTML:  <div>{{.}}</div>     -> Escapa <, >, &, "
  Contexto JS:    <script>var x={{.}}</script> -> Escapa para JS
  Contexto CSS:   <style>{{.}}</style> -> Escapa para CSS
  Contexto URL:   <a href="{{.}}">     -> Escapa para URL

Tipos seguros (no se escapan):
  template.HTML   -> HTML confiable
  template.CSS    -> CSS confiable
  template.JS     -> JavaScript confiable
  template.URL    -> URL confiable

REGLA: SIEMPRE usar html/template para generar HTML.
text/template NO escapa y es vulnerable a XSS.
`)

	// Comparacion: text/template vs html/template
	xssInput := `<script>alert("XSS")</script>`

	fmt.Println("text/template (INSEGURO para HTML):")
	textT := texttemplate.Must(texttemplate.New("text").Parse(
		`  <div>{{.}}</div>` + "\n"))
	_ = textT.Execute(os.Stdout, xssInput)

	fmt.Println("\nhtml/template (SEGURO - auto-escape):")
	htmlT := template.Must(template.New("html").Parse(
		`  <div>{{.}}</div>` + "\n"))
	_ = htmlT.Execute(os.Stdout, xssInput)

	// Tipos seguros
	fmt.Println("\ntipos seguros (template.HTML):")
	safeT := template.Must(template.New("safe").Parse(
		`  <div>{{.}}</div>` + "\n"))
	safeHTML := template.HTML(`<strong>bold text</strong>`)
	_ = safeT.Execute(os.Stdout, safeHTML)

	// Contextos de escaping
	fmt.Println("\nContextos de escaping:")
	ctxTmpl := template.Must(template.New("ctx").Parse(
		`  HTML:  <span>{{.}}</span>
  Attr:  <div title="{{.}}">text</div>
  URL:   <a href="/search?q={{.}}">link</a>
`))
	_ = ctxTmpl.Execute(os.Stdout, `hello "world" & <friends>`)

	// ============================================
	// 7. TEMPLATE HTML COMPLETO
	// ============================================
	fmt.Println("\n--- 7. Template HTML Completo ---")

	htmlPage := template.Must(template.New("page").Funcs(template.FuncMap{
		"currency": func(amount float64) string {
			return fmt.Sprintf("$%.2f", amount)
		},
		"upper": strings.ToUpper,
	}).Parse(`<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>
  {{if .User}}<h1>Welcome, {{.User.Name}}!</h1>{{end}}
  <h2>{{.Title}}</h2>
  <table>
    <tr><th>Product</th><th>Price</th><th>Stock</th></tr>
    {{range .Products}}
    <tr>
      <td>{{.Name}} ({{.Category | upper}})</td>
      <td>{{.Price | currency}}</td>
      <td>{{if .InStock}}In Stock{{else}}Out of Stock{{end}}</td>
    </tr>
    {{end}}
  </table>
  <footer>{{.Year}}</footer>
</body>
</html>`))

	fullPageData := PageData{
		Title: "Product Catalog",
		Year:  2024,
		User:  &Person{Name: "Alice"},
		Products: []Product{
			{Name: "Widget", Price: 9.99, Category: "tools", InStock: true},
			{Name: "Gadget", Price: 24.99, Category: "electronics", InStock: false},
			{Name: "Doohickey", Price: 14.50, Category: "misc", InStock: true},
		},
	}

	var htmlBuf bytes.Buffer
	_ = htmlPage.Execute(&htmlBuf, fullPageData)
	fmt.Println(htmlBuf.String())

	// ============================================
	// 8. EMAIL TEMPLATES
	// ============================================
	fmt.Println("\n--- 8. Email Templates ---")

	emailTmpl := texttemplate.Must(texttemplate.New("email").Funcs(funcMap).Parse(
		`To: {{.To}}
Subject: {{.Subject}}

Dear Customer,

{{.Body}}

Order Summary:
{{range .Items}}  - {{.Name}} (x{{.Quantity}}): {{currency .Price .Currency}}
{{end}}
Total: {{currency .Total .Currency}}

Thank you for your purchase!
Best regards, The Store Team
`))

	email := EmailData{
		To:       "alice@example.com",
		Subject:  "Order Confirmation #12345",
		Body:     "Thank you for your order. Here are the details:",
		Currency: "USD",
		Total:    84.97,
		Items: []OrderItem{
			{Name: "Widget", Quantity: 2, Price: 19.98},
			{Name: "Gadget", Quantity: 1, Price: 24.99},
			{Name: "Doohickey", Quantity: 1, Price: 40.00},
		},
	}
	_ = emailTmpl.Execute(os.Stdout, email)

	// ============================================
	// 9. CODE GENERATION TEMPLATE
	// ============================================
	fmt.Println("\n--- 9. Code Generation Template ---")

	os.Stdout.WriteString(`
TEMPLATES PARA GENERACION DE CODIGO:

Go genera mucho codigo con templates:
  - go generate + templates
  - stringer, enumer, mockgen
  - Generadores de API clients
  - ORM generators (sqlc, entgo)
`)

	type Field struct {
		Name string
		Type string
		JSON string
	}

	type StructDef struct {
		PackageName string
		StructName  string
		Fields      []Field
	}

	codegenTmpl := texttemplate.Must(texttemplate.New("codegen").Parse(
		`package {{.PackageName}}

type {{.StructName}} struct {
{{range .Fields}}	{{.Name}} {{.Type}} ` + "`" + `json:"{{.JSON}}"` + "`" + `
{{end}}}

func New{{.StructName}}() *{{.StructName}} {
	return &{{.StructName}}{}
}
`))

	structDef := StructDef{
		PackageName: "models",
		StructName:  "User",
		Fields: []Field{
			{Name: "ID", Type: "int64", JSON: "id"},
			{Name: "Name", Type: "string", JSON: "name"},
			{Name: "Email", Type: "string", JSON: "email"},
			{Name: "Active", Type: "bool", JSON: "active"},
		},
	}

	fmt.Println("Generated code:")
	_ = codegenTmpl.Execute(os.Stdout, structDef)

	// ============================================
	// 10. TEMPLATE CACHING
	// ============================================
	fmt.Println("\n--- 10. Template Caching ---")

	os.Stdout.WriteString(`
TEMPLATE CACHING - PATRON DE PRODUCCION:

// Parsear una vez al inicio (o init)
var templates = template.Must(
    template.New("").Funcs(funcMap).ParseGlob("templates/*.html"),
)

// Usar en handlers
func handler(w http.ResponseWriter, r *http.Request) {
    data := getData()
    templates.ExecuteTemplate(w, "page.html", data)
}

TIPS:
  - Parsear templates una sola vez (startup o init)
  - Usar template.Must() para fallar temprano
  - ParseGlob/ParseFiles para cargar multiples archivos
  - ExecuteTemplate para templates con nombre
  - En desarrollo: re-parsear en cada request (hot reload)

CON embed.FS (Go 1.16+):
  //go:embed templates/*
  var templateFS embed.FS

  templates = template.Must(
      template.New("").ParseFS(templateFS, "templates/*.html"),
  )
`)

	// Simular template cache
	cache := texttemplate.New("cache")
	texttemplate.Must(cache.New("header.txt").Parse("=== {{.}} ===\n"))
	texttemplate.Must(cache.New("item.txt").Parse("  * {{.}}\n"))
	texttemplate.Must(cache.New("footer.txt").Parse("--- end ---\n"))

	fmt.Println("Template cache demo:")
	_ = cache.ExecuteTemplate(os.Stdout, "header.txt", "My List")
	for _, item := range []string{"Alpha", "Beta", "Gamma"} {
		_ = cache.ExecuteTemplate(os.Stdout, "item.txt", item)
	}
	_ = cache.ExecuteTemplate(os.Stdout, "footer.txt", nil)

	// ============================================
	// 11. PATRONES AVANZADOS
	// ============================================
	fmt.Println("\n--- 11. Patrones Avanzados ---")

	// Patron: Template con metodos
	fmt.Println("Patron: Metodos en datos de template:")
	type TemplatePerson struct {
		FirstName string
		LastName  string
	}
	type TemplatePersonWrapper struct {
		P TemplatePerson
	}
	methodTmpl := texttemplate.Must(texttemplate.New("method").Funcs(texttemplate.FuncMap{
		"fullName": func(first, last string) string {
			return first + " " + last
		},
	}).Parse(
		`Full name: {{fullName .P.FirstName .P.LastName}}
`))
	_ = methodTmpl.Execute(os.Stdout, TemplatePersonWrapper{
		P: TemplatePerson{FirstName: "Alice", LastName: "Smith"},
	})

	// Patron: Render a string (no a writer)
	fmt.Println("\nPatron: Render a string:")
	renderTmpl := texttemplate.Must(texttemplate.New("render").Parse(
		`Hello, {{.Name}}! You have {{len .Friends}} friends.`))
	var renderBuf bytes.Buffer
	_ = renderTmpl.Execute(&renderBuf, Person{
		Name:    "Alice",
		Friends: []string{"Bob", "Charlie"},
	})
	result := renderBuf.String()
	fmt.Printf("Rendered string: %q\n", result)

	// Patron: Default values
	fmt.Println("\nPatron: Default values:")
	defaultTmpl := texttemplate.Must(texttemplate.New("default").Funcs(texttemplate.FuncMap{
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" || val == 0 {
				return def
			}
			return val
		},
	}).Parse(
		`Name: {{default "Anonymous" .Name}}, City: {{default "Unknown" .Address.City}}
`))
	_ = defaultTmpl.Execute(os.Stdout, Person{Name: "Alice"})
	_ = defaultTmpl.Execute(os.Stdout, Person{})

	// Patron: Whitespace control
	fmt.Println("\nPatron: Whitespace control:")
	os.Stdout.WriteString(`
WHITESPACE CONTROL:

  {{- .Field}}     -> Trim whitespace ANTES
  {{.Field -}}     -> Trim whitespace DESPUES
  {{- .Field -}}   -> Trim ambos lados

Ejemplo:
  "  {{- .Name -}}  " -> "Alice" (sin espacios)
  "  {{ .Name }}  "    -> "  Alice  " (con espacios)
`)

	wsTmpl := texttemplate.Must(texttemplate.New("ws").Parse(
		`Items: [{{range $i, $e := .}}{{if $i}}, {{end}}{{$e}}{{end}}]` + "\n"))
	_ = wsTmpl.Execute(os.Stdout, []string{"a", "b", "c"})

	// Patron: Maps en templates
	fmt.Println("Patron: Maps:")
	mapT := texttemplate.Must(texttemplate.New("mapTmpl").Parse(
		`Tags:
{{range $key, $value := .Tags}}  {{$key}}: {{$value}}
{{end}}`))
	_ = mapT.Execute(os.Stdout, Person{
		Tags: map[string]string{
			"role":       "admin",
			"department": "engineering",
			"level":      "senior",
		},
	})

	// ============================================
	// 12. ERRORES COMUNES
	// ============================================
	fmt.Println("--- 12. Errores Comunes ---")

	os.Stdout.WriteString(`
ERRORES COMUNES CON TEMPLATES:

1. Usar text/template para HTML (XSS vulnerable):
   MAL:  text/template para HTML
   BIEN: html/template para HTML

2. Parsear templates en cada request:
   MAL:  template.Parse() en el handler
   BIEN: Parse una vez, ExecuteTemplate en handler

3. Ignorar errores de Execute:
   MAL:  tmpl.Execute(w, data) // sin check error
   BIEN: if err := tmpl.Execute(&buf, data); err != nil { ... }
         // Escribir buf a w solo si no hay error

4. Olvidar Funcs() antes de Parse():
   MAL:  template.New("t").Parse(s).Funcs(fm)  // panic
   BIEN: template.New("t").Funcs(fm).Parse(s)

5. Marcar HTML como seguro sin sanitizar:
   MAL:  template.HTML(userInput)  // XSS!
   BIEN: template.HTML(sanitize(userInput))

6. No usar Must() para templates estaticos:
   MAL:  t, err := template.New("t").Parse(s)  // check en cada uso
   BIEN: t := template.Must(template.New("t").Parse(s))  // fail fast

PATRON SEGURO PARA HTTP HANDLERS:

  func render(w http.ResponseWriter, name string, data interface{}) {
      var buf bytes.Buffer
      if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
          http.Error(w, "Internal Server Error", 500)
          log.Printf("template error: %v", err)
          return
      }
      buf.WriteTo(w)
  }
`)

	// Demo error handling
	fmt.Println("Error handling demo:")
	badTmpl, err := texttemplate.New("bad").Parse(`{{.MissingField.Nested}}`)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
	} else {
		var errBuf bytes.Buffer
		err = badTmpl.Execute(&errBuf, struct{}{})
		if err != nil {
			fmt.Printf("Execute error: %v\n", err)
		}
	}

	// Suppress unused variable warnings
	_ = time.Now()

	fmt.Println("\n=== FIN DEL CAPITULO 73 ===")
}

/*
RESUMEN CAPITULO 73: TEMPLATES

text/template BASICOS:
- New(), Parse(), Execute() - flujo de trabajo basico
- {{.}} accede al dato pasado a Execute
- {{.Field}} accede a campos de struct
- Must() para fail-fast en templates estaticos

CONTROL DE FLUJO:
- {{if .Cond}}...{{else}}...{{end}}
- {{range .Slice}}...{{else}}empty{{end}}
- {{range $i, $e := .Slice}} para index
- {{with .Field}}...{{end}} para scope
- {{$var := value}} para variables

PIPELINES:
- {{.Field | func1 | func2}} estilo Unix pipes
- Funciones built-in: eq, ne, lt, gt, len, index, printf
- and, or, not para logica booleana

FUNCIONES CUSTOM (FuncMap):
- template.FuncMap{"name": func} para funciones propias
- Funcs() debe llamarse ANTES de Parse()
- Utiles para formateo, transformaciones, helpers

COMPOSICION:
- {{define "name"}}...{{end}} para definir templates
- {{template "name" .}} para invocar templates
- {{block "name" .}}default{{end}} para override
- ParseGlob/ParseFiles para cargar multiples archivos
- ExecuteTemplate para templates con nombre

html/template:
- API identica a text/template
- Auto-escaping segun contexto (HTML, JS, CSS, URL)
- Previene XSS automaticamente
- template.HTML/CSS/JS/URL para contenido confiable
- SIEMPRE usar para generar HTML

PATRONES:
- Template caching: parsear una vez, ejecutar muchas
- embed.FS para templates embebidos en binario
- Render a buffer antes de escribir a ResponseWriter
- Code generation con templates
- Email templates
- Default values con FuncMap custom
- Whitespace control con {{- y -}}

ERRORES COMUNES:
- text/template para HTML (usar html/template)
- Parsear en cada request (cachear)
- Ignorar errores de Execute
- Funcs() despues de Parse() (panic)
- template.HTML(userInput) sin sanitizar
*/
