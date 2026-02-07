// Package main - Chapter 122: Gitea - Self-Hosted Git Service in Go
package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 122: GITEA - SELF-HOSTED GIT SERVICE IN GO ===")

	// ============================================
	// WHAT IS GITEA?
	// ============================================
	fmt.Println(`
==============================
WHAT IS GITEA?
==============================

Gitea is a lightweight, self-hosted Git service written in Go.
It is a community fork of Gogs with active development. It provides:

  - Git repository hosting with web UI
  - Pull requests and code review
  - Issue tracking and project boards
  - CI/CD via Gitea Actions (GitHub Actions compatible)
  - Package registry (npm, PyPI, Docker, etc.)
  - OAuth2, LDAP, SAML authentication
  - Webhooks and API (GitHub-compatible)

WHO USES IT:
  - CERN (physics research)
  - Codeberg (public Gitea instance)
  - Many universities and enterprises
  - Homelab and self-hosting communities
  - Organizations needing on-premise Git hosting

WHY GO:
  Gitea chose Go for single-binary deployment, low resource
  usage, cross-platform support, and excellent HTTP server
  capabilities. A Gitea instance can run on a Raspberry Pi
  with ~100MB of RAM.`)

	// ============================================
	// ARCHITECTURE OVERVIEW
	// ============================================
	fmt.Println(`
==============================
ARCHITECTURE OVERVIEW
==============================

  +---------------------------------------------------+
  |                Web UI (Templates)                  |
  |  (Go html/template, JavaScript frontend)          |
  +---------------------------------------------------+
  |              API Layer (REST + GraphQL)             |
  |  (chi router, middleware chain)                    |
  +---------------------------------------------------+
  |              Service Layer                         |
  |  (Repository, Issue, PR, User, Org services)      |
  +---------------------------------------------------+
  |              Model Layer (ORM)                     |
  |  (XORM - Go ORM supporting multiple DBs)          |
  +---------------------------------------------------+
  |          Git Layer (native + go-git)               |
  |  (Shell out to git, or use go-git library)        |
  +---------------------------------------------------+
  |              Storage                               |
  |  +----------+  +----------+  +----------+         |
  |  | Database |  | Git Repos|  | LFS/     |         |
  |  | SQLite/  |  | (bare)   |  | Packages |         |
  |  | PG/MySQL |  |          |  |          |         |
  |  +----------+  +----------+  +----------+         |
  +---------------------------------------------------+

KEY DESIGN DECISIONS:
  - Single binary with embedded assets
  - SQLite as default DB (zero config)
  - Git operations via native git commands
  - Modular authentication (OAuth2, LDAP, PAM)
  - Queue system for background jobs`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
==============================
KEY GO PATTERNS IN GITEA
==============================

1. MIDDLEWARE CHAIN
   HTTP requests flow through authentication, CSRF,
   session, and context middleware. Each middleware
   is a Go function wrapping the next handler.

2. CONTEXT PATTERN
   Gitea uses a rich context object per request that
   carries user info, repository info, permissions,
   and flash messages through the handler chain.

3. SERVICE LAYER PATTERN
   Business logic is separated from HTTP handlers
   into service functions, making it testable and
   reusable across API and web routes.

4. ORM WITH XORM
   Database operations use XORM, supporting SQLite,
   PostgreSQL, MySQL, and MSSQL with the same code.

5. QUEUE SYSTEM
   Background tasks (notifications, webhooks, indexing)
   use an internal queue supporting channel, Redis,
   and LevelDB backends.

6. EMBED FOR ASSETS
   Go 1.16 embed directive bundles templates, CSS,
   JS, and translations into the single binary.`)

	// ============================================
	// DEMO: SIMPLIFIED GIT SERVICE
	// ============================================
	fmt.Println(`
==============================
DEMO: SIMPLIFIED GIT SERVICE
==============================

This demo implements core Gitea concepts:
  - Git object model (blobs, trees, commits)
  - Repository management
  - Issue tracking
  - Pull request workflow
  - User and permission model
  - Webhook notifications`)

	fmt.Println("\n--- Git Object Model ---")
	demoGitObjects()

	fmt.Println("\n--- Repository Management ---")
	demoRepositoryManagement()

	fmt.Println("\n--- Issue Tracking ---")
	demoIssueTracking()

	fmt.Println("\n--- Pull Request Workflow ---")
	demoPullRequests()

	fmt.Println("\n--- Full Mini-Gitea ---")
	demoMiniGitea()

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
==============================
PERFORMANCE TECHNIQUES
==============================

1. EFFICIENT GIT OPERATIONS
   - Shell out to native git for performance
   - Parse git output with zero-copy techniques
   - Cache frequently accessed objects (commits, trees)

2. DATABASE OPTIMIZATION
   - Careful index design for common queries
   - Batch operations for bulk updates
   - Connection pooling via XORM

3. CACHING LAYERS
   - In-memory cache for user sessions
   - Repository metadata caching
   - Avatar and asset caching

4. BACKGROUND PROCESSING
   - Webhook delivery in background goroutines
   - Email notification batching
   - Repository indexing for code search

5. RESOURCE EFFICIENCY
   - Runs on minimal hardware (Raspberry Pi)
   - SQLite for small deployments (no external DB)
   - Embedded assets eliminate file I/O`)

	// ============================================
	// GO PHILOSOPHY IN GITEA
	// ============================================
	fmt.Println(`
==============================
GO PHILOSOPHY IN GITEA
==============================

SIMPLICITY OF DEPLOYMENT:
  A single binary with SQLite. Download, run, done.
  No Docker, no PostgreSQL, no Redis required for
  basic setups. The Go philosophy of simplicity
  extends to the deployment model.

LOW RESOURCE CONSUMPTION:
  Go's efficient runtime means Gitea can serve
  hundreds of users on a machine with 512MB RAM.
  No JVM overhead, no interpreted language costs.

CROSS-PLATFORM:
  Go's cross-compilation means Gitea runs on Linux,
  macOS, Windows, ARM, and more. Same codebase,
  same features everywhere.

COMPOSITION:
  Authentication backends, storage backends, and
  database engines are swappable through interfaces.
  Users compose the deployment that fits their needs.

COMMUNITY-DRIVEN:
  Like Go itself, Gitea values clear, readable code.
  Contributions must be straightforward and well-tested.`)

	fmt.Println("\n=== END OF CHAPTER 122 ===")
}

