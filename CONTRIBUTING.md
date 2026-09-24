# Contributing

The [v0 specification](spec/totipo-vault-format-v0.md) is normative. Go code is reference/conformance evidence, and historical material under [review/](review/README.md) is provenance unless the specification explicitly incorporates it.

Protocol changes need a clear rationale and corresponding vectors/tests. Do not casually change frozen v0 wire bytes or historical protocol literals. Never silently regenerate normative vector expectations to match an implementation. Document specification contradictions with the affected sections and a minimal reproducer rather than resolving them by implementation preference.

Use Go 1.23 or later. Before proposing changes, run the relevant tests and the checks in [README.md](README.md), plus the review inventories from the repository root:

```sh
gofmt -l conformance tools review/ed25519/review.go
go -C conformance vet ./...
sha256sum -c conformance/review-inventory.sha256
sha256sum -c review/phase2/inventory.sha256
sha256sum -c review/phase3/source.sha256
```

Keep normative artifacts, generated/non-normative tests, and platform observations distinct. Explain any deliberate inventory/path update; preserve historical inputs and never run vector-writing maintenance commands as part of normal verification. Retain minimized fuzz regressions for review rather than treating them as normative expectations automatically.

See [SECURITY.md](SECURITY.md) for the unresolved private reporting channel. The project is licensed under the [Apache License 2.0](LICENSE).
