# Totipo v1/r16 simplification decisions — design checkpoint

> This document records agreed design direction for the next Totipo v1 simplification revision. It is not itself normative, does not change r15, and intentionally leaves several decisions open before the normative r16 rewrite and corpus regeneration.

The current normative specification remains **Totipo v1/r15** in
[`spec/totipo-vault-format-v1.md`](../spec/totipo-vault-format-v1.md).
The decisions below are settled except where explicitly identified as open.
MUST / MUST NOT language here records the agreed requirements for the future
rewrite; it does not amend current r15 requirements.

Starting HEAD: `5ff6977ae004a9c03a410b38f0f9d73ca2bc76d4`.
Before editing, `git status --short` was empty, and `make check` and `make verify`
passed. This checkpoint uses the local specification and agreed design direction;
no other implementation was fetched or used as a semantic oracle.

## Context and principles

Implementation of r15 through TOKEN and DEVICE wide-frontier authorship exposed
significant complexity whose security or semantic value no longer justifies it.
This is deliberate pre-RC simplification, not a finding that r15 is defective or
insecure. The direction rests on these principles:

- Totipo is a single-vault-key system.
- Possession of the unlocked vault root is the actual authority to author vault state.
- Per-device provenance keys do not create meaningful authorization because a root holder can create arbitrary device keys.
- Device and time information are useful historical metadata, not security identities.
- TOKEN heads should remain complete state snapshots.
- v1 does not need in-family forward compatibility with hypothetical future semantic versions.
- Future protocol families should define their own compatibility relationship with v1.

## Settled: remove the DEVICE object family

The v1 `DEVICE` semantic object family will be removed, including:

- DEVICE objects and DEVICE_ID;
- DEVICE public-key advertisement;
- DEVICE presentation graph and current heads;
- DEVICE display-name revisions and presentation conflict semantics;
- DEVICE rename and wide-frontier folding;
- current opaque DEVICE handling;
- first DEVICE advertisement;
- the first-TOKEN DEVICE-publication prerequisite.

There is no replacement synchronized DEVICE identity object. The former goal of
DEVICE was to tell the user approximately where a TOKEN change occurred. That
information is metadata about the TOKEN update itself, not a security principal.

## Settled: remove per-device provenance signatures

TOKEN objects will no longer carry `AUTHOR_DEVICE_ID` or `SIGNATURE`, and there
will be no DEVICE signing key. The future r16 design removes:

- P-256 device identity and DEVICE_ID derivation;
- ECDSA signing and verification;
- the DER signature profile;
- signature-context input construction;
- provenance key discovery and provenance-key collisions;
- provenance statuses `VERIFIED`, `UNRESOLVED`, and `REJECTED`;
- late provenance reclassification;
- device private-key custody and device-key binding to the local vault.

The vault root is the actual authorization boundary. A holder of that root can
generate arbitrary device keys and identities, so a device signature does not
establish a security property strong enough to justify the machinery. The
encrypted/authenticated Totipo object layer already provides integrity and
authenticity with respect to possession of the vault root.

## Settled: remove the signature-context root derivation

The root-key branch used solely for semantic provenance signatures will
disappear. The r15 `K_signature_context` has no purpose once semantic signatures
are removed. The keyed object-ID and per-object encryption derivations remain;
this checkpoint proposes no other changes to that key hierarchy.

## Settled: optional TOKEN client metadata

The old author/time attribution is replaced with:

```text
CLIENT_NAME?
CLIENT_TIME?
```

These describe the client that produced that individual TOKEN object and are
**informational metadata only**. They MUST NOT affect:

- TokenValue equality;
- causal ordering;
- parent selection or head selection;
- conflict detection or conflict resolution;
- authorization;
- validity of the semantic TokenValue;
- TOTP calculation.

Clients may prefill these automatically without asking the user and may omit
either or both. Both fields may be inaccurate.

### CLIENT_NAME