// ============================================
// GIT OBJECT MODEL
// ============================================

type GitObjectType string

const (
	BlobObject   GitObjectType = "blob"
	TreeObject   GitObjectType = "tree"
	CommitObject GitObjectType = "commit"
)

type GitObject struct {
	Type    GitObjectType
	Hash    string
	Content []byte
}

type TreeEntry struct {
	Mode string
	Name string
	Hash string
	Type GitObjectType
}

type Commit struct {
	Hash      string
	Tree      string
	Parent    string
	Author    string
	Committer string
	Message   string
	Timestamp time.Time
}

type GitObjectStore struct {
	mu      sync.RWMutex
	objects map[string]*GitObject
	refs    map[string]string
}

func NewGitObjectStore() *GitObjectStore {
	return &GitObjectStore{
		objects: make(map[string]*GitObject),
		refs:    make(map[string]string),
	}
}

func (s *GitObjectStore) HashObject(objType GitObjectType, content []byte) string {
	header := fmt.Sprintf("%s %d\x00", objType, len(content))
	h := sha1.New()
	h.Write([]byte(header))
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *GitObjectStore) StoreBlob(content []byte) string {
	hash := s.HashObject(BlobObject, content)
	s.mu.Lock()
	s.objects[hash] = &GitObject{Type: BlobObject, Hash: hash, Content: content}
	s.mu.Unlock()
	return hash
}

func (s *GitObjectStore) StoreTree(entries []TreeEntry) string {
	var buf strings.Builder
	for _, e := range entries {
		buf.WriteString(fmt.Sprintf("%s %s %s %s\n", e.Mode, string(e.Type), e.Hash, e.Name))
	}
	content := []byte(buf.String())
	hash := s.HashObject(TreeObject, content)
	s.mu.Lock()
	s.objects[hash] = &GitObject{Type: TreeObject, Hash: hash, Content: content}
	s.mu.Unlock()
	return hash
}

