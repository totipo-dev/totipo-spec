# Totipo v1/r8 — Final Consistency and Adversarial Review

**Artifact reviewed:** `totipo-vault-format-v1-r8.md`  
**Purpose:** regression-oriented review after the r8 cycle/global-ID, discovery, signature-capacity, and `AUTHOR_TIME` changes.  
**Disposition:** core architecture holds. No wire/cryptographic redesign is indicated. A small **r9 editorial/conformance-freeze cleanup** is recommended before generating normative byte vectors.

---

## Executive result

The following r8 architectural choices remain coherent under this pass:

- complete vault-authenticated TOKEN assertions;
- durable authenticated topology;
- current value availability separated from topology;
- whole-state conflict semantics;
- explicit candidate credential use;
- candidate use during incomplete discovery but not through stronger safety gates;
- convergence/reaffirmation by ordinary TOKEN over all known current heads;
- P-256 provenance separate from TOKEN authority;
- informational `AUTHOR_TIME` excluded from causality/state equality;
- deterministic 72-byte signature capacity reservation;
- global object-ID consistency and resolved-cycle fail-closed handling;
- bounded TOKEN/DEVICE folding.

No counterexample found requires reintroducing:

- accepted-head sets;
- superseded-ID sets;
- recovery generations;
- reference-coverage relations;
- field-level merge;
- lifecycle witness machinery;
- a special recovery wire object.

---

# Blocking before vector freeze

## B1. Conformance text contradicts r8 candidate-use semantics

Normative r8 correctly says:

```text
DISCOVERY_STATE = PROCESSING_INCOMPLETE
```

may still permit explicit use of an already authenticated LIVE candidate, provided:

```text
LOCAL_CONTINUITY_UNKNOWN = false
KNOWLEDGE_PERSISTENCE_BLOCKED = false
no authenticated unsupported-version block exists
```

But §56 still requires a case phrased as:

> candidate use ... is blocked by unsupported-version, continuity, **discovery**, or knowledge-persistence gates

That is stale r7 behavior.

### Fix

Replace it with two independent conformance requirements:

```text
candidate use is blocked by:
    unsupported-version
    continuity unknown
    knowledge-persistence failure

PROCESSING_INCOMPLETE:
    candidate use may proceed
    only from already-authenticated LIVE candidate
    with mandatory incomplete-discovery warning
```

This is not a protocol change; it aligns evidence with existing r8 semantics.

---

## B2. Required conformance evidence does not yet freeze the 72-byte capacity-planning rule

§45 now normatively requires:

```text
SIGNATURE_RESERVED_BYTES = 72
```

and forbids:

- exploiting a shorter actual DER signature for extra fan-in;
- retrying randomized signatures to fit an otherwise over-capacity object.

But §56 does not currently require executable evidence for those rules.

The DEVICE boundary is particularly important:

```text
max DISPLAY_NAME + 15 parents:
    reserved size = 973       // fits

max DISPLAY_NAME + 16 parents:
    reserved size = 1009      // MUST fold

same 16-parent object with a 69-byte actual signature:
    serialized size = 1006    // physically fits, but writer still MUST fold
```

### Fix

Add conformance cases for:

```text
writer capacity uses 72-byte reservation
short DER does not increase selected parent count
retry-to-fit behavior prohibited
max-size DEVICE 15 parents accepted as one planned object
max-size DEVICE 16 parents requires fold
TOKEN and DEVICE fold progress boundaries
```

This should be frozen before byte vectors so vector generators and independent implementations make the same object-shape choices.

---

# High-priority cleanup

## H1. Unsupported-version conformance list omits candidate-use blocking

§37 explicitly blocks candidate credential use when authenticated unsupported-version evidence exists.

§56's `Unsupported versions` evidence currently lists ordinary authorship and ordinary TOTP consumption but not candidate use.

Add:

```text
authenticated unsupported semantic version
    -> explicit candidate credential use blocked
```

---

## H2. Object-model overview contains stale terminology

§2 still says:

> Parent IDs are used for causality, convergence, and local reaffirmation

The current model is more accurately:

> Parent IDs express claimed causal incorporation and support durable known-graph causality and whole-state convergence/reaffirmation.

This is editorial, but §2 is where implementers form their mental model.

The overview should also mention common informational `AUTHOR_TIME`, while repeating that it is non-causal.

---

## H3. Global object-ID uniqueness should explicitly include durable unsupported-version evidence

§22.7 is broad enough in principle: one `OBJECT_ID` identifies one immutable canonical semantic plaintext.

For implementation/conformance clarity, state that this identity namespace also includes authenticated unsupported-version observations.

Thus if a client durably knows:

```text
ID X = authenticated unsupported-version semantic object
```

