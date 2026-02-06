// Package main - Chapter 110: Terraform - Infrastructure as Code
// Terraform by HashiCorp enables declarative infrastructure management
// across cloud providers. Written in Go, it demonstrates plugin
// architecture, DAG-based execution, and state management patterns.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 110: TERRAFORM - INFRASTRUCTURE AS CODE ===")

	// ============================================
	// WHAT IS TERRAFORM
	// ============================================
	fmt.Println(`
WHAT IS TERRAFORM
-----------------
Terraform is an infrastructure-as-code tool that lets you define
cloud and on-premises resources in human-readable configuration files
that you can version, reuse, and share.

Created by HashiCorp in 2014, Terraform uses a declarative approach:
you describe the desired state and Terraform figures out how to
achieve it.

KEY FACTS:
- Written in Go
- ~43,000+ GitHub stars
- Supports 3000+ providers (AWS, GCP, Azure, etc.)
- HCL (HashiCorp Configuration Language) for configs
- State-based resource tracking
- Plan-before-apply workflow
- OpenTofu is the open-source fork (2023)`)

	// ============================================
	// WHY GO FOR TERRAFORM
	// ============================================
	fmt.Println(`
WHY GO WAS CHOSEN FOR TERRAFORM
---------------------------------
1. Plugin system    - Go plugins compiled as separate binaries
2. Static binaries  - Single binary, no runtime dependencies
3. Cross-platform   - Build for Linux, macOS, Windows
4. Concurrency      - Parallel resource provisioning via goroutines
5. Performance      - Fast plan computation for large infra
6. RPC via gRPC     - Provider plugins communicate over gRPC

HashiCorp is one of the biggest Go shops - Terraform, Vault,
Consul, Nomad, Packer, and Vagrant are all written in Go.`)

	// ============================================
	// ARCHITECTURE
	// ============================================
	fmt.Println(`
TERRAFORM ARCHITECTURE
----------------------

  +-------------------------------------------------+
  |            Terraform CLI                         |
  |  +----------+  +--------+  +---------------+    |
  |  | HCL      |  | State  |  | Dependency    |    |
  |  | Parser   |  | Manager|  | Graph (DAG)   |    |
  |  +----+-----+  +---+----+  +-------+-------+    |
  |       |             |              |             |
  |  +----v-------------v--------------v----------+  |
  |  |              Core Engine                    |  |
  |  |  plan -> diff -> apply -> refresh           |  |
  |  +-----+---------------+---------------------+  |
  +---------+---------------+------------------------+
            |               |
    +-------v------+ +------v-------+
    | Provider     | | Provider     |  (gRPC plugins)
    | (AWS)        | | (GCP)        |
    +-------+------+ +------+-------+
            |               |
    +-------v------+ +------v-------+
    | AWS API      | | GCP API      |
    +--------------+ +--------------+

WORKFLOW:
  terraform init   -> Download providers
  terraform plan   -> Compute diff (desired vs actual)
  terraform apply  -> Execute changes
  terraform destroy -> Remove all resources`)

	// ============================================
	// KEY GO PATTERNS
	// ============================================
	fmt.Println(`
KEY GO PATTERNS IN TERRAFORM
-----------------------------

1. DIRECTED ACYCLIC GRAPH (DAG)
   - Resources and their dependencies form a DAG
   - Topological sort determines execution order
   - Independent resources provisioned in parallel
   - Cycle detection prevents circular dependencies

2. PROVIDER PLUGIN ARCHITECTURE
   - Each provider is a separate Go binary
   - Communication via gRPC (go-plugin framework)
   - Providers implement a standard interface:
     - Schema: declare resource types and attributes
     - Plan: compute what will change
     - Apply: make the actual changes
     - Read: refresh current state

3. STATE MANAGEMENT
   - JSON state file tracks real-world resources
   - Maps config to actual infrastructure IDs
   - Supports remote backends (S3, Consul, etc.)
   - State locking prevents concurrent modifications

4. PLAN/APPLY SEPARATION
   - Plan phase: read-only, shows what will change
   - Apply phase: executes the plan
   - Human review between plan and apply
   - Enables safe infrastructure changes

5. HCL CONFIGURATION
   - Custom DSL parsed into Go AST
   - Variables, expressions, functions
   - Modules for reusable configurations
   - Count/for_each for resource iteration`)

	// ============================================
	// PERFORMANCE TECHNIQUES
	// ============================================
	fmt.Println(`
PERFORMANCE TECHNIQUES IN TERRAFORM
------------------------------------

1. PARALLEL RESOURCE PROVISIONING
   - DAG enables parallel execution of independent resources
   - Configurable parallelism (-parallelism=N)
   - Default: 10 concurrent operations
   - Goroutine per resource with semaphore

2. PROVIDER CACHING
   - Provider binaries cached locally
   - Plugin cache directory shared across projects
   - Lock file ensures reproducible provider versions

3. STATE SNAPSHOTS
   - Incremental state updates during apply
   - Crash recovery from partial state
   - State stored after each resource operation

4. GRAPH OPTIMIZATION
   - Pruning of unchanged subgraphs
   - Targeted operations (-target=resource)
   - Refresh skipping when state is known current

5. PROVIDER MULTIPLEXING
   - Single provider process handles multiple resources
   - Reduces process spawn overhead
   - Connection reuse for API calls`)

	// ============================================
	// GO PHILOSOPHY IN TERRAFORM
	// ============================================
	fmt.Println(`
HOW TERRAFORM EXPRESSES GO PHILOSOPHY
--------------------------------------

"Composition over inheritance"
  Terraform modules compose infrastructure from smaller
  building blocks. No inheritance hierarchies.

"Accept interfaces, return structs"
  Provider interface is small and well-defined.
  Any backend (AWS, GCP, Azure) implements the same
  interface for CRUD operations.

"Clear is better than clever"
  HCL was designed for readability by non-programmers.
  Infrastructure configs should be clear to operators.

"Don't communicate by sharing memory; share memory
 by communicating"
  Provider plugins run as separate processes.
  Communication happens via gRPC, not shared memory.`)

	// ============================================
	// SIMPLIFIED TERRAFORM DEMO
	// ============================================
	fmt.Println("\n=== SIMPLIFIED TERRAFORM-LIKE ENGINE ===")

	engine := NewTerraformEngine()

	fmt.Println("\n--- Defining Configuration ---")
	engine.AddResource(Resource{
		Type: "aws_vpc",
		Name: "main",
		Config: map[string]string{
			"cidr_block": "10.0.0.0/16",
			"name":       "production-vpc",
		},
	})

	engine.AddResource(Resource{
		Type:      "aws_subnet",
		Name:      "public",
		DependsOn: []string{"aws_vpc.main"},
		Config: map[string]string{
			"vpc_id":     "${aws_vpc.main.id}",
			"cidr_block": "10.0.1.0/24",
			"name":       "public-subnet",
		},
	})

	engine.AddResource(Resource{
		Type:      "aws_subnet",
		Name:      "private",
		DependsOn: []string{"aws_vpc.main"},
		Config: map[string]string{
			"vpc_id":     "${aws_vpc.main.id}",
			"cidr_block": "10.0.2.0/24",
			"name":       "private-subnet",
		},
	})

	engine.AddResource(Resource{
		Type:      "aws_security_group",
		Name:      "web",
		DependsOn: []string{"aws_vpc.main"},
		Config: map[string]string{
			"vpc_id":      "${aws_vpc.main.id}",
			"name":        "web-sg",
			"ingress":     "80,443",
			"egress":      "0.0.0.0/0",
		},
	})

	engine.AddResource(Resource{
		Type:      "aws_instance",
		Name:      "web_server",
		DependsOn: []string{"aws_subnet.public", "aws_security_group.web"},
		Config: map[string]string{
			"ami":            "ami-12345678",
			"instance_type":  "t3.medium",
			"subnet_id":      "${aws_subnet.public.id}",
			"security_group": "${aws_security_group.web.id}",
			"name":           "web-server-1",
		},
	})

	engine.AddResource(Resource{
		Type:      "aws_rds_instance",
		Name:      "database",
		DependsOn: []string{"aws_subnet.private", "aws_security_group.web"},
		Config: map[string]string{
			"engine":         "postgres",
			"instance_class": "db.t3.medium",
			"subnet_id":      "${aws_subnet.private.id}",
			"name":           "app-database",
		},
	})

	fmt.Println("\n--- Dependency Graph ---")
	engine.PrintGraph()

	fmt.Println("\n--- Terraform Plan ---")
	plan := engine.Plan()

	fmt.Println("\n--- Terraform Apply ---")
	engine.Apply(plan)

	fmt.Println("\n--- Current State ---")
	engine.PrintState()

	fmt.Println("\n--- Modifying Configuration ---")
	engine.UpdateResource("aws_instance.web_server", map[string]string{
		"instance_type": "t3.large",
	})
	plan2 := engine.Plan()
	engine.Apply(plan2)

	fmt.Println("\n--- Terraform Destroy ---")
	engine.Destroy()

	fmt.Println(`
SUMMARY
-------
Terraform revolutionized infrastructure management.
Its Go codebase demonstrates:
- DAG-based dependency resolution and parallel execution
- Plugin architecture via gRPC for extensibility
- State management for tracking real infrastructure
- Plan/apply workflow for safe changes
- Declarative configuration parsing

Terraform made "Infrastructure as Code" mainstream.`)
}

