# Totipo v1/r16 storage-operation decisions — fourth design checkpoint

> This is a non-normative design checkpoint for the future Totipo v1/r16 simplification. It does not amend v1/r15 or begin the normative r16 rewrite.

The current normative specification remains **Totipo v1/r15** in
[`spec/totipo-vault-format-v1.md`](../spec/totipo-vault-format-v1.md).
The first three r16 checkpoints remain unchanged historical records. Where
this checkpoint resolves questions left open by the third checkpoint, this
later checkpoint records the current agreed direction for the eventual
normative r16 rewrite. MUST / MUST NOT / SHOULD / MAY language here records
future design requirements only.

Starting HEAD: `bf8f499a8ebd8fa28f8472e938db3701ac64200e`.
Before editing, `git status --short` was empty; `make check` and `make verify`
passed. All three earlier checkpoints were present. Their SHA-256 hashes are
recorded in the preservation section below.

## Relationship to previous checkpoints

The [first checkpoint](V1_R16_SIMPLIFICATION_DECISIONS.md),
[second checkpoint](V1_R16_STATE_POLICY_DECISIONS.md), and
[third checkpoint](V1_R16_STORAGE_IDENTITY_DECISIONS.md) have already settled:

- TOKEN-only semantic objects; removal of DEVICE and provenance signatures;
- flattened complete-state TOKENs, including tombstones;
- optional informational `CLIENT_NAME` / `CLIENT_TIME`;
- exact v1 TOKEN grammar within `objects-v1/`;
- causal-equivalence/SCC cycle semantics;
- fixed `MAX_PARENTS = 4` and fixed linear folding;
- descriptive rather than prescriptive state/use policy;
- authorship despite unavailable history;
- removal of protocol advisory history and a separate local object cache;
- the configured durable store as the source of protocol-visible object bytes;
- synchronization as optional and external to v1;
- retirement of mandatory `VAULT_BINDING`;
- `VAULT_FINGERPRINT` as optional recognition information.

This fourth checkpoint resolves the third checkpoint's open questions:

```text
A. Discovery/readiness/resource completeness
B. Durable object-store publication contract
C. VAULT creation/replacement details
```

It does not choose portable Java NIO versus platform-specific storage, decide
whether `fs-linux` / `platform-linux` remains, settle the fixed 1024-byte
envelope, define the public Java/session API, decide garbage collection, or
extend future-family migration/coexistence beyond already-settled boundaries.
Exact same-OBJECT_ID defensive wording remains for the normative rewrite;
this checkpoint adds no machinery for that exceptional condition.

## Settled: remove protocol discovery completeness/readiness

Future r16 has no concept of authoritative/resource-complete discovery as a
prerequisite for protocol operation. Remove `READY`, `PROCESSING_INCOMPLETE`,
authoritative readiness, resource-complete snapshots, and whole-store
readiness gates as protocol permission/readiness concepts.

The configured durable store need not be exhaustively enumerated before a
client may:

- interpret valid observed TOKEN objects;
- compute token state from the object set available to an evaluation;
- compute a TOTP from a complete known TokenValue;
- author a new complete TOKEN assertion;
- resolve a known conflict;
- publish a TOKEN;
- follow or retain known parent IDs.

A discovery/evaluation operation conceptually produces:

```text
valid observed TOKEN objects
unresolved parent references
diagnostics
```

Diagnostics may describe an unreadable candidate, wrong-sized candidate,
failed authentication, invalid canonical TOKEN grammar, directory/enumeration
failure, processing/resource limit reached, or an object that disappeared
during observation. These are not vault-wide semantic permission states.
Invalid, unavailable, or unprocessed candidates do not invalidate already
validated observed TOKEN objects. Resource exhaustion or incomplete traversal
may limit what an implementation can describe; it does not make the vault
globally unusable. This does not relax object validation rules.

### No protocol need for exhaustive absence claims

v1/r16 does not require protocol claims such as:

```text
this is every TOKEN in the store
no other TOKEN_ID exists
this TOKEN definitely has no unseen concurrent branch
this observation exhaustively represents all files present
```

The protocol therefore needs no resource-completeness predicate merely to
support those claims. Applications may implement exhaustive inventory
operations when useful. Their completeness is application/diagnostic behavior,
not Totipo semantic state.

### State is relative to evaluated valid objects

