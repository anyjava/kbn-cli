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
	if len(b.Board.Columns) == 0 {
		return
	}
	if b.RowCursor > 0 {
		b.RowCursor--
	}
}

func (b *BoardView) MoveDown() {
	if len(b.Board.Columns) == 0 {
		return
	}
	col := b.Board.Columns[b.ColCursor]
	if b.RowCursor < len(col.Cards)-1 {
		b.RowCursor++
	}
}

func (b *BoardView) MoveTo(col, row int) bool {
	if len(b.Board.Columns) == 0 {
		return false
	}
	if col < 0 || col >= len(b.Board.Columns) {
		return false
	}
	cards := b.Board.Columns[col].Cards
	if row < 0 || row >= len(cards) {
		return false
	}
	b.ColCursor = col
	b.RowCursor = row
	return true
}

func (b *BoardView) ColWidth() int {
	if len(b.Board.Columns) == 0 {
		return 0
	}
	w := b.Width / len(b.Board.Columns)
	if w < 15 {
		w = 15
	}
	return w
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
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-1]) + "…"
}