// ============================================
// TYPES
// ============================================

type ResourceStatus string

const (
	StatusPlanned  ResourceStatus = "planned"
	StatusCreating ResourceStatus = "creating"
	StatusCreated  ResourceStatus = "created"
	StatusUpdating ResourceStatus = "updating"
	StatusDestroying ResourceStatus = "destroying"
	StatusDestroyed  ResourceStatus = "destroyed"
)

type Resource struct {
	Type      string
	Name      string
	Config    map[string]string
	DependsOn []string
}

type ResourceState struct {
	Address    string
	Type       string
	Name       string
	ID         string
	Attributes map[string]string
	Status     ResourceStatus
}

type PlanAction string

const (
	ActionCreate  PlanAction = "create"
	ActionUpdate  PlanAction = "update"
	ActionDestroy PlanAction = "destroy"
	ActionNoOp    PlanAction = "no-op"
)

type PlanItem struct {
	Address string
	Action  PlanAction
	Before  map[string]string
	After   map[string]string
}

// ============================================
// TERRAFORM ENGINE
// ============================================

type TerraformEngine struct {
	resources map[string]*Resource
	state     map[string]*ResourceState
	order     []string
	nextID    int
	mu        sync.Mutex
}

func NewTerraformEngine() *TerraformEngine {
	return &TerraformEngine{
		resources: make(map[string]*Resource),
		state:     make(map[string]*ResourceState),
		nextID:    1000,
	}
}

