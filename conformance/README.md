# Go v1 reference/conformance consumer

This module supports Go 1.23 or later and is the single in-repository reference
consumer for Totipo v1/r9. It is deliberately not a production client.

From the repository root:

```sh
make test
make conformance
make verify
make race
make fuzz
```

The command also supports `-root /path/to/repository` and `-verify-only`. The
module tests execute the manifest corpus, verify its moving profile, and add
parser, crypto rejection, signing, and graph invariants. Bounded fuzzing covers
semantic parsing, encrypted envelopes, and graph arrival/disappearance order.

Packages separate fixed TLV framing, semantic object parsing, cryptographic
processing, RFC 6238 TOTP computation, durable graph semantics, and vector IO. `object.Dispatch` takes
already authenticated semantic bytes. Storage consumers must first call
`cryptov1.Keys.Open`, which authenticates the envelope, canonical padding, and
keyed object ID. Supported malformed bodies are invalid; future bodies are
never interpreted as v1.

`graph.State` models an established client's persisted authenticated knowledge
and independently tracked synchronized value availability. It is an in-memory
semantic evaluator with symbolic object IDs. It models persistence failure and
continuity gates, but does not implement storage flushes, crash recovery, pending
vault establishment, operation freshness epochs, or publication transactions.
Its author result means the frontier permits authoring after any required user
confirmation; it is not a complete writer or a confirmation bypass.

Fixture generation is a separate command described in [FORMAT.md](../vectors/FORMAT.md).
Tests never invoke it. The moving corpus needs an independent live implementation
before RC freeze. See the [implementation report](../review/V1_VECTOR_IMPLEMENTATION_REPORT.md)
for exact coverage and limitations.
