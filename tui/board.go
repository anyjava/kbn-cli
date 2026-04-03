// tui/board.go
package tui

import (
	"fmt"
	"strings"

	"github.com/anyjava/kbn/model"
	"github.com/charmbracelet/lipgloss"
)

type BoardView struct {
	Board      model.Board
	ColCursor  int // active column index
	RowCursor  int // active card index within column
	scrollOffs []int // per-column scroll offset
	Width      int
	Height     int
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
		b.adjustScroll()
	}
}

func (b *BoardView) MoveDown() {
	if len(b.Board.Columns) == 0 {
		return
	}
	col := b.Board.Columns[b.ColCursor]
	if b.RowCursor < len(col.Cards)-1 {
		b.RowCursor++
		b.adjustScroll()
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

func (b *BoardView) maxVisibleCards() int {
	cardTotal := 5 // 3 content + 2 border
	usedHeight := 3 // header(1) + margin(1) + potential "+more"(1)
	m := (b.Height - usedHeight) / cardTotal
	if m < 1 {
		m = 1
	}
	return m
}

func (b *BoardView) ensureScrollOffs() {
	needed := len(b.Board.Columns)
	for len(b.scrollOffs) < needed {
		b.scrollOffs = append(b.scrollOffs, 0)
	}
}

func (b *BoardView) adjustScroll() {
	b.ensureScrollOffs()
	if b.ColCursor >= len(b.scrollOffs) {
		return
	}
	maxVis := b.maxVisibleCards()
	off := b.scrollOffs[b.ColCursor]

	// cursor above visible area
	if b.RowCursor < off {
		b.scrollOffs[b.ColCursor] = b.RowCursor
		return
	}

	// cursor below visible area
	// After scrolling, ↑ indicator will appear if off > 0, taking 1 slot.
	// We need to account for this BEFORE deciding the new offset.
	for {
		visibleSlots := maxVis
		if off > 0 {
			visibleSlots--
		}
		if b.RowCursor < off+visibleSlots {
			break // cursor is visible
		}
		off++
	}
	b.scrollOffs[b.ColCursor] = off
}

func (b *BoardView) clampRow() {
	col := b.Board.Columns[b.ColCursor]
	if b.RowCursor >= len(col.Cards) {
		b.RowCursor = max(0, len(col.Cards)-1)
	}
	b.adjustScroll()
}

func (b *BoardView) Render() string {
	if len(b.Board.Columns) == 0 {
		return "No cards found."
	}

	colCount := len(b.Board.Columns)
	baseColWidth := b.Width / colCount
	if baseColWidth < 15 {
		baseColWidth = 15
	}
	remainder := b.Width - baseColWidth*colCount

	var columns []string
	for i, col := range b.Board.Columns {
		// Last column gets the remainder to fill full width
		colWidth := baseColWidth
		if i == colCount-1 {
			colWidth += remainder
		}
		innerWidth := colWidth - 4 // border + padding

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
		cardWidth := innerWidth
		cardInner := cardWidth - 4 // card border + padding
		if cardInner < 8 {
			cardInner = 8
		}
		maxCards := b.maxVisibleCards()
		b.ensureScrollOffs()
		scrollOff := 0
		if i < len(b.scrollOffs) {
			scrollOff = b.scrollOffs[i]
		}

		// Show scroll-up indicator
		if scrollOff > 0 {
			cardLines = append(cardLines, HelpStyle.Render(fmt.Sprintf("  ↑ %d more", scrollOff)))
			maxCards-- // one slot used by indicator
		}

		endIdx := scrollOff + maxCards
		if endIdx > len(col.Cards) {
			endIdx = len(col.Cards)
		}
		for j := scrollOff; j < endIdx; j++ {
			rendered := renderCard(col.Cards[j], cardWidth, cardInner, i == b.ColCursor && j == b.RowCursor)
			cardLines = append(cardLines, rendered)
		}

		// Show scroll-down indicator
		if endIdx < len(col.Cards) {
			remaining := len(col.Cards) - endIdx
			cardLines = append(cardLines, HelpStyle.Render(fmt.Sprintf("  ↓ %d more", remaining)))
		}

		body := strings.Join(append([]string{header}, cardLines...), "\n")

		// Clip body to b.Height lines to prevent column overflow
		bodyLines := strings.Split(body, "\n")
		if len(bodyLines) > b.Height {
			bodyLines = bodyLines[:b.Height]
			body = strings.Join(bodyLines, "\n")
		}

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

func renderCard(card model.Card, width, innerWidth int, selected bool) string {
	// Line 1: ID
	idLine := CardIDStyle.Render(truncate(card.ID, innerWidth))

	// Line 2: Title
	titleLine := CardTitleStyle.Render(truncate(card.Title, innerWidth))

	// Line 3: badges (type + priority)
	var badges []string
	if card.Type != "" {
		badges = append(badges, TypeStyle.Render(card.Type))
	}
	if card.Priority != "" {
		var pStyle lipgloss.Style
		switch card.Priority {
		case "High":
			pStyle = PriorityHighStyle
		case "Medium":
			pStyle = PriorityMediumStyle
		default:
			pStyle = PriorityLowStyle
		}
		badges = append(badges, pStyle.Render(card.Priority))
	}
	badgeLine := strings.Join(badges, " ")

	content := lipgloss.JoinVertical(lipgloss.Left, idLine, titleLine, badgeLine)

	style := CardStyle
	if selected {
		style = SelectedCardStyle
	}
	return style.Width(width).Height(3).Render(content)
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
