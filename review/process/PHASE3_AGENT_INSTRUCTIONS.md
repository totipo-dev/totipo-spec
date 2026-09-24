# Totipo Conformance Suite — Phase 3 Agent Instructions

You are continuing work in the existing Totipo repository after completion of Phase 2.

The current normative protocol specification is:

```text
spec/totipo-vault-format-v0.md
```

Treat that file as normative.

The Phase 2 implementation report established an abstract/specification-level conformance checkpoint of:

```text
388 PASS
0 FAIL
0 BLOCKED
```

for the supplied corpus, including strict Ed25519, lifecycle semantics, recovery/security-memory state machines, DEVICE_UPDATE presentation, and application-use policy.

That result does **not** establish:

- a production client;
- real OS/filesystem correctness;
- real multi-process behavior;
- crash/power-loss durability;
- end-to-end integration of authenticated object files with durable local security memory;
- remote CI execution on every configured platform;
- external security audit.

This phase is about those implementation-boundary properties.

Do not redesign the protocol merely because OS/platform behavior is difficult.

Do not edit the spec unless a genuine normative contradiction is discovered and documented.

---

# Goal

Build a minimal end-to-end Totipo v0 reference implementation and exercise the protocol at the real storage/durability boundary.

The Phase 3 deliverables should demonstrate:

1. encrypted object-file ingestion through accepted semantic state;
2. durable local security-memory persistence;
3. crash/restart behavior at normative ordering points;
4. immutable object publication;
5. crash-safe `VAULT` creation/replacement;
6. hostile-path and symlink/special-file safety;
7. concurrent/multi-process establishment and publication behavior;
8. real out-of-order/missing/reappearing object handling;
9. integration of the Phase 2 semantic/state machines with real bytes and real persistence;
10. execution of the configured CI matrix where practical.

This is still a reference/conformance implementation, not a user-facing product.

---

# Part 1 — Preserve the Phase 2 baseline

Before changing code, run and record:

```bash
go -C conformance test -count=1 ./...
go -C conformance test -race ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
(cd vectors/v0 && sha256sum -c manifest.sha256)
sha256sum -c conformance/review-inventory.sha256
sha256sum -c review/phase2/inventory.sha256
```

The baseline should remain:

```text
388 PASS
0 FAIL
0 BLOCKED
```

If it does not, stop and diagnose before Phase 3 work.

Do not regenerate normative vectors.

---

# Part 2 — Add a minimal end-to-end reference client package

Create a package tree for an integration/reference client, for example:

```text
reference/
  client/
  storage/
  localstate/
  integration/
```

or an equivalent structure consistent with the repository.

The reference client must integrate existing conformance components rather than reimplement protocol rules unnecessarily.

The minimum pipeline is:

```text
object pathname
    ↓
safe open / bounded read
    ↓
2048-byte file-length gate
    ↓
AEAD decrypt/authenticate
    ↓
envelope validation
    ↓
OBJECT_ID recomputation
    ↓
common-prefix/version dispatch
    ↓
v0 TLV parse
    ↓
strict signature verification
    ↓
dependency validation
    ↓
semantic validation
    ↓
state/oracle update
    ↓
durable local security-memory transition
```

For authorship:

```text
accepted current context
    ↓
writer-gate/freshness checks
    ↓
construct canonical unsigned bytes
    ↓
sign
    ↓
OBJECT_ID
    ↓
deterministic envelope
    ↓
encrypt
    ↓
crash-safe object publication
    ↓
durable accepted-frontier update
    ↓
success acknowledgement
```

Do not expose an operation as successful before all required durable ordering steps complete.

---

# Part 3 — Define explicit interfaces for storage and local durable state

Do not hardwire filesystem and local security-memory logic directly into semantic code.

Define narrow interfaces.

Conceptually:

