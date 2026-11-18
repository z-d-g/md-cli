# Release

md-cli is the **v2** module `github.com/z-d-g/md-cli/v2`. Tags use semantic
versioning with a `v` prefix: `v2.0.0`, `v2.1.0`, `v2.1.1-rc.1`.

## One-time flow per release

```bash
make release VERSION=2.0.0   # lint + test + dist + local tag
make publish                 # push the tag; GitHub builds and publishes
```

`make release` refuses to tag when:

- the working tree is dirty,
- `VERSION` is not `X.Y.Z` (optional `-suffix`),
- the tag already exists.

`make publish` pushes the tag. The `release` workflow then builds every
platform, verifies the archives, writes `checksums.txt`, and creates the GitHub
release. Re-running `make publish` or the workflow updates the existing release
(`--clobber`).

## What the artifacts look like

```
md-cli_2.0.0_linux_amd64.tar.gz
md-cli_2.0.0_linux_arm64.tar.gz
md-cli_2.0.0_darwin_amd64.tar.gz
md-cli_2.0.0_darwin_arm64.tar.gz
md-cli_2.0.0_windows_amd64.zip
md-cli_2.0.0_windows_arm64.zip
checksums.txt
```

Each archive contains the binary, `README.md` and `LICENSE`.

## Version metadata

The binary reports what was baked in:

```
$ md-cli -v
md-cli 2.0.0 (commit: 5023d63, built: 2026-11-18T00:29:31Z)
```

Set by `-X main.version`, `-X main.commit`, `-X main.date` in both the Makefile
and the workflow. Keep the two in sync.

## Install from a release

```bash
go install github.com/z-d-g/md-cli/v2/cmd/md-cli@latest
```

Or download an archive and move `md-cli` onto `PATH`.

## Preview without publishing

```bash
make snapshot     # dist/ built with 0.0.0-snapshot
make dist VERSION=2.0.0
```

Both are local only; nothing is pushed.

## Prereleases

A tag with a suffix (`v2.0.0-rc.1`) is published as a GitHub prerelease.
