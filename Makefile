# md-cli — build, test, release
#
# Release flow:
#   make release VERSION=2.0.0   # verify, tag v2.0.0 locally
#   make publish                 # push the tag → GitHub creates the release
#
# `make dist` builds the same archives CI does, so you can inspect them first.

BINARY      := md-cli
MODULE_PATH := github.com/z-d-g/md-cli/v2
CMD         := ./cmd/md-cli
CGO         := CGO_ENABLED=0

# Build flags
LDFLAGS := -s -w
GOFLAGS := -trimpath

# ── Version ────────────────────────────────────────────────────────
# VERSION overrides the tag name. Without it we use the nearest tag.
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
SEMVER  := $(patsubst v%,%,$(VERSION))
TAG     := v$(SEMVER)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

EXTLDFLAGS = $(LDFLAGS) \
	-X main.version=$(SEMVER) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

# ── Platforms ──────────────────────────────────────────────────────
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
DIST      := dist

# Run target file
FILE ?= README.md

# Install prefix
PREFIX ?= $(HOME)/.local
BINDIR := $(PREFIX)/bin

# Host platform
OS   := $(shell uname -s | tr A-Z a-z)
ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
  ARCH := amd64
endif
ifeq ($(ARCH),aarch64)
  ARCH := arm64
endif

# ── Default ────────────────────────────────────────────────────────
.PHONY: all
all: build

# ── Version ────────────────────────────────────────────────────────
.PHONY: version
version:
	@echo "version: $(SEMVER)"
	@echo "tag:     $(TAG)"
	@echo "commit:  $(COMMIT)"
	@echo "date:    $(DATE)"

# ── Local build ────────────────────────────────────────────────────
.PHONY: build
build:
	$(CGO) go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o bin/$(BINARY) $(CMD)

# ── Build with version info ────────────────────────────────────────
.PHONY: release-build
release-build:
	$(CGO) go build $(GOFLAGS) -ldflags="$(EXTLDFLAGS)" -o bin/$(BINARY) $(CMD)

# ── Install (platform-aware) ───────────────────────────────────────
.PHONY: install
install: release-build
ifeq ($(OS),linux)
	install -Dm755 bin/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
else ifeq ($(OS),darwin)
	install -d $(DESTDIR)$(BINDIR)
	install -m755 bin/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
else
	@mkdir -p $(DESTDIR)$(BINDIR)
	@cp bin/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY) 2>/dev/null || \
		cp bin/$(BINARY).exe $(DESTDIR)$(BINDIR)/$(BINARY).exe
endif
	@echo "installed $(BINDIR)/$(BINARY)"

.PHONY: uninstall
uninstall:
	rm -f $(DESTDIR)$(BINDIR)/$(BINARY)
	@echo "removed $(BINDIR)/$(BINARY)"

# ── Development ────────────────────────────────────────────────────
.PHONY: run
run:
	go run $(CMD) $(FILE)

.PHONY: test
test:
	$(CGO) go test -count=1 ./...

.PHONY: test-race
test-race:
	$(CGO) go test -race -count=1 ./...

.PHONY: test-cover
test-cover:
	$(CGO) go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "coverage.html"

.PHONY: cover
cover:
	$(CGO) go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

.PHONY: lint
lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

.PHONY: fmt
fmt:
	gofmt -w .

