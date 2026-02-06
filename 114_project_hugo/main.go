// Package main - Chapter 114: Hugo - Static Site Generator
// Hugo is the world's fastest framework for building websites.
// Written in Go, it demonstrates template engine usage, content
// pipeline architecture, and extreme build performance.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 114: HUGO - STATIC SITE GENERATOR ===")

	// ============================================
	// WHAT IS HUGO
	// ============================================
	fmt.Println(`
WHAT IS HUGO
------------
Hugo is an open-source static site generator known for its incredible
build speed. It takes content in Markdown, applies templates, and
generates a complete static website.

Created by Steve Francia (spf13) in 2013, Hugo is one of the most
popular static site generators alongside Jekyll, Next.js, and Gatsby.

KEY FACTS:
- Written entirely in Go
- ~78,000+ GitHub stars
- Builds thousands of pages in milliseconds
- Built-in template engine (Go templates)
- Content organization via front matter
- Taxonomies, menus, i18n built-in
- No external dependencies (single binary)
- Used by kubernetes.io, gohugo.io, 1Password blog`)

	// ============================================
	// WHY GO FOR HUGO
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR HUGO
---------------------------
1. Build speed        - Go's compilation model enables sub-second builds
2. Template engine    - text/template and html/template in stdlib
3. Concurrency        - Parallel page rendering via goroutines
4. Single binary      - No Ruby/Node.js/Python dependency
5. File I/O           - Fast filesystem operations for content reading
6. Cross-compilation  - Build for any platform`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
HUGO ARCHITECTURE
-----------------

  +---------------------------------------------+
  |              Hugo CLI                        |
  |  hugo new site / hugo server / hugo build    |
  +-------------------+-------------------------+
                      |
  +-------------------v-------------------------+
  |            Content Pipeline                  |
  |  +----------+  +----------+  +-----------+  |
  |  | Content  |  | Front    |  | Markdown  |  |
  |  | Reader   |  | Matter   |  | Parser    |  |
  |  | (fs)     |  | (YAML/   |  | (goldmark)|  |
  |  +----+-----+  | TOML)    |  +-----+-----+  |
  |       |        +----+-----+        |         |
  |       +-------------+-----+--------+         |
  |                      |                        |
  |  +-------------------v--------------------+   |
  |  |           Page Builder                  |   |
  |  |  - Taxonomies, Menus, Sections          |   |
  |  |  - Related content, Pagination          |   |
  |  +-------------------+--------------------+   |
  |                      |                        |
  |  +-------------------v--------------------+   |
  |  |         Template Engine                 |   |
  |  |  (Go html/template + Hugo functions)    |   |
  |  +-------------------+--------------------+   |
  |                      |                        |
  |  +-------------------v--------------------+   |
  |  |         Output Formats                  |   |
  |  |  HTML, RSS, JSON, AMP, sitemap          |   |
  |  +-------------------+--------------------+   |
  +----------------------|------------------------+
                         |
  +----------------------v------------------------+
  |              /public/ directory                |
  |  (complete static site ready for deployment)   |
  +-----------------------------------------------+

CONTENT STRUCTURE:
  my-site/
  +-- config.toml        (site configuration)
  +-- content/           (Markdown content)
  |   +-- posts/
  |   +-- about.md
  +-- layouts/           (templates)
  |   +-- _default/
  |   +-- partials/
  +-- static/            (static assets)
  +-- themes/            (theme packages)
  +-- public/            (generated output)`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN HUGO
-----------------------

1. CONTENT PIPELINE (Pipeline Pattern)
   - Content flows through stages: read -> parse -> transform -> render
   - Each stage can be parallelized
   - Stages connected via Go channels or function composition

2. GO TEMPLATE ENGINE
   - html/template for safe HTML output
   - Custom template functions (400+ Hugo functions)
   - Template inheritance via baseof/block/define
   - Partial templates for reusability

   {{ range .Pages }}
     <h2>{{ .Title }}</h2>
     {{ .Summary }}
   {{ end }}

3. PARALLEL PAGE RENDERING
   - Pages rendered concurrently with goroutine pool
   - Worker pool pattern with configurable parallelism
   - Independent pages rendered in parallel
   - Dependent content (taxonomies) synchronized

4. FILESYSTEM ABSTRACTION
   - Afero filesystem interface for testability
   - Union filesystem overlays themes + project
   - Memory filesystem for fast testing
   - Watching filesystem for live reload

