// Package main - Chapter 128: TiDB Internals
package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== CHAPTER 128: TiDB - DISTRIBUTED SQL DATABASE IN GO ===")

	// ============================================
	// WHAT IS TiDB?
	// ============================================
	fmt.Println(`
==============================
WHAT IS TiDB?
==============================

TiDB (Titanium Database) is a distributed SQL database built in Go.
It is MySQL-compatible, meaning applications using MySQL can switch
to TiDB with minimal changes while gaining horizontal scalability.

KEY FEATURES:
  - MySQL compatible (wire protocol and SQL syntax)
  - Horizontal scaling (add nodes, no sharding needed)
  - Distributed ACID transactions
  - Real-time HTAP (hybrid transactional/analytical processing)
  - Cloud-native architecture

ARCHITECTURE:
  +------------------+     +------------------+
  |   Application    |---->|   TiDB Server    |
  |   (MySQL client) |     |   (SQL Layer)    |
  +------------------+     +--------+---------+
                                    |
                    +---------------+---------------+
                    |                               |
           +--------v---------+           +--------v---------+
           |   TiKV Cluster   |           |   TiFlash        |
           |   (Row Store)    |           |   (Column Store) |
           |   Raft Consensus |           |   Analytics      |
           +--------+---------+           +------------------+
                    |
           +--------v---------+
           |   PD (Placement  |
           |     Driver)      |
           |   Cluster Mgmt   |
           +------------------+

COMPONENTS:
  TiDB Server:  Stateless SQL layer (parser, optimizer, executor)
  TiKV:         Distributed key-value store with Raft consensus
  PD:           Cluster manager, timestamp oracle, data placement
  TiFlash:      Columnar storage for analytical queries

WHY GO:
  TiDB server and PD are written in Go. TiKV is in Rust for
  performance. Go provides fast compilation, goroutines for
  handling thousands of concurrent connections, and excellent
  networking libraries for distributed systems.`)

	// ============================================
	// SQL TOKENIZER AND LEXER
	// ============================================
	fmt.Println("\n--- SQL Tokenizer ---")
	demoSQLTokenizer()

	// ============================================
	// SQL PARSER
	// ============================================
	fmt.Println("\n--- SQL Parser ---")
	demoSQLParser()

	// ============================================
	// QUERY PLANNER
	// ============================================
	fmt.Println("\n--- Query Planner ---")
	demoQueryPlanner()

	// ============================================
	// RAFT CONSENSUS
	// ============================================
	fmt.Println("\n--- Raft Consensus ---")
	demoRaftConsensus()

	// ============================================
	// DISTRIBUTED KEY-VALUE STORE
	// ============================================
	fmt.Println("\n--- Distributed KV Store ---")
	demoDistributedKV()

	// ============================================
	// FULL MINI-TIDB
	// ============================================
	fmt.Println("\n--- Mini-TiDB Integration ---")
	demoMiniTiDB()
}

// ============================================
// SQL TOKENIZER AND LEXER
// ============================================

type TokenType int

const (
	TokenKeyword TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString
	TokenOperator
	TokenPunctuation
	TokenStar
	TokenEOF
)

func (t TokenType) String() string {
	names := []string{"KEYWORD", "IDENT", "NUMBER", "STRING", "OP", "PUNCT", "STAR", "EOF"}
	if int(t) < len(names) {
		return names[t]
	}
	return "UNKNOWN"
}

type Token struct {
	Type    TokenType
	Value   string
	Pos     int
}

var sqlKeywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true,
	"INTO": true, "VALUES": true, "UPDATE": true, "SET": true,
	"DELETE": true, "CREATE": true, "TABLE": true, "DROP": true,
	"ALTER": true, "INDEX": true, "ON": true, "AND": true,
	"OR": true, "NOT": true, "NULL": true, "JOIN": true,
	"LEFT": true, "RIGHT": true, "INNER": true, "OUTER": true,
	"GROUP": true, "BY": true, "ORDER": true, "ASC": true,
	"DESC": true, "LIMIT": true, "OFFSET": true, "AS": true,
	"HAVING": true, "DISTINCT": true, "IN": true, "BETWEEN": true,
	"LIKE": true, "IS": true, "INT": true, "VARCHAR": true,
	"TEXT": true, "BOOLEAN": true, "PRIMARY": true, "KEY": true,
	"AUTO_INCREMENT": true, "DEFAULT": true, "NOT_NULL": true,
}

type SQLTokenizer struct {
	input   string
	pos     int
	tokens  []Token
}

func NewSQLTokenizer(input string) *SQLTokenizer {
	return &SQLTokenizer{input: input}
}

func (t *SQLTokenizer) Tokenize() []Token {
	for t.pos < len(t.input) {
		ch := t.input[t.pos]

		switch {
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			t.pos++
		case ch == '*':
			t.tokens = append(t.tokens, Token{TokenStar, "*", t.pos})
			t.pos++
		case isLetter(ch) || ch == '_':
			t.readIdentifier()
		case isDigit(ch):
			t.readNumber()
		case ch == '\'':
			t.readString()
		case ch == '(' || ch == ')' || ch == ',' || ch == ';':
			t.tokens = append(t.tokens, Token{TokenPunctuation, string(ch), t.pos})
			t.pos++
		case ch == '=' || ch == '>' || ch == '<' || ch == '!':
			t.readOperator()
		default:
			t.pos++
		}
	}

	t.tokens = append(t.tokens, Token{TokenEOF, "", t.pos})
	return t.tokens
}

