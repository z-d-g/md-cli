# md-cli

> Terminal markdown editor with live rendering. Written in Go.

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)
[![codecov](https://codecov.io/github/z-d-g/md-cli/graph/badge.svg?token=AyGyuAvKhn)](https://codecov.io/github/z-d-g/md-cli)

## Features

- **Live rendering** — headings, bold, italic, code, links, tables, lists, images
- **Syntax-aware cursor** — jumps between rendered and raw source in code blocks, tables, and lists
- **Full editing** — selection, copy/cut/paste, undo/redo, word and line operations
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
make build    # → bin/md-cli
make install  # → ~/.local/bin/md-cli
```

## Usage

```bash
md-cli file.md              # open in editor
md-cli -p file.md           # print to stdout
cat file.md | md-cli -p     # pipe from stdin
```

### Keybindings

| Key     | Action |
|---------|--------|
| Ctrl+S  | Save   |
| Ctrl+Q  | Quit   |
| F1      | Help   |


Full reference: press `F1` in the editor.

## Makefile

```bash
make              # build (same as make build)
make release      # build with version stamped into the binary
make run FILE=x.md  # run from source
make test         # run tests
make install      # copy to ~/.local/bin
make uninstall    # remove it
make clean        # delete build artifacts
```

Override defaults: `make install PREFIX=/usr/local` or `make run FILE=notes.md`.

## License

[MIT](LICENSE)