Graph/state functions operate on the valid TOKEN objects supplied to the
evaluation. For objects of the same TOKEN_ID:

```text
O1 = {A}
heads(O1) = {A}

later:

O2 = {A, B}
B lists A as parent
heads(O2) = {B}
```

The earlier result was not invalid; later information causes ordinary
recomputation. Late arrival, disappearance, unreadability, and reappearance
remain ordinary observation changes. They create no new durable freshness,
continuity, or readiness state. Descriptions must remain truthful about the
evaluated evidence and unresolved parents; unavailable objects must not be
invented or silently supplied by a second protocol cache.

## Settled: immutable TOKEN publication contract

The contract is platform-neutral. Protocol-required semantic outcomes are
distinct from possible platform implementation mechanisms. r16 must not
require `renameat2`, `O_TMPFILE`, hard links, POSIX-specific no-replace
operations, Windows-specific replace operations, a particular Java NIO call,
or an atomic compare-and-swap filesystem primitive. Those choices belong to
implementation/platform work.

### Scope

Publication addresses exactly the canonical immutable TOKEN object:

```text
objects-v1/<lowercase OBJECT_ID>
```

It requires neither complete discovery nor proof that the whole store was
processed. It does not depend on all parents being currently available,
synchronization or remote peer acknowledgement, provider propagation
completion, or durable local graph/journal insertion afterward. The configured
durable store object itself is the persistent protocol-visible fact.

### Complete staging before canonical publication

A writer MUST construct the complete intended object representation away
from the final canonical pathname. The future normative requirement must
express the outcome:

- do not intentionally expose a partially constructed canonical TOKEN object;
- use the strongest reasonable crash-safe publication mechanism available to
  the implementation/store;
- attempt the persistence operations needed by that implementation before
  reporting a new publication successful.

The implementation MUST NOT deliberately use a publication strategy whose
normal operation exposes an incomplete canonical TOKEN and then progressively
fills, truncates, or modifies that canonical file. Creating the canonical
target and writing it piece by piece is not the intended publication model;
construct complete bytes separately and use the strongest reasonable
installation mechanism available.

Not every supported store supplies atomic installation. The protocol does
not require proof that no observer can ever see an intermediate storage state
when the store cannot supply stronger publication semantics. Atomicity must
not become a protocol conformance prerequisite when the underlying store lacks
such a primitive. Attempt the required durability operations and truthfully
report success/failure; lack of atomic visibility creates no new protocol
failure/readiness state. An unexpectedly observed partial, wrong-sized, or
invalid canonical object remains ordinary invalid/unavailable diagnostic
evidence under the discovery model above.

### New-object success

A newly published TOKEN requires a good-faith crash-safe durability attempt
before publication acknowledgement. Conceptually, an implementation may
perform operations corresponding to:

```text
construct/stage complete bytes
persist staged/file contents
install/move into canonical location
persist containing namespace/directory
report success
```

This is required durability intent/outcome, not mandated exact syscalls or
ordering primitives. A successful result means:

> The implementation believes it successfully stored the complete immutable object using the durability facilities appropriate to the configured store and completed the persistence work it relies on before acknowledging success.

Success does not promise perpetual retention, deletion resistance, rollback
resistance, protection from later external replacement, hardware that tells
the truth about flushes, remote synchronization completion, or survival of
every conceivable storage failure.

### Existing exact object is idempotent success

If the canonical object already contains the exact intended object bytes,
publication is already satisfied. Return an idempotent success, conceptually:

```text
ALREADY_PRESENT_EXACT
```

The implementation MUST NOT rewrite, replace, rename, chmod, touch, or
otherwise mutate the existing object merely to reconfirm publication. No
fresh durability barrier or filesystem rewrite is required for exact-existing
success. A read-only exact-byte comparison is sufficient; the storage
publication layer need not decrypt, parse, or semantically revalidate those
bytes merely to return this result.

Identical assertions intentionally collapse to one content-addressed object.
Retries after ambiguous failure should be harmless. Unnecessary mutation can
interact badly with external synchronization systems and cause needless
conflict/version activity.

### Existing non-exact target

If the canonical pathname exists but does not contain the intended exact
bytes, ordinary publication MUST NOT intentionally overwrite, repair in
place, delete, or silently replace it. Report publication failure with no
success acknowledgement. This includes unsuitable existing targets and
content that cannot be established as exact.

