// Package main - Chapter 78: OS, Exec, and Process Management
// Go tiene excelente soporte para interactuar con el sistema operativo:
// ejecutar comandos externos, manejar procesos, variables de entorno,
// senales, y operaciones del filesystem avanzadas.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func main() {
	fmt.Println("=== OS, EXEC, AND PROCESS MANAGEMENT ===")

	// ============================================
	// EJECUTAR COMANDOS EXTERNOS
	// ============================================
	fmt.Println("\n--- exec.Command Basico ---")

	// Comando simple
	out, err := exec.Command("echo", "Hello from Go").Output()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Output: %s", out)
	}

	// Verificar si un comando existe
	path, err := exec.LookPath("go")
	if err != nil {
		fmt.Println("go no encontrado en PATH")
	} else {
		fmt.Printf("go encontrado en: %s\n", path)
	}

	// ============================================
	// CAPTURAR STDOUT Y STDERR
	// ============================================
	fmt.Println("\n--- Stdout y Stderr ---")

	cmd := exec.Command("go", "version")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\nStderr: %s\n", err, stderr.String())
	} else {
		fmt.Printf("Go version: %s", stdout.String())
	}

	// CombinedOutput captura ambos
	combined, err := exec.Command("go", "env", "GOPATH").CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("GOPATH: %s", combined)
	}

	// ============================================
	// EXEC CON CONTEXT (TIMEOUT)
	// ============================================
	fmt.Println("\n--- Command con Context ---")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, "sleep", "1")
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Comando excedio timeout!")
	} else if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Comando completado dentro del timeout")
	}

	fmt.Println(`
TIMEOUT PATTERNS:

// Con context
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "long-running-task")

// Go 1.20+: WaitDelay para cleanup graceful
cmd.WaitDelay = 5 * time.Second
// Envia SIGKILL despues de WaitDelay si no termina

// Go 1.20+: Cancel customizable
cmd.Cancel = func() error {
    return cmd.Process.Signal(syscall.SIGTERM)
}`)

	// ============================================
	// STDIN: ENVIAR INPUT A COMANDOS
	// ============================================
	fmt.Println("\n--- Stdin a Comandos ---")

	cmd = exec.Command("tr", "a-z", "A-Z")
	cmd.Stdin = strings.NewReader("hello world from go")
	var trOutput bytes.Buffer
	cmd.Stdout = &trOutput
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("tr output: %s\n", trOutput.String())
	}

	fmt.Println(`
STDIN PATTERNS:

// String como stdin
cmd.Stdin = strings.NewReader("input data")

// Pipe para escribir progresivamente
stdin, _ := cmd.StdinPipe()
go func() {
    defer stdin.Close()
    fmt.Fprintln(stdin, "line 1")
    fmt.Fprintln(stdin, "line 2")
}()

// Archivo como stdin
f, _ := os.Open("input.txt")
defer f.Close()
cmd.Stdin = f`)

	// ============================================
	// PIPES: CONECTAR COMANDOS
	// ============================================
	fmt.Println("\n--- Pipes entre Comandos ---")

	demonstratePipes()

	fmt.Println(`
PIPE PATTERNS:

// cmd1 | cmd2
cmd1 := exec.Command("echo", "hello world")
cmd2 := exec.Command("wc", "-w")

pipe, _ := cmd1.StdoutPipe()
cmd2.Stdin = pipe

cmd1.Start()
output, _ := cmd2.Output()
cmd1.Wait()

// cmd1 | cmd2 | cmd3 (cadena larga)
// Mejor usar shell:
exec.Command("sh", "-c", "cat file | grep pattern | wc -l")
// CUIDADO: solo con inputs confiables (no user input!)`)

	// ============================================
	// ENVIRONMENT VARIABLES
	// ============================================
	fmt.Println("\n--- Environment Variables ---")

	// Leer
	home := os.Getenv("HOME")
	fmt.Printf("HOME: %s\n", home)

	// Verificar si existe (Getenv retorna "" si no existe)
	gopath, exists := os.LookupEnv("GOPATH")
	if exists {
		fmt.Printf("GOPATH: %s\n", gopath)
	} else {
		fmt.Println("GOPATH no definido")
	}

	// Setear (solo afecta el proceso actual y sus hijos)
	os.Setenv("MY_APP_MODE", "debug")
	fmt.Printf("MY_APP_MODE: %s\n", os.Getenv("MY_APP_MODE"))
	os.Unsetenv("MY_APP_MODE")

	// Pasar environment a comandos
	cmd = exec.Command("sh", "-c", "echo $CUSTOM_VAR")
	cmd.Env = append(os.Environ(), "CUSTOM_VAR=hello_from_go")
	envOut, _ := cmd.Output()
	fmt.Printf("Custom env var: %s", envOut)

	os.Stdout.WriteString(`
ENVIRONMENT PATTERNS:

// Inherit todo + agregar
cmd.Env = append(os.Environ(), "KEY=value")

// Environment limpio (solo lo que defines)
cmd.Env = []string{"PATH=/usr/bin", "HOME=/tmp"}

// Expandir variables en strings
expanded := os.ExpandEnv("Home is $HOME, user is ${USER}")

// Listar todo el environment
for _, env := range os.Environ() {
    pair := strings.SplitN(env, "=", 2)
    fmt.Printf("%s = %s\n", pair[0], pair[1])
}
`)

	// ============================================
	// WORKING DIRECTORY
	// ============================================
	fmt.Println("\n--- Working Directory ---")

	cwd, _ := os.Getwd()
	fmt.Printf("Current dir: %s\n", cwd)

	// Ejecutar comando en directorio especifico
	cmd = exec.Command("ls", "-la")
	cmd.Dir = "/tmp"
	// cmd.Run()

	fmt.Println(`
DIRECTORY OPERATIONS:

os.Getwd()                  // Current working directory
os.Chdir("/path")           // Cambiar (afecta todo el proceso!)
os.UserHomeDir()            // Home del usuario
os.UserCacheDir()           // Cache dir (~/.cache en Linux)
os.UserConfigDir()          // Config dir (~/.config en Linux)
os.TempDir()                // Directorio temporal

// Para comandos, usar cmd.Dir en vez de os.Chdir
cmd := exec.Command("make", "build")
cmd.Dir = "/path/to/project"`)

	// ============================================
	// TEMPORARY FILES AND DIRECTORIES
	// ============================================
	fmt.Println("\n--- Temp Files and Dirs ---")

	// Archivo temporal
	tmpFile, err := os.CreateTemp("", "myapp-*.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Temp file: %s\n", tmpFile.Name())
		tmpFile.WriteString("temporary data")
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}

	// Directorio temporal
	tmpDir, err := os.MkdirTemp("", "myapp-*")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Temp dir: %s\n", tmpDir)
		os.RemoveAll(tmpDir)
	}

	fmt.Println(`
TEMP PATTERNS:

// Patron: "" usa os.TempDir(), el patron * es reemplazado
os.CreateTemp("", "prefix-*.suffix")  // /tmp/prefix-123456.suffix
os.CreateTemp("/custom", "app-*")     // /custom/app-123456

os.MkdirTemp("", "myapp-*")          // /tmp/myapp-123456/

// Siempre limpiar:
defer os.Remove(tmpFile.Name())
defer os.RemoveAll(tmpDir)`)

	// ============================================
	// FILEPATH OPERATIONS
	// ============================================
	fmt.Println("\n--- filepath Package ---")

	fmt.Printf("Join: %s\n", filepath.Join("home", "user", "docs", "file.txt"))
	fmt.Printf("Dir: %s\n", filepath.Dir("/home/user/file.txt"))
	fmt.Printf("Base: %s\n", filepath.Base("/home/user/file.txt"))
	fmt.Printf("Ext: %s\n", filepath.Ext("document.tar.gz"))
	fmt.Printf("Clean: %s\n", filepath.Clean("/home//user/../user/./docs"))

	abs, _ := filepath.Abs("relative/path")
	fmt.Printf("Abs: %s\n", abs)

	rel, _ := filepath.Rel("/home/user", "/home/user/docs/file.txt")
	fmt.Printf("Rel: %s\n", rel)

	fmt.Println(`
FILEPATH PATTERNS:

filepath.Join(parts...)        // Une con separador del OS
filepath.Dir(path)             // Directorio padre
filepath.Base(path)            // Nombre del archivo
filepath.Ext(path)             // Extension (.txt)
filepath.Clean(path)           // Normaliza path
filepath.Abs(path)             // Path absoluto
filepath.Rel(base, target)     // Path relativo
filepath.Split(path)           // (dir, file)
filepath.Match(pattern, name)  // Glob matching
filepath.Glob(pattern)         // Find files by pattern
filepath.EvalSymlinks(path)    // Resolver symlinks

// Walk del filesystem
filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
    if err != nil { return err }
    if !d.IsDir() && filepath.Ext(path) == ".go" {
        fmt.Println(path)
    }
    return nil
})`)

	// ============================================
	// PROCESS INFO
	// ============================================
	fmt.Println("\n--- Process Information ---")

	fmt.Printf("PID: %d\n", os.Getpid())
	fmt.Printf("PPID: %d\n", os.Getppid())
	fmt.Printf("UID: %d\n", os.Getuid())
	fmt.Printf("GID: %d\n", os.Getgid())
	fmt.Printf("Hostname: ")
	if hostname, err := os.Hostname(); err == nil {
		fmt.Println(hostname)
	}
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())
	fmt.Printf("GOOS: %s\n", runtime.GOOS)
	fmt.Printf("GOARCH: %s\n", runtime.GOARCH)

	// ============================================
	// SIGNALS
	// ============================================
	fmt.Println("\n--- Signal Handling ---")

	demonstrateSignals()

	fmt.Println(`
SIGNALS PATTERNS:

GRACEFUL SHUTDOWN:
  ctx, stop := signal.NotifyContext(context.Background(),
      syscall.SIGINT, syscall.SIGTERM)
  defer stop()

  // Iniciar servidor
  srv := &http.Server{Addr: ":8080"}
  go srv.ListenAndServe()

  // Esperar signal
  <-ctx.Done()
  log.Println("Shutting down...")

  // Cleanup con timeout
  shutdownCtx, cancel := context.WithTimeout(
      context.Background(), 30*time.Second)
  defer cancel()
  srv.Shutdown(shutdownCtx)

SIGNALS COMUNES:
  SIGINT  (Ctrl+C)    - Interrupt
  SIGTERM             - Terminate (default de kill)
  SIGHUP              - Hangup (reload config)
  SIGUSR1/SIGUSR2     - User defined
  SIGQUIT             - Quit + core dump

IGNORAR SIGNALS:
  signal.Ignore(syscall.SIGHUP)

RESET A DEFAULT:
  signal.Reset(syscall.SIGINT)`)

	// ============================================
	// FILE LOCKING
	// ============================================
	fmt.Println("\n--- File Locking ---")
	os.Stdout.WriteString(`
FILE LOCKING (para evitar multiples instancias):

func acquireLock(path string) (*os.File, error) {
    f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        return nil, err
    }

    // Intentar lock exclusivo (no bloqueante)
    err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
    if err != nil {
        f.Close()
        return nil, fmt.Errorf("otra instancia esta corriendo")
    }

    // Escribir PID
    fmt.Fprintf(f, "%d\n", os.Getpid())
    return f, nil
}

func releaseLock(f *os.File) {
    syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
    f.Close()
    os.Remove(f.Name())
}

USO:
  lock, err := acquireLock("/var/run/myapp.lock")
  if err != nil {
      log.Fatalf("Cannot start: %%v", err)
  }
  defer releaseLock(lock)
`)

	// ============================================
	// PROCESS START/MANAGEMENT
	// ============================================
	fmt.Println("\n--- Process Management ---")
	os.Stdout.WriteString(`
EXEC.CMD LIFECYCLE:
  cmd := exec.Command("server", "--port=8080")

  // Configurar antes de Start
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  cmd.Dir = "/app"
  cmd.Env = append(os.Environ(), "MODE=prod")

  // SysProcAttr para control avanzado (Linux)
  cmd.SysProcAttr = &syscall.SysProcAttr{
      Setpgid: true,  // Nuevo process group
  }

  // Start (no bloqueante)
  cmd.Start()
  pid := cmd.Process.Pid

  // Esperar (bloqueante)
  err := cmd.Wait()

  // O Run = Start + Wait
  err := cmd.Run()

MATAR PROCESO:
  cmd.Process.Signal(syscall.SIGTERM) // Graceful
  cmd.Process.Kill()                   // SIGKILL (forzoso)

  // Matar process group completo (Linux)
  syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)

EXIT:
  os.Exit(0)   // Exito (defers NO se ejecutan!)
  os.Exit(1)   // Error

  // Mejor: retornar de main
  func main() {
      if err := run(); err != nil {
          fmt.Fprintf(os.Stderr, "error: %%v\n", err)
          os.Exit(1)
      }
  }
`)

	// ============================================
	// CROSS-PLATFORM CONSIDERATIONS
	// ============================================
	fmt.Println("\n--- Cross-Platform ---")
	fmt.Println(`
DIFERENCIAS POR OS:

PATH SEPARATOR:
  filepath.Join()  // Usa \ en Windows, / en Unix
  os.PathSeparator // '/' o '\'

LINE ENDINGS:
  // Windows: \r\n, Unix: \n
  // Usar strings.TrimSpace o bufio.Scanner

COMMANDS:
  if runtime.GOOS == "windows" {
      cmd = exec.Command("cmd", "/c", "dir")
  } else {
      cmd = exec.Command("sh", "-c", "ls -la")
  }

SIGNALS:
  // SIGTERM, SIGUSR1 no existen en Windows
  // Usar build tags para platform-specific code

PERMISSIONS:
  // Windows no tiene chmod estilo Unix
  os.Chmod("file", 0755) // No-op en Windows

ENVIRONMENT:
  // Windows: case-insensitive env vars
  // Unix: case-sensitive`)

	// ============================================
	// SECURITY
	// ============================================
	fmt.Println("\n--- Security with exec ---")
	fmt.Println(`
COMMAND INJECTION PREVENTION:

// PELIGROSO: user input en shell
userInput := "file; rm -rf /"
exec.Command("sh", "-c", "cat "+userInput)  // INJECTION!

// SEGURO: argumentos separados (no pasan por shell)
exec.Command("cat", userInput)  // userInput es UN argumento

// SEGURO: si necesitas shell, sanitizar
exec.Command("sh", "-c", "cat -- "+shellescape(userInput))

// MEJOR: evitar shell completamente
cmd := exec.Command("grep", "-r", pattern, directory)
// Cada argumento es un parametro separado del exec

PRINCIPIOS:
1. NUNCA concatenar user input en strings de shell
2. Preferir argumentos separados sobre "sh -c"
3. Validar y sanitizar todo input externo
4. Usar filepath.Clean para paths
5. Limitar PATH si ejecutas binarios externos`)

	fmt.Println("\n=== FIN CAPITULO 78: OS, EXEC, AND PROCESS MANAGEMENT ===")
}

