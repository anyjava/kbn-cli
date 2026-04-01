# kbn Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a TUI kanban board viewer that reads Obsidian vault markdown files and displays them as an interactive board in the terminal.

**Architecture:** Config loading -> markdown file scanning/frontmatter parsing -> domain model (Card/Board/Column) -> bubbletea TUI with board view + preview panel. Each layer has a clean interface: config produces settings, parser produces Cards, model groups Cards into a Board, TUI renders the Board.

**Tech Stack:** Go 1.22+, cobra, bubbletea, lipgloss, glamour, yaml.v3, adrg/frontmatter

---

## File Structure

| File | Responsibility |
|------|---------------|
| `main.go` | Entry point, cobra root command setup |
| `config/config.go` | Load and parse `.kbn.yml`, resolve file paths |
| `config/config_test.go` | Config loading tests |
| `model/card.go` | Card, Column, Board types + Board construction |
| `model/card_test.go` | Board grouping/sorting/filtering tests |
| `parser/parser.go` | Scan directory, parse frontmatter, produce Cards |
| `parser/parser_test.go` | Parser tests with fixture md files |
| `parser/testdata/` | Test fixture markdown files |
| `tui/styles.go` | lipgloss style constants |
| `tui/board.go` | Board view component (columns + cards) |
| `tui/preview.go` | Markdown preview panel component |
| `tui/app.go` | Main bubbletea Model combining board + preview |

---

### Task 1: Project Setup

**Files:**
- Create: `go.mod`
- Create: `main.go`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd /Users/anyjava/_dev/kanban-cli
go mod init github.com/anyjava/kbn
```

Expected: `go.mod` created with module name.

- [ ] **Step 2: Add dependencies**

Run:
```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/glamour@latest
go get gopkg.in/yaml.v3
go get github.com/adrg/frontmatter
```

- [ ] **Step 3: Create minimal main.go**

```go
// main.go
package main

import "fmt"

func main() {
	fmt.Println("kbn")
}
```

- [ ] **Step 4: Verify build**

Run: `go build -o kbn .`
Expected: Binary `kbn` created, prints "kbn" when run.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum main.go
git commit -m "chore: initialize kbn project with dependencies"
```

---

### Task 2: Config Loading

**Files:**
- Create: `config/config.go`
- Create: `config/config_test.go`

- [ ] **Step 1: Write failing tests for config loading**

```go
// config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".kbn.yml")
	content := []byte(`vault: "/path/to/vault"
path: "my/project"
glob: "*.md"
fields:
  id: "ticket_id"
  title: "title"
  status: "status"
  priority: "priority"
  type: "type"
hidden_statuses:
  - "Closed"
  - "Archived"
`)
	os.WriteFile(cfgPath, content, 0644)

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault != "/path/to/vault" {
		t.Errorf("vault = %q, want %q", cfg.Vault, "/path/to/vault")
	}
	if cfg.Path != "my/project" {
		t.Errorf("path = %q, want %q", cfg.Path, "my/project")
	}
	if cfg.Glob != "*.md" {
		t.Errorf("glob = %q, want %q", cfg.Glob, "*.md")
	}
	if cfg.Fields.Status != "status" {
		t.Errorf("fields.status = %q, want %q", cfg.Fields.Status, "status")
	}
	if cfg.Fields.ID != "ticket_id" {
		t.Errorf("fields.id = %q, want %q", cfg.Fields.ID, "ticket_id")
	}
	if len(cfg.HiddenStatuses) != 2 || cfg.HiddenStatuses[0] != "Closed" {
		t.Errorf("hidden_statuses = %v, want [Closed, Archived]", cfg.HiddenStatuses)
	}
}

func TestLoadFromFileDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".kbn.yml")
	content := []byte(`vault: "/path/to/vault"
path: "notes"
fields:
  status: "status"
`)
	os.WriteFile(cfgPath, content, 0644)

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Glob != "*.md" {
		t.Errorf("default glob = %q, want %q", cfg.Glob, "*.md")
	}
	if cfg.HiddenStatuses != nil && len(cfg.HiddenStatuses) != 0 {
		t.Errorf("default hidden_statuses = %v, want empty", cfg.HiddenStatuses)
	}
}

func TestFullPath(t *testing.T) {
	cfg := &Config{Vault: "/vault", Path: "sub/dir"}
	got := cfg.FullPath()
	want := "/vault/sub/dir"
	if got != want {
		t.Errorf("FullPath() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./config/ -v`