func (t *SQLTokenizer) readIdentifier() {
	start := t.pos
	for t.pos < len(t.input) && (isLetter(t.input[t.pos]) || isDigit(t.input[t.pos]) || t.input[t.pos] == '_') {
		t.pos++
	}
	value := t.input[start:t.pos]
	if sqlKeywords[strings.ToUpper(value)] {
		t.tokens = append(t.tokens, Token{TokenKeyword, strings.ToUpper(value), start})
	} else {
		t.tokens = append(t.tokens, Token{TokenIdentifier, value, start})
	}
}

func (t *SQLTokenizer) readNumber() {
	start := t.pos
	for t.pos < len(t.input) && (isDigit(t.input[t.pos]) || t.input[t.pos] == '.') {
		t.pos++
	}
	t.tokens = append(t.tokens, Token{TokenNumber, t.input[start:t.pos], start})
}

func (t *SQLTokenizer) readString() {
	t.pos++
	start := t.pos
	for t.pos < len(t.input) && t.input[t.pos] != '\'' {
		t.pos++
	}
	value := t.input[start:t.pos]
	if t.pos < len(t.input) {
		t.pos++
	}
	t.tokens = append(t.tokens, Token{TokenString, value, start})
}

func (t *SQLTokenizer) readOperator() {
	start := t.pos
	t.pos++
	if t.pos < len(t.input) && t.input[t.pos] == '=' {
		t.pos++
	}
	t.tokens = append(t.tokens, Token{TokenOperator, t.input[start:t.pos], start})
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func demoSQLTokenizer() {
	queries := []string{
		"SELECT * FROM users WHERE age > 25 AND name = 'Alice'",
		"INSERT INTO orders (user_id, amount) VALUES (1, 99.99)",
		"SELECT u.name, COUNT(*) FROM users u JOIN orders o ON u.id = o.user_id GROUP BY u.name",
	}

	for _, query := range queries {
		fmt.Printf("  Query: %s\n", query)
		tokenizer := NewSQLTokenizer(query)
		tokens := tokenizer.Tokenize()
		fmt.Print("  Tokens: ")
		for _, tok := range tokens {
			if tok.Type == TokenEOF {
				break
			}
			os.Stdout.WriteString(fmt.Sprintf("[%s:%s] ", tok.Type, tok.Value))
		}
		fmt.Println()
	}
}

// ============================================
// SQL PARSER (AST)
// ============================================

type ASTNodeType string

const (
	NodeSelect ASTNodeType = "SELECT"
	NodeInsert ASTNodeType = "INSERT"
	NodeCreate ASTNodeType = "CREATE"
	NodeWhere  ASTNodeType = "WHERE"
	NodeJoin   ASTNodeType = "JOIN"
)

type ASTNode struct {
	Type     ASTNodeType
	Table    string
	Columns  []string
	Where    *WhereClause
	Joins    []JoinClause
	OrderBy  []OrderByClause
	GroupBy  []string
	Limit    int
	Values   [][]string
}

type WhereClause struct {
	Column   string
	Operator string
	Value    string
	And      *WhereClause
	Or       *WhereClause
}

type JoinClause struct {
	Type      string
	Table     string
	OnLeft    string
	OnRight   string
}

type OrderByClause struct {
	Column string
	Desc   bool
}

type SQLParser struct {
	tokens  []Token
	pos     int
}

func NewSQLParser(tokens []Token) *SQLParser {
	return &SQLParser{tokens: tokens}
}

func (p *SQLParser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{Type: TokenEOF}
}

func (p *SQLParser) advance() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *SQLParser) expect(value string) Token {
	tok := p.advance()
	if strings.ToUpper(tok.Value) != value {
		return tok
	}
	return tok
}

func (p *SQLParser) Parse() *ASTNode {
	tok := p.peek()
	switch strings.ToUpper(tok.Value) {
	case "SELECT":
		return p.parseSelect()
	case "INSERT":
		return p.parseInsert()
	default:
		return nil
	}
}

