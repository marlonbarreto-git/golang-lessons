// Package main - Chapter 098: golang.org/x/sys and golang.org/x/mod
// x/sys provides low-level OS interfaces beyond the stdlib.
// x/mod provides module-aware tooling for Go modules and semver.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"
)

func main() {
	fmt.Println("=== GOLANG.ORG/X/SYS AND GOLANG.ORG/X/MOD ===")

	// ============================================
	// X/SYS OVERVIEW
	// ============================================
	fmt.Println("\n--- golang.org/x/sys Overview ---")

	fmt.Println(`
  golang.org/x/sys provides low-level operating system
  interfaces that go beyond the standard syscall package.

  KEY PACKAGES:
    x/sys/unix     - Unix/Linux/macOS syscalls
    x/sys/windows  - Windows syscalls
    x/sys/plan9    - Plan 9 syscalls
    x/sys/cpu      - CPU feature detection
    x/sys/execabs  - Hardened os/exec.LookPath

  Install: go get golang.org/x/sys

  WHY x/sys OVER syscall:
    - syscall package is frozen (no new features)
    - x/sys/unix has ALL Linux/macOS syscalls
    - Better typed, more ergonomic API
    - Actively maintained with new OS features`)

	// ============================================
	// X/SYS/UNIX
	// ============================================
	fmt.Println("\n--- x/sys/unix: Unix System Calls ---")

	os.Stdout.WriteString(`
  RESOURCE LIMITS:

    import "golang.org/x/sys/unix"

    // Get current resource limits
    var rlimit unix.Rlimit
    err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rlimit)
    fmt.Printf("Max open files: cur=%d, max=%d\n", rlimit.Cur, rlimit.Max)

    // Set resource limits
    rlimit.Cur = 65536
    err = unix.Setrlimit(unix.RLIMIT_NOFILE, &rlimit)

    // Common limits:
    unix.RLIMIT_NOFILE  - max open file descriptors
    unix.RLIMIT_NPROC   - max processes
    unix.RLIMIT_AS       - max address space (memory)
    unix.RLIMIT_CORE     - max core file size
    unix.RLIMIT_STACK    - max stack size

  MEMORY MAPPING (mmap):

    // Map a file into memory (read-only)
    data, err := unix.Mmap(
        int(file.Fd()),    // file descriptor
        0,                  // offset
        fileSize,           // length
        unix.PROT_READ,     // protection
        unix.MAP_SHARED,    // flags
    )
    defer unix.Munmap(data)
    // data is a []byte backed by the file

    // Anonymous mapping (not backed by file)
    data, err := unix.Mmap(
        -1,
        0,
        pageSize,
        unix.PROT_READ|unix.PROT_WRITE,
        unix.MAP_ANON|unix.MAP_PRIVATE,
    )

  MEMORY LOCKING:

    // Lock memory to prevent swapping (sensitive data)
    err := unix.Mlock(sensitiveData)
    defer unix.Munlock(sensitiveData)

    // Lock all current and future mappings
    err := unix.Mlockall(unix.MCL_CURRENT | unix.MCL_FUTURE)

  IOCTL:

    // Control device parameters
    unix.IoctlSetInt(fd, request, value)
    val, err := unix.IoctlGetInt(fd, request)

    // Terminal ioctl
    winsize, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
    fmt.Printf("Terminal: %dx%d\n", winsize.Col, winsize.Row)

  EPOLL (Linux):

    // Create epoll instance
    epfd, err := unix.EpollCreate1(0)
    defer unix.Close(epfd)

    // Add file descriptor to watch
    event := unix.EpollEvent{
        Events: unix.EPOLLIN,
        Fd:     int32(fd),
    }
    unix.EpollCtl(epfd, unix.EPOLL_CTL_ADD, fd, &event)

    // Wait for events
    events := make([]unix.EpollEvent, 100)
    n, err := unix.EpollWait(epfd, events, -1)
    for i := 0; i < n; i++ {
        // Handle events[i]
    }

  KQUEUE (macOS/BSD):

    kq, err := unix.Kqueue()
    defer unix.Close(kq)

    changes := []unix.Kevent_t{{
        Ident:  uint64(fd),
        Filter: unix.EVFILT_READ,
        Flags:  unix.EV_ADD | unix.EV_ENABLE,
    }}
    events := make([]unix.Kevent_t, 10)
    n, err := unix.Kevent(kq, changes, events, nil)
` + "\n")

	// Working syscall example with stdlib
	fmt.Println("  Working example (syscall with stdlib):")
	os.Stdout.WriteString(fmt.Sprintf("    OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH))
	os.Stdout.WriteString(fmt.Sprintf("    PID: %d\n", syscall.Getpid()))
	os.Stdout.WriteString(fmt.Sprintf("    UID: %d\n", syscall.Getuid()))
	os.Stdout.WriteString(fmt.Sprintf("    GID: %d\n", syscall.Getgid()))

	wd, _ := syscall.Getwd()
	fmt.Println("    Working dir:", wd)

	var rusage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage); err == nil {
		os.Stdout.WriteString(fmt.Sprintf("    User CPU time: %d.%06ds\n",
			rusage.Utime.Sec, rusage.Utime.Usec))
		os.Stdout.WriteString(fmt.Sprintf("    Max RSS: %d KB\n", rusage.Maxrss))
	}

	// ============================================
	// X/SYS/CPU
	// ============================================
	fmt.Println("\n--- x/sys/cpu: CPU Feature Detection ---")

	fmt.Println(`
  Detect CPU features at runtime for optimized code paths.

    import "golang.org/x/sys/cpu"

    // x86/amd64 features
    if cpu.X86.HasAVX2 {
        // Use AVX2-optimized code
    }
    if cpu.X86.HasAES {
        // Use AES-NI hardware acceleration
    }

    // ARM features
    if cpu.ARM64.HasAES {
        // Use ARM crypto extensions
    }

    // Available x86 features:
    cpu.X86.HasAES       // AES-NI
    cpu.X86.HasAVX       // AVX
    cpu.X86.HasAVX2      // AVX2
    cpu.X86.HasAVX512F   // AVX-512
    cpu.X86.HasSSE2      // SSE2
    cpu.X86.HasSSE41     // SSE4.1
    cpu.X86.HasSSE42     // SSE4.2

    // Available ARM64 features:
    cpu.ARM64.HasAES     // AES
    cpu.ARM64.HasSHA2    // SHA-2
    cpu.ARM64.HasCRC32   // CRC32
    cpu.ARM64.HasATOMICS // Atomic instructions`)

	// ============================================
	// X/SYS/WINDOWS
	// ============================================
	fmt.Println("\n--- x/sys/windows: Windows Syscalls ---")

	fmt.Println(`
  Windows-specific system calls:

    import "golang.org/x/sys/windows"

    // Registry operations
    key, err := windows.OpenKey(
        windows.HKEY_LOCAL_MACHINE,
        "SOFTWARE\\Microsoft\\Windows\\CurrentVersion",
        windows.KEY_READ,
    )
    defer windows.CloseHandle(key)

    // Service management
    manager, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
    defer windows.CloseServiceHandle(manager)

    // Windows-specific file operations
    handle, err := windows.CreateFile(
        windows.StringToUTF16Ptr("file.txt"),
        windows.GENERIC_READ,
        windows.FILE_SHARE_READ,
        nil,
        windows.OPEN_EXISTING,
        windows.FILE_ATTRIBUTE_NORMAL,
        0,
    )`)

	// ============================================
	// X/MOD OVERVIEW
	// ============================================
	fmt.Println("\n--- golang.org/x/mod Overview ---")

	fmt.Println(`
  golang.org/x/mod provides Go module-aware tooling:

  KEY PACKAGES:
    x/mod/semver    - Semantic versioning
    x/mod/modfile   - Parse and edit go.mod files
    x/mod/module    - Module paths and versions
    x/mod/sumdb     - Checksum database client
    x/mod/zip       - Module zip file creation

  Install: go get golang.org/x/mod`)

	// ============================================
	// SEMVER
	// ============================================
	fmt.Println("\n--- x/mod/semver: Semantic Versioning ---")

	fmt.Println(`
  Semantic versioning follows: vMAJOR.MINOR.PATCH[-prerelease][+build]

  API:

    import "golang.org/x/mod/semver"

    // Validate
    semver.IsValid("v1.2.3")        // true
    semver.IsValid("v1.2.3-beta.1") // true
    semver.IsValid("1.2.3")         // false (missing "v" prefix!)
    semver.IsValid("v1")            // false (incomplete)

    // Compare
    semver.Compare("v1.2.3", "v1.3.0")  // -1 (less)
    semver.Compare("v2.0.0", "v1.9.9")  // +1 (greater)
    semver.Compare("v1.0.0", "v1.0.0")  //  0 (equal)

    // Extract components
    semver.Major("v1.2.3")       // "v1"
    semver.MajorMinor("v1.2.3")  // "v1.2"
    semver.Prerelease("v1.2.3-beta.1")  // "-beta.1"
    semver.Build("v1.2.3+meta")  // "+meta"

    // Canonical form
    semver.Canonical("v1.2.3-beta.1+meta")  // "v1.2.3-beta.1"

    // Max version
    semver.Max("v1.2.3", "v1.3.0")  // "v1.3.0"

  IMPORTANT: Go semver always requires "v" prefix!

  SORTING VERSIONS:

    versions := []string{"v1.2.0", "v1.10.0", "v1.3.0", "v2.0.0"}
    semver.Sort(versions)
    // ["v1.2.0", "v1.3.0", "v1.10.0", "v2.0.0"]`)

	// Working semver-like example
	fmt.Println("\n  Working example (version comparison with stdlib):")
	versions := []string{"v1.2.3", "v1.10.0", "v1.3.0", "v2.0.0-beta.1"}
	for _, v := range versions {
		os.Stdout.WriteString(fmt.Sprintf("    %s: valid=%t\n", v, isValidSemver(v)))
	}
	os.Stdout.WriteString(fmt.Sprintf("    Compare v1.2.3 vs v1.3.0: %d\n", compareSemver("v1.2.3", "v1.3.0")))
	os.Stdout.WriteString(fmt.Sprintf("    Compare v2.0.0 vs v1.9.9: %d\n", compareSemver("v2.0.0", "v1.9.9")))

	// ============================================
	// MODFILE
	// ============================================
	fmt.Println("\n--- x/mod/modfile: Parse go.mod Files ---")

	os.Stdout.WriteString(`
  Parse, modify, and write go.mod files programmatically.

  API:

    import "golang.org/x/mod/modfile"

    // Parse go.mod
    data, _ := os.ReadFile("go.mod")
    file, err := modfile.Parse("go.mod", data, nil)

    // Read module info
    fmt.Println(file.Module.Mod.Path)     // module path
    fmt.Println(file.Go.Version)          // go version

    // List requirements
    for _, req := range file.Require {
        fmt.Printf("%s %s (indirect=%t)\n",
            req.Mod.Path, req.Mod.Version, req.Indirect)
    }

    // Add a requirement
    file.AddNewRequire("github.com/pkg/errors", "v0.9.1", false)

    // Drop a requirement
    file.DropRequire("github.com/old/dep")

    // Set Go version
    file.AddGoStmt("1.22")

    // Format and write back
    newData, err := file.Format()
    os.WriteFile("go.mod", newData, 0644)

  REAL-WORLD USE CASES:
    - Dependency management tools
    - CI/CD version bumping
    - Module graph analysis
    - Automated dependency updates
` + "\n")

	// ============================================
	// MODULE
	// ============================================
	fmt.Println("--- x/mod/module: Module Paths and Versions ---")

	fmt.Println(`
  Work with module paths and versions:

    import "golang.org/x/mod/module"

    // Validate module version
    version := module.Version{
        Path:    "github.com/user/repo",
        Version: "v1.2.3",
    }
    err := module.Check(version.Path, version.Version)

    // Escape module path for filesystem
    escaped, err := module.EscapePath("github.com/User/Repo")
    // "github.com/!user/!repo" (uppercase escaped)

    // Unescape
    original, err := module.UnescapePath(escaped)

    // Check if path is valid
    err := module.CheckPath("github.com/user/repo")

    // Major version suffix
    // v2+ modules must have /v2 suffix in path:
    // github.com/user/repo/v2

  MODULE PATH RULES:
    - Must be valid import paths
    - Cannot have uppercase letters in first element
    - v2+ requires /vN suffix
    - No leading dots or underscores`)

	// ============================================
	// PRACTICAL EXAMPLES
	// ============================================
	fmt.Println("\n--- Practical Patterns ---")

	os.Stdout.WriteString(`
  PATTERN 1: CI version bumper

    func bumpVersion(current, bumpType string) string {
        major := semver.Major(current)
        mm := semver.MajorMinor(current)

        switch bumpType {
        case "major":
            n, _ := strconv.Atoi(major[1:])
            return fmt.Sprintf("v%d.0.0", n+1)
        case "minor":
            parts := strings.Split(mm[1:], ".")
            minor, _ := strconv.Atoi(parts[1])
            return fmt.Sprintf("v%s.%d.0", parts[0], minor+1)
        case "patch":
            // ... parse and increment patch
        }
    }

  PATTERN 2: Dependency analyzer

    func analyzeDeps(gomodPath string) {
        data, _ := os.ReadFile(gomodPath)
        f, _ := modfile.Parse(gomodPath, data, nil)

        direct, indirect := 0, 0
        for _, req := range f.Require {
            if req.Indirect {
                indirect++
            } else {
                direct++
            }
        }
        fmt.Printf("Direct: %d, Indirect: %d\n", direct, indirect)
    }

  PATTERN 3: Resource limit checker (production)

    func checkLimits() error {
        var rlimit unix.Rlimit
        unix.Getrlimit(unix.RLIMIT_NOFILE, &rlimit)

        if rlimit.Cur < 65536 {
            log.Printf("WARNING: open file limit is %d, recommend 65536+", rlimit.Cur)
            rlimit.Cur = 65536
            if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &rlimit); err != nil {
                return fmt.Errorf("failed to set file limit: %w", err)
            }
        }
        return nil
    }
` + "\n")

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("--- Summary ---")

	fmt.Println(`
  x/sys:
    unix     - Full Unix syscall access (mmap, epoll, rlimit, ioctl)
    windows  - Windows syscalls (registry, services, handles)
    cpu      - CPU feature detection (AVX, AES-NI, etc.)

  x/mod:
    semver   - Semantic version parsing, comparison, sorting
    modfile  - Parse/edit go.mod programmatically
    module   - Module path validation and escaping

  KEY TAKEAWAYS:
    - Use x/sys/unix instead of syscall for new code
    - Go semver always requires "v" prefix
    - modfile is essential for dependency management tools
    - CPU detection enables optimized code paths

  Install:
    go get golang.org/x/sys
    go get golang.org/x/mod`)
}

// ============================================
// SIMPLE SEMVER HELPERS (stdlib only)
// ============================================

func isValidSemver(v string) bool {
	if !strings.HasPrefix(v, "v") {
		return false
	}
	rest := v[1:]
	// Strip prerelease/build
	if idx := strings.IndexAny(rest, "-+"); idx >= 0 {
		rest = rest[:idx]
	}
	parts := strings.Split(rest, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func compareSemver(a, b string) int {
	parseMajor := func(v string) int {
		v = v[1:] // strip "v"
		dot := strings.Index(v, ".")
		if dot < 0 {
			return 0
		}
		n := 0
		for _, c := range v[:dot] {
			n = n*10 + int(c-'0')
		}
		return n
	}
	ma, mb := parseMajor(a), parseMajor(b)
	if ma < mb {
		return -1
	}
	if ma > mb {
		return 1
	}
	// Simplified: only compare major
	return 0
}