Expected: FAIL — package not found.

- [ ] **Step 3: Implement config.go**

```go
// config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Fields struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	Status   string `yaml:"status"`
	Priority string `yaml:"priority"`
	Type     string `yaml:"type"`
}

type Config struct {
	Vault          string   `yaml:"vault"`
	Path           string   `yaml:"path"`
	Glob           string   `yaml:"glob"`
	Fields         Fields   `yaml:"fields"`
	HiddenStatuses []string `yaml:"hidden_statuses"`
}

func (c *Config) FullPath() string {
	return filepath.Join(c.Vault, c.Path)
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Glob == "" {
		cfg.Glob = "*.md"
	}

	return cfg, nil
}

func Load(overridePath string) (*Config, error) {
	if overridePath != "" {
		return LoadFromFile(overridePath)
	}

	if _, err := os.Stat(".kbn.yml"); err == nil {
		return LoadFromFile(".kbn.yml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}
	globalPath := filepath.Join(home, ".config", "kbn", "config.yml")
	if _, err := os.Stat(globalPath); err == nil {
		return LoadFromFile(globalPath)
	}

	return nil, fmt.Errorf("no config file found: create .kbn.yml or ~/.config/kbn/config.yml")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./config/ -v`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add config/
git commit -m "feat: add config loading with YAML parsing and defaults"
```

---

### Task 3: Domain Model

**Files:**
- Create: `model/card.go`
- Create: `model/card_test.go`

- [ ] **Step 1: Write failing tests for Board construction**

```go
// model/card_test.go
package model

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	cards := []Card{
		{ID: "1", Title: "Task A", Status: "Done"},
		{ID: "2", Title: "Task B", Status: "In Progress"},
		{ID: "3", Title: "Task C", Status: "Done"},
		{ID: "4", Title: "Task D", Status: "Done"},
		{ID: "5", Title: "Task E", Status: "In Progress"},
	}

	board := NewBoard(cards)

	if len(board.Columns) != 2 {
		t.Fatalf("columns = %d, want 2", len(board.Columns))
	}
	// Sorted by card count descending: Done(3) > In Progress(2)
	if board.Columns[0].Name != "Done" {
		t.Errorf("first column = %q, want %q", board.Columns[0].Name, "Done")
	}
	if len(board.Columns[0].Cards) != 3 {
		t.Errorf("Done cards = %d, want 3", len(board.Columns[0].Cards))
	}
	if board.Columns[1].Name != "In Progress" {
		t.Errorf("second column = %q, want %q", board.Columns[1].Name, "In Progress")
	}
	if len(board.Columns[1].Cards) != 2 {
		t.Errorf("In Progress cards = %d, want 2", len(board.Columns[1].Cards))
	}
}

func TestNewBoardEmptyCards(t *testing.T) {
	board := NewBoard([]Card{})
	if len(board.Columns) != 0 {
		t.Errorf("columns = %d, want 0", len(board.Columns))
	}
}

func TestFilterCards(t *testing.T) {
	cards := []Card{
		{ID: "1", Title: "Open", Status: "Backlog"},
		{ID: "2", Title: "Closed", Status: "Closed"},
		{ID: "3", Title: "Archived", Status: "Archived"},
	}

	filtered := FilterCards(cards, []string{"Closed", "Archived"})
	if len(filtered) != 1 {
		t.Fatalf("filtered = %d, want 1", len(filtered))
	}
	if filtered[0].ID != "1" {
		t.Errorf("remaining card ID = %q, want %q", filtered[0].ID, "1")
	}
}