```go
type ObjectStore interface {
    ListCandidates(...) ...
    OpenObject(id ...) ...
    PublishImmutable(id ..., bytes ...) ...
    ReadVault(...) ...
    InstallInitialVault(...) ...
    ReplaceVaultSameRoot(...) ...
}

type SecurityStateStore interface {
    LoadVaultState(...) ...
    CompareAndSetEstablishment(...) ...
    PersistPendingEvidence(...) ...
    PersistFutureObservation(...) ...
    PersistAcceptedFrontier(...) ...
    PersistPresentationFrontier(...) ...
    PersistRecoveryDecision(...) ...
    Sync(...) ...
}
```

Exact APIs are implementation choices.

The interfaces must be sufficient to inject:

- crashes;
- partial durability;
- stale reads;
- concurrent writers;
- missing files;
- changed files;
- hostile path entries.

---

# Part 4 — Real filesystem adapter

Implement at least one real filesystem-backed object store for Linux first.

Prefer platform primitives that make the safety properties explicit.

The adapter must handle:

## Reading

- exact filename grammar;
- ordinary-file requirement;
- bounded reads;
- no following hostile symlinks where supported;
- reject or safely ignore directories, FIFOs, sockets, devices and other special files;
- avoid trusting size/mtime alone as immutable-content proof;
- close stable file handles correctly;
- reconsider previous failures if bytes later change.

## Temporary files

Use safe exclusive creation.

Requirements:

- unpredictable or collision-resistant temp name;
- exclusive creation;
- same filesystem/directory context as final destination where atomic rename semantics are required;
- no symlink traversal;
- no reuse of attacker-controlled existing temp path;
- restrictive permissions where applicable.

## Immutable object publication

The complete 2048-byte object file must be written and made durable before successful publication is reported.

Use:

```text
write complete temp
fsync temp
atomic/no-replace install to final object ID
fsync containing directory where supported
```

If the final immutable object already exists:

- if trustworthy bytes are identical, treat as already published;
- if bytes differ or cannot be trusted, do not overwrite it;
- surface the condition appropriately.

Do not silently replace immutable object files.

---

# Part 5 — Real `VAULT` publication

Implement the r36 bootstrap publication rules.

## Initial creation

Required ordering:

```text
construct candidate
validate candidate
derive VAULT_BINDING
durably persist PENDING(VAULT_BINDING)
confirm canonical VAULT absent
atomic no-replace install
durably flush directory metadata where supported
reopen canonical VAULT
unwrap and verify exact pending binding
durably transition to ESTABLISHED
only then report success
```

Do not truncate/write the canonical path directly.

## Same-root password rewrap

Required ordering:

```text
existing ESTABLISHED binding
construct 87-byte candidate
unwrap candidate to exact existing K_root/binding
write complete temp
durably flush temp
atomic replacement
durably flush directory metadata
report success
```

Do not allow replacement with a different root.

---

# Part 6 — Durable local security-memory implementation

Implement a real durable local-state adapter.

A small embedded database or carefully designed fsync-backed file format is acceptable.

Prefer a dependency with clear transactional/durability semantics if using a database.

Document the choice and why it is suitable.

The durable state must support at minimum:

```text
VAULT_BINDING
establishment state
REMEMBERED_TOKEN_HEADS
remembered DEVICE_UPDATE heads
pending TOKEN_UPDATE evidence
pending DEVICE_UPDATE evidence
pending abandonment decisions
pending warning acknowledgement
future-version observations
future bypass decisions
recovery state required by r36
```

The state store must live outside the hostile synchronized vault namespace.

Document explicitly:

- location/trust assumption;
- crash consistency model;
- transaction semantics;
- fsync/sync behavior;
- backup/rollback limitation.

Do not claim protection against rollback of the entire local security database unless you actually implement such protection.

---

# Part 7 — Crash injection framework

Add deterministic crash injection.

Every normative durability sequence should have named crash points.

Example:

```text
before write
after write before file fsync
after file fsync before rename
after rename before directory fsync
after directory fsync
after new security evidence persisted
before old evidence removal
after old evidence removal
```

A crash injection should:

1. stop the operation immediately;
2. discard volatile state;
3. reopen using persisted filesystem/database state;
4. run recovery;
5. assert the resulting state.

Do not simulate crashes only by returning an error while leaving in-memory state alive.