# ── Cross-platform archives (same layout as CI) ────────────────────
.PHONY: dist
dist:
	@rm -rf $(DIST)
	@mkdir -p $(DIST)
	@set -e; for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		stage=$(DIST)/$(BINARY)_$(SEMVER)_$${os}_$${arch}; \
		mkdir -p $$stage; \
		ext=""; [ "$$os" = windows ] && ext=".exe"; \
		echo "building $$os/$$arch"; \
		GOOS=$$os GOARCH=$$arch $(CGO) go build $(GOFLAGS) -ldflags="$(EXTLDFLAGS)" \
			-o $$stage/$(BINARY)$$ext $(CMD); \
		cp README.md LICENSE $$stage/; \
		if [ "$$os" = windows ]; then \
			(cd $(DIST) && zip -qr $(BINARY)_$(SEMVER)_$${os}_$${arch}.zip $(BINARY)_$(SEMVER)_$${os}_$${arch}); \
		else \
			tar -czf $(DIST)/$(BINARY)_$(SEMVER)_$${os}_$${arch}.tar.gz -C $(DIST) $(BINARY)_$(SEMVER)_$${os}_$${arch}; \
		fi; \
		rm -rf $$stage; \
	done
	@cd $(DIST) && sha256sum *.tar.gz *.zip > checksums.txt 2>/dev/null || \
		(cd $(DIST) && shasum -a 256 *.tar.gz *.zip > checksums.txt)
	@echo "-> $(DIST)"
	@ls -1 $(DIST)

# ── Tagging ────────────────────────────────────────────────────────
# Creates the annotated tag locally. Does not push.
.PHONY: tag
tag:
	@if [ "$(SEMVER)" = "0.0.0" ]; then \
		echo "VERSION required, e.g. make tag VERSION=2.0.0"; exit 1; fi
	@if ! printf '%s' "$(SEMVER)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$'; then \
		echo "VERSION must look like 2.0.0 or 2.0.0-rc.1"; exit 1; fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "working tree is dirty; commit first"; exit 1; fi
	@if git rev-parse -q --verify "refs/tags/$(TAG)" >/dev/null; then \
		echo "tag $(TAG) already exists"; exit 1; fi
	git tag -a "$(TAG)" -m "$(BINARY) $(SEMVER)"
	@echo "tagged $(TAG)"

.PHONY: publish
publish:
	@if [ -z "$$(git tag -l '$(TAG)')" ]; then \
		echo "tag $(TAG) does not exist; run 'make tag VERSION=$(SEMVER)'"; exit 1; fi
	git push origin "$(TAG)"
	@echo "pushed $(TAG) — GitHub Actions now builds the release"

# Full local release: verify, dist, tag. Push is a separate, deliberate step.
.PHONY: release
release: lint test dist tag
	@echo ""
	@echo "release $(TAG) prepared."
	@echo "inspect $(DIST)/, then run: make publish VERSION=$(SEMVER)"

.PHONY: snapshot
snapshot: SEMVER := 0.0.0-snapshot
snapshot: dist
	@echo "snapshot archives in $(DIST)"

# ── Cleanup ────────────────────────────────────────────────────────
.PHONY: clean
clean:
	rm -rf bin $(DIST) coverage.out coverage.html .testbin/
	@echo "clean"

# ── Help ───────────────────────────────────────────────────────────
.PHONY: help
help:
	@printf 'Usage: make <target> [VERSION=x.y.z]\n\n'
	@printf 'Build:\n'
	@printf '  build           Build for this platform\n'
	@printf '  install         Build with version info and install to %s\n' "$(BINDIR)"
	@printf '  uninstall       Remove the installed binary\n\n'
	@printf 'Release:\n'
	@printf '  dist            Cross-build archives into %s/\n' "$(DIST)"
	@printf '  tag             Create annotated tag %s (local only)\n' "$(TAG)"
	@printf '  publish         Push %s; CI attaches binaries to the GitHub release\n' "$(TAG)"
	@printf '  release         lint + test + dist + tag\n'
	@printf '  snapshot        Build %s with a snapshot version\n\n' "$(DIST)"
	@printf 'Dev:\n'
	@printf '  run             Run from source (FILE=%s)\n' "$(FILE)"
	@printf '  test            Unit tests\n'
	@printf '  test-race       Unit tests with the race detector\n'
	@printf '  test-cover      Tests + coverage.html\n'
	@printf '  cover           Coverage summary\n'
	@printf '  lint            go vet + gofmt + golangci-lint\n'
	@printf '  fmt             Format code\n'
	@printf '  clean           Remove build artifacts\n'
	@printf '  version         Print resolved version metadata\n'
