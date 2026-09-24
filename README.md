# md-cli

> Terminal markdown editor with live rendering. Fast, keyboard-first.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)
[![codecov](https://codecov.io/github/z-d-g/md-cli/graph/badge.svg?token=AyGyuAvKhn)](https://codecov.io/github/z-d-g/md-cli)

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

```bash
md-cli file.md              # open in editor
md-cli -p file.md           # render to stdout
cat file.md | md-cli -p     # pipe from stdin
```

## Features

- **Live rendering** — headings, bold, italic, code, links, tables, lists, images
- **Syntax-aware cursor** — switch between rendered output and raw markdown source in code blocks, tables, lists, and headings
- **Full editing** — selection, copy/cut/paste, undo/redo, word and line operations
- **Persistent cursor** — restores position per file across sessions
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

