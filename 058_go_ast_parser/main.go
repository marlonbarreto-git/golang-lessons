// Package main - Chapter 058: Go AST & Parser
// Analisis de codigo Go con go/token, go/parser, go/ast,
// formateo con go/format, y construccion de un linter simple.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== GO AST & PARSER ===")

	// ============================================
	// 1. GO/TOKEN - TOKENS Y POSICIONES
	// ============================================
	fmt.Println("\n--- 1. go/token - Tokens y posiciones ---")
	os.Stdout.WriteString(`
GO/TOKEN - SISTEMA DE TOKENS:

  token.FileSet                            // Conjunto de archivos fuente
  fset.AddFile(name, base, size)           // Agregar archivo
  token.Pos                                // Posicion en el fuente
  fset.Position(pos) token.Position        // Convertir a linea:columna

  token.Token: tipo de token
    token.IDENT, token.INT, token.FLOAT, token.STRING
    token.ADD, token.SUB, token.MUL, token.QUO
    token.FUNC, token.RETURN, token.IF, token.FOR
    token.PACKAGE, token.IMPORT, token.VAR, token.CONST

  Funciones utiles:
    token.IsKeyword(name string) bool      // Es palabra clave?
    token.IsIdentifier(name string) bool   // Es identificador valido?
    token.Lookup(name string) token.Token  // Buscar keyword
`)

	keywords := []string{"func", "return", "if", "for", "select", "go", "defer", "main", "hello", "x"}
	for _, kw := range keywords {
		isKw := token.IsKeyword(kw)
		isId := token.IsIdentifier(kw)
		tok := token.Lookup(kw)
		fmt.Printf("  %-10s keyword:%v  ident:%v  token:%s\n", kw, isKw, isId, tok)
	}

	// ============================================
	// 2. GO/PARSER - PARSEAR CODIGO GO
	// ============================================
	fmt.Println("\n--- 2. go/parser - Parsear codigo ---")
	os.Stdout.WriteString(`
GO/PARSER:

  parser.ParseFile(fset, filename, src, mode)  // Parsear un archivo
  parser.ParseDir(fset, path, filter, mode)    // Parsear directorio

  Modos:
    parser.ParseComments   // Incluir comentarios en AST
    parser.AllErrors       // Reportar todos los errores
    parser.SkipObjectResolution // Saltar resolucion de objetos

  Retorna *ast.File con la estructura completa del codigo.
`)

	src := `package example

import (
	"fmt"
	"strings"
)

// Greet returns a greeting message for the given name.
func Greet(name string) string {
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello, %s!", strings.Title(name))
}

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

func helper() {
	x := 42
	_ = x
}

// Version is the current version.
var Version = "1.0.0"

// MaxRetries is the maximum number of retries.
const MaxRetries = 3
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", src, parser.ParseComments)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	fmt.Printf("  Package: %s\n", file.Name.Name)
	fmt.Printf("  Imports: %d\n", len(file.Imports))
	for _, imp := range file.Imports {
		fmt.Printf("    - %s\n", imp.Path.Value)
	}
	fmt.Printf("  Declaraciones: %d\n", len(file.Decls))

	// ============================================
	// 3. GO/AST - RECORRER EL AST
	// ============================================
	fmt.Println("\n--- 3. go/ast - Recorrer el AST ---")
	os.Stdout.WriteString(`
GO/AST - ARBOL SINTACTICO ABSTRACTO:

  ast.File          // Archivo completo
  ast.FuncDecl      // Declaracion de funcion
  ast.GenDecl       // Declaracion general (import, var, const, type)
  ast.ValueSpec     // Especificacion de valor (var/const)
  ast.TypeSpec      // Especificacion de tipo

  Recorrido:
  ast.Inspect(node, func(n ast.Node) bool)  // Visitar nodos (pre-order)
  ast.Walk(visitor, node)                    // Visitor pattern

  ast.Print(fset, node)                      // Imprimir AST