func TestFilterCardsEmptyHidden(t *testing.T) {
	cards := []Card{
		{ID: "1", Status: "Backlog"},
		{ID: "2", Status: "Closed"},
	}
	filtered := FilterCards(cards, nil)
	if len(filtered) != 2 {
		t.Errorf("filtered = %d, want 2", len(filtered))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./model/ -v`
Expected: FAIL — package not found.

- [ ] **Step 3: Implement model/card.go**

```go
// model/card.go
package model

import "sort"

type Card struct {
	ID       string
	Title    string
	Status   string
	Priority string
	Type     string
	FilePath string
	Meta     map[string]string
}

type Column struct {
	Name  string
	Cards []Card
}

type Board struct {
	Columns []Column
}

func FilterCards(cards []Card, hiddenStatuses []string) []Card {
	if len(hiddenStatuses) == 0 {
		return cards
	}
	hidden := make(map[string]bool, len(hiddenStatuses))
	for _, s := range hiddenStatuses {
		hidden[s] = true
	}
	var result []Card
	for _, c := range cards {
		if !hidden[c.Status] {
			result = append(result, c)
		}
	}
	return result
}

func NewBoard(cards []Card) Board {
	groups := make(map[string][]Card)
	for _, c := range cards {
		groups[c.Status] = append(groups[c.Status], c)
	}

	columns := make([]Column, 0, len(groups))
	for name, cards := range groups {
		columns = append(columns, Column{Name: name, Cards: cards})
	}

	sort.Slice(columns, func(i, j int) bool {
		if len(columns[i].Cards) != len(columns[j].Cards) {
			return len(columns[i].Cards) > len(columns[j].Cards)
		}
		return columns[i].Name < columns[j].Name
	})

	return Board{Columns: columns}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./model/ -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add model/
git commit -m "feat: add Card, Column, Board model with filtering and grouping"
```

---

### Task 4: Markdown Parser

**Files:**
- Create: `parser/parser.go`
- Create: `parser/parser_test.go`
- Create: `parser/testdata/task-a.md`
- Create: `parser/testdata/task-b.md`
- Create: `parser/testdata/no-frontmatter.md`
- Create: `parser/testdata/subdir/ignored.md`

- [ ] **Step 1: Create test fixture files**

```markdown
<!-- parser/testdata/task-a.md -->
---
ticket_id: UL-001
title: Task A
status: Backlog
priority: High
type: Feature
custom_field: hello
---

## Overview

This is task A content.
```

```markdown
<!-- parser/testdata/task-b.md -->
---
ticket_id: UL-002
title: Task B
status: In Progress
priority: Medium
type: Bug
---

## Overview

This is task B content.
```

```markdown
<!-- parser/testdata/no-frontmatter.md -->

# Just a plain file

No YAML frontmatter here.
```

```markdown
<!-- parser/testdata/subdir/ignored.md -->
---
ticket_id: UL-003
title: Ignored
status: Done
---

Should not be picked up by `*.md` glob at root level.
```

- [ ] **Step 2: Write failing tests**

```go
// parser/parser_test.go
package parser

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/anyjava/kbn/config"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestParseCards(t *testing.T) {
	fields := config.Fields{
		ID:       "ticket_id",
		Title:    "title",
		Status:   "status",
		Priority: "priority",
		Type:     "type",
	}

	cards, errs := ParseCards(testdataDir(), "*.md", fields)
	if len(errs) > 0 {
		t.Logf("parse warnings: %v", errs)
	}

	// Should find task-a.md and task-b.md (not subdir/, not no-frontmatter.md due to missing status)
	found := make(map[string]bool)
	for _, c := range cards {
		found[c.ID] = true
	}

	if !found["UL-001"] {
		t.Error("missing UL-001")
	}
	if !found["UL-002"] {
		t.Error("missing UL-002")
	}
	if found["UL-003"] {
		t.Error("UL-003 in subdir should not be found")
	}
}

func TestParseCardFields(t *testing.T) {
	fields := config.Fields{
		ID:       "ticket_id",
		Title:    "title",
		Status:   "status",
		Priority: "priority",
		Type:     "type",
	}

	cards, _ := ParseCards(testdataDir(), "*.md", fields)

	var taskA *card
	for i := range cards {
		if cards[i].ID == "UL-001" {
			taskA = &cards[i]
			break
		}
	}
	if taskA == nil {
		t.Fatal("UL-001 not found")
	}
	if taskA.Title != "Task A" {
		t.Errorf("title = %q, want %q", taskA.Title, "Task A")
	}
	if taskA.Status != "Backlog" {
		t.Errorf("status = %q, want %q", taskA.Status, "Backlog")
	}
	if taskA.Priority != "High" {
		t.Errorf("priority = %q, want %q", taskA.Priority, "High")
	}
	if taskA.Type != "Feature" {
		t.Errorf("type = %q, want %q", taskA.Type, "Feature")
	}
}

type card = model.Card
```

Note: The test file needs the model import. Update the import to:

```go
// parser/parser_test.go — imports section
import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/anyjava/kbn/config"
	"github.com/anyjava/kbn/model"
)
```

And remove the type alias line at the bottom. Reference `model.Card` directly in the test where needed:

```go
	var taskA *model.Card
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./parser/ -v`
Expected: FAIL — package not found.

- [ ] **Step 4: Implement parser.go**

```go
// parser/parser.go
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/anyjava/kbn/config"
	"github.com/anyjava/kbn/model"
)