func (p *SQLParser) parseSelect() *ASTNode {
	p.advance() // consume SELECT
	node := &ASTNode{Type: NodeSelect}

	// parse columns
	for {
		tok := p.advance()
		if tok.Type == TokenStar {
			node.Columns = append(node.Columns, "*")
		} else {
			node.Columns = append(node.Columns, tok.Value)
		}
		if p.peek().Value != "," {
			break
		}
		p.advance() // consume comma
	}

	// FROM
	if strings.ToUpper(p.peek().Value) == "FROM" {
		p.advance()
		node.Table = p.advance().Value
		// optional alias
		if p.peek().Type == TokenIdentifier && !sqlKeywords[strings.ToUpper(p.peek().Value)] {
			p.advance()
		}
	}

	// JOIN
	for strings.ToUpper(p.peek().Value) == "JOIN" || strings.ToUpper(p.peek().Value) == "LEFT" ||
		strings.ToUpper(p.peek().Value) == "RIGHT" || strings.ToUpper(p.peek().Value) == "INNER" {
		join := p.parseJoin()
		node.Joins = append(node.Joins, join)
	}

	// WHERE
	if strings.ToUpper(p.peek().Value) == "WHERE" {
		p.advance()
		node.Where = p.parseWhere()
	}

	// GROUP BY
	if strings.ToUpper(p.peek().Value) == "GROUP" {
		p.advance() // GROUP
		p.advance() // BY
		for {
			node.GroupBy = append(node.GroupBy, p.advance().Value)
			if p.peek().Value != "," {
				break
			}
			p.advance()
		}
	}

	// ORDER BY
	if strings.ToUpper(p.peek().Value) == "ORDER" {
		p.advance() // ORDER
		p.advance() // BY
		for {
			col := p.advance().Value
			desc := false
			if strings.ToUpper(p.peek().Value) == "DESC" {
				desc = true
				p.advance()
			} else if strings.ToUpper(p.peek().Value) == "ASC" {
				p.advance()
			}
			node.OrderBy = append(node.OrderBy, OrderByClause{Column: col, Desc: desc})
			if p.peek().Value != "," {
				break
			}
			p.advance()
		}
	}

	return node
}

func (p *SQLParser) parseInsert() *ASTNode {
	p.advance() // INSERT
	p.expect("INTO")
	node := &ASTNode{Type: NodeInsert}
	node.Table = p.advance().Value

	// columns
	if p.peek().Value == "(" {
		p.advance()
		for {
			node.Columns = append(node.Columns, p.advance().Value)
			if p.peek().Value == ")" {
				p.advance()
				break
			}
			p.advance() // comma
		}
	}

	// VALUES
	p.expect("VALUES")
	if p.peek().Value == "(" {
		p.advance()
		var row []string
		for {
			row = append(row, p.advance().Value)
			if p.peek().Value == ")" {
				p.advance()
				break
			}
			p.advance()
		}
		node.Values = append(node.Values, row)
	}

	return node
}

func (p *SQLParser) parseJoin() JoinClause {
	joinType := "INNER"
	tok := p.peek()
	if strings.ToUpper(tok.Value) == "LEFT" || strings.ToUpper(tok.Value) == "RIGHT" {
		joinType = strings.ToUpper(tok.Value)
		p.advance()
	} else if strings.ToUpper(tok.Value) == "INNER" {
		p.advance()
	}
	p.advance() // JOIN

	table := p.advance().Value
	// optional alias
	if p.peek().Type == TokenIdentifier && !sqlKeywords[strings.ToUpper(p.peek().Value)] {
		p.advance()
	}

	p.expect("ON")
	left := p.advance().Value
	p.advance() // = operator
	right := p.advance().Value

	return JoinClause{Type: joinType, Table: table, OnLeft: left, OnRight: right}
}

func (p *SQLParser) parseWhere() *WhereClause {
	col := p.advance().Value
	op := p.advance().Value
	val := p.advance().Value

	clause := &WhereClause{Column: col, Operator: op, Value: val}

	if strings.ToUpper(p.peek().Value) == "AND" {
		p.advance()
		clause.And = p.parseWhere()
	} else if strings.ToUpper(p.peek().Value) == "OR" {
		p.advance()
		clause.Or = p.parseWhere()
	}

	return clause
}

func printAST(node *ASTNode, indent string) {
	if node == nil {
		return
	}
	os.Stdout.WriteString(fmt.Sprintf("%sType: %s\n", indent, node.Type))
	os.Stdout.WriteString(fmt.Sprintf("%sTable: %s\n", indent, node.Table))
	os.Stdout.WriteString(fmt.Sprintf("%sColumns: %v\n", indent, node.Columns))
	if node.Where != nil {
		os.Stdout.WriteString(fmt.Sprintf("%sWhere: %s %s %s\n", indent, node.Where.Column, node.Where.Operator, node.Where.Value))
		if node.Where.And != nil {
			os.Stdout.WriteString(fmt.Sprintf("%s  AND: %s %s %s\n", indent, node.Where.And.Column, node.Where.And.Operator, node.Where.And.Value))
		}
	}
	for _, j := range node.Joins {
		os.Stdout.WriteString(fmt.Sprintf("%sJoin: %s %s ON %s = %s\n", indent, j.Type, j.Table, j.OnLeft, j.OnRight))
	}
	if len(node.GroupBy) > 0 {
		os.Stdout.WriteString(fmt.Sprintf("%sGroup By: %v\n", indent, node.GroupBy))
	}
	if len(node.OrderBy) > 0 {
		for _, o := range node.OrderBy {
			dir := "ASC"
			if o.Desc {
				dir = "DESC"
			}
			os.Stdout.WriteString(fmt.Sprintf("%sOrder By: %s %s\n", indent, o.Column, dir))
		}
	}
	if len(node.Values) > 0 {
		os.Stdout.WriteString(fmt.Sprintf("%sValues: %v\n", indent, node.Values))
	}
}

