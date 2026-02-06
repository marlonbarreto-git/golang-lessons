// Package main - Chapter 048: Flag Package Deep Dive
// Command-line flag parsing with flag.String, flag.Int, flag.Bool, custom Value
// interface, FlagSet for subcommands, usage customization, and CLI patterns.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// ============================================
// CUSTOM FLAG TYPES
// ============================================

// StringSlice implements flag.Value for repeated string flags
type StringSlice []string

func (s *StringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *StringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// IPAddress implements flag.Value for IP validation
type IPAddress struct {
	IP net.IP
}

func (ip *IPAddress) String() string {
	if ip.IP == nil {
		return ""
	}
	return ip.IP.String()
}

func (ip *IPAddress) Set(value string) error {
	parsed := net.ParseIP(value)
	if parsed == nil {
		return fmt.Errorf("invalid IP address: %q", value)
	}
	ip.IP = parsed
	return nil
}

// LogLevel implements flag.Value for validated enum
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

func (l *LogLevel) String() string {
	switch *l {
	case LogDebug:
		return "debug"
	case LogInfo:
		return "info"
	case LogWarn:
		return "warn"
	case LogError:
		return "error"
	default:
		return "unknown"
	}
}

func (l *LogLevel) Set(value string) error {
	switch strings.ToLower(value) {
	case "debug":
		*l = LogDebug
	case "info":
		*l = LogInfo
	case "warn":
		*l = LogWarn
	case "error":
		*l = LogError
	default:
		return fmt.Errorf("invalid log level %q: must be debug|info|warn|error", value)
	}
	return nil
}

func main() {
	fmt.Println("=== FLAG PACKAGE DEEP DIVE ===")

	// ============================================
	// 1. FLAG BASICS
	// ============================================
	fmt.Println("\n--- 1. Flag Basics ---")
	fmt.Println(`
The flag package provides command-line flag parsing.

DEFINE FLAGS (two styles):

  Style 1 - Returns pointer:
    name := flag.String("name", "default", "description")
    port := flag.Int("port", 8080, "server port")
    debug := flag.Bool("debug", false, "enable debug")

  Style 2 - Bind to variable:
    var name string
    flag.StringVar(&name, "name", "default", "description")

BASIC TYPES:
    flag.String / flag.StringVar
    flag.Int / flag.IntVar
    flag.Int64 / flag.Int64Var
    flag.Uint / flag.UintVar
    flag.Uint64 / flag.Uint64Var
    flag.Float64 / flag.Float64Var
    flag.Bool / flag.BoolVar
    flag.Duration / flag.DurationVar

PARSE:
    flag.Parse()    // must call after defining, before using

USAGE:
    -flag value     // standard
    -flag=value     // with equals
    --flag value    // double dash also works
    -flag           // for bool flags, sets to true`)

	// Simulate flag parsing with FlagSet (to not interfere with actual args)
	demoFlags := flag.NewFlagSet("demo", flag.ContinueOnError)
	name := demoFlags.String("name", "World", "person to greet")
	port := demoFlags.Int("port", 8080, "server port")
	debug := demoFlags.Bool("debug", false, "enable debug mode")
	timeout := demoFlags.Duration("timeout", 30*time.Second, "request timeout")
	ratio := demoFlags.Float64("ratio", 0.75, "compression ratio")

	// Parse simulated args
	args := []string{"-name", "Gopher", "-port", "3000", "-debug", "-timeout", "5s"}
	demoFlags.Parse(args)

	fmt.Printf("\n  Parsed flags:\n")
	fmt.Printf("    name:    %q\n", *name)
	fmt.Printf("    port:    %d\n", *port)
	fmt.Printf("    debug:   %v\n", *debug)
	fmt.Printf("    timeout: %v\n", *timeout)
	fmt.Printf("    ratio:   %f (default, not set)\n", *ratio)

	// Remaining args
	fmt.Printf("    remaining args: %v\n", demoFlags.Args())

	// ============================================
	// 2. FLAG.ARGS AND POSITIONAL ARGUMENTS
	// ============================================
	fmt.Println("\n--- 2. flag.Args and Positional Arguments ---")
	fmt.Println(`
After flag.Parse(), remaining non-flag arguments are available via:

  flag.Args()     -> []string of remaining args
  flag.Arg(i)     -> i-th remaining arg (empty string if out of range)
  flag.NArg()     -> number of remaining args

Flags stop at first non-flag argument or "--":
  cmd -flag value arg1 arg2   -> Args() = ["arg1", "arg2"]
  cmd -flag value -- -notflag -> Args() = ["-notflag"]`)

	positional := flag.NewFlagSet("positional", flag.ContinueOnError)
	verbose := positional.Bool("v", false, "verbose")
	positional.Parse([]string{"-v", "file1.txt", "file2.txt", "file3.txt"})

	fmt.Printf("\n  verbose: %v\n", *verbose)
	fmt.Printf("  NArg: %d\n", positional.NArg())
	fmt.Printf("  Args: %v\n", positional.Args())
	fmt.Printf("  Arg(0): %q\n", positional.Arg(0))
	fmt.Printf("  Arg(1): %q\n", positional.Arg(1))

	// Double-dash to stop flag parsing
	doubleDash := flag.NewFlagSet("dd", flag.ContinueOnError)
	doubleDash.Bool("flag", false, "a flag")
	doubleDash.Parse([]string{"-flag", "--", "-not-a-flag", "arg"})
	fmt.Printf("\n  After '--': %v\n", doubleDash.Args())

	// ============================================
	// 3. CUSTOM FLAG.VALUE INTERFACE
	// ============================================
	fmt.Println("\n--- 3. Custom flag.Value Interface ---")
	fmt.Println(`
Any type implementing flag.Value can be used with flag.Var:

  type Value interface {
      String() string       // default value representation
      Set(string) error     // parse and set the value
  }

This allows custom validation, complex types, and repeated flags.`)

	// StringSlice - repeated flag
	fmt.Println("\n  StringSlice (repeated flag):")
	sliceFlags := flag.NewFlagSet("slice", flag.ContinueOnError)
	var tags StringSlice
	sliceFlags.Var(&tags, "tag", "add a tag (can repeat)")
	sliceFlags.Parse([]string{"-tag", "go", "-tag", "tutorial", "-tag", "stdlib"})
	fmt.Printf("    tags: %v\n", []string(tags))

	// IPAddress - validated flag
	fmt.Println("\n  IPAddress (validated flag):")
	ipFlags := flag.NewFlagSet("ip", flag.ContinueOnError)
	var addr IPAddress
	ipFlags.Var(&addr, "bind", "bind address")
	ipFlags.Parse([]string{"-bind", "192.168.1.100"})
	fmt.Printf("    bind: %s\n", addr.IP)

	// Invalid IP
	ipFlags2 := flag.NewFlagSet("ip2", flag.ContinueOnError)
	ipFlags2.SetOutput(os.Stderr)
	var addr2 IPAddress
	ipFlags2.Var(&addr2, "bind", "bind address")
	err := ipFlags2.Parse([]string{"-bind", "not.an.ip"})
	fmt.Printf("    invalid IP error: %v\n", err)

	// LogLevel - enum flag
	fmt.Println("\n  LogLevel (enum flag):")
	levelFlags := flag.NewFlagSet("level", flag.ContinueOnError)
	var level LogLevel = LogInfo
	levelFlags.Var(&level, "level", "log level (debug|info|warn|error)")
	levelFlags.Parse([]string{"-level", "warn"})
	fmt.Printf("    level: %s\n", level.String())

	// ============================================
	// 4. FLAG.FLAGSET FOR SUBCOMMANDS
	// ============================================
	fmt.Println("\n--- 4. flag.FlagSet for Subcommands ---")
	fmt.Println(`
flag.FlagSet allows creating independent flag sets for subcommands.

  fs := flag.NewFlagSet("name", errorHandling)

Error handling modes:
  flag.ContinueOnError -> return error
  flag.ExitOnError     -> os.Exit(2)
  flag.PanicOnError    -> panic

PATTERN FOR SUBCOMMANDS:
  switch os.Args[1] {
  case "serve":
      serveFlags.Parse(os.Args[2:])
  case "migrate":
      migrateFlags.Parse(os.Args[2:])
  }`)

	// Simulate subcommand: serve
	fmt.Println("\n  Subcommand 'serve':")
	serveFlags := flag.NewFlagSet("serve", flag.ContinueOnError)
	serveHost := serveFlags.String("host", "localhost", "bind host")
	servePort := serveFlags.Int("port", 8080, "bind port")
	serveTLS := serveFlags.Bool("tls", false, "enable TLS")
	serveFlags.Parse([]string{"-host", "0.0.0.0", "-port", "443", "-tls"})
	fmt.Printf("    host=%s, port=%d, tls=%v\n", *serveHost, *servePort, *serveTLS)

	// Simulate subcommand: migrate
	fmt.Println("\n  Subcommand 'migrate':")
	migrateFlags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	migrateDir := migrateFlags.String("dir", "./migrations", "migrations directory")
	migrateSteps := migrateFlags.Int("steps", 0, "number of steps (0=all)")
	migrateDry := migrateFlags.Bool("dry-run", false, "dry run mode")
	migrateFlags.Parse([]string{"-dir", "/app/db/migrations", "-steps", "3", "-dry-run"})
	fmt.Printf("    dir=%s, steps=%d, dry-run=%v\n", *migrateDir, *migrateSteps, *migrateDry)

	// Full subcommand dispatcher example
	fmt.Println("\n  Subcommand dispatcher pattern:")
	simulatedArgs := [][]string{
		{"serve", "-port", "9090"},
		{"version"},
		{"help"},
		{"unknown"},
	}

	for _, cmdArgs := range simulatedArgs {
		result := dispatchCommand(cmdArgs)
		fmt.Printf("    %v -> %s\n", cmdArgs, result)
	}

	// ============================================
	// 5. FLAG.USAGE CUSTOMIZATION
	// ============================================
	fmt.Println("\n--- 5. flag.Usage Customization ---")
	fmt.Println(`
Customize the usage/help message:

  flag.Usage = func() {
      fmt.Fprintf(flag.CommandLine.Output(), "Usage: mycmd [flags]\n\n")
      flag.PrintDefaults()
  }

For FlagSet:
  fs.SetOutput(w)   // redirect output
  fs.Usage = func() { ... }`)

	customFlags := flag.NewFlagSet("myapp", flag.ContinueOnError)
	var usageOutput strings.Builder
	customFlags.SetOutput(&usageOutput)
	customFlags.String("config", "config.yaml", "path to config file")
	customFlags.Int("workers", 4, "number of worker goroutines")
	customFlags.Bool("verbose", false, "enable verbose logging")

	customFlags.Usage = func() {
		fmt.Fprintf(customFlags.Output(), "myapp - A sample application\n\n")
		fmt.Fprintf(customFlags.Output(), "Usage:\n  myapp [flags] [files...]\n\nFlags:\n")
		customFlags.PrintDefaults()
	}

	customFlags.Usage()
	fmt.Printf("\n  Custom usage output:\n%s\n", usageOutput.String())

	// ============================================
	// 6. FLAG.VISIT AND VISITALL
	// ============================================
	fmt.Println("--- 6. flag.Visit and VisitAll ---")
	fmt.Println(`
  Visit(fn)    -> iterate over flags that were SET by user
  VisitAll(fn) -> iterate over ALL defined flags

Useful for detecting which flags were explicitly provided.`)

	visitFlags := flag.NewFlagSet("visit", flag.ContinueOnError)
	visitFlags.String("name", "default", "user name")
	visitFlags.Int("port", 8080, "port")
	visitFlags.Bool("debug", false, "debug")
	visitFlags.Parse([]string{"-name", "Gopher", "-debug"})

	fmt.Println("\n  Flags SET by user (Visit):")
	visitFlags.Visit(func(f *flag.Flag) {
		fmt.Printf("    -%s = %q (default: %q)\n", f.Name, f.Value.String(), f.DefValue)
	})

	fmt.Println("\n  ALL defined flags (VisitAll):")
	visitFlags.VisitAll(func(f *flag.Flag) {
		isSet := false
		visitFlags.Visit(func(sf *flag.Flag) {
			if sf.Name == f.Name {
				isSet = true
			}
		})
		fmt.Printf("    -%s = %q (set: %v)\n", f.Name, f.Value.String(), isSet)
	})

	// ============================================
	// 7. ENVIRONMENT VARIABLE FALLBACK PATTERN
	// ============================================
	fmt.Println("\n--- 7. Environment Variable Fallback Pattern ---")
	fmt.Println(`
Common pattern: flag > env var > default value.

Steps:
1. Define flag with default
2. Check if env var is set, use as default
3. Flag value overrides env var if specified`)

	fmt.Println("\n  EnvOrDefault pattern:")
	envFlags := flag.NewFlagSet("env", flag.ContinueOnError)

	// Simulate env vars
	envVars := map[string]string{
		"APP_HOST": "env-host.example.com",
		"APP_PORT": "9999",
	}

	hostDefault := "localhost"
	if v, ok := envVars["APP_HOST"]; ok {
		hostDefault = v
	}
	portDefault := 8080
	if _, ok := envVars["APP_PORT"]; ok {
		portDefault = 9999
	}

	envHost := envFlags.String("host", hostDefault, "server host (env: APP_HOST)")
	envPort := envFlags.Int("port", portDefault, "server port (env: APP_PORT)")

	// No flags set -> uses env defaults
	envFlags.Parse([]string{})
	fmt.Printf("    No flags: host=%q, port=%d (from env)\n", *envHost, *envPort)

	// Flag overrides env
	envFlags2 := flag.NewFlagSet("env2", flag.ContinueOnError)
	envHost2 := envFlags2.String("host", hostDefault, "server host")
	envPort2 := envFlags2.Int("port", portDefault, "server port")
	envFlags2.Parse([]string{"-host", "override.com", "-port", "3000"})
	fmt.Printf("    With flags: host=%q, port=%d (flag overrides)\n", *envHost2, *envPort2)

	// ============================================
	// 8. COMPARISON WITH OS.ARGS
	// ============================================
	fmt.Println("\n--- 8. Comparison with os.Args ---")
	fmt.Println(`
os.Args is the raw argument slice (no parsing):
  os.Args[0]   -> program name
  os.Args[1:]  -> all arguments as strings

flag package advantages over manual os.Args parsing:
  - Automatic type conversion
  - Default values
  - Usage/help generation
  - Error handling
  - Boolean flag shorthand (-debug vs -debug=true)
  - Standard conventions (- and -- prefixes)

When to use os.Args directly:
  - Very simple programs with 1-2 positional args
  - When you need non-standard parsing
  - Passing raw args to subprocesses`)

	fmt.Printf("\n  os.Args[0] (program name): %q\n", os.Args[0])
	fmt.Printf("  os.Args count: %d\n", len(os.Args))

	// Manual parsing vs flag (conceptual)
	fmt.Println("\n  Manual os.Args parsing example:")
	simulatedOsArgs := []string{"myapp", "-v", "--output", "result.txt", "input.txt"}
	fmt.Printf("    Raw args: %v\n", simulatedOsArgs)

	// Manual parse
	manualVerbose := false
	manualOutput := ""
	var manualFiles []string
	i := 1
	for i < len(simulatedOsArgs) {
		switch simulatedOsArgs[i] {
		case "-v", "--verbose":
			manualVerbose = true
		case "--output", "-o":
			i++
			if i < len(simulatedOsArgs) {
				manualOutput = simulatedOsArgs[i]
			}
		default:
			manualFiles = append(manualFiles, simulatedOsArgs[i])
		}
		i++
	}
	fmt.Printf("    verbose=%v, output=%q, files=%v\n", manualVerbose, manualOutput, manualFiles)

	// ============================================
	// 9. BUILDING A CLI WITH JUST FLAG
	// ============================================
	fmt.Println("\n--- 9. Building a CLI with Just Flag ---")
	fmt.Println(`
You can build sophisticated CLIs with just the flag package.
No need for cobra/urfave-cli for simple tools.

PATTERN:
  1. Define subcommand FlagSets
  2. Check os.Args[1] for subcommand name
  3. Parse remaining args with appropriate FlagSet
  4. Execute subcommand logic`)

	// Complete mini-CLI example
	fmt.Println("\n  Mini-CLI 'taskctl' simulation:")
	cliCommands := [][]string{
		{"taskctl", "add", "-title", "Buy groceries", "-priority", "high"},
		{"taskctl", "list", "-status", "pending", "-limit", "10"},
		{"taskctl", "done", "-id", "42"},
	}

	for _, cmd := range cliCommands {
		fmt.Printf("\n    $ %s\n", strings.Join(cmd, " "))
		executeTaskCLI(cmd[1:])
	}

	// ============================================
	// 10. FLAG GOTCHAS AND TIPS
	// ============================================
	fmt.Println("\n\n--- 10. Flag Gotchas and Tips ---")
	fmt.Println(`
GOTCHAS:

1. flag.Parse() must be called before accessing values
2. Bool flags: -flag sets to true, use -flag=false to set false
3. Flags stop at first non-flag argument
4. -flag is treated same as --flag
5. = is optional: -port 8080 same as -port=8080
6. flag.CommandLine is the default FlagSet used by top-level functions

TIPS:

1. Use FlagSet for testable flag parsing
2. Use Visit to detect explicitly-set flags
3. Use ContinueOnError for custom error handling
4. Group related flags in structs
5. Use flag.Var for any complex type
6. Set flag.Usage for better help messages`)

	// Gotcha: bool flag
	fmt.Println("\n  Bool flag behavior:")
	boolFS := flag.NewFlagSet("bool", flag.ContinueOnError)
	b1 := boolFS.Bool("verbose", false, "verbose")
	boolFS.Parse([]string{"-verbose"})
	fmt.Printf("    -verbose (no value): %v\n", *b1)

	boolFS2 := flag.NewFlagSet("bool2", flag.ContinueOnError)
	b2 := boolFS2.Bool("verbose", false, "verbose")
	boolFS2.Parse([]string{"-verbose=false"})
	fmt.Printf("    -verbose=false:      %v\n", *b2)

	// Config struct pattern
	fmt.Println("\n  Config struct pattern:")
	cfg := parseConfig([]string{"-host", "api.example.com", "-port", "443", "-workers", "8"})
	fmt.Printf("    Config: %+v\n", cfg)

	fmt.Println("\n=== End of Chapter 048 ===")
}

func dispatchCommand(args []string) string {
	if len(args) == 0 {
		return "error: no command specified"
	}

	switch args[0] {
	case "serve":
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
		port := fs.Int("port", 8080, "port")
		fs.Parse(args[1:])
		return fmt.Sprintf("serving on port %d", *port)
	case "version":
		return "v1.0.0"
	case "help":
		return "available commands: serve, version, help"
	default:
		return fmt.Sprintf("unknown command: %q", args[0])
	}
}

type Config struct {
	Host    string
	Port    int
	Workers int
	Debug   bool
}

func parseConfig(args []string) Config {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	cfg := Config{}
	fs.StringVar(&cfg.Host, "host", "localhost", "server host")
	fs.IntVar(&cfg.Port, "port", 8080, "server port")
	fs.IntVar(&cfg.Workers, "workers", 4, "worker count")
	fs.BoolVar(&cfg.Debug, "debug", false, "debug mode")
	fs.Parse(args)
	return cfg
}

func executeTaskCLI(args []string) {
	if len(args) == 0 {
		fmt.Println("      usage: taskctl <command> [flags]")
		return
	}

	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("add", flag.ContinueOnError)
		title := fs.String("title", "", "task title")
		priority := fs.String("priority", "normal", "priority level")
		fs.Parse(args[1:])
		fmt.Printf("      Added task: %q (priority: %s)\n", *title, *priority)

	case "list":
		fs := flag.NewFlagSet("list", flag.ContinueOnError)
		status := fs.String("status", "all", "filter by status")
		limit := fs.Int("limit", 20, "max results")
		fs.Parse(args[1:])
		fmt.Printf("      Listing tasks: status=%s, limit=%d\n", *status, *limit)

	case "done":
		fs := flag.NewFlagSet("done", flag.ContinueOnError)
		id := fs.Int("id", 0, "task ID")
		fs.Parse(args[1:])
		fmt.Printf("      Marked task #%d as done\n", *id)

	default:
		fmt.Printf("      Unknown command: %q\n", args[0])
	}
}
