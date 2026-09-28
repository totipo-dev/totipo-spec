# v1 Conformance Vectors

The [manifest](manifest.json) lists 105 current v1/r15 cases (93 baseline and 12 conditional `advisory-history`), each with an explicit
ID, kind (`bytes`, `negative`, or `semantic`), specification sections, expected
outcome, file path, and SHA-256 checksum. Case files live under `cases/<category>/`.
The manifest is the live case contract; removed/replaced pre-RC IDs are documented.

Run `make conformance` for baseline, `make conformance-all` for baseline plus
`advisory-history`, or `make verify` to validate JSON
contracts, file checksums, and the moving requirements profile. Neither command
writes vectors. See [FORMAT.md](FORMAT.md) for the language-neutral data contract,
fixed ECDSA fixture handling, and deliberate maintenance workflow.

Exact expected bytes are normative where the manifest marks them normative.
Synthetic future tails deliberately contain bytes that are not valid v1 body
TLVs. The consumer authenticates them through the same 1024-byte envelope and
parses only their frozen routing metadata.

These cases are moving pre-RC evidence. Passing them does not establish full
application conformance, independent interoperability, or production readiness.

The TOTP category contains three RFC 6238 algorithm cases, each with six standard
timestamp rows and explicit secret/counter/code values. See [FORMAT.md](FORMAT.md).

Six r10 environment cases cover exact `objects-v1/` discovery, ignored entry names
and types, wrong sizes, ignored sibling namespaces, and future-family coexistence
with/without a v1 compatibility assertion. Existing authenticated fixtures are
referenced by case ID; their bytes are not duplicated or changed.

r15 replaces obsolete blocking/reset/mandatory-retention cases with explicit
snapshot authorship, ordinary concurrency, semantic confirmation, optional-history
warnings, independent publication acknowledgement, and simple binding establishment.
Every case delta and fixed-byte comparison is recorded in the
[r15 report](../review/V1_R15_STATE_MODEL_SIMPLIFICATION_REPORT.md).

Each manifest entry has strict `applicability`: `baseline` or `conditional` with
capability `advisory-history`. Conditional expected results are real requirements
when that capability is claimed, not SKIP outcomes. The baseline profile requires
no advisory cache or cross-run remembered history.