func demoSQLParser() {
	queries := []string{
		"SELECT * FROM users WHERE age > 25 AND name = 'Alice'",
		"INSERT INTO orders (user_id, amount) VALUES (1, 99.99)",
	}

	for _, query := range queries {
		fmt.Printf("  Parsing: %s\n", query)
		tokenizer := NewSQLTokenizer(query)
		tokens := tokenizer.Tokenize()
		parser := NewSQLParser(tokens)
		ast := parser.Parse()
		printAST(ast, "    ")
		fmt.Println()
	}
}

// ============================================
// QUERY PLANNER (COST-BASED)
// ============================================

type PlanNodeType string

const (
	PlanTableScan  PlanNodeType = "TableScan"
	PlanIndexScan  PlanNodeType = "IndexScan"
	PlanFilter     PlanNodeType = "Filter"
	PlanProject    PlanNodeType = "Projection"
	PlanHashJoin   PlanNodeType = "HashJoin"
	PlanSort       PlanNodeType = "Sort"
	PlanLimit      PlanNodeType = "Limit"
	PlanAggregate  PlanNodeType = "Aggregate"
)

type PlanNode struct {
	Type       PlanNodeType
	Table      string
	Columns    []string
	Condition  string
	Cost       float64
	RowCount   int
	Children   []*PlanNode
}

type TableStats struct {
	Name      string
	RowCount  int
	Indexes   []IndexInfo
	AvgRowLen int
}

type IndexInfo struct {
	Name      string
	Columns   []string
	Unique    bool
	Cardinality int
}

type QueryPlanner struct {
	stats map[string]*TableStats
}

func NewQueryPlanner() *QueryPlanner {
	return &QueryPlanner{
		stats: map[string]*TableStats{
			"users": {
				Name: "users", RowCount: 100000, AvgRowLen: 128,
				Indexes: []IndexInfo{
					{Name: "PRIMARY", Columns: []string{"id"}, Unique: true, Cardinality: 100000},
					{Name: "idx_age", Columns: []string{"age"}, Unique: false, Cardinality: 80},
					{Name: "idx_name", Columns: []string{"name"}, Unique: false, Cardinality: 95000},
				},
			},
			"orders": {
				Name: "orders", RowCount: 500000, AvgRowLen: 64,
				Indexes: []IndexInfo{
					{Name: "PRIMARY", Columns: []string{"id"}, Unique: true, Cardinality: 500000},
					{Name: "idx_user_id", Columns: []string{"user_id"}, Unique: false, Cardinality: 100000},
					{Name: "idx_created", Columns: []string{"created_at"}, Unique: false, Cardinality: 365},
				},
			},
		},
	}
}

func (qp *QueryPlanner) Plan(ast *ASTNode) *PlanNode {
	if ast == nil {
		return nil
	}

	switch ast.Type {
	case NodeSelect:
		return qp.planSelect(ast)
	default:
		return nil
	}
}

func (qp *QueryPlanner) planSelect(ast *ASTNode) *PlanNode {
	stats := qp.stats[ast.Table]
	if stats == nil {
		stats = &TableStats{Name: ast.Table, RowCount: 1000, AvgRowLen: 64}
	}

	var scan *PlanNode
	if ast.Where != nil {
		idx := qp.findBestIndex(stats, ast.Where.Column)
		if idx != nil {
			selectivity := float64(idx.Cardinality) / float64(stats.RowCount)
			estimatedRows := int(float64(stats.RowCount) * selectivity)
			scan = &PlanNode{
				Type:      PlanIndexScan,
				Table:     ast.Table,
				Condition: fmt.Sprintf("index=%s, col=%s", idx.Name, ast.Where.Column),
				Cost:      float64(estimatedRows) * 1.0,
				RowCount:  estimatedRows,
			}
		}
	}

	if scan == nil {
		scan = &PlanNode{
			Type:     PlanTableScan,
			Table:    ast.Table,
			Cost:     float64(stats.RowCount) * float64(stats.AvgRowLen) / 1024.0,
			RowCount: stats.RowCount,
		}
	}

	current := scan

	if ast.Where != nil {
		filter := &PlanNode{
			Type:      PlanFilter,
			Condition: fmt.Sprintf("%s %s %s", ast.Where.Column, ast.Where.Operator, ast.Where.Value),
			Cost:      current.Cost * 0.1,
			RowCount:  current.RowCount / 10,
			Children:  []*PlanNode{current},
		}
		current = filter
	}

	for _, j := range ast.Joins {
		joinStats := qp.stats[j.Table]
		if joinStats == nil {
			joinStats = &TableStats{Name: j.Table, RowCount: 1000, AvgRowLen: 64}
		}

		joinScan := &PlanNode{
			Type:     PlanTableScan,
			Table:    j.Table,
			Cost:     float64(joinStats.RowCount) * float64(joinStats.AvgRowLen) / 1024.0,
			RowCount: joinStats.RowCount,
		}

		hashJoin := &PlanNode{
			Type:      PlanHashJoin,
			Condition: fmt.Sprintf("%s = %s", j.OnLeft, j.OnRight),
			Cost:      current.Cost + joinScan.Cost + float64(current.RowCount)*0.01,
			RowCount:  current.RowCount,
			Children:  []*PlanNode{current, joinScan},
		}
		current = hashJoin
	}

	if len(ast.GroupBy) > 0 {
		agg := &PlanNode{
			Type:     PlanAggregate,
			Columns:  ast.GroupBy,
			Cost:     current.Cost + float64(current.RowCount)*0.05,
			RowCount: current.RowCount / 100,
			Children: []*PlanNode{current},
		}
		current = agg
	}

	if len(ast.OrderBy) > 0 {
		sortNode := &PlanNode{
			Type:     PlanSort,
			Columns:  []string{ast.OrderBy[0].Column},
			Cost:     current.Cost + float64(current.RowCount)*0.02,
			RowCount: current.RowCount,
			Children: []*PlanNode{current},
		}
		current = sortNode
	}

	if len(ast.Columns) > 0 && ast.Columns[0] != "*" {
		proj := &PlanNode{
			Type:     PlanProject,
			Columns:  ast.Columns,
			Cost:     current.Cost * 1.01,
			RowCount: current.RowCount,
			Children: []*PlanNode{current},
		}
		current = proj
	}

	return current
}

