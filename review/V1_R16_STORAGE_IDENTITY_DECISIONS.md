# Totipo v1/r16 storage and vault-identity decisions — third design checkpoint

> This document records storage, synchronization-boundary, local-history, and vault-recognition decisions for the future Totipo v1/r16 simplification. It is not normative and does not amend the current v1/r15 specification.

> Where this checkpoint resolves questions left open by the previous r16 checkpoints, this document records the current agreed direction for the eventual normative r16 rewrite.

The current normative specification remains **Totipo v1/r15** in
[`spec/totipo-vault-format-v1.md`](../spec/totipo-vault-format-v1.md).
MUST / MUST NOT / MAY language here records future design direction only.

Starting HEAD: `f2b51026ef3d436b992e7628017fad8f26c9917f`.
Before editing, `git status --short` was empty; `make check` and `make verify`
passed. Both previous r16 checkpoints existed, were committed, and matched
their HEAD bytes. No normative r16 rewrite had started.

## Relationship to previous checkpoints

The [first checkpoint](V1_R16_SIMPLIFICATION_DECISIONS.md) and
[second checkpoint](V1_R16_STATE_POLICY_DECISIONS.md) remain unchanged historical
records. Together they have settled:

- TOKEN-only semantic object family; removal of DEVICE and provenance signatures;
- simplified flattened TOKEN grammar and optional CLIENT_NAME / CLIENT_TIME;
- exact v1 grammar within `objects-v1/`;
- causal-equivalence/SCC cycle semantics;
- complete-state TOKENs, including tombstones;
- descriptive rather than prescriptive token-state policy;
- authorship despite unavailable history;
- fixed `MAX_PARENTS = 4` and fixed linear folding.

The second checkpoint left advisory / learned-history durability,
VAULT_BINDING / local authoritative state, platform/storage implementation,
the fixed 1024-byte envelope, exact same-OBJECT_ID defensive wording, and the
public API open. This checkpoint resolves the first two and substantially
clarifies the storage abstraction. It does not select an exact portable
filesystem implementation, decide whether `platform-linux` remains, settle
the envelope, simplify discovery/readiness, specify exact VAULT creation or
replacement rules, or define the public Java API.

## Settled: a durable object store, not a synchronization protocol

> Totipo v1 operates on a configured durable object store. Synchronization is optional and external to the v1 protocol.

The protocol/store layout is deliberately synchronization-friendly, but
synchronization is not required for correct operation. Deployments may include:

```text
single local machine with no synchronization
shared removable/USB storage
shared filesystem or network mount
directory used by multiple clients at different times
directory watched by Syncthing
directory watched by Dropbox/Drive-like software
manual copy/backup workflows
```

These are deployment examples; no specific synchronization product is part of
the protocol model. Immutable content-addressed objects, explicit parent
references, and no in-place semantic-object mutation behave well when files
are copied, replicated, delayed, delivered out of order, observed by multiple
clients, or transferred manually.

> v1 correctness MUST NOT depend on the presence of a synchronization service.

A perfectly valid v1 deployment may never synchronize anything.

### The configured store is protocol-visible state

Conceptually:

```text
configured Totipo store
    VAULT
    objects-v1/
```

This is the persistent store from which a client discovers currently available
object bytes and into which it durably publishes new immutable TOKEN objects.
The protocol requires no separate `local database`, `sync database`, or
`remote graph` distinction. A client may open the same configured store that
another client used previously. No per-client graph database is required for
correctness.

### External synchronization is outside v1 semantics

The future specification should neither define nor depend on:

- synchronization acknowledgements or remote propagation completion;
- remote peer count or whether every other client has observed an object;
- provider snapshots or provider-side historical versions;
- provider conflict-copy naming conventions or remote retention policy;
- Dropbox, Syncthing, or other providers' internal behavior.

Totipo sees files/objects present in its configured store. How those files
arrived is outside the protocol. Local durable publication is distinct from
any remote propagation acknowledgement.

### Ordinary arrival and discovery remain supported

Graph semantics naturally tolerate late and out-of-order object arrival,
temporary missing dependencies, concurrent independent assertions, duplicate
arrival of the same object, and repeated discovery of unchanged objects.
With arrows denoting child-to-parent references:

```text
snapshot 1:
A

snapshot 2:
A
B -> A
```

This is ordinary late arrival. The protocol need not know whether B arrived
through synchronization, manual copy, a shared drive, another local client,
or recovery from backup.

### Multiple clients can take turns

The following must remain coherent without transferring per-client history:

```text
Client A opens configured store
Client A publishes objects durably
Client A closes

Client B later opens the same store
Client B discovers those objects
Client B publishes new objects
```

The durable store itself carries the shared protocol-visible history.
Shared removable storage is a first-class sanity case:

