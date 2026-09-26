# Totipo

Totipo is an encrypted, append-only, multi-device TOTP vault format designed for synchronization over untrusted storage.

## Current protocol work

`main` is the working tree for the current Totipo protocol design.

The current specification is:

- **Totipo Vault Format v1**
- design revision **r11**
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

## Envelope-family discovery

`objects-v1/` is the exact v1 envelope/storage-family namespace. `OBJECT_VERSION`
versions semantic content inside that fixed 1024-byte family. Unknown sibling
namespace names are not authenticated future-version evidence. A name such as
`objects-v2/` is illustrative; v1 does not parse that family's contents.

A future family claiming rolling compatibility publishes authenticated
compatibility assertions into `objects-v1/`. Old clients learn the semantic effect
only from that projection, using the existing supported/opaque routing rules.

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

The [manifest](vectors/manifest.json) currently contains 77 cases. The
[moving pre-RC profile](requirements/v1-pre-rc.json) pins their IDs and hashes;
it is not a release or a frozen RC profile. Normal checks never regenerate cases.

See the [vector contract](vectors/FORMAT.md), [Go consumer](conformance/README.md),
and [r10 integration report](review/V1_R10_INTEGRATION_REPORT.md).

## Current next step

Have the first live Totipo implementation independently consume the exact byte
and semantic corpus, including fixed signature verification and future opaque
routing. Complete platform persistence/crash and native P-256 interoperability
evidence before freezing any v1 release-candidate profile.