func (e *TerraformEngine) AddResource(r Resource) {
	addr := fmt.Sprintf("%s.%s", r.Type, r.Name)
	e.resources[addr] = &r
	fmt.Printf("  + %s defined\n", addr)
}

func (e *TerraformEngine) UpdateResource(addr string, changes map[string]string) {
	r, exists := e.resources[addr]
	if !exists {
		fmt.Printf("  Resource %s not found\n", addr)
		return
	}
	for k, v := range changes {
		old := r.Config[k]
		r.Config[k] = v
		fmt.Printf("  ~ %s: %s = %s -> %s\n", addr, k, old, v)
	}
}

func (e *TerraformEngine) PrintGraph() {
	order := e.topologicalSort()
	fmt.Println("Execution order (topological sort):")
	for i, addr := range order {
		r := e.resources[addr]
		deps := "none"
		if len(r.DependsOn) > 0 {
			deps = strings.Join(r.DependsOn, ", ")
		}
		fmt.Printf("  %d. %s (depends on: %s)\n", i+1, addr, deps)
	}

	fmt.Println("\nGraph visualization:")
	fmt.Println("  aws_vpc.main")
	fmt.Println("    |--- aws_subnet.public")
	fmt.Println("    |     |--- aws_instance.web_server")
	fmt.Println("    |--- aws_subnet.private")
	fmt.Println("    |     |--- aws_rds_instance.database")
	fmt.Println("    |--- aws_security_group.web")
	fmt.Println("          |--- aws_instance.web_server")
	fmt.Println("          |--- aws_rds_instance.database")
}

func (e *TerraformEngine) Plan() []PlanItem {
	order := e.topologicalSort()
	var plan []PlanItem

	for _, addr := range order {
		r := e.resources[addr]
		existing, hasState := e.state[addr]

		if !hasState {
			plan = append(plan, PlanItem{
				Address: addr,
				Action:  ActionCreate,
				After:   r.Config,
			})
		} else {
			changed := false
			for k, v := range r.Config {
				if existing.Attributes[k] != v {
					changed = true
					break
				}
			}
			if changed {
				plan = append(plan, PlanItem{
					Address: addr,
					Action:  ActionUpdate,
					Before:  existing.Attributes,
					After:   r.Config,
				})
			}
		}
	}

	printPlanSummary(plan)
	return plan
}

