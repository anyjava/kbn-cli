// tui/preview.go
package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
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

	if p.renderer == nil || p.rendererW != p.Width {
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(p.Width-4),
		)
		if err != nil {
			p.Content = content
			return
		}
		p.renderer = r
		p.rendererW = p.Width
	}

	rendered, err := p.renderer.Render(content)
	if err != nil {
		p.Content = content
		return
	}

	p.Content = rendered
	p.scrollOff = 0
	p.totalLines = len(strings.Split(rendered, "\n"))
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