`)

	fmt.Println("  Funciones encontradas:")
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		pos := fset.Position(fn.Pos())
		hasDoc := fn.Doc != nil
		exported := fn.Name.IsExported()

		params := ""
		if fn.Type.Params != nil {
			for _, p := range fn.Type.Params.List {
				if params != "" {
					params += ", "
				}
				for _, n := range p.Names {
					params += n.Name + " "
				}
				params += formatNode(fset, p.Type)
			}
		}

		returns := ""
		if fn.Type.Results != nil {
			for _, r := range fn.Type.Results.List {
				if returns != "" {
					returns += ", "
				}
				returns += formatNode(fset, r.Type)
			}
		}

		fmt.Printf("    func %s(%s)", fn.Name.Name, params)
		if returns != "" {
			fmt.Printf(" %s", returns)
		}
		fmt.Printf("  [line:%d exported:%v doc:%v]\n", pos.Line, exported, hasDoc)
	}

	// ============================================
	// 4. GO/AST - INSPECT Y WALK
	// ============================================
	fmt.Println("\n--- 4. go/ast - Inspect ---")

	fmt.Println("  Todos los identificadores:")
	identCount := 0
	ast.Inspect(file, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if ok {
			identCount++
			if identCount <= 10 {
				pos := fset.Position(ident.Pos())
				fmt.Printf("    %s (line %d)\n", ident.Name, pos.Line)
			}
		}
		return true
	})
	fmt.Printf("    ... total: %d identificadores\n", identCount)

	fmt.Println("\n  Literales string:")
	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if ok && lit.Kind == token.STRING {
			pos := fset.Position(lit.Pos())
			fmt.Printf("    %s (line %d)\n", lit.Value, pos.Line)
		}
		return true
	})

	fmt.Println("\n  Sentencias return:")
	ast.Inspect(file, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if ok {
			pos := fset.Position(ret.Pos())
			results := make([]string, len(ret.Results))
			for i, r := range ret.Results {
				results[i] = formatNode(fset, r)
			}
			fmt.Printf("    return %s (line %d)\n", strings.Join(results, ", "), pos.Line)
		}
		return true
	})

	// ============================================
	// 5. GO/AST - AST.PRINT (DEBUG)
	// ============================================
	fmt.Println("\n--- 5. go/ast - Print del AST ---")

	miniSrc := `package mini
func Hello() string { return "hi" }
`
	miniFset := token.NewFileSet()
	miniFile, _ := parser.ParseFile(miniFset, "mini.go", miniSrc, 0)

	fmt.Println("  AST de funcion Hello():")
	for _, d := range miniFile.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			ast.Print(miniFset, fn)
		}
	}

	// ============================================
	// 6. GO/FORMAT - FORMATEAR CODIGO
	// ============================================
	fmt.Println("\n--- 6. go/format - Formatear codigo ---")
	os.Stdout.WriteString(`
GO/FORMAT:

  format.Source(src []byte)                // Formatear codigo fuente
  format.Node(w io.Writer, fset, node)     // Formatear nodo AST

  Equivalente a gofmt. Garantiza formato estandar de Go.
`)

	uglySrc := []byte(`package main
import "fmt"
func main(){
fmt.Println(   "hello"  )
x:=1+2
if x>2{fmt.Println("big")}
}`)

	fmt.Println("  Codigo sin formato:")
	for _, line := range strings.Split(string(uglySrc), "\n") {
		fmt.Printf("    %s\n", line)
	}

	formatted, err := format.Source(uglySrc)
	if err != nil {
		fmt.Printf("  Error formateando: %v\n", err)
	} else {
		fmt.Println("\n  Codigo formateado (format.Source):")
		for _, line := range strings.Split(string(formatted), "\n") {
			fmt.Printf("    %s\n", line)
		}
	}

	// ============================================
	// 7. GO/PRINTER - IMPRIMIR AST
	// ============================================
	fmt.Println("\n--- 7. go/printer - Imprimir AST ---")
	os.Stdout.WriteString(`
GO/PRINTER:

  printer.Fprint(w io.Writer, fset, node)  // Imprimir nodo AST
  printer.Config{Mode, Tabwidth}           // Configuracion custom

  Modos: SourcePos (incluir posiciones), TabIndent, UseSpaces