`CLIENT_NAME` is an optional root-level TLV. Absence means “no client name was
supplied.” If present, its length is **0..128 UTF-8 bytes**. A present zero-length
CLIENT_NAME is valid and distinct on the wire from absence.

The existing strict UTF-8 philosophy applies: byte-counted, with no normalization,
trimming, case folding, or locale transformation. This is presentation metadata.

### CLIENT_TIME

`CLIENT_TIME` is an optional root-level TLV. If present, it contains exactly
**8 bytes**, interpreted as an unsigned big-endian integer of client-supplied
Unix seconds. Absence means “no client time was supplied.” Because absence
represents unknown, zero is an ordinary encoded numeric value, not an unknown
sentinel.

CLIENT_TIME is informational and MUST NOT establish freshness, ordering,
causality, conflict priority, or winner selection.

### Remove AUTHOR_TIME

The r15 `AUTHOR_TIME` concept disappears. Its useful historical/UI role is
replaced by optional CLIENT_TIME. No special common AUTHOR_TIME rules for folds
are retained. If multiple fold objects represent one logical authoring operation,
their client metadata policy will be defined in the r16 fold rewrite; time has
no security or ordering meaning.

## Settled: TOKEN remains complete state and STATUS stays

> Every supported TOKEN object carries a complete TokenValue.

This applies to both live and deleted state. Tombstones are not reduced to
status-only events. TOKEN state remains conceptually:

```text
STATUS
ISSUER
ACCOUNT
ALGORITHM
DIGITS
PERIOD
SECRET_BYTES
```

A deleted TOKEN continues carrying the full issuer/account/TOTP credential state:

- current state can be evaluated from heads without climbing history;
- deleted-token presentation does not require ancestry reconstruction;
- restore can act on the complete prior state;
- concurrent tombstones over different complete states remain distinguishable;
- folding remains whole-state.

The existing lifecycle concept represented by STATUS stays. The current r15 enum
names `LIVE` / `TOMBSTONE` may remain until the normative rewrite. This checkpoint
does not invent a new deletion representation. “TokenValue” here denotes the
complete semantic value called `TOKEN_VALUE` in the r15 specification.

## Settled: flatten credential fields

The nested `CREDENTIAL` wire structure is removed. `ALGORITHM`, `DIGITS`,
`PERIOD`, and `SECRET_BYTES` become root-level TOKEN TLVs. The logical
implementation/domain model may still group them as a `Credential` object if
useful; the wire format need not mirror that object model.

The entire TOKEN is already an atomic complete-state snapshot. There is no
field-level merge, so nested wire framing provides no semantic benefit.

## Settled: TOKEN conceptual order and explicit parents

The agreed root-level conceptual order is:

```text
TOKEN_ID

PARENT_COUNT
PARENT_ID...

STATUS
ISSUER
ACCOUNT
ALGORITHM
DIGITS
PERIOD
SECRET_BYTES

CLIENT_NAME?
CLIENT_TIME?
```

This reads as logical identity, causality, complete token state, and optional
event/client metadata. The normative r16 rewrite will assign/finalize exact TLV
tags. No numeric tags are assigned by this checkpoint.

Parents are not collapsed into one concatenated TLV. Retain the explicit model:

```text
PARENT_COUNT
PARENT_ID
PARENT_ID
...
```

Cardinality is explicit, each parent remains independently framed, parser
behavior is straightforward, and the representation is easy to audit. Saving
four bytes per parent is not worth changing this property. Parent IDs remain
canonical, sorted, and duplicate-free.

## Settled: field limits and guaranteed parent capacity

Retain current r15 limits:

```text
ISSUER        0..256 UTF-8 bytes
ACCOUNT       0..256 UTF-8 bytes
SECRET_BYTES  1..128 bytes
```

Issuer/account limits are not tightened to optimize uncommon wide-frontier
cases. CLIENT_NAME has its separate new 128-byte maximum.