The publication layer need not decide whether existing bytes represent
corruption, partial storage, an impossible cryptographic same-ID contradiction,
or external manipulation. Diagnostics may describe those conditions elsewhere.
Publication simply leaves the existing canonical entry untouched.

### Ambiguous failure

An error after some or all storage effects may have occurred means:

```text
success is not established
```

It does not establish that the object is absent. The intended object may
remain present. Do not create persistent protocol states such as
`PUBLICATION_UNKNOWN`, `PUBLICATION_PENDING`, or `RECOVERY_REQUIRED`.
Ordinary later observation determines what is available. Retrying the exact
same publication is safe:

```text
target absent
    -> try normal publication

target exact
    -> idempotent success

target different/unsuitable
    -> failure, leave untouched
```

### No mandatory reread/self-validation

The writer need not reopen, decrypt, parse, and semantically revalidate its
own newly written TOKEN before reporting publication success. Canonical writer
correctness is tested/conformance-covered separately. The publication backend
attempts to store the exact bytes supplied and truthfully reports whether it
believes the required storage work succeeded. Applications/clients will
ordinarily refresh/discover afterward.

### No parent publication dependency

A child TOKEN may be published even if a listed parent is not currently
available from the configured durable store. For clients lacking that object,
the edge remains unresolved; the parent's absence does not invalidate the
child. Known parent IDs may still be retained and referenced.

Fold stages remain ordinary complete TOKEN objects. There is no multi-object
transaction and no rollback of already published earlier stages if a later
stage fails.

### Namespace creation

If `objects-v1/` is absent, an implementation MAY create it as needed. Prior
existence is not a semantic requirement. Directory-creation durability is an
implementation/storage concern subject to the same good-faith durability
principles.

## Settled: VAULT initial creation

### Canonical absence is sufficient to attempt creation

Remove any requirement to prove `objects-v1/` empty before creating a new
canonical `VAULT`. Conceptually:

```text
canonical VAULT absent
    -> creation may be attempted
```

No exhaustive discovery or proof that old/orphan TOKEN-looking files are
absent is required. r16 has no protocol resource-completeness semantics.
Old objects encrypted under another root do not authenticate under a newly
generated independent root. Object-looking remnants may come from backup,
previous storage, failed synchronization, manual copying, or another vault.
Creating a new VAULT does not itself delete those files.

An application MAY warn about existing object-looking entries it happens to
observe. That is diagnostic/UI policy, not a protocol creation gate.

### Do not knowingly replace an existing canonical VAULT

Initial creation MUST NOT intentionally overwrite an already present
canonical `VAULT`. This is the desired semantic outcome, not a mandate for
an atomic no-replace primitive. Implementations should use the strongest
reasonable mechanism available. If an implementation cannot establish that
creation succeeded without knowingly replacing an observed existing canonical
VAULT, it MUST NOT report successful initial creation.

### Creation completion

Creation conceptually consists of:

```text
generate new K_root
construct complete canonical VAULT representation
attempt crash-safe durable publication
report success only when publication is believed successful
```

There is no subsequent mandatory `VAULT_BINDING` establishment, local
security-memory commit, binding journal, or `ESTABLISHED` state transition.
The already-settled `VAULT_FINGERPRINT` may be derived/exposed after
unlock/creation. Remembering an expected fingerprint is application
configuration; failure to persist it does not make protocol VAULT creation
fail.

## Settled: VAULT replacement / password change

Password change rewraps the same root:

```text
old password -> old VAULT representation -> K_root X
new password -> new VAULT representation -> same K_root X
```

A replacement intentionally changing `K_root` represents a different vault,
not an ordinary password change. `VAULT_FINGERPRINT` remains stable across
password rewrapping.

### Complete replacement representation

Construct the complete new VAULT representation separately. Do not rely on
in-place truncate-and-rewrite as the required protocol strategy.
Implementations should use their strongest reasonable crash-safe replacement
mechanism and attempt appropriate persistence of the replacement and containing
namespace before reporting success. Specify desired durable behavior without
mandating an atomic replace primitive or particular platform syscalls.

### Compare-before-replace

A replacement operation is based on the exact canonical VAULT representation
that it successfully opened:

```text
BASE = exact canonical VAULT bytes opened by the operation
```

