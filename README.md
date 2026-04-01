# md-cli

> Terminal markdown editor with live rendering. Fast, keyboard-first, adaptive themes.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)
[![codecov](https://codecov.io/github/z-d-g/md-cli/graph/badge.svg?token=AyGyuAvKhn)](https://codecov.io/github/z-d-g/md-cli)



## What

`md-cli` renders markdown in your terminal as you type — headings, bold, italic, code blocks, tables, links, images. Edit with full cursor movement, selection, undo/redo. Then save and quit.

```bash
md-cli file.md              # open in editor
md-cli -p file.md           # render to stdout
cat file.md | md-cli -p     # pipe from stdin
```

## Why

| | md-cli | glow | helix | bat |
|---|---|---|---|---|
| Live editing | ✅ | ❌ | ✅ | ❌ |
| Rendered view | ✅ | ✅ | ❌ | ✅ |
| Zero config | ✅ | ✅ | ❌ | ✅ |
| Single binary | ✅ | ✅ | ❌ | ✅ |
| Adaptive light/dark | ✅ | ❌ | ❌ | ❌ |
| Persistent cursor | ✅ | ❌ | ✅ | ❌ |

## Install

```bash
go install github.com/z-d-g/md-cli/cmd/md-cli@latest
```

Or build from source:

```bash
git clone https://github.com/z-d-g/md-cli.git && cd md-cli
make build    # → bin/md-cli
make install  # → ~/.local/bin/md-cli
```

## Usage

## Features

- **Live rendering** — headings, bold, italic, code, links, tables, lists, images
- **Syntax-aware cursor** — jumps between rendered and raw source in code blocks, tables, lists
- **Full editing** — selection, copy/cut/paste, undo/redo, word and line operations
- **Persistent cursor** — restores position per file across sessions
- **Adaptive theming** — respects terminal light/dark background automatically
- **Print mode** — render markdown to stdout without the editor

### Keybindings

| Key | Action |
|-----|--------|
| Ctrl+S | Save |
| Ctrl+Q | Quit |
| F1 | Full help |

Full reference: press `F1` in the editor.

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling

## License

[MIT](LICENSE)