For the proposed flattened grammar, explicit parent encoding, and the retained
1006-byte semantic payload maximum, fixed/mandatory overhead is:

```text
TOKEN_ID          4 + 32 = 36
PARENT_COUNT      4 +  2 =  6
STATUS            4 +  1 =  5
ISSUER header              4
ACCOUNT header             4
ALGORITHM         4 +  1 =  5
DIGITS            4 +  1 =  5
PERIOD            4 +  4 =  8
SECRET header              4
--------------------------------
fixed                     77
```

Each explicit parent costs:

```text
4-byte TLV header + 32-byte ID = 36 bytes
```

Worst-case legal TOKEN fields with both optional client metadata fields, before
adding parents:

```text
77 fixed
+ 256 ISSUER
+ 256 ACCOUNT
+ 128 SECRET_BYTES
+ (4 + 128) CLIENT_NAME
+ (4 + 8) CLIENT_TIME
= 861 bytes
```

Remaining capacity and four-parent cost:

```text
1006 - 861 = 145 bytes
4 * 36 = 144 bytes
```

> Every legal proposed r16 TOKEN is guaranteed to fit at least four explicit parents, even when ISSUER, ACCOUNT, SECRET_BYTES, CLIENT_NAME and CLIENT_TIME are simultaneously at their maximum sizes.

This arithmetic is settled as a design constraint. Whether r16 limits every
TOKEN to four parents or permits more when they fit remains **open**.

## Settled: remove OBJECT_TYPE and OBJECT_VERSION

`objects-v1/` will contain only one semantic object grammar: TOKEN. A serialized
`OBJECT_TYPE = TOKEN` field carries no information and will be removed. No DEVICE
object remains.

`OBJECT_VERSION` will also be removed. v1 does not attempt semantic forward
compatibility inside `objects-v1/`; the namespace itself defines the grammar.
Conceptually, `objects-v1/<object-id>` contains either a valid exact v1 TOKEN or
not v1 semantic state. There is no future TOKEN semantic version embedded inside
this family.

> Forward compatibility is defined between object-family specifications, not within the v1 semantic object grammar.

A future specification may introduce `objects-v2/` or another family. That
specification is responsible for defining its relationship to v1 clients.
Possible strategies include hard upgrade, read compatibility, dual writing,
explicit migration, or no coexistence. v1 does not choose among them.

## Intended r16 handling of invalid objects-v1 entries

An entry under `objects-v1/` contributes TOKEN state only if all applicable v1
checks succeed, including:

- valid candidate filename;
- correct physical object size;
- successful envelope authentication;
- valid padding/length;
- matching keyed OBJECT_ID;
- exact canonical v1 TOKEN grammar;
- valid field widths/ranges/text;
- valid TOKEN semantics.

An entry that does not validate as an exact v1 TOKEN contributes no v1 TOKEN
semantic state. Examples include random bytes, failed AEAD, malformed TLV,
unknown TLV, a missing mandatory field, wrong field order, an extra trailing
field, and authenticated but non-v1 plaintext.

These may produce local diagnostics. They do not become special future semantic
objects merely because authenticated bytes exist.

## Missing or invalid parents

If a valid TOKEN lists parent object ID P and P is not presently available as a
valid v1 TOKEN of the same TOKEN_ID, the parent edge remains unresolved. The
child itself may remain a valid complete TOKEN.

Implementations MUST NOT invent ancestry, skip through the missing parent, infer
grandparents, or reconstruct causal relationships from partial history. When the
valid parent later appears, ordinary graph recomputation may resolve the edge.

For precision about the baseline, r15 Section 22 distinguishes a rejected edge
when an accepted object at the exact ID has the wrong type/identity from an
unresolved edge for missing or invalid evidence. The direction recorded above
uses the requested same-TOKEN_ID availability condition for r16; it does not
change that current r15 distinction.

## Settled: remove opaque/future semantic machinery