A VAULT replacement operation MUST re-observe the canonical `VAULT` immediately
before its replacement attempt, when that canonical representation can be
observed, and compare the exact currently observed bytes with the exact `BASE`
representation on which the operation was based. If the canonical VAULT cannot
be read/observed sufficiently to perform this comparison, success MUST NOT be
reported for that replacement attempt.

This requires local observation and comparison, not atomic filesystem
compare-and-swap:

```text
observe current canonical bytes
compare to BASE
if different, stop
otherwise attempt replacement
```

If an intervening change is observed:

```text
CURRENT != BASE
```

the implementation MUST NOT knowingly overwrite that observed intervening
representation as part of the stale operation and MUST report the
password-change/replacement attempt as stale or unsuccessful. It may require
the application/user to reopen and retry as appropriate.

**Concurrent password rewrap.** A and B both open `V1 -> K_root X`. B first
rewraps X, replacing V1 with V2. A later prepares `V3 -> K_root X`. Blindly
replacing V2 with V3 would silently lose B's newer password change. Comparison
instead gives:

```text
A's BASE = V1
current canonical = V2

V2 != V1
    -> A does not knowingly replace
```

**Configured location changed to a different vault.** A opens
`V1 -> K_root X`. Later the canonical location contains `V2 -> K_root Y`.
A's stale operation has prepared `V3 -> K_root X`. Blind replacement would
switch the location back from vault Y to vault X. Exact base comparison
detects the observable change and stops the stale write.

### Exact bytes, not VAULT_FINGERPRINT

Do not use `VAULT_FINGERPRINT` for freshness comparison. Different password
wrappers for the same `K_root` intentionally share the same fingerprint.
Comparison must concern the exact canonical representation on which the
operation was based. Do not introduce a VAULT generation counter or revision
number merely for this purpose, or a second normative representation digest
without independent justification. Exact bytes are sufficient.

### Not an atomic/distributed CAS requirement

The compare-before-replace rule means:

> Do not knowingly overwrite an intervening canonical VAULT change that the configured store currently exposes.

It does not require proving that no other client or synchronization peer has
concurrently produced another representation anywhere. A race may still occur
between comparison and replacement if the platform lacks stronger conditional
primitives; that residual race is accepted by the protocol design. The
comparison establishes neither distributed serialization nor consensus, and
does not require proof of a globally race-free current VAULT. Remote/provider
state not currently visible in the configured store remains outside v1.

Implementations MAY use locks, native conditional replacement, atomic
compare/swap-like mechanisms, or platform-specific filesystem primitives.
r16 must not require those stronger facilities for protocol conformance.

### Ambiguous replacement failure

If an error occurs after the canonical representation may have changed, do
not report success, claim the old representation definitely remains, or claim
the new representation definitely won. Create no persistent replacement-pending
protocol state. The next ordinary open inspects the canonical `VAULT` actually
present. Crash/recovery behavior is observation-driven, without a required
journal or recovery state machine.

### Password rewrap limitations

Changing the canonical password wrapper does not cryptographically revoke
historical copies of old VAULT representations. A retained old valid wrapper
and its old password may still recover the same `K_root`. Password replacement
does not provide historical-wrapper revocation, rollback protection, deletion
resistance, or synchronization-provider version-history erasure.

This is consistent with `VAULT_FINGERPRINT`: representations may change while
the root/vault identity remains the same.

### Canonical pathname only

Only the canonical `VAULT` pathname is automatically interpreted as the active
vault bootstrap. Temporary files, provider conflict copies, backups, and
alternate names do not automatically compete as protocol-authoritative VAULT
representations. Applications or recovery tooling may surface them as
diagnostics/recovery material. No provider-specific conflict naming behavior
is defined here.

## Explicitly resolved items