5. FRONT MATTER PARSING
   - YAML, TOML, JSON, ORG mode support
   - Unmarshaled into Page struct
   - Custom parameters via map[string]interface{}
   - Cascading configuration (section -> page)`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN HUGO
-------------------------------

1. PARALLEL RENDERING
   - Worker pool renders pages concurrently
   - Number of workers = GOMAXPROCS
   - Each worker processes pages from a channel
   - 10,000+ pages built in < 1 second

2. LAZY RENDERING
   - Related content computed on demand
   - Table of contents generated when accessed
   - Summary truncated lazily
   - Avoids unnecessary computation

3. FAST MARKDOWN PARSING (Goldmark)
   - Goldmark is written in Go (no CGO)
   - Streaming parser, not DOM-based
   - Extensions loaded only when needed
   - Faster than Blackfriday (previous parser)

4. TEMPLATE CACHING
   - Compiled templates cached in memory
   - Partial templates cached by name
   - Template functions memoized where possible

5. INCREMENTAL BUILDING
   - hugo server watches for file changes
   - Only changed content re-rendered
   - Live reload via WebSocket injection
   - Sub-millisecond rebuild for single page change`)

	// ============================================
	// GO PHILOSOPHY
	// ============================================
	fmt.Println(`
HOW HUGO EXPRESSES GO PHILOSOPHY
---------------------------------

"Simplicity is prerequisite for reliability"
  Hugo's mental model is simple: content + templates = site.
  No build chain, no webpack, no node_modules.

"Make the zero value useful"
  Hugo with default config and no templates still generates
  valid HTML. Sensible defaults for everything.

"Don't communicate by sharing memory; share memory
 by communicating"
  Page rendering workers receive pages via channels.
  Results collected through output channels.

"A little copying is better than a little dependency"
  Hugo includes its own Markdown parser, image processing,
  and asset pipeline rather than shelling out to external tools.`)

	// ============================================
	// SIMPLIFIED HUGO DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED HUGO-LIKE STATIC SITE GENERATOR ===")

	site := NewHugoSite(SiteConfig{
		Title:   "My Tech Blog",
		BaseURL: "https://example.com",
		Language: "en",
	})

	fmt.Println("\n--- Adding Content ---")
	site.AddPage(Page{
		Path:    "posts/getting-started-with-go.md",
		Title:   "Getting Started with Go",
		Date:    time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		Draft:   false,
		Tags:    []string{"go", "tutorial", "beginner"},
		Section: "posts",
		Content: "Go is a statically typed, compiled language designed at Google...",
	})
	site.AddPage(Page{
		Path:    "posts/concurrency-patterns.md",
		Title:   "Concurrency Patterns in Go",
		Date:    time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
		Draft:   false,
		Tags:    []string{"go", "concurrency", "advanced"},
		Section: "posts",
		Content: "Go provides goroutines and channels as first-class concurrency primitives...",
	})
	site.AddPage(Page{
		Path:    "posts/draft-post.md",
		Title:   "Work in Progress",
		Date:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Draft:   true,
		Tags:    []string{"draft"},
		Section: "posts",
		Content: "This is a draft and should not appear in production builds...",
	})
	site.AddPage(Page{
		Path:    "posts/web-servers-in-go.md",
		Title:   "Building Web Servers in Go",
		Date:    time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC),
		Draft:   false,
		Tags:    []string{"go", "web", "http"},
		Section: "posts",
		Content: "The net/http package provides everything you need for web servers...",
	})
	site.AddPage(Page{
		Path:    "about.md",
		Title:   "About Me",
		Date:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Draft:   false,
		Section: "page",
		Content: "I'm a software engineer passionate about Go and distributed systems.",
	})

	fmt.Println("\n--- Building Site ---")
	site.Build()

	fmt.Println("\n--- Generated Output ---")
	site.PrintOutput()

	fmt.Println("\n--- Taxonomies ---")
	site.PrintTaxonomies()

	fmt.Println("\n--- Site Map ---")
	site.PrintSiteMap()

	fmt.Println("\n--- Build Statistics ---")
	site.PrintStats()

	fmt.Println(`
SUMMARY
-------
Hugo set the standard for static site generator performance.
Its Go codebase demonstrates:
- Content pipeline with parallel rendering
- Go template engine with 400+ custom functions
- Filesystem abstraction for testability
- Lazy evaluation for optimal performance
- Incremental builds with live reload

Hugo proved that build tools can be blazing fast without
sacrificing features.`)
}

