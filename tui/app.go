// tui/app.go
package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/anyjava/kbn/config"
	"github.com/anyjava/kbn/model"
	"github.com/anyjava/kbn/parser"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

type reloadMsg struct{}

type App struct {
	board         BoardView
	preview       PreviewPanel
	showHelp      bool
	searching     bool
	searchText    string
	fullBoard     model.Board // unfiltered board for search reset
	columnOrder   []string
	previewLayout string // "right" or "bottom"
	cfg           *config.Config
	showAll       bool
	width         int
	height        int
}

func NewApp(board model.Board, cfg *config.Config, showAll bool) App {
	app := App{
		board: BoardView{
			Board: board,
		},
		preview:       PreviewPanel{Visible: true},
		fullBoard:     board,
		columnOrder:   cfg.ColumnOrder,
		previewLayout: cfg.PreviewLayout,
		cfg:           cfg,
		showAll:       showAll,
	}
	return app
}

func (a App) Init() tea.Cmd {
	return a.watchFiles()
}

func (a App) watchFiles() tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return nil
		}
		dir := a.cfg.FullPath()
		watcher.Add(dir)

		for {
			select {
			case event := <-watcher.Events:
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
					if strings.HasSuffix(event.Name, ".md") {
						return reloadMsg{}
					}
				}
			case <-watcher.Errors:
				return nil
			}
		}
	}
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.relayout()
		a.updatePreview()
		return a, nil

	case reloadMsg:
		a.reload()
		return a, a.watchFiles()

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			return a.handleMouseClick(msg)
		}
		if msg.Button == tea.MouseButtonWheelUp {
			a.preview.ScrollUp()
			return a, nil
		}
		if msg.Button == tea.MouseButtonWheelDown {
			a.preview.ScrollDown()
			return a, nil
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
	key := korToEng(msg.String())
	switch key {
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

	case "J":
		a.preview.ScrollDown()

	case "K":
		a.preview.ScrollUp()

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

	case "r":
		a.reload()

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
	// Each card is 5 lines (3 content + 2 border), header area is 2 lines (border + header)
	cardHeight := 5
	headerArea := 2
	y := msg.Y - headerArea
	if y < 0 {
		return a, nil
	}
	row := y / cardHeight
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
	helpBarHeight := 1
	borderHeight := 2 // top + bottom border lines per bordered component

	if a.previewLayout == "bottom" {
		a.board.Width = a.width
		a.preview.Width = a.width
		if a.preview.Visible {
			available := a.height - helpBarHeight - borderHeight*2 // board border + preview border
			previewContent := available * 35 / 100
			boardContent := available - previewContent
			a.board.Height = boardContent
			a.preview.Height = previewContent
		} else {
			a.board.Height = a.height - helpBarHeight - borderHeight
			a.preview.Height = 0
		}
	} else {
		// "right" layout (default)
		previewWidth := 0
		if a.preview.Visible {
			previewWidth = a.width * 35 / 100
		}
		boardWidth := a.width - previewWidth
		boardHeight := a.height - helpBarHeight - borderHeight

		a.board.Width = boardWidth
		a.board.Height = boardHeight
		a.preview.Width = previewWidth
		a.preview.Height = boardHeight
	}
}

func (a *App) reload() {
	cards, _ := parser.ParseCards(a.cfg.FullPath(), a.cfg.Glob, a.cfg.Fields)
	if !a.showAll {
		cards = model.FilterCards(cards, a.cfg.HiddenStatuses)
	}
	board := model.NewBoard(cards, a.cfg.ColumnOrder)
	a.fullBoard = board
	a.board.Board = board
	a.board.clampRow()
	a.preview.filePath = "" // force preview reload
	a.updatePreview()
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
		if a.previewLayout == "bottom" {
			content = lipgloss.JoinVertical(lipgloss.Left, boardStr, previewStr)
		} else {
			content = lipgloss.JoinHorizontal(lipgloss.Top, boardStr, previewStr)
		}
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
	return HelpStyle.Render("  ←→/hl columns  ↑↓/jk cards  J/K scroll  Enter editor  p preview  r reload  / search  ? help  q quit")
}

func (a App) renderHelp() string {
	help := `
  Key Bindings (한글 키보드에서도 동작)

  ←/→  h/l     Move between columns
  ↑/↓  j/k     Move between cards
  J/K           Scroll preview panel
  Mouse wheel   Scroll preview panel
  Enter         Open in $EDITOR
  p             Toggle preview panel
  r             Reload files
  /             Search cards
  ?             Toggle this help
  q  Ctrl+C     Quit

  Press any key to close this help.
`
	return help
}

// korToEng maps Korean 2벌식 keyboard input to English equivalents.
// Also handles composed vowels from IME buffering (e.g. ㅗ+ㅣ=ㅚ).
var korEngMap = map[string]string{
	// Single jamo
	"ㅂ": "q", "ㅈ": "w", "ㄷ": "e", "ㄱ": "r", "ㅅ": "t",
	"ㅛ": "y", "ㅕ": "u", "ㅑ": "i", "ㅐ": "o", "ㅔ": "p",
	"ㅁ": "a", "ㄴ": "s", "ㅇ": "d", "ㄹ": "f", "ㅎ": "g",
	"ㅗ": "h", "ㅓ": "j", "ㅏ": "k", "ㅣ": "l",
	"ㅋ": "z", "ㅌ": "x", "ㅊ": "c", "ㅍ": "v", "ㅠ": "b",
	"ㅜ": "n", "ㅡ": "m",
	// Shift variants
	"ㅃ": "Q", "ㅉ": "W", "ㄸ": "E", "ㄲ": "R", "ㅆ": "T",
	"ㅒ": "O", "ㅖ": "P",
	// Composed vowels (IME combines these before sending)
	// Map to the LAST key pressed (most recent navigation intent)
	"ㅚ": "l", // ㅗ+ㅣ = h+l → treat as "l" (right)
	"ㅘ": "k", // ㅗ+ㅏ = h+k → treat as "k" (up)
	"ㅙ": "o", // ㅗ+ㅐ = h+o
	"ㅝ": "j", // ㅜ+ㅓ = n+j → treat as "j" (down)
	"ㅞ": "p", // ㅜ+ㅔ = n+p → treat as "p" (preview)
	"ㅟ": "l", // ㅜ+ㅣ = n+l → treat as "l" (right)
	"ㅢ": "l", // ㅡ+ㅣ = m+l → treat as "l" (right)
}

func korToEng(key string) string {
	if mapped, ok := korEngMap[key]; ok {
		return mapped
	}
	return key
}