func (qp *QueryPlanner) findBestIndex(stats *TableStats, column string) *IndexInfo {
	for i := range stats.Indexes {
		for _, col := range stats.Indexes[i].Columns {
			if col == column {
				return &stats.Indexes[i]
			}
		}
	}
	return nil
}

func printPlan(node *PlanNode, indent string) {
	if node == nil {
		return
	}
	os.Stdout.WriteString(fmt.Sprintf("%s%s", indent, node.Type))
	if node.Table != "" {
		os.Stdout.WriteString(fmt.Sprintf(" [table=%s]", node.Table))
	}
	if node.Condition != "" {
		os.Stdout.WriteString(fmt.Sprintf(" [%s]", node.Condition))
	}
	if len(node.Columns) > 0 {
		os.Stdout.WriteString(fmt.Sprintf(" cols=%v", node.Columns))
	}
	os.Stdout.WriteString(fmt.Sprintf(" (cost=%.1f, rows=%d)\n", node.Cost, node.RowCount))
	for _, child := range node.Children {
		printPlan(child, indent+"  ")
	}
}

func demoQueryPlanner() {
	planner := NewQueryPlanner()

	queries := []string{
		"SELECT * FROM users WHERE age > 25",
		"SELECT name FROM users WHERE id = 42",
	}

	for _, query := range queries {
		fmt.Printf("  Query: %s\n", query)
		tokenizer := NewSQLTokenizer(query)
		tokens := tokenizer.Tokenize()
		parser := NewSQLParser(tokens)
		ast := parser.Parse()
		plan := planner.Plan(ast)
		fmt.Println("  Execution Plan:")
		printPlan(plan, "    ")
		fmt.Println()
	}
}

// ============================================
// RAFT CONSENSUS
// ============================================

type RaftState int

const (
	Follower RaftState = iota
	Candidate
	Leader
)