func ParseCards(dir string, glob string, fields config.Fields) ([]model.Card, []error) {
	pattern := filepath.Join(dir, glob)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, []error{fmt.Errorf("glob %q: %w", pattern, err)}
	}

	var cards []model.Card
	var errs []error

	for _, path := range matches {
		card, err := parseFile(path, fields)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(path), err))
			continue
		}
		if card.Status == "" {
			continue // skip files without status field
		}
		cards = append(cards, card)
	}

	return cards, errs
}

func parseFile(path string, fields config.Fields) (model.Card, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.Card{}, err
	}
	defer f.Close()

	var meta map[string]interface{}
	_, err = frontmatter.Parse(f, &meta)
	if err != nil {
		return model.Card{}, err
	}

	card := model.Card{
		FilePath: path,
		Meta:     make(map[string]string),
	}

	mapped := map[string]*string{
		fields.ID:       &card.ID,
		fields.Title:    &card.Title,
		fields.Status:   &card.Status,
		fields.Priority: &card.Priority,
		fields.Type:     &card.Type,
	}

	for key, val := range meta {
		str := fmt.Sprintf("%v", val)
		if target, ok := mapped[key]; ok && key != "" {
			*target = str
		} else {
			card.Meta[key] = str
		}
	}

	// Fallback: use filename as ID/Title if not mapped
	if card.ID == "" {
		base := strings.TrimSuffix(filepath.Base(path), ".md")
		card.ID = base
	}
	if card.Title == "" {
		base := strings.TrimSuffix(filepath.Base(path), ".md")
		// Remove ID prefix if present (e.g., "UL-001 Task Name" -> "Task Name")
		if card.ID != "" && strings.HasPrefix(base, card.ID) {
			card.Title = strings.TrimSpace(strings.TrimPrefix(base, card.ID))
		}
		if card.Title == "" {
			card.Title = base
		}
	}

	return card, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./parser/ -v`
Expected: PASS (2 tests).

- [ ] **Step 6: Commit**

```bash
git add parser/
git commit -m "feat: add markdown parser with frontmatter field mapping"
```

---

### Task 5: TUI Styles

**Files:**
- Create: `tui/styles.go`

- [ ] **Step 1: Create styles.go**

```go
// tui/styles.go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Column header
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("62"))

	// Active column header
	ActiveHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("33"))

	// Card (normal)
	CardStyle = lipgloss.NewStyle().
		Padding(0, 1)

	// Card (selected)
	SelectedCardStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("33"))

	// Column border
	ColumnStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 0)

	// Active column border
	ActiveColumnStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(0, 0)

	// Preview panel
	PreviewStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	// Help bar
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	// Search input
	SearchStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))
)
```

- [ ] **Step 2: Verify build**

Run: `go build ./tui/`
Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add tui/styles.go
git commit -m "feat: add lipgloss styles for TUI components"
```

---

### Task 6: TUI Board View

**Files:**
- Create: `tui/board.go`

- [ ] **Step 1: Implement board.go**

This is the core board rendering component. It takes a Board model and renders columns side by side, handling cursor position and card truncation.