// demonstratePipes muestra como conectar comandos via pipes
func demonstratePipes() {
	// echo "hello world foo bar" | wc -w
	cmd1 := exec.Command("echo", "hello world foo bar")
	cmd2 := exec.Command("wc", "-w")

	// Conectar stdout de cmd1 a stdin de cmd2
	pipe, err := cmd1.StdoutPipe()
	if err != nil {
		fmt.Printf("Error creating pipe: %v\n", err)
		return
	}
	cmd2.Stdin = pipe

	// Capturar output de cmd2
	var output bytes.Buffer
	cmd2.Stdout = &output

	// Iniciar ambos
	if err := cmd1.Start(); err != nil {
		fmt.Printf("Error starting cmd1: %v\n", err)
		return
	}
	if err := cmd2.Start(); err != nil {
		fmt.Printf("Error starting cmd2: %v\n", err)
		return
	}

	cmd1.Wait()
	cmd2.Wait()

	fmt.Printf("Word count: %s", output.String())
}

// demonstrateSignals muestra signal handling basico
func demonstrateSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR1)

	// Simular recepcion
	fmt.Println("Signal handler registrado para SIGUSR1")
	fmt.Printf("Enviar signal: kill -USR1 %d\n", os.Getpid())

	signal.Stop(sigCh)
	fmt.Println("Signal handler removido")
}

// Asegurar que imports no marcados como unused se usan
var (
	_ = io.Copy
	_ = context.Background
	_ = strings.NewReader
	_ = time.Second
	_ = syscall.SIGTERM
)