func (s *GitObjectStore) StoreCommit(tree, parent, author, message string) *Commit {
	ts := time.Now()
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("tree %s\n", tree))
	if parent != "" {
		buf.WriteString(fmt.Sprintf("parent %s\n", parent))
	}
	buf.WriteString(fmt.Sprintf("author %s %d +0000\n", author, ts.Unix()))
	buf.WriteString(fmt.Sprintf("committer %s %d +0000\n", author, ts.Unix()))
	buf.WriteString(fmt.Sprintf("\n%s\n", message))

	content := []byte(buf.String())
	hash := s.HashObject(CommitObject, content)

	commit := &Commit{
		Hash:      hash,
		Tree:      tree,
		Parent:    parent,
		Author:    author,
		Committer: author,
		Message:   message,
		Timestamp: ts,
	}

	s.mu.Lock()
	s.objects[hash] = &GitObject{Type: CommitObject, Hash: hash, Content: content}
	s.mu.Unlock()

	return commit
}

func (s *GitObjectStore) UpdateRef(ref, hash string) {
	s.mu.Lock()
	s.refs[ref] = hash
	s.mu.Unlock()
}

func (s *GitObjectStore) GetRef(ref string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refs[ref]
}

func demoGitObjects() {
	store := NewGitObjectStore()

	blob1 := store.StoreBlob([]byte("# My Project\nWelcome to the project!"))
	blob2 := store.StoreBlob([]byte("package main\n\nfunc main() {}\n"))
	fmt.Printf("  Blob README.md: %s\n", blob1[:12])
	fmt.Printf("  Blob main.go:   %s\n", blob2[:12])

	tree := store.StoreTree([]TreeEntry{
		{Mode: "100644", Name: "README.md", Hash: blob1, Type: BlobObject},
		{Mode: "100644", Name: "main.go", Hash: blob2, Type: BlobObject},
	})
	fmt.Printf("  Tree (root):    %s\n", tree[:12])

	commit := store.StoreCommit(tree, "", "Alice <alice@example.com>", "Initial commit")
	fmt.Printf("  Commit:         %s\n", commit.Hash[:12])

	store.UpdateRef("refs/heads/main", commit.Hash)
	fmt.Printf("  refs/heads/main -> %s\n", store.GetRef("refs/heads/main")[:12])

	blob3 := store.StoreBlob([]byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}\n"))
	tree2 := store.StoreTree([]TreeEntry{
		{Mode: "100644", Name: "README.md", Hash: blob1, Type: BlobObject},
		{Mode: "100644", Name: "main.go", Hash: blob3, Type: BlobObject},
	})
	commit2 := store.StoreCommit(tree2, commit.Hash, "Bob <bob@example.com>", "Add hello world")
	store.UpdateRef("refs/heads/main", commit2.Hash)
	fmt.Printf("  Commit 2:       %s (parent: %s)\n", commit2.Hash[:12], commit2.Parent[:12])

	store.mu.RLock()
	fmt.Printf("  Object store: %d objects, %d refs\n", len(store.objects), len(store.refs))
	store.mu.RUnlock()
}

// ============================================
// REPOSITORY MANAGEMENT
// ============================================

type Repository struct {
	mu          sync.RWMutex
	ID          int
	Owner       string
	Name        string
	Description string
	Private     bool
	Stars       int
	Forks       int
	ObjectStore *GitObjectStore
	Created     time.Time
	Updated     time.Time
}

type RepoManager struct {
	mu     sync.RWMutex
	repos  map[string]*Repository
	nextID int
}

func NewRepoManager() *RepoManager {
	return &RepoManager{
		repos:  make(map[string]*Repository),
		nextID: 1,
	}
}

func (rm *RepoManager) CreateRepo(owner, name, description string, private bool) *Repository {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	fullName := owner + "/" + name
	repo := &Repository{
		ID:          rm.nextID,
		Owner:       owner,
		Name:        name,
		Description: description,
		Private:     private,
		ObjectStore: NewGitObjectStore(),
		Created:     time.Now(),
		Updated:     time.Now(),
	}
	rm.nextID++
	rm.repos[fullName] = repo
	return repo
}

