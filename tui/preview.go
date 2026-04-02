// tui/preview.go
package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/anyjava/kbn/model"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type PreviewPanel struct {
	Content    string
	Width      int
	Height     int
	Visible    bool
	filePath   string
	scrollOff  int
	totalLines int
	renderer   *glamour.TermRenderer
	rendererW  int // width used to create renderer
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

func (p *PreviewPanel) LoadCard(card *model.Card) {
	if card == nil {
		return
	}
	if card.FilePath == p.filePath {
		return
	}
	p.filePath = card.FilePath

	// Build ticket info header
	header := p.renderCardInfo(card)

	// Read and render markdown body
	data, err := os.ReadFile(card.FilePath)
	if err != nil {
		p.Content = header + "\nError: " + err.Error()
		return
	}

	content := string(data)
	// Strip frontmatter
	if strings.HasPrefix(content, "---") {
		if idx := strings.Index(content[3:], "---"); idx >= 0 {
			content = strings.TrimSpace(content[idx+6:])
		}
	}

	if p.renderer == nil || p.rendererW != p.Width {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(p.Width-4),
		)
		if err != nil {
			p.Content = header + "\n" + content
			return
		}
		p.renderer = r
		p.rendererW = p.Width
	}

	rendered, err := p.renderer.Render(content)
	if err != nil {
		p.Content = header + "\n" + content
		return
	}

	p.Content = header + "\n" + rendered
	p.scrollOff = 0
	p.totalLines = len(strings.Split(p.Content, "\n"))
}

var (
	infoLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(10)
	infoValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)
	infoDividerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)

func (p *PreviewPanel) renderCardInfo(card *model.Card) string {
	var lines []string

	titleLine := CardIDStyle.Render(card.ID) + " " + CardTitleStyle.Render(card.Title)
	lines = append(lines, titleLine)

	var fields []string
	if card.Status != "" {
		fields = append(fields, fmt.Sprintf("%s %s",
			infoLabelStyle.Render("Status:"),
			StatusStyle.Render(card.Status)))
	}
	if card.Type != "" {
		fields = append(fields, fmt.Sprintf("%s %s",
			infoLabelStyle.Render("Type:"),
			TypeStyle.Render(card.Type)))
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
		fields = append(fields, fmt.Sprintf("%s %s",
			infoLabelStyle.Render("Priority:"),
			pStyle.Render(card.Priority)))
	}
	lines = append(lines, fields...)

	divider := infoDividerStyle.Render(strings.Repeat("─", p.Width-6))
	lines = append(lines, divider)

	return strings.Join(lines, "\n")
}

func (p *PreviewPanel) ScrollDown() {
	viewable := p.Height - 2
	maxOff := p.totalLines - viewable
	if maxOff < 0 {
		maxOff = 0
	}
	if p.scrollOff < maxOff {
		p.scrollOff++
	}
}

func (p *PreviewPanel) ScrollUp() {
	if p.scrollOff > 0 {
		p.scrollOff--
	}
}

func (p *PreviewPanel) Render() string {
	if !p.Visible {
		return ""
	}

	content := p.Content
	if content == "" {
		content = "No card selected"
	}

	lines := strings.Split(content, "\n")
	viewable := p.Height - 2

	// Apply scroll offset
	start := p.scrollOff
	if start > len(lines) {
		start = len(lines)
	}
	lines = lines[start:]

	if len(lines) > viewable {
		lines = lines[:viewable]
	}
	content = strings.Join(lines, "\n")

	return PreviewStyle.Width(p.Width).Height(p.Height).Render(content)
}
