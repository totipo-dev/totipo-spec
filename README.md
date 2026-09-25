# Totipo

Totipo is an encrypted, append-only, multi-device TOTP vault format designed for synchronization over untrusted storage.

## Current protocol work

`main` is the working tree for the current Totipo protocol design.

The current specification is:

- **Totipo Vault Format v1**
- design revision **r9**
- pre-vector / pre-release-candidate

The normative protocol text is [`spec/totipo-vault-format-v1.md`](spec/totipo-vault-format-v1.md).

Historical v0 material is intentionally not duplicated in the current working tree. It remains part of the repository's Git history and historical tags/releases/archive references. This keeps `main` focused on the protocol being actively implemented and tested.

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

## Current next step

Freeze exact v1 routing/encoding vectors first, then full cryptographic object vectors, then semantic graph/candidate-use cases.

See:

- [`vectors/README.md`](vectors/README.md)
- [`vectors/CASE_PLAN.md`](vectors/CASE_PLAN.md)

No v1 release-candidate profile should be frozen until the vector IDs and expected bytes have stabilized.