`)

	var printerBuf bytes.Buffer
	cfg := printer.Config{Mode: printer.TabIndent, Tabwidth: 4}
	cfg.Fprint(&printerBuf, fset, file)
	lines := strings.Split(printerBuf.String(), "\n")
	fmt.Printf("  Codigo regenerado desde AST (%d lineas):\n", len(lines))
	limit := 15
	if len(lines) < limit {
		limit = len(lines)
	}
	for _, line := range lines[:limit] {
		fmt.Printf("    %s\n", line)
	}
	if len(lines) > limit {
		fmt.Println("    ...")
	}

	// ============================================
	// 8. LINTER: FUNCIONES SIN DOCUMENTACION
	// ============================================
	fmt.Println("\n--- 8. Linter: Funciones exportadas sin doc ---")

	lintSrc := `package mylib

// PublicWithDoc is properly documented.
func PublicWithDoc() {}

func PublicNoDoc() {}

func AnotherPublic(x int) int { return x * 2 }

// Helper does something.
func Helper() string { return "" }

func privateFunc() {}

func anotherPrivate() {}

// ExportedConst is documented.
const ExportedConst = 42

var UndocumentedVar = "oops"
`

	lintFset := token.NewFileSet()
	lintFile, err := parser.ParseFile(lintFset, "mylib.go", lintSrc, parser.ParseComments)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		issues := lintUndocumentedExports(lintFset, lintFile)
		if len(issues) == 0 {
			fmt.Println("  No hay problemas encontrados!")
		} else {
			fmt.Printf("  Encontrados %d problemas:\n", len(issues))
			for _, issue := range issues {
				fmt.Printf("    %s\n", issue)
			}
		}
	}

	// ============================================
	// 9. EJEMPLO: ANALISIS DE COMPLEJIDAD
	// ============================================
	fmt.Println("\n--- 9. Ejemplo: Contar complejidad ---")

	complexSrc := `package main

func simple() {
	fmt.Println("hello")
}

func medium(x int) {
	if x > 0 {
		for i := 0; i < x; i++ {
			fmt.Println(i)
		}
	} else {
		fmt.Println("negative")
	}
}

func complex(items []string) {
	for _, item := range items {
		if len(item) > 5 {
			switch item[0] {
			case 'a':
				if item == "admin" {
					fmt.Println("admin found")
				}
			case 'b':
				fmt.Println("b item")
			default:
				for _, c := range item {
					if c == 'x' {
						fmt.Println("found x")
					}
				}
			}
		} else {
			fmt.Println(item)
		}
	}
}
`

	complexFset := token.NewFileSet()
	complexFile, _ := parser.ParseFile(complexFset, "complex.go", complexSrc, 0)

	for _, decl := range complexFile.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		complexity := calculateComplexity(fn)
		level := "simple"
		if complexity > 5 {
			level = "COMPLEJA"
		} else if complexity > 3 {
			level = "media"
		}
		fmt.Printf("  func %s: complejidad=%d (%s)\n", fn.Name.Name, complexity, level)
	}

	fmt.Println("\n=== FIN CAPITULO 058 ===")
}

func formatNode(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, node)
	return buf.String()
}

func lintUndocumentedExports(fset *token.FileSet, file *ast.File) []string {
	var issues []string

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() && d.Doc == nil {
				pos := fset.Position(d.Pos())
				issues = append(issues,
					fmt.Sprintf("%s:%d: exported function %s should have a comment",
						pos.Filename, pos.Line, d.Name.Name))
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if name.IsExported() && d.Doc == nil {
							pos := fset.Position(name.Pos())
							kind := "variable"
							if d.Tok == token.CONST {
								kind = "constant"
							}
							issues = append(issues,
								fmt.Sprintf("%s:%d: exported %s %s should have a comment",
									pos.Filename, pos.Line, kind, name.Name))
						}
					}
				case *ast.TypeSpec:
					if s.Name.IsExported() && d.Doc == nil {
						pos := fset.Position(s.Pos())
						issues = append(issues,
							fmt.Sprintf("%s:%d: exported type %s should have a comment",
								pos.Filename, pos.Line, s.Name.Name))
					}
				}
			}
		}
	}

	return issues
}

func calculateComplexity(fn *ast.FuncDecl) int {
	complexity := 1
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt, *ast.RangeStmt:
			complexity++
		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		case *ast.SelectStmt:
			complexity++
		case *ast.CommClause:
			complexity++
		}
		return true
	})
	return complexity
}
