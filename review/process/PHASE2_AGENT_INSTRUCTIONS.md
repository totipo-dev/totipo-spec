# Totipo Conformance Suite — Phase 2 Agent Instructions

You are continuing work in the existing Totipo repository.

The current normative protocol specification is:

```text
spec/totipo-vault-format-v0.md
```

Treat that file as normative. The existing Go conformance implementation and reviewed vectors are evidence/implementations, not alternate protocol definitions.

The current implementation checkpoint reports:

```text
270 PASS
0 FAIL
strict Ed25519 profile: BLOCKED by missing reviewed negative corpus
stateful lifecycle/recovery work: incomplete
```

Do not redesign the protocol to make implementation easier.

Do not alter existing reviewed vector expectations unless an actual error in the reviewed artifact is established and documented.

Do not silently promote implementation behavior into normative expected behavior.

## Goal

Complete the next major conformance-evidence layer:

1. build and review a strict Ed25519 acceptance/rejection corpus;
2. build a language-neutral Totipo r36 semantic/lifecycle transition corpus;
3. build language-neutral recovery/security-memory state-machine fixtures;
4. add DEVICE_UPDATE presentation-history fixtures;
5. extend the Go implementation so it consumes those corpora independently;
6. clearly report anything still blocked rather than inventing expectations.

This phase is primarily about producing reviewed expected behavior and consuming it, not adding more byte-format design.

---

## Part 1 — Inventory before coding

First inspect the repository and identify:

```text
spec/totipo-vault-format-v0.md

vectors/v0/
conformance/
review/
```

Read the current implementation report and existing vector README.

Do not regenerate existing passing byte/crypto/bootstrap/TOTP vectors.

Confirm that the existing test suite still passes before making changes:

```bash
go -C conformance test -count=1 ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
go -C conformance test -race ./...
```

Record the baseline count.

---

## Part 2 — Strict Ed25519 acceptance corpus

The current implementation correctly does NOT claim Totipo Ed25519-profile conformance merely because Go's standard verifier accepts genuine signatures.

Close that gap.

### 2.1 Source of truth

Use Section 16 of the Totipo specification as normative.

The corpus must test Totipo's complete acceptance profile, including:

```text
DEVICE_PUBLIC_KEY exactly 32 bytes
SIGNATURE exactly 64 bytes
canonical point decoding
canonical point re-encoding
rejection of small-order points
rejection of mixed-order/non-prime-subgroup points
true [L]P subgroup membership checks
S interpreted little-endian
0 <= S < L
cofactorless verification equation
```

Do not weaken these requirements to match `crypto/ed25519.Verify`.

### 2.2 Corpus format

Create language-neutral vectors under:

```text
vectors/v0/ed25519/
```

Use JSON.

Each case should contain at minimum:

```json
{
  "id": "v0/ed25519/...",
  "description": "...",
  "public_key": "...hex...",
  "message": "...hex...",
  "signature": "...hex...",
  "expected": "ACCEPT" | "REJECT",
  "reason": "...",
  "provenance": {
    "kind": "...",
    "source": "..."
  }
}
```

Do not rely on one implementation's result as the expected answer.

### 2.3 Required positive cases

Include at minimum:

- valid RFC-style Ed25519 vector(s);
- existing Totipo TOKEN_UPDATE signature fixture;
- existing Totipo DEVICE_UPDATE signature fixture;
- at least one additional independently produced valid signature.

### 2.4 Required negative families

Create reviewed cases covering at least:

- public key wrong length;
- signature wrong length;
- non-decodable public key;
- non-decodable R;
- non-canonical public-key encoding;
- non-canonical R encoding;
- identity/small-order A;
- identity/small-order R;
- mixed-order A;
- mixed-order R;
- A outside the prime-order subgroup;
- R outside the prime-order subgroup;
- `S == L`;
- `S > L`;
- maximum 256-bit S;
- modified R;
- modified S;
- modified public key;
- modified message;
- mathematically invalid signature that still has structurally canonical encodings.

The mixed-order/subgroup cases are especially important.

### 2.5 Independent expectation generation

Do not derive expected ACCEPT/REJECT from Go's standard verifier.

