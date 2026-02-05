// Package main - Capítulo 39: CLI Applications
// Go es excelente para crear herramientas de línea de comandos.
// Aprenderás flag, cobra, y buenas prácticas para CLI.
package main

import (
	"os"
	"flag"
	"fmt"
)

func main() {
	fmt.Println("=== CLI APPLICATIONS EN GO ===")

	// ============================================
	// FLAG PACKAGE (STDLIB)
	// ============================================
	fmt.Println("\n--- flag Package ---")

	// Definir flags
	name := flag.String("name", "World", "Name to greet")
	count := flag.Int("count", 1, "Number of greetings")
	verbose := flag.Bool("verbose", false, "Enable verbose output")

	// También se puede usar punteros existentes
	var timeout int
	flag.IntVar(&timeout, "timeout", 30, "Timeout in seconds")

	// Parsear (normalmente se llama en main real)
	// flag.Parse()

	fmt.Printf("Flags definidos:\n")
	fmt.Printf("  -name string    Name to greet (default %q)\n", *name)
	fmt.Printf("  -count int      Number of greetings (default %d)\n", *count)
	fmt.Printf("  -verbose        Enable verbose output (default %v)\n", *verbose)
	fmt.Printf("  -timeout int    Timeout in seconds (default %d)\n", timeout)

	fmt.Println(`
USO:
./myapp -name="Alice" -count=3 -verbose

FLAG TYPES:
flag.String, flag.Int, flag.Bool, flag.Float64
flag.Duration, flag.Func

ARGUMENTOS POSICIONALES:
flag.Args()  // Argumentos después de los flags
flag.Arg(0)  // Primer argumento
flag.NArg()  // Número de argumentos`)
	// ============================================
	// COBRA
	// ============================================
	fmt.Println("\n--- Cobra ---")
	os.Stdout.WriteString(`
github.com/spf13/cobra - El estándar de facto para CLI en Go

INSTALACIÓN:
go get -u github.com/spf13/cobra/cobra

ESTRUCTURA TÍPICA:
myapp/
├── cmd/
│   ├── root.go      # Comando raíz
│   ├── serve.go     # Subcomando 'serve'
│   ├── migrate.go   # Subcomando 'migrate'
│   └── version.go   # Subcomando 'version'
├── internal/
├── main.go
└── go.mod

// main.go
package main

import "myapp/cmd"

func main() {
    cmd.Execute()
}

// cmd/root.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My awesome CLI application",
    Long:  ` + "`" + `A longer description of your application.` + "`" + `,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    cobra.OnInitialize(initConfig)
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.SetConfigName("config")
        viper.AddConfigPath(".")
        viper.AddConfigPath("$HOME/.myapp")
    }
    viper.AutomaticEnv()
    viper.ReadInConfig()
}

// cmd/serve.go
package cmd

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the server",
    Long:  ` + "`" + `Start the HTTP server on the specified port.` + "`" + `,
    Run: func(cmd *cobra.Command, args []string) {
        port, _ := cmd.Flags().GetInt("port")
        fmt.Printf("Starting server on port %d\n", port)
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)
    serveCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
}
`)

	// ============================================
	// URFAVE/CLI
	// ============================================
	fmt.Println("\n--- urfave/cli ---")
	os.Stdout.WriteString(`
github.com/urfave/cli/v2 - Alternativa más simple a Cobra

package main

import (
    "fmt"
    "os"

    "github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Name:  "myapp",
        Usage: "My awesome application",
        Flags: []cli.Flag{
            &cli.StringFlag{
                Name:    "config",
                Aliases: []string{"c"},
                Usage:   "Load configuration from ` + "`" + `FILE` + "`" + `",
            },
            &cli.BoolFlag{
                Name:    "verbose",
                Aliases: []string{"v"},
                Usage:   "Enable verbose output",
            },
        },
        Commands: []*cli.Command{
            {
                Name:    "serve",
                Aliases: []string{"s"},
                Usage:   "Start the server",
                Flags: []cli.Flag{
                    &cli.IntFlag{
                        Name:  "port",
                        Value: 8080,
                        Usage: "Port to listen on",
                    },
                },
                Action: func(c *cli.Context) error {
                    port := c.Int("port")
                    fmt.Printf("Starting server on port %d\n", port)
                    return nil
                },
            },
            {
                Name:  "migrate",
                Usage: "Run database migrations",
                Action: func(c *cli.Context) error {
                    fmt.Println("Running migrations...")
                    return nil
                },
            },
        },
    }

    if err := app.Run(os.Args); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
`)

	// ============================================
	// BUBBLETEA (TUI)
	// ============================================
	fmt.Println("\n--- Bubbletea (TUI) ---")
	os.Stdout.WriteString(`
github.com/charmbracelet/bubbletea - Terminal UI framework

package main

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
)

type model struct {
    choices  []string
    cursor   int
    selected map[int]struct{}
}

func initialModel() model {
    return model{
        choices:  []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},
        selected: make(map[int]struct{}),
    }
}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.choices)-1 {
                m.cursor++
            }
        case "enter", " ":
            if _, ok := m.selected[m.cursor]; ok {
                delete(m.selected, m.cursor)
            } else {
                m.selected[m.cursor] = struct{}{}
            }
        }
    }
    return m, nil
}

func (m model) View() string {
    s := "What should we buy?\n\n"
    for i, choice := range m.choices {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        checked := " "
        if _, ok := m.selected[i]; ok {
            checked = "x"
        }
        s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
    }
    s += "\nPress q to quit.\n"
    return s
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error: %v", err)
        os.Exit(1)
    }
}
`)

	// ============================================
	// INPUT/OUTPUT
	// ============================================
	fmt.Println("\n--- Input/Output ---")
	fmt.Println(`
// Leer de stdin
var input string
fmt.Scanln(&input)

// Bufio para líneas completas
reader := bufio.NewReader(os.Stdin)
line, _ := reader.ReadString('\n')

// Verificar si es terminal interactivo
if terminal.IsTerminal(int(os.Stdin.Fd())) {
    // Modo interactivo
} else {
    // Piped input
}

// Colores (con fatih/color)
import "github.com/fatih/color"

red := color.New(color.FgRed).PrintlnFunc()
red("Error: something went wrong")

green := color.New(color.FgGreen, color.Bold)
green.Println("Success!")

// Progress bars (con schollz/progressbar)
import "github.com/schollz/progressbar/v3"

bar := progressbar.Default(100)
for i := 0; i < 100; i++ {
    bar.Add(1)
    time.Sleep(40 * time.Millisecond)
}

// Spinners (con briandowns/spinner)
import "github.com/briandowns/spinner"

s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
s.Start()
time.Sleep(4 * time.Second)
s.Stop()`)
	// ============================================
	// EXIT CODES
	// ============================================
	fmt.Println("\n--- Exit Codes ---")
	os.Stdout.WriteString(`
CONVENCIONES:

0  - Éxito
1  - Error general
2  - Uso incorrecto (bad arguments)
64 - Comando no encontrado (EX_USAGE)
65 - Error de datos (EX_DATAERR)
66 - No input (EX_NOINPUT)
74 - Error de I/O (EX_IOERR)
78 - Error de configuración (EX_CONFIG)

// Salir con código
os.Exit(1)

// Mejor: retornar error y manejar en main
func run() error {
    if err := doSomething(); err != nil {
        return err
    }
    return nil
}

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
`)

	// ============================================
	// SEÑALES
	// ============================================
	fmt.Println("\n--- Señales ---")
	fmt.Println(`
import (
    "os"
    "os/signal"
    "syscall"
)

// Capturar señales para cleanup
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Capturar SIGINT y SIGTERM
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigCh
        fmt.Println("\nReceived shutdown signal")
        cancel()
    }()

    // Tu aplicación usa ctx
    if err := run(ctx); err != nil {
        os.Exit(1)
    }
}

// signal.NotifyContext (Go 1.16+)
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()`)
	// ============================================
	// CONFIGURACIÓN
	// ============================================
	fmt.Println("\n--- Configuración ---")
	fmt.Println(`
ORDEN DE PRIORIDAD (típico):
1. Flags de línea de comandos
2. Variables de entorno
3. Archivo de configuración
4. Valores por defecto

// Con Viper
viper.SetDefault("port", 8080)
viper.SetConfigName("config")
viper.AddConfigPath(".")
viper.AutomaticEnv()
viper.ReadInConfig()

// Bindings
viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))
viper.BindEnv("port", "APP_PORT")

// Obtener valor
port := viper.GetInt("port")

// Con envconfig (más simple)
import "github.com/kelseyhightower/envconfig"

type Config struct {
    Port    int           ` + "`" + `envconfig:"PORT" default:"8080"` + "`" + `
    Debug   bool          ` + "`" + `envconfig:"DEBUG"` + "`" + `
    Timeout time.Duration ` + "`" + `envconfig:"TIMEOUT" default:"30s"` + "`" + `
}

var cfg Config
envconfig.Process("MYAPP", &cfg)`)
	// ============================================
	// BUENAS PRÁCTICAS
	// ============================================
	fmt.Println("\n--- Buenas Prácticas ---")
	os.Stdout.WriteString(`
1. STDERR para errores, STDOUT para output
   fmt.Fprintf(os.Stderr, "Error: %v\n", err)

2. Exit codes consistentes

3. Help text claro y completo
   -h, --help siempre disponible

4. Versión fácil de obtener
   myapp --version
   myapp version

5. Quiet mode para scripts
   myapp -q  # Sin output excepto errores

6. Verbose mode para debugging
   myapp -v  # Output detallado

7. Graceful shutdown en señales

8. Configuración desde múltiples fuentes

9. Autocompletado de shell
   myapp completion bash > /etc/bash_completion.d/myapp

10. Man pages (con cobra-man)

11. Progress indicators para operaciones largas

12. Confirmación para operaciones destructivas
    "Are you sure? [y/N]"
`)
}

/*
RESUMEN DE CLI:

STDLIB:
flag.String(), flag.Int(), flag.Bool()
flag.Parse()
flag.Args()

COBRA:
rootCmd = &cobra.Command{...}
rootCmd.AddCommand(subCmd)
rootCmd.PersistentFlags()
cmd.Flags()

URFAVE/CLI:
app := &cli.App{Commands: [...]}
app.Run(os.Args)

TUI:
Bubbletea: tea.Model interface
Init(), Update(), View()

I/O:
os.Stdin, os.Stdout, os.Stderr
bufio.Reader
fatih/color para colores
progressbar para barras

SEÑALES:
signal.Notify(ch, syscall.SIGINT)
signal.NotifyContext(ctx, ...)

CONFIG:
1. Flags
2. Env vars
3. Config files
4. Defaults

BUENAS PRÁCTICAS:
- stderr para errores
- Exit codes consistentes
- Help y version
- Graceful shutdown
- Autocompletado
*/
