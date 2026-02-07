// Package main - Chapter 049: Path and Filepath Packages
// path (URL-style POSIX paths) vs filepath (OS-native paths): Join, Base, Dir,
// Ext, Clean, Match, Walk, WalkDir, Glob, Abs, Rel, EvalSymlinks, and more.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("=== PATH AND FILEPATH PACKAGES ===")

	// ============================================
	// 1. PATH PACKAGE (URL-STYLE / POSIX PATHS)
	// ============================================
	fmt.Println("\n--- 1. path Package (URL-style / POSIX Paths) ---")
	fmt.Println(`
The 'path' package works with FORWARD-SLASH separated paths.
Use it for: URLs, archive paths, platform-independent paths.
Do NOT use it for filesystem paths on Windows.

FUNCTIONS:
  path.Join(elem...)    -> join elements with /
  path.Base(p)          -> last element
  path.Dir(p)           -> all but last element
  path.Ext(p)           -> file extension
  path.Clean(p)         -> shortest equivalent path
  path.Match(pattern,p) -> shell-like matching
  path.Split(p)         -> (dir, file)
  path.IsAbs(p)         -> starts with /`)

	// path.Join
	fmt.Println("\n  path.Join:")
	fmt.Printf("    Join(\"a\", \"b\", \"c\") = %q\n", path.Join("a", "b", "c"))
	fmt.Printf("    Join(\"/api\", \"v2\", \"users\") = %q\n", path.Join("/api", "v2", "users"))
	fmt.Printf("    Join(\"a\", \"..\", \"b\") = %q\n", path.Join("a", "..", "b"))
	fmt.Printf("    Join(\"\", \"\") = %q\n", path.Join("", ""))
	fmt.Printf("    Join(\"a\", \"\", \"b\") = %q\n", path.Join("a", "", "b"))

	// path.Base
	fmt.Println("\n  path.Base:")
	fmt.Printf("    Base(\"/a/b/c.txt\") = %q\n", path.Base("/a/b/c.txt"))
	fmt.Printf("    Base(\"/a/b/c/\") = %q\n", path.Base("/a/b/c/"))
	fmt.Printf("    Base(\".\") = %q\n", path.Base("."))
	fmt.Printf("    Base(\"\") = %q\n", path.Base(""))
	fmt.Printf("    Base(\"/\") = %q\n", path.Base("/"))

	// path.Dir
	fmt.Println("\n  path.Dir:")
	fmt.Printf("    Dir(\"/a/b/c.txt\") = %q\n", path.Dir("/a/b/c.txt"))
	fmt.Printf("    Dir(\"/a/b/c/\") = %q\n", path.Dir("/a/b/c/"))
	fmt.Printf("    Dir(\"file.txt\") = %q\n", path.Dir("file.txt"))

	// path.Ext
	fmt.Println("\n  path.Ext:")
	fmt.Printf("    Ext(\"file.go\") = %q\n", path.Ext("file.go"))
	fmt.Printf("    Ext(\"archive.tar.gz\") = %q\n", path.Ext("archive.tar.gz"))
	fmt.Printf("    Ext(\"Makefile\") = %q\n", path.Ext("Makefile"))

	// path.Clean
	fmt.Println("\n  path.Clean:")
	fmt.Printf("    Clean(\"a//b/../c/./d\") = %q\n", path.Clean("a//b/../c/./d"))
	fmt.Printf("    Clean(\"/../a/b\") = %q\n", path.Clean("/../a/b"))
	fmt.Printf("    Clean(\"./.\") = %q\n", path.Clean("./."))

	// path.Split
	fmt.Println("\n  path.Split:")
	dir, file := path.Split("/a/b/c.txt")
	fmt.Printf("    Split(\"/a/b/c.txt\") = dir=%q, file=%q\n", dir, file)
	dir, file = path.Split("file.txt")
	fmt.Printf("    Split(\"file.txt\") = dir=%q, file=%q\n", dir, file)

	// path.Match
	fmt.Println("\n  path.Match:")
	matched, _ := path.Match("*.go", "main.go")
	fmt.Printf("    Match(\"*.go\", \"main.go\") = %v\n", matched)
	matched, _ = path.Match("cmd/*/main.go", "cmd/server/main.go")
	fmt.Printf("    Match(\"cmd/*/main.go\", \"cmd/server/main.go\") = %v\n", matched)
	matched, _ = path.Match("[abc]*.txt", "apple.txt")
	fmt.Printf("    Match(\"[abc]*.txt\", \"apple.txt\") = %v\n", matched)

	// path.IsAbs
	fmt.Println("\n  path.IsAbs:")
	fmt.Printf("    IsAbs(\"/usr/bin\") = %v\n", path.IsAbs("/usr/bin"))
	fmt.Printf("    IsAbs(\"relative\") = %v\n", path.IsAbs("relative"))

	// ============================================
	// 2. FILEPATH PACKAGE (OS-NATIVE PATHS)
	// ============================================
	fmt.Println("\n--- 2. filepath Package (OS-Native Paths) ---")
	fmt.Println(`
The 'path/filepath' package works with OS-native paths.
Uses os.PathSeparator (/ on Unix, \ on Windows).
Use this for ALL filesystem operations.

ADDITIONAL FUNCTIONS (beyond path equivalents):
  filepath.Abs(p)          -> absolute path
  filepath.Rel(base, targ) -> relative path from base to targ
  filepath.Walk(root, fn)  -> recursive walk (deprecated)
  filepath.WalkDir(root,fn)-> recursive walk (Go 1.16+, preferred)
  filepath.Glob(pattern)   -> match files on disk
  filepath.EvalSymlinks(p) -> resolve symlinks
  filepath.VolumeName(p)   -> drive letter on Windows

CONSTANTS:
  filepath.Separator     -> os.PathSeparator (/ or \)
  filepath.ListSeparator -> os.PathListSeparator (: or ;)

CONVERTERS:
  filepath.FromSlash(p)  -> replace / with os separator
  filepath.ToSlash(p)    -> replace os separator with /`)

	// filepath.Join (OS-aware)
	fmt.Println("\n  filepath.Join:")
	fmt.Printf("    Join(\"a\", \"b\", \"c\") = %q\n", filepath.Join("a", "b", "c"))
	fmt.Printf("    Join(\"/usr\", \"local\", \"bin\") = %q\n", filepath.Join("/usr", "local", "bin"))
	fmt.Printf("    Join(\"a\", \"..\", \"b\") = %q\n", filepath.Join("a", "..", "b"))

	// filepath.Abs
	fmt.Println("\n  filepath.Abs:")
	abs, _ := filepath.Abs(".")
	fmt.Printf("    Abs(\".\") = %q\n", abs)
	abs, _ = filepath.Abs("../somefile")
	fmt.Printf("    Abs(\"../somefile\") = %q\n", abs)

	// filepath.Rel
	fmt.Println("\n  filepath.Rel:")
	rel, _ := filepath.Rel("/a/b", "/a/b/c/d")
	fmt.Printf("    Rel(\"/a/b\", \"/a/b/c/d\") = %q\n", rel)
	rel, _ = filepath.Rel("/a/b/c", "/a/x/y")
	fmt.Printf("    Rel(\"/a/b/c\", \"/a/x/y\") = %q\n", rel)
	rel, _ = filepath.Rel("/a", "/a")
	fmt.Printf("    Rel(\"/a\", \"/a\") = %q\n", rel)

	// filepath.Base / Dir / Ext (same as path but OS-aware)
	fmt.Println("\n  filepath.Base/Dir/Ext:")
	testPath := filepath.Join("/home", "user", "docs", "report.pdf")
	fmt.Printf("    Path: %q\n", testPath)
	fmt.Printf("    Base: %q\n", filepath.Base(testPath))
	fmt.Printf("    Dir:  %q\n", filepath.Dir(testPath))
	fmt.Printf("    Ext:  %q\n", filepath.Ext(testPath))

	// filepath.Separator and ListSeparator
	fmt.Printf("\n  filepath.Separator: %q\n", string(filepath.Separator))
	fmt.Printf("  filepath.ListSeparator: %q\n", string(filepath.ListSeparator))

	// ============================================
	// 3. FILEPATH.FROMSLASH / TOSLASH
	// ============================================
	fmt.Println("\n--- 3. filepath.FromSlash / ToSlash ---")
	fmt.Println(`
Convert between forward-slash and OS-native paths:
  FromSlash(p) -> replace / with filepath.Separator
  ToSlash(p)   -> replace filepath.Separator with /

On Unix, these are no-ops (separator is already /).
On Windows, they convert between / and \.

Use when:
  - Reading paths from config files (always use /)
  - Storing paths in a portable format
  - Converting URL paths to file paths`)

	p := "a/b/c/file.txt"
	fmt.Printf("\n  Original: %q\n", p)
	fmt.Printf("  FromSlash: %q\n", filepath.FromSlash(p))
	fmt.Printf("  ToSlash:   %q\n", filepath.ToSlash(filepath.FromSlash(p)))

	// ============================================
	// 4. FILEPATH.MATCH AND GLOB
	// ============================================
	fmt.Println("\n--- 4. filepath.Match and Glob ---")
	fmt.Println(`
filepath.Match uses shell-like patterns (same as path.Match):
  *     -> any sequence of non-Separator characters
  ?     -> any single non-Separator character
  [abc] -> character class
  [a-z] -> character range

filepath.Glob returns matching files FROM DISK:
  matches, err := filepath.Glob("*.go")`)

	// Match patterns
	fmt.Println("\n  filepath.Match:")
	patterns := []struct {
		pattern, name string
	}{
		{"*.go", "main.go"},
		{"*.go", "dir/main.go"},
		{"**/*.go", "dir/main.go"},
		{"test_*.go", "test_utils.go"},
		{"[Mm]akefile", "Makefile"},
		{"[Mm]akefile", "makefile"},
		{"?.txt", "a.txt"},
		{"?.txt", "ab.txt"},
	}
	for _, tc := range patterns {
		m, _ := filepath.Match(tc.pattern, tc.name)
		fmt.Printf("    Match(%q, %q) = %v\n", tc.pattern, tc.name, m)
	}

	// Glob on actual filesystem
	fmt.Println("\n  filepath.Glob (actual files):")
	goFiles, _ := filepath.Glob(filepath.Join(abs, "*.go"))
	if len(goFiles) == 0 {
		// Try from project root
		goFiles, _ = filepath.Glob("/Users/marlonbarreto/Work/golang-lessons/041_strings_strconv/*.go")
	}
	for _, f := range goFiles {
		fmt.Printf("    %s\n", filepath.Base(f))
	}

	// ============================================
	// 5. FILEPATH.WALK VS WALKDIR
	// ============================================
	fmt.Println("\n--- 5. filepath.Walk vs WalkDir (Go 1.16+) ---")
	fmt.Println(`
filepath.Walk (legacy):
  func Walk(root string, fn WalkFunc) error
  type WalkFunc func(path string, info os.FileInfo, err error) error
  - Calls os.Lstat on every entry (slow)

filepath.WalkDir (Go 1.16+, preferred):
  func WalkDir(root string, fn fs.WalkDirFunc) error
  type WalkDirFunc func(path string, d fs.DirEntry, err error) error
  - Uses fs.DirEntry (lazy stat, much faster)
  - Avoids syscall for each entry

RETURN VALUES FROM CALLBACK:
  nil            -> continue walking
  fs.SkipDir     -> skip this directory
  fs.SkipAll     -> stop walking entirely (Go 1.20+)
  other error    -> stop walking, return error`)

	// Create temp directory structure for demo
	tmpDir, err := os.MkdirTemp("", "walkdemo")
	if err != nil {
		fmt.Printf("  Error creating temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Create demo structure
	dirs := []string{"src", "src/pkg", "src/cmd", "docs", "vendor", "vendor/lib"}
	files := map[string]string{
		"main.go":            "package main",
		"go.mod":             "module demo",
		"src/app.go":         "package src",
		"src/pkg/utils.go":   "package pkg",
		"src/cmd/run.go":     "package cmd",
		"docs/README.md":     "# Docs",
		"vendor/lib/lib.go":  "package lib",
	}

	for _, d := range dirs {
		os.MkdirAll(filepath.Join(tmpDir, d), 0755)
	}
	for f, content := range files {
		os.WriteFile(filepath.Join(tmpDir, f), []byte(content), 0644)
	}

	// WalkDir example
	fmt.Println("\n  WalkDir (all entries):")
	filepath.WalkDir(tmpDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(tmpDir, p)
		if relPath == "." {
			return nil
		}
		indent := strings.Repeat("  ", strings.Count(relPath, string(filepath.Separator)))
		if d.IsDir() {
			fmt.Printf("    %s[%s/]\n", indent, filepath.Base(p))
		} else {
			fmt.Printf("    %s%s\n", indent, filepath.Base(p))
		}
		return nil
	})

	// WalkDir with SkipDir (skip vendor)
	fmt.Println("\n  WalkDir (skip vendor/):")
	var goFileList []string
	filepath.WalkDir(tmpDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "vendor" {
			return fs.SkipDir
		}
		if !d.IsDir() && filepath.Ext(p) == ".go" {
			relPath, _ := filepath.Rel(tmpDir, p)
			goFileList = append(goFileList, relPath)
		}
		return nil
	})
	for _, f := range goFileList {
		fmt.Printf("    %s\n", f)
	}

	// ============================================
	// 6. FILEPATH.EVALSYMLINKS
	// ============================================
	fmt.Println("\n--- 6. filepath.EvalSymlinks ---")
	fmt.Println(`
filepath.EvalSymlinks resolves all symlinks in a path.
Returns the actual path on disk.

Useful for:
  - Resolving config file locations
  - Detecting if two paths point to same file
  - Security: preventing symlink attacks`)

	// Create a symlink for demo
	realFile := filepath.Join(tmpDir, "main.go")
	symLink := filepath.Join(tmpDir, "link_to_main.go")
	err = os.Symlink(realFile, symLink)
	if err == nil {
		resolved, _ := filepath.EvalSymlinks(symLink)
		fmt.Printf("\n  Symlink: %s\n", filepath.Base(symLink))
		fmt.Printf("  Resolved: %s\n", filepath.Base(resolved))
	} else {
		fmt.Printf("\n  Symlink creation failed: %v\n", err)
	}

	// Resolve current directory
	resolved, _ := filepath.EvalSymlinks(".")
	fmt.Printf("  EvalSymlinks(\".\") = %q\n", resolved)

	// ============================================
	// 7. WHEN TO USE PATH VS FILEPATH
	// ============================================
	fmt.Println("\n--- 7. When to Use path vs filepath ---")
	fmt.Println(`
USE 'path' FOR:
  - URL paths:          path.Join("/api", "v2", "users")
  - Archive entries:    path.Join("archive", "dir", "file.txt")
  - Abstract paths:     platform-independent path logic
  - Always uses /

USE 'path/filepath' FOR:
  - Filesystem paths:   filepath.Join("src", "main.go")
  - Walking directories: filepath.WalkDir(...)
  - File operations:    filepath.Glob, filepath.Abs
  - Symlink resolution: filepath.EvalSymlinks
  - Uses OS separator

COMMON MISTAKE:
  // WRONG: using path for filesystem
  p := path.Join("C:", "Users", "file.txt")  // "C:/Users/file.txt" on Windows

  // CORRECT: using filepath for filesystem
  p := filepath.Join("C:", "Users", "file.txt")  // "C:\Users\file.txt" on Windows`)

	// Demonstrate the difference
	fmt.Println("\n  Comparison:")
	fmt.Printf("    path.Join(\"a\", \"b\", \"c\"):     %q\n", path.Join("a", "b", "c"))
	fmt.Printf("    filepath.Join(\"a\", \"b\", \"c\"): %q\n", filepath.Join("a", "b", "c"))

	// URL path manipulation
	urlPath := "/api/v1/../v2/users/./profile"
	fmt.Printf("\n  URL path cleanup: %q -> %q\n", urlPath, path.Clean(urlPath))

	// ============================================
	// 8. PRACTICAL EXAMPLES
	// ============================================
	fmt.Println("\n--- 8. Practical Examples ---")

	// Find all files by extension
	fmt.Println("\n  Find files by extension:")
	goFilesFound := findByExtension(tmpDir, ".go")
	for _, f := range goFilesFound {
		fmt.Printf("    %s\n", f)
	}

	// Safe path joining (prevent directory traversal)
	fmt.Println("\n  Safe path join (prevent traversal):")
	baseDir := "/app/uploads"
	testPaths := []string{"photo.jpg", "../etc/passwd", "subdir/file.txt", "../../root/.ssh/id_rsa"}
	for _, tp := range testPaths {
		safe, err := safeJoin(baseDir, tp)
		if err != nil {
			fmt.Printf("    safeJoin(%q, %q) -> ERROR: %v\n", baseDir, tp, err)
		} else {
			fmt.Printf("    safeJoin(%q, %q) -> %q\n", baseDir, tp, safe)
		}
	}

	// File name without extension
	fmt.Println("\n  Filename without extension:")
	testFiles := []string{"main.go", "archive.tar.gz", "Makefile", ".gitignore", "photo.jpeg"}
	for _, f := range testFiles {
		name := fileNameWithoutExt(f)
		fmt.Printf("    %q -> %q\n", f, name)
	}

	fmt.Println("\n=== End of Chapter 049 ===")
}