```text
USB drive:
    VAULT
    objects-v1/
```

Two machines may take turns using this drive. No synchronization service
exists; the same v1 semantics apply. Future wording must preserve this case
and must not accidentally require a cloud synchronization provider.

## Settled: durable retention expectation and limits

> The configured durable store is expected, under ordinary successful operation, to retain immutable objects that were successfully durably published to it.

Totipo is not designed around successfully stored immutable objects randomly
disappearing during normal operation. Permanent disappearance may result from
storage failure, user deletion, rollback/restore, broken synchronization, or
external filesystem manipulation. Resulting missing information must be
handled gracefully; no separate security-state architecture is needed to make
such disappearance impossible or invisible.

> Totipo v1 does not provide cryptographic rollback resistance or deletion resistance against a configured durable store that loses or replaces previously stored objects.

If previously available objects disappear, do not crash or invent ancestry.
Truthfully report whatever graph/state remains observable, let unresolved
references remain unresolved, and allow new complete TOKEN assertions as
already settled. This does not promise recovery of lost objects. The exact
durable publication contract remains open below.

## Settled: remove protocol advisory history and a separate local cache

The advisory / learned-history durability question is now resolved:

- no named/normative advisory-history capability;
- no mandatory cross-run remembered graph or remembered-history database;
- no separate protocol local-object cache;
- the configured durable store is sufficient for correctness;
- implementation caching remains optional;
- no rollback/deletion-resistance promise against lost store objects.

Do not retain equivalents of advisory-history capability negotiation,
`HISTORY_REGRESSION`, `HISTORY_MEMORY_LOST`, remembered-head conformance
requirements, or cross-run advisory topology as a separate protocol layer.
Loss of previous local observations is not a protocol validity failure.
Applications may provide their own diagnostics without creating those
protocol states.

The future normative state model must not introduce a second object source:

```text
sync object store
UNION
protocol local cache
```

Do not define cache/source-merging semantics merely to defend against a
synchronization system deleting immutable objects. OS filesystem caches,
ordinary backups, implementation-local performance caches, recovery tools,
and local copies remain allowed, but are not protocol concepts.

### Implementation caching and object availability

An implementation MAY internally cache parsed objects, object bytes, indexes,
or derived graph data for performance. Such caches create no additional
protocol facts, must not silently invent objects absent from their validated
source material, remain implementation concerns, and may be discarded and
rebuilt. No persistent cache API is required by this checkpoint.

For the ordinary store-driven reader, an object is available when the
configured durable store can supply its valid bytes. A parent named by another
TOKEN whose valid bytes are absent remains unresolved. Do not introduce
`available from sync`, `available from cache`, or `available from local
security memory` states into the protocol model. A cache is not a second
normative source that silently restores missing store objects or ancestry.

This resolves the second checkpoint's durability question without making
previously validated facts false when bytes disappear. A writer may still
reference an unavailable known parent, and disappearance alone does not undo
an informed choice to assert a complete value. Neither principle requires
retaining those observations across runs or promoting remembered data into
currently available store evidence. A fresh client derives its observable
graph from the configured store and reports missing references truthfully.

## Settled: identical assertions intentionally collapse

After device provenance is removed, two independent clients in the same vault
that produce byte-identical canonical TOKEN plaintexts with the same TOKEN_ID,
parent set, complete TokenValue, CLIENT_NAME presence/value, and CLIENT_TIME
presence/value have:

```text
P1 == P2
OBJECT_ID(P1) == OBJECT_ID(P2)
```

They intentionally represent one content-addressed TOKEN object. v1 does not
preserve the multiplicity of otherwise indistinguishable authoring events.
The comparison is within the same vault/root and exact canonical grammar.

For example:

```text
current heads = {A, B}

Client 1 independently authors:
C = value X
parents = {A, B}

Client 2 independently authors exactly:
C = value X
parents = {A, B}
```

With identical optional metadata, both produce the same canonical object.
When their stores later converge, one semantic object C exists, not two
conflicting TOKEN assertions. This is desirable deduplication/idempotence.

### Provider duplicate names add no semantics

An external provider may produce duplicate/conflict filenames even for exact
identical bytes. v1 recognizes only its canonical object naming/layout rules.
Provider-generated alternate names do not create additional TOKEN assertions
merely by existing.

### Metadata differences can produce equal-valued concurrent assertions

If clients use the same TOKEN_ID, parents, and complete TokenValue, but differ
in CLIENT_NAME and/or CLIENT_TIME, their canonical plaintexts and OBJECT_IDs
differ. They may be distinct concurrent TOKEN heads C1 and C2, while:

```text
value(C1) == value(C2)
```

CLIENT_NAME and CLIENT_TIME do not participate in TokenValue equality. Current
semantic state can therefore remain unambiguous with multiple causal heads.
Client metadata is informational, not semantic state.