---

# Part 8 — Crash/restart test matrix

At minimum test:

## Initial `VAULT`

- crash before PENDING persistence;
- crash after PENDING persistence but before candidate creation;
- crash during candidate write;
- crash after candidate fsync before install;
- crash after canonical install before directory fsync;
- crash after canonical install before ESTABLISHED;
- restart completes only when canonical VAULT derives the pending binding;
- different canonical root does not complete establishment.

## Object publication

- crash during temp write;
- crash after temp fsync;
- crash after immutable install;
- crash before remembered accepted-frontier persistence;
- object may exist while local operation is not yet acknowledged accepted.

## Pending evidence

- pending evidence persisted, crash before normal processing continues;
- restart retains block;
- accepted-frontier handoff new evidence durable, crash before pending removal;
- terminal-invalidity handoff ordering.

## Future observations

- observation durable before continued processing;
- bypass durable before authorship relies on it;
- crash never erases observation merely because bypass exists.

## Remembered-frontier replacement

Test old/new evidence ordering exactly as Section 63.1 requires.

---

# Part 9 — Multi-process establishment races

Use real OS processes, not just goroutines.

Create integration tests that start two independent processes against the same configured local vault location.

Cover:

```text
creator A vs creator B
both initially observe absent VAULT
both attempt PENDING establishment
only one binding may win the configured-vault establishment state
only matching root may complete ESTABLISHED
loser must not overwrite binding or canonical VAULT
```

Test synchronization/compare-and-set semantics in the local security-state store.

A file-level no-replace operation alone is not sufficient evidence.

---

# Part 10 — Multi-process object publication races

Test two processes publishing:

## Identical object

Expected:

- same OBJECT_ID;
- one may install first;
- second observes already-existing identical immutable bytes;
- both may conclude the exact object is present;
- no corruption.

## Same final pathname, differing bytes

This should be treated as hostile/corrupt namespace behavior.

Expected:

- no overwrite;
- no silent replacement;
- no acceptance merely because pathname exists.

## Concurrent different objects

Expected independent safe publication.

---

# Part 11 — Hostile filesystem/path tests

On platforms where supported, create actual hostile entries.

Cover at minimum:

```text
object candidate pathname is symlink
object candidate pathname is FIFO
object candidate pathname is directory
object candidate pathname is socket where feasible
temp pathname already exists as symlink
canonical VAULT pathname is symlink
parent/path component substitution where feasible
```

Verify the implementation:

- does not block indefinitely on special files;
- does not follow symlinks into arbitrary paths;
- does not overwrite arbitrary targets;
- does not create unsafe side effects before cryptographic validation.

Keep OS-specific expected behavior isolated behind adapter tests.

---

# Part 12 — End-to-end object arrival scenarios

Use real encrypted files from the existing corpus or newly generated valid test objects.

Exercise actual filesystem arrivals:

## Out-of-order dependencies

```text
child appears first
parent missing
=> child PENDING
=> durable pending evidence recorded where required

parent later appears
=> child reprocessed
=> may become FULLY_VALID
```

## Object disappears after authentication

After a signature-verified pending observation:

```text
file disappears
restart
=> durable pending block remains
```

## Invalid-looking replacement

After prior authenticated observation:

```text
same pathname becomes wrong length / AEAD-invalid / substituted
=> current representation unavailable/untrustworthy
=> remembered security evidence not cleared
```

## Valid bytes reappear

The object must be reconsidered and processed normally.

---

# Part 13 — End-to-end future-version handling

Create real encrypted future-version objects using the v0-compatible envelope.

Test:

```text
authenticate/decrypt
OBJECT_ID matches
common prefix valid
OBJECT_VERSION != 0
opaque tail not parsed as v0
durable future observation persisted
v0 writing blocked unless exact durable bypass exists
```

Then:

```text
file disappears
restart
=> observation persists
```

Also verify malformed common prefix never reaches future dispatch.

---

# Part 14 — End-to-end writer gate/freshness tests

Integrate real durable state with writer preparation.

For a lifecycle resolution:

```text
derive accepted context
obtain simulated confirmation
inject new accepted head/pending evidence/future observation
attempt publish
```

If change occurs before publication linearization:

```text
abort/retry
do not publish stale operation
```

If new history arrives after publication linearization:

```text
published object may remain valid
subsequent state may expose conflict
```

Record the exact implementation linearization point.

It must correspond to the r36 rules.

---

# Part 15 — Reference client restart reconstruction

Create integration tests that:

1. build a vault and several valid objects;
2. close all processes;
3. discard all disposable derived caches;
4. restart from:
   - synchronized vault files;
   - durable security memory;
5. reconstruct state;
6. compare to the pre-restart accepted state.

Also test reconstruction with:

- missing synchronized objects;
- retained validated local immutable object copies;
- future-version evidence;
- abandoned pending history;
- presentation-history forks.

---

# Part 16 — Real local-state rollback limitation test/documentation

The spec does not claim rollback protection for the local security database itself.

Add an explicit integration/documentation test demonstrating:

```text
state S0 backup
observe/persist security evidence -> S1
restore old local database S0
```

The implementation must not claim that ordinary remembered-history guarantees survived the external rollback.

If a product-level mitigation is implemented, test it separately.

Do not silently assume the OS filesystem prevents old-backup restoration.

---

# Part 17 — Migration integration test

Build a real source vault state with >32 current token heads.

Verify:

- source operation is capacity-blocked;
- no parent/head truncation occurs;
- migration can select values locally;
- destination vault gets fresh K_root;
- destination device key is fresh;
- TOKEN_ID is fresh;
- destination root TOKEN_UPDATE contains all required creation fields;
- old vault remains unchanged;
- no retirement or erasure claim is made.

Use real CSPRNG output in this integration layer.

Assert identities differ; do not pin random bytes as normative vectors.

---

# Part 18 — CSPRNG and secret-generation integration

Test production/reference creation paths use the platform CSPRNG for:

```text
K_root
TOKEN_ID
device signing key
ARGON2_SALT
WRAP_NONCE
new generated credential secret where supported
```

Do not test statistical randomness.

Instead test:

- correct API path;
- correct lengths;
- no deterministic fixture source used in production/reference creation;
- generated credential secrets meet the 20-byte minimum.

Allow dependency injection of deterministic randomness only in tests.

---

# Part 19 — Real syscall/platform documentation

For each filesystem adapter, document:

```text
supported OS
atomic rename assumptions
no-replace primitive
symlink/no-follow primitive
file fsync behavior
directory fsync behavior
known limitations
```

Do not write generic "atomic" claims without naming the platform primitive/assumption.

Linux should be the first fully evidenced platform.

macOS and Windows can initially remain configured/partial if the exact durability semantics are not yet demonstrated.

---

# Part 20 — Remote CI execution

The Phase 2 report says CI is configured but not locally executed on other platforms.

Run the actual remote CI matrix.

At minimum collect results for:

```text
Linux
macOS
Windows
```

with the configured Go versions.

Separate:

```text
portable semantic/crypto tests
platform filesystem integration tests
race detector
```

Do not disable a failing platform integration test just to make CI green.

If a platform cannot satisfy a required primitive, report that adapter as unsupported or incomplete.

---

# Part 21 — Keep conformance layers distinct

Maintain three clear layers:

## Layer A — normative vectors

```text
vectors/v0/
```

These stay language-neutral and reviewed.

## Layer B — abstract/reference conformance

Existing Phase 1/2 Go models and state machines.

## Layer C — integration/platform conformance

Real bytes + filesystem + durable local state + process/crash behavior.

Do not promote platform-specific observations into normative protocol vectors.

---

# Part 22 — Phase 3 test classification

Use explicit test labels/categories, e.g.:

```text
integration/object-ingest
integration/object-publish
integration/vault-create
integration/vault-rewrap
integration/restart
integration/pending
integration/future
integration/multiprocess
integration/filesystem
integration/migration
```

CLI output may remain focused on normative vector categories.

Platform integration tests may be reported separately from the 388-vector conformance count.

