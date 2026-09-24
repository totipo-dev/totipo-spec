GO ?= go

.DEFAULT_GOAL := help
.PHONY: help test race conformance conformance-v0-rc1 verify verify-requirements check
.NOTPARALLEL: check

help:
	@printf '%s\n' \
	  'make test         Run the full test suite (Linux integration where supported)' \
	  'make race         Run the full suite with the race detector' \
	  'make conformance  Run the current normative corpus' \
	  'make conformance-v0-rc1  Run the frozen v0-rc1 profile' \
	  'make verify-requirements Verify profile integrity without running cases' \
	  'make verify       Verify vector and review checksums' \
	  'make check        Run all checks above'

.phase3-test-tmp:
	mkdir -m 700 -p "$@"

test: | .phase3-test-tmp
	TMPDIR="$(CURDIR)/.phase3-test-tmp" $(GO) -C conformance test -count=1 -timeout=180s ./...

race: | .phase3-test-tmp
	TMPDIR="$(CURDIR)/.phase3-test-tmp" $(GO) -C conformance test -race -count=1 -timeout=300s ./...

conformance:
	$(GO) run ./conformance/cmd/totipo-conformance ./vectors/v0

conformance-v0-rc1:
	$(GO) run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json

verify-requirements:
	$(GO) run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json --verify-only

verify:
	cd vectors/v0 && sha256sum -c manifest.sha256
	sha256sum -c conformance/review-inventory.sha256
	sha256sum -c review/phase2/inventory.sha256
	sha256sum -c review/phase3/source.sha256

check: test race conformance conformance-v0-rc1 verify verify-requirements
