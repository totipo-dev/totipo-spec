# Totipo v1/r9 — Forward-Compatibility Consistency and Simplification Review

## Result

The r9 rolling-upgrade change is internally coherent and does not require a new recovery mechanism.

The core state model remains:

```text
durable authenticated routing graph
+
whether each current supported TOKEN value is readable
```

r9 adds only one dimension to a graph node:

```text
SUPPORTED_VALID
OPAQUE_ROUTABLE
```

plus vault-level:

```text
OPAQUE_UNSCOPED
```

No accepted-head set, superseded-ID set, generation marker, handoff record, or special future-version recovery object is needed.

## Scenarios checked

### Future TOKEN causally replaces v1 TOKEN

```text
A(v1) -> C(v2 opaque)
```

v1 knows:

```text
current = C
TOKEN_ID is known
parents are known
value body is opaque
```

Result:

- ordinary current use for that token is unavailable;
- A remains an explicit supported candidate;
- unrelated tokens remain ordinary;
- v1 writing for that token is blocked.

Correct.

### Concurrent supported and future TOKEN

```text
       B(v1)
      /
A ----
      \
       C(v2 opaque)
```

Result:

- both B and C are current;
- C's value is opaque;
- B may be explicitly used as a candidate;
- v1 does not pretend it can converge over C.

Correct.

### Later supported descendant

```text
C(v2 opaque) -> R(v1 supported)
```

Once R is learned:

```text
current = R
C is historical
```

If the remaining current frontier is supported/readable/unambiguous, old v1 ordinary use resumes.

This is important: an old client never needs to parse C's body to understand that a later routable descendant superseded it.

Correct.

### Future version number does not order state

```text
A(v1) || B(v9)
```

remains concurrent without a parent relation.

Correct.

### Unknown future object type

If the authenticated object cannot be scoped to TOKEN/DEVICE:

```text
OPAQUE_UNSCOPED
```

Result:

- authoritative vault operations pause because v1 cannot determine scope;
- explicit supported candidate OTP use remains available.

This preserves degraded availability without pretending the object is irrelevant.

Correct.

### Future DEVICE

Explicit serialized `DEVICE_ID` lets v1 route future DEVICE topology without understanding future key/presentation body.

Opaque DEVICE state affects presentation/provenance only.

Correct.

## Simplification check

The common readiness predicates are useful and should remain:

```text
BASE_OPERATION_SAFE
AUTHORITATIVE_VAULT_READY
CANDIDATE_USE_READY
```

They remove repeated gate lists and cleanly encode the philosophy:

```text
trust failure:
    block trust-dependent operations

semantic uncertainty:
    degrade capability
    preserve explicit candidate access
```

Do not merge `OPAQUE_ROUTABLE` with missing-byte state:

- missing supported bytes means the semantic version is understood but value bytes are unavailable;
- opaque future state means bytes may be present, but the semantic version is not understood.

Both make current value incomplete, but the user-facing explanation is different.

Do not remove `OPAQUE_UNSCOPED`; it is the necessary conservative case for a future object an old client cannot route.

## Byte impact

TOKEN bytes are unchanged from r8.

DEVICE adds:

```text
0200 DEVICE_ID[32]
```

cost:

```text
4-byte TLV header + 32 bytes = 36 bytes
```

Thus maximum-display DEVICE writer planning changes:

```text
14 parents = 973 bytes
15 parents = 1009 bytes -> fold
```

## Verdict

r9 is a sound forward-compatibility improvement.

It better matches Totipo's degradation philosophy:

> Authenticated uncertainty should reduce assurance/capability where possible, rather than automatically making authenticated candidate material unusable.

Recommendation: freeze this semantic design, perform one mechanical vector-schema review, then start exact v1 byte vectors.