a later supposedly valid v1 TOKEN/DEVICE at the same `X` is not an alternate interpretation. It is a cryptographic/local-integrity failure.

This is collision-level defensive behavior, not an expected runtime condition.

---

## H4. Baseline rebuild should explicitly fail if graph-integrity checks fail

§34.3 already requires a resource-complete scan and normal validation.

Add one sentence:

```text
Any cycle/global-ID/durable-record integrity failure encountered while building
the replacement baseline prevents re-establishment from completing.
```

This is already implied by §§22/24, but explicit recovery-path wording is safer.

---

# Adversarial checks that continue to pass

## Missing current bytes

```text
A -> B -> C
```

If C bytes disappear:

```text
known current = C
available historical candidate = B
```

B does not become current.

The user may explicitly generate from B and may author:

```text
C -> R
```

using B only as recovery material for the confirmed complete state.

Sound.

---

## Concurrent candidates

```text
B  ||  C
```

If both are available and differ:

- ordinary use blocked;
- user may explicitly generate from either LIVE candidate;
- candidate selection does not resolve the conflict;
- convergence parents both.

If C bytes disappear:

- B remains selectable;
- C remains current and value-unavailable;
- reaffirmation still parents B and C as applicable to the current frontier.

Sound.

---

## Incomplete-discovery availability attack

An untrusted storage provider may flood the namespace with unreadable/unprocessed candidates.

r8 behavior is now appropriately split:

```text
ordinary authoritative use:
    blocked until READY

semantic authorship:
    blocked until READY

explicit candidate use:
    may continue from an already-authenticated LIVE candidate
    with warning
```

Authenticated unsupported-version evidence, continuity failure, and durable-knowledge persistence failure remain stronger blockers.

Sound.

---

## Current TOMBSTONE plus incomplete discovery

An old historical LIVE TOKEN can be reachable through explicit candidate/recovery UI while discovery is incomplete.

This does **not** make it current or modify the TOMBSTONE.

Because candidate use is explicit, non-authoritative, and the UI must show candidate status, this is consistent with the user-agency model. Implementations SHOULD present current candidates first and historical candidates only on deliberate recovery selection, as §35.5 already suggests.

No protocol change recommended.

---

## `AUTHOR_TIME`

All of the following remain semantically harmless:

```text
0
normal present-day time
far-future value
0x7fffffffffffffff
0xffffffffffffffff
```

They:

- change canonical bytes/object ID;
- are authenticated/signed metadata;
- do not change TOKEN_VALUE;
- do not create causality;
- do not choose heads;
- do not resolve conflicts.

The raw unsigned representation requirement prevents signed/time-API overflow from affecting validity.

Sound.

---

## Fold timestamp behavior

One confirmed action uses one `AUTHOR_TIME` across its fold intermediates.

If the fold is interrupted and a later fresh operation is started, that later operation may use a different time.

Because time is non-causal, neither case changes graph semantics.

Sound.

---

## Cycle/global-ID defenses

Resolved cycles and incompatible reuse of one global object ID fail closed into local continuity recovery instead of producing ordinary state.

That is the correct handling for conditions that should be cryptographically extraordinary under the object-ID construction.

Sound.

---

## Signature reservation

The reserved 72-byte size affects **writer planning only**.

The real DER signature remains serialized at its actual legal length and therefore participates normally in canonical bytes and object ID.

This cleanly separates:

```text
deterministic object-shape planning
from
randomized ECDSA encoding length
```

Sound.

---

# Mechanical observations

The normative body no longer contains active r5/r6 accepted/superseded machinery except explicit statements that those concepts are absent.

Historical revision entries correctly retain their original terminology.

One stale r8 conformance sentence remains (B1).

The main size formulas remain internally correct:

```text
TOKEN_BYTES_RESERVED =
    215 + issuer + account + secret + 36*p

DEVICE_BYTES_RESERVED =
    177 + display_name + 36*p
```

Maximum fields:

```text
TOKEN p=4 -> 999
TOKEN p=5 -> 1035

DEVICE p=15 -> 973
DEVICE p=16 -> 1009
```

---

# Recommendation

Cut a very small **v1/r9** with **no byte or semantic behavior changes**:

1. fix §56 candidate/discovery contradiction;
2. add required signature-reservation / DEVICE capacity conformance cases;
3. add candidate use to unsupported-version conformance evidence;
4. clean §2 terminology / mention `AUTHOR_TIME`;
5. explicitly include unsupported-version object IDs in the global immutable identity rule;
6. explicitly fail baseline re-establishment on graph-integrity failure.

Then perform a mechanical stale-language check only.

If that passes, freeze the semantic model and begin normative byte vectors.
