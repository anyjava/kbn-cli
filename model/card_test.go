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