func (s RaftState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

type LogEntry struct {
	Term    int
	Index   int
	Command string
	Key     string
	Value   string
}

type RaftNode struct {
	mu          sync.Mutex
	id          int
	state       RaftState
	currentTerm int
	votedFor    int
	log         []LogEntry
	commitIndex int
	lastApplied int
	peers       []*RaftNode
	votesReceived int
	kvStore     map[string]string
	isAlive     bool
}

func NewRaftNode(id int) *RaftNode {
	return &RaftNode{
		id:       id,
		state:    Follower,
		votedFor: -1,
		kvStore:  make(map[string]string),
		isAlive:  true,
	}
}

func (n *RaftNode) RequestVote(candidateTerm, candidateID, lastLogIndex, lastLogTerm int) (int, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isAlive {
		return n.currentTerm, false
	}

	if candidateTerm > n.currentTerm {
		n.currentTerm = candidateTerm
		n.state = Follower
		n.votedFor = -1
	}

	if candidateTerm < n.currentTerm {
		return n.currentTerm, false
	}

	if n.votedFor == -1 || n.votedFor == candidateID {
		n.votedFor = candidateID
		return n.currentTerm, true
	}

	return n.currentTerm, false
}

func (n *RaftNode) AppendEntries(leaderTerm, leaderID int, entries []LogEntry) (int, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.isAlive {
		return n.currentTerm, false
	}

	if leaderTerm < n.currentTerm {
		return n.currentTerm, false
	}

	n.currentTerm = leaderTerm
	n.state = Follower
	n.votedFor = leaderID

	for _, entry := range entries {
		if entry.Index > len(n.log) {
			n.log = append(n.log, entry)
		}
	}

	return n.currentTerm, true
}

func (n *RaftNode) startElection() bool {
	n.mu.Lock()
	n.currentTerm++
	n.state = Candidate
	n.votedFor = n.id
	n.votesReceived = 1
	term := n.currentTerm
	lastLogIndex := len(n.log)
	lastLogTerm := 0
	if lastLogIndex > 0 {
		lastLogTerm = n.log[lastLogIndex-1].Term
	}
	n.mu.Unlock()

	votes := 1
	for _, peer := range n.peers {
		_, granted := peer.RequestVote(term, n.id, lastLogIndex, lastLogTerm)
		if granted {
			votes++
		}
	}

	majority := (len(n.peers)+1)/2 + 1
	if votes >= majority {
		n.mu.Lock()
		n.state = Leader
		n.mu.Unlock()
		return true
	}

	n.mu.Lock()
	n.state = Follower
	n.mu.Unlock()
	return false
}

func (n *RaftNode) replicateLog(entry LogEntry) int {
	n.mu.Lock()
	n.log = append(n.log, entry)
	n.mu.Unlock()

	successes := 1
	for _, peer := range n.peers {
		_, ok := peer.AppendEntries(n.currentTerm, n.id, []LogEntry{entry})
		if ok {
			successes++
		}
	}

	majority := (len(n.peers)+1)/2 + 1
	if successes >= majority {
		n.mu.Lock()
		n.commitIndex = entry.Index
		if entry.Command == "SET" {
			n.kvStore[entry.Key] = entry.Value
		}
		for _, peer := range n.peers {
			peer.mu.Lock()
			peer.commitIndex = entry.Index
			if entry.Command == "SET" {
				peer.kvStore[entry.Key] = entry.Value
			}
			peer.mu.Unlock()
		}
		n.mu.Unlock()
	}

	return successes
}

func demoRaftConsensus() {
	nodes := make([]*RaftNode, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = NewRaftNode(i)
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			if i != j {
				nodes[i].peers = append(nodes[i].peers, nodes[j])
			}
		}
	}

	fmt.Println("  Starting election for node 0...")
	elected := nodes[0].startElection()
	fmt.Printf("  Node 0 elected leader: %v (state: %s, term: %d)\n",
		elected, nodes[0].state, nodes[0].currentTerm)

	if elected {
		entries := []struct {
			key, value string
		}{
			{"user:1", "Alice"},
			{"user:2", "Bob"},
			{"user:3", "Charlie"},
		}

		for i, e := range entries {
			entry := LogEntry{
				Term:    nodes[0].currentTerm,
				Index:   i + 1,
				Command: "SET",
				Key:     e.key,
				Value:   e.value,
			}
			acks := nodes[0].replicateLog(entry)
			os.Stdout.WriteString(fmt.Sprintf("  Replicated %s=%s to %d/%d nodes\n", e.key, e.value, acks, len(nodes)))
		}

		fmt.Println("  KV store across nodes:")
		for _, n := range nodes {
			n.mu.Lock()
			os.Stdout.WriteString(fmt.Sprintf("    Node %d (%s): %v (commit=%d)\n",
				n.id, n.state, n.kvStore, n.commitIndex))
			n.mu.Unlock()
		}

		fmt.Println("  Simulating node 3 failure...")
		nodes[3].mu.Lock()
		nodes[3].isAlive = false
		nodes[3].mu.Unlock()

		entry := LogEntry{
			Term:    nodes[0].currentTerm,
			Index:   4,
			Command: "SET",
			Key:     "user:4",
			Value:   "Diana",
		}
		acks := nodes[0].replicateLog(entry)
		os.Stdout.WriteString(fmt.Sprintf("  Replicated with node down: %d/%d acks (still majority)\n", acks, len(nodes)))
	}
}

// ============================================
// DISTRIBUTED KEY-VALUE STORE (TiKV-like)
// ============================================

type Region struct {
	mu       sync.RWMutex
	ID       int
	StartKey string
	EndKey   string
	Data     map[string]string
	Leader   int
	Replicas []int
}

type TiKVNode struct {
	mu      sync.RWMutex
	ID      int
	Regions map[int]*Region
}

type PlacementDriver struct {
	mu       sync.RWMutex
	nodes    map[int]*TiKVNode
	regions  []*Region
	nextTS   uint64
	regionID int
}

func NewPlacementDriver() *PlacementDriver {
	return &PlacementDriver{
		nodes:   make(map[int]*TiKVNode),
		nextTS:  uint64(time.Now().UnixNano()),
	}
}

func (pd *PlacementDriver) AddNode(id int) *TiKVNode {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	node := &TiKVNode{
		ID:      id,
		Regions: make(map[int]*Region),
	}
	pd.nodes[id] = node
	return node
}

func (pd *PlacementDriver) CreateRegion(startKey, endKey string) *Region {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	pd.regionID++
	region := &Region{
		ID:       pd.regionID,
		StartKey: startKey,
		EndKey:   endKey,
		Data:     make(map[string]string),
	}

	var nodeIDs []int
	for id := range pd.nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Ints(nodeIDs)

	if len(nodeIDs) > 0 {
		region.Leader = nodeIDs[0]
		for i := 0; i < len(nodeIDs) && i < 3; i++ {
			region.Replicas = append(region.Replicas, nodeIDs[i])
			pd.nodes[nodeIDs[i]].Regions[region.ID] = region
		}
	}

	pd.regions = append(pd.regions, region)
	return region
}

func (pd *PlacementDriver) GetTimestamp() uint64 {
	pd.mu.Lock()
	defer pd.mu.Unlock()
	pd.nextTS++
	return pd.nextTS
}

