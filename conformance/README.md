# Go v1 reference/conformance consumer

This module supports Go 1.23 or later and is the single in-repository reference
consumer for Totipo v1/r13. It is deliberately not a production client.

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
before RC freeze.

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
observations into the same graph evaluator. See the
[envelope-family review](../review/V1_R10_ENVELOPE_FAMILY_REVIEW.md) for the namespace
and compatibility rationale.

The r12 model keeps DEVICE causality separate from authenticated friendly names:
rename includes all supported heads, while only verified readable heads provide
names. Remote byte corruption retains durable nodes; local security-memory
corruption blocks all use. Opaque-unscoped disappearance does not clear evidence.
Explicit reset builds a complete replacement baseline and deliberately abandons
prior continuity guarantees. A separate small publication workflow checks the
first DEVICE advertisement gate; it is not a production publication system.

The r13 concrete storage path retains `OBJECT_ID` and the exact authenticated
1024-byte encrypted object for unscoped evidence. `LearnOpaque` authenticates
before insertion; `ReprocessOpaque` reauthenticates retained bytes before invoking
a compatible classifier. Existing symbolic graph cases keep their abstraction;
they alone do not prove byte retention. Reprocessing cannot bypass this boundary
with a digest-only reclassification of a concrete retained record.

`graph.Provenance` retains known TOKEN semantic evidence and recomputes attribution
when matching DEVICE/restored public-key material arrives. It updates provenance
without altering TOKEN values, topology, or authority. The baseline scan model
requires terminal classifications for every member of a fixed snapshot. Explicit
reset can discard missing intermediate ancestry and expose historical assertions
as current conflicts; a live UI must warn before asking for confirmation.
