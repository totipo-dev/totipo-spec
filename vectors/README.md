# v1 Conformance Vectors

The [manifest](manifest.json) lists 68 current v1/r9 cases, each with a permanent
ID, kind (`bytes`, `negative`, or `semantic`), specification sections, expected
outcome, file path, and SHA-256 checksum. Case files live under `cases/<category>/`.
The [case plan](CASE_PLAN.md) records the original stable IDs; all are represented.

Run `make conformance` to execute the corpus, or `make verify` to validate JSON
contracts, file checksums, and the moving requirements profile. Neither command
writes vectors. See [FORMAT.md](FORMAT.md) for the language-neutral data contract,
fixed ECDSA fixture handling, and deliberate generation workflow.

Exact expected bytes are normative where the manifest marks them normative.
Synthetic future tails deliberately contain bytes that are not valid v1 body
TLVs. The consumer authenticates them through the same 1024-byte envelope and
parses only their frozen routing metadata.

These cases are moving pre-RC evidence. Passing them does not establish full
application conformance, independent interoperability, or production readiness.
