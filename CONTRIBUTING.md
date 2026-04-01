# Contributing

Thanks for your interest in improving md-cli.

## Quick start

```bash
git clone https://github.com/z-d-g/md-cli.git
cd md-cli
make test     # run tests
make build    # build to bin/md-cli
```

## What to work on

- Check [issues](../../issues) — bugs and feature requests are tracked there
- Good first issues are labeled `good-first-issue`

## Pull requests

1. Fork, branch, make changes
2. `make test` — all tests must pass
3. `make lint` — no warnings (if available)
4. Open a PR against `main`

Keep PRs small and focused. One change per PR.

## Code style

- Go standard formatting (`gofmt`)
- Self-documenting names, no comments
- Parallel tool calls where independent
- DRY / KISS / UNIX philosophy
