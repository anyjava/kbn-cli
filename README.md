# kbn-cli

A terminal kanban board viewer for Obsidian vaults.

`kbn` reads markdown files from your Obsidian vault, groups them by their frontmatter `status` field, and displays an interactive kanban board in your terminal.

```
+------------------+------------------+------------------+
| TODO (2)         | In Progress (3)  | Done (5)         |
|------------------|------------------|------------------|
| +------+         | +------+         | +------+         |
| | UL-01|         | | UL-04|         | | UL-02|         |
| | Title|         | | Title|         | | Title|         |
| | feat |         | | bug  |         | | feat |         |
| +------+         | +------+         | +------+         |
|                  |                  |                  |
+------------------+------------------+------------------+
| Preview: markdown content rendered here               |
+-------------------------------------------------------+
```

## Install

```bash
brew tap anyjava/tap
brew install kbn-cli
```

Or build from source:

```bash
go install github.com/anyjava/kbn-cli@latest
```

## Quick Start

```bash
# Create config interactively
kbn init

# Open the kanban board
kbn
```

## Configuration

`kbn` looks for `.kbn.yml` in the current directory, then `~/.config/kbn/config.yml`.

```yaml
# .kbn.yml
vault: "/path/to/obsidian/vault"
path: "projects/MyApp"              # folder within vault
glob: "*.md"                        # file pattern to scan

fields:
  id: "ticket_id"                   # frontmatter field for card ID
  title: "title"                    # frontmatter field for card title
  status: "status"                  # frontmatter field for column grouping (required)
  priority: "priority"              # shown as badge on card
  type: "type"                      # shown as badge on card

hidden_statuses:                    # hidden by default, use --all to show
  - "Closed"

column_order:                       # custom column order (omit for card-count sorting)
  - "TODO"
  - "In Progress"
  - "Done"

preview_layout: "bottom"            # "right" or "bottom"
```

### Minimal Config

Only `vault`, `path`, and `fields.status` are required:

```yaml
vault: "/path/to/vault"
path: "notes"
fields:
  status: "status"
```

### Expected Markdown Format

Each `.md` file represents a card. Use YAML frontmatter for metadata:

```markdown
---
ticket_id: UL-001
title: Build login page
status: In Progress
priority: High
type: Feature
---

## Description

Implementation details here...
```

## Key Bindings

| Key | Action |
|-----|--------|
| `h` `l` / `Left` `Right` | Move between columns |
| `j` `k` / `Up` `Down` | Move between cards |
| `J` `K` | Scroll preview panel |
| `Enter` | Open card in `$EDITOR` |
| `p` | Toggle preview panel |
| `r` | Reload files |
| `/` | Search cards by ID or title |
| `?` | Show help |
| `q` / `Ctrl+C` | Quit |
| Mouse click | Select card |
| Mouse wheel | Scroll preview |

All key bindings work with Korean 2-beolsik keyboard layout.

## CLI Flags

```bash
kbn                      # open board
kbn --all                # include hidden statuses
kbn --config path.yml    # use specific config file
kbn --path "other/dir"   # override vault path
kbn --version            # show version
kbn init                 # create .kbn.yml interactively
```

## Features

- **Interactive TUI** - Navigate with keyboard or mouse
- **Card view** - Each ticket displayed as a card with ID, title, type, and priority badges
- **Markdown preview** - Rendered with syntax highlighting via [glamour](https://github.com/charmbracelet/glamour)
- **Flexible layout** - Preview panel on the right or bottom
- **Live reload** - Automatically updates when markdown files change
- **Configurable** - Map any frontmatter fields, set column order, hide statuses
- **Korean keyboard support** - Works with 2-beolsik layout without switching to English

## Tech Stack

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [fsnotify](https://github.com/fsnotify/fsnotify) - File watching

## License

MIT
