# md-cli

Terminal markdown editor with live rendering. Written in Go.

## Features

- **Live rendering** — headings, bold, italic, code, links, tables, lists, images
- **Syntax-aware cursor** — jumps between rendered vs raw source in code blocks, tables, and lists
- **Full editing** — selection, copy/cut/paste, undo/redo, word/line operations
- **Persistent cursor** — restores position per file across sessions
- **Adaptive theming** — respects terminal light/dark background
- **Print mode** — render markdown to stdout without the editor

## Install

```bash
go install github.com/z-d-g/md-cli/cmd/md-cli@latest
```

Or build from source:

```bash
git clone https://github.com/z-d-g/md-cli.git
cd md-cli
make build-release
```

## Usage

```bash
md-cli file.md          # open in editor
md-cli -p file.md       # print to stdout
```

Keybindings: `F1` for help. `Ctrl+Q` quit. `Ctrl+S` save.

## Release

Push a tag → GitHub Actions builds binaries for Linux, macOS, and Windows (amd64 + arm64) → uploads to [Releases](../../releases).

```bash
git tag v0.1.0
git push origin v0.1.0
# then create a release from the tag on GitHub
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — styling
- [clipboard](https://golang.design/x/clipboard) — system clipboard

## License

MIT
