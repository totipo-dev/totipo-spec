# Totipo v1/r8 — Simplification Review

**Goal:** look specifically for accidental complexity that can be removed without weakening r8 behavior.

## Result

The largest architectural simplification has already happened:

```text
r5:
    ACCEPTED_HEAD_IDS
    SUPERSEDED_ID_SET
    special local reaffirmation bookkeeping

r8:
    append-only durable known graph
    +
    current value availability
```

That simplification is holding up.

I do **not** recommend another architectural rewrite before vectors.

There are, however, a few behavior-preserving simplifications worth applying in the final cleanup revision.

---

## S1. Centralize operation readiness predicates

The same safety conditions are repeated across §§24, 32, 33, 35, 37, 38, and 55.

That repetition already caused the stale §56 contradiction.

Define common conceptual predicates once:

```text
BASE_OPERATION_SAFE =
    LOCAL_CONTINUITY_UNKNOWN == false
    and KNOWLEDGE_PERSISTENCE_BLOCKED == false
    and no authenticated unsupported-version block

AUTHORITATIVE_READY =
    BASE_OPERATION_SAFE
    and DISCOVERY_STATE == READY

CANDIDATE_USE_READY =
    BASE_OPERATION_SAFE
```

Then operation rules become:

```text
ordinary current TOTP:
    AUTHORITATIVE_READY
    + complete unambiguous LIVE current state

semantic authorship:
    AUTHORITATIVE_READY
    + operation-specific confirmation/freshness rules

explicit candidate TOTP:
    CANDIDATE_USE_READY
    + exact selected available assertion-valid LIVE TOKEN
    + incomplete-discovery warning when DISCOVERY_STATE != READY
```

This does not change semantics.

It substantially lowers the chance that future sections accidentally disagree about gates.

**Recommendation: adopt in r9.**

---

## S2. Make §38 a compact consumption matrix instead of re-specifying §35

§35 already defines candidate selection and its warnings.

§38 largely restates the gates.

A simpler normative shape would be:

```text
Mode                 Discovery   Current-state requirement
-------------------  ----------  -------------------------
Ordinary current     READY       complete + unambiguous + LIVE
Explicit candidate   any         exact available valid LIVE candidate
```

with both modes referencing the common readiness predicates from S1.

This removes duplicate prose while keeping one canonical candidate-use definition.

**Recommendation: adopt if convenient; not required before vectors.**

---

## S3. Keep parent-edge diagnostics, but make causality depend on one predicate

Current edge status is:

```text
RESOLVED
UNRESOLVED
REJECTED
```

Only `RESOLVED` affects ancestry.

`UNRESOLVED` and `REJECTED` differ only diagnostically.

Implementations can therefore structure this as:

```text
RESOLVES_CAUSALLY(child, parent_id) -> bool

EDGE_DIAGNOSTIC ->
    missing/unavailable
    wrong type
    wrong logical ID
    invalid current bytes
    ...
```

The wire/spec can retain the three labels for audit clarity; current-head algorithms only need the boolean resolved relation.

This reduces implementation coupling without another protocol revision.

**Recommendation: implementation guidance only. Do not churn the spec terminology before vectors.**

---

## S4. Do not merge DISCOVERY_STATE and KNOWLEDGE_PERSISTENCE_BLOCKED

At first glance these look like two “not ready” states.

They should stay separate:

```text
DISCOVERY incomplete:
    unprocessed/unauthenticated uncertainty
    candidate use may continue

knowledge persistence blocked:
    safety-relevant valid information was already learned
    but cannot be durably remembered
    candidate use must stop
```

Combining them would either over-block candidate mode again or under-protect learned state.

**Recommendation: keep separate.**

---

## S5. Do not remove the confirmation-context epoch

The durable graph is monotonic, which might make the epoch appear redundant.

It is still useful because **value availability is not monotonic**:

```text
C unavailable
    -> user confirms based on missing value
C becomes available and reveals information
C disappears again
```

The visible final availability set may return to its earlier shape, but the client has observed a safety-relevant transition.

The monotonic epoch prevents reuse of the old confirmation.

**Recommendation: keep.**

---

## S6. Keep `AUTHOR_TIME` required-with-zero-sentinel

Alternatives would be:

- optional TLV;
- nullable encoding;
- separate object-specific tags.

Required common:

```text
AUTHOR_TIME u64
0 = unknown
```

is simpler to parse and canonicalize.

Its non-causal rules are explicit enough to prevent it becoming a logical clock.

**Recommendation: keep.**

---

## S7. Keep DEVICE history as a DAG

DEVICE state is presentation-only, but concurrent renames still need deterministic/conflict-preserving behavior.

Removing DEVICE parents would require introducing another ordering/conflict mechanism.

The existing graph reuses the TOKEN ancestry machinery and is cheap.

**Recommendation: keep.**

---

## S8. Keep the durable graph minimal

The current minimum records are good:

```text
TOKEN:
    object ID
    token ID
    parents
    author device ID
    author time

DEVICE:
    object ID
    device ID
    parents
    public key
    author time
```

Do not add:

- TOKEN_VALUE;
- secrets;
- display names;
- signatures

to mandatory durable security memory.

Exact immutable object caching remains optional.

This keeps local security state small and avoids creating a second mandatory secret store.

**Recommendation: keep.**

---

## S9. One terminology cleanup

Use:

```text
whole-state convergence
```

as the primary protocol/application term.

Use:

```text
reaffirmation
```

only when explaining the user action of confirming a state despite unavailable/conflicting history.

This reduces the impression that “reaffirmation” is a special wire operation.

It is always ordinary TOKEN authorship.

**Recommendation: editorial cleanup only.**

---

# Simplification verdict

No architectural feature should be removed before vector freeze.

The best final simplification is **centralizing readiness predicates** and reducing duplicated operation-gate prose.

After that, the implementation model is compact:

```text
1. Validate immutable objects.
2. Durably remember authenticated topology.
3. Derive current heads from resolved ancestry.
4. Track whether each current value is available.
5. Ordinary use only for one complete unambiguous LIVE state.
6. Otherwise let the user explicitly select available candidates.
7. User-confirmed convergence publishes another ordinary complete TOKEN.
```

That is a substantially simpler client model than v0 and the early v1 drafts.
