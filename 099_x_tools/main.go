// Package main - Chapter 099: golang.org/x/tools
// x/tools provides developer tooling: static analysis (go/analysis),
// package loading (go/packages), and commands like goimports and guru.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== GOLANG.ORG/X/TOOLS ===")

	fmt.Println(`
  golang.org/x/tools provides essential Go developer tools:

  KEY PACKAGES:
    go/analysis     - Framework for writing static analyzers
    go/packages     - Load and inspect Go packages
    go/ssa          - Static Single Assignment form
    go/callgraph    - Call graph construction

  KEY COMMANDS:
    cmd/goimports   - Format + fix imports
    cmd/guru        - Source code exploration
    cmd/gorename    - Type-safe rename
    cmd/callgraph   - Call graph visualization
    cmd/godoc       - Documentation server

  Install: go get golang.org/x/tools/...`)

	// ============================================
	// GO/ANALYSIS FRAMEWORK
	// ============================================
	fmt.Println("\n--- go/analysis: Static Analysis Framework ---")

	fmt.Println(`
  The analysis framework is used by go vet and custom analyzers.
  Each analyzer inspects Go code and reports diagnostics.

  ANALYZER STRUCTURE:

    import "golang.org/x/tools/go/analysis"

    var MyAnalyzer = &analysis.Analyzer{
        Name: "mycheck",
        Doc:  "reports issues in Go code",
        Run:  run,
        Requires: []*analysis.Analyzer{
            inspect.Analyzer,  // optional: depends on other analyzers
        },
    }

  THE PASS OBJECT:

    func run(pass *analysis.Pass) (interface{}, error) {
        // pass.Fset      - file positions
        // pass.Files     - parsed AST files
        // pass.Pkg       - type-checked package
        // pass.TypesInfo - type information
        // pass.ResultOf  - results from required analyzers
        // pass.Report    - report a diagnostic
        // pass.Reportf   - report with position

        for _, file := range pass.Files {
            ast.Inspect(file, func(n ast.Node) bool {
                // Check each node
                if call, ok := n.(*ast.CallExpr); ok {
                    checkCall(pass, call)
                }
                return true
            })
        }
        return nil, nil
    }

  REPORTING DIAGNOSTICS:

    // Simple report
    pass.Reportf(node.Pos(), "found suspicious pattern")

    // Structured diagnostic with suggested fix
    pass.Report(analysis.Diagnostic{
        Pos:     node.Pos(),
        End:     node.End(),
        Message: "should use strings.EqualFold",
        SuggestedFixes: []analysis.SuggestedFix{{
            Message: "use strings.EqualFold",
            TextEdits: []analysis.TextEdit{{
                Pos:     node.Pos(),
                End:     node.End(),
                NewText: []byte(replacement),
            }},
        }},
    })`)

	// ============================================
	// BUILDING A CUSTOM ANALYZER
	// ============================================
	fmt.Println("\n--- Building a Custom Analyzer ---")

	os.Stdout.WriteString(`
  EXAMPLE: Detect fmt.Println in production code

    // analyzer.go
    package noprintln

    import (
        "go/ast"
        "golang.org/x/tools/go/analysis"
        "golang.org/x/tools/go/analysis/passes/inspect"
        "golang.org/x/tools/go/ast/inspector"
    )

    var Analyzer = &analysis.Analyzer{
        Name:     "noprintln",
        Doc:      "reports uses of fmt.Println",
        Requires: []*analysis.Analyzer{inspect.Analyzer},
        Run:      run,
    }

    func run(pass *analysis.Pass) (interface{}, error) {
        inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

        nodeFilter := []ast.Node{
            (*ast.CallExpr)(nil),
        }

        inspect.Preorder(nodeFilter, func(n ast.Node) {
            call := n.(*ast.CallExpr)
            sel, ok := call.Fun.(*ast.SelectorExpr)
            if !ok { return }
            ident, ok := sel.X.(*ast.Ident)
            if !ok { return }
            if ident.Name == "fmt" && sel.Sel.Name == "Println" {
                pass.Reportf(call.Pos(), "avoid fmt.Println in production; use structured logging")
            }
        })

        return nil, nil
    }

  REGISTER AND RUN:

    // cmd/noprintln/main.go
    package main

    import (
        "golang.org/x/tools/go/analysis/singlechecker"
        "mymodule/noprintln"
    )

    func main() {
        singlechecker.Main(noprintln.Analyzer)
    }

    // Run: go vet -vettool=$(which noprintln) ./...

  MULTI-ANALYZER:

    package main

    import (
        "golang.org/x/tools/go/analysis/multichecker"
        "mymodule/noprintln"
        "mymodule/errcheck"
    )

    func main() {
        multichecker.Main(
            noprintln.Analyzer,
            errcheck.Analyzer,
        )
    }
` + "\n")

	// ============================================
	// GO/PACKAGES
	// ============================================
	fmt.Println("--- go/packages: Load Go Packages ---")

	os.Stdout.WriteString(`
  go/packages loads Go packages with full type information.
  It's the recommended replacement for go/build.

  API:

    import "golang.org/x/tools/go/packages"

    // Load configuration
    cfg := &packages.Config{
        Mode: packages.NeedName |
              packages.NeedFiles |
              packages.NeedSyntax |
              packages.NeedTypes |
              packages.NeedTypesInfo,
        Dir:  "/path/to/project",
    }

    // Load packages
    pkgs, err := packages.Load(cfg, "./...")

    // Inspect loaded packages
    for _, pkg := range pkgs {
        fmt.Println("Package:", pkg.Name)
        fmt.Println("Path:", pkg.PkgPath)
        fmt.Println("Files:", pkg.GoFiles)

        // Access parsed syntax
        for _, file := range pkg.Syntax {
            ast.Inspect(file, func(n ast.Node) bool {
                // Walk the AST
                return true
            })
        }

        // Access type information
        for ident, obj := range pkg.TypesInfo.Defs {
            fmt.Printf("Defined: %s at %s\n",
                ident.Name, pkg.Fset.Position(ident.Pos()))
        }
    }

  LOAD MODES (performance optimization):

    packages.NeedName       // package name and import path
    packages.NeedFiles      // Go source files
    packages.NeedCompiledGoFiles // compiled files
    packages.NeedImports    // import graph
    packages.NeedDeps       // transitive dependencies
    packages.NeedExportFile // export data
    packages.NeedTypes      // type-checked types
    packages.NeedSyntax     // parsed AST
    packages.NeedTypesInfo  // type info (uses, defs, etc.)

  TIPS:
    - Only request the mode flags you need
    - NeedSyntax + NeedTypes is expensive (full type checking)
    - NeedName + NeedFiles is very fast
    - Use packages.Visit for error handling
`)

	// ============================================
	// GO/SSA
	// ============================================
	fmt.Println("\n--- go/ssa: Static Single Assignment ---")

	os.Stdout.WriteString(`
  SSA (Static Single Assignment) form represents Go programs
  where each variable is assigned exactly once.

  WHY SSA:
    - Enables powerful program analysis
    - Each value has exactly one definition point
    - Control flow is explicit
    - Simplifies data flow analysis

  API:

    import (
        "golang.org/x/tools/go/ssa"
        "golang.org/x/tools/go/ssa/ssautil"
    )

    // Build SSA from packages
    program, pkgs := ssautil.AllPackages(loadedPkgs, ssa.InstantiateGenerics)
    program.Build()

    // Inspect functions
    for _, pkg := range pkgs {
        for _, member := range pkg.Members {
            if fn, ok := member.(*ssa.Function); ok {
                fmt.Println("Function:", fn.Name())
                for _, block := range fn.Blocks {
                    for _, instr := range block.Instrs {
                        fmt.Printf("  %s\n", instr)
                    }
                }
            }
        }
    }

  SSA INSTRUCTIONS:
    *ssa.Alloc     - allocate memory
    *ssa.Call      - function call
    *ssa.Store     - store to address
    *ssa.UnOp      - unary operation
    *ssa.BinOp     - binary operation
    *ssa.Phi       - phi node (merges values from branches)
    *ssa.If        - conditional branch
    *ssa.Return    - return from function
    *ssa.MakeInterface - create interface value

  USE CASES:
    - Dead code detection
    - Nil pointer analysis
    - Escape analysis
    - Data flow tracking
    - Taint analysis
`)

	// ============================================
	// CALL GRAPH
	// ============================================
	fmt.Println("\n--- go/callgraph: Call Graph Analysis ---")

	os.Stdout.WriteString(`
  Build and analyze program call graphs.

    import (
        "golang.org/x/tools/go/callgraph"
        "golang.org/x/tools/go/callgraph/cha"    // Class Hierarchy Analysis
        "golang.org/x/tools/go/callgraph/rta"    // Rapid Type Analysis
        "golang.org/x/tools/go/callgraph/vta"    // Variable Type Analysis
        "golang.org/x/tools/go/callgraph/static" // Static calls only
    )

    // Build call graph
    cg := cha.CallGraph(program)      // most complete
    cg := vta.CallGraph(funcs, cha.CallGraph(program))  // most precise
    cg := static.CallGraph(program)   // fastest, static calls only

    // Traverse call graph
    callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
        caller := edge.Caller.Func.Name()
        callee := edge.Callee.Func.Name()
        fmt.Printf("%s -> %s\n", caller, callee)
        return nil
    })

  ALGORITHM COMPARISON:
    static - Only direct function calls (fastest, least complete)
    cha    - Includes interface dispatches (moderate)
    rta    - Type reachability analysis (better precision)
    vta    - Variable type analysis (best precision, slowest)
`)

	// ============================================
	// DEVELOPER COMMANDS
	// ============================================
	fmt.Println("\n--- Developer Commands ---")

	fmt.Println(`
  GOIMPORTS (essential):

    go install golang.org/x/tools/cmd/goimports@latest

    # Format code AND fix imports
    goimports -w file.go

    # Fix imports for entire project
    goimports -w .

    # Show diff without writing
    goimports -d file.go

    # Use as editor format-on-save
    # (configured in VS Code, GoLand, vim-go, etc.)

  DIFFERENCE FROM gofmt:
    gofmt     - only formats whitespace/style
    goimports - formats + adds/removes imports

  GURU (code exploration):

    go install golang.org/x/tools/cmd/guru@latest

    # Find all callers of a function
    guru callers mypackage.MyFunc

    # Find all implementations of an interface
    guru implements mypackage.MyInterface

    # Show the definition of a symbol
    guru definition file.go:#offset

    # Show the type of an expression
    guru describe file.go:#offset

  GORENAME (safe rename):

    go install golang.org/x/tools/cmd/gorename@latest

    # Rename a function/variable/type
    gorename -from '"mypackage".OldName' -to NewName

    # Preview changes without applying
    gorename -from '"mypackage".OldName' -to NewName -dryrun

    # Type-safe: checks all references across packages
    # Will refuse if rename causes name conflict`)

	// ============================================
	// WORKING EXAMPLE: AST INSPECTION
	// ============================================
	fmt.Println("\n--- Working Example: AST Inspection (stdlib) ---")

	src := `package example

import "fmt"

func greet(name string) {
	fmt.Println("Hello, " + name)
}

func add(a, b int) int {
	return a + b
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", src, parser.ParseComments)
	if err != nil {
		fmt.Println("  Parse error:", err)
		return
	}

	fmt.Println("  Parsed AST for example.go:")
	os.Stdout.WriteString(fmt.Sprintf("    Package: %s\n", file.Name.Name))
	fmt.Println("    Imports:")
	for _, imp := range file.Imports {
		os.Stdout.WriteString(fmt.Sprintf("      %s\n", imp.Path.Value))
	}

	fmt.Println("    Functions:")
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			params := []string{}
			for _, field := range fn.Type.Params.List {
				for _, name := range field.Names {
					params = append(params, name.Name)
				}
			}
			os.Stdout.WriteString(fmt.Sprintf("      %s(%s)\n",
				fn.Name.Name, strings.Join(params, ", ")))
		}
		return true
	})

	// Count AST nodes
	nodeCount := 0
	ast.Inspect(file, func(n ast.Node) bool {
		if n != nil {
			nodeCount++
		}
		return true
	})
	os.Stdout.WriteString(fmt.Sprintf("    Total AST nodes: %d\n", nodeCount))

	// Find function calls
	fmt.Println("    Function calls:")
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			pos := fset.Position(call.Pos())
			var callName string
			switch fn := call.Fun.(type) {
			case *ast.SelectorExpr:
				if ident, ok := fn.X.(*ast.Ident); ok {
					callName = ident.Name + "." + fn.Sel.Name
				}
			case *ast.Ident:
				callName = fn.Name
			}
			if callName != "" {
				os.Stdout.WriteString(fmt.Sprintf("      %s at line %d\n", callName, pos.Line))
			}
		}
		return true
	})

	// ============================================
	// WRITING ANALYSIS TESTS
	// ============================================
	fmt.Println("\n--- Testing Custom Analyzers ---")

	fmt.Println(`
  x/tools provides a test framework for analyzers:

    import "golang.org/x/tools/go/analysis/analysistest"

    func TestMyAnalyzer(t *testing.T) {
        testdata := analysistest.TestData()
        analysistest.Run(t, testdata, MyAnalyzer, "mypackage")
    }

  TEST FILE FORMAT:

    // testdata/src/mypackage/example.go
    package mypackage

    import "fmt"

    func bad() {
        fmt.Println("hello") // want "avoid fmt.Println"
    }

    func good() {
        log.Info("hello") // no diagnostic expected
    }

  The "want" comment syntax:
    // want "regexp matching diagnostic message"

  Multiple diagnostics on same line:
    x := foo() // want "diagnostic 1" "diagnostic 2"`)

	// ============================================
	// INTEGRATION WITH GO VET
	// ============================================
	fmt.Println("\n--- Integration with go vet ---")

	fmt.Println(`
  CUSTOM ANALYZERS WITH GO VET:

    # Build your analyzer as a binary
    go build -o myanalyzer ./cmd/myanalyzer

    # Run via go vet
    go vet -vettool=./myanalyzer ./...

    # Or install and use from PATH
    go install ./cmd/myanalyzer
    go vet -vettool=$(which myanalyzer) ./...

  POPULAR THIRD-PARTY ANALYZERS:
    staticcheck  - Comprehensive Go linter
    errcheck     - Unchecked errors
    ineffassign  - Ineffective assignments
    bodyclose    - Unclosed HTTP response bodies
    nilaway      - Nil pointer dereference detection (Uber)
    govulncheck  - Known vulnerability detection

  GOLANGCI-LINT:
    Runs many analyzers in parallel:
    golangci-lint run ./...

    Configure in .golangci.yml:
      linters:
        enable:
          - staticcheck
          - errcheck
          - govet
          - ineffassign
          - bodyclose`)

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n--- Summary ---")

	fmt.Println(`
  golang.org/x/tools ESSENTIALS:

  go/analysis:
    - Framework for static analyzers (used by go vet)
    - Analyzer struct with Name, Doc, Run, Requires
    - Pass provides AST, types, reporting
    - SuggestedFix for auto-fixable issues

  go/packages:
    - Load packages with configurable detail level
    - Replacement for go/build
    - Full type information and AST access

  go/ssa:
    - Static Single Assignment form
    - Powerful program analysis foundation
    - Dead code, nil analysis, taint tracking

  go/callgraph:
    - Call graph construction (static, CHA, RTA, VTA)
    - Dependency and reachability analysis

  Commands:
    goimports  - Format + fix imports (essential)
    guru       - Code exploration
    gorename   - Type-safe rename

  Install:
    go get golang.org/x/tools/...
    go install golang.org/x/tools/cmd/goimports@latest`)
}
