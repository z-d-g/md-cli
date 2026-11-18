# Contributing

Thanks for your interest in improving md-cli.

## Quick start

```bash
git clone https://github.com/z-d-g/md-cli.git
cd md-cli
make test     # run tests
make lint     # go vet + gofmt + golangci-lint
make build    # build to bin/md-cli
make dist     # cross-platform archives in dist/
```

Requires Go 1.27 or newer. This is the **v2** module
(`github.com/z-d-g/md-cli/v2`).

## What to work on

- Check [issues](https://github.com/z-d-g/md-cli/issues) — bugs and feature requests are tracked there
- Good first issues are labeled `good-first-issue`

## Pull requests

1. Fork, branch, make changes
2. `make test` — all tests must pass
3. `make lint` — no warnings (if available)
4. Open a PR against `main`

Keep PRs small and focused. One change per PR.

## Code style

- Go standard formatting (`gofmt`)
- Self-documenting names; comments explain *why*, never *what*
- Errors are returned by libraries; the UI logs via `slog`
- DRY / KISS / UNIX philosophy

Read [docs/INDEX.md](docs/INDEX.md) before changing parsing, rendering or the
editor core.