Do not inflate the normative vector PASS count with OS-specific test cases.

---

# Part 23 — Fault injection requirements

Every injected fault must identify:

```text
operation
crash point
durable state before crash
expected durable state after restart
expected writer eligibility
expected visible/recovery state
```

Avoid vague tests such as "simulate crash and ensure okay."

The test must state which normative ordering rule it exercises.

---

# Part 24 — Do not change r36 casually

Do not edit:

```text
spec/totipo-vault-format-v0.md
```

for platform convenience.

If Phase 3 reveals a genuine contradiction, create/update:

```text
SPEC_ISSUES.md
```

with:

```text
issue ID
normative sections
platform/primitive involved
minimal reproducer
why the requirements cannot both be satisfied
whether this is protocol, conformance, or platform-adapter scope
proposed resolution
wire compatibility impact
```

A platform limitation is not automatically a spec contradiction.

---

# Part 25 — Required deliverables

Implement/document as much as possible of:

```text
reference/client/
reference/storage/
reference/localstate/
reference/integration/
```

or equivalent.

Deliver:

- minimal end-to-end reference client;
- Linux filesystem object-store adapter;
- real durable security-state adapter;
- crash injection framework;
- restart/recovery integration tests;
- real immutable-object publication tests;
- initial `VAULT` no-replace publication tests;
- same-root `VAULT` replacement tests;
- multi-process establishment tests;
- multi-process object publication tests;
- hostile path/symlink/special-file tests;
- out-of-order/missing/reappearing object tests;
- future-version end-to-end tests;
- writer freshness/linearization integration tests;
- migration integration test with real randomness;
- platform capability documentation;
- actual CI run evidence where available.

---

# Part 26 — Phase 3 report

Produce:

```text
PHASE3_IMPLEMENTATION_REPORT.md
```

The report must include:

## Baseline

- Phase 2 normative corpus count;
- proof that existing 388 cases remain unchanged/passing.

## Reference-client coverage

- object ingest pipeline;
- object author/publish pipeline;
- restart reconstruction.

## Persistence

- local durable-state technology chosen;
- transaction/durability semantics;
- crash-point count;
- restart cases;
- ordering properties verified.

## Filesystem

- OS tested;
- actual primitives used;
- symlink/special-file tests;
- immutable publication behavior;
- `VAULT` publication behavior;
- known platform limitations.

## Concurrency

- multi-process establishment race count/result;
- object publication race count/result;
- writer freshness race count/result.

## End-to-end protocol behavior

- pending object arrival/re-entry;
- future-version evidence;
- known-history regression;
- durable abandonment/bypass;
- presentation recovery;
- migration.

## CI

- actual OS/Go matrix runs completed;
- failures/unsupported cases.

## Remaining limits

Clearly state what is still not demonstrated.

Do not claim:

```text
production readiness
power-loss proof
cross-platform durability
full Totipo v0 client conformance
```

unless the evidence genuinely supports those statements.

---

# Part 27 — Recommended implementation order

Use this order:

1. preserve Phase 2 baseline;
2. storage/security-state interfaces;
3. real durable local-state adapter;
4. Linux object-store adapter;
5. end-to-end read/validation pipeline;
6. immutable object publication;
7. `VAULT` initial/replacement publication;
8. crash injection;
9. restart reconstruction;
10. pending/future end-to-end scenarios;
11. multi-process establishment races;
12. multi-process publication races;
13. hostile filesystem cases;
14. writer freshness/linearization;
15. migration with real CSPRNG;
16. remote CI matrix.

Do not start with cross-platform abstractions before Linux behavior is well tested and documented.

---

# Completion criterion

Phase 3 is complete when the repository can demonstrate, on at least one real supported filesystem/OS:

```text
frozen protocol bytes
+
strict crypto validation
+
semantic validation
+
durable local security memory
+
crash-safe publication
+
restart reconstruction
+
multi-process race handling
```

without changing r36 or relying on in-memory-only simulations for those claims.

Any remaining platform-specific gaps must be explicitly reported rather than represented as passing.