// ============================================
// TYPES
// ============================================

type SiteConfig struct {
	Title    string
	BaseURL  string
	Language string
}

type Page struct {
	Path      string
	Title     string
	Date      time.Time
	Draft     bool
	Tags      []string
	Section   string
	Content   string
	Summary   string
	Permalink string
	OutputPath string
}

type OutputFile struct {
	Path    string
	Content string
	Size    int
}

// ============================================
// HUGO SITE
// ============================================

type HugoSite struct {
	config     SiteConfig
	pages      []*Page
	outputs    []*OutputFile
	taxonomies map[string]map[string][]*Page
	buildTime  time.Duration
	mu         sync.Mutex
}

func NewHugoSite(config SiteConfig) *HugoSite {
	return &HugoSite{
		config:     config,
		taxonomies: make(map[string]map[string][]*Page),
	}
}

func (s *HugoSite) AddPage(p Page) {
	if p.Summary == "" && len(p.Content) > 70 {
		p.Summary = p.Content[:70] + "..."
	} else if p.Summary == "" {
		p.Summary = p.Content
	}

	slug := strings.TrimSuffix(filepath.Base(p.Path), ".md")
	if p.Section == "page" {
		p.Permalink = fmt.Sprintf("%s/%s/", s.config.BaseURL, slug)
		p.OutputPath = fmt.Sprintf("public/%s/index.html", slug)
	} else {
		p.Permalink = fmt.Sprintf("%s/%s/%s/", s.config.BaseURL, p.Section, slug)
		p.OutputPath = fmt.Sprintf("public/%s/%s/index.html", p.Section, slug)
	}

	s.pages = append(s.pages, &p)

	status := ""
	if p.Draft {
		status = " [DRAFT]"
	}
	fmt.Printf("  + %s%s (%s)\n", p.Title, status, p.Path)
}

func (s *HugoSite) Build() {
	start := time.Now()

	publishable := s.getPublishablePages()
	fmt.Printf("  Pages: %d total, %d publishable, %d drafts\n",
		len(s.pages), len(publishable), len(s.pages)-len(publishable))

	fmt.Println("  Building taxonomies...")
	s.buildTaxonomies(publishable)

	fmt.Println("  Rendering pages (parallel)...")
	var wg sync.WaitGroup
	results := make(chan *OutputFile, len(publishable)+10)

	for _, page := range publishable {
		wg.Add(1)
		go func(p *Page) {
			defer wg.Done()
			output := s.renderPage(p)
			results <- output
		}(page)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		results <- s.renderIndex(publishable)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		results <- s.renderRSS(publishable)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		results <- s.renderSitemap(publishable)
	}()

	for name, tags := range s.taxonomies {
		for tag, pages := range tags {
			wg.Add(1)
			go func(n, t string, pp []*Page) {
				defer wg.Done()
				results <- s.renderTaxonomyPage(n, t, pp)
			}(name, tag, pages)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for output := range results {
		s.mu.Lock()
		s.outputs = append(s.outputs, output)
		s.mu.Unlock()
	}

	s.buildTime = time.Since(start)
	fmt.Printf("  Build complete in %v\n", s.buildTime.Round(time.Microsecond))
}

func (s *HugoSite) getPublishablePages() []*Page {
	var pages []*Page
	for _, p := range s.pages {
		if !p.Draft {
			pages = append(pages, p)
		}
	}
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Date.After(pages[j].Date)
	})
	return pages
}

func (s *HugoSite) buildTaxonomies(pages []*Page) {
	tags := make(map[string][]*Page)
	for _, p := range pages {
		for _, tag := range p.Tags {
			tags[tag] = append(tags[tag], p)
		}
	}
	s.taxonomies["tags"] = tags
	fmt.Printf("    Tags: %d unique tags\n", len(tags))
}