### Do not add EVENT_ID or an authoring nonce

Do not introduce a mandatory random event identifier solely to preserve two
indistinguishable authoring operations. Totipo is now fact/content oriented
rather than event-provenance oriented; identical-event multiplicity has no
semantic value. Mandatory nonces would manufacture unnecessary sibling
objects, while retries and identical authoring benefit from content collapse.
Future requirements must not assume that every UI action creates a distinct
object when its canonical assertion is identical.

## Settled: retire mandatory VAULT_BINDING establishment

The VAULT_BINDING / local authoritative-state question is now resolved:

- retire mandatory `VAULT_BINDING` local establishment from the future design;
- introduce `VAULT_FINGERPRINT` as stable recognition data derived from K_root;
- remembering/comparing an expected fingerprint is optional application configuration;
- mismatch is descriptive, not vault invalidity.

The old binding model pinned:

```text
configured local vault instance/location
    <-> specific K_root
```

The broader establishment design grew lifecycle concepts such as:

```text
UNESTABLISHED
PENDING(VAULT_BINDING)
ESTABLISHED(VAULT_BINDING)
```

That machinery is no longer desired. The future rewrite should remove
requirements whose sole purpose is mandatory durable local establishment:
`UNESTABLISHED` / `PENDING` / `ESTABLISHED`, mandatory `VaultBindingStore`,
binding-before-state-exposure, binding-before-use/authorship, binding mismatch
as protocol-invalid state, and binding journal/state-machine recovery.
These are concepts to retire or avoid, not a claim that all are current r15
identifiers or requirements.

Vault creation still requires its own durable publication correctness.
Removing establishment state does not weaken VAULT write durability; creation
and replacement details remain a separate open topic.

### VAULT_FINGERPRINT

Conceptually:

```text
VAULT_FINGERPRINT =
    HMAC-SHA-256(
        K_root,
        ASCII("totipo/v1/vault-fingerprint")
    )
```

The exact domain string may be finalized during the normative rewrite. The
name and semantic role are settled. The fingerprint is stable for one K_root,
unchanged by password rewrapping of that root, and different for independently
generated roots except with negligible cryptographic collision probability.
It is safe to retain/display as non-secret recognition data. It is neither
authorization, a write capability, nor protocol continuity state.

After successfully authenticating/unlocking canonical VAULT and recovering
K_root, the library can derive/expose VAULT_FINGERPRINT. An application can
say “this is vault fingerprint X” without storing K_root itself. The
fingerprint does not determine validity: successful cryptographic unlock
already established that.

### Expected fingerprint is application configuration

An application MAY remember the fingerprint normally associated with a store:

```text
configured store location
expected fingerprint?   // application-local
```

This is not mandatory protocol state. A brand-new client can open a vault
without a remembered fingerprint. The descriptive cases are:

```text
no expected fingerprint
    -> valid vault; no previous local association

expected == actual
    -> valid vault; same vault normally associated here

expected != actual
    -> valid vault; different vault than normally associated here
```

Mismatch signals a different vault to the application/user. It does not make
the opened vault cryptographically invalid or retroactively invalidate facts.
If X was expected but Y opens successfully, Y remains a valid independent
vault, valid TOKEN objects authored under Y remain Y's objects, and X's
objects remain facts of X. The location now contains a different vault than
expected; neither vault is thereby corrupt.

The application/user decides whether to cancel opening, open the different
vault, open once without changing configuration, or update/rebind local
configuration to the new fingerprint. The library should expose the actual
fingerprint and sufficient information to inform the user. Mismatch must not
become a mandatory authorship or TOTP prohibition in the semantic protocol.

### Multiple vaults are separate instances

Applications may support independent configured vault instances, normally
with distinct locations and roots:

```text
Personal:
    /vaults/personal/
    fingerprint X

Work:
    /vaults/work/
    fingerprint Y
```

These are not multiple identities within one graph. Each store/root
combination is a separate vault instance; application names such as
“Personal” and “Work” are outside v1.

### Recognition is neither first-contact authentication nor rollback protection

Without a remembered fingerprint, a new client cannot distinguish the
intended first vault from another valid vault supplied at first contact.
VAULT_FINGERPRINT recognizes prior local association; it does not authenticate
first contact.

Password/bootstrap representations can change while preserving K_root and
therefore the fingerprint. Fingerprint equality establishes neither bootstrap
freshness, the latest password wrapper, the latest object history, nor rollback
resistance. Do not reintroduce rollback claims through the new name.

## Settled: application configuration is not protocol-authoritative history

Ordinary v1 use should require no mandatory local security-memory database
in addition to the configured durable store and password/key material supplied
for opening. Applications may keep a store path, human label, expected
VAULT_FINGERPRINT, and UI preferences as application state.