func (rm *RepoManager) GetRepo(owner, name string) *Repository {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.repos[owner+"/"+name]
}

func (rm *RepoManager) ForkRepo(srcOwner, srcName, newOwner string) *Repository {
	src := rm.GetRepo(srcOwner, srcName)
	if src == nil {
		return nil
	}

	fork := rm.CreateRepo(newOwner, srcName, "Fork of "+srcOwner+"/"+srcName, false)

	src.mu.Lock()
	src.Forks++
	src.mu.Unlock()

	return fork
}

func (rm *RepoManager) StarRepo(owner, name string) {
	repo := rm.GetRepo(owner, name)
	if repo != nil {
		repo.mu.Lock()
		repo.Stars++
		repo.mu.Unlock()
	}
}

func (rm *RepoManager) SearchRepos(query string) []*Repository {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var results []*Repository
	queryLower := strings.ToLower(query)
	for _, repo := range rm.repos {
		if strings.Contains(strings.ToLower(repo.Name), queryLower) ||
			strings.Contains(strings.ToLower(repo.Description), queryLower) {
			results = append(results, repo)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Stars > results[j].Stars
	})
	return results
}

func demoRepositoryManagement() {
	rm := NewRepoManager()

	repos := []struct {
		owner, name, desc string
		private           bool
	}{
		{"alice", "webapp", "A modern web application", false},
		{"alice", "dotfiles", "My configuration files", true},
		{"bob", "api-server", "REST API server in Go", false},
		{"charlie", "ml-toolkit", "Machine learning tools", false},
	}

	for _, r := range repos {
		repo := rm.CreateRepo(r.owner, r.name, r.desc, r.private)
		fmt.Printf("  Created: %s/%s (id=%d, private=%v)\n", repo.Owner, repo.Name, repo.ID, repo.Private)
	}

	rm.StarRepo("alice", "webapp")
	rm.StarRepo("alice", "webapp")
	rm.StarRepo("bob", "api-server")

	fork := rm.ForkRepo("alice", "webapp", "bob")
	fmt.Printf("  Forked: %s/%s -> %s/%s\n", "alice", "webapp", fork.Owner, fork.Name)

	results := rm.SearchRepos("api")
	fmt.Printf("  Search 'api': %d results\n", len(results))
	for _, r := range results {
		fmt.Printf("    %s/%s - %s (stars=%d)\n", r.Owner, r.Name, r.Description, r.Stars)
	}

	webapp := rm.GetRepo("alice", "webapp")
	fmt.Printf("  alice/webapp: stars=%d, forks=%d\n", webapp.Stars, webapp.Forks)
}

// ============================================
// ISSUE TRACKING
// ============================================

type Issue struct {
	mu        sync.RWMutex
	ID        int
	RepoOwner string
	RepoName  string
	Title     string
	Body      string
	State     string
	Labels    []string
	Assignees []string
	Author    string
	Comments  []IssueComment
	Created   time.Time
	Updated   time.Time
}

type IssueComment struct {
	ID      int
	Author  string
	Body    string
	Created time.Time
}

type IssueTracker struct {
	mu     sync.RWMutex
	issues map[string][]*Issue
	nextID int
}

func NewIssueTracker() *IssueTracker {
	return &IssueTracker{
		issues: make(map[string][]*Issue),
		nextID: 1,
	}
}

func (it *IssueTracker) CreateIssue(repoOwner, repoName, author, title, body string, labels []string) *Issue {
	it.mu.Lock()
	defer it.mu.Unlock()

	issue := &Issue{
		ID:        it.nextID,
		RepoOwner: repoOwner,
		RepoName:  repoName,
		Title:     title,
		Body:      body,
		State:     "open",
		Labels:    labels,
		Author:    author,
		Created:   time.Now(),
		Updated:   time.Now(),
	}
	it.nextID++

	key := repoOwner + "/" + repoName
	it.issues[key] = append(it.issues[key], issue)
	return issue
}

