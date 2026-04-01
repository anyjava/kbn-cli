package parser

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/anyjava/kbn/config"
	"github.com/anyjava/kbn/model"
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

	var taskA *model.Card
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