Implement or use a separate review/reference calculation capable of testing the exact Section 16 point/subgroup conditions.

If the repository lacks a suitable historical strict implementation, create a small isolated reference tool for corpus review.

Keep the reference generator/reviewer outside the implementation-under-test path, e.g.:

```text
review/ed25519/
```

The final committed JSON corpus must contain fixed expected results.

Normal tests must never regenerate it.

### 2.6 Go implementation

Implement:

```text
conformance/internal/ed25519profile
```

with an API conceptually like:

```go
func Verify(publicKey, message, signature []byte) error
```

The implementation must enforce the full Totipo acceptance profile.

Add tests that run the complete external corpus.

Also run Go's standard `crypto/ed25519.Verify` against the corpus for comparison and document any behavioral differences, but do not use that comparison as protocol truth.

If a minimal external curve primitive is required, justify it and pin the dependency. Do not add a large generic crypto framework unnecessarily.

---

## Part 3 — Build the full r36 semantic transition corpus

The repository currently has only an initial symbolic lifecycle corpus.

Expand this into a reviewed, language-neutral corpus.

Create:

```text
vectors/v0/transitions/
vectors/v0/lifecycle/
```

Do not make expected results depend on Go implementation output.

### 3.1 Fixture representation

A transition/history fixture should be able to describe:

```text
objects/events
TOKEN_ID
signer identity label
context parents
asserted semantic fields
dependency availability
observation frontier
expected validation disposition
expected current token-update heads
expected field-head sets
expected displayed values where relevant
expected lifecycle-conflict tuples
expected resolution coverage
expected writer eligibility where specified
```

Use symbolic names for objects where convenient.

Example shape:

```json
{
  "id": "v0/lifecycle/delete-vs-credential-rotation",
  "events": [
    {
      "id": "O",
      "parents": [],
      "fields": {
        "STATUS": "LIVE",
        "ISSUER": "issuer",
        "ACCOUNT": "account",
        "CREDENTIAL": "k0"
      }
    }
  ],
  "view": ["D", "W"],
  "expected": {
    "token_heads": ["D", "W"],
    "lifecycle_conflicts": [
      {
        "status_head": "D",
        "credential_witness": "W"
      }
    ]
  }
}
```

The exact schema may differ, but document it.

---

## Part 4 — Mandatory lifecycle cases

Translate the Section 74 lifecycle requirements into explicit fixtures.

At minimum cover:

### Creation and ordinary field behavior

- creation requires LIVE + ISSUER + ACCOUNT + CREDENTIAL;
- issuer edit;
- account edit;
- credential rotation;
- multi-field update;
- concurrent issuer/account compose;
- concurrent issuer/credential compose;
- concurrent same-field edits conflict;
- equal-valued concurrent field revisions retain distinct lineage;
- unrelated field assertion preserves an existing conflict;
- later assertion of conflicted field resolves every visible alternative;
- no partial ordinary field merge;
- one TOKEN_UPDATE head may contain inherited field conflicts.

### Lifecycle conflicts

- deletion vs concurrent credential rotation;
- restoration vs concurrent credential rotation;
- historical non-head credential witness;
- current-credential-head-only calculation fails the named counterexample;
- hidden stale credential descendant creates conflict after earlier resolution;
- multiple credential witness branches;
- multiple STATUS heads;
- equal STATUS values on distinct revisions;
- ordinary STATUS conflict without credential witness;
- lifecycle transition detection with multiple STATUS parents.

### Lifecycle resolution

- correctly constructed status-only resolution;
- candidate provisional coverage only for its own POST*;
- candidate POST* must be empty;
- invalid candidate must contribute no coverage;
- STATUS plus another field with PRE nonempty routes to Section 51 and is invalid before candidate coverage;
- PRE empty + STATUS plus another field follows ordinary rules;
- later hidden history creates a new conflict without retroactively invalidating the earlier resolution;
- resolution may assert same scalar STATUS as an existing status head.

### Credential/status restrictions