func (it *IssueTracker) AddComment(issue *Issue, author, body string) {
	issue.mu.Lock()
	defer issue.mu.Unlock()

	issue.Comments = append(issue.Comments, IssueComment{
		ID:      len(issue.Comments) + 1,
		Author:  author,
		Body:    body,
		Created: time.Now(),
	})
	issue.Updated = time.Now()
}

func (it *IssueTracker) CloseIssue(issue *Issue) {
	issue.mu.Lock()
	defer issue.mu.Unlock()
	issue.State = "closed"
	issue.Updated = time.Now()
}

func (it *IssueTracker) ListIssues(repoOwner, repoName, state string) []*Issue {
	it.mu.RLock()
	defer it.mu.RUnlock()

	key := repoOwner + "/" + repoName
	var results []*Issue
	for _, issue := range it.issues[key] {
		issue.mu.RLock()
		if state == "" || issue.State == state {
			results = append(results, issue)
		}
		issue.mu.RUnlock()
	}
	return results
}

func demoIssueTracking() {
	tracker := NewIssueTracker()

	i1 := tracker.CreateIssue("alice", "webapp", "bob", "Login page broken",
		"The login page returns 500 error", []string{"bug", "priority:high"})
	fmt.Printf("  Issue #%d: %s [%s]\n", i1.ID, i1.Title, i1.State)

	i2 := tracker.CreateIssue("alice", "webapp", "charlie", "Add dark mode",
		"Users are requesting dark mode support", []string{"enhancement"})
	fmt.Printf("  Issue #%d: %s [%s]\n", i2.ID, i2.Title, i2.State)

	tracker.AddComment(i1, "alice", "I can reproduce this. Looking into it.")
	tracker.AddComment(i1, "alice", "Fixed in commit abc123. The session middleware was misconfigured.")

	tracker.CloseIssue(i1)

	openIssues := tracker.ListIssues("alice", "webapp", "open")
	closedIssues := tracker.ListIssues("alice", "webapp", "closed")
	fmt.Printf("  Open issues: %d, Closed issues: %d\n", len(openIssues), len(closedIssues))

	i1.mu.RLock()
	fmt.Printf("  Issue #%d has %d comments, state=%s\n", i1.ID, len(i1.Comments), i1.State)
	i1.mu.RUnlock()
}

// ============================================
// PULL REQUEST WORKFLOW
// ============================================

type PullRequest struct {
	mu         sync.RWMutex
	ID         int
	Title      string
	Body       string
	Author     string
	BaseBranch string
	HeadBranch string
	State      string
	Mergeable  bool
	Reviews    []Review
	Labels     []string
	Created    time.Time
	Merged     time.Time
}

type Review struct {
	Author  string
	State   string
	Body    string
	Created time.Time
}

type PRManager struct {
	mu     sync.RWMutex
	prs    map[string][]*PullRequest
	nextID int
}

func NewPRManager() *PRManager {
	return &PRManager{
		prs:    make(map[string][]*PullRequest),
		nextID: 1,
	}
}

func (pm *PRManager) CreatePR(repoKey, author, title, body, base, head string) *PullRequest {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pr := &PullRequest{
		ID:         pm.nextID,
		Title:      title,
		Body:       body,
		Author:     author,
		BaseBranch: base,
		HeadBranch: head,
		State:      "open",
		Mergeable:  true,
		Created:    time.Now(),
	}
	pm.nextID++
	pm.prs[repoKey] = append(pm.prs[repoKey], pr)
	return pr
}

func (pm *PRManager) AddReview(pr *PullRequest, author, state, body string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.Reviews = append(pr.Reviews, Review{
		Author:  author,
		State:   state,
		Body:    body,
		Created: time.Now(),
	})
}

func (pm *PRManager) MergePR(pr *PullRequest) bool {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if pr.State != "open" || !pr.Mergeable {
		return false
	}

	hasApproval := false
	for _, r := range pr.Reviews {
		if r.State == "approved" {
			hasApproval = true
			break
		}
	}

	if !hasApproval {
		return false
	}

	pr.State = "merged"
	pr.Merged = time.Now()
	return true
}

