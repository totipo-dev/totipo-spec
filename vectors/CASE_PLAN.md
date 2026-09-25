# Initial v1 Vector Case Plan

Create these cases in roughly this order. IDs should not be renamed after a consumer starts depending on them.

## Routing / forward compatibility

```text
v1.routing.token-v1.001
v1.routing.token-future-opaque.001
v1.routing.device-v1.001
v1.routing.device-future-opaque.001
v1.routing.unknown-type-unscoped.001
v1.routing.future-token-malformed-prefix.001
v1.routing.future-device-malformed-prefix.001
```

## Canonical encoding

```text
v1.encoding.token-root.001
v1.encoding.device-root.001
v1.encoding.parent-order.001
v1.encoding.parent-count-mismatch.001
v1.encoding.duplicate-nonrepeatable.001
v1.encoding.author-time-zero.001
v1.encoding.author-time-u64max.001
v1.encoding.utf8-boundary.001
```

## DEVICE identity / provenance

```text
v1.device.explicit-id.001
v1.device.id-mismatch.001
v1.provenance.token-verified.001
v1.provenance.token-rejected.001
v1.provenance.token-unresolved.001
v1.provenance.der-short-valid.001
v1.provenance.der-max-valid.001
```

## Capacity planning

```text
v1.size.token-max-4.001
v1.size.token-max-5-fold.001
v1.size.device-max-14.001
v1.size.device-max-15-fold.001
v1.size.short-der-no-extra-parent.001
```

## Full cryptographic objects

```text
v1.crypto.token-root.001
v1.crypto.token-child.001
v1.crypto.device-root.001
v1.crypto.future-token-opaque.001
v1.crypto.future-device-opaque.001
```

Each full crypto case should expose, as applicable:

```text
semantic bytes before signature
signature input
fixed signature fixture
final canonical semantic bytes
OBJECT_ID
derived object key
nonce
AAD
plaintext length
padded plaintext
ciphertext
GCM tag
final 1024-byte object
```

## Semantic graph / availability

```text
v1.graph.sequential.001
v1.graph.equal-concurrent.001
v1.graph.conflicting-concurrent.001
v1.graph.missing-current-value.001
v1.graph.intermediate-disappears.001
v1.graph.late-parent.001
v1.graph.cycle-integrity-failure.001
v1.graph.global-object-id-conflict.001
```

## Future versions / rolling upgrades

```text
v1.future.scoped-token.001
v1.future.concurrent-supported-opaque.001
v1.future.unrelated-token.001
v1.future.supported-descendant.001
v1.future.version-does-not-order.001
v1.future.device-presentation.001
v1.future.unscoped-authoritative-block.001
v1.future.unscoped-candidate-use.001
v1.future.disappearance-retains-routing.001
```

## Candidate credential use

```text
v1.candidate.conflict-a.001
v1.candidate.current-peer-missing.001
v1.candidate.historical-current-missing.001
v1.candidate.opaque-current.001
v1.candidate.discovery-incomplete.001
v1.candidate.persistence-block.001
v1.candidate.continuity-block.001
```

## Timestamp

```text
v1.timestamp.zero.001
v1.timestamp.normal.001
v1.timestamp.i64max.001
v1.timestamp.u64max.001
v1.timestamp.equal-value-different-times.001
v1.timestamp.fold-common-time.001
```
