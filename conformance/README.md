# Go v1 reference/conformance consumer

This module supports Go 1.23 or later and is the single in-repository reference
consumer for Totipo v1/r11. It is deliberately not a production client.

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

## Storage-family environment model

`internal/storage` evaluates an observed filesystem environment. It selects only
direct regular-file children of `objects-v1/` with 64 lowercase hexadecimal names.
It does not normalize traversal paths or follow symlink entries. Unknown siblings,
nested paths and special files never invoke their content reader. Wrong-size or
unauthenticated candidates are invalid storage observations and cannot add opaque
semantic evidence. A read error or unsafe family directory is an error requiring
incomplete discovery, not successful classification.

This is a small reference environment evaluator, not a host-filesystem adapter.
A live adapter must obtain entry kinds and bounded bytes through stable no-follow
handles, prevent namespace rebinding/escape, and honor the returned errors. No
production crash/persistence or filesystem-race guarantee is claimed here.

`OBJECT_VERSION` versions semantics within the fixed v1 family. `objects-v2/` is
an illustrative separate family and is ignored. Rolling compatibility requires
authenticated compatibility assertions in `objects-v1/`; unknown sibling namespace
names are not authenticated future-version evidence. Storage cases reuse existing
envelope fixtures, authenticate them, and feed only supported/opaque authenticated
observations into the same graph evaluator. See the [r10 report](../review/V1_R10_INTEGRATION_REPORT.md).