| Previously open item | Decision |
| --- | --- |
| Resource-complete discovery | Remove as protocol concept |
| `READY` / `PROCESSING_INCOMPLETE` | Remove as protocol permission/readiness states |
| Incomplete/resource-limited discovery | Diagnostics/unresolved evidence only; validated objects remain usable |
| Whole-store prerequisite for authorship/use | None |
| TOKEN publication and parent availability | Parents need not be currently available |
| New TOKEN publication durability | Good-faith crash-safe durable publication before success |
| Mandatory atomic filesystem install | Do not require at protocol level |
| Existing exact TOKEN object | Idempotent success; no mutation/re-fsync solely for acknowledgement |
| Existing non-exact canonical TOKEN target | Leave untouched; publication fails |
| Ambiguous TOKEN publication | No success acknowledgement; no persistent unknown state |
| Mandatory writer reread/self-validation | Remove / do not require |
| `objects-v1/` lazy creation | Permitted |
| VAULT creation preflight emptiness | Remove |
| Existing orphan TOKEN files during VAULT creation | Do not block creation |
| Initial creation over existing canonical VAULT | Do not knowingly replace |
| Post-creation VAULT_BINDING step | None |
| Password change | Rewrap same `K_root` |
| VAULT replacement freshness | Must re-observe and compare current exact bytes to the exact opened representation immediately before replacement |
| Observable intervening VAULT change | Must not knowingly overwrite; stale/failure |
| Atomic CAS/locking | Optional stronger implementation behavior; not required |
| Ambiguous VAULT replacement | No success acknowledgement; recover by ordinary reopen |
| Old password wrappers | May remain valid if retained |

## Remaining open items after this checkpoint

The main remaining protocol-design item before normative r16 is the
**fixed 1024-byte encrypted object envelope**. This checkpoint does not settle
that question.

Exact same-OBJECT_ID exceptional wording still needs final normative phrasing.
Its high-level defensive integrity-failure behavior is already settled and
does not require a new state architecture. Exact-existing publication is
byte comparison, not a new definition of that exceptional semantic condition.

Leave the following for post-r16 implementation work unless the envelope
discussion exposes a protocol dependency:

- portable Java NIO versus platform-specific storage;
- whether `fs-linux` / `platform-linux` remains;
- exact Java public/session API;
- implementation hardening choices;
- platform syscall selection.

No implementation architecture is selected. Garbage collection and
future-family migration/coexistence beyond the already-settled v1 boundary
remain outside this checkpoint's decisions.

## Baseline/current-r15 context

These future decisions must not be read as claims that every retired concept
is currently required by r15. Current r15 Section 24 describes
`PROCESSING_INCOMPLETE`, keeps successfully accepted objects usable, and already
states that v1 does not require a vault-wide operation gate. r16's direction
goes further by removing protocol discovery completeness/readiness states.

Current r15 Sections 9 and 50 describe concrete crash-safe VAULT publication
patterns, including atomic installation/replacement terminology and a fallback
when the platform lacks atomic semantics. Section 50.2 treats new immutable
object crash durability as a reliability recommendation and already permits
exact-existing acknowledgement without fresh durability barriers. This
checkpoint records the future platform-neutral durability contract; it does
not silently replace current r15 requirements or claim they already match it.

Current r15 Section 10 still requires durable VAULT_BINDING before ordinary
use/authorship. Its removal and the optional recognition role of
VAULT_FINGERPRINT were settled in the third checkpoint. “TokenValue” here is
the complete semantic value named `TOKEN_VALUE` in r15, including STATUS.

## Preservation and validation record

The three earlier checkpoints' baseline SHA-256 hashes are:

```text
0e4811836c2a9f6fc0fc590093dee94554b68cb965b5f66639da882a47a52940  review/V1_R16_SIMPLIFICATION_DECISIONS.md
99dffeeda7e41287b994c58519b1d67d20373e0dffa44df671487ede4cfac077  review/V1_R16_STATE_POLICY_DECISIONS.md
7755c3d912b63a2c251b27bc550e6ee89386294072236e424274f76a647285a0  review/V1_R16_STORAGE_IDENTITY_DECISIONS.md
```

No decision ambiguity was found while translating the agreed direction.
The deliberate limit of compare-before-replace is observation of exact bytes,
not a guarantee against races or unseen remote changes. No additional
platform, freshness, or recovery architecture is inferred from that rule.

After adding this checkpoint, `make check`, `make verify`, and
`git diff --check` passed. A separate no-index whitespace check covered this
untracked new file. All three hashes above were rechecked and matched;
the earlier checkpoints remain byte-identical. All tracked files matched
HEAD, confirming no changes to the normative specification, vectors or
generated cases, schemas, manifests, requirements profiles, conformance
implementation, generators, revision constants, README current-revision
declarations, release/tag metadata, or historical reports.

Final changed-path list (addition, left untracked and uncommitted):

```text
A review/V1_R16_STORAGE_OPERATION_DECISIONS.md
```

Normative r16 rewriting and corpus changes remain subsequent work; current
v1/r15 remains unchanged. Work is left uncommitted for manual review.
