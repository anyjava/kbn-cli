// tui/app.go
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/anyjava/kbn/model"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	board       BoardView
	preview     PreviewPanel
	showHelp    bool
	searching   bool
	searchText  string
	fullBoard   model.Board // unfiltered board for search reset
	columnOrder []string
	width       int
	height      int
}

func NewApp(board model.Board, columnOrder []string) App {
	app := App{
		board: BoardView{
			Board: board,
		},
		preview:     PreviewPanel{Visible: true},
		fullBoard:   board,
		columnOrder: columnOrder,
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

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			return a.handleMouseClick(msg)
		}

	case tea.KeyMsg:
		if a.searching {
			return a.handleSearchKey(msg)
		}
		return a.handleNormalKey(msg)
	}
	return a, nil
}

func (a App) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.showHelp {
		a.showHelp = false
		return a, nil
	}
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

func (a App) handleMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if a.showHelp || a.searching {
		return a, nil
	}
	colWidth := a.board.ColWidth()
	if colWidth == 0 {
		return a, nil
	}
	col := msg.X / colWidth
	row := msg.Y - 2 // top border (1) + header (1)
	if a.board.MoveTo(col, row) {
		a.updatePreview()
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
		if utf8.RuneCountInString(msg.String()) == 1 {
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
	a.board.Board = model.NewBoard(filtered, a.columnOrder)
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