func demoPullRequests() {
	pm := NewPRManager()

	pr := pm.CreatePR("alice/webapp", "bob", "Fix login bug",
		"Fixes #1 - session middleware configuration",
		"main", "fix/login-bug")
	fmt.Printf("  PR #%d: %s (%s -> %s)\n", pr.ID, pr.Title, pr.HeadBranch, pr.BaseBranch)

	pm.AddReview(pr, "charlie", "comment", "Looks good, but add a test please")
	fmt.Println("  Review: charlie requested changes")

	pm.AddReview(pr, "alice", "approved", "LGTM!")
	fmt.Println("  Review: alice approved")

	merged := pm.MergePR(pr)
	pr.mu.RLock()
	fmt.Printf("  Merge result: %v, state=%s\n", merged, pr.State)
	pr.mu.RUnlock()

	pr2 := pm.CreatePR("alice/webapp", "charlie", "Add dark mode",
		"Implements dark mode support. Closes #2",
		"main", "feature/dark-mode")
	canMerge := pm.MergePR(pr2)
	fmt.Printf("  PR #%d merge without approval: %v\n", pr2.ID, canMerge)
}

// ============================================
// FULL MINI-GITEA
// ============================================

type MiniGitea struct {
	repos    *RepoManager
	issues   *IssueTracker
	prs      *PRManager
	users    map[string]*User
	webhooks map[string][]WebhookConfig
	mu       sync.RWMutex
}

type User struct {
	Username string
	Email    string
	IsAdmin  bool
	Created  time.Time
}

type WebhookConfig struct {
	URL    string
	Events []string
	Active bool
}

type WebhookEvent struct {
	Type    string
	Repo    string
	Payload interface{}
}

func NewMiniGitea() *MiniGitea {
	return &MiniGitea{
		repos:    NewRepoManager(),
		issues:   NewIssueTracker(),
		prs:      NewPRManager(),
		users:    make(map[string]*User),
		webhooks: make(map[string][]WebhookConfig),
	}
}

func (g *MiniGitea) RegisterUser(username, email string, admin bool) *User {
	g.mu.Lock()
	defer g.mu.Unlock()

	user := &User{
		Username: username,
		Email:    email,
		IsAdmin:  admin,
		Created:  time.Now(),
	}
	g.users[username] = user
	return user
}

func (g *MiniGitea) AddWebhook(repoKey, url string, events []string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.webhooks[repoKey] = append(g.webhooks[repoKey], WebhookConfig{
		URL:    url,
		Events: events,
		Active: true,
	})
}

func (g *MiniGitea) TriggerWebhook(repoKey string, event WebhookEvent) []string {
	g.mu.RLock()
	hooks := g.webhooks[repoKey]
	g.mu.RUnlock()

	var triggered []string
	for _, hook := range hooks {
		if !hook.Active {
			continue
		}
		for _, evt := range hook.Events {
			if evt == event.Type || evt == "*" {
				triggered = append(triggered, hook.URL)
				break
			}
		}
	}
	return triggered
}