func (s *HugoSite) renderPage(p *Page) *OutputFile {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head><title>%s - %s</title></head>
<body>
<article>
<h1>%s</h1>
<time>%s</time>
%s
</article>
</body></html>`,
		s.config.Language, p.Title, s.config.Title,
		p.Title, p.Date.Format("2006-01-02"), p.Content)

	return &OutputFile{
		Path:    p.OutputPath,
		Content: html,
		Size:    len(html),
	}
}

func (s *HugoSite) renderIndex(pages []*Page) *OutputFile {
	var items []string
	for _, p := range pages {
		if p.Section == "posts" {
			items = append(items, fmt.Sprintf("<li><a href=\"%s\">%s</a> - %s</li>",
				p.Permalink, p.Title, p.Date.Format("2006-01-02")))
		}
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>%s</title></head>
<body><h1>%s</h1><ul>%s</ul></body></html>`,
		s.config.Title, s.config.Title, strings.Join(items, "\n"))

	return &OutputFile{
		Path:    "public/index.html",
		Content: html,
		Size:    len(html),
	}
}

func (s *HugoSite) renderRSS(pages []*Page) *OutputFile {
	var items []string
	for _, p := range pages {
		if p.Section == "posts" {
			items = append(items, fmt.Sprintf("  <item><title>%s</title><link>%s</link></item>",
				p.Title, p.Permalink))
		}
	}

	rss := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
<channel>
  <title>%s</title>
  <link>%s</link>
%s
</channel></rss>`, s.config.Title, s.config.BaseURL, strings.Join(items, "\n"))

	return &OutputFile{
		Path:    "public/index.xml",
		Content: rss,
		Size:    len(rss),
	}
}

func (s *HugoSite) renderSitemap(pages []*Page) *OutputFile {
	var urls []string
	for _, p := range pages {
		urls = append(urls, fmt.Sprintf("  <url><loc>%s</loc></url>", p.Permalink))
	}

	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
%s
</urlset>`, strings.Join(urls, "\n"))

	return &OutputFile{
		Path:    "public/sitemap.xml",
		Content: sitemap,
		Size:    len(sitemap),
	}
}

func (s *HugoSite) renderTaxonomyPage(taxonomy, term string, pages []*Page) *OutputFile {
	var items []string
	for _, p := range pages {
		items = append(items, fmt.Sprintf("<li><a href=\"%s\">%s</a></li>", p.Permalink, p.Title))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>%s: %s</title></head>
<body><h1>%s: %s</h1><ul>%s</ul></body></html>`,
		taxonomy, term, taxonomy, term, strings.Join(items, "\n"))

	path := fmt.Sprintf("public/%s/%s/index.html", taxonomy, term)
	return &OutputFile{
		Path:    path,
		Content: html,
		Size:    len(html),
	}
}

func (s *HugoSite) PrintOutput() {
	sort.Slice(s.outputs, func(i, j int) bool {
		return s.outputs[i].Path < s.outputs[j].Path
	})

	fmt.Println("  Generated files:")
	totalSize := 0
	for _, f := range s.outputs {
		fmt.Printf("    %s (%d bytes)\n", f.Path, f.Size)
		totalSize += f.Size
	}
	fmt.Printf("  Total: %d files, %d bytes\n", len(s.outputs), totalSize)
}

func (s *HugoSite) PrintTaxonomies() {
	for taxonomy, terms := range s.taxonomies {
		fmt.Printf("  Taxonomy: %s\n", taxonomy)

		var termNames []string
		for name := range terms {
			termNames = append(termNames, name)
		}
		sort.Strings(termNames)

		for _, name := range termNames {
			pages := terms[name]
			fmt.Printf("    %s (%d pages)\n", name, len(pages))
		}
	}
}

func (s *HugoSite) PrintSiteMap() {
	publishable := s.getPublishablePages()
	fmt.Println("  Site Map:")
	fmt.Printf("    %s/\n", s.config.BaseURL)
	for _, p := range publishable {
		indent := "    "
		if p.Section == "posts" {
			indent = "      "
		}
		fmt.Printf("%s%s\n", indent, p.Permalink)
	}
}

func (s *HugoSite) PrintStats() {
	fmt.Printf("  Build time:     %v\n", s.buildTime.Round(time.Microsecond))
	fmt.Printf("  Total pages:    %d\n", len(s.pages))
	fmt.Printf("  Published:      %d\n", len(s.getPublishablePages()))
	fmt.Printf("  Output files:   %d\n", len(s.outputs))

	totalSize := 0
	for _, f := range s.outputs {
		totalSize += f.Size
	}
	fmt.Printf("  Total size:     %d bytes\n", totalSize)

	tagCount := 0
	for _, terms := range s.taxonomies {
		tagCount += len(terms)
	}
	fmt.Printf("  Taxonomy terms: %d\n", tagCount)
}

func init() {
	_ = os.Stdout
}
