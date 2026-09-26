# Contributing

The [v1/r10 specification](spec/totipo-vault-format-v1.md) is normative. The Go
consumer provides reference/conformance evidence and is not production code.
Design reviews in `review/` are supporting evidence.

Protocol changes need a rationale and corresponding cases. Do not silently alter
wire bytes, semantic rules, or stable case IDs. Record contradictions with the
affected sections, a reproducer, and the smallest proposed wording change.
Never regenerate expectations merely to make a failing implementation pass.

Use the preserved Nix development environment. Before proposing changes, run:

```sh
gofmt -l conformance
make check
make race
make fuzz
```

Run `go -C conformance vet ./...` with a writable Go cache, as described in the
README. Normal verification reads the corpus; fixture generation is a separate,
explicit maintenance action documented in [vectors/FORMAT.md](vectors/FORMAT.md).
Review all byte changes and moving-profile checksum changes together. Preserve
minimized fuzz regressions without automatically promoting them to normative
vectors.

Report suspected vulnerabilities through the process in [SECURITY.md](SECURITY.md).
The project is licensed under the [Apache License 2.0](LICENSE).