func demoMiniGitea() {
	gitea := NewMiniGitea()

	alice := gitea.RegisterUser("alice", "alice@example.com", true)
	bob := gitea.RegisterUser("bob", "bob@example.com", false)
	charlie := gitea.RegisterUser("charlie", "charlie@example.com", false)
	_ = charlie

	fmt.Printf("  Users: alice(admin=%v), bob(admin=%v)\n", alice.IsAdmin, bob.IsAdmin)

	repo := gitea.repos.CreateRepo("alice", "webapp", "Production web app", false)
	blob := repo.ObjectStore.StoreBlob([]byte("# webapp\nProduction web application"))
	tree := repo.ObjectStore.StoreTree([]TreeEntry{
		{Mode: "100644", Name: "README.md", Hash: blob, Type: BlobObject},
	})
	commit := repo.ObjectStore.StoreCommit(tree, "", "alice <alice@example.com>", "Initial commit")
	repo.ObjectStore.UpdateRef("refs/heads/main", commit.Hash)
	fmt.Printf("  Repo: %s/%s, initial commit: %s\n", repo.Owner, repo.Name, commit.Hash[:12])

	gitea.AddWebhook("alice/webapp", "https://ci.example.com/hooks", []string{"push", "pull_request"})
	gitea.AddWebhook("alice/webapp", "https://chat.example.com/notify", []string{"*"})

	issue := gitea.issues.CreateIssue("alice", "webapp", "bob", "Performance regression",
		"Page load time increased by 2x after last deploy", []string{"bug", "performance"})
	fmt.Printf("  Issue #%d created by bob: %s\n", issue.ID, issue.Title)

	triggered := gitea.TriggerWebhook("alice/webapp", WebhookEvent{
		Type: "issues",
		Repo: "alice/webapp",
		Payload: map[string]interface{}{
			"action": "opened",
			"issue":  issue.ID,
		},
	})
	fmt.Printf("  Webhooks triggered for issue event: %v\n", triggered)

	pr := gitea.prs.CreatePR("alice/webapp", "bob", "Fix perf regression",
		"Fixes #3 - optimized database queries",
		"main", "fix/performance")
	fmt.Printf("  PR #%d created: %s\n", pr.ID, pr.Title)

	triggered = gitea.TriggerWebhook("alice/webapp", WebhookEvent{
		Type: "pull_request",
		Repo: "alice/webapp",
		Payload: map[string]interface{}{
			"action": "opened",
			"pr":     pr.ID,
		},
	})
	fmt.Printf("  Webhooks triggered for PR event: %v\n", triggered)

	gitea.prs.AddReview(pr, "alice", "approved", "Great fix!")
	gitea.prs.MergePR(pr)
	pr.mu.RLock()
	fmt.Printf("  PR #%d merged: state=%s\n", pr.ID, pr.State)
	pr.mu.RUnlock()

	gitea.repos.StarRepo("alice", "webapp")
	gitea.repos.StarRepo("alice", "webapp")
	gitea.repos.ForkRepo("alice", "webapp", "charlie")

	repo = gitea.repos.GetRepo("alice", "webapp")
	fmt.Printf("  Final stats: stars=%d, forks=%d\n", repo.Stars, repo.Forks)

	gitea.issues.CloseIssue(issue)
	open := gitea.issues.ListIssues("alice", "webapp", "open")
	closed := gitea.issues.ListIssues("alice", "webapp", "closed")
	fmt.Printf("  Issues: open=%d, closed=%d\n", len(open), len(closed))
}

/*
SUMMARY - GITEA: SELF-HOSTED GIT SERVICE IN GO:

WHAT IS GITEA:
- Lightweight self-hosted Git service, community fork of Gogs
- Repo hosting, pull requests, issue tracking, CI/CD (Gitea Actions)
- Package registry, OAuth2/LDAP auth, GitHub-compatible API
- Runs on minimal hardware (~100MB RAM on Raspberry Pi)

ARCHITECTURE:
- Layers: Web UI (templates) -> API (chi router) -> Service -> Model (XORM) -> Git + Storage
- Single binary with embedded assets (Go embed directive)
- SQLite as default database for zero-config deployment
- Queue system for background jobs (webhooks, notifications, indexing)

KEY GO PATTERNS DEMONSTRATED:
- Git Object Model: blobs, trees, commits with SHA-1 hashing
- Repository Management: create, fork, star, search repositories
- Issue Tracking: create, comment, close, label, filter issues
- Pull Request Workflow: create PR, review, approve, merge
- Webhook System: event-driven notifications to external services

PERFORMANCE TECHNIQUES:
- Native git commands for fast operations with cached output
- Careful database index design for common queries
- In-memory caching for sessions, repository metadata, avatars
- Background goroutines for webhook delivery and email batching
- Embedded assets eliminate runtime file I/O

GO PHILOSOPHY:
- Single binary deployment with SQLite (download, run, done)
- Low resource consumption thanks to Go's efficient runtime
- Cross-platform via Go cross-compilation
- Swappable backends through interfaces (auth, storage, database)
*/