The future r16 rewrite is expected to remove the v1 concepts whose only purpose
was semantic-version/type forward compatibility:

- routable future TOKEN and routable future DEVICE;
- `OPAQUE_ROUTABLE` and `OPAQUE_UNSCOPED`;
- future semantic version dispatch and future DEVICE dispatch;
- compatible opaque reprocessing;
- durable opaque semantic evidence;
- current opaque TOKEN authorship blocks;
- opaque DEVICE presentation/authorship blocks.

This checkpoint records that design decision; it does not edit the normative
specification or vectors.

### Unknown sibling namespaces

v1 owns `objects-v1/`. Other family namespaces are not parsed as v1 TOKEN objects.
Later specifications define their own interoperability requirements. This
checkpoint defines no further future-family coexistence policy and introduces
no mandatory blocking merely because another namespace exists.

## Outer object cryptography is not redesigned

The following remain outside the simplification decisions made so far and are
retained for this checkpoint:

- keyed OBJECT_ID;
- HKDF object key derivation;
- AES-GCM envelope;
- 1024-byte fixed encrypted object size;
- 1006-byte semantic plaintext maximum;
- zero padding model;
- filename = object ID convention.

Only provenance-signature-related cryptography is settled for removal.

## Open questions before normative r16

### A. Graph-integrity anomaly machinery

Determine whether authenticated valid TOKEN cycles and conflicting valid objects
with the same keyed OBJECT_ID are cryptographically infeasible enough that
cycles need not be a normal protocol state, global same-ID semantic
contradictions need not create continuity machinery, and implementations need
only defend against them locally. **No decision yet.**

### B. Confirmation freshness

Determine whether confirmation should bind only to the exact current accepted
TOKEN semantic context at publication time, eliminating process-local remembered
“learned then disappeared” generations. **No decision yet.**

### C. Parent maximum

Two live alternatives remain:

```text
permit every parent count that physically fits
```

versus:

```text
normative maximum = 4 parents
5+ requires fold
```

Every legal TOKEN can always fit at least four. **No decision yet.**

### D. Candidate-use policy

Determine how much explicit candidate/historical credential-use policy belongs
in the normative protocol versus the application/API layer. **No decision yet.**

### E. Advisory history

Determine whether the optional advisory-history capability should be deleted
entirely from v1. **No decision yet.**

### F. VAULT_BINDING/local authoritative state

Re-evaluate the exact local-binding requirement after removal of device-key
custody. **No decision yet.**

### G. Platform-specific implementation requirements

Re-evaluate whether desktop Java needs `platform-linux` once private signing-key
custody and stronger same-UID filesystem defenses are no longer protocol
requirements. **No decision yet.** In current r15, stronger hostile-local-filesystem
race hardening is already optional; private signing-key custody remains required.
This question does not imply otherwise or select a Java implementation strategy.

## Explicit non-decisions

This checkpoint does not yet decide:

- whether max parents is exactly 4;
- confirmation simplification;
- candidate-use simplification;
- advisory-history removal;
- graph cycle/integrity simplification;
- VAULT_BINDING simplification;
- portable NIO versus platform-linux;
- whether fixed 1024-byte objects remain optimal;
- garbage collection (GC);
- future-family migration;
- public Java API structure.

## Expected r16 impact (informational)

If the settled direction is carried into normative r16, it should eliminate
large portions of the current specification concerning:

```text
DEVICE identity and presentation
P-256 / ECDSA provenance
signature context
signature validation/provenance statuses
DEVICE graph and folding
initial DEVICE success gate
future semantic versions inside objects-v1
opaque routable/unscoped semantic states
nested credential TLV
```

The normative rewrite and corpus regeneration are subsequent work after the
remaining decisions. This artifact changes no protocol semantics, normative
revision labels, vectors, schemas, requirements profiles, conformance code,
generators, generated fixtures, or release/tag material.