func (pd *PlacementDriver) FindRegion(key string) *Region {
	pd.mu.RLock()
	defer pd.mu.RUnlock()

	for _, r := range pd.regions {
		if key >= r.StartKey && (r.EndKey == "" || key < r.EndKey) {
			return r
		}
	}
	return nil
}

func (pd *PlacementDriver) Put(key, value string) error {
	region := pd.FindRegion(key)
	if region == nil {
		return fmt.Errorf("no region found for key: %s", key)
	}

	region.mu.Lock()
	defer region.mu.Unlock()
	region.Data[key] = value
	return nil
}

func (pd *PlacementDriver) Get(key string) (string, bool) {
	region := pd.FindRegion(key)
	if region == nil {
		return "", false
	}

	region.mu.RLock()
	defer region.mu.RUnlock()
	val, ok := region.Data[key]
	return val, ok
}

func demoDistributedKV() {
	pd := NewPlacementDriver()

	for i := 0; i < 3; i++ {
		pd.AddNode(i)
	}

	pd.CreateRegion("a", "m")
	pd.CreateRegion("m", "z")

	data := map[string]string{
		"alice":   "engineer",
		"bob":     "designer",
		"charlie": "manager",
		"nancy":   "architect",
		"peter":   "developer",
	}

	for k, v := range data {
		err := pd.Put(k, v)
		if err != nil {
			fmt.Printf("  Error putting %s: %v\n", k, err)
			continue
		}
	}

	for k := range data {
		val, ok := pd.Get(k)
		if ok {
			os.Stdout.WriteString(fmt.Sprintf("  Get(%s) = %s\n", k, val))
		}
	}

	fmt.Printf("  Timestamp oracle: %d\n", pd.GetTimestamp())
	fmt.Printf("  Timestamp oracle: %d\n", pd.GetTimestamp())

	fmt.Println("  Region distribution:")
	for _, region := range pd.regions {
		region.mu.RLock()
		os.Stdout.WriteString(fmt.Sprintf("    Region %d [%s, %s): %d keys, leader=node%d, replicas=%v\n",
			region.ID, region.StartKey, region.EndKey, len(region.Data), region.Leader, region.Replicas))
		region.mu.RUnlock()
	}
}

// ============================================
// FULL MINI-TIDB
// ============================================

type MiniTiDB struct {
	pd      *PlacementDriver
	planner *QueryPlanner
	tables  map[string]*TableDef
	mu      sync.RWMutex
}

type TableDef struct {
	Name    string
	Columns []ColumnDef
	Rows    []map[string]string
	NextID  int
}

type ColumnDef struct {
	Name     string
	Type     string
	Primary  bool
	AutoInc  bool
}

func NewMiniTiDB() *MiniTiDB {
	pd := NewPlacementDriver()
	for i := 0; i < 3; i++ {
		pd.AddNode(i)
	}
	pd.CreateRegion("", "")

	return &MiniTiDB{
		pd:      pd,
		planner: NewQueryPlanner(),
		tables:  make(map[string]*TableDef),
	}
}

func (db *MiniTiDB) Execute(query string) string {
	tokenizer := NewSQLTokenizer(query)
	tokens := tokenizer.Tokenize()

	if len(tokens) == 0 {
		return "ERROR: empty query"
	}

	switch strings.ToUpper(tokens[0].Value) {
	case "CREATE":
		return db.executeCreate(tokens)
	case "INSERT":
		return db.executeInsert(tokens)
	case "SELECT":
		return db.executeSelect(tokens)
	default:
		return fmt.Sprintf("ERROR: unsupported statement: %s", tokens[0].Value)
	}
}

func (db *MiniTiDB) executeCreate(tokens []Token) string {
	db.mu.Lock()
	defer db.mu.Unlock()

	tableName := ""
	for i, tok := range tokens {
		if strings.ToUpper(tok.Value) == "TABLE" && i+1 < len(tokens) {
			tableName = tokens[i+1].Value
			break
		}
	}

	if tableName == "" {
		return "ERROR: no table name"
	}

	table := &TableDef{Name: tableName}

	inParens := false
	var colName, colType string
	for _, tok := range tokens {
		if tok.Value == "(" {
			inParens = true
			continue
		}
		if tok.Value == ")" {
			break
		}
		if !inParens {
			continue
		}

		if tok.Type == TokenIdentifier && colName == "" {
			colName = tok.Value
		} else if tok.Type == TokenKeyword && colName != "" && colType == "" {
			colType = tok.Value
		} else if tok.Value == "," || tok.Type == TokenEOF {
			if colName != "" {
				table.Columns = append(table.Columns, ColumnDef{
					Name: colName,
					Type: colType,
				})
			}
			colName = ""
			colType = ""
		}
	}
	if colName != "" {
		table.Columns = append(table.Columns, ColumnDef{
			Name: colName,
			Type: colType,
		})
	}

	db.tables[tableName] = table
	return fmt.Sprintf("OK: created table %s with %d columns", tableName, len(table.Columns))
}