```go
// tui/board.go
package tui

import (
	"fmt"
	"strings"

	"github.com/anyjava/kbn/model"
	"github.com/charmbracelet/lipgloss"
)

type BoardView struct {
	Board     model.Board
	ColCursor int // active column index
	RowCursor int // active card index within column
	Width     int
	Height    int
}

func NewBoardView(board model.Board, width, height int) BoardView {
	return BoardView{
		Board:  board,
		Width:  width,
		Height: height,
	}
}

func (b *BoardView) ActiveCard() *model.Card {
	if len(b.Board.Columns) == 0 {
		return nil
	}
	col := b.Board.Columns[b.ColCursor]
	if len(col.Cards) == 0 {
		return nil
	}
	return &col.Cards[b.RowCursor]
}

func (b *BoardView) MoveLeft() {
	if b.ColCursor > 0 {
		b.ColCursor--
		b.clampRow()
	}
}

func (b *BoardView) MoveRight() {
	if b.ColCursor < len(b.Board.Columns)-1 {
		b.ColCursor++
		b.clampRow()
	}
}

func (b *BoardView) MoveUp() {
	if b.RowCursor > 0 {
		b.RowCursor--
	}
}

func (b *BoardView) MoveDown() {
	col := b.Board.Columns[b.ColCursor]
	if b.RowCursor < len(col.Cards)-1 {
		b.RowCursor++
	}
}

func (b *BoardView) clampRow() {
	col := b.Board.Columns[b.ColCursor]
	if b.RowCursor >= len(col.Cards) {
		b.RowCursor = max(0, len(col.Cards)-1)
	}
}

func (b *BoardView) Render() string {
	if len(b.Board.Columns) == 0 {
		return "No cards found."
	}

	colCount := len(b.Board.Columns)
	colWidth := b.Width / colCount
	if colWidth < 15 {
		colWidth = 15
	}
	innerWidth := colWidth - 4 // border + padding

	var columns []string
	for i, col := range b.Board.Columns {
		// Header
		headerText := fmt.Sprintf("%s (%d)", col.Name, len(col.Cards))
		if len(headerText) > innerWidth {
			headerText = headerText[:innerWidth]
		}
		var header string
		if i == b.ColCursor {
			header = ActiveHeaderStyle.Width(innerWidth).Render(headerText)
		} else {
			header = HeaderStyle.Width(innerWidth).Render(headerText)
		}

		// Cards
		var cardLines []string
		maxCards := b.Height - 4 // header + border
		for j, card := range col.Cards {
			if j >= maxCards {
				remaining := len(col.Cards) - maxCards
				cardLines = append(cardLines, HelpStyle.Render(fmt.Sprintf("  +%d more", remaining)))
				break
			}
			label := truncate(fmt.Sprintf("%s %s", card.ID, card.Title), innerWidth-2)
			if i == b.ColCursor && j == b.RowCursor {
				label = SelectedCardStyle.Render("> " + label)
			} else {
				label = CardStyle.Render("  " + label)
			}
			cardLines = append(cardLines, label)
		}

		body := strings.Join(append([]string{header}, cardLines...), "\n")

		var colRendered string
		if i == b.ColCursor {
			colRendered = ActiveColumnStyle.Width(colWidth).Height(b.Height).Render(body)
		} else {
			colRendered = ColumnStyle.Width(colWidth).Height(b.Height).Render(body)
		}
		columns = append(columns, colRendered)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-1] + "…"
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./tui/`
Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add tui/board.go
git commit -m "feat: add board view with column rendering and cursor navigation"
```

---

### Task 7: Preview Panel

**Files:**
- Create: `tui/preview.go`

- [ ] **Step 1: Implement preview.go**

```go
// tui/preview.go
package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

type PreviewPanel struct {
	Content  string
	Width    int
	Height   int
	Visible  bool
	filePath string
}

func NewPreviewPanel(width, height int) PreviewPanel {
	return PreviewPanel{
		Width:   width,
		Height:  height,
		Visible: true,
	}
}

func (p *PreviewPanel) Toggle() {
	p.Visible = !p.Visible
}

