# Totipo

Totipo is an encrypted, append-only, multi-device TOTP vault format designed for synchronization over untrusted storage.

## Current protocol work

`main` is the working tree for the current Totipo protocol design.

The current specification is:

- **Totipo Vault Format v1**
- design revision **r9**
- moving pre-release-candidate conformance evidence

The normative protocol text is [`spec/totipo-vault-format-v1.md`](spec/totipo-vault-format-v1.md).

## Design direction

The current v1 design uses:

- immutable 1024-byte encrypted semantic objects;
- complete-state `TOKEN` assertions rather than field-level deltas;
- append-only durable authenticated topology on established clients;
- whole-state conflict handling;
- explicit candidate credential use when certainty is degraded;
- P-256 provenance separated from TOKEN state authority;
- an informational, non-causal `AUTHOR_TIME`;
- a frozen forward-compatible routing prefix so older clients can retain causal topology for future TOKEN/DEVICE versions without understanding their bodies.

A core availability principle is:

> Loss of semantic certainty should normally degrade capability rather than make authenticated candidate material unusable.

## Repository map

| Path | Role |
|---|---|
| `spec/` | Normative current protocol text |
| `vectors/` | Language-neutral current-protocol byte and semantic conformance vectors |
| `requirements/` | Moving and later frozen conformance profiles |
| `conformance/` | Go reference/conformance consumer; not production code |
| `review/` | Current-protocol design, consistency, adversarial, and simplification evidence |
| `tools/` | Repository/vector/spec verification tooling |

## Implementation evidence

The repository needs only one in-repo reference/conformance consumer: Go.

The first real Totipo implementation developed against the frozen vectors should act as the independent interoperability consumer before a v1 release candidate is frozen. There is no requirement to build a second throwaway reference implementation.

## Development and checks

Use the preserved Nix development environment (`nix develop`, or direnv with the
existing `.envrc`). It supplies Go, Make, Python, and the existing Go tooling.
The Go module supports Go 1.23 or later; the schema checker uses Python 3.9 or later.

```sh
make spec-check
make test
make conformance
make verify
make check
```

`make race` and `make fuzz` provide additional checks. Make uses a writable Go
build cache under `.direnv/` for the jailed development environment. When running
Go directly there, set `GOCACHE="$PWD/.direnv/go-build"` from the repository root.
The module tests run with `go -C conformance test ./...`.

The [manifest](vectors/manifest.json) currently contains 71 cases. The
[moving pre-RC profile](requirements/v1-pre-rc.json) pins their IDs and hashes;
it is not a release or a frozen RC profile. Normal checks never regenerate cases.

See the [vector contract](vectors/FORMAT.md), [Go consumer](conformance/README.md),
and [implementation report](review/V1_VECTOR_IMPLEMENTATION_REPORT.md).

## Current next step

Have the first live Totipo implementation independently consume the exact byte
and semantic corpus, including fixed signature verification and future opaque
routing. Complete platform persistence/crash and native P-256 interoperability
evidence before freezing any v1 release-candidate profile.
