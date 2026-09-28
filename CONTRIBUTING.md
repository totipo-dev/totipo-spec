# Contributing

The [v1/r15 specification](spec/totipo-vault-format-v1.md) is normative. The Go
consumer provides reference/conformance evidence and is not production code.
Design reviews in `review/` are supporting evidence.

Protocol changes need a rationale and corresponding cases. Do not silently alter
wire bytes, semantic rules, or case IDs. Record contradictions with the
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
README. Normal verification reads the corpus; fixture maintenance is a separate,
explicit reviewed action documented in [vectors/FORMAT.md](vectors/FORMAT.md).
Review all byte changes and moving-profile checksum changes together. Preserve
minimized fuzz regressions without automatically promoting them to normative
vectors.

Report suspected vulnerabilities through the process in [SECURITY.md](SECURITY.md).
The project is licensed under the [Apache License 2.0](LICENSE).

In r15, semantic cases evaluate explicit accepted snapshots. Local history only
supplies warnings; durable graph insertion and filesystem freshness/race proofs
are not conformance requirements. Pre-RC IDs may be removed/replaced when their
semantics disappear; document every delta and preserve frozen protocol bytes.

Every new manifest case must declare applicability as baseline or a recognized
conditional capability (`advisory-history`). Split mixed baseline protocol behavior
and optional diagnostic expectations where needed. A new capability requires a
spec definition, schema enum change, runner support, and documentation. Baseline
profile pins must equal exactly the manifest baseline set. All conditional cases
remain manifest-listed/hash-pinned and are executed by the reference capability
suite; optionality is never encoded as a SKIP or ignored failure.