func (p *PreviewPanel) LoadFile(path string) {
	if path == p.filePath {
		return
	}
	p.filePath = path

	data, err := os.ReadFile(path)
	if err != nil {
		p.Content = "Error: " + err.Error()
		return
	}

	content := string(data)
	// Strip frontmatter
	if strings.HasPrefix(content, "---") {
		if idx := strings.Index(content[3:], "---"); idx >= 0 {
			content = strings.TrimSpace(content[idx+6:])
		}
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(p.Width-4),
	)
	if err != nil {
		p.Content = content
		return
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		p.Content = content
		return
	}

	p.Content = rendered
}

func (p *PreviewPanel) Render() string {
	if !p.Visible {
		return ""
	}

	content := p.Content
	if content == "" {
		content = "No card selected"
	}

	// Limit content height
	lines := strings.Split(content, "\n")
	if len(lines) > p.Height-2 {
		lines = lines[:p.Height-2]
	}
	content = strings.Join(lines, "\n")

	return PreviewStyle.Width(p.Width).Height(p.Height).Render(content)
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./tui/`
Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add tui/preview.go
git commit -m "feat: add preview panel with glamour markdown rendering"
```

---

### Task 8: TUI App (Main Model)

**Files:**
- Create: `tui/app.go`

- [ ] **Step 1: Implement app.go**

This is the main bubbletea Model that ties board + preview together and handles all key events.

```go
// tui/app.go
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anyjava/kbn/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	board      BoardView
	preview    PreviewPanel
	showHelp   bool
	searching  bool
	searchText string
	fullBoard  model.Board // unfiltered board for search reset
	width      int
	height     int
}

func NewApp(board model.Board) App {
	app := App{
		fullBoard: board,
	}
	return app
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.relayout()
		a.updatePreview()
		return a, nil

	case tea.KeyMsg:
		if a.searching {
			return a.handleSearchKey(msg)
		}
		return a.handleNormalKey(msg)
	}
	return a, nil
}

func (a App) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit

	case "left", "h":
		a.board.MoveLeft()
		a.updatePreview()

	case "right", "l":
		a.board.MoveRight()
		a.updatePreview()

	case "up", "k":
		a.board.MoveUp()
		a.updatePreview()

	case "down", "j":
		a.board.MoveDown()
		a.updatePreview()

	case "enter":
		if card := a.board.ActiveCard(); card != nil {
			return a, a.openEditor(card.FilePath)
		}

	case "p":
		a.preview.Toggle()
		a.relayout()

	case "/":
		a.searching = true
		a.searchText = ""

	case "?":
		a.showHelp = !a.showHelp
	}
	return a, nil
}

func (a App) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		a.searching = false
		a.applySearch()
		a.updatePreview()

	case "esc":
		a.searching = false
		a.searchText = ""
		a.board.Board = a.fullBoard
		a.board.ColCursor = 0
		a.board.RowCursor = 0
		a.updatePreview()

	case "backspace":
		if len(a.searchText) > 0 {
			a.searchText = a.searchText[:len(a.searchText)-1]
		}

	default:
		if len(msg.String()) == 1 {
			a.searchText += msg.String()
		}
	}
	return a, nil
}

func (a *App) applySearch() {
	if a.searchText == "" {
		a.board.Board = a.fullBoard
		return
	}
	query := strings.ToLower(a.searchText)
	var filtered []model.Card
	for _, col := range a.fullBoard.Columns {
		for _, card := range col.Cards {
			if strings.Contains(strings.ToLower(card.ID), query) ||
				strings.Contains(strings.ToLower(card.Title), query) {
				filtered = append(filtered, card)
			}
		}
	}
	a.board.Board = model.NewBoard(filtered)
	a.board.ColCursor = 0
	a.board.RowCursor = 0
}

func (a *App) relayout() {
	previewWidth := 0
	if a.preview.Visible {
		previewWidth = a.width * 35 / 100
	}
	boardWidth := a.width - previewWidth
	boardHeight := a.height - 2 // help bar

	a.board.Width = boardWidth
	a.board.Height = boardHeight
	a.preview.Width = previewWidth
	a.preview.Height = boardHeight
}

func (a *App) updatePreview() {
	if card := a.board.ActiveCard(); card != nil {
		a.preview.LoadFile(card.FilePath)
	}
}

func (a *App) openEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return nil
	})
}

func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	if a.showHelp {
		return a.renderHelp()
	}

	boardStr := a.board.Render()
	var content string
	if a.preview.Visible {
		previewStr := a.preview.Render()
		content = lipgloss.JoinHorizontal(lipgloss.Top, boardStr, previewStr)
	} else {
		content = boardStr
	}

	helpBar := a.renderHelpBar()
	return lipgloss.JoinVertical(lipgloss.Left, content, helpBar)
}

func (a App) renderHelpBar() string {
	if a.searching {
		return SearchStyle.Render(fmt.Sprintf("/ %s█", a.searchText))
	}
	return HelpStyle.Render("  ←→/hl columns  ↑↓/jk cards  Enter editor  p preview  / search  ? help  q quit")
}

func (a App) renderHelp() string {
	help := `
  Key Bindings

  ←/→  h/l     Move between columns
  ↑/↓  j/k     Move between cards
  Enter         Open in $EDITOR
  p             Toggle preview panel
  /             Search cards
  ?             Toggle this help
  q  Ctrl+C     Quit

  Press any key to close this help.
`
	return help
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./tui/`
Expected: Build succeeds.

- [ ] **Step 3: Commit**

```bash
git add tui/app.go
git commit -m "feat: add main TUI app with keyboard navigation, search, and editor integration"
```

---

### Task 9: CLI with Cobra + Main Entry Point

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Implement main.go with cobra**

```go
// main.go
package main

import (
	"fmt"
	"os"

	"github.com/anyjava/kbn/config"
	"github.com/anyjava/kbn/model"
	"github.com/anyjava/kbn/parser"
	"github.com/anyjava/kbn/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func main() {
	var (
		showAll    bool
		configPath string
		pathOverride string
	)

	rootCmd := &cobra.Command{
		Use:   "kbn",
		Short: "Obsidian vault kanban board TUI viewer",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			if pathOverride != "" {
				cfg.Path = pathOverride
			}

			cards, errs := parser.ParseCards(cfg.FullPath(), cfg.Glob, cfg.Fields)
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "warning: %v\n", e)
			}

			if !showAll {
				cards = model.FilterCards(cards, cfg.HiddenStatuses)
			}

			board := model.NewBoard(cards)
			app := tui.NewApp(board)

			p := tea.NewProgram(app, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.Flags().BoolVar(&showAll, "all", false, "Show all cards including hidden statuses")
	rootCmd.Flags().StringVar(&configPath, "config", "", "Path to config file")
	rootCmd.Flags().StringVar(&pathOverride, "path", "", "Override vault path")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify build**

Run: `go build -o kbn .`
Expected: Binary `kbn` created.

- [ ] **Step 3: Manual smoke test**

Create a test config at `/Users/anyjava/_dev/kanban-cli/.kbn.yml`:

```yaml
vault: "/Users/anyjava/Library/Mobile Documents/iCloud~md~obsidian/Documents/sht21c@gmail.com"
path: "개발일/Underlog"
glob: "*.md"
fields:
  id: "ticket_id"
  title: "title"
  status: "status"
  priority: "priority"
  type: "type"
hidden_statuses:
  - "Closed"
```

Run: `./kbn`
Expected: TUI opens showing kanban board with Underlog tickets grouped by status.

- [ ] **Step 4: Commit**

```bash
git add main.go .kbn.yml
git commit -m "feat: add cobra CLI and wire up config -> parser -> TUI pipeline"
```

---

### Task 10: End-to-End Verification

- [ ] **Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass.

- [ ] **Step 2: Run with real vault data**

Run: `./kbn`
Verify:
- Columns appear grouped by status
- Arrow keys / hjkl navigate between columns and cards
- Preview panel shows markdown content
- `p` toggles preview panel
- `/` opens search, typing filters cards, Enter confirms, Esc resets
- `Enter` opens file in `$EDITOR`
- `q` quits

- [ ] **Step 3: Test --all flag**

Run: `./kbn --all`
Expected: Closed tickets also appear in the board.

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "chore: finalize kbn v0.1.0"
```
