# v1 Conformance Vectors

The [manifest](manifest.json) lists 85 current v1/r12 cases, each with a permanent
ID, kind (`bytes`, `negative`, or `semantic`), specification sections, expected
outcome, file path, and SHA-256 checksum. Case files live under `cases/<category>/`.
The manifest is the live case contract; all original planned IDs are represented.

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

The TOTP category contains three RFC 6238 algorithm cases, each with six standard
timestamp rows and explicit secret/counter/code values. See [FORMAT.md](FORMAT.md).

Six r10 environment cases cover exact `objects-v1/` discovery, ignored entry names
and types, wrong sizes, ignored sibling namespaces, and future-family coexistence
with/without a v1 compatibility assertion. Existing authenticated fixtures are
referenced by case ID; their bytes are not duplicated or changed.

Eight r12 cases cover sticky opaque-unscoped evidence and deliberate continuity
reset, DEVICE convergence across rejected/unresolved heads, remote versus local
corruption, initial DEVICE publication, and signature-context vault binding.
The original 77 case files remain byte-identical.
