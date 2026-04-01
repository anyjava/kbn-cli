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