If configuration containing the expected fingerprint disappears, the vault
remains valid and a new client can still unlock it. The client simply lacks
prior recognition information; this is not protocol continuity corruption.
Keep the durability domains distinct:

```text
protocol durability:
    VAULT publication/replacement
    immutable TOKEN object publication

application configuration:
    expected fingerprint
    friendly vault name
    remembered path
```

Loss of application configuration cannot retroactively invalidate successfully
published protocol objects.

## Newly resolved items

| Previously open / discussed item | Decision now |
| --- | --- |
| Advisory-history capability | Remove from protocol |
| Mandatory cross-run remembered graph | Remove |
| Protocol local object cache | Remove |
| Sync-provider behavior | Outside v1 |
| Configured store role | Durable source of protocol-visible object bytes |
| Identical independent assertions | Collapse intentionally to one content-addressed object |
| Mandatory event nonce | Do not add |
| VAULT_BINDING | Retire mandatory establishment model |
| Vault recognition | Introduce `VAULT_FINGERPRINT` |
| Fingerprint mismatch | Descriptive “different vault” signal, not invalidity |

## Open questions before normative r16

### A. Discovery/readiness/resource completeness

Still open: whether `READY`, `PROCESSING_INCOMPLETE`, or authoritative-readiness
gates remain; whether incomplete discovery is only descriptive; and how
resource exhaustion is described after the move to descriptive rather than
prescriptive protocol state.

### B. Durable object-store publication contract

Still open: exact platform-neutral publication guarantees, absent/new versus
exact-existing behavior, ambiguous durability, atomic installation/replacement
requirements, and filesystem/platform implementation obligations. The durable
store abstraction does not settle these mechanisms.

### C. VAULT creation/replacement details

Still open: initial creation with orphan `objects-v1/` entries already present,
no-replace semantics, password rewrap/replacement behavior, crash handling,
and whether any preflight emptiness requirement remains.

### D. Platform-specific Java implementation

Still open: whether `platform-linux` remains useful, whether portable NIO is
sufficient, and what belongs in protocol conformance versus implementation
hardening. No exact portable filesystem implementation is selected here.

### E. Fixed 1024-byte envelope

Still open. This checkpoint makes no envelope-size decision.

### F. Same-OBJECT_ID exceptional wording

The high-level direction remains defensive integrity failure for different
valid canonical plaintexts under one OBJECT_ID. Exact r16 wording may be
finalized during the normative rewrite. Intentional identical-content collapse
is not that exceptional condition.

### G. Public Java/session API

Still open, including exact exposure of recognition and session information.
Public Java API design is not a blocker for protocol-design completion.

## Items that need not block r16

Unless another decision exposes a protocol dependency, garbage collection,
future-family migration/coexistence beyond the already-settled v1 boundary,
application vault-list UI, backup tooling, repair tooling, and public Java
application API details do not necessarily need pre-r16 design resolution.

## Baseline terminology and scope notes

“TokenValue” is the complete semantic `TOKEN_VALUE` of r15, including STATUS.
CLIENT_NAME / CLIENT_TIME and VAULT_FINGERPRINT are future design terminology.
The store-centric vocabulary replaces the emphasis on synchronized storage
without modifying current r15 wording.

Current r15 Section 34.1 explicitly calls `advisory-history` a conformance
capability, not wire state or protocol negotiation. Baseline r15 already
requires no cross-run graph history; this checkpoint removes the named
optional capability as well. Section 34.2 already prohibits cache-only
evidence from becoming current synchronized objects, values, or ancestry.
The removed concepts above must not be read as claiming r15 mandates a
remembered graph or a store/cache union.

Current r15 Section 10 requires durable VAULT_BINDING before ordinary use or
authorship and rejects automatic open on mismatch. It already requires no
journal or pending establishment record. The lifecycle labels above describe
older/broader machinery to avoid reintroducing, not current r15 requirements.
The new fingerprint also differs from r15's optional
`SHA-256(exact 87-byte VAULT representation)` audit fingerprint: that digest
tracks representation changes, whereas VAULT_FINGERPRINT recognizes K_root
across password rewrapping. Neither establishes rollback protection.

Current r15 Section 50.2 treats immutable-object crash durability as a
reliability recommendation, while Section 50.3 requires crash-safe VAULT,
binding, and private-key custody. This checkpoint's durable-store expectation
is future direction, not a claim that r15 already specifies the eventual r16
publication contract.

Only this non-normative review artifact is added. Both earlier checkpoints
remain byte-unchanged. No specification, vectors, schemas, requirements
profile, conformance implementation, generators, generated fixtures, revision
constants, README revision, or release/tag metadata changes. Normative r16
rewriting and corpus changes remain subsequent work.