func printPlanSummary(plan []PlanItem) {
	var creates, updates, destroys int
	for _, item := range plan {
		symbol := ""
		switch item.Action {
		case ActionCreate:
			symbol = "+"
			creates++
		case ActionUpdate:
			symbol = "~"
			updates++
		case ActionDestroy:
			symbol = "-"
			destroys++
		}
		fmt.Printf("  %s %s (%s)\n", symbol, item.Address, item.Action)

		if item.Action == ActionCreate {
			for k, v := range item.After {
				fmt.Printf("      + %s = %q\n", k, v)
			}
		}
		if item.Action == ActionUpdate {
			for k, v := range item.After {
				if item.Before[k] != v {
					fmt.Printf("      ~ %s = %q -> %q\n", k, item.Before[k], v)
				}
			}
		}
	}
	fmt.Printf("\nPlan: %d to add, %d to change, %d to destroy.\n",
		creates, updates, destroys)
}

func (e *TerraformEngine) Apply(plan []PlanItem) {
	e.mu.Lock()
	defer e.mu.Unlock()

	fmt.Printf("Applying %d changes...\n", len(plan))

	for _, item := range plan {
		switch item.Action {
		case ActionCreate:
			e.nextID++
			id := fmt.Sprintf("i-%08d", e.nextID)
			e.state[item.Address] = &ResourceState{
				Address:    item.Address,
				Type:       strings.SplitN(item.Address, ".", 2)[0],
				Name:       strings.SplitN(item.Address, ".", 2)[1],
				ID:         id,
				Attributes: copyMap(item.After),
				Status:     StatusCreated,
			}
			fmt.Printf("  %s: Creating... [id=%s]\n", item.Address, id)
			time.Sleep(20 * time.Millisecond)
			fmt.Printf("  %s: Created\n", item.Address)

		case ActionUpdate:
			state := e.state[item.Address]
			for k, v := range item.After {
				state.Attributes[k] = v
			}
			state.Status = StatusCreated
			fmt.Printf("  %s: Modifying... [id=%s]\n", item.Address, state.ID)
			time.Sleep(20 * time.Millisecond)
			fmt.Printf("  %s: Modified\n", item.Address)
		}
	}

	fmt.Println("Apply complete!")
}

func (e *TerraformEngine) Destroy() {
	e.mu.Lock()
	defer e.mu.Unlock()

	order := e.topologicalSort()
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {
		order[i], order[j] = order[j], order[i]
	}

	fmt.Printf("Destroying %d resources (reverse order)...\n", len(order))
	for _, addr := range order {
		state, exists := e.state[addr]
		if !exists {
			continue
		}
		fmt.Printf("  %s: Destroying... [id=%s]\n", addr, state.ID)
		time.Sleep(15 * time.Millisecond)
		fmt.Printf("  %s: Destroyed\n", addr)
		delete(e.state, addr)
	}
	fmt.Println("Destroy complete! All resources removed.")
}

func (e *TerraformEngine) PrintState() {
	fmt.Println("Current State:")
	fmt.Printf("  %-30s %-15s %-12s %s\n", "ADDRESS", "ID", "STATUS", "KEY ATTRS")

	var addrs []string
	for addr := range e.state {
		addrs = append(addrs, addr)
	}
	sort.Strings(addrs)

	for _, addr := range addrs {
		s := e.state[addr]
		attrs := ""
		if name, ok := s.Attributes["name"]; ok {
			attrs = "name=" + name
		}
		if itype, ok := s.Attributes["instance_type"]; ok {
			if attrs != "" {
				attrs += ", "
			}
			attrs += "type=" + itype
		}
		fmt.Printf("  %-30s %-15s %-12s %s\n",
			s.Address, s.ID, string(s.Status), attrs)
	}
	fmt.Printf("  Total: %d resources managed\n", len(e.state))
}

func (e *TerraformEngine) topologicalSort() []string {
	if len(e.order) > 0 && len(e.order) == len(e.resources) {
		return e.order
	}

	inDegree := make(map[string]int)
	graph := make(map[string][]string)

	for addr := range e.resources {
		inDegree[addr] = 0
	}

	for addr, r := range e.resources {
		for _, dep := range r.DependsOn {
			graph[dep] = append(graph[dep], addr)
			inDegree[addr]++
		}
	}

	var queue []string
	for addr, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, addr)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		neighbors := graph[node]
		sort.Strings(neighbors)
		for _, next := range neighbors {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
		sort.Strings(queue)
	}

	e.order = order
	return order
}

func copyMap(m map[string]string) map[string]string {
	cp := make(map[string]string, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func init() {
	_ = os.Stdout
}