/*
SUMMARY - Chapter 049: Path and Filepath Packages

PATH PACKAGE (URL-STYLE):
- Works with forward-slash separated paths
- Use for URLs, archive paths, platform-independent paths
- path.Join, Base, Dir, Ext, Clean, Split, Match, IsAbs
- Do NOT use for filesystem paths on Windows

FILEPATH PACKAGE (OS-NATIVE):
- Works with OS-native paths (/ on Unix, \ on Windows)
- Use for ALL filesystem operations
- filepath.Join, Base, Dir, Ext, Clean, Split, Match
- filepath.Abs: resolve to absolute path
- filepath.Rel: compute relative path between two paths
- filepath.FromSlash/ToSlash: convert path separators
- filepath.EvalSymlinks: resolve symbolic links

FILEPATH.GLOB:
- Match files on disk using shell-like patterns
- * matches non-separator chars, ? matches single char
- [abc] character class, [a-z] character range

FILEPATH.WALKDIR (GO 1.16+):
- Preferred over Walk (uses fs.DirEntry, much faster)
- Return nil to continue, fs.SkipDir to skip directory
- fs.SkipAll to stop entirely (Go 1.20+)
- Walk calls os.Lstat per entry (slow), WalkDir does not

PATH VS FILEPATH:
- path: URLs, archives, abstract paths (always /)
- filepath: filesystem operations (OS separator)
- Common mistake: using path for filesystem on Windows

PRACTICAL PATTERNS:
- Safe path joining to prevent directory traversal
- Finding files by extension recursively
- Extracting filename without extension
*/

func findByExtension(root, ext string) []string {
	var results []string
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(p) == ext {
			relPath, _ := filepath.Rel(root, p)
			results = append(results, relPath)
		}
		return nil
	})
	return results
}

func safeJoin(base, userPath string) (string, error) {
	joined := filepath.Join(base, filepath.Clean("/"+userPath))
	if !strings.HasPrefix(joined, filepath.Clean(base)+string(filepath.Separator)) &&
		joined != filepath.Clean(base) {
		return "", fmt.Errorf("path traversal detected: %q resolves outside base", userPath)
	}
	return joined, nil
}

func fileNameWithoutExt(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
