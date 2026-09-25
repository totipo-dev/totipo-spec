# Agent instructions: reset `totipo-spec` main to current v1 and build pre-vector evidence

You are working in the existing `totipo-spec` repository.

The human has decided:

- v0 is historical and should not remain mixed into the current `main` working tree.
- Do **not** rewrite Git history.
- `main` should become the clean working line for Totipo v1.
- The supplied bundle's `copy-to-repo/` tree is the starting current v1 content.
- One in-repo Go reference/conformance consumer is sufficient.
- Do not create a second throwaway implementation. The first real Totipo implementation will later act as the independent interoperability consumer before a v1 release candidate.
- Do not cut a v1 release/tag yet.
- Do not change the r9 wire format or semantics unless you find a concrete contradiction; report any such contradiction instead of silently redesigning it.

## Phase 0 — protect historical v0 before deleting anything from main

From repository root:

1. Run:
   ```sh
   git status --short
   git branch --show-current
   git rev-parse HEAD
   git fetch --tags --prune
   ```
2. Require a clean working tree before destructive cleanup. If it is not clean, stop and report the exact uncommitted paths.
3. Record the current pre-reset main commit:
   ```sh
   OLD_V0_HEAD="$(git rev-parse HEAD)"
   ```
4. Inspect existing tags and releases/release metadata if available:
   ```sh
   git tag --contains "$OLD_V0_HEAD" || true
   git tag --points-at "$OLD_V0_HEAD" || true
   ```
5. Preserve the old tree even if a release tag already exists:
   ```sh
   git branch archive/v0 "$OLD_V0_HEAD"
   ```
   If `archive/v0` already exists, verify it points to the intended historical v0 line. Do not move an existing archive branch silently.
6. If authorized to push branches, push `archive/v0`. Do not force-push.
7. Never delete or retarget existing v0 tags/releases.

The goal is that someone can still inspect the exact historical v0 repository state, but new contributors looking at `main` see only the current protocol work.

## Phase 1 — reset the current working tree

Preserve repository/legal/community/infrastructure files unless they are demonstrably v0-protocol artifacts:

```text
.git/
.github/          # update workflows as required
.gitignore
.gitattributes
LICENSE
CONTRIBUTING.md
SECURITY.md

flake.nix
flake.lock
shell.nix
default.nix
.nix/             # if present
nix/              # if present
```

**Nix is part of the current project dependency/development environment and MUST be preserved.**

Do not delete or recreate the Nix setup merely because it currently mentions v0-era tooling. Instead:

1. inspect the existing Nix files;
2. preserve their dependency/toolchain intent;
3. remove only stale v0-specific commands/paths;
4. add any new Go/vector/spec-check dependencies required by the v1 tree;
5. update the lock file only if dependency changes actually require it;
6. document any Nix dependency change in the implementation report.

Remove old protocol-specific working-tree content from `main`, including the old:

```text
spec/
vectors/
requirements/
review/
conformance/
tools/
REPO_CLEANUP_REPORT.md
```

Rewrite other build orchestration such as `Makefile` or `go.work` only as needed for the new v1 tree. Preserve useful generic build/dev-environment behavior.

Then copy **the contents** of the supplied `copy-to-repo/` directory into repository root.

Run:

```sh
python3 tools/check_spec.py
```

Expected:

```text
PASS: Totipo v1/r9 structural spec checks
```

At this point, inspect the resulting tree. The current-protocol tree should be understandable without knowing v0 ever existed.

Do not add a `v0/` subdirectory back to main.

## Phase 2 — make README/build/CI reflect current v1 only

Use the supplied README as the baseline.

Update `.github/`, Makefile, Go workspace/module files, and the **preserved Nix/dev tooling** so current checks refer only to current v1 work.

For Nix specifically:

- keep the existing development shell / reproducibility model;
- keep useful existing package inputs unless they are truly obsolete;
- add the Go/tooling dependencies needed by the v1 conformance work;
- make sure a fresh Nix shell can run the documented `make`/Go/spec-check commands;
- do not casually churn `flake.lock`.

The intended developer commands should become approximately:

```sh
make spec-check
make test
make conformance
make verify
make check
```

During the first implementation pass:

- `make spec-check` runs `tools/check_spec.py`.
- `make test` runs Go tests.
- `make conformance` runs the current v1 vector corpus.
- `make verify` validates manifest/file hashes/structure.
- `make check` runs all of the above.

Do not retain `conformance-v0-rc1`, `vectors/v0`, or `requirements/v0-rc1.json` targets on current main.

Historical instructions belong on the historical Git refs, not current README.

## Phase 3 — establish v1 vector schema before vector bytes

Use:

```text
vectors/manifest.schema.json
vectors/manifest.json
vectors/CASE_PLAN.md
```

as the starting contract.

Rules:

1. Never put a case into `manifest.json` before its referenced file exists.
2. Stable case IDs are permanent after the Go consumer starts depending on them.
3. Separate:
   - exact byte cases;
   - semantic/state cases;
   - negative/parser cases.
4. Include `spec_sections` in every real manifest entry.
5. Add hashes/checksums if useful, but keep the manifest language-neutral.
6. Do not freeze `requirements/v1-rc1.json`.

## Phase 4 — implement the Go reference/conformance consumer

Create a new `conformance/` Go module/workspace suitable for the v1 spec.

This is reference/conformance code, not production code.

### Dependency policy

Keep dependencies minimal.

Prefer Go standard library for:

- SHA-256;
- HMAC-SHA-256;
- HKDF if available in the supported toolchain, otherwise the smallest standard/official solution;
- AES-256-GCM;
- ECDSA P-256;
- ASN.1 DER handling;
- JSON/test infrastructure.

Argon2id may use the official/minimal `golang.org/x/crypto/argon2` dependency.

Do not introduce general CBOR/protobuf/TLV frameworks. Implement the small canonical TLV codec directly.

### Required packages/areas

Organize cleanly, for example:

```text
conformance/
    cmd/totipo-conformance/
    internal/spec/
    internal/tlv/
    internal/crypto/
    internal/object/
    internal/graph/
    internal/vectors/
```

Exact package names may differ, but keep encoding, crypto, graph semantics, and vector IO separable.

### Implement in this order

1. Canonical TLV primitive encoding/decoding.
2. Frozen routing-prefix parser independent of the v1 body parser.
3. v1 TOKEN and DEVICE full parsers/encoders.
4. Explicit v1 DEVICE_ID derivation/equality check.
5. `SUPPORTED_VALID` / `OPAQUE_ROUTABLE` / `OPAQUE_UNSCOPED` dispatch.
6. Semantic size/capacity planning with 72-byte signature reservation.
7. Object-ID/HKDF/AES-GCM envelope processing.
8. P-256 provenance sign/verify fixtures.
9. VAULT bootstrap / Argon2id.
10. Durable-graph semantic evaluator used by language-neutral state cases.
11. Manifest-driven conformance command.

Important forward-compatibility rule:

- For an unsupported TOKEN/DEVICE version, authenticate exact semantic bytes and parse only the frozen routing prefix.
- Do **not** parse its version-specific tail as v1.
- Future semantic version number never orders the graph.
- Opaque scoped TOKEN state affects only its TOKEN_ID.
- Opaque-unscoped evidence blocks authoritative semantics but must not make explicit supported candidate material unavailable.

## Phase 5 — generate routing/encoding vectors first

Before full crypto vectors, implement these initial stable cases from `vectors/CASE_PLAN.md`:

```text
v1.routing.token-v1.001
v1.routing.token-future-opaque.001
v1.routing.device-v1.001
v1.routing.device-future-opaque.001
v1.routing.unknown-type-unscoped.001
v1.routing.future-token-malformed-prefix.001
v1.routing.future-device-malformed-prefix.001

v1.encoding.token-root.001
v1.encoding.device-root.001
v1.encoding.parent-order.001
v1.encoding.parent-count-mismatch.001
v1.encoding.author-time-zero.001
v1.encoding.author-time-u64max.001

v1.device.explicit-id.001
v1.device.id-mismatch.001

v1.size.token-max-4.001
v1.size.token-max-5-fold.001
v1.size.device-max-14.001
v1.size.device-max-15-fold.001
v1.size.short-der-no-extra-parent.001
```

### Forward-version fixture rule

Synthetic future-version fixtures are deliberate test data.

They MUST:

- use the same authenticated envelope family;
- preserve the frozen routing prefix;
- use `OBJECT_VERSION != 1`;
- contain an opaque tail that a v1 reader does not interpret.

Also include malformed-prefix and unknown-type cases to exercise `OPAQUE_UNSCOPED` vs invalid current storage.

## Phase 6 — full crypto vectors

Then create representative exact objects.

Every normative full crypto vector should record enough intermediate values to diagnose disagreement:

```text
input field values
unsigned semantic bytes
signature input
fixed private/public fixture key or fixed externally supplied signature fixture
actual canonical DER signature
final signed semantic bytes
OBJECT_ID
derived object key
nonce
AAD
SEMANTIC_LENGTH
padded plaintext
ciphertext
GCM tag
final 1024-byte object
```

### Important ECDSA rule

ECDSA signing is randomized on relevant native platforms.

Do not require independent implementations to reproduce a freshly generated signature from private key + message.

Instead:

- vectors may contain a fixed valid DER signature fixture;
- consumers verify that signature;
- encoder/signing tests verify semantic/signature-input bytes separately;
- writer capacity planning always reserves 72 bytes regardless of the fixture's actual DER length;
- never retry ECDSA signing to make an object fit.

## Phase 7 — semantic graph corpus

Port the r9 semantic model into language-neutral cases, especially:

```text
opaque future TOKEN becomes current
supported candidate remains explicitly usable
unrelated token remains ordinary
supported descendant moves opaque ancestor to history
semantic version number does not order state
opaque future DEVICE degrades presentation only
opaque-unscoped evidence blocks authoritative operations but not candidate use
current supported bytes disappear
conflicting heads
cycle => local integrity failure
global OBJECT_ID conflict => local integrity failure
```

Keep semantic cases compact; they do not all need encrypted 1024-byte fixtures.

## Phase 8 — requirements profile

Once real case IDs exist and the Go runner passes them, create:

```text
requirements/v1-pre-rc.json
```

This profile is explicitly moving.

It should pin the current required case IDs and enough metadata/checksums to detect accidental changes.

Do not create/freeze `v1-rc1.json` yet.

## Phase 9 — checks and reports

Before presenting work for human review, run:

```sh
python3 tools/check_spec.py
go test ./...
make check
git status --short
```

Also run any Go race/fuzz checks that are practical for the new parser/graph evaluator.

Produce:

```text
review/V1_VECTOR_IMPLEMENTATION_REPORT.md
```

containing:

- exact files added/removed;
- vector count by category;
- conformance PASS/FAIL/BLOCKED;
- dependencies introduced and why;
- crypto fixture-generation method;
- known gaps;
- anything in the spec that proved ambiguous while implementing.

If implementation uncovers a spec ambiguity, do not silently choose a new semantic rule. Record the ambiguity and the smallest proposed wording change.

## Phase 10 — commit boundaries

Prefer reviewable commits:

1. `archive v0 and reset main to v1/r9`
2. `add v1 vector schema and routing cases`
3. `add Go v1 conformance codec and routing parser`
4. `add v1 crypto and provenance vectors`
5. `add v1 semantic graph corpus`
6. `add moving v1 pre-rc requirements profile`
7. `add implementation evidence and CI`

Do not squash away the historical reset boundary unless the human explicitly requests it.

## Explicit non-goals

Do NOT:

- restore v0 protocol files to current main;
- delete the project's Nix development/dependency infrastructure;
- replace Nix with ad-hoc host dependencies;
- build a second synthetic/reference implementation;
- cut a release/tag;
- claim production readiness;
- add network/cloud synchronization code;
- add UI code;
- add a production password/TOTP application;
- silently change r9 wire format;
- use timestamps or semantic version numbers for causal ordering;
- make opaque future state disappear from the durable graph;
- let a v1 client author over an opaque current TOKEN.

## Final handoff

When finished, report:

1. old v0 HEAD and historical refs preserving it;
2. new main tree summary;
3. exact commits made;
4. vector manifest case count;
5. conformance result;
6. `make check` result;
7. Nix files preserved/changed and any dependency/lockfile changes;
8. any spec ambiguity or blocker;
9. recommended next step for the human's first live Totipo implementation.

Do not push a release or freeze an RC profile without explicit human direction.
