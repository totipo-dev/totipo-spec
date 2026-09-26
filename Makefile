GO ?= go
PYTHON ?= python3
# Keep Go's build cache writable in the Nix jailed development environment.
export GOCACHE ?= $(CURDIR)/.direnv/go-build

.DEFAULT_GOAL := help
.PHONY: help spec-check test race fuzz conformance verify check
.NOTPARALLEL: check

help:
	@printf '%s\n' 'make spec-check  Check the v1/r10 specification structure' 'make test        Run Go tests' 'make conformance Run all v1 cases' 'make verify      Verify manifest, files, and moving profile' 'make check       Run all required checks' 'make race        Run Go race tests' 'make fuzz        Run bounded parser fuzzing'

spec-check:
	$(PYTHON) tools/check_spec.py

test:
	$(PYTHON) -m unittest discover -s tools -p 'test_*.py'
	$(GO) -C conformance test -count=1 ./...

race:
	$(GO) -C conformance test -race -count=1 ./...

fuzz:
	$(GO) -C conformance test ./internal/object -run '^$$' -fuzz FuzzDispatch -fuzztime 10s -parallel 2
	$(GO) -C conformance test ./internal/cryptov1 -run '^$$' -fuzz FuzzOpen -fuzztime 10s -parallel 2
	$(GO) -C conformance test ./internal/graph -run '^$$' -fuzz FuzzArrivalAndDisappearance -fuzztime 10s -parallel 2

conformance:
	$(GO) run ./conformance/cmd/totipo-conformance -root .

verify:
	$(PYTHON) tools/check_vectors.py
	$(GO) run ./conformance/cmd/totipo-conformance -root . -verify-only

check: spec-check test conformance verify