- credential-only update when all current status heads LIVE;
- credential-only update when any current status head TOMBSTONE -> invalid;
- credential + STATUS=LIVE from tombstoned state when PRE empty;
- STATUS=TOMBSTONE + CREDENTIAL -> invalid;
- lifecycle conflict + issuer edit remains valid and conflict persists;
- lifecycle conflict + account edit remains valid and conflict persists;
- lifecycle conflict + credential-only update with all status heads LIVE can remain valid while lifecycle evidence persists.

---

## Part 5 — Differential lifecycle oracle

Implement a literal reference oracle separate from the main incremental model.

The point is not performance.

Use the most straightforward interpretation possible:

```text
explicit graph traversal
explicit causal closure
explicit typed-field ancestry
explicit credential-history scan
explicit raw-witness calculation
explicit CONFIRMED/coverage calculation
explicit maximal-witness filtering
```

Avoid sharing:

- ancestry caches;
- JOIN helpers;
- incremental field-state code;
- lifecycle optimization code

with the primary Go model.

Put it in a clearly separate package, e.g.:

```text
conformance/internal/referenceoracle/
```

Run every lifecycle fixture through both:

```text
primary model
literal reference oracle
```

and require identical expected state.

Also retain deliberately broken oracle variants such as:

```text
current credential heads only
confirmation covers descendants in wrong direction
resolution coverage does not propagate through STATUS ancestry
```

and prove that at least one fixture distinguishes each incorrect implementation.

---

## Part 6 — Property/randomized semantic testing

Add deterministic randomized history generation.

Use fixed seeds committed in tests.

Generate valid histories with:

- stale writers;
- same-field concurrency;
- disjoint-field concurrency;
- STATUS transitions;
- credential changes;
- lifecycle conflict creation;
- lifecycle resolution;
- hidden/revealed branches.

At every accepted state compare:

```text
primary model
reference oracle
```

Check at minimum:

```text
JOIN idempotence
JOIN commutativity
JOIN associativity
field ancestry implies causal ancestry
candidate POST* emptiness
no accepted invalid transition
no invalid candidate coverage leakage
```

Do not turn randomly generated results into normative vectors automatically.

Randomized tests are implementation evidence, not protocol vectors.

---

## Part 7 — Recovery/security-memory corpus

Create:

```text
vectors/v0/recovery/
```

These are state-machine fixtures, not merely semantic DAG fixtures.

Model local durable state explicitly.

At minimum fixture state should be able to represent:

```text
VAULT_BINDING
remembered token heads
remembered DEVICE_UPDATE heads
pending TOKEN_UPDATE evidence
pending abandonment decisions
pending warning acknowledgement
future-version observations
future bypass decisions
known-history regression
establishment state
```

Transitions should include crash/restart boundaries.

---

## Part 8 — Pending TOKEN_UPDATE recovery cases

Add fixtures for:

- signature-verified pending update blocks only its TOKEN_ID;
- evidence becomes durable before processing is considered complete;
- file disappears after pending observation -> block remains;
- restart -> block remains;
- explicit exact-ID abandonment permits authorship;
- abandonment does not erase evidence;
- newly discovered pending update is not covered by old abandonment;
- abandoned update later becomes FULLY_VALID and re-enters history;
- accepted-frontier handoff becomes durable before pending evidence is removed;
- identity-grounded terminal invalidity permits eventual evidence removal;
- wrong-length/missing/AEAD-failing current pathname does NOT establish terminal invalidity;
- known-history regression cannot be bypassed through pending recovery.

---

## Part 9 — Future-version evidence cases

Add fixtures for:

- authenticated future object creates durable vault-wide observation;
- observation survives file disappearance;
- observation survives restart;
- un-bypassed observation blocks v0 writes;
- exact durable bypass allows v0 writes from understood subset;
- bypass does NOT erase future observation;
- bypass does NOT establish completeness;
- newly observed future object re-establishes blocking;
- compatible future processing may supersede evidence only through an explicit durable handoff.

---

## Part 10 — Known-history regression

Add fixtures for:

```text
remembered X, current descendant H where X <=c H -> no regression

remembered X, unrelated current branch -> DEGRADED_KNOWN_HISTORY_LOSS

remembered fork heads, reunion descendant covers both -> no regression

remembered nonempty, current reconstructible heads empty -> degraded
```

Verify:

- degraded token cannot author;
- regressed state cannot be presented as ordinary complete current state;
- unrelated healthy token remains writable;
- pending-history abandonment cannot clear degradation;
- restoring missing validated history restores normal state;
- remembered heads cannot be deleted merely through "recovery."

---

## Part 11 — Durable replacement ordering

Add explicit crash-point fixtures around Section 63.1.

Examples:

```text
old remembered head durable
new descendant computed but not durable
crash
=> old evidence remains

new descendant durable
old evidence not yet removed
crash
=> conservative duplicate evidence allowed

new descendant durable
old evidence removed
=> valid compacted state
```

Likewise for:

- pending-to-accepted handoff;
- DEVICE_UPDATE remembered heads;
- recovery/bypass records;
- non-atomic batch updates.

Prove that a shared flush is not treated as transactional atomicity.

---

## Part 12 — DEVICE_UPDATE presentation history

Implement full DEVICE_UPDATE semantic validation and state derivation.

Add vectors covering:

- parentless DEVICE_UPDATE;
- simple rename;
- concurrent same-key names;
- later rename incorporating all same-key heads;
- identical displayed values on separate presentation heads;
- pending presentation dependency;
- pending same-key presentation evidence blocks another rename;
- exact abandonment via the Section 49.1 workflow;
- abandonment survives restart;
- abandonment remains presentation-only;
- later-valid abandoned branch re-enters presentation history;
- presentation rollback warning;
- presentation history never affects TOKEN_UPDATE validity.

---

## Part 13 — Confirmation freshness and publication linearization

Model writer-side operation preparation explicitly.

Add fixtures/events such as:

```text
confirm
observe new token head
publish attempt
=> confirmation stale

confirm
observe pending-history change
=> stale

confirm
future-observation set changes and later returns to same value
=> still stale

confirm
restart
=> stale

confirm
all inputs unchanged
publication enters linearization point
new remote history arrives afterward
=> already-published object may remain valid; later state may conflict
```

The test model must distinguish:

```text
before publication linearization point
after publication linearization point
```

Ensure new blocking observations discovered before the linearization point force abort/retry.

---

## Part 14 — Establishment state tests

Add local-state-machine tests for:

```text
UNESTABLISHED
PENDING(VAULT_BINDING)
ESTABLISHED(VAULT_BINDING)
```

Cover:

- racing initializers;
- establishment record compare-and-set/serialization;
- PENDING persisted before canonical bootstrap installation;
- crash before VAULT install;
- crash after VAULT install but before ESTABLISHED;
- conflicting existing VAULT;
- differing K_root cannot complete pending establishment;
- first open does not expose ordinary state before durable binding;
- established device key cannot silently migrate to another root.

These may use an in-memory durable-store simulator first.

Filesystem fault injection can come later.

---

## Part 15 — Capacity boundary and migration

Add fixtures for:

- exactly 32 current parent heads -> operation encodable;
- 33 current heads -> operation capacity-blocked;
- no head is silently omitted;
- no conflict/state truncation;
- no fake merge/adoption event;
- migration from a capacity-blocked source can select resolved values locally without first publishing an impossible old-vault resolution;
- migration creates fresh TOKEN_IDs and new root history;
- migration makes no continuity/retirement/erasure claim.

---

## Part 16 — Application safety profile tests

Implement tests for the r36 ordinary OTP-use/export profile.

Normal current-token use must be blocked when:

```text
known-history regression
validation incomplete/resource-limited
lifecycle conflict present
STATUS value ambiguous
not unambiguously LIVE
credential value ambiguous
credential incomplete
```

Equal-valued concurrent credential revisions may count as one value for use while lineage remains distinct.

ISSUER/ACCOUNT conflict alone need not disable OTP generation, but must remain surfaced.

A tombstoned lifecycle-conflicted token must remain discoverable in recovery/conflict UI even when not shown in ordinary active-token views.

Keep these tests clearly separated from consensus object validity.

---

## Part 17 — Filesystem and crash tests

Do not build a giant filesystem simulator yet.

First introduce an abstract storage interface for publication tests.

Then test:

- safe exclusive temp creation;
- no symlink following where hostile namespace applies;
- reject/avoid special files;
- object publication via complete temp + durable publication step;
- bootstrap no-replace initial creation;
- bootstrap atomic same-root replacement;
- interrupted temporary writes;
- conflicting canonical destination;
- directory durability hooks;
- crash points around durable local security-memory updates.

Use real filesystem integration tests only after deterministic state-machine tests pass.

Make OS-specific assumptions explicit.

---

## Part 18 — Provenance rules

Every new normative JSON case must identify provenance.

Allowed provenance kinds should include:

```text
reviewed-pinned
published-standard
spec-derived-reviewed
```

Do not call a spec-derived fixture "reviewed" merely because a test passes.

Before promoting a newly authored lifecycle/recovery fixture to normative corpus:

1. derive expected behavior manually from exact spec sections;
2. have the reference oracle agree;
3. have the primary implementation agree;
4. document the relevant sections;
5. mark it `spec-derived-reviewed` only after explicit review.

Random/property/fuzz-generated cases remain non-normative unless deliberately promoted.

---

## Part 19 — Update the manifest

After adding new normative JSON vectors:

```text
vectors/v0/manifest.sha256
```

must be regenerated explicitly.

Normal tests and CI must verify but never regenerate it.

Maintain:

```text
review-inventory.sha256
```

for historical inputs separately.

---

## Part 20 — CLI/reporting

Extend the existing CLI so categories can be filtered:

```bash
--filter v0/ed25519/
--filter v0/lifecycle/
--filter v0/recovery/
--filter v0/presentation/
```

The runner must clearly distinguish:

```text
PASS
FAIL
BLOCKED
```

Do not report unimplemented required categories as PASS with zero tests.

---

## Part 21 — Required final checks

Run at minimum:

```bash
gofmt -l conformance tools
go -C conformance fmt ./...
go -C conformance vet ./...
go -C conformance test -count=1 ./...
go -C conformance test -race ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
(cd vectors/v0 && sha256sum -c manifest.sha256)
sha256sum -c conformance/review-inventory.sha256
```

Run short smoke fuzzing for existing fuzz targets after model changes.

Add additional fuzzing only after deterministic stateful tests are stable.

---

## Part 22 — Do not change the spec casually

Do not edit:

```text
spec/totipo-vault-format-v0.md
```

just because a test is difficult to implement.

If a genuine contradiction is found, create:

```text
SPEC_ISSUES.md
```

with:

```text
issue ID
sections involved
minimal history/input
expected behavior according to section A
expected behavior according to section B
why both cannot hold
whether wire bytes are affected
proposed resolution
```

Leave the relevant test BLOCKED/FAIL until reviewed.

---

## Deliverables

Complete as much of the following as supported by normative/reviewable expectations:

```text
vectors/v0/ed25519/
vectors/v0/transitions/
vectors/v0/lifecycle/
vectors/v0/recovery/
vectors/v0/presentation/
```

plus:

```text
strict Go Ed25519 profile verifier
literal independent lifecycle oracle
expanded primary token model
DEVICE_UPDATE model
durable security-memory state machine
confirmation/publication state machine
establishment state machine
capacity/migration tests
application safety profile tests
updated CLI
updated manifest
updated README documentation
```

Produce:

```text
PHASE2_IMPLEMENTATION_REPORT.md
```

The report must include:

- baseline PASS/FAIL/BLOCKED counts before this phase;
- final counts by category;
- exact number of Ed25519 positive/negative cases;
- whether Go standard Ed25519 behavior differs from Totipo strict profile;
- lifecycle vector count;
- randomized differential-history count;
- recovery/state-machine vector count;
- DEVICE_UPDATE vector count;
- property-test counts;
- fuzz smoke results;
- new dependencies and why each was needed;
- every remaining BLOCKED requirement;
- every newly authored normative fixture and its provenance class;
- any genuine spec ambiguity/contradiction;
- commands used to reproduce the results.

Do not claim full Totipo v0 conformance while any mandatory category remains BLOCKED.

## Important Ed25519 constraint

Create the Ed25519 corpus yourself only if you can independently establish the expected results from the exact Section 16 mathematics.

If you cannot establish those expectations confidently, keep Ed25519 `BLOCKED` rather than inventing negative vectors. That is the correct failure mode for this project.