func (db *MiniTiDB) executeInsert(tokens []Token) string {
	parser := NewSQLParser(tokens)
	ast := parser.Parse()
	if ast == nil {
		return "ERROR: failed to parse INSERT"
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	table, ok := db.tables[ast.Table]
	if !ok {
		return fmt.Sprintf("ERROR: table %s does not exist", ast.Table)
	}

	for _, vals := range ast.Values {
		row := make(map[string]string)
		table.NextID++
		row["_id"] = fmt.Sprintf("%d", table.NextID)
		for i, col := range ast.Columns {
			if i < len(vals) {
				row[col] = vals[i]
			}
		}
		table.Rows = append(table.Rows, row)

		key := fmt.Sprintf("%s/%d", ast.Table, table.NextID)
		for k, v := range row {
			db.pd.Put(key+"/"+k, v)
		}
	}

	return fmt.Sprintf("OK: inserted %d row(s) into %s", len(ast.Values), ast.Table)
}

func (db *MiniTiDB) executeSelect(tokens []Token) string {
	parser := NewSQLParser(tokens)
	ast := parser.Parse()
	if ast == nil {
		return "ERROR: failed to parse SELECT"
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	table, ok := db.tables[ast.Table]
	if !ok {
		return fmt.Sprintf("ERROR: table %s does not exist", ast.Table)
	}

	var results []map[string]string
	for _, row := range table.Rows {
		if ast.Where != nil {
			val, exists := row[ast.Where.Column]
			if !exists {
				continue
			}
			if ast.Where.Operator == "=" && val != ast.Where.Value {
				continue
			}
		}
		results = append(results, row)
	}

	if len(results) == 0 {
		return "Empty set"
	}

	var sb strings.Builder
	cols := ast.Columns
	if len(cols) == 1 && cols[0] == "*" {
		cols = nil
		for _, c := range table.Columns {
			cols = append(cols, c.Name)
		}
	}

	header := strings.Join(cols, " | ")
	sb.WriteString(header + "\n")
	sb.WriteString(strings.Repeat("-", len(header)) + "\n")

	for _, row := range results {
		var vals []string
		for _, col := range cols {
			vals = append(vals, row[col])
		}
		sb.WriteString(strings.Join(vals, " | ") + "\n")
	}
	sb.WriteString(fmt.Sprintf("(%d rows)", len(results)))

	return sb.String()
}

func demoMiniTiDB() {
	db := NewMiniTiDB()

	queries := []string{
		"CREATE TABLE users (id INT, name VARCHAR, age INT, role VARCHAR)",
		"INSERT INTO users (name, age, role) VALUES ('Alice', 30, 'engineer')",
		"INSERT INTO users (name, age, role) VALUES ('Bob', 25, 'designer')",
		"INSERT INTO users (name, age, role) VALUES ('Charlie', 35, 'manager')",
		"SELECT * FROM users",
		"SELECT name, role FROM users WHERE name = 'Alice'",
	}

	for _, query := range queries {
		fmt.Printf("  > %s\n", query)
		result := db.Execute(query)
		for _, line := range strings.Split(result, "\n") {
			fmt.Printf("    %s\n", line)
		}
	}

	fmt.Println(`
  TiDB REAL ARCHITECTURE (production):
    - Full MySQL protocol compatibility
    - Distributed transactions (Percolator model)
    - Online DDL (schema changes without downtime)
    - Cost-based optimizer with statistics
    - TiFlash for columnar analytics (HTAP)
    - TiCDC for change data capture
    - Backup & Restore (BR) tool
    - TiDB Dashboard for monitoring`)
}

// Ensure imports are used
var _ = rand.Int
var _ = sort.Ints
var _ = time.Now

/* SUMMARY - CHAPTER 128: TiDB INTERNALS

Key Concepts:

1. SQL TOKENIZER
   - Lexical analysis: input -> tokens
   - Token types: keywords, identifiers, operators, literals
   - Foundation for all SQL processing

2. SQL PARSER (AST)
   - Recursive descent parser
   - Produces Abstract Syntax Tree
   - Handles SELECT, INSERT, JOINs, WHERE, GROUP BY

3. COST-BASED QUERY PLANNER
   - Table statistics (row count, index cardinality)
   - Index selection based on selectivity
   - Plan nodes: TableScan, IndexScan, HashJoin, Sort
   - Cost estimation for choosing between plans

4. RAFT CONSENSUS
   - Leader election with majority voting
   - Log replication across cluster
   - Tolerates minority node failures
   - Consistent state machine replication

5. DISTRIBUTED KEY-VALUE STORE
   - Region-based data partitioning
   - Placement Driver for cluster coordination
   - Timestamp oracle for transaction ordering
   - Multi-replica for fault tolerance

Go Patterns Demonstrated:
- Recursive descent parsing
- State machine (Raft states)
- Mutex-protected concurrent access
- Region-based sharding
- Cost-based optimization

Real TiDB Implementation:
- Uses yacc-generated parser for full MySQL syntax
- Percolator-based distributed transactions
- Coprocessor pushdown to TiKV
- Global timestamp oracle in PD
- Raft groups per region in TiKV (Rust)
- Online schema change (F1 algorithm)
*/
