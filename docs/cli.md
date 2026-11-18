# CLI

```
md-cli [options] [file.md]
```

| Flag | Meaning |
|------|---------|
| `-p`, `--print` | Render markdown to stdout |
| `-v`, `--version` | Print version |
| `-h`, `--help` | Print usage |

## Behaviour

- `-p` with piped stdin renders stdin; with files, renders each file preceded
  by `--- <file> ---`.
- Interactive mode requires a file. Piped stdin without `-p` is an error.
- Only `*.md` files are accepted as arguments.

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime failure: unreadable input, TTY error |
| 2 | Usage error: unknown flag, missing file argument |

`--print` reports every unreadable file on stderr and exits 1, so pipelines
never treat a partial render as success.
